## MODIFIED Requirements

### Requirement: Login Returns User Information

The system MUST return user information including roles and menu tree after successful login.

#### Scenario: Successful login returns user information
- **WHEN** a user logs in with correct username and password
- **THEN** the system returns userId, username, realName, email, avatar
- **AND** the system returns roles field containing the user's role key list
- **AND** the system returns menus field containing the user's accessible menu tree
- **AND** the system returns permissions field containing the user's permission identifier list
- **AND** the system returns homePath field specifying the user's home page path

#### Scenario: Super admin login
- **WHEN** a user with admin role logs in
- **THEN** the system returns all menus (without checking sys_role_menu associations)
- **AND** roles includes "admin"
- **AND** permissions includes "*:*:*" indicating all permissions

#### Scenario: Regular user login
- **WHEN** a non-super-admin user logs in
- **THEN** the system queries sys_role_menu based on user's roles to get menu IDs
- **AND** the system builds the menu tree based on menu IDs
- **AND** the system filters out disabled menus (status=0)
- **AND** the system filters out hidden menus (visible=0)

#### Scenario: User with no roles logs in
- **WHEN** a user with no assigned roles logs in
- **THEN** the system returns an empty menu tree
- **AND** roles is an empty array
- **AND** permissions is an empty array

#### Scenario: All user roles disabled
- **WHEN** all of a user's roles are disabled
- **THEN** the system returns an empty menu tree
- **AND** roles is an empty array
- **AND** permissions is an empty array

## ADDED Requirements

### Requirement: Menu Tree Structure

The returned menu tree MUST conform to frontend route generation requirements.

#### Scenario: Menu tree contains required fields
- **WHEN** the system returns the menu tree
- **THEN** each menu node contains id, parentId, name, path, component, icon, type, sort, visible, status fields
- **AND** directory type (type="D") menus contain children child nodes
- **AND** menu type (type="M") menus are leaf nodes
- **AND** button type (type="B") menus are not returned in the menu tree

#### Scenario: Menu tree sorted by sort field
- **WHEN** the system returns the menu tree
- **THEN** sibling menus are sorted by sort field in ascending order

### Requirement: Permission Identifier List

The system MUST return all permission identifiers for the user.

#### Scenario: Permission identifier aggregation
- **WHEN** a user has multiple roles
- **THEN** the system aggregates all roles' permission identifiers (deduplicated)
- **AND** permission identifiers come from the perms field of menus with type="M" or type="B"

#### Scenario: Super admin permissions
- **WHEN** a user is super admin (has admin role)
- **THEN** permissions returns ["*:*:*"]
- **AND** the frontend treats this permission identifier as having all permissions
