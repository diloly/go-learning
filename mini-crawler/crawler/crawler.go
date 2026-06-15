package crawler

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ============================================================
// 抓取模块：使用 net/http 标准库发送 HTTP 请求
// ============================================================

// FetchURL 发送 HTTP GET 请求，返回网页内容 HTML 字符串和状态码
//
// 学习重点:
//   - http.Client: 可以设置超时、重定向策略等
//   - http.NewRequest: 构造请求，可以设置 Header
//   - 设置 User-Agent: 有些网站会拦截无 UA 的请求
//   - io.ReadAll: 读取响应体所有内容
//   - defer resp.Body.Close(): 记得关闭响应体，避免资源泄漏
func FetchURL(url string) (string, int, error) {
	// 创建一个带超时的 HTTP 客户端
	// 这是生产环境的推荐做法——没有超时的客户端可能永远阻塞
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 构造 GET 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", 0, err
	}

	// 设置请求头，模拟浏览器访问
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close() // ⚠️ 务必关闭，否则会有内存泄漏

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(body), resp.StatusCode, nil
}

// ============================================================
// 解析模块：使用 goquery 解析 HTML（类似 jQuery 的 API）
//
// goquery 是 Go 生态中最流行的 HTML 解析库之一
// 安装: go get github.com/PuerkitoBio/goquery
// ============================================================

// ParseHTML 解析 HTML，提取标题、纯文本内容和所有链接
//
// 学习重点:
//   - goquery.NewDocumentFromReader: 从 reader 创建文档对象
//   - doc.Find("title"): CSS 选择器，和 jQuery 一模一样
//   - s.Text(): 获取选中元素的文本内容
//   - s.Attr("href"): 获取选中元素的属性值
//   - Each(func(i int, s *goquery.Selection)): 遍历所有匹配元素
func ParseHTML(htmlContent string) (title string, text string, links []string) {
	// 将 HTML 字符串转为 goquery 文档对象
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", "", nil
	}

	// 提取 <title> 标签内容
	title = strings.TrimSpace(doc.Find("title").Text())

	// 提取 <body> 中的纯文本
	text = strings.TrimSpace(doc.Find("body").Text())

	// 提取所有 <a href="..."> 链接
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			href = strings.TrimSpace(href)
			if href != "" {
				links = append(links, href)
			}
		}
	})

	return
}
