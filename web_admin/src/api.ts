import axios from 'axios';

// Assume gateway is running on localhost:8888 (Wait, gateway port?)
// Check backend/gateway/main.go to find port.
// Usually 8080 or 8888. stock_service is 8888. ai_service is 8889. gateway?
// I'll check gateway main.go later. Assuming 8080 for now.

const api = axios.create({
  baseURL: '/api', // Proxy in vite config
});

export interface EvaluationRecord {
  id: string;
  prediction_id: string;
  stock_code: string;
  prediction_date: string;
  initial_price: number;
  price_1d: number;
  price_2d: number;
  price_3d: number;
  score: number;
  status: string;
  stock_name?: string;
}

export const getEvaluations = async (code?: string, limit = 20, offset = 0) => {
  const params: any = { limit, offset };
  if (code) params.code = code;
  const res = await api.get<{ evaluations: EvaluationRecord[] }>('/evaluations', { params });
  return res.data;
};

export const deleteEvaluation = async (id: string) => {
  const res = await api.delete<{ success: boolean }>(`/evaluations/${id}`);
  return res.data;
};

export interface MarketTrend {
  id: number;
  source: string;
  title: string;
  summary: string;
  original_url: string;
  financial_relevance: number;
  related_sectors: string[];
  related_stocks?: string[];
  impact_analysis?: string;
  sentiment_score: number;
  impact_type: string;
  impact_scope: string;
  weight: number;
  is_still_valid: boolean;
  created_at: string;
  updated_at: string;
}

export const getMarketTrends = async (page = 1, pageSize = 20, impactType?: string) => {
  const params: any = { page, page_size: pageSize };
  if (impactType) params.impact_type = impactType;
  const res = await api.get<{ trends: MarketTrend[], total: number }>('/market/trends', { params });
  return res.data;
};

export const updateMarketTrend = async (trend: MarketTrend) => {
  const res = await api.post<{ success: boolean }>('/market/trends/update', { trend });
  return res.data;
};

export const deleteMarketTrend = async (id: number) => {
  const res = await api.delete<{ success: boolean }>(`/market/trends/${id}`);
  return res.data;
};

export default api;
