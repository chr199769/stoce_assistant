import axios from 'axios';

// Assume gateway is running on localhost:8888 (Wait, gateway port?)
// Check backend/gateway/main.go to find port.
// Usually 8080 or 8888. stock_service is 8888. ai_service is 8889. gateway?
// I'll check gateway main.go later. Assuming 8080 for now.

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
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

export interface GraphEntity {
  id: string;
  type: string;
  name: string;
  attributes?: Record<string, string>;
}

export interface GraphEvidence {
  id: string;
  category: string;
  summary: string;
  confidence: number;
  source: string;
  timestamp: number;
}

export interface GraphRelation {
  source: GraphEntity;
  target: GraphEntity;
  type: string;
  strength: number;
  evidence?: GraphEvidence;
  updated_at: number;
}

export interface GraphEvent {
  id: string;
  type: string;
  entities: GraphEntity[];
  impact_direction: number;
  impact_strength: number;
  confidence: number;
  timestamp: number;
  source: string;
  dedupe_key: string;
}

export const getGraphNeighborhood = async (params: {
  entity_type: string;
  entity_id: string;
  relation_types?: string;
  depth?: number;
  start_time?: number;
  end_time?: number;
  max_edges?: number;
}) => {
  const res = await api.get<{ entities: GraphEntity[]; relations: GraphRelation[] }>('/graph/neighborhood', { params });
  return res.data;
};

export const getGraphEntityProfile = async (params: {
  entity_type: string;
  entity_id: string;
  start_time?: number;
  end_time?: number;
}) => {
  const res = await api.get<{ entity?: GraphEntity; relations: GraphRelation[]; events: GraphEvent[] }>('/graph/entity', { params });
  return res.data;
};

export const searchGraphEvents = async (params: {
  entity_type: string;
  entity_id: string;
  event_types?: string;
  start_time?: number;
  end_time?: number;
  min_confidence?: number;
  limit?: number;
}) => {
  const res = await api.get<{ events: GraphEvent[] }>('/graph/events', { params });
  return res.data;
};

export default api;
