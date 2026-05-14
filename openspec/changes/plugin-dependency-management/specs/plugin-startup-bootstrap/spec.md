## ADDED Requirements

### Requirement: 启动自动启用必须解析并安装自动依赖

`BootstrapAutoEnable(ctx)` SHALL 对 `plugin.autoEnable` 中列出的插件执行依赖检查。对已发现、版本满足、未安装且声明为 `required: true`、`install: auto` 的依赖插件，启动自动启用必须在目标插件安装前按确定性拓扑顺序完成依赖安装。

#### Scenario: 自动启用目标插件前安装依赖
- **WHEN** `plugin.autoEnable` 包含插件 `x`
- **AND** `x` 声明自动安装硬依赖 `a`
- **AND** `a` 尚未安装
- **THEN** 启动引导先安装 `a`
- **AND** 启动引导再安装并启用 `x`

#### Scenario: 启动依赖版本不满足阻塞启动
- **WHEN** `plugin.autoEnable` 包含插件 `x`
- **AND** `x` 的硬依赖版本不满足
- **THEN** 宿主启动失败
- **AND** 错误包含目标插件、依赖插件和版本要求

### Requirement: 启动自动启用不得隐式启用依赖插件

启动自动启用流程 SHALL 只启用 `plugin.autoEnable` 中显式列出的插件。被依赖关系自动安装的插件不得因为作为依赖被安装而自动启用，除非该依赖插件自身也出现在 `plugin.autoEnable` 中。

#### Scenario: 依赖插件不在自动启用列表中
- **WHEN** 插件 `a` 被作为插件 `x` 的自动依赖安装
- **AND** `a` 不在 `plugin.autoEnable` 中
- **THEN** 启动引导只保证 `a` 已安装
- **AND** 启动引导不得启用 `a`

#### Scenario: 依赖插件也在自动启用列表中
- **WHEN** 插件 `a` 被作为插件 `x` 的自动依赖安装
- **AND** `a` 也在 `plugin.autoEnable` 中
- **THEN** 启动引导在依赖安装完成后确保 `a` 被启用

### Requirement: 集群模式下启动依赖安装必须遵守主节点副作用边界

集群模式下，启动自动启用触发的依赖安装 SHALL 遵守现有插件生命周期主节点边界。共享安装、菜单写入、发布切换和状态推进只能由主节点执行；从节点必须等待共享状态并刷新本地投影。

#### Scenario: 主节点安装自动依赖
- **WHEN** 集群模式下主节点执行 `BootstrapAutoEnable`
- **AND** 自动启用目标插件需要安装依赖插件
- **THEN** 主节点执行依赖插件安装副作用
- **AND** 主节点发布受影响插件的运行时修订或等价事件

#### Scenario: 从节点等待依赖安装结果
- **WHEN** 集群模式下从节点执行 `BootstrapAutoEnable`
- **AND** 自动启用目标插件依赖尚未在共享状态中完成安装
- **THEN** 从节点等待主节点收敛或等待窗口超时
- **AND** 从节点不得重复执行依赖安装 SQL 或共享状态写入
