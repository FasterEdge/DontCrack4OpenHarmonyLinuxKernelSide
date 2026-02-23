package core

// FasterEdge - 对称、可靠、安全的多场景边缘计算框架
// https://github.com/FasterEdge
// DontCrack开源鸿蒙(OpenHarmony)Linux内核专用版
// https://github.com/FasterEdge/DontCrack4OpenHarmonyLinuxKernelSide
// tyza66
// https://github.com/tyza66
import (
	"DontCrash/config"
	pmexec "DontCrash/exec"
	dclog "DontCrash/log"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	osexec "os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const logo = `
 ________  ________           _______          ________  ___  ___     
|\   ___ \|\   ____\         /  ___  \        |\   __  \|\  \|\  \    
\ \  \_|\ \ \  \___|        /__/|_/  /|       \ \  \|\  \ \  \\\  \   
 \ \  \ \\ \ \  \           |__|//  / /        \ \  \\\  \ \   __  \  
  \ \  \_\\ \ \  \____          /  /_/__        \ \  \\\  \ \  \ \  \ 
   \ \_______\ \_______\       |\________\       \ \_______\ \__\ \__\
    \|_______|\|_______|        \|_______|        \|_______|\|__|\|__|
`

const rootMsg = "DontCrash By FasterEdge"

// 管理器运行时状态
var (
	procState pmexec.Process
	// exitState pmexec.ProcessExit // no longer needed separately

	logCache   []string
	logMu      sync.Mutex
	fileLogger *dclog.FileLogger

	shutdownSignal = make(chan struct{})
)

// 初始化并启动管理器
func Start(cfg config.Config) {
	// 检查配置信息是否合法
	err := config.CheckConfig(cfg)
	if err != nil {
		log.Fatalf("配置检查失败: %v\n", err)
		return
	}
	// 输出启动Logo和版本等欢迎信息
	log.Printf(logo)
	log.Printf("DontCrack for OpenHarmony v%s 启动中...\n", cfg.Version)
	log.Printf("管理进程正在管理的程序: %s\n", cfg.Path)

	// 初始化全局状态
	procState.FileType, err = detectFileType(cfg.Path) // 检测文件类型
	if err != nil {
		log.Fatalf("无法检测文件类型: %v\n", err)
		return
	}
	// 若开启文件日志，初始化 fileLogger（目录与保留天数来自配置）
	if cfg.FileLogEnabled {
		procName := filepath.Base(cfg.Path)
		fl, ferr := dclog.NewFileLogger(cfg.LocalLogPath, procName, cfg.LocalLogLifeDay)
		if ferr != nil {
			log.Printf("初始化文件日志失败，继续运行但不落盘: %v", ferr)
		} else {
			fileLogger = fl
			log.Printf("文件日志启用，目录=%s，保留天数=%d", cfg.LocalLogPath, cfg.LocalLogLifeDay)
		}
	}

	log.Printf("检测到文件类型: %s", procState.FileType)
	if cfg.Args != "" {
		log.Printf("程序参数: %s", cfg.Args)
	}
	log.Printf("自动重启: %v", cfg.AutoRestart)
	log.Printf("最大重试次数: %d", cfg.RestartTimes)
	log.Printf("立即启动: %v", cfg.StartNow)
	log.Printf("HTTP端口: %d", cfg.Port)

	// 如果指定立即启动则直接启动
	if cfg.StartNow {
		log.Println("立即启动目标进程...")
		if err := startProcess(cfg); err != nil {
			log.Printf("立即启动失败: %v", err)
		}
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // 监听中断和终止信号

	// 创建HTTP服务器
	mux := http.NewServeMux()
	// 根路径返回简单文本
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, rootMsg)
	})
	// 重置重试计数并启动进程
	mux.HandleFunc("/startup", func(w http.ResponseWriter, r *http.Request) {
		if err := checkPassword(r, cfg.Password); err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintln(w, rootMsg)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		procState.RestartCount = 0
		if err := startProcess(cfg); err != nil {
			http.Error(w, fmt.Sprintf("启动进程失败: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "ok")
	})
	// 心跳，返回当前状态和日志
	mux.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if err := checkPassword(r, cfg.Password); err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintln(w, rootMsg)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		logMu.Lock()
		logsCopy := logCache
		logCache = nil
		logMu.Unlock()

		procState.ProcessMu.Lock()
		state := "stopped"
		pid := 0
		if procState.IsRunning && procState.CurrentProcess != nil {
			state = "running"
			pid = procState.CurrentProcess.Process.Pid
		}
		exit := procState.ExitInfo
		restartCnt := procState.RestartCount
		fileType := procState.FileType
		procPath := cfg.Path
		procState.ProcessMu.Unlock()

		info := HeartbeatInfo{
			Version:          cfg.Version,
			State:            state,
			Info:             "进程管理器正常运行",
			Timestamp:        time.Now().Format("2006-01-02 15:04:05"),
			Logs:             logsCopy,
			ProcessPID:       pid,
			ProcessPath:      procPath,
			RestartCount:     restartCnt,
			FileType:         fileType,
			LastExitCode:     exit.LastExitCode,
			LastExitBySignal: exit.LastExitBySignal,
			LastExitError:    exit.LastExitError,
			ProgramArgs:      cfg.Args,
			ExtraEnvRaw:      cfg.Env,
		}
		if !exit.LastExitTime.IsZero() {
			info.LastExitTime = exit.LastExitTime.Format("2006-01-02 15:04:05")
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, fmt.Sprintf("JSON序列化错误: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})
	// 停止子进程
	mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if err := checkPassword(r, cfg.Password); err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintln(w, rootMsg)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := stopProcess(); err != nil {
			http.Error(w, fmt.Sprintf("停止进程失败: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "ok")
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// 启动HTTP服务器
	go func() {
		log.Printf("HTTP服务器启动在端口 %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP服务器启动失败: %v", err)
		}
	}()

	log.Println("进程管理器运行中...")

	// 等待终止信号
	sig := <-sigChan
	log.Printf("收到信号 %v，开始关闭...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if procState.IsRunning {
		log.Println("正在停止管理的进程...")
		if err := stopProcess(); err != nil {
			log.Printf("停止进程时出错: %v", err)
		}
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP服务器关闭出错: %v", err)
	}

	close(shutdownSignal)
	log.Println("进程管理器已优雅关闭")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// buildChildEnv 合并环境变量并返回 PATH 便于日志打印
func buildChildEnv(envStr string) ([]string, string) {
	base := os.Environ()
	add := parseExtraEnv(envStr)
	env := append([]string{}, base...)
	env = append(env, add...)
	pathVal := ""
	for i := len(env) - 1; i >= 0; i-- {
		if strings.HasPrefix(env[i], "PATH=") {
			pathVal = strings.TrimPrefix(env[i], "PATH=")
			break
		}
	}
	if pathVal == "" {
		pathVal = os.Getenv("PATH")
	}
	return env, pathVal
}

// parseExtraEnv 解析以空格/分号分隔的 KEY=VAL 列表
func parseExtraEnv(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, ";", " ")
	fields := strings.Fields(s)
	var out []string
	for _, f := range fields {
		if !strings.Contains(f, "=") {
			continue
		}
		out = append(out, f)
	}
	return out
}

// runPreCommand 在 /bin/sh 中执行启动前命令
func runPreCommand(cfg config.Config) error {
	if cfg.Pre == "" {
		return nil
	}
	log.Printf("执行启动前命令: %s", cfg.Pre)
	cmd := osexec.Command("/bin/sh", "-c", cfg.Pre)
	cmd.Dir = filepath.Dir(cfg.Path)
	if env, pathVal := buildChildEnv(cfg.Env); len(env) > 0 {
		cmd.Env = env
		log.Printf("PRE环境PATH: %s", pathVal)
	}
	out, err := cmd.CombinedOutput()

	text := strings.ReplaceAll(string(out), "\r\n", "\n")
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	logMu.Lock()
	for _, line := range lines {
		if line == "" {
			continue
		}
		logCache = append(logCache, "[PRE] "+line)
		if len(logCache) > cfg.LogCapacity {
			logCache = logCache[len(logCache)-cfg.LogCapacity:]
		}
	}
	logMu.Unlock()

	if err != nil {
		return fmt.Errorf("pre命令执行失败: %v", err)
	}
	return nil
}

// 根据文件类型构造 osexec.Cmd
func createCommand(cfg config.Config) *osexec.Cmd {
	var cmd *osexec.Cmd
	args := []string{}
	if cfg.Args != "" {
		args = strings.Fields(cfg.Args)
	}

	switch procState.FileType {
	case "bash_script":
		cmdArgs := append([]string{cfg.Path}, args...)
		cmd = osexec.Command("bash", cmdArgs...)
	case "shell_script":
		cmdArgs := append([]string{cfg.Path}, args...)
		cmd = osexec.Command("/bin/sh", cmdArgs...)
	case "python_script":
		cmdArgs := append([]string{cfg.Path}, args...)
		cmd = osexec.Command("python3", cmdArgs...)
	case "nodejs_script":
		cmdArgs := append([]string{cfg.Path}, args...)
		cmd = osexec.Command("node", cmdArgs...)
	case "perl_script":
		cmdArgs := append([]string{cfg.Path}, args...)
		cmd = osexec.Command("perl", cmdArgs...)
	case "ruby_script":
		cmdArgs := append([]string{cfg.Path}, args...)
		cmd = osexec.Command("ruby", cmdArgs...)
	default:
		if len(args) > 0 {
			cmd = osexec.Command(cfg.Path, args...)
		} else {
			cmd = osexec.Command(cfg.Path)
		}
	}

	cmd.Dir = filepath.Dir(cfg.Path)
	return cmd
}

// startProcess 启动并监控子进程
func startProcess(cfg config.Config) error {
	// 重试计数由调用方控制，这里只负责一次启动
	hooks := pmexec.Hooks{
		RunPre:    func() error { return runPreCommand(cfg) },
		CreateCmd: func() (*osexec.Cmd, error) { return createCommand(cfg), nil },
		ApplyEnv: func(cmd *osexec.Cmd) error {
			if env, pathVal := buildChildEnv(cfg.Env); len(env) > 0 {
				cmd.Env = env
				log.Printf("子进程环境PATH: %s", pathVal)
			}
			return nil
		},
		OnStarted: func(cmd *osexec.Cmd) {
			log.Printf("进程启动成功，PID: %d, 路径: %s, 类型: %s", cmd.Process.Pid, cfg.Path, procState.FileType)
			if cfg.Args != "" {
				log.Printf("程序参数: %s", cfg.Args)
			}
		},
		OnStdout: func(r io.ReadCloser) { readProcessOutput(r, "STDOUT", cfg) },
		OnStderr: func(r io.ReadCloser) { readProcessOutput(r, "STDERR", cfg) },
		OnExit: func(exit pmexec.ProcessExit) {
			switch {
			case exit.StoppedByRequest:
				log.Printf("进程被请求停止，code=%d", exit.LastExitCode)
			case exit.LastExitError != "":
				log.Printf("进程异常退出，code=%d, signal=%v, 错误: %v", exit.LastExitCode, exit.LastExitBySignal, exit.LastExitError)
			default:
				log.Printf("进程正常退出，code=%d", exit.LastExitCode)
			}
		},
		Logf:         log.Printf,
		AutoRestart:  cfg.AutoRestart,
		RestartTimes: cfg.RestartTimes,
		RestartDelay: 2 * time.Second,
	}

	return procState.StartManagedProcess(hooks)
}

// stopProcess 优雅停止子进程
func stopProcess() error {
	return procState.StopManagedProcess(5 * time.Second)
}

// monitorProcess 不再在 core 中实现，交由 exec.Process monitor；保留 readProcessOutput 用于 Hooks。

// readProcessOutput 读取并缓存子进程输出
func readProcessOutput(reader io.ReadCloser, prefix string, cfg config.Config) {
	defer reader.Close()
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, cfg.LogMaxLineBytes)
	scanner.Buffer(buf, cfg.LogMaxLineBytes)

	for scanner.Scan() {
		logMsg := fmt.Sprintf("[%s] %s", prefix, scanner.Text())
		logMu.Lock()
		logCache = append(logCache, logMsg)
		if len(logCache) > cfg.LogCapacity {
			logCache = logCache[len(logCache)-cfg.LogCapacity:]
		}
		logMu.Unlock()
		if fileLogger != nil {
			fileLogger.WriteLine(logMsg)
		}
		log.Println(logMsg)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("读取进程输出时出错: %v", err)
	}
}

// checkPassword 校验请求中携带的 password，如果未配置密码则直接通过
func checkPassword(r *http.Request, expected string) error {
	if expected == "" {
		return nil
	}
	pw := r.URL.Query().Get("password")
	if pw == expected {
		return nil
	}
	return errors.New("unauthorized")
}
