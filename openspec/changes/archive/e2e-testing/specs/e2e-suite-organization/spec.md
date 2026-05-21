## ADDED Requirements

### Requirement: Plugin-full E2E must focus on plugin framework generic capability and plugin own cases
Plugin-full E2E SHALL validate host-level plugin framework behavior through generic dynamic-plugin fixtures and SHALL validate official source-plugin functionality only through plugin-owned tests under each plugin directory. It MUST NOT rely on indiscriminately re-running the complete host-only suite as its primary coverage strategy.

#### Scenario: Plugin-full covers plugin capability
- **WHEN** plugin-full E2E executes
- **THEN** it must cover official plugin own E2E tests
- **AND** root `hack/tests/e2e/extension/plugin/` can only cover host plugin framework, dynamic test plugin runtime, and generic plugin governance capability
- **AND** official source plugin menu, permission, route, i18n, task, mock data, or runtime resource verification must be closed-loop in `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`
- **AND** source plugin own E2E selection entries must remain `plugins` and `plugin:<plugin-id>`, not adding long-term scopes by official plugin business modules

#### Scenario: Host-only continues covering host full capability
- **WHEN** complete E2E verification chain executes
- **THEN** host-only E2E must continue covering host full capability scope
- **AND** plugin-full E2E must not replace host-only E2E's host baseline responsibility

#### Scenario: Root E2E does not couple official plugin IDs
- **WHEN** developer adds or modifies E2E cases, POM, support helpers, runner configuration, or execution manifest under root `hack/tests`
- **THEN** these files must not hardcode any specific source plugin ID, source plugin path, source plugin menu, source plugin mock data, source plugin test data, source plugin configuration baseline, or source plugin i18n key
- **AND** plugin-related test data, plugin-related configuration, and plugin-specific baseline must be placed in corresponding `apps/lina-plugins/<plugin-id>/hack/tests/` directory
- **AND** cases needing to verify specific source plugin behavior must be moved to corresponding `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`
- **AND** root E2E can only retain generic runner, generic discovery mechanism, and host plugin framework capability needed non-plugin-specific test assets
- **AND** root E2E governance check must block new specific source plugin information coupling from entering root test code and configuration

#### Scenario: Main framework host baseline does not default-depend on source plugins
- **WHEN** host-only E2E or host-only module E2E executes
- **THEN** it must not default require `apps/lina-plugins` to be initialized
- **AND** cases depending on official plugin installation, enablement, routes, menus, i18n, or source plugin mock data must be classified as plugin-full scope

### Requirement: E2E test file numbering must increment locally per module directory
E2E test file prefixes SHALL be scoped to the owning module directory instead of using a globally increasing sequence across the whole repository.

#### Scenario: Increment within host module
- **WHEN** developer adds test file in `hack/tests/e2e/<module>/` or its sub-module directory
- **THEN** file name must use `TC{NNN}-{brief-name}.ts`
- **AND** `TC` number must increment continuously from `TC001` within the current directory
- **AND** must not skip numbers because other host modules or source plugin directories already have larger `TC` numbers

#### Scenario: Increment within source plugin module
- **WHEN** developer adds plugin own test file in `apps/lina-plugins/<plugin-id>/hack/tests/e2e/<module>/`
- **THEN** file name must use `TC{NNN}-{brief-name}.ts`
- **AND** `TC` number must only increment continuously within that plugin's current module directory
- **AND** must not couple with host module or other plugin module test numbering

#### Scenario: Governance check blocks global numbering regression
- **WHEN** `pnpm -C hack/tests test:validate` executes
- **THEN** check must reject `TC{NNNN}-*.ts` old global four-digit numbered files
- **AND** check must reject duplicate, missing, or non-continuously-incrementing-from-TC001 test file numbers within the same module directory

### Requirement: Official plugin lifecycle regression must use representative full chain and batch smoke combination
Official plugin lifecycle regression SHALL avoid repeating the same full UI lifecycle for every official plugin when a representative full lifecycle plus per-plugin contract smoke provides equivalent behavioral coverage.

#### Scenario: Representative plugin executes full UI lifecycle
- **WHEN** official plugin lifecycle regression executes
- **THEN** at least one representative official plugin must cover installation, enablement, page accessibility, disablement, uninstallation, and menu mounting changes in complete UI chain

#### Scenario: Other official plugins execute contract smoke
- **WHEN** non-representative official plugins participate in lifecycle regression
- **THEN** tests must verify the plugin can be synchronized, installed, enabled, menu or route can be mounted, core pages are accessible
- **AND** tests must not repeat complete UI disablement, uninstallation, and missing page verification for every plugin, unless that plugin has unique lifecycle risks

#### Scenario: Unique lifecycle risks must retain dedicated coverage
- **WHEN** a specific official plugin has unique installation, enablement, uninstallation, data retention, task registration, permission, or runtime resource risks
- **THEN** that plugin must retain dedicated E2E or API-level regression coverage for that risk
