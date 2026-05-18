## Context

LinaPro's admin console currently has foundational business modules including user management, department management, post management, dictionary management, notifications, and file management. The system uses pure stateless JWT authentication, with no backend session tracking. It lacks operation auditing, system monitoring, API documentation display, runtime parameter configuration, and host-side data permission governance.

The host role management already stores three data scopes in `sys_role.data_scope`: `1` (all data), `2` (department data), `3` (self only). Current host static APIs primarily use `g.Meta permission` for menu/button permission checks. The user access context includes roles, menus, permission codes, and superadmin flag, but lacks a stable data permission snapshot. The plugin resource layer already has implementation for injecting `userColumn`/`deptColumn` based on plugin resource declarations, but host built-in business interfaces have not yet formed equivalent governance capability.

Tech stack: GoFrame v2 + MySQL + JWT (HS256), frontend Vben5 + Vue 3 + Ant Design Vue. There is a complete middleware chain (CORS -> Response Wrapper -> Context Injection -> JWT Auth), and a route definition mechanism based on `g.Meta`. Existing export functionality uses the `excelize` library for xlsx format.

## Goals / Non-Goals

**Goals:**
- Auto-record all write operations to operation log table via middleware, record login logs during login/logout flows
- Implement MySQL MEMORY engine-based session tracking with online user list query and forced offline
- Implement gopsutil-based periodic server metric collection written to MySQL, supporting multi-node distributed deployment
- Integrate Stoplight Elements as OpenAPI document UI with dynamic server URL generation from request origin
- Provide complete CRUD management for system parameters with Excel export/import
- Optimize dictionary management export/import with merged functionality
- Establish unified host data permission component for centralized parsing of caller's effective data scope
- Declare clear policies for suitable host resources, avoiding each service writing duplicate rules
- Cover user, file, user-created tasks, online sessions, and user messages in first batch
- Guarantee list, detail, export, download, update, delete, status change, and execution operations use the same data boundary
- Reuse existing `orgcap` optional organizational capabilities; safe degradation when org module is disabled

**Non-Goals:**
- Do not implement IP geolocation parsing
- Do not implement soft delete for logs (log cleanup is hard delete)
- Do not implement real-time push (WebSocket) for monitoring data refresh
- Do not implement historical metric trend charts, only display latest collection snapshot
- Do not implement alerting/alarm functionality
- Do not implement parameter caching mechanism
- Do not self-host vben5 component demo static files
- Do not add custom data scope, department-and-below, cross-department combination, or other complex scopes
- Do not filter menu, role base governance, dictionary, configuration, plugin governance, i18n, health check, system info, or other global governance data by department or self
- Do not redo tenant isolation in this change; data permission only addresses data visibility within the same tenant or single tenant
- Do not change the plugin resource data permission manifest declaration model; only reuse or extract common capabilities in the host implementation

## Decisions

### I. Audit Log System

#### 1. Operation Log Recording Approach -- Middleware Auto-Interception + API Tag Hybrid

**Choice**: Automatically intercept write operations at the middleware layer, and mark query operations needing special recording through `g.Meta` tags.

**Alternatives**:
- A. Pure middleware approach: globally intercept all write operations without modifying business code. But cannot precisely distinguish operation types and module names.
- B. Pure service-layer instrumentation: manually call log recording in each Service method. Precise but highly invasive, easy to miss.
- C. This approach (A+B hybrid): middleware auto-handles POST/PUT/DELETE, obtains module name from `g.Meta`'s `tags` field and operation description from `summary`. For GET operations needing recording (such as exports), add `operLog` tag in `g.Meta`.

**Rationale**: GoFrame's `g.Meta` already contains `tags` (module name) and `summary` (operation description); middleware can obtain these meta-info via `r.GetServeHandler()` without additional encoding.

#### 2. Operation Type (oper_type) Inference Rules

Auto-infer based on HTTP method and `operLog` tag in `g.Meta`:

| HTTP Method | Default oper_type | Description |
|-----------|---------------|------|
| POST | 1 (Create) | If path contains `import`, then 5 (Import) |
| PUT | 2 (Update) | -- |
| DELETE | 3 (Delete) | -- |
| GET + `operLog:"4"` | 4 (Export) | Only GET requests marked with operLog |
| Other | 6 (Other) | Custom tag values |

#### 3. Request/Response Parameter Recording Strategy

**Truncation Length**: Request parameters and response results each truncated to **2000 characters**. Excess is directly truncated with `...(truncated)` appended.

**Masking Rules**: Replace `password` and `Password` field values in request parameter JSON with `***`.

**Rationale**: 2000 characters are sufficient for recording most business operation parameters while preventing log table bloat in batch import scenarios. Only password fields are masked; other fields retain original values for auditing.

#### 4. Logs Do Not Use Soft Delete

Operation log and login log tables **do not have a `deleted_at` field**; cleanup operations directly execute `DELETE FROM` hard delete.

**Rationale**: Log data is audit records; soft-deleting logs causes unlimited log table growth without truly releasing storage. The purpose of cleanup is to free space, making hard delete more appropriate.

#### 5. Middleware Registration Position

Operation log middleware is registered **after** the Auth middleware, ensuring current user information (username) is already available.

```
CORS -> ResponseHandler -> Ctx -> Auth -> OperLog -> Controller
```

#### 6. Login Log Recording Position

Directly call LoginLog Service in Auth Service's login/logout methods rather than through middleware.

**Rationale**: Login endpoint is before the Auth middleware (public endpoint), so middleware cannot intercept it. Login logs need to record login results (success/failure), which can only be accurately obtained in business logic.

#### 7. Async Write

Operation log middleware writes to the database **asynchronously via goroutine** after request processing completes, avoiding impact on API response time.

### II. System Observability

#### 8. Session Storage Uses MySQL MEMORY Engine

**Choice**: MySQL MEMORY engine table `sys_online_session`

**Alternatives**:
- A) gcache in-memory cache: best performance, but lost on process restart, no cross-instance sharing
- B) Redis: supports distributed, but introduces additional component dependency
- C) MySQL MEMORY engine: stored in memory via MySQL, performance close to in-memory cache, no additional dependency

**Rationale**: MEMORY engine reduces component dependency complexity, leveraging existing MySQL infrastructure. Data loss after MySQL restart is acceptable (users re-login). By defining a `SessionStore` abstract interface, seamless switching to gcache + Redis solution in the future is possible.

#### 9. Server Monitoring Uses Periodic Collection + Database Storage

**Choice**: gopsutil periodic collection -> write to `sys_server_monitor` table -> API reads from database

**Alternatives**:
- A) Real-time query (collect at API request time): simple but no multi-node support, slow response
- B) Periodic collection to database: supports multi-node, stateless deployment, historical data queryable

**Rationale**: Periodic collection keeps Lina service stateless. Each node independently collects its own metrics and writes to database; new nodes just need deployment to automatically report.

**Collection Parameters**:
- Collection frequency: default 30 seconds
- Data retention: each node keeps only the latest record (UPSERT strategy)
- Node identification: hostname + local IP auto-obtained

#### 10. Session Active Time Tracking and Auto-Cleanup

`sys_online_session` table adds `last_active_time` field, initialized at login, updated via UPDATE operation on each request with affected row count to determine session existence. Scheduled task cleans up timed-out sessions; timeout threshold and cleanup frequency configurable through config file.

### III. System Self-Description

#### 11. OpenAPI Document UI Selection: Stoplight Elements (iframe approach)

**Choice**: Static HTML file + iframe embedding of Stoplight Elements

**Evolution Path**: Scalar -> Stoplight Elements (Web Component) -> Stoplight Elements (iframe)

**Rationale**: Scalar's API Client popup was obscured; Stoplight Elements via Web Component integration polluted global styles with CSS; iframe embedding achieves complete style isolation. Document HTML changed to static file serving, removing backend API routes, reducing system complexity.

#### 12. Dynamic Server URL Generation for API Documentation

**Problem**: The `openapi.serverUrl` in `metadata.yaml` was fixed at `http://localhost:8080`. When the service is deployed behind container port mapping, reverse proxy, or direct host-mapped port access, the API documentation request addresses would not match the browser's actual accessible address.

**Choice**: Dynamically generate `servers[0].url` from the current request origin in the `/api.json` handler.

**Implementation**:
1. Keep `apidoc.Build(ctx, server)` responsibility for building routes, plugin projections, document title, description, version, and service description.
2. In the `/api.json` HTTP handler, after document build completes and before writing JSON, generate origin from current request and override `document.Servers`.
3. Origin generation rules:
   - Scheme uses GoFrame request's `GetSchema()`, supporting `X-Forwarded-Proto` and TLS.
   - Host uses raw `r.Host`, preserving port, avoiding `GetHost()` stripping the port.
   - When request host is empty, do not override `servers`, preserving build result as fallback.
4. `metadata.yaml` retains `serverDescription` as `servers[0].description`, but no longer relies on `serverUrl` as runtime address source.
5. Stoplight static page continues using relative path `/api.json?lang=...` to load documents; both frontend proxy and backend direct access naturally determine `/api.json`'s origin from the current request.

#### 13. System Info Page Architecture

**Choice**: Backend provides `GET /api/v1/system/info` endpoint + frontend config object

- **Backend API returns**: Go version, GoFrame version, operating system, database version, system startup time, runtime duration, and other runtime information
- **Frontend config object**: project name, version, description, license, homepage link, backend component list, frontend component list
- External link addresses concentrated in frontend config file, no need to modify component code when changing

#### 14. Component Demo Approach: iframe Embedding of External Website

**Choice**: iframe embedding of `https://www.vben.pro/`

**Rationale**: vben.pro has not set X-Frame-Options restrictions, can be normally embedded. Zero size increase, zero maintenance cost. Friendly error page shown on load failure.

### IV. Runtime Configuration Management

#### 15. Parameter Settings Table Design

Table name `sys_config`, fields kept simple: id, name, key (UNIQUE), value, remark, created_at, updated_at, deleted_at.

**Decision**: `key` and `value` are MySQL reserved words, but GoFrame's ORM automatically wraps them with backticks, not affecting usage.

#### 16. Parameter Settings API Design

Follows RESTful conventions: GET `/config` (list), GET `/config/{id}` (detail), POST `/config` (create), PUT `/config/{id}` (update), DELETE `/config/{id}` (delete), GET `/config/key/{key}` (query by key name), GET `/config/export` (export), POST `/config/import` (import).

#### 17. Dictionary Merged Export/Import

New `GET /dict/export` merged export endpoint, simultaneously exporting dictionary types and dictionary data to dual-sheet Excel file. New `POST /dict/import` merged import endpoint, supporting simultaneous import of dictionary types and dictionary data. Frontend dictionary type panel uses merged interface; dictionary data panel removes independent export/import buttons.

### V. Host Data Permission Governance

#### 18. New Host Unified Data Permission Service

Data permission needs to inject query conditions based on resource tables, fields, and business relationships, and also check target record ownership before write operations. HTTP permission middleware only knows routes and `permission` tags, not business query models, target tables, and association relationships.

Therefore, a new internal service is introduced, such as `internal/service/dataperm`, responsible for:
- Obtaining user identity from current `context.Context` and `bizctx`
- Resolving user roles' effective `dataScope`
- Merging multi-role scopes: `all > department > self only > no permission`
- Determining superadmin bypass
- Generating reusable constraint results based on resource policies
- Providing `ApplyListScope(ctx, model, resourceKey)`, `EnsureRecordVisible(ctx, resourceKey, id)` or equivalent narrow interfaces

The alternative of each service reading roles and concatenating conditions itself is fast to implement but scattered, making it very difficult to audit whether detail, export, or delete paths are missed, so it is not adopted.

#### 19. Explicit Resource Policy Registration

Each governable resource must explicitly declare a policy, for example:

| resourceKey | Resource | Self Policy | Department Policy | Notes |
| --- | --- | --- | --- | --- |
| `system.user` | `sys_user` | `sys_user.id = currentUserID` | Inject database-side semi-join via org capability; fallback to resolving department user set | User table has no `dept_id` column |
| `system.file` | `sys_file` | `created_by = currentUserID` | Uploader belongs to current department user set | Table only has uploader, no department column |
| `system.job` | `sys_job` | `created_by = currentUserID` and `is_builtin=0` | Creator belongs to current department user set and `is_builtin=0` | Built-in task projection not filtered by data permission |
| `system.online-session` | `sys_online_session` | `user_id = currentUserID` | Session user belongs to current department user set | Forced offline also requires validation |
| `system.user-message` | User messages | Always current user | Always current user | Self-isolation, not widened by role scope |

Column-name auto-inference (e.g., auto-filtering when seeing `created_by`) is not adopted because different tables' `created_by` may be audit fields, projection fields, or system seed fields, with semantics not necessarily equivalent to business ownership.

#### 20. First-Batch Module Classification by Applicability

**A. First-batch host business modules integrating data permission:**

| Module | Applicable Reason | Filtering Basis | Controlled Operations |
| --- | --- | --- | --- |
| User Management | Users naturally have organizational ownership and self-identity | `sys_user.id` + `orgcap` department user set | List, detail, export, update, delete, batch delete, status, password reset, role-authorized user list |
| File Management | Files have uploader `created_by` | `sys_file.created_by` + uploader department set | List, detail, batch info, download-by-id, delete, suffix/scenario aggregation; uploaded file URL access is public |
| Cron Job Management | User-created tasks have creator `created_by` | `sys_job.created_by` + creator department set | User task list, detail, edit, delete, enable/disable, trigger, log query, log termination |
| Online User | Sessions have `user_id` | `sys_online_session.user_id` + user department set | Online list, forced offline |

**B. Modules with stronger existing self-isolation semantics:**

| Module | Existing Boundary | This Change Requirement |
| --- | --- | --- |
| User Message | Only queries and modifies current user's own messages | Maintain current user boundary; disallow `all data permission` widening |
| Current User Profile | Only reads/writes current logged-in user | Maintain self-isolation; not included in role data scope |

**C. System governance modules not suitable for data permission filtering:**

| Module | Reason for Exclusion | Governance Method |
| --- | --- | --- |
| Menu Management | Global permission topology data, no department/self ownership | Function permission + built-in protection |
| Role Management Base CRUD | Role is authorization governance data; filtering would break authorization management | Function permission + built-in role protection; `dataScope` only serves as subsequent business data scope input |
| Dictionary Management | Dictionary is cross-module enumeration infrastructure | Function permission + built-in dictionary delete protection |
| Configuration Management | Configuration is system runtime parameters | Function permission + built-in config delete protection + config cache consistency |
| Plugin Governance | Plugin state and resource references are global runtime governance | Function permission + plugin lifecycle protection |
| i18n / apidoc | Translation and document resources are global delivery resources | Function permission + i18n cache scope |
| Health Check / System Info / Public Config | Not business data ownership model | Existing public or protected read rules |
| Cache / Locks / Cluster Revision | Infrastructure state, not user business data | Component internal consistency and permission boundary |

#### 21. Department Scope Resolution via orgcap with Safe Degradation

The host core user table has no `dept_id`; department relationships are provided by optional organizational capability. Therefore department scope cannot directly concatenate `dept_id = ?` but must resolve through `orgcap`:
- Current user's department set.
- Prefer database-side `EXISTS`, `JOIN`, or equivalent semi-join to associate target resource's owning user ID with organizational relationship table.
- For providers that cannot provide database-side constraints, returning visible user ID set as fallback is allowed; this fallback must not become the default implementation for large data volume paths.

When `org-center` is not installed, not enabled, or `orgcap` provider is unavailable:
- `all data permission` does not depend on organizational capability and continues to work normally.
- `department data permission` cannot resolve department boundaries; degrades to `self only data permission` execution.
- `self only data permission` does not depend on organizational capability and continues filtering by current user.

This is safer than "degrading to all data when org is missing" and more aligned with minimal usable experience when org module is not enabled than returning empty results directly. Frontend role management synchronously hides or disables the "department data" option to reduce non-effective configurations.

#### 22. Unified Read/Write Boundary

Each integrated resource must cover:
- List queries and tree/option queries.
- Detail queries.
- Export, download-by-id, batch info, and other data-reading paths.
- Update, delete, status toggle, password reset, forced offline, trigger task, terminate task, and other write or execution operations.

Write operations must first verify the target record is within the current caller's data scope. Batch operations must verify all targets before execution; partial success is not allowed.

#### 23. Data Permission Change Reuses Access Topology Cache Invalidation

Role `dataScope`, user-role relationships, and organizational ownership affect data permission results. The current access topology cache invalidation mechanism already triggers on role and user-role write paths. Effective role data permission should be merged into the login token access snapshot, sharing the `permission-access` revision number with roles, menus, and button permissions, avoiding repeated queries of `sys_user_role` and `sys_role` for each data permission check.

This round of cache only stores "effective data scope type" (all / department / self only / no permission) and exception status, not department user sets. Department data scope continues to inject semi-join or equivalent constraints through organizational capability at the database side, preventing large department user sets from lingering in process memory. Dynamic plugin routes and host data services should also use the same identity snapshot to pass effective data scope, avoiding plugin data access paths re-scanning role tables.

### VI. General Design Decisions

#### 24. Browser and Operating System Parsing

Obtain browser and operating system information by parsing the HTTP request header `User-Agent` field. Uses `mssola/useragent` library.

#### 25. Frontend Menu and Route Structure

```
System Monitor (/monitor)
  Operation Log (/monitor/operlog)
  Login Log (/monitor/loginlog)
  Online Users (/monitor/online)
  Server Monitor (/monitor/server)

System Info (/about)
  System API Docs (/about/api-docs)
  Version Info (/about/system-info)
  Component Demo (/about/component-demo)

System Management (/system)
  Parameter Settings (/system/config)
```

#### 26. i18n Impact Assessment

Role form does not add new text, so frontend runtime language packs typically do not need changes. If new user-visible business errors are added (e.g., "target data is outside current data permission scope"), the following is required:
- Define `bizerr.Code` in module `*_code.go`.
- Add runtime error translations for `zh-CN`, `en-US`, `zh-TW`.
- If modifying API DTO documentation, supplement non-English apidoc resources and translation completeness checks.

## Risks / Trade-offs

- **[Log Loss]** Async write may lose a small number of logs on process crash -> acceptable for admin console scenarios
- **[Storage Growth]** Operation logs record request/response parameters, which will consume significant storage over long-term operation -> provide time-range-based cleanup functionality
- **[MEMORY Engine Limitations]** MEMORY engine does not support BLOB/TEXT types, all fields must use fixed-length types -> VARCHAR is sufficient
- **[MySQL Restart Loses Sessions]** All online sessions lost after MySQL restart, users must re-login -> impact is manageable for admin console scenarios
- **[Collection Interval vs Real-time]** 30-second collection interval means frontend data has up to 30-second delay -> acceptable for monitoring scenarios
- **[iframe Embedding Limitations]** Component demo depends on external website availability -> display load failure prompt, does not affect other features
- **[Reserved Word Field Names]** `key`, `value` are MySQL reserved words -> GoFrame ORM automatically handles backtick wrapping
- **[No Caching Mechanism]** Every key-name query hits database -> parameter data volume is small, query frequency is low, caching not needed at current stage
- **[Data Permission Misses Detail/Export/Write Operations]** Lists invisible but operations allowed -> Mitigation: resource integration tasks enumerate by operation type; tests cover list, detail, export/download, and write paths
- **[Large User Sets for Department Scope]** Department scope requires organizational relationship association; large user volumes may produce long `WHERE IN` user sets -> Mitigation: user management switched to database-side semi-join from first phase; organizational relationship table supplemented with covering indexes
- **[Incorrectly Including System Governance Data]** System governance data mistakenly included in data permission causing admins unable to maintain system -> Mitigation: design explicitly lists excluded modules; specification requires unregistered resource policies not to auto-apply data permission
- **[Org Capability Disabled]** Department data permission cannot express department boundaries when organizational capability is disabled -> Mitigation: backend safely degrades to self-only scope; role add/edit form hides or disables department option
- **[Multi-role User Data Scope Merge]** Multi-role user data scope merge does not meet expectations -> Mitigation: adopt widest-range merge and write into specification; test covers `all + self`, `department + self` combinations
- **[Cache Invalidation Misses Org Ownership Changes]** Organizational capability provider must trigger data permission cache invalidation on relationship changes; first phase avoids long-term caching of department user sets when no independent cache exists

## Migration Plan

1. Create host data permission service and resource policy registration, without changing existing business behavior.
2. Integrate first-batch modules' read paths, first covering list, detail, export, download-by-id.
3. Integrate first-batch modules' write paths and execution operations.
4. Supplement backend unit/integration tests and necessary E2E tests.
5. Run `make init`, backend tests, and affected E2E tests.
6. If a module's business ownership semantics are unclear, keep it unintegrated and append "subsequent evaluation" records in design.

Rollback strategy: Data permission service integrates as internal calls, no destructive database structure changes; removing integration calls restores original behavior. If new SQL indexes or fields are added, initialization must be able to rebuild them.
