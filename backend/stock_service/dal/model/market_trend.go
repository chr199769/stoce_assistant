package model

import (
	"time"
)

type MarketTrend struct {
	ID                 int64     `gorm:"primaryKey" json:"id"`
	Source             string    `gorm:"size:32;not null" json:"source"`
	Title              string    `gorm:"size:255;not null" json:"title"`
	Summary            string    `gorm:"type:text" json:"summary"`
	OriginalURL        string    `gorm:"size:512" json:"original_url"`
	FinancialRelevance int       `json:"financial_relevance"`
	RelatedSectors     string    `gorm:"type:json" json:"related_sectors"` // Stored as JSON string
	RelatedStocks      string    `gorm:"type:json" json:"related_stocks"`  // Stored as JSON string
	ImpactAnalysis     string    `gorm:"type:text" json:"impact_analysis"`
	SentimentScore     float64   `json:"sentiment_score"`
	ImpactType         string    `gorm:"size:20;default:'short_term_news'" json:"impact_type"`
	ImpactScope        string    `gorm:"size:20;default:'specific'" json:"impact_scope"`
	Weight             float64   `gorm:"default:1.0" json:"weight"`
	IsStillValid       bool      `gorm:"default:true" json:"is_still_valid"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (MarketTrend) TableName() string {
	return "market_trends"
}
