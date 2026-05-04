# plugin-config-service Specification

## Purpose

Define the general read-only configuration access contract exposed by LinaPro to source plugins and dynamic plugins. The capability lets plugins read static host configuration through business-neutral APIs while keeping plugin-specific configuration structures, defaults, and validation inside each plugin.

## Requirements
### Requirement: Plugin Configuration Service Provides General Read-Only Configuration Access

The system SHALL provide a business-neutral read-only configuration access service to source plugins through `apps/lina-core/pkg/pluginservice/config`. The service MUST allow source plugins to read host configuration file content by arbitrary configuration key and MUST NOT expose plugin-specific `GetXxx()` configuration methods for specific plugins or business modules.

#### Scenario: Plugin reads an arbitrary configuration key

- **WHEN** a source plugin reads an existing configuration key through the plugin configuration service
- **THEN** the system returns the configuration value for that key
- **AND** the configuration service does not require the key to be under a plugin-specific prefix

#### Scenario: Public component contains no plugin business configuration methods

- **WHEN** a private configuration structure is added or modified for a source plugin
- **THEN** developers define the configuration structure, defaults, and validation logic inside the plugin
- **AND** no plugin-specific `GetXxx()` method or plugin business configuration type needs to be added to `apps/lina-core/pkg/pluginservice/config`

### Requirement: Plugin Configuration Service Supports Struct Scanning and Basic Type Reads

The system SHALL support source plugins scanning a configuration section into a caller-provided struct through the generic configuration service, and SHALL support reading common types including strings, booleans, integers, and `time.Duration`. Missing or blank configuration keys MUST use the default value provided by the caller.

#### Scenario: Plugin scans a configuration section into a private struct

- **WHEN** a source plugin calls the configuration service to scan an existing configuration section
- **THEN** the system binds that section to the struct instance provided by the plugin
- **AND** the plugin can apply its own defaults and business validation after scanning

#### Scenario: Plugin reads basic type configuration with defaults

- **WHEN** a source plugin reads a missing or blank string, boolean, integer, or duration configuration key
- **THEN** the system returns the default value provided by the caller
- **AND** the missing key does not cause a failure

#### Scenario: Duration configuration parsing fails

- **WHEN** a source plugin reads a duration configuration key whose value is not a valid duration string
- **THEN** the system returns an explicit error
- **AND** the plugin caller can choose fail-fast behavior, fallback behavior, or upstream error wrapping

### Requirement: Plugin Configuration Service Preserves a Read-Only Boundary

The system SHALL limit the plugin configuration service to read-only access. The service MUST NOT expose methods for writing, saving, hot reloading, or mutating configuration at runtime.

#### Scenario: Plugin can only read configuration

- **WHEN** a source plugin depends on `apps/lina-core/pkg/pluginservice/config`
- **THEN** the public service provides only methods for reading, scanning, and parsing configuration
- **AND** it provides no capability to modify configuration files or system runtime configuration

### Requirement: Plugin Business Configuration Is Maintained Inside the Plugin

The system SHALL require each plugin's configuration structure, defaults, validation, and business semantics to be maintained inside that plugin instead of in the host generic configuration service.

#### Scenario: Monitor server plugin loads monitor configuration

- **WHEN** the `monitor-server` source plugin needs to read the monitor collection interval and retention multiplier
- **THEN** the plugin reads the existing monitor configuration keys through the generic configuration service
- **AND** the plugin maintains the monitor configuration structure, defaults, and validation logic inside itself
- **AND** the host generic configuration service does not expose `MonitorConfig` or `GetMonitor()`

### Requirement: Plugin Configuration Reads Do Not Add Distributed Cache Consistency Burden

The system SHALL limit this plugin configuration service capability to static configuration reads and SHALL NOT add a runtime mutable configuration cache. If the capability is later extended to runtime mutable configuration reads, the system MUST separately design cross-instance revision numbers, broadcast invalidation, shared cache, or an equivalent cluster consistency mechanism.

#### Scenario: Read static configuration files

- **WHEN** a source plugin reads static configuration file content through the plugin configuration service
- **THEN** the system can reuse the host's existing configuration read mechanism
- **AND** no new distributed cache invalidation or cross-instance refresh mechanism is required

#### Scenario: Future runtime mutable configuration extension

- **WHEN** the plugin configuration service needs to read configuration data that can be modified at runtime in the future
- **THEN** the design must define the authoritative data source, consistency model, invalidation trigger, cross-instance synchronization mechanism, and failure fallback strategy
- **AND** it must not rely only on single-node in-process cache to guarantee cluster consistency

### Requirement: Dynamic Plugins Read Complete Static Configuration Through the Config Host Service

The system SHALL provide the `config` host service through the dynamic plugin `hostServices` authorization model so dynamic plugins can read host GoFrame static configuration content through `lina_env.host_call`. The host service MUST allow complete configuration reads, MUST NOT require configuration keys to be under plugin-specific prefixes or resource allowlists, and MUST remain read-only.

#### Scenario: Dynamic plugin declares configuration read service

- **WHEN** a dynamic plugin declares `service: config` in `plugin.yaml` and `methods` is any non-empty subset of `get`, `exists`, `string`, `bool`, `int`, and `duration`
- **THEN** the system accepts the host service declaration and derives configuration read capability from it
- **AND** the plugin runtime authorization snapshot includes the `config` host service

#### Scenario: Dynamic plugin omits configuration read methods

- **WHEN** a dynamic plugin declares `service: config` in `plugin.yaml` without `methods`
- **THEN** the system accepts the host service declaration and defaults it to the current complete read-only configuration method set
- **AND** the config host service methods in the authorization snapshot are normalized to `get`, `exists`, `string`, `bool`, `int`, and `duration`

#### Scenario: Dynamic plugin declares an unsupported configuration method

- **WHEN** a dynamic plugin declares a method other than `get`, `exists`, `string`, `bool`, `int`, or `duration` in the `config` host service in `plugin.yaml`
- **THEN** the system rejects the host service declaration
- **AND** it does not derive configuration capability for write, save, hot reload, or runtime configuration mutation methods

#### Scenario: Dynamic plugin uses guest SDK convenience read methods

- **WHEN** dynamic plugin code calls guest SDK configuration helpers such as `Exists`, `String`, `Bool`, `Int`, or `Duration`
- **THEN** the guest SDK initiates a host_call through the corresponding `config.exists`, `config.string`, `config.bool`, `config.int`, or `config.duration` method
- **AND** the host performs existence checks or type parsing inside the read-only configuration service

#### Scenario: Dynamic plugin reads an arbitrary configuration key

- **WHEN** an authorized dynamic plugin reads any existing configuration key through `config.get`
- **THEN** the system returns the JSON representation of the configuration value for that key
- **AND** it does not reject the read by plugin ID, prefix, or key pattern

#### Scenario: Dynamic plugin reads complete configuration snapshot

- **WHEN** an authorized dynamic plugin passes an empty key or `.` to `config.get`
- **THEN** the system returns the complete static configuration snapshot JSON exposed by the GoFrame configuration system

#### Scenario: Dynamic plugin reads a missing configuration key

- **WHEN** an authorized dynamic plugin reads a missing configuration key through any read-only config method
- **THEN** the system returns a `found=false` result
- **AND** it does not treat the missing key as a host_call failure

#### Scenario: Dynamic plugin configuration service remains read-only

- **WHEN** a dynamic plugin declares or calls the `config` host service
- **THEN** the host service permission declaration supports only the read-only methods `get`, `exists`, `string`, `bool`, `int`, and `duration`
- **AND** it does not support write, save, hot reload, or runtime configuration mutation methods
