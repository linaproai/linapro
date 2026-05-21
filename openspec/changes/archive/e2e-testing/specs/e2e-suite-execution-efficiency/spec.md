## ADDED Requirements

### Requirement: Plugin-full E2E must support generic plugin entry shard execution
E2E CI workflow SHALL allow plugin-full browser regression to execute plugin management, plugin-owned tests, and plugin host seam tests as independent shards while preserving the same plugin-full startup semantics.

#### Scenario: Plugin-full shard selects module scope
- **WHEN** workflow starts plugin-full E2E shard
- **THEN** each shard must use plugin-full service startup command
- **AND** source plugin own test shards must use `plugins` or `plugin:<plugin-id>` generic entry to select test scope
- **AND** root directory shard can only select host plugin framework generic test scope, not root test file sets depending on specific official source plugins
- **AND** shard logs must display selected scope, parallel file count, serial file count, and serial isolation category

#### Scenario: Plugin-full does not maintain official plugin business alias scopes
- **WHEN** developer needs to run source plugin own E2E
- **THEN** runner must support `plugins` to run all source plugin own cases
- **AND** runner must support `plugin:<plugin-id>` to run a single source plugin's own cases
- **AND** E2E manifest should not maintain long-term alias scopes for official plugin business modules

#### Scenario: Plugin-full shard failure blocks downstream release
- **WHEN** any plugin-full E2E shard fails
- **THEN** the complete verification suite must fail
- **AND** image release or subsequent jobs depending on verification success must not execute

#### Scenario: Plugin-full shard uploads independent diagnostic evidence
- **WHEN** plugin-full E2E shard completes or fails
- **THEN** workflow must upload that shard's Playwright report, test-results, backend logs, and frontend logs
- **AND** artifact name must contain caller prefix and shard identifier, avoiding overwriting other shard evidence

### Requirement: E2E authentication page fixture must support skipping default dashboard navigation
E2E test suite SHALL provide an authenticated page fixture that creates a browser page with admin storage state without automatically navigating to the default dashboard.

#### Scenario: Test directly enters target business route
- **WHEN** test uses lightweight authenticated page fixture
- **THEN** fixture must create a page within logged-in context
- **AND** fixture must not automatically visit `/dashboard/analytics`
- **AND** test must explicitly navigate to its own target route

#### Scenario: Old adminPage fixture maintains compatibility
- **WHEN** existing tests continue using `adminPage`
- **THEN** fixture must maintain existing default dashboard available semantics
- **AND** must not require one-time migration of all existing E2E cases

### Requirement: Ordinary plugin function E2E must reuse idempotent plugin baseline
E2E test suite SHALL provide reusable plugin baseline setup for ordinary plugin-owned page tests so they do not repeatedly synchronize, install, enable, seed, and refresh plugin projection in every test case.

#### Scenario: Plugin function test declares needed plugin collection
- **WHEN** ordinary plugin function test needs one or more source plugins in available state
- **THEN** test or test suite must declare needed plugin collection through shared baseline helper
- **AND** baseline must idempotently execute plugin synchronization, installation, enablement, available mock data loading, and plugin projection refresh

#### Scenario: Plugin lifecycle test does not use ordinary baseline to cover tested state
- **WHEN** test target is plugin installation, enablement, disablement, uninstallation, upload, synchronization, or upgrade lifecycle
- **THEN** test must continue to explicitly control its own initial state and cleanup logic
- **AND** ordinary plugin baseline must not implicitly change tested plugin state in these tests

### Requirement: E2E optimization must preserve quantifiable timing verification
E2E runtime optimization SHALL preserve per-test timing evidence and compare before/after wall clock for host-only and plugin-full validation.

#### Scenario: Optimization preserves test timing records
- **WHEN** E2E workflow run completes
- **THEN** Playwright output or artifact must preserve per-test-case timing records
- **AND** CI logs must be sufficient to identify slowest file, slowest case, and shard wall clock

#### Scenario: Host-only and Plugin-full target time is reviewable
- **WHEN** this change completes verification
- **THEN** task records must document host-only and plugin-full optimization before/after time comparison
- **AND** if target time is not reached, must document remaining bottleneck and subsequent optimization scope

### Requirement: Host-only single module E2E must exclude plugin environment cases
E2E runner SHALL provide a host-only module entrypoint for running a selected host scope without requiring the official plugin workspace.

#### Scenario: Running host module without initialized official plugin workspace
- **WHEN** developer runs host-only module scope in environment without `apps/lina-plugins`
- **THEN** runner must only select host cases in that scope not depending on official plugin workspace
- **AND** runner must not require initialization of official plugin submodule

#### Scenario: Host-only module rejects plugin scope
- **WHEN** developer selects a scope requiring `apps/lina-plugins` through host-only module entry
- **THEN** runner must fail and explain that scope cannot run in host-only module mode

### Requirement: CI database health check must use explicit PostgreSQL user and database
Browser E2E CI SHALL configure PostgreSQL service health checks with explicit user and database parameters instead of relying on the runner OS user.

#### Scenario: PostgreSQL health check does not use runner user
- **WHEN** browser E2E workflow starts PostgreSQL service
- **THEN** health check command must explicitly specify `postgres` user and `linapro` database
- **AND** CI logs must not repeatedly output invalid role errors due to health check defaulting to runner user
