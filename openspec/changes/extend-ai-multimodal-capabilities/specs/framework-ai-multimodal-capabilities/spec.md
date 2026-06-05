## ADDED Requirements

### Requirement: 宿主必须通过 AI 命名空间暴露类型化多模态子能力

系统 SHALL 在`AI()`命名空间下暴露图片、向量、音频、视觉、文档、安全审核和视频等类型化子能力。根`capability.Services` MUST NOT 新增`AIImage()`、`AIEmbedding()`、`AIAudio()`等按子能力展开的根方法。

#### Scenario: 调用方获取图片能力

- **WHEN** 宿主模块、源码插件或动态插件需要使用图片生成或编辑能力
- **THEN** 调用方 MUST 通过`AI().Image()`或等价 guest 能力目录获取图片子能力
- **AND** 调用方 MUST NOT 通过根能力目录新增的`AIImage()`入口获取能力

#### Scenario: 调用方获取向量能力

- **WHEN** 调用方需要创建 embedding
- **THEN** 调用方 MUST 通过`AI().Embedding()`或等价 guest 能力目录获取向量子能力
- **AND** 向量能力 MUST 维护独立请求、响应、错误和 provider 契约

#### Scenario: 弱类型 AI 网关被拒绝

- **WHEN** 实现多模态`AI`能力
- **THEN** 系统 MUST NOT 引入`AI().Invoke(method, payload)`、`AI().Generate(capabilityType, payload)`或等价弱类型业务网关作为普通消费契约
- **AND** 每个子能力 MUST 使用自己的 DTO、方法常量、错误和状态语义

### Requirement: 多模态能力方法必须按能力族定义

系统 SHALL 使用`capabilityType + capabilityMethod`表达多模态能力方法。图片、向量、音频、视觉、文档、安全审核和视频 MUST 使用独立能力方法，MUST NOT 复用`text.generate`承载非文本语义。

#### Scenario: 图片能力方法

- **WHEN** 系统定义图片能力
- **THEN** 系统 MUST 至少支持`image.generate`和`image.edit`方法语义
- **AND** 图片方法 MUST 使用图片专属输入、资产输出和错误语义

#### Scenario: 音频能力方法

- **WHEN** 系统定义音频能力
- **THEN** 系统 MUST 使用`audio.transcribe`表达语音转文本
- **AND** 系统 MUST 使用`audio.synthesize`表达文本转语音
- **AND** 两个方法 MUST 使用各自独立的输入格式、输出格式和模型能力约束

#### Scenario: 视觉和文档方法

- **WHEN** 系统定义图像理解或文档理解能力
- **THEN** 图像、截图、图表等理解 MUST 使用`vision.analyze`
- **AND** PDF、文档、表格或引用型文档理解 MUST 使用`document.analyze`或`document.cite`
- **AND** 系统 MUST NOT 把这些方法塞入`text.generate`作为多模态消息兼容分支

#### Scenario: computer act 被排除

- **WHEN** 调用方声明或请求`computer.act`、`ui.operate`或等价 UI 控制能力
- **THEN** 宿主能力目录和动态插件 host service MUST 拒绝该方法
- **AND** 错误 MUST 表明该能力不在本轮多模态`AI`能力范围内

### Requirement: 大对象结果必须返回资产引用

系统 SHALL 对图片、音频、视频等大对象结果返回`assetRef`或受控临时资产引用。能力响应、HTTP JSON、WASM host call 和插件调用结果 MUST NOT 返回无上限 base64 或完整二进制内容。

#### Scenario: 图片生成返回资产引用

- **WHEN** `image.generate`成功生成图片
- **THEN** 响应 MUST 返回`assetRef`、`mimeType`、`sizeBytes`、`width`、`height`和生成时间投影
- **AND** 响应 MUST NOT 返回完整 base64 图片内容

#### Scenario: 音频合成返回资产引用

- **WHEN** `audio.synthesize`成功生成音频
- **THEN** 响应 MUST 返回`assetRef`、`mimeType`、`sizeBytes`和`durationMs`
- **AND** 调用方 MUST 通过受控资产能力读取或下载音频内容

#### Scenario: 渠道临时 URL 不直接外泄

- **WHEN** 渠道返回临时下载 URL 或远程资产 URL
- **THEN** provider adapter MUST 将其转换为受控资产引用或受控临时资产投影
- **AND** 响应 MUST NOT 向动态插件或前端暴露携带认证信息的渠道 URL

### Requirement: 渠道异步协议必须通过 ProviderOperationRef 表达

系统 SHALL 使用`ProviderOperationRef`表达渠道侧异步 operation。`ProviderOperationRef` MUST 是 provider 协议适配引用，MUST NOT 表示业务任务、业务队列或用户进度记录。

#### Scenario: 视频生成返回 operation 引用

- **WHEN** `video.generate`调用的渠道无法同步返回最终视频资产
- **THEN** 能力响应 MUST 返回`ProviderOperationRef`
- **AND** `ProviderOperationRef` MUST 包含不透明`operationRef`、能力方法、渠道模型投影、状态、建议下次查询时间和过期时间
- **AND** 响应 MUST NOT 创建或返回业务视频任务 ID

#### Scenario: 查询 provider operation

- **WHEN** 调用方持有有效`operationRef`并调用`video.operation.get`
- **THEN** 系统 MUST 查询渠道 operation 状态或本地最小投影
- **AND** operation 成功完成时 MUST 返回资产引用
- **AND** operation 未完成时 MUST 返回状态和`nextPollAfterMs`

#### Scenario: 业务异步由业务模块负责

- **WHEN** 业务模块需要后台生成、轮询、重试、取消、通知或绑定业务资产
- **THEN** 业务模块 MUST 创建并管理自己的业务任务、状态和数据归属
- **AND** `lina-core`和`linapro-ai-core` MUST NOT 提供`/api/ai/video-jobs`或等价具体业务任务 API

### Requirement: 多模态能力必须提供可用性和降级状态

系统 SHALL 为每个多模态子能力提供可用性和状态查询。provider 缺失、模型能力未声明、档位未配置、endpoint 禁用、密钥不可用或方法不支持时，系统 MUST 返回结构化不可用状态或业务错误。

#### Scenario: 能力方法不可用

- **WHEN** 调用方请求`audio.transcribe`
- **AND** 当前没有启用的 provider model 声明支持`audio.transcribe`
- **THEN** `AI().Audio().Status()`或等价状态能力 MUST 返回不可用原因
- **AND** 调用方法 MUST 在执行渠道请求前返回结构化不可用错误

#### Scenario: 插件禁用后能力降级

- **WHEN** `linapro-ai-core`插件被禁用、卸载或 provider factory 启动失败
- **THEN** 所有多模态子能力 MUST 返回不可用状态
- **AND** 宿主 MUST NOT 产生 500 或返回 nil service

### Requirement: 多模态调用必须保护敏感输入输出

系统 SHALL 将多模态调用视为敏感执行路径。能力层、host service 和 provider adapter MUST NOT 默认记录完整图片、音频、视频、文件、prompt、渠道响应原文或密钥。

#### Scenario: 成功调用只记录摘要

- **WHEN** 多模态能力调用成功
- **THEN** 系统记录调用日志时 MUST 只记录`capabilityType`、`capabilityMethod`、`purpose`、档位、渠道模型投影、资产引用、用量、耗时和状态等最小摘要
- **AND** 系统 MUST NOT 保存完整输入资产、完整输出资产或完整渠道响应正文

#### Scenario: 失败调用脱敏

- **WHEN** 多模态 provider 调用失败
- **THEN** 错误摘要 MUST 脱敏 API key、认证头、渠道 URL 中的凭证、请求体和响应体敏感片段
- **AND** 返回给调用方的错误 MUST 包含稳定错误码和可本地化消息键
