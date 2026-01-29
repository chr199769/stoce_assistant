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
	session := graphdb.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	listAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
MATCH (s:Entity) WHERE s.type = 1
OPTIONAL MATCH (s)-[:HAS_EVENT]->(ev:Event {source:'financial'})
OPTIONAL MATCH (s)<-[:RELATION]-(fe:Entity) WHERE fe.type = 5
WITH s, CASE WHEN ev IS NOT NULL OR fe IS NOT NULL THEN 1 ELSE 0 END AS covered
WITH collect(CASE WHEN covered = 0 THEN s.id END) AS miss
RETURN miss
`, nil)
		if err != nil {
			return nil, err
		}
		if !result.Next(ctx) {
			return []string{}, result.Err()
		}
		val, _ := result.Record().Get("miss")
		arr, _ := val.([]any)
		out := make([]string, 0, len(arr))
		for _, v := range arr {
			if v != nil {
				out = append(out, toString(v))
			}
		}
		return out, nil
	})
	if err != nil {
		fmt.Println("获取缺失列表失败:", err)
		return
	}
	missing := listAny.([]string)
	if len(missing) == 0 {
		fmt.Println("无缺失项")
		return
	}
	limit := 500
	if len(missing) < limit {
		limit = len(missing)
	}
	em := eastmoney.NewClient()
	entities := make([]*knowledge_graph.Entity, 0, limit*6)
	relations := make([]*knowledge_graph.Relation, 0, limit*6)
	total := 0
	for i := 0; i < limit; i++ {
		code := missing[i]
		reps, err := em.GetFinancialReports(ctx, code)
		if err != nil || len(reps) == 0 {
			continue
		}
		n := 5
		if len(reps) < n {
			n = len(reps)
		}
		for j := 0; j < n; j++ {
			r := reps[j]
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
			stockEntity := &knowledge_graph.Entity{Id: code, Type: knowledge_graph.EntityType_STOCK, Attributes: map[string]string{"code": code}}
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
	fmt.Printf("补齐缺失财报完成：新增 %d 条事件实体与证据关系\n", total)
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

