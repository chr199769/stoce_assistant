package sina

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock_assistant/backend/stock_service/kitex_gen/stock"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Client 处理新浪财经 API 交互
type Client struct {
	httpClient *http.Client
}

// NewClient 创建新的新浪财经 API 客户端
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetStockInfo 获取实时股票信息
// 代码格式: sh600000, sz000001
func (c *Client) GetStockInfo(ctx context.Context, code string) (*stock.StockInfo, error) {
	// 自动修复代码前缀
	code = strings.TrimSpace(code)
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") {
			code = "sh" + code
		} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
			code = "sz" + code
		}
	}

	// 新浪 API 格式: http://hq.sinajs.cn/list=sh601006
	url := fmt.Sprintf("http://hq.sinajs.cn/list=%s", code)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	// 设置 Header 模拟浏览器
	req.Header.Set("Referer", "https://finance.sina.com.cn/")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取数据失败: %v", err)
	}
	defer resp.Body.Close()

	// 新浪 API 返回 GBK/GB18030 编码，需转为 UTF-8
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var content string
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Body, _, err := transform.Bytes(decoder, rawBody)
	if err != nil {
		// 解码失败则使用原始字符串
		fmt.Printf("%s GBK 解码失败: %v. 使用原始字符串。\n", code, err)
		content = string(rawBody)
	} else {
		content = string(utf8Body)
	}
	// 响应格式: var hq_str_sh601006="大秦铁路,6.660,...";

	// 校验数据有效性
	if !strings.Contains(content, "=\"") {
		return nil, fmt.Errorf("无效股票代码或空响应. 内容: %s, URL: %s", content, url)
	}

	parts := strings.Split(content, "=\"")
	if len(parts) < 2 {
		return nil, fmt.Errorf("解析错误. 内容: %s", content)
	}

	dataStr := strings.TrimSuffix(parts[1], "\";\n")
	dataStr = strings.TrimSuffix(dataStr, "\";")

	if dataStr == "" {
		return nil, fmt.Errorf("数据为空")
	}

	fields := strings.Split(dataStr, ",")
	// 指数数据 (sh000001) 字段较少
	// 基本校验: 至少包含名称、价格、时间
	if len(fields) < 6 {
		return nil, fmt.Errorf("数据格式异常: 字段数 %d. 数据: %s, URL: %s", len(fields), dataStr, url)
	}

	// 解析字段
	name := fields[0]
	openPrice, _ := strconv.ParseFloat(fields[1], 64)
	_ = openPrice // fields[2] 是昨收
	prevClose, _ := strconv.ParseFloat(fields[2], 64)
	currentPrice, _ := strconv.ParseFloat(fields[3], 64)
	// fields[4] 最高, fields[5] 最低
	// fields[8] 成交量 (股)
	volume, _ := strconv.ParseInt(fields[8], 10, 64)
	date := fields[30]
	timeStr := fields[31]

	// 计算涨跌幅
	changePercent := 0.0
	if prevClose > 0 {
		changePercent = (currentPrice - prevClose) / prevClose * 100
	}

	return &stock.StockInfo{
		Code:          code,
		Name:          name,
		CurrentPrice:  currentPrice,
		ChangePercent: changePercent,
		Volume:        volume,
		Timestamp:     fmt.Sprintf("%s %s", date, timeStr),
	}, nil
}

type NewsItem struct {
	Title   string
	Content string
	Source  string
	Url     string
	Time    string
}

type sinaRollNewsResponse struct {
	Result struct {
		Status struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"status"`
		Data []struct {
			Title     string `json:"title"`
			Url       string `json:"url"`
			Intro     string `json:"intro"`
			MediaName string `json:"media_name"`
			Ctime     string `json:"ctime"` // Unix timestamp as string
		} `json:"data"`
	} `json:"result"`
}

// GetMarketNews 获取市场滚动新闻
func (c *Client) GetMarketNews(ctx context.Context) ([]*NewsItem, error) {
	url := "https://feed.mix.sina.com.cn/api/roll/get?pageid=153&lid=2509&k=&num=20&page=1"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result sinaRollNewsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	if result.Result.Status.Code != 0 {
		return nil, fmt.Errorf("API返回错误: %s", result.Result.Status.Msg)
	}

	var items []*NewsItem
	for _, item := range result.Result.Data {
		// 如果 Intro 为空，使用 Title
		content := item.Intro
		if content == "" {
			content = item.Title
		}

		items = append(items, &NewsItem{
			Title:   item.Title,
			Content: content,
			Source:  item.MediaName,
			Url:     item.Url,
			Time:    item.Ctime,
		})
	}

	return items, nil
}

// Get7x24LiveNews matches Sina 7x24 API
type SinaLiveNewsResponse struct {
	Result struct {
		Status struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		} `json:"status"`
		Data struct {
			Feed struct {
				List []struct {
					RichText   string `json:"rich_text"`
					CreateTime string `json:"create_time"`
					DocUrl     string `json:"doc_url"`
				} `json:"list"`
			} `json:"feed"`
		} `json:"data"`
	} `json:"result"`
}

// Get7x24LiveNews 获取新浪7x24小时直播快讯
func (c *Client) Get7x24LiveNews(ctx context.Context) ([]*NewsItem, error) {
	url := "https://zhibo.sina.com.cn/api/zhibo/feed?page=1&page_size=50&zhibo_id=152"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var sinaResp SinaLiveNewsResponse
	if err := json.Unmarshal(body, &sinaResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	if sinaResp.Result.Status.Code != 0 {
		return nil, fmt.Errorf("API返回错误: %s", sinaResp.Result.Status.Msg)
	}

	var items []*NewsItem
	for _, item := range sinaResp.Result.Data.Feed.List {
		// Clean text (Sina rich_text might have HTML or special chars)
		text := item.RichText
		// Basic cleaning if needed, usually it's plain text or minimal HTML
		items = append(items, &NewsItem{
			Title:   fmt.Sprintf("快讯: %s...", string([]rune(text))[:min(len([]rune(text)), 20)]), // 截取前20字作为标题
			Content: text,
			Source:  "新浪7x24",
			Url:     item.DocUrl,
			Time:    item.CreateTime,
		})
	}

	return items, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
