## ADDED Requirements

### Requirement: Sensitive database bootstrap commands must require explicit confirmation
The system SHALL require the host `init` and `mock` commands to receive explicit confirmation values matching the command name before executing any SQL. When the value is missing or incorrect, the command MUST refuse to execute.

#### Scenario: `init` command missing confirmation value
- **WHEN** an operator executes `go run main.go init` without passing `--confirm=init`
- **THEN** the command refuses to execute database initialization SQL
- **AND** the command outputs a clear failure reason and correct example

#### Scenario: `mock` command receives incorrect confirmation value
- **WHEN** an operator executes `go run main.go mock --confirm=init`
- **THEN** the command refuses to execute any SQL in the `mock-data` directory
- **AND** the command requires using the confirmation value matching the `mock` command

#### Scenario: Command receives correct confirmation value
- **WHEN** an operator executes `go run main.go init --confirm=init` or `go run main.go mock --confirm=mock`
- **THEN** the command allows entering the corresponding SQL scanning and execution flow

### Requirement: Makefile entries must reuse the same confirmation semantics
The system SHALL require `make init` and `make mock` entries in the repository root and `apps/lina-core` directories to use the same `confirm` confirmation values as the command implementation, failing early when missing or incorrect.

#### Scenario: Root `make init` missing confirmation variable
- **WHEN** an operator executes `make init` in the repository root
- **THEN** the Makefile refuses to continue calling the backend initialization command
- **AND** outputs the correct example `make init confirm=init`

### Requirement: Database bootstrap SQL execution must stop on first failure
The system SHALL, after `init` or `mock` enters the execution phase, immediately stop executing subsequent SQL files when any SQL file fails, and return failure status to the caller.

#### Scenario: A SQL file fails during execution
- **WHEN** a SQL file returns an execution error during `init` or `mock` execution
- **THEN** the system immediately stops executing subsequent SQL files
- **AND** the command returns failure status to `make` or the direct caller
- **AND** the log contains the failed filename and error information
