# LinaPro Plugins

`apps/lina-plugins/` is the official source-plugin workspace for LinaPro.

At the current open-source stage, the host keeps only stable core capabilities such as user management, role management, menu management, dictionary management, parameter settings, file management, scheduled job management, plugin governance, and developer support. Non-core business modules are delivered as source plugins under `apps/lina-plugins/<plugin-id>/`.

## What Lives Here

LinaPro currently ships these plugin references in this directory:

- `plugin-demo-source`: sample source plugin structure and coding reference
- `plugin-demo-dynamic`: sample dynamic WASM plugin structure and lifecycle reference
- official source plugins: first-party business plugins compiled into the host through explicit wiring

## Official Source Plugins

The repository currently includes these first-party source plugins:

- `demo-control`: demo-environment read-only request guard
- `org-center`: department management and post management
- `content-notice`: notice management
- `monitor-online`: online user query and force logout
- `monitor-server`: server monitor collection, cleanup, and query
- `monitor-operlog`: operation log persistence and governance
- `monitor-loginlog`: login log persistence and governance

Each official plugin has its own directory and follows the same baseline structure:

```text
apps/lina-plugins/<plugin-id>/
  backend/
    api/                Plugin API DTOs and route contracts
    internal/
      controller/       Plugin HTTP controllers
      service/          Plugin business services
      dao/              Plugin-local generated DAO objects when database access exists
      model/do/         Plugin-local generated DO objects when database access exists
      model/entity/     Plugin-local generated entity objects when database access exists
    hack/config.yaml    Plugin-local GoFrame codegen config
    plugin.go           Plugin backend registration entry
  frontend/pages/       Plugin pages mounted by host menus
  manifest/sql/         Plugin-owned install SQL assets
  manifest/sql/mock-data/ Optional plugin-owned mock/demo SQL assets
  manifest/sql/uninstall/ Plugin-owned uninstall SQL assets
  hack/tests/e2e/       Optional plugin-owned E2E TC files
  hack/tests/pages/     Optional plugin-owned E2E page objects
  hack/tests/support/   Optional plugin-owned E2E helpers
  plugin.yaml           Plugin manifest
  plugin_embed.go       Embedded asset registration
  README.md             English plugin guide
  README.zh-CN.md       Chinese plugin guide
```

`backend/internal/service/` is the only valid location for plugin service components. Do not create `backend/service/`.

## Host and Plugin Boundary

The host and source plugins are intentionally decoupled through stable seams instead of scattered `if pluginEnabled` branches.

- The host owns stable top-level menu catalogs such as `dashboard`, `iam`, `setting`, `scheduler`, `extension`, and `developer`.
- Plugin menus may mount only under published host catalog keys or inside the plugin's own menu tree.
- Official plugins have fixed mount points: `org-center -> org`, `content-notice -> content`, and all monitor plugins -> `monitor`.
- The host publishes stable capability seams for optional collaboration, such as auth events, audit events, org capability access, and plugin lifecycle hooks.
- Plugin-owned tables, menus, pages, hooks, and cron jobs live in the plugin directory and are installed or removed through the plugin lifecycle.

## Source Plugin Development Flow

1. Create `apps/lina-plugins/<plugin-id>/`.
2. Follow the structure used by `plugin-demo-source/`.
3. Declare metadata, menus, frontend pages, SQL assets, and optional hooks in `plugin.yaml`.
4. Keep plugin-owned backend code inside the plugin directory, place service logic under `backend/internal/service/`, and depend only on published host packages.
5. Register the plugin explicitly in `apps/lina-plugins/lina-plugins.go`.

## Plugin-Owned E2E Tests

Source plugins should keep plugin-specific Playwright coverage under `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`.
Plugin page objects and helpers should stay beside them in `hack/tests/pages/` and `hack/tests/support/`.
The host test runner discovers these tests through the generic `plugins` scope, and a single plugin can be run with `pnpm -C hack/tests test:module -- plugin:<plugin-id>` without adding a plugin-specific entry to the execution manifest.

## Source Plugin Version Upgrade

When a source plugin has already been installed in the host and you bump its
`plugin.yaml` version, discovery no longer replaces the effective host version
automatically.

- The current effective version stays pinned in `sys_plugin.version` and `release_id`.
- The higher discovered source version is stored as a prepared `sys_plugin_release`.
- Before the host is allowed to start, use the `lina-upgrade` AI skill, for example `upgrade source plugin <plugin-id>` or `upgrade all source plugins`.
- If you skip that step, host startup fails fast and prints the required skill invocation hint.

## Dynamic Plugin Notes

Dynamic WASM plugins remain supported for runtime-managed delivery scenarios. Use `plugin-demo-dynamic/` as the reference when the plugin must be uploaded, installed, enabled, disabled, and uninstalled without recompiling the host.

## References

- `apps/lina-plugins/plugin-demo-source/README.md`
- `apps/lina-plugins/plugin-demo-dynamic/README.md`
- `apps/lina-plugins/OPERATIONS.md`
