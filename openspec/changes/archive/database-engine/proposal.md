## Why

LinaPro在数据库引擎治理上经历了从`MySQL`语法遗留、多方言转译和`SQLite`开发演示链路，收敛到`PostgreSQL 14+`单一运行时基线的过程。早期兼容路径降低了局部启动门槛，但也把SQL源语法、初始化命令、插件SQL生命周期、缓存表语义、CI、镜像交付和文档口径长期绑在多数据库支持矩阵上，导致生产语义和测试语义不一致。

最终目标是让数据库层成为稳定、清晰、可持续维护的框架基础：运行时只支持`PostgreSQL 14+`，初始化与mock命令通过统一方言边界准备和执行SQL，宿主与插件SQL源使用受治理的`PostgreSQL`语法子集，原易失性表通过持久表和自然过期语义收敛。`MySQL`和`SQLite`不再作为受支持运行时数据库，也不作为SQL源约束前提。

## What Changes

- 运行时数据库支持统一为`PostgreSQL 14+`，`database.default.link`仅支持`pgsql:`前缀。
- `sqlite:`、`mysql:`和其他未知前缀在方言解析阶段明确失败，不进入启动、初始化、mock加载、集群协调或业务运行流程。
- 保留`apps/lina-core/pkg/dialect/`作为公共稳定边界，统一承载数据库准备、SQL入口、数据库版本查询、表元数据查询、驱动错误分类和驱动/ORM只读SQL分类。
- 删除`SQLite`方言实现、DDL转译器、错误分类和启动降级钩子，只保留`PostgreSQL`实现以及明确的不支持错误。
- `init`通过连接系统库`postgres`执行`PrepareDatabase`完成建库、删库和重建；`mock`只连接已初始化数据库，不负责准备数据库。
- 宿主与插件SQL资源统一使用受治理的`PostgreSQL 14+`语法子集编写，移除`MySQL`专属语法遗留，不再为了`SQLite`转译能力限制SQL源。
- `sys_online_session`、`sys_locker`、`sys_kv_cache`统一作为`PostgreSQL`普通持久表，依赖过期字段、TTL清理和读取时过期判断自然收敛。
- 配置模板、测试入口、镜像运行配置和文档统一表达`PostgreSQL-only`口径；集群、插件缓存、发布和README治理的完整历史分别由对应owner分组承载。

## Capabilities

### New Capabilities

- `postgresql-only-database-support`：定义运行时、初始化、测试、CI与交付链路只支持`PostgreSQL 14+`的能力边界。

### Modified Capabilities

- `database-dialect-abstraction`：保留方言抽象边界，但删除`SQLite`实现与转译路径，只保留`PostgreSQL`实现、元数据查询和只读SQL分类能力。
- `database-bootstrap-commands`：初始化与mock命令仅围绕`PostgreSQL`准备与SQL执行工作流运行，不再支持`SQLite`准备或转译。
- `sql-source-syntax`：继续以受治理的`PostgreSQL 14+`子集约束SQL源，但不再把`SQLite`运行时支持作为前提。
- `volatile-table-bootstrap`：易失性表自然过期规范收敛到`PostgreSQL`持久表路径。

## Impact

- 数据库引擎历史 owner 仅保留`postgresql-only-database-support`、`database-dialect-abstraction`、`database-bootstrap-commands`、`sql-source-syntax`和`volatile-table-bootstrap`。
- 集群协调、插件缓存、插件manifest生命周期、发布镜像、README本地化、项目初始化、角色授权和字典复用读取的完整规范分别由`distributed-infra`、`plugin-framework`、`devops-tooling`、`i18n`、`foundation`、`user-management`和`org-structure`承载；本分组只记录数据库收敛对这些领域的交叉影响。
- 本归档压缩不修改运行时代码、HTTP API、数据库、缓存、数据权限、前端UI、插件目录、运行时文案、运行期依赖、开发工具跨平台入口或生产构建。
- 本归档压缩不修改运行时语言包、`manifest/i18n`、`apidoc i18n JSON`、菜单、路由、按钮、错误消息或用户可见UI文案；影响仅限中文OpenSpec归档文档。
