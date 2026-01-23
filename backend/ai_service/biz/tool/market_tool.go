package tool

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"stock_assistant/backend/ai_service/biz/rpc"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
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
	trendReq := &stock.GetMarketTrendsRequest{
		Page:     1,
		PageSize: 100, // 获取更多趋势用于分析
		Sort:     "weight_desc",
	}
	if stockCode != "" {
		trendReq.RelatedStockId = stockCode
	}

	if rpc.StockClient == nil {
		sb.WriteString("Stock Service 未初始化，无法获取趋势。\n")
	} else {
		trendResp, err := rpc.StockClient.GetMarketTrends(ctx, trendReq)
		if err != nil {
			sb.WriteString(fmt.Sprintf("获取趋势失败: %v\n", err))
		} else {
			if len(trendResp.Trends) == 0 {
				if stockCode != "" {
					sb.WriteString(fmt.Sprintf("暂无与 %s 相关的趋势数据。\n", stockCode))
				} else {
					sb.WriteString("暂无趋势数据。\n")
				}
			}

			// Previously sorted manually here, now sorting is handled by DB via API parameter.

			for i, trend := range trendResp.Trends {
				sb.WriteString(fmt.Sprintf("%d. [%s] %s (相关性:%d, 权重:%.1f): %s\n",
					i+1, trend.ImpactType, trend.Title, trend.FinancialRelevance, trend.Weight, trend.Summary))
			}
		}
	}
	sb.WriteString("\n")

	// 3. 一般市场新闻与政策 (已集成到市场趋势中)
	// 此处无需再次抓取，因为 Stock Service 已经合并了快讯和新闻，并进行了分析
	// 但为了用户体验，如果趋势数据很少，可以在这里提示用户

	return sb.String(), nil
}
