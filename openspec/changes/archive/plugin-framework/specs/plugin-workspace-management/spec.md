## Requirements

### Requirement: Plugin Workspace Must Support Conversion from Submodule to Regular Directory

The system SHALL provide cross-platform developer tool commands to convert the fixed directory `apps/lina-plugins` from a Git submodule workspace to a regular directory while preserving existing plugin code. The conversion command must not delete plugin source code content; if `.gitmodules` contains other submodules, the command only removes the `apps/lina-plugins` corresponding configuration.

#### Scenario: Convert initialized plugin submodule
- **WHEN** a user runs `make plugins.init`
- **AND** `apps/lina-plugins` is an initialized submodule
- **THEN** the command removes the submodule/gitlink configuration for `apps/lina-plugins` in the parent repository
- **AND** the command removes the `apps/lina-plugins` corresponding section from `.gitmodules`
- **AND** the command cleans up `.git/config` and `.git/modules/apps/lina-plugins` corresponding submodule metadata
- **AND** the existing plugin files under `apps/lina-plugins` are preserved as regular workspace files

#### Scenario: Other submodules exist in .gitmodules
- **WHEN** a user runs `make plugins.init`
- **AND** `.gitmodules` contains both `apps/lina-plugins` and other submodule configurations
- **THEN** the command only removes the `apps/lina-plugins` corresponding section
- **AND** the `.gitmodules` file continues to retain other submodule configurations

#### Scenario: Workspace is already a regular directory
- **WHEN** a user runs `make plugins.init`
- **AND** `apps/lina-plugins` is already a regular directory and not a submodule
- **THEN** the command does not delete directory content
- **AND** the command outputs the regular directory status
- **AND** the command ends with success status

### Requirement: Plugin Sources Must Be Declared via hack/config.yaml

The system SHALL declare plugin sources through `plugins.sources` in `hack/config.yaml`. Each source SHALL contain `repo`, `root`, `ref`, and string array `items`. `apps/lina-plugins` is the fixed plugin maintenance directory and must not have the workspace path redundantly declared in configuration.

#### Scenario: Read official and custom plugin sources
- **WHEN** `hack/config.yaml` contains `plugins.sources.official` and `plugins.sources.custom`
- **AND** each source configures `repo`, `root`, `ref`, and `items`
- **THEN** the plugin tool resolves each `items` plugin ID per source
- **AND** each plugin's remote path is `<repo>@<ref>:<root>/<plugin-id>`
- **AND** the local target path is `apps/lina-plugins/<plugin-id>`

#### Scenario: items only accepts string arrays
- **WHEN** `plugins.sources.<name>.items` contains objects, numbers, or other non-string values
- **THEN** the plugin tool fails before executing any writes
- **AND** the error states `items` only supports plugin ID string arrays

#### Scenario: items uses wildcard to install all plugins under source root
- **WHEN** `plugins.sources.<name>.items` is configured as a string array containing `"*"`
- **THEN** the plugin tool scans one-level subdirectories under the source's `<root>` directory
- **AND** only directories containing `plugin.yaml` are expanded to plugin IDs
- **AND** expanded plugin IDs must still conform to plugin ID naming rules
- **AND** bare YAML alias `- *` is not a supported format; users must write the quoted `"*"`

#### Scenario: Wildcard cannot be mixed with explicit plugin IDs
- **WHEN** the same source's `items` contains both `"*"` and specific plugin IDs
- **THEN** the plugin tool fails before executing any writes
- **AND** the error states wildcard source cannot mix explicit plugin IDs

#### Scenario: Same-name plugin declared across sources
- **WHEN** multiple sources' `items` declare the same plugin ID
- **THEN** the plugin tool fails before executing any writes
- **AND** the error lists the conflicting plugin IDs and corresponding source names

#### Scenario: root uses unsafe path
- **WHEN** a source's `root` is an absolute path, contains `..`, is empty, or contains a platform drive path
- **THEN** the plugin tool rejects the configuration
- **AND** it does not access the remote repository or write to the local plugin directory

### Requirement: Plugin Install Must Copy from Configured Source to Fixed Directory

The system SHALL provide `make plugins.install` to read plugin directories from source repositories declared in `hack/config.yaml` and copy them to `apps/lina-plugins/<plugin-id>`. The install command must not write the source repository as a submodule or nested Git repository into the plugin directory.

#### Scenario: Install missing plugins from configuration
- **WHEN** a user runs `make plugins.install`
- **AND** `hack/config.yaml` declares plugin `multi-tenant`
- **AND** the remote `<root>/multi-tenant/plugin.yaml` exists
- **AND** local `apps/lina-plugins/multi-tenant` does not exist
- **THEN** the command copies the remote plugin directory to `apps/lina-plugins/multi-tenant`
- **AND** the target directory does not contain the source repository's `.git` metadata
- **AND** the command updates the plugin lock state file

#### Scenario: Still a submodule when installing
- **WHEN** a user runs `make plugins.install`
- **AND** `apps/lina-plugins` is still a submodule
- **THEN** the command fails and prompts to run `make plugins.init` first
- **AND** the command does not write to any plugin directory

#### Scenario: Target plugin directory already exists
- **WHEN** a user runs `make plugins.install`
- **AND** `apps/lina-plugins/<plugin-id>` already exists
- **THEN** the command fails by default and prompts to use `make plugins.update`
- **AND** unless the user explicitly passes `force=1`, the existing directory must not be overwritten

### Requirement: Plugin Update Must Protect Local Changes

The system SHALL provide `make plugins.update` to re-fetch plugin directories from sources in `hack/config.yaml` and update local regular directories. The update command SHALL block overwriting when the target plugin has local uncommitted changes or content inconsistent with the lock digest.

#### Scenario: Update plugin with no local changes
- **WHEN** a user runs `make plugins.update`
- **AND** `apps/lina-plugins/<plugin-id>` exists
- **AND** the plugin directory has no local uncommitted changes
- **AND** the remote source ref resolves to a new commit or content digest
- **THEN** the command updates the local plugin directory with the remote plugin directory
- **AND** the command updates the plugin lock state file's commit, version, and content digest

#### Scenario: Update encounters locally dirty plugin
- **WHEN** a user runs `make plugins.update`
- **AND** `apps/lina-plugins/<plugin-id>` has local uncommitted changes
- **THEN** the command fails by default
- **AND** the error lists the dirty plugin IDs
- **AND** the command does not overwrite local plugin content

#### Scenario: Force update dirty plugin
- **WHEN** a user runs `make plugins.update force=1`
- **AND** the target plugin has local uncommitted changes
- **THEN** the command allows overwriting the plugin directory
- **AND** the command outputs the forcibly overwritten plugin IDs
- **AND** the command updates the plugin lock state file

### Requirement: Plugin Status Check Must Provide Read-Only Diagnosis

The system SHALL provide `make plugins.status` to read-only check plugin workspace status, configured plugin status, local changes, plugin versions, lock state, and remote update state. The command must not modify `apps/lina-plugins`, `.gitmodules`, the parent repository Git index, or the lock state file.

#### Scenario: Status check on regular plugin workspace
- **WHEN** a user runs `make plugins.status`
- **AND** `apps/lina-plugins` is a regular directory
- **THEN** the command outputs workspace type as regular directory
- **AND** the command lists each configured plugin's local existence, source, ref, plugin version, and local dirty state

#### Scenario: Status check discovers still a submodule
- **WHEN** a user runs `make plugins.status`
- **AND** `apps/lina-plugins` is still a submodule
- **THEN** the command outputs submodule status
- **AND** the command prompts to run `make plugins.init` first
- **AND** the command does not modify submodule configuration

#### Scenario: Remote unreachable
- **WHEN** a user runs `make plugins.status`
- **AND** a source's remote repository is inaccessible
- **THEN** the command still outputs local state
- **AND** the remote update state for that source's plugins is marked as unknown
- **AND** the command returns the remote access error for diagnosis

#### Scenario: Local unconfigured plugins exist
- **WHEN** a user runs `make plugins.status`
- **AND** plugin directories exist under `apps/lina-plugins` that are not declared in `hack/config.yaml`
- **THEN** the command lists these unconfigured plugins
- **AND** the command does not delete or modify these directories

### Requirement: Plugin Lock State Must Be Tool-Maintained

The system SHALL record plugin install and update state in `apps/lina-plugins/.linapro-plugins.lock.yaml`. The lock state file is generated and updated by the tool for diagnosis and update decisions; the user-maintained plugin source configuration still uses `hack/config.yaml` as the single source of truth.

#### Scenario: Write lock state after install
- **WHEN** `make plugins.install` successfully installs a plugin
- **THEN** the lock state records the plugin ID, source, repo, root, ref, resolved commit, plugin manifest version, and content digest

#### Scenario: Configuration and lock state inconsistent
- **WHEN** a user runs `make plugins.status`
- **AND** a plugin exists in the lock state but is no longer declared in `hack/config.yaml`
- **THEN** the command outputs an orphaned lock entry diagnosis
- **AND** the command does not automatically delete the lock state or local plugin directory
