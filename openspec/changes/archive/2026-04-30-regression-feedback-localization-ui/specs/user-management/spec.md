## ADDED Requirements

### Requirement: User list role names must match backend-localized role display

The user management list SHALL use role display names returned by the backend and keep built-in role display consistent with role management in the same language.

#### Scenario: User list shows administrator role in English
- **WHEN** an administrator opens user management in `en-US`
- **THEN** the administrator role display matches role management
