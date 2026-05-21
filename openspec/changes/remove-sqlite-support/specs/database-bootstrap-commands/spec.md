## MODIFIED Requirements

### Requirement: SQL 引导命令不得依赖驱动多语句执行

系统 SHALL 将 `init` 和 `mock` 使用的每个 SQL 文件解析为独立语句的有序列表并逐个执行，而非依赖数据库连接字符串中的驱动级多语句支持。该规则适用于 PostgreSQL 方言：SQL 文本由现有 `splitSQLStatements` 切分，再逐句通过 GoFrame `gdb` 提交。

#### Scenario: 多语句文件按顺序逐语句运行

- **当** `init` 或 `mock` 读取包含多个 SQL 语句的目标文件时
- **则** 系统按文件中出现的顺序逐个执行这些语句
- **且** 空白片段和纯注释片段不被视为可执行语句

#### Scenario: 语句失败后立即停止执行

- **当** `init` 或 `mock` 在执行 SQL 文件中间语句时收到数据库错误时
- **则** 系统立即停止该文件中的剩余语句和所有后续 SQL 文件
- **且** 命令返回失败状态
- **且** 错误消息仍包含失败文件名以便快速定位问题

#### Scenario: PostgreSQL 模式下源 SQL 直接切分

- **当** 当前方言为 PostgreSQL 且加载某 SQL 文件时
- **则** 系统对原始 SQL 内容按 `splitSQLStatements` 切分
- **且** 每条语句逐个通过 GoFrame `gdb` 提交
- **且** 不依赖驱动多语句支持

### Requirement: 数据库引导命令必须按方言分发数据库准备逻辑

系统 SHALL 在 `init` 命令执行 SQL 资源前，根据 `database.default.link` 协议头分发到对应方言的 `PrepareDatabase`。当前唯一支持的方言为 PostgreSQL，PostgreSQL 方言通过连接系统库 `postgres` 执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`。`mock` 命令 SHALL 依赖已由 `init` 初始化完成的目标数据库，不得创建、重建或准备数据库。引导命令实现不得直接编写 PostgreSQL 专属的链接解析、`pg_terminate_backend` 或 `DROP/CREATE DATABASE` 逻辑。

`init` 是运维初始化命令，SHALL 直接使用当前配置中的数据库账号执行准备逻辑与后续 SQL 加载。该账号 MUST 具备连接系统库、创建数据库、删除数据库、终止目标库连接、连接目标库、创建表、创建索引、写入注释和写入 seed 数据的足够权限。如果权限不足，命令 SHALL 快速失败并返回明确错误；运行时服务不得因此引入额外的权限探测、账号切换、跳过建库或自动降级逻辑。

#### Scenario: PostgreSQL 链接下 init 走 PostgreSQL 方言准备

- **当** 配置文件 `database.default.link` 以 `pgsql:` 开头且运维人员运行 `make init confirm=init` 时
- **则** 命令调用当前 PostgreSQL 方言实例的 `PrepareDatabase` 创建或确认数据库存在
- **且** PrepareDatabase 通过连接系统库 `postgres` 执行 `CREATE DATABASE IF NOT EXISTS` 等价逻辑
- **且** 后续 SQL 执行连接到该数据库
- **且** 后续宿主 init SQL 直接创建业务表，不创建自定义排序规则

#### Scenario: SQLite 链接下 init 快速失败

- **当** 配置文件链接以 `sqlite:` 开头且运维人员运行 `make init confirm=init`
- **则** 命令在方言解析阶段失败
- **且** 错误消息说明 SQLite 不再支持
- **且** 命令不得创建 SQLite 父目录、SQLite 数据库文件或执行 SQL 转译

#### Scenario: rebuild 参数下 PostgreSQL 方言重建数据库

- **当** 配置文件链接以 `pgsql:` 开头且运维人员运行 `make init confirm=init rebuild=true` 时
- **则** 命令调用当前 PostgreSQL 方言实例的 `PrepareDatabase(rebuild=true)`
- **且** 实现先连接系统库 `postgres`，调用 `pg_terminate_backend` 终止活跃连接
- **且** 再执行 `DROP DATABASE IF EXISTS <目标库>`
- **且** 再执行 `CREATE DATABASE <目标库> ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`
- **且** 启动日志输出明确的 rebuild 警告
- **且** 后续宿主 init SQL 在重建后的目标库内直接创建业务表，不创建自定义排序规则

#### Scenario: PostgreSQL 系统库无法连接时 init 快速失败

- **当** PostgreSQL 方言 `PrepareDatabase` 无法连接到系统库 `postgres`
- **则** `init` 命令立即返回失败
- **且** 错误消息包含目标主机、端口、用户名与具体错误
- **且** 错误消息提示运维人员“PG 未就绪，请先启动 PostgreSQL 服务”或等价友好提示，并可附带本地 `docker run` 或仓库已有 compose 文件的启动示例

#### Scenario: PostgreSQL 初始化账号权限不足时 init 快速失败

- **当** PostgreSQL 方言 `PrepareDatabase` 或后续 init SQL 使用配置中的数据库账号执行建库、删库、终止连接、创建排序规则、建表、建索引、写注释或写入 seed 数据时收到权限错误
- **则** `init` 命令立即返回失败
- **且** 错误消息包含当前用户名、目标数据库、失败操作和 PostgreSQL 返回的具体权限错误
- **且** 系统不静默跳过数据库准备，不切换到其他隐式账号，也不提供低权限初始化路径

#### Scenario: mock 不执行数据库准备

- **当** 运维人员运行 `make mock confirm=mock` 时
- **则** 命令不调用 `Dialect.PrepareDatabase`
- **且** 命令直接使用当前配置中的 `database.default.link` 连接已初始化数据库并加载 mock SQL
- **且** 如果目标数据库、数据表或基础 seed 不存在，命令快速失败并返回数据库错误，不静默创建或重建数据库

### Requirement: 数据库引导命令必须在执行 SQL 前调用方言入口

系统 SHALL 在 `init` / `mock` 执行每个 SQL 文件前，先调用当前方言的 `TranslateDDL(ctx, sourceName, ddl)`。`sourceName` SHALL 使用源 SQL 文件路径或嵌入资产路径。当前唯一支持的 PostgreSQL 方言下转译为 no-op，SQL 文件的源文件保持 PostgreSQL 14+ 方言来源。

#### Scenario: PostgreSQL 模式下转译保持原 SQL 字节一致

- **当** 当前方言为 PostgreSQL 且 `init` 加载某 SQL 文件时
- **则** 转译后的内容与原文件字节级别一致
- **且** 后续语句分割与执行流程不受影响

#### Scenario: 不支持方言在 SQL 执行前失败

- **当** 当前配置的数据库链接为 `sqlite:`、`mysql:` 或未知前缀时
- **则** 命令在获取方言实例时失败
- **且** 不读取或执行任何 SQL 文件

#### Scenario: 转译失败时命令快速失败

- **当** 当前方言转译某 SQL 文件返回错误时
- **则** 命令立即停止后续 SQL 执行
- **且** 错误日志包含失败的 `sourceName`
- **且** 命令向调用方返回失败状态
