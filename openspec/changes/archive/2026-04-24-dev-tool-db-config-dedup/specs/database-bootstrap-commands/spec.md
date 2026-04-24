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
