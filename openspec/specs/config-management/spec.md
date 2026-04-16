# Config Management

## Purpose

定义系统参数设置模块的查询、维护、导入导出与按键读取行为，确保运行时配置能够被宿主统一管理、被业务模块稳定消费并支持后续扩展。

## Requirements

### Requirement: Config list query with pagination and filters
The system SHALL provide a paginated list query for system config parameters. The list SHALL support filtering by config name (fuzzy match), config key (fuzzy match), and creation time range. The list SHALL return config records sorted by ID descending by default.

#### Scenario: Query config list without filters
- **WHEN** user requests GET `/config` with pageNum=1 and pageSize=10
- **THEN** system returns the first 10 config records sorted by ID descending, along with total count

#### Scenario: Query config list with name filter
- **WHEN** user requests GET `/config` with name="邮件"
- **THEN** system returns only config records whose name contains "邮件"

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
- **WHEN** user requests POST `/config` with name="邮件服务地址", key="sys.mail.host", value="smtp.example.com"
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
The system SHALL support exporting config records to an Excel file. The export SHALL apply the same filters as the list query (name, key, time range). Export file name SHALL be "参数设置导出.xlsx".

#### Scenario: Export all configs
- **WHEN** user requests GET `/config/export` without filters
- **THEN** system generates and returns an Excel file containing all non-deleted config records

#### Scenario: Export filtered configs
- **WHEN** user requests GET `/config/export` with name="邮件"
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
- **WHEN** user clicks "新增" button and fills in the form (name, key, value, remark)
- **THEN** system creates the config and refreshes the table

#### Scenario: Edit config via modal
- **WHEN** user clicks "编辑" on a row
- **THEN** system opens a pre-filled modal, user edits and saves, table refreshes

#### Scenario: Delete config via popconfirm
- **WHEN** user clicks "删除" on a row and confirms
- **THEN** system deletes the config and refreshes the table

#### Scenario: Batch delete configs
- **WHEN** user selects multiple rows and clicks "批量删除" and confirms
- **THEN** system deletes all selected configs and refreshes the table

#### Scenario: Export configs with confirmation
- **WHEN** user clicks "导出" button
- **THEN** system shows export confirmation modal (selected N records or all records)
- **THEN** system downloads an Excel file named "参数设置导出.xlsx" with current filter conditions applied

### Requirement: Config menu and permissions
The system SHALL include a "参数设置" menu item under system management. Access to config operations SHALL be controlled by permissions.

#### Scenario: Menu visibility
- **WHEN** user has system:config:list permission
- **THEN** "参数设置" menu item is visible in the system management menu

#### Scenario: Permission-controlled operations
- **WHEN** user lacks system:config:add permission
- **THEN** the "新增" button is hidden on the config page

### Requirement: Import configs from Excel
The system SHALL support importing config records from an Excel file. The system SHALL provide a template download endpoint and validate imported data before persisting.

#### Scenario: Download import template
- **WHEN** user requests GET `/config/import-template`
- **THEN** system returns an Excel template with example data showing required columns: 参数名称、参数键名、参数键值、备注

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
- **WHEN** user clicks "导入" button on the config management page
- **THEN** system displays a modal with template download link, drag-and-drop upload area, file type hint (xlsx/xls), and overwrite/ignore mode switch

### Requirement: Config table design
The system SHALL provide sys_config table for storing system configuration parameters.

#### Scenario: sys_config table structure
- **WHEN** viewing sys_config table structure
- **THEN** table contains: id, name, key (UNIQUE), value, remark, created_at, updated_at, deleted_at
