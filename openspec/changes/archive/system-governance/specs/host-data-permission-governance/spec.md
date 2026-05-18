## ADDED Requirements

### Requirement: Host Must Parse Current User's Effective Data Permission Scope

The system SHALL provide a unified host data permission parsing capability that calculates the effective data scope based on the current request user, enabled roles, and role `dataScope`. Superadmin SHALL obtain the all-data scope. When a normal user has multiple enabled roles, the system SHALL adopt the widest-range merge rule: all data takes precedence over department data, department data takes precedence over self-only data. When there is no logged-in user context, no enabled roles, or the role data scope cannot be confirmed, the system SHALL fail closed and MUST NOT default to widening to all data.

#### Scenario: Superadmin Obtains All Data Scope

- **WHEN** the built-in superadmin user accesses a host resource governed by data permission
- **THEN** the system resolves the effective data scope to all data
- **AND** subsequent resource queries and write operations do not attach department or self-only scope restrictions

#### Scenario: Multi-role Adopts Widest Range

- **WHEN** a normal user simultaneously has two enabled roles with `department data` and `self only data`
- **THEN** the system resolves the effective data scope to department data
- **AND** does not use the narrower self-only range to override the department range

#### Scenario: Any Enabled Role Has All Data

- **WHEN** any of a normal user's enabled roles has `dataScope=1`
- **THEN** the system resolves the effective data scope to all data

#### Scenario: No User Context Fails Closed

- **WHEN** a host resource operation governed by data permission lacks the current user context
- **THEN** the system rejects the operation or returns empty results
- **AND** does not default to executing with all data scope

#### Scenario: Disabled Roles Do Not Participate in Data Scope Resolution

- **WHEN** a user only has disabled roles or no enabled roles
- **THEN** the system MUST NOT obtain data permission scope from these roles
- **AND** host resources governed by data permission return empty results or reject write operations

### Requirement: Host Resources Must Integrate Data Permission Through Explicit Policies

The system SHALL require each host resource integrating data permission governance to declare an explicit resource policy. The resource policy SHALL contain a stable resource key, applicable table or business resource, self-scope resolution method, department-scope resolution method, and controlled operation types. The host MUST NOT auto-infer whether a resource integrates data permission through column names; system governance interfaces without declared policies MUST NOT be auto-filtered by generic components.

#### Scenario: Inject Filter Conditions After Declaring Resource Policy

- **WHEN** `system.file` declares self-scope using `sys_file.created_by` and department-scope using uploader department set
- **THEN** file list, detail, download-by-id, and delete operations inject data permission boundaries per that policy
- **AND** uploaded file URL access `/api/v1/uploads/*path` does not require login per that policy, but must retain path normalization and file metadata record validation

#### Scenario: Governance Interfaces Without Declared Policies Are Not Auto-Filtered

- **WHEN** menu management, role management, dictionary management, or configuration management have not declared data permission resource policies
- **THEN** the system MUST NOT auto-apply data permission filtering to these tables because they have `created_at`, `updated_at`, or other common fields
- **AND** these interfaces continue to be governed by function permission and system protection rules

#### Scenario: Resource Policy Covers Read and Write Operations

- **WHEN** a host resource declares integration with data permission
- **THEN** that resource's list, detail, export, download, update, delete, status change, and execution-type operations must use the same resource policy
- **AND** batch operations must execute only after all targets pass data permission validation

### Requirement: Department Data Scope Must Be Resolved Through Organizational Capability

The system SHALL resolve department data scope through organizational capability interfaces, rather than requiring every host business table to hold a `dept_id` field. Under department scope, the system SHALL prefer the database-side constraint capability provided by organizational capability, pushing department ownership judgment down to the database for execution; when a specific organizational capability cannot provide database-side constraints, using a visible user set as fallback is allowed, and it MUST avoid generating unbounded-length user ID `IN` conditions on large data volume paths. When organizational capability is unavailable, all-data scope is unaffected; department-data scope SHALL degrade to self-only scope; self-only scope SHALL continue to execute by current user.

#### Scenario: Department Scope Resolves User Set

- **WHEN** user role data scope is department data and accessing file list
- **THEN** the system preferentially injects database-side department ownership constraint into the file list query through organizational capability
- **AND** the file list only returns files whose uploader is within the visible department scope of the current user

#### Scenario: User Management Department Scope Uses Database-Side Semi-Join

- **WHEN** user role data scope is department data and organizational capability is available
- **AND** user accesses user management list, detail, export, or write operations
- **THEN** the system MUST NOT first pull all user IDs within the department scope into application layer and concatenate them into `sys_user.id IN (...)`
- **AND** should filter target users through database-side `EXISTS`, `JOIN`, or equivalent semi-join constraints

#### Scenario: Department Scope Degrades to Self-Only When Org Capability Unavailable

- **WHEN** user role data scope is department data
- **AND** `org-center` is not installed, not enabled, or organizational capability provider is unavailable
- **THEN** the system filters list, detail, and export results by current user self-only scope
- **AND** write operations and execution-type operations only allow acting on data visible to the current user

#### Scenario: Self-Only Scope Still Executable When Org Capability Unavailable

- **WHEN** user role data scope is self-only data
- **AND** organizational capability provider is unavailable
- **THEN** the system still filters the current user's own resources by user ID

### Requirement: First-Batch Host Modules Must Integrate Data Permission by Applicability

The system SHALL include host modules with clear user or department ownership semantics in the first-batch data permission governance. First-batch integration modules include user management, file management, user-created cron jobs, and online user sessions. The user message module SHALL maintain current user self-isolation semantics. System governance modules SHALL not be included in data permission filtering.

#### Scenario: User Management Integrates Data Permission

- **WHEN** a normal user accesses user management list, detail, export, or write operations
- **THEN** the system limits target user scope by user identity and department user set resolved through organizational capability

#### Scenario: File Management Integrates Data Permission

- **WHEN** a normal user accesses file list, detail, batch info, download-by-id, or delete
- **THEN** the system limits target file scope by file uploader and uploader's department
- **AND** public uploaded file URL access is not a file management permission entry; it MUST NOT bypass path normalization, metadata existence validation, or storage backend read boundary

#### Scenario: User-Created Tasks Integrate Data Permission

- **WHEN** a normal user accesses user-created tasks in cron job management
- **THEN** the system limits target task scope by `sys_job.created_by` and creator's department
- **AND** `sys_job.is_builtin=1` built-in task projections are not filtered by data permission

#### Scenario: Online User Integrates Data Permission

- **WHEN** a normal user queries online sessions or forces offline
- **THEN** the system limits target session scope by session user and session user's department

#### Scenario: System Governance Modules Do Not Integrate Data Permission

- **WHEN** admin accesses menu, role base governance, dictionary, configuration, plugin governance, i18n, health check, system info, public config, cache, lock, or cluster coordination interfaces
- **THEN** the system does not apply role `dataScope` filtering to these governance data
- **AND** continues to control access per existing function permission, built-in protection, and component consistency rules

### Requirement: Data Permission Must Constrain Write Operations and Execution-Type Operations

The system SHALL verify whether target records are within the current caller's data permission scope before write operations and execution-type operations. When the target record is not within scope, the system SHALL return a structured business error; it MUST NOT rely solely on frontend hiding buttons, and MUST NOT only filter at the list interface.

#### Scenario: Reject Updating Invisible User

- **WHEN** user role data scope is self-only data
- **AND** the user attempts to update another user's profile or status
- **THEN** the system rejects the operation and preserves the target user's data unchanged

#### Scenario: Reject Deleting Invisible File

- **WHEN** user role data scope is department data
- **AND** the file to be deleted was uploaded by someone outside the user's department scope
- **THEN** the system rejects the delete
- **AND** does not delete the physical file or database record

#### Scenario: Batch Operations Must Validate All Targets

- **WHEN** the caller batch-deletes multiple files or users
- **AND** any target record is not within the current caller's data permission scope
- **THEN** the entire batch operation is rejected
- **AND** no partial deletion is executed on any target record

### Requirement: Data Permission Cache Must Maintain Cross-Node Consistency

The system SHALL treat role data scope, user-role relationships, and user organizational ownership as authoritative sources for data permission context. If the implementation caches data permission context or department user sets, the system MUST, after relevant authoritative data changes, discard old snapshots on all nodes within the allowed window through explicit-scoped cache revision numbers or access topology invalidation mechanisms. When cache freshness cannot be confirmed and exceeds the fault window, host resources governed by data permission MUST fail closed.

The system SHALL merge effective role data scope into the login token access snapshot cache, reusing the access topology revision number for freshness control. Host resources governed by data permission SHOULD read the effective data scope from this snapshot in the request hot path, and SHOULD NOT re-query `sys_user_role` and `sys_role` for each data permission check. Department scope's specific organizational ownership constraints MUST NOT be cached as unbounded user ID sets long-term; they should be preferentially pushed down to the database side through organizational capability.

#### Scenario: Cache Invalidation After Role Data Scope Change

- **WHEN** admin modifies a role's `dataScope`
- **THEN** the system publishes an access topology or data permission cache revision number
- **AND** subsequent requests MUST NOT indefinitely use the old data permission scope

#### Scenario: Request Hot Path Reuses Token Data Permission Snapshot

- **WHEN** user has passed permission check on a protected API and established the current token's access snapshot
- **AND** the user subsequently accesses a list, detail, or write operation interface governed by data permission
- **THEN** the system reads the effective role data scope from the token access snapshot
- **AND** does not re-query `sys_user_role` and `sys_role` for each data permission check

#### Scenario: Cache Invalidation After User-Role Relationship Change

- **WHEN** admin assigns or revokes a role for a user
- **THEN** the system publishes an access topology or data permission cache revision number
- **AND** the affected user's subsequent requests execute per the new role data scope

#### Scenario: Cache Invalidation After Organizational Ownership Change

- **WHEN** organizational capability provider updates a user's department ownership
- **THEN** the system MUST invalidate data permission caches that depend on that ownership
- **AND** department data scope queries MUST NOT indefinitely use the old department user set

### Requirement: Data Permission Errors Must Be Localized

The system SHALL use `bizerr` wrapping for data permission errors entering HTTP API responses, and provide error translations for all enabled runtime languages. When adding or modifying user-visible error text, runtime i18n JSON must be maintained synchronously. If modifying API DTO documentation metadata, non-English apidoc resources and translation completeness checks must also be maintained.

#### Scenario: Data Not Visible Error Returns Structured Response

- **WHEN** user requests detail of a record not within their data permission scope
- **THEN** the system returns a structured business error
- **AND** the response contains a stable error code, messageKey, messageParams, and English fallback

#### Scenario: Runtime Translations Complete

- **WHEN** a data permission error is returned in `zh-CN`, `en-US`, or `zh-TW` language environment
- **THEN** the response message uses the corresponding language translation
- **AND** does not expose untranslated i18n keys
