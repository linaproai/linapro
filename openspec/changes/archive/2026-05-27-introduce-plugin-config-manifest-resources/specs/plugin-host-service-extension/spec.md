## MODIFIED Requirements

### Requirement: 宿主服务访问同时受宿主服务声明推导的能力分类和资源授权约束

系统 SHALL 对每一次宿主服务调用同时执行粗粒度 capability 校验和细粒度资源授权校验，但 capability 集合必须由`hostServices`声明自动推导，而不是要求插件作者在`plugin.yaml`中重复维护第二份`capabilities`列表；任一层不满足都必须拒绝调用。只读读取型服务的`methods` MUST 表达真实 host service 调用动作，SDK typed helper 不得作为独立授权方法进入声明或运行时快照。

#### Scenario: 插件声明宿主服务策略

- **WHEN** 开发者在动态插件清单中声明`hostServices`
- **THEN** 构建器校验 service、method、资源声明（如低优先级服务的逻辑`resourceRef`、`storage.resources.paths`、URL 模式、`resources.tables`、宿主公开配置 key 或 manifest 资源路径）和策略参数是否合法
- **AND** 宿主根据这些 methods 自动推导内部 capability 分类快照
- **AND** 将归一化后的宿主服务策略写入运行时产物
- **AND** 宿主装载产物后恢复为当前 release 的服务授权快照

#### Scenario: 缺少授权的宿主服务调用被拒绝

- **WHEN** 插件调用未声明的 service、method 或未授权的资源标识
- **THEN** 宿主返回显式拒绝错误
- **AND** 宿主不执行任何目标宿主服务逻辑

#### Scenario: typed helper 不作为授权方法声明

- **WHEN** 动态插件需要读取插件配置并使用 guest SDK 的`String`、`Bool`、`Int`、`Duration`或`Scan`helper
- **THEN** `plugin.yaml`只声明`service: config`和`methods: [get]`
- **AND** 宿主授权快照只记录`config.get`
- **AND** guest SDK helper 在插件侧或共享适配层基于`get`结果完成类型转换

### Requirement: 资源型宿主服务声明属于权限申请而非自动授权

系统 SHALL 将所有资源型 hostServices 声明解释为权限申请清单，而不是插件在运行时自动拥有的资源访问权；其中`storage`当前使用`resources.paths`，`network`使用 URL 模式，`data`当前使用`resources.tables`，`hostConfig`使用`resources.keys`，`manifest`使用`resources.paths`，其余低优先级服务（`cache`、`lock`、`notify`）继续沿用逻辑`resourceRef`规划，并分别表示缓存命名空间、逻辑锁名和通知通道标识。

#### Scenario: 清单声明宿主服务资源申请

- **WHEN** 开发者在动态插件清单中声明`storage.resources.paths`、`network`的 URL 模式、`data.resources.tables`、`hostConfig.resources.keys`、`manifest.resources.paths`，或声明`cache`、`lock`、`notify`等低优先级服务的逻辑`resourceRef`
- **THEN** 这些声明表示插件申请对应宿主资源权限
- **AND** 声明本身不得直接视为运行时已授权结果

#### Scenario: 运行时只认宿主确认后的授权快照

- **WHEN** 动态插件在运行时调用一个资源型宿主服务
- **THEN** 宿主仅按安装或启用阶段确认后的授权快照执行校验
- **AND** 未被宿主确认的声明必须被拒绝
- **AND** 插件不能通过“先声明后直接调用”的方式跳过宿主授权确认

## ADDED Requirements

### Requirement: 配置和清单类只读宿主服务必须使用 get 方法声明

系统 SHALL 要求动态插件在`plugin.yaml`中为`config`、`hostConfig`和`manifest`等只读读取型宿主服务仅声明`methods: [get]`。宿主 MUST 拒绝这些服务上的写入、保存、热重载、完整快照、typed helper 或未知方法声明，并在运行时对每次`get`调用执行 service、method 和资源范围校验。

#### Scenario: 配置服务只声明 get

- **WHEN** 动态插件声明`service: config`且`methods: [get]`
- **THEN** 宿主接受声明并派生插件作用域配置读取能力
- **AND** 授权快照不包含`exists`、`string`、`bool`、`int`、`duration`或`scan`

#### Scenario: 宿主公开配置声明 key 白名单

- **WHEN** 动态插件声明`service: hostConfig`、`methods: [get]`和`resources.keys`
- **THEN** 宿主仅允许该插件读取声明且经宿主确认的公开宿主配置键
- **AND** 未声明或未确认的键在运行时被拒绝

#### Scenario: Manifest 声明资源路径白名单

- **WHEN** 动态插件声明`service: manifest`、`methods: [get]`和`resources.paths`
- **THEN** 宿主仅允许该插件读取声明且经宿主确认的当前插件`manifest/`资源路径
- **AND** 跨插件路径、绝对路径、URL 和路径穿越必须被拒绝

#### Scenario: 只读服务声明写入方法被拒绝

- **WHEN** 动态插件在`config`、`hostConfig`或`manifest`服务中声明`set`、`save`、`reload`或其他非`get`方法
- **THEN** 构建器或宿主清单校验拒绝该声明
- **AND** 运行时不得为该插件派生对应宿主服务能力
