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
    6: double sentiment_score // 0-100
    7: double policy_score    // -5 to +5
}

// --- Phase 2 Additions ---

struct SectorInfo {
    1: string code
    2: string name
    3: double change_percent
    4: double net_inflow
    5: string top_stock_name
    6: string top_stock_code
    7: string type // concept/industry
}

struct GetMarketSectorsRequest {
    1: string type (api.query="type") // concept, industry, region
    2: i32 limit (api.query="limit")
}

struct GetMarketSectorsResponse {
    1: list<SectorInfo> sectors
}

struct LimitUpStock {
    1: string code
    2: string name
    3: double price
    4: double change_percent
    5: string limit_up_type
    6: string reason
    7: bool is_broken
}

struct GetLimitUpPoolRequest {
    1: string date (api.query="date")
}

struct GetLimitUpPoolResponse {
    1: list<LimitUpStock> stocks
}

struct Kline {
    1: string date
    2: double open
    3: double close
    4: double high
    5: double low
    6: i64 volume
}

struct GetHistoricalKlineRequest {
    1: string stock_code (api.query="code")
    2: i32 days (api.query="days")
}

struct GetHistoricalKlineResponse {
    1: list<Kline> klines
}

struct EvaluationRecord {
    1: string id
    2: string prediction_id
    3: string stock_code
    4: string prediction_date
    5: double initial_price
    6: double price_1d
    7: double price_2d
    8: double price_3d
    9: double score
    10: string status
}

struct GetEvaluationsRequest {
    1: string stock_code (api.query="code")
    2: i32 limit (api.query="limit")
    3: i32 offset (api.query="offset")
}

struct GetEvaluationsResponse {
    1: list<EvaluationRecord> evaluations
}

service StockAPI {User {
    1: string id
    2: string username
    3: string created_at
}

struct GetOrCreateUserRequest {
    1: string username (api.body="username")
}

struct GetOrCreateUserResponse {
    1: User user
}

struct WatchlistItem {
    1: string stock_code
    2: list<string> tags
    3: string added_at
}

struct AddWatchlistRequest {
    1: string user_id (api.body="user_id")
    2: string stock_code (api.body="stock_code")
}

struct AddWatchlistResponse {
    1: bool success
}

struct GetWatchlistRequest {
    1: string user_id (api.query="user_id")
}

struct GetWatchlistResponse {
    1: list<WatchlistItem> items
}

struct RemoveWatchlistRequest {
    1: string user_id (api.body="user_id")
    2: string stock_code (api.body="stock_code")
}

struct RemoveWatchlistResponse {
    1: bool success
}

// Reuse existing structs
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

    // Phase 2 New: Market Overview
    GetMarketSectorsResponse GetMarketSectors(1: GetMarketSectorsRequest req) (api.get="/api/market/sectors")
    GetLimitUpPoolResponse GetLimitUpPool(1: GetLimitUpPoolRequest req) (api.get="/api/market/limit_up")
    
    // Phase 2 New: User & Watchlist
    GetOrCreateUserResponse GetOrCreateUser(1: GetOrCreateUserRequest req) (api.post="/api/user/login")
    AddWatchlistResponse AddWatchlist(1: AddWatchlistRequest req) (api.post="/api/watchlist/add")
    GetWatchlistResponse GetWatchlist(1: GetWatchlistRequest req) (api.get="/api/watchlist/list")
    RemoveWatchlistResponse RemoveWatchlist(1: RemoveWatchlistRequest req) (api.post="/api/watchlist/remove")
    
    // Phase 2 New: Fractal Data
    GetHistoricalKlineResponse GetHistoricalKline(1: GetHistoricalKlineRequest req) (api.get="/api/stock/kline")

    // Phase 3: Prediction Evaluation
    GetEvaluationsResponse GetEvaluations(1: GetEvaluationsRequest req) (api.get="/api/evaluations")
    DeleteEvaluationResponse DeleteEvaluation(1: DeleteEvaluationRequest req) (api.delete="/api/evaluations/:id")
}
