---
name: lina-perf-audit
description: "MANUAL TRIGGER ONLY. Run the LinaPro backend API performance audit only after the user explicitly asks for lina-perf-audit or confirms a full audit. This workflow performs reset DB via make init/mock, stops and restarts services, installs and enables all built-in plugins, adds stress fixtures, launches concurrent sub agents, has high elapsed time (tens of minutes to hours), and carries significant token cost. Never invoke from other skills, CI, scheduled jobs, git hooks, or ambiguous performance requests."
---

# LinaPro Performance Audit

Use this skill to run a full LinaPro backend API audit for `apps/lina-core` and all built-in plugins under `apps/lina-plugins`. The audit finds performance risks such as N+1 queries, missing indexes, unbounded list responses, repeated reads, mergeable SQL calls, blocking operations in loops, and read/query endpoints that execute write SQL such as INSERT, UPDATE, DELETE, REPLACE, TRUNCATE, or ALTER. It produces reports and issue cards only; it does not fix production code.

This skill is plain markdown so Claude Code, Codex, and other AI coding tools can read it.

## Manual Trigger Gate

This skill is destructive to the local development environment and expensive to run. It resets the database, reloads mock data, restarts services, installs and enables all built-in plugins, writes temporary audit logs, runs multiple sub agents, and may consume a large token budget.

Proceed only when the user explicitly requests a full audit, for example:

- `run lina-perf-audit`
- `execute the LinaPro performance audit`
- `perform a full backend API performance audit`
- `check all backend APIs for N+1 with the perf audit skill`

For ambiguous requests such as `the API seems slow`, `how is performance`, or `check interface performance`, ask for confirmation first. The confirmation message must mention database reset, service restart, elapsed time, sub agent fan-out, and token cost. Before confirmation, do not run `make stop`, `make init`, `make mock`, `setup-audit-env.sh`, `prepare-builtin-plugins.sh`, or `stress-fixture.sh`.

## When to Use

- The user explicitly names `lina-perf-audit` or asks to run the full LinaPro API performance audit.
- The user explicitly asks for a systematic backend API audit across `lina-core` and built-in plugins.
- The user explicitly asks to generate run reports under `temp/lina-perf-audit/<run-id>/` and persistent issue cards under `perf-issues/`.
- The user confirms an ambiguous performance request after being told the cost and destructive local effects.

## When NOT to Use

- Do not trigger from another skill such as `lina-review`, `lina-feedback`, or `lina-e2e`. Those skills may suggest that the user manually run `lina-perf-audit`, but must not invoke it.
- Do not run from CI, scheduled jobs, git hooks, automation, or background workflows.
- Do not run for a single endpoint investigation unless the user explicitly wants the full audit. Use a focused manual analysis instead.
- Do not run before user confirmation for ambiguous performance concerns.
- Do not run if the user cannot tolerate database reset, mock data reload, service restart, or the token/time cost.

## Reference Files

Load these files only when needed:

- `references/sub-agent-prompt.md`: prompt template for Stage 1 sub agents, including endpoint or small-module sharding.
- `references/severity-rubric.md`: HIGH / MEDIUM / LOW classification and anti-pattern signatures.
- `references/report-template.md`: `audits/<module>.md`, `SUMMARY.md`, and `meta.json` templates.
- `references/issue-card-template.md`: persistent `perf-issues/*.md` issue card template.
- `references/fingerprint-rule.md`: exact fingerprint generation, de-duplication, update, and cross-run rules.

## Bundled Scripts

Deterministic helper scripts are maintained inside this skill under `scripts/`. Run them from the repository root with paths like `bash .agents/skills/lina-perf-audit/scripts/scan-endpoints.sh ...`. Do not copy these scripts into `hack/` or maintain a second script set outside the skill; the skill directory is the ownership boundary for the audit workflow.

## Workflow

Run the audit in exactly three stages.

### Stage 0: Preparation

1. Confirm the user explicitly approved the full audit.
2. Create a unique `run_id` in `YYYYMMDD-HHMMSS` format and set `run_dir=temp/lina-perf-audit/<run_id>`.
3. Run `make stop`.
4. Reset local data with `make init confirm=init rebuild=true` and `make mock confirm=mock`.
5. Run `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id <run_id>` to patch audit logging, start the backend, wait for readiness, and obtain the `admin` token.
6. Run `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir <run_dir>` to discover, sync, install, enable, and load mock data for all built-in plugins.
7. Run `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir <run_dir>` after host mock data and plugin mock data are ready.
8. Run `bash .agents/skills/lina-perf-audit/scripts/scan-endpoints.sh --run-dir <run_dir>` to generate `catalog.json` from host and plugin API DTOs.
9. Run `bash .agents/skills/lina-perf-audit/scripts/probe-fixtures.sh --run-dir <run_dir>` to generate `fixtures.json` and fail fast on declared routes that are not reachable.
10. Record skipped plugins with no backend API in `meta.json` using reason `no backend API`.

Always restore the temporary logger settings and stop services on success or failure with `bash .agents/skills/lina-perf-audit/scripts/restore-audit-env.sh --run-dir <run_dir>`.

### Stage 1: Concurrent Sub-Agent Audit

Use sub agents for endpoint audit tasks. Default to one sub agent per module from `catalog.json`. If a module is too large for a single prompt, split it into endpoint shards or small module shards; each sub agent prompt must remain under 5KB and include only its assigned endpoint subset.

Each sub agent receives:

- `module`
- `endpoints[]`
- `fixtures`
- `log_path`
- `token`
- `run_dir`

Each sub agent calls its assigned endpoints serially, reads the GoFrame `Trace-ID` response header, greps `server.log` for matching SQL lines, checks relevant controller/service source, classifies findings, and writes one audit markdown file under `temp/lina-perf-audit/<run-id>/audits/`. For every GET/read/query endpoint, the sub agent must also verify that the request trace does not contain write SQL statements (`INSERT`, `UPDATE`, `DELETE`, `REPLACE`, `TRUNCATE`, `ALTER`, `DROP`, or `CREATE`). Observed write SQL in a read request is a finding even when the endpoint is fast, except when the trace also contains read SQL and every write statement only touches `sys_online_session` and/or `plugin_monitor_operlog`.

If `Trace-ID` is unavailable, the sub agent must use a fallback search by call timestamp window plus request URL and mark the evidence as `trace ID unavailable, evidence quality reduced`. It must not skip the endpoint solely because `Trace-ID` is missing.

Before launching Stage 1, load `references/sub-agent-prompt.md` and use it as the base template.

### Stage 2: Summary And Perf-Issue Cards

1. Run `bash .agents/skills/lina-perf-audit/scripts/aggregate-reports.sh --run-dir <run_dir>`.
2. The script reads all `audits/*.md`, merges results into `temp/lina-perf-audit/<run-id>/SUMMARY.md`, and classifies findings into HIGH, MEDIUM, and LOW using the report severities reviewed against `references/severity-rubric.md`.
3. The script generates or updates one persistent issue card per finding under `perf-issues/<severity>-<module>-<slug>.md`.
4. The script de-duplicates cards by fingerprint using `references/fingerprint-rule.md`.
5. The script regenerates `perf-issues/INDEX.md` with all `open` and `in-progress` cards ordered by severity.
6. The script writes final `meta.json` with start/end time, git commit, stress fixture status, sub agent count, sub agent status, skipped plugins, logger settings, and restore result.
7. Confirm `SUMMARY.md` links to every new or updated issue card using repository-relative paths.

`temp/lina-perf-audit/<run-id>/` is a per-run snapshot. `perf-issues/` is a cross-run persistent backlog and must not be placed under `temp/`.

## Destructive Endpoint Handling

Sub agents must handle destructive endpoints without damaging shared audit data.

- For DELETE, reset, clear, uninstall, and equivalent destructive operations, first create a dedicated audit fixture in the same module when a matching create endpoint exists.
- Use the returned resource ID or stable key only for the destructive call being audited.
- Attempt cleanup even if the destructive call fails.
- Mark the report entry with `autonomous fixture completed, shared data not polluted`.
- If no same-module create endpoint exists, mark the endpoint as `SKIPPED: no matching create endpoint, manual follow-up required`.
- Never uninstall built-in plugins, clear shared system logs, or destroy global state as a substitute for module-owned fixture handling.

## Severity Classification

Use three severities:

- `HIGH`: N+1 on list/detail with source evidence, missing index on potentially large data, non-batch endpoint over 1s, blocking remote/file/transaction work inside loops, or any GET/read/query endpoint whose own request trace executes non-operational write SQL.
- `MEDIUM`: small-sample N+1, missing pagination, repeated same-data reads, or multiple SELECT calls that should be merged by JOIN or WHERE IN.
- `LOW`: slightly high SQL count with indexed fast queries, application-layer filtering that can be pushed down, or static risk not observed at runtime.

Every finding must include method and path, module, trace ID or fallback marker, SQL count, write SQL count when applicable, key SQL excerpts, relative source file and line, and at least one concrete remediation suggestion.

## Read Request Side-Effect Check

GET endpoints and endpoints described as list, query, tree, option, count, health, current, or detail are expected to be read-only. Their traced SQL must not contain write operations. The check covers SQL statements whose first significant token is one of `INSERT`, `UPDATE`, `DELETE`, `REPLACE`, `TRUNCATE`, `ALTER`, `DROP`, or `CREATE`.

- Count only SQL generated by the audited endpoint trace. Do not count Stage 0 setup, login, stress fixture insertion, or autonomous fixture create/delete calls for destructive endpoint handling.
- If a read endpoint writes only `sys_online_session` and/or `plugin_monitor_operlog` while also reading data, treat that session heartbeat or operation-log write as an expected operational side effect. Record it only as a PASS note in the module audit file, and do not emit a finding, summary violation, or `perf-issues/` card for it.
- If the trace includes only those expected write tables but does not include any read SQL, keep it reportable because it no longer matches the normal read-plus-operational-side-effect pattern.
- If a read endpoint writes any other business, plugin state, runtime state, or storage table, report it as `HIGH` with anti-pattern signature prefix `read-write-side-effect`.
- The remediation must recommend moving the side effect to an explicit POST/PUT/DELETE action, splitting the endpoint into read and write operations, or replacing persistent writes with a non-mutating cache/read model.
- If source code suggests a write but runtime SQL does not show it in the sampled path, report `LOW` with a static-only note instead of fabricating runtime evidence.

## Report Schema

Per-run outputs:

```text
temp/lina-perf-audit/<run-id>/
  catalog.json
  fixtures.json
  server.log
  audits/<module-or-shard>.md
  SUMMARY.md
  meta.json
```

Persistent issue cards:

```text
perf-issues/
  HIGH-<module>-<slug>.md
  MEDIUM-<module>-<slug>.md
  LOW-<module>-<slug>.md
  INDEX.md
```

Use `references/report-template.md` for run reports and `references/issue-card-template.md` for issue cards.

## Issue Card Lifecycle

Each performance issue becomes one persistent markdown card. Stage 2 owns all card creation and updates.

- New issue: create `perf-issues/<severity>-<module>-<slug>.md` with `status: open`.
- Existing fingerprint: update `last_seen_run`, increment `seen_count`, append a history entry, and do not create a duplicate card.
- Existing card with `status: fixed` or `status: obsolete` and the issue is observed again: change status back to `open` and append a regression history entry.
- Existing card with `status: in-progress`: keep status and update observation fields.
- Regenerate `perf-issues/INDEX.md` from current cards after all updates.
- Each card body must include the fixed sections `问题描述`, `复现方式`, `证据`, `改进方案`, and `历史记录`.
- Descriptive card content and `perf-issues/INDEX.md` headings must be written in Chinese. Keep API paths, SQL excerpts, Trace IDs, fingerprints, frontmatter field names, and status enum values unchanged.

Allowed statuses are `open`, `in-progress`, `fixed`, and `obsolete`. The skill must not introduce an automated state machine beyond these field updates.

## Cross-Reference Rules

- `SUMMARY.md` must link to all issue cards created or updated in the current run using repository-relative paths.
- Each issue card history entry must reference the `run_id` and `temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`.
- Persistent cards should use repository-relative source paths, for example `apps/lina-core/internal/service/user/user.go:142`.
- Do not use absolute local machine paths in persistent cards.
- Do not move issue cards into OpenSpec archive directories. Later OpenSpec changes may consume or update cards, but archiving this skill does not archive `perf-issues/`.
