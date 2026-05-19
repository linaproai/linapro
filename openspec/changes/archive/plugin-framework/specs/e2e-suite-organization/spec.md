## Requirements

### Requirement: E2E Test Discovery Must Support Host-Only and Plugin-Full Modes

The E2E test suite SHALL clearly distinguish host tests and source plugin-owned tests. Host test discovery and execution must not require `apps/lina-plugins` to exist; plugin test discovery and execution must require the official plugin workspace to be initialized when explicitly requesting plugin scope.

#### Scenario: Host-only E2E does not require plugin workspace
- **WHEN** `apps/lina-plugins` does not exist or is empty
- **AND** a developer runs the host E2E scope
- **THEN** Playwright only discovers host-owned tests
- **AND** test discovery does not fail due to missing plugin test directories

#### Scenario: Explicit plugin E2E missing plugin workspace
- **WHEN** `apps/lina-plugins` does not exist or is empty
- **AND** a developer runs `plugins` or `plugin:<plugin-id>` test scope
- **THEN** the test entry fails fast
- **AND** the error prompts to initialize the official plugin submodule

#### Scenario: Plugin-full E2E discovers plugin-owned tests
- **WHEN** `apps/lina-plugins` submodule initialization is complete
- **AND** a developer runs `plugins` or `plugin:<plugin-id>` test scope
- **THEN** Playwright discovers the corresponding plugin's `hack/tests/e2e/TC*.ts`
- **AND** TC number uniqueness and directory attribution verification covers host and plugin tests
