## MODIFIED Requirements

### Requirement: Notification announcement menu and permissions
The system SHALL mounts the notification announcement menu as the `content-notice` source plugin menu to the host `content management` directory instead of to `system management`.

#### Scenario: Menu display
- **WHEN** `content-notice` is installed, enabled and the current user has menu access
- **THEN** `Content Management` displays the `Notification Announcement` menu item under the group
- **AND** Plugin governance is still the responsibility of `Extension Center / Plugin Management`

#### Scenario: Plug-in missing or disabled
- **WHEN** `content-notice` is not installed, not enabled, or the current user does not have access to its menu
- **THEN** The host does not display the `Notification Announcement` menu entry
- **AND** If there are no other visible submenus in `Content Management`, the parent directory will be hidden as well.

## ADDED Requirements

### Requirement: Notification announcements are delivered by the content source plugin

The The system SHALL deliver the notification announcement capability as a `content-notice` source plugin, rather than continuing as the host's default built-in module.

#### Scenario: Provides notification announcement capability when the content plugin is enabled
- **WHEN** `content-notice` is installed and enabled
- **THEN** The host exposes notification announcement related APIs, pages and menus
- **AND** This plugin continues to host the announcement content management and publishing process

#### Scenario: Hide the notification announcement entrance when the content plugin is missing
- **WHEN** `content-notice` is not installed or enabled
- **THEN** The host does not display the notification announcement menu and page entry
- **AND** The remaining core capabilities of the host continue to operate normally
