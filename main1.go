package main

//
//import (
//	"bufio"
//	"context"
//	"encoding/json"
//	"flag"
//	"fmt"
//	"io"
//	"log"
//	"net/http"
//	"os"
//	"os/exec"
//	"os/signal"
//	"path/filepath"
//	"strings"
//	"sync"
//	"syscall"
//	"time"
//)
//
//// 心跳信息结构体
//type HeartbeatInfo struct {
//	Version          string   `json:"version"`       // 当前版本
//	State            string   `json:"state"`         // 当前进程状态
//	Info             string   `json:"info"`          // 传递信息
//	Timestamp        string   `json:"timestamp"`     // 时间戳
//	Logs             []string `json:"logs"`          // 当前的一批日志
//	ProcessPID       int      `json:"process_pid"`   // 进程PID
//	ProcessPath      string   `json:"process_path"`  // 进程路径
//	RestartCount     int      `json:"restart_count"` // 重启次数
//	FileType         string   `json:"file_type"`     // 文件类型
//	LastExitCode     int      `json:"last_exit_code,omitempty"`
//	LastExitTime     string   `json:"last_exit_time,omitempty"`
//	LastExitBySignal bool     `json:"last_exit_by_signal,omitempty"`
//	LastExitError    string   `json:"last_exit_error,omitempty"`
//}
//
//// 全局状态变量
//var (
//	version = "1.0.6" // 进程管理器版本
//
//	// 命令行参数
//	programPath     = flag.String("path", "", "要管理的程序路径（支持可执行文件、shell脚本等）")
//	programArgs     = flag.String("args", "", "传递给程序的参数（可选）")
//	preCmd          = flag.String("pre", "", "启动前要执行的命令（在sh中执行，可用&&/;/||连接多条命令）")
//	autoRestart     = flag.Bool("auto-restart", false, "是否自动重启")
//	maxRetries      = flag.Int("max-retries", 3, "最大重试次数")
//	startNow        = flag.Bool("start-now", false, "是否立即启动")
//	httpPort        = flag.String("port", "11883", "HTTP服务端口")
//	extraEnvs       = flag.String("env", "", "为子进程追加环境变量，如: \"PATH=/usr/local/bin:/usr/bin FOO=bar\"；用空格或分号分隔")
//	logCapacity     = flag.Int("log-capacity", 200, "日志缓存的最大行数（默认200）")
//	logMaxLineBytes = flag.Int("log-max-line-bytes", 1048576, "单行日志的最大字节数（用于bufio.Scanner，默认1MiB）")
//
//	// 进程管理相关
//	currentProcess *exec.Cmd
//	processMu      sync.Mutex
//	isRunning      = false
//	restartCount   = 0
//	fileType       = ""
//
//	// 日志相关
//	logsCache      []string
//	logsMu         sync.Mutex
//	shutdownSignal = make(chan struct{})
//
//	// 进程退出信息
//	lastExitCode     int
//	lastExitTime     time.Time
//	lastExitBySignal bool
//	lastExitError    string
//	stoppedByRequest bool
//)
//
//// 检测文件类型
//func detectFileType(filePath string) (string, error) {
//	// 读取文件前几个字节来判断类型
//	file, err := os.Open(filePath)
//	if err != nil {
//		return "", err
//	}
//	defer file.Close()
//
//	// 读取前512字节用于检测
//	buffer := make([]byte, 512)
//	n, err := file.Read(buffer)
//	if err != nil && err != io.EOF {
//		return "", err
//	}
//
//	content := string(buffer[:n])
//
//	// 检查是否是脚本文件
//	if strings.HasPrefix(content, "#!") {
//		lines := strings.Split(content, "\n")
//		shebang := lines[0]
//
//		if strings.Contains(shebang, "bash") {
//			return "bash_script", nil
//		} else if strings.Contains(shebang, "sh") {
//			return "shell_script", nil
//		} else if strings.Contains(shebang, "python") {
//			return "python_script", nil
//		} else if strings.Contains(shebang, "node") {
//			return "nodejs_script", nil
//		} else {
//			return "script", nil
//		}
//	}
//
//	// 检查文件扩展名
//	ext := strings.ToLower(filepath.Ext(filePath))
//	switch ext {
//	case ".sh":
//		return "shell_script", nil
//	case ".py":
//		return "python_script", nil
//	case ".js":
//		return "nodejs_script", nil
//	case ".pl":
//		return "perl_script", nil
//	case ".rb":
//		return "ruby_script", nil
//	default:
//		// 检查是否是二进制可执行文件
//		if isBinaryExecutable(buffer[:n]) {
//			return "binary_executable", nil
//		}
//		return "unknown", nil
//	}
//}
//
//// 检查是否是二进制可执行文件
//func isBinaryExecutable(data []byte) bool {
//	// Linux ELF文件魔数
//	if len(data) >= 4 && data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
//		return true
//	}
//	// 其他二进制文件特征（简单检查）
//	for _, b := range data[:min(len(data), 100)] {
//		if b == 0 {
//			return true // 包含空字节，可能是二进制文件
//		}
//	}
//	return false
//}
//
//func min(a, b int) int {
//	if a < b {
//		return a
//	}
//	return b
//}
//
//// 验证程序路径和参数
//func validateParameters() error {
//	if *programPath == "" {
//		return fmt.Errorf("程序路径不能为空，请使用 -path 指定")
//	}
//
//	// 检查文件是否存在
//	if _, err := os.Stat(*programPath); os.IsNotExist(err) {
//		return fmt.Errorf("程序文件不存在: %s", *programPath)
//	}
//
//	// 转换为绝对路径
//	absPath, err := filepath.Abs(*programPath)
//	if err != nil {
//		return fmt.Errorf("无法获取绝对路径: %v", err)
//	}
//	*programPath = absPath
//
//	// 检测文件类型
//	detectedType, err := detectFileType(*programPath)
//	if err != nil {
//		return fmt.Errorf("无法检测文件类型: %v", err)
//	}
//	fileType = detectedType
//
//	// 检查文件信息
//	fileInfo, err := os.Stat(*programPath)
//	if err != nil {
//		return fmt.Errorf("无法获取文件信息: %v", err)
//	}
//
//	// 对于脚本文件，检查是否可读；对于二进制文件，检查执行权限
//	if strings.Contains(fileType, "script") {
//		// 脚本文件需要可读权限
//		if fileInfo.Mode()&0444 == 0 {
//			return fmt.Errorf("脚本文件没有读权限: %s", *programPath)
//		}
//		// 如果脚本没有执行权限，给出警告但不阻止执行
//		if fileInfo.Mode()&0111 == 0 {
//			log.Printf("警告: 脚本文件没有执行权限，将尝试通过解释器执行: %s", *programPath)
//		}
//	} else {
//		// 二进制文件需要执行权限
//		if fileInfo.Mode()&0111 == 0 {
//			return fmt.Errorf("可执行文件没有执行权限: %s", *programPath)
//		}
//	}
//
//	// 验证重试次数
//	if *maxRetries < 0 {
//		return fmt.Errorf("最大重试次数不能为负数")
//	}
//
//	// 验证日志容量参数
//	if *logCapacity <= 0 {
//		return fmt.Errorf("log-capacity 必须为正数")
//	}
//	if *logMaxLineBytes < 64*1024 { // 至少64KiB，避免极小导致频繁截断
//		log.Printf("警告: log-max-line-bytes 太小，已提升到 64KiB")
//		*logMaxLineBytes = 64 * 1024
//	}
//
//	log.Printf("检测到文件类型: %s", fileType)
//	return nil
//}
//
//// 创建适当的执行命令
//func createCommand() *exec.Cmd {
//	var cmd *exec.Cmd
//	args := []string{}
//
//	// 解析程序参数
//	if *programArgs != "" {
//		args = strings.Fields(*programArgs)
//	}
//
//	switch fileType {
//	case "bash_script":
//		cmdArgs := append([]string{*programPath}, args...)
//		cmd = exec.Command("bash", cmdArgs...)
//	case "shell_script":
//		cmdArgs := append([]string{*programPath}, args...)
//		cmd = exec.Command("/bin/sh", cmdArgs...)
//	case "python_script":
//		cmdArgs := append([]string{*programPath}, args...)
//		cmd = exec.Command("python3", cmdArgs...)
//	case "nodejs_script":
//		cmdArgs := append([]string{*programPath}, args...)
//		cmd = exec.Command("node", cmdArgs...)
//	case "perl_script":
//		cmdArgs := append([]string{*programPath}, args...)
//		cmd = exec.Command("perl", cmdArgs...)
//	case "ruby_script":
//		cmdArgs := append([]string{*programPath}, args...)
//		cmd = exec.Command("ruby", cmdArgs...)
//	default:
//		// 二进制可执行文件或其他类型，直接执行
//		if len(args) > 0 {
//			cmd = exec.Command(*programPath, args...)
//		} else {
//			cmd = exec.Command(*programPath)
//		}
//	}
//
//	// 设置工作目录为程序所在目录
//	cmd.Dir = filepath.Dir(*programPath)
//
//	return cmd
//}
//
//// 日志拦截器
//type LogInterceptor struct{}
//
//func (li *LogInterceptor) Write(p []byte) (n int, err error) {
//	logsMu.Lock()
//	defer logsMu.Unlock()
//
//	logsCache = append(logsCache, string(p))
//
//	// 限制日志缓存大小
//	if len(logsCache) > *logCapacity {
//		logsCache = logsCache[len(logsCache)-*logCapacity:]
//	}
//
//	// 同时输出到标准输出
//	fmt.Print(string(p))
//	return len(p), nil
//}
//
//// 解析 -env 参数，支持空格或分号分隔的 KEY=VAL 列表
//func parseExtraEnv(s string) []string {
//	if s == "" {
//		return nil
//	}
//	// 将分号替换为空格，再按空格切分
//	s = strings.ReplaceAll(s, ";", " ")
//	fields := strings.Fields(s)
//	var out []string
//	for _, f := range fields {
//		if !strings.Contains(f, "=") {
//			continue
//		}
//		out = append(out, f)
//	}
//	return out
//}
//
//// 组合子进程环境变量，并返回最终 PATH 便于日志打印
//func buildChildEnv() ([]string, string) {
//	base := os.Environ()
//	add := parseExtraEnv(*extraEnvs)
//	env := append([]string{}, base...)
//	env = append(env, add...)
//	// 取最终 PATH
//	pathVal := ""
//	for i := len(env) - 1; i >= 0; i-- {
//		if strings.HasPrefix(env[i], "PATH=") {
//			pathVal = strings.TrimPrefix(env[i], "PATH=")
//			break
//		}
//	}
//	if pathVal == "" {
//		pathVal = os.Getenv("PATH")
//	}
//	return env, pathVal
//}
//
//// 启动前执行的预命令（在 /bin/sh 中执行）
//func runPreCommand() error {
//	if preCmd == nil || *preCmd == "" {
//		return nil
//	}
//	log.Printf("执行启动前命令: %s", *preCmd)
//	cmd := exec.Command("/bin/sh", "-c", *preCmd)
//	cmd.Dir = filepath.Dir(*programPath)
//	// 应用环境变量
//	if env, pathVal := buildChildEnv(); len(env) > 0 {
//		cmd.Env = env
//		log.Printf("PRE环境PATH: %s", pathVal)
//	}
//	out, err := cmd.CombinedOutput()
//
//	// 把输出追加到日志缓存
//	text := strings.ReplaceAll(string(out), "\r\n", "\n")
//	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
//	logsMu.Lock()
//	for _, line := range lines {
//		if line == "" {
//			continue
//		}
//		logsCache = append(logsCache, "[PRE] "+line)
//		if len(logsCache) > *logCapacity {
//			logsCache = logsCache[len(logsCache)-*logCapacity:]
//		}
//	}
//	logsMu.Unlock()
//
//	if err != nil {
//		return fmt.Errorf("pre命令执行失败: %v", err)
//	}
//	return nil
//}
//
//// 启动进程
//func startProcess() error {
//	processMu.Lock()
//	defer processMu.Unlock()
//
//	if isRunning && currentProcess != nil {
//		return fmt.Errorf("进程已在运行，PID: %d", currentProcess.Process.Pid)
//	}
//
//	// 启动前执行预命令
//	if err := runPreCommand(); err != nil {
//		return err
//	}
//
//	// 创建新进程
//	cmd := createCommand()
//	// 应用环境变量
//	if env, pathVal := buildChildEnv(); len(env) > 0 {
//		cmd.Env = env
//		log.Printf("子进程环境PATH: %s", pathVal)
//	}
//
//	// 设置输出管道
//	stdout, err := cmd.StdoutPipe()
//	if err != nil {
//		return fmt.Errorf("创建stdout管道失败: %v", err)
//	}
//
//	stderr, err := cmd.StderrPipe()
//	if err != nil {
//		return fmt.Errorf("创建stderr管道失败: %v", err)
//	}
//
//	// 启动进程
//	if err := cmd.Start(); err != nil {
//		return fmt.Errorf("启动进程失败: %v", err)
//	}
//
//	currentProcess = cmd
//	isRunning = true
//
//	log.Printf("进程启动成功，PID: %d, 路径: %s, 类型: %s", cmd.Process.Pid, *programPath, fileType)
//	if *programArgs != "" {
//		log.Printf("程序参数: %s", *programArgs)
//	}
//
//	// 启动日志读取协程
//	go readProcessOutput(stdout, "STDOUT")
//	go readProcessOutput(stderr, "STDERR")
//
//	// 启动进程监控协程
//	go monitorProcess()
//
//	return nil
//}
//
//// 读取进程输出
//func readProcessOutput(reader io.ReadCloser, prefix string) {
//	defer reader.Close()
//	scanner := bufio.NewScanner(reader)
//	// 提高单行日志上限，默认为 1MiB，可通过 -log-max-line-bytes 调整
//	buf := make([]byte, 0, *logMaxLineBytes)
//	scanner.Buffer(buf, *logMaxLineBytes)
//
//	for scanner.Scan() {
//		logMsg := fmt.Sprintf("[%s] %s", prefix, scanner.Text())
//
//		logsMu.Lock()
//		logsCache = append(logsCache, logMsg)
//		if len(logsCache) > *logCapacity {
//			logsCache = logsCache[len(logsCache)-*logCapacity:]
//		}
//		logsMu.Unlock()
//
//		// 输出到标准输出
//		log.Println(logMsg)
//	}
//
//	if err := scanner.Err(); err != nil {
//		log.Printf("读取进程输出时出错: %v", err)
//	}
//}
//
//// 监控进程
//func monitorProcess() {
//	var proc *exec.Cmd
//	processMu.Lock()
//	proc = currentProcess
//	processMu.Unlock()
//
//	if proc == nil {
//		return
//	}
//
//	// 等待进程结束
//	err := proc.Wait()
//
//	processMu.Lock()
//	isRunning = false
//	pid := 0
//	if proc.Process != nil {
//		pid = proc.Process.Pid
//	}
//	// 记录退出信息
//	lastExitTime = time.Now()
//	lastExitCode = 0
//	lastExitBySignal = false
//	lastExitError = ""
//	if err != nil {
//		lastExitError = err.Error()
//		if ee, ok := err.(*exec.ExitError); ok {
//			ps := ee.ProcessState
//			lastExitCode = ps.ExitCode()
//			if ws, ok := ps.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
//				lastExitBySignal = true
//			}
//		}
//	}
//	wasStoppedByReq := stoppedByRequest
//	// 清除一次性标记
//	stoppedByRequest = false
//	processMu.Unlock()
//
//	if wasStoppedByReq {
//		log.Printf("进程被请求停止，PID: %d", pid)
//	} else if err != nil {
//		log.Printf("进程异常退出，PID: %d, code=%d, signal=%v, 错误: %v", pid, lastExitCode, lastExitBySignal, err)
//	} else {
//		log.Printf("进程正常退出，PID: %d, code=0", pid)
//	}
//
//	// 如果启用自动重启且还有重试次数
//	if *autoRestart && restartCount < *maxRetries {
//		restartCount++
//		log.Printf("准备重启进程，当前重试次数: %d/%d", restartCount, *maxRetries)
//		time.Sleep(2 * time.Second)
//		if err := startProcess(); err != nil {
//			log.Printf("自动重启失败: %v", err)
//		}
//	} else if *autoRestart {
//		log.Printf("已达到最大重试次数 %d，停止自动重启", *maxRetries)
//	}
//}
//
//// 停止进程
//func stopProcess() error {
//	processMu.Lock()
//	defer processMu.Unlock()
//
//	if !isRunning || currentProcess == nil {
//		return fmt.Errorf("进程未运行")
//	}
//
//	// 标记为用户请求停止，便于 monitorProcess 不将其当作崩溃
//	stoppedByRequest = true
//
//	// 先尝试优雅关闭
//	if err := currentProcess.Process.Signal(syscall.SIGTERM); err != nil {
//		log.Printf("发送SIGTERM信号失败: %v", err)
//	}
//
//	// 等待进程退出
//	done := make(chan error, 1)
//	go func(cmd *exec.Cmd) {
//		done <- cmd.Wait()
//	}(currentProcess)
//
//	select {
//	case <-time.After(5 * time.Second):
//		log.Println("进程未在5秒内退出，强制终止")
//		if err := currentProcess.Process.Kill(); err != nil {
//			return fmt.Errorf("强制终止进程失败: %v", err)
//		}
//		<-done
//	case err := <-done:
//		if err != nil {
//			log.Printf("进程退出时出现错误: %v", err)
//		}
//	}
//
//	isRunning = false
//	currentProcess = nil
//	log.Println("进程已停止")
//	return nil
//}
//
//// HTTP处理函数：启动进程
//func startupHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodGet && r.Method != http.MethodPost {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	// 重置重启计数
//	restartCount = 0
//
//	err := startProcess()
//	if err != nil {
//		http.Error(w, fmt.Sprintf("启动进程失败: %v", err), http.StatusInternalServerError)
//		return
//	}
//
//	fmt.Fprintln(w, "ok")
//}
//
//// HTTP处理函数：心跳检查
//func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodGet {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	logsMu.Lock()
//	logsCopy := logsCache
//	logsCache = nil
//	logsMu.Unlock()
//
//	processMu.Lock()
//	state := "stopped"
//	pid := 0
//	if isRunning && currentProcess != nil {
//		state = "running"
//		pid = currentProcess.Process.Pid
//	}
//	lexitCode := lastExitCode
//	lexitSignal := lastExitBySignal
//	lexitErr := lastExitError
//	lexitTime := lastExitTime
//	processMu.Unlock()
//
//	currentTime := time.Now()
//	info := HeartbeatInfo{
//		Version:          version,
//		State:            state,
//		Info:             "进程管理器正常运行",
//		Timestamp:        currentTime.Format("2006-01-02 15:04:05"),
//		Logs:             logsCopy,
//		ProcessPID:       pid,
//		ProcessPath:      *programPath,
//		RestartCount:     restartCount,
//		FileType:         fileType,
//		LastExitCode:     lexitCode,
//		LastExitBySignal: lexitSignal,
//		LastExitError:    lexitErr,
//	}
//	if !lexitTime.IsZero() {
//		info.LastExitTime = lexitTime.Format("2006-01-02 15:04:05")
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	jsonData, err := json.Marshal(info)
//	if err != nil {
//		http.Error(w, fmt.Sprintf("JSON序列化错误: %v", err), http.StatusInternalServerError)
//		return
//	}
//	w.Write(jsonData)
//}
//
//// HTTP处理函数：关闭进程
//func shutdownHandler(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodGet && r.Method != http.MethodPost {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	err := stopProcess()
//	if err != nil {
//		http.Error(w, fmt.Sprintf("停止进程失败: %v", err), http.StatusInternalServerError)
//		return
//	}
//
//	fmt.Fprintln(w, "ok")
//}
//
//// 主函数
//func main1() {
//	// 解析命令行参数
//	flag.Parse()
//
//	// 设置全局异常恢复
//	defer func() {
//		if r := recover(); r != nil {
//			log.Printf("程序发生panic: %v", r)
//		}
//	}()
//
//	// 验证参数
//	if err := validateParameters(); err != nil {
//		log.Fatalf("参数验证失败: %v", err)
//	}
//
//	log.Printf("Linux进程管理器启动")
//	log.Printf("管理目标: %s", *programPath)
//	log.Printf("文件类型: %s", fileType)
//	if *programArgs != "" {
//		log.Printf("程序参数: %s", *programArgs)
//	}
//	log.Printf("自动重启: %v", *autoRestart)
//	log.Printf("最大重试次数: %d", *maxRetries)
//	log.Printf("立即启动: %v", *startNow)
//	log.Printf("HTTP端口: %s", *httpPort)
//
//	// 如果指定立即启动
//	if *startNow {
//		log.Println("立即启动目标进程...")
//		if err := startProcess(); err != nil {
//			log.Printf("立即启动失败: %v", err)
//		}
//	}
//
//	// 设置信号处理
//	sigChan := make(chan os.Signal, 1)
//	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
//
//	// 创建HTTP服务器
//	mux := http.NewServeMux()
//	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
//		fmt.Fprintln(w, "DontCrash 1.0.0 By tyza66")
//	})
//	mux.HandleFunc("/startup", startupHandler)
//	mux.HandleFunc("/heartbeat", heartbeatHandler)
//	mux.HandleFunc("/shutdown", shutdownHandler)
//
//	server := &http.Server{
//		Addr:         ":" + *httpPort,
//		Handler:      mux,
//		ReadTimeout:  10 * time.Second,
//		WriteTimeout: 10 * time.Second,
//		IdleTimeout:  30 * time.Second,
//	}
//
//	// 启动HTTP服务器
//	go func() {
//		log.Printf("HTTP服务器启动在端口 %s", *httpPort)
//		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//			log.Printf("HTTP服务器启动失败: %v", err)
//		}
//	}()
//
//	log.Println("进程管理器运行中...")
//
//	// 等待终止信号
//	sig := <-sigChan
//	log.Printf("收到信号 %v，开始关闭...", sig)
//
//	// 优雅关闭
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	// 停止管理的进程
//	if isRunning {
//		log.Println("正在停止管理的进程...")
//		if err := stopProcess(); err != nil {
//			log.Printf("停止进程时出错: %v", err)
//		}
//	}
//
//	// 关闭HTTP服务器
//	if err := server.Shutdown(ctx); err != nil {
//		log.Printf("HTTP服务器关闭出错: %v", err)
//	}
//
//	log.Println("进程管理器已优雅关闭")
//}
