## ADDED Requirements

### Requirement: 宿主必须通过统一的方言抽象层收敛数据库引擎差异

系统 SHALL 在 `apps/lina-core/pkg/dialect/` 提供 `Dialect` 接口作为数据库引擎差异的唯一收敛点。所有数据库引擎相关的差异化行为（DDL 转译、数据库准备、集群能力查询、启动期钩子）必须通过该接口暴露，业务模块（`controller` / `service` / `model` / `dao`）不得在自身代码路径中出现 `if isMySQL / if isSQLite` 等数据库引擎判断。

#### Scenario: 业务模块不感知数据库引擎差异
- **当** 业务模块（如 `user` / `role` / `dict` / `kvcache` / `locker`）通过 DAO 层执行查询、写入、更新、删除操作时
- **则** 业务代码不包含针对数据库引擎的分支判断
- **且** 同一份业务代码在 MySQL 和 SQLite 两种引擎下行为一致

#### Scenario: 所有方言相关行为通过 Dialect 接口暴露
- **当** 宿主需要执行"DDL 转译 / 数据库准备 / 集群能力查询 / 启动期钩子"中的任一行为时
- **则** 调用方通过 `dialect.From(link)` 获取当前方言实例
- **且** 调用方仅依赖 `Dialect` 接口的方法签名，不依赖具体实现的内部细节

### Requirement: 方言根据数据库链接前缀自动分发

系统 SHALL 根据 `database.default.link` 配置的协议头自动选择对应的方言实现。`mysql:` 前缀分发到 `MySQLDialect`，`sqlite:` 前缀分发到 `SQLiteDialect`。未识别的前缀必须返回明确的错误。

#### Scenario: MySQL 链接被识别为 MySQL 方言
- **当** 配置文件 `database.default.link` 以 `mysql:` 开头时
- **则** `dialect.From(link)` 返回 `MySQLDialect` 实例
- **且** `Name()` 返回字符串 `"mysql"`
- **且** `SupportsCluster()` 返回 `true`

#### Scenario: SQLite 链接被识别为 SQLite 方言
- **当** 配置文件 `database.default.link` 以 `sqlite:` 开头时
- **则** `dialect.From(link)` 返回 `SQLiteDialect` 实例
- **且** `Name()` 返回字符串 `"sqlite"`
- **且** `SupportsCluster()` 返回 `false`

#### Scenario: 未识别的链接前缀
- **当** 配置文件 `database.default.link` 以未识别的前缀开头时
- **则** `dialect.From(link)` 返回包含前缀名与已支持前缀列表的明确错误
- **且** 系统不静默回退到任何默认方言

### Requirement: MySQL 方言 DDL 转译为无操作

`MySQLDialect.TranslateDDL` SHALL 直接返回输入字符串，不做任何修改。这保证了 MySQL 用户在引入方言抽象层后行为完全向后兼容，不会因转译副作用引入新的失败路径。

#### Scenario: MySQL 方言转译保持原文
- **当** `MySQLDialect.TranslateDDL(ctx, ddl)` 被调用时
- **则** 返回值与输入 `ddl` 字节级别完全一致
- **且** 不返回错误（除非输入本身为 `nil` 或空字符串等显式无效输入）

### Requirement: SQLite 方言 DDL 转译必须覆盖项目 DDL 子集

`SQLiteDialect.TranslateDDL` SHALL 将单一 MySQL 方言来源的 DDL 转译为可在 SQLite 上成功执行的语句。转译必须覆盖项目当前 SQL 文件中实际使用的所有 MySQL 语法，包括反引号标识符、`AUTO_INCREMENT`、`UNSIGNED`、`TINYINT` / `SMALLINT` / `LONGTEXT` 类型、`ENGINE=` / `CHARSET=` / `COLLATE=` 子句、列级与表级 `COMMENT '...'`、`INSERT IGNORE`、`DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`、表内联 `KEY` / `INDEX` / `UNIQUE KEY`、`CREATE DATABASE` / `USE` 整句。

#### Scenario: MEMORY 引擎子句被去除
- **当** 输入 DDL 包含 `... ) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='Distributed lock table';` 时
- **则** 转译结果不包含 `ENGINE=` / `CHARSET=` / `COMMENT=` 任一子句
- **且** 表创建语义保留：表本身被创建，但作为 SQLite 普通表（持久化）

#### Scenario: AUTO_INCREMENT 主键被改写
- **当** 输入 DDL 包含 `id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT` 时
- **则** 转译结果中该列定义为 `id INTEGER PRIMARY KEY AUTOINCREMENT`
- **且** 不保留 `BIGINT` / `UNSIGNED` 等 SQLite 不支持的修饰符

#### Scenario: 列类型被映射为 SQLite 等价类型
- **当** 输入 DDL 包含 `VARCHAR(64)` / `LONGTEXT` / `TINYINT` / `DECIMAL(10,2)` 任一列定义时
- **则** 转译结果对应列分别映射为 `TEXT` / `TEXT` / `INTEGER` / `NUMERIC`

#### Scenario: INSERT IGNORE 被改写
- **当** 输入 DDL 包含 `INSERT IGNORE INTO sys_user (...)` 语句时
- **则** 转译结果改写为 `INSERT OR IGNORE INTO sys_user (...)`
- **且** 写入语义（重复键时跳过）保持等价

#### Scenario: 列级与表级 COMMENT 被去除
- **当** 输入 DDL 包含 `id INT COMMENT 'User ID'` 列级注释或 `... ) ... COMMENT='User table';` 表级注释时
- **则** 转译结果不包含任何 `COMMENT '...'` 或 `COMMENT='...'` 子句
- **且** 列定义与表定义其余部分保持完整

#### Scenario: ON UPDATE CURRENT_TIMESTAMP 被去除而 DEFAULT CURRENT_TIMESTAMP 保留
- **当** 输入 DDL 包含 `created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` 时
- **则** 转译结果保留 `DEFAULT CURRENT_TIMESTAMP` 子句
- **且** 移除 `ON UPDATE CURRENT_TIMESTAMP` 子句
- **且** `updated_at` 列的实时更新由 GoFrame DAO 层在写入时自动维护

#### Scenario: 表内联索引被提取为独立 CREATE INDEX 语句
- **当** 输入 DDL 在 CREATE TABLE 内包含 `KEY idx_status (status), INDEX idx_phone (phone), UNIQUE KEY uk_name (name)` 子句时
- **则** 转译结果中 CREATE TABLE 仅保留列定义与 PRIMARY KEY、UNIQUE 约束
- **且** 内联索引被提取为表创建语句之后的 `CREATE INDEX idx_status ON tbl(status);` 等独立语句
- **且** UNIQUE KEY 转译为 `CREATE UNIQUE INDEX uk_name ON tbl(name);`

#### Scenario: CREATE DATABASE 与 USE 整句被丢弃
- **当** 输入 DDL 包含 `CREATE DATABASE IF NOT EXISTS \`linapro\` ...;` 或 `USE \`linapro\`;` 整句时
- **则** 转译结果不包含这些语句
- **且** 转译器不报错（SQLite 没有"数据库"概念，丢弃即正确语义）

#### Scenario: 反引号标识符被去除或正常化
- **当** 输入 DDL 包含 `` `id` `` / `` `sys_user` `` 等反引号包裹的标识符时
- **则** 转译结果中标识符不带反引号（直接裸写）或使用双引号
- **且** 转译结果在 SQLite 上可成功执行

### Requirement: DDL 转译失败时必须返回明确错误

`SQLiteDialect.TranslateDDL` 在遇到当前实现未覆盖的 MySQL 语法时 SHALL 返回包含输入文件名（若由调用方传入）、行号定位提示与未覆盖语法关键字的明确错误，不得静默丢弃或产生无效 SQL。

#### Scenario: 转译器遇到未覆盖的语法
- **当** 输入 DDL 包含未在覆盖范围内的 MySQL 特性（如 `FULLTEXT INDEX` / `GENERATED ALWAYS AS` / 分区子句等）时
- **则** 转译器返回错误，错误消息包含未覆盖的关键字
- **且** 调用方（`cmd init` / `cmd mock` / 插件 install pipeline）将错误向上传播
- **且** 系统不执行任何已部分转译的 SQL 内容

### Requirement: 方言必须暴露数据库准备入口

`Dialect.PrepareDatabase(ctx, link, rebuild)` SHALL 负责在执行 DDL 资源前完成方言相关的数据库准备工作。MySQL 方言执行 `CREATE DATABASE IF NOT EXISTS` 与可选的 `DROP DATABASE`；SQLite 方言执行父目录创建（`mkdir -p`）与可选的数据库文件删除。

#### Scenario: MySQL 方言准备数据库
- **当** `MySQLDialect.PrepareDatabase(ctx, link, rebuild=false)` 被调用时
- **则** 系统执行 `CREATE DATABASE IF NOT EXISTS linapro` 等价语句
- **且** 不删除已存在的数据库

#### Scenario: MySQL 方言重建数据库
- **当** `MySQLDialect.PrepareDatabase(ctx, link, rebuild=true)` 被调用时
- **则** 系统先执行 `DROP DATABASE IF EXISTS linapro`
- **且** 再执行 `CREATE DATABASE linapro`
- **且** 启动日志输出明确的 rebuild 警告

#### Scenario: SQLite 方言准备数据库文件
- **当** `SQLiteDialect.PrepareDatabase(ctx, link, rebuild=false)` 被调用且数据库文件父目录不存在时
- **则** 系统自动 `mkdir -p` 父目录
- **且** 数据库文件由后续 DDL 执行自动创建（GoFrame 驱动行为）
- **且** 已存在的数据库文件不被删除

#### Scenario: SQLite 方言重建数据库
- **当** `SQLiteDialect.PrepareDatabase(ctx, link, rebuild=true)` 被调用时
- **则** 系统先删除数据库文件（含 WAL / SHM 等附属文件）
- **且** 再确保父目录存在
- **且** 启动日志输出明确的 rebuild 警告

#### Scenario: SQLite 父目录不可创建
- **当** SQLite 数据库文件父目录创建失败（权限不足、磁盘满等）时
- **则** 系统返回包含目标路径的明确错误
- **且** 不继续后续 DDL 执行

### Requirement: 方言必须提供启动期钩子

`Dialect.OnStartup(ctx, configSvc)` SHALL 在宿主启动 bootstrap 阶段被调用一次。MySQL 方言为 no-op；SQLite 方言负责执行"强制覆盖 cluster.enabled=false + 输出警告日志"等启动期专属行为。该钩子的调用时机必须早于任何 cluster 相关初始化。

#### Scenario: MySQL 启动期钩子无副作用
- **当** `MySQLDialect.OnStartup(ctx, configSvc)` 被调用时
- **则** 钩子立即返回 `nil`
- **且** 不修改任何配置项
- **且** 不输出任何警告级别日志

#### Scenario: SQLite 启动期钩子锁定集群配置
- **当** `SQLiteDialect.OnStartup(ctx, configSvc)` 被调用时
- **则** `configSvc.IsClusterEnabled(ctx)` 在该钩子调用后稳定返回 `false`
- **且** 该覆盖优先级高于 `config.yaml` 中的 `cluster.enabled` 显式声明
- **且** 钩子向终端输出至少 4 行 `[WARNING]` 级别日志，明确告知 SQLite 模式、cluster 锁定原因、单机部署限制、不得用于生产

#### Scenario: 启动期钩子在集群初始化前执行
- **当** 宿主以 SQLite 模式启动时
- **则** `OnStartup` 在 `cluster.Service` 启动选举循环前被调用
- **且** 后续 cluster 相关组件读取到的 `IsClusterEnabled` 已为 `false`
- **且** 不会出现"先启动选举循环再被关闭"的中间状态
