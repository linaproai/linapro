## Context

The current host startup flow in `apps/lina-core/internal/cmd/cmd_http.go` already performs plugin discovery and synchronization, but that step only writes manifests into the registry. It does not advance plugins into the installed or enabled state. There are two important limitations today:

- Source plugins under `apps/lina-core/internal/service/plugin/plugin_lifecycle_source.go` still follow an explicit install/uninstall model; being discovered does not install them automatically.
- Dynamic plugins already have `desired_state/current_state` plus a master-node reconciler, but state only advances to `installed` or `enabled` after a management API explicitly triggers install/enable.

That means:

1. Plugins that must be ready as soon as the host starts still depend on administrators manually entering the plugin page and clicking install/enable.
2. Source plugins and dynamic plugins do not share one unified startup automation entry point, so deployments and demo environments still need manual follow-up actions.
3. Source-plugin route registration, plugin cron wiring, and dynamic frontend bundle warm-up all depend on the final enabled state. If bootstrap happens too late, the service may already be listening on its port before critical plugins are actually ready.

The main constraints for this design are:

- Startup policy must take effect before the UI is available, so the source of truth must be host static config instead of a plugin-management page or database data entered later.
- In cluster mode, shared lifecycle side effects must be executed only by the primary node to avoid duplicate plugin SQL, duplicate governed-resource writes, or duplicate dynamic release switches.
- The project is still greenfield, so lifecycle semantics and startup order can be adjusted directly without preserving old compatibility behavior.
- The implementation should reuse the existing `plugin` service, dynamic reconciler, authorization snapshot model, and enabled snapshot, instead of introducing new governance tables.

## Goals / Non-Goals

**Goals:**

- Provide a simplified `plugin.autoEnable` setting in the host main config file so operators or developers only need a plugin ID list to declare what must auto-enable during startup.
- Cover both source plugins and dynamic plugins so the repository does not end up with two unrelated startup-automation mechanisms.
- Make startup bootstrap happen before plugin routes, plugin cron wiring, and dynamic frontend bundle warm-up so the service is as close as possible to the target plugin state before it starts serving traffic.
- Reuse existing dynamic-plugin authorization snapshots so plugins that need governed host-service access can still participate in startup automation without storing complex authorization structures in the main config file.
- Keep failure handling predictable in both single-node and clustered modes: if a plugin listed in `plugin.autoEnable` does not become enabled successfully, the host fails fast.

**Non-Goals:**

- Do not auto-download or auto-build dynamic plugin artifacts. Startup only consumes source-plugin directories or already discoverable dynamic artifacts.
- Do not store startup policy in `plugin.yaml`; startup policy is an environment-level deployment decision, not a plugin-source contract.
- Do not add a visual editor in the plugin-management page for startup policy in this iteration; the scope is limited to host static config and backend startup behavior.
- Do not support multiple target states such as `manual`, `installed`, or `enabled`, and do not support per-plugin `required`, timeout, or authorization-detail settings.

## Decisions

### 1. Use `plugin.autoEnable` in the host main config file as the only entry point

Add `plugin.autoEnable` under `apps/lina-core/manifest/config/config.template.yaml` and parse it centrally in `apps/lina-core/internal/service/config/config_plugin.go`. The recommended shape is:

```yaml
plugin:
  dynamic:
    storagePath: "temp/output"
  autoEnable:
    - "demo-control"
    - "report-runtime"
```

Config semantics:

- `plugin.autoEnable` is an array of strings and each item is a plugin ID that must auto-enable during host startup.
- Auto-enable semantics are fixed as "install first if needed, then enable."
- Plugins listed in `plugin.autoEnable` are boot-time required plugins and the host must not enter its serving state before they have been enabled successfully.

Why:

- Static config is available before the host opens HTTP service and therefore satisfies the requirement that the policy takes effect immediately during startup.
- The common scenario is only "a small number of plugins should always be enabled on boot," and a plugin ID list is expressive enough without exposing a heavier DSL.
- Keeping the setting under the `plugin` node of the host main config matches operator expectations and lets it live alongside `plugin.dynamic.storagePath`.

**Alternatives considered:**
- Put `autoInstall/autoEnable` into `plugin.yaml`: this would freeze environment policy into plugin source and make it impossible to vary behavior by deployment.
- Store startup policy in `sys_config`: the host cannot rely on that safely before startup completes, and it turns a startup-time decision into post-startup data.
- Continue using a multi-layer structure such as `plugin.startup.policies[].desiredState`: more expressive, but too costly for the repository's main use case of a short boot-time plugin list.

### 2. Add a dedicated plugin startup bootstrap phase and move it ahead of plugin wiring

The startup sequence becomes:

1. Elect the cluster primary.
2. Scan and synchronize plugin manifests into the registry.
3. Execute `plugin startup bootstrap`: parse `plugin.autoEnable` and advance matching plugins through install/enable.
4. Refresh the enabled snapshot.
5. Only then wire plugin cron, register source-plugin HTTP routes, warm dynamic bundles, and start the runtime reconciler.

Implementation-wise, add a host entry such as `pluginSvc.BootstrapAutoEnable(ctx)` that is responsible for:

- reading `plugin.autoEnable` and building the target plugin ID set;
- installing/enabling source plugins synchronously;
- reusing existing authorization snapshots for dynamic plugins and triggering one targeted reconcile on the primary node;
- waiting within a fixed internal window until each target plugin reaches enabled or fails;
- refreshing the enabled snapshot once at the end so all later route and cron wiring sees the post-bootstrap state.

Why:

- Source-plugin routes and cron wiring already follow a pattern where registration happens first and enabled snapshot decides whether the feature is active, so bootstrap must finish before the snapshot is refreshed.
- Dynamic bundle warm-up and the host's startup readiness semantics also depend on the final enabled state. If bootstrap is delayed, the host can expose traffic before critical plugins are ready.

**Alternatives considered:**
- Keep relying on list-page/API actions to drive installation: that does not satisfy "ready at startup."
- Let the runtime reconciler converge in the background without blocking startup: unacceptable for demo-control, startup hooks, and other first-request-critical plugins.

### 3. `plugin.autoEnable` always means "ensure enabled" and implicitly includes install

`plugin.autoEnable` expresses exactly one thing: the listed plugins must be enabled during host startup. The execution semantics are:

- If a plugin is not installed, install it first.
- If a plugin is installed but not enabled, enable it.
- If a plugin is already enabled, keep it enabled.
- If a plugin is not listed in `plugin.autoEnable`, startup does not automatically install, enable, disable, or uninstall it.

Why:

- The core user need is a simpler auto-enable config, not a complete startup-state orchestration system.
- "Auto-enable implicitly includes install" matches operator intuition and covers both first-time source-plugin discovery and first-time dynamic plugin installation.
- Leaving non-listed plugins under manual governance avoids destructive reverse actions during restarts.

**Alternatives considered:**
- Expose intermediate target states such as `installed`: more flexible, but does not fit the simplification goal.

### 4. Use different execution paths for source and dynamic plugins while sharing one strategy model

The same `plugin.autoEnable` list applies to both plugin types, but execution is split internally:

- **Source plugins**
  - If not installed, run the existing source-plugin install orchestration.
  - After install, update the source-plugin enabled state.
  - In cluster mode, shared side effects such as SQL, menus, and governed resources are executed only by the primary node; followers only refresh their local snapshot from the primary's stable state.

- **Dynamic plugins**
  - If not installed, advance the plugin to the installed state first.
  - If still below enabled, set the target to `enabled`; the primary node triggers a targeted reconcile while followers only wait for shared-state convergence.
  - Reuse the current `desired_state/current_state/generation/release_id` model instead of introducing a second startup-specific state table.

Why:

- Source plugins and dynamic plugins already have different lifecycle engines: source plugins use local synchronous orchestration, while dynamic plugins use primary-node reconcile. Sharing a strategy model lowers operator complexity, but execution should still respect the existing engine boundaries.
- Even though source plugins are compiled into the host, their install SQL, menu writes, and resource ownership changes are still shared governance side effects and must not execute multiple times in a cluster.

**Alternatives considered:**
- Force source plugins into a fully reconcile-driven model too: unnecessary complexity with no current payoff.
- Add separate startup config for source and dynamic plugins: increases duplication and understanding cost.

### 5. Dynamic-plugin startup auto-enable reuses existing authorization snapshots only

If a dynamic plugin declares governed host services and appears in `plugin.autoEnable`, the host applies these rules:

- If the current release already has a host-approved authorization snapshot, reuse it during startup auto-enable.
- If the current release does not have an authorization snapshot yet, host startup fails and tells the administrator to complete a normal reviewed install/enable flow first.

Why:

- Startup auto-enable is only safe when the host has already decided what governed resources the dynamic plugin may access.
- Storing full authorization structures in the main config file would defeat the entire simplification goal.
- One reviewed management flow to establish the authorization baseline, followed by startup reuse, is a better balance between governance and usability.

**Alternatives considered:**
- Store full authorization details in the main config file: too heavy and too easy to drift.
- Allow enablement without an authorization snapshot: leads to uncertain runtime behavior under incomplete governance.

### 6. Add explicit labels and risk messaging for auto-enabled plugins in the management UI without hard-blocking manual actions

For plugins currently listed in `plugin.autoEnable`, the plugin-management UI must add these interaction rules:

- Show a read-only indicator in the list page and detail dialog so users know the plugin is governed by `plugin.autoEnable`.
- Keep runtime disable and uninstall actions available, but require an explicit warning before submission: the change takes effect immediately, but if the host config does not change the host will install and enable the plugin again on the next restart.
- Use a dedicated confirmation prompt for disable, and extend the existing uninstall confirmation with the same startup-auto-enable warning.
- The warning copy must make it explicit that permanent shutdown requires editing `plugin.autoEnable` in the host main config file.

Why:

- `plugin.autoEnable` means "ensure enabled at startup," not "forbid all manual governance at runtime." Administrators still need emergency stop-the-bleeding control.
- Disabling the buttons entirely or rejecting actions with errors would remove those emergency controls and incorrectly turn a startup policy into a runtime lock.
- Explicit labels plus warnings communicate the semantics before the user acts and reduce the risk of misunderstanding.

**Alternatives considered:**
- Hard-block disable/uninstall: reduces ambiguity but removes temporary governance control.
- Allow actions silently: simplest implementation, but users can easily assume the state will remain unchanged after restart.

## Risks / Trade-offs

- [Risk] In cluster mode, primary election may not have stabilized before bootstrap begins, so the host may not see shared lifecycle actions complete in time. -> Mitigation: use a fixed internal wait window; the primary node drives shared actions synchronously and timeout still fails fast.
- [Risk] `plugin.autoEnable` can be misunderstood relative to manual governance. -> Mitigation: define the semantics narrowly as "ensure enabled only" and add clear warnings in the plugin list, detail view, and action confirmations.
- [Risk] Dynamic plugins without an existing authorization snapshot will block startup. -> Mitigation: document clearly that those plugins must pass one manual reviewed enablement flow before startup automation can take over.
- [Risk] Moving bootstrap earlier extends cold-start time. -> Mitigation: only the small set of plugins listed in `plugin.autoEnable` are processed; all others stay manual.
- [Risk] Once source-plugin shared actions are primary-only in cluster mode, follower nodes may temporarily lag behind the enabled state on first startup. -> Mitigation: followers perform an extra wait/refresh step after bootstrap before route and cron registration.

## Migration Plan

1. Extend the host main config model and template with `plugin.autoEnable`, but keep the default as an empty list so existing deployments do not change behavior automatically.
2. Add an auto-enable bootstrap entry plus source/dynamic branching inside the `plugin` service, together with config validation.
3. Adjust the startup order in `cmd_http.go` so bootstrap runs before plugin routes, cron wiring, and bundle warm-up.
4. Add tests for:
   - `plugin.autoEnable` parsing;
   - source-plugin auto-install and auto-enable;
   - dynamic-plugin install/enable, existing authorization snapshots, and missing authorization snapshots;
   - shared action and wait behavior across cluster primary/follower nodes.
5. Release and rollback:
   - with the default empty list, deployments do not automatically change plugin state;
   - rolling back only requires removing `plugin.autoEnable` and reverting the code; plugin states that were already installed/enabled remain governable through the existing management APIs.

## Open Questions

- If future scenarios require more advanced startup governance, should the project add a second, higher-power config layer while keeping `plugin.autoEnable` as the common simple path?
