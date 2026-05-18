## Why

LinaPro 当前以 MySQL 为默认与单一 SQL 源方言、SQLite 作为开发演示方言。这种组合存在两个长期负担：

1. **MySQL 作为"源 SQL 语法"在工程上不是最优解**：MySQL 的 DDL 子集（反引号标识符、`ENGINE=`/`CHARSET=`/`COLLATE=` 子句、`UNSIGNED` 类型族、`AUTO_INCREMENT` 关键字、`INSERT IGNORE` 与 `ON DUPLICATE KEY UPDATE`、`ON UPDATE CURRENT_TIMESTAMP` 内联约束）与 ANSI SQL 偏离最大，导致 SQLite 翻译器需要 438 行正则改写并仍存在语义鸿沟（如 `ENGINE=MEMORY` 在 SQLite 上被剥离后丧失"重启即清"语义，已构成潜在 bug）。
2. **企业用户和外部部署偏好 PostgreSQL**：作为面向可持续交付的 AI 原生全栈框架，LinaPro 的默认数据库需要更贴近现代企业生产环境的偏好。PostgreSQL 是开源生态中事实上的现代数据库标准，具备更强的事务模型、数据完整性约束、JSONB 等原生能力，并且 SQL 语法更接近 ANSI 标准，使得"PG 作为单一源 → SQLite 作为开发兜底"的翻译路径更干净、规则更少、维护成本更低。

利用 CLAUDE.md 第一条规则（"全新项目，无历史遗留问题，不需要考虑兼容性"），本次变更对 MySQL 做彻底移除，把 PostgreSQL 设为唯一默认数据库与单一 SQL 源方言，SQLite 仍保留为开发/演示方言。

## What Changes

### 数据库支持范围
- **新增**：PostgreSQL 14+ 作为默认数据库与单一 SQL 源方言
- **保留**：SQLite 作为开发/演示方言（启动期警告"do not use in production"）
- **BREAKING**：完全移除 MySQL 支持（不再注册 `mysql:` link 前缀，删除 `pkg/dialect/internal/mysql/` 包，移除 GoFrame MySQL 驱动依赖）

### 方言抽象层（pkg/dialect）
- **新增** `pkg/dialect/internal/postgres/` 子包：包含 `dialect.go`、`translate.go`（PG 是源，no-op）、`error.go`（PG 错误码映射 `23505`/`40001`/`40P01`/`23502` 等）
- **删除** `pkg/dialect/internal/mysql/` 整个子包及其测试
- **重写** `pkg/dialect/internal/sqlite/translate.go`：从"MySQL → SQLite"改为"PG → SQLite"翻译
- **修改** `pkg/dialect.From()`：删除 `mysqlPrefix`，新增 `pgsqlPrefix = "pgsql:"`，保留 `sqlitePrefix`
- **新增** `Dialect.QueryTableMetadata(ctx, db, schema, names) ([]TableMeta, error)` 接口方法，用于替换硬编码的 `information_schema.TABLES` 查询

### SQL 源语法切换
- **BREAKING**：所有宿主 SQL（`apps/lina-core/manifest/sql/*.sql` 含 `mock-data/`）从 MySQL 语法改写为 PostgreSQL 语法
- **BREAKING**：所有包含实际 SQL 资源的源码插件（当前为 8 个插件：`content-notice`、`monitor-loginlog`、`monitor-online`、`monitor-operlog`、`monitor-server`、`org-center`、`plugin-demo-dynamic`、`plugin-demo-source`）的 `apps/lina-plugins/*/manifest/sql/` install/mock-data/uninstall SQL 从 MySQL 语法改写为 PostgreSQL 语法。`demo-control` 当前只有 mock-data 占位文件，无实际 SQL 改写项
- **统一规则**：使用 `INT/BIGINT GENERATED ALWAYS AS IDENTITY` 替代 `AUTO_INCREMENT`；`TIMESTAMP` 替代 `DATETIME`；删除 `ENGINE=`/`CHARSET=`/`COLLATE=`/`UNSIGNED`；表/列注释拆为独立 `COMMENT ON ...`；索引拆为独立 `CREATE INDEX`；`INSERT IGNORE INTO` 逐条审查后替换为 `INSERT INTO ... ON CONFLICT DO NOTHING`；删除 `ON UPDATE CURRENT_TIMESTAMP`
- **文本比较语义**：PostgreSQL 路径使用数据库默认 deterministic collation，不创建自定义 ICU 排序规则，也不在列定义中声明 `COLLATE linapro_ci`。业务文本键默认大小写敏感；如未来某个字段必须大小写不敏感，需单独通过规范化写入、表达式唯一索引、`citext` 或等价方案设计
- **幂等插入约束**：`ON CONFLICT DO NOTHING` 只有在目标表存在可触发冲突的 `PRIMARY KEY` / `UNIQUE` 约束时才等价于原 `INSERT IGNORE` 的重复执行跳过语义；seed 数据和具有稳定业务身份的 mock 数据必须具备覆盖记录稳定业务键的唯一约束或主键约束，再使用 `ON CONFLICT DO NOTHING` 保证重复执行幂等。日志/历史/监控类表的 mock 数据若仅为静态演示记录且业务本身不要求唯一身份，不得为了 mock 幂等强行新增影响生产写入语义的唯一约束；应按业务场景评估是否删除 mock、改为测试夹具内清理重载，或通过精确匹配静态演示行的存在性判断保证重复执行结果一致

### 易失性表（原 MEMORY 表）改造
- **BREAKING**：`sys_online_session`、`sys_locker`、`sys_kv_cache` 三张表从 MySQL `ENGINE=MEMORY` 改为 PG/SQLite 普通持久表
- 不在应用启动期执行 `TRUNCATE` 或其他全表清理操作，避免进程重启、滚动发布或 leader 切换时误清在线会话、全局锁与缓存数据
- 业务层依赖 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at` 与既有 TTL 清理路径让数据自然过期，不再依赖"引擎重启即清"语义
- SQLite 继续与 PG 保持一致：普通持久表 + TTL 自然过期

### 配置 / 工具链 / 镜像
- **BREAKING**：`config.yaml` / `config.template.yaml` / `hack/config.yaml` 默认 link 改为 `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`
- **BREAKING**：`go.mod` 移除 `github.com/go-sql-driver/mysql` 与 `gogf/gf/contrib/drivers/mysql/v2`，新增 `gogf/gf/contrib/drivers/pgsql/v2`
- **修改** 本地 PostgreSQL 启动说明：如仓库维护 compose 文件则移除 mysql 服务并新增 postgres:14+ 服务；GitHub Actions 不使用 docker-compose 启动 PG，而是参考 GoFrame workflow 通过 job `services.postgres` 直接声明 PostgreSQL service container
- **修改** `make image`：生产镜像不打包数据库（保持外部依赖架构），文档明示依赖外部 PG

### PG 特有的引导流程
- `PrepareDatabase`（PG 实现）：连接系统库 `postgres` → `pg_terminate_backend` 终止活跃连接 → `DROP DATABASE IF EXISTS` → `CREATE DATABASE linapro ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`；随后宿主 init SQL 在目标库内直接建表、建索引、写注释和写 seed 数据
- `make init rebuild=true` 在 PG 上能正确执行（不能在事务内 `DROP DATABASE`，必须连系统库）。`make init` 是运维初始化操作，使用配置中的数据库账号执行；该账号必须具备连接系统库、创建/删除目标数据库、终止目标库连接，以及在目标库内创建表、索引、注释并写入 seed 数据的足够权限。权限不足时命令快速失败并返回明确错误，运行时服务不提供低权限初始化兜底

### 业务代码影响
- PostgreSQL 不支持 MySQL 风格的 `UNSIGNED` 整数类型；所有由 MySQL `INT/BIGINT UNSIGNED` 数据列派生的 PG 字段变为 `INT/BIGINT`，`gf gen dao` 重新生成后对应实体 `uint64` → `int64`
- 全仓 grep `uint64` 仅用于识别影响面；只迁移由 MySQL `UNSIGNED` 数据列派生的 DAO / DO / Entity / API DTO / service ID 类型，不修改 wire / wasm / metrics / protobuf varint / byte count / uptime 等天然无符号语义

### 文档
- README / README.zh-CN（项目根、`apps/lina-core/`）：技术栈描述、启动前置说明、"切换到 SQLite"指南
- CLAUDE.md：技术栈段落 "GoFrame + MySQL + JWT" 改为 "GoFrame + PostgreSQL + JWT"

## Capabilities

### New Capabilities
- `sql-source-syntax`：定义 PostgreSQL 作为单一 SQL 源语法的子集约定，包括允许的语法、禁止的 PG 高级特性（避免 SQLite 翻译失败）、保留字处理、索引/注释拆分规则
- `volatile-table-bootstrap`：定义原 MEMORY 表（`sys_online_session`、`sys_locker`、`sys_kv_cache`）在 PG/SQLite 下改为普通持久表后的自然过期契约，包括不做启动期全表清空，依赖 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at` 进行 TTL 兜底、过期清理与锁抢占语义

### Modified Capabilities
- `database-dialect-abstraction`：从"MySQL/SQLite 双方言抽象"修改为"PostgreSQL/SQLite 双方言抽象"。删除 MySQL 方言相关需求与场景；新增 PostgreSQL 方言相关需求与场景（前缀分发、TranslateDDL no-op、PG 错误码分类、版本查询语句）；新增 `Dialect.QueryTableMetadata` 接口方法及其方言实现要求
- `database-bootstrap-commands`：修改 `PrepareDatabase` 方言分发场景（MySQL 替换为 PostgreSQL，PG 实现需要连系统库执行 DROP/CREATE）；修改 `TranslateDDL` 方言分发场景（MySQL no-op 替换为 PostgreSQL no-op；SQLite 翻译源从 MySQL 改为 PG）
- `cluster-deployment-mode`：修改方言枚举（MySQL → PostgreSQL）；保持"SQLite 启动期自动锁定 `cluster.enabled=false`"语义不变；保持"PG 链接下集群可启用"语义对应原 MySQL 的角色
- `project-setup`：修改"项目使用 SQLite 作为数据库"为"项目使用 PostgreSQL 作为默认数据库，SQLite 作为开发演示方言"；修改"SQL 语法 MUST 兼容 MySQL"为"SQL 语法 MUST 使用 PostgreSQL 14+ 子集，可被 SQLite 翻译执行"

## Impact

### 受影响的代码路径

**方言层**：
- `apps/lina-core/pkg/dialect/dialect.go`（前缀注册、From 工厂）
- `apps/lina-core/pkg/dialect/dialect_error.go`（错误码分类入口）
- `apps/lina-core/pkg/dialect/internal/mysql/`（**整目录删除**）
- `apps/lina-core/pkg/dialect/internal/postgres/`（**新建目录**：`dialect.go`、`translate.go`、`error.go`、`prepare.go`、`metadata.go` 及对应测试）
- `apps/lina-core/pkg/dialect/internal/sqlite/translate.go`（重写：从 MySQL → SQLite 改为 PG → SQLite）
- `apps/lina-core/pkg/dialect/internal/sqlite/dialect.go`（`OnStartup` 警告文案不变）
- `apps/lina-core/pkg/dialect/dialect_mysql_test.go`（**删除**）
- `apps/lina-core/pkg/dialect/dialect_postgres_test.go`（**新建**）

**SQL 源**（全量改写为 PG 语法）：
- `apps/lina-core/manifest/sql/001-project-init.sql`
- `apps/lina-core/manifest/sql/002-dict-type-data.sql`
- `apps/lina-core/manifest/sql/005-file-storage.sql`
- `apps/lina-core/manifest/sql/006-online-session.sql`
- `apps/lina-core/manifest/sql/007-config-management.sql`
- `apps/lina-core/manifest/sql/008-menu-role-management.sql`
- `apps/lina-core/manifest/sql/010-distributed-locker.sql`
- `apps/lina-core/manifest/sql/011-plugin-framework.sql`
- `apps/lina-core/manifest/sql/012-plugin-host-call.sql`
- `apps/lina-core/manifest/sql/013-dynamic-plugin-host-service-extension.sql`
- `apps/lina-core/manifest/sql/014-scheduled-job-management.sql`
- `apps/lina-core/manifest/sql/015-distributed-cache-consistency.sql`
- `apps/lina-core/manifest/sql/mock-data/*.sql`
- `apps/lina-plugins/*/manifest/sql/**/*.sql`（当前 8 个包含实际 SQL 资源的插件，共 22 个 SQL 文件；`demo-control` 仅有 `.gitkeep` 占位）

**配置 / 引导**：
- `apps/lina-core/manifest/config/config.yaml`、`config.template.yaml`
- `apps/lina-core/hack/config.yaml`（`gfcli.gen.dao.link`）
- `apps/lina-core/go.mod`、`go.sum`
- `apps/lina-core/main.go`（驱动 import 切换）
- `apps/lina-core/internal/cmd/cmd.go`、`cmd_init.go`、`cmd_init_database.go`、`cmd_http_runtime.go`（dialect 分发已封装，主要影响是间接的）

**业务代码（uint64 扫描点）**：
- `apps/lina-core/internal/service/jobmgmt/`（定时任务 `BIGINT UNSIGNED` IDs）
- `apps/lina-core/internal/service/locker/`（分布式锁 `INT UNSIGNED` IDs）
- `apps/lina-core/internal/service/kvcache/`（KV cache `BIGINT` IDs）
- `apps/lina-core/internal/service/config/`（config `BIGINT UNSIGNED` IDs）
- `apps/lina-core/internal/service/plugin/plugin_data_table_comment.go`（重构为 dialect.QueryTableMetadata 调用）
- 重新生成的 `apps/lina-core/internal/dao/`、`internal/model/do/`、`internal/model/entity/`
- 各 controller / api DTO 中的 ID 字段类型

**易失性表自然过期逻辑**：
- `apps/lina-core/internal/service/session/`、`internal/service/locker/`、`internal/service/kvcache/`（确认读取时过期判断、TTL 清理和锁过期抢占逻辑）
- `apps/lina-core/internal/cmd/` 与 `internal/service/cluster/`（确认启动路径不执行易失性表清空）

**容器 / 镜像 / 部署**：
- 本地 PostgreSQL 启动配置（如仓库维护 compose 文件则为 `docker-compose.yaml` 或 `hack/` 下等价文件）
- `apps/lina-core/Dockerfile`（如有 mysql client 依赖则移除）
- 任何 CI 工作流（GitHub Actions、本地测试 fixture）中起 MySQL 的步骤；GitHub Actions PG 使用 `services.postgres`，不通过 docker-compose 启动

**测试**：
- `hack/tests/fixtures/`（E2E 数据库 fixture 切换）
- `apps/lina-core/internal/service/**/*_test.go`（依赖 MySQL 行为的测试，如有）

**文档**：
- 项目根 `README.md` / `README.zh-CN.md`
- `apps/lina-core/README.md` / `README.zh-CN.md`
- `CLAUDE.md`

### API / 数据 / 依赖影响

- **HTTP API**：无变化（数据库类型不暴露到对外接口）
- **i18n**：有限影响。系统信息、framework 运行时文案和 apidoc 元数据中暴露的数据库名称/技术栈描述需要从 MySQL 更新为 PostgreSQL；不新增业务 UI 文案或新的翻译键体系，需同步维护相关运行时语言包与 apidoc i18n 资源
- **缓存一致性**：MEMORY 表改持久表后，不新增启动期清空或跨实例协调策略，也不引入新的缓存策略；`sys_online_session`/`sys_locker`/`sys_kv_cache` 分别依赖既有 `last_active_time`、`expire_time`、`expire_at` 判断、TTL 清理、锁过期抢占与 `kvcache` 跨实例一致性机制自然收敛
- **数据迁移**：**不提供数据迁移工具**。按 CLAUDE.md 第一条规则，全新项目重建数据库即可（`make init confirm=init rebuild=true`）
- **回退**：`git revert` 即可，无数据风险（数据库重建后，旧 MySQL 数据库实例可保留作为审计快照）

### 风险与缓解

- **首次启动 UX**：PG 必须先就绪，否则 `make init` 连接错误。缓解：README 显式说明先准备 PostgreSQL（本地可用 `docker run` 或 compose 文件，如存在）再 `make init`，连接错误信息友好化
- **PG 保留字（`key`/`value`/`type` 等列名）**：依赖 GoFrame ORM 自动加双引号；如不可行，SQL 源中显式加双引号；不重命名列以避免业务代码大面积改动
- **`gf gen dao` 类型变化**：由 MySQL `UNSIGNED` 数据列派生的 `uint64` → `int64` 通过全仓扫描、来源判定、集中替换和编译验证保证不漏；非数据库派生的 `uint64` 保持不变
- **构建期嵌入资源**：`prepare-packed-assets.sh` 自动重新打包 SQL 嵌入资源，无需特殊处理

### 不在本次变更范围

- **不引入 PG 特性增量**：JSONB、数组类型、CTE 物化、并行查询等 PG 高级特性本次不在 SQL 源中使用，保持纯 ANSI 子集，确保 SQLite 翻译可行；如未来需要，单独立变更
- **不调整时区策略**：仍使用 `TIMESTAMP`（不带时区），与现 `DATETIME` 语义对齐；TIMESTAMPTZ 不在本次范围
- **不调整 schema 隔离**：所有表在 `public` schema，多 schema 隔离不在本次范围
- **不双跑 E2E**：CI 主跑 PG 路径（生产路径），SQLite 仅覆盖单元/翻译测试；SQLite 完整 E2E 不纳入持续集成
