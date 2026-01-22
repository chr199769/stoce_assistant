package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	eastmoney "stock_assistant/backend/common/eastmoney"
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
	// 初始化 Stock Service 客户端
	c, err := stockservice.NewClient("stock_service", client.WithHostPorts("127.0.0.1:8888"))
	if err != nil {
		klog.Errorf("创建 stock 客户端失败: %v", err)
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
	klog.Info("盘中监控已启动")
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
	klog.Info("盘中监控已停止")
}

func (s *IntradaySentinel) loop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// 立即运行一次
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
	// 1. 获取板块排行 (概念)
	sectors, err := s.emClient.GetSectorRank(ctx, "concept", 5)
	if err != nil {
		klog.Errorf("获取板块排行失败: %v", err)
		return
	}

	// 2. 检查强势板块 (涨停数或高净流入)
	for _, sector := range sectors {
		if sector.ChangePercent > 3.0 { // 阈值: 涨幅 3%
			// 发现强势板块，创建信号
			msg := fmt.Sprintf("板块异动: %s 上涨 %.2f%% 净流入 %.2f", sector.Name, sector.ChangePercent, sector.NetInflow)
			s.saveSignal(ctx, "SECTOR_MOVE", sector.Code, msg, 0)
		}
	}

	// 3. TODO: 检查涨停池变化 (需要新的 EM API 方法或解析)
	// 目前仅记录头部板块
	if len(sectors) > 0 {
		top := sectors[0]
		klog.Infof("领涨板块: %s (+%.2f%%)", top.Name, top.ChangePercent)
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
		klog.Errorf("保存信号失败: %v", err)
	} else {
		klog.Infof("信号已保存: [%s] %s", signalType, message)
	}
}
