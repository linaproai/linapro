## ADDED Requirements

### Requirement: Controller and service layer implementation constraints

Backend production code SHALL follow the GoFrame v2 layered conventions defined by the repository: controller dependencies are injected through constructor functions, and service components are organized by convention directories and naming.

#### Scenario: Controller dependency initialization

- **WHEN** a controller depends on one or more service components
- **THEN** those dependencies are initialized in the corresponding `_new.go` constructor function
- **AND** interface methods do not call `service.New()` internally to create dependencies

#### Scenario: Service component file splitting

- **WHEN** a service component has multiple responsibility sub-modules
- **THEN** code is split into independent files by component prefix and sub-module suffix
- **AND** bare filenames unrelated to the component name are not used to host sub-module logic

### Requirement: ORM and soft-delete conformance

Backend production code SHALL use GoFrame-recommended ORM patterns for database access and follow automatic soft-delete and timestamp maintenance conventions.

#### Scenario: Querying soft-delete tables

- **WHEN** code queries a table that contains a `deleted_at` field
- **THEN** the query logic relies on GoFrame automatic soft-delete filtering
- **AND** production code does not hand-write `WhereNull(deleted_at)` or equivalent SQL conditions

#### Scenario: Updating and writing data

- **WHEN** code performs database writes, updates, or association maintenance
- **THEN** production code uses DO objects to pass `Data`
- **AND** does not manually maintain `created_at`, `updated_at`, or `deleted_at` fields that are handled by the framework

### Requirement: Exported symbol documentation completeness

Backend exported methods, structs, and key public fields SHALL carry comments that follow Go documentation conventions, suitable for doc generation and long-term maintenance.

#### Scenario: Adding or modifying exported symbols

- **WHEN** code contains exported methods, exported structs, or key exported fields
- **THEN** their declarations are preceded by adjacent, semantically clear comments
- **AND** comments are recognizable by Go doc, not just separator remarks or detached notes

### Requirement: Runtime errors must not replace explicit error handling with panic

Production backend code SHALL use `panic` only for startup, initialization, unrecoverable critical paths, `Must*` semantic constructors, or unknown panic rethrow scenarios. Ordinary requests, import/export flows, dynamic plugin input, runtime configuration reads, and recoverable resource handling paths MUST use explicit `error` returns, unified error responses, or controlled degradation.

#### Scenario: Startup unrecoverable errors use fail-fast

- **WHEN** the backend detects an unrecoverable error during process startup, driver registration, command tree initialization, or source-plugin static registration
- **THEN** the code MAY use `panic` to fail the process fast
- **AND** the panic call site MUST be in the allowlist with a reason for retaining it

#### Scenario: Ordinary business requests return errors

- **WHEN** an ordinary HTTP request, file import/export, Excel generation, or resource close operation encounters a recoverable error
- **THEN** the service or controller MUST return `error` so the unified error handling chain can generate the response
- **AND** it MUST NOT use `panic` instead of returning the error

#### Scenario: Dynamic plugin input validation fails

- **WHEN** a dynamic plugin artifact, manifest, hostServices declaration, or authorization input is invalid
- **THEN** the host MUST return a validation error with context
- **AND** plugin-provided dynamic input MUST NOT trigger a production-code panic

#### Scenario: Invalid runtime configuration values return explicitly

- **WHEN** a protected runtime configuration value has a parsing error while a snapshot is being read
- **THEN** the backend MUST expose the configuration problem through an explicit `error` return or unified error response
- **AND** write paths MUST still keep strict validation so normal management entries cannot save invalid values

#### Scenario: New panics are constrained by static checks

- **WHEN** a developer adds a `panic` call in production backend Go code
- **THEN** automated checks MUST require the call site to match the allowlist
- **AND** the allowlist entry MUST document its category and retained reason
