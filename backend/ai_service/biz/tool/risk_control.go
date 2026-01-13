package tool

import (
	"fmt"
	"strconv"
	"strings"
)

// CheckRiskControlRules checks for severe abnormal fluctuations
func CheckRiskControlRules(code string) string {
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

	// Fetch data (30 days for stock, 60 for benchmark to ensure coverage)
	stockK, err1 := GetKLineData(cleanCode, 30)
	benchK, err2 := GetKLineData(benchmarkSecId, 60)

	if err1 != nil || err2 != nil {
		return fmt.Sprintf("无法进行量化风控检查: 数据获取失败 (StockErr: %v, BenchErr: %v)", err1, err2)
	}

	// Map benchmark data by date
	benchMap := make(map[string]float64)
	for _, line := range benchK {
		parts := strings.Split(line, ",")
		if len(parts) > 2 {
			val, _ := strconv.ParseFloat(parts[2], 64)
			benchMap[parts[0]] = val
		}
	}

	var dev10, dev30, dev3 float64
	
	// Iterate stock data from latest to oldest
	// stockK is typically ascending by date
	count := 0
	for i := len(stockK) - 1; i >= 0; i-- {
		parts := strings.Split(stockK[i], ",")
		if len(parts) <= 2 {
			continue
		}
		date := parts[0]
		sPct, _ := strconv.ParseFloat(parts[2], 64)
		
		bPct, ok := benchMap[date]
		if !ok {
			// If benchmark missing, maybe assume 0 or log?
			// Usually shouldn't happen for trading days.
			bPct = 0
		}
		
		dailyDev := sPct - bPct

		if count < 3 {
			dev3 += dailyDev
		}
		if count < 10 {
			dev10 += dailyDev
		}
		if count < 30 {
			dev30 += dailyDev
		}
		count++
	}

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
