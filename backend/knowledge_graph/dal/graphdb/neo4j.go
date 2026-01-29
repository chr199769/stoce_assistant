package graphdb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"stock_assistant/backend/knowledge_graph/config"
	knowledge_graph "stock_assistant/backend/knowledge_graph/kitex_gen/knowledge_graph"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var Driver neo4j.DriverWithContext

func Init() error {
	uri := ""
	user := ""
	password := ""
	cfg := config.Get()
	if cfg != nil && cfg.Neo4j != nil {
		uri = cfg.Neo4j.URI
		user = cfg.Neo4j.User
		password = cfg.Neo4j.Password
	}
	if uri == "" {
		uri = "neo4j://localhost:7687"
	}
	if user == "" {
		user = "neo4j"
	}
	if password == "" {
		return fmt.Errorf("NEO4J_PASSWORD 未设置")
	}

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return fmt.Errorf("Neo4j 初始化失败: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return fmt.Errorf("Neo4j 连接失败: %w", err)
	}
	Driver = driver
	_ = ensureSchema(context.Background())
	return nil
}

func Available() bool {
	return Driver != nil
}

func UpsertEntities(ctx context.Context, entities []*knowledge_graph.Entity) (int, error) {
	if Driver == nil || len(entities) == 0 {
		return 0, nil
	}
	rows := make([]map[string]any, 0, len(entities))
	now := time.Now().Unix()
	for _, entity := range entities {
		if entity == nil || entity.Id == "" {
			continue
		}
		rows = append(rows, map[string]any{
			"id":         entity.Id,
			"type":       int64(entity.Type),
			"name":       entity.Name,
			"attributes": serializeAttributes(entity.Attributes),
			"created_at": now,
			"updated_at": now,
		})
	}
	if len(rows) == 0 {
		return 0, nil
	}
	_, err := executeWrite(ctx, `
UNWIND $rows AS row
MERGE (e:Entity {id: row.id})
ON CREATE SET e.created_at = row.created_at
SET e.type = row.type,
    e.name = row.name,
    e.attributes = row.attributes,
    e.updated_at = row.updated_at
`, map[string]any{"rows": rows})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func UpsertEvidences(ctx context.Context, evidences []*knowledge_graph.Evidence) (int, error) {
	if Driver == nil || len(evidences) == 0 {
		return 0, nil
	}
	rows := make([]map[string]any, 0, len(evidences))
	now := time.Now().Unix()
	for _, ev := range evidences {
		if ev == nil || ev.Id == "" {
			continue
		}
		rows = append(rows, map[string]any{
			"id":         ev.Id,
			"category":   ev.Category,
			"summary":    ev.Summary,
			"confidence": int64(ev.Confidence),
			"source":     ev.Source,
			"timestamp":  ev.Timestamp,
			"created_at": now,
			"updated_at": now,
		})
	}
	if len(rows) == 0 {
		return 0, nil
	}
	_, err := executeWrite(ctx, `
UNWIND $rows AS row
MERGE (e:Evidence {id: row.id})
ON CREATE SET e.created_at = row.created_at
SET e.category = row.category,
    e.summary = row.summary,
    e.confidence = row.confidence,
    e.source = row.source,
    e.timestamp = row.timestamp,
    e.updated_at = row.updated_at
`, map[string]any{"rows": rows})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func UpsertRelations(ctx context.Context, relations []*knowledge_graph.Relation) (int, error) {
	if Driver == nil || len(relations) == 0 {
		return 0, nil
	}
	rows := make([]map[string]any, 0, len(relations))
	evidenceRows := make([]*knowledge_graph.Evidence, 0)
	now := time.Now().Unix()
	for _, rel := range relations {
		if rel == nil || rel.Source == nil || rel.Target == nil {
			continue
		}
		updatedAt := rel.UpdatedAt
		if updatedAt == 0 {
			updatedAt = now
		}
		evidenceID := ""
		if rel.Evidence != nil {
			evidenceID = rel.Evidence.Id
			evidenceRows = append(evidenceRows, rel.Evidence)
		}
		rows = append(rows, map[string]any{
			"source_id":   rel.Source.Id,
			"source_type": int64(rel.Source.Type),
			"target_id":   rel.Target.Id,
			"target_type": int64(rel.Target.Type),
			"type":        int64(rel.Type),
			"strength":    rel.Strength,
			"evidence_id": evidenceID,
			"created_at":  updatedAt,
			"updated_at":  updatedAt,
		})
	}
	if len(rows) == 0 {
		return 0, nil
	}
	if _, err := UpsertEvidences(ctx, evidenceRows); err != nil {
		return 0, err
	}
	_, err := executeWrite(ctx, `
UNWIND $rows AS row
MERGE (s:Entity {id: row.source_id})
SET s.type = row.source_type
MERGE (t:Entity {id: row.target_id})
SET t.type = row.target_type
MERGE (s)-[r:RELATION {source_id: row.source_id, target_id: row.target_id, type: row.type}]->(t)
ON CREATE SET r.created_at = row.created_at
SET r.strength = row.strength,
    r.evidence_id = row.evidence_id,
    r.updated_at = row.updated_at
`, map[string]any{"rows": rows})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func UpsertEvents(ctx context.Context, events []*knowledge_graph.Event) (int, error) {
	if Driver == nil || len(events) == 0 {
		return 0, nil
	}
	rows := make([]map[string]any, 0, len(events))
	now := time.Now().Unix()
	for _, ev := range events {
		if ev == nil || ev.Id == "" {
			continue
		}
		entityID := ""
		entityType := int64(0)
		if len(ev.Entities) > 0 && ev.Entities[0] != nil {
			entityID = ev.Entities[0].Id
			entityType = int64(ev.Entities[0].Type)
		}
		rows = append(rows, map[string]any{
			"id":               ev.Id,
			"type":             int64(ev.Type),
			"impact_direction": int64(ev.ImpactDirection),
			"impact_strength":  int64(ev.ImpactStrength),
			"confidence":       int64(ev.Confidence),
			"timestamp":        ev.Timestamp,
			"source":           ev.Source,
			"dedupe_key":       dedupeKey(ev),
			"entity_id":        entityID,
			"entity_type":      entityType,
			"created_at":       now,
			"updated_at":       now,
		})
	}
	if len(rows) == 0 {
		return 0, nil
	}
	_, err := executeWrite(ctx, `
UNWIND $rows AS row
MERGE (e:Event {id: row.id})
ON CREATE SET e.created_at = row.created_at
SET e.type = row.type,
    e.impact_direction = row.impact_direction,
    e.impact_strength = row.impact_strength,
    e.confidence = row.confidence,
    e.timestamp = row.timestamp,
    e.source = row.source,
    e.dedupe_key = row.dedupe_key,
    e.updated_at = row.updated_at
FOREACH (_ IN CASE WHEN row.entity_id <> "" THEN [1] ELSE [] END |
  MERGE (ent:Entity {id: row.entity_id})
  SET ent.type = row.entity_type
  MERGE (ent)-[:HAS_EVENT]->(e)
)
`, map[string]any{"rows": rows})
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func GetGraphNeighborhood(ctx context.Context, req *knowledge_graph.GetGraphNeighborhoodRequest) (*knowledge_graph.GetGraphNeighborhoodResponse, error) {
	if Driver == nil || req == nil {
		return emptyNeighborhoodResponse(), nil
	}
	types := make([]int64, 0, len(req.RelationTypes))
	for _, t := range req.RelationTypes {
		types = append(types, int64(t))
	}
	limit := int64(req.MaxEdges)
	if limit <= 0 {
		limit = 100
	}
	records, err := executeRead(ctx, `
MATCH (s:Entity {id: $entity_id})
WHERE s.type = $entity_type
OPTIONAL MATCH (s)-[r:RELATION]-(t:Entity)
WHERE ($types_size = 0 OR r.type IN $types)
  AND ($start_time = 0 OR r.updated_at >= $start_time)
  AND ($end_time = 0 OR r.updated_at <= $end_time)
OPTIONAL MATCH (ev:Evidence {id: r.evidence_id})
RETURN s, r, t, ev
LIMIT $limit
`, map[string]any{
		"entity_id":   req.EntityId,
		"entity_type": int64(req.EntityType),
		"types":       types,
		"types_size":  len(types),
		"start_time":  req.StartTime,
		"end_time":    req.EndTime,
		"limit":       limit,
	})
	if err != nil {
		return nil, err
	}
	entities := make(map[string]*knowledge_graph.Entity)
	relations := make([]*knowledge_graph.Relation, 0)
	for _, record := range records {
		s := nodeToEntity(recordValue(record, "s"))
		if s != nil {
			entities[s.Id] = s
		}
		t := nodeToEntity(recordValue(record, "t"))
		if t != nil {
			entities[t.Id] = t
		}
		rel := relationFrom(recordValue(record, "r"), s, t, nodeToEvidence(recordValue(record, "ev")))
		if rel != nil {
			relations = append(relations, rel)
		}
	}
	return &knowledge_graph.GetGraphNeighborhoodResponse{
		Entities:  mapEntityValues(entities),
		Relations: relations,
	}, nil
}

func GetEntityProfile(ctx context.Context, req *knowledge_graph.GetEntityProfileRequest) (*knowledge_graph.GetEntityProfileResponse, error) {
	if Driver == nil || req == nil {
		return &knowledge_graph.GetEntityProfileResponse{
			Entity:    nil,
			Relations: make([]*knowledge_graph.Relation, 0),
			Events:    make([]*knowledge_graph.Event, 0),
		}, nil
	}
	records, err := executeRead(ctx, `
MATCH (e:Entity {id: $entity_id})
WHERE e.type = $entity_type
OPTIONAL MATCH (e)-[r:RELATION]-(t:Entity)
WHERE ($start_time = 0 OR r.updated_at >= $start_time)
  AND ($end_time = 0 OR r.updated_at <= $end_time)
OPTIONAL MATCH (ev:Evidence {id: r.evidence_id})
RETURN e, r, t, ev
`, map[string]any{
		"entity_id":   req.EntityId,
		"entity_type": int64(req.EntityType),
		"start_time":  req.StartTime,
		"end_time":    req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	var entity *knowledge_graph.Entity
	relations := make([]*knowledge_graph.Relation, 0)
	for _, record := range records {
		if entity == nil {
			entity = nodeToEntity(recordValue(record, "e"))
		}
		t := nodeToEntity(recordValue(record, "t"))
		rel := relationFrom(recordValue(record, "r"), entity, t, nodeToEvidence(recordValue(record, "ev")))
		if rel != nil {
			relations = append(relations, rel)
		}
	}

	eventResp, err := SearchEvents(ctx, &knowledge_graph.SearchEventsRequest{
		EntityType:    req.EntityType,
		EntityId:      req.EntityId,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		MinConfidence: 0,
		Limit:         200,
	})
	if err != nil {
		return nil, err
	}
	return &knowledge_graph.GetEntityProfileResponse{
		Entity:    entity,
		Relations: relations,
		Events:    eventResp.Events,
	}, nil
}

func SearchEvents(ctx context.Context, req *knowledge_graph.SearchEventsRequest) (*knowledge_graph.SearchEventsResponse, error) {
	if Driver == nil || req == nil {
		return &knowledge_graph.SearchEventsResponse{Events: make([]*knowledge_graph.Event, 0)}, nil
	}
	types := make([]int64, 0, len(req.EventTypes))
	for _, t := range req.EventTypes {
		types = append(types, int64(t))
	}
	limit := int64(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	records, err := executeRead(ctx, `
MATCH (e:Entity {id: $entity_id})-[:HAS_EVENT]->(ev:Event)
WHERE e.type = $entity_type
  AND ($types_size = 0 OR ev.type IN $types)
  AND ($start_time = 0 OR ev.timestamp >= $start_time)
  AND ($end_time = 0 OR ev.timestamp <= $end_time)
  AND ($min_confidence = 0 OR ev.confidence >= $min_confidence)
RETURN ev
ORDER BY ev.timestamp DESC
LIMIT $limit
`, map[string]any{
		"entity_id":      req.EntityId,
		"entity_type":    int64(req.EntityType),
		"types":          types,
		"types_size":     len(types),
		"start_time":     req.StartTime,
		"end_time":       req.EndTime,
		"min_confidence": int64(req.MinConfidence),
		"limit":          limit,
	})
	if err != nil {
		return nil, err
	}
	events := make([]*knowledge_graph.Event, 0, len(records))
	for _, record := range records {
		event := nodeToEvent(recordValue(record, "ev"), req.EntityId, req.EntityType)
		if event != nil {
			events = append(events, event)
		}
	}
	return &knowledge_graph.SearchEventsResponse{Events: events}, nil
}

func GetRelationStrength(ctx context.Context, req *knowledge_graph.GetRelationStrengthRequest) (*knowledge_graph.GetRelationStrengthResponse, error) {
	if Driver == nil || req == nil {
		return &knowledge_graph.GetRelationStrengthResponse{Relation: nil}, nil
	}
	records, err := executeRead(ctx, `
MATCH (s:Entity {id: $source_id, type: $source_type})-[r:RELATION {type: $relation_type}]->(t:Entity {id: $target_id, type: $target_type})
OPTIONAL MATCH (ev:Evidence {id: r.evidence_id})
RETURN s, r, t, ev
ORDER BY r.updated_at DESC
LIMIT 1
`, map[string]any{
		"source_id":     req.SourceId,
		"source_type":   int64(req.SourceType),
		"target_id":     req.TargetId,
		"target_type":   int64(req.TargetType),
		"relation_type": int64(req.RelationType),
	})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &knowledge_graph.GetRelationStrengthResponse{Relation: nil}, nil
	}
	record := records[0]
	source := nodeToEntity(recordValue(record, "s"))
	target := nodeToEntity(recordValue(record, "t"))
	relation := relationFrom(recordValue(record, "r"), source, target, nodeToEvidence(recordValue(record, "ev")))
	return &knowledge_graph.GetRelationStrengthResponse{Relation: relation}, nil
}

func DeleteEventsBySourceBefore(ctx context.Context, source string, before int64) (int, error) {
	if Driver == nil || source == "" || before <= 0 {
		return 0, nil
	}
	_, err := executeWrite(ctx, `
MATCH (e:Event)
WHERE e.source = $source AND e.timestamp < $before
DETACH DELETE e
RETURN count(e) AS deleted
`, map[string]any{
		"source": source,
		"before": before,
	})
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func CountNodes(ctx context.Context) (int64, error) {
	if Driver == nil {
		return 0, nil
	}
	records, err := executeRead(ctx, `MATCH (n) RETURN count(n) AS c`, nil)
	if err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, nil
	}
	return toInt64(recordValue(records[0], "c")), nil
}

func DeleteAll(ctx context.Context) (int64, error) {
	if Driver == nil {
		return 0, nil
	}
	count, err := CountNodes(ctx)
	if err != nil {
		return 0, err
	}
	_, err = executeWrite(ctx, `MATCH (n) DETACH DELETE n`, nil)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func UpdateEventConfidenceBySource(ctx context.Context, sources []string, value int64) (int64, error) {
	if Driver == nil || len(sources) == 0 {
		return 0, nil
	}
	_, err := executeWrite(ctx, `
UNWIND $sources AS src
MATCH (e:Event) WHERE e.source = src
SET e.confidence = $value, e.updated_at = timestamp()/1000
`, map[string]any{"sources": sources, "value": value})
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func GetEvidenceBundle(ctx context.Context, req *knowledge_graph.GetEvidenceBundleRequest) (*knowledge_graph.GetEvidenceBundleResponse, error) {
	if Driver == nil || req == nil {
		return &knowledge_graph.GetEvidenceBundleResponse{
			Bundle: &knowledge_graph.EvidenceBundle{
				Entities:  make([]*knowledge_graph.Entity, 0),
				Relations: make([]*knowledge_graph.Relation, 0),
				Events:    make([]*knowledge_graph.Event, 0),
				Score:     0,
				Coverage:  0,
			},
		}, nil
	}
	eventsResp, err := SearchEvents(ctx, &knowledge_graph.SearchEventsRequest{
		EntityType:    knowledge_graph.EntityType_STOCK,
		EntityId:      req.StockCode,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		MinConfidence: req.MinConfidence,
		Limit:         req.MaxItems,
	})
	if err != nil {
		return nil, err
	}
	graphResp, err := GetGraphNeighborhood(ctx, &knowledge_graph.GetGraphNeighborhoodRequest{
		EntityType: knowledge_graph.EntityType_STOCK,
		EntityId:   req.StockCode,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		MaxEdges:   req.MaxItems,
	})
	if err != nil {
		return nil, err
	}
	return &knowledge_graph.GetEvidenceBundleResponse{
		Bundle: &knowledge_graph.EvidenceBundle{
			Entities:  graphResp.Entities,
			Relations: graphResp.Relations,
			Events:    eventsResp.Events,
			Score:     0,
			Coverage:  int32(len(eventsResp.Events)),
		},
	}, nil
}

func executeWrite(ctx context.Context, query string, params map[string]any) (any, error) {
	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	return session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		for result.Next(ctx) {
		}
		return nil, result.Err()
	})
}

func executeRead(ctx context.Context, query string, params map[string]any) ([]*neo4j.Record, error) {
	session := Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	recordsAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		records := make([]*neo4j.Record, 0)
		for result.Next(ctx) {
			records = append(records, result.Record())
		}
		if err = result.Err(); err != nil {
			return nil, err
		}
		return records, nil
	})
	if err != nil {
		return nil, err
	}
	return recordsAny.([]*neo4j.Record), nil
}

func nodeToEntity(value any) *knowledge_graph.Entity {
	node, ok := value.(neo4j.Node)
	if !ok {
		return nil
	}
	id := toString(node.Props["id"])
	if id == "" {
		return nil
	}
	return &knowledge_graph.Entity{
		Id:         id,
		Type:       knowledge_graph.EntityType(toInt64(node.Props["type"])),
		Name:       toString(node.Props["name"]),
		Attributes: parseAttributes(toString(node.Props["attributes"])),
	}
}

func nodeToEvidence(value any) *knowledge_graph.Evidence {
	node, ok := value.(neo4j.Node)
	if !ok {
		return nil
	}
	id := toString(node.Props["id"])
	if id == "" {
		return nil
	}
	return &knowledge_graph.Evidence{
		Id:         id,
		Category:   toString(node.Props["category"]),
		Summary:    toString(node.Props["summary"]),
		Confidence: int32(toInt64(node.Props["confidence"])),
		Source:     toString(node.Props["source"]),
		Timestamp:  toInt64(node.Props["timestamp"]),
	}
}

func nodeToEvent(value any, entityID string, entityType knowledge_graph.EntityType) *knowledge_graph.Event {
	node, ok := value.(neo4j.Node)
	if !ok {
		return nil
	}
	id := toString(node.Props["id"])
	if id == "" {
		return nil
	}
	return &knowledge_graph.Event{
		Id:              id,
		Type:            knowledge_graph.EventType(toInt64(node.Props["type"])),
		Entities:        []*knowledge_graph.Entity{{Id: entityID, Type: entityType}},
		ImpactDirection: int32(toInt64(node.Props["impact_direction"])),
		ImpactStrength:  int32(toInt64(node.Props["impact_strength"])),
		Confidence:      int32(toInt64(node.Props["confidence"])),
		Timestamp:       toInt64(node.Props["timestamp"]),
		Source:          toString(node.Props["source"]),
		DedupeKey:       toString(node.Props["dedupe_key"]),
	}
}

func relationFrom(value any, source, target *knowledge_graph.Entity, evidence *knowledge_graph.Evidence) *knowledge_graph.Relation {
	rel, ok := value.(neo4j.Relationship)
	if !ok || source == nil || target == nil {
		return nil
	}
	return &knowledge_graph.Relation{
		Source:    source,
		Target:    target,
		Type:      knowledge_graph.RelationType(toInt64(rel.Props["type"])),
		Strength:  toFloat64(rel.Props["strength"]),
		Evidence:  evidence,
		UpdatedAt: toInt64(rel.Props["updated_at"]),
	}
}

func mapEntityValues(values map[string]*knowledge_graph.Entity) []*knowledge_graph.Entity {
	result := make([]*knowledge_graph.Entity, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func parseAttributes(raw string) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	var value map[string]string
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return map[string]string{}
	}
	if value == nil {
		return map[string]string{}
	}
	return value
}

func serializeAttributes(value map[string]string) string {
	if len(value) == 0 {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func dedupeKey(event *knowledge_graph.Event) string {
	if event == nil {
		return ""
	}
	if event.DedupeKey != "" {
		return event.DedupeKey
	}
	return event.Id
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	default:
		return 0
	}
}

func toFloat64(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}

func recordValue(record *neo4j.Record, key string) any {
	value, _ := record.Get(key)
	return value
}

func emptyNeighborhoodResponse() *knowledge_graph.GetGraphNeighborhoodResponse {
	return &knowledge_graph.GetGraphNeighborhoodResponse{
		Entities:  make([]*knowledge_graph.Entity, 0),
		Relations: make([]*knowledge_graph.Relation, 0),
	}
}

func ensureSchema(ctx context.Context) error {
	if Driver == nil {
		return nil
	}
	stmts := []string{
		"CREATE CONSTRAINT entity_id_unique IF NOT EXISTS FOR (e:Entity) REQUIRE e.id IS UNIQUE",
		"CREATE CONSTRAINT event_id_unique IF NOT EXISTS FOR (e:Event) REQUIRE e.id IS UNIQUE",
		"CREATE CONSTRAINT evidence_id_unique IF NOT EXISTS FOR (e:Evidence) REQUIRE e.id IS UNIQUE",
		"CREATE INDEX rel_source_target_type IF NOT EXISTS FOR ()-[r:RELATION]-() ON (r.source_id, r.target_id, r.type)",
	}
	for _, q := range stmts {
		if _, err := executeWrite(ctx, q, nil); err != nil {
			return err
		}
	}
	return nil
}
