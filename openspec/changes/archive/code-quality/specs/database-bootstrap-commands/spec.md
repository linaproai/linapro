## MODIFIED Requirements

### Requirement: SQL 引导命令不得依赖驱动多语句执行

系统 SHALL 将 `init` 和 `mock` 使用的每个 SQL 文件解析为独立语句的有序列表并逐个执行。

### Requirement: 数据库引导命令必须按方言分发数据库准备逻辑

系统 SHALL 在 `init` 命令执行 SQL 资源前，根据 `database.default.link` 协议头分发到对应方言的 `PrepareDatabase`。当前唯一支持的方言为 PostgreSQL。

#### Scenario: SQLite 链接下 init 快速失败

- **当** 配置文件链接以 `sqlite:` 开头且运维人员运行 `make init confirm=init`
- **则** 命令在方言解析阶段失败
- **且** 错误消息说明 SQLite 不再支持

### Requirement: 数据库引导命令必须在执行 SQL 前调用方言入口

系统 SHALL 在 `init` / `mock` 执行每个 SQL 文件前，先调用当前方言的 `TranslateDDL`。PostgreSQL 方言下转译为 no-op。
