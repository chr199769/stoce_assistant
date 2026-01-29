package tool

import (
	"context"
	"fmt"
	"log"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
	"strings"

	eastmoney "stock_assistant/backend/common/eastmoney"
	eimpl "stock_assistant/backend/common/provider/impl/eastmoney"
	"stock_assistant/backend/common/provider"
)

type StockPriceTool struct {
	Client       stockservice.Client
	MarketClient provider.MarketDataClient
}

func NewStockPriceTool(client stockservice.Client) *StockPriceTool {
	return &StockPriceTool{
		Client:       client,
		MarketClient: eimpl.NewMarket(eastmoney.NewClient()),
	}
}

func (t *StockPriceTool) Name() string {
	return "StockPrice"
}

func (t *StockPriceTool) Description() string {
	return "用于获取实时股票价格和信息。输入应为股票代码（例如：sh600519）。"
}

func (t *StockPriceTool) Call(ctx context.Context, input string) (string, error) {
	// Clean input: trim whitespace and take only the first line/word
	input = strings.TrimSpace(input)
	// Handle potential carriage returns or newlines
	if idx := strings.IndexAny(input, "\r\n"); idx != -1 {
		input = input[:idx]
	}
	if idx := strings.Index(input, " "); idx != -1 {
		input = input[:idx]
	}
	// Remove any potential "Observation:" prefix or suffix if the parser failed to strip it
	input = strings.TrimPrefix(input, "Observation:")
	input = strings.TrimSpace(input)

	// Ensure no control characters remain
	input = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)

	log.Printf("StockPriceTool 被调用，输入: [%s]\n", input)
	// 优先走统一 Provider 实时报价
	code := normalizeCode(input)
	if q, err := t.MarketClient.GetIntradayQuote(ctx, code); err == nil && q != nil {
		result := fmt.Sprintf("股票: %s (%s), 价格: %.2f, 涨跌幅: %.2f%%, 成交量: %d",
			q.Name, q.Code, q.Price, q.ChangePct, q.Volume)
		log.Printf("StockPriceTool 成功(Provider): %s\n", result)
		return result, nil
	}
	// 回退到 RPC (新浪)
	req := &stock.GetRealtimeRequest{Code: input}
	resp, err := t.Client.GetRealtime(ctx, req)
	if err != nil {
		log.Printf("StockPriceTool 获取实时数据错误: %v\n", err)
		return fmt.Sprintf("获取股票数据失败: %v", err), nil
	}
	if resp.Stock == nil {
		log.Printf("StockPriceTool: 未找到股票: %s\n", input)
		return "未找到股票", nil
	}
	result := fmt.Sprintf("股票: %s (%s), 价格: %.2f, 涨跌幅: %.2f%%, 成交量: %d",
		resp.Stock.Name, resp.Stock.Code, resp.Stock.CurrentPrice, resp.Stock.ChangePercent, resp.Stock.Volume)
	log.Printf("StockPriceTool 成功(RPC): %s\n", result)
	return result, nil
}

type StockAnalysisTool struct {
	EastMoneyClient *eastmoney.Client
	MarketClient    provider.MarketDataClient
	DragonClient    provider.DragonTigerClient
	NoticeClient    provider.NoticeClient
	PopularityClient provider.PopularityClient
	IndustryClient  provider.IndustryClient
	ChipClient      provider.ChipClient
}

func NewStockAnalysisTool() *StockAnalysisTool {
	return &StockAnalysisTool{
		EastMoneyClient: eastmoney.NewClient(),
		MarketClient:    eimpl.NewMarket(eastmoney.NewClient()),
		DragonClient:    eimpl.NewDragonTiger(eastmoney.NewClient()),
		NoticeClient:    eimpl.NewNotice(eastmoney.NewClient()),
		PopularityClient: eimpl.NewPopularity(eastmoney.NewClient()),
		IndustryClient:  eimpl.NewIndustry(eastmoney.NewClient()),
		ChipClient:      eimpl.NewChip(eastmoney.NewClient()),
	}
}

func (t *StockAnalysisTool) Name() string {
	return "StockAnalysis"
}

func (t *StockAnalysisTool) Description() string {
	return "用于获取高级股票分析数据，包括龙虎榜历史、筹码分布、盘口、行业信息、股吧热度和监管公告。输入应为股票代码（例如：600519）。"
}

func (t *StockAnalysisTool) Call(ctx context.Context, input string) (string, error) {
	// Clean input: trim whitespace and take only the first line/word
	input = strings.TrimSpace(input)
	// Handle potential carriage returns or newlines
	if idx := strings.IndexAny(input, "\r\n"); idx != -1 {
		input = input[:idx]
	}
	if idx := strings.Index(input, " "); idx != -1 {
		input = input[:idx]
	}
	// Remove any potential "Observation:" prefix or suffix
	input = strings.TrimPrefix(input, "Observation:")
	input = strings.TrimSpace(input)

	// Ensure no control characters remain
	input = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, input)

	log.Printf("StockAnalysisTool 被调用，输入: [%s]\n", input)

	// Fetch data
	code := normalizeCode(input)
	// 1. Industry
	industryData, err := t.IndustryClient.GetIndustryIndex(ctx, code)
	industry := "获取行业信息失败"
	if err != nil {
		log.Printf("获取行业失败: %v", err)
	} else {
		if industryData != nil {
			industry = fmt.Sprintf("行业: %s, 地域: %s, 概念: %s",
				industryData.IndustryName, industryData.RegionName, industryData.ConceptNames)
		} else {
			industry = "无行业属性数据"
		}
	}

	// 2. Order Book
	ordersData2, err := t.MarketClient.GetOrderBook(ctx, code, 5)
	orders := "获取盘口失败"
	if err != nil {
		log.Printf("获取盘口失败: %v", err)
	} else {
		orders = fmt.Sprintf("买一: %.2f/%d, 卖一: %.2f/%d, 委比: %.2f, 委差: %.2f",
			ordersData2.Bids[0].Price, ordersData2.Bids[0].Volume,
			ordersData2.Asks[0].Price, ordersData2.Asks[0].Volume,
			ordersData2.WeiBi, ordersData2.WeiCha)
	}

	// 3. Chip Distribution
	chipData, err := t.ChipClient.GetChipDistribution(ctx, code)
	chip := "获取筹码分布失败"
	if err != nil {
		log.Printf("获取筹码分布失败: %v", err)
	} else {
		if chipData == nil {
			chip = "无筹码分布数据"
		} else {
			chip = fmt.Sprintf("均价: %.2f, 胜率: %.2f%%, 90%%成本区间: %.2f-%.2f",
				chipData.AverageCost, chipData.WinnerRate, chipData.Cost90Low, chipData.Cost90High)
		}
	}

	// 4. Dragon Tiger History
	lhbData2, err := t.DragonClient.GetHistory(ctx, code, 5)
	var lhb []string
	if err != nil {
		log.Printf("获取龙虎榜历史失败: %v", err)
	} else {
		for _, item := range lhbData2 {
			lhb = append(lhb, fmt.Sprintf("%s %s 收盘:%.2f 涨跌:%.2f%% 原因:%s 净流入:%.2f",
				item.Date, item.Code, item.ClosePrice, item.ChangePercent, item.Reason, item.NetInflow))
		}
	}

	// 6. Stock Heat (Sentiment)
	heatData, err := t.PopularityClient.GetStockHeat(ctx, code)
	heat := "获取股票热度失败"
	if err != nil {
		log.Printf("获取股票热度失败: %v", err)
	} else {
		if heatData == nil {
			heat = "股吧排名: >100 (未进入前100)"
		} else {
			heat = fmt.Sprintf("股吧热度: 排名%d 分数%d 来源:%s", heatData.Rank, heatData.HeatScore, heatData.Source)
		}
	}

	// 7. Regulatory Notices (Risk)
	noticesData, err := t.NoticeClient.GetStockNotices(ctx, code, "", 10)
	var notices []string
	if err != nil {
		log.Printf("获取公告失败: %v", err)
	} else {
		for _, item := range noticesData {
			notices = append(notices, fmt.Sprintf("[%s] %s (%s)", item.Category, item.Title, item.PublishedAt.Format("2006-01-02")))
		}
	}

	// Format output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s 的分析报告:\n", input))

	sb.WriteString("\n\n[行业信息]\n")
	sb.WriteString(industry)

	sb.WriteString("\n\n[实时盘口]\n")
	sb.WriteString(orders)

	sb.WriteString("\n\n[筹码分布 (成本结构)]\n")
	sb.WriteString(chip)

	sb.WriteString("\n\n[市场情绪 & 资金]\n")
	sb.WriteString(heat)

	sb.WriteString("\n\n[龙虎榜 (最近 5 次)]\n")
	if len(lhb) > 0 {
		for _, l := range lhb {
			sb.WriteString(l + "\n")
		}
	} else {
		sb.WriteString("近期无记录。\n")
	}

	sb.WriteString("\n\n[监管公告 (风险提示)]\n")
	if len(notices) > 0 {
		for _, n := range notices {
			sb.WriteString(n + "\n")
		}
	} else {
		sb.WriteString("近期无监管公告。\n")
	}

	return sb.String(), nil
}

func normalizeCode(in string) string {
	s := strings.ToUpper(strings.TrimSpace(in))
	if strings.HasPrefix(s, "SH") || strings.HasPrefix(s, "SZ") {
		return s
	}
	if len(s) == 6 && s[0] == '6' {
		return "SH" + s
	}
	if len(s) == 6 {
		return "SZ" + s
	}
	return s
}
