## Context

LinaPro 当前已经具备源码插件、动态插件、插件宿主服务、组织能力和租户能力等边界治理基础。本次需求不是给管理工作台增加一个孤立页面，而是为系统自身、业务模块和插件提供可复用的文本 `AI` 能力入口。按照项目规范，`apps/lina-core`必须保持核心宿主边界，不能拥有渠道、模型、档位和调用日志这些具体业务存储；具体管理能力应由官方源码插件承载。

参考 `/Users/john/Workspace/github/gqcn/agent-box` 的渠道、模型和能力档位设计时，只吸收产品对象和交互经验，不复用其“主应用直接拥有 `AI` 能力实现”的边界。LinaPro 的目标架构采用“宿主定义抽象能力，插件提供实现”的模式，类似组织和租户能力。

本次仅实现文本相关 `AI` 接口，但数据模型、请求契约和日志必须保留 `capabilityType`、`capabilityMethod`、协议和 host service 方法扩展空间，后续可以自然增加图片、音频、向量等能力。模型管理不再维护能力方法声明；模型是否适配某个能力方法由管理员在档位选择和测试过程中判断。用户补充的 `thinkingEffort` 必须进入文本请求契约；平台侧将其抽象为可选枚举参数，而不是把某一家渠道的请求结构固化为框架契约。

## Goals / Non-Goals

**Goals:**

- 新增官方源码插件 `linapro-ai-core`，菜单为“智能中心”，包含“渠道”“档位管理”“调用日志”三个页面。
- 在 `apps/lina-core`新增文本 `AI` 抽象能力 `framework.ai.text.v1`，让宿主模块、源码插件和动态插件可以按档位调用文本生成。
- 首期固定提供 `text.generate` 下的 `basic`、`standard`、`advanced` 三个文本能力档位，并允许每个档位绑定一个主渠道模型。
- 渠道页面直接维护渠道和渠道支持的模型，包含增删改查、模型新增、模型同步和引用保护。
- 调用日志只记录最小审计与监控字段，支持分页筛选，不保存完整输入、完整输出、完整密钥或敏感链路内容。
- 文本请求支持 `messages`、`purpose`、`tier`、`maxOutputTokens`、`temperature`、可选 `thinkingEffort` 和 `metadata`。
- `thinkingEffort` 使用平台统一枚举 `low`、`medium`、`high`、`xhigh`、`max`；模型管理不声明支持范围，适配器按渠道能力映射、忽略或拒绝。
- 动态插件通过 `hostServices` 申请 `ai.text.generate` 权限，并由宿主统一授权、调用和审计。
- 设计时明确性能、数据权限、缓存一致性、`i18n`、插件本地规范、跨平台工具和测试影响。

**Non-Goals:**

- 不实现图片、音频、向量、重排、多模态消息、工具调用或流式输出。
- 不实现多模型自动 fallback、按成本或健康状态动态路由、配额、计费、预算或复杂网关策略。
- 不把渠道、模型、档位和调用日志表放入 `apps/lina-core`。
- 不要求所有渠道都支持 `thinkingEffort`，也不在框架契约中暴露渠道专有字段。
- 不保存模型推理过程、隐藏思考内容、完整 prompt、完整 response 或业务输入原文。
- 不为主框架新增长期维护脚本或平台专属工具入口。

## Decisions

### 1. 模块命名为“智能中心”

菜单使用“智能中心”，子菜单为“渠道”“档位管理”“调用日志”。该名称比“AI 能力中心”更适合作为长期顶级入口，既能覆盖当前文本 `AI` 能力，也能容纳后续图片、音频、向量、智能体等能力治理。

备选方案：

- “AI 能力中心”：表达准确，但名称偏技术，后续非纯 `AI API` 能力也会显得窄。
- “智能管理”：偏管理动作，不容易表达系统能力调用入口。
- “模型中心”：会让用户误解为只管理模型列表。

### 2. 宿主只定义 `framework.ai.text.v1` 抽象能力

`apps/lina-core`新增文本能力消费接口，职责限定为：

- `Available(ctx)`：返回当前文本能力是否有可用 provider。
- `Status(ctx)`：返回能力 ID、可用性、provider 插件状态和不可用原因。
- `GenerateText(ctx, request)`：按 `purpose + tier` 执行文本生成；能力方法固定为 `text.generate`，由 Go 契约中的 `CapabilityTypeText` 与 `CapabilityMethodGenerate` 表达，不允许调用方传入任意字符串改变文本生成方法。

核心契约使用自有 DTO，不返回插件内部 `DAO`、`DO`、`Entity`、缓存对象或渠道密钥结构。`lina-core`不创建具体 `AI` 配置表，不提供工作台页面，不硬编码渠道协议实现。

官方插件 `linapro-ai-core`通过 provider factory 声明 `framework.ai.text.v1` 实现。插件禁用、卸载或 provider 冲突时，消费方通过 `Available` 和 `Status` 获取降级状态；可选调用方隐藏入口或返回明确业务错误，不能出现宿主 500。

### 3. 首期文本能力采用消息数组契约

文本生成请求使用 `messages`，不使用单一 `prompt` 字段：

```text
TextGenerateRequest
  purpose string
  tier basic | standard | advanced
  messages []TextMessage
  maxOutputTokens int
  temperature *float64
  thinkingEffort *ThinkingEffort
  metadata map[string]string
```

`TextMessage` 首期只支持 `role` 和纯文本 `content`。这样能覆盖系统提示、用户输入和后续多轮上下文，同时为未来多模态 content block 保留结构扩展。`metadata` 只允许短字符串键值，用于调用来源、业务请求 ID、功能场景和审计关联，不允许承载大段 prompt。

响应返回：

```text
TextGenerateResponse
  text string
  tier string
  providerName string
  modelName string
  protocol string
  usage.inputTokens int
  usage.outputTokens int
  latencyMs int
  generatedAt int64
```

`generatedAt` 等公开时间点统一使用 Unix 毫秒时间戳。后续如果增加流式输出，应新增 `framework.ai.text.v2` 或独立 streaming 方法，不破坏 `v1` 的同步响应语义。

### 4. `thinkingEffort` 是平台抽象参数

`thinkingEffort` 是可选枚举，取值为 `low`、`medium`、`high`、`xhigh`、`max`。该字段表示调用方期望模型投入的推理强度，而不是要求模型暴露推理内容。设计规则：

- 模型管理不声明是否支持 `thinkingEffort`，也不维护支持的枚举子集。
- 调用未传 `thinkingEffort` 时，插件使用档位默认值或模型默认行为。
- 调用传入枚举范围内的 `thinkingEffort` 时，管理面只校验枚举合法性，不基于模型声明预先拒绝。
- 适配器负责把平台枚举映射到渠道协议字段；目标渠道无对应能力时不得发送专有字段，并通过测试调用或运行时调用返回结构化错误。
- 调用日志只记录请求的 `thinkingEffort` 和实际应用值，不记录模型思考过程。

默认建议：

| 档位 | 默认 `thinkingEffort` | 说明 |
|------|-----------------------|------|
| `basic` | 空或 `low` | 简单文本、摘要、提交信息等低成本场景 |
| `standard` | `medium` | 常规代码生成、代码解释、代码优化 |
| `advanced` | `high` | 复杂代码生成和跨文件推理；`xhigh`、`max` 建议仅在真实测试或运行调用确认可用后配置 |

### 5. 能力方法是档位身份的一部分

能力档位不是只按 `capabilityType + tier` 唯一，而是按 `capabilityType + capabilityMethod + tierCode` 唯一。`capabilityType` 表示能力族，例如 `text`、`image`、`embedding`、`audio`；`capabilityMethod` 表示该能力族下的具体方法，例如 `generate`、`create`、`transcribe`、`synthesize`。首期只 seed `text.generate` 的 `basic`、`standard`、`advanced` 三档。

档位管理页面可以按能力方法分组展示或筛选，但 `capabilityType` 与 `capabilityMethod` 是档位记录的不可变身份字段，不进入档位编辑表单。编辑动作只能修改启用状态、主绑定、默认参数和测试结果。后续新增 `image.generate`、`embedding.create`、`audio.transcribe`、`audio.synthesize` 时，应在同一档位表内 seed 对应能力方法下的三档记录，而不是复制一套 Text 专属表。

Go 契约层必须显式定义能力方法常量和方法语义。`framework.ai.text.v1` 的 `GenerateText` 永远映射到 `CapabilityTypeText + CapabilityMethodGenerate`；动态插件 `ai.text.generate` 也必须映射到同一能力方法。插件端档位解析缓存、调用日志和档位唯一约束使用 `capabilityType + capabilityMethod` 作为能力范围，避免后续 `audio.transcribe` 与 `audio.synthesize` 因同属 `audio` 而共享错误档位；模型管理和档位候选查询不得把模型能力声明作为筛选或限制来源。

### 6. 官方插件拥有渠道、模型、档位和日志存储

`linapro-ai-core`维护插件自有 SQL，表名使用插件命名空间：

```text
plugin_linapro_ai_provider
  id
  name
  website_url
  remark
  openai_base_url
  anthropic_base_url
  api_key_secret_ref
  enabled
  created_at
  updated_at
  deleted_at

plugin_linapro_ai_model
  id
  provider_id
  endpoint_id
  model_name
  protocol              -- openai | anthropic
  source                -- manual | api
  enabled
  created_at
  updated_at
  deleted_at

plugin_linapro_ai_tier
  id
  capability_type       -- text
  capability_method     -- generate
  code                  -- basic | standard | advanced
  display_name
  description
  default_effort
  enabled
  sort_order
  created_at
  updated_at

plugin_linapro_ai_tier_binding
  id
  tier_id
  provider_id
  model_id
  priority              -- 0 为主绑定，后续 fallback 预留
  enabled
  created_at
  updated_at
  deleted_at

plugin_linapro_ai_invocation
  id
  request_id
  capability_type
  capability_method
  purpose
  tier_code
  source_plugin_id
  tenant_id
  user_id
  provider_id
  model_id
  provider_name
  model_name
  protocol
  thinking_effort
  status
  input_tokens
  output_tokens
  latency_ms
  error_code
  error_summary
  created_at
```

密钥不得明文返回。存储优先使用宿主 secret 能力或等价加密存储，仅在插件表中保留 `api_key_secret_ref` 或脱敏摘要。渠道和模型删除必须检查档位绑定引用；被引用时拒绝删除，并返回可本地化业务错误。

固定档位和用户可见枚举显示必须接入字典或插件 `i18n` 资源治理；后端实现使用 Go 命名类型和常量，避免散落字符串。`capabilityType`、`capabilityMethod`、`protocol`、`source`、`status`、`thinkingEffort` 等枚举语义在代码中必须集中定义。

### 7. 管理 API 使用 REST 资源语义

管理 API 由 `linapro-ai-core` 注册，资源语义保持以下形态。若实现期插件统一 API 前缀要求包含插件路径，则只调整挂载前缀，不改变资源名。

```text
GET    /api/ai/providers
POST   /api/ai/providers
GET    /api/ai/providers/{id}
PUT    /api/ai/providers/{id}
DELETE /api/ai/providers/{id}
GET    /api/ai/providers/{id}/models
POST   /api/ai/providers/{id}/models
POST   /api/ai/providers/{id}/models/sync
PUT    /api/ai/models/{id}
DELETE /api/ai/models/{id}
GET    /api/ai/tiers
PUT    /api/ai/tiers/{code}
POST   /api/ai/tiers/{code}/test
GET    /api/ai/invocations
```

列表接口必须分页、可筛选并返回当前页面所需的最小投影。渠道列表通过当前页渠道 `ID` 批量返回模型摘要、脱敏密钥和端点投影，不允许前端对每个渠道逐项查询模型详情。档位列表按 `capabilityType + capabilityMethod` 一次返回 `basic`、`standard`、`advanced` 及主绑定投影。调用日志按 `created_at DESC` 分页，支持按 `capabilityType`、`capabilityMethod`、`purpose`、`tier`、`status`、`providerId`、`modelId` 和时间范围过滤。

### 8. 页面交互聚焦配置效率和可诊断性

“渠道”页面以表格为主，支持渠道新增、编辑、删除、启停和独立新增模型。渠道新增和编辑抽屉只维护渠道字段；列表模型列使用后端批量摘要投影渲染，避免进入页面后出现前端瀑布式查询。

“档位管理”页面以当前能力方法下的三档配置为第一屏核心内容，首期固定展示 `text.generate` 的 `basic`、`standard`、`advanced`，并使用稳定顺序展示。每档展示启用状态、渠道、模型、协议、`thinkingEffort` 默认值、最近测试结果和保存入口。模型选择只按渠道和启用状态过滤，展示该渠道下所有可用模型；默认 `thinkingEffort` 只校验枚举合法性，模型/协议是否支持由测试调用或运行时调用返回结构化结果。

“调用日志”页面使用表格、筛选表单和详情抽屉。详情只展示调用摘要、错误摘要、用量和耗时，不展示完整输入输出。日志数据默认平台管理员可见；如果后续开放租户侧自查，必须新增租户可见性规格。

前端实现必须复用现有 `Vben`、`vxe-table`、表单、弹窗、抽屉、操作列和 `IconifyIcon` 模式。插件禁用、未安装或无权限时，菜单和页面入口必须完全隐藏。

### 9. 动态插件通过 `ai.text.generate` 使用文本能力

动态插件在 `plugin.yaml` 中声明：

```yaml
hostServices:
  - service: ai
    methods:
      - text.generate
    resources:
      - ref: purpose:content.summary
        attributes:
          defaultTier: basic
          maxOutputTokens: "1024"
```

`hostServices` 声明是权限申请，不是自动授权。运行时调用必须同时通过 service/method 校验和资源授权校验。资源标识以 `purpose:<name>` 表达调用场景，宿主可以限制插件只能调用已授权用途和最大输出规模。动态插件调用最终进入同一个 `framework.ai.text.v1` 消费服务，并解析为 `text.generate` 能力方法；`pluginbridge`只承担 transport 和 payload 编解码。

源码插件和宿主模块通过显式注入的文本能力 service 调用，不 import `linapro-ai-core/backend/internal/**`，也不直接读取插件表。

### 10. 档位解析使用受控缓存

文本调用路径会频繁解析 `capabilityType + capabilityMethod + tier` 到渠道模型绑定。实现应在 `linapro-ai-core` provider 内维护只读解析缓存：

- 权威数据源：`linapro-ai-core` 插件数据库表。
- 缓存内容：档位、主绑定、渠道模型公开元数据和 secret 引用，不缓存完整 API key。
- 一致性模型：写后显式失效，读取 miss 时重建；允许极短暂陈旧但必须可恢复。
- 失效触发：渠道、模型、档位、绑定启停、删除、更新和插件启停成功后。
- 集群同步：`cluster.enabled=false` 使用本地失效；`cluster.enabled=true` 通过宿主插件运行时修订、共享修订号、事件广播或等价协调机制使其他节点观察到失效。
- 最大陈旧时间：正常写后失效应在当前请求完成后生效；兜底 TTL 不超过 30 秒。
- 故障降级：缓存不可用时直接读取数据库；数据库不可用时返回结构化不可用错误并记录失败日志。

缓存刷新不得清空无关插件、语言包、路由或前端 bundle 缓存。

### 11. 权限和数据可见性

智能中心管理面属于平台配置控制面。渠道、模型、档位和调用日志管理 API 除插件菜单/按钮权限外，还必须要求平台上下文；租户上下文和代管租户上下文不得执行平台配置动作。

建议权限命名：

```text
ai:provider:list
ai:provider:create
ai:provider:update
ai:provider:delete
ai:tier:list
ai:tier:update
ai:tier:test
ai:invocation:list
```

动态插件调用 `ai.text.generate` 还需要 `host:ai:text` 或等价 host service 授权分类。缺少授权时必须在执行外部渠道调用前拒绝，不写入渠道请求。

调用日志列表首期仅平台管理员可见。日志写入可以记录 `tenant_id`、`user_id`、`source_plugin_id` 和 `purpose` 作为审计投影，但不得通过日志响应泄露租户外业务输入存在性。

### 12. 外部协议适配保持窄边界

首期插件实现 OpenAI-compatible 和 Anthropic-compatible 文本适配器。适配器只接收统一文本请求和已解析模型配置，返回统一文本结果和 usage。渠道 API 地址需要支持 SDK 风格 base URL、根域名和完整资源端点规范化，避免重复拼接资源路径。

模型同步使用渠道协议对应的模型列表接口，失败时保留已有手工模型和被引用模型。模型同步结果只写入模型名称、协议、默认端点和来源，不自动推断能力方法、token 上限或 thinking 支持。

### 13. 影响分析

- `i18n`：新增用户可见菜单、页面、按钮、表单、表格、错误和 API 文档。实现期必须先读取插件 `plugin.yaml` 的 `i18n.enabled`，若启用则维护插件 `manifest/i18n/<locale>/` 和 `apidoc` 资源；若未启用，必须在任务记录中说明单语言插件判断。
- 缓存一致性：档位解析缓存是关键运行时数据，设计中已定义权威源、失效触发、集群同步、TTL 和降级策略。
- 数据权限：管理面要求平台上下文；日志首期平台可见；动态插件调用受 host service 授权和 purpose 资源限制。
- 插件边界：修改 `apps/lina-plugins/linapro-ai-core/` 前必须检查该插件根目录 `AGENTS.md`；插件业务表、SQL、i18n、前端和后端资源不得回流宿主目录。
- API 契约：新增 API 必须使用 REST 方法、完整 `g.Meta`、`dc`、`eg`、权限标签和 Unix 毫秒时间字段。
- 后端 Go：能力 service、provider adapter、controller、host service handler 和缓存组件必须显式依赖注入，不能在请求路径临时 `New()` 关键服务。
- 数据库：插件 SQL 使用 PostgreSQL 14+ 源语法，幂等 DDL/Seed，合理索引，禁止写入自增 ID，软删除和时间字段遵守 GoFrame 约束。
- 前端：新增页面和路由必须复用现有前端适配层和组件体系，涉及用户可观察行为必须补充 E2E。
- 开发工具：本次设计不新增或修改长期维护脚本；若实现期需要新增生成或治理入口，必须优先使用跨平台 Go 工具。

## Risks / Trade-offs

- [Risk] 固定三档可能不能覆盖所有业务场景。→ Mitigation：三档作为稳定系统契约，调用方通过 `purpose` 表达具体场景；后续可增加 purpose 策略而不破坏 tier 枚举。
- [Risk] 首期只有主绑定，渠道不可用时相关功能不可用。→ Mitigation：页面提供测试和状态展示，表结构预留 `priority`，后续可新增 fallback 策略。
- [Risk] `thinkingEffort` 不同渠道语义不一致。→ Mitigation：平台只定义抽象枚举，模型管理不维护支持范围；适配器按协议映射，不支持时通过测试或运行时调用返回结构化错误。
- [Risk] 调用日志可能泄露敏感业务输入。→ Mitigation：日志只保存摘要、状态、耗时、用量和错误摘要，禁止保存完整输入输出。
- [Risk] 档位配置变更后集群节点短暂读到旧绑定。→ Mitigation：写后作用域失效、共享修订或事件广播、短 TTL 和数据库兜底读取。
- [Risk] 官方插件实现过重，污染核心宿主。→ Mitigation：`lina-core`只保留 `framework.ai.text.v1` 和 host service 适配，具体 UI、SQL、渠道协议实现均在插件内闭环。

## Migration Plan

1. 新增 `linapro-ai-core` 源码插件清单、菜单、权限、SQL、后端和前端目录。
2. 插件安装 SQL 创建自有表和索引，并 seed `text` 的 `basic`、`standard`、`advanced` 三个档位。
3. 宿主启动时按插件生命周期同步菜单、权限、资源和源码插件 provider factory。
4. 初次安装后不自动创建渠道和模型；未配置档位显示“未配置”状态。
5. 回滚代码时保留插件业务表不会影响宿主核心能力；卸载插件时默认只清理治理资源，不删除业务数据，除非管理员显式执行清理。

## Open Questions

- `linapro-ai-core` 插件是否默认启用 `i18n.enabled`，还是先作为单语言官方插件交付？
- 渠道 API key 首期使用宿主 secret service 还是插件自有加密字段，需要在实现期根据现有 secret 能力确认。
- 调用日志首期是否需要开放按租户过滤的只读视图？当前设计默认仅平台管理员可见。
