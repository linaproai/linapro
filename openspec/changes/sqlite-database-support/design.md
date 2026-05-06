## Context

`LinaPro` 的数据库设计当前完全围绕 `MySQL` 展开，体现在以下事实层面：

- **链接配置**：`database.default.link` 默认值为 `mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true`，且 `cmd_init_database.go` 中的 `databaseNameFromMySQLLink` / `serverLinkFromMySQLLink` / `quoteMySQLIdentifier` 三个工具函数硬编码假设链接是 MySQL 协议
- **DDL 方言**：14 个宿主 SQL 文件 + 7 个插件 SQL 文件 + 7 个 mock SQL 文件统一使用 MySQL 方言，包含 `ENGINE=InnoDB` / `ENGINE=MEMORY` / `AUTO_INCREMENT` / `BIGINT UNSIGNED` / `TINYINT` / `LONGTEXT` / 反引号标识符 / 列级与表级 `COMMENT '...'` / `DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci` / `INSERT IGNORE` / `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` / 表内联 `KEY` 与 `INDEX` / `CREATE DATABASE IF NOT EXISTS` 等 SQLite 不支持或语法不同的特性
- **MEMORY 引擎专用语义**：`sys_locker`（010）、`sys_kv_cache`（013，动态插件宿主服务扩展）、`sys_online_session`（006）三张表使用 MySQL `MEMORY` 引擎，借其"重启即清"的语义简化分布式锁、易失性 KV 缓存与在线会话的清理逻辑
- **应用层 MySQL 专用代码**：`apps/lina-core/internal/service/kvcache/internal/mysql-memory/mysql_memory_ops.go` 第 257、272 行使用 `gdb.Raw("LAST_INSERT_ID(value_int + " + delta + ")")` 与 `Raw("SELECT LAST_INSERT_ID()")` 实现"原子自增并取回新值"，这是 MySQL 的会话级 Last-Insert-ID 技巧，SQLite 无此机制
- **集群协调依赖**：`sys_locker` / `sys_kv_cache` 既是业务存储又是多节点共享协调的载体，依赖 MySQL 跨连接共享可见性

GoFrame 已通过 `github.com/gogf/gf/contrib/drivers/sqlite/v2` 原生支持 SQLite，链接格式为 `sqlite::path/to/file.db`；DAO/DO/Entity 层的 `do.Xxx` 模式与 GoFrame 内置的查询构建器对绝大多数业务 SQL 自动方言化，加上项目已有规则禁止 `FIND_IN_SET` / `GROUP_CONCAT` / `IF()` / `ON DUPLICATE KEY UPDATE` 等数据库专用函数，DAO 层的运行时查询基本可移植。需要额外处理的只有 DDL 方言、初始化流程、MEMORY 引擎专用语义、kvcache 的 MySQL 专用 SQL 这四类。

用户对本次迭代的约束已澄清：

1. SQLite 仅服务于演示 / 个人测试场景，不支持生产
2. 不需要 SQLite ↔ MySQL 数据迁移，SQLite 数据可丢失
3. 业务模块必须对数据库引擎差异零感知，所有适配收敛在数据访问层
4. SQLite 模式必须强制 `cluster.enabled=false` 并在终端日志中输出醒目提示

## Goals / Non-Goals

**Goals:**

- 用户在 `config.yaml` 中将 `database.default.link` 改为 `sqlite::./temp/sqlite/linapro.db` 即可零依赖运行整个框架，无需启动 MySQL
- 业务模块（`controller` / `service` / `model` / `dao`）零代码改动，对底层数据库引擎差异完全无感知
- 单一 MySQL 方言 SQL 文件来源（保持"同迭代单文件"原则），不引入并存的 `xxx.mysql.sql` / `xxx.sqlite.sql`
- 插件 SQL 资源在 SQLite 模式下也可正确执行，插件源码无需任何改动
- SQLite 模式启动时强制锁定单节点模式，并在终端打印明确的警告日志，避免误用
- 默认 SQLite 数据库文件路径 `./temp/sqlite/linapro.db` 可通过修改 `link` 配置自定义；启动时若父目录不存在则自动创建
- 现有 MySQL 用户的默认配置与运行时行为完全向后兼容

**Non-Goals:**

- 不支持 SQLite 集群部署（架构上不可能，SQLite 是嵌入式单实例数据库）
- 不实现 SQLite ↔ MySQL 数据双向迁移工具
- 不实现 SQLite 模式下的"启动期生产环境检测安全网"（用户明确将其推迟到后续迭代）
- 不修改业务模块代码、不引入数据库引擎相关的业务分支判断
- 不优化 SQLite 模式下的性能（演示 / 测试场景的并发要求极低）
- 不为 SQLite 模式重新设计或裁剪现有 MEMORY 表（在 SQLite 下退化为普通表即可，不影响正确性）

## Decisions

### 决策一：单点收敛适配，定义 `pkg/dialect` 抽象层

**选择**：新增 `apps/lina-core/pkg/dialect/` 包，定义统一的 `Dialect` 接口作为数据库引擎差异收敛的唯一边界。所有方言相关逻辑（DDL 转译、数据库准备、集群能力查询、启动期钩子）都通过该接口暴露给上层调用方。

```go
// Dialect 抽象数据库引擎相关的所有差异点。
type Dialect interface {
    // Name 返回方言名（如 "mysql" / "sqlite"），用于日志与诊断输出。
    Name() string

    // TranslateDDL 将单一 MySQL 方言来源的 DDL 内容转译为目标方言可执行的语句。
    // MySQL 实现为 no-op；SQLite 实现执行 token 替换 + 结构提取。
    TranslateDDL(ctx context.Context, ddl string) (string, error)

    // PrepareDatabase 在执行 DDL 资源前准备数据库（创建库 / 文件、可选 rebuild）。
    // MySQL 实现执行 DROP/CREATE DATABASE；SQLite 实现 mkdir 父目录、可选删除文件。
    PrepareDatabase(ctx context.Context, link string, rebuild bool) error

    // SupportsCluster 表明本方言能否承担多节点共享状态协调的角色。
    // MySQL=true，SQLite=false。
    SupportsCluster() bool

    // OnStartup 在启动 bootstrap 阶段调用，用于方言相关的运行时初始化
    // （如 SQLite 模式锁定 cluster.enabled 并输出警告日志）。
    OnStartup(ctx context.Context, configSvc config.Service) error
}

// From 根据 link 的协议头返回对应方言实现。
// 链接 "mysql:..." → MySQLDialect；"sqlite::..." → SQLiteDialect。
func From(link string) (Dialect, error)
```

**理由**：

- 业务模块零感知：所有方言差异不在业务路径上出现，约束 #3 落地的物理保证
- 单一扩展点：未来支持 PostgreSQL 仅需增加一个 `Dialect` 实现，无需修改任何调用方
- 测试隔离：方言层可被独立单元测试覆盖（DDL 转译、链接解析），不依赖真实数据库
- 启动期钩子集中：`OnStartup` 让"SQLite 锁 cluster + 警告"等启动期行为有明确归宿，不散落在各处

**替代方案与拒绝理由**：

- *方案 A：在每个调用点 `if isSQLite { ... } else { ... }`*。被拒绝：违反约束 #3，业务路径出现引擎判断
- *方案 B：直接 fork 一份 SQLite 方言 SQL 文件并存*。被拒绝：违反"同迭代单文件"原则，且双份 SQL 长期会漂移
- *方案 C：把转译能力嵌入 GoFrame `gdb` 驱动层*。被拒绝：超出本项目范围，需要修改上游框架；且运行时查询已经由 `gdb` 自动方言化，需要的只是 DDL 与初始化流程的转译

### 决策二：SQLite DDL 转译策略采用"行扫描 + token 替换 + 结构提取"

**选择**：转译器不构建完整的 SQL AST，而是基于现有 `cmd_sql_split.go`（已实现 SQL 多语句分割与字符串/注释感知扫描）的能力，对每条 CREATE TABLE / INSERT 语句做：

- 词法层：纯 token 级正则替换（反引号去除、`AUTO_INCREMENT` / `UNSIGNED` / `LONGTEXT` 等关键词替换、`INSERT IGNORE` 改写）
- 结构层：识别 CREATE TABLE 的列定义块，抽取内联的 `KEY` / `INDEX` / `UNIQUE KEY` 行作为独立的 `CREATE INDEX` 语句拼接到表创建语句之后
- 行扫描层：识别并整体删除 `CREATE DATABASE ... ;` 与 `USE ... ;` 语句

**理由**：

- 当前所有 MySQL 方言 DDL 是受控来源（项目自己写的），不会出现 fancy 语法（如 generated columns / partitioning / spatial index 等），简单的行扫描即可覆盖 100% 用例
- 单元测试以 14+7+7=28 个真实 SQL 文件作为 fixture，**端到端断言转译结果可在 SQLite 上成功执行**，避免转译器实现细节漂移
- 实现规模约 200~300 LOC + 测试，避免引入完整 SQL 解析器依赖

**替代方案与拒绝理由**：

- *方案 A：引入 `vitess/sqlparser` 等完整解析器*。被拒绝：依赖太重，且我们的 DDL 子集极小，性价比低
- *方案 B：手写 schema-as-code DSL，按方言生成 DDL*。被拒绝：要求重写所有 28 个 SQL 文件，违反"演示用途、最小成本"的迭代定位

### 决策三：MEMORY 表在 SQLite 下退化为普通表，不做特殊适配

**选择**：`sys_locker` / `sys_kv_cache` / `sys_online_session` 三张原 MEMORY 表在 SQLite 模式下由 DDL 转译器去除 `ENGINE=MEMORY` 子句，变为普通持久化表。**业务逻辑零变更**。

**理由**：

- 三张表的正确性**根本不依赖** "重启即清"语义：
  - `sys_locker` 已有 `expire_time` 字段，应用层 TTL 检查兜底；重启后陈旧锁会在首次 TTL 检查时被清理
  - `sys_kv_cache` 已有 `expire_at` 字段 + 应用层 TTL，且业务规则明确"缓存是有损的，不得作为权威数据源"
  - `sys_online_session` 已有 `last_active_time` + 现有 cron 定时任务清理，重启后短暂残留几秒陈旧会话不影响功能
- "重启即清"语义在 MEMORY 引擎中是性能优化（无需扫描过期条目），在 SQLite 演示场景中性能不重要
- 不需要业务层做"SQLite 模式启动时清空表"等适配，约束 #3 落地

**替代方案与拒绝理由**：

- *方案 A：SQLite 模式下用进程内 `sync.Map` / `lru.Cache` 实现 locker / kvcache / online_session*。被拒绝：违反约束 #3，业务模块需要感知后端差异；且需要新增三个 backend 实现与启动期切换逻辑
- *方案 B：SQLite 模式启动时清空这三张表以模拟"重启即清"*。被拒绝：用户已明确"业务模块不受数据库引擎差异影响"；且模拟成本高于让业务依赖 TTL 兜底（事实上业务已经依赖 TTL）

### 决策四：`kvcache` 后端重命名为 `sqltable` 并改写为方言中性 SQL

**选择**：

- 包路径 `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` → `apps/lina-core/internal/service/kvcache/internal/sqltable/`
- 常量 `BackendMySQLMemory` → `BackendSQLTable`，字符串值 `"mysql-memory"` → `"sql-table"`
- `Incr` 操作的实现：`gdb.Raw("LAST_INSERT_ID(value_int + N)")` + `Raw("SELECT LAST_INSERT_ID()")` 改为事务内 `UPDATE sys_kv_cache SET value_int = value_int + ? WHERE owner_type=? AND cache_key=?` + `SELECT value_int FROM sys_kv_cache WHERE owner_type=? AND cache_key=?`

**理由**：

- 包名 `mysql-memory` 已经误导：在 SQLite 下它依然工作但名字暴露 MySQL 实现细节，违背抽象层意图
- 改写后的实现**单一 backend 同时跑 MySQL 与 SQLite**，不需要为 SQLite 单独新增 backend 实现，符合约束 #3
- 项目编码规范明确禁用 `FIND_IN_SET` 等 MySQL 专用函数，`LAST_INSERT_ID(v + δ)` 实质上是同类问题，理应一并消除

**替代方案与拒绝理由**：

- *方案 A：保留 `mysql-memory` 名字，新增并列的 `sqlite` backend*。被拒绝：两个 backend 实现绝大部分代码重复；启动期分发逻辑徒增复杂度
- *方案 B：使用 `RETURNING` 子句*。被拒绝：MySQL 8.0.21+ 才支持 `RETURNING`，向下兼容性不如事务内 UPDATE+SELECT

### 决策五：SQLite 模式下集群锁定通过 `OnStartup` 钩子在启动 bootstrap 中实现

**选择**：

- `SQLiteDialect.OnStartup(ctx, configSvc)` 在启动期被调用一次
- 实现：调用 `configSvc.OverrideClusterEnabledForDialect(false)`（新增内存层覆盖方法），并 `logger.Warningf(ctx, ...)` 输出明确的警告日志
- `IsClusterEnabled` 在被覆盖后稳定返回 `false`，不需要在每个调用点感知方言

**警告日志示例**：

```
[WARNING] 当前为 SQLite 模式（database.default.link = sqlite::./temp/sqlite/linapro.db）
[WARNING] SQLite 模式仅支持单节点部署，cluster.enabled 已被强制覆盖为 false
[WARNING] 所有功能在单机模式下运行；切勿将 SQLite 模式用于生产环境
[WARNING] 如需多节点集群部署，请将 database.default.link 改回 MySQL 链接并重启
```

**理由**：

- 启动期钩子让"锁定 cluster"行为有单一明确执行点，不散落在 `cluster.Service` / `config.Service` / 各个调用方
- 在 `IsClusterEnabled` 内部覆盖而非每个调用点判断，保留所有现有 cluster 联动逻辑（前端 UI 隐藏、定时任务退化、缓存协调降级）的可复用性
- 警告日志使用 `WARNING` 级别（不是 `INFO`），在终端默认日志输出中醒目可见

### 决策六：默认 SQLite 数据库文件路径 `./temp/sqlite/linapro.db`，可通过 `link` 自定义

**选择**：

- 默认值（在 `config.template.yaml` 注释中给出示例）：`sqlite::./temp/sqlite/linapro.db`
- 用户可通过修改 `link` 字段自由更改路径，例如 `sqlite::/var/lib/myapp/db.sqlite`
- `SQLiteDialect.PrepareDatabase` 在执行前自动 `mkdir -p` 父目录；目录创建失败时返回明确错误（含路径与权限提示）
- `.gitignore` 增加 `temp/sqlite/*.db*` 条目，避免数据库文件被误提交

**理由**：

- `temp/` 目录在项目中已有"临时产物"语义（CLAUDE.md 约定 e2e 截图放此），SQLite 演示 / 测试数据库语义匹配
- 相对路径 `./temp/sqlite/` 解析基准是 `gfile.Pwd()`，与 `make init` / `make dev` 的工作目录一致
- 自动 `mkdir -p` 让首次启动无需用户手动创建目录，零依赖体验

### 决策七：仅通过 `config.yaml` 切换方言，不引入命令行参数

**选择**：

- `make init` / `make mock` / `make dev` 等命令保持现状，不增加 `dialect=sqlite` / `--dialect=...` 参数
- 唯一切换方式是修改 `config.yaml` 的 `database.default.link`
- `cmd init` / `cmd mock` 内部读取 `link` 后调用 `dialect.From(link)` 自动分发

**理由**：

- 单一来源（single source of truth）避免命令行参数与配置文件状态不一致带来的混乱
- 用户切换方言是低频动作（演示场景多半一选定终）；配置文件方式足够直观
- 实现路径少一个层（无需在 Makefile 与 cmd flag 解析中传递方言参数）

## Risks / Trade-offs

| 风险 | 缓解措施 |
|---|---|
| **DDL 转译器漏覆盖某个 MySQL 语法** → 转译结果在 SQLite 上执行失败 | 单元测试以现有 28 个 SQL 文件为 fixture，端到端断言转译结果可执行；新增 SQL 文件时强制走相同测试路径 |
| **新增插件 SQL 引入未覆盖的 MySQL 语法**（变更归档后才发现） | 插件 install pipeline 在转译失败时返回明确错误，定位到具体 SQL 文件与行号；同时在 spec `plugin-manifest-lifecycle` 中明确"插件 SQL 必须能被默认方言转译器处理"的约束 |
| **SQLite 模式下无 MEMORY 表的"重启即清"语义** → 重启后短暂出现陈旧锁 / 在线会话 / 缓存条目 | 三种数据均已有 TTL / 定时任务清理，应用层在首次 TTL 检查时自动清理；演示场景对短暂残留无敏感度 |
| **`Incr` 由 `LAST_INSERT_ID(v+δ)` 改为 UPDATE+SELECT** → 高并发下事务内两次往返略慢于原方案 | 仅 `kvcache` 一处使用；演示 / 单实例场景并发要求极低；MySQL 生产场景的两次 SQL 在事务内仍是单连接内顺序执行，性能差距可忽略 |
| **用户误将 SQLite 模式用于多实例部署** | 启动期 `WARNING` 日志 + cluster 强制锁定 + 文档明确警示三重防御 |
| **`temp/sqlite/` 目录被 CI 误清理导致测试失败** | `.gitignore` 仅忽略文件不忽略目录；CI 流程在 `make init` 之前不需要保留 db 文件，初始化会自动重建 |
| **现有插件 mock SQL 包含 `INSERT INTO ... ON DUPLICATE KEY UPDATE`** | 项目规则已禁止该语法，本次仅做兜底排查；若发现违规交付应当作 bug 在本变更中一并修复 |

## Migration Plan

本次变更对现有 MySQL 用户**完全透明**，无破坏性迁移：

- 现有 `config.yaml` 默认值仍是 MySQL 链接，行为与变更前完全一致
- 现有 SQL 文件、业务模块、DAO 层均保持原样
- 唯一可观察的变化：`make init` 内部多了一次方言分发（MySQL 路径下 `TranslateDDL` 为 no-op，无开销）

对希望切换到 SQLite 的用户，迁移路径为：

1. 将 `database.default.link` 改为 `sqlite::./temp/sqlite/linapro.db`
2. 执行 `make init` → 自动创建 `temp/sqlite/` 目录与数据库文件，加载所有 DDL
3. （可选）执行 `make mock` 加载演示数据
4. 启动 `make dev`，注意终端会输出 `WARNING` 提示

如需切回 MySQL：恢复原 `link` 值即可。两种模式数据完全独立，互不影响。

## Open Questions

无。前置探索阶段已澄清所有关键决策（SQLite 用户画像、不需要数据迁移、业务模块零适配、cluster 锁定 + 警告、文件路径、包重命名、仅走配置文件、不做安全网）。
