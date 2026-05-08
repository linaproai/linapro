## ADDED Requirements

### Requirement: 启动自动启用必须同步生命周期写入后的启动快照

系统 SHALL 在一次宿主启动编排内保持插件生命周期写入与共享启动快照一致。`plugin.autoEnable` 对源码插件执行按需安装后，同一启动上下文中的后续启用、状态检查、路由接线和预热阶段必须读取到更新后的 `installed`、`status`、`desiredState` 和 `currentState` 投影。

#### Scenario: 源码插件自动安装后立即启用

- **WHEN** 宿主启动上下文已经携带插件治理启动快照
- **AND** `plugin.autoEnable` 包含一个尚未安装的源码插件
- **THEN** 自动安装完成后必须同步更新当前启动快照中的插件 registry 投影
- **AND** 后续启用检查必须将该插件识别为已安装
- **AND** 宿主启动不得因该插件报出 `Plugin is not installed`

#### Scenario: 已安装源码插件自动启用

- **WHEN** 宿主启动上下文已经携带插件治理启动快照
- **AND** `plugin.autoEnable` 包含一个已安装但未启用的源码插件
- **THEN** 启用阶段必须复用当前启动快照中的已安装状态
- **AND** 启用完成后必须同步更新当前启动快照中的启用状态投影
