## ADDED Requirements

### Requirement: Manual job trigger must require confirmation

The Run Now action for scheduled jobs SHALL show a confirmation modal before triggering execution so administrators do not accidentally run operational tasks.

#### Scenario: Trigger action asks for confirmation
- **WHEN** an administrator clicks Run Now in the scheduled-job list
- **THEN** the frontend displays a confirmation modal
- **AND** no trigger API is called before confirmation

#### Scenario: Shell trigger remains available when shell editing is blocked
- **WHEN** a runnable shell job cannot be edited because of environment or permission limits
- **THEN** the row still shows a clickable Run Now action
