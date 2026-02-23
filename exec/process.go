package exec

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// 进程的结构体
type Process struct {
	CurrentProcess *exec.Cmd  // 当前进程对象
	ProcessMu      sync.Mutex // 进程锁
	IsRunning      bool       // 进程是否在运行
	RestartCount   int        // 已重试次数
	FileType       string     // 可执行文件类型
	Pid            int        // 进程ID
	ExitInfo       ProcessExit
	stoppedByReq   bool
}

// 进程退出信息
type ProcessExit struct {
	LastExitCode     int       // 进程最后退出代码
	LastExitTime     time.Time // 进程最后退出时间
	LastExitBySignal bool      // 进程是否被信号终止
	LastExitError    string    // 进程最后退出的错误信息
	StoppedByRequest bool      // 进程是否是被请求停止的
}

// Hooks 聚合了进程启动/退出时需要的回调和策略
// RunPre / CreateCmd / ApplyEnv 由上层提供；日志与输出也交由上层处理，保持 exec 包职责单一。
type Hooks struct {
	RunPre       func() error
	CreateCmd    func() (*exec.Cmd, error)
	ApplyEnv     func(*exec.Cmd) error
	OnStarted    func(*exec.Cmd)
	OnStdout     func(io.ReadCloser)
	OnStderr     func(io.ReadCloser)
	OnExit       func(ProcessExit)
	Logf         func(string, ...interface{})
	AutoRestart  bool
	RestartTimes int
	RestartDelay time.Duration
}

// StartManagedProcess 启动子进程并在内部启动等待/回调逻辑
func (p *Process) StartManagedProcess(h Hooks) error {
	p.ProcessMu.Lock()
	defer p.ProcessMu.Unlock()

	if p.IsRunning && p.CurrentProcess != nil {
		return fmt.Errorf("进程已在运行，PID: %d", p.CurrentProcess.Process.Pid)
	}
	p.stoppedByReq = false

	if h.RunPre != nil {
		if err := h.RunPre(); err != nil {
			return err
		}
	}

	var err error
	var cmd *exec.Cmd
	if h.CreateCmd != nil {
		cmd, err = h.CreateCmd()
		if err != nil {
			return err
		}
	}
	if cmd == nil {
		return fmt.Errorf("CreateCmd 未提供")
	}
	if h.ApplyEnv != nil {
		if err := h.ApplyEnv(cmd); err != nil {
			return err
		}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动进程失败: %v", err)
	}

	p.CurrentProcess = cmd
	p.IsRunning = true
	p.Pid = cmd.Process.Pid
	if h.OnStarted != nil {
		h.OnStarted(cmd)
	}
	if h.OnStdout != nil {
		go h.OnStdout(stdout)
	}
	if h.OnStderr != nil {
		go h.OnStderr(stderr)
	}

	// 等待和自动重启逻辑放在单独 goroutine
	go p.monitor(cmd, h)
	return nil
}

// StopManagedProcess 尝试优雅停止进程
func (p *Process) StopManagedProcess(timeout time.Duration) error {
	p.ProcessMu.Lock()
	if !p.IsRunning || p.CurrentProcess == nil {
		p.ProcessMu.Unlock()
		return fmt.Errorf("进程未运行")
	}
	p.stoppedByReq = true
	cmd := p.CurrentProcess
	p.ProcessMu.Unlock()

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// 尝试继续等待，即使发送信号失败也不立即返回
	}

	done := make(chan error, 1)
	go func(c *exec.Cmd) {
		done <- c.Wait()
	}(cmd)

	select {
	case <-time.After(timeout):
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("强制终止进程失败: %v", err)
		}
		<-done
	case err := <-done:
		if err != nil {
			// 非致命，记录由外层完成
		}
	}

	p.ProcessMu.Lock()
	p.IsRunning = false
	p.CurrentProcess = nil
	p.ProcessMu.Unlock()
	return nil
}

// monitor 等待进程退出并回调 OnExit；如果需要，自动重启
func (p *Process) monitor(cmd *exec.Cmd, h Hooks) {
	err := cmd.Wait()

	p.ProcessMu.Lock()
	p.IsRunning = false
	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}
	exitInfo := ProcessExit{
		LastExitTime: time.Now(),
		LastExitCode: 0,
	}
	if err != nil {
		exitInfo.LastExitError = err.Error()
		if ee, ok := err.(*exec.ExitError); ok {
			ps := ee.ProcessState
			exitInfo.LastExitCode = ps.ExitCode()
			if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				exitInfo.LastExitBySignal = true
			}
		}
	}
	exitInfo.StoppedByRequest = p.stoppedByReq
	p.ExitInfo = exitInfo
	p.ProcessMu.Unlock()

	if h.OnExit != nil {
		h.OnExit(exitInfo)
	}

	if h.AutoRestart && (h.RestartTimes < 0 || p.RestartCount < h.RestartTimes) {
		p.RestartCount++
		delay := h.RestartDelay
		if delay == 0 {
			delay = 2 * time.Second
		}
		if h.Logf != nil {
			h.Logf("准备重启进程，当前重试次数: %d/%d", p.RestartCount, h.RestartTimes)
		}
		time.Sleep(delay)
		_ = p.StartManagedProcess(h)
	} else if h.AutoRestart && h.Logf != nil {
		h.Logf("已达到最大重试次数 %d，停止自动重启 (pid=%d)", h.RestartTimes, pid)
	}
}
