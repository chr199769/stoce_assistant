package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"stock_assistant/backend/ai_service/biz/tool/eastmoney"
	"stock_assistant/backend/ai_service/kitex_gen/stock"
	"stock_assistant/backend/ai_service/kitex_gen/stock/stockservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
)

type IntradaySentinel struct {
	stockClient stockservice.Client
	emClient    *eastmoney.Client
	running     bool
	mu          sync.Mutex
	stopChan    chan struct{}
}

func NewIntradaySentinel() *IntradaySentinel {
	// Initialize Stock Service Client
	c, err := stockservice.NewClient("stock_service", client.WithHostPorts("127.0.0.1:8888"))
	if err != nil {
		klog.Errorf("Failed to create stock client: %v", err)
	}

	return &IntradaySentinel{
		stockClient: c,
		emClient:    eastmoney.NewClient(),
		stopChan:    make(chan struct{}),
	}
}

func (s *IntradaySentinel) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go s.loop()
	klog.Info("Intraday Sentinel started")
}

func (s *IntradaySentinel) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()
	klog.Info("Intraday Sentinel stopped")
}

func (s *IntradaySentinel) loop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Run immediately once
	s.scanMarket()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.scanMarket()
		}
	}
}

func (s *IntradaySentinel) scanMarket() {
	ctx := context.Background()
	// 1. Get Sector Ranks (Concept)
	sectors, err := s.emClient.GetSectorRank(ctx, "concept", 5)
	if err != nil {
		klog.Errorf("Failed to get sector rank: %v", err)
		return
	}

	// 2. Check for strong sectors (Limit Up count, or high net inflow)
	for _, sector := range sectors {
		if sector.ChangePercent > 3.0 { // Threshold: 3% rise
			// Found a strong sector, create signal
			msg := fmt.Sprintf("Sector Alert: %s is up %.2f%% with Net Inflow %.2f", sector.Name, sector.ChangePercent, sector.NetInflow)
			s.saveSignal(ctx, "SECTOR_MOVE", sector.Code, msg, 0)
		}
	}

	// 3. TODO: Check for Limit Up Pool changes (requires new EM API method or parsing)
	// For now, let's just log the top sector
	if len(sectors) > 0 {
		top := sectors[0]
		klog.Infof("Top Sector: %s (+%.2f%%)", top.Name, top.ChangePercent)
	}
}

func (s *IntradaySentinel) saveSignal(ctx context.Context, signalType, code, message string, score float64) {
	if s.stockClient == nil {
		return
	}

	req := &stock.SaveIntradaySignalRequest{
		Signal: &stock.IntradaySignal{
			SignalType:  signalType,
			StockCode:   code,
			Description: message,
			Score:       score,
			TriggerTime: time.Now().Format(time.RFC3339),
		},
	}

	_, err := s.stockClient.SaveIntradaySignal(ctx, req)
	if err != nil {
		klog.Errorf("Failed to save signal: %v", err)
	} else {
		klog.Infof("Signal Saved: [%s] %s", signalType, message)
	}
}
