## MODIFIED Requirements

### Requirement: Workbench Header Tenant Identifier and Switcher
Workbench header SHALL display current tenant identifier when multi-tenant enabled; 1:N user's identifier right side adds switch dropdown showing all active membership tenants. Top tenant switch dropdown SHALL maintain compact style: fixed width, single-line truncation, building icon, positioned left of global search entry, with stable spacing.

#### Scenario: 1:N user switches tenant
- **WHEN** User U has 3 memberships
- **THEN** Header dropdown shows 3 options, current tenant with check mark
- **AND** Selecting other tenant triggers `/auth/switch-tenant`
- **AND** Client uses new token to re-fetch menus, permissions and workbench data
- **AND** Client forces entry to new tenant context default page

#### Scenario: Top tenant switcher style stable
- **WHEN** multi-tenant enabled and workbench header shows tenant switcher
- **THEN** Dropdown maintains fixed width with single-line truncation for long tenant names
- **AND** Dropdown shows building icon
- **AND** Dropdown positioned left of global search entry
- **AND** Dropdown maintains stable spacing from global search entry

### Requirement: Platform Admin Special Header Style
Platform admin view (`bizctx.TenantId=0`) SHALL display prominent "Platform Administrator" identifier; impersonation mode shows "Acting as Tenant X" prompt bar positioned left of tenant switcher.

### Requirement: Multi-Tenant Disabled Hides Tenant UI
When multi-tenant not enabled, workbench header SHALL not show any tenant identifier, switcher or impersonation prompt; UI degrades to single-tenant.
