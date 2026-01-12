package main

import (
	"DontCrash/config"
	"DontCrash/core"
	"flag"
)

// DontCrack开源鸿蒙(OpenHarmony)Linux内核专用版
// https://github.com/FasterEdge/DontCrack4OpenHarmonyLinuxKernelSide
// By tyza66
// https://github.com/tyza66
// 程序入口
func main() {
	// 系统版本号
	version := "1.1.20260112" // 当前系统的版本

	// 全局配置信息获得
	// 管理子进程相关
	path := flag.String("path", "", "要管理的程序路径（可选，支持可执行文件和shell脚本）")
	args := flag.String("args", "", "传递给程序的参数（可选）")
	pre := flag.String("pre", "", "启动前要执行的命令（可选，在sh中执行，可用&&/;/||连接多条命令）")
	env := flag.String("env", "", "为子进程追加环境变量，如: \"PATH=/usr/local/bin:/usr/bin FOO=bar\"；用空格或分号分隔")
	autoRestart := flag.Bool("auto-restart", false, "是否自动重启（可选)")
	maxRetries := flag.Int("max-retries", 3, "最大重试次数（可选，-1表示无限次)")
	startNow := flag.Bool("start-now", false, "是否立即启动")
	// 管理器相关
	port := flag.Int("port", 11883, "HTTP服务端口(可选)")
	password := flag.String("password", "", "管理进程的密码（可选）")
	logCapacity := flag.Int("log-capacity", 200, "日志缓存的最大行数（可选，默认200）")
	logMaxLineBytes := flag.Int("log-max-line-bytes", 1048576, "单行日志的最大字节数（可选，用于bufio.Scanner，默认1MiB）")
	fileLogEnabled := flag.Bool("file-log", false, "是否启用文件日志（可选，默认关闭）")
	localLogPath := flag.String("log-path", "/data/logs/proc_manager/", "本地日志文件目录（可选，默认/data/logs/proc_manager/进程名）")
	localLogLifeDay := flag.Int("log-life-day", 7, "本地日志文件保存天数（可选，默认7天）")

	// 解析传入的参数
	flag.Parse()

	// 将传入的配置信息转换为全局配置结构体
	config := config.ParseConfig(version, *path, *args, *pre, *env, *autoRestart,
		*maxRetries, *startNow, *port, *password, *logCapacity, *logMaxLineBytes,
		*fileLogEnabled, *localLogPath, *localLogLifeDay)

	// 携带启动参数启动管理器
	core.Start(*config)
}
