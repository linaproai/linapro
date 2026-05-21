## Why

LinaPro was originally built as a single-tenant system: `sys_*` tables have no tenant dimension, user roles are globally shared, and plugin enablement applies uniformly to all users. To evolve it into a multi-tenant full-stack framework supporting "company-internal multiple business units (BU) + platform administrator penetration", the entire stack must be reformed -- from schema, bizctx, resolution middleware, cache, JWT/session, to plugin governance.

At the same time, the existing plugin governance model ("install = enable for all users") carries excessive risk under multi-tenancy and must be split into a two-layer governance model where "platform administrators handle installation, tenant administrators handle enablement", with plugins self-protecting uninstall/disable preconditions through veto hooks.

Additionally, audit reports confirmed multiple permission boundary gaps where tenant context and resource ownership were not enforced as security boundaries alongside permission strings. The tenant switcher UI also experienced style regression after migration to plugin slots.

This iteration implements the complete multi-tenancy foundation, tenant permission boundary hardening, and tenant switcher UI stabilization as a unified user management capability.

## What Changes

### Multi-Tenancy Foundation

- **New multi-tenant capability seam (host)**: Establish stable Provider seam in `pkg/tenantcap` and `internal/service/tenantcap`; all tenant-sensitive business tables and tenant-scoped runtime state table DAOs must pass through `tenantcap.Apply` to inject `tenant_id` filtering, layered with existing `datascope`.
- **New `multi-tenant` source plugin**: Owns tenant main table, user-tenant membership table, tenant config override table; implements `tenantcap.Provider`, code-fixed resolution chain (override/JWT/session/header/subdomain/default), LifecycleGuard pre-validation, and platform management entry; tenant membership relationships unified through system user management page.
- **Schema tenant_id by responsibility**: All tenant-sensitive business tables, tenant-scoped runtime state tables, and plugin business tables add `tenant_id INT NOT NULL DEFAULT 0` (0 = PLATFORM), with corresponding query indexes upgraded to `(tenant_id, ...)` composite indexes; platform control plane tables remain globally unique without mechanical tenant_id addition.
- **bizctx adds TenantId field**: Injected from middleware layer, propagated along request chain, carried by all logs, cache, and audit.
- **JWT/session carries tenant**: Claims add `TenantId`; session store uses globally unique `token_id` to locate session while retaining `tenant_id` as ownership and validation dimension; token re-signed after login or tenant switch, old token immediately invalidated.
- **Role and user-role association tenantized**: `sys_role` adds `tenant_id`, `data_scope` expanded to `1=all data, 2=tenant data, 3=department data, 4=personal data`; `sys_user_role` adds `tenant_id`; permission/menu resolution filtered by current tenant.
- **Platform/tenant administrator capability combination model**: No platform boolean field in role table, no `platform:*`/`tenant:*` permission prefix boundaries. Platform administrator derived from `TenantId=0`, `data_scope=1` combined with unified `system:*` functional permissions; tenant administrator derived from target tenant context, tenant data permissions, and tenant-related `system:*` functional permissions. Impersonation operates as target tenant with audit marking.
- **Dict/config "platform default + tenant override"**: `sys_dict_*` and `sys_config` read path uses `(tenant_id=current) -> fallback (tenant_id=0)`; write path defaults to current tenant, platform admin can explicitly write platform layer.
- **File storage tenant isolation**: Local and object storage path prefixes separated by tenant (`/storage/t/{tenant_id}/...`).
- **Cache keys and invalidation carry tenant dimension**: `kvcache`/runtime cache keys add tenant dimension, `cluster` invalidation broadcasts carry tenant scope.
- **Plugin governance two-layer model (BREAKING)**: `plugin.yaml` adds `scope_nature` (`platform_only`/`tenant_aware`); `sys_plugin` adds `scope_nature`/`install_mode` (`global`/`tenant_scoped`); `sys_plugin_state` adds `tenant_id`; platform admin handles install/uninstall and selects `install_mode`; tenant admin handles enable/disable for `tenant_scoped` plugins.
- **Plugin lifecycle veto hooks (LifecycleGuard)**: New `CanUninstall`/`CanDisable`/`CanTenantDisable`/`CanTenantDelete` interface family; plugins self-check uninstall/disable preconditions; host aggregates multiple vetoes for unified display; supports platform admin emergency `--force` with mandatory audit.
- **Audit log full tenantization**: `monitor-operlog` and `monitor-loginlog` tables add `tenant_id` and `acting_on_behalf_of_tenant_id`; platform admin impersonation recorded in dual-track.

### Tenant Permission Boundary Hardening

- **Platform context guard**: Platform control plane APIs require platform context beyond permission strings; tenant context, acting-as-tenant context, or roles with abnormal platform permissions are rejected.
- **Assignable permissions predicate**: Role authorization tree and role create/update share the same assignable permissions set; tenant roles cannot receive platform tenant management, platform plugin governance, or global menu governance write permissions.
- **Menu write protection**: Menu create, update, delete, and status changes require platform context; tenant context can only read filtered authorization projections.
- **Plugin governance separation**: Plugin install, uninstall, enable, disable, sync, upload, install mode, and tenant provisioning policy require platform context; tenant plugin self-service continues through tenant plugin APIs.
- **Job group tenant isolation**: Job groups treated as tenant-scoped resources with list, create, update, delete, task count, and migration all limited to current tenant scope.
- **Fallback metadata**: Config and dict fallback rows return source and action metadata; frontend uses `canEdit`/`canOverride` to hide or toggle operations.

### Tenant Switcher UI Fix

- **Style stabilization**: Tenant switcher fixed width, right margin, impersonation prompt, option truncation, and dark mode styles converged to component-local scoped CSS.
- **Icon fix**: Building icon changed from Tailwind CSS utility classes to explicit `IconifyIcon` component.
- **E2E coverage**: Extended multi-tenant workbench page objects with switcher width, position, and spacing assertions.

## Capabilities

### New Capabilities

- `multi-tenancy-foundation`: Host-side multi-tenant capability seam (`tenantcap` interface, Service, bizctx integration, DAO injection discipline, Pool model schema general principles).
- `tenant-resolution`: Tenant resolution chain (override/JWT/session/header/subdomain/default), code-fixed policies and unrecognized request behavior.
- `tenant-management`: Platform administrator side tenant entity and lifecycle (create, suspend, delete).
- `tenant-membership`: User-tenant 1:N binding model (membership relationships, intra-tenant roles, status, platform/tenant administrator distinction).
- `tenant-aware-authentication`: Multi-tenant login, tenant selection, JWT tenant claim, tenant switch re-signing, platform admin impersonation.
- `tenant-config-override`: Dict/config "platform default + tenant override" read and write semantics.
- `tenant-lifecycle-events`: Tenant create/suspend/restore/delete explicit domain orchestration boundaries.
- `tenant-data-isolation`: File storage paths, cache keys, audit logs, cross-tenant operation log isolation and observability specifications.
- `tenant-platform-access-control`: Platform tenant control plane API platform context access boundary definition.
- `plugin-scope-nature`: `plugin.yaml` `scope_nature` field semantics, installation validation, immutable contract.
- `plugin-install-mode`: Platform admin installation-time `global`/`tenant_scoped` selection, mode switching rules, new tenant join strategy.
- `plugin-tenant-enablement`: Tenant admin enable/disable for `tenant_scoped` plugins, tenant-level state storage, cache invalidation.
- `plugin-lifecycle-guard`: Plugin veto hook family (`CanUninstall`/`CanDisable`/`CanTenantDisable`/`CanTenantDelete`), veto aggregation, timeout tolerance, `--force` channel, audit requirements.

### Modified Capabilities

- `user-auth`: Login flow adds tenant resolution and selection; Claims add TenantId; new tenant switch re-sign interface and token revocation rules.
- `user-management`: User identity and tenant membership relationship layering; query/create/import tenant isolation; batch edit support.
- `user-role-association`: User-role binding tenant isolation; platform context roles only bindable to platform users, tenant context roles only bindable to same-tenant active membership users.
- `role-management`: Role ownership by `tenant_id` and `data_scope` expressing data boundaries; no platform role boolean field; visibility and assignability isolated by current tenant context; assignable permissions filtering.
- `dict-management`: Read path implements platform fallback; dict type declares tenant override allowance; fallback metadata.
- `dictionary-management`: Platform dict and tenant dict visibility, write permission rules; fallback metadata.
- `config-management`: Config read path platform fallback; tenant admin only writes own tenant config; fallback metadata.
- `menu-management`: Menu governance write operations require platform context; tenant authorization tree filtered by context.
- `dept-management`: org-center plugin dept tree tenant isolation; no dependency on unimplemented tenant event bus.
- `post-management`: org-center plugin post options tenant isolation.
- `online-user`: Session storage by (tenant, token) combination; session query/kick filtered by tenant.
- `login-log`: Login log adds tenant and impersonation dual-track fields.
- `oper-log`: Operation log adds tenant and impersonation dual-track fields; force operations separately audited.
- `notice-management`: Notice tenant isolation; cross-tenant notices require platform permission.
- `user-message`: Message tenant isolation.
- `cron-job-management`: Job tenant isolation; built-in system jobs platform-level; job group tenant isolation.
- `plugin-manifest-lifecycle`: plugin.yaml adds `scope_nature`; sys_plugin adds `scope_nature`/`install_mode`; installation consistency validation; platform governance requires platform context.
- `plugin-runtime-loading`: Route global mount + request-time (tenant, plugin) enable state filtering.
- `plugin-startup-bootstrap`: Startup (plugin, tenant) dimension state cache assembly and consistency validation.
- `plugin-permission-governance`: Permissions use unified `system:*` functional namespace; platform/tenant boundaries by route plane, tenant context, data permissions, and plugin enable state; platform governance separated from tenant self-service.
- `plugin-cache-service`: Cache keys default carry tenant dimension; invalidation broadcasts tenant-scoped.
- `plugin-storage-service`: File storage paths tenant-prefixed; cross-tenant access requires explicit platform API or impersonation.
- `plugin-host-service-extension`: Host services exposed to plugins auto-forward `bizctx.TenantId` and validate tenant visibility.
- `core-host-boundary-governance`: "Tenant capability seam" listed as host stable seam (alongside orgcap).
- `module-decoupling`: Multi-tenant plugin disable/enable linked hide specification.
- `distributed-cache-coordination`: Invalidation messages and scope add `tenant_id` dimension.
- `framework-i18n-runtime-performance`: Translation package cache maintains global delivery resource dimension; dict/config tenant override cache handles tenant dimension.
- `database-bootstrap-commands`: `make init`/`make mock` defaults write PLATFORM and supports specified tenant.
- `framework-bootstrap-installer`: Startup assembly of tenantcap.Provider and validation of install_mode and scope_nature consistency.
- `dashboard-workbench`: Workbench header adds current tenant identifier and switcher; style stabilization.
- `login-page-presentation`: Login page presents tenant input/selection UI based on resolution strategy; tenant transition loading state.
- `management-workbench-i18n`: i18n text covers tenant switcher, platform admin, tenant admin and other new terms.
- `e2e-suite-organization`: e2e cases add multi-tenant groups and cross-tenant isolation matrix.
- `spec-governance`: Multi-tenant related incremental specs must explicitly declare tenant_id behavior in read/write/cache/audit four categories.
- `backend-conformance`: Adds "DAO must pass through tenantcap.Apply" hard rule.

## Impact

- **Schema**: Host tenant-sensitive `sys_*` table structures merged back to corresponding source SQL; `sys_locker`, `sys_menu`, `sys_plugin`, `sys_plugin_release`, `sys_plugin_migration`, `sys_plugin_resource_ref`, `sys_plugin_node_state`, `sys_notify_channel` and other platform control plane or global config tables do not carry `tenant_id`.
- **New plugin**: `apps/lina-plugins/multi-tenant/`, owns tenant entity, membership, config override, and platform tenant management UI; tenant membership relationships displayed and managed through host user management page.
- **Existing plugin transformation**: `org-center` all tables add `tenant_id` with tenant filtering; `monitor-loginlog`/`monitor-online`/`monitor-operlog`/`content-notice`/`demo-control`/`plugin-demo-source`/`plugin-demo-dynamic` all add `tenant_id` columns with tenant filtering; `monitor-server` remains `platform_only`.
- **Host code**: New `pkg/tenantcap` and `internal/service/tenantcap`; `bizctx` adds fields; `auth`/`session`/`role`/`dict`/`sysconfig`/`file`/`notify`/`usermsg`/`jobmgmt`/`jobhandler`/`jobmeta`/`online`/`cachecoord`/`kvcache`/`pluginruntimecache`/`plugin` fully integrate tenant context; menu directory, plugin registry directory, plugin release migration records, distributed locks, notification channel config remain platform global.
- **API**: New `/auth/login-tenants`, `/auth/switch-tenant`, `/platform/tenants/*`, `/platform/plugins/*` install_mode options, `/tenant/plugins/*` enable/disable, `/tenant/members/*` member management. All existing interfaces maintain compatibility (tenant injected from ctx, not exposed in DTO).
- **Frontend**: Login page adds tenant selector; workbench header adds tenant switcher; platform management adds "Tenant Management" page; platform and tenant admins maintain tenant membership through "User Management"; no separate "Tenant Workbench" directory, platform "Tenant Members" menu, or tenant-side member/plugin management menu.
- **Configuration**: Host `config.template.yaml` does not contain `tenant.*` section; tenant default policies defined in code.
- **Audit and observability**: Operation logs, login logs fully tenantized; platform admin impersonation dual-track recording; veto hook results and `--force` operations separately audited.
- **i18n**: New tenant management, platform admin, veto reason i18n keys, zh-CN/en-US bilingual full coverage.
- **Testing**: New e2e multi-tenant groups covering tenant isolation, platform impersonation, plugin tenant_scoped enablement, veto hooks, tenant lifecycle, login resolution strategy matrix.
- **Documentation**: `README.md` and `README.zh-CN.md` add multi-tenant chapters; plugin development guide adds `scope_nature` and `LifecycleGuard` chapters.
- **Dependencies**: No new third-party dependencies.
