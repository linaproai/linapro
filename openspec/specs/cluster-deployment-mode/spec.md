# cluster-deployment-mode Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
### Requirement: 集群部署模式配置

宿主 SHALL 提供基于配置文件的集群部署模式开关。系统必须支持 `cluster.enabled` 作为总开关，并支持 `cluster.election.lease`、`cluster.election.renewInterval` 作为选举子配置。未显式配置时，`cluster.enabled` 必须默认为 `false`。

#### Scenario: 默认按单节点模式启动
- **WHEN** 配置文件未声明 `cluster.enabled`
- **THEN** 宿主按单节点模式启动
- **AND** 当前节点被视为主节点

#### Scenario: 显式开启集群模式
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **THEN** 宿主按集群模式启动
- **AND** 选主和主节点专属行为由集群模式统一控制

### Requirement: 单节点模式主节点语义

当 `cluster.enabled=false` 时，宿主 SHALL 将当前节点视为主节点，并跳过仅为多节点部署服务的宿主协调逻辑。

#### Scenario: 单节点模式跳过选主基础设施
- **WHEN** 宿主以单节点模式启动
- **THEN** 系统不启动领导选举循环
- **AND** 系统不启动租约续期逻辑

#### Scenario: 单节点模式直接执行主节点专属任务
- **WHEN** 宿主以单节点模式运行且触发主节点专属调度逻辑
- **THEN** 当前节点直接执行该逻辑
- **AND** 不需要额外的主节点判定等待

### Requirement: 插件运行时拓扑收敛

宿主 SHALL 根据集群部署模式控制动态插件运行时的收敛方式。单节点模式下，动态插件管理操作必须在当前节点同步完成；集群模式下，仍然允许由主节点负责最终切换与收敛。

#### Scenario: 单节点模式同步完成动态插件切换
- **WHEN** 宿主以单节点模式执行动态插件安装、启用、禁用、卸载或升级
- **THEN** 当前节点同步完成目标插件的状态切换
- **AND** 不依赖宿主主节点轮询才能生效

#### Scenario: 集群模式保留主节点收敛
- **WHEN** 宿主以集群模式执行动态插件管理操作
- **THEN** 系统允许先记录目标状态
- **AND** 由主节点执行最终切换与收敛

### Requirement: 节点投影同步仅在集群模式启用

宿主 SHALL 仅在集群模式下维护动态插件的节点投影状态。单节点模式不得要求 `sys_plugin_node_state` 成为插件运行态生效的前置条件。

#### Scenario: 单节点模式不写入节点投影
- **WHEN** 宿主以单节点模式同步插件元数据或运行时状态
- **THEN** 系统不依赖 `sys_plugin_node_state` 记录当前插件状态
- **AND** 插件治理视图仍然能够根据宿主稳定状态推导出当前状态

#### Scenario: 集群模式写入节点投影
- **WHEN** 宿主以集群模式同步动态插件运行时状态
- **THEN** 系统写入或更新当前节点对应的插件投影记录
- **AND** 记录包含节点标识、目标状态、当前状态和 generation

