## 1. Schema and Initialization

- [x] 1.1 Define `tenant_id INT NOT NULL DEFAULT 0` directly in each host table's source SQL, upgrade indexes to `(tenant_id, ...)` composite indexes (idempotent); `016` only retains explanation, not migration patches or seed
- [x] 1.2 Define `tenant_id` and four-level `data_scope` (`1=all, 2=tenant, 3=dept, 4=personal`) directly in `006-menu-role-management.sql`'s `sys_role` table; no platform role boolean field
- [x] 1.3 Define `tenant_id INT NOT NULL DEFAULT 0` directly in `006-menu-role-management.sql`'s `sys_user_role` table, primary key changed to `(user_id, role_id, tenant_id)`
- [x] 1.4 Define `allow_tenant_override BOOL NOT NULL DEFAULT FALSE` directly in `002-dictionary-management.sql`'s `sys_dict_type` table
- [x] 1.5 Define `scope_nature VARCHAR(32) NOT NULL DEFAULT 'tenant_aware'`, `install_mode VARCHAR(32) NOT NULL DEFAULT 'global'` and `auto_enable_for_new_tenants BOOL NOT NULL DEFAULT FALSE` directly in `008-plugin-framework.sql`'s `sys_plugin` table
- [x] 1.6 Define `tenant_id INT NOT NULL DEFAULT 0` directly in `009-plugin-host-call.sql`'s `sys_plugin_state` table, keep `id` auto-increment technical primary key, use `(plugin_id, tenant_id, state_key)` unique index for business uniqueness, plugin enable state uses `state_key='__tenant_enabled__'`
- [x] 1.7 Add `acting_user_id`, `on_behalf_of_tenant_id`, `is_impersonation` fields to `monitor-loginlog` and `monitor-operlog` plugin schemas
- [x] 1.8 Add `tenant_id INT NOT NULL DEFAULT 0` to all plugin business tables (`plugin_*`), upgrade indexes
- [x] 1.9 Seed data: Platform super admin role (`tenant_id=0, data_scope=1`), admin user bound to platform super admin role; existing seed dict/config unified to `tenant_id=0`
- [x] 1.10 Modify existing mock-data SQL to default write `tenant_id=0`, ensure `make mock` single-tenant out-of-box experience unchanged
- [x] 1.11 Execute `make init` to rebuild database, verify all table structures and indexes correct
- [x] 1.12 Execute `make mock` to verify mock data loads successfully and single-tenant behavior not broken

## 2. Host Stable Seam (pkg/tenantcap)

- [x] 2.1 Create `apps/lina-core/pkg/tenantcap/` package, define `Provider` interface, `TenantID` type, `PLATFORM` constant, `RegisterProvider`/`CurrentProvider`/`HasProvider` (reference pkg/orgcap shape)
- [x] 2.2 Evaluate and remove unimplemented tenant lifecycle event interfaces; `pkg/tenantcap` only retains stable Provider, Resolver and tenant projection contracts
- [x] 2.3 Define `TenantInfo` struct (id, code, name, status) and `Resolver` sub-interface in `pkg/tenantcap/`

## 3. Host-Side tenantcap Service

- [x] 3.1 Create `apps/lina-core/internal/service/tenantcap/tenantcap.go` main file: `Service` interface, `serviceImpl`, `New()`, implement `Enabled`, `Apply`, `Current`, `PlatformBypass`, `EnsureTenantVisible`
- [x] 3.2 Define shared `bizerr.Code` in `apps/lina-core/pkg/tenantcap/tenantcap_code.go` (`CodeTenantRequired`, `CodeTenantForbidden`, `CodeCrossTenantNotAllowed`, `CodePlatformPermissionRequired`, `CodeTenantSuspended`, etc.)
- [x] 3.3 Implement chain-of-responsibility dispatcher in `tenantcap_resolver_chain.go`, supporting registration of multiple `Resolver` and traversal by configured order
- [x] 3.4 Implement `ReadWithPlatformFallback` helper in `tenantcap_fallback.go` (application-layer two-query merge) for dict/config "tenant override" resources
- [x] 3.5 Supplementary unit tests (self-contained, order-independent): cover `Apply` injection, `PlatformBypass`, chain order, fallback semantics

## 4. bizctx and Middleware

- [x] 4.1 Add `TenantId int`, `ActingAsTenant bool`, `ActingUserId int`, `IsImpersonation bool` fields to `internal/service/bizctx/bizctx.go`; add `SetTenant`, `SetImpersonation` methods
- [x] 4.2 Create `internal/service/middleware/middleware_tenancy.go`, implement tenancy middleware: call `tenantcap.Provider.ResolveTenant` after auth and before permission check, write to bizctx
- [x] 4.3 Register tenancy middleware in `cmd` startup; short-circuit to inject `TenantId=0` when multi-tenant not enabled
- [x] 4.4 Unit tests: cover middleware injection path, short-circuit path, resolver chain selection

## 5. Startup Consistency Validation

- [x] 5.1 Add consistency checks in `framework-bootstrap-installer` startup flow: `scope_nature` vs `install_mode`, tenant roles cannot use `data_scope=1`, platform users have no membership, multi-tenant enabled vs Provider registration
- [x] 5.2 Print clear log and panic to prevent startup on check failure
- [x] 5.3 Integration test: Intentionally construct illegal state to verify startup failure

## 6. Multi-Tenant Plugin Skeleton (lina-plugin-multi-tenant)

- [x] 6.1 Create `apps/lina-plugins/multi-tenant/` directory with `plugin.yaml` (scope_nature=platform_only, install_mode=global), `plugin_embed.go`, `backend/`, `frontend/`, `manifest/`
- [x] 6.2 Register `_ "lina-plugin-multi-tenant/backend"` in `apps/lina-plugins/lina-plugins.go`
- [x] 6.3 Build `backend/` structure: `api/`, `internal/{controller, service, dao, model/{do, entity}}`, `hack/config.yaml`, `plugin.go`
- [x] 6.4 Create `manifest/sql/001-multi-tenant-schema.sql` (all plugin-owned tables) and `manifest/sql/uninstall/`
- [x] 6.5 Create `manifest/i18n/zh-CN/plugin.json` and `manifest/i18n/en-US/plugin.json` placeholder skeletons
- [x] 6.6 plugin.yaml lists menus (platform admin side: tenants) and hidden permission points (unified `system:tenant:*`, `system:tenant:member:*`, `system:tenant:plugin:*` etc.); tenant membership managed through user management page, no independent tenant workbench directory

## 7. Multi-Tenant Plugin Schema

- [x] 7.1 `plugin_multi_tenant_tenant`: id, code (unique), name, status, remark, created_at, updated_at, deleted_at (soft delete)
- [x] 7.2 `plugin_multi_tenant_user_membership`: id, user_id, tenant_id, status, joined_at, UNIQUE (user_id, tenant_id)
- [x] 7.3 `plugin_multi_tenant_config_override`: placeholder extension, first version may not create or only create table skeleton
- [x] 7.4 No tenant quota table or placeholder execution logic in this round; quota/billing model designed in subsequent iteration
- [x] 7.5 Delete `plugin_multi_tenant_resolver_config`: resolution chain, reserved subdomains and ambiguous behavior fixed in code
- [x] 7.6 Delete `plugin_multi_tenant_event_outbox`: do not retain outbox placeholder table lacking subscription, distribution, retry and per-subscriber state
- [x] 7.7 `manifest/sql/uninstall/001-cleanup.sql`: Clean plugin tables on uninstall (preconditions guaranteed by LifecycleGuard)
- [x] 7.8 `dao/`, `model/{do,entity}/` generated via `cd apps/lina-plugins/multi-tenant/backend && make dao`

## 8. Multi-Tenant Plugin Service Layer

- [x] 8.1 `service/tenant/`: `Service` interface, `serviceImpl`, implement tenant CRUD, state machine migration, code uniqueness validation, tombstone 30-day retention
- [x] 8.2 `service/membership/`: 1:N membership CRUD, reserve `single` strategy validation, platform admin cannot add membership
- [x] 8.3 `service/resolver/`: Implement 6 Resolvers (override/jwt/session/header/subdomain/default), support config-driven chain; formal JWT priority over header/subdomain hints
- [x] 8.4 Delete unimplemented `service/lifecycle/` event outbox path; tenant creation side effects changed to explicit domain service calls
- [x] 8.5 `service/provider/`: `tenantcap.Provider` implementation, aggregates tenant + membership service, calls `tenantcap.RegisterProvider` when plugin enabled
- [x] 8.6 `service/lifecycleguard/`: Implement `CanUninstall` (active tenants exist then false), `CanDisable`, `CanTenantDelete` (subscribe to other plugin hook aggregation)
- [x] 8.7 `service/impersonate/`: Platform admin impersonation token issuance, end impersonation, dual-track log writing
- [x] 8.8 `service/resolverconfig/`: Provide code-built-in resolution strategy query and no-op validation, reject runtime strategy changes
- [x] 8.9 Service layer unit tests self-contained coverage

## 9. Multi-Tenant Plugin Controller Layer

- [x] 9.1 `controller/platform/`: `/platform/tenants/*` (CRUD), `/platform/tenants/{id}/impersonate`, `/platform/tenant/resolver-config` strategy query/validation, `/platform/users`
- [x] 9.2 `controller/tenant/`: `/tenant/members/*` (list, invite, remove, adjust role), `/tenant/members/me`
- [x] 9.3 `controller/auth/`: `/auth/login-tenants`, `/auth/select-tenant`, `/auth/switch-tenant` (overrides host original login)
- [x] 9.4 Controller generated via `cd apps/lina-plugins/multi-tenant/backend && make ctrl`, business logic written in generated files
- [x] 9.5 All controller fields hold dependency services, initialized in NewV1 (prohibit service.New() in methods)

## 10. JWT / Session / Auth Transformation

- [x] 10.1 Add `TenantId int`, `IsImpersonation bool`, `ActingUserId int` fields to `auth.Claims`; adjust issuance and parsing logic
- [x] 10.2 Modify `auth.Service.Login` to two-phase: return pre_token + tenant list (single-tenant user compatible direct issuance)
- [x] 10.3 Implement pre_token short-term single-use mechanism (60s TTL, stored in redis/session store)
- [x] 10.4 Implement `/auth/select-tenant`: validate pre_token + membership -> issue formal JWT
- [x] 10.5 Implement `/auth/switch-tenant`: validate membership -> add old token to revoke list -> re-sign
- [x] 10.6 Implement token revoke list (local memory + cluster broadcast)
- [x] 10.7 Modify `session.Store`: primary key is globally unique `token_id`, retain `tenant_id` as session ownership and validation dimension, index covers `(tenant_id, user_id)` and `(tenant_id, login_time)`
- [x] 10.8 Modify `auth.Logout`: only revoke current (tenant, token) row
- [x] 10.9 Unit tests cover Login -> SelectTenant -> SwitchTenant -> Logout full chain

## 11. DAO Injection Discipline Implementation

- [x] 11.1 Transform user service in `internal/service/user/`: multi-tenant enabled list/detail uses membership join as visibility authoritative boundary; `sys_user.tenant_id` only primary tenant; write fills primary tenant and creates membership
- [x] 11.2 Transform all reads of `sys_role` and `sys_user_role` in `internal/service/role/`; role queries filtered by current tenant context and `tenant_id`
- [x] 11.3 Transform `internal/service/menu/`: menu remains platform global, but filter by (tenant, plugin) enable state when resolving
- [x] 11.4 Transform `internal/service/dict/` to `ReadWithPlatformFallback` mode; write to current tenant; `allow_tenant_override` validation
- [x] 11.5 Same for `internal/service/sysconfig/`
- [x] 11.6 Transform `internal/service/file/` storage path prefix `/storage/t/{tid}/...`; read validation
- [x] 11.7 Transform `internal/service/notify/` notification send/query, cross-tenant broadcast platform only
- [x] 11.8 Transform `internal/service/usermsg/` message inbox filtered by tenant
- [x] 11.9 Transform `internal/service/jobmgmt/` `jobmeta/` `jobhandler/`, task execution constructs tenant bizctx
- [x] 11.10 Transform `internal/service/session/` and `internal/service/cron/`, session cleanup and scheduled task context bound to tenant
- [x] 11.11 Adjust Apply order in `internal/service/datascope/`: tenantcap first then datascope
- [x] 11.12 Grep all host code `dao.Sys*.Ctx(ctx)` to ensure no scattered omissions

## 12. Cache and Consistency

- [x] 12.1 Transform cache keys in `internal/service/kvcache/`, `internal/service/pluginruntimecache/`, `internal/service/cachecoord/` to `(tenant_id, scope, key)` form
- [x] 12.2 Create `tenantcap.CacheKey(tenant, scope, key)` helper, all cache consumers unified call
- [x] 12.3 Invalidation message schema adds `tenant_id` field; supports `cascade_to_tenants` flag
- [x] 12.4 `cluster.Service` cluster broadcast path transparent to tenant dimension
- [x] 12.5 Translation cache (framework-i18n-runtime) key and invalidation scope transformation
- [x] 12.6 Dict cache, config cache, role cache, permission cache, menu cache all bucketed by tenant
- [x] 12.7 Unit tests: cross-tenant cache isolation anti-examples, platform default cascade invalidation

## 13. Plugin Governance Transformation

- [x] 13.1 Add `scope_nature` / `install_mode` parsing and persistence in `internal/service/plugin/`, plugin.yaml parsing validation
- [x] 13.2 Transform `IsEnabled(ctx, pluginID)` to tenant-aware (reads `(plugin_id, tenant_id)`)
- [x] 13.3 Implement install_mode switching flow (global vs tenant_scoped) and force channel
- [x] 13.4 Implement `LifecycleGuard` interface family (`pkg/pluginhost/lifecycle_guard.go`) and concurrent invocation / timeout / panic recover framework
- [x] 13.5 Call `CanUninstall` hook in `plugin.Uninstall` flow and aggregate vetoes
- [x] 13.6 Call `CanDisable` hook (global) or `CanTenantDisable` (tenant_scoped) in `plugin.Disable` flow
- [x] 13.7 Call `CanTenantDelete` hook in `tenant.Delete` flow
- [x] 13.8 Implement `--force` channel (config switch `plugin.allow_force_uninstall`) and platform audit
- [x] 13.9 Unit tests: hook aggregation, timeout fail-safe, panic recover, force channel

## 14. Route and Permission Transformation

- [x] 14.1 Transform `pluginruntimecache` / `plugin.routing` request-time filtering: return 404 by (tenant, plugin) enable state
- [x] 14.2 Transform `permission` middleware: permission resolution filters disabled plugins by current tenant
- [x] 14.3 Remove permission point platform/tenant prefix constraints, unified use of `system:*`; platform/tenant boundaries constrained by route plane, tenant context, data permissions, and plugin enable state
- [x] 14.4 Unit tests: cover route 404 / menu hide / permission point filtering

## 15. org-center Plugin Tenantization

- [x] 15.1 Modify `apps/lina-plugins/org-center/manifest/sql/001-org-center-schema.sql`, all tables add `tenant_id`, indexes upgraded
- [x] 15.2 Modify mock-data/*.sql, default write `tenant_id=0` (single-tenant experience unchanged)
- [x] 15.3 Modify `org-center` plugin.yaml: `scope_nature: tenant_aware`, `default_install_mode: global`
- [x] 15.4 Transform service/dao implementation: all queries use `tenantcap.Apply`, all writes fill `tenant_id`
- [x] 15.5 Remove unimplemented `tenantcap.LifecycleSubscriber` placeholder contract
- [x] 15.6 Transform `orgcap.Provider` implementation to filter by `bizctx.TenantId`
- [x] 15.7 Unit tests self-contained, cover tenant isolation, event subscription idempotent, cascade cleanup

## 16. Other Existing Plugin Tenantization

- [x] 16.1 `monitor-loginlog`: Table adds `tenant_id`, `acting_user_id`, `on_behalf_of_tenant_id`, `is_impersonation`; dao/service/controller transformation; plugin.yaml `scope_nature: tenant_aware, default_install_mode: tenant_scoped`
- [x] 16.2 `monitor-online`: Table adds `tenant_id`, session query/kick adds tenant validation; plugin.yaml same
- [x] 16.3 `monitor-operlog`: Table adds tenant and impersonation fields; controller supports `operType` filter; plugin.yaml same
- [x] 16.4 `monitor-server`: plugin.yaml `scope_nature: platform_only`; not visible to tenant admin view
- [x] 16.5 `content-notice`: Table adds `tenant_id`, notice filtered by tenant; plugin.yaml `scope_nature: tenant_aware, default_install_mode: tenant_scoped`
- [x] 16.6 `demo-control`, `plugin-demo-source`, `plugin-demo-dynamic`: Add tenant fields (if persistent); plugin.yaml `scope_nature: tenant_aware`
- [x] 16.7 Each plugin i18n resources add `plugin.<id>.uninstall_blocked.*` and other reason key translations

## 17. Backend API and DTO

- [x] 17.1 All platform API DTOs defined in `apps/lina-plugins/multi-tenant/backend/api/platform/v1/`, English dc + eg
- [x] 17.2 All tenant API DTOs defined in `apps/lina-plugins/multi-tenant/backend/api/tenant/v1/`
- [x] 17.3 Declare `permission` tag on g.Meta, platform and tenant APIs both use unified `system:*` permission identifiers
- [x] 17.4 apidoc i18n: multi-tenant plugin maintains own `manifest/i18n/<locale>/apidoc/**/*.json`, English empty, zh-CN provides translation
- [x] 17.5 Error codes `bizerr.Code*` defined centrally in `internal/service/<module>/<module>_code.go`
- [x] 17.6 Unit tests: DTO field validation, permission validation, i18n completeness

## 18. Frontend Workbench Transformation

- [x] 18.1 Add platform/tenant/auth new interface clients in `apps/lina-vben/apps/web-antd/src/api/`
- [x] 18.2 Add `platform/tenants/`, `platform/users/` pages in `src/views/`, no runtime config page for resolution strategy
- [x] 18.3 Tenant admin entry reclaimed to user management page, no `tenant/members/`, `tenant/plugins/` left menu entry
- [x] 18.4 Transform login page `views/login/` to support tenant selector (based on resolution strategy)
- [x] 18.5 Workbench header adds tenant identifier + switcher (platform vs tenant visual distinction)
- [x] 18.6 Transform route guard, link-hide tenant UI based on multi-tenant enable state
- [x] 18.7 Add impersonation header red bar prompt and "exit impersonation" button
- [x] 18.8 Add `views/platform/plugins/install-mode-selector.vue` install-time install_mode selection dialog
- [x] 18.9 Add LifecycleGuard veto reason display dialog (supports multi-reason aggregation + i18n rendering + force second confirmation)
- [x] 18.10 Route modules `src/router/routes/modules/platform.ts`, `tenant.ts` registration
- [x] 18.11 Global Pinia store adds `useTenantStore` managing current tenant, selectable tenant list, impersonation state

## 19. Frontend i18n Resources

- [x] 19.1 Add multi-tenant related keys in `apps/lina-vben/apps/web-antd/src/locales/zh-CN.json` and `en-US.json` (menus, forms, errors, veto reasons, impersonation prompts)
- [x] 19.2 Platform multi-tenant menu i18n key namespace retained `menu.platform.*`, remove independent tenant workbench `menu.tenant.*` top-level menu namespace
- [x] 19.3 i18n completeness validation script (CI) blocks omissions

## 20. Configuration and Documentation

- [x] 20.1 Tenant default policies defined centrally in code; host config template does not provide `tenant.*`, no runtime resolution config table created or read
- [x] 20.2 `plugin.allow_force_uninstall: true` config item and documentation
- [x] 20.3 Update root `README.md` and `README.zh-CN.md` add multi-tenant chapters
- [x] 20.4 Add plugin author guide in `docs/`: scope_nature, install_mode, LifecycleGuard, tenant event subscription
- [x] 20.5 `apps/lina-plugins/multi-tenant/README.md` and `README.zh-CN.md` bilingual mirrors

## 21. Platform Context Guard (Permission Boundary Hardening)

- [x] 21.1 Add or reuse platform context guard, cover platform tenant interfaces, menu governance write operations, plugin platform governance write operations, explicitly reject acting-as-tenant context
- [x] 21.2 Integrate platform context validation in `apps/lina-plugins/multi-tenant/backend/internal/service/tenant` platform tenant list, detail, create, update, delete, enable/disable and impersonation entries
- [x] 21.3 Integrate platform context validation in `apps/lina-core/internal/service/menu` create, update, delete, status/visibility change paths, maintain permission topology revision publish failure-close
- [x] 21.4 Integrate platform context validation in `apps/lina-core/internal/service/plugin` sync, upload, install, uninstall, enable, disable, upgrade, install mode and tenant provisioning policy write paths

## 22. Role Authorization Assignable Set

- [x] 22.1 Implement centralized assignable menu/permission predicate, determine platform-only permissions by platform context, tenant context, `menu_key`, permission string and plugin governance metadata
- [x] 22.2 Update role menu tree query, tenant context filters platform tenant management, platform plugin governance and global menu governance write permissions
- [x] 22.3 Update role create and update, validate all `menuIds` belong to current context assignable set before writing `sys_role_menu`
- [x] 22.4 Confirm abnormal historical platform-only authorization cannot bypass platform control plane guard, filter tenant-state checked keys if necessary

## 23. Job Group Tenant Isolation

- [x] 23.1 Update job group list, detail and task count queries, filter by current tenant at database query stage
- [x] 23.2 Update job group create, explicitly write current `tenant_id` and validate uniqueness by `(tenant_id, code)`
- [x] 23.3 Update job group update and delete, validate target group belongs to current tenant before operation
- [x] 23.4 Update delete group migration logic, only migrate current tenant's tasks to current tenant's default group
- [x] 23.5 Inventory and adjust seed/mock job group data as needed, ensure tenant demo data does not depend on `tenant_id=0` global groups

## 24. Fallback Metadata and Frontend Visibility

- [x] 24.1 Update config API DTO, service projection and frontend types, return `sourceTenantId`, `isFallback`, `canEdit`, `canOverride`, `overrideMode`
- [x] 24.2 Update dict type and dict data API DTO, service projection and frontend types, return same fallback source and action metadata
- [x] 24.3 Update parameter settings page, hide fallback row direct edit/delete entries by action metadata, avoid must-fail detail requests
- [x] 24.4 Update dict type and dict data pages, hide fallback row direct edit/delete entries by action metadata, avoid must-fail detail requests
- [x] 24.5 If implementing tenant override creation entry in this phase, supplement clear override creation flow; if not implementing, retain `canOverride` metadata but do not display direct override button

## 25. Tenant Switcher UI Fix

- [x] 25.1 Converge tenant switcher fixed width, right margin, impersonation prompt, option truncation, dark mode styles to component-local scoped CSS
- [x] 25.2 Change building icon from Tailwind CSS utility classes to explicit `IconifyIcon` component
- [x] 25.3 Extend multi-tenant workbench E2E page objects with switcher width, position, and spacing assertions

## 26. i18n, apidoc and Error Governance

- [x] 26.1 For platform context missing, role menu assignment forbidden, job group tenant invisible and other new or reused stable `bizerr` error codes
- [x] 26.2 Supplement host and multi-tenant plugin `manifest/i18n/{zh-CN,en-US}/error.json` error translations
- [x] 26.3 Supplement related apidoc i18n JSON, ensure new errors and new response fields have bilingual descriptions
- [x] 26.4 Supplement frontend runtime language pack with fallback read-only, inherit platform default, create tenant override visible text; this phase does not add fallback visible text, only hides direct edit/delete entries and retains `canOverride` metadata

## 27. Unit Tests

- [x] 27.1 tenantcap unit tests self-contained
- [x] 27.2 bizctx unit tests
- [x] 27.3 tenancy middleware unit tests
- [x] 27.4 Resolution chain unit tests (one case set per resolver)
- [x] 27.5 LifecycleGuard hook framework unit tests (aggregation, timeout, panic, force)
- [x] 27.6 Multi-tenant plugin service layer unit tests self-contained, cover all branches
- [x] 27.7 org-center transformation unit tests self-contained
- [x] 27.8 Cache isolation unit tests (cluster mode mock)
- [x] 27.9 Platform context guard unit tests: tenant context even holding abnormal `system:tenant:*` permissions cannot access `/platform/tenants`
- [x] 27.10 Role authorization unit tests: tenant role authorization tree does not return platform nodes, role create/submit platform-only `menuIds` rejected
- [x] 27.11 Menu governance unit tests: tenant context cannot create, update, delete global menus
- [x] 27.12 Plugin governance unit tests: tenant context cannot update plugin tenant provisioning policy or execute platform plugin lifecycle governance actions
- [x] 27.13 Job group unit tests: list, create, update, delete, task count and delete migration all isolated by `tenant_id`
- [x] 27.14 Config and dict fallback unit tests: fallback rows return correct source and action metadata

## 28. E2E Tests -- Multi-Tenant Foundation

- [x] 28.1 Create `hack/tests/e2e/multi-tenant/` directory and fixtures (`multi-tenant-disabled` and `multi-tenant-enabled`)
- [x] 28.2 TC0178 Multi-tenant enabled: Platform admin installs multi-tenant plugin, system behavior correct after enable
- [x] 28.3 TC0179 Platform admin creates tenant: Basic CRUD, code validation, tombstone
- [x] 28.4 TC0180 Tenant suspend/resume: Suspended tenant users rejected on login, resume restores
- [x] 28.5 TC0181 Tenant management page does not expose archive entry
- [x] 28.6 TC0182 Tenant delete protected by LifecycleGuard: guard rejects when blocking, allows after passing
- [x] 28.7 TC0183 Multi-tenant disabled: Uninstall multi-tenant degrades to single-tenant behavior
- [x] 28.8 TC0184 1:N user login tenant selection: Returns pre_token + list + select-tenant flow
- [x] 28.9 TC0185 Tenant switch re-sign token: Old token immediately invalidated
- [x] 28.10 TC0186 Platform admin impersonation start and exit

## 29. E2E Tests -- Cross-Tenant Isolation Matrix

- [x] 29.1 TC0187 User cross-tenant isolation: Tenant A cannot see B users
- [x] 29.2 TC0188 Role cross-tenant isolation + platform roles only visible to platform admin
- [x] 29.3 TC0189 Dict cross-tenant isolation + platform fallback correct
- [x] 29.4 TC0190 Config cross-tenant isolation + platform fallback correct
- [x] 29.5 TC0191 File cross-tenant isolation + platform shared path correct
- [x] 29.6 TC0192 Notice cross-tenant isolation + platform broadcast correct
- [x] 29.7 TC0193 Job cross-tenant isolation + built-in system jobs platform-level
- [x] 29.8 TC0194 Online session cross-tenant isolation + kick validated by tenant
- [x] 29.9 TC0195 Login log/oper log cross-tenant isolation + impersonation dual-track marking
- [x] 29.10 TC0196 Dept (org-center) cross-tenant isolation
- [x] 29.11 TC0197 Post (org-center) cross-tenant isolation + same code cross-tenant allowed

## 30. E2E Tests -- Resolution Strategy and Login Flow

- [x] 30.1 TC0198 Header resolver: Pre-login X-Tenant-Code hint hits, authenticated business request cannot override JWT TenantId
- [x] 30.2 TC0199 Subdomain resolver: Pre-login subdomain hint hits + reserved subdomains ignored
- [x] 30.3 TC0200 JWT resolver: Claims hit
- [x] 30.4 TC0201 Session resolver: Post-login selection + persistent hit
- [x] 30.5 TC0202 Default resolver + ambiguous prompt mode
- [x] 30.6 TC0203 Fixed prompt ambiguity strategy, reject runtime switch
- [x] 30.7 TC0204 Override: Platform admin X-Tenant-Override valid + normal user ignored
- [x] 30.8 TC0205 Resolution chain fixed strategy, built-in strategy no-op write does not change runtime state

## 31. E2E Tests -- Plugin Governance Two-Layer

- [x] 31.1 TC0206 Platform admin installs tenant_aware plugin: install_mode selection global / tenant_scoped
- [x] 31.2 TC0207 Install platform_only plugin: install_mode forced global, tenant admin cannot see
- [x] 31.3 TC0208 Tenant admin enable/disable tenant_scoped plugin: menu/route/permission linked
- [x] 31.4 TC0209 install_mode switch global vs tenant_scoped: state correctly migrated
- [x] 31.5 TC0210 LifecycleGuard vetoes uninstall: active tenants block multi-tenant uninstall, aggregated reason
- [x] 31.6 TC0211 Force channel: Platform admin force uninstall + dual confirmation + mandatory audit
- [x] 31.7 TC0212 Hook timeout / panic fail-safe
- [x] 31.8 TC0213 Platform new tenant auto-enable policy writes tenant plugin state on new tenant creation

## 32. E2E Tests -- Platform Admin Scenarios

- [x] 32.1 TC0214 Platform admin can cross-tenant read full data
- [x] 32.2 TC0215 Platform admin impersonation operation logs dual-track recorded
- [x] 32.3 TC0216 Platform admin force operations separately audited
- [x] 32.4 TC0217 Platform admin view vs tenant admin view UI differentiation

## 33. E2E Tests -- Permission Boundary Hardening

- [x] 33.1 TC0239 Tenant role platform permission blocked: tenant role authorization tree hides platform nodes, submit platform menuIds rejected, abnormal platform permissions cannot access platform tenant interfaces
- [x] 33.2 TC0240 Tenant governance actions hidden: tenant-state menu management and plugin management platform governance actions hidden and backend rejects direct calls
- [x] 33.3 TC0241 Tenant job group isolation: tenant job groups only show current tenant, create writes current tenant, cannot update or delete out-of-scope groups
- [x] 33.4 TC0242 Config dict fallback readonly: parameter and dict fallback rows do not show direct edit entry and do not trigger not found detail requests

## 34. Cluster Consistency E2E

- [x] 34.1 TC0218 Reject runtime resolution strategy changes, no tenant-resolution shared revision created
- [x] 34.2 TC0219 Cross-node cache isolated by tenant invalidation
- [x] 34.3 TC0220 Token revoke cross-node broadcast

## 35. Performance and Regression

- [x] 35.1 Composite index performance verification: PG `EXPLAIN` verifies key queries (user list, role list, menu resolution, dict fallback)
- [x] 35.2 Existing single-tenant e2e suite passes under multi-tenant disabled fixture
- [x] 35.3 Startup consistency validation passes baseline test

## 36. Review and Archive Preparation

- [x] 36.1 Call `/lina-review` for spec and code review, process review items
- [x] 36.2 i18n completeness validation (zh-CN / en-US) no omissions
- [x] 36.3 Documentation bilingual mirror sync (README, AGENTS, SKILL, etc.)
- [x] 36.4 `openspec validate user-management --strict` passes
