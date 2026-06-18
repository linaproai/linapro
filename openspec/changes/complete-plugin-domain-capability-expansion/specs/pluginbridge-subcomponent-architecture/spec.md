## ADDED Requirements

### Requirement: 剩余动态领域方法必须由 catalog 统一覆盖
系统 SHALL 要求本变更新增的动态普通领域、插件私有资源和 runtime state 方法全部声明在`pkg/plugin/pluginbridge/protocol/hostservices`公开 catalog 中。catalog MUST 覆盖 service、method、capability、资源类型、payload 形态、guest client 发布状态和 host dispatcher 发布状态。

#### Scenario: 新增动态搜索方法
- **WHEN** 系统发布`files.search`、`jobs.search`或`dict.values.list`
- **THEN** catalog 必须声明 service、method、资源类型和 payload 类型
- **AND** guest client 与 host dispatcher 覆盖测试必须能发现遗漏

#### Scenario: 新增 runtime state 多键方法
- **WHEN** 系统发布`state.get_many`、`state.set_many`或`state.delete_many`
- **THEN** catalog 必须声明这些方法属于动态 runtime 服务
- **AND** dispatcher 必须继续校验插件和租户作用域

### Requirement: 普通领域方法必须使用 JSON envelope 或明确复用既有 codec
系统 SHALL 要求新增普通领域方法默认使用统一 JSON request/response envelope。storage、data、manifest、runtime state、notify 或`AI`若已有专用 payload codec，MUST 复用既有 codec 体系；新增专用 codec 必须先在 OpenSpec 设计中说明性能或 wire 稳定性理由。

#### Scenario: 普通领域 JSON round trip
- **WHEN** 动态插件调用`tenants.search`或`plugins.search`
- **THEN** 请求和响应通过统一 JSON envelope 编码解码
- **AND** protocol 或 typed client 测试覆盖请求响应映射

#### Scenario: 资源型能力复用既有 codec
- **WHEN** 动态插件调用`storage.stat.batch`或`data.batch_get`
- **THEN** 系统复用对应资源型 host service codec 或统一 JSON envelope 中已声明的稳定 payload
- **AND** 不新增未登记的 per-domain codec 文件

### Requirement: 动态 host service 同步测试必须覆盖所有新增方法
系统 SHALL 更新协议覆盖测试，使新增 method 缺少 public alias、guest client、dispatcher 注册、README 表格或非 WASI stub 时测试失败。

#### Scenario: dispatcher 缺少新增方法
- **WHEN** catalog 声明一个本变更新增方法
- **AND** `internal/service/plugin/internal/wasm`没有注册处理该 service/method
- **THEN** 覆盖测试失败

#### Scenario: README 缺少新增动态方法
- **WHEN** catalog 声明一个可由动态插件声明的 host service 方法
- **AND** `apps/lina-core/pkg/plugin/README.md`或`README.zh-CN.md`缺少对应方法
- **THEN** 文档同步检查或审查必须失败
