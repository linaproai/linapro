## ADDED Requirements

### Requirement: System information page must display project introduction and component descriptions by current language
The system SHALL return project description, component descriptions, and other display copy on the system information page and system information API according to the current request language. System information i18n MUST keep project positioning and component identifiers stable, localizing only user-facing descriptive text.

#### Scenario: System information displays in English
- **WHEN** a user opens the system information page or requests the system information API with `en-US`
- **THEN** the project description in the About section uses an English localized result
- **AND** frontend and backend component descriptions use English localized results
- **AND** component names, version numbers, and links keep their original values

#### Scenario: Missing component descriptions fall back to the default language
- **WHEN** a component lacks description copy in the current language
- **THEN** the system falls back to the default-language description

### Requirement: System information i18n must cover public project positioning copy
The system SHALL keep project name, project introduction, and framework positioning descriptions semantically consistent across multilingual scenarios, ensuring that LinaPro is always described as an AI-driven full-stack development framework and does not drift into a single admin system or other product boundary in another language.

#### Scenario: Unified project positioning is preserved across languages
- **WHEN** a user switches the system language and views the system information page
- **THEN** the project positioning copy changes only in language expression
- **AND** LinaPro is not described as a single backend management system or any other product positioning that deviates from framework positioning
