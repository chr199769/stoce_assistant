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

service StockService {
    GetRealtimeResponse GetRealtime(1: GetRealtimeRequest req)
    GetFinancialReportResponse GetFinancialReport(1: GetFinancialReportRequest req)
}
