## Why

LinaPro's developer tooling and operations layer previously lacked coherent governance across three areas: framework and source-plugin upgrades, development-time database configuration, and cross-platform onboarding.

**Upgrade governance** was incomplete. Framework upgrades had a working `make upgrade` path, but source plugins -- compiled into the host -- had no formal upgrade entry point. The source scan could overwrite the effective registry version with a higher version discovered in `plugin.yaml`, blurring the line between what was currently deployed and what merely existed in source. Dynamic plugins already had runtime upgrade foundations, but source-plugin upgrades needed explicit development-time parity with framework upgrades.

**Development database configuration** was duplicated. `apps/lina-core/hack/config.yaml` declared the same database connection settings twice -- once for `database.default.link` and again for `gfcli.gen.dao[].link` -- and local `init`/`mock` SQL execution depended on `multiStatements=true` in the MySQL DSN, tying command behavior to a driver-specific capability.

**Cross-platform onboarding** was fragmented. First-time users had to manually locate the repository, decide on a download method, choose an extraction directory, and separately verify whether their machine had the required dependencies. There was no unified, low-barrier entry point for getting the source code up and running.

## What Changes

- Extend `make upgrade` with upgrade scopes (`scope=framework` or `scope=source-plugin`) so one development-time entry point covers both framework and source-plugin upgrades.
- Separate the current effective source-plugin version from higher versions discovered in source; persist discovered versions as prepared releases instead of overwriting the effective version.
- Add a host startup fail-fast check that blocks startup when an installed source plugin has a higher discovered version that has not yet been upgraded.
- Implement explicit source-plugin upgrade flows with `phase=upgrade` migration bookkeeping, governance resource synchronization, and release/registry switching.
- Converge duplicated database connection settings in `hack/config.yaml` through YAML anchors and remove `multiStatements=true` from the development DSN.
- Rework local SQL execution to split files into individual statements and execute them sequentially, preserving ordered execution and fail-fast behavior without driver-level multi-statement support.
- Add cross-platform installation scripts (`install.sh` for macOS/Linux, `install.ps1` for Windows) under `hack/scripts/install/` that download source archives, deploy to a target directory with safe directory policies, and output environment health checks.
- Register the built-in log cleanup cron task through source code startup projection rather than SQL seed data.

## Capabilities

### Modified Capabilities
- `source-upgrade-governance`: Expand framework source upgrade governance into a unified development-time entry point covering both framework and source-plugin upgrades.
- `database-bootstrap-commands`: Update SQL asset-source selection by execution phase and rework local SQL execution to remove `multiStatements` dependency.
- `cron-job-management`: Project the built-in cleanup task into `sys_job` during startup rather than through delivery SQL seed data.

### New Capabilities
- `plugin-upgrade-governance`: Define source-plugin version discovery, effective-version separation, explicit development-time upgrades, and startup fail-fast checks.
- `framework-bootstrap-installer`: Provide cross-platform source code download, target directory deployment, safe extraction, environment health check, and post-installation guidance.
- `runtime-upgrade-governance`: Keep runtime business upgrade only as a directional constraint for future work.

## Impact

- The repository-root `make upgrade` entry point and `hack/upgrade-source` tool now cover both framework and source-plugin upgrades.
- Plugin registry and release synchronization logic no longer overwrites the current effective version during source scanning.
- Host startup gains a preflight source-plugin upgrade check that blocks startup when upgrades are pending.
- Source-plugin upgrades reuse `sys_plugin_release`, `sys_plugin_migration`, and governance resource-reference tables rather than a separate upgrade ledger.
- `apps/lina-core/hack/config.yaml` uses YAML anchors for a single shared database connection definition, and local SQL execution no longer depends on `multiStatements=true`.
- New `hack/scripts/install/` directory provides cross-platform installation scripts with archive download, safe directory policies, and environment health checks.
- Repository root `README.md` and `README.zh_CN.md` include quick install instructions for macOS/Linux and Windows.
- The built-in log cleanup cron task is registered through startup projection, removing its SQL seed dependency.
- Runtime business upgrades remain out of implementation scope for this iteration.
