package crawler

// ============================================================
// CrawlResult 爬取结果结构体
//
// json tag 控制序列化时的字段名
// omitempty 表示如果字段为零值，序列化时省略该字段
// ============================================================
type CrawlResult struct {
	URL    string   `json:"url"`              // 目标 URL
	Title  string   `json:"title"`            // 页面标题（<title> 标签）
	Text   string   `json:"text"`             // 页面文本内容
	Links  []string `json:"links"`            // 页面中提取的所有链接
	Status int      `json:"status_code"`      // HTTP 响应状态码
	Error  string   `json:"error,omitempty"`  // 爬取出错时的错误信息（成功时为空）
}
