## ADDED Requirements

### Requirement: JWT Claims Carry Tenant Identity
JWT Claims SHALL add `TenantId int` field; `TenantId = 0` represents platform-level token (platform admin managing platform), `TenantId > 0` represents bound to specific tenant token.

#### Scenario: Normal tenant user token after login
- **WHEN** Tenant user U logs into tenant A successfully
- **THEN** Returns token Claims containing `UserId=U`, `TenantId=A`, `TokenId=...`
- **AND** This token equivalent to "U as A" context on all non-platform APIs

#### Scenario: Platform admin token
- **WHEN** Platform admin logs in (`sys_user.tenant_id=0` and obtains platform management `system:*` permissions via `tenant_id=0,data_scope=1` role)
- **THEN** Returns token Claims `TenantId=0`
- **AND** Does not need to call select-tenant

### Requirement: Two-Phase Login Flow (Authenticate -> Select Tenant)
Authentication success and "select tenant" SHALL be two independent steps: `POST /auth/login` only does credential validation returning `pre_token` + user visible tenant list; `POST /auth/select-tenant` accepts `pre_token + tenant_id` validates membership then issues formal JWT.

#### Scenario: Multi-tenant user login flow
- **WHEN** User U submits username password to `/auth/login`
- **AND** U has 2 active memberships in multi mode
- **THEN** Returns `{ pre_token, tenants: [{id, code, name}, ...] }`, no formal JWT
- **AND** Frontend shows selector, user selects then calls `/auth/select-tenant {pre_token, tenant_id}`
- **AND** Server issues formal JWT

#### Scenario: Single-tenant user login optimization
- **WHEN** User U has only 1 active membership
- **AND** Default `multi` strategy has no ambiguity
- **THEN** `/auth/login` directly returns formal JWT (equivalent to server auto-select)
- **AND** No frontend selection needed

### Requirement: pre_token Security Constraints
`pre_token` SHALL be short-term (60s), single-use, only usable for `/auth/select-tenant` special token; not allowed for business APIs.

#### Scenario: pre_token used for business API
- **WHEN** Client uses `pre_token` to call any business API
- **THEN** Returns 401 `bizerr.CodePreTokenNotAllowedForBusiness`
- **AND** Operation log records security event

### Requirement: Tenant Switch Interface
`POST /auth/switch-tenant` SHALL re-sign token in authenticated state; old token immediately invalidated via cluster broadcast.

#### Scenario: Tenant switch old token immediately invalidated
- **WHEN** User U holds token T1 (tenant=A), calls `/auth/switch-tenant {tenant_id: B}` returns token T2 (tenant=B)
- **AND** Client continues using T1 for business API
- **THEN** Server rejects T1 (checks revoke list)
- **AND** Returns `bizerr.CodeTokenRevoked`

### Requirement: Platform Admin Impersonation Token
Platform admin SHALL through `POST /platform/tenants/{id}/impersonate` temporarily switch to "operate as tenant T perspective"; system issues impersonation token, marking `acting_user_id = platform admin`, `tenant_id = T`, `is_impersonation = true`.

#### Scenario: Platform admin starts impersonation
- **WHEN** Platform admin calls `/platform/tenants/A/impersonate`
- **THEN** Returns impersonation token
- **AND** Requests with this token `bizctx.TenantId = A`, `bizctx.ActingAsTenant = true`, `bizctx.ActingUserId = platform admin`

### Requirement: Logout and Token Revocation
`POST /auth/logout` SHALL revoke current token, delete session from session store, and broadcast cluster invalidation; if user has active sessions in multiple tenants, only revoke current token's session.

#### Scenario: Logout only revokes current token
- **WHEN** User U holds tenant A token T1 and tenant B token T2, logs out with T1 on some client
- **THEN** T1 immediately invalidated, T2 unaffected
- **AND** Session store only deletes (A, T1) row
