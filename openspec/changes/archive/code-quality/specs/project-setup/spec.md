## MODIFIED Requirements

### Requirement: 数据库配置

系统 SHALL 使用 PostgreSQL 14+ 作为唯一运行时数据库，通过 GoFrame 官方 PG 驱动连接。系统 MUST NOT 支持 SQLite、MySQL 或其他数据库作为运行时数据库。

#### Scenario: SQLite 链接被显式拒绝

- **WHEN** 配置文件 `database.default.link` 以 `sqlite:` 开头
- **THEN** 后端启动失败并返回明确错误
- **AND** 错误消息说明 SQLite 不再支持
