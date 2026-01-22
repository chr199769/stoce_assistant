package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	eastmoney "stock_assistant/backend/common/eastmoney"
)

type MarketInfoTool struct {
	EastMoneyClient *eastmoney.Client
}

func NewMarketInfoTool() *MarketInfoTool {
	return &MarketInfoTool{
		EastMoneyClient: eastmoney.NewClient(),
	}
}

// ... existing code ...

type DragonTigerTool struct {
	EastMoneyClient *eastmoney.Client
}

func NewDragonTigerTool() *DragonTigerTool {
	return &DragonTigerTool{
		EastMoneyClient: eastmoney.NewClient(),
	}
}

func (t *DragonTigerTool) Name() string {
	return "DragonTigerList"
}

func (t *DragonTigerTool) Description() string {
	return "Get daily Dragon Tiger List (Longhu Bang) data. Input can be a date (YYYY-MM-DD) or empty for today."
}

func (t *DragonTigerTool) Call(ctx context.Context, input string) (string, error) {
	date := strings.TrimSpace(input)
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	items, err := t.EastMoneyClient.GetDragonTigerList(ctx, date)
	if err != nil {
		return fmt.Sprintf("Error fetching Dragon Tiger List: %v", err), nil
	}

	if len(items) == 0 {
		return "No Dragon Tiger List data found for this date.", nil
	}

	// Seat Mapping
	seatMap := map[string]string{
		"华泰证券股份有限公司北京雍和宫证券营业部":       "赵老哥",
		"国泰君安证券股份有限公司上海江苏路证券营业部":     "章盟主",
		"中国银河证券股份有限公司北京绍兴路证券营业部":     "赵老哥",
		"东方财富证券股份有限公司拉萨团结路第二证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨团结路第一证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第二证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第一证券营业部":   "拉萨天团",
		"招商证券股份有限公司深圳益田路免税商务大厦证券营业部": "益田路", // 校长?
		"中信证券股份有限公司上海溧阳路证券营业部":       "孙哥",
	}

	// Sort by Net Inflow
	sort.Slice(items, func(i, j int) bool {
		return items[i].NetInflow > items[j].NetInflow
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Dragon Tiger List (%s) - Top 5 Net Buy:\n", date))

	for i := 0; i < 5 && i < len(items); i++ {
		item := items[i]
		sb.WriteString(fmt.Sprintf("\n%d. %s (%s) | Change: %.2f%% | Net: %.1f万\n",
			i+1, item.Name, item.Code, item.ChangePercent, item.NetInflow/10000))
		sb.WriteString(fmt.Sprintf("   Reason: %s\n", item.Reason))

		// Fetch seats for Top 3 only to save API calls in this tool
		if i < 3 {
			buySeats, sellSeats, err := t.EastMoneyClient.GetDragonTigerSeats(ctx, item.Code, date)
			if err == nil {
				sb.WriteString("   [Top Buyer]:\n")
				for k, seat := range buySeats {
					if k >= 3 {
						break
					}
					tag := ""
					if t, ok := seatMap[seat.Name]; ok {
						tag = fmt.Sprintf("[%s]", t)
					} else if strings.Contains(seat.Name, "拉萨") {
						tag = "[拉萨天团]"
					} else if strings.Contains(seat.Name, "机构专用") {
						tag = "[机构]"
					} else if strings.Contains(seat.Name, "沪股通") || strings.Contains(seat.Name, "深股通") {
						tag = "[北向]"
					}
					sb.WriteString(fmt.Sprintf("     - %s %s: %.0f万\n", seat.Name, tag, seat.NetAmt/10000))
				}

				sb.WriteString("   [Top Seller]:\n")
				for k, seat := range sellSeats {
					if k >= 3 {
						break
					}
					tag := ""
					if t, ok := seatMap[seat.Name]; ok {
						tag = fmt.Sprintf("[%s]", t)
					} else if strings.Contains(seat.Name, "拉萨") {
						tag = "[拉萨天团]"
					} else if strings.Contains(seat.Name, "机构专用") {
						tag = "[机构]"
					} else if strings.Contains(seat.Name, "沪股通") || strings.Contains(seat.Name, "深股通") {
						tag = "[北向]"
					}
					sb.WriteString(fmt.Sprintf("     - %s %s: %.0f万\n", seat.Name, tag, seat.NetAmt/10000))
				}
			}
		}
	}

	return sb.String(), nil
}

// SinaNewsResponse matches Sina 7x24 API
type SinaNewsResponse struct {
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

// GetMarketNews fetches general market news (kuaixun) using Sina 7x24 API (more reliable)
func GetMarketNews() ([]string, error) {
	url := "https://zhibo.sina.com.cn/api/zhibo/feed?page=1&page_size=50&zhibo_id=152"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sinaResp SinaNewsResponse
	if err := json.Unmarshal(body, &sinaResp); err != nil {
		return nil, fmt.Errorf("json 解析错误: %v", err)
	}

	if sinaResp.Result.Status.Code != 0 {
		return nil, fmt.Errorf("api 错误: %s", sinaResp.Result.Status.Msg)
	}

	var news []string
	for _, item := range sinaResp.Result.Data.Feed.List {
		// Clean text (Sina rich_text might have HTML or special chars)
		text := item.RichText
		// Basic cleaning if needed, usually it's plain text or minimal HTML
		news = append(news, fmt.Sprintf("[%s] %s", item.CreateTime, text))
	}
	return news, nil
}

func (t *MarketInfoTool) Name() string {
	return "MarketInfo"
}

func (t *MarketInfoTool) Description() string {
	return "综合市场情报工具。用于获取：1. 特定股票新闻（输入股票代码），2. 社交趋势（头条/百度/微博），3. 一般市场新闻和政策。输入可以是股票代码（如 '600519'）或为空以获取一般市场信息。"
}

func (t *MarketInfoTool) Call(ctx context.Context, input string) (string, error) {
	stockCode := strings.TrimSpace(input)
	if stockCode == "all" || stockCode == "market" {
		stockCode = ""
	}

	var sb strings.Builder

	// 1. 特定股票信息（如果提供了代码）
	if stockCode != "" {
		sb.WriteString(fmt.Sprintf("=== %s 的详细信息 ===\n", stockCode))

		// A. 股票新闻
		sb.WriteString("[近期新闻]\n")
		news, err := t.EastMoneyClient.GetStockNews(ctx, stockCode)
		if err != nil {
			sb.WriteString(fmt.Sprintf("获取新闻失败: %v\n", err))
		} else if len(news) == 0 {
			sb.WriteString("未找到近期新闻。\n")
		} else {
			for _, n := range news {
				sb.WriteString(fmt.Sprintf("- %s\n", n.String()))
			}
		}

		// B. 龙虎榜状态
		sb.WriteString("\n[龙虎榜状态]\n")
		onList, details, err := t.EastMoneyClient.GetDragonTigerStatus(ctx, stockCode)
		if err != nil {
			sb.WriteString(fmt.Sprintf("检查龙虎榜失败: %v\n", err))
		} else if onList {
			sb.WriteString(fmt.Sprintf("是。详情: %s\n", details))
		} else {
			sb.WriteString("否 (今日未上榜)\n")
		}
		sb.WriteString("\n")
	}

	// 2. 社交趋势 (宏观情绪)
	sb.WriteString("=== 社交趋势 (宏观情绪) ===\n")
	trends, err := GetAllTrends()
	if err != nil {
		sb.WriteString(fmt.Sprintf("获取趋势失败: %v\n", err))
	} else {
		// If input is specific stock, try to filter trends relevant to it?
		// Or just show top trends briefly?
		// Let's show all trends but maybe truncated if too long?
		// GetAllTrends returns a LOT of text.
		// Let's just append it. The LLM can handle it.
		sb.WriteString(trends)
	}
	sb.WriteString("\n")

	// 3. 一般市场新闻与政策
	sb.WriteString("=== 一般市场与政策新闻 ===\n")
	marketNews, err := GetMarketNews()
	if err != nil {
		sb.WriteString(fmt.Sprintf("获取市场新闻失败: %v\n", err))
	} else {
		// Filter logic
		keywords := []string{"马斯克", "特朗普", "政策", "行业", "板块", "Musk", "Trump", "央行", "证监会", "国务院"}
		// If stock code provided, maybe add it to keywords?
		if stockCode != "" {
			keywords = append(keywords, stockCode)
		}

		var relevantNews []string
		for _, news := range marketNews {
			for _, kw := range keywords {
				if strings.Contains(news, kw) {
					relevantNews = append(relevantNews, news)
					break
				}
			}
		}

		if len(relevantNews) > 0 {
			sb.WriteString("发现相关/重要新闻:\n")
			for _, n := range relevantNews {
				sb.WriteString(fmt.Sprintf("- %s\n", n))
			}
		} else {
			sb.WriteString("在前 50 条快讯中未发现关键人物/政策的具体提及。显示前 5 条一般新闻:\n")
			for i, n := range marketNews {
				if i >= 5 {
					break
				}
				sb.WriteString(fmt.Sprintf("- %s\n", n))
			}
		}
	}

	return sb.String(), nil
}
