package tool

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	eimpl "stock_assistant/backend/common/provider/impl/eastmoney"
	"stock_assistant/backend/common/provider"
	eastmoney "stock_assistant/backend/common/eastmoney"
)

type DragonTigerTool struct {
	Client provider.DragonTigerClient
}

func NewDragonTigerTool() *DragonTigerTool {
	return &DragonTigerTool{
		Client: eimpl.NewDragonTiger(eastmoney.NewClient()),
	}
}

func (t *DragonTigerTool) Name() string {
	return "DragonTigerList"
}

func (t *DragonTigerTool) Description() string {
	return "Get daily Dragon Tiger List (Longhu Bang) data. Input can be a date (YYYY-MM-DD) or empty for today."
}

func (t *DragonTigerTool) Call(ctx context.Context, input string) (string, error) {
	date := strings.TrimSpace(input)
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	items, err := t.Client.GetTodayList(ctx, date)
	if err != nil {
		return fmt.Sprintf("Error fetching Dragon Tiger List: %v", err), nil
	}

	if len(items) == 0 {
		return "No Dragon Tiger List data found for this date.", nil
	}

	seatMap := map[string]string{
		"华泰证券股份有限公司北京雍和宫证券营业部":       "赵老哥",
		"国泰君安证券股份有限公司上海江苏路证券营业部":     "章盟主",
		"中国银河证券股份有限公司北京绍兴路证券营业部":     "赵老哥",
		"东方财富证券股份有限公司拉萨团结路第二证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨团结路第一证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第二证券营业部":   "拉萨天团",
		"东方财富证券股份有限公司拉萨东环路第一证券营业部":   "拉萨天团",
		"招商证券股份有限公司深圳益田路免税商务大厦证券营业部": "益田路",
		"中信证券股份有限公司上海溧阳路证券营业部":       "孙哥",
	}

	sort.Slice(items, func(i, j int) bool { return items[i].NetInflow > items[j].NetInflow })

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Dragon Tiger List (%s) - Top 5 Net Buy:\n", date))

	for i := 0; i < 5 && i < len(items); i++ {
		item := items[i]
		sb.WriteString(fmt.Sprintf("\n%d. %s (%s) | Change: %.2f%% | Net: %.1f万\n",
			i+1, item.Name, item.Code, item.ChangePercent, item.NetInflow/10000))
		sb.WriteString(fmt.Sprintf("   Reason: %s\n", item.Reason))

		if i < 3 {
			seats, err := t.Client.GetSeats(ctx, date, item.Code)
			if err == nil {
				sb.WriteString("   [Top Buyer]:\n")
				buySeats := filterSeats(seats, true)
				for k, seat := range buySeats {
					if k >= 3 {
						break
					}
					tag := ""
					if t, ok := seatMap[seat.SeatName]; ok {
						tag = fmt.Sprintf("[%s]", t)
					} else if strings.Contains(seat.SeatName, "拉萨") {
						tag = "[拉萨天团]"
					} else if strings.Contains(seat.SeatName, "机构专用") {
						tag = "[机构]"
					} else if strings.Contains(seat.SeatName, "沪股通") || strings.Contains(seat.SeatName, "深股通") {
						tag = "[北向]"
					}
					sb.WriteString(fmt.Sprintf("     - %s %s: %.0f万\n", seat.SeatName, tag, seat.NetAmount/10000))
				}

				sb.WriteString("   [Top Seller]:\n")
				sellSeats := filterSeats(seats, false)
				for k, seat := range sellSeats {
					if k >= 3 {
						break
					}
					tag := ""
					if t, ok := seatMap[seat.SeatName]; ok {
						tag = fmt.Sprintf("[%s]", t)
					} else if strings.Contains(seat.SeatName, "拉萨") {
						tag = "[拉萨天团]"
					} else if strings.Contains(seat.SeatName, "机构专用") {
						tag = "[机构]"
					} else if strings.Contains(seat.SeatName, "沪股通") || strings.Contains(seat.SeatName, "深股通") {
						tag = "[北向]"
					}
					sb.WriteString(fmt.Sprintf("     - %s %s: %.0f万\n", seat.SeatName, tag, seat.NetAmount/10000))
				}
			}
		}
	}

	return sb.String(), nil
}

func filterSeats(seats []*provider.SeatRecord, buyers bool) []*provider.SeatRecord {
	var res []*provider.SeatRecord
	for _, s := range seats {
		if buyers && s.NetAmount >= 0 {
			res = append(res, s)
		}
		if !buyers && s.NetAmount < 0 {
			res = append(res, s)
		}
	}
	return res
}
