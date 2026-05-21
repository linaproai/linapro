## MODIFIED Requirements

### Requirement: 集群部署模式配置

宿主 SHALL 提供基于配置文件的集群部署模式开关。系统必须支持 `cluster.enabled` 作为总开关，并支持 `cluster.election.lease`、`cluster.election.renewInterval` 作为选举子配置。未显式配置时，`cluster.enabled` 必须默认为 `false`。当前支持的运行时数据库为 PostgreSQL；当数据库链接为 PostgreSQL 方言（`database.default.link` 以 `pgsql:` 开头）时，`cluster.enabled` 按用户配置生效，可启用集群模式。不支持的数据库方言必须在启动前失败，不得通过方言钩子改写集群配置后继续启动。

#### Scenario: 默认按单节点模式启动

- **WHEN** 配置文件未声明 `cluster.enabled`
- **THEN** 宿主按单节点模式启动
- **AND** 当前节点被视为主节点

#### Scenario: 显式开启集群模式

- **WHEN** 配置文件声明 `cluster.enabled=true` 且数据库链接为 PostgreSQL 方言
- **THEN** 宿主按集群模式启动
- **AND** 选主和主节点专属行为由集群模式统一控制

#### Scenario: SQLite 方言启动前失败

- **WHEN** 配置文件 `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件声明任意 `cluster.enabled` 值
- **THEN** 宿主在方言解析阶段启动失败
- **AND** 不启动选举循环、租约续期或节点投影同步
