## ADDED Requirements

### Requirement: Built-in metadata for the login-panel position parameter

The system MUST provide a protected built-in public-frontend parameter named `sys.auth.loginPanelLayout` to maintain the default login-panel layout.

#### Scenario: Initialize the login-panel position parameter
- **WHEN** an administrator runs the host initialization SQL
- **THEN** `sys_config` contains a built-in parameter record with key `sys.auth.loginPanelLayout`
- **AND** the default value of that record is `panel-right`
- **AND** the record includes a readable name and value descriptions for `panel-left`, `panel-center`, and `panel-right`

### Requirement: Validate the login-panel position parameter and expose it through the public-frontend config endpoint

The system MUST validate the value domain of `sys.auth.loginPanelLayout` and expose the effective value through the public-frontend config endpoint for unauthenticated pages.

#### Scenario: Reject invalid login-panel position values
- **WHEN** a user creates, updates, or imports `sys.auth.loginPanelLayout` with a value other than `panel-left`, `panel-center`, or `panel-right`
- **THEN** the system rejects the change and returns a parameter-validation error

#### Scenario: Public frontend config returns the login-panel position
- **WHEN** a browser requests `GET /config/public/frontend`
- **THEN** `auth.panelLayout` in the response equals the effective value of `sys.auth.loginPanelLayout`
- **AND** unauthenticated pages can consume that value without reading any other `sys_config` data

### Requirement: Default value and length rules for the login-page description parameter

The system MUST provide a default description value for the protected built-in public-frontend parameter `sys.auth.pageDesc`, and MUST allow a non-empty description of up to 500 characters so the login page can show richer product copy.

#### Scenario: Initialize the login-page description parameter
- **WHEN** an administrator runs the host initialization SQL
- **THEN** `sys_config` contains a built-in parameter record with key `sys.auth.pageDesc`
- **AND** the default value of that record is `Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model`

#### Scenario: Save a login-page description within 500 characters
- **WHEN** an administrator creates, updates, or imports `sys.auth.pageDesc` through system-parameter management and the value length is between 1 and 500 characters
- **THEN** the system accepts and stores the value
- **AND** `auth.pageDesc` returned by the public-frontend config endpoint matches the saved value

#### Scenario: Reject an overlong login-page description
- **WHEN** an administrator creates, updates, or imports `sys.auth.pageDesc` through system-parameter management and the value length exceeds 500 characters
- **THEN** the system rejects the change and returns a parameter-validation error
