package main

import (
	"context"
	"fmt"
	"stock_assistant/backend/knowledge_graph/config"
	"stock_assistant/backend/knowledge_graph/dal/graphdb"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		return
	}
	if err := graphdb.Init(); err != nil {
		fmt.Printf("Neo4j 初始化失败: %v\n", err)
		return
	}
	before, err := graphdb.CountNodes(context.Background())
	if err != nil {
		fmt.Printf("统计节点失败: %v\n", err)
		return
	}
	deleted, err := graphdb.DeleteAll(context.Background())
	if err != nil {
		fmt.Printf("清空图谱失败: %v\n", err)
		return
	}
	after, err := graphdb.CountNodes(context.Background())
	if err != nil {
		fmt.Printf("统计节点失败: %v\n", err)
		return
	}
	fmt.Printf("清空图谱完成: 删除前=%d 删除操作计数=%d 删除后=%d\n", before, deleted, after)
}

