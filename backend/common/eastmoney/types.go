package eastmoney

import (
	"fmt"
	"strings"
)

// FinancialData represents financial report data
type FinancialData struct {
	ReportDate   string
	TotalRevenue float64
	NetProfit    float64
	Eps          float64
	RevenueYoy   float64
	ProfitYoy    float64
}

// SectorInfo represents a market sector (industry or concept)
type SectorInfo struct {
	Code          string
	Name          string
	ChangePercent float64
	NetInflow     float64
	TopStockName  string
	TopStockCode  string
}

// SectorStockItem represents a stock within a sector
type SectorStockItem struct {
	Code          string
	Name          string
	Price         float64
	ChangePercent float64
	Volume        int64
	Amount        float64
	MarketCap     float64
}

type AStockItem struct {
	Code string
	Name string
}

// DragonTigerItem represents a stock on the Dragon Tiger List
type DragonTigerItem struct {
	Date          string
	Code          string
	Name          string
	ClosePrice    float64
	ChangePercent float64
	Reason        string
	NetInflow     float64
	BuySeats      []*DragonTigerSeat
	SellSeats     []*DragonTigerSeat
}

// DragonTigerSeat represents a trading seat on the Dragon Tiger List
type DragonTigerSeat struct {
	Name    string
	BuyAmt  float64
	SellAmt float64
	NetAmt  float64
	Tags    []string
}

// KlineItem represents a single K-line bar
type KlineItem struct {
	Date          string
	Open          float64
	Close         float64
	High          float64
	Low           float64
	Volume        int64
	ChangePercent float64
}

type StockNewsItem struct {
	Title    string
	ShowTime string
	Url      string
	Summary  string
}

type StockHeatData struct {
	Rank int
	Heat int
}

type NoticeItem struct {
	Title string
	Date  string
	Url   string
}

type OrderBookData struct {
	Buy1Price  float64
	Buy1Vol    int
	Buy2Price  float64
	Buy2Vol    int
	Buy3Price  float64
	Buy3Vol    int
	Buy4Price  float64
	Buy4Vol    int
	Buy5Price  float64
	Buy5Vol    int
	Sell1Price float64
	Sell1Vol   int
	Sell2Price float64
	Sell2Vol   int
	Sell3Price float64
	Sell3Vol   int
	Sell4Price float64
	Sell4Vol   int
	Sell5Price float64
	Sell5Vol   int
	WeiBi      float64
	WeiCha     float64
}

type ChipDistributionData struct {
	AverageCost float64
	WinnerRate  float64
	Cost90Low   float64
	Cost90High  float64
}

type IndustryIndexData struct {
	IndustryName string
	RegionName   string
	ConceptNames string
}

// RealtimeQuoteData 表示东财实时行情主要字段
type RealtimeQuoteData struct {
	Code       string
	Name       string
	Current    float64
	PrevClose  float64
	Open       float64
	High       float64
	Low        float64
	Volume     int64
	Amount     float64
	Timestamp  string
}

// String methods for formatting
func (d *StockHeatData) String() string {
	return fmt.Sprintf("Guba Rank: %d, Heat: %d", d.Rank, d.Heat)
}

func (d *IndustryIndexData) String() string {
	return fmt.Sprintf("Industry: %s, Region: %s, Concepts: %s", d.IndustryName, d.RegionName, d.ConceptNames)
}

func (d *ChipDistributionData) String() string {
	return fmt.Sprintf("AvgCost: %.2f, WinnerRate: %.2f%%, 90%%CostRange: %.2f-%.2f",
		d.AverageCost, d.WinnerRate, d.Cost90Low, d.Cost90High)
}

func (d *NoticeItem) String() string {
	return fmt.Sprintf("[%s] %s", d.Date, d.Title)
}

func (d *StockNewsItem) String() string {
	return fmt.Sprintf("[%s] %s", d.ShowTime, d.Title)
}

func (d *OrderBookData) String() string {
	var sb strings.Builder
	if d.WeiBi != 0 || d.WeiCha != 0 {
		sb.WriteString(fmt.Sprintf("Intraday Pressure: WeiBi(%.2f%%), WeiCha(%.0f) | ", d.WeiBi, d.WeiCha))
	}
	if d.Buy1Price > 0 || d.Sell1Price > 0 {
		sb.WriteString(fmt.Sprintf("Buy1: %.2f(%d), Sell1: %.2f(%d)",
			d.Buy1Price, d.Buy1Vol,
			d.Sell1Price, d.Sell1Vol))
	} else {
		sb.WriteString("Order Book Snapshot: Unavailable (Market Closed or Level-1 Restricted)")
	}
	return sb.String()
}

func (d *DragonTigerItem) String() string {
	return fmt.Sprintf("[%s] %s | Change: %.2f%% | NetBuy: %.0fWan | Reason: %s",
		d.Date, d.Name, d.ChangePercent, d.NetInflow/10000, d.Reason)
}
