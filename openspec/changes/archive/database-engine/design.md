## Context

LinaPro数据库治理最终收敛到`PostgreSQL 14+`。早期为降低开发门槛保留过`SQLite`转译和多方言分支，但这会持续牵制SQL源语法、初始化命令、插件SQL生命周期、缓存协调、CI和文档口径。项目没有兼容负担，因此运行时数据库支持矩阵直接收敛为单一`PostgreSQL`基线。

## Goals / Non-Goals

**Goals:**

- 以`PostgreSQL 14+`作为唯一受支持运行时数据库。
- 保留`pkg/dialect`作为数据库驱动、元数据查询、只读SQL分类和初始化准备的稳定边界。
- 删除`SQLite`运行、转译和启动降级残留。
- 让插件SQL、宿主SQL、初始化、mock和易失性表治理都围绕同一数据库语义验证。

**Non-Goals:**

- 不提供`SQLite`、`MySQL`或其他数据库方言兼容层。
- 不维护旧数据迁移或历史开发库升级路径。
- 不以数据库兼容为由限制宿主和插件 SQL 使用受治理的`PostgreSQL`能力。
- 不在本分组长期保存集群、插件、发布、i18n、角色、字典或基础工程能力的完整历史规范。

## Decisions

### 单一`PostgreSQL`支持矩阵

`database.default.link`只接受`pgsql:`前缀。`sqlite:`、`mysql:`和未知前缀在方言解析阶段失败，不进入启动、初始化、mock加载、集群协调或业务运行流程。方言失败要早于任何Redis探活、集群配置覆盖、业务路由注册或插件运行时启动，避免不支持数据库被单机降级或警告绕过。

### 保留方言边界但移除`SQLite`实现

`apps/lina-core/pkg/dialect/`保留为公共稳定边界，负责数据库准备、数据库版本查询、表元数据查询、驱动错误分类和只读SQL分类。该边界只保留`PostgreSQL`实现以及明确的不支持错误，不再包含`SQLite`DDL转译器、错误分类和启动降级钩子。方言抽象是驱动治理接缝，不是多数据库兼容承诺。

方言抽象的最终需求全文与当前主规范一致，归档分组不再保存重复`spec.md`副本；历史价值由本设计摘要和数据库引擎owner能力清单承载。

### 初始化和`mock`命令职责分离

`init`连接系统库`postgres`执行建库、删库和重建，然后运行宿主与插件初始化SQL。`mock`只连接已初始化的业务库并加载mock数据，不负责准备数据库。两类命令都必须显式确认、显式选择嵌入式或本地SQL资源来源、逐文件逐语句快速失败，并在执行SQL前通过当前方言入口处理。

### SQL源使用受治理的`PostgreSQL`子集

宿主与插件SQL源使用`PostgreSQL 14+`语法子集：使用`IDENTITY`自增主键、PG标准类型、独立`COMMENT ON`、独立`CREATE INDEX`、`ON CONFLICT DO NOTHING`幂等插入、双引号列标识符和默认deterministic文本比较语义；不得包含`AUTO_INCREMENT`、`ENGINE=`、`INSERT IGNORE`、`ON DUPLICATE KEY UPDATE`、`ON UPDATE CURRENT_TIMESTAMP`、`CREATE DATABASE`、`USE`或未经单独评估的PG高级特性。数据库创建和重建由`Dialect.PrepareDatabase`负责，SQL源只关心表、索引、注释和数据。

### 易失性表收敛为`PostgreSQL`持久表

`sys_online_session`、`sys_locker`和`sys_kv_cache`不再依赖`MEMORY`、`UNLOGGED`、`TEMPORARY`或`SQLite`降级语义，统一通过`PostgreSQL`普通持久表、过期字段和清理任务自然收敛。宿主启动、重启、滚动发布、leader切换和插件运行时启动不得清空这些表；读取路径必须先判断过期，会话、锁和KV缓存分别通过`last_active_time`、`expire_time`和`expire_at`判断有效性。

## Cross-Domain Impacts

- 集群部署和coordination配置受数据库收敛影响：非`PostgreSQL`链接必须在coordination启动前失败，`cluster.enabled=true`只在`PostgreSQL`支持矩阵内继续进入Redis coordination流程；当前契约由`openspec/specs/cluster-deployment-mode/spec.md`和`openspec/specs/cluster-coordination-config/spec.md`承载，历史owner为`archive/distributed-infra`。
- 插件缓存和插件manifest SQL生命周期受数据库收敛影响：单机插件缓存继续使用共享`PostgreSQL`表后端，集群模式使用coordination KV，插件SQL不再以`SQLite`转译为目标；当前契约由`openspec/specs/plugin-cache-service/spec.md`和`openspec/specs/plugin-manifest-lifecycle/spec.md`承载，历史owner为`archive/plugin-framework`。
- 发布、CI、installer和测试入口移除`SQLite`smoke或脚本通道，统一转向`PostgreSQL-only`验证；当前契约由`openspec/specs/release-image-build/spec.md`和`openspec/specs/framework-bootstrap-installer/spec.md`承载，历史owner为`archive/devops-tooling`。
- README双语镜像和管理工作台首次语言识别只承载数据库口径同步的文档/i18n影响，不属于数据库引擎owner能力；当前契约由`openspec/specs/readme-localization-governance/spec.md`和`openspec/specs/management-workbench-i18n/spec.md`承载，历史owner为`archive/i18n`。
- 项目初始化中的数据库配置场景随数据库基线改为`PostgreSQL-only`，但项目脚手架、前后端启动和开发代理仍归基础工程能力；当前契约由`openspec/specs/project-setup/spec.md`承载，历史owner为`archive/foundation`。
- 角色授权和字典复用读取只是数据库收敛期间被同步校验的业务能力，不属于数据库引擎owner；当前契约由`openspec/specs/role-management/spec.md`和`openspec/specs/dictionary-management/spec.md`承载，历史owner分别为`archive/user-management`和`archive/org-structure`。

## Risks / Trade-offs

- 开发者失去免外部数据库的本地演示路径 → 通过标准`PostgreSQL`初始化命令、配置模板和文档降低新环境成本。
- 方言抽象可能被误解为多数据库兼容承诺 → 规范明确`pkg/dialect`是驱动治理边界，不代表运行时支持多方言。
- 删除`SQLite`测试通道减少一个轻量smoke路径 → 主干验证统一转向`PostgreSQL`，避免轻量路径掩盖真实生产数据库问题。
