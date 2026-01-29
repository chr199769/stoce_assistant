package main

import (
	"fmt"
	"time"

	"stock_assistant/backend/stock_service/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := config.Init(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		return
	}
	dsn := ""
	cfg := config.Get()
	if cfg != nil && cfg.Database != nil {
		dsn = cfg.Database.MySQLDSN
	}
	if dsn == "" {
		dsn = "root:12345678@tcp(127.0.0.1:3306)/stock_assistant?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			loc, err := time.LoadLocation("Asia/Shanghai")
			if err != nil {
				return time.Now()
			}
			return time.Now().In(loc)
		},
	})
	if err != nil {
		fmt.Printf("连接 MySQL 失败: %v\n", err)
		return
	}

	result := db.Exec("DELETE FROM market_trends")
	if result.Error != nil {
		fmt.Printf("清理 market_trends 失败: %v\n", result.Error)
		return
	}
	fmt.Printf("清理 market_trends 完成，影响行数: %d\n", result.RowsAffected)
}
