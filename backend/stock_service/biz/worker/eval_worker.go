package worker

import (
	"context"
	"log"
	"math"
	"time"

	"stock_assistant/backend/common/eastmoney"
	"stock_assistant/backend/common/langfuse"
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
	rawIdx := -1
	for i, k := range klines {
		if k.Date >= predDateStr {
			// Find the closest trading day >= prediction date
			rawIdx = i
			break
		}
	}

	if rawIdx == -1 {
		// Prediction date is too new or not in history yet?
		// Or history too short.
		return
	}

	// Calculate effective base index
	// If prediction was made after 15:00, the effective T=0 is the NEXT trading day.
	// So baseIdx = rawIdx + 1
	baseIdx := rawIdx
	if eval.PredictionDate.Hour() >= 15 {
		baseIdx = rawIdx + 1
	}

	updated := false

	// We map relative days (1, 2, 3) to klines array indices
	// If baseIdx corresponds to the effective T=0, then baseIdx+1 is T+1.

	if baseIdx+1 < len(klines) {
		eval.Price1D = klines[baseIdx+1].Close
		updated = true
	}
	if baseIdx+2 < len(klines) {
		eval.Price2D = klines[baseIdx+2].Close
		updated = true
	}
	if baseIdx+3 < len(klines) {
		eval.Price3D = klines[baseIdx+3].Close
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

	// Actual change in percentage (e.g., 5.0 for 5%)
	actualChange := (eval.Price3D - eval.InitialPrice) / eval.InitialPrice * 100

	// Get prediction record
	var pred model.PredictionRecord
	mysql.DB.First(&pred, "id = ?", eval.PredictionID)

	score := 0.0

	// Check if we should use legacy scoring (if PredictedChange is 0 but Trend is set)
	// We treat 0 PredictedChange with "Up"/"Down" trend as legacy.
	usingLegacy := false
	if pred.PredictedChange == 0 && (pred.Trend == "Up" || pred.Trend == "Down" || pred.Trend == "看涨" || pred.Trend == "看跌") {
		usingLegacy = true
	}

	if usingLegacy {
		roi := (eval.Price3D - eval.InitialPrice) / eval.InitialPrice
		score = 50.0
		if pred.Trend == "看涨" || pred.Trend == "Up" {
			if roi > 0 {
				score += 20
				score += roi * 100
			} else {
				score -= 20
				score += roi * 100
			}
		} else if pred.Trend == "看跌" || pred.Trend == "Down" {
			if roi < 0 {
				score += 20
				score -= roi * 100
			} else {
				score -= 20
				score -= roi * 100
			}
		}
	} else {
		// New Logic: 50pts for direction + 50pts for accuracy

		// 1. Direction Correct
		sameDirection := false
		if pred.PredictedChange > 0 && actualChange > 0 {
			sameDirection = true
		} else if pred.PredictedChange < 0 && actualChange < 0 {
			sameDirection = true
		} else if math.Abs(pred.PredictedChange) < 1e-6 && math.Abs(actualChange) < 1e-6 {
			sameDirection = true
		}

		if sameDirection {
			score += 50
		}

		// 2. Accuracy Score: 50 * (1 - |pred - actual| / |actual|)
		diff := math.Abs(pred.PredictedChange - actualChange)
		var accuracyRatio float64

		absActual := math.Abs(actualChange)
		if absActual > 1e-6 {
			accuracyRatio = 1.0 - (diff / absActual)
		} else {
			// If actual change is ~0, penalize by absolute difference
			accuracyRatio = 1.0 - diff
		}

		if accuracyRatio < 0 {
			accuracyRatio = 0
		}

		score += 50 * accuracyRatio
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}
