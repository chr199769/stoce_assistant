package main

import (
	"context"
	"fmt"
	"log"

	"stock_assistant/backend/knowledge_graph/config"
	"stock_assistant/backend/knowledge_graph/dal/graphdb"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}
	if err := graphdb.Init(); err != nil {
		log.Fatalf("Neo4j 初始化失败: %v", err)
	}
	ctx := context.Background()
	session := graphdb.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// 总计
	totalStocks, totalEvents := aggregateCounts(ctx, session)
	fmt.Printf("已入库股票数量(含财务/公告): %d, 相关事件总数: %d\n", totalStocks, totalEvents)

	// 按股票统计 Top 50
	records, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
MATCH (s:Entity)-[:HAS_EVENT]->(ev:Event)
WHERE ev.source IN ['financial','notice'] AND s.type = 1
RETURN s.id AS stock, count(ev) AS cnt
ORDER BY cnt DESC
LIMIT 50
`, nil)
		if err != nil {
			return nil, err
		}
		list := make([]struct {
			Stock string
			Cnt   int64
		}, 0)
		for result.Next(ctx) {
			record := result.Record()
			stock, _ := record.Get("stock")
			cnt, _ := record.Get("cnt")
			list = append(list, struct {
				Stock string
				Cnt   int64
			}{
				Stock: toString(stock),
				Cnt:   toInt64(cnt),
			})
		}
		return list, result.Err()
	})
	if err != nil {
		log.Fatalf("查询 Top50 失败: %v", err)
	}
	top := records.([]struct {
		Stock string
		Cnt   int64
	})
	fmt.Println("样例（Top 50，按事件数降序）：")
	for _, item := range top {
		fmt.Printf("- %s (%d)\n", item.Stock, item.Cnt)
	}
}

func aggregateCounts(ctx context.Context, session neo4j.SessionWithContext) (int64, int64) {
	// distinct 股票数
	stocksAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
MATCH (s:Entity)-[:HAS_EVENT]->(ev:Event)
WHERE ev.source IN ['financial','notice'] AND s.type = 1
RETURN count(DISTINCT s.id) AS c
`, nil)
		if err != nil {
			return nil, err
		}
		if !result.Next(ctx) {
			return int64(0), result.Err()
		}
		val, _ := result.Record().Get("c")
		return toInt64(val), nil
	})
	if err != nil {
		return 0, 0
	}
	// 事件总数
	eventsAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
MATCH (s:Entity)-[:HAS_EVENT]->(ev:Event)
WHERE ev.source IN ['financial','notice'] AND s.type = 1
RETURN count(ev) AS c
`, nil)
		if err != nil {
			return nil, err
		}
		if !result.Next(ctx) {
			return int64(0), result.Err()
		}
		val, _ := result.Record().Get("c")
		return toInt64(val), nil
	})
	if err != nil {
		return stocksAny.(int64), 0
	}
	return stocksAny.(int64), eventsAny.(int64)
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

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int32:
		return int64(x)
	case int:
		return int64(x)
	case float64:
		return int64(x)
	case float32:
		return int64(x)
	default:
		return 0
	}
}

