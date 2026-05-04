## Context

LinaPro has backend APIs in `apps/lina-core/api` and in built-in source plugins under `apps/lina-plugins/*/backend/api`. The number of APIs keeps increasing, while the default mock data set is intentionally small. That makes N+1 behavior, repeated metadata reads, missing pagination, read endpoints with write side effects, and other SQL-pattern regressions hard to observe during ordinary manual testing.

The existing runtime already provides the key evidence path:

- `apps/lina-core/manifest/config/config.yaml` enables trace ID support in logs.
- `database.default.debug=true` writes SQL statements into the backend log.
- GoFrame v2 writes the current request trace ID to the `Trace-ID` response header.
- The project logger already preserves trace IDs, so no new middleware is required.

The workflow must be repeatable, but it is destructive and expensive: it resets local data, restarts services, installs all built-in plugins, adds audit-only stress fixtures, and fans out to multiple sub agents. For that reason the capability is implemented as a manual agent skill, not as CI automation.

## Goals / Non-Goals

Goals:

- Provide a manual-trigger-only skill that runs a full backend API audit for the host and all built-in plugins.
- Detect common API performance anti-patterns including N+1, missing indexes, missing pagination, repeated same-data reads, mergeable SQL calls, blocking loop work, and read/query endpoints that execute unexpected write SQL.
- Use sub agents for endpoint execution and evidence collection so the main agent does not carry all API context in one prompt.
- Produce stable per-run artifacts and persistent cross-run issue cards.
- Keep deterministic setup, scanning, fixture, and aggregation logic inside the skill's own `scripts/` directory.
- Avoid production runtime changes. The skill only patches local logger output during an audit run and restores it afterward.

Non-goals:

- Do not automatically fix issues discovered by the audit. Fixes are handled by later OpenSpec changes.
- Do not implement real-time monitoring or APM.
- Do not replace pprof, database `EXPLAIN`, or specialized performance tooling.
- Do not audit frontend performance.
- Do not archive individual audit run outputs. Only the skill capability and helper scripts are archived.

## Decisions

### Decision 1: Implement the audit as an agent skill

The workflow is implemented at `.agents/skills/lina-perf-audit/`. The skill owns `SKILL.md`, references, and bundled scripts.

This matches the repository's AI-governance workflow and keeps orchestration instructions close to deterministic helper scripts. A single repository-level shell script was rejected because it cannot express agent sharding, review responsibilities, evidence aggregation, and issue-card lifecycle rules.

### Decision 2: Keep endpoint audit execution in sub agents

The main agent performs Stage 0 setup, endpoint catalog generation, shard planning, status tracking, and Stage 2 aggregation. Each concrete API endpoint audit is executed by a sub agent.

Sub-agent prompts include only the assigned module or endpoint shard, fixture data, token, log path, and run directory. If a module prompt would exceed the size budget, the module is split into smaller shards. Destructive endpoints stay with the create path needed to create and clean an autonomous fixture.

### Decision 2.1: Cover all built-in plugins

Stage 0 scans every `apps/lina-plugins/*/plugin.yaml`, syncs each plugin through host plugin APIs, installs and enables it, and loads plugin mock data when a plugin provides `manifest/sql/mock-data/`.

The endpoint catalog is built from both host DTOs and plugin DTOs. Plugins with no backend API are recorded as skipped with reason `no backend API`. A plugin that declares API DTOs but has unreachable runtime routes causes environment preparation to fail instead of silently reducing coverage.

### Decision 3: Correlate requests through the default `Trace-ID` header

Sub agents call endpoints with `curl -i` or equivalent tooling, read the `Trace-ID` response header, and grep `temp/lina-perf-audit/<run-id>/server.log` for matching SQL lines.

No audit-specific middleware, custom response header, build tag, or runtime feature flag is introduced. If a trace ID is unavailable, the sub agent falls back to request time window plus URL matching and marks evidence quality as reduced.

### Decision 4: Make destructive endpoint handling autonomous

For DELETE, reset, clear, uninstall, or equivalent destructive operations, the sub agent first creates a dedicated audit fixture in the same module when a matching create endpoint exists. It then calls the destructive operation against that resource and attempts cleanup even on failure.

If no matching create endpoint exists, the sub agent marks the endpoint as skipped and records it in the manual follow-up list. It must not uninstall built-in plugins, clear global logs, or use resources from another module as substitutes.

### Decision 5: Add audit-only stress fixtures

`stress-fixture.sh` inserts additional data after host mock data and plugin mock data are ready. The goal is to make N+1 behavior visible by increasing listable row counts to tens or hundreds.

Stress data is generated only for audit runs. It is not written to `apps/lina-core/manifest/sql/`, `apps/lina-core/manifest/sql/mock-data/`, or plugin delivery SQL directories, and it disappears on the next database reset. Inserts use idempotent patterns such as `INSERT IGNORE` or existence checks.

### Decision 6: Produce two report layers

Each audit run writes a disposable snapshot under:

```text
temp/lina-perf-audit/<run-id>/
  catalog.json
  fixtures.json
  server.log
  audits/<module-or-shard>.md
  SUMMARY.md
  meta.json
```

Each finding also updates a persistent issue card under repository-root `perf-issues/`. Cards are de-duplicated by `sha256(module + ":" + method + ":" + path + ":" + severity + ":" + anti-pattern signature)`. A repeated finding updates `last_seen_run`, `seen_count`, and history instead of creating another card. If a previously fixed or obsolete card is observed again, it is reopened and marked as a regression.

`perf-issues/` is intentionally outside `temp/` and outside OpenSpec archive directories. It is a cross-run backlog consumed by later fix iterations.

### Decision 7: Use three severity levels

`HIGH` covers clearly harmful patterns: observable N+1 on list/detail endpoints, missing indexes on potentially large data, non-batch endpoints over 1s, blocking loop work, and read/query endpoints whose trace executes unexpected write SQL.

`MEDIUM` covers near-term risks: small-sample N+1, missing pagination, repeated same-data reads, or SELECT calls that should be merged by `JOIN` or `WHERE IN`.

`LOW` covers weaker or static-only evidence: slightly high SQL count with fast indexed queries, filtering that can be pushed into SQL, or a write-risk branch not observed at runtime.

### Decision 8: Treat operational read-side-effect writes as expected only in a narrow case

Read/query endpoints are reportable when their own trace writes business, plugin-state, runtime-state, or storage tables. Writes are treated as expected operational side effects only when the trace also contains read SQL and every write statement touches only `sys_online_session` or `plugin_monitor_operlog`.

This prevents session heartbeat and operation-log persistence from creating noisy cards while still catching real read-endpoint mutations.

## i18n and Cache Assessment

The skill capability itself has no runtime i18n impact. It adds developer-facing markdown, shell scripts, and audit report templates. It does not add frontend text, menus, route labels, API DTO documentation, apidoc translation resources, or runtime language bundles.

The skill itself does not introduce production cache behavior. Feedback work completed during this iteration did touch runtime plugin cache consistency and recorded the source of truth, consistency model, invalidation triggers, cross-node revision mechanism, tolerated staleness, and fallback path in `tasks.md`.

## Risks and Mitigations

- Destructive local setup: the skill is manual-trigger-only and requires explicit confirmation for ambiguous requests.
- Service or config pollution: setup backs up logger settings and restore is required on success and failure.
- Plugin coverage gaps: Stage 0 fails when declared plugin routes are unreachable.
- Report duplication: persistent cards are fingerprinted and updated instead of duplicated.
- Prompt growth: endpoint work is split across sub agents with a small prompt budget.
