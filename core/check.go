package core

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 通过魔数检查是不是二进制可执行文件
func isBinaryExecutable(data []byte) bool {
	// Linux ELF文件魔数
	if len(data) >= 4 && data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
		return true
	}

	// 其他二进制文件特征（简单检查）
	for _, b := range data[:min(len(data), 100)] {
		if b == 0 {
			return true // 包含空字节，可能是二进制文件
		}
	}
	return false
}

// 基于文件头/扩展名识别类型，用于决定如何启动
func detectFileType(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	// 读取的内容转换为字符串
	content := string(buffer[:n])

	// 首先检查是否是脚本文件
	if strings.HasPrefix(content, "#!") { // 只要第一行以 "#!" 开头，就认为是脚本文件
		lines := strings.Split(content, "\n")
		shebang := lines[0] // 获得第一行（带脚本格式）
		switch {
		case strings.Contains(shebang, "sh"): // 如果脚本格式中包含 "sh"，则认为是 shell 脚本
			return "shell_script", nil
		default:
			return "script", nil // 如果是其他脚本格式但不包含 "sh"，则认为是普通脚本，也尝试使用sh去执行，报错也将被正常记录
		}
	}

	// 通过文件扩展名进行检查
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".sh":
		return "shell_script", nil
	// 待支持的其他类型可以在这里添加
	default:
		if isBinaryExecutable(buffer[:n]) {
			return "binary_executable", nil
		}
		return "unknown", nil
	}
}

// 校验请求中携带的 password，如果未配置密码则直接通过
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
