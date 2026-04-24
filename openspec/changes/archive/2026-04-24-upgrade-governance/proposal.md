## Why

The previous `upgrade-governance` iteration turned framework source upgrades into an explicit development-time capability: developers run `make upgrade` to compare versions, overwrite framework code, and replay host SQL from the beginning. That solved the core framework-upgrade path, but plugin upgrade governance remained incomplete, especially for source plugins.

Two gaps remained:

1. **Source plugins had no formal upgrade path.** They are compiled into the host, but the system only exposed discovery, installation, and uninstallation semantics. When the version recorded in the database lagged behind the version declared in `plugin.yaml`, there was no explicit upgrade command.
2. **The current effective version and the source-discovered version were conflated.** The existing scan-and-sync flow could overwrite the registry version with the higher version found in source, even though governance and database state had not been upgraded yet.

Dynamic plugins already have runtime upgrade foundations through staged artifacts and active releases. This iteration does not generalize every plugin type into one unified upgrade platform and does not introduce rollback. It narrows the boundary as follows:

- **Framework upgrades** remain explicit development-time operations.
- **Source-plugin upgrades** become explicit development-time operations parallel to framework upgrades and must finish before host startup.
- **Dynamic-plugin upgrades** continue to use runtime upload plus install/reconcile semantics and are explicitly excluded from `make upgrade`.

## What Changes

- Extend `make upgrade` with upgrade scopes so it can run `scope=framework` or `scope=source-plugin`.
- Keep the repository-root upgrade tool as the single development-time entry point, but expand it to plan and execute both framework and source-plugin upgrades.
- Change the source-plugin governance model so `sys_plugin.version` and `release_id` represent only the current effective version.
- Persist newly discovered higher source-plugin versions as unreleased `sys_plugin_release` records instead of overwriting the effective version immediately.
- Fail host startup when an installed source plugin has a higher discovered version that has not yet been upgraded.
- Add an explicit source-plugin upgrade flow that performs version comparison, `phase=upgrade` migration bookkeeping, governance resource synchronization, and release/registry switching.
- Clarify in documentation and command help that dynamic-plugin upgrades still follow runtime upload plus install/reconcile and are not part of `make upgrade`.
- Keep rollback out of scope; failures only preserve failed state, logs, and manual recovery entry points.

## Capabilities

### Modified Capabilities
- `source-upgrade-governance`: expand framework source upgrade governance into a unified development-time entry point that covers both framework and source-plugin upgrades.
- `database-bootstrap-commands`: clarify confirmation semantics and SQL asset-source selection for `init` and `mock`.

### New Capabilities
- `plugin-upgrade-governance`: define source-plugin version discovery, effective-version separation, explicit development-time upgrades, and startup fail-fast checks.
- `runtime-upgrade-governance`: keep runtime business upgrade only as a direction-setting constraint for future work.

## Impact

- The repository-root `make upgrade` entry point and `hack/upgrade-source` tool now cover both framework and source-plugin upgrades.
- Plugin registry and release synchronization logic must stop overwriting the current effective version during source scanning.
- Host startup gains a preflight source-plugin upgrade check.
- Source-plugin upgrades reuse `sys_plugin_release`, `sys_plugin_migration`, and governance resource-reference tables rather than introducing a separate upgrade ledger.
- Documentation and command help must clearly separate source-plugin upgrades from dynamic-plugin runtime upgrades.
- Runtime business upgrades remain out of implementation scope for this iteration.
