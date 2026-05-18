## 1. Database and Infrastructure

- [x] 1.1 Create operation log and login log SQL files, including `sys_oper_log` and `sys_login_log` table creation statements, and `sys_oper_type` dictionary type and dictionary data Seed DML
- [x] 1.2 Create online user and server monitor SQL files, adding `sys_online_session` (MEMORY engine) and `sys_server_monitor` table DDL, and new system monitoring menu data
- [x] 1.3 Create parameter settings SQL file, adding `sys_config` table DDL and menu/button permission seed data
- [x] 1.4 Execute `make init` to update database, execute `make dao` to generate DAO/DO/Entity files
- [x] 1.5 Add User-Agent parsing dependency (`mssola/useragent`) and system metric collection dependency (`github.com/shirou/gopsutil/v4`)

## 2. Backend - Audit Log Module

### 2.1 Login Log

- [x] 2.1.1 Create `api/loginlog/v1/` interface definitions: List, Get, Clean, Export, Delete (batch delete)
- [x] 2.1.2 Execute `make ctrl` to generate controller skeleton
- [x] 2.1.3 Implement `internal/service/loginlog/` service layer: Create, List, Get, Clean, Export, Delete
- [x] 2.1.4 Fill controller method implementations, register login log routes in `cmd_http.go`

### 2.2 Operation Log

- [x] 2.2.1 Create `api/operlog/v1/` interface definitions: List, Get, Clean, Export, Delete (batch delete)
- [x] 2.2.2 Execute `make ctrl` to generate controller skeleton
- [x] 2.2.3 Implement `internal/service/operlog/` service layer: Create, List, Get, Clean, Export, Delete
- [x] 2.2.4 Fill controller method implementations, register operation log routes in `cmd_http.go`

### 2.3 Operation Log Middleware

- [x] 2.3.1 Implement operation log middleware `internal/service/middleware/operlog.go`: intercept write operations, parse `g.Meta` tags for module name and operation type, record request parameters (truncated + masked) and response results (truncated), calculate elapsed time, async database write
- [x] 2.3.2 Register operation log middleware after Auth middleware in `cmd_http.go`
- [x] 2.3.3 Add `operLog:"4"` tag to existing export interface `g.Meta`

## 3. Backend - Authentication Module Refactoring

- [x] 3.1 Define session storage abstract interface `SessionStore`, implement MySQL-based `DBSessionStore` (Create, Query, Delete, List filtering, TouchOrValidate, CleanupInactive)
- [x] 3.2 Refactor auth service (`internal/service/auth/`): create session record and write login log on successful login, delete session record and write login log on logout
- [x] 3.3 Refactor auth middleware (`internal/service/middleware/`): check session validity via TouchOrValidate after JWT verification, return 401 when session does not exist
- [x] 3.4 Implement inactive session auto-cleanup scheduled task, with timeout threshold and cleanup frequency configurable through config file

## 4. Backend - Online User Module

- [x] 4.1 Create online user API definition (`api/monitor/v1/`): `GET /monitor/online/list`, `DELETE /monitor/online/{tokenId}`
- [x] 4.2 Implement online user Controller and Service: list query, forced offline logic
- [x] 4.3 Register online user routes in `cmd_http.go`

## 5. Backend - Server Monitor Module

- [x] 5.1 Implement metric collection service (CPU, memory, disk, network, Go runtime, server basic info)
- [x] 5.2 Implement periodic collection task: collect once immediately on service startup, then every 30 seconds, UPSERT strategy keeping only latest record per node
- [x] 5.3 Create server monitor API definition (`api/monitor/v1/`): `GET /monitor/server`
- [x] 5.4 Implement server monitor Controller and Service: read latest monitoring data per node from database
- [x] 5.5 Register server monitor routes and periodic collection task in `cmd_http.go`

## 6. Backend - System Info Module

- [x] 6.1 Create API definition `api/sysinfo/v1/info.go`, define request/response structs for `GET /api/v1/system/info`
- [x] 6.2 Implement `internal/service/sysinfo/sysinfo.go` system info service layer, obtain runtime information
- [x] 6.3 Fill controller method implementations, register system info routes in `cmd_http.go` (within auth route group)

## 7. Backend - Parameter Settings Module

- [x] 7.1 Create API definition `api/config/v1/`: List, Get, Create, Update, Delete, ByKey, Export, Import, ImportTemplate (7 files)
- [x] 7.2 Execute `make ctrl` to generate controller skeleton
- [x] 7.3 Implement `internal/service/sysconfig/` service layer: complete CRUD, query by key name, export, import (overwrite/ignore modes)
- [x] 7.4 Fill controller method implementations, register parameter settings routes in `cmd_http.go`

## 8. Backend - Dictionary Export/Import Optimization

- [x] 8.1 New dictionary merged export endpoint (`GET /dict/export`), simultaneously exporting dictionary types and dictionary data to dual-sheet Excel file
- [x] 8.2 New dictionary merged import endpoint (`POST /dict/import`), supporting simultaneous import of dictionary types and dictionary data
- [x] 8.3 New dictionary import template download endpoint, returning template file with two sheets
- [x] 8.4 Dictionary type delete logic changed to cascade delete; deleting dictionary type also deletes associated dictionary data

## 9. Backend - Scheduled Tasks and Configuration Refactoring

- [x] 9.1 Migrate all scheduled tasks from gtimer to gcron component, using crontab expressions
- [x] 9.2 Change all hardcoded config reads to struct-based maintenance, create config structs grouped by module
- [x] 9.3 Migrate `internal/config/` to `internal/service/config/`, split into independent Go files by module
- [x] 9.4 Extract scheduled task logic from cmd_http.go into service/cron independent component

## 10. Backend - Host Data Permission Governance

### 10.1 Core Data Permission Service

- [x] 10.1.1 Review GoFrame DAO / gdb.Model integration approach using `goframe-v2` skill, avoid manual SQL concatenation
- [x] 10.1.2 Add host internal data permission service, parse current user, superadmin, enabled roles and effective `dataScope`
- [x] 10.1.3 Implement multi-role widest-range merge rules: all data > department data > self only > no permission
- [x] 10.1.4 Define data permission named types and constants; prohibit hardcoding data scope strings or bare enum semantics in business branches
- [x] 10.1.5 Define explicit per-module policy integration approach, supporting user column, department semi-join, and custom visibility check resource integration
- [x] 10.1.6 Implement list query constraint injection capability, supporting empty-scope fast-return
- [x] 10.1.7 Implement detail, write operation, and execution-type operation target record visibility check capability
- [x] 10.1.8 Define `bizerr.Code` for data permission rejection, context missing, and org capability unavailable

### 10.2 Organizational Capability and Cache Consistency

- [x] 10.2.1 Review `orgcap` current capabilities, supplement department user set resolution interfaces or adapter methods needed by data permission
- [x] 10.2.2 Implement safe degradation strategy when org capability unavailable: all data unaffected, department degrades to self-only, self-only continues to work
- [x] 10.2.3 Include role `dataScope` changes, role enable/disable, and user-role relationship changes in access topology or data permission cache invalidation
- [x] 10.2.4 Effective role data scope merged into token permission snapshot; department user set not cached; future independent cache must introduce explicit scope/revision
- [x] 10.2.5 No independent data permission cache domain this phase; reuse access topology cache invalidation and cross-node revision testing

### 10.3 User Management Integration

- [x] 10.3.1 User list and export integrate data permission, supporting all, department, and self-only scopes
- [x] 10.3.2 User detail integrates data permission; out-of-range users return structured data-not-visible error
- [x] 10.3.3 User update, status change, password reset, and role association change integrate target user visibility check
- [x] 10.3.4 User single and batch delete integrate data permission; batch operation rejects entirely if any target is not visible
- [x] 10.3.5 Role-authorized user list and user selection range integrate data permission, avoiding exposing out-of-range users on authorization page
- [x] 10.3.6 Preserve built-in admin and current user deletion protection rules, ensuring data permission does not bypass existing protections

### 10.4 File Management Integration

- [x] 10.4.1 File list integrates `sys_file.created_by` uploader range filtering
- [x] 10.4.2 File detail, batch info integrate uploader visibility check; uploaded file URL access remains public with path normalization and metadata validation
- [x] 10.4.3 File download integrates data permission; out-of-range files must not return binary content
- [x] 10.4.4 File delete integrates data permission; refuse to delete out-of-range database records and physical files
- [x] 10.4.5 File suffix, scenario, and other aggregate queries filter by current data permission scope, avoiding leaking out-of-range data existence

### 10.5 Cron Job and Log Integration

- [x] 10.5.1 User-created task list integrates `sys_job.created_by` data permission filtering
- [x] 10.5.2 Maintain `sys_job.is_builtin=1` built-in task projection without data permission filtering, continuing to use built-in task governance rules
- [x] 10.5.3 User-created task detail, edit, delete, enable, disable, reset integrate target task visibility check
- [x] 10.5.4 Manual trigger of user-created tasks checks data permission first; out-of-range tasks must not create `sys_job_log`
- [x] 10.5.5 Task log list, detail, cleanup, and terminate running log integrate parent task's data permission boundary
- [x] 10.5.6 Confirm Shell task independent permission point and data permission both apply; lacking either must reject

### 10.6 Online User and User Message Integration

- [x] 10.6.1 Online user list integrates data permission filtering by `sys_online_session.user_id`
- [x] 10.6.2 Forced offline checks target `tokenId`'s owning user against current data permission scope before execution
- [x] 10.6.3 User message unread count, list, mark read, and delete paths confirmed to maintain current user self-isolation
- [x] 10.6.4 Add user message tests confirming users with all data permission still cannot read, mark, or delete others' messages

### 10.7 i18n and API Documentation Governance

- [x] 10.7.1 Add `zh-CN`, `en-US`, `zh-TW` runtime translations for new `bizerr` errors
- [x] 10.7.2 Check whether this change modifies API DTO documentation metadata; if so, maintain non-English apidoc i18n JSON
- [x] 10.7.3 Confirm role page does not add new frontend visible fields; no new role form translations needed in frontend runtime language pack
- [x] 10.7.4 Add or update i18n completeness tests ensuring new error translations are not missing

## 11. Frontend - Operation Log Page

- [x] 11.1 Create operation log API layer: `src/api/monitor/operlog/`
- [x] 11.2 Create operation log list page: `src/views/monitor/operlog/index.vue` and `data.ts` (table + filters)
- [x] 11.3 Create operation log detail drawer component, request parameters and response results use vue-json-pretty for JSON syntax highlighting
- [x] 11.4 Implement cleanup functionality (modal with time range selection then hard delete) and batch delete functionality

## 12. Frontend - Login Log Page

- [x] 12.1 Create login log API layer: `src/api/monitor/loginlog/`
- [x] 12.2 Create login log list page: `src/views/monitor/loginlog/index.vue` and `data.ts`
- [x] 12.3 Create login log detail modal component
- [x] 12.4 Implement cleanup functionality and batch delete functionality

## 13. Frontend - Online User Page

- [x] 13.1 Create frontend API file `src/api/monitor/online/`
- [x] 13.2 Create online user page `src/views/monitor/online/index.vue` and `data.ts`: search form, VXE-Grid table, toolbar online user count, forced offline Popconfirm
- [x] 13.3 Add system monitor route module `src/router/routes/modules/monitor.ts`

## 14. Frontend - Server Monitor Page

- [x] 14.1 Create frontend API file `src/api/monitor/server/`
- [x] 14.2 Create server monitor page `src/views/monitor/server/index.vue` and sub-components: server info cards, CPU/memory circular progress bars, Go runtime info, disk usage table, network traffic info
- [x] 14.3 Implement multi-node switching logic and collapsible node list layout

## 15. Frontend - System Info Page

- [x] 15.1 Create route module `src/router/routes/modules/about.ts`, define "System Info" top-level menu with three child routes
- [x] 15.2 Create frontend config file `src/views/about/config.ts`, define configurable items
- [x] 15.3 Implement system API docs page: iframe embedding Stoplight Elements static document page
- [x] 15.4 Implement version info page: about project, backend components, frontend components three sections
- [x] 15.5 Implement component demo page: iframe embedding vben5 official demo, with load failure handling
- [x] 15.6 Create API file `src/api/about/index.ts`, call `GET /api/v1/system/info`

## 16. Frontend - Parameter Settings Page

- [x] 16.1 Create frontend API layer `src/api/system/config/`
- [x] 16.2 Create parameter settings page `src/views/system/config/index.vue`, `config-modal.vue`, `data.ts`
- [x] 16.3 Add parameter settings route to system route module

## 17. Frontend - Dictionary Export/Import Optimization

- [x] 17.1 Frontend dictionary type panel updates export/import functionality to use merged interface
- [x] 17.2 Frontend dictionary data panel removes export and import buttons
- [x] 17.3 Abstract generic export confirmation modal component `ExportConfirmModal`, reuse across all export modules
- [x] 17.4 Unify export file naming conventions across all modules

## 18. Frontend - General Improvements

- [x] 18.1 User management page status field changed to dynamically read from dictionary module
- [x] 18.2 Department/post management page status options changed to dynamically read from dictionary module
- [x] 18.3 Dictionary management page clears dictStore cache after modifying dictionary data
- [x] 18.4 Remove extra menu items from avatar dropdown menu, fix user email and nickname display
- [x] 18.5 Personal center form field adjustments (nickname required, non-required field corrections)
- [x] 18.6 Global pagination options add 100 items/page

## 19. API Documentation Completion

- [x] 19.1 Complete auth module API documentation: add dc tags to g.Meta, supplement dc and eg tags for all input/output fields
- [x] 19.2 Complete user module API documentation (13 files)
- [x] 19.3 Complete dept module API documentation (8 files)
- [x] 19.4 Complete post module API documentation (8 files)
- [x] 19.5 Complete dict module API documentation (14 files)
- [x] 19.6 Complete notice module API documentation (5 files)
- [x] 19.7 Complete loginlog module API documentation (5 files)
- [x] 19.8 Complete operlog module API documentation (5 files)
- [x] 19.9 Complete usermsg module API documentation (6 files)
- [x] 19.10 Complete sysinfo module API documentation (1 file)

## 20. API Documentation Dynamic Server URL

- [x] 20.1 Override OpenAPI `servers[0].url` in host `/api.json` handler based on current request origin
- [x] 20.2 Retain `serverDescription` as service address description, with safe fallback when host is missing
- [x] 20.3 Remove dependency on fixed `openapi.serverUrl` as runtime request address authoritative source

## 21. Backend Tests

- [x] 21.1 TC0026-TC0034: Operation log and login log list query, detail view, cleanup, export, auto-recording tests
- [x] 21.2 TC0044-TC0045: System API docs page and version info page load tests
- [x] 21.3 TC0049-TC0052: Online user list, search, forced offline, server monitor page tests
- [x] 21.4 Parameter settings page CRUD, search, export/import tests
- [x] 21.5 Dictionary merged export/import tests
- [x] 21.6 Export confirmation modal tests (all export modules)
- [x] 21.7 Dictionary modification global effect tests
- [x] 21.8 Run all E2E tests to confirm no regressions
- [x] 21.9 Add backend unit tests covering direct-mapped port, frontend proxy to backend port, and `X-Forwarded-Proto=https` origin scenarios
- [x] 21.10 New E2E case `TC0175-api-docs-request-origin.ts` verifying `/api.json` `servers[0].url` dynamically changes with frontend proxy and backend direct access

## 22. Data Permission Backend Tests

- [x] 22.1 New data permission parsing unit tests covering superadmin, multi-role widest range, disabled roles, no roles, no user context
- [x] 22.2 New org capability degradation tests covering department-to-self degradation and self-only without org capability
- [x] 22.3 New user management data permission tests covering list, detail, export, write operations, batch delete
- [x] 22.4 New file management data permission tests covering list, detail, download, delete, aggregate queries
- [x] 22.5 New cron job and task log data permission tests covering user-created tasks, built-in task projection, and log termination
- [x] 22.6 New online user data permission tests covering list and forced offline
- [x] 22.7 New user message self-isolation regression tests covering all-data-permission not widening message boundary
- [x] 22.8 Run `cd apps/lina-core && go test ./...`

## 23. Data Permission E2E Tests

- [x] 23.1 Create `hack/tests/e2e/settings/user/TC0170-user-data-permission.ts`, covering TC-170a department user list filtering, TC-170b self-only detail restriction, TC-170c out-of-range user write operation rejection
- [x] 23.2 Create `hack/tests/e2e/settings/file/TC0171-file-data-permission.ts`, covering TC-171a file list filtering, TC-171b out-of-range download rejection, TC-171c out-of-range delete rejection
- [x] 23.3 Create `hack/tests/e2e/scheduler/job/TC0172-job-data-permission.ts`, covering TC-172a user-created task list filtering, TC-172b built-in task projection visibility, TC-172c out-of-range trigger rejection
- [x] 23.4 Create `hack/tests/e2e/monitor/TC0173-online-user-data-permission.ts`, covering TC-173a online user list filtering, TC-173b out-of-range forced offline rejection
- [x] 23.5 Create `hack/tests/e2e/content/message/TC0174-user-message-self-boundary.ts`, covering TC-174a all-data-permission does not read others' messages, TC-174b does not mark others' messages, TC-174c does not delete others' messages
- [x] 23.6 Run affected E2E test cases and record results

## 24. Code Quality and Refactoring

- [x] 24.1 user/dept/file service transaction management fix; Create/Update methods use transactions to ensure data consistency
- [x] 24.2 Extract duplicate department tree traversal logic from user and post services into dept service for reuse
- [x] 24.3 Replace MySQL-specific FIND_IN_SET in department hierarchy query with cross-database generic parent_id iterative query implementation
- [x] 24.4 Fix user list query N+1 problem, batch query department info
- [x] 24.5 Dictionary type update validates Type field uniqueness
- [x] 24.6 Log export method adds record count limit to prevent memory overflow
- [x] 24.7 File upload adds filename sanitization to prevent path traversal attacks
- [x] 24.8 Notice announcement NoticeModal formRules changed to reactive object to fix Vue warn
- [x] 24.9 Fix breadcrumb link color in dark mode
- [x] 24.10 Container environment compatibility optimization: filter virtual filesystem mount points, graceful degradation on collection failure

## 25. Verification and Review

- [x] 25.1 Run `openspec validate host-data-permission-governance --strict`
- [x] 25.2 This round of data permission did not add host SQL; org-center plugin indexes verified via backend and E2E paths, database not reset
- [x] 25.3 Run affected frontend type checks or build commands
- [x] 25.4 Execute `/lina-review`, focusing on data permission gaps, cache consistency, and i18n completeness
