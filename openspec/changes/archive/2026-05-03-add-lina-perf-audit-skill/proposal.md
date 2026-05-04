## Why

LinaPro already has a growing set of backend APIs across `lina-core` and built-in plugins. During fast SDD-driven iteration, common API performance risks were still mostly found by ad hoc developer review: N+1 queries, DAO calls inside loops, missing-index full scans, unbounded list responses, blocking work inside loops, repeated configuration reads, cache misses, and GET/query endpoints that execute write SQL.

The runtime already exposes enough evidence to make this review repeatable. GoFrame writes the request trace ID to the `Trace-ID` response header, and `database.debug=true` emits SQL logs that include the trace ID. This change turns that existing observability into a reusable AI governance workflow so each audit does not need to be rebuilt manually.

## What Changes

- Add the `lina-perf-audit` agent skill. It orchestrates environment preparation, built-in plugin installation and enablement, endpoint sharding, sub-agent execution, trace-ID-based SQL lookup, source review, report aggregation, and read-request write-side-effect detection.
- Keep endpoint review work in sub agents. The main agent handles setup, task tracking, and aggregation; every API endpoint audit task is assigned to a sub agent. Large modules are split into endpoint or small-module shards so each prompt stays small.
- Require autonomous handling for destructive endpoints. A sub agent creates a dedicated resource, calls the destructive operation against that resource, and cleans it up. Destructive endpoints without a matching create path are marked for manual follow-up instead of damaging shared data.
- Read the GoFrame default `Trace-ID` response header directly. The skill does not add middleware, modify GoFrame behavior, or change production request handling.
- Add bundled helper scripts under `.agents/skills/lina-perf-audit/scripts/`:
  - `setup-audit-env.sh`: stop services, back up `logger.path` and `logger.file`, patch audit logging to a stable run directory and `server.log`, start the backend, wait for readiness, and write the `admin` token.
  - `restore-audit-env.sh`: restore logger configuration from the run directory and stop services on both success and failure paths.
  - `prepare-builtin-plugins.sh`: discover `apps/lina-plugins/*/plugin.yaml`, sync, install, enable every built-in plugin through host APIs, and load plugin mock data when present.
  - `scan-endpoints.sh`: scan host and plugin API DTOs and generate a module-grouped endpoint catalog from route metadata.
  - `probe-fixtures.sh`: call module list endpoints, collect representative resource IDs, and fail fast when declared routes are unreachable.
  - `stress-fixture.sh`: add audit-only stress data after host mock data and plugin mock data are ready so N+1 query counts become observable.
  - `aggregate-reports.sh`: merge sub-agent audit files, generate the run summary, update persistent issue cards, and regenerate the issue index.
- Define two report layers:
  - Per-run reports under `temp/lina-perf-audit/<run-id>/`, including `catalog.json`, `fixtures.json`, `server.log`, `audits/*.md`, `SUMMARY.md`, and `meta.json`.
  - Persistent issue cards under repository-root `perf-issues/`, one markdown file per performance or read-side-effect issue, de-duplicated by fingerprint and tracked with `open`, `in-progress`, `fixed`, or `obsolete` status.
- Enforce manual-trigger-only behavior. The skill description and spec state that it must not be invoked by other skills, CI, scheduled jobs, git hooks, or ambiguous performance requests. Ambiguous requests require explicit user confirmation before any destructive setup command runs.
- This change delivers the audit capability itself. It does not make the skill a production runtime dependency and does not archive individual audit run output.

## Capabilities

### New Capabilities

- `lina-perf-audit-skill`: defines the public contract for LinaPro's backend API performance and read-request side-effect audit skill, including manual triggering, sub-agent orchestration, destructive endpoint handling, trace-ID correlation, report artifacts, persistent issue cards, severity classification, and automation restrictions.

### Modified Capabilities

- None. This change adds a new governance capability and does not modify an existing capability contract.

## Impact

- Adds `.agents/skills/lina-perf-audit/SKILL.md`, reference templates, and bundled scripts for Claude Code, Codex, and other AI coding tools that can read project skills.
- Adds the baseline spec `openspec/specs/lina-perf-audit-skill/spec.md` during archive.
- Runtime impact only exists when the user manually runs the skill. A run resets the local database, reloads mock data, installs and enables built-in plugins, adds stress fixtures, restarts services, writes temporary run artifacts, and updates persistent `perf-issues/` cards.
- The skill does not change production API behavior, add middleware, modify default configuration values, or write audit stress data into delivery SQL assets.
- i18n impact: no runtime i18n resources, frontend language packs, apidoc translation resources, menus, buttons, or DTO documentation are added or changed. Skill documentation and audit reports are developer-facing assets outside the runtime localization system.
- Cache impact: the skill itself does not introduce production caches. Feedback work completed during this iteration also recorded cache consistency impact and verification in `tasks.md`.
- Dependencies: the workflow reuses existing `make init confirm=init rebuild=true`, `make mock confirm=mock`, `make stop`, GoFrame `Trace-ID`, and SQL debug logging behavior.
