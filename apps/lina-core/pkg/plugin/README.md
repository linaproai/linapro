# LinaPro Plugin Public Contracts

`apps/lina-core/pkg/plugin` contains the stable plugin-facing contracts for `lina-core`. Source plugins, dynamic plugin guests, dynamic plugin builders, and host integration code should all use this public boundary. Host-owned orchestration and persistence implementations live under `apps/lina-core/internal/service/plugin`.

## Public Components

| Component | Responsibility |
| --- | --- |
| `capability` | Defines the unified `capability.Services` directory and subpackage contracts such as users, files, i18n, jobs, AI, tenant, organization, storage, cache, and lock. Source plugins receive this directory from `pluginhost` registrar and callback inputs; dynamic plugins access only the published bridge-backed subset. Dynamic-plugin i18n resources are host-managed and no `i18n` host service is published. |
| `pluginhost` | Defines source-plugin declaration-time contracts, runtime service access, lifecycle callbacks, hook registration, HTTP route contribution, scheduled job contribution, menu filtering, permission filtering, and hosted asset constants. |
| `pluginbridge` | Provides the dynamic-plugin guest SDK. Guest code uses `pluginbridge.Declarations` during discovery or build flows and `pluginbridge.Services` at runtime. |

## Domain Capabilities

`capability.Services` is the runtime directory for host-owned domain capabilities. Source plugins consume these entries through the `capability.Services` returned by `pluginhost` registrar and callback inputs; dynamic plugins declare the matching entries that are explicitly published as dynamic `hostServices` and call them through `pluginbridge.Services`. Each domain exposes one plugin-visible `Service`; method-level contracts carry risk, authorization, data-permission, context, performance, and cache governance. Domain methods rely on the standard business `ctx` for current user, tenant, permission, and data-scope information; dynamic `hostServices` authorization stays inside the dispatcher. `I18n()` remains a source-plugin runtime capability, while dynamic-plugin i18n resources are discovered, merged, cached, and served by the host. Plugin lifecycle and plugin state belong to `Plugins()`; tenant plugin governance and tenant-filter context belong to `Tenant()`. Host-internal scope helpers are not exposed through ordinary plugin-facing access.

### Unified Services and Dynamic Plugins

Plugin-visible domains no longer provide a separate management directory. Source plugins use `Services.<Domain>()` on the `capability.Services` directory returned by `pluginhost` inputs, or a narrowed injected `*cap.Service` for reads, writes, execution, and governance actions. Plugin lifecycle and enablement state are accessed through `Services.Plugins().Lifecycle()` and `Services.Plugins().State()`; tenant plugin governance and tenant filter context are accessed through `Services.Tenant().Plugins()` and `Services.Tenant().Filter()`. `capability.Services` does not expose a separate tenant service view or top-level `PluginLifecycle()`, `PluginState()`, `TenantPluginGovernance()`, `TenantFilter()`, or `TenantTableFilter()` methods. Same-process source plugins and host adapters that need to constrain GoFrame queries use `tenantspi.ApplyPluginTableFilter(ctx, Services.Tenant().Filter(), model, qualifier)` or derive the filter from an injected `tenantcap.Service.Filter()`; ordinary `tenantcap.Service` and dynamic plugins expose only serializable tenant-filter context. Dynamic plugins may call governed plugin and tenant methods only when those methods are registered, declared, authorized, and dispatched as ordinary domain `hostServices`.

Dynamic plugins may use only the concrete methods that are registered in the dynamic `host service` catalog, declared by the plugin manifest, authorized by the host, and handled by the `WASM host-service` dispatcher. If a governed action must become available to dynamic plugins, publish a narrow, versioned dynamic method in the unified domain instead of introducing a parallel management service.

| Domain capability | Responsibility boundary | Runtime and validation path |
| --- | --- | --- |
| `APIDoc` | Resolves route text and title operation keys in localized API documentation. | Served through the capability directory or `apidoc` host calls; validation focuses on route and operation-key payloads. |
| `Auth` | Handles tenant token selection or switching, impersonation tokens, permission projections, single permission checks, and batch permission checks through `Token()` and `Authz()` sub-capabilities. | Runtime checks use current user, tenant, permission keys, and method-level authorization. |
| `AI` | Aggregates typed AI sub-capabilities for text, image, embedding, audio, vision, document, safety, and video, including method-level status projections. | Calls are authorized by method, not resources, and plugin identity is scoped to provider routing and governance. Status reads expose only availability, reason, and public provider identity. |
| `Users` | Provides base user reads, bounded user listing, visibility enforcement, governed user writes, status and credential actions, and user-role assignments under `Users().Assignment()`. | Host implementations must keep user existence and visibility checks scoped to the caller. User-role assignment methods remain source/Go contract methods unless separately registered as dynamic `hostServices`; the current dynamic `users` service publishes only projection, batch, resolution, list, and visibility methods. |
| `BizCtx` | Projects the current business request context. | Used as a read-only runtime context bridge for request user, tenant, locale, and request metadata. |
| `Dict` | Resolves dictionary value labels, lists bounded value candidates, and validates typed value visibility. | Host validation stays within visible dictionary types and values. |
| `Files` | Provides host file-center projections, bounded search, visibility enforcement, content reads, governed uploads, and creation from plugin storage into `sys_file`. | Host validation prevents plugins from probing or using file IDs outside their visible boundary. Uploads create host file-center metadata while plugin-private objects remain owned by `Storage()`. |
| `HostConfig` | Reads host configuration values through the host priority chain and exposes governed `SysConfig()` projections and writes for `sys_config` rows. | Dynamic declarations must list `resources.keys` for `get`; `SysConfig()` writes stay method-governed and source-only unless separately published. This capability is separate from plugin-scoped business config. |
| `I18n` | Reads locale and translates messages for source plugins. | Source plugins receive this through `capability.Services` from `pluginhost` inputs; dynamic plugins do not receive an `i18n` host service because their i18n resources are host-managed. |
| `Jobs` | Reads scheduled-job metadata, searches bounded job candidates, validates job visibility, and performs governed runtime job actions. | Declaration-time job contracts are submitted through `pluginbridge.Declarations.Jobs().Register(...)`; runtime services do not expose `Register`. |
| `Notifications` | Lists and reads typed notification message projections, batch-loads messages by business source, validates visibility, deletes messages, updates read state, and sends governed notifications. | Actor-scoped read, delete, and read-state calls do not require resource declarations; `messages.send` requires a `resources[].ref` boundary. |
| `Plugins` | Exposes current plugin projection, plugin registry projections, tenant plugin pages, plugin enablement state, provider enablement state, plugin-scoped config, and governed plugin lifecycle orchestration. | Runtime checks cover plugin visibility, plugin enablement/provider state, plugin-scoped config source isolation, lifecycle preconditions, and dynamic `hostServices` authorization for published governance methods. |
| `Route` | Exposes dynamic-route metadata for the current execution. | Used by runtime route dispatch without exposing host router internals. |
| `Sessions` | Reads the current online-session projection, searches sessions, batch-loads session projections, validates session visibility, and batch-reads user online status. | Host validation keeps session and user visibility scoped to the caller. |
| `Storage` | Provides plugin-private object storage operations for plugin-owned attachments, binary objects, import/export temporaries, and uninstall cleanup, including explicit batch stat, cursor list, and batch delete. | Source plugins receive plugin-scoped `Storage()` through `capability.Services` from `pluginhost` inputs; dynamic declarations use `service: storage` with `resources.paths`; writes do not create `sys_file` rows or expose provider keys or local paths. |
| `Cache` | Provides plugin-scoped cache get, set, delete, multi-key get/set/delete, increment, and expiry operations. | Dynamic declarations use `resources[].ref`; runtime dispatch validates namespace, key, value size, and positive TTL payloads. Cache remains non-authoritative plugin runtime data. |
| `Lock` | Provides plugin-visible distributed lock acquire, renew, and release operations. | Dynamic declarations use `resources[].ref`; runtime dispatch validates lock resource and lease payloads. |
| `Manifest` | Reads plugin-scoped manifest or artifact resources, including bounded multi-get and path listing. | Dynamic declarations use `resources.paths`; source and dynamic paths are resolved through plugin-scoped resource views. |
| `Org` | Provides optional organization projections such as user organization profiles, bounded department trees, department search, post options, visibility checks, department assignments, department names, and post IDs. | Provider availability is explicit; fallback services return safe neutral values when the organization provider is absent. |
| `Tenant` | Provides optional tenant context, tenant info, tenant batch/search projections, visibility checks, membership validation, accessible tenant lists, tenant plugin governance, and plugin-owned table tenant-filter context. | Provider availability is explicit; lifecycle writes and membership replacement stay in the tenant owner or host-internal SPI. Same-process Go callers can pass the ordinary filter context to `tenantspi.ApplyPluginTableFilter(...)` when they own a GoFrame model. Dynamic plugins receive only serializable filter context and must use `RecordStore` or owner-side filters for tenant isolation. |

### Plugin Distribution

Plugin manifests and lifecycle callback snapshots include `distribution`, which is normalized by the host to `marketplace` or `builtin`. Omitted values are treated as `marketplace`. `builtin` is a source-plugin-only governance mode for project components compiled with the host: the host installs, enables, and safely upgrades them during startup, and ordinary plugin-management write actions are rejected. Dynamic plugins must remain `marketplace` and cannot self-declare built-in governance.

## Dynamic-Plugin-Only Capabilities

`Runtime`, `Network`, and `RecordStore` are dynamic-plugin-only entries on the `pluginbridge.Services` directory returned by `pluginbridge.Default()` or `pluginbridge.New()`. They are not part of `capability.Services` because source plugins either already run inside the host process with native equivalents or use source-plugin data access seams instead of guest host-service wrappers.

| Capability | Public entry | Boundary reason |
| --- | --- | --- |
| `Runtime` | `pluginbridge.Default().Runtime()` or `pluginbridge.New().Runtime()` | Dynamic plugins need a WASI host-service client for logs, state, time, UUIDs, and node identity; source plugins use host-native logging and runtime context directly. |
| `Network` | `pluginbridge.Default().Network()` or `pluginbridge.New().Network()` | Dynamic plugins need governed outbound HTTP through host-service authorization; source plugins use host-native HTTP clients or injected domain services. |
| `RecordStore` | `pluginbridge.Default().RecordStore()` or `pluginbridge.New().RecordStore()`, plus `pkg/plugin/pluginbridge/recordstore` | Dynamic plugins need a guest-side facade over the data host-service protocol and typed query plans; source plugins use their own DAO or provider seams. |

New capabilities should enter `capability.Services` only when source plugins and dynamic plugins share the same stable host-owned domain contract. Dynamic-only host-service clients and guest SDKs stay under `pluginbridge`.

### `Storage()` and `Files()` Boundary

| Scenario | Use | Boundary |
| --- | --- | --- |
| A plugin stores its own attachment, generated export, binary object, or temporary import file. | `Storage()` / dynamic `service: storage` | The plugin passes a logical object path. The host scopes it by plugin ID and tenant before delegating to the active storage provider. The object stays outside host file-center lists and does not create `sys_file` metadata. |
| A plugin deletes, lists, stats, or purges objects it owns during record deletion or uninstall cleanup. | `Storage()` / dynamic `service: storage` | Cleanup uses `Delete` or bounded `List` by logical prefix. Plugins must not delete host upload roots, provider roots, or file-center rows directly. |
| A plugin references files that users already uploaded into the host file center. | `Files()` / dynamic `service: files` | The plugin receives `filecap.FileInfo` values and visibility checks for host-owned file IDs. The response must not expose DAO, DO, Entity, provider object keys, or local absolute paths. |
| A plugin command accepts host file IDs from a request. | `Files().EnsureVisible` / `files.visible.ensure` | The command checks all IDs before mutation. Missing and invisible files share the same rejection semantics to avoid existence probing. |
| A plugin needs to upload content and register it in the host file center. | `Files().Upload` / `files.upload` | The host writes through the file owner so `sys_file` receives tenant, uploader, scene, hash, and storage metadata. Dynamic direct upload is bounded; larger dynamic payloads should use `Storage().Put` first. |
| A plugin has already written an object to its private storage and needs a host file-center record. | `Files().CreateFromStorage` / `files.create_from_storage` | The host copies from the plugin-scoped `Storage()` object into file-center storage. Dynamic plugins must also declare `storage.get` for the source path. The operation does not move or delete the source object and does not expose provider keys or local paths. |

`Storage()` provider selection is configuration-free. The host uses the only
enabled storage provider plugin when exactly one is serviceable, falls back to
the built-in local provider when none is serviceable, and rejects storage calls
when multiple provider plugins are serviceable.

## Plugin Configuration Sources

Source plugins use `Services.Plugins().Config()` and dynamic plugins use
`plugins.config.get` to read the current plugin's business configuration. These
entries are plugin-scoped and do not expose arbitrary host configuration keys or
sibling plugin configuration.

Configuration source priority is section-level:

| Priority | Source | Runtime behavior |
| --- | --- | --- |
| 1 | Host main static config section `plugin.<plugin-id>` | The whole section becomes the effective plugin config source. Missing subkeys return absent or caller defaults and do not fall back to file sources. |
| 2 | Production file `plugins/<plugin-id>/config.yaml` under the host config root | Used only when `plugin.<plugin-id>` is absent. |
| 3 | Development file `apps/lina-plugins/<plugin-id>/manifest/config/config.yaml` | Used only when host static and production file sources are absent. |
| 4 | Dynamic artifact default `manifest/config/config.yaml` | Used as the final fallback for the active dynamic plugin execution context. |

`manifest/config/config.example.yaml` is a template only and is never loaded as
runtime defaults. `HostConfig()` remains the separate host configuration
capability; non-root host keys read the current `sys_config` snapshot, active
static host config, host defaults, then absent `nil`. Dynamic `hostconfig.get`
calls still require `resources.keys` authorization in `hostServices`.
Source plugins call `HostConfig().Get(ctx, key, defaultValue)` with an explicit
default value; pass `nil` to preserve the absent-key `nil` result after the host
priority chain.

## Consumer Contracts, Provider SPI, and Guest SDK

Plugin-facing packages use three separate boundaries so each caller imports only
the contract it can safely depend on:

| Boundary | Package shape | Intended callers | Must not contain |
| --- | --- | --- | --- |
| Ordinary consumer contract | `pkg/plugin/capability/<domain>cap` | Source plugins through `capability.Services` from `pluginhost` inputs, dynamic plugins through generated or bridge-backed clients, and host adapters | GoFrame HTTP request objects, provider factory registration, host-private implementation state, or GoFrame database builders |
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

Source plugins access runtime capabilities through the `capability.Services` directory returned by `pluginhost` inputs. Plugin lifecycle and state are exposed through `Services.Plugins().Lifecycle()` and `Services.Plugins().State()`. Tenant plugin governance and tenant filtering context are exposed through `Services.Tenant().Plugins()` and `Services.Tenant().Filter()`. Same-process table filtering is an explicit `tenantspi` helper call, not a separate source-plugin service-directory method or tenant-service mirror.

Dynamic plugins access published runtime capabilities through `pluginbridge.Services`. Calls are encoded through `pluginbridge/protocol`, transported through `WASI host call`, authorized by derived `HostCapabilities` and confirmed `HostServices`, then dispatched by `apps/lina-core/internal/service/plugin/internal/wasm`. Dynamic plugins do not receive the top-level compatibility shortcuts or `I18n()`, but they can call the registered governance methods under `Plugins()` and `Tenant()` when authorized by `hostServices`.

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
| `auth` | None | `host:auth:token`<br/>`host:auth:authz` | `token.tenant.select`<br/>`token.tenant.switch`<br/>`token.impersonation_token.issue`<br/>`token.impersonation_token.revoke`<br/>`authz.permissions.batch_get`<br/>`authz.permissions.batch_has`<br/>`authz.permissions.has`<br/>`authz.users.platform_admin.check` |
| `ai` | None | `host:ai:text`<br/>`host:ai`<br/>`host:ai:image`<br/>`host:ai:embedding`<br/>`host:ai:audio`<br/>`host:ai:vision`<br/>`host:ai:document`<br/>`host:ai:safety`<br/>`host:ai:video` | `text.generate`<br/>`text.method_status.get`<br/>`ai.methods.status.batch_get`<br/>`image.generate`<br/>`image.edit`<br/>`embedding.create`<br/>`audio.transcribe`<br/>`audio.synthesize`<br/>`vision.analyze`<br/>`document.analyze`<br/>`document.cite`<br/>`safety.moderate`<br/>`video.generate`<br/>`video.edit`<br/>`video.extend`<br/>`video.operation.get`<br/>`video.operation.cancel` |
| `users` | None | `host:users` | `users.current.get`<br/>`users.batch_get`<br/>`users.resolve.batch`<br/>`users.list`<br/>`users.visible.ensure` |
| `bizctx` | None | `host:bizctx` | `current.get` |
| `dict` | None | `host:dict` | `dict.value.labels.resolve`<br/>`dict.value.list`<br/>`dict.value.values.visible.ensure` |
| `files` | None | `host:files` | `files.batch_get`<br/>`files.list`<br/>`files.visible.ensure`<br/>`files.upload`<br/>`files.create_from_storage` |
| `jobs` | None | `host:jobs` | `jobs.batch_get`<br/>`jobs.list`<br/>`jobs.visible.ensure`<br/>`jobs.register` |
| `notifications` | None except `messages.send`, which uses `resources[].ref` | `host:notifications` | `messages.batch_get`<br/>`messages.list`<br/>`messages.by_source.batch_get`<br/>`messages.visible.ensure`<br/>`messages.send`<br/>`messages.delete`<br/>`messages.by_source.delete`<br/>`messages.mark_read`<br/>`messages.mark_unread` |
| `plugins` | None | `host:plugins` | `plugins.current.get`<br/>`plugins.batch_get`<br/>`plugins.registry.list`<br/>`plugins.tenant.list`<br/>`config.get`<br/>`plugins.state.enabled.check`<br/>`plugins.state.provider_enabled.check`<br/>`plugins.state.enabled_authoritative.check`<br/>`plugins.lifecycle.tenant_plugin_disable.ensure`<br/>`plugins.lifecycle.tenant_plugin_disabled.notify`<br/>`plugins.lifecycle.tenant_delete.ensure`<br/>`plugins.lifecycle.tenant_deleted.notify` |
| `route` | None | `host:route` | `metadata.get` |
| `sessions` | None | `host:sessions` | `sessions.current.get`<br/>`sessions.list`<br/>`sessions.batch_get`<br/>`sessions.users.online.batch_get`<br/>`sessions.visible.ensure` |
| `org` | None | `host:org` | `capability.available`<br/>`capability.status`<br/>`org.assignment.user_profiles.batch_get`<br/>`org.department.tree.list`<br/>`org.department.batch_get`<br/>`org.department.list`<br/>`org.post.batch_get`<br/>`org.post.options.list`<br/>`org.department.visible.ensure_many`<br/>`org.post.visible.ensure_many` |
| `tenant` | None | `host:tenant` | `capability.available`<br/>`capability.status`<br/>`tenant.context.current`<br/>`tenant.context.info`<br/>`tenant.context.platform_bypass`<br/>`tenant.directory.batch_get`<br/>`tenant.directory.list`<br/>`tenant.membership.validate`<br/>`tenant.membership.list_by_user`<br/>`tenant.directory.visible.ensure_many`<br/>`tenant.plugins.enabled.set`<br/>`tenant.plugins.defaults.provision`<br/>`tenant.filter.context` |
<!-- END generated:host-services -->

## Maintenance Notes

When plugin public contracts or dynamic `host service` descriptors change, update both `README.md` and `README.zh-CN.md` in this directory.
