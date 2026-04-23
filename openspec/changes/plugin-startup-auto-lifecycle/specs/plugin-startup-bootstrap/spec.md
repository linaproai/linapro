## ADDED Requirements

### Requirement: 宿主必须支持插件启动策略静态配置
系统 SHALL 在宿主静态配置中提供 `plugin.startup` 配置段，用于按 `pluginId` 声明插件启动期目标状态、是否必需、等待策略以及动态插件授权快照。`desiredState` MUST 只允许 `manual`、`installed`、`enabled` 三种值。

#### Scenario: 解析有效启动策略
- **WHEN** 宿主启动并读取包含多个 `plugin.startup.policies` 的有效配置
- **THEN** 宿主成功构建按 `pluginId` 索引的启动策略映射
- **AND** 每条策略都包含规范化后的目标状态与 `required` 语义

#### Scenario: 拒绝非法目标状态配置
- **WHEN** 某条启动策略声明了 `manual` / `installed` / `enabled` 之外的 `desiredState`
- **THEN** 宿主 MUST 拒绝继续启动
- **AND** 错误信息 MUST 明确指出对应的 `pluginId` 与非法字段值

### Requirement: 宿主必须在插件接线之前执行启动期 bootstrap
系统 SHALL 在插件 HTTP 路由注册、插件 cron 接线和动态前端 bundle 预热之前，先按启动策略推进命中插件的生命周期状态。

#### Scenario: 启动前推进源码插件到启用态
- **WHEN** 已发现的源码插件命中启动策略且 `desiredState=enabled`
- **THEN** 宿主在插件路由与插件 cron 注册前先完成该源码插件的安装与启用
- **AND** 随后的 enabled snapshot 读取结果中该插件处于启用状态

#### Scenario: 未配置策略的插件保持发现态或现有稳定态
- **WHEN** 插件被宿主发现，但没有匹配任何启动策略
- **THEN** 宿主仅执行常规 manifest 同步与注册表刷新
- **AND** 宿主 MUST NOT 因启动 bootstrap 而自动安装或自动启用该插件

### Requirement: 启动策略必须采用最低目标状态语义
系统 SHALL 将 `desiredState` 解释为启动期最低目标状态，而不是破坏性的精确最终状态；策略移除或降低目标状态时，宿主 MUST NOT 自动执行禁用或卸载。

#### Scenario: 已启用插件命中 installed 策略
- **WHEN** 某插件当前已处于启用状态，而启动策略声明 `desiredState=installed`
- **THEN** 宿主保持该插件启用状态不变
- **AND** 宿主 MUST NOT 将其反向收敛为仅安装状态

#### Scenario: 策略移除不触发破坏性动作
- **WHEN** 先前依赖启动策略推进过的插件在本次启动中不再匹配任何策略
- **THEN** 宿主本次启动不会自动对其执行禁用或卸载
- **AND** 后续若要降级该插件，仍需管理员显式执行生命周期操作

### Requirement: 启动期 bootstrap 必须支持必需插件与可选插件的不同失败策略
系统 SHALL 允许按策略声明插件是否为必需，并在启动期根据 `required` 与等待超时结果决定 fail-fast 或告警降级。

#### Scenario: 必需插件收敛失败阻止宿主启动
- **WHEN** 启动策略标记某插件 `required=true`，且该插件缺失、bootstrap 失败或在等待窗口内未达到目标状态
- **THEN** 宿主 MUST 终止启动流程
- **AND** 返回的错误信息 MUST 包含失败插件标识与失败阶段

#### Scenario: 可选插件收敛失败仅告警降级
- **WHEN** 启动策略标记某插件 `required=false`，且该插件缺失、bootstrap 失败或超时
- **THEN** 宿主记录结构化告警并继续启动
- **AND** 其他插件的启动 bootstrap 流程继续执行

### Requirement: 集群模式下启动期 bootstrap 必须区分共享生命周期动作与本地收敛
系统 SHALL 在集群模式下仅允许主节点执行插件共享生命周期动作（如安装 SQL、菜单写入、release 切换与共享状态推进），从节点只等待共享状态结果并刷新本地投影。

#### Scenario: 主节点执行共享插件动作
- **WHEN** 集群模式下某插件命中启动策略且需要执行安装或启用推进
- **THEN** 只有主节点执行该插件的共享安装、启用或 reconcile 动作
- **AND** 从节点 MUST NOT 重复执行同一插件的共享副作用步骤

#### Scenario: 从节点等待共享状态收敛后刷新本地视图
- **WHEN** 集群模式下从节点启动并发现某插件由启动策略要求达到 `installed` 或 `enabled`
- **THEN** 从节点等待主节点写入共享稳定状态或等待窗口超时
- **AND** 随后基于共享结果刷新本地 enabled snapshot 与运行时投影

### Requirement: 启动期启用动态插件时必须支持授权快照
系统 SHALL 允许动态插件启动策略携带宿主服务授权快照；当目标状态为 `enabled` 且插件声明了受治理 host services 时，宿主必须先处理授权，再决定启用、降级安装或 fail-fast。

#### Scenario: 启动策略提供授权快照并成功启用动态插件
- **WHEN** 动态插件命中 `desiredState=enabled` 的启动策略，且策略中提供了合法的 `authorization`
- **THEN** 宿主先持久化该 release 的授权快照
- **AND** 再把动态插件推进到启用状态

#### Scenario: 非必需动态插件缺少授权时降级为仅安装
- **WHEN** 动态插件命中 `desiredState=enabled` 的启动策略、声明了受治理 host services，但未提供授权且 `required=false`
- **THEN** 宿主最多把该插件推进到 `installed`
- **AND** 该插件保留待授权状态并记录明确告警

#### Scenario: 必需动态插件缺少授权时启动失败
- **WHEN** 动态插件命中 `desiredState=enabled` 的启动策略、声明了受治理 host services，但未提供授权且 `required=true`
- **THEN** 宿主 MUST 终止启动
- **AND** 错误信息 MUST 明确指出缺少授权快照导致无法达到目标状态
