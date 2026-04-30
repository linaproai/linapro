## ADDED Requirements

### Requirement: Dictionary form layout must remain readable in English

Dictionary create and edit forms SHALL provide enough label width in English so labels such as `Dictionary Type` and `Tag Style` do not wrap awkwardly.

#### Scenario: English dictionary labels stay readable
- **WHEN** an administrator opens dictionary forms in `en-US`
- **THEN** long labels remain readable and aligned

### Requirement: Tag Style dropdown must show readable localized options

The dictionary data form Tag Style dropdown SHALL display human-readable option text in the current language and MUST NOT expose runtime i18n keys.

#### Scenario: English tag style dropdown labels
- **WHEN** an administrator opens the Tag Style dropdown in `en-US`
- **THEN** options display labels such as `Default`, `Primary`, and `Success`

### Requirement: Built-in dictionary types and data must be editable but not deletable

System-owned dictionary types and dictionary data SHALL be marked built-in. They remain editable where allowed, but deletion MUST be blocked in both frontend and backend.

#### Scenario: Backend rejects built-in dictionary deletion
- **WHEN** a caller requests deletion of built-in dictionary type or data
- **THEN** the backend returns a structured business error and preserves the record
