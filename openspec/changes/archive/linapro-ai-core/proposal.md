## Why

LinaPro 定位为面向可持续交付的 `AI` 原生全栈框架，系统自身和插件都需要一个稳定、可治理的多模态 `AI` 能力入口，而不是让每个业务模块直接感知渠道协议、模型差异、密钥和调用审计细节。当前需要将文本、图片、向量、音频、视觉、文档、安全审核和视频等能力纳入同一个可治理的 `AI` 能力体系，同时保持 `apps/lina-core` 核心宿主边界——具体渠道、模型、档位和调用日志管理由官方源码插件 `linapro-ai-core` 承载，宿主只定义抽象能力契约。

## What Changes

- 在 `apps/lina-core` 中通过 `AI()` 命名空间暴露类型化 `AI` 子能力：`Text()`、`Image()`、`Embedding()`、`Audio()`、`Vision()`、`Document()`、`Safety()` 和 `Video()`，每个子能力维护独立 DTO、错误、状态和 provider 契约，根能力目录只暴露 `AI()` 入口。
- 将文本能力包从 `capability/aitext` 迁移到 `capability/ai/aitext`，保持 `framework.ai.text.v1` 能力 ID 和行为不变；拆分消费请求与 provider 内部请求，使 `SourcePluginID` 由 plugin-scoped service 注入。
- 新增官方源码插件 `linapro-ai-core`（菜单为"智能中心"），提供渠道管理、模型管理、档位管理、调用日志页面和 provider adapter。
- 渠道模型采用可扩展 provider endpoint 表，支持 OpenAI-compatible、Anthropic-compatible、Voyage-compatible 等多协议端点；模型管理只维护模型身份，不维护能力方法声明。
- 档位管理按能力类型组织，固定提供 `basic`、`standard`、`advanced` 三档，以 `capabilityType + capabilityMethod + tierCode` 作为唯一身份；调用参数由请求传入，不在智能中心持久化默认参数 JSON。
- 调用日志覆盖多模态方法，支持分页筛选和日志清理，只记录最小审计摘要，不保存完整输入输出或大对象内容。
- 扩展动态插件 `ai` host service 方法，新增多模态方法级授权、资源策略、payload 上限、资产引用和脱敏审计。
- 大对象结果统一使用 `assetRef` 或受控临时资产引用，禁止无上限 base64 进入 JSON 或 WASM host call。
- 渠道异步协议通过 `ProviderOperationRef` 表达，业务异步任务、队列、进度、重试和资产归属由调用方业务模块实现。
- `computer.act`、`ui.operate` 和等价 UI 控制能力不在本轮设计范围内。
- 模型名末尾 `[...]` 工具后缀在向渠道发起请求时自动去除，平台侧模型身份保持原值。
- 渠道请求兼容 `/v1` 路径自动重试和缓存修正机制。

## Capabilities

### New Capabilities

- `linapro-ai-core`：定义宿主 `AI` 命名空间聚合入口、文本子能力迁移、多模态子能力扩展、插件来源身份注入和未来子能力扩展边界。
- `linapro-ai-core-multimodal`：定义 `linapro-ai-core` 智能中心对多模态 provider、endpoint、模型身份、档位、调用参数边界、调用日志、缓存、管理页面和 provider adapter 的扩展要求。

### Modified Capabilities

- `linapro-ai-core-host-service`：扩展动态插件 `ai` host service，新增多模态方法声明、授权资源、策略属性、资产引用、provider operation 查询和最小审计要求。

## Impact

- 影响 `apps/lina-core/pkg/plugin/capability/ai`，新增类型化 `AI` 子能力接口、公共值对象、状态和 provider 契约。
- 影响 `apps/lina-core/pkg/plugin/pluginbridge`、动态插件 guest SDK、WASM host service handler 和 host service 授权校验，新增 `ai.*` 方法映射和 payload 边界。
- 新增官方源码插件目录 `apps/lina-plugins/linapro-ai-core/`，包含 `plugin.yaml`、`backend/`、`frontend/`、`manifest/sql/` 和 `i18n` 资源。
- 影响智能中心插件数据库：provider endpoint、模型身份、方法档位、调用日志和 provider operation 投影均使用幂等 SQL、合理索引和缓存失效。
- 影响 API 契约：新增渠道、模型、档位、能力方法、调用日志、provider operation 等管理 API，满足 REST 语义、分页/上限、Unix 毫秒时间和完整文档标签。
- 影响数据权限和授权：智能中心管理面作为平台配置控制面；动态插件调用按 service、method、resource、purpose 和 payload 策略授权。
- 影响缓存一致性：能力方法到 provider/model/tier 的解析缓存属于关键运行时配置，必须定义权威源、失效触发、集群同步、最大陈旧时间和故障降级。
- 影响前端和 E2E：智能中心页面支持多模态配置、provider endpoint 管理、模型身份管理和调用日志治理，并补充对应 E2E 覆盖。
- 影响 `i18n`：`linapro-ai-core` 已启用 `i18n.enabled: true`，新增菜单、按钮、表单、表格、错误和 API 文档源文本必须维护插件运行时语言包和 `apidoc` 翻译资源。
