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

service AIService {
    GetPredictionResponse GetPrediction(1: GetPredictionRequest req)
    ImageRecognitionResponse ImageRecognition(1: ImageRecognitionRequest req)
}
