package mysql

import (
	"fmt"
	"time"

	"stock_assistant/backend/stock_service/config"
	"stock_assistant/backend/stock_service/dal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	// In a real app, DSN should come from config
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// For this demo/prototype, we'll use a local DSN but wrap in try-catch logic
	// or just log error if connection fails, so we don't block service startup if DB isn't there.

	dsn := ""
	cfg := config.Get()
	if cfg != nil && cfg.Database != nil {
		dsn = cfg.Database.MySQLDSN
	}
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/stock_assistant?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				return time.Now()
			}
			return time.Now().In(loc)
		},
	})
	if err != nil {
		fmt.Printf("Warning: Failed to connect to MySQL: %v. Persistence will be disabled.\n", err)
		return
	}

	// Auto Migrate
	err = DB.AutoMigrate(
		&model.MarketSectorDaily{},
		&model.MarketLimitUpSummary{},
		&model.MarketLimitUpDetail{},
		&model.UserWatchlist{},
		&model.IntradaySignal{},
		&model.User{},
		&model.PredictionRecord{},
		&model.EvaluationRecord{},
	)
	if err != nil {
		fmt.Printf("Warning: Failed to auto migrate: %v\n", err)
	}
}
