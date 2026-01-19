package model

import "time"

type UserWatchlist struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    string    `gorm:"index;type:varchar(64)"` // String ID for flexibility (e.g. UUID)
	StockCode string    `gorm:"index;type:varchar(10)"`
	Tags      string    `gorm:"type:json"` // JSON array of tags: ["Bullish", "HighVol"]
	CreatedAt time.Time
	UpdatedAt time.Time
}
