## Requirements

### Requirement: One-Click Development Environment Startup

The system SHALL provide Makefile commands to support one-click startup of frontend and backend development environments. Repository-level development tool configuration SHALL be centrally placed in `hack/config.yaml`, where plugin source configuration SHALL use `plugins.sources` declaration and be read and executed by cross-platform `linactl` commands; default development entry points must not depend on Bash, PowerShell, or platform-specific scripts for plugin workspace management.

#### Scenario: Start development environment
- **WHEN** executing `make dev` in the project root directory
- **THEN** frontend and backend services start simultaneously

#### Scenario: Stop development environment
- **WHEN** executing `make stop` in the project root directory
- **THEN** frontend and backend services stop simultaneously

#### Scenario: Manage plugin sources through configuration
- **WHEN** `hack/config.yaml` configures `plugins.sources`
- **AND** a user runs `make plugins.install`, `make plugins.update`, or `make plugins.status`
- **THEN** the command reads the configuration through `linactl`
- **AND** the command executes plugin workspace management per the fixed `apps/lina-plugins` directory
- **AND** the command does not require users to maintain additional plugin path configuration

### Requirement: Host Initialization Commands Must Support Host-Only Workspace

The system SHALL allow developers to execute host basic initialization, build, startup, and test commands when the official source plugin workspace does not exist or is empty. Default host commands must not fail during Go workspace loading, frontend build initialization, or test discovery due to missing optional plugin directory.

#### Scenario: Backend host buildable when plugin workspace missing
- **WHEN** `apps/lina-plugins` does not exist
- **AND** a developer executes a host backend build command in `apps/lina-core` or the repository root
- **THEN** the backend host build succeeds or fails only due to host's own code errors
- **AND** the build must not fail due to missing `apps/lina-plugins`, `lina-plugins`, or `lina-plugin-*` modules

#### Scenario: Frontend host buildable when plugin workspace empty
- **WHEN** `apps/lina-plugins` is an empty directory
- **AND** a developer executes a host frontend type check or build command
- **THEN** the frontend build succeeds or fails only due to host frontend's own code errors
- **AND** plugin page scanning returns an empty set

#### Scenario: Host-only development service startup
- **WHEN** a developer runs `make dev` in a workspace without initializing the official plugin submodule
- **THEN** backend and frontend development services start in host-only mode
- **AND** source plugin-related capabilities degrade to empty set or empty state

### Requirement: Plugin Workspace Update Must Be Limited to Offline File Overwrite

The system SHALL define `plugins.install`, `plugins.update`, and direct plugin directory overwrite as development-stage offline file update capabilities. This capability can only write to or replace plugin files and tool lock state in `apps/lina-plugins/<plugin-id>` and must not modify the runtime database's effective plugin version, plugin install state, plugin enable state, plugin governance resources, or plugin business data.

#### Scenario: Update source plugin file without modifying runtime database
- **WHEN** a user runs `make plugins.update`
- **AND** the tool overwrites `apps/lina-plugins/plugin-demo` with a new version source plugin directory
- **THEN** the command only updates local plugin files and plugin lock state
- **AND** the command does not connect to the host runtime database
- **AND** the command does not modify `sys_plugin.version`, `release_id`, install state, or enable state

#### Scenario: Direct overwrite of source plugin directory
- **WHEN** a user manually overwrites plugin files in the `apps/lina-plugins/plugin-demo` directory
- **THEN** this operation only represents development-stage file changes
- **AND** the runtime database state still maintains the original effective version
- **AND** after host startup, the plugin runtime upgrade state marking handles file and database metadata differences

#### Scenario: Offline update does not execute plugin upgrade SQL
- **WHEN** a user installs or updates plugin files through the plugin workspace tool
- **THEN** the tool must not execute plugin `manifest/sql` upgrade SQL
- **AND** the tool must not call plugin custom upgrade callbacks
- **AND** the tool must not synchronize runtime governance resources such as menus, permissions, i18n, apidoc, routes, or cron
