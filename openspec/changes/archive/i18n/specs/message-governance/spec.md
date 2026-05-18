## ADDED Requirements

### Requirement: Runtime-visible messages must have explicit classification

The system SHALL classify strings in source code by runtime usage surface, and apply different governance strategies for user-visible messages, user deliverables, user display projections, developer diagnostics, operations logs, and user data. User-visible messages, user deliverables, and user display projections MUST be output through runtime i18n resources or backend localization projections; operations logs MUST use stable English and structured fields; user input and external system raw text MUST preserve original values.

#### Scenario: User-visible errors must not return hardcoded Chinese
- **WHEN** a backend business service needs to return a user-visible error
- **THEN** the error MUST carry a stable error code, runtime translation key, parameters, and English source message
- **AND** the unified response MUST output a localized `message` by request language

#### Scenario: Operations logs maintain stable English
- **WHEN** a backend service records logs used only for troubleshooting, metrics, or startup diagnostics
- **THEN** log templates MUST use stable English
- **AND** logs MUST record error codes and key parameters through structured fields

#### Scenario: User data preserves original values
- **WHEN** a string comes from user input, external interface responses, or database business names
- **THEN** the system MUST save and return the string as-is
- **AND** MUST NOT attempt to auto-translate the string as a translation key

### Requirement: Unified response must output structured error fields

The system SHALL extend the unified JSON response error payload so that runtime message errors output machine-readable `errorCode`, `messageKey`, and `messageParams` alongside the localized `message`. Frontend, plugins, and tests MUST use `errorCode` or `messageKey` to determine error semantics.

#### Scenario: Structured business error response includes localized message and stable fields
- **WHEN** a request with `Accept-Language: zh-CN` triggers a `USER_NOT_FOUND` error
- **THEN** the response JSON MUST contain `message` with a Simplified Chinese localized value
- **AND** the response JSON MUST contain `errorCode: "USER_NOT_FOUND"`
- **AND** the response JSON MUST contain the corresponding `messageKey` and `messageParams`

#### Scenario: The same error returns English display text under an English request
- **WHEN** a request with `Accept-Language: en-US` triggers the same error
- **THEN** the `errorCode` and `messageKey` MUST be consistent with the `zh-CN` request
- **AND** the `message` MUST be the English display text

### Requirement: Business error semantics must be governed by module namespace

Backend host, plugin platform components, and source plugins SHALL maintain business error semantic identifiers by business module. Business error definitions MUST be placed in the module's `*_code.go` file. All interface errors that may be returned to HTTP API, plugin calls, or other caller response payloads MUST be created/wrapped through `bizerr.NewCode`, `bizerr.WrapCode`, or equivalent wrappers. The `code` field in HTTP responses MUST use GoFrame `gcode.Code` type error codes; specific business semantics MUST be expressed by `errorCode`, `messageKey`, and `messageParams`.

#### Scenario: Host business module uses independent business semantic namespace
- **WHEN** the user module adds a new `USER_EMAIL_EXISTS` error
- **THEN** the error MUST be defined in `internal/service/user/user_code.go`
- **AND** the response `code` MUST use the closest GoFrame type error code

#### Scenario: Caller-visible errors must carry structured metadata through bizerr
- **WHEN** a controller, middleware, or business service needs to return a user-visible failure reason
- **THEN** the error MUST use a `bizerr.Code` defined in the module's `*_code.go`
- **AND** business paths MUST NOT directly return `gerror.New("...")` as caller-visible interface errors

### Requirement: Backend business errors must use runtime message resources

Backend host and source plugin business errors SHALL use the host or plugin's own runtime language packs to maintain translated text. Host error keys MUST be written to `apps/lina-core/manifest/i18n/<locale>/*.json`; plugin error keys MUST be written to the corresponding plugin's own manifest directory.

#### Scenario: Plugin business error resources belong to the plugin
- **WHEN** the `org-center` plugin adds a new error key
- **THEN** the key MUST be written to `apps/lina-plugins/org-center/manifest/i18n/<locale>/*.json`
- **AND** plugin runtime error keys MUST NOT be centralized into the `lina-core` runtime language pack

### Requirement: Import/export deliverables must render by request language

The system SHALL render Excel exports, import templates, import failure reasons, sheet names, headers, and enum values by request language. The import/export flow MUST parse the locale at the request level and reuse translation results; it MUST NOT repeatedly build runtime language packs inside batch row loops.

#### Scenario: User export outputs English headers under an English request
- **WHEN** a user requests user list export with `Accept-Language: en-US`
- **THEN** the exported Excel headers MUST use English text
- **AND** gender, status, and other enum display MUST use English text

#### Scenario: Export loop must not repeatedly build language packs
- **WHEN** the system exports 10,000 rows of data
- **THEN** translation resources MUST be resolved and cached into the request-level context when entering the export flow
- **AND** the row loop MUST only perform already-resolved translation result lookups

### Requirement: Plugin bridging and host service errors must be stable and localizable

Plugin bridging protocol, WASM host call, host service calls, and plugin manifest validation SHALL return stable status codes or error codes, and provide runtime translation keys for errors that enter admin display. Protocol-layer default developer diagnostic messages MUST use English.

#### Scenario: Host call protocol errors contain stable error codes
- **WHEN** a dynamic plugin calls host service with an illegal network URL
- **THEN** the host call response MUST contain a stable status code or error code
- **AND** the developer diagnostic message MUST use English source text
- **AND** admin display of this error MUST map through `messageKey` to the current language text

### Requirement: Plugin lifecycle and upgrade results must use message keys

Plugin install, uninstall, enable, disable, and upgrade results SHALL return or store stable `messageKey`, `messageParams`, and `errorCode`. `message` MUST be rendered from structured fields by request language.

#### Scenario: Plugin lifecycle failures can be displayed by language
- **WHEN** plugin installation fails and is displayed by the admin console
- **THEN** the backend MUST return a stable error code and translation key
- **AND** the frontend MUST use the current language to display the localized failure reason

### Requirement: Frontend user-visible text must render through i18n

The default management console and plugin frontend SHALL use `$t` or runtime language packs for user-visible text. The frontend request error interceptor MUST prioritize consuming backend-returned `messageKey/messageParams`.

#### Scenario: Monitoring page labels change after language switch
- **WHEN** a user switches from `zh-CN` to `en-US`
- **THEN** server monitoring page labels MUST switch to English
- **AND** the page MUST NOT continue displaying hardcoded Chinese labels

#### Scenario: Request errors prioritize messageKey
- **WHEN** a backend error response contains both `messageKey`, `messageParams`, and `message`
- **THEN** the frontend request interceptor MUST prioritize using `$t(messageKey, messageParams)` to display the error

### Requirement: Backend hardcoded Chinese strings must be classified

The system MUST classify Chinese string literals in backend Go source by caller-visible errors, user-visible projections, user deliverables, developer diagnostics, generated sources, test fixtures, and user-data examples.

#### Scenario: Scan output is classified
- **WHEN** a developer runs the backend hardcoded-Chinese scanner
- **THEN** each finding includes a category or allowlist reason

### Requirement: Plugin-platform developer diagnostics must be stable

Developer diagnostics in plugin bridge, filesystem, database, WASM host service, catalog, and runtime code MUST use stable English source text; if they cross a user or caller boundary, they MUST be wrapped as structured errors.

#### Scenario: Plugin protocol parsing fails
- **WHEN** a plugin bridge codec or host-service codec fails to parse a protocol payload
- **THEN** the internal diagnostic error uses stable English text

### Requirement: Generated schema text must be governed at the generation source

The system MUST NOT hand-edit Chinese comments or `description` tags in generated DAO, DO, or Entity files. When generated schema metadata enters OpenAPI or user-visible documentation, SQL comments or code generation inputs MUST be changed and regenerated.

#### Scenario: Entity descriptions enter API documentation
- **WHEN** generated Entity schema `description` metadata appears in OpenAPI documentation
- **THEN** the corresponding generation source provides English source text

### Requirement: Automated governance must block new high-risk hardcoded messages

The system SHALL provide automated scanning or test gates to identify hardcoded Chinese or mixed Chinese-English text in runtime-visible positions in Go, Vue, and TypeScript. Scanning MUST support an allowlist with classification, reason, and responsible module.

#### Scenario: Go high-risk string scanning fails
- **WHEN** a developer adds `gerror.New("部门不存在")` in production Go code
- **THEN** the hardcoded message scan MUST report this location

#### Scenario: Allowed non-runtime Chinese does not block
- **WHEN** a Chinese string appears only in comments, test fixtures, or user example data
- **THEN** the scan rules allow the string to pass
- **AND** allowlist items MUST record the reason why the string is not a runtime-visible framework message

### Requirement: Localization lookups must satisfy hot-path performance constraints

Runtime error localization, list projection, import/export, and frontend language switching SHALL reuse the existing runtime translation cache. The system MUST NOT clone or rebuild the full language pack for a single error, a single export row, or a single table column render.

#### Scenario: Single error localization only performs cache lookup
- **WHEN** the unified response middleware renders a structured business error
- **THEN** the system MUST look up the corresponding key through the current locale's runtime cache
- **AND** it MUST NOT build the full runtime message pack for that error
