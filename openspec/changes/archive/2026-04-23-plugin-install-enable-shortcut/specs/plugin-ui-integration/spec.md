## ADDED Requirements

### Requirement: Plugin installation dialog supports an install-and-enable shortcut

The system SHALL provide both `Install Only` and `Install and Enable` governance actions in the plugin installation dialog so administrators can complete immediate enablement after reviewing plugin information without returning to the list for a second step.

#### Scenario: Show the shortcut action when the user has install and enable permissions
- **WHEN** an administrator opens the installation dialog for a plugin that is not installed and the current account has both `plugin:install` and `plugin:enable`
- **THEN** the dialog shows the `Install and Enable` shortcut action
- **AND** the dialog still keeps the `Install Only` action so the administrator can explicitly choose installation without enablement

#### Scenario: Hide the shortcut action when the user only has install permission
- **WHEN** an administrator opens the installation dialog for a plugin that is not installed but the current account has `plugin:install` without `plugin:enable`
- **THEN** the dialog shows only the `Install Only` action
- **AND** the UI MUST NOT show the `Install and Enable` button so it does not imply access to a governance action outside the user's permission scope

#### Scenario: Show the real state when the second step of the composite action fails
- **WHEN** an administrator chooses `Install and Enable` in the installation dialog and the install step succeeds but the enable step fails
- **THEN** the refreshed plugin management page MUST show the plugin as `installed but disabled`
- **AND** the UI clearly tells the administrator that installation completed but enablement did not succeed and can be retried later
