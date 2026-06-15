package main

import (
	"flag"
	"fmt"
	"mini-crawler/crawler"
)

// ============================================================
// 入口：简易爬虫框架
//
// 使用方式:
//   go run main.go -w 5 https://example.com https://golang.org
//   go build -o crawler.exe   # 然后 ./crawler.exe
//
// 命令行参数:
//   -o string   输出文件 (默认 "result.json")
//   -w int      并发工作协程数 (默认 3)
//
// 如果命令行没有传入 URL，会使用默认示例 URL 做演示
// ============================================================

func main() {
	// ── 解析命令行参数 ──
	output := flag.String("o", "result.json", "输出 JSON 文件路径")
	workers := flag.Int("w", 3, "并发工作协程数")
	flag.Parse()

	// 从非 flag 参数中获取 URL 列表
	urls := flag.Args()
	if len(urls) == 0 {
		// 演示用默认 URL
		urls = []string{
			"https://httpbin.org/html",
			"https://example.com",
			"https://golang.org",
			"https://go.dev/doc/",
			"https://pkg.go.dev/std",
		}
		fmt.Println("📌 未指定 URL，使用默认示例 URL 演示")
	}

	// ── 打印启动信息 ──
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        简易爬虫框架 (Worker Pool 模式)    ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("  并发协程数: %d\n", *workers)
	fmt.Printf("  目标 URL 数: %d\n", len(urls))
	fmt.Println("────────────────────────────────────────────")

	// ── 核心: 调用爬虫引擎 ──
	// Run 内部会启动 N 个 goroutine 并发爬取
	results := crawler.Run(urls, *workers)

	// ── 保存结果到 JSON 文件 ──
	if err := crawler.SaveResults(results, *output); err != nil {
		fmt.Printf("❌ 保存结果失败: %v\n", err)
		return
	}

	// ── 输出统计 ──
	successCount := 0
	for _, r := range results {
		if r.Error == "" {
			successCount++
		}
	}

	fmt.Println("────────────────────────────────────────────")
	fmt.Printf("✅ 爬取完成!\n")
	fmt.Printf("   总计: %d 个 | 成功: %d 个 | 失败: %d 个\n",
		len(results), successCount, len(results)-successCount)
	fmt.Printf("   结果已保存到: %s\n", *output)
	fmt.Println("────────────────────────────────────────────")
	fmt.Println("💡 试试: go run main.go -w 10 <url1> <url2> ...")
}
