# database-bootstrap-commands Specification

## Purpose
Define the safety and execution rules for host bootstrap database commands, including confirmation, SQL asset sourcing, and fail-fast behavior.

## Requirements
### Requirement: Sensitive database bootstrap commands require explicit confirmation

The system SHALL require the host `init` and `mock` commands to receive an explicit confirmation value that matches the command name before executing any SQL. If the confirmation is missing or incorrect, the command MUST refuse to run. `init` and `mock` are limited to bootstrap initialization and must not act as formal upgrade commands.

#### Scenario: The `init` command is missing confirmation
- **WHEN** an operator runs `go run main.go init` without `--confirm=init`
- **THEN** the command refuses to execute initialization SQL
- **AND** the command prints a clear failure reason and a correct example

#### Scenario: The `mock` command receives the wrong confirmation value
- **WHEN** an operator runs `go run main.go mock --confirm=init`
- **THEN** the command refuses to execute any SQL under `mock-data`
- **AND** the command explains that the confirmation value must match `mock`

#### Scenario: The command receives the correct confirmation value
- **WHEN** an operator runs `go run main.go init --confirm=init` or `go run main.go mock --confirm=mock`
- **THEN** the command may enter the matching SQL scan and execution flow

#### Scenario: `init` does not create framework-upgrade bookkeeping
- **WHEN** an operator runs `go run main.go init --confirm=init` and every host SQL file succeeds
- **THEN** the command performs only host bootstrap initialization
- **AND** it does not write framework upgrade state, upgrade records, or SQL cursor metadata

### Requirement: `Makefile` entries must reuse the same confirmation semantics

The system SHALL require `make init` and `make mock` in both the repository root and `apps/lina-core` to use the same confirmation values as the command implementation, and to fail early when the confirmation value is missing or incorrect.

#### Scenario: Repository-root `make init` is missing confirmation
- **WHEN** an operator runs `make init` from the repository root without `confirm=init`
- **THEN** the `Makefile` refuses to continue
- **AND** it prints the correct example `make init confirm=init`

#### Scenario: Backend `make mock` uses the correct confirmation variable
- **WHEN** an operator runs `make mock confirm=mock` from `apps/lina-core`
- **THEN** the `Makefile` passes the confirmation value through to the backend command implementation
- **AND** the backend command continues with `mock`-specific validation and execution

### Requirement: Database bootstrap commands must choose SQL asset sources explicitly by execution phase

The system SHALL make SQL asset sourcing explicit. Runtime `lina init` and `lina mock` commands read host SQL assets from embedded FS by default, while development-time `make init` and `make mock` commands must explicitly switch to local SQL files in the source tree. The implementation MUST not infer the source from the current working directory.

#### Scenario: Runtime `init` reads embedded SQL by default
- **WHEN** an operator runs `lina init --confirm=init` from a released binary
- **THEN** the command reads embedded SQL assets from `manifest/sql/`
- **AND** it does not require a local source tree to exist

#### Scenario: Development-time `make mock` reads local SQL explicitly
- **WHEN** a developer runs `make mock confirm=mock`
- **THEN** the `Makefile` explicitly switches the command to the local SQL source
- **AND** the command reads SQL from `manifest/sql/mock-data/` in the source tree

### Requirement: Database bootstrap SQL execution must fail fast

The system SHALL stop execution immediately when any SQL file fails during `init` or `mock`, and it shall return a failure result to the caller.

#### Scenario: A SQL file fails during execution
- **WHEN** one SQL file returns an execution error during `init` or `mock`
- **THEN** the system stops executing later SQL files immediately
- **AND** the command returns a failure status to `make` or the direct caller
- **AND** logs include the failing file name and the error details

#### Scenario: Every SQL file succeeds
- **WHEN** every target SQL file succeeds during `init` or `mock`
- **THEN** the command returns a success status
- **AND** logs print the corresponding completion message

### Requirement: SQL bootstrap commands must not depend on driver multi-statement execution

The system SHALL parse each SQL file used by `init` and `mock` into an ordered list of independent statements and execute them one by one instead of relying on driver-level multi-statement support in the database connection string.

#### Scenario: Multi-statement files run statement by statement in order
- **WHEN** `init` or `mock` reads a target file that contains multiple SQL statements
- **THEN** the system executes those statements one by one in the same order they appear in the file
- **AND** blank fragments and pure comment fragments are not treated as executable statements

#### Scenario: Execution stops immediately after a statement failure
- **WHEN** `init` or `mock` receives a database error while executing a middle statement from a SQL file
- **THEN** the system immediately stops the remaining statements in that file and all later SQL files
- **AND** the command returns a failure status
- **AND** the error message still includes the failing file name so the issue can be located quickly
