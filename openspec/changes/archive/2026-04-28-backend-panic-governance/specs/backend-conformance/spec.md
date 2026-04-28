## ADDED Requirements

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
