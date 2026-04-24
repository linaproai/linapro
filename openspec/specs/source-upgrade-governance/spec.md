# source-upgrade-governance Specification

## Purpose
Define the framework-level source upgrade entry point, safety checks, metadata sourcing, and full host-SQL replay behavior used during development-time upgrades.

## Requirements
### Requirement: Framework metadata must be maintained centrally and shown directly in system info

The framework SHALL keep its name, version, description, homepage, repository URL, and license in `apps/lina-core/manifest/config/metadata.yaml`. The system-info API must return that metadata directly so the system-info page can render the project card without frontend hard-coded values.

#### Scenario: The system-info API returns framework metadata
- **WHEN** the management workbench requests system information
- **THEN** the response includes the framework name, version, description, homepage, repository URL, and license
- **AND** every value comes from the host `metadata.yaml`

### Requirement: Source upgrades must provide a unified development-time entry point

The framework SHALL provide a unified development-time source upgrade entry point invoked by `make upgrade`. That entry point MUST support both `scope=framework` and `scope=source-plugin`, and the implementation must remain in the repository-root upgrade tool rather than becoming a runtime `lina-core` command.

#### Scenario: Run a framework upgrade through `scope=framework`
- **WHEN** an operator runs `make upgrade confirm=upgrade scope=framework`
- **THEN** the system performs version checks, code replacement, and full host-SQL replay for a framework upgrade
- **AND** the operator does not need to decide manually where SQL execution should resume

#### Scenario: Run a source-plugin upgrade through `scope=source-plugin`
- **WHEN** an operator runs `make upgrade confirm=upgrade scope=source-plugin plugin=plugin-demo`
- **THEN** the system enters the source-plugin upgrade planning and execution flow
- **AND** it does not reinterpret the command as a dynamic-plugin runtime upgrade request

### Requirement: Source upgrades must complete safety checks before execution

The unified upgrade command SHALL remind operators to back up data and verify that the Git worktree is clean before any upgrade starts. If tracked files are modified, unstaged, or uncommitted, the command MUST refuse to continue.

#### Scenario: The worktree contains local changes before an upgrade
- **WHEN** an operator runs the upgrade command while the Git worktree contains modified, unstaged, or uncommitted files
- **THEN** the command refuses to continue
- **AND** it tells the operator to commit or stash the current changes first

### Requirement: Framework upgrades must read upgrade metadata only from hack config

The framework upgrade path SHALL read the current upgrade baseline from `frameworkUpgrade.version` in `apps/lina-core/hack/config.yaml` and compare it with the target upgrade version found in the target source's `apps/lina-core/hack/config.yaml`. The default upstream repository URL must also come from `frameworkUpgrade.repositoryUrl` in the same file unless the operator overrides it. The upgrade implementation MUST NOT read host runtime configuration.

#### Scenario: The target version is not higher than the current version
- **WHEN** the upgrade command resolves a target framework version that is less than or equal to the current project's `frameworkUpgrade.version`
- **THEN** the command reports that the project already uses the same or a higher framework version
- **AND** it does not overwrite code or execute SQL

#### Scenario: The upgrade command reads the upstream repository URL from hack config
- **WHEN** an operator does not pass `--repo`
- **THEN** the command reads the default upstream repository URL from `apps/lina-core/hack/config.yaml`
- **AND** it does not fall back to host runtime configuration

### Requirement: Framework upgrades must replay all host SQL from the first file

The framework upgrade path SHALL replay every host SQL file from the first file in sorted order after the target source code is applied. The process MUST NOT rely on SQL cursors or extra upgrade metadata tables.

#### Scenario: Framework upgrade replays host SQL from the beginning
- **WHEN** the upgrade command starts replaying host SQL
- **THEN** it begins with the first sorted host SQL file
- **AND** it continues in order until the final host SQL file

#### Scenario: Framework upgrade stops on the first SQL failure
- **WHEN** one host SQL file fails during replay
- **THEN** the command stops immediately
- **AND** it returns the failing SQL file and the error details
