package config

// FasterEdge - 对称、可靠、安全的多场景边缘计算框架
// https://github.com/FasterEdge
// DontCrack开源鸿蒙(OpenHarmony)Linux内核专用版
// https://github.com/FasterEdge/DontCrack4OpenHarmonyLinuxKernelSide
// tyza66
// https://github.com/tyza66
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

// 将传入的参数转换为全局配置结构体
func ParseConfig(version string, path string, args string, pre string, env string, autoRestart bool, restartTimes int, startNow bool,
	port int, password string, logCapacity int, logMaxLineBytes int, fileLogEnabled bool, localLogPath string, localLogLifeDay int) *Config {
	// 初始化空的配置结构体
	config := new(Config)
	// 通过传入的参数进行赋值
	// 系统相关
	config.Version = version // 系统版本号
	// 子进程相关
	config.Path = path                 // 子进程路径
	config.Args = args                 // 子进程参数
	config.Pre = pre                   // 启动前命令
	config.Env = env                   // 子进程环境变量
	config.AutoRestart = autoRestart   // 是否自动重启
	config.RestartTimes = restartTimes // 最大重试次数
	config.StartNow = startNow         // 是否立即启动
	// 管理器相关
	config.Port = port                       // 管理端口
	config.Password = password               // 管理密码
	config.LogCapacity = logCapacity         // 日志容量
	config.LogMaxLineBytes = logMaxLineBytes // 日志单行最大字节数
	config.FileLogEnabled = fileLogEnabled   // 是否启用文件日志
	config.LocalLogPath = localLogPath       // 本地日志文件目录
	config.LocalLogLifeDay = localLogLifeDay // 本地日志文件保存天数

	return config
}
