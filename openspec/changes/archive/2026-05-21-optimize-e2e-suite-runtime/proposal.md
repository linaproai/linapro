## Why

当前 E2E 回归耗时已经成为交付瓶颈：host-only 模式约 36 分钟，plugin-full 模式约 2 小时。日志显示 plugin-full 的主要问题是大量串行文件集中在单个 CI job 中执行，host-only 的主要问题是许多页面用例重复承担登录态页面初始化、默认 dashboard 加载和业务页面跳转成本。

本变更用于系统性降低 E2E wall clock，同时保留插件生命周期、权限、i18n、菜单和共享数据等高风险场景的隔离语义。

## What Changes

- 将 plugin-full E2E 从单个全量 job 调整为基于通用入口的 CI 分片执行，覆盖宿主插件框架通用用例与全部源码插件自有用例。
- 明确 plugin-full 与 host-only 的职责边界，避免 plugin-full 无差别重复执行完整宿主套件；根目录 E2E 只保留不依赖具体官方插件的宿主插件框架、动态测试插件与通用插件治理覆盖。
- 插件自有用例的选择只保留通用入口：`plugins` 覆盖全部源码插件，`plugin:<plugin-id>` 覆盖单个源码插件，不再维护按官方插件业务模块命名的长期别名 scope。
- 新增 host-only 单模块入口，允许在没有 `apps/lina-plugins` 的主框架环境中只运行指定宿主模块。
- 为已登录页面增加轻量 fixture，允许测试直接进入目标业务路由，避免每个用例都先加载默认 dashboard。
- 为插件 E2E 增加幂等的 suite 级 baseline 设置能力，减少普通插件页面用例在 `beforeEach` 中重复同步、安装、启用、加载 mock 数据和刷新插件投影。
- 重构耗时最高的插件生命周期用例，使完整 UI 生命周期只覆盖代表插件，其他官方插件使用 API/contract smoke 加页面可访问性验证。
- 保留每个 E2E 用例的耗时记录，并在 CI artifact 中继续上传 Playwright report、test-results 和服务日志，作为优化验收依据。
- 修正 CI PostgreSQL health check 的认证参数，减少无效健康检查日志和潜在等待抖动。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `e2e-suite-execution-efficiency`: 增加 plugin-full 分片执行、轻量认证 fixture、插件 baseline 和耗时验收要求。
- `e2e-suite-organization`: 明确 plugin-full 与 host-only 的职责边界，要求完整验证链路覆盖官方插件但不得无差别重复宿主全量套件。

## Impact

- 影响 CI workflow：`.github/workflows/reusable-test-verification-suite.yml`、`.github/workflows/reusable-e2e-tests.yml`。
- 影响 E2E runner、清单和测试说明：`hack/tests/scripts/run-suite.mjs`、`hack/tests/scripts/execution-governance.mjs`、`hack/tests/config/execution-manifest.json`、`hack/tests/README.md`、`hack/tests/README.zh-CN.md`。
- 影响 Playwright fixture 和支持工具：`hack/tests/fixtures/auth.ts`、`hack/tests/fixtures/plugin.ts`、`hack/tests/support/ui.ts`。
- 影响部分高耗时 E2E 用例，重点包括 dynamic runtime 生命周期、官方源码插件生命周期、菜单 CRUD、文件管理及相似模式的宿主页面用例。
- 不改变生产 API、数据库 schema、运行时缓存语义或用户可见功能；本变更不需要新增 i18n 资源。
