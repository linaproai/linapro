## MODIFIED Requirements

### Requirement: Username Password Login
The system SHALL supports username + password login, and returns JWT Token after successful verification. SHALL emits unified login lifecycle events during the login process (whether successful or failed). After successful login, SHALL creates session records in the `sys_online_session` table maintained by the host; if `monitor-loginlog` is enabled, the plugin completes the login log entry based on login events.

#### Scenario: Login successful
- **WHEN** User submits correct username and password to `POST /api/v1/auth/login`
- **THEN** The system returns JWT Token, and the response format is `{code: 0, message: "ok", data: {token: "..."}}`
- **AND** The host creates session records (including token_id, user information, IP, browser, operating system, etc.) in the `sys_online_session` table
- **AND** The host emits a successful login event; if `monitor-loginlog` is enabled, the plugin writes a successful login log

#### Scenario: Login failed and log plugin missing
- **WHEN** User login failed and `monitor-loginlog` is not installed, not enabled or failed to initialize
- **THEN** The system still returns the correct login failure result
- **AND** The host does not report an error due to lack of specific login log persistence implementation

### Requirement: User logs out
The system SHALL supports user logout operations. The logout operation SHALL emits unified login life cycle events and deletes the online session records maintained by the host.

#### Scenario: Logout successful
- **WHEN** Logged in user calls `POST /api/v1/auth/logout`
- **THEN** The system returns a successful response, deletes the user's session record from the `sys_online_session` table, and the front end clears the locally stored Token
- **AND** The host emits a successful logout event; if `monitor-loginlog` is enabled, the plugin will write the corresponding log

### Requirement: Certified Middleware
The system SHALL provides authentication middleware to protect APIs that require login to access. The middleware MUST verify both the JWT signature and the validity of the session maintained by the host, and MUST not rely on whether `monitor-online` is installed.

#### Scenario: Access protected interface when online user plugin is not installed
- **WHEN** The request header carries a valid `Authorization: Bearer <token>`, the corresponding session record exists in `sys_online_session`, and `monitor-online` is not installed or enabled
- **THEN** The middleware still updates the session's `last_active_time` to the current time through the UPDATE operation
- **AND** The request is processed normally and user information is injected into the request context.
