## Requirements

### Requirement: Plugin Manifest Must Support Dependency Declaration

The system SHALL allow source plugins and dynamic plugins to declare LinaPro framework version constraints and plugin dependency constraints through `dependencies` in `plugin.yaml` or equivalent runtime manifest. Plugins without `dependencies` declared SHALL remain valid and be treated as dependency-free.

#### Scenario: Parse framework version constraint and plugin dependencies
- **WHEN** a plugin manifest contains `dependencies.framework.version` and `dependencies.plugins`
- **THEN** the system parses the framework version constraint
- **AND** the system parses each plugin dependency's `id`, `version`, `required`, and `install`
- **AND** missing `required` defaults to `true`
- **AND** missing `install` defaults to `manual`

#### Scenario: Undeclared dependency plugin remains compatible
- **WHEN** a plugin manifest does not contain `dependencies`
- **THEN** the system treats the plugin as dependency-free
- **AND** manifest validation must not fail due to missing `dependencies`

#### Scenario: Dynamic plugin artifact carries dependency declaration
- **WHEN** a dynamic plugin WASM artifact's manifest custom section contains `dependencies`
- **THEN** the system parses the dependency declaration with the same semantics as source plugins
- **AND** dynamic plugin install, enable, and upgrade paths use the parsed dependency constraints

### Requirement: Plugin Dependency Declaration Must Be Structurally Validated

The system SHALL validate dependency declaration structure during manifest validation. Framework version constraints and plugin version constraints must use supported semver ranges; plugin dependency IDs must conform to plugin ID naming rules; a plugin must not depend on itself; the same manifest must not declare duplicate plugin dependencies.

#### Scenario: Reject invalid dependency fields
- **WHEN** a plugin manifest declares empty dependency ID, invalid version range, unknown install strategy, or duplicate dependency
- **THEN** manifest validation fails
- **AND** the error contains the plugin ID, dependency field path, and invalid value

#### Scenario: Reject self-dependency
- **WHEN** plugin `content-notice` declares dependency on `content-notice` in `dependencies.plugins`
- **THEN** manifest validation fails
- **AND** the error states the plugin must not depend on itself

### Requirement: Dependency Check Must Execute Before Install

The system SHALL complete dependency checking before executing any plugin install lifecycle side effects. The dependency check must validate the current LinaPro framework version, whether dependency plugins are discoverable, whether dependency versions are satisfied, whether hard dependencies are installed or auto-installable, and whether the dependency graph contains cycles.

#### Scenario: Framework version not satisfied blocks install
- **WHEN** a plugin declares `dependencies.framework.version: ">=0.7.0"`
- **AND** the current LinaPro framework version is `v0.6.0`
- **THEN** the plugin install request fails
- **AND** the system must not execute the plugin's SQL, menu sync, release switch, or state write
- **AND** the error contains the current framework version and required version range

#### Scenario: Missing hard dependency blocks install
- **WHEN** a plugin declares hard dependency on `multi-tenant`
- **AND** the plugin catalog has not discovered `multi-tenant`
- **THEN** the plugin install request fails
- **AND** the error contains the missing dependency plugin ID and target plugin ID

#### Scenario: Dependency version not satisfied blocks install
- **WHEN** a plugin declares dependency on `org-center` with version range `>=0.2.0`
- **AND** the currently available or installed `org-center` version is `v0.1.0`
- **THEN** the plugin install request fails
- **AND** the error contains the current dependency version and required version range

#### Scenario: Circular dependency blocks install
- **WHEN** the dependency graph contains `a -> b -> c -> a`
- **THEN** any install request for a plugin participating in the cycle fails
- **AND** the error contains the circular dependency chain

### Requirement: Auto Dependency Installation Must Execute in Deterministic Topological Order

When a plugin hard dependency declares `install: auto` and the dependency plugin has been discovered with a satisfied version but is not yet installed, the system SHALL automatically install the dependency before the target plugin. The auto-install plan must follow dependency topological order; dependency plugins are installed before the plugins that depend on them. Plugins at the same topological level are sorted stably by plugin ID.

#### Scenario: Auto-install uninstalled hard dependency
- **WHEN** an administrator installs plugin `content-notice`
- **AND** `content-notice` declares `multi-tenant` as `required: true` and `install: auto`
- **AND** `multi-tenant` is discovered, version satisfied, and not yet installed
- **THEN** the system first installs `multi-tenant`
- **AND** the system then installs `content-notice`
- **AND** the install result records the auto-installed dependency plugin list

#### Scenario: Auto-install does not enable dependency plugin
- **WHEN** the system auto-installs dependency plugin `multi-tenant`
- **THEN** the system only ensures the dependency plugin reaches installed state
- **AND** the system must not enable the dependency plugin by default because it was auto-installed

#### Scenario: Auto-install mid-failure stops subsequent installations
- **WHEN** the auto-install plan includes `a`, `b`, `target`
- **AND** plugin `b` installation fails
- **THEN** the system stops installing `target`
- **AND** the error contains the installed-so-far dependency list, failed plugin ID, and failure reason

### Requirement: Manual Dependency Must Block Target Plugin Install with Action Prompt

When a hard dependency declares `install: manual` and the dependency plugin is not installed, the system SHALL block the target plugin install and return the dependency list requiring manual installation first. The system must not auto-install `manual` dependencies without confirmation.

#### Scenario: Uninstalled manual hard dependency
- **WHEN** plugin `x` declares hard dependency on `org-center` with `install: manual`
- **AND** `org-center` is not yet installed
- **THEN** installing `x` fails
- **AND** the error prompts to install `org-center` first

### Requirement: Soft Dependency Must Not Block Plugin Lifecycle

When a dependency declares `required: false`, the system SHALL treat it as a soft dependency. Soft dependency absence, non-installation, or version mismatch must not block install, enable, or upgrade, but the dependency check result must mark the soft dependency as unsatisfied for API and UI display.

#### Scenario: Missing soft dependency continues install
- **WHEN** plugin `x` declares soft dependency on `optional-analytics`
- **AND** `optional-analytics` has not been discovered
- **THEN** plugin `x` installation continues
- **AND** the dependency check result marks `optional-analytics` as an unsatisfied soft dependency

### Requirement: Uninstall Must Protect Hard Dependencies of Installed Plugins

The system SHALL check installed plugins' hard dependency declarations before uninstalling a plugin. If other installed plugins hard-depend on the target plugin, the uninstall request must fail and return the downstream plugin list that depends on the target.

#### Scenario: Refuse uninstall when depended on by installed plugins
- **WHEN** plugin `content-notice` is installed and hard-depends on `multi-tenant`
- **AND** an administrator requests uninstalling `multi-tenant`
- **THEN** the uninstall request fails
- **AND** the system must not execute `multi-tenant`'s uninstall SQL, menu cleanup, or state write
- **AND** the error contains the downstream plugin `content-notice`

#### Scenario: Allow uninstall when no downstream hard dependencies
- **WHEN** no installed plugin hard-depends on the target plugin
- **THEN** the system allows continuing the existing uninstall lifecycle

### Requirement: Dependency Check Results Must Be Visible via API and UI

The system SHALL provide dependency check results for plugin management, including framework version check, dependency plugin status, version matching result, auto-install plan, manual items, soft dependency hints, circular dependencies, and uninstall blockers. The frontend SHALL use server-side results for display and must not re-implement dependency graph semantics on the client side.

#### Scenario: Display auto-install plan before install
- **WHEN** an administrator clicks install on a plugin with auto dependencies in the plugin management page
- **THEN** the frontend displays the backend-returned auto-install dependency list
- **AND** the administrator confirms before the install request is submitted

#### Scenario: Display blocker reason
- **WHEN** the backend dependency check returns framework version not satisfied or dependency version not satisfied
- **THEN** the plugin management page displays the corresponding blocker reason
- **AND** the text uses i18n resources, not hardcoded text

#### Scenario: Uninstall confirmation displays downstream dependencies
- **WHEN** an administrator attempts to uninstall a plugin that other plugins hard-depend on
- **THEN** the plugin management page displays the downstream plugin list
- **AND** the uninstall operation is prevented

### Requirement: Dependency Lifecycle Changes Must Maintain Cache Consistency

The system SHALL publish or refresh plugin runtime revision/event, enabled snapshot, frontend bundle, runtime i18n bundle, and apidoc i18n derived caches by affected plugin scope after dependency auto-installation, target plugin install, uninstall after blocker removal, source plugin upgrade, and dynamic plugin upgrade success. In cluster mode, the system must not only refresh current node memory state.

#### Scenario: Cluster mode auto-installs multiple plugins
- **WHEN** in cluster mode the primary node auto-installs dependency plugins and installs the target plugin
- **THEN** the primary node publishes plugin runtime revision or equivalent events for each affected plugin
- **AND** non-primary nodes observe the events and refresh local enabled snapshot and derived caches

#### Scenario: Read-only dependency check does not trigger cache invalidation
- **WHEN** an administrator only performs a pre-install dependency check
- **THEN** the system must not write to plugin registry, release snapshot, or cache revision number
- **AND** the system must not clear all languages and all sectors' i18n caches
