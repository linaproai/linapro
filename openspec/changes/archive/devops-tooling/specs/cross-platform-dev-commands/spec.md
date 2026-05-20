## ADDED Requirements

### Requirement: Cross-platform command primary entry point

The system MUST provide a cross-platform development command primary entry point that does not depend on GNU Make or POSIX Shell for executing common development, build, initialization, and verification tasks.

#### Scenario: Execute common commands via Go CLI

- **WHEN** a developer executes the cross-platform command entry at the project root with a supported command such as `dev`, `stop`, `status`, `build`, `wasm`, `init`, `mock`, `test`, `test-go`, or `help`
- **THEN** the system MUST execute the corresponding task through the cross-platform implementation rather than requiring the caller to install GNU Make or POSIX Shell tooling

#### Scenario: Missing external professional tool

- **WHEN** a supported command needs to call an external professional tool such as `go`, `pnpm`, `docker`, `kubectl`, `gf`, or `playwright` but the tool is missing from the current environment
- **THEN** the system MUST output the missing tool name, current command context, and an actionable installation or fix prompt

### Requirement: Windows make.cmd compatibility entry point

The system MUST provide a Windows `make.cmd` thin wrapper at the project root that delegates all real task logic to the cross-platform command primary entry point.

#### Scenario: cmd.exe uses make-style commands

- **WHEN** a Windows `cmd.exe` user executes `make dev` at the project root
- **THEN** `make.cmd` MUST forward the command and parameters to the cross-platform command primary entry point to execute the `dev` task

#### Scenario: PowerShell uses current-directory make entry

- **WHEN** a Windows PowerShell user executes `.\make dev` or `.\make.cmd dev` at the project root
- **THEN** `make.cmd` MUST forward the command and parameters to the cross-platform command primary entry point without requiring a separate `make.ps1`

#### Scenario: Parameter forwarding

- **WHEN** a Windows user executes `make init confirm=init rebuild=true`
- **THEN** `make.cmd` MUST forward `init`, `confirm=init`, and `rebuild=true` parameters as-is and return the cross-platform command primary entry point's exit code

### Requirement: Makefile compatibility layer

The system MUST retain the existing `Makefile` entry as a compatibility layer for Linux/macOS and existing CI, and MUST NOT continue maintaining complex shell orchestration logic in migrated targets.

#### Scenario: Existing make targets remain available

- **WHEN** a Linux/macOS developer executes existing `make dev`, `make build`, `make init confirm=init`, and similar commands
- **THEN** the system MUST maintain target semantics compatibility and implement task orchestration through the cross-platform command primary entry point or equivalent cross-platform tool

#### Scenario: Development services start asynchronously with final status output

- **WHEN** a developer executes `make dev` or the cross-platform command primary entry point `dev`
- **THEN** the system MUST start backend and frontend as async development service processes detachable from the current command lifecycle, record corresponding PID files, and output the current frontend/backend process status table as the terminal's final output after readiness checks complete

#### Scenario: Thin wrapper consistency

- **WHEN** the same target supports `make <target>`, `make.cmd <target>`, and the cross-platform command primary entry point
- **THEN** all three entries MUST call the same task implementation and maintain consistent parameter semantics, exit codes, and key error messages

### Requirement: Make-style parameter compatibility

The cross-platform command primary entry point MUST support existing make-style `key=value` parameters and map them to internal command options.

#### Scenario: Initialization parameter compatibility

- **WHEN** a developer executes the cross-platform command primary entry point `init confirm=init rebuild=true`
- **THEN** the system MUST recognize `confirm=init` and `rebuild=true` as initialization task parameters and maintain behavior equivalent to existing `make init confirm=init rebuild=true`

#### Scenario: Build parameter compatibility

- **WHEN** a developer executes the cross-platform command primary entry point `build platforms=linux/amd64,linux/arm64 verbose=1`
- **THEN** the system MUST recognize `platforms` and `verbose` as build task parameters and execute the build flow with specified platforms and log mode

#### Scenario: Plugin parameter compatibility

- **WHEN** a developer executes the cross-platform command primary entry point `wasm p=plugin-demo-dynamic`
- **THEN** the system MUST execute dynamic Wasm build only for the specified plugin and maintain behavior equivalent to existing `make wasm p=plugin-demo-dynamic`

### Requirement: Shell dependency governance

The system MUST migrate common command-path logic such as file copying, directory cleanup, process startup, port detection, HTTP readiness, plugin discovery, and help output to cross-platform implementations, avoiding Linux/macOS-specific commands.

#### Scenario: File copying and directory cleanup

- **WHEN** the build flow needs to prepare embedded resources or clean output directories
- **THEN** the system MUST use cross-platform file system implementations for directory creation, deletion, and copying rather than depending on `rm`, `cp`, `mkdir`, or `touch`

#### Scenario: Service status and stop

- **WHEN** a developer executes `status` or `stop`
- **THEN** the system MUST use cross-platform methods to check ports, PID files, process startup records, and service readiness, avoiding `lsof`, `pgrep`, `ps`, `kill`, `sed`, `head`, or `tr`

#### Scenario: Status command table output

- **WHEN** a developer executes `status` or `dev` completes service startup
- **THEN** the system MUST display backend and frontend service name, running status, URL, PID, PID file, and log path in a cross-platform stable table format

#### Scenario: Help output

- **WHEN** a developer executes `help`
- **THEN** the system MUST output available command descriptions from cross-platform command definitions or equivalent structures, not depending on `awk`, `sort`, or POSIX pipes

### Requirement: Documentation and verification

The system MUST provide bilingual documentation and automated verification for the cross-platform command entry, ensuring Windows, Linux, and macOS usage is clear and executable.

#### Scenario: README shows cross-platform entry

- **WHEN** a developer reads the root README and its Chinese mirror
- **THEN** documentation MUST clearly show the cross-platform recommended command, Windows `cmd.exe` `make` usage without `.cmd` suffix, PowerShell `.\make` usage, explicit `.\make.cmd` usage when needed, and Linux/macOS compatibility `make` usage

#### Scenario: Tool documentation synchronization

- **WHEN** a new cross-platform tool directory is added
- **THEN** the directory MUST provide both English `README.md` and Chinese `README.zh-CN.md` simultaneously, documenting commands, parameters, examples, output, and verification methods

#### Scenario: Automated verification

- **WHEN** cross-platform command entry implementation is complete
- **THEN** the system MUST cover parameter parsing, entry forwarding, file operations, plugin discovery, help output, and Makefile thin-wrapper consistency through Go unit tests, command-level smoke tests, static scanning, or equivalent governance verification

### Requirement: GitHub Actions Windows verification

The system MUST add Windows runner basic command verification in `.github/workflows/` workflows related to building, testing, or tool verification, ensuring cross-platform command entry and `make.cmd` work on Windows.

#### Scenario: Windows runner verifies Go CLI basic commands

- **WHEN** GitHub Actions executes cross-platform command verification on a `windows-latest` runner
- **THEN** the workflow MUST verify at least `go run ./hack/tools/linactl help`, `go run ./hack/tools/linactl status`, and one file or plugin tool command that does not depend on external services can execute successfully

#### Scenario: Windows runner verifies make.cmd

- **WHEN** GitHub Actions executes Windows entry verification on a `windows-latest` runner
- **THEN** the workflow MUST verify `make.cmd` can forward parameters to the cross-platform command primary entry point and cover at least lightweight commands like `help` or `status`

#### Scenario: cmd.exe and PowerShell usage

- **WHEN** GitHub Actions verifies Windows command entry
- **THEN** the workflow MUST cover `make <target>` or `make.cmd <target>` usage in `cmd.exe` and `.\make <target>` or `.\make.cmd <target>` usage in PowerShell, ensuring both documented Windows terminal entries can execute

#### Scenario: Windows CI cost control

- **WHEN** basic commands require external databases, Docker daemon, long-running dev services, or full frontend dependencies to execute
- **THEN** the workflow MUST use lightweight smoke, dry-run, skip heavy paths, or split subsequent verification to avoid binding Windows basic command verification to unstable or high-cost external dependencies

### Requirement: Runtime impact isolation

The cross-platform command entry change MUST NOT alter LinaPro backend runtime API, database schema, permission model, data permissions, plugin runtime contracts, frontend user interface, runtime i18n resources, or business cache behavior.

#### Scenario: Runtime resources are unaffected

- **WHEN** the cross-platform command entry change is complete
- **THEN** the system MUST NOT add or modify runtime menus, buttons, forms, API messages, plugin manifest copy, runtime i18n JSON, manifest i18n, apidoc i18n, or business cache invalidation logic

#### Scenario: External tools are wrapped, not rewritten

- **WHEN** the cross-platform command entry needs to execute building, code generation, container images, deployment, or testing tasks
- **THEN** the system MUST call `go`, `pnpm`, `docker`, `kubectl`, `gf`, `playwright`, and other professional tools through clear command wrappers, and MUST NOT rewrite these tools' core capabilities in this change
