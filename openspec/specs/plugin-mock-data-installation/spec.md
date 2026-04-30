# Plugin Mock Data Installation

## Purpose

Define optional plugin mock-data loading during manual installation and startup bootstrap, including resource discovery, transaction boundaries, rollback behavior, migration ledger records, list metadata, and uninstall cleanup warnings.

## Requirements

### Requirement: Plugin install requests must support optional mock-data loading

The manual plugin install request (`POST /plugins/{id}/install`) SHALL expose `installMockData bool`, defaulting to `false`. When explicitly true, the host SHALL execute SQL files from the plugin `manifest/sql/mock-data/` directory after install SQL succeeds. When false or omitted, the host MUST skip scanning and execution of that directory. The frontend install modal SHALL show an unchecked mock-data checkbox only when plugin metadata reports `hasMockData=true`; the option itself is sufficient user consent and MUST NOT require an additional guard, permission, or confirmation.

#### Scenario: User opts in and mock data installs successfully
- **WHEN** the user checks the mock-data checkbox in the install modal and submits
- **AND** the plugin has valid SQL files under `manifest/sql/mock-data/`
- **AND** install SQL succeeds
- **THEN** the host executes the mock-data SQL files in file-name order
- **AND** the plugin is marked installed after all mock SQL succeeds

#### Scenario: User does not opt in
- **WHEN** the user submits installation without checking the mock-data checkbox
- **AND** the plugin has SQL files under `manifest/sql/mock-data/`
- **THEN** the host does not scan or execute any file in that directory
- **AND** no mock data rows are created

#### Scenario: Plugin has no mock-data directory
- **WHEN** the frontend opens the install modal for a plugin with `hasMockData=false`
- **THEN** the mock-data checkbox is not rendered
- **AND** a manually injected `installMockData=true` request does not cause any SQL execution if no mock-data directory exists

### Requirement: Mock SQL and ledger writes must share one database transaction

When `installMockData=true`, the host SHALL execute all mock SQL files and corresponding `sys_plugin_migration` ledger writes inside one database transaction. If any mock SQL file fails, the transaction MUST roll back both data rows and ledger rows. Install SQL, menu synchronization, and plugin registration MUST NOT share this mock transaction boundary.

#### Scenario: Any mock SQL failure rolls back all mock effects
- **WHEN** install SQL succeeds and the third mock SQL file fails after two earlier files succeeded
- **THEN** the host rolls back the mock transaction
- **AND** data rows from earlier mock SQL files are absent
- **AND** `sys_plugin_migration` contains no rows for that mock phase
- **AND** the plugin remains installed without mock data

#### Scenario: Successful mock phase commits
- **WHEN** all mock SQL files execute successfully
- **THEN** the transaction commits the mock data rows
- **AND** each mock SQL file has a successful `direction='mock'` ledger row

### Requirement: Mock phase failures must return actionable failure information

When the mock phase fails and rolls back, the host SHALL return a `bizerr` response with stable error code `plugin.install.mockDataFailed`, i18n key `plugins.install.error.mockDataFailed`, and message parameters containing `pluginId`, `failedFile`, `rolledBackFiles`, and `cause`. The frontend SHALL recognize this error and show localized guidance explaining that plugin installation succeeded, mock data was rolled back, and the user may accept the current state or uninstall and reinstall after fixing SQL.

#### Scenario: Frontend shows localized mock failure guidance
- **WHEN** the backend returns `plugin.install.mockDataFailed`
- **THEN** the frontend displays plugin ID, failed SQL file, and failure cause in the current language
- **AND** the message explains that mock data was automatically rolled back

#### Scenario: Error response still reflects install success
- **WHEN** install SQL succeeded and only mock SQL failed
- **THEN** the HTTP response is an error response
- **AND** plugin registry and menu synchronization remain installed effects
- **AND** a later plugin list request shows the plugin as installed

### Requirement: Migration ledger must distinguish install, uninstall, and mock phases

The host SHALL support `MigrationDirectionMock = "mock"` alongside install and uninstall directions. `sys_plugin_migration.direction` MUST accept `install`, `uninstall`, and `mock`. Each successful mock SQL file SHALL write one `direction='mock'` ledger row with the same metadata style as install rows, and rollback MUST remove those rows together with mock data.

#### Scenario: Operators can query ledger counts by phase
- **WHEN** an operator groups `sys_plugin_migration` rows by `direction` for a plugin that installed mock data
- **THEN** the result includes separate `install` and `mock` counts
- **AND** the counts match the install SQL and mock SQL file counts

### Requirement: Source and dynamic plugins must share the same mock-data loading mechanism

Source plugins and dynamic plugins SHALL use the same `manifest/sql/mock-data/` directory convention, SQL file naming constraints, transactional execution entry point, failure response format, frontend checkbox behavior, config opt-in behavior, error code, and i18n text. Dynamic plugin artifacts MUST carry mock SQL assets so runtime loading can use the same scanner as source plugins.

#### Scenario: Dynamic plugin supports mock-data loading
- **WHEN** a dynamic plugin artifact contains `manifest/sql/mock-data/001-*.sql`
- **AND** the user installs it with mock data selected
- **THEN** the host loads mock data through the same transactional flow as source plugins
- **AND** failures return the same error code format

#### Scenario: Source and dynamic plugin UX is consistent
- **WHEN** a user opens install modals for one source plugin and one dynamic plugin that both have mock data
- **THEN** checkbox text, tooltip, default value, and submit behavior are identical

### Requirement: Plugin list must display mock-data availability in a dedicated column

The plugin management list SHALL display a dedicated mock-data availability column between status and install time. The column MUST use API metadata `hasMockData` as its only source. Plugins with mock data display a positive tag, plugins without mock data display a neutral negative tag, and the plugin name column MUST NOT duplicate that marker. The list title SHALL include a help icon explaining source versus dynamic plugins and the mock-data column meaning.

#### Scenario: Plugin with mock data displays positive availability
- **WHEN** a plugin list row has `hasMockData=true`
- **THEN** the mock-data column displays a positive availability tag
- **AND** the plugin name does not include an additional mock-data marker

#### Scenario: Plugin without mock data displays negative availability
- **WHEN** a plugin list row has `hasMockData=false`
- **THEN** the mock-data column displays a neutral negative tag

#### Scenario: Plugin list title provides help
- **WHEN** the user hovers the help icon next to the plugin list title
- **THEN** the frontend explains source and dynamic plugin differences
- **AND** it explains the mock-data column meaning

### Requirement: Uninstall confirmation must emphasize plugin-owned data cleanup risk and hard-delete semantics

The plugin uninstall modal SHALL place the plugin-owned storage cleanup option in a visually risky warning area. When the user confirms uninstall with cleanup enabled, the host SHALL execute uninstall SQL, authorization storage cleanup, menu cleanup, and resource-reference cleanup. Plugin-owned data and resource references MUST be hard-deleted or physically removed rather than left as soft-deleted stale rows.

#### Scenario: User confirms uninstall with plugin-owned data cleanup
- **WHEN** the user keeps the plugin-owned storage cleanup option checked and confirms uninstall
- **THEN** the frontend highlights the cleanup risk before confirmation
- **AND** the backend physically removes plugin-owned data or authorization storage files
- **AND** governance resource references do not remain as soft-deleted stale rows

