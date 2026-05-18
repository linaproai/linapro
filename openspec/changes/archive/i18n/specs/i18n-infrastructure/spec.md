## ADDED Requirements

### Requirement: Host must provide unified locale resolution and fallback
The host system SHALL resolve the current locale for every request and write the result into a unified business context for controllers, services, plugin bridges, and runtime translation aggregation. Locale resolution priority MUST be the `lang` query parameter, the `Accept-Language` request header, and then the system default locale. When the requested locale is unavailable, the system MUST fall back to the default locale.

#### Scenario: Query parameter overrides request locale
- **WHEN** a client requests any i18n-enabled API and explicitly passes `lang=en-US`
- **THEN** the host uses `en-US` as the effective locale for that request
- **AND** it does not continue using the `Accept-Language` result from the same request

#### Scenario: Unenabled request locale falls back to the default locale
- **WHEN** a client requests a locale that is not enabled or not supported
- **THEN** the host falls back to the system default locale
- **AND** dynamic metadata and runtime message bundles in that request are returned using the default locale

### Requirement: Host must provide unified translation resource aggregation with file-only single source of truth
The host system SHALL aggregate translation messages from host default resources, project-level resources, and plugin resources, generating runtime-usable results with a unified priority. Runtime UI file resources MAY be authored as nested JSON or flat dotted keys; after loading, the host MUST normalize them to flat keys for governance. After a user selects an enabled locale, runtime UI translation resolution MUST prefer current-locale resources. When the current locale lacks a translation, the system MUST fall back to caller-provided source default copy, an agreed default value, or a translation-key placeholder, and MUST NOT implicitly mix in another locale's default-language copy. The host core MUST NOT create or depend on `sys_i18n_locale`, `sys_i18n_message`, `sys_i18n_content` three runtime i18n persistence tables; the single source of truth for translation content is development-time JSON/YAML resources.

#### Scenario: Missing current-locale translation does not mix default-language copy
- **WHEN** the current request locale is `en-US` and the runtime UI resources for that locale lack a translation key
- **THEN** the system does not implicitly return the same key's `zh-CN` default-language value
- **AND** if the caller provides source default copy, the system returns that source copy
- **AND** if the caller does not provide source copy, the system returns the agreed default value or translation-key placeholder

#### Scenario: Plugin resource change only clears affected plugin sector
- **WHEN** a dynamic plugin is enabled, disabled, uninstalled, or upgraded
- **THEN** the system only clears the dynamic plugin sector cache and merged view related to that plugin
- **AND** other languages, host resources, and unaffected plugin resources remain valid
- **AND** the runtime translation bundle version for that language auto-increments

### Requirement: Host must provide runtime message bundle distribution with ETag negotiation
The host system SHALL provide a runtime translation bundle API and language list API, returning aggregated message resources and available locale descriptors by locale. The runtime translation bundle MUST include messages from the host, project, and currently enabled plugins, and MUST be converted to nested message objects that the frontend can consume directly. The runtime translation bundle API MUST output an `ETag` header derived from the current language and bundle version, and MUST support `If-None-Match` 304 negotiation. Any sector cache invalidation MUST trigger bundle version auto-increment.

#### Scenario: Default workspace loads runtime message bundle
- **WHEN** the frontend requests the runtime message bundle for `zh-CN`
- **THEN** the host returns aggregated messages for that locale
- **AND** the result includes merged host resources, project-level resources, and enabled plugin resources
- **AND** the response contains an `ETag` header

#### Scenario: Second request for the same bundle returns 304
- **WHEN** the frontend saves the `ETag` from the first runtime translation bundle request
- **AND** no cache invalidation has occurred on the backend between the two requests
- **AND** the frontend carries `If-None-Match` equal to the previous `ETag` in the second request
- **THEN** the backend returns `304 Not Modified` without carrying a message body

#### Scenario: Disabled plugins no longer expose translation resources
- **WHEN** a plugin is disabled or uninstalled
- **THEN** subsequent runtime message bundle results no longer include messages contributed by that plugin
- **AND** other host and enabled plugin resources remain available
- **AND** the system triggers cache invalidation for the sectors related to that plugin

#### Scenario: Runtime message bundles support nested maintenance and flat governance
- **WHEN** the host loads translation resources from files, plugins, or dynamic plugins
- **THEN** the host allows runtime UI file resources to be authored as nested JSON or flat dotted keys
- **AND** internally aggregates and overrides messages as flat keys
- **AND** returns nested object structures from runtime APIs for frontend consumption

### Requirement: Host must define stable translation-key conventions for dynamic metadata
The host system SHALL derive dynamic metadata translation keys from stable business keys to avoid separate mapping tables for menus, dictionaries, configs, plugins, roles, cron jobs, logs, system information, and similar modules. Translation key rules MUST be defined in the framework specification and kept consistent in host and plugin implementations.

#### Scenario: Menu translation keys are derived from stable business keys
- **WHEN** the host needs to return a localized title for a menu
- **THEN** the host derives the translation key from that menu's stable business key
- **AND** the same menu has consistent localized results in menu trees, route titles, and dropdown trees

#### Scenario: Plugin metadata translation keys are derived from plugin identity
- **WHEN** the host needs to return a localized plugin name or plugin description
- **THEN** the host derives the plugin translation key from `plugin_id`
- **AND** the plugin does not need to repeat additional translation key fields in `plugin.yaml`

### Requirement: Backend-owned data must be localized by the backend
The host system SHALL localize backend-owned data in backend APIs, plugin host services, and export flows. Backend-owned data includes host governance metadata, plugin governance metadata, built-in roles, built-in scheduler jobs, job groups, execution logs, audit-log route metadata, login-log statuses and messages, and other display data persisted or registered by the backend. The default workspace SHALL render API response values or runtime messages directly, and MUST NOT maintain business data translation mappings in the frontend based on Chinese source text, database IDs, stable business keys, or other backend anchors.

#### Scenario: Backend API returns already-localized governance data
- **WHEN** an administrator requests role, menu, dictionary, config, cron job, job group, execution log, audit log, or plugin governance APIs with `en-US`
- **THEN** display fields that are backend-owned and allowed to be localized are already returned in the current language
- **AND** frontend tables, details, dropdowns, and export previews use API response values directly

#### Scenario: User-editable business fields keep original values
- **WHEN** a role, job, group, config, notice, or other business record is created or edited by an administrator and the field is not explicitly integrated with multilingual business content storage
- **THEN** backend APIs and frontend pages continue to return and display the database value

### Requirement: Default workspace must fully localize framework-delivered pages and built-in display content
The host system SHALL ensure that the default workspace provides complete localized results for host pages, shared host components, built-in plugin pages, menu/tab/route titles, and framework-delivered seed, mock, and demo display content under enabled locales. This requirement covers governance metadata, example copy, and demo data delivered by the framework by default, but it MUST NOT override editable business master data that should display database field values.

#### Scenario: English sweep of default workspace pages no longer shows Chinese system copy
- **WHEN** an administrator switches the language to `en-US` and opens framework-delivered pages one by one
- **THEN** menu names, tab titles, breadcrumbs, search labels, table headers, list titles, buttons, modal titles, and empty states are all shown in English

#### Scenario: Long English copy still keeps stable usable layouts
- **WHEN** an administrator switches to `en-US` on a desktop viewport of at least `1366px`
- **THEN** search labels, table headers, form labels, buttons, and tab titles do not become unreadable because English copy is longer
- **AND** constrained areas remain readable and operable through layout adjustment, wider labels/columns, shorter default English copy, tooltips, or equivalent treatment

### Requirement: Audit logs must display route module names and operation summaries by current language
The operation log plugin SHALL save stable route anchors when recording audit logs, and SHALL reuse host apidoc i18n JSON resources to resolve and return localized module names and operation summaries according to the current request language. The host only provides generic route registration, dynamic route metadata reads, and apidoc copy resolution. It MUST NOT encode the operation log plugin's persistence fields into host plugin event payloads.

#### Scenario: Operation language differs from viewing language
- **WHEN** an administrator triggers an operation log in a Chinese environment and later switches to `en-US` to view the operation log
- **THEN** the module name and operation summary display in English

### Requirement: System API documentation must use English source copy and independent apidoc translation resources
The host system SHALL maintain readable English OpenAPI documentation source text in API DTOs. API documentation localization SHALL run while rendering `/api.json` according to the current request language, using independent `manifest/i18n/<locale>/apidoc/**/*.json` resources. These resources MUST be decoupled from default workspace runtime UI resources. `en-US` API documentation SHALL directly use English source copy, and every `manifest/i18n/en-US/apidoc/**/*.json` file SHALL remain an empty-object placeholder. Generated entity schemas, database table comments, and framework common response metadata display as supplied by their source. `eg/example` sample values SHALL keep real request/response example semantics and MUST NOT be written to or translated by apidoc i18n resources.

#### Scenario: English API documentation directly uses English source copy
- **WHEN** an administrator requests `/api.json?lang=en-US`
- **THEN** hand-authored API DTO route groups, summaries, descriptions, and parameter descriptions display in English
- **AND** English API documentation does not depend on `en-US` apidoc translation mappings

### Requirement: Host must provide i18n maintenance and validation capability
The host system SHALL provide translation resource export, missing translation checks, and resource source diagnostics to support long-term multilingual maintenance. Cache invalidation MUST provide fine-grained control by "language x sector" dimension, and MUST NOT clear the entire cache for all languages and all sectors on any single-point change. All determinations of "this key is owned by a code source" MUST be completed through an explicit namespace registry.

#### Scenario: Export translation resources for a locale
- **WHEN** an administrator or delivery tool requests export for `en-US` translation resources
- **THEN** the system returns aggregated messages for that locale

#### Scenario: Check missing translation keys
- **WHEN** an administrator or delivery tool runs a missing translation check
- **THEN** the system returns the list of translation keys missing in the current locale relative to the default locale
- **AND** translation keys registered as code-owned namespaces do not appear in missing results

### Requirement: Host must discover built-in languages via resource conventions and default configuration
The host system SHALL auto-discover built-in runtime languages from `manifest/i18n/<locale>/*.json` files, and maintain default language, multi-language toggle, display sorting, native names and other metadata through the `i18n` configuration section in the default config file. Adding a new built-in language MUST only require adding corresponding runtime JSON, apidoc JSON, plugin JSON, and optional default config metadata, and MUST NOT require adding backend Go language enumerations, SQL seeds, or frontend TS language lists. Runtime text direction is fixed to `ltr` per current host convention.

#### Scenario: Adding a new language requires no Go, SQL, or frontend TS language list changes
- **WHEN** the delivery project adds `manifest/i18n/<locale>/*.json` resources for a language
- **AND** source plugins and dynamic plugins add that language's resources following the same directory convention
- **THEN** menu, dictionary, config, scheduled tasks, plugins, roles, system info and other dynamic metadata automatically return localized results in that language
- **AND** the runtime language list automatically includes that language

#### Scenario: Disabling multi-language uses only the default language
- **WHEN** `i18n.enabled` is `false` in the default config file
- **THEN** the host request language resolution falls back to `i18n.default`
- **AND** the default management workbench hides the language switch button

### Requirement: Default management workbench must maintain fixed LTR document direction
The default management workbench SHALL fix document direction to `ltr` per current host convention. During language switching, the workbench MUST simultaneously set `<html dir>` to `ltr` and inject `direction="ltr"` into `Ant Design Vue`'s `ConfigProvider`.

#### Scenario: html direction remains LTR when switching to Traditional Chinese
- **WHEN** a user switches language to `zh-TW`
- **THEN** `document.documentElement`'s `dir` attribute remains `ltr`
- **AND** `Ant Design Vue`'s `ConfigProvider` receives `direction="ltr"`

### Requirement: Translation lookup hot path must avoid cloning the entire runtime message bundle
The host system SHALL let single-value-returning translation lookup methods directly hold a read lock on the internal message bundle and look up values when the cache hits, and MUST NOT clone or copy the entire runtime message bundle. Only methods that need to return the message set to the caller MAY clone once before returning.

#### Scenario: Single key translation lookup does not clone entire message bundle on cache hit
- **WHEN** a business module calls `Translate(ctx, key, fallback)` and the cache already exists
- **THEN** the system only holds a read lock and looks up the value, directly returning the found string
- **AND** no `cloneFlatMessageMap` or equivalent full map copy is performed

### Requirement: Runtime translation cache must support layered invalidation by language and sector
The host system SHALL organize the runtime translation message cache into a "language x sector (host / source-plugin / dynamic-plugin)" layered structure, and provide fine-grained invalidation capabilities by sector dimension. Any business-event-triggered invalidation MUST only clear the affected language or sector, and MUST NOT "clear all" caches for all languages and all sectors.

#### Scenario: Dynamic plugin enable/disable only clears that plugin's relevant sectors
- **WHEN** a dynamic plugin is enabled, disabled, or upgraded
- **THEN** the system only clears the dynamic plugin sector cache and merged view related to that plugin
- **AND** host and unaffected plugins' translation data continues to hit cache

### Requirement: Translation service interface must be split into multiple small interfaces by responsibility
The host system SHALL split the i18n translation service interface into `LocaleResolver`, `Translator`, `BundleProvider`, and `Maintainer` four small interfaces. The `Service` type MUST be a composition of these four, and business modules' `i18nSvc` field types SHALL converge to the minimum interface they actually depend on.

#### Scenario: Business modules only declare minimum interface dependencies
- **WHEN** menu / dict / sysconfig / jobmgmt and other modules need localization capabilities
- **THEN** the module declares its `i18nSvc` field as the minimum combination of `LocaleResolver` and `Translator`
- **AND** module unit tests can mock only these small interfaces without stubbing maintenance methods

### Requirement: Translation resource loader must be shared between host and plugins, UI and apidoc
The host system SHALL provide a unified `ResourceLoader` component in `pkg/i18nresource/`, accepting `Subdir`, `LocaleSubdir`, `PluginScope`, `LayoutMode` and other configuration parameters. Runtime UI translation resource loading and API documentation translation resource loading MUST be completed through different `ResourceLoader` instances, and MUST NOT each maintain duplicate implementations.

#### Scenario: Runtime bundle and apidoc share the same resource loader implementation
- **WHEN** the system loads runtime UI or apidoc translation resources
- **THEN** both pipelines complete host, source plugin, and dynamic plugin resource traversal through the same `i18nresource.ResourceLoader` implementation
- **AND** the apidoc pipeline constrains plugin namespace via `PluginScope=RestrictedToPluginNamespace` configuration

### Requirement: Business modules must own their own localization projection rules
The host system SHALL let business modules maintain localization projection rules within their own module boundaries. `internal/service/i18n` SHALL only provide foundational capabilities and MUST NOT reference business entities, business protection rules, or business translation key derivation logic.

#### Scenario: Business modules complete projection within their own boundaries
- **WHEN** menu / dict / sysconfig / jobmgmt / role / plugin runtime modules need to localize query results
- **THEN** the module derives translation keys, determines whether to skip default language, and determines whether records are protected in its own `*_i18n.go`
- **AND** `internal/service/i18n` does not import these business modules or business entities

### Requirement: Host must provide explicit registration mechanism for code-owned source text namespaces
The host system SHALL provide `RegisterSourceTextNamespace(prefix, reason string)` in the `internal/service/i18n` package. Business modules MUST register their code-owned namespaces in their own `init()`. The `i18n` package MUST NOT hardcode any specific business module's namespace prefix.

#### Scenario: Business modules register code-owned namespaces via init
- **WHEN** the `jobmgmt` package executes `init()` at project startup
- **THEN** the package registers its namespace via `i18n.RegisterSourceTextNamespace("job.handler.", "code-owned cron handler display")`
- **AND** missing checks can identify these keys as code-owned without modifying the `i18n` package source

### Requirement: WASM custom section parsing capability must be centrally provided by pluginbridge
The host system SHALL provide `ReadCustomSection` and `ListCustomSections` public capabilities in `pkg/pluginbridge/pluginbridge_wasm_section.go`. `i18n`, `apidoc`, and plugin runtime MUST read custom sections from dynamic plugin runtime artifacts through this capability, and MUST NOT maintain duplicate WASM parsing implementations.

#### Scenario: i18n reads dynamic plugin i18n section via pluginbridge
- **WHEN** the system needs to read the `i18n_assets` custom section from a dynamic plugin artifact
- **THEN** the caller completes this via `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)`
- **AND** no dedicated parsing functions exist in the `i18n` package
