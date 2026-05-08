## ADDED Requirements

### Requirement: User list role names must match backend-localized role display

The user management list SHALL use role display names returned by the backend and keep built-in role display consistent with role management in the same language.

#### Scenario: User list shows administrator role in English
- **WHEN** an administrator opens user management in `en-US`
- **THEN** the administrator role display matches role management

### Requirement: Built-in protected role display names must be localized by the backend
The system SHALL return read-only display names for framework built-in protected roles in role management APIs according to the current request language, while continuing to keep database original values for user-editable role fields. Built-in role localization MUST derive translation keys from stable `role.key` anchors. The frontend MUST NOT maintain role-name translation mappings based on `role.key` or Chinese role names.

#### Scenario: Built-in super administrator role is queried in English
- **WHEN** an administrator requests the role list with `en-US` and views the framework built-in protected role with `key=admin`
- **THEN** the backend returns the role list display name as an English projected value
- **AND** other user-editable role records in the same list continue to return database values

#### Scenario: Role detail and edit backfill keep governance values
- **WHEN** an administrator opens role detail, edit drawer, or user authorization selector
- **THEN** the backend continues to return database values as editable fields
- **AND** language switching MUST NOT write localized display names back into role governance names

### Requirement: Built-in role display must be localized consistently

Built-in protected roles and default seed roles SHALL display according to current language in framework-delivered pages. User management and role management MUST show the same localized display value for the same built-in role.

#### Scenario: Default user role displays in English
- **WHEN** an administrator opens role management in `en-US`
- **THEN** the default user role displays an English name

### Requirement: System-generated Unassigned node must be localized

System-generated virtual Unassigned department nodes SHALL be localized by current request language and maintained through backend or plugin runtime i18n resources.

#### Scenario: Unassigned displays in English
- **WHEN** an administrator opens a page with a department-tree filter in `en-US`
- **THEN** the virtual node displays as `Unassigned`

### Requirement: Position form status selector must remain readable in English

Position create and edit forms SHALL keep the status field label and options readable in English, avoiding awkward wrapping caused by insufficient space.

#### Scenario: English position status options stay readable
- **WHEN** an administrator opens a position form in `en-US`
- **THEN** the status label and options remain readable and operable

### Requirement: Service monitoring disk table must remain readable in English

The service monitoring page SHALL allocate readable disk table column widths in English so key column headers and values do not wrap unnecessarily.

#### Scenario: English disk table keeps key columns on one line
- **WHEN** an administrator opens service monitoring in `en-US`
- **THEN** `File System`, `Total`, `Used`, and `Available` headers do not wrap
