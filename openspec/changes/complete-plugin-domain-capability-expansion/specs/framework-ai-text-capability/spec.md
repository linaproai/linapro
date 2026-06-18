## ADDED Requirements

### Requirement: 文本 AI 能力必须提供方法级状态
系统 SHALL 为文本`AI`能力补齐`Text().MethodStatus`或等价方法级状态读取，并可动态发布为`text.method_status.get`或等价冻结名称。该方法 MUST 与其他`AI`子能力状态语义一致。

#### Scenario: 读取文本生成方法状态
- **WHEN** 插件读取文本生成方法状态
- **THEN** 系统返回该方法是否可用、不可用原因和稳定状态码
- **AND** 不返回 provider 私有模型配置

#### Scenario: 文本 provider 不可用
- **WHEN** 文本 provider 未启用或当前 purpose 不可用
- **THEN** 系统返回结构化 unavailable 状态
- **AND** 插件可以据此降级而不触发 provider 内部错误
