## ADDED Requirements

### Requirement: API documentation must use English source copy and independent apidoc translation resources
The host system SHALL maintain readable English OpenAPI documentation source text in host, source-plugin, and dynamic-plugin API DTOs, including route groups, summaries, descriptions, request parameter descriptions, response parameter descriptions, and fixed route projection copy. API documentation localization SHALL run while rendering `/api.json` according to the current request language, using independent `manifest/i18n/<locale>/apidoc/**/*.json` resources decoupled from runtime UI resources. Host API translations SHALL be maintained under `apps/lina-core/manifest/i18n/<locale>/apidoc/`. Plugin API translations SHALL be maintained independently by each plugin under `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/apidoc/`. The lina-core apidoc module SHALL discover and merge these resources at render time. English API documentation SHALL directly use English source copy, and every `manifest/i18n/en-US/apidoc/**/*.json` file SHALL remain an empty-object placeholder.

#### Scenario: Chinese API documentation is projected through apidoc JSON
- **WHEN** an administrator requests `/api.json?lang=zh-CN`
- **THEN** English source copy maintained in API DTOs is mapped to Chinese through stable structured keys in `manifest/i18n/zh-CN/apidoc/**/*.json`
- **AND** plugin API translation keys are provided by each plugin's own apidoc i18n JSON

#### Scenario: API copy changes validate apidoc translation coverage
- **WHEN** a developer adds or modifies OpenAPI documentation tags
- **THEN** the developer must update the owning host or plugin non-English apidoc translation resources
- **AND** automated tests or review rules must block missing non-English translations

### Requirement: API documentation translation resource loading must reuse the unified ResourceLoader
The system SHALL let apidoc translation resource loading be completed through the unified `ResourceLoader` in `pkg/i18nresource/`, with `Subdir="manifest/i18n"`, `LocaleSubdir="apidoc"`, `LayoutMode=LocaleSubdirectoryRecursive`, and `PluginScope=RestrictedToPluginNamespace`.

#### Scenario: apidoc and runtime bundle share resource loader implementation
- **WHEN** the system loads apidoc translation resources
- **THEN** the loading process completes through `i18nresource.ResourceLoader`
- **AND** no duplicate directory traversal or `wasm` parsing logic exists compared to the `i18n` package

### Requirement: API documentation metadata must align with project positioning
The system SHALL ensure that OpenAPI titles, descriptions, and page introductions used by the system API documentation page align with LinaPro's unified project positioning.

#### Scenario: Generated OpenAPI metadata
- **WHEN** the host generates or reads OpenAPI document titles and descriptions
- **THEN** titles and descriptions use semantics consistent with the LinaPro project positioning
- **AND** no longer use "LinaPro Admin API", "backend management system API documentation", or equivalent expressions
