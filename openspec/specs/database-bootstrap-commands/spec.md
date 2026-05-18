# 数据库引导命令规范

## Purpose
定义宿主数据库引导命令的安全边界、确认机制、SQL 资源来源、方言分发、转译顺序和快速失败规则，确保初始化与演示数据加载在 PostgreSQL 与 SQLite 模式下行为一致。

## Requirements
### Requirement: 敏感的数据库引导命令需要显式确认

系统 SHALL 要求宿主 `init` 和 `mock` 命令在执行任何 SQL 前接收与命令名称匹配的显式确认值。如果确认缺失或不正确，命令必须拒绝运行。`init` 和 `mock` 仅限于引导初始化，不得充当正式的升级命令。

#### Scenario: `init` 命令缺少确认
- **当** 运维人员运行 `go run main.go init` 但未带 `--confirm=init` 时
- **则** 命令拒绝执行初始化 SQL
- **且** 命令打印清晰的失败原因和正确示例

#### Scenario: `mock` 命令接收错误的确认值
- **当** 运维人员运行 `go run main.go mock --confirm=init` 时
- **则** 命令拒绝执行 `mock-data` 下的任何 SQL
- **且** 命令说明确认值必须匹配 `mock`

#### Scenario: 命令接收正确的确认值
- **当** 运维人员运行 `go run main.go init --confirm=init` 或 `go run main.go mock --confirm=mock` 时
- **则** 命令可进入匹配的 SQL 扫描和执行流程

#### Scenario: `init` 不创建框架升级记账
- **当** 运维人员运行 `go run main.go init --confirm=init` 且每个宿主 SQL 文件成功时
- **则** 命令仅执行宿主引导初始化
- **且** 不写入框架升级状态、升级记录或 SQL 游标元数据

### Requirement: `Makefile` 条目必须复用相同的确认语义

系统 SHALL 要求仓库根目录和 `apps/lina-core` 中的 `make init` 和 `make mock` 使用与命令实现相同的确认值，并在确认值缺失或不正确时提前失败。

#### Scenario: 仓库根目录 `make init` 缺少确认
- **当** 运维人员从仓库根目录运行 `make init` 但未带 `confirm=init` 时
- **则** `Makefile` 拒绝继续
- **且** 打印正确示例 `make init confirm=init`

#### Scenario: 后端 `make mock` 使用正确的确认变量
- **当** 运维人员从 `apps/lina-core` 运行 `make mock confirm=mock` 时
- **则** `Makefile` 将确认值传递给后端命令实现
- **且** 后端命令继续进行 `mock` 特定的验证和执行

### Requirement: 数据库引导命令必须按执行阶段显式选择 SQL 资源来源

系统 SHALL 使 SQL 资源来源显式化。运行时 `lina init` 和 `lina mock` 命令默认从嵌入式 FS 读取宿主 SQL 资源，而开发时 `make init` 和 `make mock` 命令必须显式切换到源码树中的本地 SQL 文件。实现不得从当前工作目录推断来源。

#### Scenario: 运行时 `init` 默认读取嵌入式 SQL
- **当** 运维人员从发布的二进制文件运行 `lina init --confirm=init` 时
- **则** 命令从 `manifest/sql/` 读取嵌入式 SQL 资源
- **且** 不要求本地源码树存在

#### Scenario: 开发时 `make mock` 显式读取本地 SQL
- **当** 开发者运行 `make mock confirm=mock` 时
- **则** `Makefile` 显式将命令切换到本地 SQL 源
- **且** 命令从源码树中的 `manifest/sql/mock-data/` 读取 SQL

### Requirement: 数据库引导 SQL 执行必须快速失败

系统 SHALL 在 `init` 或 `mock` 期间任何 SQL 文件失败时立即停止执行，并向调用方返回失败结果。

#### Scenario: SQL 文件在执行期间失败
- **当** 一个 SQL 文件在 `init` 或 `mock` 期间返回执行错误时
- **则** 系统立即停止执行后续 SQL 文件
- **且** 命令向 `make` 或直接调用方返回失败状态
- **且** 日志包含失败文件名和错误详情

#### Scenario: 每个 SQL 文件成功
- **当** 每个目标 SQL 文件在 `init` 或 `mock` 期间成功时
- **则** 命令返回成功状态
- **且** 日志打印对应的完成消息

### Requirement: SQL 引导命令不得依赖驱动多语句执行

系统 SHALL 将 `init` 和 `mock` 使用的每个 SQL 文件解析为独立语句的有序列表并逐个执行，而非依赖数据库连接字符串中的驱动级多语句支持。该规则同时适用于 PostgreSQL 与 SQLite 方言：方言转译产出后的 SQL 文本仍由现有 `splitSQLStatements` 切分，再逐句通过 GoFrame `gdb` 提交。

#### Scenario: 多语句文件按顺序逐语句运行
- **当** `init` 或 `mock` 读取包含多个 SQL 语句的目标文件时
- **则** 系统按文件中出现的顺序逐个执行这些语句
- **且** 空白片段和纯注释片段不被视为可执行语句

#### Scenario: 语句失败后立即停止执行
- **当** `init` 或 `mock` 在执行 SQL 文件中间语句时收到数据库错误时
- **则** 系统立即停止该文件中的剩余语句和所有后续 SQL 文件
- **且** 命令返回失败状态
- **且** 错误消息仍包含失败文件名以便快速定位问题

#### Scenario: SQLite 模式下转译后的多语句正常切分
- **当** 当前方言为 SQLite 且转译后的 SQL 文本包含 CREATE TABLE 语句加多条 CREATE INDEX 语句时
- **则** 系统按转译产出顺序逐句执行
- **且** 任意一条 CREATE INDEX 失败时立即停止后续语句

#### Scenario: PostgreSQL 模式下源 SQL 直接切分
- **当** 当前方言为 PostgreSQL 且加载某 SQL 文件时
- **则** 系统对原始（未转译）SQL 内容按 `splitSQLStatements` 切分
- **且** 每条语句逐个通过 GoFrame `gdb` 提交
- **且** 不依赖驱动多语句支持

### Requirement: 数据库引导命令必须按方言分发数据库准备逻辑

系统 SHALL 在 `init` 命令执行 SQL 资源前，根据 `database.default.link` 协议头分发到对应方言的 `PrepareDatabase`。PostgreSQL 方言通过连接系统库 `postgres` 执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`；SQLite 方言执行父目录创建 / 可选数据库文件删除。`mock` 命令 SHALL 依赖已由 `init` 初始化完成的目标数据库，不得创建、重建或准备数据库。引导命令实现不得直接编写 PostgreSQL 专属的链接解析、`pg_terminate_backend` 或 `DROP/CREATE DATABASE` 逻辑。

`init` 是运维初始化命令，SHALL 直接使用当前配置中的数据库账号执行准备逻辑与后续 SQL 加载。该账号 MUST 具备连接系统库、创建数据库、删除数据库、终止目标库连接、连接目标库、创建表、创建索引、写入注释和写入 seed 数据的足够权限。如果权限不足，命令 SHALL 快速失败并返回明确错误；运行时服务不得因此引入额外的权限探测、账号切换、跳过建库或自动降级逻辑。

#### Scenario: PostgreSQL 链接下 init 走 PostgreSQL 方言准备
- **当** 配置文件 `database.default.link` 以 `pgsql:` 开头且运维人员运行 `make init confirm=init` 时
- **则** 命令调用当前 PostgreSQL 方言实例的 `PrepareDatabase` 创建或确认数据库存在
- **且** PrepareDatabase 通过连接系统库 `postgres` 执行 `CREATE DATABASE IF NOT EXISTS` 等价逻辑
- **且** 后续 SQL 执行连接到该数据库
- **且** 后续宿主 init SQL 直接创建业务表，不创建自定义排序规则

#### Scenario: SQLite 链接下 init 走 SQLite 方言准备
- **当** 配置文件 `database.default.link` 以 `sqlite:` 开头且运维人员运行 `make init confirm=init` 时
- **则** 命令调用当前 SQLite 方言实例的 `PrepareDatabase`，自动创建数据库文件父目录
- **且** 后续 SQL 执行连接到该 SQLite 文件

#### Scenario: rebuild 参数下 PostgreSQL 方言重建数据库
- **当** 配置文件链接以 `pgsql:` 开头且运维人员运行 `make init confirm=init rebuild=true` 时
- **则** 命令调用当前 PostgreSQL 方言实例的 `PrepareDatabase(rebuild=true)`
- **且** 实现先连接系统库 `postgres`，调用 `pg_terminate_backend` 终止活跃连接
- **且** 再执行 `DROP DATABASE IF EXISTS <目标库>`
- **且** 再执行 `CREATE DATABASE <目标库> ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`
- **且** 启动日志输出明确的 rebuild 警告
- **且** 后续宿主 init SQL 在重建后的目标库内直接创建业务表，不创建自定义排序规则

#### Scenario: rebuild 参数下 SQLite 方言删除数据库文件
- **当** 配置文件链接以 `sqlite:` 开头且运维人员运行 `make init confirm=init rebuild=true` 时
- **则** 命令调用当前 SQLite 方言实例的 `PrepareDatabase(rebuild=true)` 删除现有数据库文件
- **且** 删除范围包括主 `.db` 文件以及可能存在的 WAL / SHM 等附属文件
- **且** 父目录被保留（不删除目录本身）

#### Scenario: PostgreSQL 系统库无法连接时 init 快速失败
- **当** PostgreSQL 方言 `PrepareDatabase` 无法连接到系统库 `postgres`
- **则** `init` 命令立即返回失败
- **且** 错误消息包含目标主机、端口、用户名与具体错误
- **且** 错误消息提示运维人员"PG 未就绪，请先启动 PostgreSQL 服务"或等价友好提示，并可附带本地 `docker run` 或仓库已有 compose 文件的启动示例

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

### Requirement: 数据库引导命令必须在执行 SQL 前调用方言转译

系统 SHALL 在 `init` / `mock` 执行每个 SQL 文件前，先调用当前方言的 `TranslateDDL(ctx, sourceName, ddl)` 将单一 PostgreSQL 14+ 方言来源的 SQL 内容转换为目标方言可执行的内容。`sourceName` SHALL 使用源 SQL 文件路径或嵌入资产路径。PostgreSQL 方言下转译为 no-op；SQLite 方言下转译产出 SQLite 兼容语句。SQL 文件的源文件保持单一 PostgreSQL 14+ 方言来源。

#### Scenario: PostgreSQL 模式下转译保持原 SQL 字节一致
- **当** 当前方言为 PostgreSQL 且 `init` 加载某 SQL 文件时
- **则** 转译后的内容与原文件字节级别一致
- **且** 后续语句分割与执行流程不受影响

#### Scenario: SQLite 模式下转译产出 SQLite 兼容语句
- **当** 当前方言为 SQLite 且 `init` 加载 `001-project-init.sql` 时
- **则** 命令先用当前 SQLite 方言实例的 `TranslateDDL(ctx, sourceName, ddl)` 转译文件内容
- **且** 再调用现有 `splitSQLStatements` 分割
- **且** 每条转译后的语句在 SQLite 上成功执行

#### Scenario: 转译失败时命令快速失败
- **当** 当前方言转译某 SQL 文件返回错误时
- **则** 命令立即停止后续 SQL 执行
- **且** 错误日志包含失败的 `sourceName`、行号提示与未覆盖关键字
- **且** 命令向调用方返回失败状态
