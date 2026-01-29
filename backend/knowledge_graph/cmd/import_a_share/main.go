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
	now := time.Now().Unix()

	market := &knowledge_graph.Entity{Id: "market:a-share", Type: knowledge_graph.EntityType_SECTOR, Name: "A股市场", Attributes: map[string]string{"type": "market"}}
	boards := []*knowledge_graph.Entity{
		{Id: "board:shanghai", Type: knowledge_graph.EntityType_SECTOR, Name: "上证", Attributes: map[string]string{"type": "trading_board", "prefix": "sh"}},
		{Id: "board:shenzhen", Type: knowledge_graph.EntityType_SECTOR, Name: "深证", Attributes: map[string]string{"type": "trading_board", "prefix": "sz"}},
		{Id: "board:chinext", Type: knowledge_graph.EntityType_SECTOR, Name: "创业板", Attributes: map[string]string{"type": "trading_board", "prefix": "sz300"}},
		{Id: "board:star", Type: knowledge_graph.EntityType_SECTOR, Name: "科创", Attributes: map[string]string{"type": "trading_board", "prefix": "sh688"}},
		{Id: "board:north", Type: knowledge_graph.EntityType_SECTOR, Name: "北证", Attributes: map[string]string{"type": "trading_board", "prefix": "bj"}},
	}
	entities := []*knowledge_graph.Entity{market}
	entities = append(entities, boards...)
	_, _ = graphdb.UpsertEntities(ctx, entities)

	boardRelEvidence := &knowledge_graph.Evidence{Id: "evd_market_boards", Category: "market", Summary: "交易板块归属", Confidence: 100, Source: "system", Timestamp: now}
	boardRels := make([]*knowledge_graph.Relation, 0, len(boards))
	for _, b := range boards {
		boardRels = append(boardRels, &knowledge_graph.Relation{
			Source:    market,
			Target:    b,
			Type:      knowledge_graph.RelationType_BELONGS_TO,
			Strength:  0.9,
			Evidence:  boardRelEvidence,
			UpdatedAt: now,
		})
	}
	_, _ = graphdb.UpsertRelations(ctx, boardRels)

	concepts, _ := em.GetSectorList(ctx, "concept")
	industries, _ := em.GetSectorList(ctx, "industry")
	sectorEntities := make([]*knowledge_graph.Entity, 0, len(concepts)+len(industries))
	sectorRels := make([]*knowledge_graph.Relation, 0, len(concepts)+len(industries))
	for _, s := range concepts {
		if s == nil || s.Name == "" {
			continue
		}
		e := &knowledge_graph.Entity{Id: "sector:" + s.Name, Type: knowledge_graph.EntityType_SECTOR, Name: s.Name, Attributes: map[string]string{"type": "concept", "code": s.Code}}
		sectorEntities = append(sectorEntities, e)
		sectorRels = append(sectorRels, &knowledge_graph.Relation{
			Source:    market,
			Target:    e,
			Type:      knowledge_graph.RelationType_BELONGS_TO,
			Strength:  0.8,
			Evidence:  &knowledge_graph.Evidence{Id: "evd_market_sector_concept_" + s.Code, Category: "sector", Summary: "概念板块归属", Confidence: 90, Source: "eastmoney", Timestamp: now},
			UpdatedAt: now,
		})
	}
	for _, s := range industries {
		if s == nil || s.Name == "" {
			continue
		}
		e := &knowledge_graph.Entity{Id: "sector:" + s.Name, Type: knowledge_graph.EntityType_SECTOR, Name: s.Name, Attributes: map[string]string{"type": "industry", "code": s.Code}}
		sectorEntities = append(sectorEntities, e)
		sectorRels = append(sectorRels, &knowledge_graph.Relation{
			Source:    market,
			Target:    e,
			Type:      knowledge_graph.RelationType_BELONGS_TO,
			Strength:  0.8,
			Evidence:  &knowledge_graph.Evidence{Id: "evd_market_sector_industry_" + s.Code, Category: "sector", Summary: "行业板块归属", Confidence: 90, Source: "eastmoney", Timestamp: now},
			UpdatedAt: now,
		})
	}
	_, _ = graphdb.UpsertEntities(ctx, sectorEntities)
	_, _ = graphdb.UpsertRelations(ctx, sectorRels)

	conceptRanks, _ := em.GetSectorRank(ctx, "concept", 500)
	industryRanks, _ := em.GetSectorRank(ctx, "industry", 500)
	leaderBySectorName := map[string]string{}
	for _, r := range conceptRanks {
		if r != nil && r.Name != "" && r.TopStockCode != "" {
			leaderBySectorName[r.Name] = normalizeCode(r.TopStockCode)
		}
	}
	for _, r := range industryRanks {
		if r != nil && r.Name != "" && r.TopStockCode != "" {
			leaderBySectorName[r.Name] = normalizeCode(r.TopStockCode)
		}
	}

	for _, s := range append(concepts, industries...) {
		if s == nil || s.Code == "" || s.Name == "" {
			continue
		}
		stocks, err := em.GetSectorStocksAll(ctx, s.Code)
		if err != nil || len(stocks) == 0 {
			continue
		}
		stockEntities := make([]*knowledge_graph.Entity, 0, len(stocks))
		stockRels := make([]*knowledge_graph.Relation, 0, len(stocks)*2)
		leaderCode := leaderBySectorName[s.Name]
		for i := 0; i < len(stocks); i++ {
			code := normalizeCode(stocks[i].Code)
			name := strings.TrimSpace(stocks[i].Name)
			if code == "" || name == "" {
				continue
			}
			if isFiltered(name) {
				continue
			}
			se := &knowledge_graph.Entity{Id: code, Type: knowledge_graph.EntityType_STOCK, Name: name, Attributes: map[string]string{"code": code}}
			stockEntities = append(stockEntities, se)
			core := i < 10
			leader := leaderCode != "" && leaderCode == code
			summary := "板块成分关系"
			if leader {
				summary = "板块龙头关系"
			} else if core {
				summary = "板块核心关系"
			}
			strength := 0.6
			if core {
				strength = 0.8
			}
			if leader {
				strength = 0.95
			}
			ev := &knowledge_graph.Evidence{
				Id:         "evd_sector_rel_" + s.Code + "_" + code,
				Category:   "sector",
				Summary:    summary,
				Confidence: 85,
				Source:     "eastmoney",
				Timestamp:  now,
			}
			stockRels = append(stockRels, &knowledge_graph.Relation{
				Source:    se,
				Target:    &knowledge_graph.Entity{Id: "sector:" + s.Name, Type: knowledge_graph.EntityType_SECTOR, Name: s.Name},
				Type:      knowledge_graph.RelationType_BELONGS_TO,
				Strength:  strength,
				Evidence:  ev,
				UpdatedAt: now,
			})
			boardID := boardOf(code)
			if boardID != "" {
				stockRels = append(stockRels, &knowledge_graph.Relation{
					Source:    se,
					Target:    boardEntity(boardID),
					Type:      knowledge_graph.RelationType_BELONGS_TO,
					Strength:  0.9,
					Evidence:  &knowledge_graph.Evidence{Id: "evd_stock_board_" + code, Category: "market", Summary: "股票市场归属", Confidence: 100, Source: "system", Timestamp: now},
					UpdatedAt: now,
				})
			}
			if len(stockEntities) >= 1000 || len(stockRels) >= 2000 {
				_, _ = graphdb.UpsertEntities(ctx, stockEntities)
				_, _ = graphdb.UpsertRelations(ctx, stockRels)
				stockEntities = stockEntities[:0]
				stockRels = stockRels[:0]
			}
		}
		_, _ = graphdb.UpsertEntities(ctx, stockEntities)
		_, _ = graphdb.UpsertRelations(ctx, stockRels)
	}
	fmt.Println("A股主干导入完成")
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

func isFiltered(name string) bool {
	n := strings.ToUpper(strings.TrimSpace(name))
	if strings.Contains(n, "ST") || strings.Contains(n, "退") {
		return true
	}
	if strings.HasPrefix(n, "N") || strings.HasPrefix(n, "C") {
		return true
	}
	return false
}

func boardOf(code string) string {
	if strings.HasPrefix(code, "sh688") {
		return "board:star"
	}
	if strings.HasPrefix(code, "sz300") {
		return "board:chinext"
	}
	if strings.HasPrefix(code, "bj") {
		return "board:north"
	}
	if strings.HasPrefix(code, "sh") {
		return "board:shanghai"
	}
	if strings.HasPrefix(code, "sz") {
		return "board:shenzhen"
	}
	return ""
}

func boardEntity(id string) *knowledge_graph.Entity {
	switch id {
	case "board:shanghai":
		return &knowledge_graph.Entity{Id: id, Type: knowledge_graph.EntityType_SECTOR, Name: "上证"}
	case "board:shenzhen":
		return &knowledge_graph.Entity{Id: id, Type: knowledge_graph.EntityType_SECTOR, Name: "深证"}
	case "board:chinext":
		return &knowledge_graph.Entity{Id: id, Type: knowledge_graph.EntityType_SECTOR, Name: "创业板"}
	case "board:star":
		return &knowledge_graph.Entity{Id: id, Type: knowledge_graph.EntityType_SECTOR, Name: "科创"}
	case "board:north":
		return &knowledge_graph.Entity{Id: id, Type: knowledge_graph.EntityType_SECTOR, Name: "北证"}
	default:
		return nil
	}
}
