## ADDED Requirements

### Requirement: WASM custom section parsing capability must be centrally provided by pluginbridge
The host system SHALL provide `ReadCustomSection(content []byte, name string) ([]byte, bool, error)` and `ListCustomSections(content []byte) (map[string][]byte, error)` public capabilities in `apps/lina-core/pkg/pluginbridge/pluginbridge_wasm_section.go`, centrally implementing `wasm` file header validation, section traversal and ULEB128 decoding. `apps/lina-core/internal/service/i18n`, `apps/lina-core/internal/service/apidoc`, and plugin runtime MUST read custom sections (such as `i18n_assets`, `apidoc_assets`) from dynamic plugin runtime artifacts through this public capability, and MUST NOT maintain duplicate WASM parsing implementations in business packages. `pluginbridge.WasmSection*` section name constants MUST be centrally maintained by the `pluginbridge` package.

#### Scenario: i18n reads dynamic plugin i18n section via pluginbridge
- **WHEN** the system needs to read the `i18n_assets` custom section from a dynamic plugin runtime artifact
- **THEN** the caller completes this via `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)`
- **AND** no dedicated parsing functions like `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` exist in the `i18n` package

#### Scenario: Fixing WASM parsing defects only requires modifying pluginbridge
- **WHEN** WASM parsing needs to be extended to support new sections, fix decoding bugs, or add boundary checks
- **THEN** modifying `pkg/pluginbridge/pluginbridge_wasm_section.go` in one place is sufficient
- **AND** the `i18n` package and plugin runtime require no duplicate changes
