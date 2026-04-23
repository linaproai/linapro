## ADDED Requirements

### Requirement: The E2E suite MUST provide layered execution entrypoints
The E2E suite SHALL provide both full-regression and fast-feedback entrypoints. At minimum, it SHALL support `smoke`, module-scoped execution, and `full`, so developers do not need to handcraft file globs or wait for the entire suite in every feedback cycle.

#### Scenario: Developer needs fast high-value feedback
- **WHEN** a developer runs the smoke entrypoint
- **THEN** the system MUST execute a curated set of high-value critical-path test files
- **AND** the developer does not need to maintain or pass an explicit file list by hand

#### Scenario: Developer wants to validate only the affected module
- **WHEN** a developer runs the module entrypoint and provides a scope
- **THEN** the system MUST resolve that scope through a predefined module mapping
- **AND** the system MUST run the corresponding directories or files instead of relying on ad-hoc temporary globs

#### Scenario: Full regression entrypoint remains available
- **WHEN** a developer or CI runs the full-regression entrypoint
- **THEN** the system MUST execute every test file that matches the `TC*.ts` suite convention

### Requirement: High-frequency authenticated tests MUST reuse pre-generated login state
Except for authentication-focused scenarios themselves, high-frequency logged-in management-workbench tests SHALL reuse pre-generated authenticated state instead of performing a full UI login in every test.

#### Scenario: Ordinary back-office tests reuse login state by default
- **WHEN** a test file uses an administrator logged-in page fixture
- **THEN** that fixture MUST load its page context from a pre-generated `storageState` or an equivalent authenticated-state artifact
- **AND** it MUST NOT perform a full login-page flow for every test by default

#### Scenario: Authentication tests still validate the real login flow
- **WHEN** a test targets login, logout, unauthenticated redirect, failed login, or another authentication behavior
- **THEN** that test MUST still be able to exercise the real authentication flow explicitly
- **AND** it MUST NOT be silently replaced by shared authenticated state

#### Scenario: Expired login state can be regenerated
- **WHEN** the pre-generated login state is missing, expired, or invalid
- **THEN** the suite preparation step MUST regenerate a usable authenticated-state artifact for later tests

### Requirement: High-cost waits MUST converge to state-based waits with explicit parallel-safety boundaries
The E2E suite SHALL prefer page, API, and component state as readiness signals, and SHALL define clear serial-versus-parallel boundaries between shared-state tests and independently runnable tests.

#### Scenario: Page objects wait for deterministic readiness signals
- **WHEN** a page object waits for tables, drawers, modals, route changes, or feedback messages to become ready
- **THEN** it MUST prefer deterministic signals such as element visibility, loading completion, API completion, or route transitions
- **AND** it MUST NOT rely on fixed-duration sleeps as its primary strategy

#### Scenario: Independent test files enter the parallel pool
- **WHEN** a test file does not rely on cross-file shared global state
- **THEN** it MUST be eligible for inclusion in a limited-worker parallel execution pool to shorten regression time

#### Scenario: Shared-state test files stay inside the serial boundary
- **WHEN** a test file mutates plugin lifecycle state, global configuration, permission matrices, or another cross-file shared state
- **THEN** it MUST be explicitly classified into the serial execution boundary
- **AND** it MUST NOT run concurrently with unrelated files if that would create instability
