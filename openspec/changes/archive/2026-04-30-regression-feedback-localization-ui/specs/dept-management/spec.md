## ADDED Requirements

### Requirement: System-generated Unassigned node must be localized

System-generated virtual Unassigned department nodes SHALL be localized by current request language and maintained through backend or plugin runtime i18n resources.

#### Scenario: Unassigned displays in English
- **WHEN** an administrator opens a page with a department-tree filter in `en-US`
- **THEN** the virtual node displays as `Unassigned`
