package worker

import (
	"context"
	"log"
	"time"

	"stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/stock_service/biz/provider/langfuse"
	"stock_assistant/backend/stock_service/dal/model"
	"stock_assistant/backend/stock_service/dal/mysql"
)

type EvalWorker struct {
	lf              *langfuse.LangfuseManager
	eastMoneyClient *eastmoney.Client
}

func NewEvalWorker() *EvalWorker {
	return &EvalWorker{
		lf:              langfuse.GetLangfuse(),
		eastMoneyClient: eastmoney.NewClient(),
	}
}

func (w *EvalWorker) Start() {
	go func() {
		time.Sleep(10 * time.Second) // Wait for DB
		log.Println("[EvalWorker] Worker started")
		ticker := time.NewTicker(1 * time.Hour)
		w.Evaluate()
		for range ticker.C {
			w.Evaluate()
		}
	}()
}

func (w *EvalWorker) Evaluate() {
	if mysql.DB == nil {
		return
	}

	var evals []model.EvaluationRecord
	if err := mysql.DB.Where("status = ?", "pending").Find(&evals).Error; err != nil {
		log.Printf("[EvalWorker] Error fetching evaluations: %v", err)
		return
	}

	if len(evals) == 0 {
		return
	}

	log.Printf("[EvalWorker] Processing %d evaluations", len(evals))

	for _, eval := range evals {
		w.processEvaluation(&eval)
	}
}

func (w *EvalWorker) processEvaluation(eval *model.EvaluationRecord) {
	// Fetch recent klines (e.g. last 10 days)
	klines, err := w.eastMoneyClient.GetKlineHistory(context.Background(), eval.StockCode, 10)
	if err != nil {
		log.Printf("Failed to get klines for %s: %v", eval.StockCode, err)
		return
	}

	// Find the prediction date index
	predDateStr := eval.PredictionDate.Format("2006-01-02")
	predIdx := -1
	for i, k := range klines {
		if k.Date >= predDateStr {
			// Find the closest trading day >= prediction date
			predIdx = i
			break
		}
	}

	if predIdx == -1 {
		// Prediction date is too new or not in history yet?
		// Or history too short.
		return
	}

	updated := false

	// We map relative days (1, 2, 3) to klines array indices
	// If predIdx corresponds to the prediction date, then predIdx+1 is T+1.
	
	if predIdx+1 < len(klines) {
		eval.Price1D = klines[predIdx+1].Close
		updated = true
	}
	if predIdx+2 < len(klines) {
		eval.Price2D = klines[predIdx+2].Close
		updated = true
	}
	if predIdx+3 < len(klines) {
		eval.Price3D = klines[predIdx+3].Close
		updated = true
		
		// Finalize
		eval.Status = "completed"
		eval.Score = w.calculateScore(eval)
		
		// Send score to Langfuse
		if w.lf != nil {
			// Retrieve TraceID from PredictionRecord
			var pred model.PredictionRecord
			if err := mysql.DB.First(&pred, "id = ?", eval.PredictionID).Error; err == nil && pred.TraceID != "" {
				_ = w.lf.Score(context.Background(), pred.TraceID, "accuracy", eval.Score, "T+3 Evaluation")
			}
		}
	}
	
	if updated {
		mysql.DB.Save(eval)
	}
}

func (w *EvalWorker) calculateScore(eval *model.EvaluationRecord) float64 {
	if eval.InitialPrice == 0 {
		return 0
	}
	
	roi := (eval.Price3D - eval.InitialPrice) / eval.InitialPrice
	
	// Get prediction trend to see if we matched direction
	var pred model.PredictionRecord
	mysql.DB.First(&pred, "id = ?", eval.PredictionID)
	
	score := 50.0
	
	// Direction match bonus
	if pred.Trend == "看涨" || pred.Trend == "Up" {
		if roi > 0 {
			score += 20 // Correct direction
			score += roi * 100 // Add ROI points (e.g. 10% gain -> +10 pts)
		} else {
			score -= 20 // Wrong direction
			score += roi * 100 // Subtract loss
		}
	} else if pred.Trend == "看跌" || pred.Trend == "Down" {
		if roi < 0 {
			score += 20
			score -= roi * 100 // Positive points for negative ROI
		} else {
			score -= 20
			score -= roi * 100
		}
	}
	
	if score > 100 { score = 100 }
	if score < 0 { score = 0 }
	
	return score
}
