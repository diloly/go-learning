package crawler

import (
	"fmt"
	"sync"
)

// ============================================================
// 并发引擎：Worker Pool 模式
//
// 这是整个项目的灵魂，集中展示 Go 并发三件套:
//   1. goroutine  —— 用 go 关键字启动轻量级线程
//   2. channel    —— 协程间通信（"不要通过共享内存来通信，而是通过通信来共享内存"）
//   3. WaitGroup  —— 等待一组协程全部完成
//
// 设计模式: Worker Pool（工作者池）
//   - 主协程: 生产者，把 URL 发送到 tasks channel
//   - N 个 Worker: 消费者，从 tasks channel 取 URL，爬取，结果发到 results channel
//   - 主协程: 收集所有结果
//
// 关键流程:
//   创建 tasks channel ──→ 启动 N 个 worker goroutine ──→ 发送任务 ──→ close(tasks)
//   ──→ worker 自动退出 for-range ──→ wg.Wait() ──→ close(results) ──→ 收集结果
// ============================================================

// Run 启动爬虫引擎: 并发爬取多个 URL，返回所有结果
//
// 参数:
//   - urls:   待爬取的 URL 列表
//   - workers: 并发工作协程数量（即同时爬多少个页面）
//
// 返回值:
//   - []CrawlResult: 爬取结果列表（顺序与输入不一定一致！）
func Run(urls []string, workers int) []CrawlResult {
	if len(urls) == 0 {
		return nil
	}
	if workers <= 0 {
		workers = 1
	}

	// ── 第 1 步: 创建两个 channel ──
	//
	// channel 是 Go 协程间通信的管道
	// 带缓冲 channel: make(chan T, size)，缓冲区满时发送方阻塞，空时接收方阻塞
	tasks := make(chan string, len(urls))   // 任务队列：传递 URL
	results := make(chan CrawlResult, len(urls)) // 结果队列：传递爬取结果

	// ── 第 2 步: 启动固定数量的 worker 协程 ──
	//
	// sync.WaitGroup 是一个计数器
	//   - Add(delta): 计数器 +delta
	//   - Done():     计数器 -1（通常在 defer 中调用）
	//   - Wait():     阻塞直到计数器归零
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)               // 告诉 WaitGroup "有一个协程要跟踪"
		go worker(i, tasks, results, &wg) // 启动 worker 协程
	}

	// ── 第 3 步: 发送所有 URL 到任务队列 ──
	for _, url := range urls {
		tasks <- url // 将 URL 发送到 tasks channel
	}
	close(tasks) // ⚠️ 关键！关闭 channel，表示"没有更多任务了"
	//
	// 为什么必须 close?
	//   worker 用 for-range 读取 channel，range 会在 channel 关闭且数据读完时退出
	//   如果不 close，worker 会永远阻塞等待新任务 → 死锁！

	// ── 第 4 步: 等待所有 worker 完成 ──
	wg.Wait()    // 阻塞，直到所有 worker 调用了 wg.Done()
	close(results) // 所有 worker 都结束了，关闭 results channel

	// ── 第 5 步: 收集结果 ──
	var resultList []CrawlResult
	for r := range results { // range 自动处理 channel 关闭
		resultList = append(resultList, r)
	}

	return resultList
}

// worker 是一个工作协程
//
// 参数中的 channel 方向注解:
//   - tasks  <-chan string   : 只读 channel（只能接收）
//   - results chan<- CrawlResult: 只写 channel（只能发送）
//   这是 Go 类型系统提供的"防呆设计"——在函数签名中标明 channel 用途
func worker(id int, tasks <-chan string, results chan<- CrawlResult, wg *sync.WaitGroup) {
	// defer wg.Done(): 函数返回时自动告诉 WaitGroup"本协程已完成"
	// 用 defer 确保即使发生 panic 也会执行 Done()，避免主协程永远等下去
	defer wg.Done()

	// for-range 从 channel 读取数据
	// 当 channel 被关闭且缓冲区为空时，for-range 自动退出
	for url := range tasks {
		fmt.Printf("[Worker %d] ▶ 开始爬取: %s\n", id, url)

		// 1. 抓取网页
		htmlContent, statusCode, err := FetchURL(url)
		if err != nil {
			fmt.Printf("[Worker %d] ✗ 失败: %s — %v\n", id, url, err)
			// 即使出错，也把结果发回，由调用方处理
			results <- CrawlResult{
				URL:    url,
				Status: statusCode,
				Error:  err.Error(),
			}
			continue // 跳过本次循环，处理下一个 URL
		}

		// 2. 解析 HTML
		title, text, links := ParseHTML(htmlContent)

		// 3. 限制文本长度，避免结果文件过大
		runes := []rune(text) // 用 []rune 处理中文字符（一个中文占多个 byte）
		if len(runes) > 500 {
			text = string(runes[:500]) + "..."
		}

		// 4. 发送结果
		results <- CrawlResult{
			URL:    url,
			Title:  title,
			Text:   text,
			Links:  links,
			Status: statusCode,
		}

		fmt.Printf("[Worker %d] ✓ 完成: %s | 标题: %s | 链接数: %d\n", id, url, title, len(links))
	}

	fmt.Printf("[Worker %d] ⏹ 退出（任务通道已关闭，无更多任务）\n", id)
}
