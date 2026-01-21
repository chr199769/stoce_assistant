package worker

import (
	"context"
	"log"
	"strings"

	// "os"

	"time"

	"stock_assistant/backend/stock_service/biz/provider/langfuse"
	"stock_assistant/backend/stock_service/biz/provider/sina"
	"stock_assistant/backend/stock_service/dal/model"
	"stock_assistant/backend/stock_service/dal/mysql"
)

type EvalWorker struct {
	lf *langfuse.LangfuseManager
}

func NewEvalWorker() *EvalWorker {
	// Initialize Langfuse client
	return &EvalWorker{
		lf: langfuse.GetLangfuse(),
	}
}

func (w *EvalWorker) Start() {
	go func() {
		// Wait for DB init
		time.Sleep(5 * time.Second)
		log.Println("[EvalWorker] Worker started")
		ticker := time.NewTicker(1 * time.Hour)
		// Initial run
		w.Evaluate()

		for range ticker.C {
			w.Evaluate()
		}
	}()
}

func (w *EvalWorker) Evaluate() {
	log.Println("[EvalWorker] Starting evaluation cycle...")
	if mysql.DB == nil {
		log.Println("[EvalWorker] DB not initialized")
		return
	}

	var signals []model.IntradaySignal

	// Find signals created > 1 day ago and not yet scored (Score = 0)
	yesterday := time.Now().Add(-24 * time.Hour)
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)

	if err := mysql.DB.Where("created_at < ? AND created_at > ? AND score = 0", yesterday, weekAgo).Find(&signals).Error; err != nil {
		log.Printf("[EvalWorker] Error fetching signals: %v", err)
		return
	}

	if len(signals) == 0 {
		log.Println("[EvalWorker] No pending signals to evaluate")
		return
	}

	log.Printf("[EvalWorker] Found %d signals to evaluate", len(signals))

	sinaClient := sina.NewClient()

	for _, sig := range signals {
		if strings.HasPrefix(sig.StockCode, "BK") {
			log.Printf("[EvalWorker] Skipping sector signal %s (not supported for evaluation yet)", sig.StockCode)
			sig.Score = -1.0
			if err := mysql.DB.Save(&sig).Error; err != nil {
				log.Printf("[EvalWorker] Failed to update skipped signal %d: %v", sig.ID, err)
			}
			continue
		}

		// Get current price
		info, err := sinaClient.GetStockInfo(context.Background(), sig.StockCode)
		if err != nil {
			log.Printf("[EvalWorker] Error getting quote for %s: %v", sig.StockCode, err)
			continue
		}

		// Determine outcome.
		// Heuristic: If ChangePercent > 0, score = 1.0 (100%). Else 0.
		var score float64
		if info.ChangePercent > 0 {
			score = 1.0
		} else {
			score = 0.0
		}

		// Update DB
		sig.Score = score * 100
		if err := mysql.DB.Save(&sig).Error; err != nil {
			log.Printf("[EvalWorker] Failed to save score for signal %d: %v", sig.ID, err)
			continue
		}

		// Send to Langfuse
		if w.lf != nil && sig.TraceID != "" {
			err := w.lf.Score(context.Background(), sig.TraceID, "accuracy", score, "Automated evaluation by EvalWorker (Sina Finance)")
			if err != nil {
				log.Printf("[EvalWorker] Failed to send score to Langfuse: %v", err)
			}
		}
	}
}
