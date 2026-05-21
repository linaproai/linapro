## MODIFIED Requirements

### Requirement: 宿主必须通过统一的方言抽象层收敛数据库引擎差异

系统 SHALL 在 `apps/lina-core/pkg/dialect/` 提供公共稳定的 `Dialect` 接口与方言辅助能力作为数据库引擎差异的唯一收敛点。当前唯一支持的具体数据库方言 SHALL 为 PostgreSQL。所有数据库引擎相关的差异化行为（数据库准备、集群能力查询、启动期钩子、驱动错误分类、数据库版本查询、表元数据查询、驱动/ORM 只读 SQL 分类）必须通过该包暴露，业务模块（`controller` / `service` / `model` / `dao`）不得在自身代码路径中出现 `if isPostgres / if isSQLite` 等数据库引擎判断。`pkg/dialect` 的公开签名 SHALL 只依赖稳定窄接口，不得暴露宿主 `internal` 包中的具体服务类型，也不得导出 PostgreSQL 方言具体实现类型；具体方言实现应收敛在 `pkg/dialect/internal/postgres` 内部子包中，由公共工厂与公共门面能力统一委托。

#### Scenario: 业务模块不感知数据库引擎差异

- **当** 业务模块（如 `user` / `role` / `dict` / `kvcache` / `locker`）通过 DAO 层执行查询、写入、更新、删除操作时
- **则** 业务代码不包含针对数据库引擎的分支判断
- **且** 同一份业务代码只承诺在 PostgreSQL 支持矩阵下运行

#### Scenario: 所有方言相关行为通过 Dialect 接口暴露

- **当** 宿主需要执行“数据库准备 / 集群能力查询 / 启动期钩子 / 数据库版本查询 / 表元数据查询 / 驱动与 ORM 只读 SQL 分类”中的任一行为时
- **则** 调用方通过 `dialect.From(link)` 或 `dialect.FromDatabase(db)` 获取当前方言实例
- **且** 调用方仅依赖 `Dialect` 接口的方法签名，不依赖具体实现的内部细节

#### Scenario: 具体方言实现不作为公共 API 暴露

- **当** 宿主、插件生命周期或工具链代码导入 `apps/lina-core/pkg/dialect` 时
- **则** 公共包不导出 `PostgresDialect` 等具体实现类型
- **且** PostgreSQL 的数据库准备、启动期行为、驱动错误分类、表元数据查询、只读 SQL 分类实现维护在 `pkg/dialect/internal/postgres` 内部子包中
- **且** 调用方只能通过 `Dialect` 接口、`dialect.From(link)` / `dialect.FromDatabase(db)` 工厂函数和 `dialect.IsRetryableWriteConflict(err)` / `dialect.SplitSQLStatements(content)` 等必要公共门面能力访问方言相关行为

#### Scenario: 驱动错误分类由 dialect 公共包提供

- **当** `kvcache incr` 等共享组件需要判断数据库写入冲突是否可重试
- **则** 调用方通过 `dialect.IsRetryableWriteConflict(err)` 判断
- **且** 调用方不得硬编码 PostgreSQL 错误文案、错误码或具体驱动错误类型
- **且** `pkg/dialect` 使用驱动暴露的结构化错误码进行分类，错误文案匹配最多只能作为方言包内部的显式兜底

#### Scenario: 数据库版本查询由 dialect 公共包提供

- **当** 宿主系统信息或 `linapro-monitor-server` 服务监控需要展示数据库版本时
- **则** 调用方通过 `dialect.DatabaseVersion(ctx, db)` 或等价 `Dialect` 方法查询
- **且** PostgreSQL 方言使用 `SELECT version()` 或等价语句，返回包含 `PostgreSQL` 引擎名与版本号的字符串
- **且** 返回给页面的版本文本必须包含数据库引擎名称与非空版本号

#### Scenario: 表元数据查询由 dialect 公共包提供

- **当** 插件 data service 等组件需要查询数据表名与表注释时
- **则** 调用方通过 `Dialect.QueryTableMetadata(ctx, db, schema, names)` 查询
- **且** 调用方不得直接编写 `information_schema.TABLES` / `pg_catalog.pg_description` 等方言特有 SQL
- **且** PostgreSQL 实现使用 `information_schema.tables` JOIN `pg_class` 的 `obj_description(oid)` 查询表注释

#### Scenario: 驱动与 ORM 只读 SQL 分类由 dialect 公共包提供

- **当** `plugindb/host` 等治理层需要允许驱动或 ORM 发出的表元数据读、schema probe、版本 probe 等无业务表只读 SQL 时
- **则** 调用方通过 `Dialect.ClassifyReadSQL(sql)` 或等价公共门面获取语义分类
- **且** 治理层不得直接硬编码 `information_schema`、`pg_catalog`、`pg_class`、`current_schema()`、`version()` 等 PostgreSQL 特有 SQL 片段
- **且** PostgreSQL 具体识别逻辑必须维护在 `pkg/dialect/internal/postgres` 内部子包中
- **且** 分类不得允许带有未授权业务表的 SQL 绕过插件数据表级授权校验

#### Scenario: dialect 公共包不暴露宿主 internal 具体类型

- **当** 插件生命周期、初始化命令或工具链代码导入 `apps/lina-core/pkg/dialect` 时
- **则** 公开接口不要求调用方引用 `apps/lina-core/internal/...` 下的具体服务类型
- **且** 启动期配置能力通过 `dialect.RuntimeConfig` 等窄接口适配
- **且** 宿主 `config.Service` 可在内部实现该窄接口后传入 `Dialect.OnStartup`

### Requirement: 方言根据数据库链接前缀自动分发

系统 SHALL 根据 `database.default.link` 配置的协议头自动选择对应的方言实现。`pgsql:` 前缀分发到 PostgreSQL 方言实现。`sqlite:`、`mysql:` 和其他未识别前缀 MUST 被识别为不支持的方言并返回明确错误，不得静默回退。未识别的前缀必须返回包含前缀名与已支持前缀列表的明确错误。调用方不得依赖或断言具体实现类型，只能依赖 `Dialect` 接口行为。

#### Scenario: PostgreSQL 链接被识别为 PostgreSQL 方言

- **当** 配置文件 `database.default.link` 以 `pgsql:` 开头时
- **则** `dialect.From(link)` 返回实现 `Dialect` 接口的 PostgreSQL 方言实例
- **且** `Name()` 返回字符串 `"postgres"`
- **且** `SupportsCluster()` 返回 `true`

#### Scenario: SQLite 链接被显式拒绝

- **当** 配置文件 `database.default.link` 以 `sqlite:` 开头时
- **则** `dialect.From(link)` 返回明确错误，错误消息包含“SQLite 不再支持”或等价说明
- **且** 错误消息列出当前已支持的前缀仅为 `pgsql:`
- **且** 系统不静默回退到任何默认方言

#### Scenario: MySQL 链接被显式拒绝

- **当** 配置文件 `database.default.link` 以 `mysql:` 开头时
- **则** `dialect.From(link)` 返回明确错误，错误消息包含“MySQL 不再支持”或等价说明
- **且** 错误消息列出当前已支持的前缀仅为 `pgsql:`
- **且** 系统不静默回退到任何默认方言

#### Scenario: 未识别的链接前缀

- **当** 配置文件 `database.default.link` 以未识别的前缀开头时
- **则** `dialect.From(link)` 返回包含前缀名与已支持前缀列表的明确错误
- **且** 系统不静默回退到任何默认方言

### Requirement: PostgreSQL 方言 DDL 入口为无操作

`Dialect.TranslateDDL(ctx, sourceName, ddl)` SHALL 接收调用方传入的 `sourceName` 诊断名。`sourceName` MUST 是源 SQL 文件路径、嵌入资产路径或调用方构造的稳定描述，用于在错误消息中定位失败来源。PostgreSQL 方言的 `TranslateDDL` SHALL 直接返回输入字符串，不做任何修改。这保证了 PostgreSQL 作为唯一 SQL 源方言时生产路径无翻译开销。

#### Scenario: PostgreSQL 方言转译保持原文

- **当** PostgreSQL 方言实例的 `TranslateDDL(ctx, sourceName, ddl)` 被调用时
- **则** 返回值与输入 `ddl` 字节级别完全一致
- **且** 不返回错误
- **且** `sourceName` 不影响 PostgreSQL no-op 转译结果

### Requirement: 方言必须暴露数据库准备入口

`Dialect.PrepareDatabase(ctx, link, rebuild)` SHALL 负责在执行 DDL 资源前完成 PostgreSQL 数据库准备工作。PostgreSQL 方言通过连接系统库 `postgres` 执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`。PostgreSQL 目标数据库创建后，后续宿主 init SQL SHALL 直接创建业务表、索引、注释并写入 seed 数据，不创建自定义排序规则。

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

#### Scenario: PostgreSQL 系统库连接失败

- **当** PostgreSQL `PrepareDatabase` 无法连接到系统库 `postgres`（PG 服务未启动、网络不通、认证失败）时
- **则** 系统返回包含目标 PG 主机、端口与具体错误的明确错误
- **且** 不继续后续 DDL 执行

### Requirement: 方言必须提供启动期钩子

`Dialect.OnStartup(ctx, runtime)` SHALL 在宿主启动 bootstrap 阶段被调用一次。`runtime` SHALL 是 `pkg/dialect` 中定义的稳定窄接口。PostgreSQL 方言为 no-op。该钩子的调用时机必须早于任何 cluster 相关初始化。

#### Scenario: PostgreSQL 启动期钩子无副作用

- **当** PostgreSQL 方言实例的 `OnStartup(ctx, runtime)` 被调用时
- **则** 钩子立即返回 `nil`
- **且** 不修改任何配置项
- **且** 不输出任何警告级别日志

#### Scenario: 启动期钩子在集群初始化前执行

- **当** 宿主启动时
- **则** `OnStartup` 在 `cluster.Service` 启动选举循环前被调用
- **且** 后续 cluster 相关组件读取到的 `IsClusterEnabled` 来自配置服务当前有效值

### Requirement: Dialect 接口必须提供表元数据查询能力

系统 SHALL 在 `Dialect` 接口中提供 `QueryTableMetadata(ctx context.Context, db gdb.DB, schema string, tableNames []string) ([]TableMeta, error)` 方法，用于查询 PostgreSQL 数据表名与表注释。`TableMeta` SHALL 至少包含 `TableName string` 与 `TableComment string` 字段。调用方（如 plugin data service）SHALL 通过该方法查询表元数据，不得直接编写 PostgreSQL 特有的 SQL（`information_schema.TABLES` / `pg_catalog.pg_description`）。

#### Scenario: PostgreSQL 实现使用 information_schema 与 pg_description

- **当** 调用方在 PostgreSQL 方言下调用 `QueryTableMetadata(ctx, db, "public", []string{"sys_user", "sys_role"})`
- **则** 实现 SQL 联接 `information_schema.tables` 与 `pg_class`
- **且** 通过 `obj_description(c.oid)` 获取表注释
- **且** 返回的 `TableMeta` 数组包含传入表名对应的 `TableName` 与 `TableComment`（无注释时为空字符串）

#### Scenario: 调用方不直接编写方言 SQL

- **当** 检查 `apps/lina-core/internal/service/plugin/plugin_data_table_comment.go` 等调用代码
- **则** 代码 MUST NOT 出现 `information_schema.TABLES` / `pg_catalog.pg_description` 字面量
- **且** 代码 MUST 通过 `dialect.FromDatabase(g.DB()).QueryTableMetadata(...)` 查询

#### Scenario: 表名传入空数组

- **当** 调用方传入空的 `tableNames` 数组
- **则** 方法返回空的 `[]TableMeta`
- **且** 不返回错误

#### Scenario: 不存在的表名被静默跳过

- **当** 调用方传入的 `tableNames` 中部分表名在数据库中不存在
- **则** 返回的 `[]TableMeta` 仅包含实际存在的表
- **且** 不为不存在的表名返回任何记录或错误
