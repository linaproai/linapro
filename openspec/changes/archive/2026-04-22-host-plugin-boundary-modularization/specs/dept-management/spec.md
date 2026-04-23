## MODIFIED Requirements

### Requirement: Department management is delivered by the organization source plugin

The The system SHALL deliver department management capabilities as `org-center` source plugins, rather than continuing as the host's default built-in module.

#### Scenario: Provides department management when the organization plugin is enabled
- **WHEN** `org-center` is installed and enabled
- **THEN** Host exposed department management API, pages and menus
- **AND** The department management menu is mounted to the host `Organization Management` directory, and the top-level `parent_key` is `org`

#### Scenario: Hide the department management entrance when the organization plugin is missing
- **WHEN** `org-center` is not installed or not enabled
- **THEN** The host does not display the department management menu and page entry
- **AND** Hosting capabilities such as user management will continue to be available according to organization downgrade rules.
