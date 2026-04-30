## ADDED Requirements

### Requirement: Built-in system parameter names and default copy must be localized in English

The config management page SHALL localize built-in system parameter names, descriptions, and default display values by current language so English environments do not show default Chinese system copy.

#### Scenario: Login and IP blacklist parameters display English metadata
- **WHEN** an administrator opens system config in `en-US`
- **THEN** built-in login, page-title, page-description, subtitle, and IP blacklist parameter metadata display in English

### Requirement: Built-in system parameters must be editable but not deletable

System-owned config records SHALL be marked as built-in. Administrators may edit their editable fields and values, but deletion of built-in records MUST be blocked in both frontend and backend.

#### Scenario: Backend rejects built-in system parameter deletion
- **WHEN** a caller bypasses the frontend and requests deletion of a built-in config record
- **THEN** the backend returns a structured business error and preserves the record
