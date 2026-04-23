## Why

Currently, LinaPro's host service has assumed multiple backend capabilities such as permissions, organization, content, monitoring, tasks, and development tools. As functions continue to increase, the left menu and host code boundaries are expanding in the direction of a "unified backend." This will weaken LinaPro's positioning as an "AI-driven full-stack development framework": although developers can start quickly after downloading the project, it is difficult to clearly distinguish which are the framework cores that must be retained and which are business modules that can be cut, replaced or evolved independently as needed.

The open source phase needs to prioritize host boundary convergence and source plugin reconstruction: the host only retains framework base capabilities such as authentication, permissions, menus, plugin management, and task scheduling; non-core modules in organization management, content management, and system monitoring are delivered in the form of source plugins, allowing developers to install and activate them on demand, and obtain a lighter backend basic disk by default, while retaining the stable menu structure and evolution path of subsequent overlay services.

## What Changes

- Reconstruct the default management background first-level menu structure to form 9 background mounting points stably provided by the host: `dashboard(dashboard)`, `Authority Management (iam)`, `Organization Management (org)`, `System Settings (setting)`, `Content Management (content)`, `System Monitoring (monitor)`, `Task Scheduling (scheduler)`, `Extension Center (extension)`, `Development Center (developer)`.
- It is stipulated that these first-level directories must be explicitly created and owned by the host with stable menu records. Plug-ins can only mount menus to these directories and cannot create new first-level directory systems on their own.
- Clarify host boundaries: `User Management`, `Role Management`, `Menu Management`, `Plug-in Management`, `Task Scheduling`, authentication session kernel, configuration and dictionary, etc. are retained in the host.
- Planning `Organization Management` as a source plugin `org-center` to carry department management and position management.
- Plan notification announcements in `content management` as source plugin `content-notice`.
- Plan `Online User`, `Service Monitoring`, `Operation Log`, `Login Log` under `System Monitoring` into 4 independent source plugins: `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog`.
- Specifies the separation of plugin management and navigation: `Extension Center' only hosts the plugin management entrance, and the actual function menu of the plugin must be hung in the host directory corresponding to the semantics.
- Specifies that empty parent directories are automatically hidden through navigation projection, but the host still retains these stable directory records to ensure long-term stability of `parent_key` resolution and subsequent expansion.
- Supplement host decoupling requirements for subsequent plugin migration: user management degrades organizational capabilities, decouples authentication sessions from online user management, and changes login logs and operation logs to host event emission + plugin dropout mode.
- Supplementary plugin storage boundary management: the host's default `init/mock' no longer creates, writes or assumes any source plugin business tables; the official source plugin business tables are uniformly migrated to the plugin scope namespace, and the host no longer retains the corresponding ORM artifacts and direct table lookup implementations.

## Capabilities

### New Capabilities

- `core-host-boundary-governance`: Defines that the open source stage host only retains the boundary constraints between the framework core and the governance base.
- `module-decoupling`: Define independent start, stop and host downgrade rules when non-core management modules are delivered as source plugins.
- `menu-management`: Introduce a stable first-level directory structure, plugin menu semantic mounting and empty directory hiding rules for the default backend.
- `plugin-manifest-lifecycle`: Supplement fixed mount points, domain-capability plugin IDs and host directory parent key constraints for official source plugins.

### Modified Capabilities

- `dept-management`: Department management capabilities are subsequently delivered by the `org-center` source plugin.
- `post-management`: Post management capabilities are subsequently delivered by the `org-center` source plugin.
- `notice-management`: The notification announcement capability is subsequently delivered by the `content-notice` source plugin and mounted from `content management`.
- `online-user`: Online user capabilities are subsequently delivered by the `monitor-online` source plugin, and the host only retains the authentication session kernel.
- `server-monitor`: The service monitoring capability is subsequently delivered by the `monitor-server` source plugin.
- `oper-log`: The operation log capability is subsequently delivered by the `monitor-operlog` source plugin, and the host emits unified audit events instead.
- `login-log`: The login log capability is subsequently delivered by the `monitor-loginlog` source plugin, and the host emits unified login events instead.
- `user-management`: User management needs to hide department/position related filters, fields and form items and keep the main functions available when the organization plugin is missing.
- `user-auth`: The authentication link needs to publish login life cycle events, but it can no longer directly rely on the specific login log persistence implementation.

## Impact

### Backend Impact

- Adjust the host menu initialization SQL and menu projection logic to provide a stable first-level directory, accurate parent `menu_key` and a hideable host directory skeleton.
- Extract host dependencies of organization, notification, and monitoring related modules to prepare for source plugin migration.
- Supplement the "host kernel/plugin display management" boundary for login, audit, online session, service monitoring and other links.
- Added official source plugin directory and explicit wiring.
- Tighten the plugin storage boundary: the host no longer holds the `dao/do/entity`, Mock data and direct table lookup logic of the plugin business table. The plugin manages the business storage independently by installing SQL and capability providers.

### Front-end impact

- Adjust the left menu hierarchy and routing grouping of the default management background.
- User management, monitoring and other pages have their UI downgraded and hidden based on plugin availability.
- The navigation results are dynamically refreshed after the plugin is started and stopped to ensure that empty parent directories are automatically converged.

### R&D and Delivery Impact

- Added official source plugin templates and naming constraints.
- Added plugin own table naming constraints and default database loading boundary constraints.
- Need to add E2E coverage for plugin start and stop, empty directory hiding, host downgrade and menu refresh.
- Subsequent module implementation must prioritize whether this capability should fall on the host or the source plugin.
