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
		return nil, fmt.Errorf("json parse error: %v", err)
	}

	if sinaResp.Result.Status.Code != 0 {
		return nil, fmt.Errorf("api error: %s", sinaResp.Result.Status.Msg)
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
	return "Comprehensive Market Intelligence Tool. Use it to fetch: 1. Specific Stock News (input stock code), 2. Social Trends (Toutiao/Baidu/Weibo), 3. General Market News & Policy. Input can be a stock code (e.g., '600519') or empty for general market info."
}

func (t *MarketInfoTool) Call(ctx context.Context, input string) (string, error) {
	stockCode := strings.TrimSpace(input)
	if stockCode == "all" || stockCode == "market" {
		stockCode = ""
	}

	var sb strings.Builder

	// 1. Specific Stock Info (if code provided)
	if stockCode != "" {
		sb.WriteString(fmt.Sprintf("=== Specific Info for %s ===\n", stockCode))

		// A. Stock News
		sb.WriteString("[Recent News]\n")
		news, err := t.EastMoneyClient.GetStockNews(ctx, stockCode)
		if err != nil {
			sb.WriteString(fmt.Sprintf("Error fetching news: %v\n", err))
		} else if len(news) == 0 {
			sb.WriteString("No recent news found.\n")
		} else {
			for _, n := range news {
				sb.WriteString(fmt.Sprintf("- %s\n", n.String()))
			}
		}

		// B. Dragon & Tiger List
		sb.WriteString("\n[Dragon & Tiger List Status]\n")
		onList, details, err := t.EastMoneyClient.GetDragonTigerStatus(ctx, stockCode)
		if err != nil {
			sb.WriteString(fmt.Sprintf("Error checking list: %v\n", err))
		} else if onList {
			sb.WriteString(fmt.Sprintf("YES. Details: %s\n", details))
		} else {
			sb.WriteString("No (Not on today's list)\n")
		}
		sb.WriteString("\n")
	}

	// 2. Social Trends (Macro Sentiment)
	sb.WriteString("=== Social Trends (Macro Sentiment) ===\n")
	trends, err := GetAllTrends()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error fetching trends: %v\n", err))
	} else {
		// If input is specific stock, try to filter trends relevant to it?
		// Or just show top trends briefly?
		// Let's show all trends but maybe truncated if too long?
		// GetAllTrends returns a LOT of text.
		// Let's just append it. The LLM can handle it.
		sb.WriteString(trends)
	}
	sb.WriteString("\n")

	// 3. General Market News & Policy
	sb.WriteString("=== General Market & Policy News ===\n")
	marketNews, err := GetMarketNews()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error fetching market news: %v\n", err))
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
			sb.WriteString("Found relevant/influential news:\n")
			for _, n := range relevantNews {
				sb.WriteString(fmt.Sprintf("- %s\n", n))
			}
		} else {
			sb.WriteString("No specific mentions of Key Figures/Policy in top 50 flash news. Showing top 5 general news:\n")
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
