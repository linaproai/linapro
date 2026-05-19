## Why

LinaPro's developer tooling and operations infrastructure lacked coherence across several dimensions: cross-platform compatibility, tool consolidation, environment management, automated governance, release quality gates, and version integrity.

**Cross-platform compatibility** was absent. The repository's development commands relied on GNU Make, POSIX Shell, and Linux/macOS-specific tools (`lsof`, `awk`, `sed`, `nohup`, `kill`). Windows users could not execute common development tasks without installing GNU Make, Git Bash, or MSYS2. There was no unified, low-barrier entry point for Windows, macOS, and Linux developers.

**Build tool fragmentation** accumulated over time. `hack/tools/image-builder`, `hack/tools/build-wasm`, and `hack/tools/runtime-i18n` existed as independent Go modules despite primarily being invoked through `linactl`. This created duplicate `go.work` entries, redundant CI fixtures, and scattered documentation paths.

**Developer environment management** was conflated with service startup. The `make dev.setup` entry mixed dependency installation with the `dev` command's service-startup semantics, and there was no lightweight cross-platform health check to verify tool requirements before development.

**OpenSpec archive governance** depended entirely on manual triggering. Completed changes lingered in the active directory, increasing noise in the active change list and causing feedback flows to misidentify completed work as pending.

**Release quality gates** were inconsistent. The release workflow did not reuse the shared test verification suite used by nightly and main CI, and there was no version governance to ensure release tags matched framework metadata.

**Upgrade governance** was incomplete. Framework upgrades worked, but source plugins lacked a formal upgrade entry point, and the source scan could overwrite effective versions with discovered versions.

**Database configuration** was duplicated, cross-platform onboarding was fragmented, and API performance auditing was ad hoc.

## What Changes

- Provide a cross-platform Go CLI (`hack/tools/linactl`) as the primary development command entry point, with a Windows `make.cmd` thin wrapper for `cmd.exe` and PowerShell compatibility.
- Consolidate `hack/tools/image-builder`, `hack/tools/build-wasm`, and `hack/tools/runtime-i18n` into `linactl/internal/` subcomponents, removing independent tool modules.
- Add `env.check` for lightweight tool-level environment health checks and `env.setup` for frontend dependency and Playwright browser installation, replacing the old `dev.setup` entry.
- Automate monthly OpenSpec archive governance through GitHub Actions, supporting configurable AI Coding tools (Codex, Claude Code, GitHub Copilot CLI) with PR-based write-back.
- Restructure release workflow to reuse the shared test verification suite with Main CI's brief test scope, add release tag version governance with `linactl release.tag.check`, and provide a controlled release tag creation workflow.
- Provide a manual nightly image build workflow that bypasses test gates, plus a memory-only Docker Compose demo launcher.
- Extend `make upgrade` with upgrade scopes for both framework and source-plugin upgrades, with effective-version separation and startup fail-fast checks.
- Converge duplicated database connection settings through YAML anchors and rework local SQL execution to remove `multiStatements` dependency.
- Add cross-platform installation scripts under `hack/scripts/install/` for new developer onboarding.
- Register the built-in log cleanup cron task through source code startup projection.
- Add the `lina-perf-audit` agent skill for automated backend API performance auditing.

## Capabilities

### Modified Capabilities
- `upgrade-governance`: Expand framework source upgrade governance into a unified development-time entry point covering both framework and source-plugin upgrades.
- `plugin-upgrade-governance`: Define source-plugin version discovery, effective-version separation, explicit development-time upgrades, and startup fail-fast checks.
- `database-bootstrap-commands`: Update SQL asset-source selection by execution phase and rework local SQL execution to remove `multiStatements` dependency.
- `cron-job-management`: Project the built-in cleanup task into `sys_job` during startup rather than through delivery SQL seed data.
- `runtime-upgrade-governance`: Keep runtime business upgrade only as a directional constraint for future work.
- `project-setup`: Adjust development environment commands, add environment check and initialization entries, and remove the old `dev.setup` entry.
- `release-image-build`: Restructure release workflow to reuse shared test verification suite, add version governance, provide manual nightly entry, and add controlled release tag creation.
- `e2e-suite-organization`: Complete E2E covers host and official plugin tests in nightly; release uses brief test scope without E2E.
- `spec-governance`: Supplement OpenSpec archive governance with controlled monthly automation.

### New Capabilities
- `cross-platform-dev-commands`: Define the project's cross-platform development command entry, Windows `make.cmd` compatibility, make-style parameter compatibility, external tool invocation boundaries, testing and documentation requirements.
- `linactl-build-tool-consolidation`: Define `linactl` as the unified carrier for image building, dynamic plugin Wasm packaging, and runtime i18n governance scanning.
- `framework-bootstrap-installer`: Provide cross-platform source code download, target directory deployment, safe extraction, environment health check, and post-installation guidance.
- `lina-perf-audit-skill`: Define the public contract for LinaPro's backend API performance and read-request side-effect audit skill.
- `monthly-openspec-archive`: Define monthly OpenSpec automatic archiving, consolidation, validation, PR write-back, AI Coding tool selection, and credential injection.
- `release-version-governance`: Define release tag version consistency enforcement, cross-platform validation tooling, and controlled release tag creation workflow.

## Impact

- The repository-root development commands are unified through `hack/tools/linactl`, with `make`, `make.cmd`, and direct `linactl` invocation as interchangeable entry points.
- `hack/tools/image-builder`, `hack/tools/build-wasm`, and `hack/tools/runtime-i18n` are removed as independent modules; their implementations live under `linactl/internal/`.
- New `env.check` and `env.setup` commands replace `dev.setup` for environment management.
- Monthly OpenSpec archive automation runs on a configurable schedule with PR-based write-back, supporting Codex, Claude Code, and GitHub Copilot CLI.
- Release workflow reuses the shared test verification suite, enforces tag-version consistency, and creates GitHub Releases after successful publishing.
- Manual nightly image build bypasses test gates for maintenance re-publishing.
- Docker Compose demo launcher provides a memory-only experience environment.
- Plugin registry and release synchronization no longer overwrites the current effective version during source scanning.
- Host startup gains a preflight source-plugin upgrade check that blocks startup when upgrades are pending.
- `apps/lina-core/hack/config.yaml` uses YAML anchors for database connection deduplication.
- New `hack/scripts/install/` provides cross-platform installation scripts.
- The built-in log cleanup cron task is registered through startup projection.
- Adds `.agents/skills/lina-perf-audit/` with automated backend API performance auditing.
