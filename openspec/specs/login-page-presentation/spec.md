# login-page-presentation Specification

## Purpose
TBD - created by archiving change login-page-simplification-and-positioning. Update Purpose after archive.
## Requirements
### Requirement: Only the username/password login entry is exposed in the current stage

The system MUST expose only the username/password login capability in the current stage and MUST NOT continue to show or preserve unfinished authentication entry points as formal public capabilities.

#### Scenario: The standard login page shows only the username/password form
- **WHEN** an unauthenticated user visits `/auth/login`
- **THEN** the page shows username, password, remember-me, and login controls
- **AND** the page does not show forgot password, registration, mobile login, QR-code login, or third-party login entry points

#### Scenario: Users visit unfinished authentication sub-routes
- **WHEN** a user visits `/auth/code-login`, `/auth/qrcode-login`, `/auth/forget-password`, or `/auth/register`
- **THEN** the system redirects back to the standard login page at `/auth/login`
- **AND** the page still exposes only the username/password login capability

### Requirement: The login panel defaults to the right-aligned layout and supports position configuration

The system MUST render the login panel in the right-aligned layout by default and MUST allow the host public-frontend config to switch it to the left, center, or right layout.

#### Scenario: The login panel defaults to the right side when no override exists
- **WHEN** a browser loads the login page and the host does not provide a login-panel position override
- **THEN** the login page uses the `panel-right` layout
- **AND** the login panel is shown on the right side of the main page area

#### Scenario: Host config overrides the login-panel position
- **WHEN** the host public-frontend config returns `auth.panelLayout` as `panel-left`, `panel-center`, or `panel-right`
- **THEN** the login page renders the corresponding layout mode
- **AND** the layout switcher in the login-page toolbar still allows switching among all three layout options

### Requirement: The default login-page description supports host configuration

The system MUST display the default login-page description when the host does not provide an override, and MUST display the configured value when the host public-frontend config provides one.

#### Scenario: The default description is shown when no override exists
- **WHEN** a browser loads the login page and the host does not provide an `auth.pageDesc` override
- **THEN** the login page shows the description `Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model`

#### Scenario: Host config overrides the login-page description
- **WHEN** the host public-frontend config returns a non-empty `auth.pageDesc`
- **THEN** the login page shows the returned description

### Requirement: Login page must support host i18n copy and language-switch refresh
The system SHALL render login-page title, description, and subtitle according to the active language, combining frontend static language resources with localized public frontend settings returned by the host. When the active language changes, the login page MUST refresh the displayed copy without requiring a new login session.

#### Scenario: Login page displays host copy in English
- **WHEN** the browser language is `en-US` and the host provides public frontend config copy for that language
- **THEN** the login page displays English title, description, and login subtitle
- **AND** static form field copy continues to be rendered from frontend static locale bundles

#### Scenario: Login-page copy refreshes after language switch
- **WHEN** a user switches the workspace language before or after login
- **THEN** host copy in the login page or authentication layout refreshes to the new language result
- **AND** the login-page component structure does not need to be reconfigured

### Requirement: Login-page i18n misses must fall back to default copy
The system SHALL fall back to the default language copy or built-in static copy when the host does not provide localized login-page text for the current language.

#### Scenario: Current language lacks login-page description translation
- **WHEN** the current language has no available localized result for `auth.pageDesc`
- **THEN** the login page falls back to the default-language description or built-in default description copy
- **AND** login-page layout and authentication flow remain usable

