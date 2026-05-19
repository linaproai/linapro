## ADDED Requirements

### Requirement: Default startup must not output SQL detail logs

The system SHALL treat SQL detail logs as an explicit diagnostic capability, not default output for normal startup. The `database.default.debug` default value in deliverable configuration MUST be `false`. When this configuration is `false`, service startup logs MUST NOT output per-statement ORM SQL detail; when the caller explicitly sets it to `true`, the system MAY output SQL detail per GoFrame behavior.

#### Scenario: Default configuration startup does not output SQL detail

- **WHEN** the backend service starts using deliverable default configuration
- **THEN** `database.default.debug` is `false`
- **AND** startup logs MUST NOT contain ORM SQL detail lines such as `SHOW FULL COLUMNS`, `SELECT ... FROM`, or `INSERT INTO`

#### Scenario: Explicit SQL debug enablement

- **WHEN** an administrator sets `database.default.debug` to `true`
- **THEN** the backend service is allowed to output ORM SQL detail
- **AND** this behavior serves only as a diagnostic mode, not as default development or production configuration

### Requirement: Startup orchestration must reuse plugin governance snapshots

The system SHALL reuse the same set of plugin governance startup snapshots within one HTTP startup orchestration. `BootstrapAutoEnable`, source-plugin HTTP route registration, runtime frontend prewarm, plugin read-only projections, and cron builtin job sync MUST NOT repeatedly construct equivalent full-table snapshots of `sys_plugin`, `sys_plugin_release`, `sys_menu`, and `sys_plugin_resource_ref`.

#### Scenario: Only one plugin catalog snapshot per startup

- **WHEN** the backend executes one HTTP startup orchestration
- **THEN** the plugin catalog startup snapshot is constructed at most once
- **AND** subsequent plugin startup phases reuse this snapshot to read `sys_plugin` and `sys_plugin_release`

#### Scenario: Only one plugin integration snapshot per startup

- **WHEN** the backend executes one HTTP startup orchestration
- **THEN** the plugin integration startup snapshot is constructed at most once
- **AND** subsequent menu and resource reference synchronization reuse this snapshot to read `sys_menu` and `sys_plugin_resource_ref`

#### Scenario: Startup writes synchronously update snapshots

- **WHEN** startup synchronization phases create or update plugin registry, release, menu, or resource reference projections
- **THEN** the system MUST synchronously update the current startup snapshot
- **AND** subsequent startup phases MUST NOT re-scan the full table to read projections that were just written

### Requirement: Plugin manifest no-op synchronization must produce no database side effects

The system SHALL implement plugin manifest synchronization as difference-driven. For plugins where registry, release snapshot, manifest menu, dynamic route permission menu, and resource ref all match, the synchronization process MUST NOT open transactions, write to the database, or perform post-write reads.

#### Scenario: No-op source-plugin synchronization does not write to database

- **WHEN** a source-plugin manifest is fully consistent with the database registry, release, menus, permissions, and resource references
- **THEN** startup synchronization for that plugin MUST NOT execute `INSERT`, `UPDATE`, or `DELETE`
- **AND** it MUST NOT produce empty `BEGIN` / `COMMIT` transactions

#### Scenario: Menu transactions only open when menu declarations change

- **WHEN** a plugin manifest's menus or dynamic route permissions change
- **THEN** the system MAY open a menu synchronization transaction and write necessary changes
- **AND** after transaction commit, the system MUST update the menu projection in the startup snapshot

#### Scenario: Release metadata only synced when release snapshot changes

- **WHEN** a plugin manifest's generated release snapshot is consistent with the current `sys_plugin_release` row
- **THEN** the system MUST NOT update release metadata
- **AND** it MUST NOT query the same release row again to refresh the release snapshot

### Requirement: Builtin job startup registration must avoid repeated persistence scans

The system SHALL treat source-code declarations as the authoritative source for builtin scheduled job execution definitions. After startup synchronizes builtin jobs, scheduler registration MUST use the declaration-derived `sys_job` projection snapshot; the persistent scheduler startup scan MUST NOT reload the same batch of `is_builtin=1` jobs as execution definitions.

#### Scenario: Builtin jobs registered from declaration-derived snapshots

- **WHEN** the backend starts and synchronizes builtin job projections
- **THEN** the system uses the synchronization-returned projection snapshot to register builtin jobs
- **AND** the registration process does not need to re-read the same builtin job rows by ID

#### Scenario: Persistent scan skips builtin jobs

- **WHEN** the persistent scheduler executes startup loading
- **THEN** the query conditions MUST exclude `is_builtin=1` jobs
- **AND** only user-created enabled jobs or non-builtin plugin jobs are loaded

### Requirement: Startup must output phase summary instead of relying on SQL detail

The system SHALL output a startup phase summary log after startup completes. The summary MUST include at least plugin scan count, plugin synchronization change count, no-op plugin count, startup snapshot construction count, builtin job projection count, and startup phase durations. The summary MUST NOT include full SQL text.

#### Scenario: Startup summary includes plugin sync statistics

- **WHEN** the backend completes plugin startup synchronization
- **THEN** the log outputs plugin scan count, changed plugin count, and no-op plugin count
- **AND** the log MUST NOT include complete SQL statement text

#### Scenario: Startup summary includes snapshot construction counts

- **WHEN** the backend completes HTTP startup orchestration
- **THEN** the log outputs catalog, integration, and job startup snapshot construction counts
- **AND** duplicate snapshot constructions within the same startup phase will be identified as regression by tests or review

### Requirement: Startup SQL efficiency must have automated regression coverage

The system SHALL provide automated tests or smoke scripts covering startup SQL efficiency key boundaries. Tests MUST NOT depend on GoFrame metadata probe exact SQL counts, but MUST constrain project-controllable behavior including default SQL detail suppression, plugin no-op synchronization with no writes, no empty transactions, and shared startup snapshot non-duplication.

#### Scenario: Default startup log smoke test

- **WHEN** a test starts the backend service using default database debug configuration
- **THEN** the test asserts startup logs do not contain ORM SQL detail
- **AND** the test asserts startup summary log exists

#### Scenario: Plugin no-op synchronization regression test

- **WHEN** a test prepares a plugin that is already synchronized with no manifest differences
- **THEN** re-executing startup synchronization MUST NOT produce write SQL
- **AND** it MUST NOT produce empty transactions

#### Scenario: Startup snapshot reuse regression test

- **WHEN** a test executes one HTTP startup orchestration or equivalent startup orchestration unit
- **THEN** catalog, integration, and job startup snapshot construction counts MUST each remain within budget
- **AND** the test fails when the budget is exceeded
