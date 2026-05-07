## ADDED Requirements

### Requirement: 宿主必须通过统一的方言抽象层收敛数据库引擎差异

系统 SHALL 在 `apps/lina-core/pkg/dialect/` 提供公共稳定的 `Dialect` 接口与方言辅助能力作为数据库引擎差异的唯一收敛点。所有数据库引擎相关的差异化行为（DDL 转译、数据库准备、集群能力查询、启动期钩子、驱动错误分类）必须通过该包暴露，业务模块（`controller` / `service` / `model` / `dao`）不得在自身代码路径中出现 `if isMySQL / if isSQLite` 等数据库引擎判断。`pkg/dialect` 的公开签名 SHALL 只依赖稳定窄接口，不得暴露宿主 `internal` 包中的具体服务类型，也不得导出 MySQL / SQLite 方言具体实现类型；具体方言实现应收敛在 `pkg/dialect/internal/mysql`、`pkg/dialect/internal/sqlite` 等内部子包中，由公共工厂与公共门面能力统一委托。

#### Scenario: 业务模块不感知数据库引擎差异
- **当** 业务模块（如 `user` / `role` / `dict` / `kvcache` / `locker`）通过 DAO 层执行查询、写入、更新、删除操作时
- **则** 业务代码不包含针对数据库引擎的分支判断
- **且** 同一份业务代码在 MySQL 和 SQLite 两种引擎下行为一致

#### Scenario: 所有方言相关行为通过 Dialect 接口暴露
- **当** 宿主需要执行"DDL 转译 / 数据库准备 / 集群能力查询 / 启动期钩子"中的任一行为时
- **则** 调用方通过 `dialect.From(link)` 获取当前方言实例
- **且** 调用方仅依赖 `Dialect` 接口的方法签名，不依赖具体实现的内部细节

#### Scenario: 具体方言实现不作为公共 API 暴露
- **当** 宿主、插件生命周期或工具链代码导入 `apps/lina-core/pkg/dialect` 时
- **则** 公共包不导出 `MySQLDialect` / `SQLiteDialect` 等具体实现类型
- **且** MySQL / SQLite 的 DDL 转译、数据库准备、启动期行为和驱动错误分类实现分别维护在 `pkg/dialect/internal/mysql` 与 `pkg/dialect/internal/sqlite` 内部子包中
- **且** 调用方只能通过 `Dialect` 接口、`dialect.From(link)` 工厂函数和 `dialect.IsRetryableWriteConflict(err)` / `dialect.SplitSQLStatements(content)` 等必要公共门面能力访问方言相关行为

#### Scenario: 驱动错误分类由 dialect 公共包提供
- **当** `kvcache incr` 等共享组件需要判断数据库写入冲突是否可重试
- **则** 调用方通过 `dialect.IsRetryableWriteConflict(err)` 判断
- **且** 调用方不得硬编码 MySQL / SQLite 错误文案、错误码或具体驱动错误类型
- **且** `pkg/dialect` 使用驱动暴露的结构化错误码进行分类，错误文案匹配最多只能作为方言包内部的显式兜底

#### Scenario: dialect 公共包不暴露宿主 internal 具体类型
- **当** 插件生命周期、初始化命令或工具链代码导入 `apps/lina-core/pkg/dialect` 时
- **则** 公开接口不要求调用方引用 `apps/lina-core/internal/...` 下的具体服务类型
- **且** 启动期配置覆盖能力通过 `dialect.RuntimeConfig` 等窄接口适配
- **且** 宿主 `config.Service` 可在内部实现该窄接口后传入 `Dialect.OnStartup`

### Requirement: 方言根据数据库链接前缀自动分发

系统 SHALL 根据 `database.default.link` 配置的协议头自动选择对应的方言实现。`mysql:` 前缀分发到 MySQL 方言实现，`sqlite:` 前缀分发到 SQLite 方言实现。未识别的前缀必须返回明确的错误。调用方不得依赖或断言具体实现类型，只能依赖 `Dialect` 接口行为。

#### Scenario: MySQL 链接被识别为 MySQL 方言
- **当** 配置文件 `database.default.link` 以 `mysql:` 开头时
- **则** `dialect.From(link)` 返回实现 `Dialect` 接口的 MySQL 方言实例
- **且** `Name()` 返回字符串 `"mysql"`
- **且** `SupportsCluster()` 返回 `true`

#### Scenario: SQLite 链接被识别为 SQLite 方言
- **当** 配置文件 `database.default.link` 以 `sqlite:` 开头时
- **则** `dialect.From(link)` 返回实现 `Dialect` 接口的 SQLite 方言实例
- **且** `Name()` 返回字符串 `"sqlite"`
- **且** `SupportsCluster()` 返回 `false`

#### Scenario: 未识别的链接前缀
- **当** 配置文件 `database.default.link` 以未识别的前缀开头时
- **则** `dialect.From(link)` 返回包含前缀名与已支持前缀列表的明确错误
- **且** 系统不静默回退到任何默认方言

### Requirement: MySQL 方言 DDL 转译为无操作

`Dialect.TranslateDDL(ctx, sourceName, ddl)` SHALL 接收调用方传入的 `sourceName` 诊断名。`sourceName` MUST 是源 SQL 文件路径、嵌入资产路径或调用方构造的稳定描述，用于在错误消息中定位失败来源。MySQL 方言的 `TranslateDDL` SHALL 直接返回输入字符串，不做任何修改。这保证了 MySQL 用户在引入方言抽象层后行为完全向后兼容，不会因转译副作用引入新的失败路径。

#### Scenario: MySQL 方言转译保持原文
- **当** MySQL 方言实例的 `TranslateDDL(ctx, sourceName, ddl)` 被调用时
- **则** 返回值与输入 `ddl` 字节级别完全一致
- **且** 不返回错误（除非输入本身为 `nil` 或空字符串等显式无效输入）
- **且** `sourceName` 不影响 MySQL no-op 转译结果

### Requirement: SQLite 方言 DDL 转译必须覆盖项目 DDL 子集

SQLite 方言的 `TranslateDDL(ctx, sourceName, ddl)` SHALL 将单一 MySQL 方言来源的 DDL / seed / mock SQL 转译为可在 SQLite 上成功执行的语句。转译必须覆盖项目当前 SQL 文件中实际使用的所有 MySQL 语法，包括反引号标识符、`AUTO_INCREMENT`、`UNSIGNED`、`TINYINT` / `SMALLINT` / `LONGTEXT` 类型、`ENGINE=` / `CHARSET=` / `COLLATE=` 子句、列级与表级 `COMMENT '...'`、`INSERT IGNORE`、`DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`、表内联 `KEY` / `INDEX` / `UNIQUE KEY` / `UNIQUE INDEX`、表达式索引、表级 `PRIMARY KEY`、`CREATE DATABASE` / `USE` 整句、当前 mock SQL 真实出现的 `CONCAT(...)`。转译覆盖范围 SHALL 以当前宿主安装 SQL、插件安装 SQL、宿主 mock SQL、插件 mock SQL 与插件卸载 SQL 的真实写法为验收基准，而不是仅覆盖文档中的示例写法。

#### Scenario: MEMORY 引擎子句被去除
- **当** 输入 DDL 包含 `... ) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='Distributed lock table';` 时
- **则** 转译结果不包含 `ENGINE=` / `CHARSET=` / `COMMENT=` 任一子句
- **且** 表创建语义保留：表本身被创建，但作为 SQLite 普通表（持久化）

#### Scenario: AUTO_INCREMENT 主键被改写
- **当** 输入 DDL 包含 `id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT` 时
- **则** 转译结果中该列定义为 `id INTEGER PRIMARY KEY AUTOINCREMENT`
- **且** 不保留 `BIGINT` / `UNSIGNED` 等 SQLite 不支持的修饰符

#### Scenario: AUTO_INCREMENT 主键真实排列被全部改写
- **当** 输入 DDL 包含 `id INT PRIMARY KEY AUTO_INCREMENT`、`id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY`、`id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT` + 表级 `PRIMARY KEY(id)` 或带反引号的等价写法时
- **则** 转译结果为 SQLite 可执行的 `INTEGER PRIMARY KEY AUTOINCREMENT` 语义
- **且** 不因主键关键字顺序、`NOT NULL` 位置或表级主键写法不同而漏转译
- **且** 现有 SQL 文件中的所有自增主键表均可在 SQLite 上创建成功

#### Scenario: 列类型被映射为 SQLite 等价类型
- **当** 输入 DDL 包含 `VARCHAR(64)` / `LONGTEXT` / `TINYINT` / `DECIMAL(10,2)` 任一列定义时
- **则** 转译结果对应列分别映射为 `TEXT` / `TEXT` / `INTEGER` / `NUMERIC`

#### Scenario: INSERT IGNORE 被改写
- **当** 输入 DDL 包含 `INSERT IGNORE INTO sys_user (...)` 语句时
- **则** 转译结果改写为 `INSERT OR IGNORE INTO sys_user (...)`
- **且** 写入语义（重复键时跳过）保持等价

#### Scenario: CONCAT 被改写为 SQLite 字符串拼接
- **当** 输入 mock SQL 包含 `CONCAT('0,', parent.id)` 语句时
- **则** 转译结果改写为 SQLite 可执行的字符串拼接表达式，如 `('0,' || parent.id)`
- **且** 转译结果在临时 SQLite 数据库上成功执行

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
- **当** 输入 DDL 在 CREATE TABLE 内包含 `KEY idx_status (status), INDEX idx_phone (phone), UNIQUE KEY uk_name (name), UNIQUE INDEX uk_node (node_name, node_ip)` 子句时
- **则** 转译结果中 CREATE TABLE 仅保留列定义与 PRIMARY KEY、UNIQUE 约束
- **且** 内联索引被提取为表创建语句之后的 `CREATE INDEX idx_status ON tbl(status);` 等独立语句
- **且** `UNIQUE KEY` 与 `UNIQUE INDEX` 均转译为 `CREATE UNIQUE INDEX ...`

#### Scenario: 表内联表达式索引被保留为 SQLite 可执行索引
- **当** 输入 DDL 在 CREATE TABLE 内包含 `UNIQUE KEY uk_plugin_org_center_dept_code ((NULLIF(code, '')))` 等当前 SQL 已使用的表达式索引时
- **则** 转译器将其提取为表创建后的独立唯一索引语句
- **且** 表达式内容保持 SQLite 可执行
- **且** 转译结果在临时 SQLite 数据库上成功执行

#### Scenario: CREATE DATABASE 与 USE 整句被丢弃
- **当** 输入 DDL 包含 `CREATE DATABASE IF NOT EXISTS \`linapro\` ...;` 或 `USE \`linapro\`;` 整句时
- **则** 转译结果不包含这些语句
- **且** 转译器不报错（SQLite 没有"数据库"概念，丢弃即正确语义）

#### Scenario: 反引号标识符被去除或正常化
- **当** 输入 DDL 包含 `` `id` `` / `` `sys_user` `` 等反引号包裹的标识符时
- **则** 转译结果中标识符不带反引号（直接裸写）或使用双引号
- **且** 转译结果在 SQLite 上可成功执行

### Requirement: DDL 转译失败时必须返回明确错误

SQLite 方言的 `TranslateDDL(ctx, sourceName, ddl)` 在遇到当前实现未覆盖的 MySQL 语法时 SHALL 返回包含 `sourceName`、行号定位提示与未覆盖语法关键字的明确错误，不得静默丢弃或产生无效 SQL。

#### Scenario: 转译器遇到未覆盖的语法
- **当** 输入 DDL 包含未在覆盖范围内的 MySQL 特性（如 `FULLTEXT INDEX` / `GENERATED ALWAYS AS` / 分区子句等）时
- **则** 转译器返回错误，错误消息包含 `sourceName`、行号提示与未覆盖的关键字
- **且** 调用方（`cmd init` / `cmd mock` / 插件 install pipeline）将错误向上传播
- **且** 系统不执行任何已部分转译的 SQL 内容

### Requirement: 方言必须暴露数据库准备入口

`Dialect.PrepareDatabase(ctx, link, rebuild)` SHALL 负责在执行 DDL 资源前完成方言相关的数据库准备工作。MySQL 方言执行 `CREATE DATABASE IF NOT EXISTS` 与可选的 `DROP DATABASE`；SQLite 方言执行父目录创建（`mkdir -p`）与可选的数据库文件删除。

#### Scenario: MySQL 方言准备数据库
- **当** MySQL 方言实例的 `PrepareDatabase(ctx, link, rebuild=false)` 被调用时
- **则** 系统执行 `CREATE DATABASE IF NOT EXISTS linapro` 等价语句
- **且** 不删除已存在的数据库

#### Scenario: MySQL 方言重建数据库
- **当** MySQL 方言实例的 `PrepareDatabase(ctx, link, rebuild=true)` 被调用时
- **则** 系统先执行 `DROP DATABASE IF EXISTS linapro`
- **且** 再执行 `CREATE DATABASE linapro`
- **且** 启动日志输出明确的 rebuild 警告

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

#### Scenario: SQLite 父目录不可创建
- **当** SQLite 数据库文件父目录创建失败（权限不足、磁盘满等）时
- **则** 系统返回包含目标路径的明确错误
- **且** 不继续后续 DDL 执行

### Requirement: 方言必须提供启动期钩子

`Dialect.OnStartup(ctx, runtime)` SHALL 在宿主启动 bootstrap 阶段被调用一次。`runtime` SHALL 是 `pkg/dialect` 中定义的稳定窄接口，至少提供方言锁定 `cluster.enabled` 所需的方法。MySQL 方言为 no-op；SQLite 方言负责执行"强制覆盖 cluster.enabled=false + 输出警告日志"等启动期专属行为。该钩子的调用时机必须早于任何 cluster 相关初始化。

#### Scenario: MySQL 启动期钩子无副作用
- **当** MySQL 方言实例的 `OnStartup(ctx, runtime)` 被调用时
- **则** 钩子立即返回 `nil`
- **且** 不修改任何配置项
- **且** 不输出任何警告级别日志

#### Scenario: SQLite 启动期钩子锁定集群配置
- **当** SQLite 方言实例的 `OnStartup(ctx, runtime)` 被调用时
- **则** `configSvc.IsClusterEnabled(ctx)` 在该钩子调用后稳定返回 `false`
- **且** 该覆盖优先级高于 `config.yaml` 中的 `cluster.enabled` 显式声明
- **且** 钩子向终端输出至少 4 行 `[WARNING]` 级别日志，明确告知 SQLite 模式、cluster 锁定原因、单机部署限制、不得用于生产

#### Scenario: 启动期钩子在集群初始化前执行
- **当** 宿主以 SQLite 模式启动时
- **则** `OnStartup` 在 `cluster.Service` 启动选举循环前被调用
- **且** 后续 cluster 相关组件读取到的 `IsClusterEnabled` 已为 `false`
- **且** 不会出现"先启动选举循环再被关闭"的中间状态
