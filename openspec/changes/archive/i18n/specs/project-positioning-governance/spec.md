## ADDED Requirements

### Requirement: Project positioning must be unified
The system SHALL define LinaPro's unified project positioning as "AI-driven full-stack development framework" and maintain this positioning consistently across project specifications, system metadata, entry copy, and repository documentation.

#### Scenario: Project-level descriptions use unified positioning
- **WHEN** developers write or update project specifications, system metadata, entry copy, or repository documentation
- **THEN** related content uses "AI-driven full-stack development framework" to describe LinaPro
- **AND** LinaPro is no longer defined as "only a backend management system" or equivalent expressions

#### Scenario: Management workspace capabilities described as subordinate
- **WHEN** documentation, page descriptions, or system metadata needs to mention menus, roles, workspace, system management and similar capabilities
- **THEN** these capabilities are described as LinaPro's default management workspace or built-in general modules
- **AND** they are not described as LinaPro's sole product boundary

### Requirement: Project-level entry copy must remain consistent
The system SHALL require project descriptions in repository entry documentation, system information page, API documentation metadata, and script output to maintain consistent positioning semantics.

#### Scenario: Multiple entries simultaneously display project descriptions
- **WHEN** a user views repository root documentation, system information page, system API documentation, or project script output
- **THEN** project descriptions across different entries maintain consistent positioning semantics
- **AND** no situation occurs where one entry describes a full-stack development framework while another still describes a backend management system
