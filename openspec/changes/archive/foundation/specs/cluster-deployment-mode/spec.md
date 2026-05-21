## MODIFIED Requirements

### Requirement: 集群部署模式配置

宿主 SHALL 提供基于配置文件的集群部署模式开关。系统必须支持 `cluster.enabled` 作为总开关。未显式配置时，`cluster.enabled` 必须默认为 `false`。当数据库链接为 SQLite 方言（`database.default.link` 以 `sqlite:` 开头）时，`cluster.enabled` 在内存层被强制覆盖为 `false`；该覆盖发生在所有 cluster 相关组件初始化之前。当数据库链接为 PostgreSQL 方言（`database.default.link` 以 `pgsql:` 开头）时，`cluster.enabled` 按用户配置生效，可启用集群模式。

#### Scenario: 默认按单节点模式启动
- **WHEN** 配置文件未声明 `cluster.enabled`
- **THEN** 宿主按单节点模式启动
- **AND** 当前节点被视为主节点

#### Scenario: 显式开启集群模式
- **WHEN** 配置文件声明 `cluster.enabled=true` 且数据库链接为 PostgreSQL 方言
- **THEN** 宿主按集群模式启动
- **AND** 选主和主节点专属行为由集群模式统一控制

#### Scenario: SQLite 方言下集群模式被强制锁定
- **WHEN** 配置文件 `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件同时声明 `cluster.enabled=true`
- **THEN** `IsClusterEnabled` 在启动后稳定返回 `false`
- **AND** 选举循环、租约续期、节点投影同步均不启动

### Requirement: SQLite 方言启动时必须输出醒目的单机模式警告

系统 SHALL 在以 SQLite 方言启动时，向终端日志输出清晰的 SQLite 单机模式提示，明确告知用户：当前为 SQLite 模式、数据库链接路径、`cluster.enabled` 已被强制锁定为 `false`、所有功能在单机模式下运行、不得用于生产环境。

#### Scenario: SQLite 启动时打印警告
- **当** 宿主以 SQLite 链接启动时
- **则** 终端日志出现 SQLite 模式启动提示
- **且** 至少一行包含完整的数据库链接字符串
- **且** 至少一行包含"cluster.enabled 强制覆盖"的明确说明
- **且** 至少一行包含"不得用于生产"的明确警示

#### Scenario: PostgreSQL 启动不输出 SQLite 相关警告
- **当** 宿主以 PostgreSQL 链接启动时
- **则** 终端日志不出现任何 SQLite 模式相关提示

### Requirement: SQLite 方言下集群专属组件必须跳过启动

系统 SHALL 在 SQLite 方言下跳过所有仅服务于多节点部署的协调组件初始化，包括领导选举循环、租约续期任务、节点投影同步、缓存修订号广播协调。

#### Scenario: SQLite 模式跳过领导选举循环
- **当** 宿主以 SQLite 方言启动时
- **则** 集群组件读取 `IsClusterEnabled` 为 `false`
- **且** 不启动选举循环和租约续期任务

#### Scenario: SQLite 模式直接执行主节点专属调度任务
- **当** 宿主以 SQLite 方言运行且触发主节点专属定时任务时
- **则** 当前节点直接执行该任务
- **且** 与 PostgreSQL 单节点模式行为一致
