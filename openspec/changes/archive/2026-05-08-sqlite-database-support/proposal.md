## Why

`LinaPro` 当前仅支持 `MySQL` 一种数据库引擎，配置文件 `database.default.link` 强制要求 `mysql:user@tcp(...)` 链接，当前宿主与插件的安装、mock、卸载 SQL 资产以及 `cmd init/mock` 流程都与 MySQL 方言深度耦合，并且 `kvcache` 中存在 MySQL 专用的 `LAST_INSERT_ID(value_int + delta)` 原子自增技巧。这给"演示 / 个人测试 / 零依赖本地试用"场景设置了不必要的门槛——希望快速试用框架的用户必须先安装并配置 MySQL，否则连 `make init` 都无法跑通。本变更引入 `SQLite` 作为面向"演示 / 测试"场景的可选数据库引擎，使用户只需修改 `config.yaml` 的一行链接配置即可零依赖运行整个框架。

## What Changes

### 新增方言抽象层

- **新增** `apps/lina-core/pkg/dialect/` 公共稳定包，定义 `Dialect` 接口（`Name()` / `TranslateDDL(ctx, sourceName, ddl)` / `PrepareDatabase()` / `SupportsCluster()` / `OnStartup()`）作为数据库引擎差异收敛的统一边界；该包暴露稳定窄接口，禁止在公开签名中绑定宿主 `internal` 具体服务类型，MySQL / SQLite 具体实现收敛到 `pkg/dialect/internal/mysql` 与 `pkg/dialect/internal/sqlite`
- **新增** MySQL 内部方言实现：`TranslateDDL` 为 no-op，`PrepareDatabase` 沿用现有 `DROP DATABASE` / `CREATE DATABASE` 行为，`SupportsCluster` 返回 `true`
- **新增** SQLite 内部方言实现：`TranslateDDL` 调用 SQLite DDL 转译器，`PrepareDatabase` 在 `rebuild=true` 时删除数据库文件，`SupportsCluster` 返回 `false`，`OnStartup` 强制覆盖 `cluster.enabled=false` 并输出启动提示日志
- **新增** SQLite DDL 转译器，覆盖以下 MySQL → SQLite 方言映射：
  - 反引号标识符去除
  - 当前所有 SQL 中真实出现的 `INT/BIGINT [UNSIGNED] AUTO_INCREMENT PRIMARY KEY` / `PRIMARY KEY AUTO_INCREMENT` / 表级 `PRIMARY KEY(id)` 组合 → SQLite 可执行的 `INTEGER PRIMARY KEY AUTOINCREMENT` 语义
  - `TINYINT / SMALLINT [UNSIGNED]` → `INTEGER`
  - `VARCHAR(N) / CHAR(N) / LONGTEXT` → `TEXT`
  - `DECIMAL(M,N)` → `NUMERIC`
  - `INSERT IGNORE INTO` → `INSERT OR IGNORE INTO`
  - 删除 `ENGINE=` / `DEFAULT CHARSET=` / `COLLATE=` / 列级与表级 `COMMENT '...'`
  - 仅删除 `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` 中的 `ON UPDATE` 部分（DAO 层已自动维护 `updated_at`）
  - 表内联 `KEY` / `INDEX` / `UNIQUE KEY` / `UNIQUE INDEX`（含当前 SQL 中的表达式索引）提取为表创建后的 `CREATE INDEX` 独立语句
  - 当前 mock SQL 真实出现的 `CONCAT(a, b, ...)` 转译为 SQLite 字符串拼接表达式 `a || b || ...`
  - `CREATE DATABASE` / `USE database` 整句删除

### 修改宿主引导流程

- **修改** `apps/lina-core/internal/cmd/cmd_init_database.go`，将 MySQL 专属的链接解析与 `DROP/CREATE DATABASE` 逻辑迁移到 MySQL 方言的 `PrepareDatabase`；`prepareInitDatabase` 改为按 `link` 协议头分发到对应方言
- **修改** `cmd_init.go` / `cmd_mock.go` 在执行 SQL 文件前先调用当前方言的 `TranslateDDL(ctx, sourceName, ddl)`，将单一的 MySQL 来源 DDL 转换为目标方言可执行语句；`sourceName` 统一使用源文件路径或嵌入资产路径，便于错误消息携带文件定位
- **明确** `make mock` 必须依赖已初始化数据库，不负责创建、重建或准备数据库；未执行 `make init` 时应快速失败并返回现有数据库错误
- **修改** SQLite 模式下，`rebuild=true` 改为删除数据库文件；启动时若数据库文件父目录不存在则自动 `mkdir -p`

### 修改集群模式开关

- **修改** `cluster-deployment-mode` 规范：在 SQLite 链接下，`cluster.enabled` 配置值在内存层被强制覆盖为 `false`，无论用户在 `config.yaml` 中写的是什么；启动期间输出清晰提示，明确告知"当前为 SQLite 模式，仅支持单节点部署，所有功能在单机模式下运行，请勿用于生产"

### 修改插件清单生命周期

- **修改** 插件安装、卸载、初始化数据加载流程在执行 `manifest/sql/` 与 `manifest/sql/uninstall/` 与 `manifest/sql/mock-data/` 下的 SQL 资源前，统一通过当前方言的 `TranslateDDL` 转译；插件源码侧无需任何改动，单一 MySQL 方言来源的 SQL 文件在 SQLite 模式下也可正确执行

### 重命名 kvcache backend 并移除 MySQL 专用语法

- **BREAKING（内部）** `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` 重命名为 `sqltable/`；常量 `BackendMySQLMemory` 重命名为 `BackendSQLTable`，字符串值由 `"mysql-memory"` 改为 `"sql-table"`
- **修改** `Incr` 操作：将 `LAST_INSERT_ID(value_int + delta)` + `SELECT LAST_INSERT_ID()` 的 MySQL 专用模式改为方言中性的 CAS 重试模式：读取当前整数快照，缺失时用 `INSERT IGNORE value_int=0` 幂等初始化，再执行带 `value_int=<snapshot>` 条件的参数化 `UPDATE` 写入新值；竞争写入导致 `affected=0` 时有限退避重试，保证成功调用线性递增
- **修改** `plugin-cache-service` 规范：默认 MySQL 交付 SQL 仍保持原始 `sys_kv_cache ENGINE=MEMORY` 表结构与引擎类型；SQLite 仅由 DDL 转译器去除引擎子句后得到普通 SQLite 表。两种模式都通过应用层 TTL 与定时清理兜底，并继续把缓存视为有损缓存

### 修改配置默认值与依赖

- **修改** `apps/lina-core/manifest/config/config.template.yaml` 与 `config.yaml` 在注释中说明 SQLite 链接示例（默认值仍为 MySQL 链接以保持 MySQL 用户的零修改体验）
- **新增** `go.mod` 依赖：`github.com/gogf/gf/contrib/drivers/sqlite/v2`
- **确认** SQLite 默认路径位于已被仓库根 `.gitignore` 忽略的 `temp/` 目录下，无需新增专用 SQLite 忽略条目

### 新增 E2E 与单元测试

- **新增** SQLite DDL 转译器单元测试，按当前仓库真实 SQL 资产自动扫描宿主安装 SQL、插件安装 SQL、宿主 mock SQL、插件 mock SQL 与插件卸载 SQL，断言转译结果可在 SQLite 上成功执行
- **新增** SQLite 模式 E2E 用例：通过测试夹具在启动前写入测试配置文件来切换 `database.default.link`，不引入命令行参数或环境变量作为运行时方言来源，验证业务模块在 SQLite 引擎下行为一致、零感知；主 CI 仅运行轻量后端 SQLite smoke，完整 SQLite E2E 保留为手动验证入口
- **新增** 启动期 cluster 锁定与启动提示日志的单元测试

## Capabilities

### New Capabilities

- `database-dialect-abstraction`：定义数据库方言抽象层、`Dialect` 接口与基于链接前缀的方言分发机制，作为 MySQL 与 SQLite 等多种数据库引擎差异的唯一收敛点；定义 SQLite DDL 转译器对 MySQL 方言 DDL 的覆盖范围与可执行结果保证

### Modified Capabilities

- `database-bootstrap-commands`：新增按方言分发的初始化 / 重建语义；新增执行 SQL 资源前必须经方言转译的需求；新增 SQLite 数据库文件路径与父目录自动创建语义
- `cluster-deployment-mode`：新增 SQLite 链接下 `cluster.enabled` 强制为 `false` 的需求；新增启动期提示日志的可见性需求
- `plugin-cache-service`：明确 MySQL 交付 SQL 中 `sys_kv_cache` 保持既有 `ENGINE=MEMORY` 表结构与引擎类型；SQLite 模式仅在执行期通过 DDL 转译器去除引擎子句，并继续把缓存视为有损缓存
- `plugin-manifest-lifecycle`：新增插件 SQL 资源（`manifest/sql/` / `uninstall/` / `mock-data/`）执行前必须经当前方言转译的需求

## Impact

### 受影响代码

- `apps/lina-core/pkg/dialect/`（新增）
- `apps/lina-core/internal/cmd/cmd_init_database.go`、`cmd_init.go`、`cmd_mock.go`（重构）
- `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` → `sqltable/`（重命名 + 重写 `Incr`）
- `apps/lina-core/internal/service/kvcache/kvcache_backend.go`（常量名调整）
- `apps/lina-core/internal/service/plugin/`（插件 install/uninstall pipeline 接入方言转译）
- `apps/lina-core/internal/cmd/cmd_http_runtime.go` 或等价启动 bootstrap 入口（SQLite 启动期锁 cluster + 启动提示日志）
- `apps/lina-core/manifest/config/config.template.yaml`、`config.yaml`（注释新增 SQLite 链接示例）

### 受影响测试

- 新增方言转译器单元测试（覆盖当前宿主 / 插件安装、mock、卸载 SQL 资产）
- 新增启动期 cluster 锁定与日志输出的单元测试
- 复用现有 E2E 套件，新增 SQLite 模式参数化执行通道；主 CI 使用后端 SQLite smoke 覆盖启动提示、单节点 health 与管理员登录链路

### 受影响依赖

- `apps/lina-core/go.mod` 增加 `github.com/gogf/gf/contrib/drivers/sqlite/v2`
- `.gitignore` 不需要新增 SQLite 专用条目；仓库根已有 `temp/` 忽略规则，覆盖默认 SQLite 数据库文件路径

### 不受影响

- 业务模块（`controller` / `service` / `model` / `dao`）：零代码改动，对数据库引擎差异无感知（约束 #3）
- 已有 MySQL 用户：默认配置仍为 MySQL，行为完全向后兼容
- 插件源码：单一 MySQL 方言 SQL 文件继续工作，SQLite 由方言层透明转译
- `gf gen dao` / `gf gen ctrl` 工作流：保持不变
- `apidoc` / i18n 资源：不受方言切换影响
