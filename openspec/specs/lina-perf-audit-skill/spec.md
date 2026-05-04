# lina-perf-audit-skill Specification

## Purpose

Define the governance contract for the manually triggered `lina-perf-audit` agent skill, which audits LinaPro backend APIs for performance risks and read-request write side effects across the host and all built-in plugins.

## Requirements

### Requirement: Manual trigger only

The `lina-perf-audit` skill SHALL declare itself as manual-trigger-only and MUST NOT be invoked by automation, CI/CD pipelines, scheduled jobs, git hooks, or other skills. Ambiguous performance requests MUST require user confirmation before any destructive audit preparation runs.

#### Scenario: SKILL.md frontmatter declares the trigger constraint

- **WHEN** `.agents/skills/lina-perf-audit/SKILL.md` is loaded
- **THEN** its description contains the phrase `MANUAL TRIGGER ONLY`
- **AND** it lists the expected resource cost, including database reset, service restart, built-in plugin enablement, stress fixtures, sub-agent fan-out, elapsed time, and token cost
- **AND** it lists the forbidden invocation paths, including other skills, CI, scheduled jobs, git hooks, and ambiguous performance requests

#### Scenario: Ambiguous requests require confirmation

- **WHEN** the user says something ambiguous such as `the API seems slow`, `how is performance`, or `check interface performance`
- **THEN** the skill asks whether to start the full audit before running any Stage 0 setup command
- **AND** the confirmation text mentions database reset, service restart, elapsed time, sub-agent fan-out, and token cost
- **AND** the skill does not run `make stop`, `make init`, `make mock`, `setup-audit-env.sh`, `prepare-builtin-plugins.sh`, or `stress-fixture.sh` before confirmation

#### Scenario: Other skills do not invoke this skill automatically

- **WHEN** another skill such as `lina-review`, `lina-feedback`, or `lina-e2e` observes a performance concern
- **THEN** that skill MUST NOT invoke `lina-perf-audit`
- **AND** it may only suggest that the user manually trigger `lina-perf-audit`

### Requirement: Three-stage audit workflow

The `lina-perf-audit` skill MUST execute a full audit in three stages: Stage 0 preparation, Stage 1 concurrent sub-agent audit, and Stage 2 summary plus issue-card aggregation. Single-run artifacts MUST be written under `temp/lina-perf-audit/<run-id>/`.

#### Scenario: Helper scripts stay inside the skill boundary

- **WHEN** deterministic helper logic is needed for setup, plugin preparation, endpoint scanning, fixture probing, stress fixtures, or report aggregation
- **THEN** the scripts live under `.agents/skills/lina-perf-audit/scripts/`
- **AND** `SKILL.md`, references, OpenSpec documents, and issue-card templates use the skill-owned script paths
- **AND** no second copy of the skill-private scripts is maintained under `hack/`

#### Scenario: Stage 0 prepares a complete audit environment

- **WHEN** the user has confirmed the full audit
- **THEN** the skill creates a unique `run-id` in `YYYYMMDD-HHMMSS` format
- **AND** it stops services, resets the local database with `make init confirm=init rebuild=true` and `make mock confirm=mock`, patches audit logging through `setup-audit-env.sh`, installs and enables all built-in plugins through `prepare-builtin-plugins.sh`, adds audit-only stress data through `stress-fixture.sh`, scans endpoints through `scan-endpoints.sh`, and probes fixtures through `probe-fixtures.sh`
- **AND** all generated artifacts are written under `temp/lina-perf-audit/<run-id>/`
- **AND** temporary logger settings are restored on success or failure through `restore-audit-env.sh`

#### Scenario: Stage 1 uses sub agents for endpoint tasks

- **WHEN** Stage 0 completes
- **THEN** the skill builds endpoint tasks from `catalog.json` for host and built-in plugin modules
- **AND** every API endpoint audit task is executed by a sub agent rather than serially by the main agent
- **AND** large modules are split into endpoint or small-module shards so each sub-agent prompt remains below the configured prompt budget
- **AND** each sub agent writes exactly its assigned audit output to `temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`

#### Scenario: Stage 2 produces stable outputs

- **WHEN** all sub agents finish
- **THEN** the main agent runs report aggregation
- **AND** `SUMMARY.md` lists findings by HIGH, MEDIUM, and LOW severity
- **AND** `meta.json` records run timing, git commit, stress-fixture status, sub-agent status, skipped plugins, logger settings, and restore result
- **AND** every new or updated persistent issue card is linked from the run summary

### Requirement: All built-in plugins are covered

The skill MUST audit backend APIs for every built-in plugin in the repository and MUST NOT limit coverage to plugins already installed or enabled before the audit starts.

#### Scenario: Stage 0 installs and enables built-in plugins

- **WHEN** Stage 0 prepares the audit environment
- **THEN** it discovers plugins from `apps/lina-plugins/*/plugin.yaml`
- **AND** it syncs, installs, and enables every built-in plugin through host plugin management APIs
- **AND** it loads plugin mock data when a plugin provides `manifest/sql/mock-data/`
- **AND** any install or enable failure fails Stage 0 with the failing plugin list
- **AND** a successfully discovered plugin with no backend API is recorded as skipped with reason `no backend API`

#### Scenario: Endpoint catalog includes host and plugin DTOs

- **WHEN** `scan-endpoints.sh` generates `catalog.json`
- **THEN** the catalog includes route metadata from `apps/lina-core/api/**/v1/*.go` and `apps/lina-plugins/*/backend/api/**/v1/*.go`
- **AND** declared plugin routes that are not reachable at runtime fail setup or fixture probing instead of silently entering normal audit output

### Requirement: Trace-ID based SQL evidence

Sub agents MUST use GoFrame's default `Trace-ID` response header to correlate endpoint calls with SQL log lines. The skill MUST NOT add audit-only middleware, custom response headers, or production behavior changes.

#### Scenario: Sub agent obtains trace ID from the response

- **WHEN** a sub agent calls an endpoint
- **THEN** it reads the `Trace-ID` response header
- **AND** it uses that trace ID to find matching SQL lines in `temp/lina-perf-audit/<run-id>/server.log`
- **AND** it does not depend on new middleware or custom headers

#### Scenario: Trace ID is unavailable

- **WHEN** an endpoint response does not contain `Trace-ID`
- **THEN** the sub agent searches by request time window and request URL
- **AND** the audit entry states `trace ID unavailable, evidence quality reduced`
- **AND** the endpoint is not skipped solely because the trace ID is missing

### Requirement: Destructive endpoints are handled with autonomous fixtures

Sub agents MUST avoid damaging shared audit data when calling destructive endpoints such as DELETE, uninstall, clear, reset, or equivalent operations.

#### Scenario: Destructive endpoint has a matching create endpoint

- **WHEN** a sub agent audits a destructive endpoint and the same module has a matching create endpoint
- **THEN** it creates a dedicated audit fixture first
- **AND** it calls the destructive endpoint only against that fixture
- **AND** it attempts cleanup even if the destructive call fails
- **AND** the report states that the autonomous fixture completed without polluting shared data

#### Scenario: Destructive endpoint has no matching create endpoint

- **WHEN** no same-module create endpoint exists for a destructive operation
- **THEN** the sub agent marks the endpoint as `SKIPPED: no matching create endpoint, manual follow-up required`
- **AND** the endpoint appears in the manual follow-up section of `SUMMARY.md`
- **AND** the sub agent does not use another module's resources as a substitute

### Requirement: Severity classification and read-request side-effect detection

The skill MUST classify findings as HIGH, MEDIUM, or LOW and MUST include evidence and remediation guidance. The audit MUST check performance risks and read/query endpoints that execute unexpected write SQL.

#### Scenario: HIGH severity is assigned

- **WHEN** an endpoint has list/detail N+1 with source evidence, a missing index on potentially large data, a non-batch response over 1 second, blocking loop work, or unexpected write SQL in a read/query endpoint trace
- **THEN** the finding is marked `severity: HIGH`
- **AND** `SUMMARY.md` lists it in the HIGH section
- **AND** remediation includes a concrete batching, pagination, indexing, asynchronous, or endpoint-semantics fix

#### Scenario: MEDIUM severity is assigned

- **WHEN** an endpoint shows small-sample N+1, missing pagination, repeated same-data reads, or multiple SELECT calls that should be merged with `JOIN` or `WHERE IN`
- **THEN** the finding is marked `severity: MEDIUM`

#### Scenario: LOW severity is assigned

- **WHEN** an endpoint has slightly high SQL count with fast indexed queries, application-layer filtering that can be pushed down, or static-only evidence not observed at runtime
- **THEN** the finding is marked `severity: LOW`
- **AND** remediation states that the issue is observational or lower-priority

#### Scenario: Read request write side effects are detected

- **WHEN** a sub agent calls a GET, list, query, tree, options, count, detail, current, or equivalent read endpoint
- **THEN** it checks the endpoint's trace for write SQL whose first significant token is `INSERT`, `UPDATE`, `DELETE`, `REPLACE`, `TRUNCATE`, `ALTER`, `DROP`, or `CREATE`
- **AND** unexpected writes are reported as HIGH with anti-pattern signature prefix `read-write-side-effect`
- **AND** Stage 0 setup, login, stress fixture writes, and autonomous fixture create/delete calls are not counted as the audited read endpoint's writes
- **AND** if the trace contains read SQL and every write statement only touches `sys_online_session` or `plugin_monitor_operlog`, the write is recorded as an expected operational PASS note and no finding, summary violation, or issue card is created
- **AND** if the trace writes other business, plugin-state, runtime-state, or storage tables, the finding remains reportable

#### Scenario: Findings are traceable

- **WHEN** a report contains a finding
- **THEN** it includes method and path, module name, trace ID or fallback marker, SQL count, write SQL count when applicable, key SQL excerpts, source file and line, and at least one concrete remediation suggestion

### Requirement: Audit does not modify production delivery assets

The skill MAY temporarily patch local logger output during an audit run, but it MUST restore the original logger settings and MUST NOT modify source code, API DTOs, SQL delivery assets, frontend code, or default runtime configuration as part of the audit itself.

#### Scenario: Logger settings are patched and restored

- **WHEN** Stage 0 needs stable log output for SQL evidence
- **THEN** `setup-audit-env.sh` backs up the original `logger.path` and `logger.file`
- **AND** it sets `logger.path` to the run directory and `logger.file` to `server.log`
- **AND** `restore-audit-env.sh` restores the exact original values on success or failure
- **AND** `SUMMARY.md` records the logger path, logger file, and restore result

#### Scenario: Delivery assets are not changed by the audit

- **WHEN** the skill discovers a performance issue
- **THEN** it writes reports and issue cards only
- **AND** it does not modify `apps/lina-core/api/`, `apps/lina-core/internal/`, `apps/lina-plugins/`, `manifest/sql/`, or frontend source as part of the audit run
- **AND** remediation text states that fixes should be implemented through later OpenSpec changes

### Requirement: Persistent cross-run issue cards

The skill MUST write each discovered performance or read-side-effect issue to a persistent markdown card under repository-root `perf-issues/` and MUST de-duplicate cards by fingerprint across runs.

#### Scenario: New issue creates one card

- **WHEN** Stage 2 observes a new finding
- **THEN** it creates `perf-issues/<severity>-<module>-<slug>.md`
- **AND** the card frontmatter contains `id`, `severity`, `module`, `endpoint`, `status`, `first_seen_run`, `last_seen_run`, `seen_count`, and `fingerprint`
- **AND** the card body contains the sections `问题描述`, `复现方式`, `证据`, `改进方案`, and `历史记录`
- **AND** descriptive card text and `perf-issues/INDEX.md` headings are written in Chinese, while API paths, SQL excerpts, Trace IDs, fingerprints, frontmatter field names, and status values keep their machine-readable originals

#### Scenario: Reproduction steps are self-contained

- **WHEN** an engineer reads only a single issue card
- **THEN** the `复现方式` section gives the commands needed from a clean local environment through the endpoint request and expected SQL observation
- **AND** it does not depend on undeclared temporary variables from the original run

#### Scenario: Existing fingerprint updates the existing card

- **WHEN** a later run observes a finding with the same fingerprint
- **THEN** Stage 2 updates `last_seen_run`, increments `seen_count`, and appends history
- **AND** it does not create a duplicate file
- **AND** if the existing card has `status: fixed` or `status: obsolete`, it changes status back to `open` and records a regression history entry

#### Scenario: Issue cards and run reports cross-reference each other

- **WHEN** Stage 2 completes
- **THEN** the run `SUMMARY.md` links to all cards created or updated in that run
- **AND** each card history entry references the run ID and the relevant `audits/<module-or-shard>.md` path
- **AND** `perf-issues/INDEX.md` lists all `open` and `in-progress` cards ordered by severity

#### Scenario: Issue cards are not OpenSpec archive artifacts

- **WHEN** the OpenSpec change is archived
- **THEN** `perf-issues/` remains at repository root
- **AND** OpenSpec archive does not move issue cards into `openspec/changes/archive/` or `openspec/specs/`

### Requirement: Stress fixtures are audit-only

Stress fixture data MUST exist only during audit runs and MUST NOT be written to host or plugin delivery SQL directories.

#### Scenario: Stress fixture script generates temporary scale data

- **WHEN** Stage 0 needs additional rows to make N+1 behavior observable
- **THEN** the skill runs `.agents/skills/lina-perf-audit/scripts/stress-fixture.sh`
- **AND** the script runs after host mock data and plugin mock data are ready
- **AND** it inserts data in dependency order using idempotent patterns such as `INSERT IGNORE` or prior existence checks
- **AND** it does not write files under `apps/lina-core/manifest/sql/`, `apps/lina-core/manifest/sql/mock-data/`, or plugin delivery SQL directories
