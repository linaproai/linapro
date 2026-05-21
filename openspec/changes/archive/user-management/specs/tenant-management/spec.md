## ADDED Requirements

### Requirement: Tenant Entity Table Structure
`multi-tenant` plugin SHALL maintain `plugin_multi_tenant_tenant` table with fields at least including `id` (primary key), `code` (globally unique, only allows ASCII lowercase/digits/hyphens matching `[a-z0-9-]{2,32}`), `name` (display name, can contain Chinese or other Unicode), `status` (active/suspended/deleted), `remark`, `created_at`, `updated_at`, `deleted_at` (soft delete); tenant entity table, API DTO and management page SHALL NOT expose plan/套餐 field.

#### Scenario: Create tenant
- **WHEN** Platform admin calls `POST /platform/tenants` to create tenant `code=acme, name=ACME Group`
- **THEN** System validates code not in reserved subdomain list, matches ASCII character set and length, and does not allow Chinese or other Unicode characters
- **AND** Writes `plugin_multi_tenant_tenant` row with status=active
- **AND** Directly executes new tenant plugin default enable strategy domain service

#### Scenario: Chinese tenant code rejected
- **WHEN** Platform admin attempts to create tenant `code=研发部, name=研发部`
- **THEN** Returns `bizerr.CodeTenantCodeInvalid`
- **AND** Prompts tenant code can only use `[a-z0-9-]{2,32}`, display name can use Chinese

#### Scenario: Tenant code conflict
- **WHEN** Creating tenant `code=acme` when same code tenant already exists
- **THEN** Returns `bizerr.CodeTenantCodeDuplicated`
- **AND** No data written

### Requirement: Tenant Lifecycle State Machine
Tenant SHALL flow between `active`, `suspended`, `deleted` (soft delete) three states; only following transitions allowed:
- active -> suspended (suspend)
- suspended -> active (resume)
- active/suspended -> deleted (via platform admin second confirmation + LifecycleGuard pass)

#### Scenario: Suspend tenant
- **WHEN** Platform admin suspends tenant T
- **THEN** T's member login rejected (returns `TENANT_SUSPENDED`)
- **AND** T's existing tokens immediately invalidated
- **AND** T's data retained, tenant members cannot read or write
- **AND** Platform admin can perform read-only troubleshooting and recovery through `/platform/*`; business writes must go through impersonation or dedicated platform API with audit

#### Scenario: Delete tenant
- **WHEN** Platform admin deletes active or suspended tenant T
- **THEN** System executes all plugins' `CanTenantDelete` hooks
- **AND** All pass then `plugin_multi_tenant_tenant` row `deleted_at` set to current time
- **AND** Does not write outbox or trigger unimplemented cross-plugin event bus

### Requirement: Platform Admin Permission Requirements
Tenant CRUD operations SHALL require requesting user to be platform admin (`bizctx.TenantId = 0`, effective data permission is all data permissions and holds corresponding `system:tenant:*` functional permission); non-platform admin requests to `/platform/tenants/*` must return 403.

#### Scenario: Tenant admin attempts platform API access
- **WHEN** Tenant admin requests `GET /platform/tenants`
- **THEN** Returns 403 `bizerr.CodePlatformPermissionRequired`
- **AND** Operation log records permission denial event

### Requirement: Tenant List Query and Filtering
`GET /platform/tenants` SHALL support filtering by `code`, `name`, `status` and pagination; query results do not expose tenants where `deleted_at != NULL`, and return `createdAt` field corresponding to `created_at` for management page display.

#### Scenario: Default query excludes deleted
- **WHEN** Platform admin calls `GET /platform/tenants`
- **THEN** Only returns tenants where `deleted_at IS NULL`
- **AND** Default sorted by `created_at desc`

### Requirement: Tenant Management Page Inline Operations
Tenant management page SHALL provide impersonation entry in each tenant row; impersonation button SHALL show brief description on hover. Tenant management page SHALL NOT provide member management button, member drawer, member dialog or jump to system user management inline entry; tenant membership relationships, tenant-filtered user list, user tenant ownership adjustment and tenant role maintenance SHALL be unified through system user management page.

#### Scenario: Tenant management page does not host member management
- **WHEN** Platform admin views tenant management list
- **THEN** Row operations do not show member management entry
- **AND** Frontend does not open tenant member drawer or dialog
- **AND** Tenant membership maintenance through system user management page's tenant filter and user tenant ownership fields

### Requirement: Tenant Quota Not Modeled
`multi-tenant` plugin SHALL NOT create tenant quota table, write quota mock data or execute quota checks in this round; tenant-level quota/billing capability SHALL be re-designed in subsequent iteration.

#### Scenario: Plugin install does not create quota placeholder table
- **WHEN** Multi-tenant plugin installs
- **THEN** Does not create `plugin_multi_tenant_quota` table
- **AND** Plugin does not read, write or validate tenant quota in any business path

### Requirement: Tenant Code Immutable and Non-Reusable
Tenant `code` once created SHALL not be modifiable; after tenant deletion, its `code` SHALL retain 30-day tombstone before reuse (`plugin_multi_tenant_tenant` retains soft delete row, new same code creation rejected).

#### Scenario: Attempt to reuse deleted tenant code
- **WHEN** Tenant `acme` deleted 10 days ago
- **AND** Platform admin attempts to create new tenant `code=acme`
- **THEN** Returns `bizerr.CodeTenantCodeReserved` (reason i18n)
- **AND** Creation rejected
