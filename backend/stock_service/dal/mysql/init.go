package mysql

import (
	"fmt"
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
	
	dsn := "root:root@tcp(127.0.0.1:3306)/stock_assistant?charset=utf8mb4&parseTime=True&loc=Local"
	
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Warning: Failed to connect to MySQL: %v. Persistence will be disabled.\n", err)
		return
	}

	// Auto Migrate
	err = DB.AutoMigrate(
		&model.MarketSectorDaily{},
		&model.MarketLimitUpSummary{},
		&model.MarketLimitUpDetail{},
	)
	if err != nil {
		fmt.Printf("Warning: Failed to auto migrate: %v\n", err)
	}
}
