## Context

This consolidation merges archived changes that address developer tooling and operational infrastructure for LinaPro:

1. **Upgrade governance** -- establishing a unified development-time upgrade entry point for both framework and source-plugin upgrades, with effective-version separation and startup fail-fast validation.
2. **Development database configuration deduplication** -- removing duplicated connection settings and `multiStatements` dependency from the host development toolchain.
3. **Framework bootstrap installer** -- providing cross-platform installation scripts for quick source code deployment with environment health checks.
4. **Performance audit skill** -- turning existing runtime observability into a reusable AI governance workflow for repeatable backend API performance and read-side-effect auditing.

These changes share a common theme: improving the developer experience around tooling, configuration, and operational workflows. The merged design organizes them by functional area.

## Goals / Non-Goals

**Goals:**
- Provide a single `make upgrade` entry point that supports both framework and source-plugin upgrades with explicit scope parameters.
- Separate effective source-plugin versions from discovered source versions to prevent governance state corruption.
- Block host startup when installed source plugins have pending upgrades.
- Remove duplicated database connection settings from the host development config file.
- Remove `multiStatements=true` dependency from local SQL bootstrap execution.
- Establish cross-platform installation scripts as a first-landing entry point for new developers.
- Project the built-in log cleanup task through startup code rather than SQL seed data.
- Provide a manual-trigger-only skill that runs a full backend API audit for the host and all built-in plugins.
- Detect common API performance anti-patterns including N+1, missing indexes, missing pagination, repeated same-data reads, mergeable SQL calls, blocking loop work, and read/query endpoints that execute unexpected write SQL.
- Produce stable per-run artifacts and persistent cross-run issue cards for the audit workflow.

**Non-Goals:**
- Do not implement rollback commands, automatic rollback, or rollback SQL directories.
- Do not build a runtime business-system upgrade platform in this iteration.
- Do not change runtime config structures or the delivery strategy under `manifest/config/`.
- Do not automatically install system dependencies during the bootstrap installation flow.
- Do not introduce a `git clone`-based primary installation path.
- Do not automatically fix issues discovered by the audit. Fixes are handled by later OpenSpec changes.
- Do not implement real-time monitoring or APM.
- Do not replace pprof, database `EXPLAIN`, or specialized performance tooling.
- Do not audit frontend performance.
- Do not archive individual audit run outputs. Only the skill capability and helper scripts are archived.

## Upgrade Governance

### Decision 1: Keep `make upgrade` as the only development-time upgrade entry point

`make upgrade` remains the single top-level development-time upgrade command, but it now accepts explicit scope parameters (`scope=framework` or `scope=source-plugin`). The repository-root tool lives under `hack/upgrade-source`, with `main.go` kept at the root and the real implementation split into focused internal components for framework and source-plugin upgrades.

### Decision 2: Source-plugin upgrades must be explicit and pre-startup

Source plugins are compiled into the host, so they must be upgraded before the host starts. Startup may scan and validate, but it must not automatically execute source-plugin migrations. The shared development-time upgrade command supports both single-plugin (`plugin=<id>`) and bulk (`plugin=all`) upgrades.

### Decision 3: Source plugins must separate effective version from discovered version

`sys_plugin.version` and `sys_plugin.release_id` represent only the effective source-plugin version. Higher versions discovered in source are written as prepared releases and do not take effect until an explicit upgrade completes. This prevents the source scan from overwriting governance state.

### Decision 4: Reuse release, migration, and resource-reference ledgers

Source-plugin upgrades reuse `sys_plugin_release`, `sys_plugin_migration`, and `sys_plugin_resource_ref` rather than introducing a separate upgrade metadata stack. The upgrade records entries with `phase=upgrade` and synchronizes menus, permissions, and governance resource references.

### Decision 5: Reuse existing install SQL assets for upgrade execution

The iteration does not introduce `manifest/sql/upgrade/`. Source-plugin upgrades reuse the existing plugin SQL assets and record the execution under `phase=upgrade`, relying on idempotent SQL rules already required by the project.

### Decision 6: Host startup must fail fast when a source-plugin upgrade is pending

After source scanning, startup compares the effective version with the highest discovered source version. If an installed source plugin is behind, startup fails before routes, cron jobs, or other plugin runtime hooks become active. The error message includes the plugin ID, effective version, discovered version, and the recommended `make upgrade` command.

### Decision 7: Dynamic-plugin upgrades keep their runtime model

Dynamic plugins continue to upgrade through upload plus install/reconcile. `make upgrade` must never scan, switch, or migrate dynamic-plugin releases. This boundary reflects different delivery models rather than a design inconsistency.

### Decision 8: Framework upgrades must read upgrade metadata only from hack config

The framework upgrade path reads the current upgrade baseline from `frameworkUpgrade.version` in `apps/lina-core/hack/config.yaml` and compares it with the target upgrade version found in the target source's `hack/config.yaml`. The default upstream repository URL comes from `frameworkUpgrade.repositoryUrl` in the same file. The upgrade implementation does not read host runtime configuration.

### Decision 9: Framework upgrades must replay all host SQL from the first file

After the target source code is applied, the framework upgrade replays every host SQL file from the first file in sorted order. The process does not rely on SQL cursors or extra upgrade metadata tables. Execution stops immediately on the first SQL failure.

### Decision 10: Rollback stays out of scope

Failures stop execution immediately, preserve failure records and logs, and require manual repair. The iteration does not attempt automated recovery.

## Development Database Configuration

### Decision 11: Reuse one development database connection in host `hack/config.yaml` through YAML anchors

Define one shared database connection anchor in the file and let both `database.default.link` and `gfcli.gen.dao[].link` reference it. Remove `multiStatements=true` from the shared DSN so local SQL execution and `gf gen dao` use the same base connection settings. This fixes single-file configuration drift directly without adding extra scripts or a template rendering step.

### Decision 12: Split SQL statements explicitly in the command layer

Add SQL splitting helpers under `apps/lina-core/internal/cmd/` and turn each SQL file into an ordered statement list that is executed one statement at a time. The splitter ignores blank fragments and correctly handles common comments and semicolons inside string literals. `executeSQLAssetsWithExecutor` keeps its fail-fast behavior, but the execution granularity changes from "run the whole file at once" to "run statements inside the file in order and stop the full bootstrap as soon as any statement fails."

### Decision 13: Cover the behavior change with command-layer unit tests

Extend existing command-layer tests with statement splitting and sequential execution tests. The new tests focus on four cases: multi-statement files executing in order, blank/comment fragments being skipped, semicolons inside string literals not causing a split, and immediate termination after a statement failure.

## Framework Bootstrap Installer

### Decision 14: Use dual entry point scripts

Add two entry points under `hack/scripts/install/`: `install.sh` for macOS/Linux and `install.ps1` for Windows PowerShell. Both share consistent core parameter semantics. `curl | bash` works well for Unix-like environments but not for native Windows users; PowerShell requires native download and extraction capabilities. Splitting entry points makes script logic clearer.

### Decision 15: Source code acquisition uniformly uses archive download

The installation script defaults to downloading archive files from GitHub/Codeload based on the platform, with Unix preferring `tar.gz` and Windows preferring `zip`. Archive download is lighter than `git clone`, does not require repository history, and does not depend on Git being pre-installed. Through repository and `ref` parameters, it supports the official default repository, specified branches, specified tags, and development version distribution.

### Decision 16: Target directory uses explicit mode selection with safe defaults

By default, source code is deployed to a new subdirectory under the current working directory. When the user explicitly passes a current-directory-mode parameter, extraction results are placed directly in the current directory. When the user passes a target path, the script deploys to the specified directory. In all modes, if the target directory is non-empty, the script defaults to refusing to continue unless the user explicitly passes an overwrite parameter. The script extracts to a temporary directory first, identifies the unique root directory dynamically, then moves it to the final position.

### Decision 17: Environment health check only, no automatic dependency installation

After source code deployment, the script checks for the presence of key dependencies (Go, Node.js, pnpm, MySQL, make) and outputs version information or missing dependency prompts. It does not automatically call package managers. Doing a good job on dependency checking and next-step command guidance already significantly reduces onboarding costs while keeping the script predictable and auditable.

### Decision 18: Repository scripts as single source of truth

`hack/scripts/install/install.sh` and `hack/scripts/install/install.ps1` serve as the sole source of installation logic. The publicly available `https://linapro.ai/install.sh` / `install.ps1` should reuse the same content, at most doing a thin wrapper or redirect mapping, and must not diverge from the repository script long-term.

## Cron Job Management

### Decision 19: Project built-in cleanup task through startup code

The built-in `host:cleanup-job-logs` task is registered through host source code and projected into `sys_job` during service startup. Delivery SQL does not write initialization seed data for this task. The default `cron_expr` triggers daily at midnight and the task has `is_builtin=1`.

## Performance Audit Skill

### Decision 20: Implement the audit as an agent skill

The workflow is implemented at `.agents/skills/lina-perf-audit/`. The skill owns `SKILL.md`, references, and bundled scripts. This matches the repository's AI-governance workflow and keeps orchestration instructions close to deterministic helper scripts. A single repository-level shell script was rejected because it cannot express agent sharding, review responsibilities, evidence aggregation, and issue-card lifecycle rules.

### Decision 21: Keep endpoint audit execution in sub agents

The main agent performs Stage 0 setup, endpoint catalog generation, shard planning, status tracking, and Stage 2 aggregation. Each concrete API endpoint audit is executed by a sub agent. Sub-agent prompts include only the assigned module or endpoint shard, fixture data, token, log path, and run directory. If a module prompt would exceed the size budget, the module is split into smaller shards. Destructive endpoints stay with the create path needed to create and clean an autonomous fixture.

### Decision 22: Cover all built-in plugins

Stage 0 scans every `apps/lina-plugins/*/plugin.yaml`, syncs each plugin through host plugin APIs, installs and enables it, and loads plugin mock data when a plugin provides `manifest/sql/mock-data/`. The endpoint catalog is built from both host DTOs and plugin DTOs. Plugins with no backend API are recorded as skipped with reason `no backend API`. A plugin that declares API DTOs but has unreachable runtime routes causes environment preparation to fail instead of silently reducing coverage.

### Decision 23: Correlate requests through the default `Trace-ID` header

Sub agents call endpoints with `curl -i` or equivalent tooling, read the `Trace-ID` response header, and grep `temp/lina-perf-audit/<run-id>/server.log` for matching SQL lines. No audit-specific middleware, custom response header, build tag, or runtime feature flag is introduced. If a trace ID is unavailable, the sub agent falls back to request time window plus URL matching and marks evidence quality as reduced.

### Decision 24: Make destructive endpoint handling autonomous

For DELETE, reset, clear, uninstall, or equivalent destructive operations, the sub agent first creates a dedicated audit fixture in the same module when a matching create endpoint exists. It then calls the destructive operation against that resource and attempts cleanup even on failure. If no matching create endpoint exists, the sub agent marks the endpoint as skipped and records it in the manual follow-up list. It must not uninstall built-in plugins, clear global logs, or use resources from another module as substitutes.

### Decision 25: Add audit-only stress fixtures

`stress-fixture.sh` inserts additional data after host mock data and plugin mock data are ready. The goal is to make N+1 behavior visible by increasing listable row counts to tens or hundreds. Stress data is generated only for audit runs. It is not written to `apps/lina-core/manifest/sql/`, `apps/lina-core/manifest/sql/mock-data/`, or plugin delivery SQL directories, and it disappears on the next database reset. Inserts use idempotent patterns such as `INSERT IGNORE` or existence checks.

### Decision 26: Produce two report layers

Each audit run writes a disposable snapshot under `temp/lina-perf-audit/<run-id>/` containing `catalog.json`, `fixtures.json`, `server.log`, `audits/<module-or-shard>.md`, `SUMMARY.md`, and `meta.json`. Each finding also updates a persistent issue card under repository-root `perf-issues/`. Cards are de-duplicated by `sha256(module + ":" + method + ":" + path + ":" + severity + ":" + anti-pattern signature)`. A repeated finding updates `last_seen_run`, `seen_count`, and history instead of creating another card. If a previously fixed or obsolete card is observed again, it is reopened and marked as a regression. `perf-issues/` is intentionally outside `temp/` and outside OpenSpec archive directories. It is a cross-run backlog consumed by later fix iterations.

### Decision 27: Use three severity levels

`HIGH` covers clearly harmful patterns: observable N+1 on list/detail endpoints, missing indexes on potentially large data, non-batch endpoints over 1s, blocking loop work, and read/query endpoints whose trace executes unexpected write SQL. `MEDIUM` covers near-term risks: small-sample N+1, missing pagination, repeated same-data reads, or SELECT calls that should be merged by `JOIN` or `WHERE IN`. `LOW` covers weaker or static-only evidence: slightly high SQL count with fast indexed queries, filtering that can be pushed into SQL, or a write-risk branch not observed at runtime.

### Decision 28: Treat operational read-side-effect writes as expected only in a narrow case

Read/query endpoints are reportable when their own trace writes business, plugin-state, runtime-state, or storage tables. Writes are treated as expected operational side effects only when the trace also contains read SQL and every write statement touches only `sys_online_session` or `plugin_monitor_operlog`. This prevents session heartbeat and operation-log persistence from creating noisy cards while still catching real read-endpoint mutations.

## Risks / Trade-offs

- Historical source-plugin releases still reference the same evolving source tree rather than frozen artifacts. The iteration accepts this limitation to deliver a clear upgrade path first.
- Startup fail-fast adds friction for local development, but it is safer than running a host whose compiled source plugins are newer than its governance state.
- Without rollback, recovery is more manual. That is an intentional boundary for this iteration.
- Source and dynamic plugins keep different upgrade triggers reflecting different delivery models.
- A custom SQL splitter may miss edge cases. Mitigation: targeted tests against current delivered SQL style, prioritizing common boundaries such as strings, line comments, and block comments.
- When dependencies are not automatically installed, some users still need to manually supplement their environment. Mitigation: clear missing-item output and next-step commands.
- Current directory mode can easily damage existing files. Mitigation: explicitly triggered parameter with non-empty directory refusal.
- Destructive local setup from the audit skill: the skill is manual-trigger-only and requires explicit confirmation for ambiguous requests.
- Service or config pollution from audit runs: setup backs up logger settings and restore is required on success and failure.
- Plugin coverage gaps in audit: Stage 0 fails when declared plugin routes are unreachable.
- Report duplication in audit: persistent cards are fingerprinted and updated instead of duplicated.
- Prompt growth in audit: endpoint work is split across sub agents with a small prompt budget.
