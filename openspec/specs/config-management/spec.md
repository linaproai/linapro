# Config Management

## Purpose

Define configuration management behavior, including localized import/export metadata, built-in parameter display, and deletion protection for system-owned records.

## Requirements

### Requirement: Config export and import headers must be resolved via translation keys by current language

The system SHALL resolve column headers (`name`, `key`, `value`, `remark`, `createdAt`, `updatedAt`) in config Excel export and import flows through `config.field.<name>` translation keys for the current request language. Backend Go source MUST NOT maintain literal English/Chinese header maps. Adding a new language only requires adding corresponding `config.field.*` resources under `apps/lina-core/manifest/i18n/<locale>/*.json`.

#### Scenario: Export uses current-language headers
- **WHEN** an administrator exports configs in a non-default runtime language
- **THEN** the Excel column headers use that language's `config.field.*` translations
- **AND** backend source does not contain duplicate literal header maps

#### Scenario: Adding a language requires no backend code change
- **WHEN** the project enables a new built-in language and provides `config.field.*` resources
- **THEN** config import and export headers display in that language
- **AND** no source code change in config services is required

### Requirement: Built-in system parameter names and default copy must be localized in English

The config management page SHALL localize built-in system parameter names, descriptions, and default display values by current language so English environments do not show default Chinese system copy.

#### Scenario: Login and IP blacklist parameters display English metadata
- **WHEN** an administrator opens system config in `en-US`
- **THEN** built-in login, page-title, page-description, subtitle, and IP blacklist parameter metadata display in English
- **AND** the page does not show Chinese built-in labels for those parameters

#### Scenario: Built-in public frontend copy can project English display content
- **WHEN** the config list displays default login-page title, description, or subtitle in `en-US`
- **THEN** the visible display content uses an English projection or English default
- **AND** edit details still preserve stable `configKey` and the actual stored value

#### Scenario: Config localization resources stay complete
- **WHEN** built-in config translation keys are added or changed
- **THEN** `zh-CN`, `en-US`, and `zh-TW` runtime resources keep matching key coverage
- **AND** missing-translation checks report no newly missing built-in config keys

### Requirement: Built-in system parameters must be editable but not deletable

System-owned config records SHALL be marked as built-in. Administrators may edit their editable fields and values, but deletion of built-in records MUST be blocked in both frontend and backend.

#### Scenario: Built-in system parameter delete action is disabled
- **WHEN** an administrator views built-in config rows
- **THEN** delete actions are disabled and do not open delete confirmation
- **AND** hover text explains that built-in system data cannot be deleted
- **AND** edit actions remain available

#### Scenario: Backend rejects built-in system parameter deletion
- **WHEN** a caller bypasses the frontend and requests deletion of a built-in config record
- **THEN** the backend returns a structured business error and preserves the record
- **AND** non-built-in config records remain deletable under existing permission and validation rules

