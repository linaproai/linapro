## Why

LinaPro's admin console, after completing the foundational business modules (user, department, post, dictionary, notifications, etc.), lacked core system governance capabilities: administrators could not track user operation behavior and login history, could not view online users and server runtime status, could not browse backend API documentation and system runtime information, and could not dynamically adjust system parameters at runtime. Additionally, while role management provided `dataScope` (all data / department data / self only) configuration, the host core interfaces mainly only checked menu and button permissions, without consistently applying data scope to host business data queries, details, exports, downloads, updates, deletes, and execution operations. This change addresses system observability, audit security, self-description, runtime configuration, and host-side data permission governance.

## What Changes

### Audit Log System
- New **Operation Log** module: middleware automatically intercepts and records all write operations (POST/PUT/DELETE) and specially marked query operations (such as exports), recording operator, operation time, operation type, request parameters, response results, elapsed time, and other information
- New **Login Log** module: automatically records login behavior during user login/logout, including login username, login time, IP, browser, operating system, login result, and other information
- Both log types support conditional filtering, detail viewing, time-range-based cleanup (hard delete), batch deletion, and xlsx export
- Operation log recording uses a hybrid middleware auto-interception + API definition tag approach; request parameters and response results undergo length truncation and password masking

### System Observability
- New **Online User** feature: session tracking based on MySQL MEMORY engine, displaying current online user list with admin support for forced offline of specified users
- New **Server Monitor** feature: periodic server metric collection via gopsutil (CPU, memory, disk, network traffic), written to database, supporting multi-node distributed deployment; frontend displays server information, metrics, and Go runtime information
- Refactored existing authentication middleware: creates session records on login, deletes session records on logout, and validates session validity during request processing (supporting forced offline and timeout cleanup)

### System Self-Description
- New **System API Docs** page: integrates Stoplight Elements OpenAPI document UI via iframe embedding, displays backend API documentation with online testing support; dynamically generates `servers[0].url` from the request origin instead of using a fixed `http://localhost:8080`
- New **System Info** page: displays project introduction, backend/frontend technology component lists; backend provides system runtime information API
- New **Component Demo** page: embeds vben5 official website demo page via iframe, with load failure handling

### Runtime Configuration Management
- New **Parameter Settings** module: `sys_config` table stores system parameters in key-value form, supporting full CRUD, key-name query, Excel export/import (with overwrite/ignore modes)
- Optimized **Dictionary Management** export/import: supports merged export/import of dictionary types and dictionary data into dual-sheet Excel files, simplifying the operation workflow

### Host Data Permission Governance
- New unified host data permission parsing capability: resolves the current user's effective data scope based on enabled roles and `dataScope`, with superadmin bypass and multi-role widest-range merging
- New explicit resource policy registration for host resources: each governable resource declares its own filtering strategy by user column, department column, or organizational capability; no automatic column-name inference
- First-batch integration of suitable host business modules:
  - User Management: filters by user identity and department user set for list, detail, export, update, delete, status change, password reset, and role-authorized user list operations
  - File Management: filters by uploader `created_by` and uploader's department for list, detail, batch info, download-by-id, delete, and aggregate queries; uploaded file URL access remains public with path normalization and metadata validation
  - Cron Job Management: filters user-created tasks (`is_builtin=0`) by `created_by` and creator's department; built-in task projections maintain system governance semantics
  - Online User: filters online session list and forced-offline by session `user_id` and user's department
  - User Message: maintains current user self-isolation semantics; data permission must not widen this boundary
- Explicitly excluded from data permission filtering: menu management, role base CRUD, dictionary, configuration, plugin governance, i18n, health check, system info, public config, cache, locks, and cluster coordination
- Organizational capability degradation: when org-center is unavailable, "department data" scope degrades to "self only" scope

## Capabilities

### New Capabilities
- `oper-log`: Operation log auto-recording (middleware interception), list query, detail view, time-range cleanup, batch delete, export
- `login-log`: Login log auto-recording, list query, detail view, time-range cleanup, batch delete, export
- `online-user`: Online user management including session tracking (MySQL MEMORY engine), online user list query, forced offline, inactive session auto-cleanup
- `server-monitor`: Server monitoring including periodic metric collection (CPU, memory, disk, network, Go runtime), multi-node database writes, frontend metric display with node switching
- `system-api-docs`: System API documentation page integrating Stoplight Elements to display OpenAPI docs with online testing, dynamically generating server URLs from request origin
- `system-info`: Version info page displaying project info, backend/frontend component info; backend provides system runtime info API
- `component-demo`: Component demo page with iframe embedding of vben5 official demo, with load failure handling
- `config-management`: Complete CRUD management for system parameter settings including list query, create, edit, delete, key-name query, Excel export/import
- `dict-import`: Dictionary type and data import functionality, plus merged export/import optimization
- `host-data-permission-governance`: Unified host data permission parsing, resource policy registration, module applicability classification, query injection, write-operation boundary validation, cache consistency, and first-batch host resource integration

### Modified Capabilities
- `user-auth`: Authentication flow refactored -- login writes login log and creates session record, logout writes login log and deletes session record, middleware adds session validity check to support forced offline
- `role-management`: Role data scope is no longer just a stored field; it must serve as input to host data permission parsing, and changes must trigger access topology cache invalidation
- `user-management`: User data queries and mutations must follow the caller's data permission scope
- `online-user`: Online session list and forced offline must follow the caller's data permission scope
- `cron-job-management`: User-created tasks and associated execution logs must follow the caller's data permission scope; built-in task projections maintain system governance semantics
- `user-message`: User messages maintain current user self-isolation boundary; data permission must not widen this boundary

## Impact

- **Database**: New tables `sys_oper_log`, `sys_login_log`, `sys_online_session` (MEMORY engine), `sys_server_monitor`, `sys_config`; new system monitoring menu and dictionary data
- **Backend API**: New RESTful interfaces for operation logs, login logs, online users, server monitoring, system info, parameter settings
- **Backend Services**: New service modules for operlog, loginlog, online-user, server-monitor, sysinfo, sysconfig, dataperm; new operation log middleware; refactored auth service and middleware
- **Backend Dependencies**: New `github.com/shirou/gopsutil/v4` (system metric collection), `mssola/useragent` (UA parsing)
- **Frontend Routes**: New "System Monitor" menu (operation logs, login logs, online users, server monitor), "System Info" menu (system API docs, version info, component demo), parameter settings menu item
- **Frontend Views**: New page components under monitor/ and about/ directories, system/config/ parameter settings page
- **Frontend Dependencies**: Introduced `@scalar/api-reference` (OpenAPI document rendering)
- **Dictionary Data**: New `sys_oper_type` (operation type), `sys_login_status` (login status) dictionary types
- **Data Permission**: New `dataperm` internal service; affected backend services for user, file, job, session, and message modules; access topology cache integration for data scope changes
- **i18n**: New bizerr error codes for data permission rejection must include zh-CN, en-US, zh-TW runtime translations
