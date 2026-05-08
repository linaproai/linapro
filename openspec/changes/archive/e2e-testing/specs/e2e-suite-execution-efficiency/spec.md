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

### Requirement: E2E shared global state tests MUST declare isolation categories
The E2E suite SHALL classify tests that mutate or depend on cross-file shared global state. Files that mutate plugin lifecycle, runtime i18n bundle versions, public frontend configuration, system parameters, dictionaries, menu or role permission matrices, shared database seed data, or filesystem-backed plugin artifacts MUST declare an isolation category and MUST be routed into an execution boundary that prevents unsafe parallel overlap.

#### Scenario: Plugin lifecycle test is classified as serial
- **WHEN** a test file installs, enables, disables, uninstalls, uploads, syncs, or upgrades a plugin
- **THEN** the E2E execution manifest MUST classify that file with a plugin lifecycle isolation category
- **AND** the full-regression runner MUST keep that file out of the parallel pool

#### Scenario: Global configuration test is classified as serial
- **WHEN** a test file mutates system parameters, public frontend configuration, dictionaries, menu permissions, role permissions, or other shared governance data
- **THEN** the E2E execution manifest MUST classify that file with the matching shared-state category
- **AND** the full-regression runner MUST keep that file out of the parallel pool unless an explicit documented safe exception exists

#### Scenario: Validator rejects unclassified high-risk tests
- **WHEN** the E2E validator detects high-risk operations in a test file that is not serial or not classified
- **THEN** validation MUST fail with a message that identifies the file, detected risk category, and expected manifest action

### Requirement: E2E cache revalidation tests MUST tolerate legitimate global version refreshes
The E2E suite SHALL test cache and ETag behavior by validating protocol semantics instead of assuming that global resource versions remain unchanged for the duration of a full regression run. Conditional requests MUST verify the request precondition and the correctness of either a not-modified response or a refreshed-resource response.

#### Scenario: Conditional request hits unchanged resource version
- **WHEN** a cache test sends a conditional request with an ETag that still matches the current resource version
- **THEN** the test MUST accept a `304 Not Modified` response
- **AND** it MUST verify that the returned ETag matches the cached ETag and that no body is required

#### Scenario: Conditional request observes refreshed resource version
- **WHEN** a cache test sends a conditional request with an ETag that no longer matches because another legitimate test or lifecycle operation refreshed the resource version
- **THEN** the test MUST accept a `200 OK` response only if the response includes a new ETag that differs from the cached ETag
- **AND** it MUST verify that the refreshed response body is present and valid

#### Scenario: Cache test still verifies conditional request behavior
- **WHEN** a cache test reloads a page or resource that should use persistent cache metadata
- **THEN** the test MUST verify that the request carries the expected conditional header or equivalent cache precondition
- **AND** it MUST NOT pass solely because the resource endpoint returned a successful body

### Requirement: E2E prerequisites MUST be fixture-owned and idempotent
The E2E suite SHALL make plugin state, mock data, authenticated state, and shared filesystem prerequisites explicit through reusable fixtures or support helpers. A test file MUST be independently runnable without relying on another test file to create plugin rows, install source plugins, load mock SQL, refresh frontend plugin projection, or create reusable authenticated state.

#### Scenario: Test depends on a source plugin
- **WHEN** a test file needs a source plugin page, API, menu, or mock data
- **THEN** the test MUST call a shared fixture/helper that idempotently syncs, installs, enables, and refreshes the plugin projection
- **AND** the helper MUST load plugin mock SQL only when the plugin provides a matching mock-data resource

#### Scenario: Test depends on generated user or business data
- **WHEN** a test file creates users, departments, posts, notices, files, plugin records, or import/export data
- **THEN** the test MUST use unique names or stable test prefixes
- **AND** it MUST clean up its own data in `finally`, `afterEach`, or `afterAll` without depending on cross-file cleanup

#### Scenario: Test reads business state under localized UI
- **WHEN** a test needs to compare business counts, identities, permissions, or state transitions under different languages
- **THEN** it MUST use stable API fields such as IDs, codes, permission keys, label keys, or numeric counters for the business assertion
- **AND** localized UI text MUST be asserted separately as presentation behavior

### Requirement: E2E full-regression reports MUST expose serial and parallel boundaries
The E2E full-regression runner SHALL report enough information to make execution isolation auditable. Reports MUST show which files were run in the parallel pool, which files were run in the serial pool, and which isolation categories caused files to be serialized.

#### Scenario: Full regression starts
- **WHEN** a developer or CI starts the full-regression entrypoint
- **THEN** the runner MUST print or persist a summary of parallel file count, serial file count, and configured worker count
- **AND** it MUST include the isolation categories represented in the serial set

#### Scenario: Module-scoped regression starts
- **WHEN** a developer runs a module-scoped E2E command
- **THEN** the runner MUST apply the same serial-versus-parallel split to the resolved module files
- **AND** it MUST report any serialized files and categories within that module scope
