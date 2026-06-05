## ADDED Requirements

### Requirement: 动态插件必须通过 ai.text.generate 调用文本 AI

系统 SHALL 在动态插件宿主服务体系中新增 `ai` service family，首期仅开放 `text.generate` 方法。动态插件 MUST 通过 `hostServices` 声明申请该能力，并由宿主授权快照确认后才能调用。

#### Scenario: 插件声明文本 AI 宿主服务

- **WHEN** 动态插件在 `plugin.yaml` 中声明 `service: ai` 和 `methods: [text.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为文本 `AI` 调用权限申请
- **AND** 声明 MUST 支持 `resources` 中以 `purpose:<name>` 表达调用用途
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限
- **AND** 运行时 MUST 将该方法映射为 `capabilityType=text` 与 `capabilityMethod=generate`

#### Scenario: 未声明插件调用被拒绝

- **WHEN** 动态插件未声明或未获确认 `ai.text.generate` 授权却发起文本 `AI` 调用
- **THEN** 宿主 MUST 在执行 `framework.ai.text.v1` 或渠道调用前拒绝
- **AND** 宿主 MUST 返回结构化授权错误

### Requirement: ai.text.generate 必须受 service、method 和资源授权约束

系统 SHALL 对每一次 `ai.text.generate` 调用同时校验 service、method、`purpose` 资源、调用来源和策略属性。任一校验失败时，宿主 MUST 拒绝调用，并且 MUST NOT 执行外部渠道请求。

#### Scenario: 授权 purpose 调用成功进入文本能力

- **WHEN** 动态插件已获 `ai.text.generate` 授权
- **AND** 调用请求的 `purpose` 与授权快照中的 `purpose:<name>` 匹配
- **THEN** host service handler MUST 将请求转换为 `framework.ai.text.v1` 的 `GenerateText` 请求
- **AND** 调用 MUST 复用同一个文本能力消费 service 和 provider 可用性语义
- **AND** 档位解析 MUST 使用 `text.generate + tier` 作为能力范围

#### Scenario: 未授权 purpose 被拒绝

- **WHEN** 动态插件请求未在授权快照中确认的 `purpose`
- **THEN** 宿主 MUST 拒绝调用
- **AND** 宿主 MUST NOT 解析档位绑定、读取渠道密钥或执行渠道 API 请求

#### Scenario: 策略属性限制输出规模

- **WHEN** 授权资源声明包含 `maxOutputTokens` 等策略属性
- **THEN** 宿主 MUST 在转发到 `framework.ai.text.v1` 前校验请求参数不超过授权范围
- **AND** 超出范围的请求 MUST 被拒绝或按显式策略收敛，且不得静默扩大权限

### Requirement: ai.text.generate 调用契约必须支持文本参数和 thinkingEffort

系统 SHALL 为动态插件 `ai.text.generate` 定义 DTO 化调用载荷。载荷 MUST 支持 `purpose`、`tier`、`messages`、`maxOutputTokens`、`temperature`、可选 `thinkingEffort` 和短字符串 `metadata`，并与 `framework.ai.text.v1` 保持语义一致。

#### Scenario: 动态插件传入 thinkingEffort

- **WHEN** 动态插件调用 `ai.text.generate` 并传入 `thinkingEffort`
- **THEN** 宿主 MUST 校验该值属于 `low`、`medium`、`high`、`xhigh`、`max`
- **AND** 授权通过后 MUST 将该值传递给 `framework.ai.text.v1`
- **AND** 模型不支持该 effort 时 MUST 返回文本能力的结构化业务错误

#### Scenario: 动态插件不传 thinkingEffort

- **WHEN** 动态插件调用 `ai.text.generate` 未传 `thinkingEffort`
- **THEN** 宿主 MUST 保持字段为空并由档位默认值或模型默认行为决定实际 effort
- **AND** 宿主 MUST NOT 在 host service 层硬编码某个渠道专有 effort 值

#### Scenario: 动态插件 metadata 有界

- **WHEN** 动态插件在 `ai.text.generate` 中传入 `metadata`
- **THEN** 宿主 MUST 只接受短字符串键值用于调用来源、业务请求 ID 或审计关联
- **AND** 宿主 MUST 拒绝或截断超出契约边界的大段输入

### Requirement: ai.text.generate 必须记录最小审计信息

系统 SHALL 对动态插件发起的 `ai.text.generate` 调用记录最小宿主服务审计和智能中心调用日志。审计 MUST 支持诊断授权、耗时、状态和来源插件，但 MUST NOT 保存完整输入、完整输出、隐藏思考内容或密钥。

#### Scenario: 成功调用记录来源

- **WHEN** 动态插件通过 `ai.text.generate` 成功生成文本
- **THEN** 宿主服务审计 MUST 记录 `pluginId`、service、method、purpose 摘要、结果状态和耗时
- **AND** 智能中心调用日志 MUST 记录 `sourcePluginId`、`purpose`、档位、渠道模型投影、`thinkingEffort`、token 用量和耗时

#### Scenario: 失败调用记录脱敏错误

- **WHEN** 动态插件调用 `ai.text.generate` 失败
- **THEN** 宿主服务审计和智能中心调用日志 MUST 记录失败状态与脱敏错误摘要
- **AND** 审计和日志 MUST NOT 包含完整 `messages`、API key、认证头或渠道响应原文

### Requirement: ai service 必须拒绝首期未开放的方法

系统 SHALL 将 `ai` service family 设计为可扩展宿主服务族，但首期 MUST 只开放 `text.generate`。图片、音频、向量、重排、工具调用、流式输出或其他 `AI` 方法在未有正式规范和授权模型前 MUST 被拒绝。

#### Scenario: 未开放 image 方法被拒绝

- **WHEN** 动态插件声明或调用 `ai.image.generate`
- **THEN** 构建器、宿主清单校验或运行时 handler MUST 拒绝该方法
- **AND** 错误 MUST 指出该 `AI` host service method 当前不受支持

#### Scenario: 后续能力独立扩展

- **WHEN** 系统后续新增图片、音频或向量能力
- **THEN** 新方法 MUST 明确定义 `capabilityType`、`capabilityMethod`、资源授权、请求响应 DTO、审计字段和与框架 capability 的适配关系
- **AND** 新方法 MUST NOT 改变现有 `ai.text.generate` 的授权和同步文本响应语义
