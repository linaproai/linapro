## ADDED Requirements

### Requirement: WASM host service runtime 必须由实例持有

系统 SHALL 让动态插件 WASM host service 分发器通过显式 runtime 实例读取运行期依赖。WASM host service 的 domain capability directory、插件配置 factory、宿主配置 service、manifest service factory 和其他共享宿主能力 MUST 来自启动期构造并注入的实例，不得通过包级`Configure*`函数、包级`atomic.Pointer`快照或包级默认实例作为生产调用事实源。

#### Scenario: 宿主启动配置 WASM host service

- **WHEN** 宿主启动并初始化动态插件 runtime
- **THEN** 启动路径创建 WASM host service runtime 实例
- **AND** 该实例显式接收共享 capability directory、config factory、host config service 和 manifest factory
- **AND** 生产启动后不调用`wasm.Configure*`配置包级状态

#### Scenario: 动态插件执行 host call

- **WHEN** 动态插件通过统一 host service 协议执行 host call
- **THEN** 分发器从当前 runtime 实例读取依赖
- **AND** 不读取包级 snapshot
- **AND** 多个测试或多 service 实例可以持有彼此隔离的 WASM runtime

#### Scenario: 缺少 WASM runtime 依赖

- **WHEN** 启动期未提供 WASM host service 必需依赖
- **THEN** 构造函数返回错误
- **AND** 错误在启动装配阶段暴露
- **AND** 不允许请求路径首次 host call 时才因包级配置缺失失败
