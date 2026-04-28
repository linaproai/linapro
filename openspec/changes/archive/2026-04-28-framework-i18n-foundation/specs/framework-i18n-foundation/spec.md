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

### Requirement: Host must provide unified translation resource aggregation and override priority
The host system SHALL aggregate translation messages from host default resources, project-level resources, plugin resources, and database overrides, and generate runtime-usable results with a unified priority. Runtime UI file resources MAY be authored as nested JSON or flat dotted keys; after loading, the host MUST normalize them to flat keys for governance. Database overrides, import/export, missing checks, and source diagnostics MUST use flat keys. Database overrides MUST have higher priority than file resources. After a user selects an enabled locale, runtime UI translation resolution MUST prefer current-locale resources. When the current locale lacks a translation, the system MUST fall back to caller-provided source default copy, an agreed default value, or a translation-key placeholder, and MUST NOT implicitly mix in another locale's default-language copy. Default-locale fallback translation methods MAY be called only when a specific business scenario explicitly allows cross-language default fallback.

#### Scenario: Database overrides replace file baselines
- **WHEN** a translation key has a baseline value in project files and a database override for the same `key + locale`
- **THEN** the runtime aggregated result returns the database override
- **AND** export or diagnostic results can identify that the value came from a database override

#### Scenario: Missing current-locale translation does not mix default-language copy
- **WHEN** the current request locale is `en-US` and the runtime UI resources for that locale lack a translation key
- **THEN** the system does not implicitly return the same key's `zh-CN` default-language value
- **AND** if the caller provides source default copy, the system returns that source copy
- **AND** if the caller does not provide source copy, the system returns the agreed default value or translation-key placeholder

#### Scenario: Explicit default-locale fallback is limited to scenarios that allow mixed languages
- **WHEN** a maintenance, diagnostic, or business content scenario explicitly requires viewing the default-language baseline when the current locale is missing
- **THEN** the system may call a default-locale fallback translation method
- **AND** the method name and comments must clearly state that it returns default-language content when the current locale is missing

### Requirement: Host must provide runtime message bundle distribution
The host system SHALL provide runtime message bundle and locale list APIs that return aggregated message resources and available locale descriptors by locale for the default workspace and host-embedded plugin pages. Runtime message bundles MUST include messages from the host, project, and currently enabled plugins, and MUST be converted to nested message objects that the frontend can consume directly.

#### Scenario: Default workspace loads runtime message bundle
- **WHEN** the frontend requests the runtime message bundle for `zh-CN`
- **THEN** the host returns aggregated messages for that locale
- **AND** the result includes merged host resources, project-level resources, and enabled plugin resources

#### Scenario: Frontend obtains host-supported locale list
- **WHEN** the frontend requests the runtime locale list
- **THEN** the host returns supported locale codes, default-locale markers, display names, and native names
- **AND** display names are returned according to the current request locale, while native names remain in their own language

#### Scenario: Runtime message bundles support nested maintenance and flat governance
- **WHEN** the host loads translation resources from files, plugins, or the database
- **THEN** the host allows runtime UI file resources to be authored as nested JSON or flat dotted keys
- **AND** internally aggregates and overrides messages as flat keys
- **AND** returns nested object structures from runtime APIs so the frontend can merge them directly into the `vue-i18n` message tree

#### Scenario: Disabled plugins no longer expose translation resources
- **WHEN** a plugin is disabled or uninstalled
- **THEN** subsequent runtime message bundle results no longer include messages contributed by that plugin
- **AND** other host and enabled plugin resources remain available

### Requirement: Host must define stable translation-key conventions for dynamic metadata
The host system SHALL derive dynamic metadata translation keys from stable business keys to avoid separate mapping tables for menus, dictionaries, configs, plugins, roles, cron jobs, logs, system information, and similar modules. Translation key rules for menus, dictionaries, configs, plugins, roles, scheduler data, logs, and system information MUST be defined in the framework specification and kept consistent in host and plugin implementations.

#### Scenario: Menu translation keys are derived from stable business keys
- **WHEN** the host needs to return a localized title for a menu
- **THEN** the host derives the translation key from that menu's stable business key
- **AND** the same menu has consistent localized results in menu trees, route titles, and dropdown trees

#### Scenario: Plugin metadata translation keys are derived from plugin identity
- **WHEN** the host needs to return a localized plugin name or plugin description
- **THEN** the host derives the plugin translation key from `plugin_id`
- **AND** the plugin does not need to repeat additional translation key fields in `plugin.yaml`

### Requirement: Backend-owned data must be localized by the backend
The host system SHALL localize backend-owned data in backend APIs, plugin host services, and export flows. Backend-owned data includes host governance metadata, plugin governance metadata, built-in roles, built-in scheduler jobs, job groups, execution logs, audit-log route metadata, login-log statuses and messages, and other display data persisted or registered by the backend. The default workspace SHALL render API response values or runtime messages directly, and MUST NOT maintain business data translation mappings in the frontend based on Chinese source text, database IDs, stable business keys, `handlerRef`, `dict_type`, `config_key`, `menu_key`, or other backend anchors.

#### Scenario: Backend API returns already-localized governance data
- **WHEN** an administrator requests role, menu, dictionary, config, cron job, job group, execution log, audit log, or plugin governance APIs with `en-US`
- **THEN** display fields that are backend-owned and allowed to be localized are already returned in the current language
- **AND** frontend tables, details, dropdowns, and export previews use API response values directly
- **AND** no frontend formatter needs to query Chinese source-text mapping tables to derive English display values

#### Scenario: User-editable business fields keep original values
- **WHEN** a role, job, group, config, notice, or other business record is created or edited by an administrator and the field is not explicitly integrated with multilingual business content storage
- **THEN** backend APIs and frontend pages continue to return and display the database value
- **AND** the system does not automatically rewrite that user-editable field based on current locale, field name, stable key, or historical seed content

#### Scenario: Frontend removes backend-data translation mappings
- **WHEN** the frontend renders backend-returned governance data, log data, or business content
- **THEN** the frontend MUST NOT add or continue depending on mappings such as `display-l10n` that translate by backend data value
- **AND** the frontend keeps only static UI `$t` copy, runtime message consumption, permission-code readability formatting, status styling, and layout adaptation responsibilities

### Requirement: Host must provide a common multilingual business content storage model
The host system SHALL provide a multilingual content storage model independent of specific business tables for managing multilingual titles, summaries, descriptions, bodies, and similar business content. The model MUST support at least `business_type`, `business_id`, `field`, `locale`, and `content`, and MUST support fallback to the default locale. For business master data editable in the management workspace, such as departments, posts, roles, dictionaries, parameters, notices, and scheduler data, management lists, details, edit backfill, and selectors MUST keep database values by default. Unless the specific field of that record has explicitly been integrated with multilingual business content storage, the system MUST NOT rewrite its display value based only on stable keys, seed mappings, or frontend heuristics. Menu governance is host navigation metadata: menu tree lists, parent menu displays, role menu trees, and related read-only selectors SHALL continue to return localized titles from stable `menu_key` anchors, but the editable `name` field in menu detail forms and edit backfill MUST keep the database value. The only exception is framework-built-in governance records that are protected and cannot be edited or deleted in the current governance page; those records MAY provide name localization in read-only list display positions based on stable business anchors, while details, edit backfill, and selectors MUST still keep database values.

#### Scenario: Business module reads multilingual titles
- **WHEN** a business module reads the title field of a business record according to the current locale
- **THEN** the system first returns the content value matching `business_type + business_id + field + locale`
- **AND** if the current locale is missing, it falls back to default-locale content

#### Scenario: Business modules keep defaults when multilingual content is not integrated
- **WHEN** a business module has not written multilingual content records for a field
- **THEN** the system can still return the module's default-language or original field value
- **AND** reading does not fail because multilingual records are missing

#### Scenario: Editable master data in management pages keeps original values when content translation is not integrated
- **WHEN** an administrator views editable business records such as departments, posts, roles, dictionaries, parameters, notices, or scheduler data in an English environment
- **THEN** lists, details, form backfill, and tree/dropdown selectors display the original database names, remarks, titles, or body values
- **AND** the system does not project those records to English only because they have anchors such as `dict_type` or `config_key`

#### Scenario: Menu governance page shows localized titles in English while keeping editable fields original
- **WHEN** an administrator views the menu management list, parent menu display, or role menu tree in an `en-US` environment
- **THEN** menu titles are returned as English or other localized results based on stable `menu_key`
- **AND** editable `name` fields in menu detail forms and edit backfill continue to show database values

#### Scenario: Built-in protected super administrator role may display English in list views
- **WHEN** an administrator opens the role management page in an `en-US` environment and views the framework built-in protected super administrator role
- **THEN** the backend role list API returns an English projected name for that record based on stable `role.key`
- **AND** other editable role records in the same list continue to display database values
- **AND** the super administrator role's details, edit backfill, and selector semantics still keep database values

### Requirement: Default workspace must fully localize framework-delivered pages and built-in display content
The host system SHALL ensure that the default workspace provides complete localized results for host pages, shared host components, built-in plugin pages, menu/tab/route titles, and framework-delivered `seed`, `mock`, and `demo` display content under enabled locales. This requirement covers governance metadata, example copy, and demo data delivered by the framework by default, but it MUST NOT override editable business master data in the management workspace that should display database field values. If a category of display content has not yet been integrated with common message translation, the system MUST integrate it with existing i18n infrastructure through stable business anchors, plugin resources, or default content translation records.

#### Scenario: English sweep of default workspace pages no longer shows Chinese system copy
- **WHEN** an administrator switches the language to `en-US` and opens framework-delivered host pages and built-in plugin pages one by one
- **THEN** menu names, tab titles, breadcrumbs, search labels, table headers, list titles, buttons, modal titles, and empty states are all shown in English
- **AND** no residual Chinese host shared component copy or route titles remain

#### Scenario: Default examples and seed display content are localized in English
- **WHEN** an administrator views framework default dashboard examples, scheduler default data, notice demo content, plugin example pages, and other delivered seed data in an `en-US` environment
- **THEN** those framework-delivered default display contents return English or the localized projection for the corresponding language
- **AND** users do not need to manually modify built-in SQL, plugin manifests, or frontend source code to remove default Chinese display values
- **AND** if the same page contains administrator-editable business record fields, those fields still display database values unless the field explicitly maintains corresponding multilingual content

#### Scenario: Long English copy still keeps stable usable layouts
- **WHEN** an administrator switches to `en-US` on a desktop viewport of at least `1366px` and opens framework-delivered list pages, search areas, forms, drawers, and modals
- **THEN** search labels, table headers, form labels, buttons, and tab titles do not become unreadable, obscure key fields, or move important information out of the visible area by default because English copy is longer
- **AND** constrained areas remain readable and operable through layout adjustment, wider labels/columns, shorter default English copy, tooltips, or equivalent treatment

### Requirement: Audit logs must display route module names and operation summaries by current language
The operation log plugin SHALL save stable route anchors such as route ownership, request method, public path, and apidoc operation key when recording audit logs, and SHALL preserve route metadata `tags` and `summary` source text as fallback values. In list, detail, and export scenarios, the operation log plugin SHALL reuse host `apidoc i18n JSON` resources to resolve and return localized module names and operation summaries according to the current request language. When the target language lacks a translation, the system MUST fall back to the source text saved in the log record. The host only provides generic route registration, dynamic route metadata reads, and apidoc copy resolution. It MUST NOT encode the operation log plugin's persistence fields into host plugin event payloads. The system MUST NOT copy API route copy into runtime UI locale bundles, and MUST NOT save only the current `tags` and `summary` text at operation time for all-language display.

#### Scenario: Operation language differs from viewing language
- **WHEN** an administrator triggers an operation log in a Chinese environment and later switches to `en-US` to view the operation log list or detail
- **THEN** the module name and operation summary in the operation log display in English
- **AND** viewing the same log again in `zh-CN` displays it in Chinese

#### Scenario: Audit log export uses current language
- **WHEN** an administrator exports operation logs in an `en-US` environment
- **THEN** exported module names, operation summaries, operation types, operation statuses, and headers display in English
- **AND** text from other languages is not mixed in because the log was created under another locale

#### Scenario: Missing audit route translations fall back safely
- **WHEN** an audit log records route anchors but the current language lacks matching apidoc translations
- **THEN** the system returns the route source text stored in the log
- **AND** list, detail, and export views do not display blank values or raw translation keys

#### Scenario: Operation log persistence structure belongs to the plugin
- **WHEN** the operation log plugin collects and saves audit records through global middleware
- **THEN** the host only provides generic capabilities such as dynamic route metadata lookup, route registration context, and apidoc copy resolution
- **AND** the host MUST NOT define or distribute audit record payloads that map one-to-one to operation log plugin persistence table fields
- **AND** the host dynamic route runtime MUST NOT expose operation-log-named data structures or structures bounded by operation log field projection
- **AND** plugin custom metadata in dynamic plugin route contracts can pass only through generic `meta`; the host MUST NOT define or validate operation-log-specific route fields or operation type enums
- **AND** changes to the operation log plugin table structure, create input, or export fields MUST NOT require host plugin service event encoding changes

### Requirement: System API documentation must use English source copy and independent apidoc translation resources
The host system SHALL maintain readable English OpenAPI documentation source text in host, source-plugin, and dynamic-plugin API DTOs, including route groups, summaries, descriptions, request parameter descriptions, response parameter descriptions, and fixed route projection copy that can enter API documentation. API documentation localization SHALL run while rendering `/api.json` according to the current request language, using independent `manifest/i18n/<locale>/apidoc/**/*.json` resources. These resources MUST be decoupled from default workspace runtime UI resources in `manifest/i18n/<locale>/*.json`. Host API translations SHALL be maintained under `apps/lina-core/manifest/i18n/<locale>/apidoc/`. Plugin API translations SHALL be maintained independently by each plugin under `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/apidoc/`. The lina-core apidoc module SHALL discover source plugin embedded resources and dynamic plugin runtime artifact apidoc resources at render time, merge them, and apply them uniformly. Plugin `plugins.*` translation keys MUST NOT be centralized into lina-core apidoc resources. Non-English apidoc files MAY be authored as nested JSON or flat dotted keys; after loading, the host MUST normalize them to stable structured keys. Repeated metadata such as standard response wrappers, pagination fields, and common time fields MAY use host `core.common.*` fallback keys, while specific structure keys MUST take precedence when present. English API documentation SHALL directly use English source copy, and every `manifest/i18n/en-US/apidoc/**/*.json` file SHALL remain an empty-object placeholder to avoid duplicate English mappings. If GoFrame, generated `entity` schemas, or database table comment metadata enters API documentation, the system SHALL display it as supplied by its source. It MUST NOT maintain Chinese-to-English temporary conversion tables in the apidoc service layer, and MUST NOT rely on restored `en-US/apidoc` placeholder mappings or non-English apidoc JSON mappings as fallback. To change that display language, the corresponding source data or generation source must be changed. The system MUST NOT write Chinese source text or opaque `i18n` key placeholders into hand-authored API DTO documentation tags. `eg/example` sample values SHALL keep real request/response example semantics and MUST NOT be written to or translated by apidoc i18n resources.

#### Scenario: English API documentation directly uses English source copy
- **WHEN** an administrator opens system API documentation in English or requests `/api.json?lang=en-US`
- **THEN** hand-authored API DTO route groups, summaries, descriptions, request parameter descriptions, and response parameter descriptions from the host, source plugins, and dynamic plugins display in English
- **AND** English API documentation does not depend on `en-US` apidoc translation mappings, and JSON files under `manifest/i18n/en-US/apidoc/` exist only as empty-object placeholders
- **AND** generated `entity` schemas, database table comments, and framework common response metadata are not temporarily translated by the apidoc service layer and display as supplied by their source
- **AND** non-English apidoc JSON MUST NOT maintain generated schema translation keys such as `internal.model.entity.*`

#### Scenario: Chinese API documentation is projected through apidoc JSON
- **WHEN** an administrator opens system API documentation in Chinese or requests `/api.json?lang=zh-CN`
- **THEN** English source copy maintained in API DTOs is mapped to Chinese through stable structured keys in `manifest/i18n/zh-CN/apidoc/**/*.json`
- **AND** API documentation translation resources do not enter runtime UI language bundles and do not depend on default workspace frontend static locale bundles
- **AND** plugin API `plugins.*` translation keys are provided by each plugin's own apidoc i18n JSON and applied after lina-core apidoc merging
- **AND** standard response fields, pagination fields, and common time fields can use `core.common.*` fallback translations to avoid repeating the same copy under every API structure
- **AND** `eg/example` sample values continue to display the original examples from DTOs or OpenAPI projection

#### Scenario: API copy changes validate apidoc translation coverage
- **WHEN** a developer adds or modifies OpenAPI documentation tags in host, source-plugin, or dynamic-plugin API DTOs
- **THEN** the developer must update the owning host or plugin non-English `manifest/i18n/<locale>/apidoc` translation resources with stable structured keys, and MUST NOT use English source copy as JSON keys
- **AND** automated tests or review rules must block missing non-English translations, Chinese source text, and opaque placeholders in API documentation, while `en-US` empty placeholder files and `eg/example` sample values do not require translations

### Requirement: Host must provide i18n maintenance and validation capability
The host system SHALL provide locale enablement, translation message import/export, missing translation checks, and resource source diagnostics to support long-term multilingual maintenance for delivered projects. Delivered projects and plugins MUST follow unified directory conventions and translation key rules when adding i18n resources.

#### Scenario: Export translation resources for a locale
- **WHEN** an administrator or delivery tool requests export for `en-US` translation resources
- **THEN** the system returns aggregated messages or an importable file for that locale
- **AND** the exported content can be used for later batch maintenance and re-import

#### Scenario: Check missing translation keys
- **WHEN** an administrator or delivery tool runs a missing translation check
- **THEN** the system returns the list of translation keys missing in the current locale relative to the default locale
- **AND** the result can identify whether each missing key belongs to host modules, project resources, or plugin resource scope
