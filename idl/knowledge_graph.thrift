namespace go knowledge_graph

enum EntityType {
  STOCK = 1
  COMPANY = 2
  INDUSTRY = 3
  SECTOR = 4
  EVENT = 5
  POLICY = 6
  MACRO = 7
  FLOW = 8
}

enum RelationType {
  BELONGS_TO = 1
  AFFECTS = 2
  BENEFITS = 3
  FLOWS_INTO = 4
  SUPPLIES = 5
}

enum EventType {
  NEWS = 1
  POLICY = 2
  REPORT = 3
  RUMOR = 4
  CAPITAL = 5
}

struct Entity {
  1: string id
  2: EntityType type
  3: string name
  4: map<string, string> attributes
}

struct Evidence {
  1: string id
  2: string category
  3: string summary
  4: i32 confidence
  5: string source
  6: i64 timestamp
}

struct Relation {
  1: Entity source
  2: Entity target
  3: RelationType type
  4: double strength
  5: Evidence evidence
  6: i64 updated_at
}

struct Event {
  1: string id
  2: EventType type
  3: list<Entity> entities
  4: i32 impact_direction
  5: i32 impact_strength
  6: i32 confidence
  7: i64 timestamp
  8: string source
  9: string dedupe_key
}

struct EvidenceBundle {
  1: list<Entity> entities
  2: list<Relation> relations
  3: list<Event> events
  4: double score
  5: i32 coverage
}

struct GetGraphNeighborhoodRequest {
  1: EntityType entity_type
  2: string entity_id
  3: list<RelationType> relation_types
  4: i32 depth
  5: i64 start_time
  6: i64 end_time
  7: i32 max_edges
}

struct GetGraphNeighborhoodResponse {
  1: list<Entity> entities
  2: list<Relation> relations
}

struct GetEntityProfileRequest {
  1: EntityType entity_type
  2: string entity_id
  3: i64 start_time
  4: i64 end_time
}

struct GetEntityProfileResponse {
  1: Entity entity
  2: list<Relation> relations
  3: list<Event> events
}

struct SearchEventsRequest {
  1: EntityType entity_type
  2: string entity_id
  3: list<EventType> event_types
  4: i64 start_time
  5: i64 end_time
  6: i32 min_confidence
  7: i32 limit
}

struct SearchEventsResponse {
  1: list<Event> events
}

struct GetEvidenceBundleRequest {
  1: string stock_code
  2: i64 start_time
  3: i64 end_time
  4: i32 min_confidence
  5: i32 max_items
}

struct GetEvidenceBundleResponse {
  1: EvidenceBundle bundle
}

struct GetRelationStrengthRequest {
  1: EntityType source_type
  2: string source_id
  3: EntityType target_type
  4: string target_id
  5: RelationType relation_type
}

struct GetRelationStrengthResponse {
  1: Relation relation
}

struct UpsertEntitiesRequest {
  1: list<Entity> entities
}

struct UpsertEntitiesResponse {
  1: i32 upserted
}

struct UpsertRelationsRequest {
  1: list<Relation> relations
}

struct UpsertRelationsResponse {
  1: i32 upserted
}

struct UpsertEventsRequest {
  1: list<Event> events
}

struct UpsertEventsResponse {
  1: i32 upserted
}

struct UpsertEvidencesRequest {
  1: list<Evidence> evidences
}

struct UpsertEvidencesResponse {
  1: i32 upserted
}

service KnowledgeGraphService {
  GetGraphNeighborhoodResponse GetGraphNeighborhood(1: GetGraphNeighborhoodRequest req)
  GetEntityProfileResponse GetEntityProfile(1: GetEntityProfileRequest req)
  SearchEventsResponse SearchEvents(1: SearchEventsRequest req)
  GetEvidenceBundleResponse GetEvidenceBundle(1: GetEvidenceBundleRequest req)
  GetRelationStrengthResponse GetRelationStrength(1: GetRelationStrengthRequest req)
  UpsertEntitiesResponse UpsertEntities(1: UpsertEntitiesRequest req)
  UpsertRelationsResponse UpsertRelations(1: UpsertRelationsRequest req)
  UpsertEventsResponse UpsertEvents(1: UpsertEventsRequest req)
  UpsertEvidencesResponse UpsertEvidences(1: UpsertEvidencesRequest req)
}
