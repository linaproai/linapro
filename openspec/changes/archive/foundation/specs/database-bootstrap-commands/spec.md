## MODIFIED Requirements

### Requirement: SQL 引导命令不得依赖驱动多语句执行

系统 SHALL 将 `init` 和 `mock` 使用的每个 SQL 文件解析为独立语句的有序列表并逐个执行，而非依赖数据库连接字符串中的驱动级多语句支持。该规则同时适用于 PostgreSQL 与 SQLite 方言。

#### Scenario: 多语句文件按顺序逐语句运行
- **当** `init` 或 `mock` 读取包含多个 SQL 语句的目标文件时
- **则** 系统按文件中出现的顺序逐个执行这些语句
- **且** 空白片段和纯注释片段不被视为可执行语句

#### Scenario: 语句失败后立即停止执行
- **当** `init` 或 `mock` 在执行 SQL 文件中间语句时收到数据库错误时
- **则** 系统立即停止该文件中的剩余语句和所有后续 SQL 文件
- **且** 命令返回失败状态
- **且** 错误消息仍包含失败文件名以便快速定位问题

### Requirement: 数据库引导命令必须按方言分发数据库准备逻辑

系统 SHALL 在 `init` 命令执行 SQL 资源前，根据 `database.default.link` 协议头分发到对应方言的 `PrepareDatabase`。PostgreSQL 方言通过连接系统库 `postgres` 执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`；SQLite 方言执行父目录创建 / 可选数据库文件删除。`mock` 命令 SHALL 依赖已由 `init` 初始化完成的目标数据库，不得创建、重建或准备数据库。

`init` 是运维初始化命令，SHALL 直接使用当前配置中的数据库账号执行。该账号 MUST 具备连接系统库、创建数据库、删除数据库、终止目标库连接、创建表、索引、注释并写入 seed 数据的足够权限。权限不足时命令 SHALL 快速失败并返回明确错误。

#### Scenario: PostgreSQL 链接下 init 走 PostgreSQL 方言准备
- **当** 配置文件以 `pgsql:` 开头且运行 `make init confirm=init` 时
- **则** 命令调用 PostgreSQL 方言的 `PrepareDatabase`
- **且** 通过连接系统库 `postgres` 执行创建数据库逻辑
- **且** 后续宿主 init SQL 直接创建业务表，不创建自定义排序规则

#### Scenario: SQLite 链接下 init 走 SQLite 方言准备
- **当** 配置文件以 `sqlite:` 开头且运行 `make init confirm=init` 时
- **则** 命令调用 SQLite 方言的 `PrepareDatabase`，自动创建数据库文件父目录

#### Scenario: rebuild 参数下 PostgreSQL 方言重建数据库
- **当** 配置以 `pgsql:` 开头且运行 `make init confirm=init rebuild=true` 时
- **则** 先连接系统库 `postgres`，调用 `pg_terminate_backend` 终止活跃连接
- **且** 再执行 `DROP DATABASE IF EXISTS` + `CREATE DATABASE ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`

#### Scenario: PostgreSQL 系统库无法连接时 init 快速失败
- **当** PostgreSQL 方言 `PrepareDatabase` 无法连接到系统库时
- **则** `init` 命令立即返回失败
- **且** 错误消息提示"PG 未就绪，请先启动 PostgreSQL 服务"

#### Scenario: mock 不执行数据库准备
- **当** 运维人员运行 `make mock confirm=mock` 时
- **则** 命令不调用 `Dialect.PrepareDatabase`
- **且** 直接连接已初始化数据库并加载 mock SQL

### Requirement: 数据库引导命令必须在执行 SQL 前调用方言转译

系统 SHALL 在 `init` / `mock` 执行每个 SQL 文件前，先调用当前方言的 `TranslateDDL(ctx, sourceName, ddl)` 将 PostgreSQL 14+ 方言来源的 SQL 内容转换为目标方言可执行的内容。PostgreSQL 方言下转译为 no-op；SQLite 方言下转译产出 SQLite 兼容语句。

#### Scenario: PostgreSQL 模式下转译保持原 SQL 字节一致
- **当** 当前方言为 PostgreSQL 且 `init` 加载某 SQL 文件时
- **则** 转译后的内容与原文件字节级别一致

#### Scenario: SQLite 模式下转译产出 SQLite 兼容语句
- **当** 当前方言为 SQLite 且 `init` 加载 SQL 文件时
- **则** 命令先用 SQLite 方言的 `TranslateDDL` 转译文件内容
- **且** 每条转译后的语句在 SQLite 上成功执行

#### Scenario: 转译失败时命令快速失败
- **当** 当前方言转译某 SQL 文件返回错误时
- **则** 命令立即停止后续 SQL 执行
- **且** 错误日志包含失败的 `sourceName`、行号提示与未覆盖关键字
