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

service StockService {
    GetRealtimeResponse GetRealtime(1: GetRealtimeRequest req)
    GetFinancialReportResponse GetFinancialReport(1: GetFinancialReportRequest req)
    
    // New methods
    GetMarketSectorsResponse GetMarketSectors(1: GetMarketSectorsRequest req)
    GetLimitUpPoolResponse GetLimitUpPool(1: GetLimitUpPoolRequest req)
}
