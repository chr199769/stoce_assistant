import client from './client';
import {
  RealtimeResponse,
  PredictionResponse,
  PredictionRequest,
  ImageRecognitionResponse,
  MarketReviewRequest,
  MarketReviewResponse,
  MarketAnalysisRequest,
  MarketAnalysisResponse,
  GetSectorStocksResponse,
  GetMarketSectorsResponse,
  GetDragonTigerListResponse,
  AddWatchlistResponse,
  RemoveWatchlistResponse,
  GetWatchlistResponse
} from '../types';

export const getRealtime = async (code: string): Promise<RealtimeResponse> => {
  const response = await client.get<RealtimeResponse>(`/api/stocks/${code}/realtime`);
  return response.data;
};

export const recognizeStockImage = async (imageUri: string, imageType: string, imageName: string): Promise<ImageRecognitionResponse> => {
  const formData = new FormData();
  formData.append('image', {
    uri: imageUri,
    type: imageType,
    name: imageName,
  } as any);

  const response = await client.post<ImageRecognitionResponse>('/api/image/recognize', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    timeout: 60000, // Explicitly set timeout to 60s
  });
  return response.data;
};

export const getPrediction = async (req: PredictionRequest): Promise<PredictionResponse> => {
  const response = await client.post<PredictionResponse>(`/api/prediction/${req.code}`, {
    days: req.days,
    include_news: req.include_news,
    model: req.model,
  });
  return response.data;
};

export const marketReview = async (req: MarketReviewRequest): Promise<MarketReviewResponse> => {
  const response = await client.post<MarketReviewResponse>('/api/market/review', {
    date: req.date,
    focus_sectors: req.focus_sectors,
  });
  return response.data;
};

export const analyzeMarket = async (req: MarketAnalysisRequest): Promise<MarketAnalysisResponse> => {
  const response = await client.post<MarketAnalysisResponse>('/api/market/analysis', {
    date: req.date,
  });
  return response.data;
};

export const getSectorStocks = async (sectorCode: string): Promise<GetSectorStocksResponse> => {
  const response = await client.get<GetSectorStocksResponse>(`/api/stock/sector/stocks?sector_code=${sectorCode}`);
  return response.data;
};

export const getMarketSectors = async (limit: number = 20, type: string = 'concept'): Promise<GetMarketSectorsResponse> => {
  const response = await client.get<GetMarketSectorsResponse>(`/api/market/sectors?limit=${limit}&type=${type}`);
  return response.data;
};

export const getDragonTigerList = async (date: string): Promise<GetDragonTigerListResponse> => {
  const response = await client.get<GetDragonTigerListResponse>(`/api/stock/dragontiger/list?date=${date}`);
  return response.data;
};

export const addWatchlist = async (userId: string, stockCode: string): Promise<AddWatchlistResponse> => {
  const response = await client.post<AddWatchlistResponse>('/api/watchlist/add', {
    user_id: userId,
    stock_code: stockCode,
  });
  return response.data;
};

export const removeWatchlist = async (userId: string, stockCode: string): Promise<RemoveWatchlistResponse> => {
  const response = await client.post<RemoveWatchlistResponse>('/api/watchlist/remove', {
    user_id: userId,
    stock_code: stockCode,
  });
  return response.data;
};

export const getWatchlist = async (userId: string): Promise<GetWatchlistResponse> => {
  const response = await client.get<GetWatchlistResponse>(`/api/watchlist/list?user_id=${userId}`);
  return response.data;
};
