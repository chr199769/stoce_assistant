package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/knowledge_graph/config"
	"stock_assistant/backend/knowledge_graph/dal/graphdb"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Println("初始化配置失败:", err)
		return
	}
	if err := graphdb.Init(); err != nil {
		fmt.Println("Neo4j 初始化失败:", err)
		return
	}
	ctx := context.Background()
	em := eastmoney.NewClient()
	// 从图谱读取现有股票列表，避免外部接口不稳定
	type stockItem struct{ Code, Name string }
	stocks := make([]stockItem, 0, 5000)
	{
		session := graphdb.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
		defer session.Close(ctx)
		itemsAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			result, err := tx.Run(ctx, `MATCH (s:Entity) WHERE s.type = 1 RETURN s.id AS code, s.name AS name`, nil)
			if err != nil {
				return nil, err
			}
			list := make([]stockItem, 0)
			for result.Next(ctx) {
				rec := result.Record()
				code, _ := rec.Get("code")
				name, _ := rec.Get("name")
				list = append(list, stockItem{Code: toString(code), Name: toString(name)})
			}
			return list, result.Err()
		})
		if err != nil {
			fmt.Println("读取图谱股票列表失败:", err)
			return
		}
		stocks = itemsAny.([]stockItem)
	}
	entities := make([]*knowledge_graph.Entity, 0, 5000)
	relations := make([]*knowledge_graph.Relation, 0, 10000)
	total := 0
	start := time.Now()
	for _, s := range stocks {
		if s.Code == "" {
			continue
		}
		reps, err := em.GetFinancialReports(ctx, s.Code)
		if err != nil || len(reps) == 0 {
			continue
		}
		limit := 5
		if len(reps) < limit {
			limit = len(reps)
		}
		code := normalizeCode(s.Code)
		for i := 0; i < limit; i++ {
			r := reps[i]
			if r == nil || strings.TrimSpace(r.ReportDate) == "" {
				continue
			}
			ts := parseDate(r.ReportDate)
			eventID := fmt.Sprintf("event:financial:%s:%s", code, r.ReportDate)
			evEntity := &knowledge_graph.Entity{
				Id:   eventID,
				Type: knowledge_graph.EntityType_EVENT,
				Name: "财报",
				Attributes: map[string]string{
					"code":          code,
					"report_date":   r.ReportDate,
					"total_revenue": fmt.Sprintf("%.2f", r.TotalRevenue),
					"net_profit":    fmt.Sprintf("%.2f", r.NetProfit),
					"eps":           fmt.Sprintf("%.2f", r.Eps),
					"revenue_yoy":   fmt.Sprintf("%.2f", r.RevenueYoy),
					"profit_yoy":    fmt.Sprintf("%.2f", r.ProfitYoy),
				},
			}
			entities = append(entities, evEntity)
			stockEntity := &knowledge_graph.Entity{Id: code, Type: knowledge_graph.EntityType_STOCK, Name: s.Name, Attributes: map[string]string{"code": code}}
			entities = append(entities, stockEntity)
			summary := fmt.Sprintf("营收 %.2f, 净利 %.2f, EPS %.2f, 营收同比 %.2f%%, 净利同比 %.2f%%", r.TotalRevenue, r.NetProfit, r.Eps, r.RevenueYoy, r.ProfitYoy)
			evidence := &knowledge_graph.Evidence{Id: "evd_fin_" + eventID, Category: "financial", Summary: summary, Confidence: 100, Source: "financial", Timestamp: ts}
			relations = append(relations, &knowledge_graph.Relation{
				Source:    evEntity,
				Target:    stockEntity,
				Type:      knowledge_graph.RelationType_AFFECTS,
				Strength:  1,
				Evidence:  evidence,
				UpdatedAt: ts,
			})
			total++
		}
		// flush periodically
		if len(entities) >= 4000 || len(relations) >= 8000 {
			_, _ = graphdb.UpsertEntities(ctx, entities)
			_, _ = graphdb.UpsertRelations(ctx, relations)
			entities = entities[:0]
			relations = relations[:0]
		}
	}
	if len(entities) > 0 {
		_, _ = graphdb.UpsertEntities(ctx, entities)
	}
	if len(relations) > 0 {
		_, _ = graphdb.UpsertRelations(ctx, relations)
	}
	elapsed := time.Since(start)
	fmt.Printf("已补齐财报数据：新增事件实体与证据关系 %d 条，用时 %s\n", total, elapsed.String())
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func normalizeCode(raw string) string {
	code := strings.ToLower(strings.TrimSpace(raw))
	code = strings.TrimPrefix(code, "sh")
	code = strings.TrimPrefix(code, "sz")
	code = strings.TrimPrefix(code, "bj")
	if len(code) != 6 {
		return ""
	}
	if code[0] == '6' {
		return "sh" + code
	}
	if code[0] == '0' || code[0] == '3' {
		return "sz" + code
	}
	if code[0] == '8' {
		return "bj" + code
	}
	return ""
}

func parseDate(value string) int64 {
	if value == "" {
		return time.Now().Unix()
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	t, err := time.ParseInLocation("2006-01-02", value, loc)
	if err != nil {
		return time.Now().Unix()
	}
	return t.Unix()
}
