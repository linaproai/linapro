# Dept Management

## Purpose

Define the department management tree query, create, update, delete and associate user capabilities provided by the `org-center` source plugin to ensure that the organizational structure can be stably maintained by the organization plugin and support reference by other business modules.
## Requirements
### Requirement: Department list query
System SHALL provides a tree list query interface for departments (without paging).

#### Scenario: Query department list
- **WHEN** calls `GET /api/v1/dept`
- **THEN** returns a flat list of all department data, and the front end builds a tree structure through parentId
- **THEN** Sort by order_num in ascending order

#### Scenario: Department list supports conditional filtering
- **WHEN** Pass in the filter parameter `name` (department name) or `status` (status) when querying
- **THEN** `name` uses fuzzy matching (LIKE)
- **THEN** `status` uses exact match

#### Scenario: Department list excludes deleted records
- **WHEN** Query department list
- **THEN** Soft deleted records are not included in the results

### Requirement: Create department
System SHALL provides an interface for creating departments.

#### Scenario: Department created successfully
- **WHEN** Call `POST /api/v1/dept` and submit parentId, name, orderNum and other fields
- **THEN** The system creates a department, automatically calculates the ancestors field (such as "0,1,2"), and returns success

#### Scenario: Create root department
- **WHEN** parentId is 0 when creating a department
- **THEN** This department is the root department and ancestors is "0"

#### Scenario: Required field verification
- **WHEN** name is missing when creating department
- **THEN** The system returns parameter verification error

### Requirement: Update department
System SHALL provides an interface for updating department information.

#### Scenario: Update department successful
- **WHEN** Call `PUT /api/v1/dept/{id}` and submit the fields to be updated
- **THEN** The system updates the corresponding department information and returns success

#### Scenario: Departments cannot be set as subdepartments of themselves
- **WHEN** Set parentId to its own ID or its own sub-department ID when updating the department
- **THEN** The system returns an error message, prompting that the superior department cannot be itself or its sub-department.

#### Scenario: Update sub-department ancestors synchronously when updating department
- The parentId of the **WHEN** department has changed
- **THEN** The system automatically updates the ancestors field of this department and all sub-departments

### Requirement: Delete department
System SHALL provides an interface for deleting departments.

#### Scenario: Department deleted successfully
- **WHEN** calls `DELETE /api/v1/dept/{id}`
- **THEN** department is soft deleted

#### Scenario: Departments with sub-departments cannot be deleted
- **WHEN** Delete a department with sub-departments
- **THEN** The system returns an error message, prompting that there are sub-departments under this department and the sub-departments MUST be deleted first.

#### Scenario: Departments with associated users cannot be deleted
- **WHEN** Delete a department that has an associated user in `plugin_org_center_user_dept`
- **THEN** The system returns an error message, prompting that there is a user under this department and the user MUST be removed first.

### Requirement: View department details
System SHALL provides department details query interface.

#### Scenario: Query department details
- **WHEN** calls `GET /api/v1/dept/{id}`
- **THEN** returns complete information about the department

### Requirement: Department tree structure interface
System SHALL provides a department tree interface for use with the TreeSelect component.

#### Scenario: Get the complete department tree
- **WHEN** calls `GET /api/v1/dept/tree`
- **THEN** returns tree structure data, each node contains id, label (department name), children

#### Scenario: Get the department tree excluding the specified node
- **WHEN** calls `GET /api/v1/dept/exclude/{id}`
- **THEN** Returns a list of departments excluding this node and all its child nodes
- **THEN** is used to select the superior department when editing a department (to avoid circular references)

### Requirement: Department data table design
The system SHALL provides the `plugin_org_center_dept` table and the `plugin_org_center_user_dept` related table.

#### Scenario: plugin_org_center_dept table structure
- **WHEN** View `plugin_org_center_dept` table structure
- **THEN** table contains: id, parent_id, ancestors, name, order_num, leader (INTEGER, referencing sys_user.id), phone, email, status, remark, created_at, updated_at, deleted_at

#### Scenario: plugin_org_center_user_dept related table structure
- **WHEN** View the `plugin_org_center_user_dept` table structure
- **THEN** table contains: user_id (INTEGER), dept_id (INTEGER), joint primary key
- **THEN** user_id refers to sys_user.id, dept_id refers to `plugin_org_center_dept`.id

### Requirement: Department management frontend tree form
System SHALL uses VXE-Grid's tree mode to display the department level on the department management page.

#### Scenario: Tree display
- **WHEN** Open the department management page
- **THEN** Use VXE-Grid treeConfig (parentField: 'parentId', rowField: 'id', transform: true) to render the tree table
- **THEN** Expand all nodes by default

#### Scenario: Expand/Collapse operation
- **WHEN** Click the "Expand All" button on the toolbar
- **THEN** Expand all tree nodes
- **WHEN** Click the "Collapse All" button on the toolbar
- **THEN** Collapse all tree nodes
- **WHEN** Double click on a row
- **THEN** switches the expanded/collapsed state of the node

#### Scenario: table column definition
- **WHEN** View department list table
- **THEN** Show the following columns: Department Name (Tree Node), Sort, Status (DictTag Rendering), Creation Time, Action

#### Scenario: Row action button
- **WHEN** View the action column of each row
- **THEN** displays three buttons: edit (ghost), add sub-department (ghost, green), delete (ghost, red, Popconfirm confirmation)

### Requirement: Department edit drawer
The system SHALL provides a 600px width Drawer for adding and editing departments.

#### Scenario: Add new department form
- **WHEN** Click the "Add Root Department" or "Add Sub-Department" button
- **THEN** Open Drawer. The form fields include: superior department (TreeSelect, display full path), department name (required), sorting (required, default 0), person in charge (Select, disabled), contact number (regular verification), email (email verification), status (RadioGroup button style)
- **THEN** When adding a new sub-department, the superior department will automatically fill in the current department.

#### Scenario: Edit Department Form
- **WHEN** Click the edit button
- **THEN** Open Drawer and load existing data
- **THEN** The person in charge field becomes available (Select), and the option list is the user under the department (query through `plugin_org_center_user_dept`)
- **THEN** Parent department TreeSelect excludes itself and sub-department nodes

### Requirement: DeptTree reusable components
System SHALL provides reusable DeptTree components for user management and position management.

#### Scenario: DeptTree component function
- **WHEN** Use DeptTree component
- **THEN** Displays department tree structure, supports search, refresh, and radio selection
- **THEN** Bind the selected department ID through v-model:selectDeptId
- **THEN** Expand all nodes by default

#### Scenario: DeptTree Search
- **WHEN** Enter keywords in the search box
- **THEN** Filter to display matching department nodes

#### Scenario: DeptTree Refresh
- **WHEN** Click the refresh button
- **THEN** Reload department tree data and trigger the reload event

### Requirement: Department initialization data
System SHALL provides basic department initialization data.

#### Scenario: Initialize department structure
- **WHEN** Execute v0.2.0 database migration script
- **THEN** Create the following department structure:
  - Lina Technology (root department, id=1)
    - R&D department (id=2)
    - Marketing department (id=3)
    - Test department (id=4)
    - Finance department (id=5)
    - Operation and maintenance department (id=6)

### Requirement: Department management is delivered by the organization source plugin

The The system SHALL deliver department management capabilities as `org-center` source plugins, rather than continuing as the host's default built-in module.

#### Scenario: Provides department management when the organization plugin is enabled
- **WHEN** `org-center` is installed and enabled
- **THEN** Host exposed department management API, pages and menus
- **AND** The department management menu is mounted to the host `Organization Management` directory, and the top-level `parent_key` is `org`

#### Scenario: Hide the department management entrance when the organization plugin is missing
- **WHEN** `org-center` is not installed or not enabled
- **THEN** The host does not display the department management menu and page entry
- **AND** Hosting capabilities such as user management will continue to be available according to organization downgrade rules.

