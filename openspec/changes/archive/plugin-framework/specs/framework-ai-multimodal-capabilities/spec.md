# 多模态 AI 能力 owner 迁移

## Purpose
保留多模态 AI 方法语义与安全边界，迁移契约包与动态方法发布 owner。

## Requirements

### Requirement: 宿主必须通过 AI 命名空间暴露类型化多模态子能力

系统 SHALL 由`linapro-ai-core/backend/cap/aicap`在`AI()`命名空间下暴露图片、向量、音频、视觉、文档、安全审核和视频等类型化子能力。根`lina-core/pkg/plugin/capability.Services` MUST NOT 新增`AIImage()`、`AIEmbedding()`、`AIAudio()`等按子能力展开的根方法，也 MUST NOT 继续作为多模态`AI`契约 owner。

#### Scenario: 调用方获取图片能力

- **WHEN** 宿主模块、源码插件或动态插件需要使用图片生成或编辑能力
- **THEN** 调用方 MUST 通过 owner 插件`AI().Image()`或等价 owner guest 能力目录获取图片子能力
- **AND** 调用方 MUST NOT 通过 core 根能力目录新增的`AIImage()`入口获取能力

#### Scenario: 调用方获取向量能力

- **WHEN** 调用方需要创建 embedding
- **THEN** 调用方 MUST 通过 owner 插件`AI().Embedding()`或等价 owner guest 能力目录获取向量子能力
- **AND** 向量能力 MUST 维护独立请求、响应、错误和 provider 契约

#### Scenario: 弱类型 AI 网关被拒绝

- **WHEN** 实现多模态`AI`能力
- **THEN** 系统 MUST NOT 引入`AI().Invoke(method, payload)`、`AI().Generate(capabilityType, payload)`或等价弱类型业务网关作为普通消费契约
- **AND** 每个子能力 MUST 使用自己的 DTO、方法常量、错误和状态语义

### Requirement: 多模态能力方法必须按能力族定义

系统 SHALL 在 owner 插件契约中使用`capabilityType + capabilityMethod`表达多模态能力方法。图片、向量、音频、视觉、文档、安全审核和视频 MUST 使用独立能力方法，MUST NOT 复用`text.generate`承载非文本语义。owner 契约 MUST 为这些方法维护方法常量、DTO、风险和资源形态；当且仅当对应 provider 运行时路径就绪时，owner descriptor 才 MUST 把该方法发布到动态授权 catalog。

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
- **THEN** owner 能力目录和动态插件 host service MUST 拒绝该方法
- **AND** 错误 MUST 表明该能力不在本轮多模态`AI`能力范围内

### Requirement: 多模态能力必须提供可用性和降级状态

系统 SHALL 由 owner 插件为每个多模态子能力提供可用性和状态查询。provider 缺失、模型能力未声明、档位未配置、endpoint 禁用、密钥不可用或方法不支持时，系统 MUST 返回结构化不可用状态或业务错误。动态插件读取多模态状态时 MUST 通过 owner-aware host service 和授权快照进入 owner handler。

#### Scenario: 能力方法不可用

- **WHEN** 调用方请求`audio.transcribe`
- **AND** 当前没有启用的 provider model 声明支持`audio.transcribe`
- **THEN** `AI().Audio().Status()`或等价状态能力 MUST 返回不可用原因
- **AND** 调用方法 MUST 在执行渠道请求前返回结构化不可用错误

#### Scenario: 插件禁用后能力降级

- **WHEN** `linapro-ai-core`插件被禁用、卸载或 provider factory 启动失败
- **THEN** 所有多模态子能力 MUST 返回不可用状态
- **AND** 宿主 MUST NOT 产生 500 或返回 nil service


### Requirement: 多模态动态方法不得继续由 core codec 拥有

系统 SHALL 将多模态`AI`动态方法的请求响应 DTO、codec、method 常量和 guest helper 迁移到 owner 插件公开 bridge SDK 或 owner 契约包。core `pluginbridge`只保留通用 owner-aware envelope、授权、转发和错误映射，不得继续维护`HostServiceAIImageGenerateRequest`、`HostServiceAIAudioTranscribeRequest`等 AI 专属生产 owner 别名。

#### Scenario: 图片方法动态调用

- **WHEN** 动态插件调用`image.generate`
- **THEN** owner bridge SDK MUST 使用 owner 契约中的图片 DTO 编码 payload
- **AND** core dispatcher MUST 按 owner descriptor 转发
- **AND** core 不得进入图片生成专属 dispatcher 分支

#### Scenario: 视频 operation 动态调用

- **WHEN** 动态插件调用`video.operation.get`或`video.operation.cancel`
- **THEN** owner bridge SDK MUST 维护 operation 请求响应契约
- **AND** core 只校验 owner、service、version、method、授权快照和资源范围
