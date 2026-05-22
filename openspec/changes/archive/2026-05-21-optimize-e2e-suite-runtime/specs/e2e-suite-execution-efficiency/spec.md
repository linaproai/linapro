## ADDED Requirements

### Requirement: Plugin-full E2E 必须支持通用插件入口分片执行
E2E CI workflow SHALL allow plugin-full browser regression to execute plugin management, plugin-owned tests, and plugin host seam tests as independent shards while preserving the same plugin-full startup semantics.

#### Scenario: Plugin-full 分片选择模块范围
- **WHEN** workflow 启动 plugin-full E2E 分片
- **THEN** 每个分片必须使用 plugin-full 服务启动命令
- **AND** 源码插件自有测试分片必须使用 `plugins` 或 `plugin:<plugin-id>` 通用入口选择测试范围
- **AND** 根目录分片只能选择宿主插件框架通用测试范围，不得选择依赖具体官方源码插件的根测试文件集合
- **AND** 分片日志必须显示选择的 scope、并行文件数、串行文件数和串行隔离类别

#### Scenario: Plugin-full 不维护官方插件业务别名 scope
- **WHEN** 开发者需要运行源码插件自有 E2E
- **THEN** runner 必须支持 `plugins` 运行全部源码插件自有用例
- **AND** runner 必须支持 `plugin:<plugin-id>` 运行单个源码插件自有用例
- **AND** E2E manifest 不应为官方插件业务模块维护长期别名 scope

#### Scenario: Plugin-full 分片失败阻止下游发布
- **WHEN** 任一 plugin-full E2E 分片失败
- **THEN** 完整验证套件必须失败
- **AND** 依赖完整验证成功的镜像发布或后续 job 不得执行

#### Scenario: Plugin-full 分片上传独立诊断证据
- **WHEN** plugin-full E2E 分片完成或失败
- **THEN** workflow 必须上传该分片的 Playwright report、test-results、后端日志和前端日志
- **AND** artifact 名称必须包含调用方前缀和分片标识，避免覆盖其他分片证据

### Requirement: E2E 认证页面 fixture 必须支持跳过默认 dashboard 导航
E2E 测试套件 SHALL provide an authenticated page fixture that creates a browser page with admin storage state without automatically navigating to the default dashboard.

#### Scenario: 测试直接进入目标业务路由
- **WHEN** 测试使用轻量认证页面 fixture
- **THEN** fixture 必须创建已登录上下文中的页面
- **AND** fixture 不得自动访问 `/dashboard/analytics`
- **AND** 测试必须显式导航到自身需要验证的目标路由

#### Scenario: 旧 adminPage fixture 保持兼容
- **WHEN** 现有测试继续使用 `adminPage`
- **THEN** fixture 必须保持现有默认 dashboard 可用语义
- **AND** 不得要求一次性迁移所有现有 E2E 用例

### Requirement: 普通插件功能 E2E 必须复用幂等插件 baseline
E2E 测试套件 SHALL provide reusable plugin baseline setup for ordinary plugin-owned page tests so they do not repeatedly synchronize, install, enable, seed, and refresh plugin projection in every test case.

#### Scenario: 插件功能测试声明所需插件集合
- **WHEN** 普通插件功能测试需要一个或多个源码插件处于可用状态
- **THEN** 测试或测试套件必须通过共享 baseline 辅助声明所需插件集合
- **AND** baseline 必须幂等执行插件同步、安装、启用、可用 mock 数据加载和插件投影刷新

#### Scenario: 插件生命周期测试不使用普通 baseline 覆盖被测状态
- **WHEN** 测试目标是插件安装、启用、禁用、卸载、上传、同步或升级生命周期
- **THEN** 测试必须继续显式控制自己的初始状态和清理逻辑
- **AND** 普通插件 baseline 不得在这些测试中隐式改变被测插件状态

### Requirement: E2E 优化必须保留可量化耗时验收
E2E runtime optimization SHALL preserve per-test timing evidence and compare before/after wall clock for host-only and plugin-full validation.

#### Scenario: 优化后保留测试耗时记录
- **WHEN** E2E workflow 运行完成
- **THEN** Playwright 输出或 artifact 必须保留每个测试用例的耗时记录
- **AND** CI 日志必须足以识别最慢文件、最慢用例和分片 wall clock

#### Scenario: Host-only 和 Plugin-full 目标耗时可复核
- **WHEN** 本变更完成验证
- **THEN** 任务记录必须写明 host-only 与 plugin-full 优化前后的耗时对比
- **AND** 若未达到目标耗时，必须说明剩余瓶颈和后续优化范围

### Requirement: Host-only 单模块 E2E 必须排除插件环境用例
E2E runner SHALL provide a host-only module entrypoint for running a selected host scope without requiring the official plugin workspace.

#### Scenario: 未初始化官方插件工作区时运行宿主模块
- **WHEN** 开发者在没有 `apps/lina-plugins` 的环境中运行 host-only module scope
- **THEN** runner 必须只选择该 scope 中不依赖官方插件工作区的宿主用例
- **AND** runner 不得要求初始化官方插件 submodule

#### Scenario: Host-only module 拒绝插件 scope
- **WHEN** 开发者通过 host-only module 入口选择需要 `apps/lina-plugins` 的 scope
- **THEN** runner 必须失败并说明该 scope 不能在 host-only module 模式运行

### Requirement: CI 数据库健康检查必须使用显式 PostgreSQL 用户和数据库
Browser E2E CI SHALL configure PostgreSQL service health checks with explicit user and database parameters instead of relying on the runner OS user.

#### Scenario: PostgreSQL 健康检查不使用 runner 用户
- **WHEN** browser E2E workflow 启动 PostgreSQL service
- **THEN** health check 命令必须显式指定 `postgres` 用户和 `linapro` 数据库
- **AND** CI 日志不得因健康检查默认使用 runner 用户而反复输出无效角色错误
