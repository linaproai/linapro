## MODIFIED Requirements

### Requirement: Sensitive database bootstrap commands require explicit confirmation

The system SHALL require the host `init` and `mock` commands to receive an explicit confirmation value that matches the command name before executing any SQL. If the confirmation is missing or incorrect, the command MUST refuse to run. `init` and `mock` are limited to host bootstrap assets, must not act as formal upgrade commands, and must not replace `upgrade`.

#### Scenario: The `init` command is missing confirmation
- **WHEN** an operator runs `go run main.go init` without `--confirm=init`
- **THEN** the command refuses to execute initialization SQL
- **AND** the command prints a clear failure reason and a correct example

#### Scenario: The `mock` command receives the wrong confirmation value
- **WHEN** an operator runs `go run main.go mock --confirm=init`
- **THEN** the command refuses to execute any SQL under `mock-data`
- **AND** the command explains that the confirmation value must match `mock`

#### Scenario: `init` does not create framework-upgrade bookkeeping
- **WHEN** an operator runs `go run main.go init --confirm=init` and every host SQL file succeeds
- **THEN** the command performs only host bootstrap initialization
- **AND** it does not write framework upgrade state, upgrade records, or SQL cursor metadata

### Requirement: Database bootstrap commands must select SQL asset sources by execution phase

The system SHALL make SQL asset sourcing explicit. Runtime `lina init` and `lina mock` commands read host SQL assets from embedded FS by default, while development-time `make init` and `make mock` commands must explicitly switch to local SQL files in the source tree. The implementation MUST not infer the source implicitly from the current working directory.

#### Scenario: Runtime `init` reads embedded SQL by default
- **WHEN** an operator runs `lina init --confirm=init` from a released binary
- **THEN** the command reads embedded SQL assets from `manifest/sql/`
- **AND** it does not require a local source tree to exist

#### Scenario: Development-time `make mock` reads local SQL explicitly
- **WHEN** a developer runs `make mock confirm=mock`
- **THEN** the `Makefile` explicitly switches the command to the local SQL source
- **AND** the command reads SQL from `manifest/sql/mock-data/` in the source tree

## ADDED Requirements

### Requirement: SQL bootstrap commands must not depend on driver multi-statement execution

The system SHALL parse each SQL file used by the host `init` and `mock` commands into an ordered list of independent SQL statements and execute them one by one instead of relying on driver-level `multiStatements` support in the database connection string. If any statement fails, the system MUST stop the remaining statements in the current file and all later SQL files, and return a failure result to the caller.

#### Scenario: Multi-statement SQL files run statement by statement in order
- **WHEN** `init` or `mock` reads a target file that contains multiple SQL statements
- **THEN** the system executes those statements one by one in the same order they appear in the file
- **AND** blank fragments and pure comment fragments are not treated as executable statements

#### Scenario: Execution stops immediately after a statement failure
- **WHEN** `init` or `mock` receives a database error while executing a middle statement from a SQL file
- **THEN** the system immediately stops the remaining statements in that file and all later SQL files
- **AND** the command returns a failure status
- **AND** the error message still includes the failing file name so the issue can be located quickly
