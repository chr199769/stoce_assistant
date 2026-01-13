package tool

import (
	"context"
	"fmt"
	"log"
	"stock_assistant/backend/stock_service/kitex_gen/stock"
	"stock_assistant/backend/stock_service/kitex_gen/stock/stockservice"
	"strings"
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
	return "Useful for getting the realtime stock price and information. Input should be the stock code (e.g., sh600519)."
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

	log.Printf("StockPriceTool called with cleaned input: [%s]\n", input)
	req := &stock.GetRealtimeRequest{Code: input}
	resp, err := t.Client.GetRealtime(ctx, req)
	if err != nil {
		log.Printf("StockPriceTool GetRealtime error: %v\n", err)
		// Return error as string observation so the Agent knows it failed
		return fmt.Sprintf("Error fetching stock data: %v", err), nil
	}
	if resp.Stock == nil {
		log.Printf("StockPriceTool: Stock not found for input: %s\n", input)
		return "Stock not found", nil
	}
	result := fmt.Sprintf("Stock: %s (%s), Price: %.2f, Change: %.2f%%, Volume: %d",
		resp.Stock.Name, resp.Stock.Code, resp.Stock.CurrentPrice, resp.Stock.ChangePercent, resp.Stock.Volume)
	log.Printf("StockPriceTool success: %s\n", result)
	return result, nil
}

type StockAnalysisTool struct {
}

func NewStockAnalysisTool() *StockAnalysisTool {
	return &StockAnalysisTool{}
}

func (t *StockAnalysisTool) Name() string {
	return "StockAnalysis"
}

func (t *StockAnalysisTool) Description() string {
	return "Useful for getting advanced stock analysis data including Dragon Tiger List history, Chip Distribution, Order Book, Industry info, Northbound Funds, Guba Popularity, and Regulatory Notices. Input should be the stock code (e.g., 600519)."
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

	log.Printf("StockAnalysisTool called with input: [%s]\n", input)

	// Fetch data
	// 1. Industry
	industry, err := GetIndustryIndex(input)
	if err != nil {
		log.Printf("Error fetching industry: %v", err)
		industry = "Error fetching industry info"
	}

	// 2. Order Book
	orders, err := GetOrderBook(input)
	if err != nil {
		log.Printf("Error fetching order book: %v", err)
		orders = "Error fetching order book"
	}

	// 3. Chip Distribution
	chip, err := GetChipDistribution(input)
	if err != nil {
		log.Printf("Error fetching chip distribution: %v", err)
		chip = "Error fetching chip distribution"
	}

	// 4. Dragon Tiger History
	lhb, err := GetDragonTigerHistory(input, 5)
	if err != nil {
		log.Printf("Error fetching LHB history: %v", err)
	}

	// 5. Northbound Funds (Market Level)
	northbound, err := GetNorthboundFunds()
	if err != nil {
		log.Printf("Error fetching Northbound funds: %v", err)
		northbound = "Error fetching Northbound funds"
	}

	// 6. Stock Heat (Sentiment)
	heat, err := GetStockHeat(input)
	if err != nil {
		log.Printf("Error fetching stock heat: %v", err)
		heat = "Error fetching stock heat"
	}

	// 7. Regulatory Notices (Risk)
	notices, err := GetStockNotices(input, []string{"监管", "问询", "关注函", "立案", "警示"})
	if err != nil {
		log.Printf("Error fetching notices: %v", err)
	}

	// Format output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Analysis for %s:\n", input))

	sb.WriteString("\n[Industry Info]\n")
	sb.WriteString(industry)

	sb.WriteString("\n\n[Order Book (Intraday)]\n")
	sb.WriteString(orders)

	sb.WriteString("\n\n[Chip Distribution (Cost Structure)]\n")
	sb.WriteString(chip)

	sb.WriteString("\n\n[Market Sentiment & Funds]\n")
	sb.WriteString(northbound + "\n")
	sb.WriteString(heat)

	sb.WriteString("\n\n[Dragon Tiger List (Last 5)]\n")
	if len(lhb) > 0 {
		for _, l := range lhb {
			sb.WriteString(l + "\n")
		}
	} else {
		sb.WriteString("No recent records.\n")
	}

	sb.WriteString("\n\n[Regulatory Notices (Risk Alert)]\n")
	if len(notices) > 0 {
		for _, n := range notices {
			sb.WriteString(n + "\n")
		}
	} else {
		sb.WriteString("No recent regulatory notices.\n")
	}

	return sb.String(), nil
}
