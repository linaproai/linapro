## MODIFIED Requirements

### Requirement: WASM custom section parsing MUST be provided centrally by pluginbridge

The host system SHALL provide `ReadCustomSection(content []byte, name string) ([]byte, bool, error)` and `ListCustomSections(content []byte) (map[string][]byte, error)` public capabilities through the `apps/lina-core/pkg/pluginbridge` ecosystem, centrally implementing `wasm` file header validation, section traversal, and ULEB128 decoding. This capability may be exposed by the `pluginbridge` root package facade or a responsibility-scoped subcomponent such as `pluginbridge/artifact`, but the protocol implementation must have only one authoritative location. `apps/lina-core/internal/service/i18n`, `apps/lina-core/internal/service/apidoc`, and the plugin runtime must use this public capability to read custom sections (such as `i18n_assets`, `apidoc_assets`) from dynamic plugin runtime artifacts, and must not maintain duplicate WASM parsing implementations in business packages. `pluginbridge.WasmSection*` section name constants or their subcomponent equivalents must be maintained centrally by the pluginbridge ecosystem.

#### Scenario: i18n reads dynamic plugin i18n section through pluginbridge

- **WHEN** The system needs to read the `i18n_assets` custom section from a dynamic plugin runtime artifact
- **THEN** The caller completes the read through `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)` or the equivalent entry point in `pluginbridge/artifact`
- **AND** The `i18n` package does not contain dedicated parsing functions such as `parseWasmCustomSectionsForI18N` or `readWasmULEB128ForI18N`

#### Scenario: WASM parsing defect fixes only require changes to the pluginbridge ecosystem

- **WHEN** WASM parsing needs to be extended to support new sections, fix decoding errors, or add boundary checks
- **THEN** Modifying the authoritative implementation in the corresponding `pkg/pluginbridge` artifact/wasm section subcomponent is sufficient
- **AND** The `i18n` package and plugin runtime do not need duplicate changes
