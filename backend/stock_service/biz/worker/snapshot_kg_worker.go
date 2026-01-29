package worker

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"stock_assistant/backend/common/eastmoney"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"
	"stock_assistant/backend/stock_service/biz/rpc"
)

type SnapshotKGWorker struct {
	ctx      context.Context
	cancel   context.CancelFunc
	emClient *eastmoney.Client
}

func NewSnapshotKGWorker() *SnapshotKGWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &SnapshotKGWorker{
		ctx:      ctx,
		cancel:   cancel,
		emClient: eastmoney.NewClient(),
	}
}

func (w *SnapshotKGWorker) Start() {
	if rpc.KnowledgeGraphClient == nil {
		log.Println("[SnapshotKGWorker] 客户端未初始化，无法启动")
		return
	}
	log.Println("[SnapshotKGWorker] 启动图谱快照入库 Worker...")
	go w.loop()
}

func (w *SnapshotKGWorker) Stop() {
	w.cancel()
}

func (w *SnapshotKGWorker) loop() {
	noticeTicker := time.NewTicker(4 * time.Hour)
	finTicker := time.NewTicker(24 * time.Hour)
	defer noticeTicker.Stop()
	defer finTicker.Stop()
	w.processNotices()
	w.processFinancials()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-noticeTicker.C:
			w.processNotices()
		case <-finTicker.C:
			w.processFinancials()
		}
	}
}

func (w *SnapshotKGWorker) processNotices() {
	ctx, cancel := context.WithTimeout(w.ctx, 60*time.Second)
	defer cancel()
	stocks, err := w.emClient.GetAStockList(ctx)
	if err != nil || len(stocks) == 0 {
		log.Printf("[SnapshotKGWorker] 获取A股列表失败: %v", err)
		return
	}
	now := time.Now().Unix()
	batch := 0
	entities := make([]*knowledge_graph.Entity, 0, 1000)
	relations := make([]*knowledge_graph.Relation, 0, 2000)
	for _, s := range stocks {
		if s == nil || s.Code == "" {
			continue
		}
		nitems, err := w.emClient.GetStockNotices(ctx, s.Code, nil)
		if err != nil || len(nitems) == 0 {
			continue
		}
		for _, it := range nitems {
			if it == nil || strings.TrimSpace(it.Title) == "" {
				continue
			}
			ts := parseDate(it.Date)
			eventID := fmt.Sprintf("event:notice:%s:%s:%s", s.Code, it.Date, hashKey(it.Title))
			eventEntity := &knowledge_graph.Entity{
				Id:   eventID,
				Type: knowledge_graph.EntityType_EVENT,
				Name: "公告",
				Attributes: map[string]string{
					"code":  normalizeCode(s.Code),
					"title": it.Title,
					"url":   it.Url,
					"date":  it.Date,
				},
			}
			entities = append(entities, eventEntity)
			stockEntity := &knowledge_graph.Entity{Id: normalizeCode(s.Code), Type: knowledge_graph.EntityType_STOCK, Name: s.Name, Attributes: map[string]string{"code": normalizeCode(s.Code)}}
			entities = append(entities, stockEntity)
			summary := trimText(it.Title, 200)
			evidence := &knowledge_graph.Evidence{Id: "evd_notice_" + eventID, Category: "notice", Summary: summary, Confidence: 100, Source: "notice", Timestamp: ts}
			relations = append(relations, &knowledge_graph.Relation{
				Source:    eventEntity,
				Target:    stockEntity,
				Type:      knowledge_graph.RelationType_AFFECTS,
				Strength:  1,
				Evidence:  evidence,
				UpdatedAt: ts,
			})
			batch++
		}
		if batch >= 1000 {
			break
		}
	}
	if len(entities) > 0 {
		_, _ = rpc.KnowledgeGraphClient.UpsertEntities(w.ctx, &knowledge_graph.UpsertEntitiesRequest{Entities: entities})
	}
	if len(relations) > 0 {
		_, _ = rpc.KnowledgeGraphClient.UpsertRelations(w.ctx, &knowledge_graph.UpsertRelationsRequest{Relations: relations})
	}
	log.Printf("[SnapshotKGWorker] 公告入库完成: %d 条", batch)
	_ = now
}

func (w *SnapshotKGWorker) processFinancials() {
	ctx, cancel := context.WithTimeout(w.ctx, 120*time.Second)
	defer cancel()
	stocks, err := w.emClient.GetAStockList(ctx)
	if err != nil || len(stocks) == 0 {
		log.Printf("[SnapshotKGWorker] 获取A股列表失败: %v", err)
		return
	}
	batch := 0
	entities := make([]*knowledge_graph.Entity, 0, 2000)
	relations := make([]*knowledge_graph.Relation, 0, 4000)
	for _, s := range stocks {
		if s == nil || s.Code == "" {
			continue
		}
		reports, err := w.emClient.GetFinancialReports(ctx, s.Code)
		if err != nil || len(reports) == 0 {
			continue
		}
		for _, r := range reports {
			if r == nil || r.ReportDate == "" {
				continue
			}
			ts := parseDate(r.ReportDate)
			eventID := fmt.Sprintf("event:financial:%s:%s", s.Code, r.ReportDate)
			eventEntity := &knowledge_graph.Entity{
				Id:   eventID,
				Type: knowledge_graph.EntityType_EVENT,
				Name: "财报",
				Attributes: map[string]string{
					"code":          normalizeCode(s.Code),
					"report_date":   r.ReportDate,
					"total_revenue": fmt.Sprintf("%.2f", r.TotalRevenue),
					"net_profit":    fmt.Sprintf("%.2f", r.NetProfit),
					"eps":           fmt.Sprintf("%.2f", r.Eps),
					"revenue_yoy":   fmt.Sprintf("%.2f", r.RevenueYoy),
					"profit_yoy":    fmt.Sprintf("%.2f", r.ProfitYoy),
				},
			}
			entities = append(entities, eventEntity)
			stockEntity := &knowledge_graph.Entity{Id: normalizeCode(s.Code), Type: knowledge_graph.EntityType_STOCK, Name: s.Name, Attributes: map[string]string{"code": normalizeCode(s.Code)}}
			entities = append(entities, stockEntity)
			summary := fmt.Sprintf("营收 %.2f, 净利 %.2f, EPS %.2f, 营收同比 %.2f%%, 净利同比 %.2f%%", r.TotalRevenue, r.NetProfit, r.Eps, r.RevenueYoy, r.ProfitYoy)
			evidence := &knowledge_graph.Evidence{Id: "evd_fin_" + eventID, Category: "financial", Summary: trimText(summary, 300), Confidence: 100, Source: "financial", Timestamp: ts}
			relations = append(relations, &knowledge_graph.Relation{
				Source:    eventEntity,
				Target:    stockEntity,
				Type:      knowledge_graph.RelationType_AFFECTS,
				Strength:  1,
				Evidence:  evidence,
				UpdatedAt: ts,
			})
			batch++
		}
		if batch >= 2000 {
			break
		}
	}
	if len(entities) > 0 {
		_, _ = rpc.KnowledgeGraphClient.UpsertEntities(w.ctx, &knowledge_graph.UpsertEntitiesRequest{Entities: entities})
	}
	if len(relations) > 0 {
		_, _ = rpc.KnowledgeGraphClient.UpsertRelations(w.ctx, &knowledge_graph.UpsertRelationsRequest{Relations: relations})
	}
	log.Printf("[SnapshotKGWorker] 财务入库完成: %d 条", batch)
}

func normalizeCode(code string) string {
	c := strings.ToLower(strings.TrimSpace(code))
	c = strings.TrimPrefix(c, "sh")
	c = strings.TrimPrefix(c, "sz")
	c = strings.TrimPrefix(c, "bj")
	if len(c) != 6 {
		return ""
	}
	if c[0] == '6' {
		return "sh" + c
	}
	if c[0] == '0' || c[0] == '3' {
		return "sz" + c
	}
	if c[0] == '8' {
		return "bj" + c
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
	layouts := []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return t.Unix()
		}
	}
	return time.Now().Unix()
}

func hashKey(text string) string {
	sum := sha1.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}

func trimText(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}

func (w *SnapshotKGWorker) processTHSConcepts() { return }
