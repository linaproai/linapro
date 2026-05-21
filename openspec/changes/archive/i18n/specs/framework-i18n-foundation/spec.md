## MODIFIED Requirements

### Requirement: Host must discover built-in languages through resource conventions and default configuration

The host system SHALL automatically discover built-in runtime languages from `manifest/i18n/<locale>/*.json` files, and maintain default language, multi-language switch, display sorting, native names, and other metadata through the `i18n` configuration section in the default configuration file. LinaPro default delivery SHALL only include `zh-CN` and `en-US` language resources. Adding a built-in language must only require adding the corresponding runtime JSON, apidoc JSON, plugin JSON, and optional default configuration metadata, without needing to add backend Go language enumerations, SQL seed, or frontend TS language lists. Runtime text direction is fixed as `ltr` per current host convention, without needing to maintain `direction` in configuration. Runtime language list, missing translation checks, resource source diagnostics, runtime translation pack API, ETag negotiation, and frontend persistent cache must automatically cover the default-enabled `zh-CN` and `en-US`.

#### Scenario: Default runtime language list only contains Simplified Chinese and English
- **WHEN** the project starts with default `manifest/i18n` resources and default configuration
- **THEN** the `/i18n/runtime/locales` API returns a language list containing only `zh-CN` and `en-US`
- **AND** `zh-CN` is marked as the default language
- **AND** each language's direction field is the fixed value `ltr`

#### Scenario: Adding a language without Go, SQL, or frontend TS language list changes
- **WHEN** a delivery project adds `manifest/i18n/<locale>/*.json` and `manifest/i18n/<locale>/apidoc/**/*.json` resources for a language
- **AND** source plugins and dynamic plugins add resources for that language following the same directory convention
- **AND** if default language, sorting, native name, or enabled/disabled state needs control, only the `i18n` configuration section in the default configuration file is modified
- **THEN** menus, dictionaries, configurations, scheduled tasks, plugins, roles, system info, and other dynamic metadata automatically return localized results for that language
- **AND** the runtime language list automatically includes that language
- **AND** no modification to backend Go constants, SQL seed, frontend `SUPPORT_LANGUAGES`, or third-party language switching branches is needed

#### Scenario: Disabling multi-language uses only the default language
- **WHEN** `i18n.enabled` is `false` in the default configuration file
- **AND** the user's browser or request parameters pass a non-default language
- **THEN** host request language parsing falls back to `i18n.default`
- **AND** the `/i18n/runtime/locales` response marks the multi-language switch as off, returning only the default language descriptor
- **AND** the default management workbench hides the language switch button, loading static language packs, runtime translation packs, and public frontend configuration in the default language

#### Scenario: Removing a language from the locales list disables it
- **WHEN** the project has multiple `manifest/i18n/<locale>/*.json` resources
- **AND** the default configuration file `i18n.locales` lists only some languages
- **THEN** `/i18n/runtime/locales` returns only languages listed in `i18n.locales`
- **AND** requests for unlisted languages fall back to `i18n.default`

### Requirement: Default management workbench must maintain fixed LTR document direction

The default management workbench SHALL fix document direction as `ltr` per current host convention. When switching languages, the workbench must simultaneously set `<html dir>` to `ltr` and inject `direction="ltr"` into `Ant Design Vue`'s `ConfigProvider`. The frontend must not maintain a static RTL language registry, and adding languages must not require modifying direction-related TypeScript branches.

#### Scenario: html direction remains LTR when switching default built-in languages
- **WHEN** a user switches language to `zh-CN` or `en-US` in the default management workbench
- **THEN** `document.documentElement`'s `dir` attribute remains `ltr`
- **AND** `Ant Design Vue`'s `ConfigProvider` receives `direction="ltr"`

#### Scenario: Default built-in language page text completeness is sufficient
- **WHEN** a user opens framework default delivery list pages, drawers, and dialogs in the default built-in language environment
- **THEN** page text displays in the current language, layout does not block core operations
- **AND** RTL mirrored layout is not needed

### Requirement: Host must provide runtime translation pack distribution capability

The host system SHALL provide runtime translation pack API and language list API, returning aggregated message resources and current available language descriptor information by language, for the default management workbench and host-embedded plugin pages to load. Runtime translation packs must be able to simultaneously include host, source plugin, and currently enabled dynamic plugin i18n messages, converting to nested message objects that the frontend can directly consume. The runtime translation pack API must output an `ETag` header in the response, derived from the current language and runtime translation pack version; the system must accept the `If-None-Match` header from requests, returning `304 Not Modified` without message body when matched. Any sector cache invalidation must trigger automatic runtime translation pack version increment, ensuring different ETag values for different translation pack content in the same language.

#### Scenario: Default workbench loads runtime translation pack
- **WHEN** the frontend requests the `zh-CN` runtime translation pack
- **THEN** the host returns the aggregated message set for that language
- **AND** the result includes merged results from host resources, source plugin resources, and enabled dynamic plugin resources
- **AND** the response includes an `ETag` header

#### Scenario: Frontend obtains host-supported language list
- **WHEN** the frontend requests the runtime language list
- **THEN** the host returns multi-language switch, currently supported language codes, default language marker, display name, native name, and fixed LTR text direction
- **AND** the display name is returned in the current request language, the native name maintains the text of the corresponding language itself
- **AND** the default delivery list only contains `zh-CN` and `en-US`

#### Scenario: Runtime language pack supports layered maintenance while maintaining flat governance
- **WHEN** the host loads translation resources from files, source plugins, or dynamic plugins
- **THEN** the host allows runtime UI file resources to use layered JSON or flat dot-separated key format
- **AND** the host internally unifies messages to flat keys
- **AND** the runtime API returns results to the frontend as nested object structures for direct merging into the frontend `vue-i18n` message tree

#### Scenario: Disabled plugin translation resources are no longer exposed
- **WHEN** a plugin is disabled or uninstalled
- **THEN** subsequent runtime translation pack results no longer include translation messages contributed by that plugin
- **AND** other host and enabled plugin resources remain available
- **AND** the system correspondingly triggers cache invalidation for sectors related to that plugin, automatically incrementing the runtime translation pack version

#### Scenario: Second request for the same translation pack returns 304
- **WHEN** the frontend saved the `ETag` from the first runtime translation pack request
- **AND** no backend cache invalidation occurred between the two requests
- **AND** the frontend carries `If-None-Match` equal to the previous `ETag` in the second request
- **THEN** the backend returns `304 Not Modified` without message body

### Requirement: English regression scan must cover framework delivery pages and seed display content

The default management workbench SHALL provide English regression coverage for framework delivery pages, ensuring system-generated content, default seed display content, and static UI text do not retain Chinese text.

#### Scenario: English regression page does not contain Chinese system text
- **WHEN** an administrator switches to `en-US` and opens the workbench, user management, role management, department management, position management, dictionary management, system configuration, service monitoring, and scheduled tasks
- **THEN** framework delivery titles, buttons, form labels, table columns, system-generated nodes, built-in record displays, and confirmation dialogs use English
- **AND** user-editable business fields are only localized when explicitly included in framework delivery projection rules

#### Scenario: English layout regression screenshot check
- **WHEN** Playwright captures position form, dictionary form, and service monitoring disk table in `en-US`
- **THEN** critical labels, options, headers, and values do not unreadably wrap or overlap
- **AND** screenshot results serve as part of acceptance evidence

#### Scenario: Version info menu title localization consistency
- **WHEN** an administrator views the version info menu item under the development center
- **THEN** Simplified Chinese and English each display the corresponding language title
- **AND** the English title is `Version Info`
