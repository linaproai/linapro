## Why

LinaPro's host service has accumulated multiple backend capabilities -- permissions, organization, content, monitoring, tasks, and development tools -- that blur the line between framework core and optional business modules. This weakens the project's positioning as an "AI-driven full-stack development framework": developers can start quickly, but cannot easily distinguish which capabilities are essential and which can be cut, replaced, or evolved independently.

Beyond boundary slimming, the project also needs environment-level governance. Demo environments require a data-protection layer that preserves full browse/query behavior while preventing accidental or malicious writes. This cannot be solved by administrator discipline alone; it needs a systematic, plugin-driven guard.

This change addresses both concerns: converging the host into its framework core and management base while delivering non-core modules as source plugins, and adding a demo-control source plugin that enforces read-only protection through the same plugin lifecycle and global middleware seams.

## What Changes

- Reconstruct the default management backend first-level menu structure to form 9 stable mount points provided by the host: `dashboard`, `iam`, `org`, `setting`, `content`, `monitor`, `scheduler`, `extension`, `developer`.
- Establish that these first-level directories are explicitly created and owned by the host; plugins can only mount menus to these directories and cannot create new first-level directory systems.
- Clarify host boundaries: user management, role management, menu management, plugin management, task scheduling, authentication session kernel, configuration, and dictionary remain in the host.
- Deliver non-core management modules as official source plugins:
  - `org-center` -- department management and position management.
  - `content-notice` -- notification announcements.
  - `monitor-online` -- online user query and forced offline management.
  - `monitor-server` -- service monitoring collection, cleaning, and display.
  - `monitor-operlog` -- operation log query, details, export, cleaning, and page.
  - `monitor-loginlog` -- login log query, details, export, cleaning, and page.
  - `demo-control` -- demo-mode write-protection guard.
- Introduce stable capability seams (capability interfaces, event hooks, HTTP route/middleware registrars, Cron registers) instead of scattering `if pluginEnabled` branches throughout host code.
- Define plugin-local ORM generation, plugin-scoped table naming (`plugin_<plugin_id_snake_case>`), and plugin-owned storage lifecycle boundaries.
- Add a `demo-control` source plugin that blocks write operations across `/*` based on HTTP Method semantics when enabled via `plugin.autoEnable`, preserving a minimal session and plugin-governance whitelist.

## Capabilities

### New Capabilities

- `core-host-boundary-governance`: Defines that the open source stage host only retains the boundary constraints between the framework core and the governance base.
- `module-decoupling`: Defines independent start, stop, and host downgrade rules when non-core management modules are delivered as source plugins.
- `menu-management`: Introduces a stable first-level directory structure, plugin menu semantic mounting, and empty directory hiding rules for the default backend.
- `plugin-manifest-lifecycle`: Supplements fixed mount points, domain-capability plugin IDs, and host directory parent key constraints for official source plugins.
- `demo-control-guard`: Defines a demo-control source plugin governed by `plugin.autoEnable`, together with global write-interception rules based on HTTP Method semantics.

### Modified Capabilities

- `dept-management`: Department management capabilities are delivered by the `org-center` source plugin.
- `post-management`: Position management capabilities are delivered by the `org-center` source plugin.
- `notice-management`: Notification announcement capabilities are delivered by the `content-notice` source plugin, mounted under `content management`.
- `online-user`: Online user capabilities are delivered by the `monitor-online` source plugin; the host only retains the authentication session kernel.
- `server-monitor`: Service monitoring capabilities are delivered by the `monitor-server` source plugin.
- `oper-log`: Operation log capabilities are delivered by the `monitor-operlog` source plugin; the host emits unified audit events instead.
- `login-log`: Login log capabilities are delivered by the `monitor-loginlog` source plugin; the host emits unified login events instead.
- `user-management`: User management hides department/position-related filters, fields, and form items when the organization plugin is missing.
- `user-auth`: The authentication link publishes login lifecycle events without directly depending on specific login log persistence implementations.
- `config-management`: Log TraceID output is controlled only by static configuration, not runtime parameters.
- `plugin-hook-slot-extension`: Dynamic plugins reuse public bridge components for ABI, codec, and typed guest controller adaptation.
- `plugin-http-slot-extension`: The host publishes backend extension points through callback registration (HTTP routes, Cron, menu/permission filters).

## Impact

### Backend Impact

- Adjust host menu initialization SQL and menu projection logic to provide a stable first-level directory, accurate parent `menu_key`, and a hideable host directory skeleton.
- Extract host dependencies of organization, notification, and monitoring modules to prepare for source plugin migration.
- Supplement host kernel/plugin display management boundaries for login, audit, online session, and service monitoring links.
- Add official source plugin directories and explicit wiring entries.
- Tighten plugin storage boundaries: the host no longer holds `dao/do/entity`, mock data, or direct table lookup logic for plugin business tables.
- Add `demo-control` source plugin with global HTTP middleware integration for write protection.
- Publish unified event contracts for login lifecycle and audit events, removing direct host dependencies on specific log implementations.

### Frontend Impact

- Adjust the left menu hierarchy and routing grouping of the default management backend.
- User management, monitoring, and other pages degrade their UI based on plugin availability (hidden fields, filters, tree selectors).
- Navigation results dynamically refresh after plugin start/stop to ensure empty parent directories automatically converge.

### R&D and Delivery Impact

- Add official source plugin templates and naming constraints.
- Add plugin-owned table naming constraints and default database loading boundary constraints.
- Add E2E coverage for plugin start/stop, empty directory hiding, host downgrade, menu refresh, and demo-control write interception.
- Subsequent module implementation must evaluate whether a capability belongs to the host or a source plugin.
