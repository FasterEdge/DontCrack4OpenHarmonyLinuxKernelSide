package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// runMode 定义子进程的运行模式
// normal: 正常输出直到 lifetime 到期或收到信号退出
// crash: 模拟崩溃退出（退出码非0）
// hang: 挂起不退出，用于测试管理器超时和强杀
// graceful: 等待信号后优雅退出
func main() {
	mode := flag.String("mode", "normal", "运行模式: normal|crash|hang|graceful")
	interval := flag.Duration("interval", time.Second, "输出间隔")
	lifetime := flag.Duration("lifetime", 0, "运行时长，0 表示不主动退出")
	stateFile := flag.String("state-file", "", "用于记录重启计数的文件路径，可选")
	message := flag.String("message", "", "自定义输出前缀")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// 读取并更新环境中的重启计数
	restartEnv := os.Getenv("RESTART_ENV_COUNT")
	restartCount := parseIntDefault(restartEnv, 0)
	if restartEnv != "" {
		restartCount++
		// 不修改父进程环境，只打印观察
		log.Printf("env restart count -> %d", restartCount)
	}

	// 状态文件计数，用于跨重启观察
	if *stateFile != "" {
		cnt := readCounter(*stateFile)
		cnt++
		_ = os.WriteFile(*stateFile, []byte(fmt.Sprintf("%d", cnt)), 0644)
		log.Printf("state-file %s count -> %d", *stateFile, cnt)
	}

	// 打印启动信息
	log.Printf("childproc start | pid=%d | mode=%s | interval=%s | lifetime=%s | msg=%s", os.Getpid(), *mode, interval.String(), lifetime.String(), *message)
	log.Printf("args: %s", strings.Join(os.Args, " "))
	log.Printf("env EXTRA_INFO=%s", os.Getenv("EXTRA_INFO"))
	log.Printf("env RESTART_ENV_COUNT=%s", restartEnv)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	start := time.Now()
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	exitCode := 0
	crashCode := 42

	for {
		select {
		case <-ctx.Done():
			log.Println("received signal, exiting gracefully")
			return
		case t := <-ticker.C:
			// 周期性输出 stdout
			log.Printf("tick at %s", t.Format(time.RFC3339Nano))
			// 额外模拟 stderr 输出
			fmt.Fprintf(os.Stderr, "stderr burst at %s\n", t.Format(time.RFC3339Nano))
		default:
			// 检查运行时长
			if *lifetime > 0 && time.Since(start) > *lifetime {
				switch *mode {
				case "crash":
					log.Println("lifetime reached, simulating crash")
					os.Exit(crashCode)
				case "graceful", "normal":
					log.Println("lifetime reached, exiting normally")
					return
				}
			}
			if *mode == "hang" {
				// 持续消耗 CPU 轻度循环，偶尔打印
				time.Sleep(200 * time.Millisecond)
				if rand.Intn(20) == 0 {
					log.Printf("hang mode heartbeat pid=%d", os.Getpid())
				}
				continue
			}
			// 让出 CPU
			time.Sleep(50 * time.Millisecond)
		}
	}

	os.Exit(exitCode)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func readCounter(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	v, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return v
}
