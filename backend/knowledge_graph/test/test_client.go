package testclient

import (
	"context"
	"fmt"
	"log"
	"time"

	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"
	knowledge_graph_service "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph/knowledgegraphservice"

	"github.com/cloudwego/kitex/client"
)

func Run() {
	// 创建知识图谱服务客户端
	client, err := knowledge_graph_service.NewClient("knowledge_graph", client.WithHostPorts("localhost:8890"))
	if err != nil {
		log.Fatalf("创建客户端失败: %v\n", err)
	}

	ctx := context.Background()

	// 测试1: 写入实体
	fmt.Println("=== 测试1: 写入实体 ===")
	entityReq := &knowledge_graph.UpsertEntitiesRequest{
		Entities: []*knowledge_graph.Entity{
			{
				Id:   "600519",
				Type: knowledge_graph.EntityType_STOCK,
				Name: "贵州茅台",
				Attributes: map[string]string{
					"industry": "白酒",
					"sector":   "食品饮料",
				},
			},
			{
				Id:   "白酒行业",
				Type: knowledge_graph.EntityType_INDUSTRY,
				Name: "白酒行业",
				Attributes: map[string]string{
					"category": "消费",
				},
			},
		},
	}

	_, err = client.UpsertEntities(ctx, entityReq)
	if err != nil {
		log.Fatalf("写入实体失败: %v\n", err)
	}
	fmt.Printf("写入实体成功\n")

	// 测试2: 写入关系
	fmt.Println("=== 测试2: 写入关系 ===")
	relationReq := &knowledge_graph.UpsertRelationsRequest{
		Relations: []*knowledge_graph.Relation{
			{
				Source: &knowledge_graph.Entity{
					Id:   "600519",
					Type: knowledge_graph.EntityType_STOCK,
				},
				Target: &knowledge_graph.Entity{
					Id:   "白酒行业",
					Type: knowledge_graph.EntityType_INDUSTRY,
				},
				Type:     knowledge_graph.RelationType_BELONGS_TO,
				Strength: 1.0,
				Evidence: &knowledge_graph.Evidence{
					Id:         "test_evidence_1",
					Category:   "industry_belonging",
					Summary:    "贵州茅台属于白酒行业",
					Confidence: 95,
					Source:     "测试来源",
					Timestamp:  time.Now().Unix(),
				},
			},
		},
	}

	_, err = client.UpsertRelations(ctx, relationReq)
	if err != nil {
		log.Fatalf("写入关系失败: %v\n", err)
	}
	fmt.Printf("写入关系成功\n")

	// 测试3: 查询实体邻居
	fmt.Println("=== 测试3: 查询实体邻居 ===")
	neighborReq := &knowledge_graph.GetGraphNeighborhoodRequest{
		EntityId:   "600519",
		EntityType: knowledge_graph.EntityType_STOCK,
		Depth:      1,
		MaxEdges:   10,
	}

	neighborResp, err := client.GetGraphNeighborhood(ctx, neighborReq)
	if err != nil {
		log.Fatalf("查询邻居失败: %v\n", err)
	}
	fmt.Printf("查询到 %d 个实体，%d 个关系\n", len(neighborResp.Entities), len(neighborResp.Relations))
	for _, entity := range neighborResp.Entities {
		fmt.Printf("实体: %s (%s)\n", entity.Name, entity.Id)
	}
	for _, relation := range neighborResp.Relations {
		fmt.Printf("关系: %s -> %s (%s)\n", relation.Source.Id, relation.Target.Id, relation.Type.String())
	}

	// 测试4: 查询实体详情
	fmt.Println("=== 测试4: 查询实体详情 ===")
	profileReq := &knowledge_graph.GetEntityProfileRequest{
		EntityId:   "600519",
		EntityType: knowledge_graph.EntityType_STOCK,
	}

	profileResp, err := client.GetEntityProfile(ctx, profileReq)
	if err != nil {
		log.Fatalf("查询实体详情失败: %v\n", err)
	}
	if profileResp.Entity != nil {
		fmt.Printf("实体详情: %s (%s), 属性: %v\n", profileResp.Entity.Name, profileResp.Entity.Id, profileResp.Entity.Attributes)
	}
	fmt.Printf("关联关系数: %d, 关联事件数: %d\n", len(profileResp.Relations), len(profileResp.Events))

	// 测试5: 写入事件
	fmt.Println("=== 测试5: 写入事件 ===")
	eventReq := &knowledge_graph.UpsertEventsRequest{
		Events: []*knowledge_graph.Event{
			{
				Id:              "event_1",
				Type:            knowledge_graph.EventType_NEWS,
				Entities:        []*knowledge_graph.Entity{{Id: "600519", Type: knowledge_graph.EntityType_STOCK}},
				ImpactDirection: 1,
				ImpactStrength:  8,
				Confidence:      85,
				Timestamp:       time.Now().Unix(),
				Source:          "测试新闻源",
				DedupeKey:       "test_news_600519_20260123",
			},
		},
	}

	_, err = client.UpsertEvents(ctx, eventReq)
	if err != nil {
		log.Fatalf("写入事件失败: %v\n", err)
	}
	fmt.Printf("写入事件成功\n")

	// 测试6: 查询事件
	fmt.Println("=== 测试6: 查询事件 ===")
	searchEventsReq := &knowledge_graph.SearchEventsRequest{
		EntityId:   "600519",
		EntityType: knowledge_graph.EntityType_STOCK,
		Limit:      10,
	}

	searchEventsResp, err := client.SearchEvents(ctx, searchEventsReq)
	if err != nil {
		log.Fatalf("查询事件失败: %v\n", err)
	}
	fmt.Printf("查询到 %d 个事件\n", len(searchEventsResp.Events))
	for _, event := range searchEventsResp.Events {
		fmt.Printf("事件: %s, 类型: %s, 置信度: %d\n", event.Id, event.Type.String(), event.Confidence)
	}

	// 测试7: 获取证据包
	fmt.Println("=== 测试7: 获取证据包 ===")
	evidenceBundleReq := &knowledge_graph.GetEvidenceBundleRequest{
		StockCode: "600519",
		StartTime: time.Now().AddDate(0, 0, -7).Unix(),
		EndTime:   time.Now().AddDate(0, 0, 1).Unix(),
		MaxItems:  20,
	}

	evidenceBundleResp, err := client.GetEvidenceBundle(ctx, evidenceBundleReq)
	if err != nil {
		log.Fatalf("获取证据包失败: %v\n", err)
	}
	bundle := evidenceBundleResp.Bundle
	fmt.Printf("证据包: %d 个实体, %d 个关系, %d 个事件, 评分: %.2f\n",
		len(bundle.Entities), len(bundle.Relations), len(bundle.Events), bundle.Score)

	fmt.Println("=== 所有测试完成 ===")
}
