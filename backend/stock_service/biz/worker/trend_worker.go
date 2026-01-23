package worker

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"stock_assistant/backend/stock_service/biz/provider/crawler"
	"stock_assistant/backend/stock_service/biz/provider/sina"
	"stock_assistant/backend/stock_service/biz/rpc"
	"stock_assistant/backend/stock_service/dal/model"
	"stock_assistant/backend/stock_service/dal/mysql"
	"stock_assistant/backend/stock_service/kitex_gen/ai"
)

type TrendWorker struct {
	ctx        context.Context
	cancel     context.CancelFunc
	sinaClient *sina.Client
}

func NewTrendWorker() *TrendWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &TrendWorker{
		ctx:        ctx,
		cancel:     cancel,
		sinaClient: sina.NewClient(),
	}
}

func (w *TrendWorker) Start() {
	if rpc.AIClient == nil {
		log.Println("[TrendWorker] AI Client 未初始化，无法启动")
		return
	}
	log.Println("[TrendWorker] 启动市场趋势分析 Worker...")

	go w.fetchLoop()
	go w.cleanLoop()
}

func (w *TrendWorker) Stop() {
	w.cancel()
}

func (w *TrendWorker) fetchLoop() {
	ticker := time.NewTicker(1 * time.Hour) // 每1小时抓取一次
	defer ticker.Stop()

	// 启动时立即执行一次
	w.processNewTrends()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.processNewTrends()
		}
	}
}

func (w *TrendWorker) cleanLoop() {
	ticker := time.NewTicker(24 * time.Hour) // 每天清理一次
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.cleanupOldTrends()
		}
	}
}

// processNewTrends 抓取并分析新趋势
func (w *TrendWorker) processNewTrends() {
	log.Println("[TrendWorker] 开始抓取市场热点...")

	rawItems := w.fetchHotTopics()
	if len(rawItems) == 0 {
		log.Println("[TrendWorker] 未获取到热点内容")
		return
	}

	req := &ai.ProcessMarketTrendsRequest{
		Items: rawItems,
	}

	resp, err := rpc.AIClient.ProcessMarketTrends(w.ctx, req)
	if err != nil {
		log.Printf("[TrendWorker] 调用 AI 分析失败: %v", err)
		return
	}

	if len(resp.Items) == 0 {
		log.Println("[TrendWorker] AI 未返回有效趋势")
		return
	}

	log.Printf("[TrendWorker] AI 分析完成，获取到 %d 个趋势", len(resp.Items))
	w.saveTrends(resp.Items)
}

// cleanupOldTrends 清理过时趋势
func (w *TrendWorker) cleanupOldTrends() {
	log.Println("[TrendWorker] 开始清理过时趋势...")

	// 1. 硬性清理：超过7天的短期消息，直接标记无效
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	mysql.DB.Model(&model.MarketTrend{}).
		Where("impact_type = ? AND created_at < ? AND is_still_valid = ?", "short_term", sevenDaysAgo, true).
		Update("is_still_valid", false)

	// 2. 智能清理：超过24小时的短期消息，重新评估
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	var oldTrends []model.MarketTrend
	mysql.DB.Where("impact_type = ? AND created_at < ? AND is_still_valid = ?", "short_term", oneDayAgo, true).
		Find(&oldTrends)

	for _, trend := range oldTrends {
		// 构造请求，让 AI 判断是否还有效
		// 这里简化逻辑：将旧趋势作为 raw_content 发给 AI
		// 如果 AI 重新生成的趋势中，包含该内容且权重依然较高，则保留

		req := &ai.ProcessMarketTrendsRequest{
			Items: []*ai.RawTrendItem{
				{
					Title:   trend.Title,
					Content: trend.Summary,
					Source:  "validity_check",
				},
			},
		}

		resp, err := rpc.AIClient.ProcessMarketTrends(w.ctx, req)
		if err != nil {
			continue
		}

		// 如果 AI 没有返回趋势，或者返回的趋势权重很低，则标记为无效
		isValid := false
		if len(resp.Items) > 0 {
			// 假设如果 AI 还能识别出趋势，且权重 > 0.3，则认为有效
			if resp.Items[0].Weight > 0.3 {
				isValid = true
			}
		}

		if !isValid {
			log.Printf("[TrendWorker] 趋势已过时: %s", trend.Title)
			mysql.DB.Model(&trend).Update("is_still_valid", false)
		}
	}
}

func (w *TrendWorker) saveTrends(trends []*ai.AnalyzedTrendItem) {
	for _, t := range trends {
		// 查重：简单通过 Title 查重
		var count int64
		mysql.DB.Model(&model.MarketTrend{}).Where("title = ?", t.Title).Count(&count)
		if count > 0 {
			continue
		}

		sectorsJSON, _ := json.Marshal(t.RelatedSectors)
		stocksJSON, _ := json.Marshal(t.RelatedStocks)

		trendModel := &model.MarketTrend{
			Title:              t.Title,
			Summary:            t.Summary,
			ImpactType:         t.ImpactType,
			Weight:             t.Weight,
			Source:             t.Source,
			OriginalURL:        t.Url,
			FinancialRelevance: int(t.FinancialRelevance),
			SentimentScore:     t.SentimentScore,
			RelatedSectors:     string(sectorsJSON),
			RelatedStocks:      string(stocksJSON),
			ImpactAnalysis:     t.ImpactAnalysis,
			IsStillValid:       true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		mysql.DB.Create(trendModel)
	}
}

// fetchHotTopics 抓取多源热点
func (w *TrendWorker) fetchHotTopics() []*ai.RawTrendItem {
	f, err := os.OpenFile("stock_service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	log.Println("[TrendWorker] 正在从新浪财经抓取数据...")

	var items []*ai.RawTrendItem

	// sinaNews, err := w.sinaClient.GetMarketNews(w.ctx)
	// if err != nil {
	// 	log.Printf("[TrendWorker] 抓取新浪新闻失败: %v", err)
	// } else {
	// 	for _, news := range sinaNews {
	// 		items = append(items, &ai.RawTrendItem{
	// 			Title:   news.Title,
	// 			Content: news.Content,
	// 			Source:  news.Source,
	// 			Url:     news.Url,
	// 		})
	// 	}
	// }

	// 抓取快讯
	// liveNews, err := w.sinaClient.Get7x24LiveNews(w.ctx)
	// if err != nil {
	// 	log.Printf("[TrendWorker] 抓取新浪快讯失败: %v", err)
	// } else {
	// 	for _, news := range liveNews {
	// 		items = append(items, &ai.RawTrendItem{
	// 			Title:   news.Title,
	// 			Content: news.Content,
	// 			Source:  news.Source,
	// 			Url:     news.Url,
	// 		})
	// 	}
	// }

	// 抓取今日头条热搜
	// toutiaoNews, err := crawler.GetToutiaoHotTrends()
	// if err != nil {
	// 	log.Printf("[TrendWorker] 抓取头条热搜失败: %v", err)
	// } else {
	// 	for _, news := range toutiaoNews {
	// 		items = append(items, &ai.RawTrendItem{
	// 			Title:   news.Title,
	// 			Content: news.Content,
	// 			Source:  news.Source,
	// 			Url:     news.Url,
	// 		})
	// 	}
	// }

	// 抓取华尔街见闻
	// wscnNews, err := crawler.GetWallstreetCNHotTrends("day")
	// if err != nil {
	// 	log.Printf("[TrendWorker] 抓取华尔街见闻失败: %v", err)
	// } else {
	// 	for _, news := range wscnNews {
	// 		items = append(items, &ai.RawTrendItem{
	// 			Title:   news.Title,
	// 			Content: news.Content,
	// 			Source:  news.Source,
	// 			Url:     news.Url,
	// 		})
	// 	}
	// }

	// 抓取澎湃新闻
	// paperNews, err := crawler.GetThePaperHotTrends()
	// if err != nil {
	// 	log.Printf("[TrendWorker] 抓取澎湃新闻失败: %v", err)
	// } else {
	// 	for _, news := range paperNews {
	// 		items = append(items, &ai.RawTrendItem{
	// 			Title:   news.Title,
	// 			Content: news.Content,
	// 			Source:  news.Source,
	// 			Url:     news.Url,
	// 		})
	// 	}
	// }

	// 抓取百度热搜
	// baiduNews, err := crawler.GetBaiduHotTrends()
	// if err != nil {
	// 	log.Printf("[TrendWorker] 抓取百度热搜失败: %v", err)
	// } else {
	// 	for _, news := range baiduNews {
	// 		items = append(items, &ai.RawTrendItem{
	// 			Title:   news.Title,
	// 			Content: news.Content,
	// 			Source:  news.Source,
	// 			Url:     news.Url,
	// 		})
	// 	}
	// }

	// 抓取财联社电报
	clsNews, err := crawler.GetCailianPressTelegraph()
	if err != nil {
		log.Printf("[TrendWorker] 抓取财联社电报失败: %v", err)
	} else {
		for _, news := range clsNews {
			items = append(items, &ai.RawTrendItem{
				Title:   news.Title,
				Content: news.Content,
				Source:  news.Source,
				Url:     news.Url,
			})
		}
	}

	log.Printf("[TrendWorker] 成功抓取 %d 条新闻(含多渠道聚合)", len(items))
	return items
}
