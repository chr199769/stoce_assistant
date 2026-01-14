package tool

import (
	"context"
	"fmt"
	"log"
	"strings"

	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
)

type SectorTool struct {
	Client stockservice.Client
}

func NewSectorTool(client stockservice.Client) *SectorTool {
	return &SectorTool{Client: client}
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
