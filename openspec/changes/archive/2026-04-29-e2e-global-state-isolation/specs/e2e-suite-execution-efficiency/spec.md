## ADDED Requirements

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
