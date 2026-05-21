## MODIFIED Requirements

### Requirement: 宿主必须通过统一的方言抽象层收敛数据库引擎差异

系统 SHALL 在 `apps/lina-core/pkg/dialect/` 提供公共稳定的 `Dialect` 接口与方言辅助能力作为数据库引擎差异的唯一收敛点。所有数据库引擎相关的差异化行为（DDL 转译、数据库准备、集群能力查询、启动期钩子、驱动错误分类、数据库版本查询、表元数据查询）必须通过该包暴露，业务模块不得在自身代码路径中出现 `if isPostgres / if isSQLite` 等数据库引擎判断。

#### Scenario: 业务模块不感知数据库引擎差异
- **当** 业务模块通过 DAO 层执行查询、写入、更新、删除操作时
- **则** 业务代码不包含针对数据库引擎的分支判断
- **且** 同一份业务代码在 PostgreSQL 和 SQLite 两种引擎下行为一致

#### Scenario: 所有方言相关行为通过 Dialect 接口暴露
- **当** 宿主需要执行 DDL 转译、数据库准备、集群能力查询、启动期钩子、数据库版本查询或表元数据查询时
- **则** 调用方通过 `dialect.From(link)` 或 `dialect.FromDatabase(db)` 获取当前方言实例
- **且** 调用方仅依赖 `Dialect` 接口的方法签名

#### Scenario: 驱动错误分类由 dialect 公共包提供
- **当** 共享组件需要判断数据库写入冲突是否可重试
- **则** 调用方通过 `dialect.IsRetryableWriteConflict(err)` 判断
- **且** 调用方不得硬编码 PostgreSQL / SQLite 错误文案或具体驱动错误类型

#### Scenario: 表元数据查询由 dialect 公共包提供
- **当** 插件 data service 等组件需要查询数据表名与表注释时
- **则** 调用方通过 `Dialect.QueryTableMetadata(ctx, db, schema, names)` 查询
- **且** PostgreSQL 实现使用 `information_schema.tables` JOIN `pg_class` 的 `obj_description(oid)` 查询表注释
- **且** SQLite 实现从 `sqlite_master` 查询表名，注释字段返回空字符串

### Requirement: 方言根据数据库链接前缀自动分发

系统 SHALL 根据 `database.default.link` 配置的协议头自动选择对应的方言实现。`pgsql:` 前缀分发到 PostgreSQL 方言，`sqlite:` 前缀分发到 SQLite 方言。`mysql:` 前缀 MUST 被识别为不支持的方言并返回明确错误。

#### Scenario: PostgreSQL 链接被识别为 PostgreSQL 方言
- **当** 配置文件 `database.default.link` 以 `pgsql:` 开头时
- **则** `dialect.From(link)` 返回 PostgreSQL 方言实例
- **且** `Name()` 返回 `"postgres"`
- **且** `SupportsCluster()` 返回 `true`

#### Scenario: SQLite 链接被识别为 SQLite 方言
- **当** 配置文件 `database.default.link` 以 `sqlite:` 开头时
- **则** `dialect.From(link)` 返回 SQLite 方言实例
- **且** `Name()` 返回 `"sqlite"`
- **且** `SupportsCluster()` 返回 `false`

#### Scenario: MySQL 链接被显式拒绝
- **当** 配置文件 `database.default.link` 以 `mysql:` 开头时
- **则** `dialect.From(link)` 返回明确错误
- **且** 错误消息列出当前已支持的前缀（`pgsql:`、`sqlite:`）

### Requirement: PostgreSQL 方言 DDL 转译为无操作

PostgreSQL 方言的 `TranslateDDL` SHALL 直接返回输入字符串，不做任何修改。这保证了 PostgreSQL 作为单一 SQL 源方言时生产路径无翻译开销。

### Requirement: SQLite 方言 DDL 转译必须覆盖项目 PG SQL 子集

SQLite 方言的 `TranslateDDL` SHALL 将单一 PostgreSQL 14+ 方言来源的 DDL / seed / mock SQL 转译为可在 SQLite 上成功执行的语句。转译必须覆盖项目当前 SQL 文件中实际使用的所有 PostgreSQL 语法。

#### Scenario: GENERATED IDENTITY 主键被改写
- **当** 输入 DDL 包含 `id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY` 时
- **则** 转译结果中该列定义为 `id INTEGER PRIMARY KEY AUTOINCREMENT`

#### Scenario: 整数类型被映射为 INTEGER
- **当** 输入 DDL 包含 `INT` / `BIGINT` / `SMALLINT` 任一非主键列定义时
- **则** 转译结果对应列映射为 `INTEGER`

#### Scenario: COMMENT ON 语句被丢弃
- **当** 输入 DDL 包含 `COMMENT ON TABLE` 或 `COMMENT ON COLUMN` 语句时
- **则** 转译结果不包含这些语句
- **且** 转译器输出 Debug 日志便于诊断

#### Scenario: INSERT ... ON CONFLICT DO NOTHING 被保留
- **当** 输入 DDL 包含 `INSERT INTO ... ON CONFLICT DO NOTHING` 时
- **则** 转译结果保留原文（SQLite 3.24+ 兼容）

#### Scenario: 已有 SQL 文件全部翻译成功
- **当** 转译器接收当前宿主与插件全部 SQL 文件内容时
- **则** 翻译产出对应的 SQLite 兼容语句序列
- **且** 在临时 SQLite 数据库上逐句执行均成功

### Requirement: DDL 转译失败时必须返回明确错误

SQLite 方言的 `TranslateDDL` 在遇到当前实现未覆盖的 PostgreSQL 语法时 SHALL 返回包含 `sourceName`、行号定位提示与未覆盖语法关键字的明确错误。

### Requirement: 方言必须暴露数据库准备入口

`Dialect.PrepareDatabase(ctx, link, rebuild)` SHALL 负责在执行 DDL 资源前完成方言相关的数据库准备工作。PostgreSQL 方言通过连接系统库执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`；SQLite 方言执行父目录创建与可选的数据库文件删除。

### Requirement: 方言必须提供启动期钩子

`Dialect.OnStartup(ctx, runtime)` SHALL 在宿主启动 bootstrap 阶段被调用一次。PostgreSQL 方言为 no-op；SQLite 方言负责执行"强制覆盖 cluster.enabled=false + 输出警告日志"等启动期专属行为。

## ADDED Requirements

### Requirement: Dialect 接口必须提供表元数据查询能力

系统 SHALL 在 `Dialect` 接口中提供 `QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error)` 方法，用于跨方言查询数据表名与表注释。`TableMeta` SHALL 至少包含 `TableName string` 与 `TableComment string` 字段。

#### Scenario: PostgreSQL 实现使用 information_schema 与 pg_description
- **当** 调用方在 PostgreSQL 方言下调用 `QueryTableMetadata` 时
- **则** 实现 SQL 联接 `information_schema.tables` 与 `pg_class`
- **且** 通过 `obj_description(c.oid)` 获取表注释

#### Scenario: SQLite 实现使用 sqlite_master
- **当** 调用方在 SQLite 方言下调用 `QueryTableMetadata` 时
- **则** 实现从 `sqlite_master` 查询表名
- **且** 注释字段固定为空字符串
