package tool

import (
	"context"
	"fmt"
	"log"
	"strings"

	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
	eastmoney "stock_assistant/backend/common/eastmoney"
	eimpl "stock_assistant/backend/common/provider/impl/eastmoney"
	"stock_assistant/backend/common/provider"
)

type SectorTool struct {
	Client          stockservice.Client
	SectorClient    provider.SectorClient
}

func NewSectorTool(client stockservice.Client) *SectorTool {
	return &SectorTool{
		Client:          client,
		SectorClient:    eimpl.NewSector(eastmoney.NewClient()),
	}
}

func (t *SectorTool) Name() string {
	return "SectorAnalysis"
}

func (t *SectorTool) Description() string {
	return "用于获取市场板块排行和涨停（情绪）数据。输入可以是 'concept'（默认）、'industry' 或 'limit_up'。"
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
	rawStocks, err := t.SectorClient.GetSectorStocks(ctx, sectorCode)
	if err != nil {
		return fmt.Sprintf("获取板块股票失败: %v", err), nil
	}

	var candidates []*eastmoney.SectorStockItem
	for _, item := range rawStocks {
		if strings.Contains(item.Name, "ST") || strings.Contains(item.Name, "退") {
			continue
		}
		if strings.HasPrefix(item.Name, "N") || strings.HasPrefix(item.Name, "C") {
			continue
		}
		candidates = append(candidates, &eastmoney.SectorStockItem{
			Code: item.Code,
			Name: item.Name,
		})
	}

	leaders := t.selectLeaders(candidates)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("板块龙头 (影响力前 5):\n"))
	for _, l := range leaders {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", l.Name, l.Code))
	}

	return sb.String(), nil
}

func (t *SectorTool) selectLeaders(candidates []*eastmoney.SectorStockItem) []*eastmoney.SectorStockItem {
	// 简单的按成交额降序排序
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
			if neededSZ < 0 {
				neededSZ = 0
			}
			if 5-(len(leaders)+1) < neededSZ {
				canPick = false
			}
		} else if isSZ {
			neededSH := 2 - shCount
			if neededSH < 0 {
				neededSH = 0
			}
			if 5-(len(leaders)+1) < neededSH {
				canPick = false
			}
		} else {
			canPick = false
		}

		if canPick {
			leaders = append(leaders, item)
			if isSH {
				shCount++
			}
			if isSZ {
				szCount++
			}
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
		log.Printf("SectorTool 获取板块数据错误: %v", err)
		return fmt.Sprintf("获取板块数据失败: %v", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("前 10 %s 板块:\n", strings.Title(rankType)))
	for i, s := range resp.Sectors {
		sb.WriteString(fmt.Sprintf("%d. %s: %.2f%% (净流入: %.2f 万), 领涨股: %s\n",
			i+1, s.Name, s.ChangePercent, s.NetInflow/10000, s.TopStockName))
	}
	limitUp, err := t.getLimitUpPool(ctx)
	if err == nil && strings.TrimSpace(limitUp) != "" {
		sb.WriteString("\n[市场情绪·涨停池]\n")
		sb.WriteString(strings.TrimSpace(limitUp))
	}
	return sb.String(), nil
}

func (t *SectorTool) getLimitUpPool(ctx context.Context) (string, error) {
	req := &stock.GetLimitUpPoolRequest{}
	resp, err := t.Client.GetLimitUpPool(ctx, req)
	if err != nil {
		log.Printf("SectorTool 获取涨停池错误: %v", err)
		return fmt.Sprintf("获取涨停数据失败: %v", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("涨停池摘要 (总计: %d):\n", len(resp.Stocks)))

	// 按高度分组 (LimitUpType)
	groups := make(map[string][]string)
	for _, s := range resp.Stocks {
		groups[s.LimitUpType] = append(groups[s.LimitUpType], fmt.Sprintf("%s(%s)", s.Name, s.Reason))
	}

	// 排序顺序: 通常按连板数降序，但 map 是无序的。
	// 我们只需遍历常见键进行显示
	keys := []string{"5连板", "4连板", "3连板", "2连板", "首板"}
	for _, k := range keys {
		if stocks, ok := groups[k]; ok {
			sb.WriteString(fmt.Sprintf("[%s]: %s\n", k, strings.Join(stocks, ", ")))
			delete(groups, k)
		}
	}
	// 剩余的
	for k, stocks := range groups {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", k, strings.Join(stocks, ", ")))
	}

	return sb.String(), nil
}
