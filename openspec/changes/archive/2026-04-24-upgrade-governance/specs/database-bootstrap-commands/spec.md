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
