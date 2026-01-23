namespace go stock

struct StockInfo {
    1: string code
    2: string name
    3: double current_price
    4: double change_percent
    5: i64 volume
    6: string timestamp
}

struct GetRealtimeRequest {
    1: string code
}

struct GetRealtimeResponse {
    1: StockInfo stock
}

struct FinancialData {
    1: string report_date
    2: double total_revenue // 营业总收入
    3: double net_profit    // 归母净利润
    4: double eps           // 每股收益
    5: double revenue_yoy   // 营收同比增长率
    6: double profit_yoy    // 净利润同比增长率
}

struct GetFinancialReportRequest {
    1: string code
}

struct GetFinancialReportResponse {
    1: list<FinancialData> reports
}

// --- New Structures for Sector & Sentiment ---

struct SectorInfo {
    1: string code
    2: string name
    3: double change_percent
    4: double net_inflow
    5: string top_stock_name
    6: string top_stock_code
    7: string type // "concept" or "industry"
}

struct GetMarketSectorsRequest {
    1: string type // "concept" or "industry", default "concept"
    2: i32 limit   // default 20
}

struct GetMarketSectorsResponse {
    1: list<SectorInfo> sectors
}

struct LimitUpStock {
    1: string code
    2: string name
    3: double price
    4: double change_percent
    5: string limit_up_type // e.g., "首板", "2连板"
    6: string reason
    7: bool is_broken
}

struct GetLimitUpPoolRequest {
    1: string date // Optional, YYYY-MM-DD
}

struct GetLimitUpPoolResponse {
    1: list<LimitUpStock> stocks
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
    1: string sector_code
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
    1: string date // YYYY-MM-DD
}

struct GetDragonTigerListResponse {
    1: list<DragonTigerItem> items
}

// --- Smart Trading Enhancement Suite ---

// User Identity
struct User {
    1: string id
    2: string username
    3: string created_at
}

struct GetOrCreateUserRequest {
    1: string username
}

struct GetOrCreateUserResponse {
    1: User user
}

// Cloud Watchlist
struct WatchlistItem {
    1: string stock_code
    2: list<string> tags
    3: string added_at
}

struct AddWatchlistRequest {
    1: string user_id
    2: string stock_code
}

struct AddWatchlistResponse {
    1: bool success
}

struct GetWatchlistRequest {
    1: string user_id
}

struct GetWatchlistResponse {
    1: list<WatchlistItem> items
}

struct RemoveWatchlistRequest {
    1: string user_id
    2: string stock_code
}

struct RemoveWatchlistResponse {
    1: bool success
}

// Intraday Signal
struct IntradaySignal {
    1: string stock_code
    2: string signal_type
    3: double score
    4: string description
    5: string trigger_time
    6: string trace_id
}

struct SaveIntradaySignalRequest {
    1: IntradaySignal signal
}

struct SaveIntradaySignalResponse {
    1: bool success
}

struct GetIntradaySignalsRequest {
    1: string date // Optional, YYYY-MM-DD
}

struct GetIntradaySignalsResponse {
    1: list<IntradaySignal> signals
}

// Fractal Prediction
struct Kline {
    1: string date
    2: double open
    3: double close
    4: double high
    5: double low
    6: i64 volume
}

struct GetHistoricalKlineRequest {
    1: string stock_code
    2: i32 days // Number of days to retrieve
}

struct GetHistoricalKlineResponse {
    1: list<Kline> klines
}

// --- Prediction Evaluation ---

struct PredictionRecord {
    1: string id
    2: string stock_code
    3: string prediction_date
    4: string content
    5: double confidence
    6: string trend // "up", "down", "neutral"
    7: double target_price // optional
    8: double stop_loss_price // optional
    9: string trace_id // Langfuse Trace ID
    10: string policy_impact_scope // "specific", "sector", "market_wide"
    11: double predicted_change // Percentage change
}

struct SavePredictionRequest {
    1: PredictionRecord record
}

struct SavePredictionResponse {
    1: bool success
}

struct EvaluationRecord {
    1: string id
    2: string prediction_id
    3: string stock_code
    4: string stock_name
    5: string prediction_date
    6: double initial_price
    7: double price_1d
    8: double price_2d
    9: double price_3d
    10: double score
    11: string status // "pending", "completed"
}

struct GetEvaluationsRequest {
    1: string stock_code
    2: i32 limit
    3: i32 offset
}

struct GetEvaluationsResponse {
    1: list<EvaluationRecord> evaluations
}

struct DeleteEvaluationRequest {
    1: string id
}

struct DeleteEvaluationResponse {
    1: bool success
}

service StockService {
    GetRealtimeResponse GetRealtime(1: GetRealtimeRequest req)
    GetFinancialReportResponse GetFinancialReport(1: GetFinancialReportRequest req)
    
    // New methods
    GetMarketSectorsResponse GetMarketSectors(1: GetMarketSectorsRequest req)
    GetLimitUpPoolResponse GetLimitUpPool(1: GetLimitUpPoolRequest req)

    // Phase 2: Sector Details & Dragon Tiger List
    GetSectorStocksResponse GetSectorStocks(1: GetSectorStocksRequest req)
    GetDragonTigerListResponse GetDragonTigerList(1: GetDragonTigerListRequest req)

    // Smart Trading Enhancement Suite
    GetOrCreateUserResponse GetOrCreateUser(1: GetOrCreateUserRequest req)
    AddWatchlistResponse AddWatchlist(1: AddWatchlistRequest req)
    GetWatchlistResponse GetWatchlist(1: GetWatchlistRequest req)
    RemoveWatchlistResponse RemoveWatchlist(1: RemoveWatchlistRequest req)
    
    SaveIntradaySignalResponse SaveIntradaySignal(1: SaveIntradaySignalRequest req)
    GetIntradaySignalsResponse GetIntradaySignals(1: GetIntradaySignalsRequest req)
    
    GetHistoricalKlineResponse GetHistoricalKline(1: GetHistoricalKlineRequest req)

    // Prediction Evaluation
    SavePredictionResponse SavePrediction(1: SavePredictionRequest req)
    GetEvaluationsResponse GetEvaluations(1: GetEvaluationsRequest req)
    DeleteEvaluationResponse DeleteEvaluation(1: DeleteEvaluationRequest req)

    // Market Intelligence
    GetMarketTrendsResponse GetMarketTrends(1: GetMarketTrendsRequest req)
    UpdateMarketTrendResponse UpdateMarketTrend(1: UpdateMarketTrendRequest req)
    DeleteMarketTrendResponse DeleteMarketTrend(1: DeleteMarketTrendRequest req)
}

struct MarketTrend {
    1: i64 id
    2: string source
    3: string title
    4: string summary
    5: string original_url
    6: i32 financial_relevance
    7: list<string> related_sectors
    8: double sentiment_score
    9: string impact_type
    10: double weight
    11: bool is_still_valid
    12: string created_at
    13: string updated_at
    14: list<string> related_stocks
    15: string impact_analysis
    16: string impact_scope // "specific", "sector", "market_wide"
}

struct GetMarketTrendsRequest {
    1: i32 page = 1
    2: i32 page_size = 20
    3: string impact_type // optional filter
    4: string related_stock_id // optional filter: specific stock code
    5: string query // optional search keyword
    6: string sort // optional sort order, e.g. "weight_desc", "created_at_desc"
}

struct GetMarketTrendsResponse {
    1: list<MarketTrend> trends
    2: i64 total
}

struct UpdateMarketTrendRequest {
    1: MarketTrend trend
}

struct UpdateMarketTrendResponse {
    1: bool success
}

struct DeleteMarketTrendRequest {
    1: i64 id
}

struct DeleteMarketTrendResponse {
    1: bool success
}
