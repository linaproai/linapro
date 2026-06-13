## ADDED Requirements

### Requirement: WASM host service 公共 helper 必须归属公共层

系统 SHALL 将跨领域复用的 WASM host service helper 归属到`wasm`公共 host service 文件或`hostservicedispatch`公共层。具体领域文件 MUST 只承载该领域 service/method 的 transport 适配、授权前置和领域能力调用，不得承载`CapabilityContext`构造、统一 envelope 辅助或 registry 公共响应逻辑。

#### Scenario: 公共 capability context helper 不在用户领域文件

- **WHEN** 静态检索`apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_users.go`
- **THEN** 文件中不得定义`capabilityContextForHostCall`
- **AND** 用户领域 handler 仍通过同包公共 helper 构造`CapabilityContext`

#### Scenario: 新增普通领域 host service 不修改统一入口分发分支

- **WHEN** 开发者新增一个普通领域 host service 或 method
- **THEN** 开发者只需要新增或修改领域文件中的 handler 实现
- **AND** 开发者通过显式 registry 注册条目接入该 service/method
- **AND** `wasm_host_service.go`不得新增按 service family 直接分发的 switch 分支

#### Scenario: 公共 helper 不扩大运行期依赖来源

- **WHEN** 公共 helper 被迁移到公共层
- **THEN** helper 继续使用调用路径已有的`hostCallContext`和启动期注入能力
- **AND** 不得通过包级默认实例、`init()`或临时`New()`创建`auth`、`session`、`i18n`、`cache`或插件 runtime 依赖
