## MODIFIED Requirements

### Requirement: 非 PostgreSQL 数据库链接必须在 coordination 启动前失败
系统仅支持 PostgreSQL 运行时数据库。`sqlite:`、`mysql:` 或未知数据库链接 MUST 在方言解析阶段失败，不得进入 Redis coordination 探活、集群配置覆盖或业务启动流程。

#### Scenario: SQLite 配置了 Redis coordination
- **WHEN** `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件声明 `cluster.enabled=true`
- **AND** 配置文件声明 `cluster.coordination=redis`
- **THEN** 宿主启动失败并返回 SQLite 不再支持的明确错误
- **AND** 系统不输出 SQLite 单机模式警告或进入单机降级
- **AND** 系统不得连接 Redis

#### Scenario: SQLite 配置了单机模式
- **WHEN** `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件声明 `cluster.enabled=false`
- **THEN** 宿主启动失败并返回 SQLite 不再支持的明确错误
- **AND** 系统不得进入单机缓存、单机 coordination 或业务启动流程
