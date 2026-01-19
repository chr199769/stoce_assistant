package model

import "time"

type IntradaySignal struct {
	ID          uint      `gorm:"primaryKey"`
	StockCode   string    `gorm:"index;type:varchar(10)"`
	SignalType  string    `gorm:"index;type:varchar(32)"` // e.g., "VolumeSpike", "SectorRotation"
	TraceID     string    `gorm:"index;type:varchar(64)"` // Link to Langfuse Trace
	Score       float64   `gorm:"type:decimal(5,2)"`      // 0-100 score
	Description string    `gorm:"type:text"`
	TriggerTime time.Time `gorm:"index"`
	CreatedAt   time.Time
}
