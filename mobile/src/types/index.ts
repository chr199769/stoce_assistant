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

export interface MarketReviewRequest {
  date?: string;
  focus_sectors?: string[];
}

export interface MarketReviewResponse {
  summary: string;
  confidence: number;
  date: string;
}

// Phase 2: Sector & Dragon Tiger Types
export interface SectorStockItem {
  code: string;
  name: string;
  price: number;
  change_percent: number;
  volume: number;
  amount: number;
  market_cap: number;
}

export interface GetSectorStocksResponse {
  stocks: SectorStockItem[];
}

export interface DragonTigerSeat {
  name: string;
  buy_amt: number;
  sell_amt: number;
  net_amt: number;
  tags: string[];
}

export interface DragonTigerItem {
  code: string;
  name: string;
  close_price: number;
  change_percent: number;
  reason: string;
  net_inflow: number;
  buy_seats: DragonTigerSeat[];
  sell_seats: DragonTigerSeat[];
}

export interface GetDragonTigerListResponse {
  items: DragonTigerItem[];
}
