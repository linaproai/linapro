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

#### Scenario: Reusing cross-module public enum contracts

- **WHEN** backend API DTO fields use stable enums or value types shared across multiple API modules
- **THEN** the DTO uses public components in `apps/lina-core/pkg` organized by contract boundary, or existing stable public components
- **AND** public components SHALL use small domain-contract naming; a unified `pkg/enums` package MUST NOT be created
- **AND** domain-private enums SHALL remain in their owning API or service package, not merged solely to eliminate literal duplication
- **AND** type abstraction MUST NOT change existing JSON field names, field values, defaults, or API documentation example semantics

### Requirement: RESTful batch delete endpoints

The system SHALL provide batch delete endpoints for users and roles using `DELETE` with repeated query parameters, reusing all single-record protection rules inside a single transaction.

#### Scenario: Batch delete with repeated query parameters

- **WHEN** a caller invokes `DELETE /api/v1/user?ids=1&ids=2&ids=3` or `DELETE /api/v1/role?ids=1&ids=2&ids=3`
- **THEN** the system processes all IDs in a single transaction
- **AND** any protection rule violation (built-in admin, current user, super administrator role) rejects the entire batch
- **AND** the DTO uses `Ids []int json:"ids" v:"required|min-length:1"` with English `dc` and `eg` tags
- **AND** the `g.Meta` carries the corresponding permission tag (`system:user:remove` or `system:role:remove`)

### Requirement: Response instant fields use Unix timestamps in milliseconds

Public HTTP JSON response DTO fields that represent exact instants SHALL return Unix timestamps in milliseconds. The JSON type MUST be numeric, and the Go DTO type MUST be `int64` or `*int64`. Public response DTOs MUST NOT directly expose `time.Time`, `*time.Time`, `gtime.Time`, `*gtime.Time`, or formatted strings for instant fields. This rule only constrains the API response DTO boundary and does not constrain internal `dao`, `entity`, `do`, or `service` models.

#### Scenario: Defining an instant response field

- **WHEN** a developer adds or modifies `createdAt`, `updatedAt`, `deletedAt`, `loginAt`, `startedAt`, `endedAt`, `expiredAt`, or another public HTTP JSON response DTO field that represents an exact instant
- **THEN** the field type is `int64` or `*int64`
- **AND** the field value represents a Unix timestamp in milliseconds
- **AND** the field `dc` tag or API documentation includes `Unix timestamp in milliseconds`

#### Scenario: Defining a calendar-date response field

- **WHEN** a developer defines `birthday`, `businessDate`, `periodDate`, or another public HTTP JSON response DTO field that represents only a calendar date and not an exact instant
- **THEN** the field may use a `YYYY-MM-DD` string
- **AND** the field `dc` tag or API documentation clearly states `date-only` semantics
- **AND** callers do not need to infer an exact instant in the server time zone from the field

#### Scenario: Migrating an existing time-field contract

- **WHEN** a developer plans to migrate an existing public API instant field from a string or time-object response to a Unix timestamp in milliseconds
- **THEN** the migration is evaluated through the active OpenSpec change that owns the API contract change
- **AND** frontend parsing, page formatting, `apidoc i18n` resources, and related test impact are evaluated together

### Requirement: API response must not directly expose database entities

Backend API response DTOs SHALL use independent response struct definitions for caller-visible fields, and MUST NOT directly embed or return DAO-generated database entity types. The response boundary SHALL explicitly map only the fields the current interface needs, avoiding passive changes to the external API contract when database fields are added, renamed, or internal governance fields change.

#### Scenario: Defining read response DTOs

- **WHEN** an API defines list, detail, option, tree, file info, or current user profile read responses
- **THEN** the response type SHALL use an independent DTO within the API package
- **AND** the response type SHALL NOT embed `entity.*`
- **AND** response lists SHALL NOT directly use `[]*entity.*`

#### Scenario: Mapping entities to response DTOs

- **WHEN** controllers or response assembly logic generate API responses from service-layer entities, projections, or domain objects
- **THEN** response assembly SHALL map fields one by one for only permitted fields
- **AND** database entity pointers MUST NOT be directly assigned to response fields

#### Scenario: Hiding internal and sensitive fields

- **WHEN** database entities contain passwords, credentials, soft-delete timestamps, storage paths, hashes, internal tenant governance fields, cache versions, or other internal implementation fields
- **THEN** API response DTOs SHALL NOT expose these fields by default
- **AND** only fields that are explicitly part of the current interface contract and have been evaluated for permissions, data permissions, and business boundaries may be added to response DTOs
