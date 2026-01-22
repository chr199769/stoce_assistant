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

export default api;
