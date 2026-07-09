## Context

LinaPro 当前已具备源码插件、动态插件、插件宿主服务、组织能力和租户能力等边界治理基础。本次需求不是给管理工作台增加孤立页面，而是为系统自身、业务模块和插件提供可复用的多模态 `AI` 能力入口。按照项目规范，`apps/lina-core` 必须保持核心宿主边界，不能拥有渠道、模型、档位和调用日志这些具体业务存储；具体管理能力应由官方源码插件承载。

已有基础包括：`framework.ai.text.v1` 文本能力、`AI().Text()` 命名空间入口、`linapro-ai-core` 智能中心插件和动态插件 `ai.text.generate` host service 协议。本次变更在此基础上扩展到多模态能力，覆盖图片、向量、音频、视觉、文档、安全审核和视频，明确排除 `computer.act`。

## Goals / Non-Goals

**Goals:**

- 在 `AI()` 命名空间下定义 `Image()`、`Embedding()`、`Audio()`、`Vision()`、`Document()`、`Safety()` 和 `Video()` 类型化子能力，每个子能力维护独立 DTO、错误和 provider 契约。
- 扩展 `linapro-ai-core` 智能中心，支持 provider endpoint 管理、模型身份管理、能力方法档位、调用日志和 provider adapter。
- 将 provider 协议端点建模为可扩展 endpoint 表，支持 OpenAI-compatible、Anthropic-compatible、Voyage-compatible 等多协议。
- 定义 `assetRef` 边界，禁止大段 base64 进入 JSON 或 WASM host call。
- 定义 `ProviderOperationRef`，让 `linapro-ai-core` 适配渠道异步协议，但不实现具体业务任务。
- 扩展动态插件 `ai` host service 方法，形成方法级授权和资源策略。
- 调用参数由每次请求传入，不在智能中心持久化默认参数 JSON。
- 模型管理只维护模型身份，不维护能力方法声明。

**Non-Goals:**

- 不实现 `computer.act`、桌面控制、浏览器控制或 UI 自动操作类能力。
- 不把视频生成、批量转写、文档处理等业务任务表放进 `lina-core` 或 `linapro-ai-core`。
- 不新增 `/api/ai/video-jobs`、`/api/ai/audio-jobs` 等业务任务 API。
- 不让业务模块直接读取 `linapro-ai-core` 的 `DAO`、`DO`、`Entity`、provider adapter 或密钥结构。
- 不要求所有 provider 都实现所有能力；能力可用性由管理员选择、档位测试、运行时调用结果和 provider adapter 决定。
- 不在调用日志中保存完整输入、完整输出、音视频原文、文件内容、渠道响应原文或密钥。
- 不引入弱类型 `AI().Invoke(method, payload)` 网关。

## Decisions

### 1. AI 命名空间采用类型化子能力

```go
type Service interface {
    Text() aitext.Service
    Image() aiimage.Service
    Embedding() aiembedding.Service
    Audio() aiaudio.Service
    Vision() aivision.Service
    Document() aidocument.Service
    Safety() aisafety.Service
    Video() aivideo.Service
}
```

`ai.Service` 只负责命名空间聚合和插件身份绑定，不负责渠道选择、档位解析、授权判断或弱类型方法分发。每个子能力包拥有自己的请求、响应、状态、provider factory 和错误语义。根 `capability.Services` 只暴露 `AI()`，不随 AI 子能力增加而扩展为 `AIImage()`、`AIEmbedding()` 等方法。

### 2. 能力方法按 provider-generic 语义命名

| 能力族 | 方法 |
|--------|------|
| `text` | `generate` |
| `image` | `generate`、`edit` |
| `embedding` | `create` |
| `audio` | `transcribe`、`synthesize` |
| `vision` | `analyze` |
| `document` | `analyze`、`cite` |
| `safety` | `moderate` |
| `video` | `generate`、`edit`、`extend`、`operation.get`、`operation.cancel` |

`capabilityType + capabilityMethod` 是档位、调用日志和授权的核心身份。

### 3. 资产结果通过引用返回

图片、音频和视频结果统一使用 `assetRef`，包含 `mimeType`、`sizeBytes`、`width`、`height`、`durationMs` 等投影。`assetRef` 可以是宿主文件能力、插件资产能力或受控临时对象引用，不把二进制内容放入 host service JSON。

### 4. 渠道异步只暴露 ProviderOperationRef

```text
ProviderOperationRef
  operationRef          -- 不透明引用
  capabilityType
  capabilityMethod
  providerName
  modelName
  status
  nextPollAfterMs
  expiresAt
```

业务模块可以保存 `operationRef` 随后查询 provider 进度，但业务 job、状态机、重试策略、通知和资产归属由业务模块决定。

### 5. Provider endpoint 作为单一事实来源

provider 主表只保存渠道元数据；协议端点、基础地址和密钥引用统一由 `plugin_linapro_ai_provider_endpoint` 表承载。一个 provider 可以同时维护多个协议端点，新增协议时只新增 endpoint 记录和受控协议值，不修改 provider 主表结构。

### 6. 模型管理只维护模型身份

模型记录只保存渠道、默认 endpoint、模型名称、协议、来源和启用状态。模型管理不保存、展示、筛选或编辑 `capabilityType`、`capabilityMethod`、token 上限、输入输出模态、`thinkingEffort` 支持范围。模型是否适配某个方法由管理员通过档位绑定、测试调用和运行时结果判断。

### 7. 消费请求与 provider 请求分离

普通消费方看到的 `GenerateRequest` 不包含 `SourcePluginID`。源码插件通过 `ServicesForPlugin(services, pluginID).AI().Text()` 获取 plugin-scoped service；动态插件由 host-call 上下文提供 `pluginID`。

### 8. 档位按能力方法管理，调用参数由请求传入

档位以 `capabilityType + capabilityMethod + tierCode` 作为唯一身份。每个能力方法拥有 `basic`、`standard`、`advanced` 三档及主绑定。`maxOutputTokens`、图片尺寸、音频格式、视频时长等调用参数由调用方在每次 AI 请求中显式传入，不在智能中心持久化默认参数 JSON。

### 9. 动态插件授权按方法和资源拆分

动态插件 `plugin.yaml` 中 `service: ai` 的方法映射到能力分类，例如 `host:ai:image:generate`、`host:ai:embedding:create`、`host:ai:audio:transcribe`。资源策略用于约束 `purpose`、最大输出、最大输入资产数量、最大字节数、允许 mime 类型、默认档位、是否允许 provider operation 和是否允许取消。

### 10. 档位解析缓存覆盖所有方法

- 权威数据源：`linapro-ai-core` 插件数据库。
- 缓存键：`capabilityType + capabilityMethod + tier`。
- 缓存内容：启用档位、主绑定、provider endpoint、model 和 secret 引用，不缓存密钥明文或调用参数。
- 失效触发：provider、endpoint、model、tier 和 binding 写入成功后通过共享修订号发布失效。
- 集群同步：`cluster.enabled=true` 时使用宿主统一集群协调机制。
- 最大陈旧时间：兜底 TTL 不超过 30 秒。
- 故障降级：缓存不可用时读数据库重建；数据库不可用时返回结构化不可用错误。

### 11. 权限和数据可见性

智能中心管理面属于平台配置控制面。provider、endpoint、model、tier、binding、log 和 operation 摘要管理 API 必须要求平台上下文和对应权限。日志首期仅平台管理员可见。动态插件调用不走智能中心管理权限，走 host service 授权。

### 12. 外部协议适配保持窄边界

实现 OpenAI-compatible、Anthropic-compatible 和 Voyage-compatible 适配器。适配器只接收统一请求和已解析模型配置，返回统一结果和 usage。渠道 API 地址支持 SDK 风格 base URL、根域名和完整资源端点规范化，并在 404 时自动重试 `/v1` 路径并缓存成功修正。模型名末尾 `[...]` 工具后缀在向渠道发起请求时自动去除。

### 13. 智能中心只管理能力元数据，不变成任务中心

`linapro-ai-core` 扩展后只拥有：provider 和 endpoint 配置、secret 引用、模型身份、能力方法档位和绑定、最小调用日志、provider operation 最小投影。不拥有业务任务表、业务对象关联、用户进度页面、业务重试、通知、审批或发布流程。

## Risks / Trade-offs

- [Risk] 子能力接口数量增长，`ai.Service` 实现面变大。→ Mitigation：每个子能力保持窄接口和 fallback service，根 `capability.Services` 不再追加 `AI*()` 方法。
- [Risk] 资产引用依赖现有文件/存储能力成熟度。→ Mitigation：首期定义 `assetRef` 契约和临时资产投影，具体存储接缝按现有文件能力或插件存储能力实现。
- [Risk] `ProviderOperationRef` 可能被误用为业务 job。→ Mitigation：规格明确 operation 只代表渠道协议状态，业务模块必须自建任务和业务归属。
- [Risk] 不同 provider 的多模态能力差异很大。→ Mitigation：模型管理不声明方法能力，管理员通过档位绑定和测试确认适配性；adapter 只实现已支持的 provider 协议，不支持时返回结构化不可用错误。
- [Risk] provider endpoint 模型使模型创建必须选择端点。→ Mitigation：模型 API、前端表单和 E2E 均要求显式 `endpointId`。
- [Risk] 多模态调用日志泄露敏感输入。→ Mitigation：只记录摘要、资产引用、状态、用量、耗时和脱敏错误，不保存完整文件、音频、图片或响应正文。
- [Risk] 视频渠道能力变化快。→ Mitigation：视频能力按 provider-generic 方法和 operation 引用设计，不绑定特定渠道长期契约。
- [Risk] 固定三档可能不能覆盖所有业务场景。→ Mitigation：三档作为稳定系统契约，调用方通过 `purpose` 表达具体场景。
- [Risk] 首期只有主绑定，渠道不可用时相关功能不可用。→ Mitigation：页面提供测试和状态展示，表结构预留 `priority` 供后续 fallback 策略。
- [Risk] 档位配置变更后集群节点短暂读到旧绑定。→ Mitigation：写后作用域失效、共享修订号、短 TTL 和数据库兜底读取。
