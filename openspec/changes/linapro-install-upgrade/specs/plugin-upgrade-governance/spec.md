## MODIFIED Requirements

### Requirement: Source-plugin upgrades must be explicit development-time operations

The system SHALL require source-plugin upgrades to be executed through the `lina-upgrade` Claude Code skill (located at `.claude/skills/lina-upgrade/`) instead of being repaired automatically during host startup. The skill MUST support both single-plugin and bulk source-plugin upgrades through its source-plugin sub-flow.

#### Scenario: Upgrade one source plugin explicitly

- **WHEN** a developer instructs the AI to upgrade a single source plugin (e.g., "upgrade plugin-demo to the version found in source")
- **THEN** the AI invokes the `lina-upgrade` skill in source-plugin sub-flow with the plugin id `plugin-demo`
- **AND** the skill generates and executes an upgrade plan only for `plugin-demo`
- **AND** does not trigger upgrades for other source plugins or any dynamic plugin

#### Scenario: Upgrade all source plugins in one run

- **WHEN** a developer instructs the AI to upgrade all source plugins in bulk
- **THEN** the AI invokes the `lina-upgrade` skill in source-plugin sub-flow with the special selector `all`
- **AND** the skill scans all source plugins and processes pending upgrades in a deterministic order
- **AND** prints explicit skip results for plugins that are not installed or do not require upgrades

### Requirement: Host startup must verify that source-plugin upgrades are complete

The host SHALL scan source plugins during startup and then validate whether any installed source plugin has a higher discovered version than the effective version. If such a plugin exists, the host MUST refuse to start and print the matching upgrade instruction directing the user to invoke the `lina-upgrade` skill.

#### Scenario: A pending source-plugin upgrade blocks startup

- **WHEN** the host starts and discovers that `plugin-demo` is effectively running `v0.1.0` while source discovery reports `v0.5.0`
- **THEN** the startup flow fails
- **AND** the error message includes the plugin ID, the effective version, the discovered version, and the recommended action (invoke the `lina-upgrade` skill via the AI tooling, with example phrasing such as "ask Claude Code to run lina-upgrade for plugin-demo")

### Requirement: Dynamic-plugin upgrades stay on the runtime model

The system SHALL keep dynamic-plugin upgrades on the existing runtime upload plus install/reconcile model. The development-time `lina-upgrade` skill MUST NOT scan, migrate, or switch dynamic-plugin releases.

#### Scenario: The development-time upgrade skill ignores dynamic plugins

- **WHEN** a developer invokes the `lina-upgrade` skill in source-plugin sub-flow
- **THEN** the skill does not scan or switch any dynamic-plugin release
- **AND** dynamic plugins continue to upgrade only through upload plus install/reconcile
