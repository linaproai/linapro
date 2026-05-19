## MODIFIED Requirements

### Requirement: One-click development environment startup

The system SHALL provide Makefile commands supporting one-click startup of frontend and backend development environments, and SHALL provide independent local development environment check and initialization entries.

#### Scenario: Start development environment
- **WHEN** `make dev` is executed at the project root
- **THEN** frontend and backend services start simultaneously

#### Scenario: Stop development environment
- **WHEN** `make stop` is executed at the project root
- **THEN** frontend and backend services stop simultaneously

#### Scenario: Check development environment
- **WHEN** `make env.check` is executed at the project root
- **THEN** the system SHALL display a bordered table showing local Go, Node.js, pnpm, Vite, Playwright, and PostgreSQL with name, current version, required version, satisfied status, and remarks
- **AND** the check SHALL NOT start development services, connect to business databases, or modify local dependencies or configuration files

#### Scenario: Initialize development environment
- **WHEN** `make env.setup` is executed at the project root
- **THEN** the system SHALL install or confirm frontend dependencies and install Playwright Chromium browser with required system dependencies
- **AND** the command SHALL inherit the functional semantics of the original `make dev.setup`

#### Scenario: Old initialization entry is removed
- **WHEN** a developer executes `make dev.setup` or `linactl dev.setup`
- **THEN** the system SHALL NOT treat it as a supported development environment initialization entry
- **AND** help information and fix prompts SHALL point to `make env.setup`
