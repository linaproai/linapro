# Framework I18n Foundation

## Purpose

Define the host i18n foundation, including language discovery, runtime bundle distribution, shared resource loading, cache invalidation, and maintenance capabilities.
## Requirements
### Requirement: Host translation service interface must be split into multiple small interfaces by responsibility
The host system SHALL split the i18n translation service interface into `LocaleResolver`, `Translator`, `BundleProvider`, and `Maintainer` four small interfaces. Each small interface MUST only carry one responsibility: `LocaleResolver` resolves request language and context language; `Translator` provides translation lookup and error localization; `BundleProvider` outputs runtime translation bundles and language lists; `Maintainer` provides resource export, missing checks, source diagnostics, and cache invalidation. The `Service` type MUST be a composition of these four small interfaces, and business modules' `i18nSvc` field types SHALL converge to the minimum interface they actually depend on rather than the entire `Service`. The host core i18n service must not provide database translation override import or generic business content multilingual persistence interfaces.

#### Scenario: Business modules only declare minimum interface dependencies
- **WHEN** `menu` / `dict` / `sysconfig` / `jobmgmt` and other modules need localization translation capabilities
- **THEN** the module declares its `i18nSvc` field as the minimum combination of `LocaleResolver` and `Translator` in its own struct
- **AND** module unit tests can mock only these two small interfaces without stubbing maintenance methods

#### Scenario: Controller tests complete via minimum interface stubs
- **WHEN** testing `i18n` management controllers (export, diagnostics, missing checks)
- **THEN** tests can individually mock the `Maintainer` interface
- **AND** do not need to simultaneously provide placeholder implementations for `Translator` / `BundleProvider`

### Requirement: Host must discover built-in languages via resource conventions and default configuration
The host system SHALL auto-discover built-in runtime languages from `manifest/i18n/<locale>/*.json` files, and maintain default language, multi-language toggle, display sorting, native names and other metadata that cannot be derived from filenames through the `i18n` configuration section in the default config file. Adding a new built-in language MUST only require adding corresponding runtime JSON, apidoc JSON, plugin JSON, and optional default config metadata, and MUST NOT require adding backend Go language enumerations, SQL seeds, or frontend TS language lists. Runtime text direction is fixed to `ltr` per current host convention, and MUST NOT require maintaining `direction` in configuration. Runtime language list, missing translation checks, resource source diagnostics, runtime translation bundle API, ETag negotiation, and frontend persistent cache MUST automatically cover `zh-TW`.

#### Scenario: Traditional Chinese auto-joins runtime language list from resource files after startup
- **WHEN** the project has `manifest/i18n/zh-TW/*.json`
- **AND** the default config file `i18n.locales` provides `nativeName` for `zh-TW`
- **AND** after service startup the frontend requests the runtime language list
- **THEN** the `/i18n/runtime/locales` API returns a language list containing `zh-TW`
- **AND** `zh-TW` is marked as non-default language
- **AND** `zh-TW`'s direction field is the fixed value `ltr`

#### Scenario: Adding a new language requires no Go, SQL, or frontend TS language list changes
- **WHEN** the delivery project adds `manifest/i18n/<locale>/*.json` and `manifest/i18n/<locale>/apidoc/**/*.json` resources for a language
- **AND** source plugins and dynamic plugins add that language's resources following the same directory convention
- **AND** if default language, sorting, native name, or enable/disable state needs to be controlled, only modify the `i18n` config section in the default config file
- **THEN** menu, dictionary, config, scheduled tasks, plugins, roles, system info and other dynamic metadata automatically return localized results in that language
- **AND** the runtime language list automatically includes that language
- **AND** no changes to backend Go constants, SQL seeds, frontend `SUPPORT_LANGUAGES`, or third-party locale switch branches are needed

#### Scenario: Disabling multi-language uses only the default language
- **WHEN** `i18n.enabled` is `false` in the default config file
- **AND** the user's browser or request parameters pass a non-default language
- **THEN** the host request language resolution falls back to `i18n.default`
- **AND** the `/i18n/runtime/locales` response marks the multi-language toggle as off, and only returns the default language descriptor
- **AND** the default management workbench hides the language switch button, loading static language packs, runtime translation bundles, and public frontend config in the default language

#### Scenario: Removing a language from the locales list disables it
- **WHEN** the project has multiple `manifest/i18n/<locale>/*.json` resources
- **AND** the default config file `i18n.locales` only lists a subset of those languages
- **THEN** `/i18n/runtime/locales` only returns languages listed in `i18n.locales`
- **AND** requests for unlisted languages fall back to `i18n.default`

### Requirement: Default management workbench must maintain fixed LTR document direction
The default management workbench SHALL fix document direction to `ltr` per current host convention. During language switching, the workbench MUST simultaneously set `<html dir>` to `ltr` and inject `direction="ltr"` into `Ant Design Vue`'s `ConfigProvider`. The frontend must not maintain a static RTL language registry, and must not require modifying direction-related TypeScript branches when adding new languages.

#### Scenario: html direction remains LTR when switching to Traditional Chinese
- **WHEN** a user switches language to `zh-TW` in the default management workbench
- **THEN** `document.documentElement`'s `dir` attribute remains `ltr`
- **AND** `Ant Design Vue`'s `ConfigProvider` receives `direction="ltr"`
- **AND** switching back to `zh-CN` or `en-US` still keeps `dir` as `ltr`

#### Scenario: Traditional Chinese page copy completeness is sufficient
- **WHEN** a user opens framework default-delivered list pages, drawers and modals in Traditional Chinese environment
- **THEN** page copy is displayed in Traditional Chinese and layout does not block core operations
- **AND** RTL mirrored layout is not required

### Requirement: Translation resource loader must be shared between host and plugins, UI and apidoc
The host system SHALL provide a unified `ResourceLoader` component in the `pkg/i18nresource` package, accepting `Subdir`, `LocaleSubdir`, `PluginScope`, `LayoutMode` and other configuration parameters, centrally implementing the discovery and loading logic for "host embedded resources -> source plugin resources -> dynamic plugin resources". Runtime UI translation resource loading and API documentation translation resource loading MUST be completed through different `ResourceLoader` instances, and MUST NOT each maintain duplicate implementations, and MUST NOT cause the API documentation module to reverse-depend on runtime `internal/service/i18n` just to reuse the loader. Source plugin apidoc namespace isolation MUST be achieved by `ResourceLoader` configuration rather than additional duplicate code.

#### Scenario: Runtime bundle and apidoc share the same resource loader implementation
- **WHEN** the system loads runtime UI translation resources or apidoc translation resources
- **THEN** both pipelines complete host, source plugin, and dynamic plugin resource traversal through the same `i18nresource.ResourceLoader` implementation
- **AND** the apidoc pipeline constrains plugin namespace via `PluginScope=RestrictedToPluginNamespace` configuration
- **AND** the runtime UI pipeline allows plugins to contribute arbitrary keys via `PluginScope=Open` configuration

### Requirement: Host must provide runtime translation bundle distribution capability
The host system SHALL provide a runtime translation bundle API and language list API, returning aggregated message resources and current available language descriptor information by language, for the default management workbench and host embedded plugin pages to load. The runtime translation bundle MUST be able to simultaneously contain host, source plugin, and currently enabled dynamic plugin i18n messages, and convert them to nested message objects directly consumable by the frontend on output. The runtime translation bundle API MUST output an `ETag` header in the response, with a value derived from the current language and runtime translation bundle version that must differ when the version changes; the system MUST receive the `If-None-Match` header from requests and return `304 Not Modified` without carrying a message body when matched. Any sector cache invalidation MUST trigger runtime translation bundle version auto-increment, ensuring different bundle contents for the same language have different ETags.

#### Scenario: Default workbench loads runtime translation bundle
- **WHEN** the frontend requests the runtime translation bundle with `zh-CN`
- **THEN** the host returns the aggregated message set for that language
- **AND** the result contains the merged result of host resources, source plugin resources, and enabled dynamic plugin resources
- **AND** the response contains an `ETag` header

#### Scenario: Frontend obtains host-supported language list
- **WHEN** the frontend requests the runtime language list
- **THEN** the host returns the multi-language toggle, current supported language codes, default language marker, display name, native name, and fixed LTR text direction
- **AND** display names are returned in the current request language, while native names remain in the corresponding language's own text
- **AND** the list includes `zh-CN`, `en-US` and `zh-TW`

#### Scenario: Runtime language pack supports hierarchical maintenance while maintaining flat governance
- **WHEN** the host loads translation resources from files, source plugins, or dynamic plugins
- **THEN** the host allows runtime UI file resources to use hierarchical JSON or flat dotted key format
- **AND** the host internally aggregates messages uniformly as flat keys
- **AND** the runtime API returns results to the frontend as nested object structures, for direct merging into the frontend `vue-i18n` message tree

#### Scenario: Disabled plugin's translation resources are no longer exposed
- **WHEN** a plugin is disabled or uninstalled
- **THEN** subsequent runtime translation bundle results no longer contain translation messages contributed by that plugin
- **AND** other host and enabled plugin resources remain available
- **AND** the system correspondingly triggers cache invalidation for the sectors related to that plugin, and the runtime translation bundle version auto-increments

#### Scenario: Second request for the same translation bundle returns 304
- **WHEN** the frontend saves the `ETag` from the first runtime translation bundle request
- **AND** no cache invalidation has occurred on the backend between the two requests
- **AND** the frontend carries `If-None-Match` equal to the previous `ETag` in the second request
- **THEN** the backend returns `304 Not Modified` without carrying a message body

### Requirement: Host must provide internationalization maintenance and validation capabilities
The host system SHALL provide translation resource export, missing translation check, and resource source diagnostics capabilities to support delivery projects in maintaining multilingual resources during development. Delivery projects and plugins MUST follow unified directory conventions and translation key specifications when adding i18n resources. Cache invalidation MUST provide fine-grained control by "language x sector (host / source-plugin / dynamic-plugin)" dimension, and MUST NOT clear the entire cache for all languages and all sectors on any single-point change. All determinations of "this key is owned by a code source" MUST be completed through an explicit namespace registry, and the system MUST NOT hardcode any specific business module's namespace prefix in the i18n package. The host core must not create or depend on `sys_i18n_locale`, `sys_i18n_message`, `sys_i18n_content` three runtime i18n persistence tables; the single source of truth for translation content is development-time JSON/YAML resources.

#### Scenario: Export translation resources for a language
- **WHEN** an administrator or delivery tool requests exporting `en-US` translation resources
- **THEN** the system returns the aggregated message result for that language
- **AND** the export content can be used for offline proofreading and writing back to the corresponding JSON resource file

#### Scenario: Check missing translation keys
- **WHEN** an administrator or delivery tool executes a missing translation check
- **THEN** the system returns the list of translation keys missing in the current language relative to the default language
- **AND** the results can locate which host module, project resource, or plugin resource scope the missing keys belong to
- **AND** translation keys registered as code-owned namespaces do not appear in missing results

#### Scenario: Plugin resource change only clears affected plugin sector
- **WHEN** a dynamic plugin is enabled, disabled, uninstalled, or upgraded
- **THEN** the system only clears the dynamic plugin sector cache and merged view related to that plugin
- **AND** other languages, host resources, and unaffected plugin resources remain valid
- **AND** the runtime translation bundle version for that language auto-increments, ensuring the frontend can detect changes on next negotiation

### Requirement: English regression sweep must cover framework-delivered pages and seed display content
The default management workbench SHALL provide English regression coverage for framework-delivered pages from manual feedback, ensuring system-generated content, default seed display content, and static UI copy do not retain Chinese text.

#### Scenario: English regression pages contain no Chinese system copy
- **WHEN** an administrator switches to `en-US` and opens workbench, user management, role management, department management, post management, dictionary management, system config, service monitoring, and scheduled jobs
- **THEN** framework-delivered titles, buttons, form labels, table columns, system-generated nodes, built-in record displays, and confirmation modals use English
- **AND** user-editable business fields are localized only when explicitly included in framework-delivered projection rules

#### Scenario: English layout regressions are screenshot checked
- **WHEN** Playwright captures post forms, dictionary forms, and service monitoring disk tables in `en-US`
- **THEN** key labels, options, headers, and values do not wrap unreadably or overlap
- **AND** screenshot results are part of acceptance evidence

#### Scenario: Version information menu title is localized consistently
- **WHEN** an administrator views the version information menu entry under the development center
- **THEN** Simplified Chinese, English, and Traditional Chinese locales each show the locale-appropriate title
- **AND** the English title is `Version Info`

### Requirement: Runtime locale JSON values must avoid markdown-only code markers
Runtime translation JSON SHALL avoid markdown-style backtick markers in user-visible strings because ordinary UI rendering does not apply code highlighting and raw backticks reduce readability.

#### Scenario: Locale JSON strings are displayed as plain UI text
- **WHEN** frontend, host, or plugin locale JSON strings contain file paths, parameter examples, wildcards, or extensions
- **THEN** strings display the content directly
- **AND** strings do not wrap those values in backticks
- **AND** automated checks prevent locale JSON strings from reintroducing backticks

