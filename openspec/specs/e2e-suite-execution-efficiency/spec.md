## Purpose
定义 E2E 测试套件的执行效率、共享状态隔离、缓存重验证容忍度和前置条件治理要求。
## Requirements
### Requirement:E2E 共享全局状态测试必须声明隔离类别
E2E 测试套件 SHALL 分类变更或依赖跨文件共享全局状态的测试。变更插件生命周期、运行时 i18n 包版本、公共前端配置、系统参数、字典、菜单或角色权限矩阵、共享数据库种子数据或文件系统支持的插件产物的文件，必须声明隔离类别，并必须路由到防止不安全并行重叠的执行边界。

#### Scenario:插件生命周期测试分类为串行
- **当** 测试文件安装、启用、禁用、卸载、上传、同步或升级插件时
- **则** E2E 执行清单必须将该文件分类为插件生命周期隔离类别
- **且** 全回归运行器必须将该文件排除在并行池外

#### Scenario:全局配置测试分类为串行
- **当** 测试文件变更系统参数、公共前端配置、字典、菜单权限、角色权限或其他共享治理数据时
- **则** E2E 执行清单必须将该文件分类为匹配的共享状态类别
- **且** 全回归运行器必须将该文件排除在并行池外，除非存在显式文档化的安全例外

#### Scenario:验证器拒绝未分类的高风险测试
- **当** E2E 验证器在未串行或未分类的测试文件中检测到高风险操作时
- **则** 验证必须失败，并显示标识文件、检测到的风险类别和预期清单操作的消息

### Requirement:E2E 缓存重验证测试必须容忍合法的全局版本刷新
E2E 测试套件 SHALL 通过验证协议语义而非假设全局资源版本在全回归运行期间保持不变来测试缓存和 ETag 行为。条件请求必须验证请求前提条件以及未修改响应或刷新资源响应的正确性。

#### Scenario:条件请求命中未变更的资源版本
- **当** 缓存测试发送带有仍匹配当前资源版本的 ETag 的条件请求时
- **则** 测试必须接受 `304 Not Modified` 响应
- **且** 必须验证返回的 ETag 匹配缓存的 ETag 且不需要响应体

#### Scenario:条件请求观察到刷新的资源版本
- **当** 缓存测试发送带有因其他合法测试或生命周期操作刷新了资源版本而不再匹配的 ETag 的条件请求时
- **则** 测试必须仅在响应包含与缓存 ETag 不同的新 ETag 时接受 `200 OK` 响应
- **且** 必须验证刷新的响应体存在且有效

#### Scenario:缓存测试仍验证条件请求行为
- **当** 缓存测试重新加载应使用持久缓存元数据的页面或资源时
- **则** 测试必须验证请求携带了预期的条件头或等效缓存前提条件
- **且** 不得仅因资源端点返回了成功响应体而通过

### Requirement:E2E 前置条件必须由固件拥有且幂等
E2E 测试套件 SHALL 通过可复用的固件或支持辅助器使插件状态、模拟数据、认证状态和共享文件系统前置条件显式化。测试文件必须可独立运行，不依赖其他测试文件创建插件行、安装源码插件、加载模拟 SQL、刷新前端插件投影或创建可复用认证状态。

#### Scenario:测试依赖源码插件
- **当** 测试文件需要源码插件页面、API、菜单或模拟数据时
- **则** 测试必须调用幂等同步、安装、启用和刷新插件投影的共享固件/辅助器
- **且** 辅助器必须仅在插件提供匹配的模拟数据资源时加载插件模拟 SQL

#### Scenario:测试依赖生成的用户或业务数据
- **当** 测试文件创建用户、部门、岗位、通知、文件、插件记录或导入/导出数据时
- **则** 测试必须使用唯一名称或稳定测试前缀
- **且** 必须在 `finally`、`afterEach` 或 `afterAll` 中清理自己的数据，不依赖跨文件清理

#### Scenario:测试在本地化 UI 下读取业务状态
- **当** 测试需要在不同语言下比较业务计数、标识、权限或状态转换时
- **则** 必须使用 ID、代码、权限键、标签键或数字计数器等稳定 API 字段进行业务断言
- **且** 本地化 UI 文本必须作为展示行为单独断言

### Requirement:E2E 全回归报告必须暴露串行和并行边界
E2E 全回归运行器 SHALL 报告足够信息使执行隔离可审计。报告必须显示哪些文件在并行池中运行、哪些文件在串行池中运行、以及哪些隔离类别导致文件被串行化。

#### Scenario:全回归启动
- **当** 开发者或 CI 启动全回归入口点时
- **则** 运行器必须打印或持久化并行文件数、串行文件数和配置的工作线程数摘要
- **且** 必须包含串行集中表示的隔离类别

#### Scenario:模块范围回归启动
- **当** 开发者运行模块范围 E2E 命令时
- **则** 运行器必须对解析的模块文件应用相同的串行与并行拆分
- **且** 必须报告该模块范围内的任何串行化文件和类别

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

