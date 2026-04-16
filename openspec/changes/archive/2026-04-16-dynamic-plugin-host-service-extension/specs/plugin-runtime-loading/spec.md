## ADDED Requirements

### Requirement: 动态插件运行时产物携带宿主服务治理快照

系统 SHALL 让动态插件运行时产物携带结构化宿主服务声明、方法授权和资源申请快照，供宿主在装载时恢复当前 release 的宿主服务治理信息；宿主内部 capability 分类快照必须基于该`hostServices`快照自动推导，而不是由 guest 额外嵌入第二份作者侧声明。

#### Scenario: 构建器写入宿主服务快照

- **WHEN** 构建器生成动态插件运行时产物
- **THEN** 构建器将归一化后的`hostServices`声明写入专用自定义节
- **AND** 不再写入作者侧顶层`capabilities`自定义节
- **AND** 对未知 service、method 或非法策略参数直接报错

#### Scenario: 宿主恢复宿主服务快照

- **WHEN** 宿主装载一个动态插件运行时产物
- **THEN** 宿主恢复结构化宿主服务策略并据此推导能力分类集合
- **AND** 将其挂到当前 active release 的运行时 manifest
- **AND** 缺失或损坏的宿主服务快照会阻止对应宿主服务进入可执行状态

### Requirement: 动态插件运行时按统一宿主服务协议执行 Host Call

系统 SHALL 根据动态插件声明的宿主服务协议版本，通过统一宿主服务分发器执行 Host Call，不再为历史探索性实现保留平行公开协议。

#### Scenario: 运行时统一走宿主服务分发器

- **WHEN** 一个动态插件声明了结构化宿主服务协议
- **THEN** 宿主通过统一宿主服务分发器处理该调用
- **AND** 宿主不得再为同类新增能力暴露新的专用 opcode
