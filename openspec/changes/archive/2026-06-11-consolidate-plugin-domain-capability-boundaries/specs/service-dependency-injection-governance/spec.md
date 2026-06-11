## ADDED Requirements

### Requirement: 动态领域能力配置必须复用启动期共享服务目录

系统 SHALL 要求动态插件普通领域能力配置复用启动期构造的同一个`capability.Services`目录。`ConfigureWasmHostServices` MUST 逐项接收并传递启动期共享依赖，MUST NOT 为普通领域能力创建新的服务图、领域专用全局目录、通用 service locator 或聚合依赖结构体。

#### Scenario: ConfigureWasmHostServices 配置普通领域能力

- **WHEN** 宿主启动配置动态插件`WASM host service`
- **THEN** 普通领域能力通过启动期共享的`capability.Services`实例一次性注入
- **AND** `WASM`分发层不得为`AI`、`User`、`Org`、`Tenant`或其他普通领域维护第二个共享实例来源

#### Scenario: 缺失领域能力目录

- **WHEN** `ConfigureWasmHostServices`收到`nil`领域能力目录
- **THEN** 配置入口必须返回错误
- **AND** 不得用包级默认实例、空实现或临时`New()`补齐运行期依赖

#### Scenario: 测试配置动态领域能力

- **WHEN** 单元测试需要验证动态领域 host service 分发
- **THEN** 测试必须构造自包含的`capability.Services`替身并调用`ConfigureDomainHostServices`
- **AND** 涉及包级状态时必须保存原值并通过`t.Cleanup`恢复
