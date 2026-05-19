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


---

## Tenant Permission Boundary Hardening Design

## Context

审计报告确认当前权限链路多处只检查认证态和权限字符串，没有把“当前租户上下文是否允许访问该资源类型”和“目标资源归属是否匹配当前租户”作为同等级安全边界。最严重路径是租户用户在角色授权中分配平台租户管理权限，重新登录后后端将这些权限视为有效，从而读取 `/api/v1/platform/tenants` 的全租户数据。

现有多租户规格已经定义平台 bypass 只能在 `tenant_id=0`、非代管、全部数据权限时成立；`sys_menu`、`sys_plugin*` 等是平台控制面表；`sys_config`、`sys_dict_*` 具备平台默认值 + 租户覆盖语义。本变更不是重做多租户模型，而是把这些既有边界落实到缺失的服务路径上。

## Goals / Non-Goals

**Goals:**

- 为平台租户接口、菜单写操作、插件平台治理动作建立一致的平台上下文 guard。
- 在角色授权树展示和角色创建/更新提交中使用同一套可分配权限集合，防止前端过滤被绕过。
- 将任务分组明确收敛为租户作用域资源，保证读写、计数和删除迁移均不跨租户。
- 为配置和字典 fallback 行提供来源和动作元数据，使前端能区分平台继承值、租户覆盖值和可创建覆盖的状态。
- 补齐业务错误 i18n、apidoc i18n、前端运行时语言包、单元测试和 E2E 回归。
- 明确缓存一致性和数据权限影响，避免修复过程中引入单机缓存或本地状态假设。

**Non-Goals:**

- 不新增租户自定义菜单模型；租户态不得直接写 `sys_menu`。
- 不重新设计插件生命周期、插件安装模式或租户插件自服务协议，只补齐平台治理 guard 和前端显隐。
- 不在本阶段完整实现“点击 fallback 行一键创建租户覆盖”的复杂编辑流程；首版允许隐藏编辑并预留 `canOverride` / `overrideMode` 元数据。
- 不为配置/字典 fallback 引入新的数据库 schema，除非实现时发现现有字段无法表达已定义元数据。

## Decisions

### 1. 平台上下文 guard 作为权限字符串之外的强制条件

平台控制面 API 的授权条件定义为：

```text
authenticated
+ permission string
+ current tenant context is platform
+ current request is not acting-as-tenant / impersonation
+ effective data scope allows platform bypass when cross-tenant data is read
```

实现时优先复用 `tenantcap.Service.PlatformBypass(ctx)` 或等价现有上下文 helper。若某路径只需要判断平台上下文而不是跨租户数据 bypass，也必须显式拒绝 `ActingAsTenant=true` 的代管态，避免平台管理员进入租户视角后仍操作平台全局资源。

替代方案是只移除租户角色中的平台权限。该方案无法防御脏数据、历史授权或直接写入 `sys_role_menu` 产生的异常权限，因此不能作为唯一边界。

### 2. 角色授权展示和提交共享可分配权限谓词

角色授权树过滤和 `Role.Create/Update` 的 `menuIds` 校验必须调用同一套可分配权限谓词。租户上下文下，该集合排除：

- 平台租户管理和平台控制面菜单/按钮。
- 全局菜单治理写权限。
- 插件安装、卸载、启用、禁用、同步、上传、安装模式、新租户自动启用策略等平台插件治理权限。
- 其他通过 `menu_key`、权限字符串命名空间、插件清单元数据或稳定平台目录标记为 platform-only 的权限。

平台上下文可分配平台权限和普通租户通用权限，但仍要遵循受保护内置角色、数据权限和插件启用状态等既有规则。

替代方案是只在前端隐藏平台节点。该方案会被直接 API 提交绕过，审计已证明不足。

### 3. `sys_menu` 写操作平台化，租户定制另行建模

`sys_menu` 当前是全局权限拓扑和路由治理模型。菜单创建、更新、删除、状态变更等写操作必须要求平台上下文，并在成功后继续发布权限拓扑修订号。租户上下文可以读取当前上下文可见的授权投影，但不能修改全局表。

如未来支持租户自定义菜单，必须新增租户作用域模型或扩展字段，并设计清晰的合并和缓存失效语义；本变更不把租户写入直接落到 `sys_menu`。

### 4. 插件平台治理和租户插件自服务分离

插件生命周期和平台策略仍由宿主插件治理服务负责，且要求平台上下文。租户管理员只通过 `/tenant/plugins` 或 multi-tenant 插件提供的租户插件服务修改当前租户的 `tenant_scoped` 启用状态。

这避免 `plugin:edit` 同时表示“平台治理插件”和“租户自服务启停插件”。实现时前端应隐藏租户态插件管理页中的平台治理动作；后端仍必须拒绝越权调用。

### 5. 任务分组使用现有 `tenant_id` 成为租户资源

`sys_job_group` 已包含 `tenant_id` 且唯一键为 `(tenant_id, code)`，因此本变更将任务分组定义为租户作用域资源：

- 创建时写入当前 `tenant_id`。
- 列表和详情按当前租户过滤；平台上下文管理平台分组或通过明确平台接口跨租户治理。
- 更新、删除和任务计数在操作前校验目标分组归属。
- 删除分组时只迁移当前租户内的任务到当前租户默认分组。

替代方案是把任务分组定义为平台全局资源，但这与现有表结构和审计中租户态任务调度页面语义不一致。

### 6. Fallback 行返回动作元数据

配置和字典列表返回 effective row 时，增加来源和动作字段：

- `sourceTenantId`: 记录真实来源租户，平台默认行为 `0`。
- `isFallback`: 当前请求租户看到的记录是否来自平台默认。
- `canEdit`: 当前上下文是否能直接编辑该行。
- `canOverride`: 当前上下文是否能为 fallback 行创建租户覆盖。
- `overrideMode`: 首版可返回 `none` / `createTenantOverride` 等稳定枚举。

首版前端对 `isFallback && !canEdit` 的行隐藏编辑入口或显示只读状态；若 `canOverride=true`，可保留后续“一键创建覆盖”的入口设计，但不在本变更强制完整实现。

### 7. i18n、缓存、数据权限和开发工具影响

- i18n：新增或调整的业务错误必须补齐 `manifest/i18n/{zh-CN,en-US}/error.json`、apidoc i18n、前端运行时语言包；前端按钮、提示、只读/fallback 状态不得硬编码。
- 缓存：角色菜单写入、菜单治理写入和插件治理写入继续使用权限拓扑或插件运行时共享修订号；配置/字典覆盖写入按租户作用域失效，平台默认变更按 fallback 影响范围级联失效。不得引入仅当前节点可见的本地缓存控制。
- 数据权限：读取类在数据库查询阶段注入租户/数据权限过滤；详情和写操作在操作前校验目标资源可见性；平台控制面只允许平台上下文。
- 开发工具：本变更不新增或修改开发工具/脚本；如实现阶段需要新增验证入口，优先使用 Go 工具链并保持跨平台。

## Risks / Trade-offs

- [Risk] 历史租户角色已经持有平台权限，修复后用户短期看到权限减少或相关页面消失。→ 后端以强 guard 拒绝，角色授权树过滤异常权限，并通过测试覆盖“异常权限仍不能访问平台接口”。
- [Risk] 使用权限字符串或菜单目录判断 platform-only 容易遗漏新权限。→ 谓词应优先使用稳定 `menu_key`、平台目录、插件清单元数据和集中常量，新增平台治理权限必须加入同一分类测试。
- [Risk] 任务分组历史 `tenant_id=0` 数据在租户态不可见可能影响现有演示数据。→ 实现阶段盘点 mock/seed 数据；需要时将租户演示分组迁移到 mock 租户，平台默认分组保留平台上下文。
- [Risk] Fallback 元数据增加 DTO 字段可能影响前端类型和旧调用方。→ 字段向后兼容新增，旧调用方可忽略；前端类型同步更新。
- [Risk] 权限拓扑和插件状态修复在集群模式下只刷新本节点。→ 必须复用现有共享修订号/事件广播；相关测试和审查记录说明单机与集群分支。

## Migration Plan

1. 先增加平台上下文 guard、可分配权限谓词和后端单元测试，确保高危接口默认失败关闭。
2. 再补任务分组租户过滤和 fallback 元数据，保持 API 字段向后兼容。
3. 最后调整前端显隐、i18n 和 E2E，验证租户态页面不再暴露平台治理动作。
4. 若发现历史租户角色已绑定 platform-only 菜单，本阶段不强制删除历史关系；权限解析和平台 guard 保证不生效，后续可新增治理任务清理脏授权。
5. 回滚时可移除前端显隐和新增字段消费，但不应回滚平台 guard；安全 guard 属于失败关闭修复。

## Open Questions

- 是否在本变更中提供 fallback 行“一键创建租户覆盖”入口，还是仅返回 `canOverride` 并隐藏编辑？当前建议首版隐藏编辑并保留元数据。
- 平台上下文管理任务分组是否需要新增跨租户分组管理页面？当前建议仅保持平台上下文管理平台分组，不新增页面。
