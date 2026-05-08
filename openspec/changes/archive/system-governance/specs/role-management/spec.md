## MODIFIED Requirements

### Requirement: Must Support Data Scope Options

The system SHALL support simplified data scope values of all data, department data, and self-created data. The role data scope SHALL be used by the host data permission governance capability as input for resolving the current user's effective data scope, not merely a display field on the role record. After role data scope creation, update, enable, disable, role menu permission change, and user-role relationship change, the system SHALL trigger access topology or data permission cache invalidation so that subsequent protected APIs do not indefinitely use the old data scope. When organizational management capability is not enabled, the role add and edit forms SHALL hide or disable the department-data option to prevent configuring data scopes that cannot take effect.

#### Scenario: Configure Data Scope

- **WHEN** role data scope is set to one of the supported values
- **THEN** the system stores that value for host data permission governance use

#### Scenario: Role Data Scope Participates in Data Permission Resolution

- **WHEN** user has an enabled role with `dataScope=2`
- **THEN** host data permission governance capability treats that role as a department-data scope input
- **AND** host resources governed by data permission limit data per department scope

#### Scenario: Hide Department-Data Option When Org Capability Not Enabled

- **WHEN** `org-center` is not installed, not enabled, or organizational capability provider is unavailable
- **AND** admin opens the role add or edit form
- **THEN** the data permission options do not allow selecting the department-data scope
- **AND** newly created roles default to the available option among all-data or self-only

#### Scenario: Role Data Scope Change Triggers Cache Invalidation

- **WHEN** admin modifies a role's `dataScope` and saves successfully
- **THEN** the system publishes an access topology or data permission cache revision number
- **AND** users with that role MUST NOT indefinitely use the pre-modification data scope in subsequent requests

#### Scenario: Disabled Roles No Longer Contribute Data Scope

- **WHEN** admin disables a role
- **THEN** users associated with that role no longer obtain that role's menu permissions on login
- **AND** host data permission governance capability no longer uses that role's `dataScope`
