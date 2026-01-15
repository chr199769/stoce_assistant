package tool

import (
	"context"
	"fmt"
	"log"
	"strings"

	"stock_assistant/backend/ai_service/biz/tool/eastmoney"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
)

type SectorTool struct {
	Client          stockservice.Client
	EastMoneyClient *eastmoney.Client
}

func NewSectorTool(client stockservice.Client) *SectorTool {
	return &SectorTool{
		Client:          client,
		EastMoneyClient: eastmoney.NewClient(),
	}
}

func (t *SectorTool) Name() string {
	return "SectorAnalysis"
}

func (t *SectorTool) Description() string {
	return "Useful for getting market sector rankings and limit-up (sentiment) data. Input can be 'concept' (default), 'industry' or 'limit_up'."
}

func (t *SectorTool) Call(ctx context.Context, input string) (string, error) {
	input = strings.TrimSpace(strings.ToLower(input))
	
	if strings.Contains(input, "limit") || strings.Contains(input, "sentiment") || strings.Contains(input, "涨停") {
		return t.getLimitUpPool(ctx)
	}

	rankType := "concept"
	if strings.Contains(input, "industry") || strings.Contains(input, "行业") {
		rankType = "industry"
	}

	return t.getSectorRank(ctx, rankType)
}

func (t *SectorTool) GetSectorDetail(ctx context.Context, sectorCode string) (string, error) {
	// 1. Get Stocks Raw Data
	rawStocks, err := t.EastMoneyClient.GetSectorStocksRaw(ctx, sectorCode)
	if err != nil {
		return fmt.Sprintf("Error fetching sector stocks: %v", err), nil
	}

	// 2. Leader Selection Logic
	// Filter ST and New Stocks
	var candidates []*eastmoney.SectorStockItem
	for _, item := range rawStocks {
		if strings.Contains(item.Name, "ST") || strings.Contains(item.Name, "退") {
			continue
		}
		if strings.HasPrefix(item.Name, "N") || strings.HasPrefix(item.Name, "C") {
			continue
		}
		candidates = append(candidates, item)
	}

	// Calculate Score: 0.6*Amount + 0.4*MarketCap (Simplified proxy for influence)
	// Or better: Just sort by Amount (Turnover) which indicates liquidity and attention
	// Let's use Amount as primary factor for "Main Force" attention
	// And check for Limit Up status
	
	// Sort by Amount Desc
	// Bubble sort for simplicity (small list < 100) or just iterate to find max
	// Go sort needs import sort
	
	// We will just return the top 5 by Amount, ensuring mix of SH/SZ
	// Since we can't import sort easily without updating imports, let's implement simple selection
	
	leaders := t.selectLeaders(candidates)
	
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Sector Leaders (Top 5 by Influence):\n"))
	for _, l := range leaders {
		sb.WriteString(fmt.Sprintf("- %s (%s): %.2f%%, Amount: %.1f亿\n", 
			l.Name, l.Code, l.ChangePercent, l.Amount/100000000))
	}
	
	return sb.String(), nil
}

func (t *SectorTool) selectLeaders(candidates []*eastmoney.SectorStockItem) []*eastmoney.SectorStockItem {
	// Simple manual sort by Amount Desc
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].Amount > candidates[i].Amount {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	
	var leaders []*eastmoney.SectorStockItem
	shCount, szCount := 0, 0
	
	for _, item := range candidates {
		if len(leaders) >= 5 {
			break
		}
		
		isSH := strings.HasPrefix(item.Code, "6")
		isSZ := strings.HasPrefix(item.Code, "0") || strings.HasPrefix(item.Code, "3")
		
		canPick := true
		if isSH {
			neededSZ := 2 - szCount
			if neededSZ < 0 { neededSZ = 0 }
			if 5 - (len(leaders) + 1) < neededSZ { canPick = false }
		} else if isSZ {
			neededSH := 2 - shCount
			if neededSH < 0 { neededSH = 0 }
			if 5 - (len(leaders) + 1) < neededSH { canPick = false }
		} else {
			canPick = false
		}
		
		if canPick {
			leaders = append(leaders, item)
			if isSH { shCount++ }
			if isSZ { szCount++ }
		}
	}
	return leaders
}

func (t *SectorTool) getSectorRank(ctx context.Context, rankType string) (string, error) {
	req := &stock.GetMarketSectorsRequest{
		Type:  rankType,
		Limit: 10,
	}
	resp, err := t.Client.GetMarketSectors(ctx, req)
	if err != nil {
		log.Printf("SectorTool GetMarketSectors error: %v", err)
		return fmt.Sprintf("Error fetching sector data: %v", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Top 10 %s Sectors:\n", strings.Title(rankType)))
	for i, s := range resp.Sectors {
		sb.WriteString(fmt.Sprintf("%d. %s: %.2f%% (NetInflow: %.2f Wan), Top: %s\n", 
			i+1, s.Name, s.ChangePercent, s.NetInflow/10000, s.TopStockName))
	}
	return sb.String(), nil
}

func (t *SectorTool) getLimitUpPool(ctx context.Context) (string, error) {
	req := &stock.GetLimitUpPoolRequest{}
	resp, err := t.Client.GetLimitUpPool(ctx, req)
	if err != nil {
		log.Printf("SectorTool GetLimitUpPool error: %v", err)
		return fmt.Sprintf("Error fetching limit-up data: %v", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Limit-Up Pool Summary (Total: %d):\n", len(resp.Stocks)))
	
	// Group by height (LimitUpType)
	groups := make(map[string][]string)
	for _, s := range resp.Stocks {
		groups[s.LimitUpType] = append(groups[s.LimitUpType], fmt.Sprintf("%s(%s)", s.Name, s.Reason))
	}

	// Sort order: usually by board count desc, but map is unordered.
	// We just iterate common keys for display
	keys := []string{"5连板", "4连板", "3连板", "2连板", "首板"}
	for _, k := range keys {
		if stocks, ok := groups[k]; ok {
			sb.WriteString(fmt.Sprintf("[%s]: %s\n", k, strings.Join(stocks, ", ")))
			delete(groups, k)
		}
	}
	// Remaining
	for k, stocks := range groups {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", k, strings.Join(stocks, ", ")))
	}

	return sb.String(), nil
}
