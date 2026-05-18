## ADDED Requirements

### Requirement: Non-core management modules are delivered as source plugins.

The system SHALL deliver non-core modules in organization management, content management, and system monitoring as source plugins that developers can install and enable on demand.

#### Scenario: Planning Organization and Content Modules
- **WHEN** The host delivers default background capabilities
- **THEN** Department management and position management are provided by the `org-center` source plugin
- **AND** Notification announcements are provided by the `content-notice` source plugin

#### Scenario: Planning system monitoring module
- **WHEN** Host delivery system monitoring related capabilities
- **THEN** Online users, service monitoring, operation logs and login logs are provided by independent source plugins.
- **AND** Their plugin IDs are `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog`

### Requirement: The monitoring plugin MUST support independent installation and startup and shutdown.

The system SHALL treat online users, service monitoring, operation logs and login logs as four independent source plugins instead of a single monitoring plugin suite.

#### Scenario: Only some monitoring plugins are installed
- **WHEN** The administrator only installs or enables some monitoring plugins
- **THEN** The host only displays the monitoring menu corresponding to these installed and enabled plugins.
- **AND** Uninstalled monitoring plugins will not block other monitoring plugins from running.

#### Scenario: Disable individual monitoring plugins
- **WHEN** Administrator disables any of `monitor-online`, `monitor-server`, `monitor-operlog` or `monitor-loginlog`
- **THEN** The host only hides the function entrance corresponding to the plugin
- **AND** Other monitoring plugins and host core links continue to operate normally

### Requirement: The host MUST be gracefully downgraded when the plugin is missing.

The system SHALL ensure that host principal functions continue to be available when the source plugin is missing, not installed, or not enabled.

#### Scenario: Access user management when organization plugin is missing
- **WHEN** `org-center` is not installed or not enabled
- **THEN** The user management page and interface still work normally
- **AND** Fields, filter items, tree selectors and form items related to departments and positions are safely hidden or ignored

#### Scenario: The host continues to process requests when the log plugin is missing
- **WHEN** `monitor-operlog` or `monitor-loginlog` is not installed or enabled
- **THEN** The host core request link still executes normally
- **AND** Capabilities related to corresponding log persistence enter the controlled downgrade process
- **AND** No authentication, authentication or ordinary business requests will fail due to missing log plugin

### Requirement: The online user plugin MUST not carry the authentication main link

The system SHALL ensure that `monitor-online` only carries online user management capabilities and will not take over the host authentication main link.

#### Scenario: Online user plugin missing
- **WHEN** `monitor-online` is not installed or not enabled
- **THEN** The host still performs login, logout, protected interface authentication and session timeout cleanup normally
- **AND** The host continues to use its own session truth source to maintain `last_active_time` and timeout determination

#### Scenario: Online user plugin execution forced offline
- **WHEN** `monitor-online` has been installed and enforced offline management
- **THEN** The plugin uses the session management capability provided by the host to invalidate the specified session.
- **AND** Plugin does not have JWT validation, session hit refresh or timeout clean source

### Requirement: The log plugin accepts non-core logs and logs into the database through host events.

The system SHALL decouple the logging responsibilities of login logs and operation logs into "host-emitted events + plugin on-demand subscription persistence".

#### Scenario: Login log plugin is enabled
- **WHEN** The user has successfully logged in, failed to log in, or successfully logged out.
- **THEN** The host launches the unified login event first
- **AND** `monitor-loginlog` completes the logout and subsequent query management after subscribing to the event

#### Scenario: Operation log plugin is enabled
- **WHEN** User triggered write operation or audited query with `operLog` tag
- **THEN** The host launches unified audit events first
- **AND** `monitor-operlog` completes the logout and subsequent query management after subscribing to the event

#### Scenario: Logging plugin is not enabled
- **WHEN** `monitor-loginlog` or `monitor-operlog` is not installed, not enabled, or failed to initialize
- **THEN** The host continues processing the original authentication or request process
- **AND** The host does not return an error due to lack of specific log persistence implementation

### Requirement: The backend database access of the source plugin is closed loop within the plugin.

The system SHALL require official source plugins to maintain independent GoFrame ORM code generation configurations in their respective `backend/` directories, and complete database access through the plugin's local `dao/do/entity` to avoid re-reliance on the host `dao/model` package or long-term retention of scattered `g.DB().Model(...)` direct connection implementations.

#### Scenario: Maintain independent codegen configuration for plugin backend
- **WHEN** team creates or maintains official source plugin backends
- **THEN** The plugin `backend/` directory contains `hack/config.yaml`
- **AND** Developers can directly execute `gf gen dao` in the `backend/` directory of the plugin
- **AND** The generated results fall into `internal/dao`, `internal/model/do` and `internal/model/entity` local to the plugin

#### Scenario: The plugin service accesses the plugin's own table or shared reading table
- **WHEN** `backend/internal/service/` of `org-center`, `content-notice`, `monitor-loginlog`, `monitor-operlog`, `monitor-server` or `plugin-demo-source` to access the database
- **THEN** plugin service uses `dao/do/entity` generated locally by the plugin
- **AND** Access to shared read tables such as `sys_user` and `sys_dict_data` is also completed through the plugin's local generation of artifacts
- **AND** The plugin backend does not directly depend on the host `dao/model` package
- **AND** The host no longer retains the ORM artifacts of these plugin business tables in parallel

#### Scenario: The current version does not directly access the source plugin of the database.
- **WHEN** The current version of an official source plugin only completes business processing through the host's stable capabilities.
- **THEN** The plugin still retains the local `backend/hack/config.yaml`
- **AND** If new database access is added in the future, the plugin's local `gf gen dao` and `dao/do/entity` structures will continue to be used.

### Requirement: Source plugins have independent storage lifecycles and namespaces

The system SHALL establish clear data table naming and loading boundaries for official source plugins, so that plugin own storage and host core storage can be stably distinguished in the same database.

#### Scenario: Plugin installation business table
- **WHEN** Official source plugin creates its own business table
- **THEN** Created by installing SQL under the plugin `manifest/sql/`
- **AND** Host `manifest/sql/` does not create these tables
- **AND** Host `manifest/sql/mock-data/` does not write to these tables

#### Scenario: Naming of the planning plugin's own business tables
- **WHEN** team designs new business physical table for official source plugin
- **THEN** table name uses `plugin_<plugin_id_snake_case>` scope naming
- **AND** Single table plugins preferentially use `plugin_<plugin_id_snake_case>` as the complete table name.
- **AND** Multi-table plugin adds business suffix as needed on this basis (such as `plugin_org_center_dept`)
- **AND** Host core table prefix `sys_` is no longer used

#### Scenario: Uninstall plugin and clean data
- **WHEN** Administrator uninstalls the plugin and chooses to clean its business data
- **THEN** The plugin `manifest/sql/uninstall/` is responsible for deleting the plugin scope business table
- **AND** The host does not additionally maintain the cleaning SQL of the plugin business table
