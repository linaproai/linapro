## ADDED Requirements

### Requirement: The config-management component must have a unit-test coverage gate
The system SHALL maintain repeatable unit tests for the `apps/lina-core/internal/service/config` config-management component, and SHALL use package-level coverage verification as a delivery gate before that component is considered ready.

#### Scenario: Package-level coverage meets the delivery bar
- **WHEN** a maintainer runs `go test ./internal/service/config -cover` from `apps/lina-core`
- **THEN** the command succeeds
- **AND** the reported package-level statement coverage is not lower than `80%`

### Requirement: Critical config-management branches must have automated regression protection
The system SHALL add automated unit tests for critical helper logic inside the config-management component, including high-risk branches around defaults and fallbacks, cache or snapshot reuse, and invalid input or error propagation.

#### Scenario: Plugin and public-frontend config helper logic changes
- **WHEN** a change touches plugin dynamic storage paths, protected public-frontend config key checks, or the shared validation entry point
- **THEN** unit tests cover the normal read path
- **AND** cover default-value or compatibility-fallback behavior
- **AND** cover invalid input or empty-value defensive behavior

#### Scenario: Runtime-parameter cache and revision synchronization logic changes
- **WHEN** a change touches runtime-parameter snapshot caches, the revision controller, or shared-KV synchronization logic
- **THEN** unit tests cover cache-hit or local-reuse behavior
- **AND** cover rebuilds after revision changes
- **AND** cover error propagation and defensive behavior for shared-KV read failures, invalid cached values, or equivalent exceptional cases
