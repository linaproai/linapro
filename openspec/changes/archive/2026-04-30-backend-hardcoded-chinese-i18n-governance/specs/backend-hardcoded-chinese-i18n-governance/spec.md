## ADDED Requirements

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

### Requirement: Plugin errors and language resources must belong to the plugin

When a source-plugin backend adds or changes user-visible errors, export text, page summaries, demo prompts, or business projection text, the corresponding runtime i18n resources MUST live under that plugin's own `manifest/i18n/<locale>/*.json` files.

#### Scenario: Plugin business errors are localized by plugin resources
- **WHEN** a source plugin adds business errors
- **THEN** error definitions use stable plugin-namespace codes and message keys
- **AND** `zh-CN`, `en-US`, and `zh-TW` resources are maintained in that plugin directory

### Requirement: User-visible projections and deliverables must render by language

Backend-owned display fields, export headers, import-template fields, import failure reasons, status fallbacks, and runtime configuration display reasons MUST render by request language or return structured values that the frontend can render.

#### Scenario: Export files vary by language
- **WHEN** a user triggers the same export API in different runtime languages
- **THEN** system headers and enum display values in the exported file use the current request language
- **AND** user data remains unchanged

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
- **AND** the apidoc service does not maintain temporary Chinese-to-English conversion tables

### Requirement: Regression scans must cover high-risk positions

The project MUST provide backend hardcoded-Chinese scanner gates for caller-visible errors, user-visible fields, export headers, status labels, plugin diagnostics, and structured error fallbacks.

#### Scenario: New Chinese gerror text is blocked
- **WHEN** a developer adds Chinese `gerror.New*`, `gerror.Wrap*`, `errors.New`, or `fmt.Errorf` text in handwritten non-test Go files
- **THEN** the scanner reports the new item as a violation
