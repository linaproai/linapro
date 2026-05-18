# 数据库方言抽象规范

## Purpose

定义宿主数据库方言抽象层，用于统一收敛 PostgreSQL 与 SQLite 等数据库引擎差异，并为 SQL 转译、数据库准备、启动期钩子和表元数据查询提供稳定的公共接缝。

## Requirements
### Requirement: 宿主必须通过统一的方言抽象层收敛数据库引擎差异

系统 SHALL 在 `apps/lina-core/pkg/dialect/` 提供公共稳定的 `Dialect` 接口与方言辅助能力作为数据库引擎差异的唯一收敛点。所有数据库引擎相关的差异化行为（DDL 转译、数据库准备、集群能力查询、启动期钩子、驱动错误分类、数据库版本查询、表元数据查询）必须通过该包暴露，业务模块（`controller` / `service` / `model` / `dao`）不得在自身代码路径中出现 `if isPostgres / if isSQLite` 等数据库引擎判断。`pkg/dialect` 的公开签名 SHALL 只依赖稳定窄接口，不得暴露宿主 `internal` 包中的具体服务类型，也不得导出 PostgreSQL / SQLite 方言具体实现类型;具体方言实现应收敛在 `pkg/dialect/internal/postgres`、`pkg/dialect/internal/sqlite` 等内部子包中，由公共工厂与公共门面能力统一委托。

#### Scenario: 业务模块不感知数据库引擎差异
- **当** 业务模块（如 `user` / `role` / `dict` / `kvcache` / `locker`）通过 DAO 层执行查询、写入、更新、删除操作时
- **则** 业务代码不包含针对数据库引擎的分支判断
- **且** 同一份业务代码在 PostgreSQL 和 SQLite 两种引擎下行为一致

#### Scenario: 所有方言相关行为通过 Dialect 接口暴露
- **当** 宿主需要执行"DDL 转译 / 数据库准备 / 集群能力查询 / 启动期钩子 / 数据库版本查询 / 表元数据查询"中的任一行为时
- **则** 调用方通过 `dialect.From(link)` 或 `dialect.FromDatabase(db)` 获取当前方言实例
- **且** 调用方仅依赖 `Dialect` 接口的方法签名，不依赖具体实现的内部细节

#### Scenario: 具体方言实现不作为公共 API 暴露
- **当** 宿主、插件生命周期或工具链代码导入 `apps/lina-core/pkg/dialect` 时
- **则** 公共包不导出 `PostgresDialect` / `SQLiteDialect` 等具体实现类型
- **且** PostgreSQL / SQLite 的 DDL 转译、数据库准备、启动期行为、驱动错误分类、表元数据查询实现分别维护在 `pkg/dialect/internal/postgres` 与 `pkg/dialect/internal/sqlite` 内部子包中
- **且** 调用方只能通过 `Dialect` 接口、`dialect.From(link)` / `dialect.FromDatabase(db)` 工厂函数和 `dialect.IsRetryableWriteConflict(err)` / `dialect.SplitSQLStatements(content)` 等必要公共门面能力访问方言相关行为

#### Scenario: 驱动错误分类由 dialect 公共包提供
- **当** `kvcache incr` 等共享组件需要判断数据库写入冲突是否可重试
- **则** 调用方通过 `dialect.IsRetryableWriteConflict(err)` 判断
- **且** 调用方不得硬编码 PostgreSQL / SQLite 错误文案、错误码或具体驱动错误类型
- **且** `pkg/dialect` 使用驱动暴露的结构化错误码进行分类，错误文案匹配最多只能作为方言包内部的显式兜底

#### Scenario: 数据库版本查询由 dialect 公共包提供
- **当** 宿主系统信息或 `monitor-server` 服务监控需要展示数据库版本时
- **则** 调用方通过 `dialect.DatabaseVersion(ctx, db)` 或等价 `Dialect` 方法查询
- **且** PostgreSQL 方言使用 `SELECT version()` 或等价语句，返回包含 `PostgreSQL` 引擎名与版本号的字符串
- **且** SQLite 方言使用 `SELECT sqlite_version()`，返回包含 `SQLite` 引擎名与版本号的字符串
- **且** 任一方言下不得返回空版本
- **且** 返回给页面的版本文本必须包含数据库引擎名称与非空版本号

#### Scenario: 表元数据查询由 dialect 公共包提供
- **当** 插件 data service 等组件需要查询数据表名与表注释时
- **则** 调用方通过 `Dialect.QueryTableMetadata(ctx, db, schema, names)` 查询
- **且** 调用方不得直接编写 `information_schema.TABLES` / `pg_catalog.pg_description` / `sqlite_master` 等方言特有 SQL
- **且** PostgreSQL 实现使用 `information_schema.tables` JOIN `pg_class` 的 `obj_description(oid)` 查询表注释
- **且** SQLite 实现从 `sqlite_master` 查询表名，注释字段返回空字符串（SQLite 不支持表注释）

#### Scenario: dialect 公共包不暴露宿主 internal 具体类型
- **当** 插件生命周期、初始化命令或工具链代码导入 `apps/lina-core/pkg/dialect` 时
- **则** 公开接口不要求调用方引用 `apps/lina-core/internal/...` 下的具体服务类型
- **且** 启动期配置覆盖能力通过 `dialect.RuntimeConfig` 等窄接口适配
- **且** 宿主 `config.Service` 可在内部实现该窄接口后传入 `Dialect.OnStartup`

### Requirement: 方言根据数据库链接前缀自动分发

系统 SHALL 根据 `database.default.link` 配置的协议头自动选择对应的方言实现。`pgsql:` 前缀分发到 PostgreSQL 方言实现，`sqlite:` 前缀分发到 SQLite 方言实现。`mysql:` 前缀 MUST 被识别为不支持的方言并返回明确错误（与未知前缀同样处理），不得静默回退。未识别的前缀必须返回包含前缀名与已支持前缀列表的明确错误。调用方不得依赖或断言具体实现类型，只能依赖 `Dialect` 接口行为。

#### Scenario: PostgreSQL 链接被识别为 PostgreSQL 方言
- **当** 配置文件 `database.default.link` 以 `pgsql:` 开头时
- **则** `dialect.From(link)` 返回实现 `Dialect` 接口的 PostgreSQL 方言实例
- **且** `Name()` 返回字符串 `"postgres"`
- **且** `SupportsCluster()` 返回 `true`

#### Scenario: SQLite 链接被识别为 SQLite 方言
- **当** 配置文件 `database.default.link` 以 `sqlite:` 开头时
- **则** `dialect.From(link)` 返回实现 `Dialect` 接口的 SQLite 方言实例
- **且** `Name()` 返回字符串 `"sqlite"`
- **且** `SupportsCluster()` 返回 `false`

#### Scenario: MySQL 链接被显式拒绝
- **当** 配置文件 `database.default.link` 以 `mysql:` 开头时
- **则** `dialect.From(link)` 返回明确错误，错误消息包含"mysql 方言不再支持"或等价说明
- **且** 错误消息列出当前已支持的前缀（`pgsql:`、`sqlite:`）
- **且** 系统不静默回退到任何默认方言

#### Scenario: 未识别的链接前缀
- **当** 配置文件 `database.default.link` 以未识别的前缀开头时
- **则** `dialect.From(link)` 返回包含前缀名与已支持前缀列表的明确错误
- **且** 系统不静默回退到任何默认方言

### Requirement: PostgreSQL 方言 DDL 转译为无操作

`Dialect.TranslateDDL(ctx, sourceName, ddl)` SHALL 接收调用方传入的 `sourceName` 诊断名。`sourceName` MUST 是源 SQL 文件路径、嵌入资产路径或调用方构造的稳定描述，用于在错误消息中定位失败来源。PostgreSQL 方言的 `TranslateDDL` SHALL 直接返回输入字符串，不做任何修改。这保证了 PostgreSQL 作为单一 SQL 源方言时生产路径无翻译开销。

#### Scenario: PostgreSQL 方言转译保持原文
- **当** PostgreSQL 方言实例的 `TranslateDDL(ctx, sourceName, ddl)` 被调用时
- **则** 返回值与输入 `ddl` 字节级别完全一致
- **且** 不返回错误（除非输入本身为 `nil` 或空字符串等显式无效输入）
- **且** `sourceName` 不影响 PostgreSQL no-op 转译结果

### Requirement: SQLite 方言 DDL 转译必须覆盖项目 PG SQL 子集

SQLite 方言的 `TranslateDDL(ctx, sourceName, ddl)` SHALL 将单一 PostgreSQL 14+ 方言来源的 DDL / seed / mock SQL 转译为可在 SQLite 上成功执行的语句。转译必须覆盖项目当前 SQL 文件中实际使用的所有 PostgreSQL 语法，包括 `INT/BIGINT GENERATED ALWAYS AS IDENTITY [PRIMARY KEY]`、`SMALLINT` / `BIGINT` / `INT` / `VARCHAR(n)` / `CHAR(n)` / `TEXT` / `BYTEA` / `TIMESTAMP` / `DECIMAL(m,n)` / `NUMERIC` / `REAL` / `DOUBLE PRECISION` 类型族、双引号包裹的保留字标识符、`COMMENT ON TABLE/COLUMN ... IS '...'` 独立注释语句、`CREATE INDEX [IF NOT EXISTS] ... ON ...` 与 `CREATE UNIQUE INDEX [IF NOT EXISTS] ... ON ...` 独立索引语句、`INSERT INTO ... ON CONFLICT DO NOTHING`。转译覆盖范围 SHALL 以当前宿主安装 SQL、插件安装 SQL、宿主 mock SQL、插件 mock SQL 与插件卸载 SQL 的真实写法为验收基准。SQLite 不支持 `TRUNCATE` 语法，本次变更不得在应用启动路径或 SQL 源中依赖 `TRUNCATE TABLE`。

#### Scenario: GENERATED IDENTITY 主键被改写
- **当** 输入 DDL 包含 `id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY` 或 `id BIGINT GENERATED ALWAYS AS IDENTITY` 时
- **则** 转译结果中该列定义为 `id INTEGER PRIMARY KEY AUTOINCREMENT`
- **且** 不保留 `BIGINT` / `INT` / `GENERATED` / `ALWAYS` / `IDENTITY` 等 SQLite 不需要的修饰符
- **且** SQLite 自增列必须类型为 `INTEGER`（不能是 `BIGINT`），符合 SQLite 自增约束

#### Scenario: 整数类型被映射为 INTEGER
- **当** 输入 DDL 包含 `INT` / `BIGINT` / `SMALLINT` 任一非主键列定义时
- **则** 转译结果对应列映射为 `INTEGER`

#### Scenario: 字符串类型被映射为 TEXT
- **当** 输入 DDL 包含 `VARCHAR(n)` / `CHAR(n)` / `TEXT` 任一列定义时
- **则** 转译结果对应列映射为 `TEXT`

#### Scenario: BYTEA 被映射为 BLOB
- **当** 输入 DDL 包含 `BYTEA` 列定义时
- **则** 转译结果对应列映射为 `BLOB`

#### Scenario: DECIMAL 被映射为 NUMERIC
- **当** 输入 DDL 包含 `DECIMAL(m, n)` 或 `NUMERIC(m, n)` 列定义时
- **则** 转译结果对应列映射为 `NUMERIC`

#### Scenario: TIMESTAMP 列保留为 SQLite 兼容类型
- **当** 输入 DDL 包含 `TIMESTAMP` 列定义时
- **则** 转译结果对应列映射为 `DATETIME` 或 `TEXT`，由 GoFrame 自动序列化时间值

#### Scenario: 双引号保留字列名保留双引号
- **当** 输入 DDL 包含 `"key" VARCHAR(128) NOT NULL` 等双引号包裹的列定义时
- **则** 转译结果保留双引号包裹（SQLite 也接受双引号作为标识符引号）
- **且** DML 中的双引号列引用同样保留

#### Scenario: COMMENT ON 语句被丢弃
- **当** 输入 DDL 包含 `COMMENT ON TABLE <表名> IS '...';` 或 `COMMENT ON COLUMN <表名>.<列名> IS '...';` 整句时
- **则** 转译结果不包含这些语句
- **且** 转译器不报错（SQLite 不支持表/列注释，丢弃即正确语义）

#### Scenario: 自定义排序规则被拒绝
- **当** 输入 DDL 包含 `CREATE COLLATION` 或自定义列级排序规则时
- **则** 转译器返回错误
- **且** 项目 SQL 源不得通过 SQLite 翻译器隐式模拟自定义排序规则语义

#### Scenario: INSERT ... ON CONFLICT DO NOTHING 被保留
- **当** 输入 DDL 包含 `INSERT INTO <表名> (...) VALUES (...) ON CONFLICT DO NOTHING;` 时
- **则** 转译结果保留原文（SQLite 3.24+ 兼容）
- **且** 写入语义（unique 约束冲突时跳过）保持等价

#### Scenario: CREATE INDEX 与 CREATE UNIQUE INDEX 被保留
- **当** 输入 DDL 包含 `CREATE INDEX [IF NOT EXISTS] <索引名> ON <表名> (<列名>);` 或 `CREATE UNIQUE INDEX [IF NOT EXISTS] ...` 时
- **则** 转译结果保留原语句（SQLite 兼容）

#### Scenario: 已有 SQL 文件全部翻译成功
- **当** 转译器接收当前宿主 12 个 SQL 文件、宿主 mock SQL、8 个包含实际 SQL 资源的插件 install/mock-data/uninstall SQL 全部内容时
- **则** 翻译产出对应的 SQLite 兼容语句序列
- **且** 在临时 SQLite 数据库上逐句执行均成功

### Requirement: DDL 转译失败时必须返回明确错误

SQLite 方言的 `TranslateDDL(ctx, sourceName, ddl)` 在遇到当前实现未覆盖的 PostgreSQL 语法时 SHALL 返回包含 `sourceName`、行号定位提示与未覆盖语法关键字的明确错误，不得静默丢弃或产生无效 SQL。被翻译器拒绝的 PG 高级特性包括但不限于：`JSONB` / `JSON` 类型与运算符、数组类型、`GENERATED ALWAYS AS (expr) STORED` 计算列、`CREATE EXTENSION`、`CREATE FUNCTION`、`CREATE TRIGGER`、`CREATE TYPE`、`MERGE` 语句、`WITH RECURSIVE`、`LATERAL` 联接、`PARTITION OF` 子句、`SERIAL` / `BIGSERIAL` 简写。

#### Scenario: 转译器遇到未覆盖的 PG 语法
- **当** 输入 DDL 包含未在覆盖范围内的 PG 高级特性（如 `JSONB` / `CREATE TRIGGER` / `MERGE` / `SERIAL` 等）时
- **则** 转译器返回错误，错误消息包含 `sourceName`、行号提示与未覆盖的关键字
- **且** 调用方（`cmd init` / `cmd mock` / 插件 install pipeline）将错误向上传播
- **且** 系统不执行任何已部分转译的 SQL 内容

### Requirement: 方言必须暴露数据库准备入口

`Dialect.PrepareDatabase(ctx, link, rebuild)` SHALL 负责在执行 DDL 资源前完成方言相关的数据库准备工作。PostgreSQL 方言通过连接系统库 `postgres` 执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`；SQLite 方言执行父目录创建（`mkdir -p`）与可选的数据库文件删除。PostgreSQL 目标数据库创建后，后续宿主 init SQL SHALL 直接创建业务表、索引、注释并写入 seed 数据，不创建自定义排序规则。

#### Scenario: PostgreSQL 方言准备数据库
- **当** PostgreSQL 方言实例的 `PrepareDatabase(ctx, link, rebuild=false)` 被调用且目标库不存在时
- **则** 系统连接到 PG 系统库 `postgres`
- **且** 执行 `CREATE DATABASE <目标库名> ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`
- **且** 不删除已存在的数据库
- **且** 后续宿主 init SQL 不创建自定义排序规则

#### Scenario: PostgreSQL 方言重建数据库
- **当** PostgreSQL 方言实例的 `PrepareDatabase(ctx, link, rebuild=true)` 被调用且目标库存在时
- **则** 系统连接到 PG 系统库 `postgres`
- **且** 先执行 `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()` 终止活跃连接
- **且** 再执行 `DROP DATABASE IF EXISTS <目标库名>`
- **且** 再执行 `CREATE DATABASE <目标库名> ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`
- **且** 启动日志输出明确的 rebuild 警告
- **且** 后续宿主 init SQL 不创建自定义排序规则

#### Scenario: SQLite 方言准备数据库文件
- **当** SQLite 方言实例的 `PrepareDatabase(ctx, link, rebuild=false)` 被调用且数据库文件父目录不存在时
- **则** 系统自动 `mkdir -p` 父目录
- **且** 数据库文件由后续 DDL 执行自动创建（GoFrame 驱动行为）
- **且** 已存在的数据库文件不被删除

#### Scenario: SQLite 方言重建数据库
- **当** SQLite 方言实例的 `PrepareDatabase(ctx, link, rebuild=true)` 被调用时
- **则** 系统先删除数据库文件（含 WAL / SHM 等附属文件）
- **且** 再确保父目录存在
- **且** 启动日志输出明确的 rebuild 警告

#### Scenario: PostgreSQL 系统库连接失败
- **当** PostgreSQL `PrepareDatabase` 无法连接到系统库 `postgres`（PG 服务未启动、网络不通、认证失败）时
- **则** 系统返回包含目标 PG 主机、端口与具体错误的明确错误
- **且** 不继续后续 DDL 执行

#### Scenario: SQLite 父目录不可创建
- **当** SQLite 数据库文件父目录创建失败（权限不足、磁盘满等）时
- **则** 系统返回包含目标路径的明确错误
- **且** 不继续后续 DDL 执行

### Requirement: 方言必须提供启动期钩子

`Dialect.OnStartup(ctx, runtime)` SHALL 在宿主启动 bootstrap 阶段被调用一次。`runtime` SHALL 是 `pkg/dialect` 中定义的稳定窄接口，至少提供方言锁定 `cluster.enabled` 所需的方法。PostgreSQL 方言为 no-op；SQLite 方言负责执行"强制覆盖 cluster.enabled=false + 输出警告日志"等启动期专属行为。该钩子的调用时机必须早于任何 cluster 相关初始化。

#### Scenario: PostgreSQL 启动期钩子无副作用
- **当** PostgreSQL 方言实例的 `OnStartup(ctx, runtime)` 被调用时
- **则** 钩子立即返回 `nil`
- **且** 不修改任何配置项
- **且** 不输出任何警告级别日志

#### Scenario: SQLite 启动期钩子锁定集群配置
- **当** SQLite 方言实例的 `OnStartup(ctx, runtime)` 被调用时
- **则** `configSvc.IsClusterEnabled(ctx)` 在该钩子调用后稳定返回 `false`
- **且** 该覆盖优先级高于 `config.yaml` 中的 `cluster.enabled` 显式声明
- **且** 钩子向终端输出启动提示日志，明确告知 SQLite 模式、cluster 锁定原因、单机部署限制、不得用于生产

#### Scenario: 启动期钩子在集群初始化前执行
- **当** 宿主以 SQLite 模式启动时
- **则** `OnStartup` 在 `cluster.Service` 启动选举循环前被调用
- **且** 后续 cluster 相关组件读取到的 `IsClusterEnabled` 已为 `false`
- **且** 不会出现"先启动选举循环再被关闭"的中间状态

### Requirement: Dialect 接口必须提供表元数据查询能力

系统 SHALL 在 `Dialect` 接口中提供 `QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error)` 方法，用于跨方言查询数据表名与表注释。`TableMeta` SHALL 至少包含 `TableName string` 与 `TableComment string` 字段。调用方（如 plugin data service）SHALL 通过该方法查询表元数据，不得直接编写方言特有的 SQL（`information_schema.TABLES` / `pg_catalog.pg_description` / `sqlite_master`）。

#### Scenario: PostgreSQL 实现使用 information_schema 与 pg_description
- **当** 调用方在 PostgreSQL 方言下调用 `QueryTableMetadata(ctx, db, "public", []string{"sys_user", "sys_role"})`
- **则** 实现 SQL 联接 `information_schema.tables` 与 `pg_class`
- **且** 通过 `obj_description(c.oid)` 获取表注释
- **且** 返回的 `TableMeta` 数组包含传入表名对应的 `TableName` 与 `TableComment`（无注释时为空字符串）

#### Scenario: SQLite 实现使用 sqlite_master
- **当** 调用方在 SQLite 方言下调用 `QueryTableMetadata(ctx, db, schema, names)`
- **则** 实现从 `sqlite_master` 查询 `name` 列得到表名
- **且** 返回的 `TableMeta` 数组中所有 `TableComment` 字段固定为空字符串
- **且** schema 参数在 SQLite 实现中可被忽略（SQLite 无 schema 概念）

#### Scenario: 调用方不直接编写方言 SQL
- **当** 检查 `apps/lina-core/internal/service/plugin/plugin_data_table_comment.go` 等调用代码
- **则** 代码 MUST NOT 出现 `information_schema.TABLES` / `pg_catalog.pg_description` / `sqlite_master` 字面量
- **且** 代码 MUST 通过 `dialect.FromDatabase(g.DB()).QueryTableMetadata(...)` 查询

#### Scenario: 表名传入空数组
- **当** 调用方传入空的 `tableNames` 数组
- **则** 方法返回空的 `[]TableMeta`
- **且** 不返回错误

#### Scenario: 不存在的表名被静默跳过
- **当** 调用方传入的 `tableNames` 中部分表名在数据库中不存在
- **则** 返回的 `[]TableMeta` 仅包含实际存在的表
- **且** 不为不存在的表名返回任何记录或错误
