## ADDED Requirements

### Requirement: Dynamic plugin runtime derived caches must invalidate across nodes

After dynamic plugin install, enable, disable, uninstall, upgrade, or same-version refresh, the system SHALL use the unified cache coordination mechanism to invalidate or refresh plugin runtime derived caches on all nodes.

#### Scenario: Non-primary node observes plugin runtime revision change

- **WHEN** the primary node completes a dynamic plugin runtime state transition in cluster mode and publishes a plugin runtime cache revision
- **THEN** non-primary nodes observe the new revision on the next request path or watcher path
- **AND** non-primary nodes refresh the plugin enabled snapshot
- **AND** non-primary nodes invalidate plugin frontend bundle, runtime i18n bundle, and Wasm compilation cache

#### Scenario: Non-primary node does not keep exposing capability after plugin disable

- **WHEN** a dynamic plugin is disabled or uninstalled on the primary node
- **THEN** non-primary nodes MUST NOT continue exposing that plugin's menu, frontend assets, or dynamic route capabilities from stale local cache beyond the staleness window allowed by the plugin runtime cache domain

### Requirement: Wasm compilation cache must bind to artifact checksum or generation

The system SHALL bind dynamic-plugin Wasm compilation cache to the artifact checksum or generation of the current active release. It MUST NOT decide cache reuse only by mutable artifact path.

#### Scenario: Same-version dynamic plugin refresh recompiles

- **WHEN** a dynamic plugin is refreshed with the same version but a changed artifact checksum
- **THEN** after nodes observe the plugin runtime revision change, they MUST NOT keep hitting the Wasm compilation cache for the old checksum
- **AND** the next dynamic route or dynamic task execution must compile or load from the new artifact

#### Scenario: Same artifact path but different checksum

- **WHEN** the active release artifact path is the same as the old cache path but the checksum differs
- **THEN** the system treats it as a different compilation cache entry
- **AND** the old entry must be invalidated or naturally cleaned up

### Requirement: Dynamic plugin artifact archive must support same-version refresh consistency

The system SHALL ensure that the active release after same-version refresh points to verifiable new artifact content, and that other nodes can use shared release state to decide whether local cache is stale.

#### Scenario: Same-version refresh writes new artifact

- **WHEN** a plugin same-version refresh commits new artifact content
- **THEN** the system updates the active release checksum or generation
- **AND** publishes a plugin runtime cache revision
- **AND** other nodes can use the active release checksum or generation to decide whether local cache needs rebuilding

#### Scenario: Old artifact cleanup does not affect current active release

- **WHEN** the system cleans old dynamic plugin artifacts
- **THEN** the artifact referenced by the current active release MUST NOT be deleted
- **AND** artifacts still referenced by local caches but no longer active can be cleaned later according to the retention policy

## MODIFIED Requirements

### Requirement: WASM custom section parsing capability must be centrally provided through pluginbridge

The host system SHALL provide `ReadCustomSection(content []byte, name string) ([]byte, bool, error)` and `ListCustomSections(content []byte) (map[string][]byte, error)` public capabilities through the `apps/lina-core/pkg/pluginbridge` system, centrally implementing `wasm` file header validation, section traversal, and ULEB128 decoding. This capability may be exposed through the `pluginbridge` root package facade or through responsibility-scoped subcomponents such as `pluginbridge/artifact`, but the protocol implementation must have only one authoritative location. `apps/lina-core/internal/service/i18n`, `apps/lina-core/internal/service/apidoc`, and the plugin runtime must use this public capability to read custom sections (such as `i18n_assets`, `apidoc_assets`) from dynamic plugin runtime artifacts, and MUST NOT maintain duplicate WASM parsing implementations in business packages. `pluginbridge.WasmSection*` section name constants or their subcomponent equivalents must be centrally maintained by the pluginbridge system.

#### Scenario: i18n reads dynamic plugin i18n section through pluginbridge

- **WHEN** the system needs to read the `i18n_assets` custom section from a dynamic plugin runtime artifact
- **THEN** the caller completes the read through `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)` or the equivalent entry in `pluginbridge/artifact`
- **AND** the `i18n` package does not contain dedicated parsing functions such as `parseWasmCustomSectionsForI18N` or `readWasmULEB128ForI18N`

#### Scenario: Fixing WASM parsing defects only requires modifying the pluginbridge system

- **WHEN** WASM parsing needs to be extended to support new sections, fix decoding errors, or add boundary checks
- **THEN** modifying the authoritative implementation in `pkg/pluginbridge`'s corresponding artifact/wasm section subcomponent is sufficient
- **AND** the `i18n` package and plugin runtime do not need duplicate changes
