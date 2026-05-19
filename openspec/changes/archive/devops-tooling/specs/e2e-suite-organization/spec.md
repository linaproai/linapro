## ADDED Requirements

### Requirement: Nightly full E2E must cover host and official plugin tests

The nightly verification pipeline's full E2E SHALL simultaneously execute host E2E and official plugin self-owned E2E. The test entry SHALL use the existing E2E governance scope selection `e2e` and `plugins`, covering the complete browser regression scope that the release workflow does not execute.

#### Scenario: Nightly full E2E selects host and plugin scope
- **WHEN** the nightly workflow executes full E2E
- **THEN** the E2E runner selects the host `e2e` scope
- **AND** the E2E runner selects the official plugin `plugins` scope
- **AND** Playwright discovers and executes plugin self-owned cases from `apps/lina-plugins/<plugin-id>/hack/tests/e2e/TC*.ts`

#### Scenario: Missing plugin E2E blocks nightly image publishing
- **WHEN** the nightly workflow executes full E2E
- **AND** the official plugin workspace is missing, empty, or the plugin E2E scope resolves to empty
- **THEN** the E2E phase fails
- **AND** the nightly image publishing job MUST NOT execute

#### Scenario: E2E failure evidence is uploaded
- **WHEN** the nightly full E2E completes or fails
- **THEN** the workflow uploads Playwright report, test-results, backend logs, and frontend logs
- **AND** artifact names can distinguish the nightly execution source

### Requirement: E2E official plugin verification must be semantically distinct from host-only tests

The E2E test suite SHALL preserve the semantic distinction between host-only and plugin-full. Workflows that enable full E2E SHALL simultaneously run host-only E2E and plugin-full E2E; plugin-full E2E MUST NOT be replaced by host-only E2E, and official plugin workspace absence MUST NOT silently degrade to host-only E2E.

#### Scenario: Both host-only and plugin-full E2E run when full E2E is enabled
- **WHEN** a workflow executes the full E2E gate
- **THEN** the workflow SHALL run the host-only E2E entry
- **AND** the workflow SHALL run the plugin-full E2E entry
- **AND** downstream image publishing jobs SHALL wait for both E2E jobs to succeed

#### Scenario: Plugin-full E2E does not degrade to host-only E2E
- **WHEN** a workflow executes full E2E
- **AND** official plugin tests cannot be discovered
- **THEN** the workflow reports plugin-full verification failure
- **AND** the workflow MUST NOT treat running only host E2E results as full verification passing

#### Scenario: Host-only test entry does not affect plugin-full full verification
- **WHEN** the repository has a host-only E2E entry
- **AND** a workflow executes plugin-full E2E
- **THEN** the workflow MUST NOT use the host-only E2E entry to replace plugin-full E2E
- **AND** the workflow's test log should show that the plugin test scope has been selected
