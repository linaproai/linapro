## MODIFIED Requirements

### Requirement: 数据库配置

系统 SHALL 使用 PostgreSQL 14+ 作为默认数据库，通过 GoFrame 官方 PG 驱动 `gogf/gf/contrib/drivers/pgsql/v2` 连接。系统 SHALL 同时支持 SQLite 作为开发/演示方言，通过 GoFrame 内置的 SQLite driver 连接。系统 MUST NOT 支持 MySQL 作为运行时数据库。所有 SQL 源文件 MUST 使用 PostgreSQL 14+ 语法子集编写，可被 SQLite 方言翻译执行。PostgreSQL 默认路径 SHALL 使用数据库默认 deterministic collation，不创建或依赖自定义排序规则；业务文本键默认大小写敏感。

#### Scenario: PostgreSQL 默认数据库连接
- **WHEN** 后端服务启动且 `database.default.link` 以 `pgsql:` 开头
- **THEN** 后端通过 GoFrame PG 驱动连接到 PostgreSQL 数据库
- **AND** 服务启动不创建、删除或重建数据库
- **AND** 数据库创建、重建和 SQL 加载仅由 `make init confirm=init` / `make init confirm=init rebuild=true` 等运维初始化命令触发
- **AND** 业务文本键的唯一约束和等值匹配按 PostgreSQL 默认大小写敏感语义工作

#### Scenario: SQLite 开发演示模式
- **WHEN** 开发者将 `database.default.link` 改为 `sqlite::@file(./temp/sqlite/linapro.db)`
- **THEN** 后端通过 GoFrame SQLite 驱动连接到 SQLite 数据库文件
- **AND** SQLite 数据库文件父目录由 `make init` 的方言准备逻辑创建
- **AND** 启动日志输出"不得用于生产"的明确警示

#### Scenario: MySQL 链接被显式拒绝
- **WHEN** 配置文件 `database.default.link` 以 `mysql:` 开头
- **THEN** 后端启动失败并返回明确错误
- **AND** 错误消息说明 MySQL 不再支持，并列出当前支持的方言（`pgsql:`、`sqlite:`）
- **AND** 不静默回退到任何默认方言

#### Scenario: SQL 语法兼容性
- **WHEN** 编写 SQL schema 和查询
- **THEN** 所有 SQL 语句 MUST 使用 PostgreSQL 14+ 语法子集
- **AND** 不使用 PG 高级特性（JSONB、CREATE TRIGGER、CREATE FUNCTION、SERIAL/BIGSERIAL 简写、MERGE、WITH RECURSIVE 等）
- **AND** 可通过方言层 `TranslateDDL` 翻译为 SQLite 兼容语句
- **AND** MUST NOT 包含 MySQL 特有语法（AUTO_INCREMENT、UNSIGNED、ENGINE=、INSERT IGNORE、ON DUPLICATE KEY UPDATE 等）
- **AND** MUST NOT 创建或依赖自定义 collation；需要大小写不敏感语义的具体字段必须单独通过 OpenSpec 设计
