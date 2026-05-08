## Context

This consolidation merges three archived changes that all address developer tooling and operational infrastructure for LinaPro:

1. **Upgrade governance** -- establishing a unified development-time upgrade entry point for both framework and source-plugin upgrades, with effective-version separation and startup fail-fast validation.
2. **Development database configuration deduplication** -- removing duplicated connection settings and `multiStatements` dependency from the host development toolchain.
3. **Framework bootstrap installer** -- providing cross-platform installation scripts for quick source code deployment with environment health checks.

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

**Non-Goals:**
- Do not implement rollback commands, automatic rollback, or rollback SQL directories.
- Do not build a runtime business-system upgrade platform in this iteration.
- Do not change runtime config structures or the delivery strategy under `manifest/config/`.
- Do not automatically install system dependencies during the bootstrap installation flow.
- Do not introduce a `git clone`-based primary installation path.

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

## Risks / Trade-offs

- Historical source-plugin releases still reference the same evolving source tree rather than frozen artifacts. The iteration accepts this limitation to deliver a clear upgrade path first.
- Startup fail-fast adds friction for local development, but it is safer than running a host whose compiled source plugins are newer than its governance state.
- Without rollback, recovery is more manual. That is an intentional boundary for this iteration.
- Source and dynamic plugins keep different upgrade triggers reflecting different delivery models.
- A custom SQL splitter may miss edge cases. Mitigation: targeted tests against current delivered SQL style, prioritizing common boundaries such as strings, line comments, and block comments.
- When dependencies are not automatically installed, some users still need to manually supplement their environment. Mitigation: clear missing-item output and next-step commands.
- Current directory mode can easily damage existing files. Mitigation: explicitly triggered parameter with non-empty directory refusal.
