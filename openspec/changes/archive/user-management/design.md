## Context

LinaPro was originally architected with "single tenant + single administrator" as the premise: `sys_user`, `sys_role`, and all `sys_*` business tables have no tenant dimension; `bizctx` only carries user identity and locale; data permissions are layered through `datascope` (role data_scope) and `orgcap` (department tree); plugin governance is single-layer (platform admin install + enable, affects everyone).

To support "company-internal multiple BUs (initial phase) + future SaaS hard isolation" multi-tenant form, the core constraints are:

1. **No phasing, complete landing in one iteration**: Project has no legacy burden, Pool model one-shot is cheaper than phased.
2. **Maximize pluginization**: Tenant entity, membership, lifecycle and other high-change-frequency parts are held by `multi-tenant` plugin; host only retains the thinnest capability seam.
3. **Maintain single-tenant out-of-box usability**: When `multi-tenant` plugin is not installed, system behavior is equivalent to today (tenant-sensitive data treated as `tenant_id=0` PLATFORM tenant, platform control plane data remains global).
4. **Reuse existing "stable capability seam" pattern**: `pkg/orgcap` + `internal/service/orgcap` + plugin Provider three-segment paradigm has been verified; this design directly reuses its shape.

stakeholders:
- Framework core developers (this repository maintainers)
- Framework operator administrators (platform administrators) and tenant administrators (downstream)
- Plugin authors (need to understand `scope_nature` and `LifecycleGuard` contracts)
- AI collaborative development flow (need OpenSpec governance to ensure spec consistency)

## Goals / Non-Goals

**Goals:**

- Provide complete multi-tenant capability (schema, resolution, JWT/session, cache, file, audit, plugin governance, UI) without breaking single-tenant out-of-box experience.
- Encapsulate all variable policies (whether to enable multi-tenancy, how to resolve tenant, tenant entity management, user-tenant binding, tenant config override, tenant lifecycle events) in `multi-tenant` plugin.
- Upgrade "plugin governance" to "platform admin install + tenant admin enable" two-layer model, and through `scope_nature` + `install_mode` explicitly express plugin scope.
- Migrate "uninstall/disable pre-check" from host hardcoding to plugin self-check (`LifecycleGuard` veto hooks), allowing platform admin emergency `--force` with mandatory audit.
- Provide platform admin vs tenant admin capability combination, cross-tenant impersonation and observability under 1:N membership model.
- Provide unified fallback read path helper for dict, config and other "platform default + tenant override" semantic resources.
- Fully integrate i18n (zh-CN / en-US), data permissions (datascope), cache consistency (cluster-aware invalidation), audit (loginlog/operlog dual-track).

**Non-Goals:**

- Not implement schema-per-tenant or DB-per-tenant isolation mode; first version isolation mode fixed by code default as `pool`, not exposed in host config file.
- Not implement "batch platform panel for enabling plugins per tenant" advanced form; tenant admin single-tenant enable/disable is sufficient, platform admin batch operations left for future extension.
- Not implement "tenant-level quota/billing" logic; no quota table or placeholder execution logic created in this round.
- Not implement "tenant-level brand customization" (logo, color scheme, domain SSL, etc.); only leave schema extension position for `tenant_branding`.
- Not introduce distributed transactions, cross-tenant data migration, tenant merge/split or other complex operations.
- Not break existing "single-tenant out-of-box experience": When multi-tenant plugin not installed, all tenant-sensitive table `tenant_id` defaults to 0, platform control plane data remains global, all filtering logic no-op, all existing e2e cases pass without modification.

## Decisions

### 1. Multi-Tenancy Foundation

#### 1.1 Isolation Model: Pool (Single DB + tenant_id Column)

**Choice**: Tenant-sensitive `sys_*` business tables, tenant-scoped runtime state tables and plugin business tables add `tenant_id INT NOT NULL DEFAULT 0`, indexes upgraded to `(tenant_id, ...)` composite indexes. Platform control plane tables remain globally unique, no mechanical tenant field addition. `sys_user` in 1:N mode uses membership as user visibility authoritative boundary, `sys_user.tenant_id` only represents primary tenant/default login tenant.

**Table Responsibility Classification**:
- Tenant-sensitive business tables: Users, roles, dicts, configs, files, sessions, notify messages/deliveries, jobs/job logs, cache KV, cache revisions, and plugin business data tables must carry `tenant_id`.
- Tenant-scoped runtime state tables: `sys_plugin_state` must carry `tenant_id` for expressing platform-level or tenant-level plugin enable state and plugin state KV.
- Platform control plane/global config tables: `sys_locker`, `sys_menu`, `sys_plugin`, `sys_plugin_release`, `sys_plugin_migration`, `sys_plugin_resource_ref`, `sys_plugin_node_state`, `sys_notify_channel` do not carry `tenant_id`; their tenant differences are expressed through `sys_plugin_state`, role/menu assignment, notify messages/deliveries, or dedicated business tables.

**Rationale**:
- Self-consistent with "plugin-driven" goal (other modes require host-level connection routing, cannot be pluginized).
- Pool is Q1 soft isolation standard answer, and PG's (tenant_id, x) composite index performance is stable.
- Future upgrade to schema-per-tenant: add `SET search_path` through middleware layer, business code unchanged.

**Default Value**: Code constant `DefaultIsolationMode = "pool"`. Host `config.template.yaml` does not expose `tenant.isolation.mode`.

#### 1.2 User-Tenant Cardinality: 1:N Membership (Default), Preserve 1:1 Strategy Extension

**Choice**: Global identity (`sys_user.username` globally unique) + `plugin_multi_tenant_user_membership` many-to-many binding. Code default `DefaultCardinality = "multi"`, allowing one user to have multiple active memberships; `single` as optional strategy for controlled management settings preserved, not exposed through host config file in current version.

**Rationale**:
- Internal BU scenario "same employee serving multiple BUs" is common.
- 1:N is true superset of 1:1; `single` mode simply rejects binding second membership.
- Slack/Notion/GitHub Orgs/Atlassian all adopt this model; Salesforce historical 1:1 is industry counter-example.

#### 1.3 Platform Admin vs Tenant Admin Capability Combination

**Choice**: `sys_role` adds `tenant_id INT`, `data_scope` expanded to `1=all data, 2=tenant data, 3=department data, 4=personal data`; `sys_user_role` adds `tenant_id INT`, UNIQUE `(user_id, role_id, tenant_id)`. No platform role boolean field maintained.

**Semantics**:
- `tenant_id=0` roles belong to platform context, can configure `data_scope=1` for platform global data permissions; whether can execute platform operations still depends on unified `system:*` functional permissions.
- `tenant_id>0` roles belong to tenant context, only visible and assignable within owning tenant; tenant roles prohibit `data_scope=1`, can only use `2`, `3` or `4`.
- Permission points only express functional actions, unified use of `system:*`; platform/tenant boundaries constrained by API route plane, current tenant context, data permissions, and plugin enable state.

#### 1.4 bizctx and Tenant Resolution Middleware

**Choice**: `bizctx.Context` adds `TenantId int` field; new `tenancy` middleware positioned after `auth`, before business processing.

**Resolution Chain** (code-fixed order: `[override, jwt, session, header, subdomain, default]`):
- `override`: X-Tenant-Override (platform admin only)
- `jwt`: Claims.TenantId
- `session`: Session persisted current tenant
- `header`: X-Tenant-Code (pre-login/ no formal JWT requests as tenant hint)
- `subdomain`: {code}.{root_domain} (pre-login/ no formal JWT requests as tenant hint; rootDomain currently fixed empty, subdomain resolution disabled)
- `default`: User's primary tenant (tenant_id=0 or membership first)

**Unrecognized Request Behavior** (`on_ambiguous`):
- `prompt` (code-fixed): Returns `TENANT_REQUIRED` error code, frontend shows selector.

#### 1.5 tenantcap Seam and DAO Injection Discipline

**New**:
- `pkg/tenantcap/`: Provider interface, registration functions, TenantID type, PLATFORM constant.
- `internal/service/tenantcap/`: `Service` interface, `Apply(model, col)`, `Current(ctx)`, `Enabled(ctx)`, `EnsureTenantVisible(ctx, tenantID)`, `PlatformBypass(ctx)`.

**DAO Injection Discipline**:
- All host service reading tenant-sensitive `sys_*` tables, tenant-scoped runtime state tables, or plugin business tables with `tenant_id` must be wrapped through `tenantcap.Apply(ctx, model, "tenant_id")`; platform control plane tables not injected through tenantcap, but only accessible through platform/host governance service.
- When layered with `datascope.ApplyUserScope`, tenantcap first then datascope (tenant isolation highest priority).
- `Apply` is no-op when multi-tenant not enabled, injects `WHERE tenant_id = ?` when enabled.

#### 1.6 JWT/Session Tenant Binding

**Choice**:
- JWT Claims add `TenantId int` (0 = platform).
- Session storage uses globally unique `token_id` as primary key, `tenant_id` only as session ownership, list filtering, data permissions, and request-time claim/session consistency validation dimension.
- Tenant switch: call `/auth/switch-tenant`, server validates membership -> deletes old session -> re-signs token -> returns new token. Old token added to revoke list (short-term + cluster broadcast).

#### 1.7 Dict/Config Platform Fallback

**Read Path** (encapsulated in `tenantcap.ReadWithPlatformFallback`):
1. First query `tenant_id = current_tenant`, if result found return directly.
2. Otherwise fallback to `tenant_id = 0` (PLATFORM default), if found return.
3. Both empty return empty.

**Write Path**:
- Default writes `bizctx.TenantId` (tenant admin can only write own tenant).
- Platform admin explicitly calls `WritePlatformDefault(...)` to write `tenant_id=0`.
- Dict type "whether allow tenant override" controlled by `sys_dict_type.allow_tenant_override BOOL`; false means only platform admin can modify, all tenants share.

### 2. Plugin Governance Two-Layer Model

#### 2.1 scope_nature Field

**plugin.yaml new field**:
```yaml
id: <plugin-id>
scope_nature: platform_only | tenant_aware  # required
default_install_mode: global | tenant_scoped # optional, effective when scope_nature=tenant_aware, default tenant_scoped
```

**Semantics**:
- `platform_only`: Plugin only runs at platform level, affects all tenants. Tenant admin MUST NOT see or control its enablement. `install_mode` forced to `global`.
- `tenant_aware`: Plugin supports platform-level or tenant-level operation. `install_mode` selected by platform admin at install time as `global` or `tenant_scoped`.

#### 2.2 install_mode Field

`sys_plugin.install_mode` values:
- `global`: Plugin effective for all tenants, enable/disable controlled by platform admin (`sys_plugin_state.tenant_id=0 AND state_key='__tenant_enabled__'` single row).
- `tenant_scoped`: Plugin enabled per tenant independently, enable/disable controlled by each tenant admin (`sys_plugin_state.tenant_id>0 AND state_key='__tenant_enabled__'` multiple rows).

**install_mode Switching Rules**:
| Switch Direction | Allowed | Processing |
|---|---|---|
| `global -> tenant_scoped` | Yes | Auto-initialize `enabled=current global state` for all active tenants |
| `tenant_scoped -> global` | Warning | Platform admin only; second confirmation + forced audit; immediately force-enable for all tenants |
| `scope_nature` change | No | Only through plugin version upgrade with manifest change, requires migration script |

#### 2.3 LifecycleGuard Veto Hooks

**Contract** (`pkg/pluginhost/lifecycle_guard.go`):
- `CanUninstaller.CanUninstall(ctx) (ok bool, reason string, err error)`
- `CanDisabler.CanDisable(ctx) (ok, reason, err)`
- `CanTenantDisabler.CanTenantDisable(ctx, tenantID int) (ok, reason, err)`
- `CanTenantDeleter.CanTenantDelete(ctx, tenantID int) (ok, reason, err)`

**Invocation and Aggregation**:
- Host before executing uninstall/disable/tenant delete, enumerates all installed plugins -> type asserts each `CanXxx` interface -> concurrent invocation (timeout 5s/hook).
- Any returns `ok=false` -> overall reject; **still executes all hooks** to collect reason set (no short-circuit).
- `reason` must be i18n key; frontend renders by locale.

**`--force` Channel**:
- Platform admin can select "force uninstall" in UI, must text-input plugin ID for second confirmation.
- Force operations separately write audit log, recording all bypassed reasons and triggering user.
- Config `plugin.allow_force_uninstall: true|false` (default true, strict compliance scenarios can disable).

### 3. Tenant Data Isolation

#### 3.1 File Storage Path Tenant Prefix

- Local storage: file path prefix `/storage/t/{tenant_id}/yyyy/mm/dd/...`, `tenant_id=0` for platform shared.
- Object storage (if enabled): object key prefix `t/{tenant_id}/...`.
- Cross-tenant reference: not allowed; if sharing needed (e.g., platform logo) explicitly write `tenant_id=0`.
- File read interface: auth validates file `tenant_id` matches `bizctx.TenantId`; platform admin in management platform mode can only cross-tenant access through explicit `/platform/*` read-only interface.

#### 3.2 Cache Keys Carry Tenant Dimension

All runtime cache keys in tenant-sensitive scenarios SHALL carry `tenant_id` dimension; cache key construction unified through helper `CacheKey(tenant, scope, key)`, prohibit scattered concatenation.

#### 3.3 Cluster Invalidation Broadcast Carries Tenant Scope

`distributed-cache-coordination` invalidation messages SHALL carry `tenant_id` field; invalidation must explicitly declare scope:
- Single tenant: `tenant_id = T` (only invalidate that tenant's scope cache).
- Platform default change: `tenant_id = 0` + `cascade_to_tenants = true` (invalidate all tenants' corresponding scope).
- Global fallback (rare): `tenant_id = -1` (explicit "all tenants clear" flag, must have audit log).

#### 3.4 Audit Log Tenant and Impersonation Dual-Track

`monitor-operlog` and `monitor-loginlog` SHALL contain:
- `tenant_id`: Operation tenant context (= `bizctx.TenantId`).
- `acting_user_id`: Actual operator's global user_id.
- `on_behalf_of_tenant_id`: Only has value during impersonation, = platform admin's target tenant.
- `is_impersonation`: Boolean flag.

### 4. Platform Access Control

#### 4.1 Platform Context Guard

Platform control plane API authorization conditions:
```
authenticated
+ permission string
+ current tenant context is platform
+ current request is not acting-as-tenant / impersonation
+ effective data scope allows platform bypass when cross-tenant data is read
```

Implementation prioritizes reusing `tenantcap.Service.PlatformBypass(ctx)` or equivalent existing context helper. If a path only needs to check platform context rather than cross-tenant data bypass, it must also explicitly reject `ActingAsTenant=true` acting-as-tenant state.

#### 4.2 Assignable Permissions Predicate

Role authorization tree filtering and `Role.Create/Update` `menuIds` validation must call the same assignable permissions predicate. Under tenant context, this set excludes:
- Platform tenant management and platform control plane menu/buttons.
- Global menu governance write permissions.
- Plugin install, uninstall, enable, disable, sync, upload, install mode, new tenant auto-enable policy and other platform plugin governance permissions.
- Other permissions marked as platform-only through `menu_key`, permission string namespace, plugin manifest metadata, or stable platform directory.

#### 4.3 Menu Write Protection

`sys_menu` is currently the global permission topology and route governance model. Menu create, update, delete, status change and other write operations must require platform context, and after success continue to publish permission topology revision number. Tenant context can read current context visible authorization projection, but cannot modify global table.

### 5. Tenant UI

#### 5.1 Workbench Header Tenant Identifier and Switcher

Workbench header SHALL display current tenant identifier (tenant name + code) when multi-tenant enabled; 1:N user's identifier right side adds switch dropdown showing all active membership tenants. Top tenant switch dropdown SHALL maintain compact style consistent with host header controls: fixed width, single-line truncation, building icon, positioned left of global search entry, with stable spacing from search entry.

#### 5.2 Platform Admin Special Header Style

Platform admin view (`bizctx.TenantId=0`) SHALL display prominent "Platform Administrator" identifier (color bar/label) in header, visually distinct from tenant view; impersonation mode shows "Acting as Tenant X" prompt bar.

#### 5.3 Login Page Presentation

Login page SHALL present different UI based on current effective resolution strategy:
- subdomain strategy: No tenant input box (auto-resolved from URL).
- header strategy: Show tenant code input box (sent as pre-login `X-Tenant-Code` hint).
- jwt/session/default strategy: No tenant input, after login with username/password, select-tenant decides.

When 1:N user login returns pre_token + tenant list, frontend switches login form content area to tenant dropdown selector (showing code + name); account password login form and tenant selector must not display simultaneously.

#### 5.4 Tenant Transition Loading State

After selecting tenant, prioritize showing "entering tenant" transition interface with loading indicator, avoiding account/password form briefly re-appearing during token exchange, user info loading, and route navigation.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| **Massive changes, one-shot landing prone to regression** | Strict e2e isolation cases + datascope existing query filtering precedent + tenancy bypass unified helper (avoid scattered judgment); tenant-sensitive table `tenant_id` defaults 0 + Apply no-op ensures behavior completely equivalent when multi-tenant not enabled |
| **DAO injection discipline omission leads to cross-tenant data leak** | Hard rules through `backend-conformance` spec + `lina-review` audit checklist "DAO must pass tenantcap.Apply" check; e2e isolation cases cover each query interface's anti-example |
| **Cache key missing tenant dimension causes dirty read** | Centralized cache key construction in helper functions; cases verify "tenant A switch to B sees A's old data" anti-example |
| **JWT tenant switch causes race condition** | Old token immediately written to revoke list + cluster broadcast; client switch flow requires "first call switch-tenant get new token, then use new token request" |
| **`scope_nature` mis-label causes plugin misplacement** | Install-time validation: platform_only disallows install_mode=tenant_scoped; tenant admin cannot see platform_only plugins even if enabled |
| **LifecycleGuard hook timeout blocks UI** | 5s hook timeout + fail-safe (timeout = ok=false); concurrent execution of all hooks |
| **Platform admin `--force` abuse** | Mandatory audit + second confirmation input plugin ID + config switch can disable |
| **Tenant resolution strategy mis-change causes site-wide 401** | First version resolution strategy fixed in code, no runtime modification or DB override |
| **org-center transformation affects existing dept tree cases** | Mock data writes PLATFORM(0), single-tenant scenario dept tree behavior equivalent to today |
| **Tenant suspend/delete plugin data residual** | Through `CanTenantDelete` hooks requiring all plugins holding tenant data to participate in confirmation |
| **Multi-tenant performance degradation** | All original indexes upgraded to `(tenant_id, ...)`; PG `EXPLAIN` verifies key queries |

## Migration Plan

Project has no legacy burden, adopts **reset database + initialize by new schema**:

1. **Host SQL source files directly define final schema**: Add `tenant_id` directly to each tenant-sensitive table's source SQL; platform control plane tables remain globally unique.
2. **New plugin SQL** `apps/lina-plugins/multi-tenant/manifest/sql/001-multi-tenant-schema.sql`.
3. **org-center plugin SQL change** `apps/lina-plugins/org-center/manifest/sql/001-org-center-schema.sql` same iteration modification.
4. **Execute `make init`**: Rebuild database; execute `make mock` to load demo data.
5. **Code deployment**: Core changes like `tenantcap` and middleware go live with version.
6. **Rollback strategy**: Roll back to previous git tag + re-`make init` (project stage allows database reset).

## Clarified Decisions

- **Tenant code (tenant code)**: Only allows ASCII lowercase letters, digits and hyphens, format fixed as `[a-z0-9-]{2,32}`; Chinese or other Unicode characters not allowed. Chinese, English display names only written to `name`.
- **TenantId priority**: Formal JWT is authoritative tenant identity for normal business requests; header/subdomain only pre-login hints; tenant switch must go through `/auth/switch-tenant` re-sign.
- **Platform bypass and impersonation**: Only `TenantId=0` management platform mode full bypass; impersonation `TenantId=T` filters by T, only audit marks platform admin acting on behalf.
- **Suspended tenant boundary**: Tenant members cannot read/write; platform admin can perform read-only troubleshooting and recovery through `/platform/*`. Business writes must go through impersonation or dedicated platform API.
- **install_mode switch back to global**: Platform admin confirms then forces all tenants enable that plugin, with mandatory audit.
- **User list isolation boundary**: 1:N mode membership is authoritative boundary for tenant visibility; `sys_user.tenant_id` only represents primary tenant/default login tenant, cannot be sole filter condition for user list.
- **Platform admin impersonation token audience**: Current version does not split separate audience, still uses formal token structure, distinguished by `is_impersonation`, `acting_user_id` and audit fields.
- **Tenant-level i18n custom text**: Not landed in this version; schema extension position `plugin_multi_tenant_i18n_override` not created.
- **Platform admin full platform aggregate view**: Current version provides "tenant overview + user aggregation + operation log full" basic page, deep dashboard left for extension.
- **Quota enforcement timing**: No table created, no mock data, no hook implemented in this version; tenant-level quota/billing re-modeled in subsequent iteration.
