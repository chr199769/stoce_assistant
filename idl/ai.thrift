namespace go ai

struct PredictionResult {
    1: string code
    2: double confidence
    3: string analysis
    4: string news_summary
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
}

service AIService {
    GetPredictionResponse GetPrediction(1: GetPredictionRequest req)
    ImageRecognitionResponse ImageRecognition(1: ImageRecognitionRequest req)
    MarketReviewResponse MarketReview(1: MarketReviewRequest req)
    MarketAnalysisResponse AnalyzeMarket(1: MarketAnalysisRequest req)
}
