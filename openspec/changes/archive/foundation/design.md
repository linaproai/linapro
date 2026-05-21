## Context

LinaPro 基础设施层存在三个相互独立但共同影响开发效率的问题：

1. **数据库引擎**：以 MySQL 为默认 SQL 源方言，与 ANSI SQL 偏离最大，SQLite 翻译器维护成本高；企业用户偏好 PostgreSQL。
2. **启动效率**：启动后 10 秒内约 98 条 SQL，前 4 秒约 57 条属于启动链路，主要来自默认 SQL debug 配置、插件快照重复构造、no-op 空事务和写后回读。
3. **跨平台开发**：开发入口依赖 GNU Make 和 POSIX Shell，Windows 用户无法顺利执行常用命令。

项目是全新项目，没有历史遗留问题和技术债务，设计方案不需要考虑兼容性。

## Goals / Non-Goals

**Goals:**

- 彻底移除 MySQL 支持，建立 PostgreSQL 为单一 SQL 源方言，保留 SQLite 为开发演示方言
- 重写 SQLite 翻译器为"PG → SQLite"翻译路径
- 原 MEMORY 表改造为持久表 + 自然过期语义
- 默认启动日志不输出 SQL 明细，减少启动期重复查询和空事务
- 建立启动 SQL 统计口径和自动化回归测试
- 提供不依赖 GNU Make 的跨平台开发命令入口
- 保持 Linux/macOS 用户现有 `make` 工作流可用

**Non-Goals:**

- 不引入 PG 高级特性（JSONB、数组类型、CTE 物化、并行查询等）
- 不调整时区策略，不引入多 schema 隔离
- 不修改 GoFrame ORM 元数据探测机制
- 不改变插件安装、启用、禁用、卸载的用户可见语义
- 不重写 GoFrame CLI、Docker、kubectl、pnpm、Playwright 或 Go toolchain
- 不改变后端运行时 API、数据库结构、权限模型、插件运行时契约或前端用户界面

## Decisions

### 1. Database Engine Migration

#### D1: PostgreSQL 作为单一 SQL 源语法（替代 MySQL）

**决策**：把 PG 设为唯一 SQL 源方言，所有 `manifest/sql/*.sql` 用 PG 14+ 子集编写。SQLite 通过翻译器从 PG 改写为 SQLite 兼容 SQL。

**为什么**：
- PG 比 MySQL 更接近 ANSI SQL，PG → SQLite 的翻译规则比 MySQL → SQLite 少
- PG 是默认方言，源 SQL = 默认方言确保生产路径无翻译开销
- PG 的标识符引号策略与代码风格天然契合
- 业务文本值的比较与唯一性采用 PG 默认 deterministic collation，保持大小写敏感

**备选方案**：保留 MySQL 作为源（与移除 MySQL 矛盾）；使用 SQL92 中性语法（增加抽象层）；使用 SQLite 作为源（太宽松）

#### D2: MEMORY 表改造为普通持久表 + 自然过期

**决策**：`sys_online_session`、`sys_locker`、`sys_kv_cache` 三张表在 PG 和 SQLite 上都是普通持久表。启动、重启、滚动发布和 leader 切换时不执行 `TRUNCATE` 或全表清理；数据由业务层 `last_active_time`、`expire_time`、`expire_at` 与 TTL 清理任务自然过期。

**为什么**：
- PG 没有真正等价 MySQL MEMORY 的引擎
- 改造后所有方言行为一致
- 启动期清空会在进程重启、滚动发布、leader 重新选举时误踢在线会话、删除有效锁或清空仍可用缓存

**备选方案**：PG 用 UNLOGGED TABLE（不在重启时清空，语义错位）；PG 用 TEMP TABLE（会话级，无法跨连接共享）；启动期 TRUNCATE（误删未过期数据）

#### D3: INSERT IGNORE → INSERT ... ON CONFLICT DO NOTHING

**决策**：所有 `INSERT IGNORE INTO` 逐条审查后改写。seed 数据和具有稳定业务身份的 mock 数据必须存在覆盖稳定业务键的约束，再改为 `INSERT INTO ... ON CONFLICT DO NOTHING`。日志/历史/监控类 mock 数据不得为了幂等强行新增唯一约束，应通过精确存在性判断保证重复执行结果一致。

**为什么**：
- PG 14+ 和 SQLite 3.24+ 都支持 `ON CONFLICT DO NOTHING`
- 不写显式冲突列等同于"任意 unique 约束冲突都跳过"
- `ON CONFLICT DO NOTHING` 不是无条件去重；目标表只有未写入的自增主键时重复执行会插入新行

#### D4: ON UPDATE CURRENT_TIMESTAMP 全部移除

**决策**：从所有 SQL 源中删除 `ON UPDATE CURRENT_TIMESTAMP` 子句。`updated_at` 的实时更新由 GoFrame DAO 层自动维护。

**为什么**：PG 没有等价的内联子句；SQLite 也没有等价；GoFrame 已明确自动维护时间字段。

#### D5: 保留字列名处理策略

**决策**：依赖 GoFrame ORM 自动加双引号；SQL 源 DDL 中对保留字列名使用双引号包裹。不重命名列。

**为什么**：重命名列会扩散到 entity / DTO / 业务代码 / 前端，改动面巨大。GoFrame PG 驱动实测通过保留字列写入和查询。

#### D6: PG 的 PrepareDatabase 通过系统库执行 DROP/CREATE

**决策**：PG `PrepareDatabase` 连接到 `postgres` 系统库，执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`。`make init` 使用配置中的数据库账号，权限不足时快速失败。

**为什么**：PG 不能在事务内 DROP DATABASE，不能在有活跃连接时 DROP。运维初始化失败应显式暴露权限或环境问题。

#### D7: Dialect.QueryTableMetadata 接口抽象

**决策**：`Dialect` 接口新增 `QueryTableMetadata(ctx, db, schema, names) ([]TableMeta, error)` 方法。PG 实现使用 `information_schema.tables` JOIN `pg_class` 的 `obj_description(oid)`；SQLite 实现从 `sqlite_master` 查询，注释固定空字符串。

**为什么**：当前硬编码 `information_schema.TABLES` 是 MySQL 风格，PG 上语义不同。SQLite 完全不支持表注释。抽到 dialect 接口后业务代码不再感知方言差异。

#### D8: 数据库派生 ID 类型 uint64 → int64

**决策**：所有由 MySQL `INT/BIGINT UNSIGNED` 数据列派生的字段重新生成 entity 后变为 `int64`。不修改非数据库派生的 `uint64`（wire / wasm / metrics / protobuf 等）。

**为什么**：PG 不支持 UNSIGNED 类型。ID 类字段实际值远未达到 `int64` 上限。

#### D9: CI 主跑 PG，SQLite 仅覆盖单元/翻译测试

**决策**：E2E 测试默认连接 PG（生产路径），不双跑 SQLite。SQLite 仅覆盖翻译器单测、宿主启动 smoke 和开发演示。

#### D10: 本地 PG 启动与 CI PG service container

**决策**：`make dev` 不负责启动或管理数据库。开发者自行准备 PG。GitHub Actions 使用 `services.postgres` 声明 PG service container。

#### D11: i18n 影响评估

有限影响：系统信息、framework 运行时文案和 apidoc 元数据中的数据库名称/技术栈描述需要从 MySQL 更新为 PostgreSQL。

#### D12: 缓存一致性评估

不引入新的缓存策略。MEMORY → 持久表后不做启动期全表清空，沿用既有 `cluster.Service` 抽象和 TTL 清理路径。

#### D13: PostgreSQL 文本比较采用默认 deterministic collation

**决策**：不创建自定义 ICU 排序规则，所有文本列使用数据库创建时的 deterministic collation（`LC_COLLATE 'C' LC_CTYPE 'C'`）。业务文本键默认大小写敏感。

**为什么**：自定义 ICU 非确定性排序规则会引入初始化权限要求、ICU 版本升级后的维护点和比较成本。大小写不敏感需求应按字段精确建模。

### 2. Startup SQL Efficiency

#### 决策一: 把 SQL 明细日志和真实 SQL 优化拆开处理

**决策**：交付型默认配置统一为 `database.default.debug=false`，保留注释说明如何临时开启。

**理由**：启动日志刷屏首先是配置问题，关闭 debug 不影响真实行为。SQL 明细日志属于诊断工具，不适合作为默认开发体验。

**替代方案**：保留 `debug=true`，仅过滤启动阶段 SQL 日志（需要侵入 GoFrame 日志处理链，收益低）

#### 决策二: 引入一次启动链路内共享的 StartupContext

**决策**：在 HTTP 启动编排中创建一次启动上下文，包含 catalog/integration/job 启动快照和可选的启动统计采集器。`BootstrapAutoEnable`、插件 HTTP 路由注册、runtime frontend prewarm、cron builtin sync 等启动阶段复用该上下文。

**理由**：当前已有 snapshot 设计，问题在于生命周期太短、每个阶段重新构造。将快照作用域提升到一次启动编排内，可以减少重复全表读取。

**替代方案**：把 snapshot 做成进程级缓存（会引入跨请求一致性问题）

#### 决策三: 插件同步 no-op 时不得进入事务或写后回读

**决策**：插件 manifest 同步拆成两步：计算期望投影 → 与启动快照比较；只有存在差异时才进入事务。菜单同步新增 `PluginMenusMatch` 或等价比较能力。

**理由**：日志中多个空 BEGIN/COMMIT 来自无变化同步路径。插件数量增长时，空事务按插件数线性增长。

**替代方案**：保留事务但调低事务日志级别（只降低可见噪音，未减少数据库操作）

#### 决策四: 写后优先更新启动快照，只有必要时回读数据库

**决策**：插入路径优先使用 `InsertAndGetId` 构造 entity 并写入 snapshot；更新路径使用 `existing + data` 合成最新 entity 并写入 snapshot；只有依赖数据库默认值或复杂触发结果时才回读。

**理由**：当前写后回读会制造额外查询。启动同步使用的字段大多来自 manifest 或 DO projection，本地可确定。

#### 决策五: 内置定时任务注册只使用声明派生快照

**决策**：内置任务注册使用 `SyncBuiltinJobs` 返回的 projection 直接调用 `RegisterJobSnapshot`，持久化扫描只加载用户创建任务和非内置启用任务。

**理由**：内置任务执行定义权威来源是源码声明，`sys_job` 行只是治理投影。避免启动期既 upsert 内置行又从 `sys_job` 读取同一批内置行。

#### 决策六: 启动 SQL 统计用于回归门禁，不追求绝对零 SQL

**决策**：测试不断言精确 SQL 条数，而是断言：默认配置不输出 SQL 明细、插件同步 no-op 不写库、启动共享快照构造次数在预算内、启动摘要日志包含阶段耗时和差异统计。

**理由**：GoFrame 版本、方言和测试环境会影响元数据探测条数。精确 SQL 条数测试容易脆弱。

### 3. Cross-Platform Dev Commands

#### 1. 使用 Go CLI 作为跨平台主入口

**决策**：新增或扩展 `hack/tools/linactl`，提供统一子命令。Go CLI 负责跨平台路径处理、文件复制、进程启动、HTTP readiness、端口检测、日志文件、子命令执行和错误输出。

**选择 Go 的原因**：项目后端与现有工具链已以 Go 为基础；Go 标准库具备跨平台能力；Windows 用户在后端开发前已经需要 Go；现有 `hack/tools/*` 已采用 Go 工具模式。

**备选方案**：Node.js CLI（会让后端任务依赖 Node 环境）；Taskfile/just/Mage（引入新的全局工具安装要求）

#### 2. make.cmd 只作为 Windows 薄包装

**决策**：根目录提供 `make.cmd`，内容只转发参数到 Go CLI。`cmd.exe` 用户可执行 `make dev`，PowerShell 用户使用 `.\make dev`。`make.cmd` 透传所有参数。不默认新增 `make.ps1`。

#### 3. Makefile 保留为兼容层

**决策**：现有 `make <target>` 入口不立即移除。目标实现逐步瘦身为调用 Go CLI。迁移顺序按风险和收益分批：低风险工具目标 → 开发服务目标 → 构建目标 → 验证目标 → 子模块目标。

#### 4. 兼容 make 风格参数

**决策**：Go CLI 支持现有 `key=value` 参数（如 `init confirm=init rebuild=true`）。CLI 内部将 `key=value` 归一化为选项结构。

#### 5. 测试策略以工具行为为中心

**决策**：不涉及用户可观察页面，不需要 E2E 测试。优先使用 Go 单元测试和命令级 smoke 测试覆盖参数解析、文件复制、插件扫描、服务状态检测和 Makefile 薄包装一致性。

#### 6. GitHub Actions 必须覆盖 Windows 基本命令

**决策**：增加 Windows runner 验证，至少覆盖 `go run ./hack/tools/linactl help`、`go run ./hack/tools/linactl status`、`.\make help`、`.\make status` 等轻量命令。

## SQL 改写规则与示例

### 类型映射表

| MySQL 源 | PG 14+ 目标 | SQLite 翻译目标 |
|---|---|---|
| `INT PRIMARY KEY AUTO_INCREMENT` | `INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `BIGINT UNSIGNED NOT NULL AUTO_INCREMENT` | `BIGINT GENERATED ALWAYS AS IDENTITY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `BIGINT UNSIGNED` | `BIGINT` | `INTEGER` |
| `INT UNSIGNED` | `INT` | `INTEGER` |
| `TINYINT` | `SMALLINT` | `INTEGER` |
| `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` |
| `CHAR(n)` | `CHAR(n)` | `TEXT` |
| `TEXT` / `LONGTEXT` / `MEDIUMTEXT` | `TEXT` | `TEXT` |
| `BLOB` / `LONGBLOB` / `MEDIUMBLOB` | `BYTEA` | `BLOB` |
| `DATETIME` | `TIMESTAMP` | `DATETIME` |
| `DECIMAL(m,n)` | `DECIMAL(m,n)` | `NUMERIC` |
| `FLOAT` / `DOUBLE` | `REAL` / `DOUBLE PRECISION` | `REAL` |

### 子句移除规则

| MySQL 源子句 | PG 处理 | SQLite 处理 |
|---|---|---|
| `ENGINE=InnoDB` / `ENGINE=MEMORY` | 删除 | 删除 |
| `DEFAULT CHARSET=utf8mb4` | 删除 | 删除 |
| `COLLATE=utf8mb4_general_ci` | 删除 | 删除 |
| `UNSIGNED` | 删除 | 删除 |
| `ON UPDATE CURRENT_TIMESTAMP` | 删除 | 删除 |
| 反引号 `` `id` `` | 删除或 `"id"` | 删除 |

### 注释处理

MySQL 内联 `COMMENT '...'` 拆为独立 `COMMENT ON TABLE/COLUMN ... IS '...'`。SQLite 翻译器丢弃 `COMMENT ON` 语句并输出 Debug 日志。

### 索引处理

MySQL 内联 `KEY`/`UNIQUE KEY` 拆为独立 `CREATE INDEX`/`CREATE UNIQUE INDEX`。索引命名约定：`idx_{表名}_{列名}` 或 `uk_{表名}_{列名}`。

### INSERT IGNORE 改写

```sql
-- MySQL 源
INSERT IGNORE INTO sys_user (username, password) VALUES ('admin', '$2a$10$...');

-- PG 改写
INSERT INTO sys_user (username, password) VALUES ('admin', '$2a$10$...')
ON CONFLICT DO NOTHING;
```

### 保留字列名处理

PG DDL 中对保留字列名使用双引号包裹：

```sql
CREATE TABLE sys_config (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "key" VARCHAR(128) NOT NULL,
    "value" TEXT
);
```

### MEMORY 表改造示例

```sql
-- MySQL 源
CREATE TABLE IF NOT EXISTS sys_locker (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    holder VARCHAR(64) DEFAULT '',
    expire_time DATETIME NOT NULL,
    UNIQUE KEY uk_name (name)
) ENGINE=MEMORY;

-- PG 改写
CREATE TABLE IF NOT EXISTS sys_locker (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    holder VARCHAR(64) NOT NULL DEFAULT '',
    expire_time TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX uk_sys_locker_name ON sys_locker (name);
```

## 错误码映射

`pkg/dialect/internal/postgres/error.go` 关键定义：

```go
const (
    pgErrUniqueViolation       = "23505"
    pgErrSerializationFailure  = "40001"
    pgErrDeadlockDetected      = "40P01"
    pgErrLockNotAvailable      = "55P03"
    pgErrCheckViolation        = "23514"
    pgErrForeignKeyViolation   = "23503"
    pgErrNotNullViolation      = "23502"
)
```

SQLite 错误码映射保持不变（`SQLITE_BUSY` / `SQLITE_LOCKED` 视为可重试）。

## Risks / Trade-offs

### Database Engine Migration Risks

- **[R1] GoFrame PG 驱动行为差异**：实测通过保留字列、IDENTITY 列回填、连接字符串格式。保留受控集成测试作为升级回归保护。
- **[R2] 连接字符串格式**：`pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable` 已验证。不需要 MySQL 路径中的 `parseTime=true`、`loc=Local` 等参数。
- **[R3] 保留字列名实测**：`key`/`value` 最小表写入和查询已通过真实 PG 集成测试。
- **[R4] MEMORY 表改造后的过期数据收敛**：不在启动期清空数据，单元测试覆盖过期判断与清理路径，多进程集群模拟测试覆盖 leader 选举和锁过期抢占。
- **[R5] PG 14+ 可用性**：PG 14 是 2021 年发布，主流云厂商均支持。
- **[R6] 嵌入资源构建脚本兼容性**：`prepare-packed-assets.sh` 不做语法预检查，运行时翻译器负责验证。
- **[R7] uint64 替换回归风险**：全仓扫描后逐个分析使用场景，仅迁移 MySQL UNSIGNED 数据列派生类型。
- **[R8] PG 容器启动慢**：GitHub Actions 使用 `services.postgres` + `pg_isready` healthcheck。
- **[R9] 低权限数据库账号**：`make init` 使用配置中的账号，权限不足时快速失败。
- **[R10] INSERT IGNORE 机械替换**：为每个目标表列出幂等依据，日志/历史类 mock 表不得强造唯一约束。
- **[R11] 默认大小写敏感语义**：明确记录为产品语义，未来需要大小写不敏感时单独设计。
- **[R12] 本地 Docker 守护进程不可用**：镜像验证在 CI 或 Docker 可用的开发机上补充。

### Startup SQL Efficiency Risks

- **[Risk] 启动共享快照过期**：快照仅在同一启动编排内使用；所有写入路径必须同步更新快照。
- **[Risk] no-op 比较遗漏字段**：比较函数必须覆盖持久化字段；测试覆盖四类差异。
- **[Risk] 关闭默认 SQL debug 后排查不便**：配置注释保留开启方式；启动摘要日志保留关键阶段耗时。
- **[Risk] 统计测试在不同环境下波动**：测试约束项目可控行为，不断言框架元数据探测的绝对条数。
- **[Risk] Cron 首轮 monitor 任务被误判为启动 SQL**：统计口径区分 host startup phase 和 first scheduled job phase。

### Cross-Platform Dev Commands Risks

- **[Risk] 一次性迁移导致回归面过大**：分批迁移，先低风险目标。
- **[Risk] Go CLI 与 Makefile 行为分叉**：Makefile 和 make.cmd 只做薄包装，真实逻辑只保留在 Go CLI。
- **[Risk] Windows 进程树停止语义不同**：Go CLI 使用 PID 文件、端口检测和子进程启动句柄组合处理。
- **[Risk] go run 每次编译影响启动速度**：初期接受该成本，后续可提供 `go install` 或缓存二进制入口。
- **[Risk] make.cmd 与 GNU Make 命名冲突**：仅在 Windows 项目根目录作为本地脚本使用。
- **[Risk] 文档双入口造成困惑**：文档按"跨平台推荐入口"和"兼容入口"分组。
- **[Risk] Windows CI 过慢**：首期只验证基本命令、入口转发和轻量 smoke。

## Migration Plan

1. **方言层骨架**：删除 `internal/mysql/`，新建 `internal/postgres/`，扩展 `Dialect` 接口，更新 `From()` 注册
2. **GoFrame PG 驱动 spike**：最小 main.go 验证连接、IDENTITY 列、保留字列名
3. **SQL 源改写**：宿主与插件 SQL 全量改写为 PG 语法
4. **MEMORY 表改造**：DDL 改持久表 + 确认启动路径不执行清空
5. **SQLite 翻译器重写**：替换输入语法基准
6. **配置切换**：config.yaml / hack/config.yaml / go.mod 切到 PG
7. **启动 SQL 优化**：调整默认 debug 配置、引入 StartupContext、实现 no-op fast path、优化 cron 注册
8. **跨平台工具**：新增 linactl Go CLI、make.cmd、Makefile 薄包装化
9. **测试与验证**：单元测试 + 集成测试 + E2E + 启动 smoke + Windows CI
10. **文档同步**：README / README.zh-CN / CLAUDE.md

回退策略：`git revert` 即可，无数据风险。
