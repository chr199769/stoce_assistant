package tool

import (
	"context"
	"fmt"
	"strings"
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
	// Fetch real news using EastMoney API
	news, err := GetStockNews(input)
	if err != nil {
		return fmt.Sprintf("Error fetching news: %v", err), nil
	}
	if len(news) == 0 {
		return "No recent news found for this stock.", nil
	}
	return strings.Join(news, "\n"), nil
}
