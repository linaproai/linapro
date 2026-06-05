## ADDED Requirements

### Requirement: 动态插件 guest SDK 必须通过 AI 命名空间调用文本 AI

系统 SHALL 在动态插件 guest 侧通过 `AI().Text()` 暴露文本 `AI` 能力。guest SDK 的 `AI().Text().GenerateText(...)` MUST 继续使用既有 `ai.text.generate` host service 协议，并保持 `host:ai:text`、`purpose:<name>` 和策略属性授权语义不变。

#### Scenario: 动态插件通过 AI 命名空间生成文本

- **WHEN** 动态插件需要调用文本 `AI` 生成能力
- **THEN** guest 代码 MUST 通过 `guest.Default().AI().Text().GenerateText(...)` 或等价能力目录入口发起调用
- **AND** guest SDK MUST NOT 继续要求调用方使用根目录 `AIText()` 方法

#### Scenario: guest AI Text 调用进入既有 host service

- **WHEN** guest SDK 执行 `AI().Text().GenerateText(...)`
- **THEN** SDK MUST 构造既有 `service: ai`、`method: text.generate` host service 调用
- **AND** 请求资源 MUST 继续使用 `purpose:<name>` 表达授权用途
- **AND** 宿主 MUST 在执行文本能力或渠道调用前完成 service、method、资源和策略属性校验

#### Scenario: 动态插件协议不因 Go 入口重构改变

- **WHEN** 系统将 guest 侧调用入口从 `AIText()` 重构为 `AI().Text()`
- **THEN** 动态插件 `plugin.yaml` 中的 `hostServices` 声明格式 MUST 保持 `service: ai` 和 `methods: [text.generate]`
- **AND** `host:ai:text` 的能力分类、`maxOutputTokens` 等资源策略和脱敏审计语义 MUST 保持不变

#### Scenario: 未开放的 AI 子能力仍被拒绝

- **WHEN** 动态插件通过 `AI()` 命名空间尝试调用尚未规范化的图片、向量、音频或其他 `AI` 子能力
- **THEN** guest SDK 或宿主 MUST 返回不支持错误
- **AND** 宿主 MUST NOT 因存在 `AI()` 聚合入口而自动授予未声明子能力
