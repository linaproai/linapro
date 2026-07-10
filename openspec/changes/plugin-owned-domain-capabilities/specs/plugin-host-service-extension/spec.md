## ADDED Requirements

### Requirement: 动态 hostServices 必须支持 owner-aware 能力声明

系统 SHALL 扩展动态插件`hostServices`声明，使 plugin-owned 领域能力可以通过`service`、`owner`、`version`、`methods`和`resources`表达。`owner`字段 MUST 使用 owner 插件 ID，`version`字段 MUST 使用 owner capability 协议版本。core-owned 宿主内核能力 MAY 继续省略`owner`，并按现有 service/method 语义处理。plugin-owned 能力 MUST 使用显式`owner`字段，不得只通过拼接型 service key 表达 owner、capability 和 version。

#### Scenario: 声明 owner AI 服务

- **WHEN** 动态插件需要调用`linapro-ai-core`发布的`AI`文本生成方法
- **THEN** `plugin.yaml hostServices` MUST 声明`service: ai`、`owner: linapro-ai-core`、`version: v1`和`methods: [text.generate]`
- **AND** 宿主 MUST 将该声明归一化到授权快照中，保留 owner、service、version、method 和资源范围

#### Scenario: core-owned 服务保持现有声明

- **WHEN** 动态插件声明`service: storage`、`service: cache`或`service: runtime`
- **THEN** 清单 MAY 继续省略`owner`
- **AND** 宿主 MUST 按既有 core-owned service catalog 和资源授权语义校验

#### Scenario: 拼接型 owner service key 被拒绝

- **WHEN** 动态插件使用`service: plugin:linapro-ai-core:ai:v1`或等价拼接 service key 声明 owner 能力
- **THEN** 清单校验 MUST 失败
- **AND** 错误必须提示使用结构化`owner`和`version`字段

### Requirement: owner-aware host service catalog 必须由 core 和 owner descriptor 合并

系统 SHALL 将动态 host service catalog 拆分为 core-owned 静态 catalog 与 plugin-owned owner descriptor 投影。core MUST 继续维护宿主内核和宿主通用能力 catalog；owner 插件 MUST 通过 capability descriptor 发布其动态方法、请求响应 codec 标识、风险、资源形态、方法状态和文档信息。宿主在构建、安装、启用、升级和 API/UI 展示时 MUST 合并两类 catalog，并在 owner 插件缺失、禁用或版本不满足时给出结构化诊断。

#### Scenario: owner descriptor 发布 AI 方法

- **WHEN** `linapro-ai-core`发布`ai.v1` descriptor
- **THEN** descriptor MUST 至少包含`text.generate`、`text.method_status.get`、`ai.methods.status.batch_get`以及已发布的多模态方法
- **AND** core 不得在`pluginbridge/protocol/hostservices/catalog.go`中继续硬编码这些 AI 方法作为生产 owner

#### Scenario: 动态升级 diff 展示 owner 来源

- **WHEN** 动态插件升级改变 owner 能力方法声明
- **THEN** 升级预览 MUST 展示 owner 插件 ID、service、version、method、资源变化和是否需要重新授权
- **AND** 前端不得只展示无 owner 的`service + method`字符串

### Requirement: owner-aware dispatcher 必须通用转发而非领域 switch

系统 SHALL 为 plugin-owned 动态能力提供通用 dispatcher。dispatcher MUST 从授权快照和运行时请求 envelope 中读取 owner、service、version、method 和资源标识，校验调用插件依赖、owner 启用状态、method 授权和资源范围后，通过 capability descriptor 定位 owner handler。dispatcher MUST NOT 为每个 owner 领域维护独立 Go `switch`、专属 codec 文件或专属业务分发函数。

#### Scenario: 调用 AI 文本生成

- **WHEN** 动态插件调用`owner=linapro-ai-core service=ai version=v1 method=text.generate`
- **THEN** dispatcher MUST 校验调用插件已声明并满足`linapro-ai-core`依赖
- **AND** dispatcher MUST 校验该 method 已授权并由 owner descriptor 注册
- **AND** dispatcher MUST 将 payload 转发给`linapro-ai-core`注册的 handler
- **AND** dispatcher 不得进入 core 内置`dispatchAITextGenerate`分支

#### Scenario: 未授权 owner 方法

- **WHEN** 动态插件调用未授权或未注册的 owner method
- **THEN** dispatcher MUST 返回结构化 denied 或 not found 错误
- **AND** owner handler 不得被调用

### Requirement: owner-aware 动态调用错误必须使用稳定 envelope

系统 SHALL 使用现有 host call response envelope 返回 owner-aware 动态调用结果。owner 插件返回的业务错误 MUST 映射为稳定 host call 状态、errorCode、messageKey、messageParams 和英文 fallback；错误摘要 MUST 脱敏 provider 密钥、认证头、完整请求体、完整响应体和内部路由配置。

#### Scenario: owner 插件不可用

- **WHEN** 动态插件调用的 owner 插件未安装、未启用或版本不满足
- **THEN** host call response MUST 返回 capability unavailable 或 dependency blocker 语义
- **AND** 响应不得泄露 owner 插件内部初始化错误、密钥或数据库结构

#### Scenario: provider 失败脱敏

- **WHEN** `linapro-ai-core`调用外部 provider 失败
- **THEN** owner handler MUST 返回脱敏后的结构化业务错误
- **AND** host call envelope MUST 保留稳定错误码和可本地化消息键
