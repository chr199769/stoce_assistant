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
		news, err := GetStockNews(stockCode)
		if err != nil {
			sb.WriteString(fmt.Sprintf("Error fetching news: %v\n", err))
		} else if len(news) == 0 {
			sb.WriteString("No recent news found.\n")
		} else {
			for _, n := range news {
				sb.WriteString(fmt.Sprintf("- %s\n", n))
			}
		}
		
		// B. Dragon & Tiger List
		sb.WriteString("\n[Dragon & Tiger List Status]\n")
		onList, details, err := GetDragonTigerStatus(stockCode)
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
