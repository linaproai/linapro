## Context

LinaPro is a GoFrame + Vben5 full-stack management framework. The backend uses Controller-Service-DAO three-layer architecture with JWT authentication. The frontend uses Vue 3 + Ant Design Vue with pnpm monorepo. Prior to this change, the system had user management, department management, post management, and dictionary management, but lacked menu management and role management for fine-grained permission control. The login page also exposed unfinished authentication entry points and lacked host-configurable layout governance.

### Existing Architecture

- **Backend**: GoFrame v2, Controller-Service-DAO three-layer architecture
- **Frontend**: Vben5 + Vue 3 + Ant Design Vue, pnpm monorepo
- **Auth**: JWT Token, session stored in database
- **Permission**: hardcoded admin role, no dynamic permission control
- **Login page**: built on `@vben/common-ui`'s `AuthenticationLogin` component with existing `authPageLayout` preference supporting `panel-left`, `panel-center`, `panel-right`

### Reference Projects

The ruoyi-plus-vben5 project provides complete menu and role management implementation. This design references its data model and interaction patterns while following LinaPro's architecture conventions.

## Goals / Non-Goals

**Goals:**

- Implement menu management with tree hierarchy, three menu types (directory/menu/button)
- Implement role management with menu association, user association, and data-scope settings
- Extend user management with role assignment support
- Extend login authentication to return menu tree for frontend dynamic route generation
- Implement simplified data permissions (all/dept-only/self-only)
- Standardize login page to expose only username/password login with right-aligned default layout
- Manage login-panel position and page description through host public-frontend system parameters

**Non-Goals:**

- Do not implement full six-level data-scope range (all/custom/dept/dept-and-below/self-only/dept-and-below-or-self)
- Do not implement actual data-scope filtering logic (future iteration)
- Do not implement role permission editing page
- Do not implement menu permission parent-child non-linkage option
- Do not implement real forgot-password, registration, SMS-code login, QR-code login, or third-party login flows
- Do not change the login API, JWT behavior, session governance, or permission payload structure
- Do not redesign the login-page visual style; only adjust exposure strategy and default layout behavior

## Decisions

### 1. Database Design

Four new tables implement the menu-role-user association:

```
sys_menu (menu table)
├── id, parent_id, name, path, component, perms, icon
├── type (D=directory, M=menu, B=button)
├── sort, visible, status, is_frame, is_cache
├── query_param, remark
└── created_at, updated_at, deleted_at

sys_role (role table)
├── id, name, key, sort
├── data_scope (1=all, 2=dept-only, 3=self-only)
├── status, remark
└── created_at, updated_at, deleted_at

sys_role_menu (role-menu association)
├── role_id, menu_id
└── PRIMARY KEY (role_id, menu_id)

sys_user_role (user-role association)
├── user_id, role_id
└── PRIMARY KEY (user_id, role_id)
```

**Rationale**: Reference ruoyi-plus-vben5 mature design. Menus and roles support soft delete. Many-to-many association tables provide flexibility for one-user-many-roles and one-role-many-menus relationships.

### 2. Menu Type Design

Three menu types (D/M/B):

| Type | Identifier | Description | Characteristics |
|------|-----------|-------------|-----------------|
| Directory | D | Container for child menus | Has icon, route address, sort order |
| Menu | M | Actual page | Has component path, permission key, caching flag |
| Button | B | In-page button permission | Only permission key, no route |

**Rationale**: Directory type organizes menu hierarchy. Menu type corresponds to accessible pages. Button type controls in-page operation permissions.

### 3. Login Authentication Returns Menu Tree

Login returns user menu tree through `/user/info` response:

```json
{
  "userId": 1,
  "username": "admin",
  "realName": "Administrator",
  "avatar": "/avatar.png",
  "roles": ["admin"],
  "menus": [
    {
      "id": 1,
      "parentId": 0,
      "name": "system-management",
      "path": "system",
      "icon": "ant-design:setting-outlined",
      "type": "D",
      "children": [...]
    }
  ],
  "permissions": ["system:user:list", "system:user:add"],
  "homePath": "/dashboard"
}
```

**Rationale**: Vben5 framework supports backend-mode dynamic routing. Returning menu tree in one request reduces round trips. Permission identifiers enable button-level access control.

### 4. Menu Parent-Child Linkage

When assigning menus to roles, parent-child linkage is enabled by default:
- Checking a parent menu automatically checks all child menus
- Unchecking a child menu automatically unchecks the parent (until another child is checked)

**Rationale**: Simplifies user operation, matches intuition, reduces configuration errors.

### 5. Data Permission Scope

Three simplified data-scope levels:

| Value | Meaning | Description |
|-------|---------|-------------|
| 1 | All data | Can view all data |
| 2 | Dept data | Can only view own department data |
| 3 | Self data | Can only view self-created data |

**Rationale**: Covers common scenarios, reduces implementation complexity, extensible later.

### 6. Reuse Vben's Existing `authPageLayout` Enum for Login Panel Position

The underlying preference system already supports three layout values (`panel-left`, `panel-center`, `panel-right`). This change reuses that existing enum and standardizes LinaPro's default at `panel-right`.

**Rationale**: Reuses current layout switcher and rendering logic instead of maintaining a second mapping. A right-aligned default matches the current visual focus.

**Alternatives considered**:
- Hard-code to center and remove enum: fails configurability requirement
- Add custom position field mapped to CSS classes: duplicates existing capability

### 7. Close Unfinished Login Methods at Page-Entry and Route-Entry Layers

Hiding buttons is insufficient because the router still registers unfinished pages. Two-layer shutdown:
- Explicitly disable forgot-password, registration, mobile-login, QR-code, third-party entry points when assembling `AuthenticationLogin`
- Remove or redirect unfinished auth routes back to `/auth/login`

**Rationale**: Keeps both visible page entries and accessible routing surface aligned with product statement.

**Alternatives considered**:
- Hide only buttons: dead pages remain reachable
- Delete unfinished page files entirely: increases rework cost when implementing later

### 8. Login-Panel Position in Public-frontend Whitelist via `sys.auth.loginPanelLayout`

Login-panel position is presentation config that unauthenticated pages must read. New protected key: `sys.auth.loginPanelLayout`, allowed values: `panel-left`, `panel-center`, `panel-right`, grouped under `auth`.

Frontend runtime sync maps value into `preferences.app.authPageLayout`. Key belongs under `auth` group because it only affects unauthenticated auth pages.

**Rationale**: Existing public-frontend config endpoint already serves unauthenticated pages. No new API needed.

**Alternatives considered**:
- Put field under `ui`: too easy to confuse with workspace layout
- Change only frontend default: does not satisfy administrator configurability requirement

### 9. Keep Frontend Layout Switcher with Host System Parameters as Default Source

Vben login toolbar includes a layout switcher. This change keeps that capability but lets the host system parameter define the initial default state:
- LinaPro built-in default becomes `panel-right`
- Host can override via `sys.auth.loginPanelLayout` in `sys_config`
- Toolbar switcher still works as immediate per-session preview

### 10. API Design

**Menu Management API**:
| Method | Path | Description |
|--------|------|-------------|
| GET | /menu | Get menu list (tree) |
| GET | /menu/:id | Get menu detail |
| POST | /menu | Create menu |
| PUT | /menu/:id | Update menu |
| DELETE | /menu/:id | Delete menu |
| GET | /menu/treeselect | Get menu dropdown tree |
| GET | /menu/role/:roleId | Get role's menu tree |

**Role Management API**:
| Method | Path | Description |
|--------|------|-------------|
| GET | /role | Get role list (paginated) |
| GET | /role/:id | Get role detail |
| POST | /role | Create role |
| PUT | /role/:id | Update role |
| DELETE | /role/:id | Delete role |
| PUT | /role/:id/status | Update role status |
| GET | /role/:id/users | Get role's user list |
| POST | /role/:id/users | Assign users to role |
| DELETE | /role/:id/users/:userId | Remove user from role |

**User Management API Extensions**:
| Method | Path | Description |
|--------|------|-------------|
| GET | /role/options | Get role dropdown options |

## Risks / Trade-offs

### Risk 1: Menu Deletion Breaks Role Permissions

**Risk**: Deleting a menu leaves invalid data in `sys_role_menu`.

**Mitigation**: Cascade-delete `sys_role_menu` records when deleting menus. Use database-level cascading or application-level cleanup.

### Risk 2: Role Deletion Loses User Permissions

**Risk**: Deleting a role leaves invalid data in `sys_user_role` and `sys_role_menu`.

**Mitigation**: Cascade-delete `sys_user_role` and `sys_role_menu` records when deleting roles. Show warning about how many users are assigned.

### Risk 3: Super Admin Permission Handling

**Risk**: How to handle admin role's menu permissions.

**Mitigation**: Pre-configure `admin` role with all menu permissions (code-level bypass, not dependent on association table). Non-admin roles get permissions from `sys_role_menu`.

### Risk 4: Frontend Dynamic Route Refresh

**Risk**: Dynamic routes lost on page refresh.

**Mitigation**: Frontend route guard detects refresh and re-requests menu API. Or store menu tree in localStorage (with expiry handling).

### Risk 5: Bookmarks Pointing to Old Auth Sub-routes

**Risk**: External links may still point to old auth sub-routes.

**Mitigation**: Redirect unfinished auth routes back to `/auth/login` instead of leaving 404s.

### Risk 6: Host Config vs Frontend Layout Switching Priority

**Risk**: Both host system parameters and frontend local switching can affect panel position.

**Mitigation**: Public-frontend config defines startup default; runtime switching is per-session browser behavior only.

### Risk 7: New Public-frontend Parameter Without SQL Seed

**Risk**: Adding parameter without updating SQL seed makes it unmanageable.

**Mitigation**: Update built-in parameter list, validation logic, and tests together.

## Migration Plan

### Deployment Steps

1. Execute SQL migration scripts to create new tables and seed data
2. Add `sys.auth.loginPanelLayout` and `sys.auth.pageDesc` to `sys_config`
3. Deploy backend new version
4. Deploy frontend new version
5. Verify login, menu, role, and login-page presentation functionality

### Initialization Data

- Pre-configure `admin` role (super administrator)
- Pre-configure `user` role (regular user)
- Pre-configure system menus (system management, user management, etc.)
- Assign `admin` role to default administrator user
- Seed `sys.auth.loginPanelLayout=panel-right` and `sys.auth.pageDesc` default value

### Rollback Strategy

- SQL rollback scripts to drop new tables
- Backend rollback to previous version
- Frontend rollback to previous version

## Open Questions

1. **Menu sort field**: Use `sort` or `order_num`? Decision: use `sort`, consistent with post table naming.
2. **Role identifier field**: Use `key` or `role_key`? Decision: use `key`, concise.
3. **Menu name i18n**: Decision: `name` field stores i18n key (e.g., `menu.system.user`), frontend translates based on language pack.
4. **Future toolbar layout persistence**: Should a future iteration persist toolbar layout changes as user-specific preference? For now, host default is page-load source of truth; personalized preference governance is out of scope.
