namespace go api

struct RealtimeResponse {
    1: string code
    2: string name
    3: double current_price
    4: double change_percent
    5: i64 volume
    6: string timestamp
}

struct GetRealtimeRequest {
    1: string code (api.path="code")
}

struct PredictionRequest {
    1: string code (api.path="code")
    2: i32 days (api.body="days")
    3: bool include_news (api.body="include_news")
    4: string model (api.body="model")
}

struct PredictionResponse {
    1: string code
    2: double confidence
    3: string analysis
    4: string news_summary
}

struct ImageRecognitionRequest {
}

struct RecognizedStock {
    1: string code
    2: string name
}

struct ImageRecognitionResponse {
    1: list<RecognizedStock> stocks
}

struct FinancialData {
    1: string report_date
    2: double total_revenue
    3: double net_profit
    4: double eps
    5: double revenue_yoy
    6: double profit_yoy
}

struct GetFinancialReportRequest {
    1: string code (api.path="code")
}

struct GetFinancialReportResponse {
    1: list<FinancialData> reports
}

struct MarketReviewRequest {
    1: string date (api.body="date") // Optional: YYYY-MM-DD
    2: list<string> focus_sectors (api.body="focus_sectors") // Optional
}

struct MarketReviewResponse {
    1: string summary
    2: string sector_analysis
    3: string sentiment_analysis
    4: list<string> key_risks
    5: list<string> opportunities
}

struct MarketAnalysisRequest {
    1: string date (api.body="date")
}

struct MarketAnalysisResponse {
    1: list<string> hot_stocks
    2: list<string> recommended_stocks
    3: list<string> risks
    4: list<string> opportunities
    5: string analysis_summary
}

service StockAPI {
    RealtimeResponse GetRealtime(1: GetRealtimeRequest req) (api.get="/api/stocks/:code/realtime")
    PredictionResponse GetPrediction(1: PredictionRequest req) (api.post="/api/prediction/:code")
    ImageRecognitionResponse RecognizeStockImage(1: ImageRecognitionRequest req) (api.post="/api/image/recognize")
    GetFinancialReportResponse GetFinancialReport(1: GetFinancialReportRequest req) (api.get="/api/stocks/:code/financial")
    MarketReviewResponse MarketReview(1: MarketReviewRequest req) (api.post="/api/market/review")
    MarketAnalysisResponse AnalyzeMarket(1: MarketAnalysisRequest req) (api.post="/api/market/analysis")
    
    // Phase 2: Sector Details & Dragon Tiger List
    GetSectorStocksResponse GetSectorStocks(1: GetSectorStocksRequest req) (api.get="/api/stock/sector/stocks")
    GetDragonTigerListResponse GetDragonTigerList(1: GetDragonTigerListRequest req) (api.get="/api/stock/dragontiger/list")
}

struct SectorStockItem {
    1: string code
    2: string name
    3: double price
    4: double change_percent
    5: i64 volume
    6: double amount
    7: double market_cap
}

struct GetSectorStocksRequest {
    1: string sector_code (api.query="sector_code")
}

struct GetSectorStocksResponse {
    1: list<SectorStockItem> stocks
}

struct DragonTigerSeat {
    1: string name
    2: double buy_amt
    3: double sell_amt
    4: double net_amt
    5: list<string> tags
}

struct DragonTigerItem {
    1: string code
    2: string name
    3: double close_price
    4: double change_percent
    5: string reason
    6: double net_inflow
    7: list<DragonTigerSeat> buy_seats
    8: list<DragonTigerSeat> sell_seats
}

struct GetDragonTigerListRequest {
    1: string date (api.query="date")
}

struct GetDragonTigerListResponse {
    1: list<DragonTigerItem> items
}
