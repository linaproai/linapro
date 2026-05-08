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
- [x] 3.6 Update the related documentation, including the current OpenSpec artifacts, `apps/lina-core/README.md`, `apps/lina-core/README.zh_CN.md`, command help, and plugin governance guidance.

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
- [x] 5.6 Update the repository root `README.md` and `README.zh_CN.md`, adding quick install examples for `macOS/Linux` and `Windows`, parameter descriptions, and official entry point mapping.

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

## Feedback

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
