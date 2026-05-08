## MODIFIED Requirements

### Requirement: Automatic recording of operation logs

The system SHALL use the global audit middleware declared on the host unified HTTP registration portal through the `monitor-operlog` source plugin to automatically emit unified audit events for all write operations (POST/PUT/DELETE) and query operations marked with the `operLog` tag. The host only provides managed global middleware registration joints and unified event distribution, and does not retain fixed operation log business middleware. When `monitor-operlog` is installed and enabled, the plugin middleware participates in the request chain and persists the logs to the `plugin_monitor_operlog` table; when the plugin is unavailable, the host core request link MUST bypass the collection logic and continue normal execution.

#### Scenario: Operation log plugin is enabled
- **WHEN** The user initiates an audited request and `monitor-operlog` is installed and enabled
- **THEN** `monitor-operlog` wraps matching requests through the host's global HTTP middleware registrar
- **AND** The host emits unified audit events
- **AND** `monitor-operlog` writes a corresponding operation log record

#### Scenario: Operation log plugin is missing or disabled
- **WHEN** The user initiates an audited request but `monitor-operlog` is not installed, not enabled, or fails to initialize
- **THEN** Host bypass plugin self-registered audit middleware logic
- **AND** The host still completes the original business request normally
- **AND** The host does not return an error due to lack of specific operation log implementation.

#### Scenario: Downstream middleware ends the request early
- **WHEN** The global audit middleware of `monitor-operlog` has wrapped a request, and the subsequent middleware or processor ends the current request early after writing a response.
- **THEN** The audit middleware can still read the current response snapshot and emit matching audit events after `Next` returns
- **AND** Ending the request early will not cause the operation log to be omitted.

## ADDED Requirements

### Requirement: Operation log type uses semantic string constants

The system SHALL use string constants with business semantics to express operation log types, instead of propagating position-sensitive integer encodings such as `1~6` in the host, plugins, interfaces and storage layers.

#### Scenario: Write semantic types when audit events are logged into the database
- **WHEN** The host emits an operation log audit event
- **THEN** `monitor-operlog` writes `oper_type` using strongly typed constants
- **AND** The persistent value of `oper_type` is one of `create`, `update`, `delete`, `export`, `import`, `other`

#### Scenario: Operation log interface returns semantic type
- **WHEN** Administrator queries or exports operation logs
- **THEN** The `operType` field in the interface returns a semantic string value
- **AND** The front end continues to render the corresponding localized labels through the `sys_oper_type` dictionary

### Requirement: The operation log management interface is delivered by the source plugin

The system SHALL deliver operation log query, details, export, cleaning and page capabilities as the `monitor-operlog` source plugin.

#### Scenario: Expose the management entrance when the plugin is enabled
- **WHEN** `monitor-operlog` is installed and enabled
- **THEN** The host exposes operation log query, details, export, cleanup interface and frontend page
- **AND** The plugin menu is mounted to the host `system monitoring` directory, and the top-level `parent_key` is `monitor`

#### Scenario: Hide the management entrance when the plugin is missing
- **WHEN** `monitor-operlog` is not installed or not enabled
- **THEN** The host does not display the operation log menu and page entry
- **AND** Ordinary service request links continue to operate normally
