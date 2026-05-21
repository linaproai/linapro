## MODIFIED Requirements

### Requirement: 数据库配置

系统 SHALL 使用 PostgreSQL 14+ 作为唯一运行时数据库，通过 GoFrame 官方 PG 驱动 `gogf/gf/contrib/drivers/pgsql/v2` 连接。系统 MUST NOT 支持 SQLite、MySQL 或其他数据库作为运行时数据库。所有 SQL 源文件 MUST 使用 PostgreSQL 14+ 语法编写，并在 PostgreSQL 数据库上直接执行。PostgreSQL 默认路径 SHALL 使用数据库默认 deterministic collation，不创建或依赖自定义排序规则；业务文本键默认大小写敏感。

#### Scenario: PostgreSQL 默认数据库连接

- **WHEN** 后端服务启动且 `database.default.link` 以 `pgsql:` 开头
- **THEN** 后端通过 GoFrame PG 驱动连接到 PostgreSQL 数据库
- **AND** 服务启动不创建、删除或重建数据库
- **AND** 数据库创建、重建和 SQL 加载仅由 `make init confirm=init` / `make init confirm=init rebuild=true` 等运维初始化命令触发
- **AND** 业务文本键的唯一约束和等值匹配按 PostgreSQL 默认大小写敏感语义工作

#### Scenario: SQLite 链接被显式拒绝

- **WHEN** 配置文件 `database.default.link` 以 `sqlite:` 开头
- **THEN** 后端启动失败并返回明确错误
- **AND** 错误消息说明 SQLite 不再支持，并列出当前支持的方言仅为 `pgsql:`
- **AND** 不静默回退到任何默认方言

#### Scenario: MySQL 链接被显式拒绝

- **WHEN** 配置文件 `database.default.link` 以 `mysql:` 开头
- **THEN** 后端启动失败并返回明确错误
- **AND** 错误消息说明 MySQL 不再支持，并列出当前支持的方言仅为 `pgsql:`
- **AND** 不静默回退到任何默认方言

#### Scenario: SQL 语法兼容性

- **WHEN** 编写 SQL schema 和查询
- **THEN** 所有 SQL 语句 MUST 使用 PostgreSQL 14+ 语法
- **AND** MUST NOT 包含 MySQL 特有语法（AUTO_INCREMENT、UNSIGNED、ENGINE=、INSERT IGNORE、ON DUPLICATE KEY UPDATE 等）
- **AND** MUST NOT 包含 SQLite 特有语法或依赖 SQLite 文件数据库行为
- **AND** MUST NOT 创建或依赖自定义 collation；需要大小写不敏感语义的具体字段必须单独通过 OpenSpec 设计
