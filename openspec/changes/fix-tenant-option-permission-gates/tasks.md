# Tasks

## 1. 规范与实现

- [x] 1.1 更新 `tenant-platform-access-control` 增量规范，明确租户上下文前端不得主动请求平台租户控制面候选接口。
- [x] 1.2 在宿主前端租户选项加载和租户 store 中增加权限门禁，避免无权限时请求 `platform/tenants` 或 `auth/login-tenants`。
- [x] 1.3 调整 `linapro-tenant-core` 头部租户切换器的加载触发条件，确保权限码变化后可重试、有权限前不误请求。

## 2. 测试与验证

- [x] 2.1 新增前端单元测试覆盖租户选项加载门禁和租户上下文当前租户优先逻辑。
- [x] 2.2 创建 `apps/lina-plugins/linapro-tenant-core/hack/tests/e2e/host-integration/TC003-tenant-admin-no-platform-tenant-requests.ts`。
- [x] 2.3 实现 TC-3a：租户管理员访问用户管理不请求 `platform/tenants` 或 `auth/login-tenants`。
- [x] 2.4 实现 TC-3b：租户管理员访问角色管理不请求 `platform/tenants` 或 `auth/login-tenants`。

## 3. 影响记录与审查

- [x] 3.1 记录本次反馈修复的根因、影响分析、i18n/缓存/数据权限/开发工具/测试影响判断和验证命令。
- [x] 3.2 运行变更范围测试、`openspec validate fix-tenant-option-permission-gates --strict` 和静态检查。
- [x] 3.3 完成实现和验证后调用 `lina-review` 审查。

## 执行记录

### 根因

`loadUserTenantOptions` 在平台态会先调用 `authLoginTenants(userId)`，并在本地候选为空时回退 `platformTenantList`；这两条路径没有检查当前权限快照。`tenantStore.ensureTenantOptions` 也会被头部租户切换器在布局加载时立即触发，平台态直接请求 `platform/tenants`，租户态带 `userId` 时请求 `auth/login-tenants`。因此租户管理员处于租户上下文、缺少平台租户控制面权限时，前端仍发出了后端正确拒绝的请求。

### 影响分析

- 修改文件：`apps/lina-vben/apps/web-antd/src/store/tenant.ts`、`apps/lina-vben/apps/web-antd/src/store/tenant-permissions.ts`、`apps/lina-vben/apps/web-antd/src/views/system/user/tenant-options.ts`、用户管理相关页面、`apps/lina-plugins/linapro-tenant-core/frontend/slots/layout/header/actions/tenant-switcher.vue`、插件 `host-integration/TC003` E2E 和本 OpenSpec 变更。
- 受影响模块：宿主前端租户上下文状态、用户管理租户筛选/表单候选加载、多租户插件头部租户切换器。
- `i18n` 影响：不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、插件清单或语言包资源；插件启用 `i18n`，但本次只调整加载门禁和测试标题，不需要修改插件 `manifest/i18n`。
- 缓存一致性影响：不修改后端缓存、分布式失效或权限快照缓存；前端仅在无平台租户列表权限时清理本地候选，避免复用陈旧候选。
- 数据权限影响：不新增或修改后端数据操作接口；修复后减少无权限控制面请求，不扩大租户管理员数据可见性。
- 开发工具跨平台影响：不修改脚本、`Makefile`、`linactl`、CI 或构建入口。
- HTTP API/数据库影响：不修改路由、DTO、权限标签、OpenAPI 元数据、SQL、DAO、DO 或 Entity。
- 接口性能影响：减少无权限路径的前端瀑布式候选请求；不新增列表、批量或聚合接口调用。
- 插件 README 同步审查：不修改 `apps/lina-core/pkg/plugin` 插件能力，也不修改 `linapro-tenant-core` 插件清单或对外能力说明，无需同步 README。
- E2E 质量审查：触发，原因是 GitHub issue 76 属于用户可观察的权限/接口联动 bugfix。新增 TC003 覆盖原问题复现场景，断言用户管理和角色管理页面加载期间 `platform/tenants` 与 `auth/login-tenants` 请求计数均为 0，并在关键页面加载后截图审查。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/plugin.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`。

### 验证

- `pnpm -C apps/lina-vben exec vitest run --dom apps/web-antd/src/views/system/user/tenant-options.test.ts apps/web-antd/src/plugins/management-capabilities.test.ts`：通过，7 个测试通过。
- `pnpm -C apps/lina-vben exec vue-tsc --noEmit --skipLibCheck -p apps/web-antd/tsconfig.json`：通过。
- `pnpm -C apps/lina-vben exec eslint apps/web-antd/src/store/tenant.ts apps/web-antd/src/store/tenant-permissions.ts apps/web-antd/src/views/system/user/tenant-options.ts apps/web-antd/src/views/system/user/tenant-options.test.ts apps/web-antd/src/views/system/user/index.vue apps/web-antd/src/views/system/user/user-drawer.vue apps/web-antd/src/views/system/user/user-batch-edit-modal.vue`：通过。
- `pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-tenant-core/hack/tests/e2e/host-integration/TC003-tenant-admin-no-platform-tenant-requests.ts --workers=1 --reporter=list`：通过，2 个测试通过。
- E2E 截图审查：`temp/20260706/084343-tc003-user-management.png`、`temp/20260706/084351-tc003-role-management.png` 已人工检查，无 403 错误提示、原始 i18n key 或明显布局错乱。
- `pnpm -C hack/tests test:validate`：通过，校验 253 个 E2E 文件。
- `openspec validate fix-tenant-option-permission-gates --strict`：通过。
- `git diff --check && git -C apps/lina-plugins diff --check`：通过。
