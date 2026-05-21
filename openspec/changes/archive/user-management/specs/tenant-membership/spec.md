## ADDED Requirements

### Requirement: 1:N User-Tenant Membership
`multi-tenant` plugin SHALL maintain `plugin_multi_tenant_user_membership(user_id, tenant_id, status, joined_at)` table, UNIQUE constraint `(user_id, tenant_id)`, allowing same user to have multiple active memberships.

#### Scenario: User joins second tenant
- **WHEN** User U already member of tenant A, joins tenant B
- **THEN** `user_membership` table adds `(U, B, active, ...)` row
- **AND** U can select to enter A or B after login

### Requirement: 1:1 Compatibility Strategy Reservation
System SHALL run with code default `multi`, allowing same user to have multiple active memberships; `single` preserved as optional strategy for controlled management settings. When `single` strategy enabled, system SHALL reject any write operation that would give user multiple active memberships.

#### Scenario: single mode adding second membership
- **WHEN** `single` strategy enabled and user U already has active membership in tenant A
- **AND** Platform admin attempts to add U to tenant B
- **THEN** Returns `bizerr.CodeMembershipExceedsCardinality`
- **AND** Write rejected

### Requirement: Platform Admin is Special User Without Membership
`tenant_id=0` platform context roles SHALL only be assignable to `sys_user.tenant_id = 0` users; platform admin users MUST NOT have `plugin_multi_tenant_user_membership` rows.

#### Scenario: Platform admin attempts to join tenant
- **WHEN** Platform admin U (tenant_id=0) attempted to join tenant A
- **THEN** Returns `bizerr.CodePlatformUserCannotJoinTenant`
- **AND** Write rejected; if U needs to operate in tenant, should use impersonation

### Requirement: Intra-Tenant Management Capability
Intra-tenant management capability SHALL be derived from current tenant context, role data permissions and `system:*` functional permissions combined, not maintaining extra admin boolean field in membership table.

#### Scenario: Tenant admin visible menus
- **WHEN** Tenant admin logs into that tenant
- **THEN** Menu through "User Management" hosts tenant membership management, can continue using "Role Management" and other tenant-level management items
- **AND** Does not include independent "Tenant Workbench" directory, `/tenant/members`, `/tenant/plugins` or "Tenant Management", "System Plugin Install" and other platform-level menus

### Requirement: User Kicked from Tenant
When a membership is deleted or status set to `removed`, system SHALL immediately invalidate that user's all tokens, sessions and permission caches in that tenant, but retain global `sys_user` and other tenant memberships.

#### Scenario: Remove membership triggers session invalidation
- **WHEN** Tenant admin removes user U from tenant A
- **THEN** U's sessions/tokens in tenant A immediately invalidated
- **AND** U's sessions in other tenants unaffected
- **AND** Operation log records `tenant_id = A`, `acting_user_id = tenant admin`

### Requirement: User Visible Tenant List
`GET /auth/login-tenants` and `GET /tenant/membership/me` SHALL return current user's all `status=active` memberships and corresponding tenant basic info (id, code, name).

#### Scenario: 1:N user login tenant selection
- **WHEN** User U authenticates successfully
- **AND** U has 2 active memberships
- **THEN** `/auth/login-tenants` returns list of length 2
- **AND** Frontend shows selector, user selects then calls `/auth/select-tenant`

### Requirement: Tenant Switch Token Re-sign
`POST /auth/switch-tenant` SHALL accept `target_tenant_id`, validate current user has active membership in target tenant, then:
1. Add old token to revoke list (short-term + cluster broadcast).
2. Re-sign JWT carrying new `TenantId`.
3. Delete old session, create new session.
4. Return new token and new menu/permissions.

#### Scenario: 1:N user tenant switch success
- **WHEN** User U holds tenant A token, calls `/auth/switch-tenant {target_tenant_id: B}`
- **AND** U has active membership in B
- **THEN** Old token immediately invalidated
- **AND** Returns new token, Claims `TenantId = B`

#### Scenario: Switch to tenant without membership
- **WHEN** User U calls `/auth/switch-tenant {target_tenant_id: C}`
- **AND** U has no active membership in C
- **THEN** Returns `bizerr.CodeTenantMembershipMissing`
- **AND** Old token not revoked
