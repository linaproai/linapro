## ADDED Requirements

### Requirement: 宿主能力目录必须通过 AI 命名空间暴露 AI 能力

系统 SHALL 在根 `capability.Services` 中通过 `AI() ai.Service` 暴露 `AI` 能力族。根能力目录 MUST NOT 直接暴露 `AIText()`、`AIImage()`、`AIEmbedding()` 或其他按 `AI` 子能力展开的根方法。

#### Scenario: 源码插件获取文本 AI 能力

- **WHEN** 源码插件通过宿主发布的能力目录获取文本 `AI` 能力
- **THEN** 插件 MUST 通过 `services.AI().Text()` 获取文本能力
- **AND** 插件 MUST NOT 通过 `services.AIText()` 获取文本能力

#### Scenario: 宿主内部模块获取文本 AI 能力

- **WHEN** 宿主内部模块可选消费文本 `AI` 能力
- **THEN** 模块 MUST 通过显式注入的能力目录调用 `AI().Text()`
- **AND** 模块 MUST NOT 直接依赖 `linapro-ai-core/backend/internal/**` 或 provider adapter

#### Scenario: 根能力目录新增后续 AI 子能力

- **WHEN** 系统后续新增图片、向量、音频或其他 `AI` 子能力
- **THEN** 新子能力 MUST 挂载到 `ai.Service` 下
- **AND** 新子能力 MUST NOT 在根 `capability.Services` 上新增 `AI*()` 方法

### Requirement: AI 聚合服务必须只承担子能力聚合职责

系统 SHALL 使用 `ai.Service` 聚合 `AI` 子能力。`ai.Service` MUST 只暴露类型化子能力入口，例如 `Text() aitext.Service`，MUST NOT 作为弱类型 `AI` 网关执行运行时 method 分发。

#### Scenario: 文本能力通过 Text 入口访问

- **WHEN** 调用方需要执行同步文本生成
- **THEN** 调用方 MUST 使用 `AI().Text().GenerateText(...)`
- **AND** `Text()` 返回的 service MUST 保持 `framework.ai.text.v1` 的状态、降级和错误语义

#### Scenario: 弱类型 AI 网关被拒绝

- **WHEN** 实现 `AI` 能力聚合服务
- **THEN** 系统 MUST NOT 引入 `Generate(ctx, capabilityType, payload)`、`Invoke(ctx, method, payload)` 或等价弱类型业务网关作为普通消费契约
- **AND** 文本、图片、向量等子能力 MUST 维护各自的 DTO、错误和授权边界

### Requirement: 文本 AI 能力必须归属 AI 命名空间

系统 SHALL 将文本 `AI` 能力包归属到 `apps/lina-core/pkg/plugin/capability/ai/aitext`。生产代码 MUST 使用该新路径引用文本能力契约，旧 `apps/lina-core/pkg/plugin/capability/aitext` 路径 MUST 不再作为生产消费入口保留。

#### Scenario: 生产代码引用文本 AI 契约

- **WHEN** 宿主、源码插件或动态插件生产代码引用文本 `AI` 契约
- **THEN** 代码 MUST import `lina-core/pkg/plugin/capability/ai/aitext`
- **AND** 代码 MUST NOT import `lina-core/pkg/plugin/capability/aitext`

#### Scenario: 文本能力行为保持不变

- **WHEN** 文本能力包迁移到 `capability/ai/aitext`
- **THEN** `framework.ai.text.v1` 的 capability ID、`Available(ctx)`、`Status(ctx)`、`GenerateText(ctx, request)` 和 provider factory 语义 MUST 保持不变
- **AND** 迁移 MUST NOT 新增渠道、模型、档位或调用日志宿主存储

### Requirement: 文本 AI 来源身份必须由能力服务注入

系统 SHALL 将文本生成消费请求与 provider 内部请求分离。普通调用方可见的 `GenerateRequest` MUST NOT 要求填写 `SourcePluginID`；插件来源身份 MUST 由 plugin-scoped 能力 service 或动态插件 host-call 上下文注入到 provider 请求。

#### Scenario: 源码插件调用注入插件来源

- **WHEN** 源码插件通过 `ServicesForPlugin(services, pluginID).AI().Text()` 发起文本生成
- **THEN** 文本能力 service MUST 将该 `pluginID` 作为 provider 请求的来源插件标识
- **AND** 普通调用方 MUST NOT 在消费请求中自行填写或伪造 `SourcePluginID`

#### Scenario: 动态插件调用注入插件来源

- **WHEN** 动态插件通过 `ai.text.generate` host service 发起文本生成
- **THEN** `WASM` host service handler MUST 使用 host-call 上下文中的 `pluginID` 作为 provider 请求来源
- **AND** 该来源 MUST 与智能中心调用日志和宿主服务审计中的来源插件保持一致

#### Scenario: 宿主内部调用不伪造插件来源

- **WHEN** 宿主内部模块直接使用 `AI().Text()` 发起文本生成
- **THEN** 文本能力 service MUST 保持来源为空或使用规范定义的宿主来源标识
- **AND** 宿主内部调用 MUST NOT 被记录为任意源码插件或动态插件来源
