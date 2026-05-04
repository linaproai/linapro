## 1. Preparation

- [x] 1.1 Create `.agents/skills/lina-perf-audit/` with `SKILL.md`, `references/`, and `scripts/`.
- [x] 1.2 Add `.agents/skills/lina-perf-audit/scripts/README.md` and `README.zh_CN.md` explaining that scripts are owned by the skill.
- [x] 1.3 Verify that the GoFrame `Trace-ID` response header is available in this repository and that the same trace ID can be found in SQL debug logs.

## 2. Helper Scripts

- [x] 2.1 Implement `setup-audit-env.sh`: stop services, back up `logger.path` and `logger.file`, patch audit logging to `temp/lina-perf-audit/<run-id>/server.log`, start the backend, wait for health, and write the `admin/admin123` token.
- [x] 2.2 Implement `restore-audit-env.sh`: restore logger settings from run metadata and stop services on both success and failure paths.
- [x] 2.3 Implement `prepare-builtin-plugins.sh`: scan built-in plugin manifests, sync, install, enable all built-in plugins, and load plugin mock data when present.
- [x] 2.4 Implement `scan-endpoints.sh`: scan host and built-in plugin API DTOs, parse route metadata, group endpoints by module, generate `catalog.json`, and record plugins without backend APIs as skipped.
- [x] 2.5 Implement `probe-fixtures.sh`: call list endpoints to collect sample resource IDs into `fixtures.json` and fail when declared runtime routes are unreachable.
- [x] 2.6 Implement `stress-fixture.sh`: insert audit-only stress data after host and plugin mock data are loaded, use idempotent inserts, and verify delivery SQL directories are not modified.
- [x] 2.7 Implement `aggregate-reports.sh`: merge sub-agent reports, generate `SUMMARY.md`, create or update persistent `perf-issues/` cards by fingerprint, and regenerate `perf-issues/INDEX.md`.

## 3. Skill Documentation

- [x] 3.1 Add frontmatter with `MANUAL TRIGGER ONLY`, destructive setup warnings, resource-cost notes, and automation restrictions.
- [x] 3.2 Document explicit use cases and the confirmation gate for ambiguous performance requests.
- [x] 3.3 Document forbidden use cases: automatic invocation from other skills, CI, scheduled jobs, git hooks, background workflows, or single-endpoint investigations unless a full audit is requested.
- [x] 3.4 Document the three-stage workflow: preparation, concurrent sub-agent audit, and summary/issue-card aggregation.
- [x] 3.5 Document the sub-agent prompt payload, prompt-size budget, and output contract.
- [x] 3.6 Document destructive endpoint handling, including create-call-delete cleanup and manual follow-up when no create endpoint exists.
- [x] 3.7 Document HIGH / MEDIUM / LOW severity classification.
- [x] 3.8 Document the per-run report schema and persistent issue-card schema.
- [x] 3.9 Document issue-card lifecycle, fingerprint de-duplication, reopening regressions, and index regeneration.
- [x] 3.10 Document cross-reference rules and the fact that `perf-issues/` is not moved into OpenSpec archives.

## 4. References

- [x] 4.1 Add `references/sub-agent-prompt.md`.
- [x] 4.2 Add `references/severity-rubric.md`.
- [x] 4.3 Add `references/report-template.md`.
- [x] 4.4 Add `references/issue-card-template.md`.
- [x] 4.5 Add `references/fingerprint-rule.md`.

## 5. Skill README

- [x] 5.1 Add `.agents/skills/lina-perf-audit/README.md` and `README.zh_CN.md`.
- [x] 5.2 Add a short note in the root command documentation that `lina-perf-audit` is a manually triggered performance audit skill.

## 6. Dry-Run Verification

- [x] 6.1 Run one full dry run from a reset local environment.
- [x] 6.2 Verify `logger.path` and `logger.file` are restored after the run.
- [x] 6.3 Verify `apps/lina-core/manifest/sql/` and `apps/lina-core/manifest/sql/mock-data/` have no residual diff after the run.
- [x] 6.4 Verify the run directory contains module audit files, `SUMMARY.md`, and `meta.json`.
- [x] 6.5 Verify the run either detects at least one HIGH N+1 case or explicitly records that no HIGH case was found.
- [x] 6.6 Verify at least one destructive endpoint uses the autonomous create-call-delete lifecycle.
- [x] 6.7 Verify `perf-issues/` cards are created with valid filenames, frontmatter, and required body sections.
- [x] 6.8 Verify `perf-issues/INDEX.md` lists all open and in-progress cards by severity.
- [x] 6.9 Run a second dry run and verify fingerprint de-duplication updates existing cards instead of creating duplicates.
- [x] 6.10 Mark one card as `fixed`, rerun aggregation, and verify the card reopens when the issue is observed again.
- [x] 6.11 Verify `perf-issues/` is outside `temp/`, is not ignored, is not deleted by temp cleanup, and is not moved by OpenSpec archive.

## 7. Manual Trigger Verification

- [x] 7.1 Confirm `SKILL.md` description contains `MANUAL TRIGGER ONLY`.
- [x] 7.2 Simulate an ambiguous request and verify the skill asks for confirmation before running destructive setup.
- [x] 7.3 Grep other skill files and confirm no other skill references or invokes `lina-perf-audit`.

## 8. Review and Validation

- [x] 8.1 Run `openspec validate add-lina-perf-audit-skill --type change --strict`.
- [x] 8.2 Run `lina-review` for code and specification compliance.
- [x] 8.3 Fix review findings and rerun review until no critical issue remains.

## Execution Record

- Main dry-run directory: `temp/lina-perf-audit/20260501-233924/`.
- Endpoint catalog: `171` endpoints across `26` modules, including `121` host endpoints and `50` built-in plugin endpoints. `demo-control` was marked skipped because it has no backend API.
- Sub-agent reports: `22` module or shard reports under `audits/`, covering all `26` modules. `core-small` combined `core:auth`, `core:health`, `core:i18n`, `core:publicconfig`, and `core:sysinfo`.
- Initial summary: `18` HIGH, `6` MEDIUM, and `0` LOW findings. After the operational side-effect exception was added, expected writes limited to `sys_online_session` and `plugin_monitor_operlog` were filtered out of persistent issue cards.
- Persistent issue cards: cards were generated under repository-root `perf-issues/` with required frontmatter and body sections. Fingerprint de-duplication, `seen_count`, `last_seen_run`, and fixed-card regression reopening were verified with isolated card lifecycle runs.
- Environment restoration: `restore-audit-env.sh` restored `apps/lina-core/manifest/config/config.yaml`; service status showed frontend and backend stopped; delivery SQL directories had no residual diff.
- Script checks: `bash -n .agents/skills/lina-perf-audit/scripts/*.sh` passed.
- JSON checks: run JSON artifacts passed `python3 -m json.tool`.
- Build check: `go build -o temp/bin/lina-perf-audit-build-check ./apps/lina-core` passed.
- OpenSpec validation: `openspec validate add-lina-perf-audit-skill --type change --strict` passed.

## Feedback Completed

- [x] FB-1: Ensure the audit scope covers all built-in plugins.
- [x] FB-2: Move cross-run issue cards to repository-root `perf-issues/`.
- [x] FB-3: Move the skill to the generic `.agents/skills/lina-perf-audit/` directory.
- [x] FB-4: Add checks for query/read requests that execute write SQL.
- [x] FB-5: Move helper scripts into the skill-owned `scripts/` directory.
- [x] FB-6: Write persistent issue-card descriptions in Chinese while keeping machine-readable fields unchanged.
- [x] FB-7: Do not create issue cards for read requests that only write `sys_online_session` or `plugin_monitor_operlog` while also reading data.
- [x] FB-8: Fix the job-log list dynamic-plugin i18n metadata N+1 case.
- [x] FB-9: Change the dynamic plugin `host-call-demo` endpoint from GET to POST because it writes runtime state.
- [x] FB-10: Reduce repeated dynamic-plugin localization metadata reads in job list/detail and keyword search paths.
- [x] FB-11: Replace per-job-group job counts with grouped batch counting.
- [x] FB-12: Reduce repeated menu/plugin runtime metadata reads in menu list paths.
- [x] FB-13: Reduce repeated dynamic-plugin and release-state reads in plugin list projection.
- [x] FB-14: Replace per-menu role association inserts with batch insertion.
- [x] FB-15: Batch localize monitor operation-log route metadata.
- [x] FB-16: Add cluster-aware plugin runtime cache revision coordination.
- [x] FB-17: Optimize the dynamic-plugin reconciler in cluster mode by using shared revisions plus a low-frequency fallback scan.

## Feedback Verification Summary

- FB-8 through FB-15 targeted API verification was run under `temp/lina-perf-audit/perf-issues-regression-20260502-205731/`. The original N+1, repeated metadata reads, loop writes, and read-endpoint write-side-effect issues were not reproduced.
- FB-16 cache consistency assessment: the authority is `sys_plugin`, `sys_plugin_release`, and artifact storage; cluster synchronization uses `sys_kv_cache` revisions; single-node mode remains local; cluster mode invalidates through shared revisions; normal trigger delay is about 2 seconds; recovery includes a fallback scan.
- FB-17 reconciler assessment: in cluster mode the reconciler polls a shared revision every 2 seconds and performs a full scan only when a new revision is observed or the 5-minute fallback window elapses. Single-node mode keeps direct local behavior.
- Focused tests recorded in the active iteration included `go test` for `jobmgmt`, `i18n`, `plugin/...`, `role`, `apidoc`, `cmd`, `monitor-operlog`, and `plugin-demo-dynamic`; `openspec validate add-lina-perf-audit-skill --type change --strict`; and `git diff --check`.
- i18n impact: the skill itself has no runtime i18n or apidoc i18n impact. Feedback code changes either had no user-visible text change or updated the relevant apidoc resources when the dynamic plugin endpoint method changed.
- Cache impact: the skill itself does not add a production cache. Feedback cache changes explicitly document authority, consistency, invalidation, cross-node revision synchronization, bounded staleness, and fallback behavior.
