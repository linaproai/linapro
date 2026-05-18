## ADDED Requirements

### Requirement: REST semantics and path style unification

Backend APIs SHALL maintain consistency in path, HTTP method, and resource semantics. Read operations, write operations, and deletions SHALL each use the repository's agreed REST conventions.

#### Scenario: Defining read endpoints

- **WHEN** defining list queries, detail queries, option queries, tree queries, or export endpoints
- **THEN** the endpoint uses `GET`
- **AND** the path uses resource-based naming rather than action-based naming

#### Scenario: Defining write-operation endpoints

- **WHEN** defining create, update, status-change, import, or delete endpoints
- **THEN** the endpoint uses `POST`, `PUT`, or `DELETE` respectively
- **AND** the same resource family maintains consistent path style across the repository

### Requirement: Path parameter binding unification

Backend API DTOs SHALL unify path parameter declaration and binding conventions, avoiding mixed parameter styles within the same repository.

#### Scenario: Declaring path parameters

- **WHEN** an endpoint path contains a resource identifier or sub-resource identifier
- **THEN** `g.Meta` uses `{param}` syntax to declare path parameters
- **AND** input DTO fields use `json:"param"` for parameter naming; the repository does not mix `p` and `json` tags
- **AND** output DTOs continue using `json` tags for response fields

### Requirement: API documentation tag completeness

Backend API input/output structures SHALL provide clear `dc` and `eg` tags for all visible fields, ensuring auto-generated OpenAPI documentation is directly understandable and debuggable.

#### Scenario: Defining input/output fields

- **WHEN** adding or modifying input/output fields in API DTOs
- **THEN** each field includes a `dc` tag describing its business meaning
- **AND** each field includes a directly debuggable `eg` example value

#### Scenario: Defining enum or optional fields

- **WHEN** a field represents a status, type, toggle, or optional filter condition
- **THEN** the `dc` tag explicitly lists all possible values and their meanings
- **AND** documentation does not require callers to read implementation code to infer parameter semantics

### Requirement: RESTful batch delete endpoints

The system SHALL provide batch delete endpoints for users and roles using `DELETE` with repeated query parameters, reusing all single-record protection rules inside a single transaction.

#### Scenario: Batch delete with repeated query parameters

- **WHEN** a caller invokes `DELETE /api/v1/user?ids=1&ids=2&ids=3` or `DELETE /api/v1/role?ids=1&ids=2&ids=3`
- **THEN** the system processes all IDs in a single transaction
- **AND** any protection rule violation (built-in admin, current user, super administrator role) rejects the entire batch
- **AND** the DTO uses `Ids []int json:"ids" v:"required|min-length:1"` with English `dc` and `eg` tags
- **AND** the `g.Meta` carries the corresponding permission tag (`system:user:remove` or `system:role:remove`)
