## ADDED Requirements

### Requirement: AI 命名空间必须支持跨子能力方法状态批量读取
系统 SHALL 在`AI`命名空间提供跨子能力方法状态批量读取能力，并可动态发布为`ai.methods.status.batch_get`或等价冻结名称。响应 MUST 只包含能力、方法、可用性、禁用原因和结构化 unavailable 信息，不得暴露 provider 配置、密钥、模型映射或内部路由策略。

#### Scenario: 批量读取 AI 方法状态
- **WHEN** 插件请求文本、图像、音频或视觉等多个`AI`方法状态
- **THEN** 系统批量返回每个方法的可用性状态
- **AND** provider 未启用或方法不可用时返回结构化状态而不是泄露 provider 内部配置

#### Scenario: AI provider 配置不暴露
- **WHEN** 插件读取`AI`方法状态
- **THEN** 响应不得包含 API key、供应商私有 endpoint、模型路由表或内部 provider 优先级
