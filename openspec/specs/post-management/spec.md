# Post Management

## Purpose

Define the position management query, maintenance, department association and option reading behaviors provided by the `org-center` source plugin to ensure that position data can be stably coordinated with the organizational structure and user management capabilities.
## Requirements
### Requirement: Job list query
The system SHALL provides a paging list query interface for positions and supports filtering by department.

#### Scenario: Check the job list
- **WHEN** Call `GET /api/v1/post` and pass in the paging parameters `pageNum` and `pageSize`
- **THEN** returns the position list and total number in the format of `{list: [...], total: number}`

#### Scenario: Filter jobs by department
- **WHEN** Pass in the `deptId` parameter when querying
- **THEN** Return only positions belonging to this department

#### Scenario: Job list supports conditional filtering
- **WHEN** Pass in the filter parameters `code` (position code), `name` (position name) or `status` (status) when querying
- **THEN** `code` and `name` use fuzzy matching (LIKE)
- **THEN** `status` uses exact match

#### Scenario: Exclude deleted records from the job list
- **WHEN** Query job list
- **THEN** Soft deleted records are not included in the results

### Requirement: Create a position
The system SHALL provides an interface for creating positions.

#### Scenario: Position created successfully
- **WHEN** Call `POST /api/v1/post` and submit fields such as deptId, code, name, sort etc.
- **THEN** The system creates the position and returns success

#### Scenario: Duplicate position code
- **WHEN** Submit the existing code value when creating a position
- **THEN** The system returns an error message, indicating that the position code already exists

#### Scenario: Required field verification
- **WHEN** Missing deptId, code or name when creating position
- **THEN** The system returns parameter verification error

### Requirement: Update position
The system SHALL provides an interface for updating position information.

#### Scenario: Position updated successfully
- **WHEN** Call `PUT /api/v1/post/{id}` and submit the fields to be updated
- **THEN** The system updates the corresponding position information and returns success

#### Scenario: Update non-existent positions
- **WHEN** Update a non-existent position ID
- **THEN** The system returns an error message

### Requirement: Delete position
The system SHALL provides an interface for deleting positions and supports batch deletion.

#### Scenario: Delete a single position
- **WHEN** calls `DELETE /api/v1/post/{id}`
- **THEN** posts are soft deleted

#### Scenario: Delete positions in batches
- **WHEN** calls `DELETE /api/v1/post/{ids}`, ids is multiple IDs separated by commas
- **THEN** All specified positions are soft deleted

#### Scenario: Posts with associated users cannot be deleted
- **WHEN** Delete a post with associated user in `plugin_org_center_user_post`
- **THEN** The system returns an error message, prompting that there is a user under this position and the user MUST be removed first.

### Requirement: View job details
The system SHALL provides a job details query interface.

#### Scenario: Check job details
- **WHEN** calls `GET /api/v1/post/{id}`
- **THEN** Returns complete information for this position

### Requirement: Export position
The system SHALL provides the function of exporting the position list to an Excel file.

#### Scenario: Export jobs
- **WHEN** Call `GET /api/v1/post/export` and pass in filter parameters
- **THEN** returns Excel file stream
- **THEN** The exported fields include: position code, position name, sorting, status, remarks, creation time

### Requirement: Position department tree interface
The system SHALL provides a department tree interface for left-side filtering of position management, including the "unassigned department" virtual node.

#### Scenario: Get the job department tree
- **WHEN** calls `GET /api/v1/post/dept-tree`
- **THEN** Returns department tree structure data

#### Scenario: Unassigned department virtual node
- **WHEN** Department tree returns data
- **THEN** contains an "Unassigned Department" virtual node with id -1

#### Scenario: Filter jobs by unassigned department
- **WHEN** Pass in `deptId=-1` when querying the job list
- **THEN** Returns all positions with dept_id 0 (positions without assigned departments)

### Requirement: Get job options by department
The system SHALL provides an interface for obtaining position options by department for users to edit forms.

#### Scenario: Get the position options under the department
- **WHEN** Call `GET /api/v1/post/option-select` and pass in the `deptId` parameter
- **THEN** Returns a list of all normal positions in the department, including id and name

#### Scenario: There are no positions under the department
- **WHEN** There are no positions under the queried department.
- **THEN** returns an empty list

### Requirement: Position data table design
The The system SHALL provides the `plugin_org_center_post` table and the `plugin_org_center_user_post` related table.

#### Scenario: plugin_org_center_post table structure
- **WHEN** View `plugin_org_center_post` table structure
- **THEN** table contains: id, dept_id (INTEGER, referencing `plugin_org_center_dept`.id), code (VARCHAR, UNIQUE), name, sort, status, remark, created_at, updated_at, deleted_at

#### Scenario: plugin_org_center_user_post association table structure
- **WHEN** View `plugin_org_center_user_post` table structure
- **THEN** table contains: user_id (INTEGER), post_id (INTEGER), joint primary key
- **THEN** user_id refers to sys_user.id, post_id refers to `plugin_org_center_post`.id

### Requirement: Position management frontend left tree and right table layout
The system SHALL adopts the layout of department tree on the left + position list on the right on the position management page.

#### Scenario: layout structure
- **WHEN** Open the position management page
- **THEN** The DeptTree component (260px width) is displayed on the left and the position list (flex-1) is displayed on the right

#### Scenario: Department screening linkage
- **WHEN** Select a department on the left
- **THEN** The job list on the right is automatically filtered by the department
- **WHEN** Cancel department selection
- **THEN** All positions are displayed on the right

#### Scenario: table column definition
- **WHEN** View the job list table
- **THEN** Display the following columns: checkbox, position code, position name, sort, status (DictTag rendering), creation time, operation

#### Scenario: Toolbar operations
- **WHEN** View toolbar
- **THEN** Display: new button (primary), batch delete button (danger, enabled after checked), export button

#### Scenario: Row action button
- **WHEN** View the action column of each row
- **THEN** Displays two buttons: Edit (ghost), Delete (ghost, red, Popconfirm to confirm)

### Requirement: Position editing drawer
The The system SHALL provides a 600px width Drawer for adding and editing positions.

#### Scenario: Position form fields
- **WHEN** Open the position editing Drawer
- **THEN** Form fields include: department (TreeSelect, required, display full path), position name (required), position code (required), sort (required, default 0), status (RadioGroup button style, default normal), remarks (Textarea, full width)
- **THEN** The form uses a 2-column grid layout

### Requirement: Position initialization data
The system SHALL provides basic job initialization data.

#### Scenario: Initialize position data
- **WHEN** Execute v0.2.0 database migration script
- **THEN** Create the following job data:
  - General Manager (code: CEO, dept: Lina Technology, sort: 1)
  - Technical Director (code: CTO, dept: R&D department, sort: 2)
  - Project manager (code: PM, dept: R&D department, sort: 3)
  - Development engineer (code: DEV, dept: R&D department, sort: 4)
  - Test engineer (code: QA, dept: test department, sort: 5)

### Requirement: Position management is delivered by the organization source plugin

The The system SHALL deliver position management capabilities as an `org-center` source plugin, rather than continuing as the host's default built-in module.

#### Scenario: Provides position management when the organization plugin is enabled
- **WHEN** `org-center` is installed and enabled
- **THEN** The host exposes position management API, page and menu
- **AND** The position management menu is mounted to the host `Organization Management` directory, and the top-level `parent_key` is `org`

#### Scenario: Hide the position management entrance when the organization plugin is missing
- **WHEN** `org-center` is not installed or not enabled
- **THEN** The host does not display the position management menu and page entry
- **AND** Hosting capabilities such as user management will continue to be available according to organization downgrade rules.

