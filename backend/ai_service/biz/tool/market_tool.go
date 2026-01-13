package tool

import (
	"context"
	"fmt"
	"strings"
)

type MarketInfoTool struct{}

func NewMarketInfoTool() *MarketInfoTool {
	return &MarketInfoTool{}
}

func (t *MarketInfoTool) Name() string {
	return "MarketInfo"
}

func (t *MarketInfoTool) Description() string {
	return "Fetches Dragon & Tiger List status, general market news, and specific influential news (Musk/Trump/Policy)."
}

func (t *MarketInfoTool) Call(ctx context.Context, input string) (string, error) {
	// Input is stock code, but this tool provides broader context + specific stock context
	stockCode := strings.TrimSpace(input)
	
	var sb strings.Builder

	// 1. Dragon & Tiger List
	sb.WriteString("=== Dragon & Tiger List (Longhu Bang) ===\n")
	if stockCode != "" {
		onList, details, err := GetDragonTigerStatus(stockCode)
		if err != nil {
			sb.WriteString(fmt.Sprintf("Error checking list: %v\n", err))
		} else if onList {
			sb.WriteString(fmt.Sprintf("YES. Details: %s\n", details))
		} else {
			sb.WriteString("No (Not on today's list)\n")
		}
	} else {
		sb.WriteString("No stock code provided for Longhu Bang check.\n")
	}
	sb.WriteString("\n")

	// 2. General Market News & Policy
	sb.WriteString("=== Market & Policy News ===\n")
	marketNews, err := GetMarketNews()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error fetching market news: %v\n", err))
	} else {
		// Filter for "Musk", "Trump", "Policy", "Industry"
		keywords := []string{"马斯克", "特朗普", "政策", "行业", "板块", "Musk", "Trump"}
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
			sb.WriteString("Found relevant influential news:\n")
			for _, n := range relevantNews {
				sb.WriteString(fmt.Sprintf("- %s\n", n))
			}
		} else {
			sb.WriteString("No specific mentions of Musk/Trump/Policy in top 20 flash news. Showing top 3 general news:\n")
			for i, n := range marketNews {
				if i >= 3 {
					break
				}
				sb.WriteString(fmt.Sprintf("- %s\n", n))
			}
		}
	}

	return sb.String(), nil
}
