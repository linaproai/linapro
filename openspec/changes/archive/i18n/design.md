## Context

LinaPro is positioned as an AI-driven full-stack development framework. The frontend already had `vue-i18n` base with static locale bundles, and workspace requests carried `Accept-Language`, but the host lacked framework-level i18n capability. Menus, dictionaries, system parameter metadata, public login-page configuration, system information, plugin `plugin.yaml` metadata, backend errors, import/export content, and many backend-returned labels were still stored and returned as single-language text. The `Service` interface had 18 methods carrying multiple responsibility categories. Business modules independently decided projection rules. Resource loading was duplicated between apidoc and runtime bundles. WASM parsing leaked into the i18n package. Backend code contained widespread hardcoded Chinese strings. The repository documentation described LinaPro as a "backend management system" rather than a framework.

Since this is a new project, no historical compatibility layer is required. This design directly establishes unified i18n infrastructure, performance optimization, message governance, and project positioning without preserving long-term dual-track behavior.

## I18n Infrastructure

### Three-Layer Model

Internationalized text is split by source and lifecycle:

- **Static UI copy**: Frontend framework and page-local static copy continues to be maintained by local `vue-i18n` JSON.
- **Dynamic metadata**: Menus, dictionaries, system parameter metadata, system information, plugin descriptions, and similar metadata are projected by the backend according to locale.
- **Business content**: Business modules that need multilingual titles, descriptions, or body content adopt the common multilingual content model within their own module boundaries.

Static UI copy and dynamic metadata have different sources. Mixing them blurs maintenance boundaries. Menus, dictionaries, and system configuration metadata are backend-governed data and should be projected by the backend rather than hard-coded across frontend combinations.

### Resource Governance Model

i18n resources use files as the single source of truth. The host and delivered projects maintain base locale resource files. Plugins declare translation files in their own directories. The only required change for adding a new language is corresponding JSON resources in host/plugin/frontend, with optional changes being `i18n` metadata in the default config file.

Runtime persistence tables (`sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content`) were initially designed but later removed. Database overrides create a dual-source truth, making auditing, missing checks, and delivery write-back more complex. The lowest-risk path for adding new languages is "supplement resources + optional YAML metadata", not modifying SQL, backend Go, frontend TS and cache invalidation strategy.

Resource files are organized by locale directory and semantic domain:

```text
manifest/i18n/
  en-US/
    framework.json    # Framework name, description, language name, common display text
    menu.json         # Menu, route titles, navigation text
    dict.json         # Dictionary type names, item labels, enum display text
    config.json       # System config names, descriptions, groups
    error.json        # bizerr business errors, validation errors
    artifact.json     # Import/export headers, templates, failure reasons
    public-frontend.json  # Login page, public pages, default console text
    apidoc/
      common.json
      core-api-user.json
```

Source plugins use the same layout with resources owned by the plugin directory. Runtime keys are governed by existing namespaces: `error.<domain>.<case>`, `artifact.<module>.<section>.<field>`, `menu.<menu_key>.title`, `dict.<dict_type>.name`, `config.<config_key>.name`, `plugin.<plugin_id>.name`, etc.

### Locale Resolution

A host `LocaleResolver` and request-context injection mechanism resolves locale with priority:

1. Query parameter `lang`
2. `Accept-Language` request header
3. System default locale (from `i18n.default` in default config file)

The resolved `locale` is available to controllers, services, and plugin host bridges. When `i18n.enabled=false`, the host only accepts the default language, the frontend hides the language switch button and loads messages in the default language.

### Translation Key Conventions

Translation keys are derived from stable business keys:

- Menu title: `menu.<menu_key>.title`
- Dictionary type name: `dict.<dict_type>.name`
- Dictionary label: `dict.<dict_type>.<value>.label`
- Config name/remark: `config.<config_key>.name`, `config.<config_key>.remark`
- Plugin name/description: `plugin.<plugin_id>.name`, `plugin.<plugin_id>.description`
- System information: `systemInfo.component.<section>.<name>.description`
- Public frontend config: `publicFrontend.<group>.<field>`
- Config field headers: `config.field.<name>`
- Error messages: `error.<domain>.<case>`
- Role names: `role.builtin.<key>.name`

Runtime UI resource files may be authored as nested JSON or flat dotted keys; the host normalizes them to flat keys for governance. API endpoints return nested message objects for frontend consumption.

### Language Discovery and Configuration

Built-in languages are auto-discovered from host `manifest/i18n/<locale>/*.json` files. The `i18n` configuration section in the default config file maintains metadata that cannot be derived from filenames:

```yaml
i18n:
  default: zh-CN
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文
    - locale: zh-TW
      nativeName: 繁體中文
```

Adding a new built-in language only requires adding corresponding JSON resources and optional config metadata. No backend Go constants, SQL seeds, or frontend TS language lists need modification. Third-party library locales (dayjs, antd, vxe) are derived by language code convention. Document direction is fixed to `ltr` per current host convention.

### Service Interface

The `i18n.Service` large interface is split into four smaller interfaces by responsibility:

- `LocaleResolver`: Resolves request language and context language.
- `Translator`: Provides translation lookup and error localization (`Translate`, `TranslateSourceText`, `TranslateOrKey`, `TranslateWithDefaultLocale`, `LocalizeError`).
- `BundleProvider`: Outputs runtime translation bundles and language lists (`BuildRuntimeMessages`, `ListRuntimeLocales`, `BundleVersion`).
- `Maintainer`: Provides resource export, missing checks, source diagnostics, and cache invalidation (`ExportMessages`, `CheckMissingMessages`, `DiagnoseMessages`, `InvalidateRuntimeBundleCache`).

`serviceImpl` implements all four interfaces. Business modules' `i18nSvc` field types converge to the minimum interface they actually depend on (typically `LocaleResolver + Translator`).

### ResourceLoader

A shared `ResourceLoader` component in `pkg/i18nresource/` accepts `Subdir`, `LocaleSubdir`, `PluginScope`, `LayoutMode`, and `ValueMode` configuration parameters, centrally implementing the discovery and loading logic for host embedded resources, source plugin resources, and dynamic plugin resources. Both runtime UI translation resource loading and API documentation translation resource loading are completed through different `ResourceLoader` instances, eliminating ~280 lines of duplicate code while preventing apidoc from reverse-depending on `internal/service/i18n`.

### Plugin I18n Integration

Plugins add a standard `manifest/i18n/<locale>/` resource directory. The host handles plugin i18n resources during source-plugin synchronization and discovery, dynamic plugin installation or upgrade (where resource snapshots are written for the release), plugin enablement (where resources join runtime message aggregation), and plugin disablement or uninstallation (where resources are removed from runtime aggregation). Plugin i18n resources are delivered with the plugin version and managed consistently by the host.

## Performance Optimization

### Zero-Copy Hot Path

`Translate` / `TranslateSourceText` / `TranslateOrKey` / `TranslateWithDefaultLocale` directly hold a read lock on `runtimeBundleCache` and look up values, no longer going through `buildRuntimeMessageCatalog -> cloneFlatMessageMap`. Only `BuildRuntimeMessages` (output to frontend) and `ExportMessages` retain clone semantics. 99% of `Translate` call paths only read 1 key; cloning the entire 800+ key map was pure waste. Concurrent read-only access to `map[string]string` is safe with `sync.RWMutex`.

### Layered Cache

The `runtimeBundleCache` is refactored to a layered structure by `locale x sector`:

- `host`: Immutable, loaded once at startup.
- `plugins`: Source plugins, refreshed when source registry changes.
- `dynamic`: Dynamic plugins, refreshed by plugin lifecycle hooks.
- `merged`: Merged view by priority, invalidated when any sub-layer changes.

Invalidation granularity: dynamic plugin enable/disable only clears that plugin ID's dynamic sub-layer; source plugin registry changes invalidate the plugins layer; resource metadata or test-triggered invalidation clears by locale/sector. Each invalidation triggers `version.Add(1)`, driving frontend ETag negotiation.

### ETag and 304 Negotiation

The `/i18n/runtime/messages` API outputs `ETag: "<locale>-<bundleVersion>"` and `Cache-Control: private, must-revalidate`. When the request carries `If-None-Match` matching the current ETag, the API returns `304 Not Modified` with no message body. The backend maintains a `bundleVersion` atomic counter that auto-increments on any invalidation.

### Frontend Persistent Cache

The frontend `runtime-i18n.ts` switches from raw `fetch` to `requestClient`, integrating auth/error/degradation chain. It persists `{locale, etag, messages, savedAt}` to `localStorage` with a 7-day TTL. On subsequent page loads or language switches, persistent data renders immediately, then asynchronously negotiates with `If-None-Match` in the background. On 304 hit, persistent data stays unchanged. This enables zero-network language switching.

### Load Messages Split

Frontend `loadMessages` handles three things independently by failure semantics:

1. Runtime bundle failure -> hit persistent cache or fallback; user notification via degradation mechanism.
2. Public config failure -> fire-and-forget without blocking.
3. Third-party library locale -> must await (dayjs, antd, vxe locale modules).

## Message Governance

### Message Classification

Runtime messages are classified into six categories:

- `UserMessage`: API errors, business validation, frontend toasts, admin result prompts -- must use error codes or translation keys.
- `UserArtifact`: Excel headers, sheet names, import template examples, import failure reasons, export enum values -- must render by request language.
- `UserProjection`: Menus, dictionaries, roles, tasks, audit/operation logs and other backend-owned display data -- must use stable business keys for projection.
- `DeveloperDiagnostic`: Plugin protocol, WASM host call, manifest validation, CLI diagnostics -- must have stable machine codes, default English source messages.
- `OpsLog`: Server logs and metrics -- use English and structured fields, do not participate in runtime i18n.
- `UserData`: User input, external system content, database business values -- preserved and returned as-is.

### Structured Error Model

The `bizerr` package in `apps/lina-core/pkg/bizerr` provides unified business error construction. Each component defines business errors centrally in a dedicated `*_code.go` file using `bizerr.MustDefine`. Business code only references component error definitions:

```go
var CodeDictTypeExists = bizerr.MustDefine(
    "DICT_TYPE_EXISTS",
    "Dictionary type already exists",
    gcode.CodeInvalidParameter,
)
```

The unified response middleware identifies structured error metadata and outputs:

```json
{
  "code": 65,
  "message": "Dictionary type already exists",
  "errorCode": "DICT_TYPE_EXISTS",
  "messageKey": "error.dict.type.exists",
  "messageParams": {},
  "data": null
}
```

`code` is the GoFrame type error code expressing the error category. `message` is the display text resolved by the server according to request language. `errorCode` and `messageKey` are for frontend, tests, plugins, or third-party callers to make stable business semantic judgments. Business semantic identifiers are governed by module namespace: host modules use `<MODULE>_<CASE>` format, plugin modules use `<PLUGIN>_<MODULE>_<CASE>` format.

### Import/Export Localization

Import/export flows parse the current locale at the request level and reuse translation results. Batch pre-fetching makes localization cost proportional to field count, not row count. The `excelutil` package continues to only handle Excel file operations; the business service is responsible for passing in localized sheet names, headers, enum text, and failure reasons.

### Plugin Error Contracts

Plugin bridging protocol, WASM host call, and host service protocol return stable status codes and error codes. Default error source messages use English developer diagnostics. When these errors enter the admin UI, they are rendered using `messageKey` and locale. Dynamic plugin guest-returned JSON errors support `errorCode/messageKey/messageParams/message`, and the host preserves structured fields when passing through.

### Frontend Error Handling

The default console request error handling priority is:

1. If the backend returns `messageKey`, the frontend renders using `$t(messageKey, messageParams)`.
2. Otherwise, use the backend's `message` already localized by request language.
3. Otherwise, use the request library's default error text.

Page-level messages must use `$t` or runtime language packs. Directly writing Chinese or mixed Chinese-English strings in user-visible locations is prohibited.

### Scanning and Governance

A Go tool under `hack/tools/runtime-i18n` maintains runtime message scanning and language pack coverage checks. The tool uses auditable rule patterns to identify high-risk positions:

- Go: `gerror.New*`, `gerror.Wrap*`, `panic(gerror...)`, `Reason/Message/Fallback` fields, export header arrays, status text mappings, plugin bridging error construction.
- Vue/TypeScript: `title/label/placeholder`, template text nodes, `message.*`, `notification.*`, `Modal.confirm`, table column definitions.
- Allowlist: comments, test fixtures, user example data, technical units, protocol constants, English operations logs.

### Log and Audit Boundaries

Operations logs use stable English and structured fields, recording `errorCode/messageKey` and key parameters. They do not depend on the current request language. Operation logs, login logs, task logs, and plugin upgrade results store stable codes and parameters, and are projected by request language for lists, details, and exports.

## Module Boundaries

### Business Module Projection Ownership

Business modules maintain localization projection rules within their own module boundaries. `internal/service/i18n` only provides foundational capabilities such as language resolution, translation lookup, resource loading, caching, and missing checks. It does not reference business entities, business protection rules, or business translation key derivation logic.

Each business module owns its `*_i18n.go` or equivalent file:
- `menu_i18n.go`: Menu translation key derivation and projection.
- `dict_i18n.go`: Dictionary default-language skip strategy and `dict.*` key conventions.
- `sysconfig_i18n.go`: Config projection and field header translation keys.
- `jobmgmt_i18n.go`: Built-in tasks and default task group protection rules.
- `role.go`: Built-in admin role projection rules.
- Plugin runtime: Plugin metadata projection rules.

The initial `LocaleProjector` centralized approach was rejected because it would cause the i18n foundation service to reverse-couple with business entities and business protection rules, violating core host boundary and module decoupling principles.

### Source-Text Namespace Registration

`RegisterSourceTextNamespace(prefix, reason string)` provides explicit registration for code-owned namespaces. Business modules register their namespaces in their own `init()`. Missing translation checks, override source diagnostics, and import/export identify "namespaces whose translation keys are owned by code sources" by querying this registry. The i18n package does not hardcode any specific business module's namespace prefix.

### Projection Rules

For business master data editable in the management workspace (departments, posts, roles, dictionaries, parameters, notices, scheduler data), management lists, details, edit backfill, and selectors keep database values by default. Unless a specific field has explicitly been integrated with multilingual business content storage, the system does not rewrite its display value based on stable keys or seed mappings.

Menu governance is host navigation metadata: menu tree lists, parent menu displays, role menu trees, and related read-only selectors return localized titles from stable `menu_key` anchors. The editable `name` field in menu detail forms keeps the database value.

The only exception is framework-built-in governance records that are protected and cannot be edited or deleted. Those records provide name localization in read-only list display positions based on stable business anchors, while details, edit backfill, and selectors still keep database values.

## UI and Content Localization

### English Page Sweep

Framework-delivered pages are checked in `en-US` for Chinese system copy, seed display leakage, localized role names, generated department nodes, built-in config metadata, and public workbench content. This covers:
- Dashboard and shared shell surfaces (statistic cards, workspace demo copy, user profile display, route titles, tab-title refresh).
- Access management (user, role, menu page search fields, table headers, action buttons, drawers, authorization modals).
- Organization center (department/post pages, tree selectors, drawer forms, plugin menus).
- System settings (dictionary, parameter, file pages plus import/export/upload modals).
- Content management (notice lists, edit/preview modals).
- System monitoring (online users, service monitoring, operation logs, login logs).
- Scheduler center (job/group/execution-log pages, job forms).
- Extension center (plugin management, development center, dynamic plugin example pages).
- Host shared components (export confirmation, tree select, upload/cropper, profile center, security settings).

### Layout Adaptation

Forms and tables with long English labels receive layout adjustments so labels and critical table columns remain readable. Search labels, table headers, form labels, buttons, and tab titles must not become unreadable because English copy is longer. Constrained areas remain readable through layout adjustment, wider labels/columns, shorter default English copy, tooltips, or equivalent treatment.

### Menu and Navigation

Dynamic plugin route permission buttons mount under the owning plugin menu. Menu tree rows expose direct title-click expansion with pointer affordance. Button titles under resource menus use short action words such as `Query`, `Create`, `Update`, `Delete`, and `Export`.

### Built-in Data Protection

Built-in dictionaries, system parameters, and similar governance records are editable but protected from deletion in both frontend and backend. The backend returns structured business errors when deletion of built-in records is requested.

### Role Display Alignment

User management and role management show the same localized display value for the same built-in role. The user management list uses role display names returned by the backend.

### Workbench

The dashboard workbench uses runtime i18n for all copy. It displays real navigation entries and operational semantics instead of template demo content. User theme preference takes precedence over public frontend defaults. The brand logo icon adds a subtle cyan edge glow in dark mode.

### Demo Control

When the `demo-control` plugin is enabled, the system rejects plugin governance writes (synchronization, dynamic package upload, installation, uninstallation, enablement, disablement) while allowing plugin management reads.

## Project Governance

### Unified Positioning

LinaPro is uniformly positioned as an "AI-driven full-stack development framework". All project-level descriptions, system metadata, and user-visible copy maintain this positioning. The default management workspace, system management modules, user permission modules and similar capabilities are LinaPro's default entry points and built-in general capabilities, not the project's sole product boundary.

### Core Host Boundary

`apps/lina-core` is the framework's core host service, providing general module interface capabilities, component capabilities, system governance capabilities, and plugin extension capabilities. Its design prioritizes generality, stability, and reusability, and must not be strongly bound to specific workspace page display structures, interaction details, or frontend framework implementations.

Workspace adaptation interfaces (menu route projection, current-user workspace startup data, tree selectors, dropdown options) are explicitly classified as workspace adaptation interfaces rather than general domain interfaces. When a requirement only changes table columns, filters, tree selectors, workspace aggregation, route assembly, or other page-specific display structures, the system should prioritize solving through workspace adaptation interfaces or frontend adaptation layers.

### README Governance

All directory-level main documentation uses English `README.md` as the primary document. A Chinese mirror `README.zh-CN.md` must exist in the same directory. Both documents maintain consistent structure and information, differing only in language. Updates to one require synchronized updates to the other. The repository root provides both files as the project entry point.

### Database Bootstrap Safety

The host `init` and `mock` commands require explicit confirmation values matching the command name (`confirm=init` for init, `confirm=mock` for mock). The repository root `Makefile`, `apps/lina-core/Makefile`, and backend command implementation all follow the same confirmation semantics. The common SQL execution flow stops at the first error and returns failure status.

### Plugin Installation Governance

Source plugins default to "discovered but not installed" governance form. Both source and dynamic plugins go through a unified install review dialog before installation. Dynamic plugins form an authorization snapshot during installation; subsequent enables reuse the snapshot without repeating the confirmation dialog. Plugin resource interfaces are checked by the plugin's own resource permissions, not by additional governance permissions like `plugin:query`.

## Risks / Trade-offs

- **[Risk]** After removing clone semantics from `Translate`, if any business code assumes the returned map is writable and modifies it, the cache will be corrupted. -> Mitigation: `Translate` series only returns `string`, with no map exposed to business callers.
- **[Risk]** Sector cache refactoring involves multiple invalidation call sites. -> Mitigation: Replace bare calls with `Maintainer.InvalidateRuntimeBundleCache(scope)`, with scope explicitly passed by callers.
- **[Risk]** ETag negotiation may show stale translations when frontend persistence and backend bundleVersion are inconsistent. -> Mitigation: Every invalidation path increments version; persistent 7-day TTL provides fallback.
- **[Risk]** Interface splitting may cause downstream modules to change type signatures on a large scale. -> Mitigation: `Service` still exists as a composite type; changing field types to smaller interfaces is optional per-module.
- **[Risk]** Traditional Chinese may cause `CheckMissingMessages` to be permanently red due to missing translation resources. -> Mitigation: Traditional Chinese manifest completion is an independent task; CI threshold for `zh-TW` matches `en-US`.
- **[Risk]** Translation key count grows rapidly. -> Mitigation: Use namespace conventions and missing checks; split resource files by locale directory and semantic domain.
- **[Risk]** Batch export localization affects performance. -> Mitigation: Use request-level batch pre-fetching and map lookups; prohibit building bundles inside row loops.
- **[Risk]** Scan rule false positives or misses. -> Mitigation: First phase outputs warn/report; after stabilization, switch to blocking; allowlist carries classification and reason.
- **[Risk]** README governance scope expands to affect many files. -> Mitigation: Rules and entry points first, then incremental subdirectory governance.
- **[Trade-off]** ResourceLoader abstraction adds an extra layer of indirect calls. -> Accepted: eliminating ~280 lines of duplicate implementation far outweighs one abstraction layer.
- **[Trade-off]** Fixed LTR sacrifices future configuration flexibility for automatic direction switching. -> Accepted: current host positioning prioritizes reducing configuration complexity.
- **[Trade-off]** Removing database overrides means online translation hotfixing via API is no longer possible. -> Accepted: current priority is reducing new language complexity; future hotfix capability can be designed as an optional plugin.
