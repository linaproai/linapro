# AI 能力命名空间 owner 迁移

## Purpose
保留 AI 命名空间与类型化子能力语义，并将非核心 AI 契约 owner 迁移到 linapro-ai-core。

## Requirements

### Requirement: 宿主能力目录必须通过 AI 命名空间暴露 AI 能力

系统 SHALL 由`linapro-ai-core`在`backend/cap/aicap`发布`AI`能力族命名空间。源码消费插件 MUST 通过 owner 插件公开契约获取`AI`类型化子能力，例如`AI().Text()`或注入的`aitext.Service`。动态插件 MUST 通过 owner 插件公开 bridge SDK 和 owner-aware host service 调用对应子能力。根`lina-core/pkg/plugin/capability.Services` MUST NOT 继续作为`AI`生产契约 owner，也 MUST NOT 直接暴露`AIText()`、`AIImage()`、`AIEmbedding()`或其他按`AI`子能力展开的根方法。

#### Scenario: 源码插件获取文本 AI 能力

- **WHEN** 源码插件需要使用文本`AI`能力
- **THEN** 插件 MUST 依赖`linapro-ai-core`的`backend/cap/aicap`公开契约
- **AND** 插件 MUST 通过`AI().Text()`、注入的`aitext.Service`或 owner helper 获取文本能力
- **AND** 插件 MUST NOT 通过 core 根能力目录`services.AIText()`获取文本能力

#### Scenario: 宿主内部模块获取文本 AI 能力

- **WHEN** 宿主内部模块可选消费文本`AI`能力
- **THEN** 模块 MUST 通过显式注入的 owner 能力引用或通用 capability descriptor 解析后的契约调用`AI().Text()`
- **AND** 模块 MUST NOT 直接依赖`linapro-ai-core/backend/internal/**`或 provider adapter

#### Scenario: 根能力目录新增后续 AI 子能力

- **WHEN** 系统后续新增图片、向量、音频或其他`AI`子能力
- **THEN** 新子能力 MUST 挂载到`linapro-ai-core/backend/cap/aicap`发布的`AI`命名空间下
- **AND** 新子能力 MUST NOT 在 core 根`capability.Services`上新增`AI*()`方法

### Requirement: AI 聚合服务必须只承担子能力聚合职责

系统 SHALL 使用`aicap.Service`聚合`AI`子能力。`aicap.Service` MUST 只暴露类型化子能力入口，例如`Text() aitext.Service`、`Image()`或`Embedding()`，MUST NOT 作为弱类型`AI`网关执行运行时 method 分发。通用动态 dispatcher 可以按 descriptor 路由 host call，但不得把弱类型 payload 网关暴露为普通源码插件消费契约。

#### Scenario: 文本能力通过 Text 入口访问

- **WHEN** 调用方需要执行同步文本生成
- **THEN** 调用方 MUST 使用`AI().Text().GenerateText(...)`或注入的`aitext.Service.GenerateText(...)`
- **AND** `Text()`返回的 service MUST 保持`plugin.linapro-ai-core.ai.text.v1`的状态、降级和错误语义

#### Scenario: 弱类型 AI 网关被拒绝

- **WHEN** 实现`AI`能力聚合服务
- **THEN** 系统 MUST NOT 引入`Generate(ctx, capabilityType, payload)`、`Invoke(ctx, method, payload)`或等价弱类型业务网关作为普通消费契约
- **AND** 文本、图片、向量等子能力 MUST 维护各自的 DTO、错误和授权边界

### Requirement: 文本 AI 能力必须归属 AI 命名空间

系统 SHALL 将文本`AI`能力包归属到`apps/lina-plugins/linapro-ai-core/backend/cap/aicap/aitext`或 owner 插件内等价`aicap`子包。生产代码 MUST 使用该 owner 插件路径引用文本能力契约。`apps/lina-core/pkg/plugin/capability/aicap`或历史`capability/ai/aitext`路径 MUST 不再作为生产消费入口保留。

#### Scenario: 生产代码引用文本 AI 契约

- **WHEN** 宿主、源码插件或动态插件生产代码引用文本`AI`契约
- **THEN** 代码 MUST import `lina-plugin-linapro-ai-core/backend/cap/aicap/aitext`或 owner 插件公开等价路径
- **AND** 代码 MUST NOT import `lina-core/pkg/plugin/capability/aicap/aitext`作为生产契约 owner

#### Scenario: 文本能力行为保持不变

- **WHEN** 文本能力包迁移到`linapro-ai-core/backend/cap/aicap/aitext`
- **THEN** `plugin.linapro-ai-core.ai.text.v1`的 capability ID、`Available(ctx)`、`Status(ctx)`、`GenerateText(ctx, request)`和 provider factory 语义 MUST 保持不变
- **AND** 迁移 MUST NOT 将渠道、模型、档位或调用日志业务存储移入`lina-core`


### Requirement: AI owner 迁移必须覆盖全部已发布子能力方法

系统 SHALL 将当前 core catalog 中已发布的`AI`方法作为同一 owner 迁移面处理，包括文本、图片、向量、音频、视觉、文档、安全审核、视频和 operation 方法。这些方法的公开 DTO、方法常量和错误语义 MUST 迁到`linapro-ai-core/backend/cap/aicap`，不得继续由 core 拥有。动态 descriptor 仅 MUST 发布当前具备真实运行时路径的方法，避免授权 catalog 展示永远不可用的方法；尚未接线的多模态方法保留在 owner 契约中，待 provider SPI 落地后再发布到 descriptor。

#### Scenario: 契约全量迁出 core，动态方法按可运行路径发布

- **WHEN** `linapro-ai-core`发布`ai.v1`动态 descriptor
- **THEN** owner 契约包 MUST 覆盖文本和多模态公开 DTO/方法常量
- **AND** 动态 descriptor MUST 至少发布`text.generate`、`text.method_status.get`和`ai.methods.status.batch_get`
- **AND** 静态检索 MUST 证明 core 不再保留未说明的`AI`专属 host service 方法 owner
