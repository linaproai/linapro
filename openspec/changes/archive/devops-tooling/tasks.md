## 1. P0 Framework metadata unification

- [x] 1.1 Add a framework metadata section to `metadata.yaml` so name, version, description, homepage, repository URL, and license are managed in one place.
- [x] 1.2 Return that framework metadata from the system-info API and drive the top project card on the system-info page from backend data.

## 2. P0 Formal source upgrade command

- [x] 2.1 Add the repository-root `hack/upgrade-source` development-time tool and wire it into `make upgrade` with explicit confirmation.
- [x] 2.2 Perform backup reminders, Git dirty-worktree checks, current-version loading, and target-version comparison before upgrade execution.
- [x] 2.3 Implement target-tag fetch and local framework code overlay.
- [x] 2.4 Replay host SQL from the first file in order after the target source is applied.
- [x] 2.5 Exit safely with a clear message when the target version is not higher than the current project version.

## 3. Source-plugin upgrade governance

- [x] 3.1 Extend `make upgrade` and the repository-root upgrade tool so they support `scope=framework|source-plugin`, `plugin=<id|all>`, and a shared `dry-run` plan mode.
- [x] 3.2 Adjust source-plugin scan and governance sync logic so `sys_plugin.version` and `release_id` always represent the current effective version and higher discovered versions only become prepared releases.
- [x] 3.3 Implement the explicit source-plugin upgrade flow: version comparison, single-plugin and bulk plans, `phase=upgrade` SQL execution, menu and permission synchronization, governance resource-reference synchronization, and release/registry switching.
- [x] 3.4 Add a startup-time pending-upgrade check for source plugins. If an installed plugin has a higher discovered version that has not been upgraded, block startup and print the matching `make upgrade` command.
- [x] 3.5 Clarify the dynamic-plugin upgrade boundary so runtime upload plus install/reconcile remains the only upgrade path and `make upgrade` never takes over that flow.
- [x] 3.6 Update the related documentation, including the current OpenSpec artifacts, `apps/lina-core/README.md`, `apps/lina-core/README.zh-CN.md`, command help, and plugin governance guidance.

## 4. Development database configuration deduplication

- [x] 4.1 Update `apps/lina-core/hack/config.yaml` to use YAML anchors for the host development-only database connection settings and remove `multiStatements=true`.
- [x] 4.2 Review and update any development-only consumers of `hack/config.yaml` so upgrade tooling and local commands still read the unified connection settings correctly.
- [x] 4.3 Implement SQL file splitting and statement-by-statement execution under `apps/lina-core/internal/cmd/` while preserving ordered execution and fail-fast semantics.
- [x] 4.4 Adjust error and log context so statement failures still identify the relevant SQL file.

## 5. Cross-platform installation scripts

- [x] 5.1 Add `install.sh` and `install.ps1` under the repository root `hack/scripts/install/`, uniformly defining core parameter semantics and help messages for repository, `ref`, current directory / specified directory, overwrite protection, etc.
- [x] 5.2 Implement source archive download, temporary directory extraction, dynamic top-level directory identification, and target directory deployment logic in both scripts, ensuring the main flow does not depend on `git clone`.
- [x] 5.3 Implement safe directory policy: default deployment to a safe directory, explicit support for current directory mode, and refusal to continue execution when the target directory is non-empty and overwrite is not allowed.
- [x] 5.4 Add environment health check output for key dependencies such as `Go`, `Node.js`, `pnpm`, `MySQL`, `make`, as well as project path hints, at the end of the installation scripts.
- [x] 5.5 Output unified post-installation next-step guidance, clearly listing recommended operations such as `make init`, `make mock`, `make dev` and related notes.
- [x] 5.6 Update the repository root `README.md` and `README.zh-CN.md`, adding quick install examples for `macOS/Linux` and `Windows`, parameter descriptions, and official entry point mapping.

## 6. Cron job management: built-in cleanup task projection

- [x] 6.1 Remove SQL seed data for the built-in cleanup task from `sys_job`, changing it to only be generated through source code registration and startup projection synchronization.

## 7. P1 reserved direction

- [x] 7.1 Keep runtime business-system upgrades as a future direction in the current change, without implementing them in this iteration.

## 8. Verification

- [x] 8.1 Add automated tests for version comparison, target-tag resolution, and Git worktree cleanliness checks.
- [x] 8.2 Add automated tests for full host-SQL replay during upgrades.
- [x] 8.3 Add unit tests for the effective-version vs discovered-version split across uninstalled, same-version, and higher-discovered-version scenarios.
- [x] 8.4 Add tests for source-plugin upgrade commands covering single-plugin upgrades, `plugin=all`, dry-run, lower-version rejection, and not-installed plugin handling.
- [x] 8.5 Add startup fail-fast tests that confirm the host refuses to start when a source-plugin upgrade is pending.
- [x] 8.6 Add regression coverage confirming the development-time upgrade command does not interfere with dynamic-plugin runtime upgrades.
- [x] 8.7 Add command-layer unit tests that cover multi-statement splitting, comment/blank skipping, semicolons inside strings, and failure interruption.
- [x] 8.8 Run the affected Go unit tests and record the results to confirm stable behavior after the development tooling changes.
- [x] 8.9 Add automated verification or minimal executable smoke tests for the installation scripts, covering core behaviors such as parameter parsing, archive download URL generation, directory protection, and error prompts.
- [x] 8.10 Run script-related verification to confirm that the `macOS/Linux` and `Windows` entry points maintain consistency in core parameter contracts.

## 9. Performance audit skill: preparation

- [x] 9.1 Create `.agents/skills/lina-perf-audit/` with `SKILL.md`, `references/`, and `scripts/`.
- [x] 9.2 Add `.agents/skills/lina-perf-audit/scripts/README.md` and `README.zh-CN.md` explaining that scripts are owned by the skill.
- [x] 9.3 Verify that the GoFrame `Trace-ID` response header is available in this repository and that the same trace ID can be found in SQL debug logs.

## 10. Performance audit skill: helper scripts

- [x] 10.1 Implement `setup-audit-env.sh`: stop services, back up `logger.path` and `logger.file`, patch audit logging to `temp/lina-perf-audit/<run-id>/server.log`, start the backend, wait for health, and write the `admin/admin123` token.
- [x] 10.2 Implement `restore-audit-env.sh`: restore logger settings from run metadata and stop services on both success and failure paths.
- [x] 10.3 Implement `prepare-builtin-plugins.sh`: scan built-in plugin manifests, sync, install, enable all built-in plugins, and load plugin mock data when present.
- [x] 10.4 Implement `scan-endpoints.sh`: scan host and built-in plugin API DTOs, parse route metadata, group endpoints by module, generate `catalog.json`, and record plugins without backend APIs as skipped.
- [x] 10.5 Implement `probe-fixtures.sh`: call list endpoints to collect sample resource IDs into `fixtures.json` and fail when declared runtime routes are unreachable.
- [x] 10.6 Implement `stress-fixture.sh`: insert audit-only stress data after host and plugin mock data are loaded, use idempotent inserts, and verify delivery SQL directories are not modified.
- [x] 10.7 Implement `aggregate-reports.sh`: merge sub-agent reports, generate `SUMMARY.md`, create or update persistent `perf-issues/` cards by fingerprint, and regenerate `perf-issues/INDEX.md`.

## 11. Performance audit skill: documentation

- [x] 11.1 Add frontmatter with `MANUAL TRIGGER ONLY`, destructive setup warnings, resource-cost notes, and automation restrictions.
- [x] 11.2 Document explicit use cases and the confirmation gate for ambiguous performance requests.
- [x] 11.3 Document forbidden use cases: automatic invocation from other skills, CI, scheduled jobs, git hooks, background workflows, or single-endpoint investigations unless a full audit is requested.
- [x] 11.4 Document the three-stage workflow: preparation, concurrent sub-agent audit, and summary/issue-card aggregation.
- [x] 11.5 Document the sub-agent prompt payload, prompt-size budget, and output contract.
- [x] 11.6 Document destructive endpoint handling, including create-call-delete cleanup and manual follow-up when no create endpoint exists.
- [x] 11.7 Document HIGH / MEDIUM / LOW severity classification.
- [x] 11.8 Document the per-run report schema and persistent issue-card schema.
- [x] 11.9 Document issue-card lifecycle, fingerprint de-duplication, reopening regressions, and index regeneration.
- [x] 11.10 Document cross-reference rules and the fact that `perf-issues/` is not moved into OpenSpec archives.

## 12. Performance audit skill: references

- [x] 12.1 Add `references/sub-agent-prompt.md`.
- [x] 12.2 Add `references/severity-rubric.md`.
- [x] 12.3 Add `references/report-template.md`.
- [x] 12.4 Add `references/issue-card-template.md`.
- [x] 12.5 Add `references/fingerprint-rule.md`.

## 13. Performance audit skill: README

- [x] 13.1 Add `.agents/skills/lina-perf-audit/README.md` and `README.zh-CN.md`.
- [x] 13.2 Add a short note in the root command documentation that `lina-perf-audit` is a manually triggered performance audit skill.

## 14. Performance audit skill: dry-run verification

- [x] 14.1 Run one full dry run from a reset local environment.
- [x] 14.2 Verify `logger.path` and `logger.file` are restored after the run.
- [x] 14.3 Verify `apps/lina-core/manifest/sql/` and `apps/lina-core/manifest/sql/mock-data/` have no residual diff after the run.
- [x] 14.4 Verify the run directory contains module audit files, `SUMMARY.md`, and `meta.json`.
- [x] 14.5 Verify the run either detects at least one HIGH N+1 case or explicitly records that no HIGH case was found.
- [x] 14.6 Verify at least one destructive endpoint uses the autonomous create-call-delete lifecycle.
- [x] 14.7 Verify `perf-issues/` cards are created with valid filenames, frontmatter, and required body sections.
- [x] 14.8 Verify `perf-issues/INDEX.md` lists all open and in-progress cards by severity.
- [x] 14.9 Run a second dry run and verify fingerprint de-duplication updates existing cards instead of creating duplicates.
- [x] 14.10 Mark one card as `fixed`, rerun aggregation, and verify the card reopens when the issue is observed again.
- [x] 14.11 Verify `perf-issues/` is outside `temp/`, is not ignored, is not deleted by temp cleanup, and is not moved by OpenSpec archive.

## 15. Performance audit skill: manual trigger verification

- [x] 15.1 Confirm `SKILL.md` description contains `MANUAL TRIGGER ONLY`.
- [x] 15.2 Simulate an ambiguous request and verify the skill asks for confirmation before running destructive setup.
- [x] 15.3 Grep other skill files and confirm no other skill references or invokes `lina-perf-audit`.

## 16. Performance audit skill: review and validation

- [x] 16.1 Run `openspec validate add-lina-perf-audit-skill --type change --strict`.
- [x] 16.2 Run `lina-review` for code and specification compliance.
- [x] 16.3 Fix review findings and rerun review until no critical issue remains.

## Feedback (existing iterations)

- [x] **FB-1**: Converge `make upgrade` implementation under `hack/upgrade-source/` and read only database connection and upgrade metadata from `apps/lina-core/hack/config.yaml`.
- [x] **FB-2**: Let `init` and `mock` switch SQL asset sources by execution phase: runtime uses embedded SQL, while development-time `Makefile` entries use local SQL files explicitly.
- [x] **FB-3**: Treat `homepage` as the official website and add a separate repository URL field for system-info presentation and upgrade tooling.
- [x] **FB-4**: Re-group `internal/cmd` unit tests by command responsibility instead of preserving file names tied to removed helpers.
- [x] **FB-5**: Keep non-test logic for `init`, `mock`, and `upgrade` close to their corresponding command files or `cmd.go` to avoid scattered helpers.
- [x] **FB-6**: Do not introduce upgrade state tables, upgrade record tables, or SQL cursor tables for source upgrades.
- [x] **FB-7**: Replay host SQL from the first SQL file during `make upgrade` instead of relying on persisted execution position.
- [x] **FB-8**: Rename the development-time tool directory from `hack/upgrade-framework` to `hack/upgrade-source` and keep only `main.go` at the component root.
- [x] **FB-9**: Extract source-plugin upgrade governance into an independent host-side component that is reused by both `make upgrade` and startup validation. Keep the `pkg` layer limited to stable contracts and small facades.
- [x] **FB-10**: Add automated validation for source-plugin upgrade governance, including unit tests for the host-side `sourceupgrade` component/facade and E2E coverage in `TC0106` for the "new version discovered but not yet activated" scenario.
- [x] **FB-11**: Fix the runtime WASM oversize-upload E2E assertion so it no longer hard-codes the outdated `16MB` limit and restore full `e2e/extension/plugin` regression coverage.
- [x] **FB-12**: Change the installation script directory convention from `hack/scripts/` to `hack/scripts/install/`, and synchronize path references in proposal, design, spec, and tasks.
- [x] **FB-13**: Add reusable local/CI execution entry points for installation script smoke tests, avoiding verification methods that only rely on manual commands.
- [x] **FB-14**: Add Chinese and English documentation under `hack/scripts/install/`, describing the installation script's purpose, parameters, usage, and verification methods.
- [x] **FB-15**: When the user does not explicitly pass `ref`, default to resolving and installing the latest stable tag version of the repository; if no stable tag exists, fall back to the default main branch, and display the final resolved reference value in the output.
- [x] **FB-16**: Migrate repository-level standalone Go tools to `hack/tools/`, and synchronize directory references in build, upgrade, test, and specification documents.
- [x] **FB-17**: Remove SQL seed data for the built-in cleanup task from `sys_job`, changing it to only be generated through source code registration and startup projection synchronization.
- [x] **FB-18**: Split the core HTTP startup function and add key logic comments to reduce the single-function complexity of `cmd_http.go`.
- [x] **FB-19**: Unify development build and backend runtime relative paths to the repository root `temp/`, avoiding the generation of duplicate `temp` directories under `apps/lina-core`.
- [x] **FB-20**: Fix the issue where the cron expression column in the cron job list has insufficient contrast and is hard to read in dark theme.

## Feedback (performance audit skill)

- [x] **FB-PA-1**: Ensure the audit scope covers all built-in plugins.
- [x] **FB-PA-2**: Move cross-run issue cards to repository-root `perf-issues/`.
- [x] **FB-PA-3**: Move the skill to the generic `.agents/skills/lina-perf-audit/` directory.
- [x] **FB-PA-4**: Add checks for query/read requests that execute write SQL.
- [x] **FB-PA-5**: Move helper scripts into the skill-owned `scripts/` directory.
- [x] **FB-PA-6**: Write persistent issue-card descriptions in Chinese while keeping machine-readable fields unchanged.
- [x] **FB-PA-7**: Do not create issue cards for read requests that only write `sys_online_session` or `plugin_monitor_operlog` while also reading data.
- [x] **FB-PA-8**: Fix the job-log list dynamic-plugin i18n metadata N+1 case.
- [x] **FB-PA-9**: Change the dynamic plugin `host-call-demo` endpoint from GET to POST because it writes runtime state.
- [x] **FB-PA-10**: Reduce repeated dynamic-plugin localization metadata reads in job list/detail and keyword search paths.
- [x] **FB-PA-11**: Replace per-job-group job counts with grouped batch counting.
- [x] **FB-PA-12**: Reduce repeated menu/plugin runtime metadata reads in menu list paths.
- [x] **FB-PA-13**: Reduce repeated dynamic-plugin and release-state reads in plugin list projection.
- [x] **FB-PA-14**: Replace per-menu role association inserts with batch insertion.
- [x] **FB-PA-15**: Batch localize monitor operation-log route metadata.
- [x] **FB-PA-16**: Add cluster-aware plugin runtime cache revision coordination.
- [x] **FB-PA-17**: Optimize the dynamic-plugin reconciler in cluster mode by using shared revisions plus a low-frequency fallback scan.

## Performance Audit Execution Record

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

## Performance Audit Feedback Verification Summary

- FB-PA-8 through FB-PA-15 targeted API verification was run under `temp/lina-perf-audit/perf-issues-regression-20260502-205731/`. The original N+1, repeated metadata reads, loop writes, and read-endpoint write-side-effect issues were not reproduced.
- FB-PA-16 cache consistency assessment: the authority is `sys_plugin`, `sys_plugin_release`, and artifact storage; cluster synchronization uses `sys_kv_cache` revisions; single-node mode remains local; cluster mode invalidates through shared revisions; normal trigger delay is about 2 seconds; recovery includes a fallback scan.
- FB-PA-17 reconciler assessment: in cluster mode the reconciler polls a shared revision every 2 seconds and performs a full scan only when a new revision is observed or the 5-minute fallback window elapses. Single-node mode keeps direct local behavior.
- Focused tests recorded in the active iteration included `go test` for `jobmgmt`, `i18n`, `plugin/...`, `role`, `apidoc`, `cmd`, `monitor-operlog`, and `plugin-demo-dynamic`; `openspec validate add-lina-perf-audit-skill --type change --strict`; and `git diff --check`.
- i18n impact: the skill itself has no runtime i18n or apidoc i18n impact. Feedback code changes either had no user-visible text change or updated the relevant apidoc resources when the dynamic plugin endpoint method changed.
- Cache impact: the skill itself does not add a production cache. Feedback cache changes explicitly document authority, consistency, invalidation, cross-node revision synchronization, bounded staleness, and fallback behavior.
