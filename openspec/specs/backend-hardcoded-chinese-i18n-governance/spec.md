# Backend Hardcoded Chinese I18n Governance

## Purpose

Define how backend Chinese string literals are identified, classified, remediated, and guarded so caller-visible behavior, exported artifacts, API documentation, and developer diagnostics remain localizable and stable.

## Requirements

### Requirement: Backend hardcoded Chinese strings must be classified

The system MUST classify Chinese string literals in backend Go source by caller-visible errors, user-visible projections, user deliverables, developer diagnostics, generated sources, test fixtures, and user-data examples.

#### Scenario: Scan output is classified
- **WHEN** a developer runs the backend hardcoded-Chinese scanner
- **THEN** each finding includes a category or allowlist reason
- **AND** uncategorized handwritten non-test Go Chinese strings are reported as issues to handle

#### Scenario: Tests and generated files do not block business cleanup
- **WHEN** the scanner finds Chinese text in `_test.go`, `internal/dao`, `internal/model/do`, or `internal/model/entity`
- **THEN** those findings are counted separately
- **AND** developers MUST NOT hand-edit generated DAO, DO, or Entity files

### Requirement: Caller-visible errors must be structured

Business, authorization, validation, and user-visible failure reasons that can reach HTTP APIs, source-plugin APIs, dynamic-plugin routes, WASM host services, plugin host services, or unified response payloads MUST use `bizerr` or an equivalent structured error format.

#### Scenario: Business error reaches an HTTP caller
- **WHEN** a backend service returns a caller-visible error
- **THEN** the unified response contains a stable `errorCode`
- **AND** the unified response contains a `messageKey`
- **AND** the unified response localizes `message` by request language
- **AND** the call site MUST NOT directly return Chinese `gerror.New*`, `errors.New`, or `fmt.Errorf` text

#### Scenario: Module defines its own error codes
- **WHEN** a module adds a visible business error
- **THEN** the error is defined in the module's own `*_code.go`
- **AND** the definition includes an English fallback, stable machine error code, and runtime i18n key
- **AND** call sites create or wrap errors through the defined code variable

### Requirement: Plugin errors and language resources must belong to the plugin

When a source-plugin backend adds or changes user-visible errors, export text, page summaries, demo prompts, or business projection text, the corresponding runtime i18n resources MUST live under that plugin's own `manifest/i18n/<locale>/*.json` files.

#### Scenario: Plugin business errors are localized by plugin resources
- **WHEN** `org-center`, `content-notice`, `monitor-loginlog`, `monitor-operlog`, `plugin-demo-source`, `plugin-demo-dynamic`, or another source plugin adds business errors
- **THEN** the error definitions use stable plugin-namespace codes and message keys
- **AND** `zh-CN`, `en-US`, and `zh-TW` translation resources are maintained in that plugin directory
- **AND** lina-core runtime language packs MUST NOT centrally own plugin business-error translations

#### Scenario: Plugin exported artifacts are localized
- **WHEN** a plugin generates Excel, CSV, or another user deliverable
- **THEN** headers, sheet names, and enum display values use plugin runtime i18n resources
- **AND** user-entered or database business content is exported unchanged

### Requirement: User-visible projections and deliverables must render by language

Backend-owned user-visible display fields, export headers, import-template fields, import failure reasons, status fallbacks, and runtime configuration display reasons MUST render by request language or return structured values that the frontend can render.

#### Scenario: Export files vary by language
- **WHEN** a user triggers the same export API in different runtime languages
- **THEN** system headers and system enum display values in the exported file use the current request language
- **AND** user data in the exported file remains unchanged

#### Scenario: Backend projection fields vary by language
- **WHEN** a user requests department trees, post trees, system information, or runtime configuration display fields
- **THEN** backend-generated labels, units, and reason text use the current request language
- **OR** the backend returns structured values and codes that the frontend renders by language

### Requirement: Plugin-platform developer diagnostics must be stable

Developer diagnostic errors in plugin bridges, plugin file systems, plugin database services, WASM host services, and plugin catalog or runtime validation MUST use stable English source text; if such errors cross into user UI or caller response boundaries, they MUST be wrapped as structured errors or structured plugin error payloads.

#### Scenario: Plugin protocol parsing fails
- **WHEN** a plugin bridge codec or host-service codec fails to parse a protocol payload
- **THEN** the internal diagnostic error uses stable English text
- **AND** protocol callers MUST NOT rely on localized natural-language text to identify error types

#### Scenario: Plugin management API exposes platform errors
- **WHEN** a plugin-platform internal error reaches a plugin management API response
- **THEN** the response carries a stable error code or message key
- **AND** the user-facing display text can be localized by runtime language

### Requirement: Generated schema text must be governed at the generation source

The system MUST NOT hand-edit Chinese comments or `description` tags in generated DAO, DO, or Entity files. When generated schema metadata enters OpenAPI or user-visible documentation, the corresponding SQL comments or code generation inputs MUST be changed and regenerated.

#### Scenario: Entity descriptions enter API documentation
- **WHEN** generated Entity schema `description` metadata appears in OpenAPI documentation
- **THEN** the corresponding SQL table or field comments, or generation source, MUST provide English source text
- **AND** the apidoc service MUST NOT maintain temporary Chinese-to-English conversion tables

#### Scenario: DAO artifacts are regenerated
- **WHEN** SQL comments or generation inputs change to govern schema text
- **THEN** developers run the repository's DAO generation flow
- **AND** generated results remain reproducible without manual edits

### Requirement: Regression scans must cover high-risk positions

The project MUST provide backend runtime hardcoded-Chinese scanner gates for high-risk positions including caller-visible errors, user-visible fields, export headers, status labels, plugin diagnostics, and structured error fallbacks.

#### Scenario: New Chinese gerror text is blocked
- **WHEN** a developer adds Chinese `gerror.New*`, `gerror.Wrap*`, `errors.New`, or `fmt.Errorf` text in handwritten non-test Go files
- **THEN** the scanner reports the new item as a violation
- **AND** the violation MUST be converted to `bizerr`, stable English developer diagnostics, or a justified allowlist entry

#### Scenario: Allowlist entries explain their boundary
- **WHEN** a Chinese string is allowed to remain
- **THEN** the allowlist records the file, category, retention reason, and applicability scope
- **AND** user-visible errors and user deliverable text MUST NOT be exempted only through the allowlist

