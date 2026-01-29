package api

import (
	"context"
	"fmt"
	"strings"

	"stock_assistant/backend/gateway/biz/rpc"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type GraphNeighborhoodReq struct {
	EntityType    string `query:"entity_type"`
	EntityId      string `query:"entity_id"`
	RelationTypes string `query:"relation_types"`
	Depth         int32  `query:"depth"`
	StartTime     int64  `query:"start_time"`
	EndTime       int64  `query:"end_time"`
	MaxEdges      int32  `query:"max_edges"`
}

type GraphEntityProfileReq struct {
	EntityType string `query:"entity_type"`
	EntityId   string `query:"entity_id"`
	StartTime  int64  `query:"start_time"`
	EndTime    int64  `query:"end_time"`
}

type GraphEventsReq struct {
	EntityType    string `query:"entity_type"`
	EntityId      string `query:"entity_id"`
	EventTypes    string `query:"event_types"`
	StartTime     int64  `query:"start_time"`
	EndTime       int64  `query:"end_time"`
	MinConfidence int32  `query:"min_confidence"`
	Limit         int32  `query:"limit"`
}

type GraphEntity struct {
	Id         string            `json:"id"`
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
}

type GraphEvidence struct {
	Id         string `json:"id"`
	Category   string `json:"category"`
	Summary    string `json:"summary"`
	Confidence int32  `json:"confidence"`
	Source     string `json:"source"`
	Timestamp  int64  `json:"timestamp"`
}

type GraphRelation struct {
	Source    GraphEntity   `json:"source"`
	Target    GraphEntity   `json:"target"`
	Type      string        `json:"type"`
	Strength  float64       `json:"strength"`
	Evidence  *GraphEvidence `json:"evidence,omitempty"`
	UpdatedAt int64         `json:"updated_at"`
}

type GraphEvent struct {
	Id              string       `json:"id"`
	Type            string       `json:"type"`
	Entities        []GraphEntity `json:"entities"`
	ImpactDirection int32        `json:"impact_direction"`
	ImpactStrength  int32        `json:"impact_strength"`
	Confidence      int32        `json:"confidence"`
	Timestamp       int64        `json:"timestamp"`
	Source          string       `json:"source"`
	DedupeKey       string       `json:"dedupe_key"`
}

type GraphNeighborhoodResp struct {
	Entities  []GraphEntity  `json:"entities"`
	Relations []GraphRelation `json:"relations"`
}

type GraphEntityProfileResp struct {
	Entity    *GraphEntity    `json:"entity"`
	Relations []GraphRelation `json:"relations"`
	Events    []GraphEvent    `json:"events"`
}

type GraphEventsResp struct {
	Events []GraphEvent `json:"events"`
}

func GetGraphNeighborhood(ctx context.Context, c *app.RequestContext) {
	var req GraphNeighborhoodReq
	if err := c.BindAndValidate(&req); err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	if req.EntityType == "" || req.EntityId == "" {
		c.String(consts.StatusBadRequest, "entity_type 与 entity_id 必填")
		return
	}
	entityType, err := knowledge_graph.EntityTypeFromString(strings.ToUpper(req.EntityType))
	if err != nil {
		c.String(consts.StatusBadRequest, "entity_type 无效")
		return
	}
	relationTypes, err := parseRelationTypes(req.RelationTypes)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	depth := req.Depth
	if depth <= 0 {
		depth = 1
	}
	maxEdges := req.MaxEdges
	if maxEdges <= 0 {
		maxEdges = 200
	}

	rpcReq := &knowledge_graph.GetGraphNeighborhoodRequest{
		EntityType:    entityType,
		EntityId:      req.EntityId,
		RelationTypes: relationTypes,
		Depth:         depth,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		MaxEdges:      maxEdges,
	}
	rpcResp, err := rpc.KnowledgeGraphClient.GetGraphNeighborhood(ctx, rpcReq)
	if err != nil {
		c.String(consts.StatusInternalServerError, err.Error())
		return
	}
	resp := GraphNeighborhoodResp{
		Entities:  mapEntities(rpcResp.Entities),
		Relations: mapRelations(rpcResp.Relations),
	}
	c.JSON(consts.StatusOK, resp)
}

func GetGraphEntityProfile(ctx context.Context, c *app.RequestContext) {
	var req GraphEntityProfileReq
	if err := c.BindAndValidate(&req); err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	if req.EntityType == "" || req.EntityId == "" {
		c.String(consts.StatusBadRequest, "entity_type 与 entity_id 必填")
		return
	}
	entityType, err := knowledge_graph.EntityTypeFromString(strings.ToUpper(req.EntityType))
	if err != nil {
		c.String(consts.StatusBadRequest, "entity_type 无效")
		return
	}
	rpcReq := &knowledge_graph.GetEntityProfileRequest{
		EntityType: entityType,
		EntityId:   req.EntityId,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	}
	rpcResp, err := rpc.KnowledgeGraphClient.GetEntityProfile(ctx, rpcReq)
	if err != nil {
		c.String(consts.StatusInternalServerError, err.Error())
		return
	}
	resp := GraphEntityProfileResp{
		Entity:    mapEntity(rpcResp.Entity),
		Relations: mapRelations(rpcResp.Relations),
		Events:    mapEvents(rpcResp.Events),
	}
	c.JSON(consts.StatusOK, resp)
}

func SearchGraphEvents(ctx context.Context, c *app.RequestContext) {
	var req GraphEventsReq
	if err := c.BindAndValidate(&req); err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	if req.EntityType == "" || req.EntityId == "" {
		c.String(consts.StatusBadRequest, "entity_type 与 entity_id 必填")
		return
	}
	entityType, err := knowledge_graph.EntityTypeFromString(strings.ToUpper(req.EntityType))
	if err != nil {
		c.String(consts.StatusBadRequest, "entity_type 无效")
		return
	}
	eventTypes, err := parseEventTypes(req.EventTypes)
	if err != nil {
		c.String(consts.StatusBadRequest, err.Error())
		return
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	rpcReq := &knowledge_graph.SearchEventsRequest{
		EntityType:    entityType,
		EntityId:      req.EntityId,
		EventTypes:    eventTypes,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		MinConfidence: req.MinConfidence,
		Limit:         limit,
	}
	rpcResp, err := rpc.KnowledgeGraphClient.SearchEvents(ctx, rpcReq)
	if err != nil {
		c.String(consts.StatusInternalServerError, err.Error())
		return
	}
	resp := GraphEventsResp{
		Events: mapEvents(rpcResp.Events),
	}
	c.JSON(consts.StatusOK, resp)
}

func parseRelationTypes(raw string) ([]knowledge_graph.RelationType, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	result := make([]knowledge_graph.RelationType, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		t, err := knowledge_graph.RelationTypeFromString(strings.ToUpper(value))
		if err != nil {
			return nil, fmt.Errorf("relation_types 无效")
		}
		result = append(result, t)
	}
	return result, nil
}

func parseEventTypes(raw string) ([]knowledge_graph.EventType, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	result := make([]knowledge_graph.EventType, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		t, err := knowledge_graph.EventTypeFromString(strings.ToUpper(value))
		if err != nil {
			return nil, fmt.Errorf("event_types 无效")
		}
		result = append(result, t)
	}
	return result, nil
}

func mapEntity(entity *knowledge_graph.Entity) *GraphEntity {
	if entity == nil {
		return nil
	}
	return &GraphEntity{
		Id:         entity.Id,
		Type:       entity.Type.String(),
		Name:       entity.Name,
		Attributes: entity.Attributes,
	}
}

func mapEntities(entities []*knowledge_graph.Entity) []GraphEntity {
	result := make([]GraphEntity, 0, len(entities))
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		result = append(result, GraphEntity{
			Id:         entity.Id,
			Type:       entity.Type.String(),
			Name:       entity.Name,
			Attributes: entity.Attributes,
		})
	}
	return result
}

func mapEvidence(evidence *knowledge_graph.Evidence) *GraphEvidence {
	if evidence == nil {
		return nil
	}
	return &GraphEvidence{
		Id:         evidence.Id,
		Category:   evidence.Category,
		Summary:    evidence.Summary,
		Confidence: evidence.Confidence,
		Source:     evidence.Source,
		Timestamp:  evidence.Timestamp,
	}
}

func mapRelations(relations []*knowledge_graph.Relation) []GraphRelation {
	result := make([]GraphRelation, 0, len(relations))
	for _, relation := range relations {
		if relation == nil {
			continue
		}
		source := mapEntity(relation.Source)
		target := mapEntity(relation.Target)
		if source == nil || target == nil {
			continue
		}
		result = append(result, GraphRelation{
			Source:    *source,
			Target:    *target,
			Type:      relation.Type.String(),
			Strength:  relation.Strength,
			Evidence:  mapEvidence(relation.Evidence),
			UpdatedAt: relation.UpdatedAt,
		})
	}
	return result
}

func mapEvents(events []*knowledge_graph.Event) []GraphEvent {
	result := make([]GraphEvent, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}
		result = append(result, GraphEvent{
			Id:              event.Id,
			Type:            event.Type.String(),
			Entities:        mapEntities(event.Entities),
			ImpactDirection: event.ImpactDirection,
			ImpactStrength:  event.ImpactStrength,
			Confidence:      event.Confidence,
			Timestamp:       event.Timestamp,
			Source:          event.Source,
			DedupeKey:       event.DedupeKey,
		})
	}
	return result
}
