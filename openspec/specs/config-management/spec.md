# Config Management

## Purpose

Define the query, maintenance, import and export, and key reading behaviors of the system parameter setting module to ensure that the runtime configuration can be uniformly managed by the host, stably consumed by the business module, and support subsequent expansion.
## Requirements
### Requirement: Config list query with pagination and filters
The system SHALL provide a paginated list query for system config parameters. The list SHALL support filtering by config name (fuzzy match), config key (fuzzy match), and creation time range. The list SHALL return config records sorted by ID descending by default.

#### Scenario: Query config list without filters
- **WHEN** user requests GET `/config` with pageNum=1 and pageSize=10
- **THEN** system returns the first 10 config records sorted by ID descending, along with total count

#### Scenario: Query config list with name filter
- **WHEN** user requests GET `/config` with name="mail"
- **THEN** system returns only config records whose name contains "mail"

#### Scenario: Query config list with key filter
- **WHEN** user requests GET `/config` with key="smtp"
- **THEN** system returns only config records whose key contains "smtp"

#### Scenario: Query config list with time range filter
- **WHEN** user requests GET `/config` with beginTime="2025-01-01" and endTime="2025-12-31"
- **THEN** system returns only config records created within the specified time range

### Requirement: Get config detail by ID
The system SHALL provide an endpoint to retrieve a single config record by its ID.

#### Scenario: Get existing config detail
- **WHEN** user requests GET `/config/{id}` with a valid config ID
- **THEN** system returns the complete config record including id, name, key, value, remark, created_at, updated_at

#### Scenario: Get non-existent config detail
- **WHEN** user requests GET `/config/{id}` with a non-existent ID
- **THEN** system returns an error indicating the config does not exist

### Requirement: Create config
The system SHALL allow creating a new config parameter with name, key, value, and optional remark. The key field MUST be unique across all non-deleted config records.

#### Scenario: Create config with valid data
- **WHEN** user requests POST `/config` with name="Mail service address", key="sys.mail.host", value="smtp.example.com"
- **THEN** system creates the config record and returns the new record ID

#### Scenario: Create config with duplicate key
- **WHEN** user requests POST `/config` with a key that already exists
- **THEN** system returns an error indicating the key already exists

### Requirement: Update config
The system SHALL allow updating an existing config record's name, key, value, and remark fields. The updated key MUST remain unique.

#### Scenario: Update config with valid data
- **WHEN** user requests PUT `/config/{id}` with updated value="new-value"
- **THEN** system updates the record and sets updated_at to current time

#### Scenario: Update config with duplicate key
- **WHEN** user requests PUT `/config/{id}` with a key that belongs to another config
- **THEN** system returns an error indicating the key already exists

### Requirement: Delete config
The system SHALL support soft-deleting a config record by ID. The system SHALL also support batch deletion of multiple records.

#### Scenario: Delete a single config
- **WHEN** user requests DELETE `/config/{id}` with a valid ID
- **THEN** system soft-deletes the record by setting deleted_at timestamp

#### Scenario: Delete a non-existent config
- **WHEN** user requests DELETE `/config/{id}` with a non-existent ID
- **THEN** system returns an error indicating the config does not exist

### Requirement: Get config by key
The system SHALL provide an endpoint to retrieve a config value by its key name, for internal use by other modules.

#### Scenario: Get config by existing key
- **WHEN** user requests GET `/config/key/{key}` with an existing key
- **THEN** system returns the config record matching the key

#### Scenario: Get config by non-existent key
- **WHEN** user requests GET `/config/key/{key}` with a non-existent key
- **THEN** system returns an error indicating the config key does not exist

### Requirement: Export configs to Excel
The The system SHALL support exporting config records to an Excel file. The export SHALL apply the same filters as the list query (name, key, time range). Export file name SHALL be "parameter settings export.xlsx".

#### Scenario: Export all configs
- **WHEN** user requests GET `/config/export` without filters
- **THEN** system generates and returns an Excel file containing all non-deleted config records

#### Scenario: Export filtered configs
- **WHEN** user requests GET `/config/export` with name="mail"
- **THEN** system generates an Excel file containing only config records whose name matches the filter

#### Scenario: Export selected configs
- **WHEN** user selects N records and requests export with ids parameter
- **THEN** system generates an Excel file containing only the selected config records

### Requirement: Config management page UI
The frontend SHALL provide a config management page under system management menu. The page SHALL include a search bar, toolbar, data table, and create/edit modal.

#### Scenario: Display config list page
- **WHEN** user navigates to the config management page
- **THEN** page displays a search bar (name, key, time range), toolbar (export, batch delete, add), and a VXE-Grid table with columns: checkbox, name, key, value, remark, updated_at, actions (edit/delete)

#### Scenario: Create config via modal
- **WHEN** user clicks "New" button and fills in the form (name, key, value, remark)
- **THEN** system creates the config and refreshes the table

#### Scenario: Edit config via modal
- **WHEN** user clicks "edit" on a row
- **THEN** system opens a pre-filled modal, user edits and saves, table refreshes

#### Scenario: Delete config via popconfirm
- **WHEN** user clicks "delete" on a row and confirms
- **THEN** system deletes the config and refreshes the table

#### Scenario: Batch delete configs
- **WHEN** user selects multiple rows and clicks "bulk delete" and confirms
- **THEN** system deletes all selected configs and refreshes the table

#### Scenario: Export configs with confirmation
- **WHEN** user clicks "Export" button
- **THEN** system shows export confirmation modal (selected N records or all records)
- **THEN** system downloads an Excel file named "parameter settings export.xlsx" with current filter conditions applied

### Requirement: Config menu and permissions
The The system SHALL include a "Parameter Settings" menu item under system management. Access to config operations SHALL be controlled by permissions.

#### Scenario: Menu visibility
- **WHEN** user has system:config:list permission
- **THEN** "Parameter Settings" menu item is visible in the system management menu

#### Scenario: Permission-controlled operations
- **WHEN** user lacks system:config:add permission
- **THEN** the "Add" button is hidden on the config page

### Requirement: Import configs from Excel
The system SHALL support importing config records from an Excel file. The system SHALL provide a template download endpoint and validate imported data before persisting.

#### Scenario: Download import template
- **WHEN** user requests GET `/config/import-template`
- **THEN** system returns an Excel template with example data showing required columns: parameter name, parameter key name, parameter key value, remarks

#### Scenario: Import with valid data
- **WHEN** user uploads a valid Excel file to POST `/config/import`
- **THEN** system validates all rows, creates records, and returns success count

#### Scenario: Import with validation errors
- **WHEN** user uploads an Excel file with invalid data (missing required fields, duplicate keys)
- **THEN** system rejects the entire import and returns error details with row numbers and reasons

#### Scenario: Import with overwrite mode
- **WHEN** user uploads an Excel file with `updateSupport=true` and the file contains keys that already exist
- **THEN** system updates existing records with the imported values

#### Scenario: Import with ignore mode
- **WHEN** user uploads an Excel file with `updateSupport=false` (default) and the file contains keys that already exist
- **THEN** system skips existing records and only creates new records

#### Scenario: Import modal UI
- **WHEN** user clicks "Import" button on the config management page
- **THEN** system displays a modal with template download link, drag-and-drop upload area, file type hint (xlsx/xls), and overwrite/ignore mode switch

### Requirement: Config table design
The system SHALL provide sys_config table for storing system configuration parameters.

#### Scenario: sys_config table structure
- **WHEN** viewing sys_config table structure
- **THEN** table contains: id, name, key (UNIQUE), value, remark, created_at, updated_at, deleted_at

### Requirement: Built-in Runtime Parameter Metadata
The system SHALL provide built-in metadata records for host-consumed runtime parameters so administrators can manage effective host behavior directly from config management.

#### Scenario: Initialize built-in runtime parameters
- **WHEN** an administrator executes the host initialization SQL
- **THEN** `sys_config` contains `sys.jwt.expire`, `sys.session.timeout`, `sys.upload.maxSize`, and `sys.login.blackIPList`
- **AND** each record includes a readable name, a default value, and a format description

### Requirement: Built-in Runtime Parameter Protection
The system SHALL validate built-in runtime parameter values and SHALL protect stable host-owned keys from rename or deletion.

#### Scenario: Reject invalid built-in runtime parameter values
- **WHEN** a user creates, updates, or imports one of the built-in runtime parameters with an invalid value format
- **THEN** the system rejects the change and returns a validation error

#### Scenario: Reject rename or deletion of built-in runtime parameter keys
- **WHEN** a user attempts to rename or delete a built-in runtime parameter key already consumed by the host
- **THEN** the system rejects the operation and keeps the parameter record intact

### Requirement: Upload Size Parameter Must Drive Host Runtime Behavior
The system SHALL ensure that `sys.upload.maxSize` is enforced by the host upload chain instead of existing only as editable metadata.

#### Scenario: Upload size change takes effect immediately
- **WHEN** an administrator updates `sys.upload.maxSize` to `1`
- **THEN** subsequent upload requests are validated against a 1 MB limit
- **AND** uploads above the configured limit are rejected

### Requirement: Multi-Instance Runtime Parameter Cache Synchronization
The system SHALL use a local snapshot plus shared revision strategy for protected runtime parameter reads so hot paths do not query `sys_config` on every request.

#### Scenario: Runtime reads hit the local snapshot
- **WHEN** a node repeatedly reads protected runtime parameters while the shared revision has not changed
- **THEN** the node reuses its local in-memory snapshot
- **AND** it does not need to query `sys_config` on every read

#### Scenario: Parameter changes propagate to other instances
- **WHEN** a protected runtime parameter changes on one instance
- **THEN** the writing instance clears its local snapshot and bumps the shared revision
- **AND** other instances rebuild their local snapshots during the next synchronization cycle

### Requirement: Public Frontend Setting Metadata
The system SHALL provide built-in metadata for safe public frontend settings used by branding, login-page presentation, and workspace theme bootstrap.

#### Scenario: Initialize public frontend settings
- **WHEN** an administrator executes the host initialization SQL
- **THEN** `sys_config` contains the built-in public frontend setting keys used by the login page and workspace bootstrap
- **AND** each record includes a readable name, a default value, and a format description

### Requirement: Public Frontend Setting Protection
The system SHALL validate built-in public frontend setting values and SHALL protect their stable keys from rename or deletion.

#### Scenario: Reject invalid public frontend setting values
- **WHEN** a user creates, updates, or imports a built-in public frontend setting with an invalid enum, boolean, or required-text value
- **THEN** the system rejects the change and returns a validation error

#### Scenario: Reject rename or deletion of public frontend setting keys
- **WHEN** a user attempts to rename or delete a built-in public frontend setting key already consumed by the login page or admin workspace
- **THEN** the system rejects the operation and keeps the parameter record intact

### Requirement: Login and Workspace Consume Public Frontend Settings
The system SHALL expose a safe whitelist endpoint for public frontend settings and SHALL let the login page and admin workspace consume that contract.

#### Scenario: Public frontend settings are available before login
- **WHEN** a browser loads the login page without an authenticated session
- **THEN** the frontend can read the whitelisted branding and presentation settings through the public endpoint
- **AND** the endpoint does not expose arbitrary `sys_config` keys

#### Scenario: Updated branding is reflected after refresh
- **WHEN** an administrator updates public frontend settings and a user refreshes the login page or workspace
- **THEN** the refreshed UI shows the updated branding, copy, and theme defaults

### Requirement: Log TraceID output switch is only controlled by static configuration files

The The system SHALL turns off TraceID output in the log by default, and only allows the control of whether to output TraceID through the `logger.extensions.traceIDEnabled` static switch in `config.yaml`.

#### Scenario: The log does not output TraceID by default when it is not explicitly enabled.
- **WHEN** `logger.extensions.traceIDEnabled` is not declared in the configuration file
- **THEN** Host logs and HTTP Server logs do not output the TraceID field by default

#### Scenario: Configuration file explicitly enables TraceID output
- **WHEN** `logger.extensions.traceIDEnabled` is explicitly set to `true` in the configuration file
- **THEN** Host log and HTTP Server log output TraceID field

#### Scenario: TraceID system parameters are no longer exposed when initializing built-in parameters.
- **WHEN** Administrator executes host initialization SQL
- **THEN** `sys_config` does not contain `sys.logger.traceID.enabled` records

### Requirement: Built-in metadata for the login-panel position parameter

The system MUST provide a protected built-in public-frontend parameter named `sys.auth.loginPanelLayout` to maintain the default login-panel layout.

#### Scenario: Initialize the login-panel position parameter
- **WHEN** an administrator runs the host initialization SQL
- **THEN** `sys_config` contains a built-in parameter record with key `sys.auth.loginPanelLayout`
- **AND** the default value of that record is `panel-right`
- **AND** the record includes a readable name and value descriptions for `panel-left`, `panel-center`, and `panel-right`

### Requirement: Validate the login-panel position parameter and expose it through the public-frontend config endpoint

The system MUST validate the value domain of `sys.auth.loginPanelLayout` and expose the effective value through the public-frontend config endpoint for unauthenticated pages.

#### Scenario: Reject invalid login-panel position values
- **WHEN** a user creates, updates, or imports `sys.auth.loginPanelLayout` with a value other than `panel-left`, `panel-center`, or `panel-right`
- **THEN** the system rejects the change and returns a parameter-validation error

#### Scenario: Public frontend config returns the login-panel position
- **WHEN** a browser requests `GET /config/public/frontend`
- **THEN** `auth.panelLayout` in the response equals the effective value of `sys.auth.loginPanelLayout`
- **AND** unauthenticated pages can consume that value without reading any other `sys_config` data

### Requirement: Default value and length rules for the login-page description parameter

The system MUST provide a default description value for the protected built-in public-frontend parameter `sys.auth.pageDesc`, and MUST allow a non-empty description of up to 500 characters so the login page can show richer product copy.

#### Scenario: Initialize the login-page description parameter
- **WHEN** an administrator runs the host initialization SQL
- **THEN** `sys_config` contains a built-in parameter record with key `sys.auth.pageDesc`
- **AND** the default value of that record is `Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model`

#### Scenario: Save a login-page description within 500 characters
- **WHEN** an administrator creates, updates, or imports `sys.auth.pageDesc` through system-parameter management and the value length is between 1 and 500 characters
- **THEN** the system accepts and stores the value
- **AND** `auth.pageDesc` returned by the public-frontend config endpoint matches the saved value

#### Scenario: Reject an overlong login-page description
- **WHEN** an administrator creates, updates, or imports `sys.auth.pageDesc` through system-parameter management and the value length exceeds 500 characters
- **THEN** the system rejects the change and returns a parameter-validation error

### Requirement: The default upload size must be unified at 20 MB
The system SHALL set the platform default value of `sys.upload.maxSize` to `20`, and database initialization, config-template defaults, and runtime upload fallbacks SHALL all use that same value unless an administrator explicitly overrides it.

#### Scenario: Host initialization writes the 20 MB default
- **WHEN** an administrator runs the host initialization SQL
- **THEN** the default value of `sys.upload.maxSize` in `sys_config` is `20`
- **AND** the default value read by config management for that built-in parameter is also `20`

#### Scenario: Runtime default remains 20 MB when no override is provided
- **WHEN** the host handles a `multipart` upload request without any administrator override for the upload-size setting
- **THEN** file-upload validation enforces a 20 MB limit
- **AND** the friendly error message triggered by the default limit returns wording equivalent to "file size cannot exceed 20 MB"

### Requirement: All default upload-size sources must stay consistent
The system SHALL keep the database seed value, config-template default, and host static fallback value for `sys.upload.maxSize` consistent so different startup paths do not expose different default upload limits.

#### Scenario: The host starts from the default template
- **WHEN** an operator generates runtime config from the host default `config.template.yaml` and does not change the upload limit separately
- **THEN** the host reads a default upload size of 20 MB
- **AND** that default matches the `sys.upload.maxSize` default written by the host initialization SQL

### Requirement: The config-management component must have a unit-test coverage gate
The system SHALL maintain repeatable unit tests for the `apps/lina-core/internal/service/config` config-management component, and SHALL use package-level coverage verification as a delivery gate before that component is considered ready.

#### Scenario: Package-level coverage meets the delivery bar
- **WHEN** a maintainer runs `go test ./internal/service/config -cover` from `apps/lina-core`
- **THEN** the command succeeds
- **AND** the reported package-level statement coverage is not lower than `80%`

### Requirement: Critical config-management branches must have automated regression protection
The system SHALL add automated unit tests for critical helper logic inside the config-management component, including high-risk branches around defaults and fallbacks, cache or snapshot reuse, and invalid input or error propagation.

#### Scenario: Plugin and public-frontend config helper logic changes
- **WHEN** a change touches plugin dynamic storage paths, protected public-frontend config key checks, or the shared validation entry point
- **THEN** unit tests cover the normal read path
- **AND** cover default-value or compatibility-fallback behavior
- **AND** cover invalid input or empty-value defensive behavior

#### Scenario: Runtime-parameter cache and revision synchronization logic changes
- **WHEN** a change touches runtime-parameter snapshot caches, the revision controller, or shared-KV synchronization logic
- **THEN** unit tests cover cache-hit or local-reuse behavior
- **AND** cover rebuilds after revision changes
- **AND** cover error propagation and defensive behavior for shared-KV read failures, invalid cached values, or equivalent exceptional cases

### Requirement: Config metadata must return localized names and remarks for the current language
The system SHALL return localized config metadata for config list, import/export templates, and protected-setting projections. Config metadata localization MUST use stable config keys as translation anchors and MUST NOT change the actual config key or stored config value.

#### Scenario: Config list returns English metadata
- **WHEN** an administrator queries the config list with `en-US`
- **THEN** the config names and remarks in the response use English localized values
- **AND** `configKey` and `configValue` keep their original governance semantics

#### Scenario: Config management keeps original config values
- **WHEN** an administrator opens the parameter setting edit modal, exports data, or imports filled data with `en-US`, and a config item's `configValue` is Chinese seed copy
- **THEN** the system MUST NOT write the current-language projected value back as the config governance value
- **AND** `configKey` and `configValue` continue to participate in edit, import, export, and audit flows using the database's original values
- **AND** parameter setting detail backfill keeps the database value by default, avoiding editable master data changes caused only by language switching

#### Scenario: Import templates and export headers return localized metadata
- **WHEN** an administrator downloads a parameter setting import template or exports parameter setting data with `en-US`
- **THEN** template instructions, header titles, and metadata-related prompts use English localized copy
- **AND** the `configKey` and `configValue` columns keep their original governance semantics and actual exported content

#### Scenario: Missing config metadata translations fall back to the default language
- **WHEN** a config item lacks a name or remark translation in the current language
- **THEN** the system falls back to default-language metadata or baseline names
- **AND** config read and write capabilities remain unaffected

### Requirement: Public frontend config copy must support i18n projection
The system SHALL let the public frontend config endpoint return localized brand and authentication copy according to the current request language, while keeping non-textual fields such as layout and theme mode stable.

#### Scenario: Login-page public config returns English copy
- **WHEN** the browser requests the public frontend config endpoint with `en-US`
- **THEN** the returned app name, login page title, login page description, and login subtitle are English localized results
- **AND** non-copy fields such as `panelLayout`, `themeMode`, and `layout` keep their original values

#### Scenario: Copy-like public parameters are projected only at consumer endpoints
- **WHEN** `sys.app.name`, `sys.auth.pageTitle`, `sys.auth.pageDesc`, `sys.auth.loginSubtitle`, or `sys.ui.watermark.content` still store default seed values, and the browser requests the public frontend config endpoint with `en-US`
- **THEN** the public frontend config endpoint returns the corresponding English localized display copy
- **AND** the same `configValue` in the parameter setting management API keeps the original database value and does not add display translation fields

#### Scenario: Refreshed workspace shows the latest localized brand copy
- **WHEN** an administrator updates public frontend copy for a locale and refreshes the login page or workspace
- **THEN** the refreshed page shows the new localized brand name and login display copy
- **AND** no page component code changes are required

