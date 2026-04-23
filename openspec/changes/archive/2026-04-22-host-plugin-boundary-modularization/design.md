## Overview

This change aims to "slim down the host in the open source stage" and adopts the design of "host stable mount points + official source plugins are placed according to semantics" instead of stacking all expansion capabilities into the plugin center. The host is responsible for providing the framework core, management base, and stable navigation skeleton; non-core modules such as organization, content, and monitoring are delivered as source plugins, and participate in menus, routing, permissions, and frontend page assembly through the plugin life cycle.

## Goals

- Make host boundaries clear and readable: developers can know at a glance which are framework cores and which are optional modules.
- Keep the main background scene easy to use: the naming of the first-level menu continues to follow the common cognition of the management background, rather than abstract platform terms.
- Supports independent installation, activation, deactivation, and uninstallation of source plugins, and ensures smooth host degradation when plugins are missing.
- Provide a stable menu mounting point for the subsequent addition of official plugins and business modules, and no longer frequently change the first-level directory.

## Non-Goals

- Platform capabilities such as commercial plugin market, signature authorization, and billing distribution will not be introduced this time.
- The task scheduling capability will not be pluginized this time.
- This time it is not required to complete all code migration implementation at one time; the focus is to fix the boundaries, menus and migration order first.

## Menu Architecture

### Stable Top-Level Catalogs

The host permanently provides the following first-level directories and stable parent `menu_key`, and maintains them as the real `sys_menu` directory records owned by the host for long-term maintenance:

- `workbench` -> `dashboard`
- `Permission Management` -> `iam`
- `Organization Management` -> `org`
- `System settings` -> `setting`
- `Content Management` -> `content`
- `System Monitoring` -> `monitor`
- `Task Scheduling` -> `scheduler`
- `Extension Center` -> `extension`
- `Development Center` -> `developer`

These directory records are always created and owned by the host; whether empty directories are displayed is determined by the menu projection layer, rather than relying on plugins to dynamically create or delete parent directories.

### Menu Tree

```text
workbench
  - Analysis page
  - workbench
Permission management
  - User management
  - Role management
  - Menu management
Organizational management
  - Department management
  - Position management
System settings
  - Dictionary management
  - Parameter settings
  - File management
Content management
  - Notices and announcements
System monitoring
  - Online users
  - Service monitoring
  - Operation log
  - Login log
Task scheduling
  - Task management
  - Group management
  - Execution log
Extension Center
  - Plug-in management
Development Center
  - Interface documentation
  - System information
```

### Navigation Rules

- The first-level directory is created by the host and is not dynamically added by the plugin.
- The plugin function menu must be hung in the host directory corresponding to the semantics.
- `Extension Center` only displays the plugin management entrance and does not include the actual business menu.
- If a directory at a certain level does not have any visible submenus, the parent directory is automatically hidden in the navigation projection, but the host still retains the corresponding stable directory record.
- After the plugin status changes, the host triggers the current user menu and dynamic routing to refresh, and finally converges to the latest status.

## Host Boundary

### Host-Retained Capabilities

The following abilities must remain in the host:

- Authentication, JWT, login state analysis, user context.
- User management, role management, menu management, permission verification.
- Plug-in registry, installation/uninstallation, start and stop, menu synchronization, and management resource synchronization.
- Dictionary capabilities, parameter configuration capabilities, and file basic capabilities.
- Task scheduling platform capabilities.
-Host unified event/Hook mechanism.
- First-level directory records and menu projection management logic owned by the host.
- Governance and development assistance capabilities provided by `Extension Center` and `Development Center`.

### Source-Plugin Capabilities

The following official source plugins are fixed this time:

- `org-center`
- Menu mount: `org`
  - Bearing: department management, position management.
- `content-notice`
- Menu mounting: `content`
  - Bearer: notification announcement.
- `monitor-online`
- Menu mount: `monitor`
  - Bearing: online user query and forced offline management.
- `monitor-server`
- Menu mount: `monitor`
  - Bearer: service monitoring collection, cleaning and display.
- `monitor-operlog`
- Menu mount: `monitor`
  - Bearing: operation log query, export, cleaning and details.
- `monitor-loginlog`
- Menu mount: `monitor`
  - Bearing: login log query, export, cleaning and details.

## Critical Decoupling Strategy

### Organization Decoupling

Current user management directly relies on the organization tree, department fields and position options. If `org-center` is not installed, the host must still ensure that user management is available. Therefore, we need to abstract the organizational capability provider first:

- The department column in the user list has been changed to optional display.
- The department tree and position selector in the user form are changed to be displayed on demand based on organization plugin capability detection.
- Organization information in user details is changed to optional extended data and is no longer considered a host hard field.
- The host business logic only relies on the "organizational capability interface" and does not directly rely on the `dept` / `post` service implementation.
- `orgcap` serves as a host capability seam, retaining only the interfaces, DTOs, empty implementations and call entries owned by the host; `org-center` registers its real implementation through the stable Provider after installation.
- Even if the plugin business tables share the same database as the host, the host must not directly hold or query the physical tables, ORM artifacts and associated writing logic of `org-center`.
- When `org-center` is missing, the user management page degrades to the core user management view without left-tree filtering and department/position fields.

### Online Session vs Online User Plugin

Although `Online User` needs to be an independent source plugin, the host authentication link still relies on online session validity verification. The current authentication middleware directly calls the session store to do `TouchOrValidate`, therefore:

- The host retains the authenticated session kernel, `sys_online_session` source of truth, and session storage abstraction.
- The host continues to be responsible for logging in to create a session, logging out to delete a session, refreshing the active time when a request is reached, timeout determination and cleanup tasks.
- The host publishes independent session DTO / filter / result contracts to `monitor-online` to avoid directly exposing internal `session` type aliases to plugins.
- `monitor-online` is only responsible for reading session projections, displaying online user lists and performing forced offline management.
- `monitor-online` does not have JWT validation, session timeout semantics, `last_active_time` maintenance or cleanup task truth sources.
- When the plugin is not installed, the host can still log in, log out, and verify session timeout normally.

### Login Log / Oper Log Eventization

The current login log and operation log are directly dependent on the specific log service by the host core link. In order to achieve plugin, it needs to be changed to:

- The host defines a unified login event contract, covering scenarios such as successful login, failed login, and successful logout.
- The host defines a unified audit event contract, covering write operations and audited query operations with the `operLog` tag.
- The host core authentication link and request link are only responsible for emitting events and no longer directly rely on the specific `loginlog` / `operlog` implementation.
- `monitor-loginlog` and `monitor-operlog` complete logging, querying, exporting and cleaning after subscribing to events.
- When the plugin is not installed, the host event emission link can still be executed, but it is not forced to be dropped from the library, nor is it allowed to block authentication, authentication or ordinary business requests.

In order to prevent the host from continuing to retain the plugin-specific operation log middleware, and at the same time allowing the source plugin to obtain the complete request results, HTTP seams need to be further converged:

- The host publishes the routing register and the global middleware register at the same time on the unified HTTP registration entrance of the source plugin, instead of retaining a dedicated host hanging point for `monitor-operlog`.
- GoFrame HOOK is suitable for pre- and post-observation, but it is not suitable as the only implementation of operation logs, because the request to end early will skip the post-HOOK, and the complete completion status result cannot be stably obtained.
- `monitor-operlog` registers its own audit middleware through the global middleware register encapsulated by the host, and reuses the host's unified audit event distribution.
- The host uniformly packages the start and stop switches of the global middleware that the plugin self-registers; when the plugin is deactivated, its logic is directly bypassed, and the HTTP routing tree is not required to be rebuilt.

### Capability Seams Instead of Placeholder Branches

This split does not accept the implementation method of "the host retains the complete business skeleton + scattering `if pluginEnabled` judgments everywhere". A more stable and tidy seam between hosts and plugins is needed:

- The host retains a unified capability interface, such as the organizational capability `orgcap`, which is implemented by the host core module only relying on interfaces, DTOs and empty; after the plugin is installed, the real capability implementation is provided through the registered Provider/Adapter.
- The host uses unified Hook events to emit events to links such as login logs and operation logs, instead of retaining specific plugin branches in authentication and middleware links.
- Source plugins take over their own HTTP API and scheduled tasks through the route register and Cron register, and do not require the host to reserve static binding code for these plugins.
- The host only retains the truth source and management base that really belong to the framework core, such as `sys_online_session` session truth source, plugin life cycle, menu management and task scheduling base.
- When the plugin is missing, the host smoothly degrades through the "zero value/empty collection/no extra fields" semantics of the capability interface, instead of splicing different code paths after judging everywhere.

### Plugin-Local ORM Generation

The database access of the source plugin backend needs to be closed in the plugin directory, rather than flowing back to the host `dao/model` artifact or long-term dependency scattered `g.DB().Model(...)` statements:

- The `backend/` directory of each official source plugin maintains an independent `hack/config.yaml` to ensure that developers can execute `gf gen dao` after entering the plugin backend directory.
- The plugin locally generates and maintains `internal/dao`, `internal/model/do`, and `internal/model/entity`, which are directly consumed by the plugin service.
- When the plugin reads the host shared table (such as `sys_user`, `sys_dict_data`), it also completes the query through the plugin's local codegen artifact to avoid relying on the host `dao/model` package again.
- Once a business table completes the plugin migration, the host must simultaneously delete the `dao/do/entity` corresponding to the table and implement the direct table lookup to avoid the formation of "double ORM artifacts + double storage entries".
- Source code plugins that do not directly check the library in the current version still retain the local codegen configuration, so that they can continue to use the unified structure during subsequent expansion.

### Plugin-Owned Storage Lifecycle and Naming

The data storage of the source plugin must be clearly identifiable outside the host boundary and cannot continue to be disguised as the host's built-in business table:

- The host `make init` only initializes the host kernel table and necessary Seed data, and does not create any source plugin business tables.
- The host `make mock` no longer writes the plugin business table; plugin demo data should be loaded by the plugin installation SQL, plugin's own mock resources or plugin exclusive commands.
- The business physical tables of the official source plugin uniformly use the plugin scope naming `plugin_<plugin_id_snake_case>`; single-table plugins preferentially use the complete table name directly, and multi-table plugins add business suffixes as needed, such as `plugin_org_center_dept`, `plugin_content_notice`, `plugin_monitor_loginlog`; the `sys_` prefix is ​​only reserved for the host core table.
- If the plugin needs to read the host's shared management data, it is only allowed to explicitly rely on the host's shared table (such as `sys_user`, `sys_dict_data`, `sys_online_session`), and is not allowed to reversely require the host to hold the plugin business table.
- Plug-in installation/uninstall SQL, plugin local `gf gen dao` configuration, plugin service and test data must be consistent around the above plugin scope table name.

### Server Monitor Migration

`monitor-server` can be migrated as a relatively complete independent plugin:

- The plugin has collectors, cleaning tasks, data tables, query interfaces and pages.
- The host only provides task scheduling base and plugin life cycle capabilities.
- Plug-in startup and shutdown should be linked to collection task registration and cancellation.

## Plugin Manifest Rules

- Plugin `id` must use `kebab-case`, but does not require `plugin-` prefix.
- Unified usage areas of official plugins - capability styles: `org-center`, `content-notice`, `monitor-online`, etc.
- The plugin menu key continues to use the `plugin:<plugin-id>:<menu-key>` format to ensure governance consistency.
- Plugin `parent_key` must point to the host stable directory key or the same plugin internal menu key.
- The top-level mounting relationships of official plugins are fixed as: `org-center -> org`, `content-notice -> content`, `monitor-online -> monitor`, `monitor-server -> monitor`, `monitor-operlog -> monitor`, `monitor-loginlog -> monitor`.

## Migration Order

### Phase 1: Governance Foundation

- Fixed first-level directory and host parent `menu_key`.
- Create and maintain 9 first-level directory records owned by the host.
- Adjust menu SQL, menu projection and navigation hiding rules.
- Fixed plugin ID and menu mounting constraints.

### Phase 2: Event and Boundary Extraction

- Pump the login event contract with the publisher.
- Extract audit event contracts and publishers.
- Extract organizational capability interface.
- Draw a clear boundary between authentication session kernel and online user management.

### Phase 3: Independent Monitor Plugins

- Migrate `monitor-operlog`.
- Migrate `monitor-loginlog`.
- Migrate `monitor-server`.
- Finally moved to `monitor-online`.

### Phase 4: Organization and Content Plugins

- Migrate `org-center`.
- Migrated `content-notice`.

## Risks and Mitigations

### Risk: User Management Loses Required Fields

- Risk: When the organization plugin is not installed, the user management page and interface still assume that the department/position must exist.
- Mitigation: Do capability detection and field downgrade first, then migrate the plugin.

### Risk: Authentication Depends on Online User Plugin

- Risk: Moving all online users out of the host will destroy the main authentication link.
- Mitigation: Keep the host session kernel and only migrate the display and management capabilities.

### Risk: Menu Tree Becomes Empty or Fragmented

- Risk: An empty parent directory appears when the plugin is not installed, or the first-level directory only exists in the frontend projection, causing `parent_key` to be unable to be stably parsed.
- Mitigation: First-level directory hosting, stable directory records are permanent, empty parent directories are only hidden in the navigation projection layer, and semantic mounting rules are fixed.

### Risk: Too Many Small Plugins Increase Maintenance Cost

- Risk: After monitoring is split into 4 plugins, documentation, testing and sample maintenance will increase.
- Mitigation: Accept this split as a product requirement, while unifying plugin templates, menu mounting rules and test specifications to reduce marginal costs.
