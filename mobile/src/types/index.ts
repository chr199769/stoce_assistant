export interface RealtimeResponse {
  code: string;
  name: string;
  current_price: number;
  change_percent: number;
  volume: number;
  timestamp: string;
}

export interface PredictionResponse {
  code: string;
  confidence: number;
  analysis: string;
  news_summary: string;
}

export interface PredictionRequest {
  code: string;
  days: number;
  include_news: boolean;
  model: string;
}

export interface RecognizedStock {
  code: string;
  name: string;
}

export interface ImageRecognitionResponse {
  stocks: RecognizedStock[];
}
