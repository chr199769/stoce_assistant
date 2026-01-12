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
