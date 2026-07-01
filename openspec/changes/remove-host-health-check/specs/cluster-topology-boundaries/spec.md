## MODIFIED Requirements

### Requirement: 节点身份必须贯穿 coordination 事件
系统 SHALL 在 coordination lock、revision event、插件运行时事件和系统信息诊断中携带稳定 node ID。node ID MUST 由 cluster/topology 层统一提供。

#### Scenario: 发布事件包含 sourceNode
- **WHEN** 节点发布 cache invalidation event
- **THEN** event payload 包含当前 node ID
- **AND** 接收节点可忽略或诊断来自自身的重复事件

#### Scenario: 系统信息诊断包含 node ID
- **WHEN** 查询系统信息
- **THEN** 响应包含当前 node ID
- **AND** 响应包含当前节点是否为 primary
