package main

import (
	"fmt"
	"log"

	"stock_assistant/backend/stock_service/dal/model"
	"stock_assistant/backend/stock_service/dal/mysql"

	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env from various locations relative to script execution
	// We assume we might run this from backend/stock_service/
	_ = godotenv.Load(".env")
	_ = godotenv.Load("backend/stock_service/.env")

	fmt.Println("Initializing MySQL connection...")
	// Initialize DB
	mysql.Init()

	if mysql.DB == nil {
		log.Fatal("Database connection failed")
	}

	db := mysql.DB
    
    // Criteria:
    // impact_scope != 'market_wide'
    // AND (related_sectors is empty/null)
    // AND (related_stocks is empty/null)
    
    // Note on JSON fields:
    // Empty might be stored as:
    // 1. NULL
    // 2. Empty string ""
    // 3. JSON empty array "[]"
    // 4. JSON null "null" (less likely but possible)

    // Debug: Print first 10 records to inspect data
    var trends []model.MarketTrend
    if err := db.Limit(10).Find(&trends).Error; err == nil {
        fmt.Println("--- Sample Data ---")
        for _, t := range trends {
            fmt.Printf("ID: %d, Scope: %s, Sectors: %s, Stocks: %s\n", t.ID, t.ImpactScope, t.RelatedSectors, t.RelatedStocks)
        }
        fmt.Println("-------------------")
    }

    scopeCondition := "impact_scope != 'market_wide'"
    
    // We want to delete if BOTH sectors and stocks are effectively empty
    // Effective empty means: NULL, "", "[]", "null"
    
    sectorCondition := "(related_sectors IS NULL OR related_sectors = '' OR related_sectors = '[]' OR related_sectors = 'null')"
    stockCondition := "(related_stocks IS NULL OR related_stocks = '' OR related_stocks = '[]' OR related_stocks = 'null')"

    // 1. Count
    var count int64
    query := db.Model(&model.MarketTrend{}).
        Where(scopeCondition).
        Where(sectorCondition).
        Where(stockCondition)
    
    if err := query.Count(&count).Error; err != nil {
        log.Fatalf("Failed to count records: %v", err)
    }

    fmt.Printf("Found %d invalid trends matching criteria:\n", count)
    fmt.Println(" - Impact Scope != 'market_wide'")
    fmt.Println(" - Related Sectors is empty")
    fmt.Println(" - Related Stocks is empty")

    if count == 0 {
        fmt.Println("No records to clean based on empty fields.")
    } else {
        // 2. Delete
        // MarketTrend doesn't have DeletedAt, so this is a hard delete.
        result := query.Delete(&model.MarketTrend{})
        if result.Error != nil {
            log.Fatalf("Failed to delete records: %v", result.Error)
        }
        fmt.Printf("Successfully deleted %d records.\n", result.RowsAffected)
    }

    // 3. Clean up garbled titles (UTF-8 truncation artifacts)
    fmt.Println("Checking for garbled titles (containing '')...")
    var garbledCount int64
    garbledQuery := db.Model(&model.MarketTrend{}).Where("title LIKE ?", "%%")
    
    if err := garbledQuery.Count(&garbledCount).Error; err != nil {
        log.Printf("Failed to count garbled records: %v", err)
    } else {
        fmt.Printf("Found %d records with garbled titles.\n", garbledCount)
        if garbledCount > 0 {
            res := garbledQuery.Delete(&model.MarketTrend{})
            if res.Error != nil {
                log.Printf("Failed to delete garbled records: %v", res.Error)
            } else {
                fmt.Printf("Successfully deleted %d garbled records.\n", res.RowsAffected)
            }
        }
    }
}
