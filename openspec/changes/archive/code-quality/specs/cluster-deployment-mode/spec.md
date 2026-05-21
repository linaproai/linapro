## MODIFIED Requirements

### Requirement: 集群部署模式配置

宿主 SHALL 提供基于配置文件的集群部署模式开关。当数据库链接为 PostgreSQL 方言时，`cluster.enabled` 按用户配置生效。不支持的数据库方言必须在启动前失败，不得通过方言钩子改写集群配置后继续启动。

#### Scenario: SQLite 方言启动前失败

- **WHEN** 配置文件 `database.default.link` 以 `sqlite:` 开头
- **THEN** 宿主在方言解析阶段启动失败
- **AND** 不启动选举循环、租约续期或节点投影同步
