package tool

import (
	"context"
	"fmt"
	"log"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
	"strings"

	eastmoney "stock_assistant/backend/common/eastmoney"
)

type StockPriceTool struct {
	Client stockservice.Client
}

func NewStockPriceTool(client stockservice.Client) *StockPriceTool {
	return &StockPriceTool{Client: client}
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
	req := &stock.GetRealtimeRequest{Code: input}
	resp, err := t.Client.GetRealtime(ctx, req)
	if err != nil {
		log.Printf("StockPriceTool 获取实时数据错误: %v\n", err)
		// Return error as string observation so the Agent knows it failed
		return fmt.Sprintf("获取股票数据失败: %v", err), nil
	}
	if resp.Stock == nil {
		log.Printf("StockPriceTool: 未找到股票: %s\n", input)
		return "未找到股票", nil
	}
	result := fmt.Sprintf("股票: %s (%s), 价格: %.2f, 涨跌幅: %.2f%%, 成交量: %d",
		resp.Stock.Name, resp.Stock.Code, resp.Stock.CurrentPrice, resp.Stock.ChangePercent, resp.Stock.Volume)
	log.Printf("StockPriceTool 成功: %s\n", result)
	return result, nil
}

type StockAnalysisTool struct {
	EastMoneyClient *eastmoney.Client
}

func NewStockAnalysisTool() *StockAnalysisTool {
	return &StockAnalysisTool{
		EastMoneyClient: eastmoney.NewClient(),
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
	// 1. Industry
	industryData, err := t.EastMoneyClient.GetIndustryIndex(ctx, input)
	industry := "获取行业信息失败"
	if err != nil {
		log.Printf("获取行业失败: %v", err)
	} else {
		industry = industryData.String()
	}

	// 2. Order Book
	ordersData, err := t.EastMoneyClient.GetOrderBook(ctx, input)
	orders := "获取盘口失败"
	if err != nil {
		log.Printf("获取盘口失败: %v", err)
	} else {
		orders = ordersData.String()
	}

	// 3. Chip Distribution
	chipData, err := t.EastMoneyClient.GetChipDistribution(ctx, input)
	chip := "获取筹码分布失败"
	if err != nil {
		log.Printf("获取筹码分布失败: %v", err)
	} else if chipData == nil {
		chip = "无筹码分布数据"
	} else {
		chip = chipData.String()
	}

	// 4. Dragon Tiger History
	lhbData, err := t.EastMoneyClient.GetDragonTigerHistory(ctx, input, 5)
	var lhb []string
	if err != nil {
		log.Printf("获取龙虎榜历史失败: %v", err)
	} else {
		for _, item := range lhbData {
			lhb = append(lhb, item.String())
		}
	}

	// 6. Stock Heat (Sentiment)
	heatData, err := t.EastMoneyClient.GetStockHeat(ctx, input)
	heat := "获取股票热度失败"
	if err != nil {
		log.Printf("获取股票热度失败: %v", err)
	} else if heatData == nil {
		heat = "股吧排名: >100 (未进入前100)"
	} else {
		heat = heatData.String()
	}

	// 7. Regulatory Notices (Risk)
	noticesData, err := t.EastMoneyClient.GetStockNotices(ctx, input, []string{"监管", "问询", "关注函", "立案", "警示"})
	var notices []string
	if err != nil {
		log.Printf("获取公告失败: %v", err)
	} else {
		for _, item := range noticesData {
			notices = append(notices, item.String())
		}
	}

	// 8. Quantitative Risk Control (Severe Abnormal Fluctuation)
	riskCheck := CheckRiskControlRules(ctx, t.EastMoneyClient, input)

	// Format output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s 的分析报告:\n", input))

	sb.WriteString("\n[量化风控检查]\n")
	sb.WriteString(riskCheck)

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
