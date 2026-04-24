## Context

The host already has two foundational capabilities that matter directly for this requirement:

1. The host main config file already supports `plugin.autoEnable`, which can automatically install and enable specified source plugins during startup.
2. The host already exposes a global HTTP middleware registration seam to source plugins through `pluginhost.HTTPRegistrar`, and plugins such as `monitor-operlog` already use that seam.

The remaining gap is a unified write-protection switch for demo environments. Relying on administrators to avoid write operations manually cannot really protect demo data. Hard-coding the behavior inside host middleware would also blur the plugin boundary the project is trying to preserve.

There is another key constraint. Since the host already has a startup-time static enablement mechanism in `plugin.autoEnable`, there is no need to add a second boolean switch for demo behavior. Otherwise the system would need to manage both "is the plugin enabled" and "is demo mode enabled," which increases understanding and operational cost. It is cleaner to treat the presence of `demo-control` inside `plugin.autoEnable` as the single indicator that demo protection is on.

## Goals / Non-Goals

**Goals:**

- Carry the demo capability through an official source plugin named `demo-control` instead of scattering the logic through host core middleware.
- Use `plugin.autoEnable` as the only switch that controls whether `demo-control` is auto-enabled during startup.
- When `demo-control` is enabled, block non-query write operations across system APIs by primarily reusing RESTful `HTTP Method` semantics.
- Preserve the smallest necessary whitelist so the demo environment can still complete basic session actions such as login and logout.

**Non-Goals:**

- Do not add a dedicated frontend page, banner, toolbar hint, or workspace visual marker for demo mode in this change.
- Do not expand demo control into fine-grained resource whitelists, role-level exceptions, or module-specific bypasses.
- Do not design a separate middleware declaration model for dynamic plugins in this change; dynamic plugins are covered only through the existing unified `/api/v1` request path.
- Do not introduce database tables, install SQL, extra host boolean config, or new runtime-parameter schema just for this feature.

## Decisions

### 1. Use whether `plugin.autoEnable` includes `demo-control` as the only switch

The host does not add an independent boolean such as `demo.control.enabled`. It reuses `plugin.autoEnable` directly. The resulting behavior is:

- If `plugin.autoEnable` does not include `demo-control`, demo protection is off by default.
- If `plugin.autoEnable` includes `demo-control`, the host installs and enables the plugin during startup and demo protection becomes active.

Why:

- Users only need to understand one rule: when the plugin is enabled, the capability is active.
- `plugin.autoEnable` is already the host's static startup-time configuration entry and fits the deployment-time nature of demo-mode enablement.
- Disabling demo protection is as simple as removing `demo-control` from `plugin.autoEnable` and restarting the host.

**Alternatives considered:**
- Add a static boolean `demo.control.enabled`: technically workable, but it creates a second switch alongside plugin enablement and adds cognitive overhead.
- Store the toggle in `sys_config`: dynamic UI editing would be possible, but it creates a circular governance problem because the protection would be managed through the same system it protects.

### 2. Carry request interception through an official source plugin instead of a built-in host middleware

Add a source plugin under `apps/lina-plugins/demo-control/` and connect it to the full request chain through `registrar.GlobalMiddlewares().Bind("/*", ...)`.

Why:

- The requirement is a classic environment-governance capability and is a good first-party example of how source plugins can reuse host global middleware seams.
- The host core middleware chain remains generic and does not become tightly coupled to the specific semantics of "demo mode."
- Plugin lifecycle governance stays consistent, and `plugin.autoEnable` can continue to act as the environment-level switch.

**Alternatives considered:**
- Add the logic directly under `apps/lina-core/internal/service/middleware/`: simplest implementation, but it pushes an optional governance capability back into the host core.

### 3. Base write protection on `HTTP Method` and preserve the smallest whitelist

The `demo-control` plugin's global middleware applies these rules:

- When the plugin is enabled, allow `GET`, `HEAD`, and `OPTIONS` by default across `/*`.
- When the plugin is enabled, reject write-oriented methods such as `POST`, `PUT`, and `DELETE` by default.
- Preserve `POST /api/v1/auth/login` and `POST /api/v1/auth/logout` so the demo environment remains usable.
- Preserve install, uninstall, enable, and disable endpoints for plugins other than `demo-control` itself so plugin-governance behavior can still be demonstrated.

Why:

- The user explicitly suggested using RESTful `HTTP Method` semantics as the control basis, and the repository already requires query operations to use `GET`, so this strategy aligns with the project's API-governance rules.
- Method-based evaluation is low-cost and broad in coverage. The `/*` scope protects host APIs, source-plugin APIs, dynamic-plugin APIs, and any future non-`/api/v1` write entry points.
- Login and logout are session actions rather than business-data writes. Without a whitelist, the demo environment would become unusable.
- Plugin install/uninstall/enable/disable operations remain important demo-time governance actions, but `demo-control` itself must stay read-only so the protection cannot be disabled accidentally from inside the protected mode.
- The plugin only applies these rules while it is enabled. When disabled, the request chain behaves exactly as before.

**Alternatives considered:**
- Maintain an endpoint-by-endpoint URL whitelist: more precise, but far heavier to maintain and not aligned with the goal of enforcing governance through RESTful semantics.
- No whitelist at all: would block login and logout and make the demo environment unusable.

### 4. Show `plugin.autoEnable` examples in config templates but keep the default state disabled

Why:

- The original requirement was "disabled by default." Once the separate boolean switch is removed, the cleanest representation is simply to omit `demo-control` from the default `plugin.autoEnable` list.
- The config template can still include a clear example showing that demo protection becomes active as soon as `demo-control` is added.
- This preserves a safe default developer experience while reusing the existing plugin-startup governance mechanism.

**Alternatives considered:**
- Put `demo-control` into the default `plugin.autoEnable` list: that would force all default environments into read-only mode and does not match the default-disabled requirement.

## Risks / Trade-offs

- [Risk] A method-based policy means any interface that violates the repository's RESTful contract could bypass the protection or be blocked incorrectly. -> Mitigation: the project already requires all APIs to follow RESTful semantics, so the implementation intentionally enforces that rule rather than adding compatibility branches.
- [Risk] Using plugin enablement as the only switch means changing demo mode requires a config edit plus host restart. -> Mitigation: demo protection is an environment-level governance strategy, and startup-time config is clearer than adding a second boolean switch.
- [Risk] A whitelist that is too narrow or too broad can harm usability or protection strength. -> Mitigation: preserve only the minimum session-entry whitelist and enforce it through explicit tests.

## Migration Plan

1. Add the `demo-control` source plugin and wire it into the plugin workspace, `go.work`, and the shared plugin registration entry.
2. Add `plugin.autoEnable` examples to the host config template and make it clear that adding `demo-control` enables demo protection.
3. Add unit tests that verify no interception when the plugin is disabled, and verify read-only enforcement across `/*` plus login/logout and plugin-governance whitelist behavior when the plugin is enabled.
4. After release, enabling the feature in a demo environment only requires adding `demo-control` to `plugin.autoEnable` and restarting the host. Rolling back only requires removing the plugin ID and restarting.

## Open Questions

- Should a later iteration expose the current demo-mode state to the frontend workspace so it can show a read-only banner or hide write buttons? That question does not block this backend read-only protection and is left out of the current iteration.
