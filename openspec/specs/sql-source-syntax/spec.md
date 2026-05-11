# sql-source-syntax Specification

## Purpose
TBD - created by archiving change switch-default-database-to-postgres. Update Purpose after archive.
## Requirements
### Requirement: 项目所有 SQL 源文件 MUST 使用 PostgreSQL 14+ 语法子集编写

系统 SHALL 把 PostgreSQL 14+ 语法作为所有 SQL 源文件的唯一基准方言。`apps/lina-core/manifest/sql/` 下的宿主 SQL 文件、`apps/lina-core/manifest/sql/mock-data/` 下的宿主 mock 数据文件、`apps/lina-plugins/<plugin-id>/manifest/sql/` 下的所有插件 install/mock-data/uninstall SQL 文件 MUST 全部使用 PostgreSQL 14+ 语法编写，不得包含 MySQL、SQLite 或其他方言特有语法。SQL 源文件在 PG 数据库上 MUST 直接可执行（除 `CREATE DATABASE` 外，库的创建由方言层 `PrepareDatabase` 钩子单独处理）。SQL 源 MUST 使用 PostgreSQL 默认 deterministic collation，不得创建或依赖自定义排序规则。

#### Scenario: 宿主 SQL 在 PG 上直接执行
- **WHEN** 在已通过 `PrepareDatabase` 准备好的 PG 数据库上逐句执行 `apps/lina-core/manifest/sql/001-project-init.sql` 至 `015-distributed-cache-consistency.sql`
- **THEN** 每个语句 MUST 成功执行且不报语法错误
- **AND** 创建出的表结构、索引、约束、注释 MUST 与设计预期一致

#### Scenario: 插件 SQL 在 PG 上直接执行
- **WHEN** 在已通过插件 install pipeline 引导的 PG 数据库上逐句执行任一插件的 `manifest/sql/001-*.sql` 安装脚本
- **THEN** 每个语句 MUST 成功执行
- **AND** 不需要任何 PG 之外的方言转译

#### Scenario: SQL 源不包含 MySQL 特有语法
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `AUTO_INCREMENT`、`UNSIGNED`、`ENGINE=`、`DEFAULT CHARSET=`、`COLLATE=`、`TINYINT`、`LONGTEXT`、`MEDIUMTEXT`、`MEDIUMBLOB`、`LONGBLOB`、`VARBINARY`、`DATETIME`、反引号标识符、`INSERT IGNORE`、`ON DUPLICATE KEY UPDATE`、`ON UPDATE CURRENT_TIMESTAMP`、`KEY ... (...)` 内联索引、`UNIQUE KEY ... (...)` 内联索引、内联表/列 `COMMENT '...'`、`USE \`db\`` 中任何一项

### Requirement: SQL 源 MUST 使用默认 deterministic 文本比较语义

系统 SHALL 使用 PostgreSQL 默认 deterministic collation 提供文本比较、排序和唯一约束语义。SQL 源文件 MUST NOT 创建自定义 collation，也 MUST NOT 在列定义中声明 `COLLATE linapro_ci`、`COLLATE NOCASE` 或其他非默认排序规则。业务文本键默认大小写敏感：仅大小写不同的用户名、配置 key、字典类型、角色 key、菜单 key、插件 ID 或业务编码 SHALL 被视为不同值。若未来某个具体业务字段确实需要大小写不敏感语义，必须通过新的 OpenSpec 变更单独设计规范化写入、`lower(...)` 表达式唯一索引、`citext` 或等价方案，并评估 SQLite 翻译、索引性能和查询语义。

#### Scenario: 不创建自定义排序规则
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `CREATE COLLATION`
- **AND** 文件中 MUST NOT 出现 `linapro_ci`

#### Scenario: 业务文本键默认大小写敏感
- **WHEN** 表包含用户名、邮箱、手机号、字典类型、字典值、配置 key、角色 key、菜单 key、插件 ID、业务编码或其他稳定业务文本键
- **THEN** 对应列定义 MUST 使用普通 `VARCHAR` / `CHAR` / `TEXT` 类型
- **AND** 唯一索引或唯一约束 MUST 按 PostgreSQL 默认语义允许仅大小写不同的不同值

#### Scenario: SQLite 保持默认文本比较语义
- **WHEN** SQLite 方言转译 SQL 源中的文本列定义
- **THEN** 转译结果 MUST NOT 追加 `COLLATE NOCASE`
- **AND** SQLite 开发演示路径与 PostgreSQL 默认路径保持大小写敏感的唯一约束语义

### Requirement: SQL 源 MUST 使用 PG IDENTITY 列定义自增主键

系统 SHALL 要求所有自增主键列使用 PostgreSQL 的 `GENERATED ALWAYS AS IDENTITY` 子句声明。整数类型 SHALL 使用 `INT` 或 `BIGINT`（不带 `UNSIGNED`），不得使用 `SERIAL` / `BIGSERIAL` 简写形式（保留字段类型一致性）。

#### Scenario: INT 自增主键定义
- **WHEN** 表需要一个自增的 32 位整数主键
- **THEN** 列定义 MUST 是 `id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`

#### Scenario: BIGINT 自增主键定义
- **WHEN** 表需要一个自增的 64 位整数主键
- **THEN** 列定义 MUST 是 `id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`

#### Scenario: 不使用 SERIAL 简写
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `SERIAL` 或 `BIGSERIAL` 关键字

### Requirement: SQL 源 MUST 使用 PG 兼容的类型映射

系统 SHALL 要求所有数据列使用 PG 14+ 标准类型，并遵循统一的类型映射规则：整数族（`INT` / `BIGINT` / `SMALLINT`）、字符串（`VARCHAR(n)` / `CHAR(n)` / `TEXT`）、二进制（`BYTEA`）、时间（`TIMESTAMP`，不带时区）、定点数（`DECIMAL(m, n)`）、浮点数（`REAL` / `DOUBLE PRECISION`）。

#### Scenario: 时间列使用 TIMESTAMP
- **WHEN** 表需要存储日期时间
- **THEN** 列类型 MUST 是 `TIMESTAMP`（不带时区，与原 MySQL `DATETIME` 语义对齐）
- **AND** MUST NOT 使用 `TIMESTAMPTZ` / `DATE` / `TIME WITH TIME ZONE`

#### Scenario: 二进制列使用 BYTEA
- **WHEN** 表需要存储二进制数据
- **THEN** 列类型 MUST 是 `BYTEA`
- **AND** MUST NOT 使用 `BLOB` / `LONGBLOB` / `MEDIUMBLOB` / `VARBINARY`

#### Scenario: 长文本列使用 TEXT
- **WHEN** 表需要存储长度不固定的长文本
- **THEN** 列类型 MUST 是 `TEXT`
- **AND** MUST NOT 使用 `LONGTEXT` / `MEDIUMTEXT`

### Requirement: SQL 源 MUST 把表/列注释拆为独立 COMMENT ON 语句

系统 SHALL 要求所有表/列注释通过独立的 `COMMENT ON TABLE` 或 `COMMENT ON COLUMN` 语句声明，紧跟在 `CREATE TABLE` 之后，与 PG 标准用法一致。`CREATE TABLE` 语句体内 MUST NOT 包含任何内联 `COMMENT '...'` 子句。

#### Scenario: 表级注释
- **WHEN** 表需要注释说明
- **THEN** SQL 源 MUST 在 `CREATE TABLE` 后追加独立语句 `COMMENT ON TABLE <表名> IS '<注释>';`
- **AND** MUST NOT 使用内联 `... ) COMMENT='...';` 写法

#### Scenario: 列级注释
- **WHEN** 列需要注释说明
- **THEN** SQL 源 MUST 在 `CREATE TABLE` 后追加独立语句 `COMMENT ON COLUMN <表名>.<列名> IS '<注释>';`
- **AND** MUST NOT 使用列定义末尾的 `COMMENT '...'` 内联子句

### Requirement: SQL 源 MUST 把索引拆为独立 CREATE INDEX 语句

系统 SHALL 要求除 PRIMARY KEY 外的所有索引（包括普通索引、唯一索引、表达式索引）通过独立的 `CREATE INDEX` 或 `CREATE UNIQUE INDEX` 语句声明，紧跟在 `CREATE TABLE` 之后。`CREATE TABLE` 语句体内 MUST NOT 包含 `KEY` / `INDEX` / `UNIQUE KEY` / `UNIQUE INDEX` 内联索引子句。索引命名 SHALL 使用 `idx_{表名}_{列名}` 或 `uk_{表名}_{列名}` 约定，跨表唯一。

#### Scenario: 普通索引
- **WHEN** 列需要普通索引
- **THEN** SQL 源 MUST 在 `CREATE TABLE` 后追加独立语句 `CREATE INDEX idx_<表名>_<列名> ON <表名> (<列名>);`

#### Scenario: 唯一索引
- **WHEN** 列需要唯一索引
- **THEN** SQL 源 MUST 在 `CREATE TABLE` 后追加独立语句 `CREATE UNIQUE INDEX uk_<表名>_<列名> ON <表名> (<列名>);`
- **AND** MUST NOT 使用 `UNIQUE KEY` / `UNIQUE INDEX` 内联子句

#### Scenario: 表内不出现内联索引
- **WHEN** 扫描 `CREATE TABLE` 语句体
- **THEN** 语句体内只允许出现列定义、`PRIMARY KEY` 约束、`UNIQUE` 单列内联约束（仅作为列级约束，不带索引名）、`CHECK` 约束
- **AND** 语句体内 MUST NOT 出现 `KEY`、`INDEX`、`UNIQUE KEY`、`UNIQUE INDEX` 等命名索引子句

### Requirement: SQL 源 MUST 显式保证 INSERT 幂等

系统 SHALL 要求所有 Seed DML 与具有稳定业务身份的 mock 数据脚本中需要重复执行不报错且结果一致的 INSERT 语句使用 PG 标准的 `INSERT INTO ... ON CONFLICT DO NOTHING` 语法。每条用于声明幂等的 `ON CONFLICT DO NOTHING` INSERT MUST 由目标表上可触发冲突的 `PRIMARY KEY` 或 `UNIQUE` 约束支撑；该约束 MUST 覆盖该 seed/mock 记录的稳定业务键，而不是依赖未写入的自增主键。日志、历史、监控类表的 mock 数据若仅为静态演示历史记录且业务本身不要求唯一身份，MUST NOT 为了 mock 重复执行幂等强行新增会限制真实业务写入语义的唯一约束；实施者 MUST 按业务场景评估并记录处理策略，通过删除 mock、迁移为测试夹具内重置加载，或精确匹配静态演示行的存在性判断来保证重复执行结果一致。MUST NOT 将字段较多、易遗漏条件的 `WHERE NOT EXISTS` 作为默认替代方案。MUST NOT 使用 `INSERT IGNORE INTO` / `ON DUPLICATE KEY UPDATE` 等 MySQL 特有语法。MUST NOT 使用 `INSERT INTO ... ON CONFLICT (col) DO UPDATE SET ...` 形式（避免覆盖用户数据，符合 CLAUDE.md "禁止 ON DUPLICATE KEY UPDATE" 精神）。

#### Scenario: 使用 ON CONFLICT DO NOTHING 的幂等插入
- **WHEN** SQL 源选择使用 `ON CONFLICT DO NOTHING` 实现重复执行幂等
- **THEN** 语法 MUST 是 `INSERT INTO <表名> (...) VALUES (...) ON CONFLICT DO NOTHING;`
- **AND** 目标表 MUST 存在能覆盖该 INSERT 稳定业务键的 `PRIMARY KEY` 或 `UNIQUE` 约束
- **AND** 重复执行时 MUST 实际触发该冲突并跳过写入，而不是插入重复记录

#### Scenario: 具有稳定业务身份但缺少唯一约束时必须补充约束
- **WHEN** SQL 源目标表缺少覆盖 seed 记录或具有稳定业务身份的 mock 记录的自然唯一键
- **THEN** 实施者 MUST 补充符合业务语义的 `UNIQUE` 约束或唯一索引
- **AND** 该约束 MUST 覆盖该 seed/mock 记录的稳定业务键
- **AND** INSERT MUST 使用该约束触发 `ON CONFLICT DO NOTHING`

#### Scenario: 没有自然唯一键的日志历史 mock 数据不得强造唯一约束
- **WHEN** 日志、历史、监控类 mock 表只有自增主键，且真实业务允许多条相似记录存在
- **THEN** 实施者 MUST NOT 仅为了 mock 数据幂等而新增限制真实业务写入的唯一约束
- **AND** 实施者 MUST 记录处理策略：删除该 mock 数据、迁移为测试夹具内重置加载，或对少量静态演示行使用精确存在性判断
- **AND** 重复执行测试 MUST 断言最终数据库状态与单次执行后一致

#### Scenario: INSERT IGNORE 迁移必须逐条审查幂等依据
- **WHEN** 将现有 `INSERT IGNORE INTO` 迁移为 PostgreSQL 语法
- **THEN** 实施者 MUST 为每个目标表记录其幂等依据或静态日志历史处理策略（如 `sys_user.username`、`sys_dict_type.type`、`sys_dict_data(dict_type,value)`、关联表复合主键、日志历史静态行存在性判断等）
- **AND** 对缺少唯一约束但具有稳定业务身份的目标表，必须补充符合业务语义的唯一约束
- **AND** 对日志、历史、监控类目标表，必须根据业务场景评估是否需要唯一约束，不得仅为了 mock 数据重复执行而新增不符合业务语义的唯一约束
- **AND** 文本业务键对应的唯一约束按 PostgreSQL 默认 deterministic collation 工作，允许仅大小写不同的不同值
- **AND** 不得只做机械文本替换

#### Scenario: 不使用 INSERT IGNORE
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `INSERT IGNORE` 关键字

#### Scenario: 不使用 ON CONFLICT DO UPDATE
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `ON CONFLICT ... DO UPDATE SET` 子句

### Requirement: SQL 源 MUST NOT 包含 ON UPDATE CURRENT_TIMESTAMP 子句

系统 SHALL 要求所有 SQL 源文件不包含 `ON UPDATE CURRENT_TIMESTAMP` 内联子句。`updated_at` 列的实时更新 SHALL 由 GoFrame DAO 层在 Insert/Update/Save 操作时自动维护。`created_at` 列保留 `DEFAULT CURRENT_TIMESTAMP` 用于初始默认值。

#### Scenario: updated_at 列定义
- **WHEN** 表需要 `updated_at` 列
- **THEN** 列定义 MUST 是 `updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`
- **AND** MUST NOT 包含 `ON UPDATE CURRENT_TIMESTAMP` 子句
- **AND** 实时更新由 GoFrame 自动维护

#### Scenario: 扫描确认无 ON UPDATE 子句
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `ON UPDATE CURRENT_TIMESTAMP` 关键字

### Requirement: SQL 源 MUST 使用双引号包裹所有列标识符

系统 SHALL 要求所有 SQL 源文件中的列标识符在 DDL 与 DML 中统一使用 PostgreSQL 双引号包裹（如 `"id"`、`"username"`、`"created_at"`、`"key"`、`"value"`）。该规则覆盖列定义、`PRIMARY KEY`、索引列、`COMMENT ON COLUMN`、`INSERT` 列清单、子查询投影、`WHERE` / `JOIN` / `ORDER BY` / `GROUP BY` 中的列引用以及表达式中的列引用，避免大小写折叠、保留字冲突和后续 SQL 维护口径不一致。MUST NOT 因为保留字而重命名业务列，业务代码层依赖 GoFrame ORM 的自动引号行为或在 do/entity 字段标签中显式声明。

#### Scenario: 表字段定义统一转义
- **WHEN** 任一 `CREATE TABLE` 语句定义列
- **THEN** 每个列名 MUST 使用双引号包裹，例如 `"id" INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`
- **AND** 复合主键 MUST 写为 `PRIMARY KEY ("user_id", "role_id")`

#### Scenario: 索引与注释字段统一转义
- **WHEN** SQL 源声明索引或列注释
- **THEN** 索引列清单 MUST 使用双引号包裹，例如 `CREATE INDEX idx_sys_user_status ON sys_user ("status");`
- **AND** 列注释 MUST 写为 `COMMENT ON COLUMN sys_user."username" IS 'Username';`

#### Scenario: DML 字段统一转义
- **WHEN** SQL 源声明 `INSERT`、`SELECT`、`WHERE`、`JOIN`、`ORDER BY` 或 `GROUP BY` 中的列引用
- **THEN** 所有列标识符 MUST 使用双引号包裹，例如 `INSERT INTO sys_user ("username", "password") ...`
- **AND** 别名引用 MUST 写为 `u."username"` 或 `existing."user_name"`

#### Scenario: 业务列标识符保持原名
- **WHEN** 决定保留字列的处理策略
- **THEN** 业务列名 MUST NOT 因为转义要求或 PG 保留字冲突而重命名
- **AND** entity 与业务代码字段名保持原有命名

### Requirement: SQL 源 MUST NOT 包含 CREATE DATABASE 或 USE 语句

系统 SHALL 要求所有 SQL 源文件不包含 `CREATE DATABASE`、`DROP DATABASE` 或 `USE <database>` 任一语句。库的创建、删除、重建 SHALL 由 `Dialect.PrepareDatabase` 钩子在 SQL 加载前通过系统库连接独立执行。

#### Scenario: SQL 源不直接管理数据库
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `CREATE DATABASE` / `DROP DATABASE` / `USE <db>` 任一关键字组合

#### Scenario: 数据库管理由方言层钩子负责
- **WHEN** 运维人员运行 `make init confirm=init` 或 `make init confirm=init rebuild=true`
- **THEN** 库的创建/重建由 `Dialect.PrepareDatabase` 完成
- **AND** 后续 SQL 文件加载只关心表与数据，不关心库本身

### Requirement: SQL 源 MUST NOT 使用 PG 高级特性以保证 SQLite 翻译可行

系统 SHALL 限制 SQL 源使用纯 PG 14+ ANSI 子集，不使用 SQLite 翻译器无法处理的 PG 高级特性。被禁止的特性包括但不限于：`JSONB` / `JSON` 列类型与运算符（如 `->`、`->>`、`@>`）、数组类型与运算符、`GENERATED ALWAYS AS (expr) STORED` 计算列、`CREATE EXTENSION`、`CREATE FUNCTION`、`CREATE TRIGGER`、`CREATE TYPE`、`CREATE SCHEMA`（除 `public` 外）、`DOMAIN` 自定义域、`MERGE` 语句、`WITH RECURSIVE` 递归 CTE、`LATERAL` 联接、`TABLESAMPLE`、`PARTITION OF` 子句、`EXCLUSION CONSTRAINT`、`SERIAL` / `BIGSERIAL` 简写。如确有业务需要使用上述特性，必须新立 OpenSpec 变更评估对 SQLite 方言支持的影响。

#### Scenario: SQL 源不使用 JSONB
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `JSONB` / `JSON` 列类型，也 MUST NOT 出现 `->`、`->>`、`@>`、`<@`、`?` 等 JSON 运算符

#### Scenario: SQL 源不使用 PG 触发器与函数
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `CREATE TRIGGER` / `CREATE FUNCTION` / `CREATE PROCEDURE` 任一关键字

#### Scenario: SQL 源仅在 public schema 创建对象
- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `CREATE SCHEMA <非 public 名>` 语句
- **AND** 所有表、索引、约束 MUST 隐式创建在 `public` schema 下

### Requirement: SQL 源 MUST 满足跨方言执行的幂等性要求

系统 SHALL 要求所有 SQL 源文件可重复执行且结果一致（沿用 CLAUDE.md "SQL执行幂等性规范"）。`CREATE TABLE` 必须使用 `IF NOT EXISTS` 子句，`CREATE INDEX` 必须使用 `IF NOT EXISTS` 子句（PG 9.5+ 支持）；`INSERT` 语句使用有实际冲突依据的 `ON CONFLICT DO NOTHING`；`COMMENT ON` 语句天然幂等（重复执行覆盖原值）。

#### Scenario: 表创建幂等
- **WHEN** SQL 源创建表
- **THEN** 语法 MUST 是 `CREATE TABLE IF NOT EXISTS <表名> (...);`

#### Scenario: 索引创建幂等
- **WHEN** SQL 源创建索引
- **THEN** 语法 MUST 是 `CREATE INDEX IF NOT EXISTS <索引名> ON <表名> (<列名>);` 或 `CREATE UNIQUE INDEX IF NOT EXISTS ...`

#### Scenario: 重复执行不报错
- **WHEN** 在已经初始化过的数据库上重复执行同一个 SQL 源文件
- **THEN** 所有语句 MUST 成功执行且不引发错误
- **AND** 数据库最终状态 MUST 与单次执行后一致

