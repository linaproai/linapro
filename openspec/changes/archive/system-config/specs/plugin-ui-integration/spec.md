## ADDED Requirements

### Requirement: Plugin Management Supports a Detail Dialog
The system SHALL provide a detail dialog for plugin records in plugin management so administrators can inspect plugin governance information without leaving the list page.

#### Scenario: Administrator opens plugin details from the list
- **WHEN** an administrator clicks the detail action for a plugin record
- **THEN** the system opens a detail dialog
- **AND** the dialog shows the plugin's identity, type, version, description, installation state, status, authorization requirements, authorization status, install time, and update time

### Requirement: Dynamic Plugin Details Merge Requested and Effective Host-Service Scope
The system SHALL present dynamic-plugin host-service information in a way that distinguishes declared intent from effective authorization only when the two differ.

#### Scenario: Requested and authorized host-service scopes are identical
- **WHEN** a dynamic plugin has the same requested and authorized host-service scope
- **THEN** the dialog merges them into one effective scope presentation
- **AND** the UI avoids duplicate explanatory text

#### Scenario: Requested and authorized host-service scopes differ
- **WHEN** a dynamic plugin's authorized host-service scope differs from its requested scope
- **THEN** the dialog shows requested scope and authorized scope as distinct sections
- **AND** resource groups use one semantic heading plus a resource list instead of repeating the same prefix for every resource item

### Requirement: Empty Host-Service State Is Only Shown for Dynamic Plugins
The system SHALL show an empty host-service state only when a dynamic plugin has no host-service governance data.

#### Scenario: Source plugin without host-service governance data
- **WHEN** a source plugin has no host-service governance information
- **THEN** the detail dialog does not show an unnecessary host-service empty-state block

### Requirement: Dynamic Plugin Demo Records Support Pagination
The system SHALL let administrators page through demo records on the dynamic plugin sample page.

#### Scenario: Administrator switches demo-record pages
- **WHEN** an administrator changes the page in the dynamic plugin demo record list
- **THEN** the page reloads the target page and updates the range summary and current page state accordingly
