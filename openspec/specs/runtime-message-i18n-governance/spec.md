## ADDED Requirements

### Requirement: Runtime-visible messages must have explicit classification

The system SHALL classify strings in source code by runtime usage surface, and apply different governance strategies for user-visible messages, user deliverables, user display projections, developer diagnostics, operations logs, and user data. User-visible messages, user deliverables, and user display projections MUST be output through runtime i18n resources or backend localization projections; operations logs MUST use stable English and structured fields; user input and external system raw text MUST preserve original values and MUST NOT be auto-translated.

#### Scenario: User-visible errors must not return hardcoded Chinese

- **WHEN** a backend business service needs to return a user-visible error
- **THEN** the error MUST carry a stable error code, runtime translation key, parameters, and English source message
- **AND** the unified response MUST output a localized `message` by request language
- **AND** business services MUST NOT construct errors through Chinese free text like `gerror.New("字典类型不存在")`

#### Scenario: Operations logs maintain stable English

- **WHEN** a backend service records logs used only for troubleshooting, metrics, or startup diagnostics
- **THEN** log templates MUST use stable English
- **AND** logs MUST record error codes, resource IDs, plugin IDs, paths, or underlying errors through structured fields
- **AND** logs MUST NOT depend on the current request language to generate localized text

#### Scenario: User data preserves original values

- **WHEN** a string comes from user input, external interface responses, database business names, or file content
- **THEN** the system MUST save and return the string as-is
- **AND** the system MUST NOT attempt to auto-translate the string as a translation key

### Requirement: Unified response must output structured error fields

The system SHALL extend the unified JSON response error payload so that runtime message errors output machine-readable `errorCode`, `messageKey`, and `messageParams` alongside the localized `message`. Frontend, plugins, and tests MUST use `errorCode` or `messageKey` to determine error semantics and MUST NOT rely on the natural language `message` for business logic.

#### Scenario: Structured business error response includes localized message and stable fields

- **WHEN** a request with `Accept-Language: zh-CN` triggers a `USER_NOT_FOUND` error
- **THEN** the response JSON MUST contain `message` with a Simplified Chinese localized value
- **AND** the response JSON MUST contain `errorCode: "USER_NOT_FOUND"`
- **AND** the response JSON MUST contain the corresponding `messageKey`
- **AND** the response JSON MUST contain `messageParams` used for message formatting

#### Scenario: The same error returns English display text under an English request

- **WHEN** a request with `Accept-Language: en-US` triggers the same `USER_NOT_FOUND` error
- **THEN** the `errorCode` and `messageKey` in the response JSON MUST be consistent with the `zh-CN` request
- **AND** the `message` in the response JSON MUST be the English display text
- **AND** the response MUST NOT contain hardcoded Chinese fallback text

#### Scenario: Unstructured errors still fall back through existing localization

- **WHEN** legacy code or a third-party library returns an unstructured error
- **THEN** the unified response MUST continue attempting translation through the existing `LocalizeError` logic
- **AND** if translation is not possible, the system MUST return a controlled fallback message
- **AND** new business errors MUST NOT continue to use unstructured free text as the primary implementation

### Requirement: Backend business errors must use runtime message resources

Backend host and source plugin business errors SHALL use the host or plugin's own runtime language packs to maintain translated text. Host error keys MUST be written to `apps/lina-core/manifest/i18n/<locale>/*.json`; plugin error keys MUST be written to the corresponding `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/*.json`. Runtime error messages MUST NOT be written to or reuse `manifest/i18n/<locale>/apidoc` resources.

#### Scenario: Host business error resources are complete

- **WHEN** the host adds a new `error.dict.type.exists` error key
- **THEN** the `zh-CN`, `en-US`, and `zh-TW` host runtime language packs MUST all contain the key
- **AND** the missing translation check MUST fail when any target language is missing

#### Scenario: Plugin business error resources belong to the plugin

- **WHEN** the `org-center` plugin adds a new `plugin.org-center.error.deptNotFound` error key
- **THEN** the key MUST be written to `apps/lina-plugins/org-center/manifest/i18n/<locale>/*.json`
- **AND** plugin runtime error keys MUST NOT be centralized into the `lina-core` runtime language pack

### Requirement: Business error semantics must be governed by module namespace

Backend host, plugin platform components, and source plugins SHALL maintain business error semantic identifiers by business module. Business error definitions MUST be placed in the module's `*_code.go` file; business implementations MUST only reference `bizerr.Code` variables defined in that file and MUST NOT hardcode machine error codes, translation keys, or borrow error definitions from other modules at business call sites. All interface errors that may be returned to HTTP API, plugin calls, source plugin backend interfaces, WASM host service, or other caller response payloads MUST be created/wrapped through `bizerr.NewCode`, `bizerr.WrapCode`, or equivalent wrappers. The `code` field in HTTP responses MUST use GoFrame `gcode.Code` type error codes to express the error category; specific business semantics MUST be expressed by `errorCode`, `messageKey`, and `messageParams`.

#### Scenario: Host business module uses independent business semantic namespace

- **WHEN** the user module adds a new `USER_EMAIL_EXISTS` error
- **THEN** the error MUST be defined in `internal/service/user/user_code.go`
- **AND** the `errorCode` MUST use the user module prefix, e.g. `USER_EMAIL_EXISTS`
- **AND** the response `code` MUST use the closest GoFrame type error code, e.g. `gcode.CodeInvalidParameter`
- **AND** other modules such as role, menu, dictionary MUST NOT reuse this business semantic identifier

#### Scenario: Error code definitions are decoupled from business usage

- **WHEN** dictionary module business code needs to return a "dictionary type already exists" error
- **THEN** the business code MUST use `bizerr.NewCode(CodeDictTypeExists)` or equivalent wrapper
- **AND** the business code MUST NOT write `"DICT_TYPE_EXISTS"`, `"error.dict.type.exists"`, or raw numeric error codes
- **AND** the response `code` MUST come from the GoFrame type error code bound to that definition

#### Scenario: Caller-visible errors must carry structured metadata through bizerr

- **WHEN** a controller, middleware, business service, or plugin host service needs to return a parameter error, authentication error, business validation failure, or user-visible failure reason
- **THEN** the error MUST use a `bizerr.Code` defined in the module's `*_code.go`
- **AND** the error MUST be returned through `bizerr.NewCode`, `bizerr.WrapCode`, or equivalent wrapper
- **AND** business paths MUST NOT directly return `gerror.New("请选择要导入的文件")`, `gerror.NewCode(gcode.CodeInvalidParameter, "error.xxx")`, `errors.New(...)`, or `fmt.Errorf(...)` as caller-visible interface errors
- **AND** low-level technical errors are only allowed as causes if they are wrapped by `bizerr.WrapCode` into business semantic errors before reaching the return boundary

#### Scenario: Source plugin internal modules are also divided by namespace

- **WHEN** the `org-center` plugin adds department module and post module errors
- **THEN** department module errors MUST use `ORG_CENTER_DEPT_*` or equivalent module prefix
- **AND** post module errors MUST use `ORG_CENTER_POST_*` or equivalent module prefix
- **AND** plugin error definitions MUST NOT be written into host `lina-core` error definition files

### Requirement: Import/export deliverables must render by request language

The system SHALL render Excel exports, import templates, import failure reasons, sheet names, headers, statuses, genders, operation types, and operation results — and other user deliverable text — by request language. The import/export flow MUST parse the locale at the request level and reuse the translation results needed by the module; it MUST NOT repeatedly build runtime language packs inside batch row loops.

#### Scenario: User export outputs English headers under an English request

- **WHEN** a user requests user list export with `Accept-Language: en-US`
- **THEN** the exported Excel headers MUST use English text
- **AND** gender, status, and other enum display MUST use English text
- **AND** the exported file MUST NOT contain hardcoded Chinese headers or statuses like `用户名`, `正常`, `停用`

#### Scenario: Dictionary import failure reasons return in request language

- **WHEN** a user imports an Excel file containing invalid dictionary types with `Accept-Language: zh-TW`
- **THEN** the failure reasons in the import results MUST use Traditional Chinese text
- **AND** failure reasons MUST preserve parameters such as row number, field name, and invalid value
- **AND** underlying technical errors MUST NOT be directly concatenated into user display text causing mixed Chinese-English

#### Scenario: Export loop must not repeatedly build language packs

- **WHEN** the system exports 10,000 rows of data
- **THEN** translation resources MUST be resolved and cached into the request-level context by module or key set when entering the export flow
- **AND** the row loop MUST only perform already-resolved translation result lookups, parameter formatting, or user data writing

### Requirement: Plugin bridging and host service errors must be stable and localizable

Plugin bridging protocol, WASM host call, host service calls, plugin manifest validation, plugin resource validation, and dynamic plugin runtime errors SHALL return stable status codes or error codes, and provide runtime translation keys for errors that enter admin display. Protocol-layer default developer diagnostic messages MUST use English, and admin display MUST be localized by request language.

#### Scenario: Host call protocol errors contain stable error codes

- **WHEN** a dynamic plugin calls host service with an illegal network URL
- **THEN** the host call response MUST contain a stable status code or error code
- **AND** the developer diagnostic message MUST use English source text
- **AND** admin display of this error MUST map through `messageKey` or error code to the current language text

#### Scenario: Plugin manifest validation errors no longer use mixed Chinese-English

- **WHEN** plugin manifest validation finds an illegal menu key
- **THEN** the validation error MUST use a stable error code and parameters describing the plugin ID, field name, actual value, and expected rule
- **AND** user-visible display MUST be generated through runtime translation resources
- **AND** it MUST NOT return mixed Chinese-English free text like `插件菜单 key 必须使用当前插件前缀 plugin:<id>:*`

### Requirement: Plugin lifecycle and upgrade results must use message keys

Plugin install, uninstall, enable, disable, auto-enable, source plugin upgrade, and dynamic plugin publish governance results SHALL return or store stable `messageKey`, `messageParams`, and `errorCode`. If the interface still needs a simple display field, `message` MUST be rendered from structured fields by request language or command locale.

#### Scenario: Source plugin returns structured result when no upgrade is needed

- **WHEN** the source plugin's current effective version equals the discovered version
- **THEN** the upgrade result MUST contain a stable `messageKey` indicating no upgrade is needed
- **AND** the result MUST contain parameters such as plugin ID, current version, and discovered version
- **AND** it MUST NOT only return a fixed Chinese string like `当前源码插件已是最新版本，无需升级。`

#### Scenario: Plugin lifecycle failures can be displayed by language

- **WHEN** plugin installation fails and is displayed by the admin console
- **THEN** the backend MUST return a stable error code and translation key
- **AND** the frontend MUST use the current language to display the localized failure reason
- **AND** logs MUST retain English diagnostics and structured parameters for troubleshooting

### Requirement: Frontend user-visible text must render through i18n

The default management console and plugin frontend SHALL use `$t` or runtime language packs for user-visible text such as page titles, form labels, table columns, empty states, prompts, confirm dialogs, toasts, tooltips, and time units. The frontend request error interceptor MUST prioritize consuming backend-returned `messageKey/messageParams`, and only then fall back to the backend's already-localized `message`.

#### Scenario: Monitoring page labels change after language switch

- **WHEN** a user switches from `zh-CN` to `en-US`
- **THEN** database info, server info, service info, disk column names, empty states, and time units on the server monitoring page MUST switch to English
- **AND** the page MUST NOT continue displaying hardcoded Chinese labels

#### Scenario: Online users page column names use translation keys

- **WHEN** a user opens the online users page
- **THEN** query form labels and table column headers MUST render through `$t` or runtime language packs
- **AND** `data.ts` MUST NOT retain hardcoded Chinese labels like `用户账号`, `登录账号`, `部门名称`

#### Scenario: Request errors prioritize messageKey

- **WHEN** a backend error response contains both `messageKey`, `messageParams`, and `message`
- **THEN** the frontend request interceptor MUST prioritize using `$t(messageKey, messageParams)` to display the error
- **AND** when the frontend lacks the translation key, it MUST fall back to the backend `message`

### Requirement: Audit and operation display must store stable semantics and project by language

Operation logs, login logs, task logs, notification messages, plugin upgrade results, and other user-facing or exportable run records SHALL preferentially store stable type codes, status codes, translation keys, and parameters. Lists, details, exports, and message previews MUST project display text by request language. The system MUST NOT persist only already-localized Chinese display values as the sole semantic source.

#### Scenario: Operation log export displays operation type by request language

- **WHEN** a user exports operation logs with `en-US`
- **THEN** operation type and operation status MUST render in English based on stable codes
- **AND** the export MUST NOT depend on persisted Chinese strings like `导出`, `成功`, `失败`

#### Scenario: Task log details preserve machine semantics and localized display

- **WHEN** a user views task log details
- **THEN** the backend or frontend MUST use task status code, handler key, and error code to project display text
- **AND** original stdout/stderr or user script output MUST preserve original values

### Requirement: Automated governance must block new high-risk hardcoded messages

The system SHALL provide automated scanning or test gates to identify hardcoded Chinese or mixed Chinese-English text in runtime-visible positions in Go, Vue, and TypeScript. Scanning MUST support an allowlist, but each exception MUST state the classification, reason, and responsible module. When adding or modifying runtime translation keys, the missing translation check MUST cover all enabled built-in languages.

#### Scenario: Go high-risk string scanning fails

- **WHEN** a developer adds `gerror.New("部门不存在")` in production Go code
- **THEN** the hardcoded message scan MUST report this location
- **AND** the check MUST require changing to a structured error with a runtime translation key

#### Scenario: Frontend high-risk string scanning fails

- **WHEN** a developer adds `title: '操作'` in a Vue page table column
- **THEN** the frontend scan MUST report this location
- **AND** the check MUST require changing to `$t` or a runtime language pack key

#### Scenario: Allowed non-runtime Chinese does not block

- **WHEN** a Chinese string appears only in comments, test fixtures, user example data, or explicitly annotated allowlist items
- **THEN** the scan rules allow the string to pass
- **AND** allowlist items MUST record the reason why the string is not a runtime-visible framework message

### Requirement: Localization lookups must satisfy hot-path performance constraints

Runtime error localization, list projection, import/export, and frontend language switching SHALL reuse the existing runtime translation cache. The system MUST NOT clone or rebuild the full language pack for a single error, a single export row, a single table column render, or a single log projection.

#### Scenario: Single error localization only performs cache lookup

- **WHEN** the unified response middleware renders a structured business error
- **THEN** the system MUST look up the corresponding key through the current locale's runtime cache
- **AND** it MUST NOT build the full runtime message pack for that error

#### Scenario: Batch list projection reuses the same request language context

- **WHEN** the backend returns an operation log list containing 1,000 records
- **THEN** the system MUST reuse the locale and translation lookup capability within the same request
- **AND** it MUST NOT repeatedly parse the request language or load language resources for each record
