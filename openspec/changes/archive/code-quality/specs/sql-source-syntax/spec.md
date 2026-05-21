## MODIFIED Requirements

### Requirement: SQL 源使用 PG 高级特性前必须单独评估

系统 SHALL 以 PostgreSQL 14+ 为唯一 SQL 源与执行方言。SQL 源默认使用项目约定的 PostgreSQL 14+ 可治理子集；使用 JSONB、数组类型、计算列、CREATE EXTENSION、CREATE FUNCTION、CREATE TRIGGER 等 PostgreSQL 高级特性前，必须新立 OpenSpec 变更评估。不再为了 SQLite 翻译能力限制 SQL 源。
