# 设计文档：switch-default-database-to-postgres

## Context

### 当前状态

LinaPro 项目当前的数据库支撑栈：

- **GoFrame 驱动**：`gogf/gf/contrib/drivers/mysql/v2`、`gogf/gf/contrib/drivers/sqlite/v2`
- **方言抽象层**：`apps/lina-core/pkg/dialect/`，包含 `Dialect` 接口、`From(link)` 工厂、`internal/mysql/`、`internal/sqlite/` 两个内部子包
- **SQL 源语法**：MySQL（单一源），SQLite 通过 `internal/sqlite/translate.go` 用 438 行正则改写
- **默认配置**：`database.default.link = mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?...`
- **数据表特征**：3 张原 `ENGINE=MEMORY` 语义表（`sys_online_session`、`sys_locker`、`sys_kv_cache`）依赖 MySQL MEMORY 引擎"重启即清"语义；其余表用 InnoDB。此前文档中的旧会话表名和旧 KV cache 表名均为错误名称，正确表名统一为 `sys_online_session` 与 `sys_kv_cache`

### 决策驱动

1. CLAUDE.md 第一条规则："全新项目，没有历史遗留问题和技术债务，因此在设计方案和实施的时候不需要考虑兼容性"——给出彻底移除 MySQL 的合法性
2. CLAUDE.md "跨数据库兼容"约定：业务代码已经禁用 MySQL 特有函数（`FIND_IN_SET`、`GROUP_CONCAT`、`IF()` 等），方言切换在业务层无大面积改动
3. 现有方言抽象层已经验证可行（SQLite 翻译器已存在），新增/替换方言是水平扩展

### 约束

- **PG 14+ 最低版本**：保证 `GENERATED ALWAYS AS IDENTITY`、`ON CONFLICT DO NOTHING`、成熟事务与约束能力可用。本次 SQL 源禁止使用 `MERGE`、JSONB 等 PG 高级特性，以保证 SQLite 翻译路径稳定
- **PG 文本比较语义**：PostgreSQL 路径使用数据库默认 deterministic collation，不创建自定义 ICU 排序规则。业务文本键默认大小写敏感；如未来具体字段需要大小写不敏感语义，需单独设计规范化写入、表达式索引、`citext` 或等价方案
- **SQLite 3.24+ 最低版本**：保证 `INSERT ... ON CONFLICT DO NOTHING` 可用（SQLite 3.24+ 支持 PG 风格的 ON CONFLICT 子句）
- **GoFrame 行为**：必须验证 GoFrame ORM 在 PG 上对保留字列名（`key`、`value`、`type`）的自动加引号行为
- **嵌入资源构建**：`prepare-packed-assets.sh` 自动重新打包 `manifest/sql/*` 到 `internal/packed/manifest/sql`，无需手动维护
- **i18n 治理**：CLAUDE.md 要求所有功能改动评估 i18n 影响，本次变更存在有限 i18n 影响：系统信息、framework 运行时文案和 apidoc 元数据中的数据库名称/技术栈描述需要从 MySQL 更新为 PostgreSQL
- **缓存一致性治理**：CLAUDE.md 要求所有缓存设计明确分布式一致性方案；本次变更不引入新的缓存策略，MEMORY → 持久表后不做启动期全表清空，相关数据依赖 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at` 与既有 TTL 清理路径自然过期

### 利益相关方

- **开发者**：本地启动需要 PG，不再是 MySQL；文档提供本地 PG 启动方式（可使用 `docker run` 或仓库已有 compose 文件）
- **CI/CD**：流水线 fixture 切换；E2E 主跑 PG；GitHub Actions 通过 `services.postgres` 启动 PG，不使用 docker-compose
- **生产部署**：默认 PG（外部托管或独立集群），文档显式不再支持 MySQL
- **插件作者**：插件 SQL 改写为 PG 语法；不影响 plugin.yaml 与运行时接口

## Goals / Non-Goals

### Goals

- **G1**：彻底移除 MySQL 支持，删除 `pkg/dialect/internal/mysql/` 包及驱动依赖
- **G2**：建立 `pkg/dialect/internal/postgres/` 子包，实现完整的 `Dialect` 接口
- **G3**：把 PostgreSQL 设为单一 SQL 源方言，重写所有宿主与插件 SQL 文件
- **G4**：重写 SQLite 翻译器为"PG → SQLite"翻译路径
- **G5**：把原 MEMORY 表改造为持久表 + 自然过期语义，不在应用启动期清空表数据
- **G6**：抽象 `Dialect.QueryTableMetadata` 接口，替换硬编码的 `information_schema.TABLES` 查询
- **G7**：实现 PG 特有的 `PrepareDatabase` 流程（连系统库 + 终止活跃连接 + DROP/CREATE）
- **G8**：默认配置、本地 PG 启动说明、CI service container 配置和文档全量切换到 PG
- **G9**：在 PostgreSQL 下使用默认 deterministic collation，避免自定义 ICU 排序规则带来的权限、维护和性能成本

### Non-Goals

- **NG1**：不引入 PG 高级特性（JSONB、数组类型、CTE 物化、并行查询等）。SQL 源保持 ANSI 子集，确保 SQLite 翻译可行
- **NG2**：不调整时区策略。继续使用 `TIMESTAMP`（不带时区），与现 `DATETIME` 语义对齐
- **NG3**：不引入多 schema 隔离。所有表在 `public` schema
- **NG4**：不双跑 E2E。CI 主跑 PG（生产路径），SQLite 仅覆盖单元/翻译测试
- **NG5**：不提供数据迁移工具。按 CLAUDE.md 第一条规则，重建数据库即可
- **NG6**：不重命名 PG 保留字列名（`key`/`value`/`type`），依赖 GoFrame ORM 自动加引号或 SQL 源显式加双引号
- **NG7**：不调整插件运行时接口或插件清单格式

## Decisions

### D1：PostgreSQL 作为单一 SQL 源语法（替代 MySQL）

**决策**：把 PG 设为唯一 SQL 源方言，所有 `manifest/sql/*.sql` 用 PG 14+ 子集编写。SQLite 通过翻译器从 PG 改写为 SQLite 兼容 SQL。

**为什么**：
- PG 比 MySQL 更接近 ANSI SQL，PG → SQLite 的翻译规则比 MySQL → SQLite 少
- PG 是默认方言，源 SQL = 默认方言确保生产路径无翻译开销，最稳
- PG 的标识符引号策略（不引号 = 全小写、双引号区分大小写）与代码风格天然契合
- 业务文本值的比较与唯一性采用 PG 默认 deterministic collation，保持大小写敏感和索引行为可预测

**备选方案**：
- **保留 MySQL 作为源**：与 G1 矛盾（不再支持 MySQL，把它当源很奇怪）
- **使用 SQL92 中性语法**：增加抽象层，调试痛苦，错失 PG 的强约束（CHECK、IDENTITY 等）
- **使用 SQLite 作为源**：SQLite 太宽松（动态类型、无 schema 约束），错失 PG 类型系统的好处

### D2：MEMORY 表改造为普通持久表 + 自然过期

**决策**：`sys_online_session`、`sys_locker`、`sys_kv_cache` 三张表在 PG 和 SQLite 上都是普通持久表。宿主启动、重启、滚动发布和集群 leader 切换时不执行 `TRUNCATE`、`DELETE` 或任何全表清理操作；表内数据分别由业务层 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at`、TTL 清理任务和锁过期抢占逻辑自然过期。

**为什么**：
- PG 没有真正等价 MySQL MEMORY 的引擎（UNLOGGED 不会重启清空，TEMP 是会话级）
- 改造后所有方言行为一致（PG/SQLite/未来其他方言）
- 启动期清空会在应用进程重启、滚动发布、leader 重新选举时误踢在线会话、删除有效锁或清空仍可用缓存，不符合运行时可靠性预期
- 业务层已经依赖会话最近活跃时间、锁过期时间和 KV cache 过期时间做 TTL 兜底，继续沿用自然过期模型即可

**备选方案**：
- **PG 用 UNLOGGED TABLE**：UNLOGGED 不在重启时清空（只在崩溃恢复时丢数据），语义错位
- **PG 用 TEMP TABLE**：TEMP 是会话级的，无法跨连接共享，分布式锁用不了
- **启动期 TRUNCATE**：会把进程重启误建模为数据库重启，影响滚动发布、集群 leader 切换和仍未过期的数据，被否

### D3：`INSERT IGNORE` → `INSERT ... ON CONFLICT DO NOTHING`

**决策**：所有 SQL 源中的 `INSERT IGNORE INTO ...` 必须逐条审查后改写。seed 数据和具有稳定业务身份的 mock 数据必须存在覆盖稳定业务键的 `PRIMARY KEY` 或 `UNIQUE` 约束，再改为 `INSERT INTO ... ON CONFLICT DO NOTHING`，冲突列由表约束自动识别（不写显式 `ON CONFLICT (col)`）。目标表没有自然唯一键、且 INSERT 不显式写入主键时，不得机械替换为裸 `ON CONFLICT DO NOTHING`。对于 `sys_job_log`、`sys_notify_message`、`sys_notify_delivery`、`plugin_monitor_loginlog`、`plugin_monitor_operlog` 等日志/历史/监控类表的 mock 数据，不得仅为了 mock 重复执行幂等而新增会限制真实业务写入的唯一约束；静态演示日志/历史记录改为精确匹配演示行的 `WHERE NOT EXISTS` 存在性判断，保证重复 `make mock` 结果一致。

**为什么**：
- PG 14+ 和 SQLite 3.24+ 都支持 `ON CONFLICT DO NOTHING` 语法
- 不写显式冲突列等同于"任意 unique 约束冲突都跳过"，在存在真实冲突依据时与 MySQL `INSERT IGNORE` 语义最接近
- `ON CONFLICT DO NOTHING` 不是无条件去重；如果目标表只有自增主键且 INSERT 未写入该主键，重复执行不会产生冲突，会插入重复行。日志/历史/监控类 mock 数据不能用不合理的唯一约束掩盖，也不能在交付型 mock SQL 中保留重复追加语义
- 翻译器无需推断冲突列（避免复杂的 SQL 上下文解析）

**实施要求**：
- 为每个现有 `INSERT IGNORE INTO` 目标表记录幂等依据，例如 `sys_user.username`、`sys_dict_type.type`、`sys_dict_data(dict_type,value)`、`sys_menu.menu_key`、`sys_role.key`、`sys_config.key`、`sys_online_session.token_id`、关联表复合主键等
- 对 `plugin_monitor_loginlog`、`plugin_monitor_operlog`、`sys_job_log`、`sys_notify_message`、`sys_notify_delivery` 等日志/历史/监控类 mock 表，先判断真实业务是否存在稳定唯一身份；只有业务语义确实要求唯一时才补充唯一约束或唯一索引
- 对缺少自然唯一键且业务不要求唯一身份的 mock/log/history 数据，不得为了迁移方便增加不符合业务语义的唯一约束；实施者必须在 SQL 改写记录中说明处理策略（例如删除该 mock、迁移为测试夹具内重置加载，或对少量静态演示日志使用精确存在性判断）
- 实施中不得为继续保留无唯一约束的 seed/mock 目标表而机械新增字段较多、易遗漏的 `WHERE NOT EXISTS`、临时 staging 表或其他绕开业务唯一性评估的加载策略；少量静态演示日志/历史行允许使用覆盖演示身份字段的存在性判断，且必须由重复加载测试验证结果一致
- 在 PG 和 SQLite 翻译路径上重复执行 host init、host mock、plugin install/mock SQL，断言关键表行数与业务状态不变

**备选方案**：
- **写显式 `ON CONFLICT (col)`**：要求 SQL 作者显式标注，翻译器无负担但 SQL 噪声大
- **PG 用 `ON CONFLICT DO NOTHING`，SQLite 翻译为 `INSERT OR IGNORE`**：可行但 SQLite 翻译多一条规则；既然 SQLite 3.24+ 也支持 ON CONFLICT，统一即可

### D4：`ON UPDATE CURRENT_TIMESTAMP` 全部移除，依赖 GoFrame 自动维护

**决策**：从所有 SQL 源中删除 `ON UPDATE CURRENT_TIMESTAMP` 子句。`updated_at` 列的实时更新由 GoFrame DAO 层在 Insert/Update/Save 时自动写入。

**为什么**：
- PG 没有等价的内联子句（需要 BEFORE UPDATE 触发器，破坏 ANSI 语法）
- SQLite 也没有等价（现状已经被 SQLite 翻译器移除）
- CLAUDE.md "软删除与时间维护规范"已经明确"禁止手动设置时间字段"，框架会自动处理
- 移除后，所有方言行为一致（无方言层开销）

**备选方案**：
- **PG 写 BEFORE UPDATE 触发器**：需要为每张表写触发器函数，污染 SQL 源；与 GoFrame 自动维护重复
- **保留 SQL 子句，由翻译器移除**：失去 D4 的统一性（PG 又特殊处理）

### D5：保留字列名（`key`/`value`/`type`）的处理策略

**决策**：依赖 GoFrame ORM 在 PG 上自动加双引号；在 SQL 源 DDL 中也使用双引号包裹这些列名（如 `"key" VARCHAR(64) NOT NULL`）确保 DDL 可执行。**不重命名列**。

**为什么**：
- 重命名列会扩散到 entity / api DTO / 业务代码 / 前端 API 客户端，改动面巨大
- GoFrame ORM 对 `key` / `value` 保留字列的写入与查询已在 PostgreSQL 实例上实测通过，写入业务代码层零改动
- DDL 显式加双引号保证 SQL 源在 PG 上无歧义

**实施验证结果**：
1. 写一个最小 PG 表（`CREATE TABLE test_kv ("key" VARCHAR, "value" TEXT)`）
2. 用 GoFrame ORM 写入并查询，验证保留字列访问成功
3. 用 `InsertAndGetId(...)` 写入 `GENERATED ALWAYS AS IDENTITY` 表，验证返回新 ID 正确
4. 结论：当前 GoFrame PG 驱动满足 D5 策略，不需要业务代码层手动加引号或列重命名

**备选方案**：
- **重命名为 `config_key`/`config_value`/`dict_type_code`**：业务代码大改，被否
- **不加双引号，依赖 PG 不区分大小写裸标识符**：`key` 在 PG 是真保留字，裸写会报错

### D6：PG 的 `PrepareDatabase` 通过系统库执行 DROP/CREATE

**决策**：PG `PrepareDatabase` 实现连接到 `postgres` 系统库（通过修改 link 字符串中的 dbname），执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`，然后让 GoFrame 重新连接到目标库。`make init` 是运维初始化命令，直接使用当前配置中的数据库账号；该账号必须具备连接系统库、创建数据库、删除数据库、终止目标库连接，以及在目标库内创建表、索引、注释并写入 seed 数据的足够权限。权限不足时命令快速失败并返回明确错误。运行时服务不提供低权限初始化、账号切换或跳过建库的兜底模式。

**为什么**：
- PG 不能在事务内 `DROP DATABASE`
- PG 不能在有活跃连接时 DROP（包括 GoFrame 自身的连接池）
- 必须连"非目标库"执行管理操作
- 运维初始化失败应显式暴露权限或环境问题，不在运行时引入额外的权限探测、账号切换、跳过数据库准备或低权限初始化逻辑

**实现要点**（伪代码骨架）：

```go
// pkg/dialect/internal/postgres/prepare.go
func PrepareDatabase(ctx context.Context, link string, rebuild bool) error {
    target, sysLink, err := splitLinkToSystemAndTarget(link)
    if err != nil { return err }

    // 1. 连接到系统库
    sysDB, err := openTransientDB(sysLink)
    if err != nil { return err }
    defer sysDB.Close()

    // 2. 检查目标库是否存在
    exists, err := databaseExists(sysDB, target.dbName)
    if err != nil { return err }

    if rebuild && exists {
        // 3a. 终止活跃连接
        _, err = sysDB.Exec(ctx,
            `SELECT pg_terminate_backend(pid) FROM pg_stat_activity
             WHERE datname=$1 AND pid<>pg_backend_pid()`,
            target.dbName)
        if err != nil { return err }
        // 3b. DROP
        _, err = sysDB.Exec(ctx,
            fmt.Sprintf("DROP DATABASE IF EXISTS %s",
                quoteIdentifier(target.dbName)))
        if err != nil { return err }
        exists = false
    }

    if !exists {
        // 4. CREATE
        _, err = sysDB.Exec(ctx,
            fmt.Sprintf("CREATE DATABASE %s ENCODING 'UTF8' "+
                "LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0",
                quoteIdentifier(target.dbName)))
        if err != nil { return err }
    }
    return nil
}
```

**为什么 `LC_COLLATE 'C' LC_CTYPE 'C'`**：跨平台 PG 默认 locale 不一致（macOS 是 `C`，Linux 容器可能是 `en_US.utf8` 或 `C.UTF-8`）。强制 `C` 保证数据库默认排序行为一致，避免不同环境复现失败。业务文本列使用该默认 deterministic collation，默认大小写敏感；如未来具体字段需要大小写不敏感语义，应通过单独变更设计规范化写入、表达式索引、`citext` 或等价方案。

### D7：`Dialect.QueryTableMetadata` 接口抽象

**决策**：`Dialect` 接口新增方法：

```go
type TableMeta struct {
    TableName    string
    TableComment string
}

type Dialect interface {
    // ... 现有方法
    QueryTableMetadata(ctx context.Context, db gdb.DB,
        schema string, tableNames []string) ([]TableMeta, error)
}
```

PG 实现（`pkg/dialect/internal/postgres/metadata.go`）核心查询：

```sql
SELECT
    t.table_name,
    COALESCE(obj_description(c.oid), '') AS table_comment
FROM information_schema.tables t
LEFT JOIN pg_class c ON c.relname = t.table_name AND c.relkind = 'r'
WHERE t.table_schema = $1 AND t.table_name = ANY($2)
```

SQLite 实现：从 `sqlite_master` 取表名，注释字段固定返回空字符串（SQLite 不支持表注释）。

调用方（`apps/lina-core/internal/service/plugin/plugin_data_table_comment.go`）改为：

```go
dialect, err := dialect.FromDatabase(g.DB())
if err != nil { return nil, err }
metas, err := dialect.QueryTableMetadata(ctx, g.DB(), "public", tableNames)
```

**为什么**：
- 当前硬编码 `information_schema.TABLES WHERE TABLE_SCHEMA=?` 是 MySQL 风格，PG 上 `TABLE_SCHEMA` 语义是 schema（默认 `public`），不是 database 名
- PG 的列注释要查 `pg_catalog.pg_description`，不在 `information_schema`
- SQLite 完全不支持表注释，需要单独处理
- 抽到 dialect 接口后，业务代码不再感知方言差异

### D8：数据库派生 ID 类型 `uint64` → `int64`

**决策**：PostgreSQL 官方整数类型为有符号的 `smallint` / `integer` / `bigint`，不支持 MySQL 风格的 `UNSIGNED INT` / `UNSIGNED BIGINT`。GoFrame 的字段本地类型映射也依赖数据库字段类型字符串中出现 `unsigned` 时才映射为 `uint` / `uint64`；PG 表结构不会提供该标记。因此，所有由 MySQL `INT/BIGINT UNSIGNED` 数据列派生的字段重新生成 entity 后变为 `int64`，对应 DAO / DO / Entity / API DTO / service ID 类型同步迁移为 `int64`。

本次迁移不修改非数据库派生的 `uint64`：包括 protobuf varint 编码、WASM ABI 指针/长度打包、监控指标（内存/磁盘/网络字节数、uptime、GC pause 等）、哈希、二进制协议、第三方库接口等天然无符号语义。

**为什么**：
- ID 类字段实际值远未达到 `int64` 上限（最大 9.2 × 10^18），`int64` 完全够用
- 数据库派生 ID 若继续保留 `uint64`，需要在业务代码层做 cast，破坏类型一致性
- 重新生成的 entity 类型与数据库派生业务代码必须匹配，否则编译失败
- 非数据库派生 `uint64` 保留原类型可以避免破坏 wire / wasm / metrics 等领域语义

**实施步骤**（写入 tasks.md）：
1. 全仓 `grep -rn 'uint64'` 定位所有声明点
2. 按来源分组：MySQL `UNSIGNED` 数据列派生类型、非数据库派生无符号语义、第三方/协议接口类型
3. 仅将 MySQL `UNSIGNED` 数据列派生类型替换为 `int64`
4. 保留 wire / wasm / metrics / protobuf varint / byte count / uptime 等天然无符号语义
5. 编译验证并回归测试相关业务逻辑

### D9：CI 主跑 PG，SQLite 仅覆盖单元/翻译测试

**决策**：E2E 测试默认连接 PG（生产路径），不双跑 SQLite。SQLite 仅在以下场景覆盖：
- 单元测试：`pkg/dialect/internal/sqlite/translate_test.go` 的翻译规则单测
- 集成测试：宿主启动 + init SQL 在 SQLite 上跑通的最小 smoke test
- 开发演示：开发者手动切换到 SQLite 链接验证

**为什么**：
- 生产路径必须严格验证（PG E2E 是必须的）
- SQLite 双跑 E2E 资源消耗翻倍，价值不高（SQLite 是演示，不是生产路径）
- SQLite 翻译规则的正确性由翻译器单元测试 + smoke test 保证

### D10：本地 PG 启动与 CI PG service container

**决策**：`make dev` 仍只负责编译并启动前后端，不负责启动或管理数据库。开发者必须在运行 `make init` / `make dev` 前自行准备 PG；本地可使用 `docker run` 或仓库已有 compose 文件（如后续维护 compose 文件则提供 postgres:14+ 服务定义）。GitHub Actions 不通过 docker-compose 启动 PostgreSQL，而是参考 GoFrame 的 `.github/workflows/ci-main.yml`，在 job 中使用 `services.postgres` 声明 PG service container、端口映射和 healthcheck。

```yaml
services:
  postgres:
    image: postgres:14-alpine
    env:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: linapro
      TZ: Asia/Shanghai
    ports:
      - 5432:5432
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
```

**为什么**：
- 数据库属于运行环境依赖，不应由 `make dev` 隐式管理
- 本地启动方式与 CI 启动方式分离，避免 GitHub Actions 依赖 compose 文件
- GitHub Actions service container 的 healthcheck 与 GoFrame workflow 模式一致，能等待 PG 就绪
- 如仓库后续维护 compose 文件，只服务于本地开发体验，不作为 CI 的 PG 启动方式

### D11：i18n 影响评估——有限影响

**决策**：本次变更存在有限 i18n 影响，需要把已有系统信息、framework 运行时文案和 apidoc 元数据中暴露的数据库名称或技术栈描述从 MySQL 更新为 PostgreSQL。除此之外，不新增业务 UI 文案、不引入新的翻译键体系。

**评估依据**：
- `sysinfo` 等系统信息响应、framework 运行时语言包和 apidoc 元数据会展示数据库/技术栈描述，必须与默认 PostgreSQL 保持一致
- 本次不改变业务页面菜单、按钮、表单、表格或错误提示语义，因此不需要新增业务翻译键
- 错误码文案（如 `DIALECT_LINK_REQUIRED`、`DIALECT_UNSUPPORTED`）已经使用 `bizerr` 框架的 i18n 占位符，仅文案中的支持前缀列表有微调（`mysql:` → `pgsql:`），通过 `bizerr` 参数机制处理，不需要新建翻译键
- `OnStartup` 启动期日志只写入终端日志，不进入前端 UI
- 需要同步维护前端运行时语言包、宿主 manifest i18n 资源和 apidoc i18n JSON 中相关 MySQL/PostgreSQL 元数据，避免残留过期描述

按 CLAUDE.md i18n 治理要求，本判断在 design.md 显式记录。

### D12：缓存一致性评估——无新增策略

**决策**：本次变更**不引入新的缓存策略**，沿用既有 `cluster.Service` 抽象。

**评估依据**：
- MEMORY → 持久表后不在启动期执行全表清空，也不新增跨实例清空协调
- 单节点与集群模式均依赖 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at`、TTL 清理、锁抢占和已有缓存一致性机制让数据自然过期
- `sys_kv_cache` 的应用层缓存策略不变（仍由 `kvcache` 服务维护跨实例一致性）
- `sys_online_session` 的过期清理仍由 `session.CleanupInactive()` 定时任务基于 `last_active_time` 驱动
- `sys_locker` 的锁过期仍由 `expire_time` 列驱动，与方言无关

按 CLAUDE.md 缓存一致性治理要求，本判断在 design.md 显式记录。

### D13：PostgreSQL 文本比较采用默认 deterministic collation

**决策**：PostgreSQL 初始化 SQL 不创建自定义 ICU 排序规则，所有文本列默认使用数据库创建时的 deterministic collation（当前 `PrepareDatabase` 创建库时使用 `LC_COLLATE 'C' LC_CTYPE 'C'`）。业务文本键默认大小写敏感，`admin` 与 `Admin`、`sys_config.key` 中仅大小写不同的值、字典类型中仅大小写不同的值均被视为不同值。

如未来某个具体字段必须具备大小写不敏感语义，不在本次 PostgreSQL 默认切换中统一处理；应单独通过 OpenSpec 设计并验证以下方案之一：写入前规范化为小写、`lower(column)` 表达式唯一索引、`citext` 扩展或等价机制，并同时评估 SQLite 翻译、索引性能、查询语义和运维权限。

**为什么**：
- 自定义 ICU 非确定性排序规则会引入初始化权限要求、ICU 版本升级后的 `REINDEX`/刷新维护点，以及 B-tree 索引去重限制和比较成本
- 项目是全新项目，不需要为了兼容 MySQL 默认 collation 保留大小写不敏感业务键语义
- 默认 deterministic collation 更符合 PostgreSQL 原生行为，性能、索引和运维模型更简单可预测
- 大小写不敏感需求应按字段精确建模，而不是作为所有业务文本键的默认数据库层行为

**SQLite 翻译策略**：
- SQLite 翻译器不再接受或模拟 `linapro_ci`
- 项目 SQL 源出现 `CREATE COLLATION` 时快速失败，提示该语法超出当前受控 PG 子集
- 文本列按普通 `VARCHAR` / `CHAR` / `TEXT` 转译为 SQLite `TEXT`，不追加 `COLLATE NOCASE`

**验证要求**：
- PG 集成测试覆盖默认唯一索引允许 `admin` 与 `Admin` 作为大小写不同的不同值共存
- SQLite 翻译测试覆盖文本列不生成 `COLLATE NOCASE`，项目 SQL 资产全文翻译执行成功
- SQL 扫描确认宿主和插件 SQL 中不存在 `linapro_ci`、`CREATE COLLATION` 或 `COLLATE NOCASE`

## SQL 改写规则与示例

### 类型映射表

| MySQL 源 | PG 14+ 目标 | SQLite 翻译目标 |
|---|---|---|
| `INT PRIMARY KEY AUTO_INCREMENT` | `INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `BIGINT UNSIGNED NOT NULL AUTO_INCREMENT` | `BIGINT GENERATED ALWAYS AS IDENTITY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `BIGINT UNSIGNED` | `BIGINT` | `INTEGER` |
| `INT UNSIGNED` | `INT` | `INTEGER` |
| `TINYINT` | `SMALLINT` | `INTEGER` |
| `SMALLINT` | `SMALLINT` | `INTEGER` |
| `INT` | `INT` | `INTEGER` |
| `BIGINT` | `BIGINT` | `INTEGER` |
| `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` |
| `CHAR(n)` | `CHAR(n)` | `TEXT` |
| `TEXT` | `TEXT` | `TEXT` |
| `LONGTEXT` / `MEDIUMTEXT` | `TEXT` | `TEXT` |
| `BLOB` / `LONGBLOB` / `MEDIUMBLOB` | `BYTEA` | `BLOB` |
| `VARBINARY(n)` | `BYTEA` | `BLOB` |
| `DATETIME` | `TIMESTAMP` | `DATETIME`（GoFrame 自动序列化） |
| `DECIMAL(m,n)` | `DECIMAL(m,n)` | `NUMERIC` |
| `FLOAT` / `DOUBLE` | `REAL` / `DOUBLE PRECISION` | `REAL` |

### 子句移除规则

| MySQL 源子句 | PG 处理 | SQLite 处理 |
|---|---|---|
| `ENGINE=InnoDB` | 删除 | 删除 |
| `ENGINE=MEMORY` | 删除（改用持久表 + 自然过期） | 删除（同 PG） |
| `DEFAULT CHARSET=utf8mb4` | 删除（PG 用库级 ENCODING） | 删除 |
| `COLLATE=utf8mb4_general_ci` | 删除（PG 用库级 LC_COLLATE） | 删除 |
| `UNSIGNED` | 删除 | 删除 |
| `ON UPDATE CURRENT_TIMESTAMP` | 删除（GoFrame 自动维护） | 删除 |
| 反引号 `` `id` `` | 删除（裸标识符）或 `"id"` | 删除（裸标识符） |

### 注释处理

**MySQL 源**：
```sql
CREATE TABLE sys_user (
    id INT PRIMARY KEY AUTO_INCREMENT COMMENT 'User ID',
    username VARCHAR(64) NOT NULL COMMENT 'Username'
) ENGINE=InnoDB COMMENT='User information table';
```

**PG 改写**：
```sql
CREATE TABLE sys_user (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(64) NOT NULL
);
COMMENT ON TABLE sys_user IS 'User information table';
COMMENT ON COLUMN sys_user.id IS 'User ID';
COMMENT ON COLUMN sys_user.username IS 'Username';
```

**SQLite 翻译**（注释丢弃）：
```sql
CREATE TABLE sys_user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL
);
-- COMMENT ON 语句被翻译器过滤掉
```

SQLite 翻译器在过滤 `COMMENT ON TABLE/COLUMN` 语句时输出 Debug 日志，日志包含源 SQL 文件与行号，便于诊断注释元数据被有意丢弃的情况。该过滤规则只匹配 PostgreSQL 注释语句，不影响独立的索引创建或维护语句。

### 索引处理

**MySQL 源**：
```sql
CREATE TABLE sys_user (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    phone VARCHAR(20),
    UNIQUE(username),
    KEY idx_phone (phone),
    KEY idx_status (status)
);
```

**PG 改写**（每个索引拆为独立语句）：
```sql
CREATE TABLE sys_user (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(64) NOT NULL,
    phone VARCHAR(20)
);
CREATE UNIQUE INDEX uk_sys_user_username ON sys_user (username);
CREATE INDEX idx_sys_user_phone ON sys_user (phone);
CREATE INDEX idx_sys_user_status ON sys_user (status);
```

**索引命名约定**：`idx_{表名}_{列名}` 或 `uk_{表名}_{列名}`，跨表唯一，避免 PG schema 全局命名空间冲突。

### INSERT IGNORE 改写

**MySQL 源**：
```sql
INSERT IGNORE INTO sys_user (username, password)
VALUES ('admin', '$2a$10$...');
```

**PG 改写**：
```sql
INSERT INTO sys_user (username, password)
VALUES ('admin', '$2a$10$...')
ON CONFLICT DO NOTHING;
```

该写法依赖 `sys_user.username` 唯一约束；该唯一约束按 PostgreSQL 默认 deterministic collation 工作，允许 `admin` 与 `Admin` 作为不同值共存。日志、历史、监控类 mock 表如果没有真实业务唯一身份，不得为了 mock 重复执行幂等新增人工唯一约束。例如 `plugin_monitor_loginlog`、`plugin_monitor_operlog`、`sys_job_log`、`sys_notify_message`、`sys_notify_delivery` 允许多条相似真实业务记录存在，不能通过 `(user_name, ip, login_time)`、请求摘要或消息内容等人工组合唯一键限制生产写入；本迭代保留的静态演示日志行通过精确存在性判断避免重复加载。

SQLite 翻译保留 `ON CONFLICT DO NOTHING` 原文（SQLite 3.24+ 兼容）。

### 默认文本列处理

**PG 源**：
```sql
CREATE TABLE IF NOT EXISTS sys_user (
    "id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "username" VARCHAR(64) NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_username ON sys_user ("username");
```

**SQLite 翻译**：
```sql
CREATE TABLE IF NOT EXISTS sys_user (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,
    "username" TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_username ON sys_user ("username");
```

该设计使用 PostgreSQL 和 SQLite 的默认大小写敏感唯一约束语义，`admin` 与 `Admin` 被视为不同值。

### 保留字列名处理

**PG DDL 源（显式双引号）**：
```sql
CREATE TABLE sys_config (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    "key" VARCHAR(128) NOT NULL,
    "value" TEXT
);
CREATE UNIQUE INDEX uk_sys_config_key ON sys_config ("key");
```

**SQLite 翻译**：保留双引号（SQLite 也接受双引号作为标识符引号）：
```sql
CREATE TABLE sys_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    "key" TEXT NOT NULL,
    "value" TEXT
);
```

### CREATE DATABASE / USE 处理

**MySQL 源**：
```sql
CREATE DATABASE IF NOT EXISTS `linapro` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `linapro`;
```

**PG 改写**：**整段删除**。库的创建由 `Dialect.PrepareDatabase` 钩子在连接前用系统库执行（参见 D6）。SQL 源文件中只保留表定义。

### MEMORY 表改造示例

**MySQL 源**（`010-distributed-locker.sql`）：
```sql
CREATE TABLE IF NOT EXISTS sys_locker (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    holder VARCHAR(64) DEFAULT '',
    expire_time DATETIME NOT NULL,
    UNIQUE KEY uk_name (name)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='Distributed lock table';
```

**PG 改写**：
```sql
CREATE TABLE IF NOT EXISTS sys_locker (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    holder VARCHAR(64) NOT NULL DEFAULT '',
    expire_time TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX uk_sys_locker_name ON sys_locker (name);
COMMENT ON TABLE sys_locker IS 'Distributed lock table';
```

运行期语义：

- 不新增 `TruncateVolatileTables` 或等价启动期清空入口。
- 宿主启动、重启、滚动发布和集群 leader 切换不得清空 `sys_online_session`、`sys_locker`、`sys_kv_cache`。
- `sys_online_session` 继续由会话过期清理任务基于 `last_active_time` 删除过期记录。
- `sys_locker` 继续由锁获取逻辑基于 `expire_time` 判断过期并允许抢占，必要时由既有清理任务回收。
- `sys_kv_cache` 继续由 KV cache 服务基于 `expire_at` 读取时判定过期并按既有策略清理。

PG 原生支持 `TRUNCATE TABLE ... RESTART IDENTITY`，但本次变更不在应用启动路径使用该语句。SQLite 不支持 `TRUNCATE` 语法；如未来需要在 SQLite 中实现显式清空，应使用 `DELETE FROM <table>; DELETE FROM sqlite_sequence WHERE name='<table>';` 等价方案，而不是直接执行 `TRUNCATE`。

## SQLite 翻译器（PG → SQLite）规则速记

新版翻译器 `pkg/dialect/internal/sqlite/translate.go` 应处理：

| 输入（PG）| 输出（SQLite）|
|---|---|
| `INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY` | `INTEGER PRIMARY KEY AUTOINCREMENT` |
| `BIGINT GENERATED ALWAYS AS IDENTITY` | `INTEGER PRIMARY KEY AUTOINCREMENT`（注意：SQLite 自增列必须叫 INTEGER，不能 BIGINT） |
| `INT` / `BIGINT` / `SMALLINT` | `INTEGER` |
| `VARCHAR(n)` / `TEXT` / `CHAR(n)` | `TEXT` |
| `BYTEA` | `BLOB` |
| `TIMESTAMP` | `DATETIME`（GoFrame 自动序列化） |
| `DECIMAL(m,n)` | `NUMERIC` |
| `COMMENT ON TABLE/COLUMN ...` | 整句丢弃，并输出 Debug 日志 |
| `INSERT ... ON CONFLICT DO NOTHING` | 保留（SQLite 3.24+ 原生支持） |
| `CREATE INDEX/UNIQUE INDEX ... ON x (col)` | 保留 |
| `DROP INDEX IF EXISTS ...` / `REINDEX ...` | 保留 |
| 双引号标识符 `"key"` | 保留 |

**未支持快速失败**：与现有规则一致，遇到 `JSONB`、`GENERATED ALWAYS AS (expr) STORED`、`CREATE EXTENSION`、`CREATE FUNCTION` 等 PG 特有语法时，翻译器返回明确错误（含 sourceName、行号、关键字）。

## 错误码映射

`pkg/dialect/internal/postgres/error.go` 关键定义：

```go
const (
    pgErrUniqueViolation       = "23505" // unique_violation
    pgErrSerializationFailure  = "40001" // serialization_failure
    pgErrDeadlockDetected      = "40P01" // deadlock_detected
    pgErrLockNotAvailable      = "55P03" // lock_not_available
    pgErrCheckViolation        = "23514" // check_violation
    pgErrForeignKeyViolation   = "23503" // foreign_key_violation
    pgErrNotNullViolation      = "23502" // not_null_violation
)

func IsRetryableWriteConflict(err error) bool {
    var pgErr *pq.Error // 或 pgconn.PgError，依驱动
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case pgErrSerializationFailure, pgErrDeadlockDetected, pgErrLockNotAvailable:
            return true
        }
    }
    return false
}
```

SQLite 错误码映射保持不变（`SQLITE_BUSY` / `SQLITE_LOCKED` 视为可重试）。

## Risks / Trade-offs

### R1：GoFrame PG 驱动行为差异风险

**风险**：GoFrame 的 PG 驱动在某些场景下（如保留字列名引号、连接字符串解析、IDENTITY 列回填）的行为可能与 MySQL 驱动不一致。

**实测结论**：使用 Postgres.app PG 18 与链接 `pgsql:postgres:postgres@tcp(127.0.0.1:55432)/linapro?sslmode=disable` 验证通过：驱动可连通、`SELECT version()` 成功、保留字列 `key` / `value` 可通过 GoFrame ORM 正常写入和查询、`GENERATED ALWAYS AS IDENTITY` 的 `InsertAndGetId` 返回值正确。该结论已由 `apps/lina-core/pkg/dialect/internal/postgres/dialect_integration_test.go` 的受控集成测试覆盖。

**缓解**：
- 保留受控集成测试，默认不依赖外部 PG；设置 `LINA_TEST_PGSQL_LINK` 后可在真实 PG 上重复验证关键驱动行为
- 如未来 GoFrame 或 PG 驱动升级导致行为变化，优先通过该测试暴露，再评估业务代码层显式引号或显式 `RETURNING id` 兜底

### R2：连接字符串格式确认

**风险**：GoFrame 的 PG link 字符串格式（`pgsql:user:pass@tcp(host:port)/db?...`）需要实测确认参数支持范围。

**实测结论**：`pgsql:postgres:postgres@tcp(127.0.0.1:55432)/linapro?sslmode=disable` 可正常连接并执行查询；不需要 MySQL 路径中的 `parseTime=true`、`loc=Local` 或 `multiStatements=true` 参数。项目 SQL 引导路径逐句解析并逐句提交，不依赖驱动级多语句开关。

**缓解**：
- 默认配置保持 `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`
- 如部署环境需要 SSL、时区或连接池参数，按 PG 驱动实际支持的 URL query 参数显式配置，不沿用 MySQL 专有参数

### R3：保留字列名实测可能落到"GoFrame 不自动加引号"分支

**风险**：D5 的策略依赖 GoFrame ORM 自动加双引号；如果不自动加，业务代码会大面积报"语法错误"。

**实测结论**：`key` / `value` 最小表写入和查询已通过真实 PG 集成测试；当前实现未落到兜底分支。

**缓解**：
- 保持 DDL 中对保留字列名使用双引号，业务代码继续通过 GoFrame ORM 访问
- 保留受控 PG 集成测试作为升级回归保护
- 若未来回归，兜底方案仍为业务代码层显式查询或列重命名，但当前不需要启用

### R4：MEMORY 表改造后的过期数据收敛风险

**风险**：MEMORY → 持久表后，数据库重启不再自动丢弃 `sys_online_session`、`sys_locker`、`sys_kv_cache` 的数据。如果 TTL 清理、锁过期抢占或缓存过期判断存在缺口，过期记录可能比预期保留更久。

**缓解**：
- 不在启动期清空数据，避免进程重启或 leader 切换误删除仍未过期的数据
- 单元测试覆盖 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at` 的过期判断与清理路径
- 多进程集群模拟测试覆盖 leader 选举、leader 切换、锁过期抢占、KV cache 跨实例一致性和会话过期清理

### R5：PG 14+ 在某些环境（老旧 Linux 发行版、企业内网）不可用

**风险**：明示 PG 14+ 后，部署环境受限。

**缓解**：
- PG 14 是 2021 年发布，主流云厂商（AWS RDS、Aliyun RDS、GCP Cloud SQL）均支持
- 本地容器示例使用 `postgres:14-alpine` 满足开发环境
- 文档中明确"如需更老 PG 版本，请提交 issue 评估"

### R6：嵌入资源构建脚本对 PG SQL 语法的兼容性

**风险**：`prepare-packed-assets.sh` 在打包 SQL 资源时如果做语法检查（如 `sqlite3 :memory: < file.sql`），可能在 PG 语法上失败。

**缓解**：
- 检查 `prepare-packed-assets.sh` 当前实现（不应做语法预检查，因为它需要适配多方言）
- 如确实做了 SQLite 预检查，调整为只做文件复制和列表生成，不做语法预检查；运行时翻译器负责验证

### R7：业务代码 `uint64` 替换的回归风险

**风险**：`uint64` → `int64` 看似简单，但若有业务逻辑显式依赖无符号语义（如位运算、模运算），可能产生隐性 bug。

**缓解**：
- 全仓 `grep -rn 'uint64'` 后，逐个分析使用场景并标注来源
- 仅迁移 MySQL `UNSIGNED` 数据列派生的 DAO / DO / Entity / API DTO / service ID 类型
- 保留 wire / wasm / metrics / protobuf varint / byte count / uptime 等天然无符号语义
- 单元测试覆盖关键数据库派生 ID 模块（jobmgmt、locker、kvcache、config）
- E2E 跑通完整业务流（登录、创建任务、加锁、读写缓存）

### R8：PG 容器启动慢 → CI 偶发失败

**风险**：CI 中 PG 容器启动时间不可预测，可能在应用启动前未就绪。

**缓解**：
- GitHub Actions 中使用 `services.postgres.options` 配置 `pg_isready` healthcheck，参考 GoFrame workflow
- 本地文档给出 `docker run` 或 compose 文件（如存在）启动方式
- `make init` 命令在连接失败时给友好提示："PG 未就绪，请先启动 PostgreSQL 服务"并给出本地启动示例

### R9：`make init` 使用低权限数据库账号失败

**风险**：运维可能在 `database.default.link` 中配置只具备运行时读写权限的账号。该账号无法连接系统库、创建数据库、删除数据库、终止连接或创建表/索引时，`make init` / `make init rebuild=true` 会失败。

**缓解**：
- 明确 `make init` 是运维初始化命令，使用配置中的账号直接执行数据库准备与建表写入逻辑
- 初始化账号必须具备足够权限；权限不足时快速失败并返回具体 PG 错误，不做静默降级、跳过建库、运行时账号切换或低权限兜底
- README 增加托管 PG / RDS 场景说明：必须使用具备建库、建表、建索引和写入 seed 数据权限的初始化账号运行 init

### R10：`INSERT IGNORE` 机械替换导致重复 mock/seed 数据

**风险**：MySQL `INSERT IGNORE` 常被误解为"重复执行自动去重"。PG 的 `ON CONFLICT DO NOTHING` 只在实际触发主键、唯一约束或排他约束冲突时跳过写入；如果目标表只有未写入的自增主键，重复执行会插入新行，导致 mock 日志、通知、演示记录等数据膨胀。

**缓解**：
- SQL 改写阶段为每个 `INSERT IGNORE INTO` 目标表列出幂等依据或日志/历史类静态演示行处理策略
- 对无自然唯一键的 mock/log/history 表，必须先评估业务是否需要唯一身份；禁止为了迁移方便增加不符合业务语义、会限制真实日志/历史写入的唯一约束
- 不把字段较多的 `WHERE NOT EXISTS` 或无冲突依据的裸 `ON CONFLICT DO NOTHING` 作为默认替代方案；若业务上无法定义稳定唯一键，必须记录处理策略（删除该 mock、使用测试夹具重置加载，或对少量静态演示日志使用覆盖演示身份字段的存在性判断）
- 集成测试在 PG 与 SQLite 翻译路径上重复执行 host init、host mock、plugin install/mock SQL，并断言关键表行数与业务状态不变

### R11：默认大小写敏感语义的产品边界

**风险**：去掉自定义 ICU 排序规则后，PostgreSQL 默认唯一索引允许 `admin` 与 `Admin` 这类仅大小写不同的业务键共存。如果未来产品要求某些字段大小写不敏感，不能依赖当前数据库默认行为自动满足。

**决策结论**：本项目是全新项目，优先选择 PostgreSQL 原生 deterministic collation，以降低初始化权限、ICU 版本升级、索引维护和比较性能成本。业务文本键默认大小写敏感是明确产品语义。

**缓解**：
- 文档和规格明确记录默认大小写敏感语义
- PG 集成测试覆盖默认唯一索引允许大小写不同值共存
- 如未来需要大小写不敏感字段，单独 OpenSpec 设计规范化写入、`lower(...)` 表达式唯一索引、`citext` 或等价机制

### R12：本地 Docker 守护进程不可用导致镜像打包验证受阻

**风险**：`make image` 最终需要 Docker daemon 完成镜像构建与打包。若本地 Docker Desktop 或 Docker CLI 无响应，镜像验证会被环境阻塞，无法代表代码或 Dockerfile 本身失败。

**实测结论**：`make image-build tag=postgres-validation` 已通过并生成 `temp/image` 构建产物；`make image tag=postgres-validation` 到达 `docker build` 阶段后因本地 Docker CLI/daemon 持续无响应被超时终止，`docker version` 也在输出 client 元数据后挂起。因此本地仅剩真实 Docker 打包动作受环境限制。

**缓解**：
- 在 Docker daemon 可用的 CI 或开发机上重新运行 `make image tag=postgres-validation`
- 保持镜像不内置数据库，运行时仍依赖外部 PostgreSQL
- 当前环境下用 `go build ./...`、完整单元测试、PG 初始化、完整 E2E 和 `make image-build` 覆盖代码与资源打包路径

## Migration Plan

### 部署步骤（项目内部）

1. **方言层骨架（A 阶段）**：删除 `internal/mysql/`，新建 `internal/postgres/`，扩展 `Dialect` 接口（`QueryTableMetadata`），更新 `From()` 注册
2. **GoFrame PG 驱动实测（spike）**：最小 main.go 验证连接、IDENTITY 列、保留字列名
3. **SQL 源改写（B 阶段）**：宿主 12 个 SQL 文件 + mock-data + 8 个包含实际 SQL 资源的插件，按 §SQL 改写规则与示例 逐个改写
4. **默认文本比较语义**：宿主 init SQL 不创建自定义排序规则，业务文本键使用 PostgreSQL 默认 deterministic collation，SQLite 翻译路径不追加 `COLLATE NOCASE`
5. **MEMORY 表改造（D 阶段）**：DDL 改持久表 + 确认启动路径不执行清空 + 验证自然过期语义
6. **SQLite 翻译器重写**：替换 `internal/sqlite/translate.go` 的输入语法基准
7. **`plugin_data_table_comment.go` 重构**：抽到 `Dialect.QueryTableMetadata`
8. **PG `PrepareDatabase`**：连系统库 + 终止连接 + DROP/CREATE 实现
9. **数据库派生 `uint64` → `int64`**：全仓扫描 + 来源判定 + 仅替换 MySQL `UNSIGNED` 数据列派生类型 + 编译验证
10. **DAO 重新生成**：`make dao` 在 PG 上运行
11. **配置切换**：`config.yaml` / `hack/config.yaml` / `go.mod` 同步切到 PG
12. **本地/CI PG 启动配置**：本地 PG 启动说明与 GitHub Actions `services.postgres` 配置
13. **文档双语同步**：README / README.zh-CN / CLAUDE.md
14. **测试**：单元测试 + 集成测试 + E2E 验证
15. **手动验证清单**：`make init` / `make mock` / `make dev` / 登录 / 关键模块功能

### 回退策略

- 本次变更前的 git commit 即是回退点
- `git revert <merge-commit>` 即可恢复 MySQL 路径（无数据风险，因为 PG 库可与原 MySQL 库共存）
- 已有 PG 数据库实例可保留作为审计快照

### 升级路径

- 因 CLAUDE.md 第一条规则，**不提供数据迁移工具**
- 现有用户升级时，通过 `make init confirm=init rebuild=true` 重建 PG 数据库
- mock 数据可通过 `make mock confirm=mock` 重新加载

## Open Questions

### Q1：GoFrame PG 驱动的 `parseTime`/`loc` 等价参数？

MySQL link 中的 `parseTime=true&loc=Local` 控制时间字段解析。PG link 是否有等价参数？

**结论**：不需要沿用 MySQL 的 `parseTime=true` 或 `loc=Local`。PG 默认链接使用 `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`；本地验证使用端口 `55432` 的 Postgres.app 实例，除端口外格式一致。

### Q2：GoFrame 在 PG 上的 `multiStatements` 行为？

MySQL link 中的 `multiStatements=true` 允许多语句执行。PG 是否需要等价参数？

**结论**：不需要等价参数。项目 SQL 引导命令已经按"逐句解析逐句提交"处理（见 `database-bootstrap-commands` spec），不依赖驱动级多语句支持；默认 PG 链接不包含 `multiStatements`。

### Q3：是否在本地 PG 启动方式中暴露 5432 端口到宿主机？

**默认是**（便于本地工具如 pgAdmin / DBeaver 直连调试）。但安全考虑下可只在容器网络内暴露。

**决策**：本地启动示例保持默认暴露 5432 到宿主机，便于开发体验。GitHub Actions 使用 service container 端口映射，不通过 docker-compose。生产部署文档明示外部 PG 应配置防火墙。

### Q4：SQLite 翻译器对 `COMMENT ON ...` 的处理粒度？

PG 源中每个表会有多条 `COMMENT ON COLUMN ... IS '...'` 语句。SQLite 翻译器有两种处理方式：
- **整句丢弃**：扫到 `^COMMENT ON ` 开头的语句直接跳过
- **解析后丢弃**：解析 SQL 语句类型，遇 `COMMENT` 类型整体跳过

**推荐**：整句丢弃（正则即可），实现简单。在 `translate.go` 中加一条规则：

```go
reCommentOn = regexp.MustCompile(`(?is)^\s*COMMENT\s+ON\b`)
// 翻译时遇到此模式 → 返回空字符串（跳过该语句）
```
