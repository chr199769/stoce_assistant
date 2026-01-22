package tool

import (
	"context"
	"fmt"
	"log"
	"strings"

	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"
	eastmoney "stock_assistant/backend/common/eastmoney"
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
	// 1. 获取股票原始数据
	rawStocks, err := t.EastMoneyClient.GetSectorStocksRaw(ctx, sectorCode)
	if err != nil {
		return fmt.Sprintf("获取板块股票失败: %v", err), nil
	}

	// 2. 龙头选择逻辑
	// 过滤 ST 和新股
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

	// 计算得分: 0.6*成交额 + 0.4*市值 (作为影响力的简化代理)
	// 或者更好: 直接按成交额排序，这表示流动性和关注度
	// 让我们使用成交额作为“主力”关注的主要因素
	// 并检查涨停状态

	// 按成交额降序排序
	// 为了简单起见使用冒泡排序（列表很小 < 100），或者直接遍历查找最大值
	// Go 的 sort 需要引入 sort 包

	// 我们将只返回成交额前 5 名，确保沪深两市混合
	// 由于不想轻易引入 sort 包而更新 imports，我们实现简单的选择逻辑

	leaders := t.selectLeaders(candidates)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("板块龙头 (影响力前 5):\n"))
	for _, l := range leaders {
		sb.WriteString(fmt.Sprintf("- %s (%s): %.2f%%, 成交额: %.1f亿\n",
			l.Name, l.Code, l.ChangePercent, l.Amount/100000000))
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
