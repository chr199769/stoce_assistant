package main

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/kitex/client"
	"stock_assistant/backend/ai_service/kitex_gen/ai"
	"stock_assistant/backend/ai_service/kitex_gen/ai/aiservice"
)

func main() {
	log.Println("Starting verification client...")
	
	// Create Client
	c, err := aiservice.NewClient("ai_service", client.WithHostPorts("localhost:8889"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	testMarketReview(c)
	testPrediction(c)
}

func testMarketReview(c aiservice.Client) {
	log.Println("--- Testing MarketReview ---")
	req := &ai.MarketReviewRequest{
		Date:         time.Now().Format("2006-01-02"),
		FocusSectors: []string{"Semiconductor", "AI"},
	}

	log.Printf("Calling MarketReview for date: %s", req.Date)
	
	_, err := c.MarketReview(context.Background(), req)
	if err != nil {
		log.Printf("Error calling MarketReview: %v", err)
		return
	}

	log.Println("Successfully received MarketReview response!")
	// log.Printf("Summary: %s", resp.Summary)
}

func testPrediction(c aiservice.Client) {
	log.Println("--- Testing GetPrediction ---")
	req := &ai.GetPredictionRequest{
		Code:        "600519", // Moutai
		Days:        3,
		IncludeNews: true,
	}
	
	log.Printf("Calling GetPrediction for code: %s", req.Code)

	resp, err := c.GetPrediction(context.Background(), req)
	if err != nil {
		log.Printf("Error calling GetPrediction: %v", err)
		return
	}
	
	log.Println("Successfully received GetPrediction response!")
	if resp.Result_ != nil {
		log.Printf("Confidence: %.2f", resp.Result_.Confidence)
		log.Printf("Analysis: %s...", resp.Result_.Analysis[:100]) // Print first 100 chars
	}
}
