## ADDED Requirements

### Requirement: 项目所有 SQL 源文件 MUST 使用 PostgreSQL 14+ 语法子集编写

系统 SHALL 把 PostgreSQL 14+ 语法作为所有 SQL 源文件的唯一基准方言。`apps/lina-core/manifest/sql/` 下的宿主 SQL 文件、`apps/lina-core/manifest/sql/mock-data/` 下的宿主 mock 数据文件、`apps/lina-plugins/<plugin-id>/manifest/sql/` 下的所有插件 install/mock-data/uninstall SQL 文件 MUST 全部使用 PostgreSQL 14+ 语法编写，不得包含 MySQL、SQLite 或其他方言特有语法。SQL 源文件在 PG 数据库上 MUST 直接可执行（除 `CREATE DATABASE` 外，库的创建由方言层 `PrepareDatabase` 钩子单独处理）。SQL 源 MUST 使用 PostgreSQL 默认 deterministic collation，不得创建或依赖自定义排序规则。

#### Scenario: 宿主 SQL 在 PG 上直接执行
- **WHEN** 在已通过 `PrepareDatabase` 准备好的 PG 数据库上逐句执行宿主 SQL 文件
- **THEN** 每个语句 MUST 成功执行且不报语法错误
- **AND** 创建出的表结构、索引、约束、注释 MUST 与设计预期一致

#### Scenario: 插件 SQL 在 PG 上直接执行
- **WHEN** 在已通过插件 install pipeline 引导的 PG 数据库上逐句执行任一插件的安装脚本
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

系统 SHALL 要求所有自增主键列使用 PostgreSQL 的 `GENERATED ALWAYS AS IDENTITY` 子句声明。整数类型 SHALL 使用 `INT` 或 `BIGINT`（不带 `UNSIGNED`），不得使用 `SERIAL` / `BIGSERIAL` 简写形式。

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

系统 SHALL 要求所有数据列使用 PG 14+ 标准类型：整数族（`INT` / `BIGINT` / `SMALLINT`）、字符串（`VARCHAR(n)` / `CHAR(n)` / `TEXT`）、二进制（`BYTEA`）、时间（`TIMESTAMP`，不带时区）、定点数（`DECIMAL(m, n)`）、浮点数（`REAL` / `DOUBLE PRECISION`）。

#### Scenario: 时间列使用 TIMESTAMP
- **WHEN** 表需要存储日期时间
- **THEN** 列类型 MUST 是 `TIMESTAMP`（不带时区）
- **AND** MUST NOT 使用 `TIMESTAMPTZ` / `DATE` / `TIME WITH TIME ZONE`

#### Scenario: 二进制列使用 BYTEA
- **WHEN** 表需要存储二进制数据
- **THEN** 列类型 MUST 是 `BYTEA`
- **AND** MUST NOT 使用 `BLOB` / `LONGBLOB` / `MEDIUMBLOB` / `VARBINARY`

### Requirement: SQL 源 MUST 把表/列注释拆为独立 COMMENT ON 语句

系统 SHALL 要求所有表/列注释通过独立的 `COMMENT ON TABLE` 或 `COMMENT ON COLUMN` 语句声明。`CREATE TABLE` 语句体内 MUST NOT 包含任何内联 `COMMENT '...'` 子句。

### Requirement: SQL 源 MUST 把索引拆为独立 CREATE INDEX 语句

系统 SHALL 要求除 PRIMARY KEY 外的所有索引通过独立的 `CREATE INDEX` 或 `CREATE UNIQUE INDEX` 语句声明。`CREATE TABLE` 语句体内 MUST NOT 包含 `KEY` / `INDEX` / `UNIQUE KEY` / `UNIQUE INDEX` 内联索引子句。索引命名 SHALL 使用 `idx_{表名}_{列名}` 或 `uk_{表名}_{列名}` 约定，跨表唯一。

### Requirement: SQL 源 MUST 显式保证 INSERT 幂等

系统 SHALL 要求所有 Seed DML 与具有稳定业务身份的 mock 数据脚本中需要重复执行不报错且结果一致的 INSERT 语句使用 PG 标准的 `INSERT INTO ... ON CONFLICT DO NOTHING` 语法。每条用于声明幂等的 `ON CONFLICT DO NOTHING` INSERT MUST 由目标表上可触发冲突的 `PRIMARY KEY` 或 `UNIQUE` 约束支撑。日志、历史、监控类表的 mock 数据若仅为静态演示历史记录且业务本身不要求唯一身份，MUST NOT 为了 mock 重复执行幂等强行新增会限制真实业务写入语义的唯一约束。MUST NOT 使用 `INSERT IGNORE INTO` / `ON DUPLICATE KEY UPDATE` / `ON CONFLICT ... DO UPDATE SET` 等语法。

### Requirement: SQL 源 MUST NOT 包含 ON UPDATE CURRENT_TIMESTAMP 子句

系统 SHALL 要求所有 SQL 源文件不包含 `ON UPDATE CURRENT_TIMESTAMP` 内联子句。`updated_at` 列的实时更新 SHALL 由 GoFrame DAO 层自动维护。

### Requirement: SQL 源 MUST 使用双引号包裹所有列标识符

系统 SHALL 要求所有 SQL 源文件中的列标识符在 DDL 与 DML 中统一使用 PostgreSQL 双引号包裹。该规则覆盖列定义、`PRIMARY KEY`、索引列、`COMMENT ON COLUMN`、`INSERT` 列清单、子查询投影、`WHERE` / `JOIN` / `ORDER BY` / `GROUP BY` 中的列引用以及表达式中的列引用。MUST NOT 因为保留字而重命名业务列。

### Requirement: SQL 源 MUST NOT 包含 CREATE DATABASE 或 USE 语句

系统 SHALL 要求所有 SQL 源文件不包含 `CREATE DATABASE`、`DROP DATABASE` 或 `USE <database>` 任一语句。库的创建、删除、重建 SHALL 由 `Dialect.PrepareDatabase` 钩子在 SQL 加载前通过系统库连接独立执行。

### Requirement: SQL 源 MUST NOT 使用 PG 高级特性以保证 SQLite 翻译可行

系统 SHALL 限制 SQL 源使用纯 PG 14+ ANSI 子集，不使用 SQLite 翻译器无法处理的 PG 高级特性。被禁止的特性包括但不限于：`JSONB` / `JSON`、数组类型、`GENERATED ALWAYS AS (expr) STORED`、`CREATE EXTENSION`、`CREATE FUNCTION`、`CREATE TRIGGER`、`CREATE TYPE`、`CREATE SCHEMA`（除 `public` 外）、`DOMAIN`、`MERGE`、`WITH RECURSIVE`、`LATERAL`、`TABLESAMPLE`、`PARTITION OF`、`EXCLUSION CONSTRAINT`、`SERIAL` / `BIGSERIAL`。

### Requirement: SQL 源 MUST 满足跨方言执行的幂等性要求

系统 SHALL 要求所有 SQL 源文件可重复执行且结果一致。`CREATE TABLE` 必须使用 `IF NOT EXISTS` 子句，`CREATE INDEX` 必须使用 `IF NOT EXISTS` 子句；`INSERT` 语句使用有实际冲突依据的 `ON CONFLICT DO NOTHING`；`COMMENT ON` 语句天然幂等。
