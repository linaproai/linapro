# Plugin Package

`apps/lina-core/pkg/plugin` is the public `Go` contract surface for the `Lina Core` plugin system. It keeps plugin-facing contracts in one place while runtime implementations, persistence, lifecycle orchestration, and workbench adapters stay behind host-owned services.

The package has three main jobs:

- Publish stable host capability interfaces for source plugins and dynamic plugins.
- Publish source-plugin registration contracts for routes, hooks, cron jobs, lifecycle callbacks, embedded assets, and governance filters.
- Publish dynamic-plugin bridge contracts for `Wasm` route execution, guest runtime helpers, artifact metadata, and governed `hostServices` calls declared in `plugin.yaml`.

## Boundary

This package is a contract boundary, not a business implementation layer. It must not expose host `DAO`, `DO`, `Entity`, raw `gdb.Model` builders to ordinary plugins, concrete workbench page structures, or host implementation packages.

Source plugins usually use `pluginhost` plus the typed services returned from `pluginhost.Services`. Dynamic plugins usually use `pluginbridge/guest` for guest-side host-service clients and low-level route execution helpers.

## Directory Structure

| Path | Responsibility |
|------|----------------|
| `capability/` | Stable host capability directory shared by source plugins and dynamic plugins. It owns aggregate `Services`, `AdminServices`, plugin-scoped service binding, and typed domain contracts. |
| `pluginbridge/` | Dynamic-plugin bridge namespace. It owns public protocol facades, guest runtime helpers, bridge contracts, and codecs for `Wasm` host calls. |
| `pluginhost/` | Source-plugin host namespace. It owns compile-time registration facades and runtime callback contracts for embedded source plugins. |

## `capability` Packages

`capability` defines the stable host capability directory. Host runtime services implement these interfaces, while plugin code should depend on these narrow contracts.

| Path | Responsibility |
|------|----------------|
| `capability/` | Aggregate `Services`, `AdminServices`, and `ServicesForPlugin` binding for plugin-scoped views. |
| `capability/aicap/` | Root `AI` namespace that aggregates typed `AI` sub capabilities and injects source plugin identity with `ForPlugin`. |
| `capability/aicap/aicommon/` | Shared `AI` value objects, capability types, methods, tiers, asset references, provider projections, operation references, usage, and unavailable status helpers. |
| `capability/aicap/aitext/` | Text generation capability contract, provider contract, fallback behavior, and source identity binding. |
| `capability/aicap/aiimage/` | Image generation and image editing capability contract. |
| `capability/aicap/aiembedding/` | Embedding creation capability contract. |
| `capability/aicap/aiaudio/` | Audio transcription and synthesis capability contract. |
| `capability/aicap/aivision/` | Vision analysis capability contract for images, screenshots, diagrams, and frames. |
| `capability/aicap/aidocument/` | Document analysis and citation-aware document capability contract. |
| `capability/aicap/aisafety/` | Safety moderation capability contract for text and asset inputs. |
| `capability/aicap/aivideo/` | Video generation, editing, extension, and provider operation capability contract. |
| `capability/apidoccap/` | API-documentation text lookup and route operation-key helpers for source plugins and dynamic routes. |
| `capability/authcap/` | Authentication and authorization namespace that aggregates token and authorization sub capabilities. |
| `capability/authcap/authz/` | Authorization-domain read and management contracts without exposing host role, menu, or permission tables. |
| `capability/authcap/token/` | Tenant token selection, tenant switching, and impersonation token handoff contracts without exposing host JWT internals. |
| `capability/bizctxcap/` | Read-only business context projection for user, tenant, impersonation, and platform-bypass state. |
| `capability/cachecap/` | Plugin-scoped runtime cache contract with string and integer value primitives. |
| `capability/capmodel/` | Shared storage-agnostic domain primitives such as `CapabilityContext`, actor/source metadata, batch results, page results, and localized labels. |
| `capability/configcap/` | Governed runtime configuration capability contracts for reading and managing host-owned config projections. |
| `capability/dictcap/` | Dictionary-domain label resolution and refresh contracts without exposing dictionary tables. |
| `capability/filecap/` | Governed file projections and file deletion contracts without exposing physical paths or storage tables. |
| `capability/hostconfigcap/` | Read-only host configuration access for authorized source plugins and dynamic `hostconfig` calls. |
| `capability/i18ncap/` | Runtime locale and translation lookup contracts for source plugins. |
| `capability/infracap/` | Infrastructure status projections and refresh contracts without leaking concrete runtime backends. |
| `capability/jobcap/` | Scheduled-job projections and governed job execution or status management contracts. |
| `capability/manifestcap/` | Plugin-scoped read-only access to resources under `manifest/` for source and dynamic plugins. |
| `capability/notifycap/` | Notification projections and governed message send/delete contracts without exposing notification tables. |
| `capability/orgcap/` | Optional organization capability, provider registration, user-department/post projections, and organization scope seams. |
| `capability/plugincap/` | Plugin-governance projections plus plugin-local config, state, lifecycle, registry, and tenant-default management contracts. |
| `capability/recordstore/` | Guest-side governed ORM-style facade over authorized dynamic-plugin `data` service tables. |
| `capability/routecap/` | Dynamic route metadata projection attached to the current request. |
| `capability/sessioncap/` | Online-session search, batch read, and revocation contracts without exposing session storage. |
| `capability/tenantcap/` | Optional tenant capability, provider registration, tenant resolution, visibility checks, tenant switching, and tenant scope seams. |
| `capability/usercap/` | User-domain projections, search, visibility checks, and governed status management without exposing `sys_user`. |

## `pluginbridge` Responsibilities

`pluginbridge` is the dynamic-plugin protocol area. It is used by the host runtime, the dynamic plugin builder, and `Wasm` guest code.

| Path | Responsibility |
|------|----------------|
| `pluginbridge/` | Namespace package documenting that bridge protocol and guest helpers live in subpackages. |
| `pluginbridge/contract/` | Stable bridge `ABI` constants, route contracts, request/response envelopes, identity snapshots, execution-source values, lifecycle contracts, cron contracts, and validation helpers. |
| `pluginbridge/guest/` | Guest runtime helpers for exported allocation/execution functions, route dispatch, request binding, router matching, guest-side host-service clients, and raw host-call transport. |
| `pluginbridge/protocol/` | Public low-level protocol facade for bridge envelopes, artifact metadata, host-call payloads, `hostServices` payloads, codecs, capability constants, and the public `hostServices` catalog. |

Use `pluginbridge/protocol.HostServiceDescriptors()` when tooling needs the language-neutral `hostServices` catalog. It is the public source developers and tooling should consume for service and method discovery.

## `pluginhost` Responsibilities

`pluginhost` is the source-plugin contract area. Source plugins are compiled into the host and register their backend contribution through grouped facades.

| Area | Responsibility |
|------|----------------|
| `SourcePlugin` | Root grouped source-plugin registration contract. It exposes `Assets()`, `Lifecycle()`, `Hooks()`, `HTTP()`, `Cron()`, and `Governance()`. |
| `Services` | Runtime service directory passed to source-plugin registration and callbacks. It embeds ordinary `capability.Services` and adds source-plugin-only `Admin()` and `TenantFilter()` seams. |
| Asset registration | `UseEmbeddedFiles` binds plugin-owned embedded files so the host can serve manifest and public assets. |
| Lifecycle registration | Registers precondition, custom upgrade, cleanup, and post-notification callbacks for install, upgrade, disable, uninstall, tenant disable/delete, and install-mode changes. |
| Hook registration | Registers callback-style hook handlers for published backend extension points. |
| HTTP registration | Registers source-plugin routes under the plugin API namespace and captures route bindings for host governance. |
| Cron registration | Registers guarded cron jobs that check plugin enablement at execution time and expose primary-node inspection. |
| Governance registration | Registers menu and permission filters used by the host governance pipeline. |

Published backend extension points include `auth.login.succeeded`, `auth.login.failed`, `auth.logout.succeeded`, `plugin.installed`, `plugin.enabled`, `plugin.disabled`, `plugin.uninstalled`, `plugin.upgraded`, `system.started`, `http.route.register`, `cron.register`, `menu.filter`, and `permission.filter`.

## `plugin.yaml` `hostServices`

Dynamic plugins declare host access in `plugin.yaml` with `hostServices`. Each declaration uses a service name, method names, and the resource declaration shape required by that service.

Example:

```yaml
hostServices:
  - service: runtime
    methods:
      - log.write
      - state.get
      - state.set
  - service: storage
    methods:
      - put
      - get
      - list
    paths:
      - exports/
  - service: data
    methods:
      - list
      - get
    tables:
      - plugin_demo_reports
  - service: network
    methods:
      - request
    resources:
      - ref: https://api.example.com/v1/*
  - service: ai
    methods:
      - text.generate
    resources:
      - ref: purpose:report.summary
        attributes:
          defaultTier: standard
```

Resource declaration shapes:

| Resource kind | Declaration fields | Services |
|---------------|--------------------|----------|
| `none` | No `paths`, `tables`, `keys`, or `resources`. | `runtime`, `cron`, `config`, `org`, `tenant` |
| `path` | `paths` | `storage`, `manifest` |
| `table` | `tables` | `data` |
| `key` | `keys` | `hostconfig` |
| `resource` | `resources[].ref` plus service-specific governance fields. | `network`, `cache`, `lock`, `notify`, `ai` |

`data` service tables are plugin-owned in production validation. A dynamic plugin must not declare host core tables such as `sys_*`.

`network` resources use authorized `http` or `https` URL patterns. `ai` resources use `purpose:<name>` refs and may include governed attributes such as `defaultTier`, `maxOutputTokens`, `maxPayloadBytes`, `maxInputAssets`, `maxOutputAssets`, `maxAssetBytes`, `allowOperation`, `allowOperationCancel`, and `allowedMimeTypes`.

`config`, `hostconfig`, and `manifest` default to `get` when methods are omitted. The dynamic guest config helpers such as `Exists`, `String`, `Bool`, `Int`, and `Duration` map through `config.get`; `plugin.yaml` should still declare `config.get`.

## Dynamic Plugin `hostServices` Catalog

This chapter lists the `hostServices` service names dynamic plugins can declare in `plugin.yaml`, the methods under each service, and the purpose of each method. The machine-readable public source is `pluginbridge/protocol.HostServiceDescriptors()`.

### Service Summary

| Service | Resource kind | Purpose |
|---------|---------------|---------|
| `runtime` | `none` | Runtime logs, plugin-scoped state, host time, host-generated IDs, and node identity. |
| `cron` | `none` | Dynamic-plugin cron declaration during host-side discovery. |
| `storage` | `path` | Governed plugin storage object operations under authorized logical paths. |
| `network` | `resource` | Governed outbound `HTTP` requests to authorized URL patterns. |
| `data` | `table` | Governed reads and mutations against plugin-owned authorized tables. |
| `cache` | `resource` | Governed cache reads, writes, integer increments, and expiration updates. |
| `lock` | `resource` | Governed distributed lock acquire, renew, and release operations. |
| `notify` | `resource` | Governed notification message dispatch. |
| `config` | `none` | Read-only access to the current plugin runtime configuration. |
| `hostconfig` | `key` | Read-only access to explicitly authorized host configuration keys. |
| `manifest` | `path` | Read-only access to plugin-scoped resources under `manifest/`. |
| `ai` | `resource` | Governed typed `AI` calls authorized by `purpose:<name>` resources. |
| `org` | `none` | Organization capability status and user organization projections. |
| `tenant` | `none` | Tenant capability status, current tenant, visibility, membership, and switch checks. |

### Method Details

#### `runtime`

| Method | Purpose |
|--------|---------|
| `log.write` | Write one structured runtime log entry for the plugin. |
| `state.get` | Read one plugin-scoped runtime state value. |
| `state.set` | Write one plugin-scoped runtime state value. |
| `state.delete` | Delete one plugin-scoped runtime state value. |
| `info.now` | Return host time information. |
| `info.uuid` | Return one host-generated unique identifier. |
| `info.node` | Return host node identity information. |

#### `cron`

| Method | Purpose |
|--------|---------|
| `register` | Register one dynamic-plugin cron contract with the current host-side discovery collector. |

#### `storage`

| Method | Purpose |
|--------|---------|
| `put` | Write one governed storage object. |
| `get` | Read one governed storage object. |
| `delete` | Delete one governed storage object. |
| `list` | List governed storage objects under one authorized prefix. |
| `stat` | Read metadata for one governed storage object. |

#### `network`

| Method | Purpose |
|--------|---------|
| `request` | Execute one governed outbound `HTTP` request. |

#### `data`

| Method | Purpose |
|--------|---------|
| `list` | Execute one governed paged list query against an authorized plugin-owned table. |
| `get` | Read one governed record by key from an authorized plugin-owned table. |
| `create` | Create one governed record in an authorized plugin-owned table. |
| `update` | Update one governed record in an authorized plugin-owned table. |
| `delete` | Delete one governed record in an authorized plugin-owned table. |
| `transaction` | Execute one governed transaction over structured data mutations. |

#### `cache`

| Method | Purpose |
|--------|---------|
| `get` | Read one governed cache value. |
| `set` | Write one governed cache value. |
| `delete` | Remove one governed cache value. |
| `incr` | Increment one governed cache integer value. |
| `expire` | Update one governed cache expiration policy. |

#### `lock`

| Method | Purpose |
|--------|---------|
| `acquire` | Acquire one governed distributed lock. |
| `renew` | Renew one governed distributed lock. |
| `release` | Release one governed distributed lock. |

#### `notify`

| Method | Purpose |
|--------|---------|
| `send` | Send one governed notification message. |

#### `config`

| Method | Purpose |
|--------|---------|
| `get` | Read one current-plugin configuration value as `JSON`. |

`config` only publishes `get` in `plugin.yaml`. Guest helpers such as `Exists`, `String`, `Bool`, `Int`, and `Duration` are convenience adapters over `config.get`.

#### `hostconfig`

| Method | Purpose |
|--------|---------|
| `get` | Read one authorized host configuration value. |

#### `manifest`

| Method | Purpose |
|--------|---------|
| `get` | Read one plugin-scoped manifest resource. |

#### `ai`

| Method | Purpose |
|--------|---------|
| `text.generate` | Execute one governed text generation request. |
| `image.generate` | Execute one governed image generation request. |
| `image.edit` | Execute one governed image editing request. |
| `embedding.create` | Execute one governed embedding request. |
| `audio.transcribe` | Execute one governed audio transcription request. |
| `audio.synthesize` | Execute one governed audio synthesis request. |
| `vision.analyze` | Execute one governed visual analysis request. |
| `document.analyze` | Execute one governed document analysis request. |
| `document.cite` | Execute one governed citation-aware document request. |
| `safety.moderate` | Execute one governed safety moderation request. |
| `video.generate` | Execute one governed video generation request. |
| `video.edit` | Execute one governed video editing request. |
| `video.extend` | Execute one governed video extension request. |
| `video.operation.get` | Read one governed provider operation. |
| `video.operation.cancel` | Cancel one governed provider operation. |

#### `org`

| Method | Purpose |
|--------|---------|
| `capability.available` | Report whether the organization capability is available. |
| `capability.status` | Read organization capability status. |
| `users.dept_assignments.list` | List user department assignments in batch. |
| `users.dept_info.get` | Read one user's department identifier and name. |
| `users.dept_name.get` | Read one user's department name. |
| `users.dept_ids.get` | Read one user's department identifiers. |
| `users.post_ids.get` | Read one user's post identifiers. |

#### `tenant`

| Method | Purpose |
|--------|---------|
| `capability.available` | Report whether the tenant capability is available. |
| `capability.status` | Read tenant capability status. |
| `tenants.current` | Read the current request tenant. |
| `tenants.platform_bypass` | Report whether tenant filtering may be bypassed. |
| `tenants.visible.ensure` | Validate that the current user can access one tenant. |
| `users.tenant_membership.validate` | Validate one user's tenant membership. |
| `users.tenants.list` | List tenants visible to one user. |
| `tenants.switch.validate` | Validate one tenant switch target. |

Reserved governance entries currently exist for `secret.resolve`, `event.publish`, and `queue.enqueue`. They are part of the descriptor catalog for future governance alignment, but they are not published guest-callable methods and should not be used for executable dynamic plugin calls until host dispatcher and guest SDK support is added.

## Developer Guide

- Use `capability.Services` when a source plugin or host package needs ordinary read-oriented plugin-facing capabilities.
- Use `capability.AdminServices` only for trusted source-plugin management commands. Keep those dependencies narrow and pass `CapabilityContext` through domain services.
- Use `pluginbridge/guest.Default()` or `pluginbridge/guest.New()` inside dynamic plugin guest code for `runtime`, `storage`, `network`, `recordstore`, `cache`, `lock`, `config`, `notify`, `cron`, `hostconfig`, `manifest`, `org`, `tenant`, `plugin`, and `AI` clients.
- Use `pluginhost.SourcePlugin` from source-plugin registrars to declare embedded files, routes, lifecycle callbacks, hooks, cron jobs, and governance filters.
- Use `pluginbridge/contract` for dynamic route and artifact contracts, `pluginbridge/guest` for guest route execution, and `pluginbridge/protocol` for low-level payloads or tooling that needs the public protocol catalog.
- Keep new capability contracts storage-agnostic, batch-oriented for high-volume reads, and explicit about data visibility. Do not add a capability method that forces plugins to know host table names or host cache keys.
