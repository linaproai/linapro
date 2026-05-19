## 1. Cross-Platform Dev Commands: Tool Skeleton and Entry Points

- [x] 1.1 Add `hack/tools/linactl` Go CLI tool directory with `go.mod`, command dispatch framework, error output conventions, and exit code handling
- [x] 1.2 Add `hack/tools/linactl` to `go.work`, and provide English `README.md` and Chinese `README.zh-CN.md`
- [x] 1.3 Implement make-style `key=value` parameter parsing, covering `confirm`, `rebuild`, `platforms`, `verbose`, `p`, and other existing parameters
- [x] 1.4 Add root `make.cmd` that forwards parameters to `go run ./hack/tools/linactl`

## 2. Cross-Platform Dev Commands: Low-Risk Target Migration

- [x] 2.1 Implement `help` command, outputting cross-platform command list, parameter descriptions, and Windows/Linux/macOS entry examples
- [x] 2.2 Implement `pack.assets` command (renamed from `prepare-packed-assets`), replacing `hack/scripts/prepare-packed-assets.sh` directory cleanup and manifest resource copying logic
- [x] 2.3 Implement `wasm` command with plugin scanning, dynamic type identification, and specified plugin build parameters
- [x] 2.4 Update root and sub-module Makefiles to call `linactl` for migrated low-risk targets

## 3. Cross-Platform Dev Commands: Development Service Migration

- [x] 3.1 Implement `status` command, displaying backend and frontend ports, PID files, log paths, and service readiness in cross-platform table format
- [x] 3.2 Implement `dev` command, orchestrating Wasm build, packed assets preparation, backend build, frontend Vite startup, log writing, and HTTP readiness wait as async processes
- [x] 3.3 Implement `stop` command, cross-platform stopping of identifiable backend and frontend processes and cleaning stale PID files
- [x] 3.4 Update `hack/makefiles/dev.mk` to call `linactl` for `dev`, `stop`, and `status`

## 4. Cross-Platform Dev Commands: Build, Database, and Test Migration

- [x] 4.1 Implement `build` command, orchestrating frontend build, resource embedding, dynamic plugin build, and multi-platform backend build
- [x] 4.2 Implement `image-build` (later renamed to `image.build`) and `image` cross-platform wrappers
- [x] 4.3 Implement `init` and `mock` commands with confirmation parameters, anti-misoperation prompts, and PostgreSQL connection failure diagnostics
- [x] 4.4 Implement `test`, `test-go`, `check-runtime-i18n`, and `check-runtime-i18n-messages` cross-platform wrappers
- [x] 4.5 Implement `cli.install`, `ctrl`, `dao`, `enums`, `service`, `pb`, `pbentity`, and other GoFrame CLI target cross-platform wrappers

## 5. Cross-Platform Dev Commands: Script Governance and Compatibility

- [x] 5.1 Evaluate and delete or downgrade `.sh` scripts no longer used as primary paths, marking retained historical entries as compatibility
- [x] 5.2 Converge root `Makefile` and `hack/makefiles/*.mk` migrated targets into thin wrappers
- [x] 5.3 Converge `apps/lina-core/Makefile` and `apps/lina-plugins/Makefile` migrated targets into thin wrappers
- [x] 5.4 Clarify entry priority and behavioral consistency between `make.cmd`, Makefile, and `linactl`

## 6. Cross-Platform Dev Commands: Documentation and Verification

- [x] 6.1 Update root `README.md` and `README.zh-CN.md` with cross-platform recommended entry, Windows `cmd.exe`, PowerShell, and Linux/macOS usage
- [x] 6.2 Update `hack/tools/README.md` and `hack/tools/README.zh-CN.md` with `linactl` tool responsibilities and maintenance rules
- [x] 6.3 Add Go unit tests covering parameter parsing, command dispatch, file copying, plugin scanning, help output, and error messages
- [x] 6.4 Add command-level smoke verification covering `make.cmd` parameter forwarding, Makefile thin-wrapper consistency, and key command exit codes
- [x] 6.5 Update `.github/workflows/` related GitHub Actions to add `windows-latest` basic command verification
- [x] 6.6 Cover `go run ./hack/tools/linactl help`, `go run ./hack/tools/linactl status`, and at least one lightweight file or plugin tool command in Windows CI
- [x] 6.7 Cover `cmd.exe` `make.cmd`/`make` and PowerShell `.\make.cmd` usage in Windows CI
- [x] 6.8 Run `go test` covering new tools and affected Go tool modules
- [x] 6.9 Run `openspec validate` for the cross-platform dev commands change

## 7. Linactl Build Tool Consolidation

- [x] 7.1 Migrate `hack/tools/image-builder` implementation to `hack/tools/linactl/internal/imagebuilder`
- [x] 7.2 Update `linactl image` and `linactl image.build` commands to call internal `imagebuilder`
- [x] 7.3 Migrate `hack/tools/build-wasm/internal/builder` to `hack/tools/linactl/internal/wasmbuilder`
- [x] 7.4 Update `linactl wasm` to directly call `wasmbuilder` with plugin workspace preparation, dry-run, and output directory semantics
- [x] 7.5 Migrate `hack/tools/runtime-i18n` to `hack/tools/linactl/internal/runtimei18n`
- [x] 7.6 Update `linactl i18n.check` to directly call `internal/runtimei18n`
- [x] 7.7 Update `linactl/go.mod` and root `go.work`, remove old independent tool modules, and maintain `pluginbridge` dependencies
- [x] 7.8 Unify internal component source file naming to component name prefix convention
- [x] 7.9 Update CI fixtures, test helpers, E2E, and documentation references to new paths
- [x] 7.10 Delete `hack/tools/image-builder`, `hack/tools/build-wasm`, and `hack/tools/runtime-i18n` old independent tool directories
- [x] 7.11 Update bilingual README documentation

## 8. Dev Environment Commands

- [x] 8.1 Add `linactl env.check` command to detect Go, Node.js, pnpm, Vite, Playwright, PostgreSQL and output bordered table with name, current version, required version, satisfied status, and remarks
- [x] 8.2 Migrate original `linactl dev.setup` implementation to `linactl env.setup`, preserving frontend dependency and Playwright Chromium installation behavior
- [x] 8.3 Remove `dev.setup` from `linactl` registry and root Make targets, add `env.check`/`env.setup` Make targets
- [x] 8.4 Update Playwright missing browser prompts and documentation/help output to point to `make env.setup`
- [x] 8.5 Add or update `hack/tools/linactl` unit tests covering environment command registration, old command removal, environment check table, and command file naming governance

## 9. Image Build Workflows: Release Test and Build

- [x] 9.1 Record differences between nightly and release workflow jobs, permissions, triggers, and artifacts
- [x] 9.2 Rename release workflow to `.github/workflows/release-test-and-build.yml` with workflow name `Release Test and Build`
- [x] 9.3 Introduce `verification-suite` job calling `reusable-test-verification-suite.yml` with Main CI's brief test scope (no E2E)
- [x] 9.4 Let `release-image` depend on tag version check and `verification-suite` via `needs`
- [x] 9.5 Preserve release tag validation, multi-architecture image push, `latest` floating tag, and manifest inspect
- [x] 9.6 Add `create-github-release` job after successful tag validation, test suite, and image publishing
- [x] 9.7 Extract E2E, Redis cluster smoke, host-only build smoke, image publishing, and official plugin workspace preflight into reusable workflows and composite actions
- [x] 9.8 Add bilingual comments to all workflow files, jobs, and inline steps

## 10. Image Build Workflows: Manual Nightly Image Build

- [x] 10.1 Add `.github/workflows/manual-nightly-image-build.yml` with `workflow_dispatch` trigger only
- [x] 10.2 Call `.github/workflows/reusable-image-publish.yml` directly with no test job or `needs` test dependency
- [x] 10.3 Configure same nightly image name, platforms, `tag-mode`, floating tag, and official plugin full build mode
- [x] 10.4 Add `hack/deploy/docker-compose.yaml` memory-only demo launcher with PostgreSQL service
- [x] 10.5 Add `hack/deploy/config.yaml` with runtime configuration separated from Compose startup script
- [x] 10.6 Add `hack/deploy/tests/docker-compose.yaml` for manual development container

## 11. Release Version Governance

- [x] 11.1 Add `linactl release.tag.check` command reading `metadata.yaml` and validating tag against `framework.version`
- [x] 11.2 Add unit tests for tag match, tag mismatch, invalid version format, missing version field, and environment variable fallback
- [x] 11.3 Add leading `release-tag-version-check` job in `Release Test and Build` workflow
- [x] 11.4 Make all release test and image publishing jobs depend on version check job
- [x] 11.5 Add controlled `Create Release Tag` manual workflow using GitHub App installation token
- [x] 11.6 Document GitHub tag ruleset configuration and `GITHUB_TOKEN` re-trigger limitation in `CONTRIBUTING.md`

## 12. Monthly OpenSpec Archive Automation

- [x] 12.1 Adjust `.gitignore` to allow committing `.github/codex/config.template.toml` and key-free templates
- [x] 12.2 Add `.github/workflows/monthly-openspec-archive.yml` main routing workflow with schedule and `workflow_dispatch`
- [x] 12.3 Add Codex reusable workflow `.github/workflows/monthly-openspec-archive-codex.yml`
- [x] 12.4 Add Claude Code reusable workflow `.github/workflows/monthly-openspec-archive-cc.yml`
- [x] 12.5 Add GitHub Copilot CLI reusable workflow `.github/workflows/monthly-openspec-archive-copilot.yml`
- [x] 12.6 Add shared composite actions: `monthly-openspec-setup`, `monthly-openspec-detect-changes`, `monthly-openspec-assert-archive-complete`, `monthly-openspec-validate`, `monthly-openspec-finalize-pr`
- [x] 12.7 Add shared prompt files `.github/prompts/monthly-openspec-auto-archive.zh-CN.md` and `.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`
- [x] 12.8 Implement runtime credential injection from GitHub Secrets and Variables
- [x] 12.9 Implement OpenSpec completion pre-check, conditional consolidation, phase-based fail-fast, change scope protection, and PR write-back

## 13. Upgrade Governance: Framework and Source-Plugin

- [x] 13.1 Add repository-root `hack/upgrade-source` development-time tool wired into `make upgrade` with explicit confirmation
- [x] 13.2 Perform backup reminders, Git dirty-worktree checks, current-version loading, and target-version comparison
- [x] 13.3 Implement target-tag fetch and local framework code overlay
- [x] 13.4 Replay host SQL from the first file in order after target source application
- [x] 13.5 Extend `make upgrade` with `scope=framework|source-plugin`, `plugin=<id|all>`, and shared `dry-run` plan mode
- [x] 13.6 Adjust source-plugin scan and governance sync so `sys_plugin.version` and `release_id` always represent the effective version
- [x] 13.7 Implement explicit source-plugin upgrade flow with `phase=upgrade` SQL execution, menu/permission synchronization, and release/registry switching
- [x] 13.8 Add startup-time pending-upgrade check that blocks startup when upgrades are pending
- [x] 13.9 Clarify dynamic-plugin upgrade boundary: runtime upload plus install/reconcile remains the only upgrade path

## 14. Framework Metadata Unification

- [x] 14.1 Add framework metadata section to `metadata.yaml` for centralized name, version, description, homepage, repository URL, and license management
- [x] 14.2 Return framework metadata from the system-info API and drive the project card from backend data

## 15. Development Database Configuration

- [x] 15.1 Update `apps/lina-core/hack/config.yaml` to use YAML anchors and remove `multiStatements=true`
- [x] 15.2 Review and update development-only consumers of `hack/config.yaml`
- [x] 15.3 Implement SQL file splitting and statement-by-statement execution preserving ordered execution and fail-fast semantics
- [x] 15.4 Adjust error and log context so statement failures identify the relevant SQL file

## 16. Cross-Platform Installation Scripts

- [x] 16.1 Add `install.sh` and `install.ps1` under `hack/scripts/install/` with core parameter semantics
- [x] 16.2 Implement source archive download, temporary directory extraction, dynamic top-level directory identification, and target directory deployment
- [x] 16.3 Implement safe directory policy with non-empty directory refusal
- [x] 16.4 Add environment health check output for Go, Node.js, pnpm, MySQL, and make
- [x] 16.5 Output unified post-installation next-step guidance
- [x] 16.6 Update repository root README with quick install examples

## 17. Cron Job Management

- [x] 17.1 Remove SQL seed data for built-in cleanup task from `sys_job`, using source code registration and startup projection

## 18. Performance Audit Skill: Preparation and Scripts

- [x] 18.1 Create `.agents/skills/lina-perf-audit/` with `SKILL.md`, `references/`, and `scripts/`
- [x] 18.2 Implement `setup-audit-env.sh`, `restore-audit-env.sh`, `prepare-builtin-plugins.sh`, `scan-endpoints.sh`, `probe-fixtures.sh`, `stress-fixture.sh`, and `aggregate-reports.sh`

## 19. Performance Audit Skill: Documentation and References

- [x] 19.1 Add SKILL.md frontmatter with `MANUAL TRIGGER ONLY`, destructive setup warnings, and automation restrictions
- [x] 19.2 Document three-stage workflow, sub-agent prompt payload, destructive endpoint handling, severity classification, report schema, and issue-card lifecycle
- [x] 19.3 Add reference files: `sub-agent-prompt.md`, `severity-rubric.md`, `report-template.md`, `issue-card-template.md`, `fingerprint-rule.md`
- [x] 19.4 Add README.md and README.zh-CN.md

## 20. Verification and Cross-Cutting Concerns

- [x] 20.1 Add automated tests for version comparison, target-tag resolution, and Git worktree cleanliness checks
- [x] 20.2 Add automated tests for full host-SQL replay during upgrades
- [x] 20.3 Add unit tests for effective-version vs discovered-version split
- [x] 20.4 Add tests for source-plugin upgrade commands
- [x] 20.5 Add startup fail-fast tests
- [x] 20.6 Add command-layer unit tests for multi-statement splitting, comment/blank skipping, semicolons inside strings, and failure interruption
- [x] 20.7 Run affected Go unit tests and record results
- [x] 20.8 Add automated verification for installation scripts
- [x] 20.9 Run performance audit dry-run and verify all outputs
- [x] 20.10 Confirm manual trigger constraint and no cross-skill invocation

## Feedback (Upgrade Governance and Database Bootstrap)

- [x] **FB-1**: Converge `make upgrade` under `hack/upgrade-source/` and read only database connection and upgrade metadata from hack config
- [x] **FB-2**: Let `init` and `mock` switch SQL asset sources by execution phase
- [x] **FB-3**: Treat `homepage` as official website and add separate repository URL field
- [x] **FB-4**: Re-group `internal/cmd` unit tests by command responsibility
- [x] **FB-5**: Keep non-test logic close to corresponding command files
- [x] **FB-6**: Do not introduce upgrade state tables or SQL cursor tables
- [x] **FB-7**: Replay host SQL from the first file during `make upgrade`
- [x] **FB-8**: Rename development-time tool directory to `hack/upgrade-source`
- [x] **FB-9**: Extract source-plugin upgrade governance into independent host-side component
- [x] **FB-10**: Add automated validation for source-plugin upgrade governance
- [x] **FB-11**: Fix runtime WASM oversize-upload E2E assertion
- [x] **FB-12**: Change installation script directory convention to `hack/scripts/install/`
- [x] **FB-13**: Add reusable local/CI execution entry points for installation script smoke tests
- [x] **FB-14**: Add Chinese and English documentation under `hack/scripts/install/`
- [x] **FB-15**: Default to latest stable tag version when user does not pass `ref`
- [x] **FB-16**: Migrate repository-level standalone Go tools to `hack/tools/`
- [x] **FB-17**: Remove SQL seed data for built-in cleanup task
- [x] **FB-18**: Split core HTTP startup function and add key logic comments
- [x] **FB-19**: Unify development build and backend runtime relative paths to `temp/`
- [x] **FB-20**: Fix cron expression column contrast in dark theme

## Feedback (Performance Audit Skill)

- [x] **FB-PA-1**: Cover all built-in plugins in audit scope
- [x] **FB-PA-2**: Move cross-run issue cards to repository-root `perf-issues/`
- [x] **FB-PA-3**: Move skill to `.agents/skills/lina-perf-audit/`
- [x] **FB-PA-4**: Add checks for query/read requests executing write SQL
- [x] **FB-PA-5**: Move helper scripts into skill-owned `scripts/` directory
- [x] **FB-PA-6**: Write persistent issue-card descriptions in Chinese
- [x] **FB-PA-7**: Filter operational side-effect writes to `sys_online_session` or `plugin_monitor_operlog`
- [x] **FB-PA-8**: Fix job-log list dynamic-plugin i18n metadata N+1
- [x] **FB-PA-9**: Change dynamic plugin `host-call-demo` endpoint from GET to POST
- [x] **FB-PA-10**: Reduce repeated dynamic-plugin localization metadata reads
- [x] **FB-PA-11**: Replace per-job-group job counts with grouped batch counting
- [x] **FB-PA-12**: Reduce repeated menu/plugin runtime metadata reads
- [x] **FB-PA-13**: Reduce repeated dynamic-plugin and release-state reads in plugin list
- [x] **FB-PA-14**: Replace per-menu role association inserts with batch insertion
- [x] **FB-PA-15**: Batch localize monitor operation-log route metadata
- [x] **FB-PA-16**: Add cluster-aware plugin runtime cache revision coordination
- [x] **FB-PA-17**: Optimize dynamic-plugin reconciler in cluster mode

## Feedback (Cross-Platform Dev Commands)

- [x] **FB-CP-1**: Allow `.github/workflows/main-ci.yml` to trigger on all branch push and pull requests
- [x] **FB-CP-2**: Fix E2E plugin menu i18n assertion missing "Configure" button
- [x] **FB-CP-3**: Fix E2E role data permission dropdown multi-tenant enabled state regression
- [x] **FB-CP-4**: Fix E2E user management sort and self-operation assertion using unstable table locator
- [x] **FB-CP-5**: Fix E2E file management data permission case misusing org plugin department scope
- [x] **FB-CP-6**: Fix E2E monitor plugin case depending on unstable list order
- [x] **FB-CP-7**: Fix E2E dynamic plugin test fixture missing current plugin manifest multi-tenant field
- [x] **FB-CP-8**: Fix E2E multi-tenant scenario assertions inconsistent with current platform tenant menu contract
- [x] **FB-CP-9**: Fix E2E data permission regression case misusing org plugin department scope
- [x] **FB-CP-10**: Fix E2E user management edit case using unstable batch button locator
- [x] **FB-CP-11**: Fix E2E plugin example data column help icon test using old locator
- [x] **FB-CP-12**: Fix E2E dict type delete case depending on transient success toast
- [x] **FB-CP-13**: README Windows `make.cmd` entry should show omitting `.cmd` suffix first
- [x] **FB-CP-14**: Multi-tenant plugin should clean deprecated tenant resolution config
- [x] **FB-CP-15**: `make dev` and `make status` should restore cross-platform status table output
- [x] **FB-CP-16**: GitHub Actions reusable workflow call with undefined `run-windows-command-smoke` input
- [x] **FB-CP-17**: `make dev` should start frontend/backend as async processes with final status table
- [x] **FB-CP-18**: GitHub Actions Go unit tests expose SQLite plugin lifecycle test cleanup missing explicit WHERE
- [x] **FB-CP-19**: GitHub Actions should clean Node.js 20 action deprecation, Go cache restore tar failure, and Windows runner redirect warning
- [x] **FB-CP-20**: SQLite SQL translator outputs redundant debug log when skipping PostgreSQL comment statements
- [x] **FB-CP-21**: `go.work` declares Go 1.26.2 causing CI Go commands to fail with Go 1.25
- [x] **FB-CP-22**: Nightly test workflow should be integrated into nightly build workflow

## Feedback (Linactl Build Tool Consolidation)

- [x] **FB-CT-1**: Merge `hack/tools/runtime-i18n` governance scanning into `linactl` internal component
- [x] **FB-CT-2**: Unify internal component source file naming to component name prefix
- [x] **FB-CT-3**: GitHub Actions `plugin-command-smoke` isolated fixture missing `linactl` local replace module

## Feedback (Dev Environment Commands)

- [x] **FB-ENV-1**: `make env.check` should use bordered table and detect PostgreSQL server version via database connection
- [x] **FB-ENV-2**: PostgreSQL version detection must use Go database connection, not `psql` client tool
- [x] **FB-ENV-3**: `TestRunEnvSetupInstallsFrontendAndPlaywright` should not fail in CI without real `pnpm`

## Feedback (Image Build Workflows)

- [x] **FB-IMG-1**: Fix `turbo.json` strict env mode not including `LINAPRO_SOURCE_PLUGINS` in `globalEnv`
- [x] **FB-IMG-2**: Fix auto-enabled tenant-scoped plugins not provisioning tenant entries for existing active tenants
- [x] **FB-IMG-3**: Fix Tailwind v4 CSS not scanning source plugin pages in production build
- [x] **FB-IMG-4**: Add memory-only Docker Compose demo launcher
- [x] **FB-IMG-5**: Add necessary comments to Docker Compose demo launcher
- [x] **FB-IMG-6**: Separate runtime configuration from Compose startup script into `config.yaml`
- [x] **FB-IMG-7**: Switch demo Docker Compose from SQLite to PostgreSQL service
- [x] **FB-IMG-8**: Adjust test Docker Compose to PostgreSQL plus manual development container

## Feedback (Release Test and Build)

- [x] **FB-REL-1**: Split `hack/tools/image-builder/main.go` into focused source files
- [x] **FB-REL-2**: Extract nightly/release E2E, Redis smoke, build smoke, and image publishing into reusable workflows and composite actions
- [x] **FB-REL-3**: Align release workflow verification matrix with nightly
- [x] **FB-REL-4**: Add bilingual comments to all workflow files, jobs, and inline steps
- [x] **FB-REL-5**: Rename `make image-build` to `make image.build`
- [x] **FB-REL-6**: Add plugin command smoke reusable workflow to main CI
- [x] **FB-REL-7**: Add make command smoke reusable workflow and extract complex main CI jobs into reusable workflows
- [x] **FB-REL-8**: Fix backend unit test workflow calling nonexistent `prepare-packed-assets` Make target
- [x] **FB-REL-9**: Add unified `make pack.assets` root entry
- [x] **FB-REL-10**: Use `make` entry instead of direct `linactl` in CI when corresponding Make target exists
- [x] **FB-REL-11**: Fix OpenAPI apidoc Chinese translation missing shared DTO keys
- [x] **FB-REL-12**: Fix SQLite smoke calling old `linactl prepare-packed-assets` command
- [x] **FB-REL-13**: Fix plugin command smoke asserting old `remote=current` key-value format
- [x] **FB-REL-14**: Fix `make dev` plugin-full mode Go workspace module duplicate error
- [x] **FB-REL-15**: Fix E2E Playwright browser revision mismatch between npm and pnpm lock files
- [x] **FB-REL-16**: Extract unified test verification suite reusable workflow
- [x] **FB-REL-17**: Make each verification suite job controlled by independent required input
- [x] **FB-REL-18**: Move shared-state-intensive E2E cases to serial isolation and default parallel workers to 1
- [x] **FB-REL-19**: Fix E2E dict type delete and user batch edit retry state loss and overlay interception
- [x] **FB-REL-20**: Release workflow should reuse shared test verification suite with brief test scope
- [x] **FB-REL-21**: Fix E2E `TC0063-auth-menu` missing explicit button permissions
- [x] **FB-REL-22**: Fix E2E login tenant-state misidentifying host platform directory as multi-tenant
- [x] **FB-REL-23**: Fix Go unit tests shared PostgreSQL schema pollution with `-p=1` serialization
- [x] **FB-REL-24**: Fix E2E `TC0223d` tenant-state simulation via real tenant menu signal
- [x] **FB-REL-25**: Split long E2E flows and reduce non-assertion navigation to prevent timeout
- [x] **FB-REL-26**: Increase Playwright global default test timeout from 60s to 180s
- [x] **FB-REL-27**: Fix `runDev` unit tests failing without real `pnpm` in CI
- [x] **FB-REL-28**: Add automatic GitHub Release creation after successful release publishing
- [x] **FB-REL-29**: Fix Make command smoke dev fixture missing `linactl` local replace dependencies
- [x] **FB-REL-30**: Fix `en-US` OpenAPI apidoc bundle containing actual translation keys
- [x] **FB-REL-31**: Add `pull_request` trigger to main CI workflow

## Feedback (Release Version Governance)

- [x] **FB-VER-1**: GitHub tag ruleset bypass description should clarify actor-level configuration
- [x] **FB-VER-2**: Controlled tag workflow should use GitHub App installation token
- [x] **FB-VER-3**: Move release tag documentation to CONTRIBUTING.md
- [x] **FB-VER-4**: Remove independent release tag section from root README

## Feedback (Monthly OpenSpec Archive)

- [x] **FB-OA-1**: Codex config template `base_url` should be injected from GitHub Secret at runtime
- [x] **FB-OA-2**: Unify `base_url` template placeholder style
- [x] **FB-OA-3**: Support switching AI Coding tool via `AI_CODING_TOOL` GitHub Variable
- [x] **FB-OA-4**: Isolate tool implementations in reusable workflows
- [x] **FB-OA-5**: Extract shared prompt files to `.github/prompts/`
- [x] **FB-OA-6**: Rename Codex config template to `config.template.toml`
- [x] **FB-OA-7**: Use PR-based write-back instead of direct push to default branch
- [x] **FB-OA-8**: Extract common runner preparation and PR finalization to composite actions
- [x] **FB-OA-9**: Switch to monthly 1st-of-month Asia/Shanghai schedule
- [x] **FB-OA-10**: Change provider base URLs and model to GitHub Variables
- [x] **FB-OA-11**: Stream AI tool execution logs to GitHub Actions step log
- [x] **FB-OA-12**: Support GitHub Copilot CLI as AI Coding tool
- [x] **FB-OA-13**: Support Copilot reasoning effort configuration
- [x] **FB-OA-14**: Fail fast after each archive phase
- [x] **FB-OA-15**: Fix `openspec validate --all` failing on existing main spec format
