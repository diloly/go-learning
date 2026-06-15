// main.go — go-cli-tool 的入口。
// 演示 Go 标准库 flag 做命令行参数解析，以及子命令的路由分发模式。
//
// 使用方式：
//
//	文件管理器：
//	  go run main.go fm list [--dir .] [--full]
//	  go run main.go fm create --path test.txt
//	  go run main.go fm delete --path test.txt
//	  go run main.go fm rename --old test.txt --new hello.txt
//	  go run main.go fm size [--dir .]
//
//	日志工具：
//	  go run main.go log write --level info --msg "hello world" [--file app.log]
//	  go run main.go log write --level fatal --msg "something wrong"

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-cli-tool/filemanager"
	"go-cli-tool/logger"
)

// subcommandUsage 显示子命令帮助。
func mainUsage() {
	fmt.Fprint(os.Stderr, `go-cli-tool — CLI 工具箱

子命令:
  fm   文件管理器（列表 / 创建 / 删除 / 重命名 / 大小统计）
  log  简易日志工具（按级别写入本地日志文件）

使用 "go-cli-tool <子命令> -h" 查看对应子命令的帮助。

`)
}

func main() {
	// ── 顶层解析：取出子命令名称 ──
	// flag 标准库默认只处理 os.Args[1:]，所以我们先手动取第一个参数。
	if len(os.Args) < 2 {
		mainUsage()
		os.Exit(1)
	}

	subCmd := os.Args[1]

	switch subCmd {
	case "fm":
		runFM(os.Args[2:])
	case "log":
		runLog(os.Args[2:])
	default:
		mainUsage()
		fmt.Fprintf(os.Stderr, "未知子命令: %s\n", subCmd)
		os.Exit(1)
	}
}

// ══════════════════════════════════════════════
// 文件管理器子命令
// ══════════════════════════════════════════════

func fmUsage() {
	fmt.Fprint(os.Stderr, `文件管理器 - 子命令:

  list     列出目录内容
  create   创建空文件或目录
  delete   删除文件或空目录
  rename   重命名/移动文件或目录
  size     统计目录总大小

使用 "go-cli-tool fm <子命令> -h" 查看详情。
`)
}

func runFM(args []string) {
	if len(args) < 1 {
		fmUsage()
		os.Exit(1)
	}

	action := args[0]

	switch action {
	case "list":
		fs := flag.NewFlagSet("fm list", flag.ExitOnError)
		dir := fs.String("dir", ".", "目标目录")
		full := fs.Bool("full", false, "是否输出绝对路径")
		_ = fs.Parse(args[1:]) // 直接用，ExitOnError 模式下错误会自己退出

		entries, err := filemanager.List(*dir, *full)
		if err != nil {
			log.Fatalf("列出目录失败: %v", err)
		}
		for _, e := range entries {
			fmt.Println(e)
		}

	case "create":
		fs := flag.NewFlagSet("fm create", flag.ExitOnError)
		path := fs.String("path", "", "要创建的文件或目录路径")
		isDir := fs.Bool("dir", false, "创建目录（而非文件）")
		_ = fs.Parse(args[1:])

		if *path == "" {
			fmt.Fprintln(os.Stderr, "请指定 --path")
			fs.Usage()
			os.Exit(1)
		}

		var err error
		if *isDir {
			err = filemanager.CreateDir(*path)
		} else {
			err = filemanager.CreateFile(*path)
		}
		if err != nil {
			log.Fatalf("创建失败: %v", err)
		}
		fmt.Printf("✅ 已创建: %s\n", *path)

	case "delete":
		fs := flag.NewFlagSet("fm delete", flag.ExitOnError)
		path := fs.String("path", "", "要删除的文件或空目录路径")
		recursive := fs.Bool("r", false, "递归删除（慎用！）")
		_ = fs.Parse(args[1:])

		if *path == "" {
			fmt.Fprintln(os.Stderr, "请指定 --path")
			fs.Usage()
			os.Exit(1)
		}

		var err error
		if *recursive {
			err = filemanager.RemoveAll(*path)
		} else {
			err = filemanager.Remove(*path)
		}
		if err != nil {
			log.Fatalf("删除失败: %v", err)
		}
		fmt.Printf("🗑️ 已删除: %s\n", *path)

	case "rename":
		fs := flag.NewFlagSet("fm rename", flag.ExitOnError)
		oldPath := fs.String("old", "", "原路径")
		newPath := fs.String("new", "", "新路径")
		_ = fs.Parse(args[1:])

		if *oldPath == "" || *newPath == "" {
			fmt.Fprintln(os.Stderr, "请指定 --old 和 --new")
			fs.Usage()
			os.Exit(1)
		}
		if err := filemanager.Rename(*oldPath, *newPath); err != nil {
			log.Fatalf("重命名失败: %v", err)
		}
		fmt.Printf("📝 已重命名: %s → %s\n", *oldPath, *newPath)

	case "size":
		fs := flag.NewFlagSet("fm size", flag.ExitOnError)
		dir := fs.String("dir", ".", "目标目录")
		_ = fs.Parse(args[1:])

		bytes, human, err := filemanager.Size(*dir)
		if err != nil {
			log.Fatalf("统计大小失败: %v", err)
		}
		fmt.Printf("📦 %s — %s（%d 字节）\n", *dir, human, bytes)

	default:
		fmUsage()
		fmt.Fprintf(os.Stderr, "未知操作: %s\n", action)
		os.Exit(1)
	}
}

// ══════════════════════════════════════════════
// 日志工具子命令
// ══════════════════════════════════════════════

func logUsage() {
	fmt.Fprint(os.Stderr, `日志工具 - 子命令:

  write   写入一条日志到文件

使用 "go-cli-tool log write -h" 查看详情。
`)
}

func runLog(args []string) {
	if len(args) < 1 {
		logUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "write":
		fs := flag.NewFlagSet("log write", flag.ExitOnError)
		level := fs.String("level", "info", "日志级别: info / warn / error / fatal")
		msg := fs.String("msg", "", "日志消息内容")
		file := fs.String("file", "app.log", "日志文件路径")
		_ = fs.Parse(args[1:])

		if *msg == "" {
			fmt.Fprintln(os.Stderr, "请指定 --msg")
			fs.Usage()
			os.Exit(1)
		}

		l, err := logger.New(*file)
		if err != nil {
			log.Fatalf("初始化日志失败: %v", err)
		}
		defer l.Close()

		switch *level {
		case "info":
			l.Info(*msg)
		case "warn":
			l.Warn(*msg)
		case "error":
			l.Error(*msg)
		case "fatal":
			// Fatal 内部会 os.Exit(1)，所以不需要再打印提示
			l.Fatal(*msg)
		default:
			fmt.Fprintf(os.Stderr, "未知级别: %s（可用: info/warn/error/fatal）\n", *level)
			os.Exit(1)
		}
		fmt.Printf("✅ 已写入 [%s] → %s\n", *level, *file)

	default:
		logUsage()
		fmt.Fprintf(os.Stderr, "未知操作: %s\n", args[0])
		os.Exit(1)
	}
}
