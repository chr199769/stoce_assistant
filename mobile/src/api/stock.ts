import client from './client';
import { RealtimeResponse, PredictionResponse, PredictionRequest, ImageRecognitionResponse } from '../types';

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
