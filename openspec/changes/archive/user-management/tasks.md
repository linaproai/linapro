## 1. Database Design and Migration

- [x] 1.1 Create SQL migration file `008-menu-role-management.sql`
- [x] 1.2 Create `sys_menu` table
- [x] 1.3 Create `sys_role` table
- [x] 1.4 Create `sys_role_menu` table
- [x] 1.5 Create `sys_user_role` table
- [x] 1.6 Add menu-related dictionary types (menu status, display status, menu type)
- [x] 1.7 Add initial role data (admin, user roles)
- [x] 1.8 Add initial menu data (system management menus and sub-menus)
- [x] 1.9 Assign admin role to default administrator user
- [x] 1.10 Add `sys.auth.loginPanelLayout` and `sys.auth.pageDesc` to `sys_config` seed data with default values

## 2. Backend - Menu Management Module

- [x] 2.1 Run `make dao` to generate menu-related DAO/DO/Entity
- [x] 2.2 Create menu API definition files `api/menu/v1/`
- [x] 2.3 Define menu list query endpoint `GET /menu`
- [x] 2.4 Define menu detail endpoint `GET /menu/:id`
- [x] 2.5 Define menu creation endpoint `POST /menu`
- [x] 2.6 Define menu update endpoint `PUT /menu/:id`
- [x] 2.7 Define menu deletion endpoint `DELETE /menu/:id`
- [x] 2.8 Define menu dropdown tree endpoint `GET /menu/treeselect`
- [x] 2.9 Define role menu tree endpoint `GET /menu/role/:roleId`
- [x] 2.10 Run `make ctrl` to generate menu controller skeleton
- [x] 2.11 Implement menu service layer `internal/service/menu/menu.go`
- [x] 2.12 Implement menu list query (tree structure)
- [x] 2.13 Implement menu detail query
- [x] 2.14 Implement menu creation (name uniqueness validation)
- [x] 2.15 Implement menu update
- [x] 2.16 Implement menu deletion (cascade child menus, clean role associations)
- [x] 2.17 Implement menu dropdown tree endpoint
- [x] 2.18 Implement role menu tree endpoint

## 3. Backend - Role Management Module

- [x] 3.1 Run `make dao` to generate role-related DAO/DO/Entity
- [x] 3.2 Create role API definition files `api/role/v1/`
- [x] 3.3 Define role list query endpoint `GET /role`
- [x] 3.4 Define role detail endpoint `GET /role/:id`
- [x] 3.5 Define role creation endpoint `POST /role`
- [x] 3.6 Define role update endpoint `PUT /role/:id`
- [x] 3.7 Define role deletion endpoint `DELETE /role/:id`
- [x] 3.8 Define role status toggle endpoint `PUT /role/:id/status`
- [x] 3.9 Define role dropdown options endpoint `GET /role/options`
- [x] 3.10 Define role user list endpoint `GET /role/:id/users`
- [x] 3.11 Define role assign users endpoint `POST /role/:id/users`
- [x] 3.12 Define remove user authorization endpoint `DELETE /role/:id/users/:userId`
- [x] 3.13 Run `make ctrl` to generate role controller skeleton
- [x] 3.14 Implement role service layer `internal/service/role/role.go`
- [x] 3.15 Implement role list query (paginated)
- [x] 3.16 Implement role detail query (including menu ID list)
- [x] 3.17 Implement role creation (name and key uniqueness validation)
- [x] 3.18 Implement role update (including menu association update)
- [x] 3.19 Implement role deletion (clean menu and user associations)
- [x] 3.20 Implement role status toggle
- [x] 3.21 Implement role dropdown options endpoint
- [x] 3.22 Implement role user list query
- [x] 3.23 Implement role assign users
- [x] 3.24 Implement remove user authorization

## 4. Backend - User Management Extension

- [x] 4.1 Extend user list API to return roleIds and roleNames
- [x] 4.2 Extend user detail API to return roleIds
- [x] 4.3 Extend user creation to support roleIds parameter
- [x] 4.4 Extend user update to support roleIds parameter
- [x] 4.5 Extend user deletion to clean sys_user_role associations

## 5. Backend - Login Authentication Extension

- [x] 5.1 Extend `/user/info` endpoint response structure
- [x] 5.2 Implement user role query logic
- [x] 5.3 Implement user menu tree construction logic
- [x] 5.4 Implement user permission identifier aggregation logic
- [x] 5.5 Handle super admin special logic (return all menus)
- [x] 5.6 Handle user with no roles (empty menu logic)

## 6. Backend - Login-Page Config Parameters

- [x] 6.1 Add protected public-frontend metadata, default values, domain validation, and whitelist projection for `sys.auth.loginPanelLayout`
- [x] 6.2 Add protected public-frontend metadata, default values, and length validation for `sys.auth.pageDesc`
- [x] 6.3 Update built-in system-parameter seed list in host initialization SQL
- [x] 6.4 Add backend config unit tests covering valid values, invalid value rejection, and public-frontend output

## 7. Frontend - Menu Management Page

- [x] 7.1 Create menu management API file `src/api/system/menu/`
- [x] 7.2 Create menu management route `src/router/routes/modules/system.ts`
- [x] 7.3 Create menu management page `src/views/system/menu/index.vue`
- [x] 7.4 Create menu form drawer `src/views/system/menu/menu-drawer.vue`
- [x] 7.5 Create menu data definitions `src/views/system/menu/data.ts`
- [x] 7.6 Implement menu tree table display
- [x] 7.7 Implement menu search functionality
- [x] 7.8 Implement menu creation functionality
- [x] 7.9 Implement menu editing functionality
- [x] 7.10 Implement menu deletion functionality (with cascade)
- [x] 7.11 Implement menu status toggle
- [x] 7.12 Implement menu icon selector

## 8. Frontend - Role Management Page

- [x] 8.1 Create role management API file `src/api/system/role/`
- [x] 8.2 Create role management route
- [x] 8.3 Create role management page `src/views/system/role/index.vue`
- [x] 8.4 Create role form drawer `src/views/system/role/role-drawer.vue`
- [x] 8.5 Create role data definitions `src/views/system/role/data.ts`
- [x] 8.6 Create role menu selection component `src/components/tree/MenuSelectTable.vue`
- [x] 8.7 Create role user assignment page `src/views/system/role/authUser.vue`
- [x] 8.8 Implement role list display
- [x] 8.9 Implement role search functionality
- [x] 8.10 Implement role creation (including menu selection)
- [x] 8.11 Implement role editing (including menu selection)
- [x] 8.12 Implement role deletion
- [x] 8.13 Implement role status toggle
- [x] 8.14 Implement role user assignment
- [x] 8.15 Implement remove user authorization

## 9. Frontend - User Management Extension

- [x] 9.1 Extend user list to add role column
- [x] 9.2 Extend user form to add role selector
- [x] 9.3 Extend user creation to support role association
- [x] 9.4 Extend user editing to support role association

## 10. Frontend - Login Authentication Extension

- [x] 10.1 Update user info type definitions, add menus and permissions
- [x] 10.2 Update useUserStore to store menu and permission info
- [x] 10.3 Configure frontend dynamic route generation logic
- [x] 10.4 Implement menu-tree-based route registration
- [x] 10.5 Implement button-level permission directive v-access

## 11. Frontend - Login Page Presentation

- [x] 11.1 Adjust frontend default preferences and public-frontend runtime sync so login page defaults to `panel-right`
- [x] 11.2 Adjust login-page assembly so forgot password, registration, mobile login, QR-code login, and third-party login entry points are explicitly hidden
- [x] 11.3 Adjust auth-route registration so unfinished auth subpages redirect to `/auth/login`
- [x] 11.4 Update login-page description to match product positioning
- [x] 11.5 Update frontend runtime tests for public-frontend config sync

## 12. Backend - Server Monitor Optimization

- [x] 12.1 Create SQL migration `009-server-monitor-updated-at.sql` adding `updated_at` field
- [x] 12.2 Run `make dao` to regenerate DAO/DO/Entity
- [x] 12.3 Modify `CollectAndStore` method to use framework auto-handled time fields
- [x] 12.4 Modify `GetLatest` method to sort by `updated_at`
- [x] 12.5 Update frontend "collection time" label to "data update time"
- [x] 12.6 Add `CleanupStale` method to `servermon` module
- [x] 12.7 Extract cron job registration logic to `cron/cron_servermon_cleanup.go`

## 13. E2E Tests

- [x] 13.1 Create menu management test `TC0060-menu-crud.ts`
- [x] 13.2 Test menu creation
- [x] 13.3 Test menu editing
- [x] 13.4 Test menu deletion
- [x] 13.5 Test menu tree display
- [x] 13.6 Create role management test `TC0061-role-crud.ts`
- [x] 13.7 Test role creation
- [x] 13.8 Test role editing (including menu selection)
- [x] 13.9 Test role deletion
- [x] 13.10 Test role user assignment
- [x] 13.11 Create user-role association test `TC0062-user-role.ts`
- [x] 13.12 Test user creation with role selection
- [x] 13.13 Test user editing with role modification
- [x] 13.14 Test user list displays roles
- [x] 13.15 Create login menu test `TC0063-auth-menu.ts`
- [x] 13.16 Test menu displays correctly after login
- [x] 13.17 Test different role users see different menus
- [x] 13.18 Create role form defaults test `TC0064-role-form-defaults.ts`
- [x] 13.19 Create login page presentation test `TC0102-login-page-presentation.ts`
- [x] 13.20 Test unfinished entry points stay hidden
- [x] 13.21 Test login page defaults to right-aligned layout
- [x] 13.22 Test system parameter switches layout

## 14. Integration and Verification

- [x] 14.1 Run backend unit tests
- [x] 14.2 Run frontend type checks
- [x] 14.3 Run full E2E test suite
- [x] 14.4 Manually verify super admin login menu
- [x] 14.5 Manually verify regular user login menu
- [x] 14.6 Manually verify button-level permission control
- [x] 14.7 Manually verify login page presentation and system-parameter behavior

## Feedback

- [x] **FB-1**: Menu page data display failure due to query form dictionary option loading method error causing VXE-Grid initialization failure
- [x] **FB-2**: Menu status search dropdown showing incorrect options
- [x] **FB-3**: Empty column on right side of table
- [x] **FB-4**: Status and display column label styles inconsistent with reference project
- [x] **FB-5**: Form row spacing and remark input style inconsistent with reference project
- [x] **FB-6**: Menu management page query form dictionary dropdowns empty; switched from `getDictOptionsSync()` to `getDictOptions()` for async loading
- [x] **FB-9**: Menu status search dropdown showing extra options due to test data pollution from TC0013; fixed with independent test dictionary type and SQL cleanup
- [x] **FB-10**: Edit menu drawer not showing edited menu content; root cause: `handleEdit` passes `isEdit` but drawer checks `update`
- [x] **FB-11**: Parent menu dropdown tree not showing child menus in new/edit drawer; root cause: backend returns tree but frontend treated as flat list
- [x] **FB-12**: Edit menu parent selector should disable current menu and descendants to prevent circular reference
- [x] **FB-13**: Remark input in new/edit menu drawer too small; replaced with Textarea component
- [x] **FB-14**: Clicking "add" on specific menu should show that menu as parent in the add panel; root cause: `resetForm()` called after `setFieldValue()`
- [x] **FB-16**: Role sort InputNumber default value correctly configured
- [x] **FB-17**: Data permission field has `rules: 'required'` for mandatory selection
- [x] **FB-18**: Role menu permission tree shows type icons and permission checkboxes; added Type and Icon fields to MenuTreeNode
- [x] **FB-19**: Role sort InputNumber default not displaying; changed `getDrawerSchema()` from async to sync
- [x] **FB-20**: Data permission field defaults to "all data permissions"; confirmed `defaultValue: 1` correct after FB-19 fix
- [x] **FB-21**: Role status Select dictionary loading timing issue; changed to RadioGroup with hardcoded options
- [x] **FB-22**: Server monitor data storage optimization; added `updated_at` field and `CleanupStale` method
- [x] **FB-24**: Improve cron job service layer encapsulation; added `CleanupStale` to servermon, extracted cron registration
- [x] **FB-login-1**: Expand `sys.auth.pageDesc` length limit to 500 characters with validation and regression coverage
- [x] **FB-login-2**: Update default login-page description and standardize default login-panel layout to right side

## Menu Structure Summary
```
Dashboard (sort:0)
  Analysis
  Workspace
System Management (sort:1)
  User Management + 7 buttons
  Role Management + 4 buttons
  Menu Management + 4 buttons
  Department Management + 4 buttons
  Post Management + 5 buttons
  Dictionary Management + 5 buttons
  Notice Announcement + 4 buttons
  Parameter Settings + 5 buttons
  File Management + 4 buttons
  Message List (hidden)
  Role Auth Users (hidden)
System Monitor (sort:2)
  Online Users + 2 buttons
  Server Monitor
  Operation Log + 4 buttons
  Login Log + 4 buttons
System Info (sort:3)
  System API
  Version Info
Personal Center (hidden, sort:99)
```


---

## Tenant Permission Boundary Hardening Tasks

## 1. 后端权限边界

- [x] 1.1 新增或复用平台上下文 guard，覆盖平台租户接口、菜单治理写操作、插件平台治理写操作，并显式拒绝代管租户上下文
- [x] 1.2 在 `apps/lina-plugins/multi-tenant/backend/internal/service/tenant` 的平台租户列表、详情、创建、更新、删除、启停和代管入口接入平台上下文校验
- [x] 1.3 在 `apps/lina-core/internal/service/menu` 的创建、更新、删除、状态/显隐变更路径接入平台上下文校验，并保持权限拓扑修订号发布失败时失败关闭
- [x] 1.4 在 `apps/lina-core/internal/service/plugin` 的插件同步、上传、安装、卸载、启用、禁用、升级、安装模式和 tenant provisioning policy 写路径接入平台上下文校验

## 2. 角色授权可分配集合

- [x] 2.1 实现集中可分配菜单/权限谓词，按平台上下文、租户上下文、`menu_key`、权限字符串和插件治理元数据判定 platform-only 权限
- [x] 2.2 更新角色菜单树查询，租户上下文过滤平台租户管理、平台插件治理和全局菜单治理写权限
- [x] 2.3 更新角色创建和更新，在写入 `sys_role_menu` 前校验所有 `menuIds` 均属于当前上下文可分配集合
- [x] 2.4 确认异常历史 platform-only 授权不会绕过平台控制面 guard，并在必要时过滤租户态 checked keys

## 3. 任务分组租户隔离

- [x] 3.1 更新任务分组列表、详情和任务计数查询，按当前租户在数据库查询阶段过滤
- [x] 3.2 更新任务分组创建，显式写入当前 `tenant_id` 并按 `(tenant_id, code)` 校验唯一性
- [x] 3.3 更新任务分组更新和删除，操作前校验目标分组属于当前租户
- [x] 3.4 更新删除分组迁移逻辑，只迁移当前租户内任务到当前租户默认分组
- [x] 3.5 盘点并按需调整 seed/mock 任务分组数据，确保租户演示数据不依赖 `tenant_id=0` 全局分组

## 4. Fallback 元数据与前端显隐

- [x] 4.1 更新配置 API DTO、服务投影和前端类型，返回 `sourceTenantId`、`isFallback`、`canEdit`、`canOverride`、`overrideMode`
- [x] 4.2 更新字典类型和字典数据 API DTO、服务投影和前端类型，返回同一套 fallback 来源与动作元数据
- [x] 4.3 更新参数设置页面，按动作元数据隐藏 fallback 行的直接编辑/删除入口，并避免必失败详情请求
- [x] 4.4 更新字典类型和字典数据页面，按动作元数据隐藏 fallback 行的直接编辑/删除入口，并避免必失败详情请求
- [x] 4.5 若本阶段实现创建租户覆盖入口，补齐明确的覆盖创建流程；若不实现，保留 `canOverride` 元数据但不显示直接覆盖按钮

## 5. i18n、apidoc 和错误治理

- [x] 5.1 为平台上下文缺失、角色菜单分配 forbidden、任务分组租户不可见等新增或复用稳定 `bizerr` 错误码
- [x] 5.2 补齐宿主和 multi-tenant 插件的 `manifest/i18n/{zh-CN,en-US}/error.json` 错误翻译
- [x] 5.3 补齐相关 apidoc i18n JSON，确保新增错误和新增响应字段有中英文说明
- [x] 5.4 补齐前端运行时语言包中 fallback 只读、继承平台默认、创建租户覆盖等可见文案；本阶段未新增 fallback 可见文案，仅隐藏直接编辑/删除入口并保留 `canOverride` 元数据

## 6. 自动化测试

- [x] 6.1 后端单元测试：租户上下文即使持有异常 `system:tenant:*` 权限也不能访问 `/platform/tenants`
- [x] 6.2 后端单元测试：租户角色授权树不返回平台节点，角色创建/更新提交 platform-only `menuIds` 被拒绝
- [x] 6.3 后端单元测试：租户上下文不能创建、更新、删除全局菜单
- [x] 6.4 后端单元测试：租户上下文不能更新插件 tenant provisioning policy 或执行平台插件生命周期治理动作
- [x] 6.5 后端单元测试：任务分组列表、创建、更新、删除、任务计数和删除迁移均按 `tenant_id` 隔离
- [x] 6.6 后端单元测试：配置和字典 fallback 行返回正确来源与动作元数据
- [x] 6.7 新增 multi-tenant 插件 E2E `TC0239-tenant-role-platform-permission-blocked.ts`，覆盖租户角色授权树隐藏平台节点、提交平台 `menuIds` 被拒绝、异常平台权限不能访问平台租户接口
- [x] 6.8 新增 E2E `TC0240-tenant-governance-actions-hidden.ts`，覆盖租户态菜单管理和插件管理平台治理动作隐藏且后端拒绝直接调用
- [x] 6.9 新增 E2E `TC0241-tenant-job-group-isolation.ts`，覆盖租户任务分组只显示本租户、创建写入当前租户、不能更新或删除范围外分组
- [x] 6.10 新增 E2E `TC0242-config-dict-fallback-readonly.ts`，覆盖参数和字典 fallback 行不显示直接编辑入口且不触发 not found 详情请求

## 7. 验证与审查

- [x] 7.1 运行受影响 Go 包测试，至少覆盖 `apps/lina-core/internal/service/{menu,role,plugin,jobmgmt,sysconfig,dict}` 和 `apps/lina-plugins/multi-tenant/backend/internal/service/tenant`
  - 已通过：`cd apps/lina-core && go test ./internal/service/menu ./internal/service/role ./internal/service/jobmgmt ./internal/service/sysconfig ./internal/service/dict -count=1`
  - 已通过：`cd apps/lina-core && go test ./internal/service/plugin -run 'TestPluginGovernanceMethodsRejectTenantContext|TestUpdateTenantProvisioningPolicy' -count=1`
  - 已通过：`cd apps/lina-plugins/multi-tenant/backend && GOWORK=off go test ./internal/service/tenant -count=1`
  - 已通过：使用完整宿主配置副本并将 `database.default.link` 指向临时 SQLite 库 `/tmp/linapro-plugin-test.Vl4Jba/linapro-plugin-test.db`，执行 `GF_GCFG_PATH=/tmp/linapro-plugin-test.Vl4Jba GF_GCFG_FILE=config.yaml go run main.go init --confirm=init --sql-source=local --rebuild=true` 初始化宿主 schema，再补齐 multi-tenant 插件 `manifest/sql/001-multi-tenant-schema.sql` 后运行 `GF_GCFG_PATH=/tmp/linapro-plugin-test.Vl4Jba GF_GCFG_FILE=config.yaml go test ./internal/service/plugin -count=1`
  - 说明：共享开发 Postgres 库上直接运行 `cd apps/lina-core && go test ./internal/service/plugin -count=1` 曾因历史插件治理残留失败，错误为 `Plugin <id> cannot be changed because installed plugins depend on it`；隔离初始化库完整通过后，判定该失败来自共享库状态污染而非当前工作区代码或测试用例。
- [x] 7.2 如修改 Controller 构造、路由绑定、启动编排或 API 签名，运行 `cd apps/lina-core && go test ./internal/cmd -count=1` 或更窄但覆盖绑定的测试
  - 已通过：`cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] 7.3 运行前端 typecheck、i18n 检查、E2E TypeScript 校验和新增 Playwright 用例
  - 已通过：`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`
  - 已通过：`go run ./hack/tools/linactl i18n.check`
  - 已通过：`cd hack/tests && pnpm exec tsc --noEmit -p tsconfig.json`
  - 已通过：`cd hack/tests && pnpm test:validate`
  - 已通过：重启当前工作区服务后，`cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test apps/lina-plugins/multi-tenant/hack/tests/e2e/tenant-isolation/TC0239-tenant-role-platform-permission-blocked.ts apps/lina-plugins/multi-tenant/hack/tests/e2e/platform-admin/TC0240-tenant-governance-actions-hidden.ts apps/lina-plugins/multi-tenant/hack/tests/e2e/tenant-isolation/TC0241-tenant-job-group-isolation.ts apps/lina-plugins/multi-tenant/hack/tests/e2e/tenant-isolation/TC0242-config-dict-fallback-readonly.ts`
- [x] 7.4 运行 `openspec validate tenant-permission-boundary-hardening --strict`
  - 已通过：`openspec validate tenant-permission-boundary-hardening --strict`
- [x] 7.5 运行相关 diff 空白检查，并记录 i18n、缓存一致性、数据权限和开发工具影响结论
  - 已通过：`git diff --check`
  - 已通过：`git -C apps/lina-plugins diff --check`
  - i18n 结论：已同步宿主和 multi-tenant 插件错误资源、宿主 apidoc i18n JSON 与前端类型；本阶段 fallback 可见操作仅隐藏直接编辑/删除入口，未新增前端运行时可见文案。
  - 缓存一致性结论：未新增缓存；菜单、角色和插件治理继续复用既有权限拓扑、插件运行时状态和显式作用域失效机制；任务分组无独立缓存。
  - 数据权限结论：平台控制面新增平台上下文强 guard；角色授权可分配集合、任务分组读写迁移、配置/字典 fallback 元数据均按当前租户上下文收敛。
  - 开发工具结论：未新增或修改长期维护开发工具/脚本；验证复用现有 `linactl`、Go、pnpm 和 Playwright 入口。
- [x] 7.6 完成实现后调用 `lina-review` 进行代码和规范审查
  - 已完成：审查发现 `SyncSourcePlugins` 仍可作为公开写入口绕过新增平台 guard，已补齐同一 `ensurePlatformGovernance` 校验并纳入 `TestPluginGovernanceMethodsRejectTenantContext` 覆盖。
  - 审查结论：未发现新的必须阻断问题；`plugin` 全包测试的动态/源码插件生命周期失败仍作为 7.1 阻断和剩余风险保留。

## Feedback

- [x] **FB-1**: 平台 admin 在监控插件已安装启用时，系统监控目录只显示服务监控，缺少登录日志、在线用户和操作日志
  - 插件状态：通过 `/api/v1/plugins` 确认 `monitor-loginlog`、`monitor-online`、`monitor-operlog` 均为 `installed=1`、`enabled=1`、`runtimeState=normal`，安装模式为 `tenant_scoped` 且 `scopeNature=tenant_aware`。
  - 根因：插件集成层的已启用快照为进程级共享状态，租户上下文菜单过滤可能把租户范围插件写入为 disabled，平台 admin 后续复用该租户派生快照时会错误隐藏登录日志、在线用户和操作日志菜单。
  - 修复：`storeLoadedEnabledSnapshot` 改为接收 `ctx`，仅平台租户上下文允许刷新共享启用快照，避免租户态过滤污染平台 admin 菜单过滤结果。
  - 自动化验证已通过：`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -run 'TestStoreLoadedEnabledSnapshotBackfillsSharedState|TestTenantSnapshotDoesNotOverwritePlatformSnapshot|TestPlatformOnlyGlobalPluginRemainsEnabledInTenantContext' -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -count=1`。
  - 接口验证已通过：`/api/v1/menus/all` 返回 `/monitor/online`、`/monitor/server`、`/monitor/operlog`、`/monitor/loginlog`，`/api/v1/user/info` 返回对应 `monitor:*` 权限。
  - Playwright 验证已通过：admin 登录后展开系统监控目录，可见 `在线用户`、`服务监控`、`操作日志`、`登录日志`，截图保存为 `temp/screenshots/admin-monitor-menus.png`。
  - i18n 影响：本次未新增、修改或删除用户可见文案和翻译键。
  - 缓存一致性影响：未新增缓存键或缓存后端；修复范围为进程内插件启用快照写入边界，禁止租户上下文覆盖平台快照，分布式插件状态仍沿用既有运行时状态与拓扑失效机制。
  - 数据权限影响：未新增数据读写接口；仅修正平台 admin 菜单可见性，租户态菜单过滤边界不扩大。
  - lina-review 结论：未发现新的阻断问题；变更未修改 API、Controller、路由绑定或数据库访问面，Go 编译门禁由 `go test ./internal/service/plugin/internal/integration -count=1` 覆盖，新增测试可复现并防止租户上下文污染平台共享快照。
- [x] **FB-2**: tenant-user 上传图片后文件管理列表不可见，但平台 admin 可见；同时审查其他业务模块是否存在类似租户归属漏写问题
  - 根因：`file.Upload` 写入 `sys_file` 时未设置 `tenant_id`，租户上传的文件落到平台租户 `0`；文件管理列表已经按当前租户和角色数据权限过滤，所以 tenant-user 查不到，平台 admin 可见。
  - 修复：文件上传新文件和重复 hash 复用路径均写入当前 `tenant_id`，重复 hash 查询限定同一租户，保证跨租户不会复用范围外文件元数据。
  - 同类修复：用户 Excel 导入写入 `sys_user.tenant_id`，并在需要时同事务写入租户成员关系；字典类型/数据导入和参数配置导入均按当前租户判断存在性、写入当前租户，并避免租户覆盖导入误更新平台 fallback 行；调度器创建运行中/终态任务日志时继承 `sys_job.tenant_id`。
  - 同类审查：角色、角色菜单、用户角色、任务分组/任务创建、通知消息/投递、在线会话、KV cache、插件状态、动态插件 host state、源码插件 content-notice/org-center/monitor 登录和操作日志等路径已显式设置或传递租户；插件目录、插件发布、插件节点状态、源码插件治理数据属于平台/运行时治理边界，本次未改。
  - 自动化验证已通过：`cd apps/lina-core && go test ./internal/service/file -count=1`、`cd apps/lina-core && go test ./internal/service/sysconfig -count=1`、`cd apps/lina-core && go test ./internal/service/dict -count=1`、`cd apps/lina-core && go test ./internal/service/user -count=1`、`cd apps/lina-core && go test ./internal/service/jobmgmt ./internal/service/jobmgmt/internal/scheduler -count=1`。
  - 治理验证已通过：`openspec validate tenant-permission-boundary-hardening --strict`、`git diff --check`。
  - i18n 影响：未新增、修改或删除用户可见文案、错误码、API DTO 或翻译键。
  - 缓存一致性影响：未新增缓存或缓存失效路径；修复仅调整持久化租户归属，读取侧继续复用既有租户/数据权限过滤。
  - 数据权限影响：写入归属与已有读取过滤对齐；文件列表、配置/字典 fallback 读取、用户列表和任务日志列表不会再因写入 `tenant_id=0` 泄漏或隐藏租户数据。
  - 开发工具影响：未新增或修改开发工具、脚本或跨平台入口。
  - lina-review 结论：审查发现导入和文件 hash 复用路径不能用平台绕过式 `ApplyTenantScope` 表达“当前租户存在性”，已改为显式 `tenant_id = CurrentTenantID(ctx)` 并重新运行上述测试；未发现新的阻断问题。
- [x] **FB-3**: tenant-user mock 租户用户不应看到平台级服务监控菜单
  - 根因：multi-tenant mock 数据为 `tenant-user` 动态授予平台管理子树以外的所有启用菜单；`monitor-server` 是 `platform_only` + `global` 插件，因此仍会被授权并在租户上下文显示为服务监控。
  - 修复：`tenant-user` mock 授权 SQL 在排除平台管理子树的基础上，额外排除 `plugin:monitor-server:%` 菜单键，避免授予服务监控页面及其权限节点。
  - 自动化验证已通过：`cd apps/lina-plugins/multi-tenant && GOWORK=off go test . -run TestMultiTenantMockDataContainsTenantUserIsolationAccount -count=1`、`cd apps/lina-plugins/multi-tenant && GOWORK=off go test . -count=1`。
  - 治理验证已通过：`openspec validate tenant-permission-boundary-hardening --strict`。
  - i18n 影响：仅调整 mock SQL 注释和测试断言，不新增、修改或删除用户可见文案、翻译键、manifest i18n 或 apidoc 资源。
  - 缓存一致性影响：不新增缓存读写或失效路径；变更仅影响重新初始化或重新加载 mock 数据后的角色菜单关系。
  - 数据权限影响：收窄 `tenant-user` mock 角色菜单授权，避免租户演示账号获得平台级服务监控入口；未改变运行时权限解析、真实租户插件启用规则或业务数据访问接口。
  - 开发工具影响：未新增或修改开发工具、脚本或跨平台入口。
  - lina-review 结论：未发现新的阻断问题；变更范围仅为 mock SQL 授权谓词、对应测试断言和任务记录，未修改生产 Go 代码、API、Controller、路由绑定、缓存逻辑或 i18n 资源。
- [x] **FB-4**: 配置和字典 fallback `overrideMode` 等 API DTO 使用裸 `string` 表达枚举语义，需审查并修复同类 API 枚举字段
  - 根因：配置和字典 fallback 元数据中的 `overrideMode` 是有限动作模式，但 API DTO 使用裸 `string`；同类问题还存在于定时任务、任务日志、任务处理器、菜单、插件治理、公共前端配置、健康检查、i18n 诊断、用户消息来源和列表排序方向等 DTO 字段。
  - 修复：为上述 API 包新增或补齐命名枚举类型与常量，并在 Controller 边界显式完成服务层字符串/领域枚举到 API DTO 枚举的转换；JSON wire 值保持不变。
  - 同类审查：`dictType`/字典 `type`、文件 `scene`、分组 `code`、错误 `reason`/`errorCode`、缓存 `scope`、租户候选 `status` 等保留为 `string`，原因分别是字典类型标识、字典模块维护的业务值、开放业务编码、机器错误码/说明、显式缓存作用域或 provider-owned 生命周期状态，不属于宿主 API 可封闭维护的 Go 枚举。
  - 自动化验证已通过：`cd apps/lina-core && go test ./api/... -count=1`、`cd apps/lina-core && go test ./api ./internal/controller/config ./internal/controller/dict ./internal/controller/job ./internal/controller/joblog ./internal/controller/jobhandler ./internal/controller/menu ./internal/controller/user ./internal/controller/plugin ./internal/controller/publicconfig ./internal/controller/health ./internal/controller/i18n ./internal/controller/usermsg ./internal/controller/file ./internal/controller/jobgroup -count=1`、`cd apps/lina-core && go test ./internal/cmd -count=1`。
  - 治理验证已通过：`openspec validate tenant-permission-boundary-hardening --strict`、`git diff --check`。
  - i18n 影响：未新增、修改或删除用户可见文案、错误码、运行时语言包、`manifest/i18n` 或 `apidoc i18n JSON`；仅收紧 Go API DTO 类型。
  - 缓存一致性影响：不涉及缓存读写、失效、刷新或跨实例协调；JSON wire 值不变。
  - 数据权限影响：不新增或扩大数据操作接口，不改变查询、详情、写操作或导出过滤边界；仅调整 DTO 类型和投影转换。
  - 开发工具影响：未新增或修改长期维护开发工具、脚本或跨平台执行入口。
  - lina-review 结论：未发现新的阻断问题；后端 Go 编译门禁、启动绑定包测试、OpenSpec 严格校验和空白检查均已通过。
- [x] **FB-5**: 登录页选择租户后不应短暂回显账号密码输入界面，应展示进入租户的过渡加载状态直到后台路由完成
  - 根因：`selectTenant` 在拿到租户访问令牌后立即清空 `pendingPreToken`，登录页条件渲染因此短暂切回账号密码表单；随后 `fetchUserInfo` 与后台路由跳转完成后才离开登录页。
  - 修复：新增 `tenantLoginTransitioning` 状态，选择租户后优先展示 `login-tenant-transition` 过渡界面，避免账号密码表单在令牌交换、用户信息加载和后台路由跳转期间回显。
  - 自动化验证已通过：`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`、`cd hack/tests && pnpm exec tsc --noEmit -p tsconfig.json`、`cd hack/tests && pnpm test:validate`、`cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test apps/lina-plugins/multi-tenant/hack/tests/e2e/tenant-switching/TC0184-login-select-tenant.ts --config playwright.config.ts --project=chromium`。
  - Playwright 截图验证已通过：使用延迟的 `/auth/select-tenant` mock 捕获选择租户后的页面，确认显示“正在进入租户”与加载环且账号密码输入框不存在，截图保存为 `temp/screenshots/login-tenant-transition.png`。
  - 治理验证已通过：`go run ./hack/tools/linactl i18n.check`、`openspec validate tenant-permission-boundary-hardening --strict`、`git diff --check -- apps/lina-vben/apps/web-antd/src/store/auth.ts apps/lina-vben/apps/web-antd/src/views/_core/authentication/login.vue apps/lina-vben/apps/web-antd/src/locales/langs/zh-CN/pages.json apps/lina-vben/apps/web-antd/src/locales/langs/en-US/pages.json hack/tests/pages/MultiTenantPage.ts openspec/changes/tenant-permission-boundary-hardening/tasks.md`。
  - i18n 影响：新增登录租户过渡界面中英文运行时文案，未修改 `manifest/i18n` 或 `apidoc i18n JSON`。
  - 缓存一致性影响：不涉及缓存读写、失效或跨实例协调。
  - 数据权限影响：不涉及后端数据接口或数据权限过滤；仅调整前端登录流程的过渡展示状态。
  - 开发工具影响：未新增或修改默认开发工具、脚本或跨平台执行入口。
  - lina-review 结论：未发现新的阻断问题；审查范围限定为前端登录状态切换、运行时语言包、共享 E2E 页面对象和任务记录，未修改 Go 生产代码、REST API、Controller、路由绑定、数据库、缓存或数据权限路径。
- [x] **FB-6**: API 层 `xxx_enum.go` 文件仅承载过渡枚举定义，应合并到对应模块主 `xxx.go` 文件以保持 API 代码紧凑
  - 根因：FB-4 为修复裸 `string` 枚举字段新增了多个 API 枚举定义文件，但这些 `xxx_enum.go` 只承载过渡类型与常量，降低了 API DTO 文件的聚合度。
  - 修复：将配置、字典、文件、任务、任务分组、任务日志、任务处理器、菜单、插件、公共前端配置、健康检查、i18n 运行时语言和用户消息等 API 枚举定义合并到对应模块主 DTO 文件或最相关的核心 DTO 文件，并删除所有 `api/*/v1/*_enum.go`。
  - 自动化验证已通过：`cd apps/lina-core && go test ./api/... -count=1`、`cd apps/lina-core && go test ./api ./internal/controller/config ./internal/controller/dict ./internal/controller/job ./internal/controller/joblog ./internal/controller/jobhandler ./internal/controller/menu ./internal/controller/user ./internal/controller/plugin ./internal/controller/publicconfig ./internal/controller/health ./internal/controller/i18n ./internal/controller/usermsg ./internal/controller/file ./internal/controller/jobgroup -count=1`、`cd apps/lina-core && go test ./internal/cmd -count=1`。
  - 治理验证已通过：`find apps/lina-core/api -path '*/v1/*_enum.go' -print | sort` 无残留输出、`openspec validate tenant-permission-boundary-hardening --strict`、`git diff --check`。
  - i18n 影响：不新增、修改或删除用户可见文案、错误码、运行时语言包、`manifest/i18n` 或 `apidoc i18n JSON`；JSON wire 值不变。
  - 缓存一致性影响：不涉及缓存读写、失效、刷新、跨实例协调或最大陈旧时间变化。
  - 数据权限影响：不新增或扩大数据操作接口，不改变读取、详情、写操作、导出或聚合过滤边界。
  - 开发工具影响：未新增或修改长期维护开发工具、脚本或跨平台执行入口。
  - lina-review 结论：未发现新的阻断问题；变更仅为 API 层类型定义文件组织调整和必要文件顶部注释补齐，Go 编译门禁、启动绑定包测试、OpenSpec 严格校验和空白检查均已通过。
