## MODIFIED Requirements

### Requirement: Built-in role display must be localized consistently

Built-in protected roles and default seed roles SHALL display according to current language in framework-delivered pages. User management and role management MUST show the same localized display value for the same built-in role.

#### Scenario: Default user role displays in English
- **WHEN** an administrator opens role management in `en-US`
- **THEN** the default user role displays an English name

#### Scenario: User management role display matches role management
- **WHEN** an administrator opens user management in `en-US`
- **THEN** the administrator user's role name matches the English role-management display
