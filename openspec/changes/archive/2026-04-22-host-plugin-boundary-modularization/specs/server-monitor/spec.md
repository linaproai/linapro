## ADDED Requirements

### Requirement: Service monitoring is delivered by an independent source plugin

System SHALL delivers the service monitoring capability as a `monitor-server` source plugin instead of continuing as the host's default built-in module.

#### Scenario: Provides capabilities when the service monitoring plugin is enabled
- **WHEN** `monitor-server` is installed and enabled
- **THEN** Host exposure service monitoring collection, cleaning, query and page capabilities
- **AND** The plugin menu is mounted to the host `system monitoring` directory, and the top-level `parent_key` is `monitor`

#### Scenario: Smooth degradation when service monitoring plugin is missing
- **WHEN** `monitor-server` is not installed or not enabled
- **THEN** The host does not display the service monitoring menu and page entry
- **AND** Other monitoring plugins and host core capabilities continue to operate normally