## Why

LinaPro needs a complete user-management foundation that covers menu-based permission control, role assignment, login-page presentation governance, and system-parameter-driven configuration. Without these capabilities the system cannot enforce fine-grained access control, administrators cannot assign differentiated permissions to different roles, and the login page exposes unfinished entry points that mislead users into thinking unsupported authentication flows are available.

The menu and role management subsystem provides the core RBAC backbone: menus form a tree hierarchy of directories, pages, and buttons; roles bind menus to define permission scopes; users receive roles to inherit permissions; and the login flow returns the aggregated menu tree so the frontend can generate routes dynamically. The login-page presentation layer ensures that only the implemented username/password path is visible, defaults to a right-aligned layout, and lets administrators configure the layout and description through host-managed system parameters.

## What Changes

- **New menu management module**: full CRUD for system menus with tree hierarchy support (directory/menu/button types), external links, caching, icon selection, i18n keys, and status/visibility control.
- **New role management module**: full CRUD for roles with menu-permission assignment using parent-child linkage, user assignment to roles, data-scope settings (all/dept-only/self-only), and role status control.
- **Extended user management**: user creation and editing now support role selection; user list and detail views display associated role information; user deletion cleans up role associations.
- **Extended login authentication flow**: login returns the user's aggregated menu tree, role list, and permission identifiers; the frontend uses the menu tree for dynamic route generation and button-level access control.
- **Login page presentation governance**: hide unfinished authentication entry points (forgot password, registration, mobile login, QR-code login, third-party login); default to right-aligned login panel; manage login-panel position and page description through host public-frontend system parameters.
- **Server monitor data storage optimization**: use `updated_at` for tracking latest data collection time per node, with stale-data cleanup for dynamic environments.

## Capabilities

### New Capabilities

- `menu-management`: System menu management with tree hierarchy, three menu types (directory/menu/button), external links, caching, icon selection, status and visibility control.
- `role-management`: System role management with CRUD, menu-permission assignment, user assignment, data-scope settings, and status control.
- `user-role-association`: User-to-role association support in user management, including role selection in forms and role display in lists.
- `login-page-presentation`: Login page exposes only username/password login, hides unfinished entry points, defaults to right-aligned layout, and supports host-configurable layout and description.

### Modified Capabilities

- `user-auth`: Login authentication returns user roles, menu tree, and permission identifiers; supports dynamic route generation and button-level access control.
- `user-management`: User list, detail, creation, update, and deletion extended with role association support.
- `config-management`: Built-in public-frontend system parameters extended with login-panel position (`sys.auth.loginPanelLayout`) and login-page description (`sys.auth.pageDesc`).

## Impact

### Backend

- New database tables: `sys_menu`, `sys_role`, `sys_role_menu`, `sys_user_role`
- New API modules: `menu`, `role`
- Modified API modules: `auth` (returns menu tree), `user` (adds role association)
- New dictionary types: menu status, display status, menu type
- New system parameters: `sys.auth.loginPanelLayout`, `sys.auth.pageDesc`
- Modified config service: public-frontend whitelist, validation rules, and seed data

### Frontend

- New pages: menu management, role management, role-user assignment
- Modified pages: user management (adds role selection and display)
- Modified login flow: dynamic route generation based on menu tree, button-level permission directives
- Modified login page: hide unfinished entry points, right-aligned default layout, host-configurable layout and description
- Modified auth routes: unfinished auth subpages redirect to `/auth/login`

### Permission Control

- Login returns menu tree based on user roles; frontend controls page access via menu tree
- Button-level permissions controlled through menu button type (B) entries
- Admin role bypasses menu association checks and receives all permissions
