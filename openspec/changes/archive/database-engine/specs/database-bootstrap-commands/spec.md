## MODIFIED Requirements

### Requirement: SQL引导命令不得依赖驱动多语句执行

系统SHALL将`init`和`mock`使用的每个SQL文件解析为独立语句的有序列表并逐个执行，而非依赖数据库连接字符串中的驱动级多语句支持。该规则适用于`PostgreSQL`方言：SQL文本由现有`splitSQLStatements`切分，再逐句通过GoFrame`gdb`提交。

#### Scenario: 多语句文件按顺序逐语句运行

- **当**`init`或`mock`读取包含多个SQL语句的目标文件时
- **则**系统按文件中出现的顺序逐个执行这些语句
- **且**空白片段和纯注释片段不被视为可执行语句

#### Scenario: 语句失败后立即停止执行

- **当**`init`或`mock`在执行SQL文件中间语句时收到数据库错误时
- **则**系统立即停止该文件中的剩余语句和所有后续SQL文件
- **且**命令返回失败状态
- **且**错误消息仍包含失败文件名以便快速定位问题

#### Scenario: PostgreSQL模式下源SQL直接切分

- **当**当前方言为`PostgreSQL`且加载某SQL文件时
- **则**系统对原始SQL内容按`splitSQLStatements`切分
- **且**每条语句逐个通过GoFrame`gdb`提交
- **且**不依赖驱动多语句支持

### Requirement: 数据库引导命令必须按方言分发数据库准备逻辑

系统SHALL在`init`命令执行SQL资源前，根据`database.default.link`协议头分发到对应方言的`PrepareDatabase`。当前唯一支持的方言为`PostgreSQL`，`PostgreSQL`方言通过连接系统库`postgres`执行`pg_terminate_backend`、`DROP DATABASE IF EXISTS`和`CREATE DATABASE`。`mock`命令SHALL依赖已由`init`初始化完成的目标数据库，不得创建、重建或准备数据库。引导命令实现不得直接编写`PostgreSQL`专属的链接解析、`pg_terminate_backend`或`DROP/CREATE DATABASE`逻辑。

`init`是运维初始化命令，SHALL直接使用当前配置中的数据库账号执行准备逻辑与后续SQL加载。该账号MUST具备连接系统库、创建数据库、删除数据库、终止目标库连接、连接目标库、创建表、创建索引、写入注释和写入seed数据的足够权限。如果权限不足，命令SHALL快速失败并返回明确错误；运行时服务不得因此引入额外的权限探测、账号切换、跳过建库或自动降级逻辑。

#### Scenario: PostgreSQL链接下init走PostgreSQL方言准备

- **当**配置文件`database.default.link`以`pgsql:`开头且运维人员运行`make init confirm=init`时
- **则**命令调用当前`PostgreSQL`方言实例的`PrepareDatabase`创建或确认数据库存在
- **且**实现通过连接系统库`postgres`完成数据库准备
- **且**后续SQL执行连接到该数据库

#### Scenario: rebuild参数下PostgreSQL方言重建数据库

- **当**配置文件链接以`pgsql:`开头且运维人员运行`make init confirm=init rebuild=true`时
- **则**命令调用当前`PostgreSQL`方言实例的`PrepareDatabase(rebuild=true)`
- **且**实现先终止目标库的活跃连接
- **且**再执行`DROP DATABASE IF EXISTS <目标库>`与`CREATE DATABASE <目标库> ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`
- **且**启动日志输出明确的rebuild警告

#### Scenario: 不支持的数据库链接下init快速失败

- **当**配置文件链接以`sqlite:`、`mysql:`或未知前缀开头且运维人员运行`make init confirm=init`
- **则**命令在方言解析阶段失败
- **且**错误消息说明当前数据库方言不再支持或无法识别
- **且**命令不得创建数据库文件、执行方言转译或读取业务SQL文件

#### Scenario: mock不执行数据库准备

- **当**运维人员运行`make mock confirm=mock`时
- **则**命令不调用`Dialect.PrepareDatabase`
- **且**命令直接使用当前配置中的`database.default.link`连接已初始化数据库并加载mock SQL
- **且**如果目标数据库、数据表或基础seed不存在，命令快速失败并返回数据库错误，不静默创建或重建数据库

### Requirement: 数据库引导命令必须在执行SQL前调用方言入口

系统SHALL在`init`与`mock`执行每个SQL文件前，先调用当前方言的`TranslateDDL(ctx, sourceName, ddl)`。`sourceName`SHALL使用源SQL文件路径或嵌入资产路径。当前唯一支持的`PostgreSQL`方言下转译为no-op，SQL文件保持`PostgreSQL 14+`方言来源。

#### Scenario: PostgreSQL模式下转译保持原SQL字节一致

- **当**当前方言为`PostgreSQL`且`init`加载某SQL文件时
- **则**转译后的内容与原文件字节级别一致
- **且**后续语句分割与执行流程不受影响

#### Scenario: 不支持方言在SQL执行前失败

- **当**当前配置的数据库链接为`sqlite:`、`mysql:`或未知前缀时
- **则**命令在获取方言实例时失败
- **且**不读取或执行任何SQL文件

#### Scenario: 转译失败时命令快速失败

- **当**当前方言转译某SQL文件返回错误时
- **则**命令立即停止后续SQL执行
- **且**错误日志包含失败的`sourceName`
- **且**命令向调用方返回失败状态
