## Why

Today the install and enable flow for both source plugins and dynamic plugins still mainly depends on post-startup administrative actions. That prevents some plugins from reaching a usable state automatically even though they must be ready as soon as the host starts, such as demo-control, startup-time governance extensions, or plugins that need to register capabilities before the first request arrives. We need a startup-time auto-enable mechanism that stays simple enough for the common case and is configured directly in the host service's main config file instead of introducing an overly heavy strategy model for a small set of "enable on boot" plugins.

## What Changes

- Add a simplified `plugin.autoEnable` setting to the host main config file that uses only a list of plugin IDs to declare which plugins must auto-enable during startup.
- Define "auto-enable" as "install first if needed, then enable" so the host does not need to expose extra startup-state options such as `manual`, `installed`, or `enabled`.
- Add a dedicated `plugin startup bootstrap` phase to the host HTTP startup flow: discover plugins and sync the registry first, then advance only the plugins listed in `plugin.autoEnable` through install and enable.
- Change source-plugin lifecycle semantics so discovered source plugins stay in a discovered-only state by default; they only install and enable automatically when an administrator does so explicitly or the plugin is listed in `plugin.autoEnable`.
- Reuse the existing dynamic-plugin `desired_state/current_state` and reconciler model. For dynamic plugins that declare governed host-service resources, startup auto-enable reuses an existing authorization snapshot instead of pushing complex authorization structures into the main config file.
- Make failure behavior explicit: because `plugin.autoEnable` is an explicit list of boot-time required plugins, the host fail-fast behavior applies whenever any listed plugin is missing or fails to enable.

## Capabilities

### New Capabilities
- `plugin-startup-bootstrap`: Define the host capability to install, enable, and wait for plugins to converge during startup according to static config, covering source plugins, dynamic plugins, missing-plugin handling, failure strategy, and dynamic-plugin authorization snapshots.

### Modified Capabilities
- `plugin-manifest-lifecycle`: Adjust the default lifecycle semantics after source-plugin discovery and add rules that allow startup policy to trigger install/enable transitions.

## Impact

- The affected code is mainly under `apps/lina-core/internal/cmd/cmd_http.go`, `apps/lina-core/internal/service/config/`, `apps/lina-core/internal/service/plugin/`, and its `internal/runtime` / `internal/catalog` submodules.
- The host main config template `apps/lina-core/manifest/config/config.template.yaml` needs a new plugin config shape together with matching read and validation logic.
- New startup-bootstrap unit and integration tests are required to cover source plugins, dynamic plugins, existing authorization snapshots, cluster behavior, and fail-fast cases.
- No public RESTful API changes are involved. The change focuses on host startup behavior, the config model, and plugin lifecycle convergence.
