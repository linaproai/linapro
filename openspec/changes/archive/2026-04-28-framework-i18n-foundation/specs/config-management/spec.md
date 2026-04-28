## ADDED Requirements

### Requirement: Config metadata must return localized names and remarks for the current language
The system SHALL return localized config metadata for config list, import/export templates, and protected-setting projections. Config metadata localization MUST use stable config keys as translation anchors and MUST NOT change the actual config key or stored config value.

#### Scenario: Config list returns English metadata
- **WHEN** an administrator queries the config list with `en-US`
- **THEN** the config names and remarks in the response use English localized values
- **AND** `configKey` and `configValue` keep their original governance semantics

#### Scenario: Config management keeps original config values
- **WHEN** an administrator opens the parameter setting edit modal, exports data, or imports filled data with `en-US`, and a config item's `configValue` is Chinese seed copy
- **THEN** the system MUST NOT write the current-language projected value back as the config governance value
- **AND** `configKey` and `configValue` continue to participate in edit, import, export, and audit flows using the database's original values
- **AND** parameter setting detail backfill keeps the database value by default, avoiding editable master data changes caused only by language switching

#### Scenario: Import templates and export headers return localized metadata
- **WHEN** an administrator downloads a parameter setting import template or exports parameter setting data with `en-US`
- **THEN** template instructions, header titles, and metadata-related prompts use English localized copy
- **AND** the `configKey` and `configValue` columns keep their original governance semantics and actual exported content

#### Scenario: Missing config metadata translations fall back to the default language
- **WHEN** a config item lacks a name or remark translation in the current language
- **THEN** the system falls back to default-language metadata or baseline names
- **AND** config read and write capabilities remain unaffected

### Requirement: Public frontend config copy must support i18n projection
The system SHALL let the public frontend config endpoint return localized brand and authentication copy according to the current request language, while keeping non-textual fields such as layout and theme mode stable.

#### Scenario: Login-page public config returns English copy
- **WHEN** the browser requests the public frontend config endpoint with `en-US`
- **THEN** the returned app name, login page title, login page description, and login subtitle are English localized results
- **AND** non-copy fields such as `panelLayout`, `themeMode`, and `layout` keep their original values

#### Scenario: Copy-like public parameters are projected only at consumer endpoints
- **WHEN** `sys.app.name`, `sys.auth.pageTitle`, `sys.auth.pageDesc`, `sys.auth.loginSubtitle`, or `sys.ui.watermark.content` still store default seed values, and the browser requests the public frontend config endpoint with `en-US`
- **THEN** the public frontend config endpoint returns the corresponding English localized display copy
- **AND** the same `configValue` in the parameter setting management API keeps the original database value and does not add display translation fields

#### Scenario: Refreshed workspace shows the latest localized brand copy
- **WHEN** an administrator updates public frontend copy for a locale and refreshes the login page or workspace
- **THEN** the refreshed page shows the new localized brand name and login display copy
- **AND** no page component code changes are required
