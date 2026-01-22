package model

import (
	"time"

	"gorm.io/gorm"
)

type PredictionRecord struct {
	ID             string `gorm:"primaryKey"`
	StockCode      string `gorm:"index"`
	PredictionDate time.Time
	Content        string `gorm:"type:text"`
	Confidence     float64
	Trend          string
	TargetPrice    float64
	StopLossPrice  float64
	TraceID        string `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type EvaluationRecord struct {
	ID             string `gorm:"primaryKey"`
	PredictionID   string `gorm:"index"`
	StockCode      string
	StockName      string
	PredictionDate time.Time
	InitialPrice   float64
	Price1D        float64
	Price2D        float64
	Price3D        float64
	Score          float64
	Status         string // "pending", "completed"
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
