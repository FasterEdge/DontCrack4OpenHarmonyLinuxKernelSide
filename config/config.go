package config

type Config struct {
	// 系统相关
	Version string // 系统版本号(主方法中写死定义)
	// 子进程相关
	Path         string // 打开子进程的可执行文件路径(必填)
	Args         string // 打开子进程所携带的参数(可选)
	Pre          string // 启动前要执行的命令(可选，只能一条但可以是sh脚本或&&连接的命令)
	Env          string // 子进程的环境变量(可选，如: "PATH=/usr/local/bin:/usr/bin FOO=bar")
	AutoRestart  bool   // 子进程退出后是否自动重启(可选，默认关闭)
	RestartTimes int    // 最大重试次数(可选，默认3次，-1表示无限次)
	StartNow     bool   // 是否立即启动子进程(可选，默认关闭)
	// 管理器相关
	Port            int    // 管理进程使用的进程端口(可选，默认11883)
	Password        string // 管理进程使用的密码(可选)
	LogCapacity     int    // 日志容量(可选，默认10MB)
	LogMaxLineBytes int    // 日志单行最大字节数(可选，默认1KB)
	FileLogEnabled  bool   // 是否启用文件日志(可选，默认关闭)
	LocalLogPath    string // 日志文件目录(可选，默认/data/logs/proc_manager/进程名)
	LocalLogLifeDay int    // 本地日志文件保存天数(可选，默认7天，新日志进来的时候检查日志文件名)
}

func ParseConfig(version string, path string, args string, pre string, env string, autoRestart bool, restartTimes int, startNow bool,
	port int, password string, logCapacity int, logMaxLineBytes int, fileLogEnabled bool, localLogPath string, localLogLifeDay int) *Config {
	// 初始化空的配置结构体
	config := new(Config)
	// 通过传入的参数进行赋值
	// 系统相关
	config.Version = version // 系统版本号
	// 子进程相关
	config.Path = path
	config.Args = args
	config.Pre = pre
	config.Env = env
	config.AutoRestart = autoRestart
	config.RestartTimes = restartTimes
	config.StartNow = startNow
	// 管理器相关
	config.Port = port
	config.Password = password
	config.LogCapacity = logCapacity
	config.LogMaxLineBytes = logMaxLineBytes
	config.FileLogEnabled = fileLogEnabled
	config.LocalLogPath = localLogPath
	config.LocalLogLifeDay = localLogLifeDay

	return config
}
