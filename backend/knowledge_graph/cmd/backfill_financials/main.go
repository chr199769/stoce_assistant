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

	recordsAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
MATCH (s:Entity)-[:HAS_EVENT]->(ev:Event)
WHERE s.type = 1 AND ev.source = 'financial'
RETURN s.id AS stock, ev.id AS event_id
`, nil)
		if err != nil {
			return nil, err
		}
		pairs := make([]struct {
			Stock   string
			EventID string
		}, 0)
		for result.Next(ctx) {
			rec := result.Record()
			stock, _ := rec.Get("stock")
			eid, _ := rec.Get("event_id")
			pairs = append(pairs, struct {
				Stock   string
				EventID string
			}{
				Stock:   toString(stock),
				EventID: toString(eid),
			})
		}
		return pairs, result.Err()
	})
	if err != nil {
		fmt.Println("读取历史财报事件失败:", err)
		return
	}
	pairs := recordsAny.([]struct {
		Stock   string
		EventID string
	})
	em := eastmoney.NewClient()
	entities := make([]*knowledge_graph.Entity, 0, len(pairs))
	relations := make([]*knowledge_graph.Relation, 0, len(pairs))
	added := 0
	for _, p := range pairs {
		date := parseReportDateFromEventID(p.EventID)
		code := p.Stock
		reports, err := em.GetFinancialReports(ctx, code)
		if err != nil || len(reports) == 0 {
			continue
		}
		var match *eastmoney.FinancialData
		for _, r := range reports {
			if r != nil && strings.TrimSpace(r.ReportDate) == date {
				match = r
				break
			}
		}
		if match == nil {
			continue
		}
		eventID := "event:financial:" + code + ":" + date
		evEntity := &knowledge_graph.Entity{
			Id:   eventID,
			Type: knowledge_graph.EntityType_EVENT,
			Name: "财报",
			Attributes: map[string]string{
				"code":          code,
				"report_date":   date,
				"total_revenue": fmt.Sprintf("%.2f", match.TotalRevenue),
				"net_profit":    fmt.Sprintf("%.2f", match.NetProfit),
				"eps":           fmt.Sprintf("%.2f", match.Eps),
				"revenue_yoy":   fmt.Sprintf("%.2f", match.RevenueYoy),
				"profit_yoy":    fmt.Sprintf("%.2f", match.ProfitYoy),
			},
		}
		entities = append(entities, evEntity)
		stockEntity := &knowledge_graph.Entity{Id: code, Type: knowledge_graph.EntityType_STOCK}
		entities = append(entities, stockEntity)
		summary := fmt.Sprintf("营收 %.2f, 净利 %.2f, EPS %.2f, 营收同比 %.2f%%, 净利同比 %.2f%%", match.TotalRevenue, match.NetProfit, match.Eps, match.RevenueYoy, match.ProfitYoy)
		ts := parseDate(date)
		evidence := &knowledge_graph.Evidence{Id: "evd_fin_" + eventID, Category: "financial", Summary: summary, Confidence: 100, Source: "financial", Timestamp: ts}
		relations = append(relations, &knowledge_graph.Relation{
			Source:    evEntity,
			Target:    stockEntity,
			Type:      knowledge_graph.RelationType_AFFECTS,
			Strength:  1,
			Evidence:  evidence,
			UpdatedAt: ts,
		})
		added++
		if added%500 == 0 {
			if len(entities) > 0 {
				_, _ = graphdb.UpsertEntities(ctx, entities)
				entities = entities[:0]
			}
			if len(relations) > 0 {
				_, _ = graphdb.UpsertRelations(ctx, relations)
				relations = relations[:0]
			}
		}
	}
	if len(entities) > 0 {
		_, _ = graphdb.UpsertEntities(ctx, entities)
	}
	if len(relations) > 0 {
		_, _ = graphdb.UpsertRelations(ctx, relations)
	}
	fmt.Printf("已回填财报summary并建立证据关系: %d 条\n", added)

	// 覆盖检查
	allStocksAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `MATCH (s:Entity) WHERE s.type = 1 RETURN collect(DISTINCT s.id) AS codes`, nil)
		if err != nil {
			return nil, err
		}
		if !result.Next(ctx) {
			return []string{}, result.Err()
		}
		val, _ := result.Record().Get("codes")
		arr, _ := val.([]any)
		out := make([]string, 0, len(arr))
		for _, v := range arr {
			out = append(out, toString(v))
		}
		return out, nil
	})
	if err != nil {
		fmt.Println("读取股票列表失败:", err)
		return
	}
	withFinAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
MATCH (s:Entity)
WHERE s.type = 1
OPTIONAL MATCH (s)-[:HAS_EVENT]->(ev:Event {source:'financial'})
OPTIONAL MATCH (s)<-[:RELATION]-(fe:Entity)
WHERE fe.type = 4 // EVENT
RETURN collect(DISTINCT s.id) AS all_codes,
       collect(DISTINCT CASE WHEN ev IS NOT NULL OR fe IS NOT NULL THEN s.id END) AS covered
`, nil)
		if err != nil {
			return nil, err
		}
		if !result.Next(ctx) {
			return struct {
				All     []string
				Covered []string
			}{}, result.Err()
		}
		rec := result.Record()
		allVal, _ := rec.Get("all_codes")
		coveredVal, _ := rec.Get("covered")
		allArr, _ := allVal.([]any)
		covArr, _ := coveredVal.([]any)
		all := make([]string, 0, len(allArr))
		cov := make([]string, 0, len(covArr))
		for _, v := range allArr {
			all = append(all, toString(v))
		}
		for _, v := range covArr {
			if v != nil {
				cov = append(cov, toString(v))
			}
		}
		return struct {
			All     []string
			Covered []string
		}{All: all, Covered: cov}, nil
	})
	if err != nil {
		fmt.Println("覆盖检查失败:", err)
		return
	}
	all := allStocksAny.([]string)
	cov := withFinAny.(struct {
		All     []string
		Covered []string
	}).Covered
	missing := diff(all, cov)
	fmt.Printf("股票总数: %d, 已有财报/事件覆盖: %d, 缺失: %d\n", len(all), len(cov), len(missing))
	if len(missing) > 0 {
		fmt.Println("缺失样例（前20）:")
		for i := 0; i < len(missing) && i < 20; i++ {
			fmt.Printf("- %s\n", missing[i])
		}
	}
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

func parseReportDateFromEventID(eid string) string {
	parts := strings.Split(eid, "_")
	if len(parts) >= 3 {
		return parts[2]
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

func diff(a, b []string) []string {
	set := map[string]struct{}{}
	for _, x := range b {
		set[x] = struct{}{}
	}
	miss := make([]string, 0, len(a))
	for _, x := range a {
		if _, ok := set[x]; !ok {
			miss = append(miss, x)
		}
	}
	return miss
}

