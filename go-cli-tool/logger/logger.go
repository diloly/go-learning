// Package logger 提供简易日志工具。
// 支持 4 个级别：INFO / WARN / ERROR / FATAL。
// 每条日志包含：时间戳、级别、消息，一行一条追加到文件。
package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Level 日志级别类型。
type Level int

const (
	INFO Level = iota
	WARN
	ERROR
	FATAL
)

// levelNames 级别与字符串的映射。
var levelNames = map[Level]string{
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// ──────────────────────────────────────────────
// Logger 核心结构
// ──────────────────────────────────────────────

// Logger 持有日志文件的句柄，负责写入与关闭。
// 使用 sync.Mutex 保证并发安全（多 goroutine 写入不混乱）。
type Logger struct {
	mu     sync.Mutex
	file   *os.File
	path   string
	closed bool
}

// New 创建（或追加打开）一个日志文件，返回 Logger 实例。
// 文件不存在则自动创建，已存在则追加写入。
func New(path string) (*Logger, error) {
	// os.O_APPEND | os.O_CREATE | os.O_WRONLY
	// 追加写入、文件不存在时创建、只写模式
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	return &Logger{file: f, path: path}, nil
}

// write 是内部写入方法，输出格式：
//
//	[2006-01-02 15:04:05] [INFO]  你的消息
func (l *Logger) write(level Level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		fmt.Fprintf(os.Stderr, "[logger 已关闭] 丢弃日志: %s\n", msg)
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] [%s]  %s\n", now, levelNames[level], msg)
	_, _ = l.file.WriteString(line)
}

// ──────────────────────────────────────────────
// 公开 API：按级别写入
// ──────────────────────────────────────────────

func (l *Logger) Info(msg string)  { l.write(INFO, msg) }
func (l *Logger) Warn(msg string)  { l.write(WARN, msg) }
func (l *Logger) Error(msg string) { l.write(ERROR, msg) }

// Fatal 写入后还会在 stderr 打印消息并退出进程（os.Exit(1)）。
func (l *Logger) Fatal(msg string) {
	l.write(FATAL, msg)
	fmt.Fprintln(os.Stderr, "[FATAL]", msg)
	os.Exit(1)
}

// Close 关闭日志文件。写入后务必调用。
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	return l.file.Close()
}
