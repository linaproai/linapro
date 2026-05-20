## REMOVED Requirements

### Requirement: SQLite 方言必须禁止集群 coordination

**Reason**: SQLite 不再是受支持数据库，`sqlite:` 链接必须在方言解析阶段失败，不再进入 cluster coordination 配置判断。

**Migration**: 使用 PostgreSQL 数据库。单机模式设置 `cluster.enabled=false`，集群模式设置 `cluster.enabled=true` 且 `cluster.coordination=redis`。
