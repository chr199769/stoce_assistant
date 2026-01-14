package model

import "time"

type MarketSectorDaily struct {
	ID            uint      `gorm:"primaryKey"`
	Date          string    `gorm:"index;type:varchar(10)"` // YYYY-MM-DD
	Time          string    `gorm:"type:varchar(8)"`        // HH:MM:SS
	SectorCode    string    `gorm:"index;type:varchar(20)"`
	Name          string    `gorm:"type:varchar(50)"`
	ChangePercent float64   `gorm:"type:decimal(10,2)"`
	NetInflow     float64   `gorm:"type:decimal(20,2)"`
	TopStockName  string    `gorm:"type:varchar(50)"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type MarketLimitUpSummary struct {
	ID             uint      `gorm:"primaryKey"`
	Date           string    `gorm:"uniqueIndex;type:varchar(10)"`
	LimitUpCount   int       `gorm:"type:int"`
	BrokenCount    int       `gorm:"type:int"`
	HighestBoard   int       `gorm:"type:int"` // 最高连板数
	SentimentScore float64   `gorm:"type:decimal(5,2)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type MarketLimitUpDetail struct {
	ID          uint      `gorm:"primaryKey"`
	Date        string    `gorm:"index;type:varchar(10)"`
	StockCode   string    `gorm:"index;type:varchar(10)"`
	StockName   string    `gorm:"type:varchar(20)"`
	BoardCount  int       `gorm:"type:int"` // 连板数
	LimitUpType string    `gorm:"type:varchar(20)"` // 首板, 2连板
	Reason      string    `gorm:"type:varchar(100)"`
	IsBroken    bool      `gorm:"type:boolean"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
