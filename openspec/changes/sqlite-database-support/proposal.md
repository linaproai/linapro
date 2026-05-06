## Why

`LinaPro` 当前仅支持 `MySQL` 一种数据库引擎，配置文件 `database.default.link` 强制要求 `mysql:user@tcp(...)` 链接，所有 14 个宿主 SQL 文件、7 个插件 SQL 文件以及 `cmd init/mock` 流程都与 MySQL 方言深度耦合，并且 `kvcache` 中存在 MySQL 专用的 `LAST_INSERT_ID(value_int + delta)` 原子自增技巧。这给"演示 / 个人测试 / 零依赖本地试用"场景设置了不必要的门槛——希望快速试用框架的用户必须先安装并配置 MySQL，否则连 `make init` 都无法跑通。本变更引入 `SQLite` 作为面向"演示 / 测试"场景的可选数据库引擎，使用户只需修改 `config.yaml` 的一行链接配置即可零依赖运行整个框架。

## What Changes

### 新增方言抽象层

- **新增** `apps/lina-core/pkg/dialect/` 包，定义 `Dialect` 接口（`Name()` / `TranslateDDL()` / `PrepareDatabase()` / `SupportsCluster()` / `OnStartup()`）作为数据库引擎差异收敛的统一边界
- **新增** `MySQLDialect` 实现：`TranslateDDL` 为 no-op，`PrepareDatabase` 沿用现有 `DROP DATABASE` / `CREATE DATABASE` 行为，`SupportsCluster` 返回 `true`
- **新增** `SQLiteDialect` 实现：`TranslateDDL` 调用 SQLite DDL 转译器，`PrepareDatabase` 在 `rebuild=true` 时删除数据库文件，`SupportsCluster` 返回 `false`，`OnStartup` 强制覆盖 `cluster.enabled=false` 并输出警告日志
- **新增** SQLite DDL 转译器，覆盖以下 MySQL → SQLite 方言映射：
  - 反引号标识符去除
  - `INT/BIGINT [UNSIGNED] AUTO_INCREMENT` → `INTEGER PRIMARY KEY AUTOINCREMENT`
  - `TINYINT / SMALLINT [UNSIGNED]` → `INTEGER`
  - `VARCHAR(N) / CHAR(N) / LONGTEXT` → `TEXT`
  - `DECIMAL(M,N)` → `NUMERIC`
  - `INSERT IGNORE INTO` → `INSERT OR IGNORE INTO`
  - 删除 `ENGINE=` / `DEFAULT CHARSET=` / `COLLATE=` / 列级与表级 `COMMENT '...'`
  - 仅删除 `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` 中的 `ON UPDATE` 部分（DAO 层已自动维护 `updated_at`）
  - 表内联 `KEY` / `INDEX` / `UNIQUE KEY` 提取为表创建后的 `CREATE INDEX` 独立语句
  - `CREATE DATABASE` / `USE database` 整句删除

### 修改宿主引导流程

- **修改** `apps/lina-core/internal/cmd/cmd_init_database.go`，将 MySQL 专属的链接解析与 `DROP/CREATE DATABASE` 逻辑迁移到 `MySQLDialect.PrepareDatabase`；`prepareInitDatabase` 改为按 `link` 协议头分发到对应方言
- **修改** `cmd_init.go` / `cmd_mock.go` 在执行 SQL 文件前先调用当前方言的 `TranslateDDL`，将单一的 MySQL 来源 DDL 转换为目标方言可执行语句
- **修改** SQLite 模式下，`rebuild=true` 改为删除数据库文件；启动时若数据库文件父目录不存在则自动 `mkdir -p`

### 修改集群模式开关

- **修改** `cluster-deployment-mode` 规范：在 SQLite 链接下，`cluster.enabled` 配置值在内存层被强制覆盖为 `false`，无论用户在 `config.yaml` 中写的是什么；启动期间输出醒目的 `WARNING` 日志，明确告知"当前为 SQLite 模式，仅支持单节点部署，所有功能在单机模式下运行，请勿用于生产"

### 修改插件清单生命周期

- **修改** 插件安装、卸载、初始化数据加载流程在执行 `manifest/sql/` 与 `manifest/sql/uninstall/` 与 `manifest/sql/mock-data/` 下的 SQL 资源前，统一通过当前方言的 `TranslateDDL` 转译；插件源码侧无需任何改动，单一 MySQL 方言来源的 SQL 文件在 SQLite 模式下也可正确执行

### 重命名 kvcache backend 并移除 MySQL 专用语法

- **BREAKING（内部）** `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` 重命名为 `sqltable/`；常量 `BackendMySQLMemory` 重命名为 `BackendSQLTable`，字符串值由 `"mysql-memory"` 改为 `"sql-table"`
- **修改** `Incr` 操作：将 `LAST_INSERT_ID(value_int + delta)` + `SELECT LAST_INSERT_ID()` 的 MySQL 专用模式改为方言中性的"事务内 `UPDATE ... SET value_int = value_int + ?` + `SELECT value_int`"模式
- **修改** `plugin-cache-service` 规范：将"MySQL `MEMORY` 表"等引擎专属用语软化为"易失性缓存表 / ephemeral cache table"等中性表述，使规范同时适用于 MySQL `MEMORY` 引擎和 SQLite 普通表两种存储后端

### 修改配置默认值与依赖

- **修改** `apps/lina-core/manifest/config/config.template.yaml` 与 `config.yaml` 在注释中说明 SQLite 链接示例（默认值仍为 MySQL 链接以保持 MySQL 用户的零修改体验）
- **新增** `go.mod` 依赖：`github.com/gogf/gf/contrib/drivers/sqlite/v2`
- **新增** `.gitignore` 条目：`temp/sqlite/*.db*`（若 `temp/` 未被忽略）

### 新增 E2E 与单元测试

- **新增** SQLite DDL 转译器单元测试，对 14 个宿主 SQL 文件 + 7 个插件 SQL 文件 + 7 个 mock SQL 文件分别构建 fixture，断言转译结果可在 SQLite 上成功执行
- **新增** SQLite 模式 E2E 用例：复用现有用例集，验证业务模块在 SQLite 引擎下行为一致、零感知
- **新增** 启动期 cluster 锁定与警告日志的单元测试

## Capabilities

### New Capabilities

- `database-dialect-abstraction`：定义数据库方言抽象层、`Dialect` 接口与基于链接前缀的方言分发机制，作为 MySQL 与 SQLite 等多种数据库引擎差异的唯一收敛点；定义 SQLite DDL 转译器对 MySQL 方言 DDL 的覆盖范围与可执行结果保证

### Modified Capabilities

- `database-bootstrap-commands`：新增按方言分发的初始化 / 重建语义；新增执行 SQL 资源前必须经方言转译的需求；新增 SQLite 数据库文件路径与父目录自动创建语义
- `cluster-deployment-mode`：新增 SQLite 链接下 `cluster.enabled` 强制为 `false` 的需求；新增启动期警告日志的可见性需求
- `plugin-cache-service`：将"MySQL `MEMORY` 表"等引擎专属表述改为"易失性缓存表"中性表述；新增"重启不丢失"语义在 SQLite 模式下亦成立的兼容性说明（缓存仍按有损缓存对待，关键业务状态不依赖缓存表）
- `plugin-manifest-lifecycle`：新增插件 SQL 资源（`manifest/sql/` / `uninstall/` / `mock-data/`）执行前必须经当前方言转译的需求

## Impact

### 受影响代码

- `apps/lina-core/pkg/dialect/`（新增）
- `apps/lina-core/internal/cmd/cmd_init_database.go`、`cmd_init.go`、`cmd_mock.go`（重构）
- `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` → `sqltable/`（重命名 + 重写 `Incr`）
- `apps/lina-core/internal/service/kvcache/kvcache_backend.go`（常量名调整）
- `apps/lina-core/internal/service/plugin/`（插件 install/uninstall pipeline 接入方言转译）
- `apps/lina-core/internal/cmd/cmd_http_runtime.go` 或等价启动 bootstrap 入口（SQLite 启动期锁 cluster + 警告日志）
- `apps/lina-core/manifest/config/config.template.yaml`、`config.yaml`（注释新增 SQLite 链接示例）

### 受影响测试

- 新增方言转译器单元测试（覆盖 28 个 SQL 文件）
- 新增启动期 cluster 锁定与日志输出的单元测试
- 复用现有 E2E 套件，新增 SQLite 模式参数化执行通道

### 受影响依赖

- `apps/lina-core/go.mod` 增加 `github.com/gogf/gf/contrib/drivers/sqlite/v2`
- `.gitignore` 增加 `temp/sqlite/*.db*` 条目

### 不受影响

- 业务模块（`controller` / `service` / `model` / `dao`）：零代码改动，对数据库引擎差异无感知（约束 #3）
- 已有 MySQL 用户：默认配置仍为 MySQL，行为完全向后兼容
- 插件源码：单一 MySQL 方言 SQL 文件继续工作，SQLite 由方言层透明转译
- `gf gen dao` / `gf gen ctrl` 工作流：保持不变
- `apidoc` / i18n 资源：不受方言切换影响
