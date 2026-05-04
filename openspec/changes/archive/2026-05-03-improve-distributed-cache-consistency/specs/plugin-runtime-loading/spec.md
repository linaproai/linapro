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
