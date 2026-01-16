## ADDED Requirements
### Requirement: Data Authenticity
The system MUST NOT return hardcoded mock data for market sentiment or stock information in production environments or default fallback paths.

#### Scenario: API Failure
- **WHEN** the external data provider API fails or returns invalid data
- **THEN** the system MUST return an error or an empty dataset, NOT fake data
