## ADDED Requirements

### Requirement: 动态插件 AI host service 必须支持多模态方法声明

系统 SHALL 扩展动态插件`service: ai`host service，使插件可以声明图片、向量、音频、视觉、文档、安全审核和视频方法。每个方法 MUST 映射到明确的`capabilityType + capabilityMethod`和独立授权分类。

#### Scenario: 插件声明图片生成能力

- **WHEN** 动态插件在`plugin.yaml`中声明`service: ai`和`methods: [image.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为`image.generate`权限申请
- **AND** 运行时 MUST 将方法映射为`capabilityType=image`和`capabilityMethod=generate`
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限

#### Scenario: 插件声明音频能力

- **WHEN** 动态插件声明`audio.transcribe`或`audio.synthesize`
- **THEN** 宿主 MUST 分别识别为不同 host service 方法
- **AND** 两个方法 MUST 使用独立 payload 契约、资源策略和授权分类

#### Scenario: computer act 声明被拒绝

- **WHEN** 动态插件声明`computer.act`、`ui.operate`或等价 UI 控制方法
- **THEN** 清单校验或运行时 MUST 拒绝该声明
- **AND** 错误 MUST 表明该方法不属于本轮`ai`host service 支持范围

### Requirement: 多模态 AI host service 必须按方法和资源授权

系统 SHALL 对每一次多模态`ai`host service 调用同时校验 service、method、resource、调用来源和策略属性。任一校验失败时，宿主 MUST 在读取密钥、解析档位或调用渠道前拒绝请求。

#### Scenario: 授权资源匹配后调用成功

- **WHEN** 动态插件已获`image.generate`授权
- **AND** 请求的`purpose`、输入 mime 类型、最大输入资产数和最大输出数量均满足授权资源策略
- **THEN** host service handler MUST 将请求转换为`AI().Image().Generate(...)`或等价能力调用
- **AND** 调用 MUST 复用对应子能力的 provider 可用性和错误语义

#### Scenario: 未授权 purpose 被拒绝

- **WHEN** 动态插件请求未在授权快照中确认的`purpose`
- **THEN** 宿主 MUST 拒绝调用
- **AND** 宿主 MUST NOT 读取渠道 endpoint、secret 引用或模型绑定

#### Scenario: payload 超限被拒绝

- **WHEN** 动态插件上传或引用的输入资产数量、字节数、mime 类型、输出数量或 token 上限超过授权策略
- **THEN** 宿主 MUST 拒绝调用或按显式策略收敛
- **AND** 宿主 MUST NOT 静默扩大插件授权范围

### Requirement: AI host service 大对象 payload 必须使用资产引用

系统 SHALL 要求动态插件多模态`ai`host service 使用`assetRef`或受控临时资产引用传递大对象。host service 请求和响应 MUST NOT 传输无上限 base64 或完整二进制内容。

#### Scenario: 图片输入使用资产引用

- **WHEN** 动态插件调用`vision.analyze`并提供图片输入
- **THEN** 请求 MUST 使用`assetRef`、mime 类型和大小投影引用图片
- **AND** 宿主 MUST 校验该资产引用对当前插件和请求上下文可访问

#### Scenario: 音频输出使用资产引用

- **WHEN** 动态插件调用`audio.synthesize`成功
- **THEN** 响应 MUST 返回输出音频的`assetRef`和摘要投影
- **AND** 响应 MUST NOT 返回完整音频 base64

### Requirement: AI host service 必须支持 provider operation 查询边界

系统 SHALL 允许动态插件在获得授权后使用 provider operation 查询方法跟踪渠道异步 operation。operation 查询 MUST 表达渠道协议状态，MUST NOT 表达业务任务状态。

#### Scenario: 视频生成返回 provider operation

- **WHEN** 动态插件调用`video.generate`
- **AND** 渠道返回异步 operation
- **THEN** host service 响应 MUST 返回不透明`operationRef`、状态、渠道模型投影、`nextPollAfterMs`和过期时间
- **AND** 响应 MUST NOT 返回业务任务 ID

#### Scenario: 查询 operation 状态

- **WHEN** 动态插件调用`video.operation.get`
- **AND** 插件已获该 operation 所属方法和资源授权
- **THEN** 宿主 MUST 返回 operation 当前状态或完成后的资产引用
- **AND** 宿主 MUST NOT 返回 provider 原始认证 URL、密钥或完整响应正文

#### Scenario: 未授权取消被拒绝

- **WHEN** 动态插件调用`video.operation.cancel`
- **AND** 授权资源未允许取消或 provider 不支持取消
- **THEN** 宿主 MUST 拒绝调用并返回结构化错误
- **AND** 宿主 MUST NOT 执行 provider 取消请求

### Requirement: 多模态 AI host service 必须记录最小审计

系统 SHALL 对动态插件多模态`ai`host service 调用记录最小审计信息。审计 MUST 支持诊断插件、方法、资源、耗时、状态和错误，但 MUST NOT 保存完整输入输出、大对象内容、渠道响应原文或密钥。

#### Scenario: 成功调用记录摘要

- **WHEN** 动态插件通过多模态`ai`host service 成功调用 provider
- **THEN** 宿主服务审计 MUST 记录`pluginId`、service、method、purpose、授权资源摘要、状态和耗时
- **AND** 智能中心调用日志 MUST 记录来源插件、能力方法、渠道模型投影、资产引用摘要和用量摘要

#### Scenario: 失败调用脱敏

- **WHEN** 多模态`ai`host service 调用失败
- **THEN** 审计和调用日志 MUST 记录失败状态、稳定错误码和脱敏错误摘要
- **AND** 审计和日志 MUST NOT 包含完整文件内容、音视频内容、API key、认证头或渠道响应原文
