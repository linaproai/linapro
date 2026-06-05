## Why

`add-ai-text-capability-center`和`refactor-ai-capability-namespace`已经建立了`AI().Text()`、`framework.ai.text.v1`和`linapro-ai-core`智能中心的基础。下一步需要把图片、向量、音频、视觉、文档、安全审核和视频等能力纳入同一个可治理的`AI`能力体系，避免业务模块直接感知渠道协议、模型差异、密钥和调用审计细节。

本次设计也需要明确边界：`lina-core`和`linapro-ai-core`只管理`AI`能力、元数据、渠道模型适配和最小调用治理，不承载“生成视频任务”“业务后台队列”“用户进度通知”等具体业务场景实现。

## What Changes

- 在`AI()`命名空间下新增类型化子能力设计：`Image()`、`Embedding()`、`Audio()`、`Vision()`、`Document()`、`Safety()`和`Video()`，保持各自 DTO、错误、状态和 provider 契约独立。
- 新增能力方法建议：`image.generate`、`image.edit`、`embedding.create`、`audio.transcribe`、`audio.synthesize`、`vision.analyze`、`document.analyze`、`document.cite`、`safety.moderate`、`video.generate`、`video.edit`、`video.extend`、`video.operation.get`和可选`video.operation.cancel`。
- 保持`computer.act`明确不进入本轮设计，不新增 UI 操作代理、桌面控制或浏览器控制类宿主能力。
- 为二进制和大对象结果定义统一资产引用边界，图片、音频和视频结果 MUST 返回`assetRef`或受控临时资产引用，MUST NOT 在 JSON、WASM host call 或插件调用结果中返回大段 base64。
- 为渠道异步协议定义`ProviderOperationRef`边界。`linapro-ai-core`可以适配渠道的`create -> poll -> download`协议，但业务异步任务、队列、状态、重试、通知和资产归属 MUST 由调用方业务模块实现。
- 扩展`linapro-ai-core`元数据模型，支持通用 provider endpoint、模型身份、能力方法档位、最小调用日志和 provider operation 投影；调用参数由每次`AI`请求传入，不在智能中心持久化默认参数 JSON。
- 将 provider 协议端点直接建模为可扩展的`provider_endpoint`或等价插件自有表，provider 主表不承载按协议命名的固定端点列，避免每新增协议就污染 provider 主表。
- 扩展动态插件`ai` host service 方法和授权模型，新增多模态方法级授权、资源策略、payload 上限、资产引用和脱敏审计要求。
- 智能中心管理页面从文本能力扩展为多模态配置；档位管理当前版本按能力类型`Tab`组织交互，内部仍映射到方法级档位身份，不新增具体业务任务页面，例如不提供`/ai/video-jobs`。
- 本轮实现范围建议分期：优先完成通用契约、元数据重构、host service DTO 和 OpenAI/Anthropic/Voyage 等 provider adapter 的窄实现；视频 adapter 可先完成协议形态和 operation 投影，不强制完整业务编排。

## Capabilities

### New Capabilities

- `framework-ai-multimodal-capabilities`：定义`AI`命名空间下图片、向量、音频、视觉、文档、安全审核和视频等类型化框架能力、请求响应契约、资产结果、provider operation 引用和降级语义。
- `linapro-ai-core-multimodal-plugin`：定义`linapro-ai-core`智能中心对多模态 provider、endpoint、模型身份、档位、调用参数边界、调用日志、缓存、管理页面和 provider adapter 的扩展要求。

### Modified Capabilities

- `plugin-host-service-extension`：扩展动态插件`ai` host service，新增多模态方法声明、授权资源、策略属性、资产引用、provider operation 查询和最小审计要求。

## Impact

- 影响`apps/lina-core/pkg/plugin/capability/ai`或等价宿主能力目录，新增类型化`AI`子能力接口、公共值对象、状态和 provider 契约。
- 影响`apps/lina-core/pkg/plugin/pluginbridge`、动态插件 guest SDK、WASM host service handler 和 host service 授权校验，新增`ai.*`方法映射和 payload 边界。
- 影响`apps/lina-plugins/linapro-ai-core/`，需要扩展插件自有 SQL、DAO 生成、后端 API、service、provider adapter、管理页面、菜单/权限和`i18n`资源。
- 影响智能中心插件数据库：需要为 provider endpoint、模型身份、方法档位、调用日志和 provider operation 投影设计幂等 SQL、索引和缓存失效；不新增方法默认参数表。
- 影响 API 契约：新增或修改渠道、模型、档位、能力方法、调用日志、provider operation 查询等管理 API，必须满足 REST 语义、分页/上限、Unix 毫秒时间和完整文档标签。
- 影响数据权限和授权：智能中心管理面继续作为平台配置控制面；动态插件调用按 service、method、resource、purpose 和 payload 策略授权；日志首期仅平台可见。
- 影响缓存一致性：能力方法到 provider/model/tier 的解析缓存属于关键运行时配置，必须定义权威源、失效触发、集群同步、最大陈旧时间和故障降级。
- 影响前端和 E2E：智能中心页面需要支持档位能力类型`Tab`、provider endpoint 管理、模型身份管理和资产/operation 摘要展示，并补充对应 E2E 覆盖。
- 影响`i18n`：`linapro-ai-core`已启用`i18n.enabled: true`，新增菜单、按钮、表单、表格、错误和 API 文档源文本必须维护插件运行时语言包和`apidoc`翻译资源。
