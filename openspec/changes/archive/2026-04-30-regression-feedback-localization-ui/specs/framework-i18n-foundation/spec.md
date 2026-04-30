## ADDED Requirements

### Requirement: English regression sweep must cover framework-delivered pages and seed display content

The default management workbench SHALL provide English regression coverage for framework-delivered pages from manual feedback, ensuring system-generated content, default seed display content, and static UI copy do not retain Chinese text.

#### Scenario: English regression pages contain no Chinese system copy
- **WHEN** an administrator opens framework-delivered pages in `en-US`
- **THEN** titles, buttons, form labels, table columns, generated nodes, built-in displays, and confirmation modals use English

### Requirement: Runtime locale JSON values must avoid markdown-only code markers

Runtime translation JSON SHALL avoid markdown-style backtick markers in user-visible strings because ordinary UI rendering does not apply code highlighting.

#### Scenario: Locale JSON strings are displayed as plain UI text
- **WHEN** locale JSON strings contain file paths, examples, wildcards, or extensions
- **THEN** strings display the content directly without backticks
