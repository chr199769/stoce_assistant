package tool

import (
	"context"
	"fmt"
	"strings"

	eastmoney "stock_assistant/backend/common/eastmoney"
)

// CheckRiskControlRules checks for severe abnormal fluctuations
func CheckRiskControlRules(ctx context.Context, client *eastmoney.Client, code string) string {
	var benchmarkSecId string
	var boardName string

	// Clean code
	cleanCode := code
	if len(code) > 6 {
		cleanCode = code[len(code)-6:]
	}

	if strings.HasPrefix(cleanCode, "688") {
		benchmarkSecId = "1.000688" // STAR 50
		boardName = "科创板"
	} else if strings.HasPrefix(cleanCode, "6") {
		benchmarkSecId = "1.000001" // SSE Composite
		boardName = "沪市主板"
	} else if strings.HasPrefix(cleanCode, "3") {
		benchmarkSecId = "0.399006" // ChiNext
		boardName = "创业板"
	} else if strings.HasPrefix(cleanCode, "8") || strings.HasPrefix(cleanCode, "4") || strings.HasPrefix(cleanCode, "9") {
		benchmarkSecId = "0.899050" // BSE 50 (Best guess)
		boardName = "北交所"
	} else {
		benchmarkSecId = "0.399107" // SZSE A Share
		boardName = "深市主板"
	}

	// Fetch data (60 days for stock to ensure enough history for 30-day calc, 60 for benchmark)
	// We need index 30 (31st data point) for 30-day deviation check
	stockK, err1 := client.GetKlineHistory(ctx, cleanCode, 60)
	benchK, err2 := client.GetKlineHistory(ctx, benchmarkSecId, 60)

	if err1 != nil || err2 != nil {
		return fmt.Sprintf("无法进行量化风控检查: 数据获取失败 (StockErr: %v, BenchErr: %v)", err1, err2)
	}

	// Map benchmark data by date (Close Price for interval calc)
	benchCloseMap := make(map[string]float64)
	for _, item := range benchK {
		benchCloseMap[item.Date] = item.Close
	}

	// Filter valid trading days from stockK (reverse order: latest to oldest)
	var validStockK []*eastmoney.KlineItem
	for i := len(stockK) - 1; i >= 0; i-- {
		item := stockK[i]
		// Check for non-trading days (Volume = 0)
		if item.Volume > 0 {
			validStockK = append(validStockK, item)
		}
	}

	// Helper to calculate interval deviation
	calcDeviation := func(days int) float64 {
		if len(validStockK) < days {
			fmt.Printf("[RiskControl] Not enough data for %d days check. Available: %d\n", days, len(validStockK))
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
			fmt.Printf("[RiskControl] Warning: Data length (%d) <= days (%d), using oldest available day index %d\n", len(validStockK), days, actualDays)
		}

		startClose := validStockK[actualDays].Close
		baseDate := validStockK[actualDays].Date

		// Get Interval Start Date (T-(days-1)) for display clarity
		// This is the first day INCLUDED in the interval
		intervalStartDate := "N/A"
		if actualDays > 0 {
			intervalStartDate = validStockK[actualDays-1].Date
		}

		// Calculate Stock Interval Pct
		stockPct := (endClose - startClose) / startClose * 100

		// Calculate Benchmark Interval Pct
		benchEndClose, ok1 := benchCloseMap[endDate]
		benchStartClose, ok2 := benchCloseMap[baseDate]

		benchPct := 0.0
		if !ok1 || !ok2 {
			fmt.Printf("[RiskControl] Benchmark data missing for dates: Base(%s)=%v, End(%s)=%v. Assuming 0%% benchmark change.\n", baseDate, ok2, endDate, ok1)
		} else {
			benchPct = (benchEndClose - benchStartClose) / benchStartClose * 100
		}

		deviation := stockPct - benchPct

		fmt.Printf("[RiskControl] Interval: %d Days | Range: [%s, %s] (Base: %s)\n", days, intervalStartDate, endDate, baseDate)
		fmt.Printf("    Stock: %.2f -> %.2f (%.2f%%)\n", startClose, endClose, stockPct)
		if ok1 && ok2 {
			fmt.Printf("    Bench: %.2f -> %.2f (%.2f%%)\n", benchStartClose, benchEndClose, benchPct)
		}
		fmt.Printf("    Deviation: %.2f%%\n", deviation)

		return deviation
	}

	dev3 := calcDeviation(3)
	dev10 := calcDeviation(10)
	dev30 := calcDeviation(30)

	var risks []string

	// BSE Rule: 3 days ±40%
	if boardName == "北交所" {
		if dev3 >= 40 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警 (北交所): 近3日累计涨幅偏离值达 %.2f%% (阈值 40%%)", dev3))
		} else if dev3 <= -40 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警 (北交所): 近3日累计跌幅偏离值达 %.2f%% (阈值 -40%%)", dev3))
		}
	} else {
		// Main Boards
		if dev10 >= 100 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警: 10日累计涨幅偏离值达 %.2f%% (阈值 100%%)", dev10))
		} else if dev10 <= -50 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警: 10日累计跌幅偏离值达 %.2f%% (阈值 -50%%)", dev10))
		}

		if dev30 >= 200 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警: 30日累计涨幅偏离值达 %.2f%% (阈值 200%%)", dev30))
		} else if dev30 <= -70 {
			risks = append(risks, fmt.Sprintf("⚠️ 严重异动预警: 30日累计跌幅偏离值达 %.2f%% (阈值 -70%%)", dev30))
		}
	}

	if len(risks) > 0 {
		return strings.Join(risks, "\n")
	}
	return fmt.Sprintf("✅ 偏离值检查通过 (%s): 10日偏离 %.2f%%, 30日偏离 %.2f%%", boardName, dev10, dev30)
}
