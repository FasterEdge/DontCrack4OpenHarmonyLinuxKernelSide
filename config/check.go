package config

import (
	"errors"
	"fmt"
	"os"
)

// 检查配置的合法性
func CheckConfig(config Config) error {
	// 根据定义的配置结构体进行检查
	if config.Version == "" { // 系统版本号不能为空
		// 这是一个检查开关，若将代码中的版本号清空则直接跳过检查，返回 nil
		fmt.Println("系统版本号为空，跳过配置检查")
		return nil
	}
	if config.Path == "" { // 子进程路径不能为空
		return fmt.Errorf("所管理进程的Path不可以为空")
	} else {
		// 检查指定地方的文件存在与否（并且是文件而不是文件夹）
		exists := fileExists(config.Path)
		if !exists {
			return fmt.Errorf("所管理进程的Path指定的文件不存在: %s", config.Path)
		}
	}

	// 数值边界校验：非法值会导致运行时 panic 或服务失效
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port 必须在 1..65535 之间, 当前值: %d", config.Port)
	}
	if config.LogCapacity < 0 {
		return fmt.Errorf("log-capacity 不能为负数, 当前值: %d", config.LogCapacity)
	}
	if config.LogMaxLineBytes <= 0 {
		return fmt.Errorf("log-max-line-bytes 必须大于 0, 当前值: %d", config.LogMaxLineBytes)
	}
	if config.RestartTimes < -1 {
		return fmt.Errorf("max-retries 只能为 -1(无限) 或非负数, 当前值: %d", config.RestartTimes)
	}
	if config.LocalLogLifeDay < 0 {
		return fmt.Errorf("log-life-day 不能为负数, 当前值: %d", config.LocalLogLifeDay)
	}
	if config.ProbeCmd != "" {
		if config.ProbeInterval <= 0 {
			return fmt.Errorf("probe-interval 必须大于 0, 当前值: %d", config.ProbeInterval)
		}
		if config.ProbeTimeout <= 0 {
			return fmt.Errorf("probe-timeout 必须大于 0, 当前值: %d", config.ProbeTimeout)
		}
		if config.ProbeFailureLimit <= 0 {
			return fmt.Errorf("probe-failure-limit 必须大于 0, 当前值: %d", config.ProbeFailureLimit)
		}
	}

	return nil
}

// 检查指定位置的文件存在并且是文件而不是文件夹
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // 如要忽略符号链接用 os.Lstat
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false // 文件不存在
		}
		// 其他错误（权限不足等）也视为“不存在”
		return false
	}
	return !info.IsDir()
}
