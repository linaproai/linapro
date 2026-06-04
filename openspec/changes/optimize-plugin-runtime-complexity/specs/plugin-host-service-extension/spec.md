## ADDED Requirements

### Requirement: WASM host service 生产依赖必须由启动期显式配置

系统 SHALL 由宿主启动期统一配置 WASM host service dispatcher 所需的 cache、lock、notify、storage config、plugin config、manifest、AI、organization、tenant 和其他 host service 运行期依赖。生产代码中的 WASM host service 包级变量 MUST NOT 默认调用 `New()` 创建 cache、config、session、notify、lock、plugin service 或 capability manager 等关键服务实例。缺失配置 MUST 以显式初始化错误或 host call internal error 暴露。

#### Scenario: 启动期配置共享 cache host service

- **WHEN** 宿主 HTTP runtime 初始化动态插件 WASM host service
- **THEN** `cache` host service 使用启动期已创建并选定共享后端的 `kvcache.Service`
- **AND** WASM host service 包不得保留可被生产路径使用的默认 `kvcache.New()` fallback

#### Scenario: storage host service 使用启动期配置服务

- **WHEN** 动态插件调用 `storage` host service
- **THEN** storage dispatcher 使用启动期注入的插件动态存储配置 reader
- **AND** 不通过包级默认 `config.New()` 读取独立配置服务图

#### Scenario: 测试 fixture 显式注入 host service 依赖

- **WHEN** 单元测试或集成测试执行 WASM host service dispatcher
- **THEN** 测试 fixture 显式调用配置入口注入 fake、memory 或共享测试服务
- **AND** 测试不得依赖生产包级默认服务实例

### Requirement: host service 配置入口必须覆盖完整依赖集

系统 SHALL 为 WASM host service 提供一个启动期可调用的完整配置入口，覆盖所有已发布 host service dispatcher 的必需运行期依赖。新增 host service dispatcher 时，系统 MUST 同步更新配置入口和缺失配置测试，确保新依赖不会绕过启动装配。

#### Scenario: 新增 host service 后配置覆盖测试失败

- **WHEN** 开发者新增一个 WASM host service dispatcher
- **AND** 未把该 dispatcher 的必需依赖纳入统一配置入口或覆盖测试
- **THEN** 自动化测试失败
- **AND** 失败信息指向缺失的 host service 配置项
