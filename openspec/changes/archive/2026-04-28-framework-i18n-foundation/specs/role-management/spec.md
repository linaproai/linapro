## ADDED Requirements

### Requirement: Built-in protected role display names must be localized by the backend
The system SHALL return read-only display names for framework built-in protected roles in role management APIs according to the current request language, while continuing to keep database original values for user-editable role fields. Built-in role localization MUST derive translation keys from stable `role.key` anchors. The frontend MUST NOT maintain role-name translation mappings based on `role.key` or Chinese role names.

#### Scenario: Built-in super administrator role is queried in English
- **WHEN** an administrator requests the role list with `en-US` and views the framework built-in protected role with `key=admin`
- **THEN** the backend returns the role list display name as an English projected value
- **AND** other user-editable role records in the same list continue to return database values
- **AND** the frontend role list renders API response values directly and no longer calls frontend role seed mapping helpers

#### Scenario: Role detail and edit backfill keep governance values
- **WHEN** an administrator opens role detail, edit drawer, user authorization selector, or any page that affects governance data backfill
- **THEN** the backend continues to return database values as editable fields or selector semantic values
- **AND** language switching MUST NOT write localized display names back into role governance names
