package exec

import (
	"os/exec"
	"sync"
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
}

// 进程退出信息
type ProcessExit struct {
	LastExitCode     int       // 进程最后退出代码
	LastExitTime     time.Time // 进程最后退出时间
	LastExitBySignal bool      // 进程是否被信号终止
	LastExitError    string    // 进程最后退出的错误信息
	StoppedByRequest bool      // 进程是否是被请求停止的
}
