## MODIFIED Requirements

### Requirement: SQL 源使用 PG 高级特性前必须单独评估

系统 SHALL 以 PostgreSQL 14+ 为唯一 SQL 源与执行方言。SQL 源默认使用项目约定的 PostgreSQL 14+ 可治理子集；使用 `JSONB` / `JSON` 列类型与运算符、数组类型与运算符、`GENERATED ALWAYS AS (expr) STORED` 计算列、`CREATE EXTENSION`、`CREATE FUNCTION`、`CREATE TRIGGER`、`CREATE TYPE`、`CREATE SCHEMA`（除 `public` 外）、`DOMAIN` 自定义域、`MERGE` 语句、`WITH RECURSIVE` 递归 CTE、`LATERAL` 联接、`TABLESAMPLE`、`PARTITION OF` 子句、`EXCLUSION CONSTRAINT`、`SERIAL` / `BIGSERIAL` 简写等 PostgreSQL 高级特性前，必须新立 OpenSpec 变更评估可维护性、升级策略、索引性能、DAO 兼容性、备份恢复和测试覆盖。不再为了 SQLite 翻译能力限制 SQL 源。

#### Scenario: SQL 源不使用 JSONB

- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `JSONB` / `JSON` 列类型，也 MUST NOT 出现 `->`、`->>`、`@>`、`<@`、`?` 等 JSON 运算符，除非对应 OpenSpec 变更已经明确批准

#### Scenario: SQL 源不使用 PG 触发器与函数

- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `CREATE TRIGGER` / `CREATE FUNCTION` / `CREATE PROCEDURE` 任一关键字，除非对应 OpenSpec 变更已经明确批准

#### Scenario: SQL 源仅在 public schema 创建对象

- **WHEN** 扫描所有 SQL 源文件
- **THEN** 文件中 MUST NOT 出现 `CREATE SCHEMA <非 public 名>` 语句
- **AND** 所有表、索引、约束 MUST 隐式创建在 `public` schema 下
