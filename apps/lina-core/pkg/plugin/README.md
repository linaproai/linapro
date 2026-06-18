# LinaPro Plugin Public Contracts

`apps/lina-core/pkg/plugin` contains the stable plugin-facing contracts for `lina-core`. Source plugins, dynamic plugin guests, dynamic plugin builders, and host integration code should all use this public boundary. Host-owned orchestration and persistence implementations live under `apps/lina-core/internal/service/plugin`.

## Public Components

| Component | Responsibility |
| --- | --- |
| `capability` | Defines the unified `capability.Services` directory and subpackage contracts such as users, files, i18n, jobs, AI, tenant, organization, storage, cache, and lock. Source plugins receive the full directory directly; dynamic plugins access only the published bridge-backed subset. Dynamic-plugin i18n resources are host-managed and no `i18n` host service is published. |
| `pluginhost` | Defines source-plugin declaration-time contracts, runtime service access, lifecycle callbacks, hook registration, HTTP route contribution, scheduled job contribution, menu filtering, permission filtering, and hosted asset constants. |
| `pluginbridge` | Provides the dynamic-plugin guest SDK. Guest code uses `pluginbridge.Declarations` during discovery or build flows and `pluginbridge.Services` at runtime. |

## Domain Capabilities

`capability.Services` is the runtime directory for host-owned domain capabilities. Source plugins consume these entries through `pluginhost.Services`; dynamic plugins declare the matching entries that are explicitly published as dynamic `hostServices` and call them through `pluginbridge.Services`. `I18n()` remains a source-plugin runtime capability, while dynamic-plugin i18n resources are discovered, merged, cached, and served by the host. Trusted source-plugin management commands stay in `capability.AdminServices`, and host-internal scope helpers are not exposed to ordinary plugin-facing access.

### `AdminServices` and Dynamic Plugins

`capability.AdminServices` is intentionally exposed only through `pluginhost.Services.Admin()` for trusted source plugins. Dynamic plugins do not have an `Admin()` entry in `pluginbridge.Services`, so they cannot directly consume domain `AdminService` interfaces such as `sessioncap.AdminService` or `notifycap.AdminService`.

Dynamic plugins may use only the concrete methods that are explicitly published as dynamic `hostServices`, declared by the plugin manifest, authorized by the host, and registered in the `WASM host-service` dispatcher. For example, the current `sessions` dynamic service publishes read, search, batch, user-online-status, and visibility methods; it does not publish `sessioncap.AdminService.Revoke`. If a management command must become available to dynamic plugins, add a narrow, versioned `host-service` method for that command instead of exposing the full `AdminServices` directory.

| Domain capability | Responsibility boundary | Runtime and validation path |
| --- | --- | --- |
| `APIDoc` | Resolves route text and title operation keys in localized API documentation. | Served through the capability directory or `apidoc` host calls; validation focuses on route and operation-key payloads. |
| `Auth`/`Authz` | Handles tenant token selection or switching, impersonation tokens, permission projections, single permission checks, and batch permission checks. | Runtime checks use current user, tenant, and permission keys; management commands remain in `AdminServices.Auth()`. |
| `AI` | Aggregates typed AI sub-capabilities for text, image, embedding, audio, vision, document, safety, and video, including method-level status projections. | Calls are authorized by method, not resources, and plugin identity is injected into provider requests for audit and governance. Status reads expose only availability, reason, and public provider identity. |
| `Users` | Provides current user projection, user batch lookup, batch user resolution, user search, and visibility enforcement. | Host implementations must keep user existence and visibility checks scoped to the caller. |
| `BizCtx` | Projects the current business request context. | Used as a read-only runtime context bridge for request user, tenant, locale, and request metadata. |
| `Dict` | Resolves dictionary value labels, lists bounded value candidates, and validates typed value visibility. | Host validation stays within visible dictionary types and values. |
| `Files` | Provides host file-center projections, bounded search, and visibility enforcement for existing `sys_file` resources. | Host validation prevents plugins from probing or using file IDs outside their visible boundary; it does not write, read, delete, or list plugin-private objects. |
| `HostConfig` | Reads governed host runtime configuration values. | Dynamic declarations must list `resources.keys`; source plugins receive a narrow read-only service. |
| `I18n` | Reads locale, translates messages, and finds message keys for source plugins. | Source plugins receive this through `pluginhost.Services`; dynamic plugins do not receive an `i18n` host service because their i18n resources are host-managed. |
| `Infra` | Reads infrastructure component status projections. | Validation focuses on visible component IDs and read-only status projection. |
| `Jobs` | Reads scheduled-job metadata, searches bounded job candidates, validates job visibility, and registers dynamic jobs. | Declaration-time job contracts are validated before runtime job discovery and execution; runtime search and visibility checks stay read-only. |
| `Notifications` | Reads typed notification message projections, batch-loads messages by business source, validates message visibility, and sends governed notifications. | Read calls do not require resource declarations and stay actor-visible; `messages.send` requires a `resources[].ref` boundary. |
| `Plugins` | Exposes current plugin projection, plugin registry projections, tenant plugin pages, capability status, plugin-scoped config, enablement state, and tenant lifecycle hooks. | Runtime checks cover plugin visibility, provider enablement, authoritative state, and tenant lifecycle preconditions. |
| `Route` | Exposes dynamic-route metadata for the current execution. | Used by runtime route dispatch without exposing host router internals. |
| `Sessions` | Reads the current online-session projection, searches sessions, batch-loads session projections, validates session visibility, and batch-reads user online status. | Host validation keeps session and user visibility scoped to the caller. |
| `Storage` | Provides plugin-private object storage operations for plugin-owned attachments, binary objects, import/export temporaries, and uninstall cleanup, including explicit batch stat, cursor list, and batch delete. | Source plugins receive plugin-scoped `Storage()` through `pluginhost.Services`; dynamic declarations use `service: storage` with `resources.paths`; writes do not create `sys_file` rows or expose provider keys or local paths. |
| `Cache` | Provides plugin-scoped cache get, set, delete, multi-key get/set/delete, increment, and expiry operations. | Dynamic declarations use `resources[].ref`; runtime dispatch validates namespace, key, value size, and TTL payloads. Cache remains non-authoritative plugin runtime data. |
| `Lock` | Provides plugin-visible distributed lock acquire, renew, and release operations. | Dynamic declarations use `resources[].ref`; runtime dispatch validates lock resource and lease payloads. |
| `Manifest` | Reads plugin-scoped manifest or artifact resources, including bounded multi-get and path listing. | Dynamic declarations use `resources.paths`; source and dynamic paths are resolved through plugin-scoped resource views. |
| `Org` | Provides optional organization projections such as user organization profiles, bounded department trees, department search, post options, visibility checks, department assignments, department names, and post IDs. | Provider availability is explicit; fallback services return safe neutral values when the organization provider is absent. |
| `Tenant` | Provides optional tenant context, tenant info, tenant batch/search projections, visibility checks, membership validation, accessible tenant lists, batch user tenant lists, and tenant switching validation. | Provider availability is explicit; host filters apply tenant scope without exposing tenant storage internals. |

## Dynamic-Plugin-Only Capabilities

`Runtime`, `Network`, and `RecordStore` are dynamic-plugin-only entries on `pluginbridge.Services`. They are not part of `capability.Services` because source plugins either already run inside the host process with native equivalents or use source-plugin data access seams instead of guest host-service wrappers.

| Capability | Public entry | Boundary reason |
| --- | --- | --- |
| `Runtime` | `pluginbridge.Services.Runtime()` | Dynamic plugins need a WASI host-service client for logs, state, time, UUIDs, and node identity; source plugins use host-native logging and runtime context directly. |
| `Network` | `pluginbridge.Services.Network()` | Dynamic plugins need governed outbound HTTP through host-service authorization; source plugins use host-native HTTP clients or injected domain services. |
| `RecordStore` | `pluginbridge.Services.RecordStore()` and `pkg/plugin/pluginbridge/recordstore` | Dynamic plugins need a guest-side facade over the data host-service protocol and typed query plans; source plugins use their own DAO or provider seams. |

New capabilities should enter `capability.Services` only when source plugins and dynamic plugins share the same stable host-owned domain contract. Dynamic-only host-service clients and guest SDKs stay under `pluginbridge`.

### `Storage()` and `Files()` Boundary

| Scenario | Use | Boundary |
| --- | --- | --- |
| A plugin stores its own attachment, generated export, binary object, or temporary import file. | `Storage()` / dynamic `service: storage` | The plugin passes a logical object path. The host scopes it by plugin ID and tenant before delegating to the active storage provider. The object stays outside host file-center lists and does not create `sys_file` metadata. |
| A plugin deletes, lists, stats, or purges objects it owns during record deletion or uninstall cleanup. | `Storage()` / dynamic `service: storage` | Cleanup uses `Delete` or bounded `List` by logical prefix. Plugins must not delete host upload roots, provider roots, or file-center rows directly. |
| A plugin references files that users already uploaded into the host file center. | `Files()` / dynamic `service: files` | The plugin receives `filecap.FileProjection` values and visibility checks for host-owned file IDs. The response must not expose DAO, DO, Entity, provider object keys, or local absolute paths. |
| A plugin command accepts host file IDs from a request. | `Files().EnsureVisible` / `files.visible.ensure` | The command checks all IDs before mutation. Missing and invisible files share the same rejection semantics to avoid existence probing. |

`Storage()` provider selection is configuration-free. The host uses the only
enabled storage provider plugin when exactly one is serviceable, falls back to
the built-in local provider when none is serviceable, and rejects storage calls
when multiple provider plugins are serviceable.

## Consumer Contracts, Provider SPI, and Guest SDK

Plugin-facing packages use three separate boundaries so each caller imports only
the contract it can safely depend on:

| Boundary | Package shape | Intended callers | Must not contain |
| --- | --- | --- | --- |
| Ordinary consumer contract | `pkg/plugin/capability/<domain>cap` | Source plugins through `pluginhost.Services`, dynamic plugins through generated or bridge-backed clients, and host adapters | GoFrame database builders, GoFrame HTTP request objects, provider factory registration, or host-private implementation state |
| Source-plugin provider SPI | `pkg/plugin/capability/<domain>cap/<domain>spi` | Source plugins that implement a host domain provider, plus host capability assembly code | Dynamic-plugin guest SDK imports or WASM host-service wire contracts |
| Dynamic-plugin guest SDK | `pkg/plugin/pluginbridge` and its dynamic-only subpackages | WASM guest code and dynamic plugin builders | Provider SPI imports or source-plugin registration APIs |

Provider factory declarations belong to `pluginhost.Declarations.Providers()`.
Source provider plugins declare factories there with domain-specific methods such
as `ProvideTenant`, `ProvideOrg`, and `ProvideAIText`. Host startup owns the
provider manager instances and injects the shared managers into host capability
services; ordinary `capability` packages do not keep package-level provider
registries.

## Host Domain Implementation

`apps/lina-core/internal/service/plugin` is the host-side plugin domain component. The root package exposes a unified facade covering plugin discovery, management lists, install, enable, disable, uninstall, runtime upgrades, source upgrades, runtime route dispatch, frontend asset serving, dependency checks, and capability wiring.

## Declaration-Time and Runtime Capabilities

### Declaration-Time Capabilities

Declaration-time capabilities are the plugin's static declarations and registration output. The host uses them before business execution to build governance state.

Source plugins express declaration-time contracts through `pluginhost.Declarations`, including `Assets()`, `Lifecycle()`, `Hooks()`, `HTTP()`, `Jobs()`, and `Access()`.

Dynamic plugins express declaration-time contracts through `plugin.yaml`, WASM custom sections, `pluginbridge.Declarations.Routes().Group(...)`, `pluginbridge.Declarations.Jobs().Register(...)`, and embedded `protocol` contracts, such as routes, jobs, lifecycle handlers, backend resources, frontend assets, SQL, i18n resources, and `hostServices`.

### Runtime Capabilities

Runtime capabilities are the services available while plugin business logic is executing.

Source plugins access runtime capabilities through `pluginhost.Services`; this interface embeds ordinary `capability.Services` and additionally provides trusted source-plugin-only capabilities such as `Admin()` and `TenantFilter()`.

Dynamic plugins access published runtime capabilities through `pluginbridge.Services`. Calls are encoded through `pluginbridge/protocol`, transported through `WASI host call`, authorized by derived `HostCapabilities` and confirmed `HostServices`, then dispatched by `apps/lina-core/internal/service/plugin/internal/wasm`. This runtime directory does not expose source-plugin-only entries such as `Admin()`, `TenantFilter()`, or `I18n()`.

For each guest execution, the host builds a request-local `HostServices` authorization snapshot and every host call still checks `service`, `method`, and resource identity against that snapshot.

## Dynamic Plugin Host Service Declarations

Minimal shape:

```yaml
hostServices:
  - service: runtime
    methods:
      - log.write
```

Resource-scoped shape:

```yaml
hostServices:
  - service: storage
    methods: [get, put, put.init, put.chunk, put.commit, put.abort]
    resources:
      paths:
        - reports/
  - service: data
    methods: [list, get]
    resources:
      tables:
        - plugin_acme_demo_report
  - service: hostconfig
    methods: [get]
    resources:
      keys:
        - i18n.default
  - service: network
    methods: [request]
    resources:
      - url: https://*.example.com/api
  - service: notifications
    methods: [messages.send]
    resources:
      - ref: inbox
        attributes:
          channel: inbox
```

## Declarable Host Services

<!-- BEGIN generated:host-services -->
| Service | Resource declaration | Derived capability | Methods |
| --- | --- | --- | --- |
| `runtime` | None | `host:runtime` | `log.write`<br/>`state.get`<br/>`state.get_many`<br/>`state.set`<br/>`state.set_many`<br/>`state.delete`<br/>`state.delete_many`<br/>`info.now`<br/>`info.uuid`<br/>`info.node` |
| `storage` | `resources.paths` | `host:storage` | `put`<br/>`put.init`<br/>`put.chunk`<br/>`put.commit`<br/>`put.abort`<br/>`get`<br/>`delete`<br/>`delete.batch`<br/>`list`<br/>`list.cursor`<br/>`stat`<br/>`stat.batch` |
| `network` | `resources[].url` | `host:http:request` | `request` |
| `data` | `resources.tables` | `host:data:read`<br/>`host:data:mutate` | `list`<br/>`get`<br/>`batch_get`<br/>`create`<br/>`update`<br/>`delete`<br/>`transaction` |
| `cache` | `resources[].ref` | `host:cache` | `get`<br/>`get_many`<br/>`set`<br/>`set_many`<br/>`delete`<br/>`delete_many`<br/>`incr`<br/>`expire` |
| `lock` | `resources[].ref` | `host:lock` | `acquire`<br/>`renew`<br/>`release` |
| `secret` | `resources[].ref` | `host:secret` | `resolve` reserved |
| `event` | `resources[].ref` | `host:event:publish` | `publish` reserved |
| `queue` | `resources[].ref` | `host:queue:enqueue` | `enqueue` reserved |
| `hostconfig` | `resources.keys` | `host:hostconfig` | `get` |
| `manifest` | `resources.paths` | `host:manifest` | `get`<br/>`get_many`<br/>`list` |
| `apidoc` | None | `host:apidoc` | `route_text.resolve`<br/>`route_texts.resolve`<br/>`route_title_operation_keys.find` |
| `auth` | None | `host:auth:token` | `tenant.select`<br/>`tenant.switch`<br/>`impersonation_token.issue`<br/>`impersonation_token.revoke` |
| `authz` | None | `host:authz` | `permissions.batch_get`<br/>`permissions.batch_has`<br/>`permissions.has`<br/>`users.platform_admin.check` |
| `ai` | None | `host:ai:text`<br/>`host:ai`<br/>`host:ai:image`<br/>`host:ai:embedding`<br/>`host:ai:audio`<br/>`host:ai:vision`<br/>`host:ai:document`<br/>`host:ai:safety`<br/>`host:ai:video` | `text.generate`<br/>`text.method_status.get`<br/>`ai.methods.status.batch_get`<br/>`image.generate`<br/>`image.edit`<br/>`embedding.create`<br/>`audio.transcribe`<br/>`audio.synthesize`<br/>`vision.analyze`<br/>`document.analyze`<br/>`document.cite`<br/>`safety.moderate`<br/>`video.generate`<br/>`video.edit`<br/>`video.extend`<br/>`video.operation.get`<br/>`video.operation.cancel` |
| `users` | None | `host:users` | `users.current.get`<br/>`users.batch_get`<br/>`users.resolve.batch`<br/>`users.search`<br/>`users.visible.ensure` |
| `bizctx` | None | `host:bizctx` | `current.get` |
| `dict` | None | `host:dict` | `labels.resolve`<br/>`dict.values.list`<br/>`values.visible.ensure` |
| `files` | None | `host:files` | `files.batch_get`<br/>`files.search`<br/>`files.visible.ensure` |
| `infra` | None | `host:infra` | `status.batch_get` |
| `jobs` | None | `host:jobs` | `jobs.batch_get`<br/>`jobs.search`<br/>`jobs.visible.ensure`<br/>`jobs.register` |
| `notifications` | None for reads; `messages.send` uses `resources[].ref` | `host:notifications` | `messages.batch_get`<br/>`messages.by_source.batch_get`<br/>`messages.visible.ensure`<br/>`messages.send` |
| `plugins` | None | `host:plugins` | `plugins.current.get`<br/>`plugins.batch_get`<br/>`plugins.search`<br/>`plugins.tenant.list`<br/>`plugins.capabilities.status.batch_get`<br/>`plugins.enabled.check`<br/>`plugins.provider_enabled.check`<br/>`plugins.enabled_authoritative.check`<br/>`config.get`<br/>`lifecycle.tenant_plugin_disable.ensure`<br/>`lifecycle.tenant_plugin_disabled.notify`<br/>`lifecycle.tenant_delete.ensure`<br/>`lifecycle.tenant_deleted.notify` |
| `route` | None | `host:route` | `metadata.get` |
| `sessions` | None | `host:sessions` | `sessions.current.get`<br/>`sessions.search`<br/>`sessions.batch_get`<br/>`sessions.users.online.batch_get`<br/>`sessions.visible.ensure` |
| `org` | None | `host:org` | `capability.available`<br/>`capability.status`<br/>`users.dept_assignments.list`<br/>`users.org_profiles.batch_get`<br/>`users.dept_info.get`<br/>`users.dept_name.get`<br/>`users.dept_ids.get`<br/>`users.post_ids.get`<br/>`depts.tree.list`<br/>`depts.search`<br/>`posts.options.list`<br/>`depts.visible.ensure`<br/>`posts.visible.ensure` |
| `tenant` | None | `host:tenant` | `capability.available`<br/>`capability.status`<br/>`tenants.current`<br/>`tenants.current_info.get`<br/>`tenants.platform_bypass`<br/>`tenants.visible.ensure`<br/>`tenants.batch_get`<br/>`tenants.search`<br/>`users.tenant_membership.validate`<br/>`users.tenants.list`<br/>`users.tenants.batch_list`<br/>`tenants.visible.batch_ensure`<br/>`tenants.switch.validate` |
<!-- END generated:host-services -->

## Maintenance Notes

When plugin public contracts or dynamic `host service` descriptors change, update both `README.md` and `README.zh-CN.md` in this directory.
