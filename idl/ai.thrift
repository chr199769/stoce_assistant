namespace go ai

struct PredictionResult {
    1: string code
    2: double confidence
    3: string analysis
    4: string news_summary
    5: string trace_id
}

struct GetPredictionRequest {
    1: string code
    2: i32 days
    3: bool include_news
    4: string model
}

struct GetPredictionResponse {
    1: PredictionResult result
}

struct ImageRecognitionRequest {
    1: binary image_data
    2: string model
}

struct RecognizedStock {
    1: string code
    2: string name
}

struct ImageRecognitionResponse {
    1: list<RecognizedStock> stocks
}

struct MarketReviewRequest {
    1: string date
    2: list<string> focus_sectors
}

struct MarketReviewResponse {
    1: string summary
    2: string sector_analysis
    3: string sentiment_analysis
    4: list<string> key_risks
    5: list<string> opportunities
}

struct MarketAnalysisRequest {
    1: string date
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

service AIService {
    GetPredictionResponse GetPrediction(1: GetPredictionRequest req)
    ImageRecognitionResponse ImageRecognition(1: ImageRecognitionRequest req)
    MarketReviewResponse MarketReview(1: MarketReviewRequest req)
    MarketAnalysisResponse AnalyzeMarket(1: MarketAnalysisRequest req)
    ProcessMarketTrendsResponse ProcessMarketTrends(1: ProcessMarketTrendsRequest req)
}

struct ProcessMarketTrendsRequest {
    1: list<RawTrendItem> items
}

struct RawTrendItem {
    1: string title
    2: string source
    3: string url
    4: string content
}

struct ProcessMarketTrendsResponse {
    1: list<AnalyzedTrendItem> items
}

struct AnalyzedTrendItem {
    1: string title
    2: string source
    3: string url
    4: string summary
    5: i32 financial_relevance // 0-10
    6: list<string> related_sectors
    7: double sentiment_score
    8: string impact_type // "policy_long_term", "short_term_news"
    9: double weight
    10: list<string> related_stocks
    11: string impact_analysis
    12: string impact_scope // "specific", "sector", "market_wide"
}
