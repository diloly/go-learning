// Package filemanager 提供基础的文件与目录操作。
// 涵盖：遍历列表、创建/删除/重命名、目录大小统计。
package filemanager

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ──────────────────────────────────────────────
// 列出目录内容
// ──────────────────────────────────────────────

// List 返回 dir 中所有条目（文件 + 目录）的格式化列表。
// full = true 输出绝对路径，否则输出相对路径。
func List(dir string, full bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取目录 %s 失败: %w", dir, err)
	}

	var result []string
	for _, e := range entries {
		info, _ := e.Info()
		path := e.Name()
		if full {
			path = filepath.Join(dir, e.Name())
		}

		// 用 os.FileMode 判断类型，附加标识符
		var modeInfo string
		if info != nil {
			modeInfo = info.Mode().String()
		}
		if e.IsDir() {
			result = append(result, fmt.Sprintf("[DIR]  %s  (%s)", path, modeInfo))
		} else {
			size := ""
			if info != nil {
				size = humanSize(info.Size())
			}
			result = append(result, fmt.Sprintf("[FILE] %s  %s  (%s)", path, size, modeInfo))
		}
	}
	return result, nil
}

// ──────────────────────────────────────────────
// 创建
// ──────────────────────────────────────────────

// CreateFile 创建空文件。如果父目录不存在则先创建。
// os.Create 会创建或截断已存在的文件。
func CreateFile(path string) error {
	// 确保父目录存在
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("创建父目录失败: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	return f.Close()
}

// CreateDir 创建目录（等价于 mkdir -p）。
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ──────────────────────────────────────────────
// 删除
// ──────────────────────────────────────────────

// Remove 删除单个文件或空目录，等价于 rm / rmdir。
// 如果路径不存在或目录非空会返回 error。
func Remove(path string) error {
	return os.Remove(path) // 只删文件或空目录
}

// RemoveAll 递归删除整个路径，等价于 rm -rf。
// ⚠️ 慎用，删除不可恢复。
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// ──────────────────────────────────────────────
// 重命名 / 移动
// ──────────────────────────────────────────────

// Rename 重命名或移动文件/目录，等价于 mv。
// 跨设备会失败（不同磁盘分区不支持）。
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// ──────────────────────────────────────────────
// 目录大小统计
// ──────────────────────────────────────────────

// Size 递归遍历 dir 下所有文件，累加字节数。
// 返回字节数 bytes、人类可读字符串 human。
func Size(dir string) (bytes int64, human string, err error) {
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e // 遇到权限问题会直接返回错误
		}
		if !d.IsDir() {
			info, infoErr := d.Info()
			if infoErr != nil {
				return infoErr
			}
			bytes += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, "", fmt.Errorf("统计目录大小失败: %w", err)
	}
	human = humanSize(bytes)
	return
}

// ──────────────────────────────────────────────
// 辅助函数
// ──────────────────────────────────────────────

// humanSize 将字节数转为可读形式（KB / MB / GB）。
func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}
