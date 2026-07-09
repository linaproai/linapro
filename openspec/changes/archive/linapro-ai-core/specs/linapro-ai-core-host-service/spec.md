## ADDED Requirements

### Requirement: 动态插件必须通过 ai service 调用文本和多模态 AI 能力

系统 SHALL 在动态插件宿主服务体系中提供 `ai` service family，首期开放 `text.generate` 方法，并支持图片、向量、音频、视觉、文档、安全审核和视频等多模态方法声明。动态插件 MUST 通过 `hostServices` 声明申请该能力，并由宿主授权快照确认后才能调用。

#### Scenario: 插件声明文本 AI 宿主服务

- **WHEN** 动态插件在 `plugin.yaml` 中声明 `service: ai` 和 `methods: [text.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为文本 `AI` 调用权限申请
- **AND** 声明 MUST 支持 `resources` 中以 `purpose:<name>` 表达调用用途
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限
- **AND** 运行时 MUST 将该方法映射为 `capabilityType=text` 与 `capabilityMethod=generate`

#### Scenario: 插件声明图片生成能力

- **WHEN** 动态插件在 `plugin.yaml` 中声明 `service: ai` 和 `methods: [image.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为 `image.generate` 权限申请
- **AND** 运行时 MUST 将方法映射为 `capabilityType=image` 和 `capabilityMethod=generate`
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限

#### Scenario: 插件声明音频能力

- **WHEN** 动态插件声明 `audio.transcribe` 或 `audio.synthesize`
- **THEN** 宿主 MUST 分别识别为不同 host service 方法
- **AND** 两个方法 MUST 使用独立 payload 契约、资源策略和授权分类

#### Scenario: 未声明插件调用被拒绝

- **WHEN** 动态插件未声明或未获确认 `ai.text.generate` 授权却发起文本 `AI` 调用
- **THEN** 宿主 MUST 在执行 `framework.ai.text.v1` 或渠道调用前拒绝
- **AND** 宿主 MUST 返回结构化授权错误

#### Scenario: computer act 声明被拒绝

- **WHEN** 动态插件声明 `computer.act`、`ui.operate` 或等价 UI 控制方法
- **THEN** 清单校验或运行时 MUST 拒绝该声明
- **AND** 错误 MUST 表明该方法不属于本轮 `ai` host service 支持范围

### Requirement: ai service 必须受 service、method 和资源授权约束

系统 SHALL 对每一次 `ai` host service 调用同时校验 service、method、resource、`purpose`、调用来源和策略属性。任一校验失败时，宿主 MUST 拒绝调用，并且 MUST NOT 执行外部渠道请求或读取渠道密钥。

#### Scenario: 授权 purpose 调用成功进入文本能力

- **WHEN** 动态插件已获 `ai.text.generate` 授权
- **AND** 调用请求的 `purpose` 与授权快照中的 `purpose:<name>` 匹配
- **THEN** host service handler MUST 将请求转换为 `framework.ai.text.v1` 的 `GenerateText` 请求
- **AND** 调用 MUST 复用同一个文本能力消费 service 和 provider 可用性语义
- **AND** 档位解析 MUST 使用 `text.generate + tier` 作为能力范围

#### Scenario: 授权资源匹配后多模态调用成功

- **WHEN** 动态插件已获 `image.generate` 授权
- **AND** 请求的 `purpose`、输入 mime 类型、最大输入资产数和最大输出数量均满足授权资源策略
- **THEN** host service handler MUST 将请求转换为 `AI().Image().Generate(...)` 或等价能力调用
- **AND** 调用 MUST 复用对应子能力的 provider 可用性和错误语义

#### Scenario: 未授权 purpose 被拒绝

- **WHEN** 动态插件请求未在授权快照中确认的 `purpose`
- **THEN** 宿主 MUST 拒绝调用
- **AND** 宿主 MUST NOT 解析档位绑定、读取渠道 endpoint、secret 引用或执行渠道 API 请求

#### Scenario: 策略属性限制输出规模

- **WHEN** 授权资源声明包含 `maxOutputTokens` 等策略属性
- **THEN** 宿主 MUST 在转发到 `framework.ai.text.v1` 或对应子能力前校验请求参数不超过授权范围
- **AND** 超出范围的请求 MUST 被拒绝或按显式策略收敛，且不得静默扩大权限

#### Scenario: payload 超限被拒绝

- **WHEN** 动态插件上传或引用的输入资产数量、字节数、mime 类型、输出数量或 token 上限超过授权策略
- **THEN** 宿主 MUST 拒绝调用或按显式策略收敛
- **AND** 宿主 MUST NOT 静默扩大插件授权范围

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

### Requirement: AI host service 大对象 payload 必须使用资产引用

系统 SHALL 要求动态插件多模态 `ai` host service 使用 `assetRef` 或受控临时资产引用传递大对象。host service 请求和响应 MUST NOT 传输无上限 base64 或完整二进制内容。

#### Scenario: 图片输入使用资产引用

- **WHEN** 动态插件调用 `vision.analyze` 并提供图片输入
- **THEN** 请求 MUST 使用 `assetRef`、mime 类型和大小投影引用图片
- **AND** 宿主 MUST 校验该资产引用对当前插件和请求上下文可访问

#### Scenario: 音频输出使用资产引用

- **WHEN** 动态插件调用 `audio.synthesize` 成功
- **THEN** 响应 MUST 返回输出音频的 `assetRef` 和摘要投影
- **AND** 响应 MUST NOT 返回完整音频 base64

### Requirement: AI host service 必须支持 provider operation 查询边界

系统 SHALL 允许动态插件在获得授权后使用 provider operation 查询方法跟踪渠道异步 operation。operation 查询 MUST 表达渠道协议状态，MUST NOT 表达业务任务状态。

#### Scenario: 视频生成返回 provider operation

- **WHEN** 动态插件调用 `video.generate`
- **AND** 渠道返回异步 operation
- **THEN** host service 响应 MUST 返回不透明 `operationRef`、状态、渠道模型投影、`nextPollAfterMs` 和过期时间
- **AND** 响应 MUST NOT 返回业务任务 ID

#### Scenario: 查询 operation 状态

- **WHEN** 动态插件调用 `video.operation.get`
- **AND** 插件已获该 operation 所属方法和资源授权
- **THEN** 宿主 MUST 返回 operation 当前状态或完成后的资产引用
- **AND** 宿主 MUST NOT 返回 provider 原始认证 URL、密钥或完整响应正文

#### Scenario: 未授权取消被拒绝

- **WHEN** 动态插件调用 `video.operation.cancel`
- **AND** 授权资源未允许取消或 provider 不支持取消
- **THEN** 宿主 MUST 拒绝调用并返回结构化错误
- **AND** 宿主 MUST NOT 执行 provider 取消请求

### Requirement: ai service 必须记录最小审计信息

系统 SHALL 对动态插件 `ai` host service 调用记录最小宿主服务审计和智能中心调用日志。审计 MUST 支持诊断插件、方法、资源、授权、耗时、状态和来源插件，但 MUST NOT 保存完整输入、完整输出、完整资产、隐藏思考内容、渠道响应原文或密钥。

#### Scenario: 文本成功调用记录来源

- **WHEN** 动态插件通过 `ai.text.generate` 成功生成文本
- **THEN** 宿主服务审计 MUST 记录 `pluginId`、service、method、purpose 摘要、结果状态和耗时
- **AND** 智能中心调用日志 MUST 记录 `sourcePluginId`、`purpose`、档位、渠道模型投影、`thinkingEffort`、token 用量和耗时

#### Scenario: 多模态成功调用记录摘要

- **WHEN** 动态插件通过多模态 `ai` host service 成功调用 provider
- **THEN** 宿主服务审计 MUST 记录 `pluginId`、service、method、purpose、授权资源摘要、状态和耗时
- **AND** 智能中心调用日志 MUST 记录来源插件、能力方法、渠道模型投影、资产引用摘要和用量摘要

#### Scenario: 失败调用脱敏

- **WHEN** 动态插件调用 `ai` host service 失败
- **THEN** 宿主服务审计和智能中心调用日志 MUST 记录失败状态、稳定错误码和脱敏错误摘要
- **AND** 审计和日志 MUST NOT 包含完整 `messages`、完整文件内容、音视频内容、API key、认证头或渠道响应原文

### Requirement: ai service 必须拒绝首期未开放的方法

系统 SHALL 将 `ai` service family 设计为可扩展宿主服务族，但首期 MUST 只开放 `text.generate` 和已定义规范的多模态方法。图片、音频、向量、重排、工具调用、流式输出或其他未定义规范的 `AI` 方法在未有正式规范和授权模型前 MUST 被拒绝。

#### Scenario: 未开放方法被拒绝

- **WHEN** 动态插件声明或调用未定义规范的 `ai.*` 方法
- **THEN** 构建器、宿主清单校验或运行时 handler MUST 拒绝该方法
- **AND** 错误 MUST 指出该 `AI` host service method 当前不受支持

#### Scenario: 后续能力独立扩展

- **WHEN** 系统后续新增图片、音频或向量能力
- **THEN** 新方法 MUST 明确定义 `capabilityType`、`capabilityMethod`、资源授权、请求响应 DTO、审计字段和与框架 capability 的适配关系
- **AND** 新方法 MUST NOT 改变现有 `ai.text.generate` 的授权和同步文本响应语义

### Requirement: 动态插件 guest SDK 必须通过 AI 命名空间调用

系统 SHALL 在动态插件 guest 侧通过 `AI().Text()` 和 `AI().Image()` 等入口暴露 `AI` 能力。guest SDK 的调用 MUST 继续使用既有 `ai` host service 协议，并保持 `host:ai:text`、`purpose:<name>` 和策略属性授权语义不变。

#### Scenario: 动态插件通过 AI 命名空间生成文本

- **WHEN** 动态插件需要调用文本 `AI` 生成能力
- **THEN** guest 代码 MUST 通过 `guest.Default().AI().Text().GenerateText(...)` 或等价能力目录入口发起调用
- **AND** guest SDK MUST NOT 继续要求调用方使用根目录 `AIText()` 方法

#### Scenario: guest AI 调用进入既有 host service

- **WHEN** guest SDK 执行 `AI().Text().GenerateText(...)` 或 `AI().Image().Generate(...)`
- **THEN** SDK MUST 构造既有 `service: ai` 和对应 `method` 的 host service 调用
- **AND** 请求资源 MUST 继续使用 `purpose:<name>` 表达授权用途
- **AND** 宿主 MUST 在执行能力调用或渠道调用前完成 service、method、资源和策略属性校验

#### Scenario: 动态插件协议不因 Go 入口重构改变

- **WHEN** 系统将 guest 侧调用入口从 `AIText()` 重构为 `AI().Text()`
- **THEN** 动态插件 `plugin.yaml` 中的 `hostServices` 声明格式 MUST 保持 `service: ai` 和对应 `methods`
- **AND** `host:ai:text` 的能力分类、`maxOutputTokens` 等资源策略和脱敏审计语义 MUST 保持不变

#### Scenario: 未开放的 AI 子能力仍被拒绝

- **WHEN** 动态插件通过 `AI()` 命名空间尝试调用尚未规范化的图片、向量、音频或其他 `AI` 子能力
- **THEN** guest SDK 或宿主 MUST 返回不支持错误
- **AND** 宿主 MUST NOT 因存在 `AI()` 聚合入口而自动授予未声明子能力
