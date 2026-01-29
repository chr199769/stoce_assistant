package main

import (
	"context"
	"fmt"

	"stock_assistant/backend/knowledge_graph/dal/graphdb"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"
)

type KnowledgeGraphServiceImpl struct{}

func (s *KnowledgeGraphServiceImpl) UpsertEntities(ctx context.Context, req *knowledge_graph.UpsertEntitiesRequest) (resp *knowledge_graph.UpsertEntitiesResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	count, err := graphdb.UpsertEntities(ctx, req.Entities)
	if err != nil {
		return nil, err
	}
	return &knowledge_graph.UpsertEntitiesResponse{Upserted: int32(count)}, nil
}

func (s *KnowledgeGraphServiceImpl) UpsertRelations(ctx context.Context, req *knowledge_graph.UpsertRelationsRequest) (resp *knowledge_graph.UpsertRelationsResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	count, err := graphdb.UpsertRelations(ctx, req.Relations)
	if err != nil {
		return nil, err
	}
	return &knowledge_graph.UpsertRelationsResponse{Upserted: int32(count)}, nil
}

func (s *KnowledgeGraphServiceImpl) UpsertEvents(ctx context.Context, req *knowledge_graph.UpsertEventsRequest) (resp *knowledge_graph.UpsertEventsResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	count, err := graphdb.UpsertEvents(ctx, req.Events)
	if err != nil {
		return nil, err
	}
	return &knowledge_graph.UpsertEventsResponse{Upserted: int32(count)}, nil
}

func (s *KnowledgeGraphServiceImpl) UpsertEvidences(ctx context.Context, req *knowledge_graph.UpsertEvidencesRequest) (resp *knowledge_graph.UpsertEvidencesResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	count, err := graphdb.UpsertEvidences(ctx, req.Evidences)
	if err != nil {
		return nil, err
	}
	return &knowledge_graph.UpsertEvidencesResponse{Upserted: int32(count)}, nil
}

func (s *KnowledgeGraphServiceImpl) GetGraphNeighborhood(ctx context.Context, req *knowledge_graph.GetGraphNeighborhoodRequest) (resp *knowledge_graph.GetGraphNeighborhoodResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	return graphdb.GetGraphNeighborhood(ctx, req)
}

func (s *KnowledgeGraphServiceImpl) GetEntityProfile(ctx context.Context, req *knowledge_graph.GetEntityProfileRequest) (resp *knowledge_graph.GetEntityProfileResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	return graphdb.GetEntityProfile(ctx, req)
}

func (s *KnowledgeGraphServiceImpl) SearchEvents(ctx context.Context, req *knowledge_graph.SearchEventsRequest) (resp *knowledge_graph.SearchEventsResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	return graphdb.SearchEvents(ctx, req)
}

func (s *KnowledgeGraphServiceImpl) GetEvidenceBundle(ctx context.Context, req *knowledge_graph.GetEvidenceBundleRequest) (resp *knowledge_graph.GetEvidenceBundleResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	return graphdb.GetEvidenceBundle(ctx, req)
}

func (s *KnowledgeGraphServiceImpl) GetRelationStrength(ctx context.Context, req *knowledge_graph.GetRelationStrengthRequest) (resp *knowledge_graph.GetRelationStrengthResponse, err error) {
	if err := ensureGraphDB(); err != nil {
		return nil, err
	}
	return graphdb.GetRelationStrength(ctx, req)
}

func ensureGraphDB() error {
	if graphdb.Available() {
		return nil
	}
	return fmt.Errorf("neo4j 未连接")
}
