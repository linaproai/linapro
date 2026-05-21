## Why

LinaPro 后端在依赖管理、API 合同和数据库支持三个维度上存在治理缺口。宿主 Controller、Middleware、插件服务适配器和部分插件后端在运行期各自调用 `service.New()` 构造独立服务图，导致缓存敏感路径上的实例不一致隐患。API 响应直接嵌入数据库实体，可能暴露密码、软删除字段、存储路径等实现细节。公共枚举值在多处重复定义，增加了契约漂移风险。仓库同时维护 PostgreSQL 与 SQLite 两条运行路径，导致 SQL 源治理、插件生命周期和 CI 门禁持续承担非生产数据库路径成本。

## What Changes

### Explicit Dependency Injection

- 后端宿主和源码插件的 Controller、Middleware、Service、插件宿主服务适配器和 WASM host service 不再在业务构造函数或请求/回调路径中隐式创建关键依赖服务。
- 构造函数改为逐项接收接口型显式依赖；禁止通过 `Dependencies`、`Deps`、`Options` 等聚合结构体整体传递多个接口对象。
- 高风险缓存一致性路径必须复用同一套运行期服务实例，包括认证中间件、权限服务、插件管理、插件运行时缓存、运行时配置、i18n、session hot state、source plugin registrar 和 WASM host service。
- 为源码插件和动态插件 host service 提供宿主发布的依赖入口，插件通过 registrar 获取宿主能力适配器。
- 初始化与注册 API 必须返回 error 给调用方决策，禁止在内部直接 panic 处理可预期错误。
- 更新项目规范和 lina-review 审查标准，增加静态扫描防止回归。

### API Contract Hardening

- 将宿主 API 响应中的 `entity.*` 嵌入替换为显式响应 DTO，只暴露必要字段。
- 提取公共枚举契约：`pkg/listorder`（排序方向）、`pkg/tenantoverride`（租户覆盖模式）、`pkg/statusflag`（通用 0/1 状态标志）。
- 菜单类型复用 `pkg/menutype`，插件桥接常量复用 `pkg/pluginbridge`。
- 清理文件详情页运行时翻译和宿主 apidoc i18n 中已不再暴露的响应字段翻译。
- 统一源码插件 API DTO 命名和存放方式，移除 `*Entity` 命名和软删除字段暴露。

### Database Support Convergence

- 运行时数据库支持收敛到 PostgreSQL 14+，移除 SQLite 作为开发、演示和测试目标的支持。
- 移除 GoFrame SQLite 驱动注册、SQLite 方言实现、SQLite DDL 转译器和 SQLite 专属启动钩子。
- 移除 SQLite 专属后端测试、E2E 用例、CI smoke workflow 和 `pnpm test:sqlite*` 入口。
- 保留 `pkg/dialect` 抽象但只保留 PostgreSQL 具体实现，`sqlite:` 链接必须明确失败。
- SQL 源不再受 SQLite 转译能力约束，可围绕 PostgreSQL 14+ 语法子集治理。

## Capabilities

### New Capabilities

- `service-dependency-injection-governance`：定义宿主、源码插件、动态插件 host service 和审查流程的显式依赖注入、共享实例和隐式构造禁止规则。
- `postgresql-only-database-support`：定义 LinaPro 运行时、初始化、测试和交付链路只支持 PostgreSQL 14+ 数据库的能力边界。

### Modified Capabilities

- `backend-conformance`：将现有控制器和服务层构造约束升级为显式依赖注入规范。
- `distributed-cache-coordination`：补充缓存敏感服务必须复用运行期同一服务实例的要求。
- `plugin-http-slot-extension`：补充源码插件 HTTP 注册回调应通过宿主发布依赖完成构造。
- `plugin-host-service-extension`：补充插件 host service 适配器必须由宿主运行期统一构造。
- `api-contract-consistency`：API 响应不得直接暴露数据库实体，公共枚举契约统一管理。
- `database-dialect-abstraction`：方言抽象移除 SQLite 方言，只保留 PostgreSQL 实现。
- `database-bootstrap-commands`：初始化和 mock 命令不再为 SQLite 准备数据库或转译 SQL。
- `cluster-deployment-mode`：移除 SQLite 专属单机锁定与启动警告。
- `cluster-coordination-config`：移除 SQLite 禁止集群 coordination 的特殊规则。
- `sql-source-syntax`：SQL 源不再受 SQLite 转译能力约束。
- `plugin-cache-service`：缓存服务规范移除 SQLite 持久化和 SQLite 锁冲突语义。
- `volatile-table-bootstrap`：易失性表引导规范移除 SQLite 场景。
- `release-image-build`：共享 CI 与 release 简要门禁移除 SQLite smoke。
- `project-setup`：数据库配置从 PostgreSQL + SQLite 调整为 PostgreSQL-only。

## Impact

- 影响 `apps/lina-core/internal/controller/**`、`apps/lina-core/internal/service/**`、`apps/lina-core/internal/cmd/**` 的构造方式。
- 影响 `apps/lina-core/api` 中 API DTO 类型定义和控制器响应组装。
- 影响 `apps/lina-core/pkg/dialect`、`apps/lina-core/pkg/dbdriver`、启动初始化命令。
- 影响 `apps/lina-plugins/**/backend/` 源码插件的依赖传递方式和 API DTO。
- 影响 `.github/workflows/*`、`.github/image/config.runtime.yaml`、`hack/tests`、`hack/makefiles/database.mk`。
- 不涉及前端 UI 交互变更；若实现过程中发现用户可见错误文案或 apidoc 变化，应按 i18n 规范同步维护。
