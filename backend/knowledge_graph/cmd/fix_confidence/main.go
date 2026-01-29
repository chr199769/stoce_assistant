package main

import (
	"context"
	"fmt"
	"stock_assistant/backend/knowledge_graph/config"
	"stock_assistant/backend/knowledge_graph/dal/graphdb"
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
	updated, err := graphdb.UpdateEventConfidenceBySource(ctx, []string{"financial", "notice"}, 100)
	if err != nil {
		fmt.Println("修复置信度失败:", err)
		return
	}
	fmt.Printf("已将 financial/notice 事件的置信度修复为 100，受影响事件数: %d\n", updated)
}
