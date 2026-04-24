## Context

`upgrade-governance` already introduced an executable framework source-upgrade path: developers run `make upgrade`, compare the current and target framework versions, overwrite code, and replay host SQL from the first file. What remained incomplete was plugin upgrade governance, especially for source plugins.

Dynamic plugins already have staged artifacts, active releases, and upgrade-phase migration behavior. Source plugins, however, were still treated as discovered/installable units without a formal upgrade entry point. Worse, the source scan could overwrite the registry version with the higher version found in `plugin.yaml`, which blurred the line between the version currently in effect and the version merely discovered in source.

Because source plugins are compiled and shipped together with the host, their upgrade semantics are closer to framework upgrades than to dynamic-plugin runtime hot upgrades. This iteration therefore keeps source-plugin upgrades strictly as development-time operations and uses host startup only for fail-fast validation, not for automatic repair.

Rollback remains intentionally out of scope. The design focuses on explicit plans, visible execution, reusable governance ledgers, and safe startup blocking.

## Goals / Non-Goals

**Goals**
- Extend `make upgrade` so one development-time entry point supports framework and source-plugin upgrades.
- Separate current effective source-plugin versions from higher versions discovered in source.
- Reuse existing release, migration, and governance-resource tables for source-plugin upgrades.
- Block host startup when installed source plugins still require an explicit upgrade.
- Clarify that dynamic-plugin upgrades remain runtime upload plus install/reconcile flows.
- Update the relevant OpenSpec artifacts and delivery documentation.

**Non-Goals**
- Do not implement rollback commands, automatic rollback, or rollback SQL directories.
- Do not build a runtime business-system upgrade platform in this iteration.
- Do not add a dedicated `manifest/sql/upgrade/` directory for plugins.
- Do not replace the existing dynamic-plugin runtime reconciler model with a development-time command.

## Decisions

### 1. Keep `make upgrade` as the only development-time upgrade entry point

`make upgrade` remains the single top-level development-time upgrade command, but it now accepts explicit scope parameters. The repository-root tool lives under `hack/upgrade-source`, with `main.go` kept at the root and the real implementation split into focused internal components for framework and source-plugin upgrades.

### 2. Source-plugin upgrades must be explicit and pre-startup

Source plugins are compiled into the host, so they must be upgraded before the host starts. Startup may scan and validate, but it must not automatically execute source-plugin migrations.

### 3. Source plugins must separate effective version from discovered version

`sys_plugin.version` and `sys_plugin.release_id` represent only the effective source-plugin version. Higher versions discovered in source are written as prepared releases and do not take effect until an explicit upgrade completes.

### 4. Reuse release, migration, and resource-reference ledgers

Source-plugin upgrades reuse `sys_plugin_release`, `sys_plugin_migration`, and `sys_plugin_resource_ref` rather than introducing a separate upgrade metadata stack.

### 5. Reuse existing install SQL assets for upgrade execution

The iteration does not introduce `manifest/sql/upgrade/`. Source-plugin upgrades reuse the existing plugin SQL assets and record the execution under `phase=upgrade`, relying on idempotent SQL rules already required by the project.

### 6. Host startup must fail fast when a source-plugin upgrade is pending

After source scanning, startup compares the effective version with the highest discovered source version. If an installed source plugin is behind, startup fails before routes, cron jobs, or other plugin runtime hooks become active.

### 7. Dynamic-plugin upgrades keep their runtime model

Dynamic plugins continue to upgrade through upload plus install/reconcile. `make upgrade` must never scan, switch, or migrate dynamic-plugin releases.

### 8. Rollback stays out of scope

Failures stop execution immediately, preserve failure records and logs, and require manual repair. The iteration does not attempt automated recovery.

## Risks / Trade-offs

- Historical source-plugin releases still reference the same evolving source tree rather than frozen artifacts. The iteration accepts this limitation to deliver a clear upgrade path first.
- Startup fail-fast adds friction for local development, but it is safer than running a host whose compiled source plugins are newer than its governance state.
- Without rollback, recovery is more manual. That is an intentional boundary for this iteration.
- Source and dynamic plugins keep different upgrade triggers. That difference reflects different delivery models rather than a design inconsistency.

## Migration Plan

1. Extend `make upgrade` and `hack/upgrade-source` to support `scope=framework|source-plugin` and `plugin=<id|all>`.
2. Split source-plugin scan results into effective-version state and discovered higher-version releases.
3. Implement source-plugin upgrade planning and execution with `phase=upgrade` bookkeeping.
4. Add host-startup fail-fast validation for pending source-plugin upgrades.
5. Add regression coverage for version drift, dry-run planning, single-plugin upgrades, bulk upgrades, and startup blocking.
6. Update documentation and command help so the development-time and runtime upgrade boundaries remain clear.

## Open Questions

- None.
