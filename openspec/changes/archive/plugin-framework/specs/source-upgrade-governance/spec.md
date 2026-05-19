## Requirements

### Requirement: Framework Metadata Must Be Centrally Maintained and Displayed in System Information

The framework SHALL store its name, version, description, homepage, repository URL, and license in `apps/lina-core/manifest/config/metadata.yaml`. The system information API must directly return this metadata so the system information page can render the project card without frontend hardcoded values.

#### Scenario: System information API returns framework metadata
- **WHEN** the management workspace requests system information
- **THEN** the response includes framework name, version, description, homepage, repository URL, and license
- **AND** each value comes from the host `metadata.yaml`

## REMOVED Requirements

### Requirement: Source Upgrade Must Provide Unified Development-Time Entry

**Reason**: The old development-time upgrade entry was an unimplemented command and should not continue as the framework or source plugin upgrade specification requirement. Source plugin and dynamic plugin file overwrites are development-stage offline updates; runtime state and data upgrades must be handled through the host's post-startup plugin runtime upgrade flow.

**Migration**: Delete old development-time upgrade entry related documentation, tasks, and error prompts. Source plugin offline file updates use `plugins.update` or direct plugin directory overwrite; plugin runtime upgrade uses the plugin management page and `POST /plugins/{id}/upgrade`. Framework source upgrade itself is not in scope of this plugin runtime upgrade change.

### Requirement: Source Upgrade Must Complete Safety Check Before Execution

**Reason**: This requirement is specific to the old development-time upgrade command. Since the command is not a current delivery capability, the related Git working tree safety check must not be retained as a specification requirement for a non-existent command.

**Migration**: Local change protection for plugin offline file overwrite is handled by `plugin-workspace-management`'s `plugins.update` dirty blocking and `force=1` rules. Runtime upgrade safety checks are handled by plugin runtime upgrade pre-check, permission validation, and confirmation dialog.

### Requirement: Framework Upgrade Must Only Read Upgrade Metadata from hack Config

**Reason**: This requirement binds to the old framework source code upgrade tool and upgrade entry, does not belong to current implemented capabilities, and is not part of plugin runtime upgrade scope.

**Migration**: Retain framework metadata display capability. If framework source upgrade is redesigned in the future, it should be proposed as a separate change without reusing the plugin runtime upgrade flow.

### Requirement: Framework Upgrade Must Replay All Host SQL from First File

**Reason**: This requirement binds to the old framework source code upgrade tool and host SQL replay strategy, does not belong to current plugin runtime upgrade capabilities, and retaining it would mislead users into thinking a development-time upgrade entry exists.

**Migration**: Plugin runtime upgrade only executes the target plugin's own `manifest/sql` upgrade resources and governance synchronization, and is not responsible for full host framework SQL replay. Future framework upgrade strategy requires independent design and implementation.
