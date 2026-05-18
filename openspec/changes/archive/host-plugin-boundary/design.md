## Overview

This change establishes a clear boundary between the LinaPro host framework and its plugin ecosystem. The host retains only framework core capabilities -- authentication, permissions, menus, plugin management, task scheduling, configuration, and dictionary -- while delivering non-core management modules (organization, content, monitoring) as official source plugins. It also introduces a demo-control source plugin that enforces environment-level write protection through the same plugin lifecycle and global middleware seams, and refactors the plugin bridge infrastructure into responsibility-scoped subcomponents.

The design follows two complementary principles: the host provides stable mount points and capability seams, while plugins provide domain-specific business logic and participate in menus, routing, permissions, and page assembly through the plugin lifecycle. The plugin bridge infrastructure follows the same principle of responsibility separation, with each subcomponent owning a clear slice of the bridge protocol.

## Goals

- Make host boundaries clear and readable: developers can identify framework cores versus optional modules at a glance.
- Keep the main background scene easy to use: the naming of first-level menus continues to follow common management backend conventions.
- Support independent installation, activation, deactivation, and uninstallation of source plugins, with smooth host degradation when plugins are missing.
- Provide stable menu mounting points for subsequent official plugins and business modules.
- Protect demo environments through a plugin-driven write guard that reuses RESTful HTTP Method semantics.
- Organize `pkg/pluginbridge` into responsibility-scoped subcomponents so that plugin developers and host maintainers can locate bridge capabilities by their purpose rather than scanning a flat root package.

## Non-Goals

- Platform capabilities such as commercial plugin market, signature authorization, and billing distribution are not introduced.
- Task scheduling capability is not pluginized.
- Complete code migration is not required in one iteration; the focus is fixing boundaries, menus, and migration order.
- No dedicated frontend page, banner, or toolbar hint for demo mode in this iteration.
- No fine-grained resource whitelists, role-level exceptions, or module-specific bypasses for demo control.
- The pluginbridge subcomponent refactoring does not change the dynamic plugin Wasm bridge protocol, host call entry, host service method names, or payload field numbering.
- No new external dependencies or code generation flows are introduced for the pluginbridge restructuring.

## Menu Architecture

### Stable Top-Level Catalogs

The host permanently provides the following first-level directories and stable parent `menu_key`, maintained as real `sys_menu` directory records owned by the host:

- `workbench` -> `dashboard`
- `Permission Management` -> `iam`
- `Organization Management` -> `org`
- `System Settings` -> `setting`
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
Permission Management
  - User management
  - Role management
  - Menu management
Organization Management
  - Department management
  - Position management
System Settings
  - Dictionary management
  - Parameter settings
  - File management
Content Management
  - Notices and announcements
System Monitoring
  - Online users
  - Service monitoring
  - Operation log
  - Login log
Task Scheduling
  - Task management
  - Group management
  - Execution log
Extension Center
  - Plugin management
Development Center
  - Interface documentation
  - System information
```

### Navigation Rules

- The first-level directory is created by the host and is not dynamically added by plugins.
- Plugin function menus must be mounted in the host directory corresponding to the semantics.
- `Extension Center` only displays the plugin management entrance and does not include the actual business menu.
- If a directory at a certain level does not have any visible submenus, the parent directory is automatically hidden in the navigation projection, but the host still retains the corresponding stable directory record.
- After plugin status changes, the host triggers the current user menu and dynamic routing to refresh.

## Host Boundary

### Host-Retained Capabilities

The following abilities must remain in the host:

- Authentication, JWT, login state analysis, user context.
- User management, role management, menu management, permission verification.
- Plugin registry, installation/uninstallation, start and stop, menu synchronization, and management resource synchronization.
- Dictionary capabilities, parameter configuration capabilities, and file basic capabilities.
- Task scheduling platform capabilities.
- Host unified event/Hook mechanism.
- First-level directory records and menu projection management logic.
- Governance and development assistance capabilities provided by `Extension Center` and `Development Center`.

### Source-Plugin Capabilities

The following official source plugins are defined:

- `org-center` -- Menu mount: `org`. Carries department management and position management.
- `content-notice` -- Menu mount: `content`. Carries notification announcements.
- `monitor-online` -- Menu mount: `monitor`. Carries online user query and forced offline management.
- `monitor-server` -- Menu mount: `monitor`. Carries service monitoring collection, cleaning, and display.
- `monitor-operlog` -- Menu mount: `monitor`. Carries operation log query, details, export, cleaning, and page.
- `monitor-loginlog` -- Menu mount: `monitor`. Carries login log query, details, export, cleaning, and page.
- `demo-control` -- No menu mount. Carries demo-mode write-protection guard via global HTTP middleware.

## Critical Decoupling Strategy

### Organization Decoupling

Current user management directly relies on the organization tree, department fields, and position options. If `org-center` is not installed, the host must still ensure user management is available. Therefore:

- The department column in the user list is changed to optional display.
- The department tree and position selector in the user form are displayed on demand based on organization plugin capability detection.
- Organization information in user details is changed to optional extended data, no longer a host hard field.
- The host business logic only relies on the organizational capability interface and does not directly depend on `dept` / `post` service implementations.
- `orgcap` serves as a host capability seam, retaining only interfaces, DTOs, empty implementations, and call entries owned by the host; `org-center` registers its real implementation through the stable Provider after installation.
- Even if plugin business tables share the same database as the host, the host must not directly hold or query the physical tables, ORM artifacts, or associated writing logic of `org-center`.
- When `org-center` is missing, the user management page degrades to the core user management view without left-tree filtering and department/position fields.

### Online Session vs Online User Plugin

The host authentication link relies on online session validity verification. The current authentication middleware directly calls the session store for `TouchOrValidate`. Therefore:

- The host retains the authenticated session kernel, `sys_online_session` source of truth, and session storage abstraction.
- The host continues to be responsible for login session creation, logout session deletion, active time refresh on request, timeout determination, and cleanup tasks.
- The host publishes independent session DTO/filter/result contracts to `monitor-online` to avoid directly exposing internal `session` type aliases to plugins.
- `monitor-online` is only responsible for reading session projections, displaying online user lists, and performing forced offline management.
- `monitor-online` does not have JWT validation, session timeout semantics, `last_active_time` maintenance, or cleanup task truth sources.
- When the plugin is not installed, the host can still log in, log out, and verify session timeout normally.

### Login Log / Oper Log Eventization

The current login log and operation log are directly dependent on specific log services by host core links. To achieve pluginization:

- The host defines a unified login event contract covering successful login, failed login, and successful logout scenarios.
- The host defines a unified audit event contract covering write operations and audited query operations with the `operLog` tag.
- The host core authentication link and request link only emit events and no longer directly depend on specific `loginlog` / `operlog` implementations.
- `monitor-loginlog` and `monitor-operlog` complete logging, querying, exporting, and cleaning after subscribing to events.
- When the plugin is not installed, the host event emission link still executes but does not block authentication or ordinary business requests.

To prevent the host from retaining plugin-specific operation log middleware while allowing source plugins to obtain complete request results:

- The host publishes the routing register and global middleware register on the unified HTTP registration entrance for source plugins.
- `monitor-operlog` registers its own audit middleware through the global middleware register and reuses the host's unified audit event distribution.
- The host uniformly manages start/stop switches of plugin-registered global middleware; when deactivated, the logic is directly bypassed without rebuilding the HTTP routing tree.

### Capability Seams Instead of Placeholder Branches

This design does not accept "host retains complete business skeleton + scattered `if pluginEnabled` judgments everywhere." A more stable seam is needed:

- The host retains a unified capability interface (e.g., `orgcap`), implemented only through interfaces, DTOs, and empty implementations; after plugin installation, the real capability is provided through registered Provider/Adapter.
- The host uses unified Hook events to emit events to links such as login logs and operation logs.
- Source plugins take over their own HTTP API and scheduled tasks through route registers and Cron registers.
- When the plugin is missing, the host smoothly degrades through the "zero value/empty collection/no extra fields" semantics of the capability interface.

### Plugin-Local ORM Generation

The database access of source plugin backends is closed within the plugin directory:

- Each official source plugin's `backend/` directory maintains an independent `hack/config.yaml` for `gf gen dao`.
- The plugin locally generates and maintains `internal/dao`, `internal/model/do`, and `internal/model/entity`.
- When the plugin reads host shared tables (e.g., `sys_user`, `sys_dict_data`), it also uses the plugin's local codegen artifact.
- Once a business table completes plugin migration, the host deletes the corresponding `dao/do/entity` and direct table lookup implementation.

### Plugin-Owned Storage Lifecycle and Naming

Source plugin data storage must be clearly identifiable outside the host boundary:

- The host `make init` only initializes host kernel tables and necessary Seed data; it does not create source plugin business tables.
- The host `make mock` no longer writes to plugin business tables; plugin demo data is loaded by plugin installation SQL or plugin-specific mock resources.
- Official source plugin business tables uniformly use the plugin scope naming `plugin_<plugin_id_snake_case>`; the `sys_` prefix is only reserved for host core tables.
- If the plugin needs to read host shared management data, it only explicitly relies on host shared tables (e.g., `sys_user`, `sys_dict_data`, `sys_online_session`).

### Server Monitor Migration

`monitor-server` is migrated as a relatively complete independent plugin:

- The plugin has collectors, cleaning tasks, data tables, query interfaces, and pages.
- The host only provides task scheduling base and plugin lifecycle capabilities.
- Plugin startup and shutdown are linked to collection task registration and cancellation.

## Pluginbridge Subcomponent Architecture

### Context

`pkg/pluginbridge` is the shared infrastructure for the dynamic plugin Wasm bridge. It is used by the host runtime, WASM host functions, `plugindb`, dynamic plugin samples, and guest code. The root package directory contained approximately 41 production Go files spanning multiple responsibility layers: stable ABI and manifest contracts, bridge envelope codec, WASM artifact helpers, host call protocol, host service protocol, and guest SDK. Placing all of these in a single root package made it difficult for users to distinguish "the API plugin authors should use" from "the protocol details host runtime maintainers care about."

### Subcomponent Package Structure

The refactored `pkg/pluginbridge` is organized into the following subcomponent packages:

```text
pkg/pluginbridge/
  pluginbridge.go      # Root package facade: aliases and wrappers
  contract/            # ABI, route, cron, execution source stable contracts
  codec/               # Bridge request/response envelope codec
  artifact/            # Wasm section constants, custom section reading, runtime metadata
  hostcall/            # Host call opcode, generic host call envelope and status codes
  hostservice/         # Host service spec, capability derivation, payload codec
  guest/               # Guest runtime, controller dispatcher, BindJSON, host service clients
```

Each subcomponent has clear responsibilities:

- `contract` owns stable bridge contracts: `BridgeSpec`, `RouteContract`, `CronContract`, `ExecutionSource`, `HostServiceSpec`, and related DTOs.
- `codec` owns bridge request/response/route/identity/HTTP snapshot envelope encoding and decoding, including protobuf wire tools in its `internal` package.
- `artifact` owns WASM section constants (`WasmSectionI18NAssets`, `WasmSectionApidocAssets`, etc.), `RuntimeArtifactMetadata`, `ReadCustomSection`, and `ListCustomSections`.
- `hostcall` owns host call opcodes, status codes, `HostCallResponseEnvelope`, and generic host call codec.
- `hostservice` owns host service spec, capability derivation, manifest codec, and all host service method/payload codec (runtime, storage, network, data, cache, lock, config, notify, cron).
- `guest` owns guest runtime, typed guest controller dispatcher, context response helpers, `BindJSON`/`WriteJSON`, `ErrorClassifier`, and host service client helpers.

### Root Package Facade

The root package `pluginbridge` is preserved as a backward-compatible facade. Public constants, types, and functions continue to be accessible through the root package using:

- `type X = contract.X` for type aliases
- `const X = contract.X` for constant aliases
- `func EncodeRequestEnvelope(...) { return codec.EncodeRequestEnvelope(...) }` for wrapper functions
- `func Runtime() guest.RuntimeHostService { return guest.Runtime() }` for facade delegates

This ensures existing host code, dynamic plugin samples, and user plugin code that imports `lina-core/pkg/pluginbridge` continues to compile and behave identically. The facade does not replicate protocol implementation logic; all implementation resides in exactly one authoritative subcomponent.

### Dependency Direction

The subcomponent dependency graph flows from low-level to high-level:

```text
contract
  ^
codec --> internal/wire
  ^
artifact --> internal/wasmsection
  ^
hostservice --> contract, codec internal wire
  ^
hostcall --> hostservice
  ^
guest --> contract, codec, hostcall, hostservice
  ^
pluginbridge facade --> all subcomponents
```

Low-level contract and protocol subcomponents must not import the root package facade or the guest SDK. Any subcomponent's `internal` package serves only that subcomponent or sibling packages within the same parent path, and must not become a cross-domain utility dump.

### Host Internal Import Migration

Host internal code is migrated to use precise subcomponent imports:

- Runtime artifact parsing uses `pluginbridge/artifact`.
- Wasm executor uses `pluginbridge/codec`, `pluginbridge/hostcall`, `pluginbridge/hostservice`.
- Manifest and route contract validation uses `pluginbridge/contract`, `pluginbridge/hostservice`.

Dynamic plugin sample code may migrate to `pluginbridge/guest` but the root-package facade remains available as a compatibility path.

### Verification Strategy

Verification centers on protocol invariance, not file counts:

- `EncodeRequestEnvelope` / `DecodeRequestEnvelope` byte-level round trip must remain identical.
- All host service payload `Marshal` / `Unmarshal` round trips must remain identical.
- WASM custom section reading error boundaries must remain identical.
- `HostCallResponseEnvelope` and structured host service envelope must remain identical.
- Guest runtime, typed controller dispatcher, `BindJSON`/`WriteJSON` behavior must remain identical.
- Root package facade calls and direct subcomponent calls must produce identical results.

## Demo Control Guard

### Design

The `demo-control` source plugin provides environment-level write protection through the host's published global HTTP middleware registration seam (`pluginhost.HTTPRegistrar`).

### Switch Mechanism

The host does not add an independent boolean such as `demo.control.enabled`. It reuses `plugin.autoEnable` directly:

- If `plugin.autoEnable` does not include `demo-control`, demo protection is off by default.
- If `plugin.autoEnable` includes `demo-control`, the host installs and enables the plugin during startup and demo protection becomes active.
- Disabling demo protection is as simple as removing `demo-control` from `plugin.autoEnable` and restarting the host.

### Write Interception Rules

When `demo-control` is enabled:

- `GET`, `HEAD`, and `OPTIONS` requests are allowed across `/*`.
- `POST`, `PUT`, and `DELETE` requests are rejected by default with a clear read-only demo message.
- Login (`POST /api/v1/auth/login`) and logout (`POST /api/v1/auth/logout`) are preserved as a session whitelist.
- Install, uninstall, enable, and disable operations for plugins other than `demo-control` itself are preserved as a plugin-governance whitelist.
- `demo-control` cannot change its own governance state while enabled (self-protection).

### Why a Source Plugin

- The requirement is a classic environment-governance capability and a first-party example of how source plugins reuse host global middleware seams.
- The host core middleware chain remains generic and does not become coupled to "demo mode" semantics.
- Plugin lifecycle governance stays consistent.

## Plugin Manifest Rules

- Plugin `id` must use `kebab-case` without requiring a `plugin-` prefix.
- Official plugins use domain-capability naming: `org-center`, `content-notice`, `monitor-online`, etc.
- Plugin menu key continues to use the `plugin:<plugin-id>:<menu-key>` format.
- Plugin `parent_key` must point to the host stable directory key or the same plugin internal menu key.
- Top-level mounting relationships: `org-center -> org`, `content-notice -> content`, `monitor-online -> monitor`, `monitor-server -> monitor`, `monitor-operlog -> monitor`, `monitor-loginlog -> monitor`.

## Migration Order

### Phase 1: Governance Foundation

- Fix first-level directory and host parent `menu_key`.
- Create and maintain 9 first-level directory records owned by the host.
- Adjust menu SQL, menu projection, and navigation hiding rules.
- Fix plugin ID and menu mounting constraints.

### Phase 2: Event and Boundary Extraction

- Extract login event contract with publisher.
- Extract audit event contracts and publishers.
- Extract organizational capability interface.
- Draw a clear boundary between authentication session kernel and online user management.

### Phase 3: Independent Monitor Plugins

- Migrate `monitor-operlog`.
- Migrate `monitor-loginlog`.
- Migrate `monitor-server`.
- Migrate `monitor-online`.

### Phase 4: Organization and Content Plugins

- Migrate `org-center`.
- Migrate `content-notice`.

### Phase 5: Environment Governance

- Add `demo-control` source plugin with global middleware integration.
- Wire into `plugin.autoEnable` as the demo-mode switch.

### Phase 6: Pluginbridge Subcomponent Restructuring

- Create subcomponent package skeletons with package-level documentation.
- Migrate low-dependency capabilities (`contract`, `artifact`, `codec`) and maintain root package facade compilation.
- Migrate `hostservice` and `hostcall`, preserving existing serialization tests and adding facade consistency tests.
- Migrate `guest` SDK, update dynamic plugin samples or add compatibility tests.
- Update host internal imports to use precise subcomponent packages while preserving root-package external compatibility.
- Run related Go tests, `wasip1/wasm` builds, and OpenSpec validation.

## Risks and Mitigations

### Risk: User Management Loses Required Fields

When the organization plugin is not installed, the user management page and interface still assume department/position must exist. Mitigation: do capability detection and field downgrade first, then migrate the plugin.

### Risk: Authentication Depends on Online User Plugin

Moving all online users out of the host will destroy the main authentication link. Mitigation: keep the host session kernel and only migrate the display and management capabilities.

### Risk: Menu Tree Becomes Empty or Fragmented

An empty parent directory appears when the plugin is not installed, or the first-level directory only exists in the frontend projection. Mitigation: first-level directory hosting, stable directory records are permanent, empty parent directories are only hidden in the navigation projection layer.

### Risk: Too Many Small Plugins Increase Maintenance Cost

After monitoring is split into 4 plugins, documentation, testing, and sample maintenance will increase. Mitigation: accept this split as a product requirement while unifying plugin templates, menu mounting rules, and test specifications.

### Risk: Method-Based Demo Policy Bypasses RESTful Violations

Any interface that violates the repository's RESTful contract could bypass demo protection or be blocked incorrectly. Mitigation: the project already requires all APIs to follow RESTful semantics, so the implementation intentionally enforces that rule.

### Risk: Demo Mode Requires Config Edit Plus Host Restart

Using plugin enablement as the only switch means changing demo mode requires a config edit and host restart. Mitigation: demo protection is an environment-level governance strategy, and startup-time config is clearer than adding a second boolean switch.

### Risk: Pluginbridge Subcomponent Split Causes Import Cycles

Subcomponent refactoring may introduce circular dependencies if dependency direction is not enforced. Mitigation: move contract types and pure implementation utilities first, then migrate upper-level packages; root package facade connects last, and subcomponents are prohibited from importing the root package.

### Risk: Facade Alias/Wrapper Omissions Break Existing Callers

Missing aliases or wrappers in the root-package facade could cause compilation failures in existing code. Mitigation: run `go test ./pkg/pluginbridge/... ./internal/service/plugin/... ./pkg/plugindb/...` and build dynamic plugin samples during the refactoring process.

### Risk: Serialization Field Numbers or Defaults Change During Split

Protocol fields or default values could be inadvertently modified when code is moved between packages. Mitigation: preserve and migrate existing codec round-trip tests, and add facade-vs-subcomponent consistency tests.

### Risk: Too Many Fine-Grained Subcomponents Cause Import Confusion

Excessively granular subcomponent packages could make it difficult for users to choose the right import. Mitigation: establish only 5-6 stable subcomponents; the root-package facade and `guest` package serve as the primary entry points for plugin authors, with documentation explaining recommended imports.
