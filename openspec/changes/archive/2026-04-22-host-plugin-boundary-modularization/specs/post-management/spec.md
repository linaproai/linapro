## ADDED Requirements

### Requirement: Position management is delivered by the organization source plugin

The The system SHALL deliver position management capabilities as an `org-center` source plugin, rather than continuing as the host's default built-in module.

#### Scenario: Provides position management when the organization plugin is enabled
- **WHEN** `org-center` is installed and enabled
- **THEN** The host exposes position management API, page and menu
- **AND** The position management menu is mounted to the host `Organization Management` directory, and the top-level `parent_key` is `org`

#### Scenario: Hide the position management entrance when the organization plugin is missing
- **WHEN** `org-center` is not installed or not enabled
- **THEN** The host does not display the position management menu and page entry
- **AND** Hosting capabilities such as user management will continue to be available according to organization downgrade rules.
