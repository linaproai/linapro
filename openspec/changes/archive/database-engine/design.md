## Context

LinaPro 数据库治理最终收敛到`PostgreSQL 14+`。早期为降低开发门槛保留过`SQLite`转译和多方言分支，但这会持续牵制 SQL 源语法、初始化命令、插件 SQL 生命周期、缓存协调、CI 和文档口径。项目没有兼容负担，因此运行时数据库支持矩阵直接收敛为单一`PostgreSQL`基线。

## Goals / Non-Goals

**Goals:**

- 以`PostgreSQL 14+`作为唯一受支持运行时数据库。
- 保留`pkg/dialect`作为数据库驱动、元数据查询、只读 SQL 分类和初始化准备的稳定边界。
- 删除`SQLite`运行、转译、测试、CI、文档和开发入口残留。
- 让插件 SQL、宿主 SQL、初始化、mock、发布镜像和集群协调都围绕同一数据库语义验证。

**Non-Goals:**

- 不提供`SQLite`、`MySQL`或其他数据库方言兼容层。
- 不维护旧数据迁移或历史开发库升级路径。
- 不以数据库兼容为由限制宿主和插件 SQL 使用受治理的`PostgreSQL`能力。

## Decisions

### 单一 PostgreSQL 支持矩阵

`database.default.link`只接受`pgsql:`前缀。`sqlite:`、`mysql:`和未知前缀在方言解析阶段失败，不进入启动、初始化、mock 加载、cluster coordination 或业务运行流程。配置模板、镜像运行配置、README、CI 和测试文档统一使用`PostgreSQL-only`口径。

### 保留方言边界但移除 SQLite 实现

`apps/lina-core/pkg/dialect/`保留为公共稳定边界，负责数据库准备、数据库版本查询、表元数据查询、驱动错误分类和只读 SQL 分类。该边界只保留`PostgreSQL`实现以及明确的不支持错误，不再包含`SQLite`DDL 转译器、错误分类和启动降级钩子。

### 初始化和 mock 命令职责分离

`init`连接系统库`postgres`执行建库、删库和重建，然后运行宿主与插件初始化 SQL。`mock`只连接已初始化的业务库并加载 mock 数据，不负责准备数据库。所有 SQL 资源使用受治理的`PostgreSQL 14+`语法子集，并保持可重复执行。

### 易失性表收敛为 PostgreSQL 持久表

`sys_online_session`、`sys_locker`和`sys_kv_cache`不再依赖`MEMORY`或`SQLite`降级语义，统一通过`PostgreSQL`普通表、过期字段和清理任务自然收敛。插件缓存服务在单机模式下使用 SQL 表后端，在集群模式下交给 coordination KV 后端；缓存仍然是有损缓存，不作为业务权威数据源。

### 集群和发布链路先校验数据库边界

`cluster.enabled=true`只在`PostgreSQL`支持矩阵内生效。不支持方言必须在 coordination 或业务服务启动前失败。CI、nightly、release、共享验证模板和镜像说明均不再包含`SQLite`smoke 或 package script 入口。

## Risks / Trade-offs

- 开发者失去免外部数据库的本地演示路径 → 通过标准`PostgreSQL`初始化命令、配置模板和文档降低新环境成本。
- 方言抽象可能被误解为多数据库兼容承诺 → 规范明确`pkg/dialect`是驱动治理边界，不代表运行时支持多方言。
- 删除`SQLite`测试通道减少一个轻量 smoke 路径 → 主干验证统一转向`PostgreSQL`，避免轻量路径掩盖真实生产数据库问题。
