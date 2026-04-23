# User Auth

## Purpose

The user authentication function is responsible for system login, logout, session verification, login state maintenance, and user information and permission data return after login to ensure stable and consistent frontend and backend authentication processes.
## Requirements
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

### Requirement: Log in and return user information

The system SHALL returns user information, including roles and menu trees, after a successful user login.

#### Scenario: Login successful User information returned
- **WHEN** User logs in with correct username and password
- **THEN** The system returns userId, username, realName, email, avatar
- **AND** The system returns a roles field containing a list of keys for all roles of the user
- **AND** The system returns a menus field containing a user-accessible menu tree
- **AND** The system returns a permissions field containing a list of permission identifiers owned by the user
- **AND** The system returns the homePath field, specifying the user's home page path

#### Scenario: Super Admin Login
- User login for the **WHEN** admin role
- **THEN** The system returns all menus (without checking the sys_role_menu association)
- **AND** roles contain "admin"
- **AND** permissions contain "*: *: *" for all permissions

#### Scenario: Normal user login
- **WHEN** Non-Super Admin User Login
- **THEN** The system queries sys_role_menu based on the user's role to get a list of menu IDs
- **AND** The system builds a menu tree based on a list of menu IDs
- **AND** The system filters out menus with inactive status (status = 0)
- **AND** The system filters out hidden menus (visible = 0)

#### Scenario: User is logged in without role
- **WHEN** User logged in without assigning any role
- **THEN** The system returns an empty menu tree
- **AND** roles is an empty array
- **AND** permissions is an empty array

#### Scenario: Deactivate all user roles
- **WHEN** All roles of the user are deactivated
- **THEN** The system returns an empty menu tree
- **AND** roles is an empty array
- **AND** permissions is an empty array

### Requirement: Certified Lifecycle Events Available for Plugin Subscription
The system SHALL publishes authentication lifecycle events such as login success and logout success as controlled hooks to enabled plugins.

#### Scenario: Publish authentication event after successful login
- **WHEN** User logged in successfully
- **THEN** Host distributes events to plugins subscribed to `auth.login.succeeded'
- **AND** The event contains the host's exposed user identity and client context

#### Scenario: Post certification event after successful logout
- **WHEN** User logged out successfully
- **THEN** Host distributes events to plugins subscribed to `auth.logout.succeeded'
- **AND** Event distribution does not change the original logout success semantics

### Requirement: Plugin Authentication Extension Failure Does Not Affect Authentication Results
The system SHALL guarantees that the extension failure of the plugin on the authentication event will not change the final result of the login or logout.

#### Scenario: Login successful Plugin error in Hook
- **WHEN** A plugin failed in logon success event handling
- **THEN** users still receive successful login results
- **AND** The system logs the plugin failure information for troubleshooting purposes

### Requirement: JWT Token validity period configuration
System SHALL supports configuring the JWT Token validity period through the `jwt.expire` duration string in `config.yaml`.

#### Scenario: Use the new duration to configure the Token validity period
- **WHEN** Administrator sets `jwt.expire=24h` in `config.yaml`
- **THEN** The system MUST use this duration value as the JWT Token validity period

## ADDED Requirements

### Requirement: Menu tree structure

The menu tree returned by the system MUST meet the requirements for frontend route generation.

#### Scenario: Menu tree contains necessary fields
- **WHEN** The system returns to the menu tree
- **THEN** Each menu node contains id, parentId, name, path, component, icon, type, sort, visible, status fields
- **AND** The menu of directory type (type="D") contains children child nodes
- **AND** The menu of menu type (type="M") is a leaf node
- **AND** Button type (type="B") is not returned in the menu tree

#### Scenario: Menu tree sorted by sort field
- **WHEN** The system returns to the menu tree
- **THEN** The same level menu is sorted in ascending order by the sort field.

### Requirement: List of permission identifiers

System SHALL returns all permission IDs of the user.

#### Scenario: Permission ID aggregation
- **WHEN** User has multiple roles
- **THEN** The system aggregates the permission identifiers of all roles (removal of duplicates)
- **AND** The permission ID comes from the perms field of type="M" or type="B" in the menu table

#### Scenario: Super administrator privileges
- **WHEN** The user is a super administrator (has admin role)
- **THEN** permissions returns ["*:*:*"]
- **AND** The front end determines that this permission mark has all permissions

### Requirement: Runtime-Configured JWT Expiry
The system SHALL allow `sys.jwt.expire` to control the lifetime of newly issued JWT tokens at runtime and SHALL fall back to static configuration when no runtime override exists.

#### Scenario: Runtime JWT expiry takes effect
- **WHEN** an administrator maintains `sys.jwt.expire=24h`
- **THEN** newly issued JWT tokens use that duration as their effective expiry time

### Requirement: Runtime-Configured Login IP Blacklist
The system SHALL allow `sys.login.blackIPList` to control login IP blacklisting at runtime.

#### Scenario: Login request is denied by the configured blacklist
- **WHEN** a login request originates from an IP or CIDR range matched by `sys.login.blackIPList`
- **THEN** the system rejects the login attempt
- **AND** the login log records the failure reason that the login IP is forbidden
