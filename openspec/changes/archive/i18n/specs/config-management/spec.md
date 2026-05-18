## ADDED Requirements

### Requirement: Config metadata must return localized names and remarks for the current language
The system SHALL return localized config metadata for config list, import/export templates, and protected-setting projections. Config metadata localization MUST use stable config keys as translation anchors and MUST NOT change the actual config key or stored config value.

#### Scenario: Config list returns English metadata
- **WHEN** an administrator queries the config list with `en-US`
- **THEN** the config names and remarks in the response use English localized values
- **AND** `configKey` and `configValue` keep their original governance semantics

#### Scenario: Config management keeps original config values
- **WHEN** an administrator opens the parameter setting edit modal with `en-US`, and a config item's `configValue` is Chinese seed copy
- **THEN** the system MUST NOT write the current-language projected value back as the config governance value
- **AND** parameter setting detail backfill keeps the database value by default

#### Scenario: Import templates and export headers return localized metadata
- **WHEN** an administrator downloads a parameter setting import template or exports data with `en-US`
- **THEN** template instructions, header titles, and metadata-related prompts use English localized copy
- **AND** the `configKey` and `configValue` columns keep their original governance semantics

### Requirement: Config export and import headers must be resolved via translation keys by current language
The system SHALL resolve column headers in the export and import pipelines by current request language through `config.field.<name>` translation keys, and MUST NOT maintain English/Chinese literal mapping tables in backend Go source.

#### Scenario: Traditional Chinese environment export contains Traditional Chinese headers
- **WHEN** an administrator requests `GET /config/export` in the `zh-TW` environment
- **THEN** the exported Excel column headers are displayed in Traditional Chinese
- **AND** no hardcoded literal mappings like `englishLabels` / `chineseLabels` exist in backend code

#### Scenario: Adding a new language requires no backend Go code changes
- **WHEN** the project enables a new built-in language and provides `manifest/i18n/<locale>/*.json` resources
- **THEN** config import and export headers automatically display in that language

### Requirement: Public frontend config copy must support i18n projection
The system SHALL let the public frontend config endpoint return localized brand and authentication copy according to the current request language, while keeping non-textual fields stable.

#### Scenario: Login-page public config returns English copy
- **WHEN** the browser requests the public frontend config endpoint with `en-US`
- **THEN** the returned app name, login page title, login page description, and login subtitle are English localized results
- **AND** non-copy fields such as `panelLayout`, `themeMode`, and `layout` keep their original values

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
