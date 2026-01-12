package config

import (
	"errors"
	"fmt"
	"os"
)

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

	return nil
}

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
