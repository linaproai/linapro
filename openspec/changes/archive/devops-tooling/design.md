## Context

This consolidation merges archived changes addressing developer tooling and operational infrastructure for LinaPro. The changes share a common theme: improving the developer experience around tooling, configuration, release governance, and operational workflows. The merged design organizes them by functional area.

## Goals / Non-Goals

**Goals:**
- Provide a unified cross-platform development command entry point through `linactl`.
- Consolidate independent build tools into `linactl/internal/` subcomponents.
- Separate environment management from service startup with `env.check` and `env.setup`.
- Automate monthly OpenSpec archive governance with configurable AI Coding tools.
- Enforce release tag version consistency and provide controlled release tag creation.
- Restructure release workflow to reuse shared test verification suite.
- Provide a manual nightly image build entry and memory-only Docker Compose launcher.
- Provide a single `make upgrade` entry point for both framework and source-plugin upgrades.
- Remove duplicated database connection settings and `multiStatements` dependency.
- Establish cross-platform installation scripts for new developers.
- Project the built-in log cleanup task through startup code.
- Provide a manual-trigger-only skill for backend API performance auditing.

**Non-Goals:**
- Do not rewrite GoFrame CLI, Docker, kubectl, pnpm, Playwright, or Go toolchain.
- Do not require Windows users to install GNU Make, Git Bash, MSYS2, or Cygwin.
- Do not change runtime API, database schema, permissions, plugin runtime contracts, or frontend UI.
- Do not implement rollback commands or automatic rollback.
- Do not build a runtime business-system upgrade platform.
- Do not automatically install system dependencies during bootstrap installation.
- Do not automatically fix issues discovered by the performance audit.
- Do not implement real-time monitoring or APM.

## Cross-Platform Dev Commands

### Decision 1: Use Go CLI as the cross-platform primary entry point

New or extended `hack/tools/linactl` provides unified subcommands (`dev`, `stop`, `status`, `build`, `wasm`, `init`, `mock`, `test`, `test-go`, `help`, `env.check`, `env.setup`, `image`, `image.build`, `i18n.check`, `release.tag.check`, `pack.assets`, `plugins.*`, etc.). The Go CLI handles cross-platform path processing, file copying, process startup, HTTP readiness, port detection, log files, subcommand execution, and error output.

Go was chosen because the project backend and existing toolchain are already Go-based, Go's standard library provides cross-platform file system, process, HTTP, path, and environment variable handling, Windows users already need Go for backend development, and existing `hack/tools/*` already follow the Go tool pattern.

### Decision 2: `make.cmd` serves only as a Windows thin wrapper

The repository root provides `make.cmd` that forwards parameters to the Go CLI. It does not carry business logic, duplicate complex logic, or maintain a separate task implementation. `cmd.exe` users execute `make dev`; PowerShell users use `.\make dev` or `.\make.cmd dev`. All parameters are forwarded transparently.

### Decision 3: `Makefile` remains as a compatibility layer

The existing `make <target>` entry is not removed. Target implementations are progressively thinned to call the Go CLI. Migration order: low-risk targets first (`help`, `status`, `pack.assets`, `wasm`), then development service targets (`dev`, `stop`), then build targets (`build`), then verification targets (`test-go`, `test-scripts`), then sub-module targets.

### Decision 4: Support make-style parameters

The Go CLI supports existing make-style `key=value` parameters (e.g., `init confirm=init rebuild=true`, `build platforms=linux/amd64,linux/arm64 verbose=1`, `wasm p=plugin-demo-dynamic`). Internally, `key=value` pairs are normalized to option structures.

### Decision 5: Test strategy centered on tool behavior

This change belongs to the development toolchain and does not involve user-observable pages, so E2E tests are not needed. Go unit tests and command-level smoke tests cover parameter parsing, file copying, plugin scanning, help output, and Makefile thin-wrapper consistency.

### Decision 6: GitHub Actions must cover Windows basic commands

`.github/workflows/` related workflows add `windows-latest` runner verification covering at least `go run ./hack/tools/linactl help`, `go run ./hack/tools/linactl status`, and one lightweight file or plugin tool command. Windows CI verification covers both `cmd.exe` (`make <target>`) and PowerShell (`.\make <target>`) usage. Heavy builds, database initialization, Docker images, and dev services use lightweight smoke or dry-run modes.

## Linactl Build Tool Consolidation

### Decision 7: Use `linactl/internal` subcomponents for tool implementations

Image building implementation goes into `hack/tools/linactl/internal/imagebuilder`, dynamic plugin packaging into `hack/tools/linactl/internal/wasmbuilder`, and runtime i18n governance scanning into `hack/tools/linactl/internal/runtimei18n`. Command files only orchestrate parameters, plugin workspace preparation, and output.

### Decision 8: Allow `linactl` compile-time dependency on `lina-core/pkg/pluginbridge`

`wasmbuilder` needs `pluginbridge` artifact section, lifecycle, route, and host service contract definitions. Keeping this as a compile-time dependency eliminates subprocess tool boundaries. The `linactl/go.mod` explicitly declares the `lina-core` local dependency for `GOWORK=off` compilation.

### Decision 9: Keep public commands stable and delete old independent entries

User-facing entries remain `make image`, `make image.build`, `make wasm`, `make i18n.check` and their `linactl` equivalents. Old tool directories are deleted; repository tests, CI, and documentation switch to public entries or internal package tests.

## Dev Environment Commands

### Decision 10: `env.check` performs tool-level smoke detection only

`env.check` checks for Go, Node.js, pnpm, Vite, Playwright, and PostgreSQL using `exec.LookPath` and version commands. It outputs a bordered ASCII table with name, current version, required version, satisfied status, and remarks. It does not start dev services, connect to business databases, or modify local dependencies.

### Decision 11: PostgreSQL detection uses Go database connection

PostgreSQL version detection reads `apps/lina-core/manifest/config/config.yaml` database connection settings and queries `SHOW server_version` via Go `database/sql`, rather than depending on the `psql` client tool.

### Decision 12: Version requirements are centrally defined in `linactl`

Each tool's minimum or project-required version (e.g., Go 1.25+, PostgreSQL 14+) is defined within the `env.check` command file. Adjusting requirements does not require Makefile changes.

### Decision 13: `env.setup` reuses the original `dev.setup` implementation path

The command file is renamed to `command_env.setup.go` with function `runEnvSetup`, continuing to call `frontend.EnsureDeps` and the Playwright installation command. Behavior migration maintains compatibility.

## Image Build Workflows

### Decision 14: Release workflow reuses shared test verification suite

The release workflow calls `.github/workflows/reusable-test-verification-suite.yml` with Main CI's brief test scope (no E2E). The `release-image` job depends on tag version check and `verification-suite` via `needs`. Any test failure, cancellation, or timeout prevents GHCR login, image push, and `latest` update.

### Decision 15: Release workflow creates GitHub Release after successful publishing

After tag validation, shared test suite, and GHCR image publishing all succeed, the release workflow creates a GitHub Release with title `LinaPro Release <tag>`. If any gate fails, no Release is created.

### Decision 16: Manual nightly image build bypasses test gates

A separate `workflow_dispatch`-only workflow directly calls the reusable image publish workflow without test dependencies. It uses the same nightly tag strategy, multi-architecture platforms, official plugin full build mode, and GHCR publish permissions. The scheduled nightly workflow continues to require test gates.

### Decision 17: Nightly demo image provides memory-only Docker Compose launcher

`hack/deploy/docker-compose.yaml` provides a demo launcher using PostgreSQL service, memory-only `tmpfs` data directories, runtime configuration from `hack/deploy/config.yaml` injected as read-only config, and startup sequence: PostgreSQL health check, `init --rebuild=true`, `mock`, HTTP service. The `hack/deploy/tests/docker-compose.yaml` provides a manual development container based on `loads/ubuntu:24.04-npm`.

### Decision 18: Multi-architecture Docker image building uses per-platform host binaries

Multi-platform image builds prepare host binaries for each target platform. Docker buildx pushes multi-architecture manifests. Without `push=1`, multi-platform builds fail fast.

## OpenSpec Archive Automation

### Decision 19: Independent monthly workflow with configurable AI Coding tool

A `Monthly OpenSpec Archive` workflow uses `schedule` and `workflow_dispatch` triggers. The schedule uses UTC month-end cron groups covering Asia/Shanghai 1st-of-month 00:00, with leap-year deduplication for February. The `AI_CODING_TOOL` GitHub Variable selects the tool (`codex`, `cc`, or `copilot`; default `codex`).

### Decision 20: Tool-specific reusable workflows with shared composite actions

Codex, Claude Code, and GitHub Copilot CLI implementations are isolated in separate reusable workflows. Common governance steps (setup, change detection, archive completion assertion, validation, PR finalization) are extracted to local composite actions under `.github/actions/monthly-openspec-*`.

### Decision 21: Shared prompt files across AI tools

Auto-archive and archive consolidation prompts are maintained in `.github/prompts/` and referenced by all tool-specific workflows via stdin.

### Decision 22: Phase-based fail-fast with OpenSpec validation

After auto-archive, `openspec list --json` verifies no completed changes remain. After archive consolidation, `openspec validate --all` runs. Any phase failure stops subsequent phases and prevents PR creation. Change scope protection only allows `openspec/**` modifications.

### Decision 23: PR-based write-back instead of direct push

Archive results are written to a fixed maintenance branch `automation/monthly-openspec-archive` and a PR is created or updated targeting the default branch. No direct push to the default branch.

### Decision 24: Runtime credential injection

AI tool credentials and provider `base_url` are injected at runtime from GitHub Secrets and Variables. Real API keys and endpoints never enter the repository workspace, artifacts, or commit history.

### Decision 25: AI tool execution logs stream to GitHub Actions

`codex exec`, `claude -p`, and `copilot -p` stdout/stderr are streamed to the current Actions step log via `tee`, while artifact logs are preserved for post-hoc review. Log passthrough must not mask AI tool process failure exit codes.

## Release Version Governance

### Decision 26: `linactl release.tag.check` as the single version consistency entry point

The command reads `apps/lina-core/manifest/config/metadata.yaml` `framework.version` and compares it with the provided tag. Validation includes: version non-empty, tag non-empty, exact match, release format compliance, and Docker tag compatibility.

### Decision 27: Release tag format limited to Docker-compatible SemVer subset

Allowed: `vMAJOR.MINOR.PATCH` and `vMAJOR.MINOR.PATCH-prerelease` (e.g., `v0.2.0`, `v0.2.1-rc.1`). Disallowed: SemVer build metadata `+build` because Docker tags do not support `+`.

### Decision 28: Existing tag push release workflow gains a leading version check job

All test and image publishing jobs depend on the `release-tag-version-check` job via `needs`. Even if someone bypasses the recommended flow and pushes a tag directly, incorrect tags will not produce release artifacts.

### Decision 29: Controlled release tag creation via GitHub App

A `Create Release Tag` manual workflow reads `framework.version`, runs the same `linactl release.tag.check` validation, and creates and pushes a matching tag using a GitHub App installation token generated from `RELEASE_APP_CLIENT_ID` and `RELEASE_APP_PRIVATE_KEY`. The tag ruleset bypass must be configured to the GitHub App actor, not to a token string.

## Upgrade Governance

### Decision 30: Keep `make upgrade` as the only development-time upgrade entry point

`make upgrade` accepts explicit scope parameters (`scope=framework` or `scope=source-plugin`). The implementation lives under `hack/upgrade-source/` with `main.go` at the root and internal components for framework and source-plugin upgrades.

### Decision 31: Source plugins must separate effective version from discovered version

`sys_plugin.version` and `sys_plugin.release_id` represent only the effective source-plugin version. Higher versions discovered in source are written as prepared releases and do not take effect until an explicit upgrade completes.

### Decision 32: Host startup must fail fast when a source-plugin upgrade is pending

After source scanning, startup compares the effective version with the highest discovered source version. If an installed source plugin is behind, startup fails with the plugin ID, effective version, discovered version, and recommended `make upgrade` command.

### Decision 33: Reuse release, migration, and resource-reference ledgers

Source-plugin upgrades reuse `sys_plugin_release`, `sys_plugin_migration`, and `sys_plugin_resource_ref` rather than introducing a separate upgrade metadata stack. Upgrade records entries with `phase=upgrade` and synchronizes menus, permissions, and governance resource references.

### Decision 34: Framework upgrades replay all host SQL from the first file

After the target source code is applied, the framework upgrade replays every host SQL file from the first file in sorted order. Execution stops immediately on the first SQL failure.

## Development Database Configuration

### Decision 35: Reuse one development database connection through YAML anchors

Define one shared database connection anchor in `apps/lina-core/hack/config.yaml` and let both `database.default.link` and `gfcli.gen.dao[].link` reference it. Remove `multiStatements=true` from the shared DSN.

### Decision 36: Split SQL statements explicitly in the command layer

Add SQL splitting helpers under `apps/lina-core/internal/cmd/` that turn each SQL file into an ordered statement list executed one statement at a time. The splitter ignores blank fragments and handles common comments and semicolons inside string literals. `executeSQLAssetsWithExecutor` keeps its fail-fast behavior with statement-level granularity.

## Framework Bootstrap Installer

### Decision 37: Dual entry point scripts

`install.sh` for macOS/Linux and `install.ps1` for Windows PowerShell under `hack/scripts/install/`. Both share consistent core parameter semantics. Archive download (tar.gz for Unix, zip for Windows) is preferred over `git clone`.

### Decision 38: Target directory uses explicit mode selection with safe defaults

Default: new subdirectory under current working directory. Current-directory mode: extract directly. Specified directory: deploy to target. Non-empty target directory without explicit overwrite parameter: refuse to continue. Extraction uses temporary directory first, then moves to final position.

### Decision 39: Environment health check only, no automatic dependency installation

After deployment, check for Go, Node.js, pnpm, MySQL, and make presence/version. Output post-installation guidance with project path and recommended next commands. Do not call package managers automatically.

## Cron Job Management

### Decision 40: Project built-in cleanup task through startup code

The `host:cleanup-job-logs` task is registered through host source code and projected into `sys_job` during service startup. Delivery SQL does not write initialization seed data. Default `cron_expr` triggers daily at midnight with `is_builtin=1`.

## Performance Audit Skill

### Decision 41: Implement the audit as an agent skill

The workflow lives at `.agents/skills/lina-perf-audit/` with `SKILL.md`, references, and bundled scripts. Manual trigger only; no invocation from CI, scheduled jobs, git hooks, or other skills.

### Decision 42: Three-stage audit workflow

Stage 0 preparation (environment setup, plugin installation, endpoint scanning, stress fixtures), Stage 1 concurrent sub-agent audit, Stage 2 summary and persistent issue-card aggregation. Per-run artifacts under `temp/lina-perf-audit/<run-id>/`.

### Decision 43: Trace-ID based SQL evidence with destructive endpoint handling

Sub agents use GoFrame's default `Trace-ID` response header to correlate endpoint calls with SQL log lines. Destructive endpoints use autonomous create-call-delete fixtures. Read/query endpoints that execute unexpected write SQL are reported as HIGH severity, with operational writes to `sys_online_session` or `plugin_monitor_operlog` treated as expected.

### Decision 44: Persistent cross-run issue cards with fingerprint de-duplication

Each finding writes a persistent markdown card under repository-root `perf-issues/`. Cards are de-duplicated by fingerprint. Repeated findings update `last_seen_run` and `seen_count`. Previously fixed cards are reopened when issues recur.

## Risks / Trade-offs

- Go CLI compilation adds startup latency to each command invocation. Accepted for cross-platform consistency; cached binary entry can be added later.
- `make.cmd` may conflict with real GNU Make on systems that have both. Mitigated by Windows-only local script usage.
- Build tool consolidation makes `linactl` heavier. Accepted because tools are LinaPro-repository-specific.
- Monthly auto-archive may leave recently completed changes in the active directory for up to a month. Mitigated by manual trigger entry.
- UTC cron cannot directly express Beijing-time month-start. Mitigated by grouped cron with leap-year deduplication.
- Manual nightly bypass may publish images from unverified commits. Documented as maintenance re-publish entry.
- Without upgrade rollback, recovery is manual. Intentional boundary for this iteration.
- Custom SQL splitter may miss edge cases. Mitigated by targeted tests against current SQL style.
- Destructive local setup from audit skill requires explicit confirmation for ambiguous requests.
- Historical source-plugin releases reference evolving source trees rather than frozen artifacts. Accepted to deliver clear upgrade path first.
