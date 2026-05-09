# 实施任务清单

> 实施前请先阅读：
> - `proposal.md` 了解整体动机与影响面
> - `design.md` 了解关键决策、SQL 改写规则、伪代码骨架
> - `specs/sql-source-syntax/spec.md` 了解 PG 源 SQL 子集约定
> - `specs/volatile-table-bootstrap/spec.md` 了解易失性表自然过期契约
> - `specs/database-dialect-abstraction/spec.md` 与 `specs/database-bootstrap-commands/spec.md` 了解方言层接口
> - CLAUDE.md 中关于 i18n 治理、缓存一致性治理、bugfix 测试要求等规范

## 1. 实施前的 spike 验证（高优先，必须最先做）

- [x] 1.1 写最小 main.go 验证 GoFrame `gogf/gf/contrib/drivers/pgsql/v2` 驱动可连通：使用 link `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/postgres?sslmode=disable` 连接 PG 14+ 实例，执行 `SELECT version()` 成功返回。结果记录到 design.md Q1/Q2，确认最终 link 格式；本机 Docker 阻塞时以 Postgres.app PG18 完成实测，PG14 覆盖由 GitHub Actions `postgres:14-alpine` service 与同一受控集成测试入口保证
- [x] 1.2 在 PG 上建测试表 `CREATE TABLE test_kv ("key" VARCHAR(64) NOT NULL, "value" TEXT, PRIMARY KEY("key"))`，用 `gf gen dao` 生成 entity，再用 GoFrame ORM 写入并查询保留字列。验证 GoFrame 是否自动加双引号；记录结果到 design.md D5/R3
- [x] 1.3 在 PG 上建带 IDENTITY 列的测试表 `CREATE TABLE test_id (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, name VARCHAR(64))`，用 GoFrame `InsertAndGetId(...)` 写入并验证返回的新 ID 正确。如失败，记录回退方案（手动 `RETURNING id` 查询）
- [x] 1.4 验证 GoFrame PG 驱动对 `parseTime`/`loc` 等时间相关参数的需求；如不需要，确认 link 字符串简化为 `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`
- [x] 1.5 检查 `hack/scripts/prepare-packed-assets.sh` 当前是否对 SQL 文件做 SQLite 语法预检查；如有，调整为只做文件复制和列表生成，不做语法预检查
- [x] 1.6 早期 spike 曾验证 PostgreSQL ICU 非确定性排序规则可执行；该方案后续已由 FB-14 取代，最终设计改为不创建自定义排序规则，使用 PostgreSQL 默认 deterministic collation

## 2. 方言层骨架改造（pkg/dialect）

- [x] 2.1 删除 `apps/lina-core/pkg/dialect/internal/mysql/` 整个目录及其测试文件（`dialect.go`、`error.go`、`dialect_test.go`）
- [x] 2.2 删除 `apps/lina-core/pkg/dialect/dialect_mysql_test.go`
- [x] 2.3 新建 `apps/lina-core/pkg/dialect/internal/postgres/` 目录
- [x] 2.4 新建 `apps/lina-core/pkg/dialect/internal/postgres/dialect.go`：实现 `Name()`（返回 `"postgres"`）、`SupportsCluster()`（返回 `true`）、`OnStartup()`（no-op）、`DatabaseVersion()`（用 `SELECT version()`）；保留 `TranslateDDL` 为 no-op（PG 是源方言，直接返回原 ddl）
- [x] 2.5 新建 `apps/lina-core/pkg/dialect/internal/postgres/error.go`：定义 PG 错误码常量（`23505`/`40001`/`40P01`/`55P03`/`23514`/`23503`/`23502`）；实现 `IsRetryableWriteConflict(err)` 识别 `40001`/`40P01`/`55P03`
- [x] 2.6 新建 `apps/lina-core/pkg/dialect/internal/postgres/prepare.go`：实现 `PrepareDatabase(ctx, link, rebuild)`，包含连接系统库、`pg_terminate_backend` 终止活跃连接、`DROP DATABASE IF EXISTS`、`CREATE DATABASE ENCODING 'UTF8' LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0`（伪代码骨架见 design.md D6）
- [x] 2.7 新建 `apps/lina-core/pkg/dialect/internal/postgres/metadata.go`：实现 `QueryTableMetadata(ctx, db, schema, names)`，使用 `information_schema.tables` JOIN `pg_class` 通过 `obj_description(c.oid)` 查询表注释（伪代码骨架见 design.md D7）
- [x] 2.8 新建 `apps/lina-core/pkg/dialect/internal/postgres/dialect_test.go`、`error_test.go`、`prepare_test.go`、`metadata_test.go`：单元测试覆盖 Name 返回固定为 `"postgres"`、错误码识别、准备数据库流程、元数据查询
- [x] 2.9 修改 `apps/lina-core/pkg/dialect/dialect.go`：删除 `mysqlPrefix = "mysql:"` 与对应分支，新增 `pgsqlPrefix = "pgsql:"` 与 `postgresDialect` 包装结构体；修改 `From()` 工厂函数；当 `mysql:` 前缀传入时返回明确的"不再支持"错误（含支持前缀列表）
- [x] 2.10 在 `Dialect` 接口中新增 `QueryTableMetadata(ctx, db, schema, names) ([]TableMeta, error)` 方法签名，定义 `TableMeta struct { TableName string; TableComment string }` 类型
- [x] 2.11 在 `pkg/dialect/internal/sqlite/dialect.go` 中实现 `QueryTableMetadata`：从 `sqlite_master` 查询表名，注释字段固定空字符串
- [x] 2.12 新建 `apps/lina-core/pkg/dialect/dialect_postgres_test.go`：测试 `pgsql:` 前缀分发、`mysql:` 前缀拒绝错误、`SupportsCluster()=true` 行为

## 3. SQLite 翻译器重写（PG → SQLite）

- [x] 3.1 重写 `apps/lina-core/pkg/dialect/internal/sqlite/translate.go`：用 PG 14+ 语法作为输入基准，按 design.md §SQLite 翻译器规则速记 实现翻译规则
- [x] 3.2 实现 `GENERATED ALWAYS AS IDENTITY` 主键改写为 `INTEGER PRIMARY KEY AUTOINCREMENT`（注意：SQLite 自增列必须是 INTEGER，不能是 BIGINT）
- [x] 3.3 实现整数类型映射：`INT/BIGINT/SMALLINT` → `INTEGER`
- [x] 3.4 实现字符串类型映射：`VARCHAR(n)/CHAR(n)/TEXT` → `TEXT`
- [x] 3.5 实现 `BYTEA` → `BLOB` 映射
- [x] 3.6 实现 `TIMESTAMP` → `DATETIME` 映射（保持 GoFrame 时间序列化兼容）
- [x] 3.7 实现 `DECIMAL(m,n)/NUMERIC(m,n)` → `NUMERIC` 映射
- [x] 3.8 实现 `COMMENT ON TABLE/COLUMN ...` 整句丢弃规则（正则匹配 `^\s*COMMENT\s+ON\b` 开头的语句）
- [x] 3.9 早期实现过自定义排序规则的 SQLite 转译；该方案后续已由 FB-14 取代，最终 SQLite 翻译器不再模拟自定义排序规则，遇到 `CREATE COLLATION` 会快速失败
- [x] 3.10 确认 SQLite 翻译器不需要支持 `TRUNCATE`：SQLite 原生不支持 `TRUNCATE TABLE`，本次 SQL 源与应用启动路径不得依赖该语法；如未来需要显式清空，应单独通过方言能力设计 `DELETE FROM x; DELETE FROM sqlite_sequence WHERE name='x';` 等价方案
- [x] 3.11 保留 `INSERT ... ON CONFLICT DO NOTHING`（SQLite 3.24+ 兼容）、`CREATE INDEX/UNIQUE INDEX [IF NOT EXISTS] ...`、双引号包裹的标识符
- [x] 3.12 实现快速失败逻辑：识别并拒绝 PG 高级特性（`JSONB`、`CREATE TRIGGER`、`CREATE FUNCTION`、`SERIAL`、`MERGE`、`WITH RECURSIVE`、`GENERATED ALWAYS AS (expr) STORED` 等），返回带 sourceName/行号/关键字的明确错误
- [x] 3.13 重写 `apps/lina-core/pkg/dialect/internal/sqlite/translate_test.go` 与 `dialect_sqlite_translate_test.go`：覆盖每条翻译规则的正反用例、默认文本列不追加 `COLLATE NOCASE` 的用例 + 项目实际 SQL 文件全文翻译用例
- [x] 3.14 删除原"MySQL → SQLite"翻译相关的过时正则与测试用例（如 `reCreateDatabase`、`reUseDatabase`、`reInsertIgnore`、`reIntegerAuto`、`reUnsupportedKeywords` 中针对 MySQL 关键字的部分）

## 4. 宿主 SQL 源改写为 PG 语法

- [x] 4.1 改写 `apps/lina-core/manifest/sql/001-project-init.sql`：删除 `CREATE DATABASE` 与 `USE` 语句；不创建自定义排序规则；`sys_user` 表使用 `INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`、`SMALLINT` 替代 `TINYINT`、`TIMESTAMP` 替代 `DATETIME`；文本列使用 PostgreSQL 默认 deterministic collation；删除 `ENGINE=`/`CHARSET=`/`COLLATE=` MySQL 子句；表/列注释拆为独立 `COMMENT ON ...`；索引拆为独立 `CREATE INDEX/UNIQUE INDEX`；`INSERT IGNORE` 按幂等依据改为 `INSERT INTO ... ON CONFLICT DO NOTHING`
- [x] 4.2 改写 `apps/lina-core/manifest/sql/002-dict-type-data.sql`：同 4.1 规则；注意 `sys_dict_type.type` 与 `sys_dict_data.value` 列名是 PG 保留字，DDL 中用 `"type"` 与 `"value"` 双引号包裹
- [x] 4.3 改写 `apps/lina-core/manifest/sql/005-file-storage.sql`：同 4.1 规则
- [x] 4.4 改写 `apps/lina-core/manifest/sql/006-online-session.sql`：删除 `ENGINE=MEMORY`，改为普通持久表（参见任务组 5）
- [x] 4.5 改写 `apps/lina-core/manifest/sql/007-config-management.sql`：同 4.1 规则；`sys_config.key`、`sys_config.value` 是 PG 保留字，DDL 与 DML 都要用双引号
- [x] 4.6 改写 `apps/lina-core/manifest/sql/008-menu-role-management.sql`：同 4.1 规则
- [x] 4.7 改写 `apps/lina-core/manifest/sql/010-distributed-locker.sql`：删除 `ENGINE=MEMORY`，改为普通持久表（参见任务组 5）
- [x] 4.8 改写 `apps/lina-core/manifest/sql/011-plugin-framework.sql`：同 4.1 规则
- [x] 4.9 改写 `apps/lina-core/manifest/sql/012-plugin-host-call.sql`：同 4.1 规则
- [x] 4.10 改写 `apps/lina-core/manifest/sql/013-dynamic-plugin-host-service-extension.sql`：删除 `ENGINE=MEMORY`（用于 `sys_kv_cache`），改为普通持久表；其他表按 4.1 规则
- [x] 4.11 改写 `apps/lina-core/manifest/sql/014-scheduled-job-management.sql`：同 4.1 规则；`BIGINT UNSIGNED` 改为 `BIGINT`；注意 `BIGINT GENERATED ALWAYS AS IDENTITY` 写法
- [x] 4.12 改写 `apps/lina-core/manifest/sql/015-distributed-cache-consistency.sql`：同 4.1 规则
- [x] 4.13 改写所有 `apps/lina-core/manifest/sql/mock-data/*.sql`：`INSERT IGNORE INTO` 按幂等依据改为 `INSERT INTO ... ON CONFLICT DO NOTHING`；删除 MySQL 特有函数；不显式写 IDENTITY 列值
- [x] 4.13.1 盘点宿主 SQL 中每个 `INSERT IGNORE INTO` 目标表并记录幂等依据或静态日志历史处理策略：seed 数据和具有稳定业务身份的 mock 数据必须具备覆盖稳定业务键的 `PRIMARY KEY` / `UNIQUE` 约束来触发 `ON CONFLICT DO NOTHING`；`sys_job_log`、`sys_notify_message`、`sys_notify_delivery` 等日志/历史类 mock 表不得仅为了 mock 数据幂等而新增限制真实业务写入的唯一约束，已改为精确存在性判断保证重复加载结果一致；文本业务键唯一约束使用 PostgreSQL 默认大小写敏感语义
- [x] 4.14 在 PG 数据库上逐个执行改写后的 SQL 文件，验证语法正确、约束、注释、索引创建符合预期；任意失败立即修正
- [x] 4.15 用重写后的 SQLite 翻译器，把 PG 源 SQL 翻译为 SQLite 兼容语句，在内存 SQLite 数据库上执行，验证全部成功

## 5. MEMORY 表改造（持久表 + 自然过期）

- [x] 5.1 在任务 4.4、4.7、4.10 中 SQL 已改为持久表，确认 DDL 不含 `ENGINE=MEMORY` 与 `UNLOGGED`/`TEMPORARY` 关键字
- [x] 5.2 检查 `apps/lina-core/internal/cmd/`、`internal/service/cluster/`、`internal/service/session/`、`internal/service/locker/`、`internal/service/kvcache/`，确认宿主启动、重启、leader 选举回调和插件运行时启动路径不对 `sys_online_session`、`sys_locker`、`sys_kv_cache` 执行 `TRUNCATE`、无条件全表 `DELETE` 或序列重置
- [x] 5.3 确认 `sys_online_session` 读取和 `session.CleanupInactive()` 均基于 `last_active_time` 处理过期会话；补充或更新单元测试覆盖过期会话不被视为有效
- [x] 5.4 确认 `sys_locker` 获取锁路径基于 `expire_time` 判断过期并允许抢占/覆盖过期锁；补充或更新单元测试覆盖过期锁抢占
- [x] 5.5 确认 `sys_kv_cache` 读取路径基于 `expire_at` 判定过期记录为未命中；补充或更新单元测试覆盖过期缓存读取
- [x] 5.6 写集成测试：在 PG 数据库上启动宿主，写入未过期会话、锁、KV cache，重启宿主后验证未过期记录仍按业务规则可用；写入过期记录后验证自然过期/清理路径生效
- [x] 5.7 写多进程集群模拟测试：启动至少两个宿主进程共用同一 PG 数据库，验证 leader 选举、leader 切换、节点重启不会清空易失性表，并验证过期锁抢占、会话过期清理、KV cache 跨实例一致性仍正常

## 6. plugin_data_table_comment 重构

- [x] 6.1 修改 `apps/lina-core/internal/service/plugin/plugin_data_table_comment.go`：删除硬编码的 `SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.TABLES WHERE TABLE_SCHEMA=? AND TABLE_NAME IN(?)` 查询
- [x] 6.2 引入 `lina-core/pkg/dialect`，调用 `dialect.FromDatabase(g.DB()).QueryTableMetadata(ctx, g.DB(), "public", tableNames)` 替代原查询
- [x] 6.3 将查询结果（`[]dialect.TableMeta`）映射为现有调用方期望的数据结构
- [x] 6.4 写单元测试，覆盖 PG 与 SQLite 两个分支的元数据查询行为

## 7. 配置文件 / 工具链 / 驱动切换

- [x] 7.1 修改 `apps/lina-core/manifest/config/config.yaml`：默认 link 改为 `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`；删除 `loc=Local&parseTime=true&multiStatements=true` 等 MySQL 特有参数（按 1.4 spike 结果定）
- [x] 7.2 修改 `apps/lina-core/manifest/config/config.template.yaml`：默认 link 同步改为 PG；注释中的 SQLite 示例链接保留
- [x] 7.3 修改 `apps/lina-core/hack/config.yaml`：`shared.localDatabaseLink` 改为 PG link；`gfcli.gen.dao.link` 同步切换
- [x] 7.4 修改 `apps/lina-core/go.mod`：移除 `github.com/go-sql-driver/mysql` 与 `github.com/gogf/gf/contrib/drivers/mysql/v2`；新增 `github.com/gogf/gf/contrib/drivers/pgsql/v2`
- [x] 7.5 修改 `apps/lina-core/main.go`：删除 mysql 驱动 import，新增 pgsql 驱动 import
- [x] 7.6 在 `apps/lina-core/` 目录运行 `go mod tidy`，确认依赖更新成功
- [x] 7.7 修改任何静态导入 mysql 驱动的源文件（搜索 `github.com/go-sql-driver/mysql` 与 `gogf/gf/contrib/drivers/mysql`），删除或替换为 pgsql

## 8. 数据库派生 uint64 → int64

- [x] 8.1 在 `apps/lina-core/` 全仓 `grep -rn 'uint64'`，输出所有命中点；按来源分组为 MySQL `UNSIGNED` 数据列派生、wire/wasm/metrics/protobuf 等天然无符号语义、第三方接口三类
- [x] 8.2 在 `internal/service/jobmgmt/`、`internal/service/locker/`、`internal/service/kvcache/`、`internal/service/config/` 等模块中识别由 MySQL `UNSIGNED` 数据列派生的 `uint64` 字段、参数、变量
- [x] 8.3 在对应 api DTO（`api/*/v1/*.go`）中识别由 MySQL `UNSIGNED` 数据列派生的 `uint64` 字段
- [x] 8.4 仅将 MySQL `UNSIGNED` 数据列派生类型替换为 `int64`，并相应调整初始化、比较、传递逻辑；wire / wasm / metrics / protobuf varint / byte count / uptime 等天然无符号语义必须保持 `uint64`
- [x] 8.5 编译验证：`cd apps/lina-core && go build ./...` 必须通过
- [x] 8.6 单元测试覆盖：相关模块的现有 `*_test.go` 必须全部通过

## 9. DAO 重新生成

- [x] 9.1 启动本地 PG 容器（可用 README 中的 `docker run` 示例或仓库已有 compose 文件），等待 `pg_isready` 返回成功；GitHub Actions 中使用 `services.postgres` 而不是 docker-compose
- [x] 9.2 运行 `make init confirm=init` 在 PG 上执行宿主 init SQL；任意 SQL 文件失败立即修正
- [x] 9.3 在 `apps/lina-core/` 运行 `make dao` 重新生成 `internal/dao/`、`internal/model/do/`、`internal/model/entity/` 全部 Go 文件
- [x] 9.4 编译验证：`cd apps/lina-core && go build ./...` 必须通过；如有数据库派生 `uint64` 残留，回到任务组 8 修复；非数据库派生 `uint64` 保持原类型
- [x] 9.5 检查生成的 entity 类型：所有 `*BIGINT UNSIGNED*` 字段已变为 `int64`；保留字列名（`Key`、`Value`、`Type`）字段在 entity 中字段名保持原大小写命名（GoFrame 默认风格）
- [x] 9.6 对 8 个包含实际 SQL 资源的插件（`content-notice`、`monitor-loginlog`、`monitor-online`、`monitor-operlog`、`monitor-server`、`org-center`、`plugin-demo-dynamic`、`plugin-demo-source`）分别运行其插件目录下的 `make dao`（如插件配置了独立 codegen），重新生成插件 DAO；`demo-control` 当前只有 `.gitkeep` 占位，无 DAO 生成项
- [x] 9.7 同样在插件业务代码中扫描 `uint64`，仅替换 MySQL `UNSIGNED` 数据列派生类型，保留监控指标、WASM ABI、协议编码等天然无符号语义

## 10. 插件 SQL 改写

- [x] 10.1 改写 `apps/lina-plugins/content-notice/manifest/sql/001-content-notice-schema.sql`：按任务 4.1 规则；卸载 SQL `manifest/sql/uninstall/001-content-notice-schema.sql` 同步；mock 数据 `manifest/sql/mock-data/001-content-notice-mock-data.sql` 按幂等依据改写 `INSERT IGNORE`
- [x] 10.2 改写 `apps/lina-plugins/monitor-loginlog/` 下所有 SQL 文件（schema/uninstall/mock-data）
- [x] 10.3 改写 `apps/lina-plugins/monitor-operlog/` 下所有 SQL 文件
- [x] 10.4 改写 `apps/lina-plugins/monitor-server/` 下所有 SQL 文件
- [x] 10.5 改写 `apps/lina-plugins/org-center/` 下所有 SQL 文件;注意此插件可能有表达式索引（参考现有 SQLite 翻译器的 `uk_plugin_org_center_dept_code` 用例），改为 PG 表达式索引语法 `CREATE UNIQUE INDEX ... ON x ((NULLIF(code, '')))`
- [x] 10.6 改写 `apps/lina-plugins/monitor-online/` 下所有 SQL 文件
- [x] 10.7 改写 `apps/lina-plugins/plugin-demo-dynamic/` 下所有 SQL 文件
- [x] 10.8 改写 `apps/lina-plugins/plugin-demo-source/` 下所有 SQL 文件
- [x] 10.9 确认 `apps/lina-plugins/demo-control/manifest/sql/` 当前只有 `.gitkeep` 占位文件，无实际 SQL 改写；若实施前新增 SQL 文件，必须纳入本任务组
- [x] 10.9.1 盘点插件 SQL 中每个 `INSERT IGNORE INTO` 目标表并记录幂等依据或静态日志历史处理策略：字典类 seed 依赖 `sys_dict_type.type` / `sys_dict_data(dict_type,value)`；关联表依赖复合主键；`plugin_monitor_loginlog`、`plugin_monitor_operlog` 等日志类 mock 表不得仅为了 mock 数据幂等而新增限制真实日志写入的唯一约束，已改为精确存在性判断保证重复加载结果一致；`plugin_content_notice`、`plugin_demo_source_record` 等业务演示数据如具有稳定业务身份则补充符合业务语义的唯一约束；文本业务键唯一约束使用 PostgreSQL 默认大小写敏感语义
- [x] 10.10 启动宿主，依次安装/启用每个包含实际 SQL 资源的插件，验证插件 install SQL 在 PG 上成功执行；触发插件功能（loginlog 写日志、monitor-server 上报指标等）确认表结构与业务逻辑正常
- [x] 10.11 卸载每个包含 uninstall SQL 的插件，验证 uninstall SQL 在 PG 上成功执行
- [x] 10.12 通过 SQLite 翻译器验证插件 SQL 在 SQLite 上也可执行成功

## 11. 本地 PG 启动 / GitHub Actions PG 服务 / 镜像 / 启动脚本

- [x] 11.1 更新本地 PostgreSQL 启动说明：优先提供 `docker run` 示例（postgres:14-alpine，`POSTGRES_USER=postgres`、`POSTGRES_PASSWORD=postgres`、`POSTGRES_DB=linapro`、端口 5432:5432、`pg_isready` 健康检查）；如仓库已有 compose 文件，则同步移除 mysql 服务并新增等价 postgres 服务
- [x] 11.2 修改 GitHub Actions 工作流：参考 `/Users/john/Workspace/github/gogf/gf/.github/workflows/ci-main.yml:113`，使用 job `services.postgres` 声明 `postgres:14-alpine` service container、环境变量、端口和 `pg_isready` healthcheck；CI 不使用 docker-compose 启动 PostgreSQL
- [x] 11.3 确认应用容器或本地开发路径不隐式管理数据库；如仓库维护 compose 文件，再按需配置 `depends_on` 和数据卷用于本地开发
- [x] 11.4 检查 `apps/lina-core/Dockerfile`：如有 mysql 客户端工具依赖，移除；不需打包 PG 客户端
- [ ] 11.5 调整 `make image` 流程，确认镜像构建仍然成功;镜像不内置数据库（外部依赖）
- [x] 11.6 保持 `make dev` 只编译并启动前后端，不自动启动或管理数据库；在 README 显式说明运行 `make init` / `make dev` 前需要由开发者或 CI 准备 PG（本地可用 `docker run` 或仓库已有 compose 文件，CI 使用 `services.postgres`）
- [x] 11.7 调整 `make init` 命令的连接错误提示：当 `pgsql` link 连接失败时，提示运维人员"PG 未就绪，请先启动 PostgreSQL 服务"，并给出本地启动示例

## 12. README / CLAUDE.md / 双语文档同步

- [x] 12.1 修改项目根 `README.md`：技术栈一节 "MySQL" 改为 "PostgreSQL（默认）/ SQLite（开发演示）"；快速开始一节增加启动 PostgreSQL 前置说明，优先给出 `docker run` 示例，如仓库已有 compose 文件再说明对应命令
- [x] 12.2 修改项目根 `README.zh-CN.md`：内容与 12.1 同步
- [x] 12.3 修改 `apps/lina-core/README.md`：数据库相关描述同步
- [x] 12.4 修改 `apps/lina-core/README.zh-CN.md`：内容与 12.3 同步
- [x] 12.5 修改 `CLAUDE.md`：技术栈段落 "GoFrame + MySQL + JWT" 改为 "GoFrame + PostgreSQL + JWT"
- [x] 12.6 在 README/技术文档中增加"切换到 SQLite 开发演示"指南段落（修改 `database.default.link` 为 `sqlite::@file(./temp/sqlite/linapro.db)`）
- [x] 12.7 在 README 中增加"PG 14+ 最低版本"说明
- [x] 12.8 在 README 中增加"切换到外部托管 PG（如 RDS / Aliyun PolarDB）"配置指南
- [x] 12.8.1 在 README 中说明 `make init` 是运维初始化操作，使用配置中的数据库账号执行；账号必须具备连接系统库、建库/删库、终止连接、建表、建索引、写注释和写 seed 数据的权限；权限不足时命令失败，运行时服务不做账号权限兜底
- [x] 12.9 检查所有 OpenSpec 归档文档中关于"MySQL"的引用，是否需要在归档说明中追加"已被 switch-default-database-to-postgres 取代"批注（仅在引用关键决策时追加，普通引用不动）
- [x] 12.10 验证双语 README 内容一致（关键段落对照）

## 13. 单元测试 / 集成测试 / E2E

- [x] 13.1 单元测试覆盖：`pkg/dialect/internal/postgres/*_test.go`、重写后的 `pkg/dialect/internal/sqlite/translate_test.go`、`pkg/dialect/dialect_postgres_test.go`、易失性表自然过期相关 service 测试、`plugin_data_table_comment_test.go`
- [x] 13.2 集成测试 1：在 PG 数据库上 `make init confirm=init` + `make mock confirm=mock`，启动宿主，验证管理员登录、菜单加载、字典查询、用户列表、定时任务列表、分布式锁加锁/释放、KV cache 读写均正常
- [x] 13.2.1 在 PG 数据库上重复执行 `make init confirm=init` 与 `make mock confirm=mock`，断言 host seed/mock 关键表行数和业务状态不变，尤其覆盖新增业务唯一约束触发的 `ON CONFLICT DO NOTHING` 幂等插入
- [x] 13.3 集成测试 2：在 PG 数据库上 `make init confirm=init rebuild=true`，验证已存在的 PG 数据库被正确 DROP/CREATE，且 `pg_terminate_backend` 终止活跃连接成功
- [x] 13.4 集成测试 3：切换到 SQLite link 跑 `make init confirm=init`，验证 SQLite 翻译器把 PG 源 SQL 转译后逐句成功执行；启动宿主跑核心功能 smoke
- [x] 13.4.1 在 SQLite link 上重复执行 host init/mock 与插件 install/mock SQL，断言关键表行数和业务状态不变
- [x] 13.5 E2E 用例（lina-e2e 技能新建，TC ID 按 lina-e2e 规范分配）：登录 / 字典管理 / 配置管理（含保留字列读写）/ 定时任务 / 分布式锁 / 文件上传下载 / 插件安装卸载 / 监控指标 完整业务流，全部在 PG 模式下跑通
- [x] 13.6 E2E/集成用例：宿主启动 + 重启场景，验证 `sys_online_session`、`sys_locker`、`sys_kv_cache` 三张易失性表在重启后不会被清空，未过期记录继续可用，过期记录按业务规则失效
- [x] 13.7 新增多进程集群模拟测试，并纳入 GitHub Actions：至少启动两个宿主进程共用 PG，验证 leader 选举、leader 切换、节点重启、主节点专属任务、缓存修订号/广播协调、分布式锁和会话/缓存自然过期行为；已补充 `TestClusterTwoHostProcessesSharePostgreSQL` 启动两个独立宿主进程、使用 `LINAPRO_NODE_ID` 区分节点、共享同一 PostgreSQL 数据库，覆盖单主健康模式、leader 停止后的接任，以及 `sys_online_session`、`sys_locker`、`sys_kv_cache` 启动不清空
- [x] 13.8 新增 SQL 幂等性检查测试或脚本：扫描 `ON CONFLICT DO NOTHING` 语句并确认声明幂等的目标表存在覆盖稳定业务键的 `PRIMARY KEY` / `UNIQUE` 约束；对日志/历史/监控类 mock 表确认已记录业务评估策略并使用精确存在性判断保持重复加载结果一致；任何无冲突依据但声称幂等的裸 `ON CONFLICT DO NOTHING` 均视为失败；文本业务键唯一约束使用 PostgreSQL 默认大小写敏感语义
- [x] 13.8.1 新增默认 collation 验证：在 PG 上验证用户名等文本业务键仅大小写不同会被唯一索引视为不同值；在 SQLite 翻译路径验证对应列不追加 `COLLATE NOCASE`
- [x] 13.9 单元测试 / 集成测试 / E2E 全部通过；任何失败必须修复后再标记任务完成。Docker 镜像真实打包和全新 clone README 流程因本机 Docker daemon 无响应仍作为 11.5 / 14.1 环境阻塞项保留未完成，不计入本项的代码与功能测试完成口径

## 14. 手动验证清单（实施完成前必跑）

- [ ] 14.1 全新 clone 项目，按 README 步骤启动本地 PostgreSQL（`docker run` 或仓库已有 compose 文件）→ `make init confirm=init` → `make mock confirm=mock` → `make dev`，验证一键启动成功
- [x] 14.2 浏览器访问 http://localhost:5666，使用 admin/admin123 登录成功
- [x] 14.3 用户管理：列表加载、新建用户、编辑、禁用/启用、删除（含数据权限校验）
- [x] 14.4 角色管理：列表、菜单分配、用户授权
- [x] 14.5 字典管理：字典类型 CRUD、字典数据 CRUD（含保留字列 `value` / `type` 读写）
- [x] 14.6 配置管理：配置项 CRUD（含保留字列 `key` / `value` 读写）
- [x] 14.7 部门 / 岗位（如启用对应模块）：CRUD
- [x] 14.8 文件管理：上传 / 下载 / 删除
- [x] 14.9 定时任务：创建 / 启用 / 立即执行 / 查看日志
- [x] 14.10 监控：服务监控页面（`monitor-server` 插件）显示数据库版本含 "PostgreSQL"
- [x] 14.11 插件管理：安装 / 启用 / 禁用 / 卸载 各个内置插件，验证 install/uninstall SQL 在 PG 上成功
- [x] 14.12 重启宿主，验证未过期的会话、锁、KV 缓存不会被启动流程清空；再验证过期的会话、锁、KV 缓存按自然过期规则失效或被清理，且持久数据（用户、角色、字典等）保留
- [x] 14.13 切换到 SQLite link（`sqlite::@file(./temp/sqlite/linapro.db)`），重新 `make init confirm=init`，验证启动日志输出"不得用于生产"警告，所有 cluster 组件不启动，核心功能可用
- [x] 14.14 切换到 `mysql:` link，验证启动失败并报"mysql 方言不再支持"明确错误

## 15. 文档完整性自检与提交前检查

- [x] 15.1 在 design.md 显式声明本次变更存在有限 i18n 影响：涉及 sysinfo/framework/apidoc 元数据从 MySQL 改为 PostgreSQL
- [x] 15.2 在 design.md 显式声明本次变更"不引入新缓存策略"（已在 D12 完成，确认无遗漏）
- [x] 15.3 检查所有改动涉及的双语 README 内容一致（项目根 + apps/lina-core）
- [x] 15.4 运行 `openspec validate switch-default-database-to-postgres` 通过验证
- [x] 15.5 调用 `/lina-review` 技能进行全面变更审查（CLAUDE.md 要求 archive 前必须做）
- [ ] 15.6 审查通过后，准备 git commit；commit 不在本变更任务清单内，遵循用户指令明确触发后再执行
- [ ] 15.7 准备 archive：调用 `/opsx:archive` 触发归档流程；归档前再次调用 `/lina-review`

## 16. 风险监控点（实施过程中持续关注）

- [x] 16.1 GoFrame PG 驱动行为有任何与预期不符的发现，立即记录在 design.md 对应 Risks 段落，并更新缓解策略
- [x] 16.2 保留字列名实测落到"GoFrame 不自动加引号"分支时，按 design.md R3 兜底方案处理（do/ 字段标签显式引号 → 列重命名最后兜底）
- [x] 16.3 嵌入资源构建脚本对 PG SQL 不兼容时，按 design.md R6 调整脚本（去掉语法预检查）
- [x] 16.4 业务代码 `uint64` → `int64` 替换发现依赖无符号语义的代码，单独评估解决方案，不强行替换
- [x] 16.5 集群模式下若出现自然过期、leader 切换或跨实例缓存一致性问题，优先通过多进程集群模拟测试复现，并调整既有 cluster/cache/session/locker/kvcache 逻辑；不得通过启动期清空易失性表规避问题
- [x] 16.6 SQL 改写阶段发现任何 `ON CONFLICT DO NOTHING` 没有实际冲突依据，立即回到任务 4.13.1 / 10.9.1 补齐符合业务语义的唯一约束或唯一索引；具有稳定业务身份的 seed/mock 数据不得通过临时 `WHERE NOT EXISTS`、无冲突依据的裸 `ON CONFLICT DO NOTHING` 或其他替代加载策略规避唯一性要求。日志/历史/监控类静态演示行确实不应新增业务唯一约束时，必须按任务 4.13.1 / 10.9.1 记录策略，并使用覆盖演示身份字段的精确存在性判断保证重复加载结果一致
- [x] 16.7 自定义 `linapro_ci` 排序规则方案已由 FB-14 移除；最终 SQL 不创建自定义排序规则，初始化账号不再需要创建排序规则权限

## Verification Notes

- 2026-05-09: PostgreSQL local verification used Postgres.app PG 18 on `127.0.0.1:55432` with link `pgsql:postgres:postgres@tcp(127.0.0.1:55432)/linapro?sslmode=disable`. This intentionally differs from README's default Docker `5432` port because Docker Desktop is not responsive in this environment.
- PG 14 coverage is wired into `.github/workflows/reusable-backend-unit-tests.yml` through `services.postgres` using `postgres:14-alpine` and `LINA_TEST_PGSQL_LINK=pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`; the current workstation only has Postgres.app 18 available, so local execution evidence remains PG18 while CI covers the PG14 minimum.
- Spike coverage is now captured in `apps/lina-core/pkg/dialect/internal/postgres/dialect_integration_test.go` and `prepare_test.go`: driver connectivity, `SELECT version()`、reserved `key`/`value` ORM access、identity `LastInsertId`、default collation case-distinct business keys、metadata lookup、and `PrepareDatabase(rebuild=true)` active-connection termination all pass when `LINA_TEST_PGSQL_LINK` is provided.
- Full backend unit tests passed with `cd apps/lina-core && GF_GCFG_PATH="$PWD/../../temp/test-config" go test -count=1 ./...`.
- PostgreSQL multi-process cluster integration passed with `cd apps/lina-core && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:55432)/linapro?sslmode=disable' go test -run TestClusterTwoHostProcessesSharePostgreSQL -count=1 ./internal/service/cluster -v`.
- A real multi-process startup race was found and fixed: concurrent source-plugin release metadata sync now treats PostgreSQL unique-key conflicts on `sys_plugin_release(plugin_id, release_version)` as a reread/update path instead of a startup-fatal insert error.
- Full E2E passed with `cd hack/tests && E2E_DB_PORT=55432 E2E_BROWSER_CHANNEL=chrome pnpm test`: 363 passed, 7 skipped. Skips are environment-gated SQLite/demo-control cases, not PostgreSQL failures.
- Final regression verification after review fixes passed: `cd apps/lina-core && GF_GCFG_PATH="$PWD/../../temp/test-config" go test -count=1 ./...`; PostgreSQL integration subset `LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:55432)/linapro?sslmode=disable' GF_GCFG_PATH="$PWD/../../temp/test-config" go test -count=1 ./internal/service/cluster ./internal/service/plugin/internal/catalog ./pkg/dialect/internal/postgres`; and race-focused subset `go test -race -run 'TestClusterTwoHostProcessesSharePostgreSQL|TestService_Lock_ConcurrentFreshLockRace|TestPrepareDatabaseRebuildWithPostgreSQL' -count=1 ./internal/service/cluster ./internal/service/locker ./pkg/dialect/internal/postgres`.
- `make image-build tag=postgres-validation` passed and staged image artifacts under `temp/image`; `make image tag=postgres-validation` reached `docker build` but Docker Desktop/CLI did not complete. `docker version` also hung after printing client metadata, so image Docker packaging remains environment-blocked while build artifacts are verified.
- Archive MySQL references were reviewed with `rg -n "MySQL|mysql" openspec/changes/archive`. They are historical decisions/specs from archived changes; this active change supersedes them, and no targeted archive note is needed before this change itself is archived.
- FB-11 verification passed after quoting all SQL field identifiers: `cd apps/lina-core && go test -count=1 ./pkg/dialect`; PostgreSQL asset smoke `LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:55432)/linapro?sslmode=disable' go test -run TestPostgreSQLProjectSQLAssetsSmoke -count=1 ./pkg/dialect -v`; host bootstrap `make init confirm=init rebuild=true`; host mock load `make mock confirm=mock`; and `openspec validate switch-default-database-to-postgres`.
- FB-12 verification passed after updating `lina-review` SQL review rules: `openspec validate switch-default-database-to-postgres`; manual review confirmed the skill now checks PostgreSQL double-quoted field identifiers and no longer lists `INSERT IGNORE` as a compliant idempotency pattern.
- FB-14 verification passed after removing custom collation usage: `openspec validate switch-default-database-to-postgres`; `cd apps/lina-core && go test -count=1 ./pkg/dialect ./pkg/dialect/internal/sqlite ./pkg/dialect/internal/postgres ./internal/service/kvcache/internal/sqltable ./internal/service/plugin/internal/wasm`; PostgreSQL asset smoke `LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:55432)/linapro?sslmode=disable' go test -run TestPostgreSQLProjectSQLAssetsSmoke -count=1 ./pkg/dialect -v`; and `rg` confirmed host/plugin SQL assets contain no `COLLATE`, `linapro_ci`, or `CREATE COLLATION`.
- FB-13 verification passed after centralizing GoFrame DB driver registration and the plugindb host driver factory: `cd apps/lina-core && go test -count=1 ./pkg/dbdriver ./pkg/plugindb/host ./pkg/dialect ./pkg/dialect/internal/postgres`; `cd apps/lina-core && GF_GCFG_PATH="$PWD/../../temp/test-config" go test -count=1 ./...`; and `openspec validate switch-default-database-to-postgres`.
- FB-15 verification passed after removing dictionary management created-time table columns: `cd apps/lina-vben && pnpm -F @lina/web-antd run typecheck`; `cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test e2e/settings/dict/TC0176-dict-created-time-column-hidden.ts`; `cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test e2e/settings/dict/TC0012-dict-type-crud.ts e2e/settings/dict/TC0013-dict-data-crud.ts`; and `openspec validate switch-default-database-to-postgres`. i18n impact is limited to removing two existing table header usages, so no translation resource changes are required; no API, cache, or data-permission behavior changed.
- FB-16 verification passed after standardizing Chinese README mirror names to `README.zh-CN.md`: the legacy underscore README pattern scan returned no matches outside ignored dependency and temp directories; the legacy underscore README filename search returned no source files; `openspec validate switch-default-database-to-postgres`; and `openspec validate readme-localization-governance --type spec`. i18n impact is limited to documentation filename governance and README links; runtime locale resources, API behavior, cache behavior, and data-permission behavior are unchanged.
- FB-17 verification passed after splitting plugin test utilities by responsibility: `cd apps/lina-core && GF_GCFG_PATH="$PWD/../../temp/test-config" go test -count=1 ./internal/service/plugin/...`; `cd apps/lina-core && GF_GCFG_PATH="$PWD/../../temp/test-config" go test -count=1 ./...`; and `openspec validate switch-default-database-to-postgres`. i18n, runtime cache behavior, API behavior, and data-permission behavior are unchanged because the change only reorganizes test helper source files.
- FB-18 verification passed after removing the README filename Go governance test and narrowing feedback testing rules: the removed test file no longer exists; stale reference scanning found no references to the removed test file, its package path, or its old test command; legacy underscore README pattern and filename scans still returned no source matches; `openspec validate switch-default-database-to-postgres`; and `openspec validate readme-localization-governance --type spec`. This is a project governance feedback item, so no unit test or E2E test is required by the updated rule; runtime behavior, API behavior, cache behavior, data-permission behavior, and i18n resources are unchanged.
- FB-19 verification passed after adding SQLite translation debug diagnostics for skipped `COMMENT ON TABLE/COLUMN` statements and explicit index-preservation coverage: `cd apps/lina-core && go test -count=1 ./pkg/dialect/internal/sqlite`; `cd apps/lina-core && go test -count=1 ./pkg/dialect`; `cd apps/lina-core && go test -run 'TestTranslateDDLDropsCommentsWithDebugLogPreservingIndexes|TestSQLiteTranslateDDLExecutesPostgreSQLFixture' -count=3 ./pkg/dialect/internal/sqlite ./pkg/dialect`; and `openspec validate switch-default-database-to-postgres`. Review found no issues: the logger uses the propagated translation context, comment filtering remains scoped to PostgreSQL comment metadata, and `CREATE INDEX` / `CREATE UNIQUE INDEX` / `DROP INDEX` / `REINDEX` statements stay on the compatible SQL path. i18n resources, runtime cache behavior, API behavior, and role data-permission behavior are unchanged.
- FB-20 verification passed after converting the `apps/lina-core/Makefile` terminal/help-facing Chinese text to English: `rg -n "[\\p{Han}]" apps/lina-core/Makefile` returned no matches; `make -C apps/lina-core init` printed the English confirmation guidance; and `openspec validate switch-default-database-to-postgres` passed. This is a project governance feedback item, so no unit test or E2E test is required; runtime behavior, API behavior, cache behavior, data-permission behavior, and i18n resources are unchanged.
- FB-21 verification passed after moving the remaining root `Makefile` command implementations into `hack/makefiles`: `make help`; `make -n init confirm=init`; and `make -n mock confirm=mock` all passed. A root target scan confirmed the root `Makefile` now only keeps shared variables and split makefile includes. This is a project governance feedback item, so no unit test or E2E test is required; runtime behavior, API behavior, cache behavior, data-permission behavior, and i18n resources are unchanged.
- FB-22 verification passed after resolving opened tab titles from stable runtime i18n keys at render time: `cd apps/lina-vben && pnpm -F @lina/web-antd run typecheck`; `cd hack/tests && pnpm run test:validate`; `cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test e2e/i18n/TC0177-tab-title-language-switch.ts`; and `openspec validate switch-default-database-to-postgres`. i18n impact is limited to frontend tab title relocalization behavior; no runtime language resource, backend API, cache behavior, or role data-permission behavior changed.

## Feedback

- [x] **FB-1**: 修复 `TestSyncBuiltinJobsPrunesRemovedBuiltins` 对内置任务清理调度器移除顺序的脆弱断言
- [x] **FB-2**: 根据需求澄清修正 PostgreSQL 切换提案、设计、规范和任务边界（易失性表自然过期、init 权限模型、插件清单、MERGE/SQLite/INSERT 语义、make dev 职责、多进程集群测试）
- [x] **FB-3**: 在项目规范与 `lina-review` 中增加单元测试自包含、顺序无关的审查要求
- [x] **FB-4**: 补充 `INSERT IGNORE` 迁移到 `ON CONFLICT DO NOTHING` 时的幂等性审查要求，禁止无冲突依据的机械替换
- [x] **FB-5**: 调整缺唯一键场景的处理优先级：先评估并补充业务唯一约束，不把字段较多的 `WHERE NOT EXISTS` 作为默认幂等方案
- [x] **FB-6**: 根据需求澄清收敛 PostgreSQL 大小写不敏感语义、init 权限要求、seed/mock 唯一约束要求和数据库方言名称
- [x] **FB-7**: 根据需求澄清修正 KV 表名、日志历史 mock 唯一约束策略、数据库派生 `uint64` 迁移范围和 GitHub Actions PG 启动方式
- [x] **FB-8**: 早期补充 `001-project-init.sql` 顶部 `CREATE COLLATION` 用途注释；该语句后续已由 FB-14 移除
- [x] **FB-9**: 审查并说明 PostgreSQL 表/列注释不能像 MySQL 一样内联到 `CREATE TABLE` 字段定义中
- [x] **FB-10**: 早期将 `001-project-init.sql` 顶部 `CREATE COLLATION` 用途注释补充为中英文双语说明；该语句后续已由 FB-14 移除
- [x] **FB-11**: 为所有数据库 SQL 中的字段标识符统一增加 PostgreSQL 双引号转义
- [x] **FB-12**: 将 SQL 字段标识符必须使用 PostgreSQL 双引号转义的要求加入 `lina-review` 审查规则
- [x] **FB-13**: 收敛 GoFrame 数据库驱动注册与插件数据驱动工厂
- [x] **FB-14**: 移除自定义 `linapro_ci` 排序规则，改用 PostgreSQL 默认 deterministic collation 以降低维护和性能成本
- [x] **FB-15**: 移除字典类型和字典数据列表中的创建时间列，释放字段名称与字典标签展示宽度
- [x] **FB-16**: 统一中文 README 镜像文件名为 `README.zh-CN.md`，与标准 i18n locale 标识保持一致
- [x] **FB-17**: 按功能职责拆分插件测试工具源文件，降低单文件维护负担
- [x] **FB-18**: 收窄反馈测试要求，移除 README 命名治理的 Go 单元测试并改用治理验证
- [x] **FB-19**: SQLite SQL 翻译器丢弃 `COMMENT ON TABLE/COLUMN` 注释语句时输出 Debug 日志，并确保创建和更新索引语句不会被误丢弃
- [x] **FB-20**: 将 `apps/lina-core/Makefile` 中会输出到终端或帮助文本的中文提示统一改为英文
- [x] **FB-21**: 将根目录 `Makefile` 中剩余命令迁移到 `hack/makefiles` 对应命令文件中
- [x] **FB-22**: 切换 i18n 语言时已打开 tab 页标题应联动切换到当前语言
