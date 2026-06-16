## ADDED Requirements

### Requirement: 插件运行时组合 delegate 不得静默伪装成功

系统 SHALL 将插件服务内部用于打破启动构造循环的 delegate 限定为最小组合接缝。delegate MUST 提供可诊断的绑定状态；当运行期写入、副作用发布、缓存刷新、依赖校验或认证事件回调在未绑定状态下被调用时，系统 MUST 返回明确错误，不得静默返回成功。只读投影方法在接口无法返回错误时 MAY fail-closed 或返回输入投影，但不得让调用方误以为副作用已经执行。

#### Scenario: 未绑定 runtime delegate 处理认证事件

- **WHEN** `RuntimeDelegate` 尚未绑定插件根服务
- **AND** 调用方触发登录成功、登录失败或登出回调
- **THEN** delegate 返回明确错误
- **AND** 不报告认证事件副作用已经成功执行

#### Scenario: 插件根服务构造后绑定 delegate

- **WHEN** 宿主启动构造插件根服务
- **THEN** 启动装配在插件根服务创建完成后显式绑定 runtime delegate
- **AND** 测试或审查可以确认运行期使用前 delegate 已处于绑定状态

### Requirement: 插件内部 cache 和升级 adapter 必须暴露缺失依赖

系统 SHALL 要求插件内部 cache notifier、dependency validator、source upgrade cache publisher 和 cache freshener 等窄 adapter 在依赖缺失时返回明确错误。生产构造路径 MUST 传入根插件服务或对应窄接口实例；adapter 不得因为 service 为 nil 而返回 nil。

#### Scenario: upgrade cache publisher 缺少根服务

- **WHEN** source upgrade 流程调用未绑定根服务的 cache publisher
- **THEN** publisher 返回明确错误
- **AND** 不发布插件运行时缓存已失效的假成功结果

#### Scenario: dependency validator 缺少根服务

- **WHEN** 生命周期或升级流程调用未绑定根服务的 dependency validator
- **THEN** validator 返回明确错误
- **AND** 不把依赖校验视为通过
