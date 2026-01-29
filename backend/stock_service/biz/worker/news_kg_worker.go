package worker

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"stock_assistant/backend/common/eastmoney"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"
	"stock_assistant/backend/stock_service/biz/provider/crawler"
	"stock_assistant/backend/stock_service/biz/provider/sina"
	"stock_assistant/backend/stock_service/biz/rpc"
	"stock_assistant/backend/stock_service/kitex_gen/ai"
)

type NewsKGWorker struct {
	ctx        context.Context
	cancel     context.CancelFunc
	sinaClient *sina.Client
	emClient   *eastmoney.Client
}

type trendMeta struct {
	url       string
	timestamp int64
	content   string
}

func NewNewsKGWorker() *NewsKGWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &NewsKGWorker{
		ctx:        ctx,
		cancel:     cancel,
		sinaClient: sina.NewClient(),
		emClient:   eastmoney.NewClient(),
	}
}

func (w *NewsKGWorker) Start() {
	if rpc.KnowledgeGraphClient == nil || rpc.AIClient == nil {
		log.Println("[NewsKGWorker] 客户端未初始化，无法启动")
		return
	}
	log.Println("[NewsKGWorker] 启动图谱新闻入库 Worker...")
	go w.fetchLoop()
}

func (w *NewsKGWorker) Stop() {
	w.cancel()
}

func (w *NewsKGWorker) fetchLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	w.processNews()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.processNews()
		}
	}
}

func (w *NewsKGWorker) processNews() {
	rawItems, metaMap := w.fetchTrendSources()
	macroItems := w.fetchMacroData()
	if len(rawItems) == 0 && len(macroItems) == 0 {
		log.Println("[NewsKGWorker] 未获取到趋势数据")
		return
	}

	entities := make(map[string]*knowledge_graph.Entity)
	events := make([]*knowledge_graph.Event, 0)
	relations := make([]*knowledge_graph.Relation, 0)
	eventSeen := make(map[string]struct{})

	if len(rawItems) > 0 {
		resp, err := rpc.AIClient.ProcessMarketTrends(w.ctx, &ai.ProcessMarketTrendsRequest{Items: rawItems})
		if err != nil {
			log.Printf("[NewsKGWorker] 调用 AI 分析失败: %v", err)
			return
		}
		if len(resp.Items) > 0 {
			for _, trend := range resp.Items {
				if trend == nil || strings.TrimSpace(trend.Title) == "" {
					continue
				}
				meta := metaMap[trend.Title+"|"+trend.Source]
				eventTimestamp := meta.timestamp
				if eventTimestamp == 0 {
					eventTimestamp = time.Now().Unix()
				}
				eventID := buildTrendEventID(trend.Title, trend.Source, eventTimestamp)
				eventEntity := &knowledge_graph.Entity{
					Id:   eventID,
					Type: knowledge_graph.EntityType_EVENT,
					Name: trend.Title,
					Attributes: map[string]string{
						"source": trend.Source,
						"url":    meta.url,
					},
				}
				entities[eventEntity.Id] = eventEntity

				targets := make([]*knowledge_graph.Entity, 0)
				for _, sector := range trend.RelatedSectors {
					sectorName := strings.TrimSpace(sector)
					if sectorName == "" {
						continue
					}
					sectorEntity := buildSectorEntity(sectorName)
					entities[sectorEntity.Id] = sectorEntity
					targets = append(targets, sectorEntity)
				}

				for _, rawCode := range trend.RelatedStocks {
					code := normalizeStockCode(rawCode)
					if code == "" {
						continue
					}
					stockEntity := w.buildStockEntity(code)
					if stockEntity == nil {
						continue
					}
					entities[stockEntity.Id] = stockEntity
					targets = append(targets, stockEntity)
				}

				if len(targets) == 0 {
					continue
				}

				summary := strings.TrimSpace(trend.Summary)
				if summary == "" {
					summary = strings.TrimSpace(trend.ImpactAnalysis)
				}
				if summary == "" {
					summary = strings.TrimSpace(meta.content)
				}
				if summary == "" {
					summary = trend.Title
				}
				summary = trimEvidenceText(summary, 300)
				confidence := computeTrendConfidence(trend.FinancialRelevance, trend.Weight)
				strength := computeStrength(confidence)
				evidence := &knowledge_graph.Evidence{
					Id:         "evd_" + eventID,
					Category:   "trend",
					Summary:    summary,
					Confidence: confidence,
					Source:     trend.Source,
					Timestamp:  eventTimestamp,
				}

				for _, target := range targets {
					relations = append(relations, &knowledge_graph.Relation{
						Source:    eventEntity,
						Target:    target,
						Type:      knowledge_graph.RelationType_AFFECTS,
						Strength:  strength,
						Evidence:  evidence,
						UpdatedAt: eventTimestamp,
					})
				}

				if _, ok := eventSeen[eventID]; !ok {
					eventSeen[eventID] = struct{}{}
					events = append(events, &knowledge_graph.Event{
						Id:              eventID,
						Type:            mapTrendEventType(trend.ImpactType),
						Entities:        targets,
						ImpactDirection: 0,
						ImpactStrength:  1,
						Confidence:      confidence,
						Timestamp:       eventTimestamp,
						Source:          trend.Source,
						DedupeKey:       eventID,
					})
				}
			}
		}
	}

	w.appendMacroEvents(entities, &relations, &events, eventSeen, macroItems)

	w.appendSectorCoreStocks(entities, &relations)

	if len(events) == 0 && len(relations) == 0 {
		log.Println("[NewsKGWorker] 未生成可写入的图谱数据")
		return
	}

	entityList := make([]*knowledge_graph.Entity, 0, len(entities))
	for _, ent := range entities {
		entityList = append(entityList, ent)
	}

	var err error
	_, err = rpc.KnowledgeGraphClient.UpsertEntities(w.ctx, &knowledge_graph.UpsertEntitiesRequest{Entities: entityList})
	if err != nil {
		log.Printf("[NewsKGWorker] 实体写入失败: %v", err)
		return
	}

	if len(relations) > 0 {
		_, err = rpc.KnowledgeGraphClient.UpsertRelations(w.ctx, &knowledge_graph.UpsertRelationsRequest{Relations: relations})
		if err != nil {
			log.Printf("[NewsKGWorker] 关系写入失败: %v", err)
			return
		}
	}

	if len(events) > 0 {
		_, err = rpc.KnowledgeGraphClient.UpsertEvents(w.ctx, &knowledge_graph.UpsertEventsRequest{Events: events})
		if err != nil {
			log.Printf("[NewsKGWorker] 事件写入失败: %v", err)
			return
		}
	}
	log.Printf("[NewsKGWorker] 写入实体 %d 个，关系 %d 个，事件 %d 个", len(entityList), len(relations), len(events))
}

func (w *NewsKGWorker) fetchTrendSources() ([]*ai.RawTrendItem, map[string]trendMeta) {
	metaMap := make(map[string]trendMeta)
	items := make([]*ai.RawTrendItem, 0)

	var mu sync.Mutex
	appendItems := func(news []*crawler.NewsItem) {
		if len(news) == 0 {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		for _, item := range news {
			if item == nil {
				continue
			}
			title := strings.TrimSpace(item.Title)
			if title == "" {
				continue
			}
			content := strings.TrimSpace(item.Content)
			if content == "" {
				content = title
			}
			raw := &ai.RawTrendItem{
				Title:   title,
				Content: content,
				Source:  strings.TrimSpace(item.Source),
				Url:     strings.TrimSpace(item.Url),
			}
			items = append(items, raw)
			metaMap[raw.Title+"|"+raw.Source] = trendMeta{
				url:       raw.Url,
				timestamp: parseNewsTime(item.Time),
				content:   content,
			}
		}
	}

	sources := []func(){
		func() {
			if news, err := crawler.GetCailianPressTelegraph(); err == nil {
				appendItems(news)
			}
		},
		func() {
			if news, err := crawler.GetWallstreetCNHotTrends("day"); err == nil {
				appendItems(news)
			}
		},
		func() {
			if news, err := crawler.GetThePaperHotTrends(); err == nil {
				appendItems(news)
			}
		},
		func() {
			if news, err := crawler.GetAkShareNews(); err == nil {
				appendItems(news)
			}
		},
	}

	var wg sync.WaitGroup
	wg.Add(len(sources))
	for _, fn := range sources {
		go func(run func()) {
			defer wg.Done()
			run()
		}(fn)
	}
	wg.Wait()

	log.Printf("[NewsKGWorker] 获取趋势来源 %d 条", len(items))
	return items, metaMap
}

func (w *NewsKGWorker) fetchMacroData() []*crawler.MacroItem {
	items, err := crawler.GetAkShareMacroData()
	if err != nil {
		log.Printf("[NewsKGWorker] 获取宏观数据失败: %v", err)
		return nil
	}
	if len(items) == 0 {
		return nil
	}
	log.Printf("[NewsKGWorker] 获取宏观数据 %d 条", len(items))
	return items
}

func (w *NewsKGWorker) appendMacroEvents(entities map[string]*knowledge_graph.Entity, relations *[]*knowledge_graph.Relation, events *[]*knowledge_graph.Event, eventSeen map[string]struct{}, items []*crawler.MacroItem) {
	if len(items) == 0 {
		return
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		indicator := strings.TrimSpace(item.Indicator)
		label := strings.TrimSpace(item.Label)
		if indicator == "" && label == "" {
			continue
		}
		if label == "" {
			label = indicator
		}
		macroID := "macro:" + indicator
		if indicator == "" {
			macroID = "macro:" + label
		}
		macroEntity := &knowledge_graph.Entity{
			Id:   macroID,
			Type: knowledge_graph.EntityType_MACRO,
			Name: label,
			Attributes: map[string]string{
				"indicator": indicator,
				"unit":      strings.TrimSpace(item.Unit),
				"source":    strings.TrimSpace(item.Source),
			},
		}
		entities[macroEntity.Id] = macroEntity

		eventTimestamp := parseNewsTime(item.Time)
		if eventTimestamp == 0 {
			eventTimestamp = time.Now().Unix()
		}
		eventID := buildMacroEventID(indicator, label, item.Value, eventTimestamp)
		eventEntity := &knowledge_graph.Entity{
			Id:   eventID,
			Type: knowledge_graph.EntityType_EVENT,
			Name: label,
			Attributes: map[string]string{
				"source": strings.TrimSpace(item.Source),
				"unit":   strings.TrimSpace(item.Unit),
			},
		}
		entities[eventEntity.Id] = eventEntity

		value := strings.TrimSpace(item.Value)
		summary := label
		if value != "" {
			unit := strings.TrimSpace(item.Unit)
			if unit != "" {
				summary = fmt.Sprintf("%s: %s %s", label, value, unit)
			} else {
				summary = fmt.Sprintf("%s: %s", label, value)
			}
		}

		evidence := &knowledge_graph.Evidence{
			Id:         "evd_" + eventID,
			Category:   "macro",
			Summary:    trimEvidenceText(summary, 300),
			Confidence: 70,
			Source:     strings.TrimSpace(item.Source),
			Timestamp:  eventTimestamp,
		}

		*relations = append(*relations, &knowledge_graph.Relation{
			Source:    eventEntity,
			Target:    macroEntity,
			Type:      knowledge_graph.RelationType_AFFECTS,
			Strength:  computeStrength(70),
			Evidence:  evidence,
			UpdatedAt: eventTimestamp,
		})

		if _, ok := eventSeen[eventID]; !ok {
			eventSeen[eventID] = struct{}{}
			*events = append(*events, &knowledge_graph.Event{
				Id:              eventID,
				Type:            knowledge_graph.EventType_REPORT,
				Entities:        []*knowledge_graph.Entity{macroEntity},
				ImpactDirection: 0,
				ImpactStrength:  1,
				Confidence:      70,
				Timestamp:       eventTimestamp,
				Source:          strings.TrimSpace(item.Source),
				DedupeKey:       eventID,
			})
		}
	}
}

func trimEvidenceText(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}

func (w *NewsKGWorker) appendSectorCoreStocks(entities map[string]*knowledge_graph.Entity, relations *[]*knowledge_graph.Relation) {
	ctx, cancel := context.WithTimeout(w.ctx, 20*time.Second)
	defer cancel()

	conceptList, err := w.emClient.GetSectorList(ctx, "concept")
	if err != nil {
		log.Printf("[NewsKGWorker] 获取概念板块失败: %v", err)
	}
	industryList, err := w.emClient.GetSectorList(ctx, "industry")
	if err != nil {
		log.Printf("[NewsKGWorker] 获取行业板块失败: %v", err)
	}
	relationKey := make(map[string]struct{})
	process := func(list []*eastmoney.SectorInfo) {
		for _, item := range list {
			if item == nil || item.Code == "" || strings.TrimSpace(item.Name) == "" {
				continue
			}
			sectorEntity := buildSectorEntityWithCode(item.Name, item.Code)
			entities[sectorEntity.Id] = sectorEntity
			stocks, err := w.emClient.GetSectorStocksRaw(ctx, item.Code)
			if err != nil || len(stocks) == 0 {
				continue
			}
			sort.Slice(stocks, func(i, j int) bool {
				return strings.Compare(stocks[i].Name, stocks[j].Name) < 0
			})
			limit := 10
			if len(stocks) < limit {
				limit = len(stocks)
			}
			ev := &knowledge_graph.Evidence{
				Id:         "evd_sector_" + item.Code,
				Category:   "sector",
				Summary:    "板块核心股票清单",
				Confidence: 85,
				Source:     "板块核心股票",
				Timestamp:  time.Now().Unix(),
			}
			for i := 0; i < limit; i++ {
				code := normalizeStockCode(stocks[i].Code)
				name := strings.TrimSpace(stocks[i].Name)
				if code == "" || name == "" {
					continue
				}
				up := strings.ToUpper(name)
				if strings.Contains(up, "ST") || strings.Contains(up, "退") || strings.HasPrefix(up, "N") || strings.HasPrefix(up, "C") {
					continue
				}
				stockEntity := w.buildStockEntity(code)
				if stockEntity == nil {
					continue
				}
				entities[stockEntity.Id] = stockEntity
				key := stockEntity.Id + "|" + sectorEntity.Id
				if _, ok := relationKey[key]; ok {
					continue
				}
				relationKey[key] = struct{}{}
				*relations = append(*relations, &knowledge_graph.Relation{
					Source:    stockEntity,
					Target:    sectorEntity,
					Type:      knowledge_graph.RelationType_BELONGS_TO,
					Strength:  computeStrength(85),
					Evidence:  ev,
					UpdatedAt: ev.Timestamp,
				})
			}
		}
	}
	process(conceptList)
	process(industryList)
}

func (w *NewsKGWorker) buildStockEntity(code string) *knowledge_graph.Entity {
	name := code
	info, err := w.sinaClient.GetStockInfo(context.Background(), code)
	if err == nil && info != nil && info.Name != "" {
		name = info.Name
	}
	return &knowledge_graph.Entity{
		Id:   code,
		Type: knowledge_graph.EntityType_STOCK,
		Name: name,
		Attributes: map[string]string{
			"code": code,
		},
	}
}

func buildSectorEntity(name string) *knowledge_graph.Entity {
	id := "sector:" + name
	return &knowledge_graph.Entity{
		Id:   id,
		Type: knowledge_graph.EntityType_SECTOR,
		Name: name,
		Attributes: map[string]string{
			"name": name,
		},
	}
}

func buildSectorEntityWithCode(name, code string) *knowledge_graph.Entity {
	id := "sector:" + name
	return &knowledge_graph.Entity{
		Id:   id,
		Type: knowledge_graph.EntityType_SECTOR,
		Name: name,
		Attributes: map[string]string{
			"name": name,
			"code": code,
		},
	}
}

func matchSector(list []*eastmoney.SectorInfo, name string, aliases []string) (string, string) {
	candidates := append([]string{name}, aliases...)
	for _, item := range list {
		if item == nil {
			continue
		}
		for _, key := range candidates {
			if key != "" && strings.Contains(item.Name, key) {
				return item.Code, item.Name
			}
		}
	}
	return "", ""
}

func parseNewsTime(value string) int64 {
	if value == "" {
		return time.Now().Unix()
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02", time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, value, loc); err == nil {
			return t.Unix()
		}
	}
	return time.Now().Unix()
}

func buildTrendEventID(title, source string, ts int64) string {
	data := strings.Join([]string{title, source, strconv.FormatInt(ts, 10)}, "|")
	sum := sha1.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

func buildMacroEventID(indicator, label, value string, ts int64) string {
	data := strings.Join([]string{indicator, label, value, strconv.FormatInt(ts, 10)}, "|")
	sum := sha1.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

func normalizeStockCode(raw string) string {
	code := strings.ToLower(strings.TrimSpace(raw))
	code = strings.TrimPrefix(code, "sh")
	code = strings.TrimPrefix(code, "sz")
	if len(code) != 6 {
		return ""
	}
	if code[0] == '6' {
		return "sh" + code
	}
	if code[0] == '0' || code[0] == '3' {
		return "sz" + code
	}
	return ""
}

func computeTrendConfidence(financialRelevance int32, weight float64) int32 {
	base := int32(40)
	if financialRelevance > 0 {
		base += financialRelevance * 5
	}
	if weight >= 1.2 {
		base += 5
	}
	if base > 90 {
		return 90
	}
	if base < 40 {
		return 40
	}
	return base
}

func computeStrength(confidence int32) float64 {
	strength := 0.3 + float64(confidence)/100*0.5
	if strength > 0.8 {
		return 0.8
	}
	if strength < 0.3 {
		return 0.3
	}
	return strength
}

func mapTrendEventType(impactType string) knowledge_graph.EventType {
	lower := strings.ToLower(impactType)
	if strings.Contains(lower, "policy") {
		return knowledge_graph.EventType_POLICY
	}
	return knowledge_graph.EventType_NEWS
}
