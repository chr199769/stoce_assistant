package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	eastmoney "stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/ai_service/biz/tool/fractal"
)

type SectorCorrelation struct {
	SectorCode      string  `json:"sector_code"`
	SectorName      string  `json:"sector_name"`
	Correlation     float64 `json:"correlation"`
	ConstituentUsed int     `json:"constituent_used"`
	DaysUsed        int     `json:"days_used"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("用法: sector_correlation <stock_code> [days=60] [topK=5] [maxConstituents=10]")
	}
	stockCode := strings.TrimSpace(os.Args[1])
	days := 60
	topK := 5
	maxConstituents := 10
	if len(os.Args) >= 3 {
		fmt.Sscanf(os.Args[2], "%d", &days)
	}
	if len(os.Args) >= 4 {
		fmt.Sscanf(os.Args[3], "%d", &topK)
	}
	if len(os.Args) >= 5 {
		fmt.Sscanf(os.Args[4], "%d", &maxConstituents)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := eastmoney.NewClient()

	klines, err := client.GetKlineHistory(ctx, stockCode, days)
	if err != nil || len(klines) == 0 {
		log.Fatalf("获取股票K线失败: code=%s days=%d err=%v", stockCode, days, err)
	}

	stockChange := make(map[string]float64)
	var dateOrder []string
	for _, k := range klines {
		stockChange[k.Date] = k.ChangePercent
		dateOrder = append(dateOrder, k.Date)
	}

	idxInfo, err := client.GetIndustryIndex(ctx, stockCode)
	if err != nil || idxInfo == nil {
		log.Fatalf("获取行业/概念标签失败: code=%s err=%v", stockCode, err)
	}

	concepts := splitNames(idxInfo.ConceptNames)
	industry := strings.TrimSpace(idxInfo.IndustryName)

	industryList, _ := client.GetSectorRank(ctx, "industry", 500)
	conceptList, _ := client.GetSectorRank(ctx, "concept", 500)

	nameToSector := make(map[string]*eastmoney.SectorInfo)
	for _, s := range industryList {
		nameToSector[normalize(s.Name)] = s
	}
	for _, s := range conceptList {
		nameToSector[normalize(s.Name)] = s
	}

	var candidates []*eastmoney.SectorInfo
	if industry != "" {
		if sec, ok := nameToSector[normalize(industry)]; ok {
			candidates = append(candidates, sec)
		} else {
			candidates = append(candidates, matchSectorsByName(industry, industryList)...)
		}
	}
	for _, c := range concepts {
		if sec, ok := nameToSector[normalize(c)]; ok {
			candidates = append(candidates, sec)
		} else {
			candidates = append(candidates, matchSectorsByName(c, conceptList)...)
		}
	}
	if len(candidates) == 0 {
		log.Printf("行业板块数量: %d 概念板块数量: %d", len(industryList), len(conceptList))
		printSample("行业示例", industryList, 20)
		printSample("概念示例", conceptList, 20)
		log.Fatalf("未找到匹配的板块候选: 行业=%s 概念=%v", industry, concepts)
	}

	var results []SectorCorrelation
	for _, sec := range candidates {
		usedCount, sectorSeries := buildSectorSeries(ctx, client, sec.Code, dateOrder, maxConstituents, days)
		if usedCount == 0 || len(sectorSeries) == 0 {
			continue
		}
		var x, y []float64
		for i := range sectorSeries {
			d := dateOrder[i]
			y = append(y, sectorSeries[i])
			x = append(x, stockChange[d])
		}
		corr := fractal.CalculatePearsonCorrelation(x, y)
		results = append(results, SectorCorrelation{
			SectorCode:      sec.Code,
			SectorName:      sec.Name,
			Correlation:     corr,
			ConstituentUsed: usedCount,
			DaysUsed:        len(x),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Correlation > results[j].Correlation
	})
	if len(results) > topK {
		results = results[:topK]
	}

	out, _ := json.MarshalIndent(map[string]interface{}{
		"stock_code": stockCode,
		"days":       days,
		"top_k":      topK,
		"results":    results,
	}, "", "  ")
	fmt.Println(string(out))
}

func splitNames(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ReplaceAll(s, "，", ",")
	parts := strings.Split(s, ",")
	var res []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func matchSectorsByName(name string, list []*eastmoney.SectorInfo) []*eastmoney.SectorInfo {
	n := normalize(name)
	var res []*eastmoney.SectorInfo
	for _, s := range list {
		sn := normalize(s.Name)
		if strings.Contains(sn, n) || strings.Contains(n, sn) {
			res = append(res, s)
		}
	}
	return res
}

func printSample(title string, list []*eastmoney.SectorInfo, n int) {
	if n > len(list) {
		n = len(list)
	}
	var names []string
	for i := 0; i < n; i++ {
		names = append(names, list[i].Name)
	}
	log.Printf("%s: %v", title, names)
}

func buildSectorSeries(ctx context.Context, client *eastmoney.Client, sectorCode string, dateOrder []string, maxConstituents int, days int) (int, []float64) {
	stocks, err := client.GetSectorStocksRaw(ctx, sectorCode)
	if err != nil || len(stocks) == 0 {
		return 0, nil
	}
	if len(stocks) > maxConstituents {
		stocks = stocks[:maxConstituents]
	}
	dateIndex := make(map[string]int)
	for i, d := range dateOrder {
		dateIndex[d] = i
	}
	sum := make([]float64, len(dateOrder))
	count := make([]int, len(dateOrder))
	used := 0
	for _, s := range stocks {
		kl, err := client.GetKlineHistory(ctx, s.Code, days)
		if err != nil || len(kl) == 0 {
			continue
		}
		used++
		for _, k := range kl {
			if idx, ok := dateIndex[k.Date]; ok {
				sum[idx] += k.ChangePercent
				count[idx]++
			}
		}
	}
	if used == 0 {
		return 0, nil
	}
	series := make([]float64, len(dateOrder))
	effective := 0
	for i := range dateOrder {
		if count[i] > 0 {
			series[i] = sum[i] / float64(count[i])
			effective++
		} else {
			series[i] = 0
		}
	}
	if effective == 0 {
		return 0, nil
	}
	return used, series
}
