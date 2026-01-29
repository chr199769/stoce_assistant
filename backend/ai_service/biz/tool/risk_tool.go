package tool

import (
	"context"
	"fmt"
	"strings"

	eastmoney "stock_assistant/backend/common/eastmoney"
	eimpl "stock_assistant/backend/common/provider/impl/eastmoney"
)

func CheckRiskControlRules(ctx context.Context, code string) string {
	client := eimpl.NewMarket(eastmoney.NewClient())

	var benchmarkSecId string
	var boardName string

	// Clean code
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	if strings.HasPrefix(cleanCode, "688") {
		benchmarkSecId = "SH000688" // 科创50
		boardName = "科创板"
	} else if strings.HasPrefix(cleanCode, "6") {
		benchmarkSecId = "SH000001" // 上证指数
		boardName = "沪市主板"
	} else if strings.HasPrefix(cleanCode, "3") {
		benchmarkSecId = "SZ399006" // 创业板指
		boardName = "创业板"
	} else if strings.HasPrefix(cleanCode, "8") || strings.HasPrefix(cleanCode, "4") || strings.HasPrefix(cleanCode, "9") {
		benchmarkSecId = "SZ899050" // 北证50
		boardName = "北交所"
	} else {
		benchmarkSecId = "SZ399107" // 深证A指
		boardName = "深市主板"
	}

	// Fetch data (60 days for stock to ensure enough history for 30-day calc, 60 for benchmark)
	// We need index 30 (31st data point) for 30-day deviation check
	stockCode := normalizeProviderCode(cleanCode)
	stockK, err1 := client.GetKlines(ctx, stockCode, "day", 60)
	benchK, err2 := client.GetKlines(ctx, benchmarkSecId, "day", 60)

	if err1 != nil || err2 != nil {
		return fmt.Sprintf("无法进行量化风控检查: 数据获取失败 (StockErr: %v, BenchErr: %v)", err1, err2)
	}

	// Map benchmark data by date (Close Price for interval calc)
	benchCloseMap := make(map[string]float64)
	for _, it := range benchK {
		benchCloseMap[it.Time.Format("2006-01-02")] = it.Close
	}

	// Filter valid trading days from stockK (reverse order: latest to oldest)
	var validStockK []*eastmoney.KlineItem // 用于索引，但数据来源为 provider；仅取必要字段
	for i := len(stockK) - 1; i >= 0; i-- {
		it := stockK[i]
		if it.Volume > 0 {
			// 构造最小结构以复用下方逻辑
			validStockK = append(validStockK, &eastmoney.KlineItem{
				Date:          it.Time.Format("2006-01-02"),
				Open:          it.Open,
				Close:         it.Close,
				High:          it.High,
				Low:           it.Low,
				Volume:        it.Volume,
				ChangePercent: it.ChangePct,
			})
		}
	}

	// Helper to calculate interval deviation
	calcDeviation := func(days int) float64 {
		if len(validStockK) < days {
			return 0
		}

		// End Date Data (Latest)
		endClose := validStockK[0].Close
		endDate := validStockK[0].Date

		// Start Date Data (The day BEFORE the interval starts)
		// For 10 days interval, we need the closing price of the 11th day back as base
		actualDays := days
		if len(validStockK) <= days {
			actualDays = len(validStockK) - 1
		}

		startClose := validStockK[actualDays].Close
		baseDate := validStockK[actualDays].Date

		// Get Interval Start Date (T-(days-1)) for display clarity
		// This is the first day INCLUDED in the interval
		// intervalStartDate 仅用于展示，计算不需要

		// Calculate Stock Interval Pct
		stockPct := (endClose - startClose) / startClose * 100

		// Calculate Benchmark Interval Pct
		benchEndClose, ok1 := benchCloseMap[endDate]
		benchStartClose, ok2 := benchCloseMap[baseDate]

		benchPct := 0.0
		if !ok1 || !ok2 {
		} else {
			benchPct = (benchEndClose - benchStartClose) / benchStartClose * 100
		}

		deviation := stockPct - benchPct

		return deviation
	}

	dev3 := calcDeviation(3)
	dev10 := calcDeviation(10)
	dev30 := calcDeviation(30)

	var risks []string

	// BSE Rule: 3 days ±40%
	if boardName == "北交所" {
		if dev3 >= 40 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警（北交所）：近3日累计涨幅偏离值达 %.2f%%（阈值 40%%）", dev3))
		} else if dev3 <= -40 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警（北交所）：近3日累计跌幅偏离值达 %.2f%%（阈值 -40%%）", dev3))
		}
	} else {
		// Main Boards
		if dev10 >= 100 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警：10日累计涨幅偏离值达 %.2f%%（阈值 100%%）", dev10))
		} else if dev10 <= -50 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警：10日累计跌幅偏离值达 %.2f%%（阈值 -50%%）", dev10))
		}

		if dev30 >= 200 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警：30日累计涨幅偏离值达 %.2f%%（阈值 200%%）", dev30))
		} else if dev30 <= -70 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警：30日累计跌幅偏离值达 %.2f%%（阈值 -70%%）", dev30))
		}
	}

	if len(risks) > 0 {
		return strings.Join(risks, "\n")
	}
	return fmt.Sprintf("✅ 偏离值检查通过（%s）：10日偏离 %.2f%%，30日偏离 %.2f%%", boardName, dev10, dev30)
}

func normalizeProviderCode(s string) string {
	u := strings.ToUpper(strings.TrimSpace(s))
	if len(u) == 6 && u[0] == '6' {
		return "SH" + u
	}
	if len(u) == 6 {
		return "SZ" + u
	}
	if strings.HasPrefix(u, "SH") || strings.HasPrefix(u, "SZ") {
		return u
	}
	return u
}
