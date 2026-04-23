# Oper Log

## Purpose

Define the automatic recording, querying, deleting and exporting behavior of the operation logs undertaken by the `monitor-operlog` source plugin to ensure that the system's key write operations and specified read operations have traceable and auditable operation traceability.
## Requirements
### Requirement: Automatic recording of operation logs
The system SHALL uses the global audit middleware declared on the host unified HTTP registration portal through the `monitor-operlog` source plugin to automatically emit unified audit events for all write operations (POST/PUT/DELETE) and query operations marked with the `operLog` tag. The host only provides managed global middleware registration joints and unified event distribution, and does not retain fixed operation log business middleware. When `monitor-operlog` is installed and enabled, the plugin middleware participates in the request chain and persists the logs to the `plugin_monitor_operlog` table; when the plugin is unavailable, the host core request link MUST bypass the collection logic and continue normal execution.

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

### Requirement: Operational logging content
The system SHALL record the following operation information: module name (title, `tags` tag from `g.Meta'), operation name (oper_summary, `summary` tag from `g.Meta'), operation type (oper_type), request method (request_method), request URL (oper_url), operator username (oper_name), operation IP (oper_ip), request parameters (oper_param), response result (json_result), operation status (status), error information (error_msg), cost_time (cost_time), operation time (oper_time).

#### Scenario: Document full operation information
- **WHEN** One write operation completed (successful or unsuccessful)
- **THEN** The log record contains all the above fields, `status` is 0 (success) or 1 (failure), `cost_time` is the request processing time (milliseconds)

#### Scenario: Request parameter length truncation
- **WHEN** Request parameter JSON length exceeds 2000 characters
- **THEN** truncate to 2000 characters and append `... (truncated)`

#### Scenario: Response result length truncation
- **WHEN** Response result JSON length exceeds 2000 characters
- **THEN** truncate to 2000 characters and append `... (truncated)`

#### Scenario: Password Field Desensitization
- **WHEN** The request parameter contains a `password` or `Password` field
- **THEN** Replace the field value with `* * *`

#### Scenario: Operation type uses semantic string constants
- **WHEN** System records, queries, or exports action logs
- **THEN** `oper_type` uses semantic strings like `create`, `update`, `delete`, `export`, `import`, `other`
- **AND** host and plugin code multiplex these values through strongly typed constants instead of scattered hardcoded or `1 ~ 6` integer numbers

### Requirement: Operation log list query
The system SHALL provides the operation log paging query interface `get/api/v1/operlog`, which supports filtering by operation module, operator, operation type, status, and time range.

#### Scenario: Paging Query Action Log
- **WHEN** Admin request `get/api/v1/operlog? pageNum = 1 & pageSize = 10`
- **THEN** Returns a paginated list of operation logs, arranged in reverse order by operation time

#### Scenario: Filter by criteria
- **WHEN** Admin requests a query with filters (e.g. `title = User&operName = admin&operType = create&status = 0 & beginTime = 2026-01-01 & endTime = 2026-03-15`)
- **THEN** returns log records that meet all conditions

### Requirement: View operation log details
The system SHALL provides the operation log detail interface `get/api/v1/operlog/{id}` and returns the complete log information.

#### Scenario: View log details
- **WHEN** admin request `get/api/v1/operlog/1`
- **THEN** returns all field information for the log, including full request parameters and response results

### Requirement: Operation log cleanup by time range
The system SHALL provides the operation log cleaning interface `delete/api/v1/operlog/clean`, which supports hard deletion of log records by time range.

#### Scenario: Clean up logs by time range
- **WHEN** Admin request `delete/api/v1/operlog/clean? beginTime = 2026-01-01 & endTime = 2026-01-31`
- **THEN** Hard delete all operation log records within the time range, returning the number of deleted records

#### Scenario: Clean up all logs
- **WHEN** Admin request `delete/api/v1/operlog/clean` (without time parameter)
- **THEN** hard delete all operation log records

### Requirement: Bulk deletion of operation logs
The system SHALL provides the operation log bulk deletion interface `delete/api/v1/operlog/{ids}`, which supports hard deletion of log records by ID list.

#### Scenario: Bulk delete by ID
- **WHEN** admin request `delete/api/v1/operlog/1,2,3`
- **THEN** Hard delete the operation log record with the specified ID, and return the number of deleted records

#### Scenario: frontend bulk delete operation
- **WHEN** Admin checks one or more action logs and clicks the "Delete" button
- **THEN** pops up a confirmation dialog box showing the selected number, confirming and performing batch deletion

### Requirement: Operation log export
The system SHALL provides the operation log export interface `get/api/v1/operlog/export`, which is exported to the xlsx format according to the current filter conditions.

#### Scenario: Export by filter
- **WHEN** Admin request `get/api/v1/operlog/export? title = User&status = 0`
- **THEN** returns an xlsx file with all eligible log records, including all fields

#### Scenario: Export all
- **WHEN** Admin request `get/api/v1/operlog/export` (without filters)
- **THEN** returns an xlsx file with all action logs

### Requirement: Operation log frontend page
The system SHALL provides the operation log management page under the frontend system monitoring menu through the `monitor-operlog` source plugin.

#### Scenario: Operation Log List Page
- **WHEN** Admin access action log page
- **THEN** shows the action log table, including the filter area (module name, operator, operation type, status, time range), table columns (module name, operation name, operator, operation IP, operation status, operation date, operation time), toolbar (empty, export, delete buttons), line actions (view details button).The delete button is grayed out when no record is checked.

#### Scenario: View details drawer
- **WHEN** Admin clicks "Details" button on a log
- **THEN** Open the right drawer to show the full information of the log, including the formatted request parameters and the response result JSON

#### Scenario: Cleanup actions
- **WHEN** Admin clicks "Clean" button
- **THEN** pops up a confirmation dialog box with a time range selector, confirmation and cleanup

#### Scenario: Export Actions
- **WHEN** Admin clicks "Export" button
- **THEN** Export the action log as an xlsx file with the current filters and download it

### Requirement: Operation log type uses semantic string constants

The system SHALL uses string constants with business semantics to express operation log types, instead of propagating position-sensitive integer encodings such as `1~6` in the host, plugins, interfaces and storage layers.

#### Scenario: Write semantic types when audit events are logged into the database
- **WHEN** The host emits an operation log audit event
- **THEN** `monitor-operlog` writes `oper_type` using strongly typed constants
- **AND** The persistent value of `oper_type` is one of `create`, `update`, `delete`, `export`, `import`, `other`

#### Scenario: Operation log interface returns semantic type
- **WHEN** Administrator queries or exports operation logs
- **THEN** The `operType` field in the interface returns a semantic string value
- **AND** The front end continues to render the corresponding localized labels through the `sys_oper_type` dictionary

### Requirement: The operation log management interface is delivered by the source plugin

The The system SHALL deliver operation log query, details, export, cleaning and page capabilities as the `monitor-operlog` source plugin.

#### Scenario: Expose the management entrance when the plugin is enabled
- **WHEN** `monitor-operlog` is installed and enabled
- **THEN** The host exposes operation log query, details, export, cleanup interface and frontend page
- **AND** The plugin menu is mounted to the host `system monitoring` directory, and the top-level `parent_key` is `monitor`

#### Scenario: Hide the management entrance when the plugin is missing
- **WHEN** `monitor-operlog` is not installed or not enabled
- **THEN** The host does not display the operation log menu and page entry
- **AND** Ordinary service request links continue to operate normally

