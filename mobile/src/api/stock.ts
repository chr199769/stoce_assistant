import client from './client';
import { RealtimeResponse, PredictionResponse, PredictionRequest } from '../types';

export const getRealtime = async (code: string): Promise<RealtimeResponse> => {
  const response = await client.get<RealtimeResponse>(`/api/stocks/${code}/realtime`);
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
