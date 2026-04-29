## ADDED Requirements

### Requirement: API documentation translation resource loading must reuse the unified ResourceLoader
The system SHALL let the `apidoc` translation resource loading pipeline be completed through the unified `ResourceLoader` provided by the `pkg/i18nresource` package, and MUST NOT maintain an independent "host embedded resources -> source plugin embedded resources -> dynamic plugin runtime resources" traversal implementation in the `apidoc` package, and MUST NOT reverse-depend on `internal/service/i18n` to reuse the loader. The `apidoc` pipeline MUST declare `Subdir = "manifest/i18n"`, `LocaleSubdir = "apidoc"`, `LayoutMode = LocaleSubdirectoryRecursive` and `PluginScope = RestrictedToPluginNamespace` through `ResourceLoader` configuration parameters, and retain multi-file, hierarchical JSON and flat dotted key as three maintenance methods; the system still normalizes to stable structured keys during merging.

#### Scenario: apidoc and runtime bundle share resource loader implementation
- **WHEN** the system loads `apidoc` translation resources for a language
- **THEN** the loading process completes host embedded resources, source plugin embedded resources, and dynamic plugin runtime resources discovery and merging through `i18nresource.ResourceLoader`
- **AND** no duplicate directory traversal or `wasm` parsing logic exists in the `apidoc` package compared to the `i18n` package

#### Scenario: Plugin namespace isolation still takes effect
- **WHEN** a source plugin's `apidoc` translation resource declares keys `plugins.<plugin-id>.routes.*`
- **THEN** that plugin's resources are only allowed to contribute keys prefixed with `plugins.<plugin-id>.`
- **AND** the system ignores keys from other plugin namespaces or the host namespace

### Requirement: API documentation must support Traditional Chinese display
The system SHALL support system API documentation display in Traditional Chinese after adding `zh-TW` as a built-in language. The host and all source plugins, dynamic plugins MUST provide `manifest/i18n/zh-TW/apidoc/**/*.json` translation resources, ensuring that API groups, summaries, descriptions, and parameter descriptions returned by `/api.json?lang=zh-TW` are all displayed in Traditional Chinese. `en-US` API documentation continues to directly use English source text from API DTOs, without depending on `zh-TW` resources.

#### Scenario: Loading API documentation in Traditional Chinese environment
- **WHEN** an administrator opens system API documentation in the `zh-TW` environment, or requests `/api.json?lang=zh-TW`
- **THEN** the host, source plugins, and dynamic plugins' route groups, API summaries, API descriptions, request parameter and response parameter descriptions are displayed in Traditional Chinese
- **AND** when translations are missing, falls back to the English source text maintained in API DTOs, without displaying blanks or translation keys
