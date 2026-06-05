## Why

LinaPro 的定位是面向可持续交付的 `AI` 原生全栈框架，系统自身和插件都需要一个稳定、可治理的文本 `AI` 能力入口，而不是让每个业务模块直接感知渠道、协议、模型和密钥细节。当前已明确需要渠道管理、模型管理、能力档位和调用日志，但这些具体管理能力应按组织、多租户能力的模式由官方插件承载，`apps/lina-core`仅保留稳定抽象接口。

## What Changes

- 新增“智能中心”官方源码插件 `linapro-ai-core`，提供渠道、模型管理、能力档位和调用日志页面。
- 渠道页面直接维护渠道及其支持模型，支持添加、编辑、删除、同步或手工维护模型，并保护被能力档位引用的渠道和模型。
- 能力档位页面固定提供 `basic`、`standard`、`advanced` 三档文本生成能力配置，每档绑定一个主渠道模型，并以 `capabilityType + capabilityMethod + tierCode` 作为唯一身份，为后续图片、音频、向量等能力方法保留扩展边界。
- 调用日志页面记录文本 `AI` 调用的最小审计信息，支持分页、筛选、耗时和用量分析，默认不保存完整输入、完整输出或密钥。
- `apps/lina-core`新增抽象文本能力契约 `framework.ai.text.v1`，只暴露 `Available`、`Status`、`GenerateText` 等稳定接口，不拥有渠道、模型、档位或日志存储。
- 动态插件新增受治理的 `ai` host service，首期仅开放 `text.generate`，并将 `text.generate` 映射为 `capabilityType=text`、`capabilityMethod=generate`；授权资源与调用参数为未来 `image.generate`、`audio.transcribe`、`audio.synthesize`、`embedding.create` 等能力方法预留扩展。
- 文本生成请求预留可选 `thinkingEffort` 参数，枚举参考 Claude Opus effort 级别：`low`、`medium`、`high`、`xhigh`、`max`；模型管理不声明支持范围，是否可用由管理员测试、真实调用和 provider adapter 的结构化错误反馈判断。
- 首期不实现流式输出、工具调用、多模态消息、图片生成、音频处理、向量生成、自动 fallback 路由或成本配额。

## Capabilities

### New Capabilities

- `framework-ai-text-capability`：定义宿主文本 `AI` 抽象能力、请求响应契约、能力状态、降级语义和 `thinkingEffort` 扩展参数。
- `linapro-ai-core-plugin`：定义官方智能中心插件的渠道、模型管理、档位管理、调用日志、插件自有存储、前端页面和权限治理。

### Modified Capabilities

- `plugin-host-service-extension`：新增动态插件 `ai.text.generate` 宿主服务授权、资源声明、调用审计和与 `framework.ai.text.v1` 的适配要求。

## Impact

- 影响 `apps/lina-core/pkg/plugin/capability` 或等价宿主能力目录，新增文本 `AI` 消费契约和动态插件 guest 侧调用入口。
- 影响 `apps/lina-core/pkg/plugin/pluginbridge`、host service 注册表和授权快照，新增 `ai` service family 的 `text.generate` 方法。
- 新增官方源码插件目录 `apps/lina-plugins/linapro-ai-core/`，包含 `plugin.yaml`、`backend/`、`frontend/`、`manifest/sql/` 和按插件 `i18n` 配置维护的资源。
- 新增插件自有表，建议使用 `plugin_linapro_ai_provider`、`plugin_linapro_ai_model`、`plugin_linapro_ai_tier`、`plugin_linapro_ai_tier_binding`、`plugin_linapro_ai_invocation` 等插件命名空间；模型表只保存模型身份字段，档位和调用日志必须保存 `capability_type` 与 `capability_method`。
- 新增插件 API，资源路径以插件 API 前缀承载，管理资源语义采用 `/api/ai/providers`、`/api/ai/tiers`、`/api/ai/invocations` 等名词化路径；若实现期发现官方源码插件统一前缀已有更严格约定，则以插件路由约定为准并保持资源名不变。
- 影响权限、数据权限、缓存一致性、`i18n` 和 E2E 验证：管理接口要求平台上下文和插件权限；档位解析属于运行时关键配置，需要明确缓存权威源、失效和集群同步；用户可见菜单和页面需要按插件 `i18n` 策略治理。
