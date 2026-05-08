## ADDED Requirements

### Requirement: Login page must support host i18n copy and language-switch refresh
The system SHALL render login-page title, description, and subtitle according to the active language, combining frontend static language resources with localized public frontend settings returned by the host. When the active language changes, the login page MUST refresh the displayed copy without requiring a new login session.

#### Scenario: Login page displays host copy in English
- **WHEN** the browser language is `en-US` and the host provides public frontend config copy for that language
- **THEN** the login page displays English title, description, and login subtitle
- **AND** static form field copy continues to be rendered from frontend static locale bundles

#### Scenario: Login-page copy refreshes after language switch
- **WHEN** a user switches the workspace language before or after login
- **THEN** host copy in the login page or authentication layout refreshes to the new language result

### Requirement: Login-page i18n misses must fall back to default copy
The system SHALL fall back to the default language copy or built-in static copy when the host does not provide localized login-page text for the current language.

#### Scenario: Current language lacks login-page description translation
- **WHEN** the current language has no available localized result for `auth.pageDesc`
- **THEN** the login page falls back to the default-language description or built-in default description copy
- **AND** login-page layout and authentication flow remain usable
