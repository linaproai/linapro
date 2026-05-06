## ADDED Requirements

### Requirement: 数据库引导命令必须按方言分发数据库准备逻辑

系统 SHALL 在 `init` / `mock` 命令执行 SQL 资源前，根据 `database.default.link` 协议头分发到对应方言的 `PrepareDatabase`。MySQL 方言执行 `CREATE DATABASE` / 可选 `DROP DATABASE`；SQLite 方言执行父目录创建 / 可选数据库文件删除。引导命令实现不得直接编写 MySQL 专属的链接解析或 `DROP/CREATE DATABASE` 逻辑。

#### Scenario: MySQL 链接下 init 走 MySQL 方言准备
- **当** 配置文件 `database.default.link` 以 `mysql:` 开头且运维人员运行 `make init confirm=init` 时
- **则** 命令调用 `MySQLDialect.PrepareDatabase` 创建或确认数据库存在
- **且** 后续 SQL 执行连接到该数据库

#### Scenario: SQLite 链接下 init 走 SQLite 方言准备
- **当** 配置文件 `database.default.link` 以 `sqlite:` 开头且运维人员运行 `make init confirm=init` 时
- **则** 命令调用 `SQLiteDialect.PrepareDatabase`，自动创建数据库文件父目录
- **且** 后续 SQL 执行连接到该 SQLite 文件

#### Scenario: rebuild 参数下 SQLite 方言删除数据库文件
- **当** 配置文件链接以 `sqlite:` 开头且运维人员运行 `make init confirm=init rebuild=true` 时
- **则** 命令调用 `SQLiteDialect.PrepareDatabase(rebuild=true)` 删除现有数据库文件
- **且** 删除范围包括主 `.db` 文件以及可能存在的 WAL / SHM 等附属文件
- **且** 父目录被保留（不删除目录本身）

### Requirement: 数据库引导命令必须在执行 SQL 前调用方言转译

系统 SHALL 在 `init` / `mock` 执行每个 SQL 文件前，先调用当前方言的 `TranslateDDL` 将单一 MySQL 方言来源的 SQL 内容转换为目标方言可执行的内容。MySQL 方言下转译为 no-op；SQLite 方言下转译产出 SQLite 兼容语句。SQL 文件的源文件保持单一 MySQL 方言来源。

#### Scenario: MySQL 模式下转译保持原 SQL 字节一致
- **当** 当前方言为 MySQL 且 `init` 加载某 SQL 文件时
- **则** 转译后的内容与原文件字节级别一致
- **且** 后续语句分割与执行流程不受影响

#### Scenario: SQLite 模式下转译产出 SQLite 兼容语句
- **当** 当前方言为 SQLite 且 `init` 加载 `001-project-init.sql` 时
- **则** 命令先用 `SQLiteDialect.TranslateDDL` 转译文件内容
- **且** 再调用现有 `splitSQLStatements` 分割
- **且** 每条转译后的语句在 SQLite 上成功执行

#### Scenario: 转译失败时命令快速失败
- **当** 当前方言转译某 SQL 文件返回错误时
- **则** 命令立即停止后续 SQL 执行
- **且** 错误日志包含失败的源文件名
- **且** 命令向调用方返回失败状态

## MODIFIED Requirements

### Requirement: SQL 引导命令不得依赖驱动多语句执行

系统 SHALL 将 `init` 和 `mock` 使用的每个 SQL 文件解析为独立语句的有序列表并逐个执行，而非依赖数据库连接字符串中的驱动级多语句支持。该规则同时适用于 MySQL 与 SQLite 方言：方言转译产出后的 SQL 文本仍由现有 `splitSQLStatements` 切分，再逐句通过 GoFrame `gdb` 提交。

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
