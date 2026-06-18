## ADDED Requirements

### Requirement: 阶段一动态领域方法必须由 host service catalog 覆盖

系统 SHALL 要求`expand-plugin-domain-capabilities`新增的动态普通领域方法全部声明在`pkg/plugin/pluginbridge/protocol/hostservices`公开 catalog 中。catalog MUST 覆盖 service、method、capability、资源类型、payload 形态、guest client 发布状态和 host dispatcher 发布状态；内部 descriptor 不得维护第二份手写方法表。

#### Scenario: 新增动态当前用户方法

- **WHEN** 系统发布`users.current.get`
- **THEN** catalog 必须声明该方法的 service、method、资源类型和 JSON payload 形态
- **AND** guest client 与 host dispatcher 覆盖测试必须能发现遗漏

#### Scenario: 新增动态字典值校验方法

- **WHEN** 系统发布`values.visible.ensure`
- **THEN** catalog 必须声明该方法归属`dict`服务和`host:dict`能力资源
- **AND** dispatcher 必须只做授权、解码、委托和错误映射，不重新实现字典业务规则

### Requirement: 阶段一普通领域动态方法必须使用 JSON envelope

系统 SHALL 要求阶段一新增的`users.current.get`、`users.resolve.batch`、`permissions.batch_has`、`values.visible.ensure`和`sessions.current.get`使用普通领域 JSON request/response envelope。除非 OpenSpec 另行声明性能或 wire 稳定性例外，系统 MUST NOT 为这些方法新增专用 per-domain protowire codec。

#### Scenario: JSON envelope round trip

- **WHEN** 动态插件 guest client 调用阶段一新增普通领域方法
- **THEN** 请求和响应通过统一 JSON envelope 编码解码
- **AND** protocol 测试必须覆盖新增请求响应的 round trip 或等价 typed client 映射

#### Scenario: 不新增专用 codec

- **WHEN** 开发者实现阶段一动态普通领域方法
- **THEN** 不得新增`protocol_hostservice_users_current_codec.go`或同类 per-domain codec 文件
- **AND** 特殊 codec 例外必须先写入 OpenSpec 设计
