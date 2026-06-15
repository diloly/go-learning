package crawler

import (
	"encoding/json"
	"os"
)

// ============================================================
// 存储模块：将爬取结果序列化为 JSON 并保存到本地文件
//
// 学习重点:
//   - encoding/json: Go 标准库的 JSON 编解码
//   - json.NewEncoder: 流式编码器，适合写文件
//   - SetIndent: 美化输出，便于阅读
//   - os.Create: 创建/覆盖文件
// ============================================================

// SaveResults 将爬取结果序列化为 JSON，保存到指定文件路径
func SaveResults(results []CrawlResult, filepath string) error {
	// os.Create: 创建文件（如果已存在则清空）
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	// defer f.Close(): 函数结束时关闭文件
	// 注意: 在写入后关闭文件很重要，确保数据刷入磁盘
	defer f.Close()

	// json.NewEncoder: 创建一个 JSON 编码器，将数据编码并写入文件
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ") // 美化输出: 2 空格缩进，方便人类阅读

	// Encode: 将结果列表编码为 JSON 数组并写入文件
	return encoder.Encode(results)
}
