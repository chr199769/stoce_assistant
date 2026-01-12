package tool

import (
	"context"
)

type NewsTool struct{}

func NewNewsTool() *NewsTool {
	return &NewsTool{}
}

func (t *NewsTool) Name() string {
	return "StockNews"
}

func (t *NewsTool) Description() string {
	return "Useful for getting recent news about a stock. Input should be the stock code."
}

func (t *NewsTool) Call(ctx context.Context, input string) (string, error) {
	// Mock news
	return "1. Recent financial report shows 20% growth.\n2. New product launch announced next month.\n3. Industry sector is recovering.", nil
}
