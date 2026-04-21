# LinaPro Plugins

`apps/lina-plugins/` is the official source-plugin workspace for LinaPro.

At the current open-source stage, the host keeps only stable core capabilities such as user management, role management, menu management, dictionary management, parameter settings, file management, scheduled job management, plugin governance, and developer support. Non-core business modules are delivered as source plugins under `apps/lina-plugins/<plugin-id>/`.

## What Lives Here

LinaPro currently ships three plugin references in this directory:

- `plugin-demo-source`: sample source plugin structure and coding reference
- `plugin-demo-dynamic`: sample dynamic WASM plugin structure and lifecycle reference
- official source plugins: first-party business plugins compiled into the host through explicit wiring

## Official Source Plugins

The repository currently includes these first-party source plugins:

- `org-center`: department management and post management
- `content-notice`: notice management
- `monitor-online`: online user query and force logout
- `monitor-server`: server monitor collection, cleanup, and query
- `monitor-operlog`: operation log persistence and governance
- `monitor-loginlog`: login log persistence and governance

Each official plugin has its own directory and follows the same baseline structure:

```text
apps/lina-plugins/<plugin-id>/
  backend/              Plugin backend entry and hook/resource declarations
  frontend/pages/       Plugin page wrappers mounted by host menus
  manifest/sql/         Plugin-owned install and uninstall SQL assets
  plugin.yaml           Plugin manifest
  plugin_embed.go       Embedded asset registration
  README.md             English plugin guide
  README.zh_CN.md       Chinese plugin guide
```

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
4. Keep plugin-owned backend code inside the plugin directory and depend only on published host packages.
5. Register the plugin explicitly in `apps/lina-plugins/lina-plugins.go`.

## Dynamic Plugin Notes

Dynamic WASM plugins remain supported for runtime-managed delivery scenarios. Use `plugin-demo-dynamic/` as the reference when the plugin must be uploaded, installed, enabled, disabled, and uninstalled without recompiling the host.

## References

- `apps/lina-plugins/plugin-demo-source/README.md`
- `apps/lina-plugins/plugin-demo-dynamic/README.md`
- `apps/lina-plugins/OPERATIONS.md`
