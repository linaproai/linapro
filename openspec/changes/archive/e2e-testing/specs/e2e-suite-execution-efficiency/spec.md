## ADDED Requirements

### Requirement: E2E 测试套件必须提供分层且按环境分责的执行入口
E2E 测试套件必须同时提供 full regression、fast feedback 和 environment-scoped entrypoints。至少必须支持 `smoke`、`module`、`full`、`host-only` 与 `host-only module`，并且 plugin-full 必须通过通用插件入口选择测试范围，而不是依赖手工 globs 或长期维护的插件业务别名。

#### Scenario: 开发者需要关键路径快速反馈
- **WHEN** 开发者运行 smoke 入口
- **THEN** 系统必须执行预先声明的高价值关键路径用例集合
- **AND** 开发者不需要手工维护文件列表

#### Scenario: 开发者只验证受影响模块
- **WHEN** 开发者运行 module 入口并提供 scope
- **THEN** 系统必须通过 execution manifest 解析该 scope
- **AND** 只执行对应目录或文件集合，而不是依赖临时 glob

#### Scenario: host-only module 拒绝插件工作区范围
- **WHEN** 开发者通过 `test:host:module` 选择 `plugins`、`plugin:<plugin-id>` 或其他依赖 `apps/lina-plugins` 的 scope
- **THEN** runner 必须直接失败
- **AND** 说明该 scope 只能在 plugin-full 环境运行

#### Scenario: plugin-full 使用通用插件入口
- **WHEN** plugin-full 运行源码插件自有 E2E
- **THEN** runner 必须使用 `plugins` 或 `plugin:<plugin-id>` 选择测试范围
- **AND** 根路径通用插件分片只能选择宿主插件框架相关 scope，例如 `extension:plugin`

### Requirement: 高频已登录用例必须复用预生成登录态并支持轻量认证页面
除认证行为本身外，高频已登录管理台用例必须复用预生成认证态，而不是在每个用例中重复执行完整 UI 登录。同时，测试套件必须提供不自动进入默认 dashboard 的轻量页面 fixture，用于直接进入目标业务路由。

#### Scenario: 普通后台用例默认复用登录态
- **WHEN** 测试文件使用管理员已登录页面 fixture
- **THEN** fixture 必须从预生成 `storageState` 或等效认证态产物创建页面上下文
- **AND** 默认不再为每个用例重复执行完整登录流程

#### Scenario: 认证场景仍走真实登录链路
- **WHEN** 测试验证登录、登出、未认证跳转、登录失败或其他认证行为
- **THEN** 该测试必须仍可显式走真实认证流程
- **AND** 不得被共享登录态静默替换

#### Scenario: 轻量认证页面跳过默认 dashboard
- **WHEN** 测试使用轻量认证页面 fixture
- **THEN** fixture 必须创建已登录页面上下文
- **AND** 不得自动访问 `/dashboard/analytics`
- **AND** 测试必须自行导航到目标业务路由

#### Scenario: 登录态失效时自动再生
- **WHEN** 预生成登录态缺失、过期或失效
- **THEN** suite 准备步骤必须重新生成可用的认证态产物

### Requirement: 高成本等待必须收敛为状态驱动等待并定义串并行边界
E2E 套件必须优先使用页面、组件、API 和路由状态作为就绪信号，并显式区分可并行文件与共享状态串行文件。

#### Scenario: Page object 使用确定性就绪信号
- **WHEN** page object 需要等待表格、抽屉、弹窗、路由变化或反馈消息完成
- **THEN** 它必须优先使用可见性、loading 完成、API 完成或路由切换等确定性信号
- **AND** 不得把固定时长睡眠作为主要策略

#### Scenario: 独立文件进入并行池
- **WHEN** 某个测试文件不依赖跨文件共享全局状态
- **THEN** 它必须有资格进入受控 worker 数量的并行执行池

#### Scenario: 共享状态文件保留在串行边界内
- **WHEN** 测试文件会修改插件生命周期、全局配置、权限矩阵或其他跨文件共享状态
- **THEN** 它必须被显式归入串行执行边界
- **AND** 不得与无关文件并发执行造成不稳定

### Requirement: E2E 共享全局状态测试必须声明隔离类别并通过风险校验
E2E 套件必须对会修改或依赖跨文件共享全局状态的测试声明 isolation category。涉及插件生命周期、运行时 i18n bundle 版本、公共前端配置、系统参数、字典、菜单或角色权限矩阵、共享数据库种子数据、文件系统插件产物等风险面的文件，必须进入安全的执行边界，并接受自动化风险校验。

#### Scenario: 插件生命周期文件被归类为串行
- **WHEN** 测试文件会安装、启用、禁用、卸载、上传、同步或升级插件
- **THEN** execution manifest 必须为该文件声明插件生命周期隔离类别
- **AND** full runner 必须将其排除在并行池外

#### Scenario: 共享治理数据修改被归类为串行
- **WHEN** 测试文件修改系统参数、公共配置、字典、菜单权限、角色权限或其他共享治理数据
- **THEN** execution manifest 必须为其声明匹配的共享状态类别
- **AND** 除非存在带理由的安全例外，否则 full runner 必须将其保留在串行边界内

#### Scenario: validator 拒绝未分类高风险文件
- **WHEN** validator 通过静态启发式检测到高风险操作，但该文件未进入串行边界或缺少分类
- **THEN** 校验必须失败
- **AND** 输出包含文件路径、风险类别和期望 manifest 修复动作的错误信息

### Requirement: E2E 前置条件必须由 fixture 持有并支持幂等插件 baseline
E2E 套件必须通过共享 fixture 或 support helper 显式管理插件状态、mock 数据、认证态和共享文件系统前置条件。普通插件功能测试必须复用幂等 baseline，而生命周期测试必须保持对被测状态的显式控制。

#### Scenario: 测试依赖源码插件能力
- **WHEN** 测试文件依赖源码插件页面、API、菜单或 mock 数据
- **THEN** 它必须调用共享 helper 幂等地完成插件同步、安装、启用与投影刷新
- **AND** 只有插件提供对应 mock-data 资源时才加载 mock SQL

#### Scenario: 普通插件页面测试复用 baseline
- **WHEN** 普通插件功能测试声明一个或多个必需插件
- **THEN** suite 或 shard 级 baseline 必须复用该前置条件
- **AND** 不得在每个测试中重复执行相同的同步、安装、启用和投影刷新步骤

#### Scenario: 生命周期测试不被 baseline 污染
- **WHEN** 测试目标是插件安装、启用、禁用、卸载、上传、同步或升级生命周期
- **THEN** 该测试必须继续显式控制自己的初始状态与清理逻辑
- **AND** 普通 baseline 不得隐式改变被测插件状态

#### Scenario: 测试数据保持单文件自包含
- **WHEN** 测试文件创建用户、部门、岗位、公告、文件、插件记录或导入导出数据
- **THEN** 它必须使用唯一名称或稳定测试前缀
- **AND** 在 `finally`、`afterEach` 或 `afterAll` 中自行清理，不依赖跨文件清理

### Requirement: 缓存与业务状态断言必须遵循协议语义和稳定字段
E2E 套件必须以协议语义验证缓存和条件请求行为，并在多语言或跨环境断言中优先使用稳定业务字段，而不是依赖易变的展示文案。

#### Scenario: 条件请求命中未变化资源版本
- **WHEN** 缓存测试使用仍然匹配当前资源版本的 ETag 发起条件请求
- **THEN** 测试必须接受 `304 Not Modified`
- **AND** 验证返回的 ETag 与缓存值一致且不需要响应体

#### Scenario: 条件请求观察到合法的资源刷新
- **WHEN** 条件请求使用的 ETag 因其他合法测试或生命周期操作而失效
- **THEN** 测试必须接受 `200 OK`
- **AND** 只有在响应包含不同于旧值的新 ETag 和有效响应体时才判定通过

#### Scenario: 缓存测试仍验证条件头行为
- **WHEN** 缓存测试重载使用持久化缓存元数据的页面或资源
- **THEN** 它必须验证请求携带了预期条件头或等效前置条件
- **AND** 不得仅因为接口返回成功 body 就判定通过

#### Scenario: 多语言环境下断言业务状态
- **WHEN** 测试需要比较业务数量、身份、权限或状态流转
- **THEN** 它必须使用稳定 API 字段、ID、code、permission key、labelKey 或数值计数器进行业务断言
- **AND** 本地化文案应作为独立的表现层断言

### Requirement: 全量回归与优化结果必须输出可审计证据
E2E full runner、module runner 和 CI workflow 必须输出足够的执行证据，使串并行边界、分片行为和优化收益都可复核。

#### Scenario: full regression 启动时输出边界摘要
- **WHEN** 开发者或 CI 启动 full regression 入口
- **THEN** runner 必须输出并行文件数、串行文件数和 worker 数量摘要
- **AND** 必须包含串行集合对应的隔离类别

#### Scenario: module 或 shard 回归延续同一边界语义
- **WHEN** 开发者运行 module 命令或 CI 运行 shard
- **THEN** runner 必须对解析后的文件集合应用同一套串并行拆分逻辑
- **AND** 报告被串行化的文件与类别摘要

#### Scenario: 优化后保留耗时证据
- **WHEN** E2E workflow 完成
- **THEN** Playwright 输出或 artifact 必须保留每个测试用例的耗时记录
- **AND** CI 日志必须足以定位最慢文件、最慢用例和分片 wall clock

### Requirement: Browser E2E CI 基础设施必须支持通用插件分片与显式数据库健康检查
Browser E2E CI 必须在保持现有 plugin-full 启动语义的前提下，支持基于通用插件入口的分片执行，并且数据库健康检查必须显式指定 PostgreSQL 用户和数据库。

#### Scenario: plugin-full 分片失败阻断下游流程
- **WHEN** 任一 plugin-full 分片失败
- **THEN** 完整 verification suite 必须失败
- **AND** 依赖验证成功的下游 job 不得执行

#### Scenario: plugin-full 分片上传独立诊断证据
- **WHEN** plugin-full 分片完成或失败
- **THEN** workflow 必须上传该分片的 Playwright report、test-results、后端日志和前端日志
- **AND** artifact 名称必须包含分片标识，避免覆盖其他分片证据

#### Scenario: PostgreSQL 健康检查显式指定用户和数据库
- **WHEN** browser E2E workflow 启动 PostgreSQL service
- **THEN** health check 命令必须显式指定 `postgres` 用户和 `linapro` 数据库
- **AND** CI 日志不得因默认使用 runner 用户而反复输出无效角色错误
