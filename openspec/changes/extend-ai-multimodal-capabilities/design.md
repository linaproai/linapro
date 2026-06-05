## Context

`add-ai-text-capability-center`已经建立`framework.ai.text.v1`文本能力和`linapro-ai-core`智能中心，`refactor-ai-capability-namespace`已经把消费入口收敛为`AI().Text()`。现有基础说明了两个边界：

- `lina-core`负责稳定、可复用的宿主能力接缝和动态插件 host service 协议。
- `linapro-ai-core`负责渠道、模型、档位、调用日志和 provider adapter，不把具体存储回流到宿主。

本次变更在这个基础上扩展多模态能力。需要覆盖的能力包括图片、向量、音频、视觉、文档、安全审核和视频；明确排除`computer.act`。视频需要特别处理：渠道可能采用异步生成协议，但框架和智能中心不能变成业务任务中心。业务异步任务、队列、进度展示、通知、重试和资产归属必须由调用`AI`能力的业务模块负责。

外部能力口径截至`2026-06-02`：OpenAI 具备图片、音频、视觉、文件输入、安全审核和 Sora 视频相关 API；Anthropic/Claude 具备视觉、PDF、工具类能力，但未看到原生视频生成 API；Voyage 等渠道支持多模态 embedding。渠道能力变化快，因此平台契约必须按 provider-generic 方法设计，而不是绑定某一家渠道的长期产品形态。

## Goals / Non-Goals

**Goals:**

- 在`AI()`命名空间下定义`Image()`、`Embedding()`、`Audio()`、`Vision()`、`Document()`、`Safety()`和`Video()`类型化子能力。
- 让每个子能力维护独立 DTO、错误、provider 契约和方法常量，拒绝弱类型`AI().Invoke(method, payload)`网关。
- 扩展动态插件`service: ai`的 host service 方法，形成方法级授权和资源策略。
- 定义二进制和大对象结果的`assetRef`边界，禁止大段 base64 进入 JSON 或 WASM host call。
- 定义`ProviderOperationRef`，让`linapro-ai-core`能适配渠道异步协议，但不实现具体业务任务。
- 将`linapro-ai-core`渠道端点从固定列重构为可扩展 endpoint 模型，支撑 OpenAI-compatible、Anthropic-compatible、Voyage-compatible 等协议。
- 扩展能力方法档位、调用参数边界、调用日志、缓存失效、平台权限和`i18n`治理；模型管理只维护模型身份，不维护能力方法声明或默认参数模板。

**Non-Goals:**

- 不实现`computer.act`、桌面控制、浏览器控制或 UI 自动操作类能力。
- 不把视频生成、批量转写、文档处理等业务任务表放进`lina-core`或`linapro-ai-core`。
- 不新增`/api/ai/video-jobs`、`/api/ai/audio-jobs`等业务任务 API。
- 不让业务模块直接读取`linapro-ai-core`的`DAO`、`DO`、`Entity`、provider adapter 或密钥结构。
- 不要求所有 provider 都实现所有能力；能力可用性由管理员选择、档位测试、运行时调用结果和 provider adapter 决定。
- 不在调用日志中保存完整输入、完整输出、音视频原文、文件内容、渠道响应原文或密钥。
- 不把 web search、file search、code execution 作为本轮主能力实现；这些可作为后续 agent/tool 能力另行设计。

## Decisions

### 1. `AI`命名空间继续采用类型化子能力

目标形态：

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

`ai.Service`只负责命名空间聚合和插件身份绑定。每个子能力包拥有自己的请求、响应、状态、provider factory 和错误语义。这样新增`Image()`或`Embedding()`只扩大`AI`命名空间，不污染根`capability.Services`。

拒绝方案是引入：

```go
Invoke(ctx, method string, payload any) (any, error)
```

该方案会把授权、审计、资产边界、向量维度、音频格式、视频 operation 状态都推到运行时字符串分发里，契约不清晰，也不利于插件静态声明和 E2E 覆盖。

### 2. 能力方法按 provider-generic 语义命名

本轮建议方法集合：

| 能力族 | 方法 |
|--------|------|
| `image` | `generate`、`edit` |
| `embedding` | `create` |
| `audio` | `transcribe`、`synthesize` |
| `vision` | `analyze` |
| `document` | `analyze`、`cite` |
| `safety` | `moderate` |
| `video` | `generate`、`edit`、`extend`、`operation.get`、`operation.cancel` |

`capabilityType + capabilityMethod`仍然是档位、调用日志和授权的核心身份。`vision.analyze`和`document.analyze`不塞进`text.generate`，避免文本能力被多模态 content block 侵蚀。

`video.operation.get`和`video.operation.cancel`表达的是 provider operation 协议适配，不是业务任务查询。业务模块如果要给用户展示生成进度，应创建自己的业务任务并保存`operationRef`。

### 3. 资产结果必须通过引用返回

图片、音频和视频的结果统一使用资产引用：

```text
AssetResult
  assetRef
  mimeType
  sizeBytes
  checksum
  width
  height
  durationMs
  createdAt
```

`assetRef`可以是宿主文件能力、插件资产能力或受控临时对象的引用。具体实现应复用现有文件/存储能力接缝，不把二进制内容放入 host service JSON。对于临时 URL 或 provider download URL，必须先转换为受控引用或受控临时资产投影，避免把渠道认证 URL 暴露给插件或前端。

### 4. 渠道异步只暴露`ProviderOperationRef`

视频和部分音频/图片渠道可能不是同步完成。`linapro-ai-core`可以保存最小 provider operation 投影：

```text
ProviderOperationRef
  operationRef
  capabilityType
  capabilityMethod
  providerName
  modelName
  status
  nextPollAfterMs
  expiresAt
```

`operationRef`必须是不透明引用。调用方不能从中解析 provider 请求 ID、账号、区域、URL 或密钥。业务模块可以保存`operationRef`，随后调用`video.operation.get`查询 provider 进度；但业务 job、业务状态机、重试策略、通知和资产归属由业务模块决定。

这使 AI 层保持“渠道协议适配和能力治理”职责，不承载“生成某类业务视频”这样的场景逻辑。

### 5. `linapro-ai-core`扩展为能力元数据中心，不变成任务中心

`linapro-ai-core`扩展后仍只拥有：

- provider 和 endpoint 配置。
- secret 引用或加密密文。
- 模型身份。
- `capabilityType + capabilityMethod + tier`档位和绑定。
- provider 支持矩阵；调用参数由每次`AI`请求传入，不在智能中心持久化任意默认参数 JSON。
- 最小调用日志。
- provider operation 的最小投影和状态查询适配。

它不拥有：

- 业务任务表。
- 业务对象关联。
- 用户进度页面。
- 业务重试、通知、审批或发布流程。
- 面向某个业务领域的生成记录。

智能中心页面可以展示“provider operation 摘要”用于诊断，但不能替代业务模块的任务管理页面。

### 6. Provider endpoint 作为单一事实来源

provider 主表只保存渠道元数据；协议端点、基础地址和密钥引用必须从初始设计开始收敛到独立 endpoint 表：

```text
plugin_linapro_ai_provider_endpoint
  id
  provider_id
  protocol
  base_url
  secret_ref
  enabled
  metadata_json
  created_at
  updated_at
  deleted_at
```

`protocol`使用受控枚举，例如`openai`、`anthropic`、`voyage`、`openai-compatible`。端点删除必须检查模型和档位引用；被引用时拒绝或要求先解除模型引用。provider 主表保留展示名称、网站、备注和启用状态。

该模型让一个 provider 同时维护多个协议端点，新增协议时只新增 endpoint 记录和受控协议值，不修改 provider 主表结构。

### 7. 模型管理只维护模型身份

模型记录只保存渠道、默认 endpoint、模型名称、协议、来源和启用状态等身份字段。模型管理不保存、展示、筛选或编辑`capabilityType`、`capabilityMethod`、token 上限、输入输出模态、`thinkingEffort`支持范围或 operation 支持状态。

能力方法仍然是档位、调用、日志和 host service 的身份维度，但不是模型管理的事实源。渠道模型列表、档位候选模型查询、档位绑定校验和运行时解析缓存均不得依赖模型能力声明筛选或拒绝模型。模型是否适配某个方法由管理员选择模型后通过测试调用和真实运行结果判断；provider adapter 对协议不支持、参数不支持或渠道错误返回结构化错误。

### 8. 动态插件授权按方法和资源拆分

动态插件仍使用：

```yaml
hostServices:
  - service: ai
    methods:
      - image.generate
      - embedding.create
      - audio.transcribe
```

方法映射到能力分类，例如：

```text
host:ai:image:generate
host:ai:embedding:create
host:ai:audio:transcribe
host:ai:audio:synthesize
host:ai:vision:analyze
host:ai:document:analyze
host:ai:safety:moderate
host:ai:video:generate
```

资源策略用于约束`purpose`、最大输出、最大输入资产数量、最大字节数、允许 mime 类型、默认档位、是否允许 provider operation 和是否允许取消。任一校验失败必须在读取密钥或调用渠道前拒绝。

### 9. 缓存仍以能力方法解析为核心

档位解析缓存从文本扩展到所有方法：

- 权威源：`linapro-ai-core`插件数据库。
- 缓存键：`capabilityType + capabilityMethod + tier + optional purpose`。
- 缓存内容：启用档位、主绑定、provider endpoint、model 和 secret 引用，不缓存密钥明文或调用参数。
- 失效触发：provider、endpoint、model、tier 和 binding 写入成功后。
- 集群同步：`cluster.enabled=true`时使用宿主统一集群协调、修订号、事件广播或等价机制。
- 最大陈旧时间：正常写后失效应在事务成功后生效，兜底 TTL 不超过 30 秒。
- 降级：缓存不可用时读数据库重建；数据库不可用时返回结构化不可用错误。

### 10. 数据权限和平台权限

智能中心仍是平台配置控制面。provider、endpoint、model、tier、binding、log 和 operation 摘要管理 API 必须要求平台上下文和对应权限。日志首期仅平台管理员可见；如果后续要给租户或业务模块自查，必须另立租户可见规格，避免通过日志暴露跨租户业务输入存在性。

动态插件调用不走智能中心管理权限，而走 host service 授权。管理权限和调用权限不能混用。

### 11. `i18n`和前端页面

`linapro-ai-core/plugin.yaml`已启用`i18n.enabled: true`。新增菜单、表单字段、表格列、按钮、状态、错误和 API 文档源文本必须使用英文源文本，并维护插件`manifest/i18n/<locale>/`和`manifest/i18n/<locale>/apidoc/`资源。前端表格、抽屉、弹窗、操作列和筛选项继续使用现有`Vben`和`vxe-table`模式。

页面结构建议：

- “渠道”：provider、endpoint 和模型摘要，支持端点维护和模型维护，不展示或编辑模型能力方法声明。
- “档位管理”：按能力类型`Tab`展示`basic`、`standard`、`advanced`，当前版本每个类型映射到默认能力方法完成绑定和测试；调用参数由具体调用请求传入，不在页面维护默认参数 JSON。
- “调用日志”：支持 capability type、method、provider、model、purpose、status、time range 过滤，只展示脱敏摘要。

不新增“视频任务”“音频任务”页面。

## Risks / Trade-offs

- [Risk] 子能力接口数量增长，`ai.Service`实现面变大。→ Mitigation：每个子能力保持窄接口和 fallback service，根`capability.Services`不再追加`AI*()`方法；测试桩按需实现。
- [Risk] 资产引用依赖现有文件/存储能力成熟度。→ Mitigation：首期定义`assetRef`契约和临时资产投影，具体存储接缝按现有文件能力或插件存储能力实现。
- [Risk] `ProviderOperationRef`可能被误用为业务 job。→ Mitigation：规格明确 operation 只代表渠道协议状态，业务模块必须自建任务和业务归属。
- [Risk] 不同 provider 的多模态能力差异很大。→ Mitigation：模型管理不声明方法能力，管理员通过档位绑定和测试确认适配性；adapter 只实现已支持的 provider 协议，不支持时返回结构化不可用错误。
- [Risk] provider endpoint 模型使模型创建必须选择端点。→ Mitigation：模型 API、前端表单和 E2E 均要求显式`endpointId`，provider endpoint 表是端点配置的单一事实来源。
- [Risk] 多模态调用日志泄露敏感输入。→ Mitigation：只记录摘要、资产引用、状态、用量、耗时和脱敏错误，不保存完整文件、音频、图片或响应正文。
- [Risk] 视频渠道能力变化快。→ Mitigation：视频能力按 provider-generic 方法和 operation 引用设计，不绑定 Sora 长期契约。

## Implementation Plan

1. 在现有`AI().Text()`基础上新增子能力包和公共值对象，保持文本能力 ID 和行为不变。
2. 扩展`pluginbridge`host service 常量、校验和 guest SDK，新增多模态`ai.*`方法，但默认未授权。
3. 修改`linapro-ai-core`插件 SQL：新增 provider endpoint、method 参数和 operation 投影表；模型表只保存身份字段，provider 主表不包含固定端点列或密钥引用列。
4. 更新 DAO 生成、后端 API、service、provider adapter 和缓存失效路径。
5. 更新智能中心前端页面、菜单权限、语言包和`apidoc`资源。
6. 按能力分批实现 provider adapter：embedding、vision/document/safety、image/audio、video operation。
7. 回滚时保留插件业务表和日志；宿主能力回退到仅`Text()`时，未启用的子能力返回结构化不可用错误。

## Open Questions

- `assetRef`首期应复用现有文件管理能力、插件 storage host service，还是新增最小`AI`临时资产投影？
- `embedding.create`是否首期只支持文本 embedding，还是同步支持 Voyage 多模态 embedding 的图片/视频输入？
- `video.operation.cancel`是否首期开放给动态插件，还是只允许管理面和业务模块通过 Go capability 调用？
- `safety.moderate`是否作为所有多模态 provider 调用的前置策略，还是仅作为显式能力方法供业务模块调用？
