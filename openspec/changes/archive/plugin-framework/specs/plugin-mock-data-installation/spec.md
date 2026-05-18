## Requirements

### Requirement: Install Request Supports Optional Mock-Data Loading

The install request SHALL expose `installMockData` (default false). When true, the host executes mock SQL from `manifest/sql/mock-data/` after install SQL succeeds.

#### Scenario: Mock data installs successfully
- **WHEN** user checks mock-data and install SQL succeeds
- **THEN** host executes mock SQL in file-name order

#### Scenario: User does not opt in
- **WHEN** user submits without checking mock-data
- **THEN** host skips mock-data directory entirely

### Requirement: Mock SQL and Ledger Share One Transaction

All mock SQL files and `sys_plugin_migration` rows execute in one transaction. Any failure rolls back all mock effects while preserving installed state.

### Requirement: Mock Failures Return Actionable Information

Mock phase failure returns `bizerr` with `plugin.install.mockDataFailed` code, i18n key, and parameters including `pluginId`, `failedFile`, `rolledBackFiles`, `cause`.

### Requirement: Migration Ledger Distinguishes Mock Phase

`sys_plugin_migration.direction` supports `install`, `uninstall`, and `mock`.

### Requirement: Source and Dynamic Plugins Share Mock Mechanism

Both use the same directory convention, transactional execution, error format, and frontend behavior.

### Requirement: Plugin List Displays Mock-Data Availability

The management list shows a dedicated mock-data column based on `hasMockData` metadata.

### Requirement: Uninstall Cleanup Emphasizes Hard-Delete Semantics

The uninstall modal highlights plugin-owned data cleanup risk. Confirmed cleanup physically removes data and resource references.
