## ADDED Requirements

### Requirement: Tenant Resolution Chain of Responsibility
System SHALL provide code-fixed tenant resolution chain of responsibility, attempting `[override, jwt, session, header, subdomain, default]` resolvers in order before each HTTP request enters business processing; first resolver returning non-empty `TenantID` wins. In normal authenticated business requests, JWT `TenantId` is authoritative tenant identity; `X-Tenant-Code` and subdomain only serve as pre-login or `pre_token` stage tenant hints, must not override formal JWT.

#### Scenario: Default resolution order
- **WHEN** Logged-in request arrives with JWT containing `TenantId`
- **AND** Request header `X-Tenant-Code` also exists
- **THEN** `jwt` resolver hits before `header`, adopts JWT tenant
- **AND** If `X-Tenant-Code` differs from JWT tenant, system ignores hint and records security audit event; normal user cannot use header to temporarily switch tenant

#### Scenario: Resolution chain fixed in code
- **WHEN** System starts or processes tenant resolution requests
- **THEN** Uses code-built-in resolution chain `[override, jwt, session, header, subdomain, default]`
- **AND** Does not read host config file or plugin database table to override chain order

#### Scenario: Pre-login header hint
- **WHEN** Unauthenticated login request carries `X-Tenant-Code: acme`
- **THEN** `header` resolver can use `acme` as tenant hint
- **AND** After authentication success, only when user has acme's active membership is auto-selection or candidate tenant return allowed

### Requirement: Override Resolver
`override` resolver SHALL parse `X-Tenant-Override` request header, but only effective when current user is in platform context with all data permissions and `system:tenant:impersonate` permission; otherwise header ignored.

#### Scenario: Platform admin legal impersonation
- **WHEN** Platform admin request carries `X-Tenant-Override: acme`
- **THEN** `bizctx.TenantId` resolves to `acme.id`, `bizctx.ActingAsTenant = true`
- **AND** Subsequent queries and writes filtered as tenant acme's normal tenant view
- **AND** Operation log carries `acting_user_id = platform admin`, `on_behalf_of_tenant_id = acme.id`

#### Scenario: Normal user attempts override
- **WHEN** Normal user (no platform permission) request carries `X-Tenant-Override: acme`
- **THEN** Override resolver treats header as illegal and skips
- **AND** Operation log records warn-level security event

### Requirement: Subdomain Resolver and Reserved Subdomains
`subdomain` resolver SHALL extract tenant subdomain from request host by valid `root_domain`, excluding `reserved` list (default `[www, api, admin, static, docs]`). First version `root_domain` fixed empty by code default, backend update endpoint does not accept non-empty value; therefore subdomain resolution currently disabled, subsequent iteration opens settings to enable.

#### Scenario: Extract tenant subdomain
- **WHEN** Request host is `acme.app.com` and `root_domain = app.com`
- **THEN** Extract `acme` as tenant code
- **AND** Call Provider to resolve to TenantID

#### Scenario: Current version rootDomain not open
- **WHEN** Platform admin attempts to save non-empty `rootDomain`
- **THEN** Backend rejects save
- **AND** Runtime resolution strategy continues using empty `rootDomain`, subdomain resolver returns empty and lets chain continue

#### Scenario: Hit reserved subdomain
- **WHEN** Request host is `www.app.com`
- **THEN** Subdomain resolver returns empty, chain continues
- **AND** Does not affect jwt/session subsequent resolution

### Requirement: Unrecognized Request Handling Strategy
When all resolvers return no valid TenantID, system SHALL return `TENANT_REQUIRED` error code by code-fixed `prompt` strategy. This strategy only applies to pre-login, `pre_token` or internal fallback chains without formal JWT tenant identity; must not be used to override JWT `TenantId` in authenticated business requests.

#### Scenario: Fixed prompt mode first login
- **WHEN** Unauthenticated 1:N user first login without any tenant hint
- **THEN** API returns `TENANT_REQUIRED` error code and user-visible tenant list
- **AND** Frontend shows selector, user selects then continues via `/auth/select-tenant`

### Requirement: Resolver Configuration Source
Resolution chain order, reserved subdomains, `root_domain` and ambiguous behavior SHALL be fixed in code, with code comments explaining rule logic, priority and disable boundaries. System SHALL NOT create or read `plugin_multi_tenant_resolver_config` table; host `config.template.yaml` SHALL NOT expose `tenant.resolution.*` config items.

#### Scenario: Code-fixed resolution strategy
- **WHEN** System resolves tenant identity or validates tenant code hits reserved tag
- **THEN** Directly uses code-built-in resolution chain, reserved subdomains and `prompt` ambiguous behavior
- **AND** Does not execute database queries, runtime config saves or cluster config broadcasts

#### Scenario: rootDomain currently disabled
- **WHEN** Request host has subdomain
- **THEN** Due to code-fixed `rootDomain = ""`, subdomain resolver returns empty and lets chain continue
- **AND** When rootDomain opens later, must redesign config source and distributed consistency strategy

### Requirement: Resolution Result and Consistency
Resolver results (subdomain/header -> TenantID mapping) SHALL use `plugin_multi_tenant_tenant` table as authoritative data source; current implementation does not introduce resolution config cache or config change broadcast. Tenant code once created cannot be modified, therefore no code change resolution cache remapping.

#### Scenario: Tenant deletion triggers invalidation
- **WHEN** Platform admin deletes tenant `acme`
- **THEN** Subsequent queries by `plugin_multi_tenant_tenant` authoritative data source no longer hit valid TenantID
- **AND** Does not depend on resolution config cache invalidation
