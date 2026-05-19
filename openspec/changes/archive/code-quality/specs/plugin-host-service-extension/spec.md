## ADDED Requirements

### Requirement: Plugin host service adapters must be uniformly constructed by the host runtime

The system SHALL have the host runtime uniformly construct and publish source-plugin and dynamic-plugin host service adapters. Adapters MUST reuse startup-shared host service instances or shared backends, and MUST NOT construct isolated host service graphs on their own in plugin call paths.

#### Scenario: Source plugin uses host service adapters

- **WHEN** a source plugin calls host capabilities published by `pkg/pluginservice/*`
- **THEN** the capability adapter is constructed by the host runtime and passed to the plugin through the registrar
- **AND** the adapter reuses host-shared auth, session, notify, config, i18n, pluginstate, or other dependencies
- **AND** plugin production paths MUST NOT create the adapter with no parameters

#### Scenario: Dynamic plugin host service calls shared host capabilities

- **WHEN** a dynamic plugin calls cache, lock, notify, config, runtime, storage, or data host capabilities through the unified host service protocol
- **THEN** the host service handler uses shared host services or shared backends injected by the plugin runtime
- **AND** the handler MUST NOT create independent cache, lock, notify, config, or plugin service instances on each call

#### Scenario: WASM host service configuration entry is injected during startup

- **WHEN** the host starts and initializes WASM host services
- **THEN** the startup path explicitly configures shared dependencies for cache, lock, notify, storage, config, and runtime host services
- **AND** package-level default instances MUST NOT continue serving as actual runtime dependencies after production startup
