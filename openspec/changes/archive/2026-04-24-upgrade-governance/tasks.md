## 1. P0 Framework metadata unification

- [x] 1.1 Add a framework metadata section to `metadata.yaml` so name, version, description, homepage, repository URL, and license are managed in one place.
- [x] 1.2 Return that framework metadata from the system-info API and drive the top project card on the system-info page from backend data.

## 2. P0 Formal source upgrade command

- [x] 2.1 Add the repository-root `hack/upgrade-source` development-time tool and wire it into `make upgrade` with explicit confirmation.
- [x] 2.2 Perform backup reminders, Git dirty-worktree checks, current-version loading, and target-version comparison before upgrade execution.
- [x] 2.3 Implement target-tag fetch and local framework code overlay.
- [x] 2.4 Replay host SQL from the first file in order after the target source is applied.
- [x] 2.5 Exit safely with a clear message when the target version is not higher than the current project version.

## 3. Verification

- [x] 3.1 Add automated tests for version comparison, target-tag resolution, and Git worktree cleanliness checks.
- [x] 3.2 Add automated tests for full host-SQL replay during upgrades.

## 4. P1 reserved direction

- [x] 4.1 Keep runtime business-system upgrades as a future direction in the current change, without implementing them in this iteration.

## 5. P0 Source-plugin upgrade governance

- [x] 5.1 Extend `make upgrade` and the repository-root upgrade tool so they support `scope=framework|source-plugin`, `plugin=<id|all>`, and a shared `dry-run` plan mode.
- [x] 5.2 Adjust source-plugin scan and governance sync logic so `sys_plugin.version` and `release_id` always represent the current effective version and higher discovered versions only become prepared releases.
- [x] 5.3 Implement the explicit source-plugin upgrade flow: version comparison, single-plugin and bulk plans, `phase=upgrade` SQL execution, menu and permission synchronization, governance resource-reference synchronization, and release/registry switching.
- [x] 5.4 Add a startup-time pending-upgrade check for source plugins. If an installed plugin has a higher discovered version that has not been upgraded, block startup and print the matching `make upgrade` command.
- [x] 5.5 Clarify the dynamic-plugin upgrade boundary so runtime upload plus install/reconcile remains the only upgrade path and `make upgrade` never takes over that flow.
- [x] 5.6 Update the related documentation, including the current OpenSpec artifacts, `apps/lina-core/README.md`, `apps/lina-core/README.zh_CN.md`, command help, and plugin governance guidance.

## 6. Verification

- [x] 6.1 Add unit tests for the effective-version vs discovered-version split across uninstalled, same-version, and higher-discovered-version scenarios.
- [x] 6.2 Add tests for source-plugin upgrade commands covering single-plugin upgrades, `plugin=all`, dry-run, lower-version rejection, and not-installed plugin handling.
- [x] 6.3 Add startup fail-fast tests that confirm the host refuses to start when a source-plugin upgrade is pending.
- [x] 6.4 Add regression coverage confirming the development-time upgrade command does not interfere with dynamic-plugin runtime upgrades.

## Feedback

- [x] **FB-6**: Converge `make upgrade` implementation under `hack/upgrade-source/` and read only database connection and upgrade metadata from `apps/lina-core/hack/config.yaml`.
- [x] **FB-7**: Let `init` and `mock` switch SQL asset sources by execution phase: runtime uses embedded SQL, while development-time `Makefile` entries use local SQL files explicitly.
- [x] **FB-5**: Treat `homepage` as the official website and add a separate repository URL field for system-info presentation and upgrade tooling.
- [x] **FB-4**: Re-group `internal/cmd` unit tests by command responsibility instead of preserving file names tied to removed helpers.
- [x] **FB-3**: Keep non-test logic for `init`, `mock`, and `upgrade` close to their corresponding command files or `cmd.go` to avoid scattered helpers.
- [x] **FB-1**: Do not introduce upgrade state tables, upgrade record tables, or SQL cursor tables for source upgrades.
- [x] **FB-2**: Replay host SQL from the first SQL file during `make upgrade` instead of relying on persisted execution position.
- [x] **FB-8**: Rename the development-time tool directory from `hack/upgrade-framework` to `hack/upgrade-source` and keep only `main.go` at the component root.
- [x] **FB-9**: Extract source-plugin upgrade governance into an independent host-side component that is reused by both `make upgrade` and startup validation. Keep the `pkg` layer limited to stable contracts and small facades.
- [x] **FB-10**: Add automated validation for source-plugin upgrade governance, including unit tests for the host-side `sourceupgrade` component/facade and E2E coverage in `TC0106` for the "new version discovered but not yet activated" scenario.
- [x] **FB-11**: Fix the runtime WASM oversize-upload E2E assertion so it no longer hard-codes the outdated `16MB` limit and restore full `e2e/extension/plugin` regression coverage.
