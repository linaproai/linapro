## ADDED Requirements

### Requirement: Dynamic-plugin authorization review shows a route exposure list

The system SHALL display the current release's dynamic-route list in the installation or enablement review dialog for a dynamic plugin. This list MUST appear as a governance section parallel to, but semantically separate from, `hostServices` authorization. The host-service authorization section MUST appear before the route-information section.

#### Scenario: The installation review dialog shows declared routes
- **WHEN** the host prepares to install a dynamic plugin whose current release declares dynamic-route contracts and the administrator opens the review dialog
- **THEN** the dialog shows the route list for that current release
- **AND** each route shows at least the HTTP method, real host public path, access level, permission key, and summary
- **AND** the routes appear in an independent review section rather than inside host-service cards or resource lists
- **AND** the host-service authorization section appears before the route-information section

#### Scenario: The enablement review dialog shows the routes that are about to take effect
- **WHEN** a dynamic plugin release is about to be enabled and the enablement review dialog is opened
- **AND** that release declares dynamic-route contracts
- **THEN** the host shows the routes declared by that pending release rather than routes from a historical release
- **AND** administrators can review the exact exposure that will become effective after enablement

#### Scenario: Long route lists are collapsed by default
- **WHEN** the installation or enablement review dialog opens and the current release declares more than two dynamic routes
- **THEN** the route-information section shows only the first two routes by default
- **AND** the host provides an explicit expand action so administrators can view the remaining routes
- **AND** after expansion, the dialog still shows the full public path, access level, and permission key for every route

#### Scenario: Plugins without declared routes do not render a redundant block
- **WHEN** the installation or enablement review dialog opens for a dynamic plugin whose current release declares no dynamic routes
- **THEN** the host does not render a redundant route-list block
- **AND** the existing plugin information and host-service authorization sections continue to behave normally

### Requirement: The dynamic-plugin detail dialog reuses the route exposure review structure

The system SHALL show the current release's dynamic-route list in the dynamic-plugin detail dialog and reuse the same route-information structure used by the authorization review dialog.

#### Scenario: The detail dialog shows route information
- **WHEN** an administrator opens the detail dialog for a dynamic plugin whose current release declares dynamic-route contracts
- **THEN** the dialog shows an independent "Registered Routes" section
- **AND** the section appears after the host-service information section
- **AND** each route shows at least the HTTP method, real host public path, access level, permission key, and summary

#### Scenario: Long route lists stay collapsed in the detail dialog
- **WHEN** the dynamic-plugin detail dialog contains more than two routes
- **THEN** the dialog shows only the first two routes by default
- **AND** administrators can expand the list to view the full set

### Requirement: Governance section titles must be visually distinct in the authorization and detail dialogs

The system SHALL style governance section titles such as "Host Service Authorization Scope," "Host Service Information," and "Registered Routes" so they are clearly distinguishable from body text and help administrators identify section boundaries quickly.

#### Scenario: Authorization-dialog section titles have a clear hierarchy
- **WHEN** an administrator opens the installation or enablement review dialog for a dynamic plugin
- **THEN** titles such as "Host Service Authorization Scope" and "Registered Routes" are visibly distinguished through size, weight, accent color, or an equivalent hierarchy treatment
- **AND** administrators can quickly identify the boundaries between governance sections

#### Scenario: Detail-dialog section titles follow the same visual language
- **WHEN** an administrator opens the dynamic-plugin detail dialog
- **THEN** titles such as "Host Service Information" and "Registered Routes" use the same clear hierarchy pattern
- **AND** the visual treatment remains consistent with the authorization dialog

### Requirement: Long plugin descriptions must not break the detail layout

The system SHALL render long plugin descriptions in the detail dialog without compressing the rest of the base information table into an unreadable layout.

#### Scenario: Long descriptions use a dedicated full-width row
- **WHEN** the plugin description is long in the detail dialog
- **THEN** the host renders the description as a dedicated full-width row inside the base information table
- **AND** administrators can still read the full description without losing the readability of other metadata fields

### Requirement: The detail dialog hides the redundant authorization-requirement field

The system SHALL hide the "Authorization Requirement" field in the plugin detail dialog so it does not duplicate the meaning already conveyed by authorization state and other governance fields.

#### Scenario: The detail dialog no longer shows the authorization-requirement field
- **WHEN** an administrator opens the plugin detail dialog
- **THEN** the base information table does not render a standalone "Authorization Requirement" field
- **AND** more valuable fields such as authorization state and plugin description remain visible
