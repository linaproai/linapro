## Context

`add-ai-text-capability-center` 已经让 `apps/lina-core` 发布 `framework.ai.text.v1` 文本 `AI` 能力，并让官方源码插件 `linapro-ai-core` 提供渠道、模型、档位和调用日志实现。当前问题不在文本能力本身，而在能力入口形态：`capability.Services` 直接暴露 `AIText() aitext.Service`，后续继续新增 `AIImage()`、`AIEmbedding()` 会让根服务目录承载过多同族能力。

现有动态插件 host service 协议已经使用 `service: ai` 和 `method: text.generate`，这说明运行时协议层已经具备 `AI` 服务族概念。Go 侧源码插件和 guest SDK 应与这一结构对齐，通过 `AI().Text()` 表达统一入口和子能力访问。

本变更属于宿主通用能力和插件能力接缝重构，不属于工作台展示适配层能力。实现必须保持 `apps/lina-core` 对渠道、模型、档位和调用日志的宿主边界不变，不能把 `linapro-ai-core` 内部实现、`DAO`、`DO`、`Entity` 或缓存对象暴露给调用方。

## Goals / Non-Goals

**Goals:**

- 将根能力目录从 `AIText() aitext.Service` 调整为 `AI() ai.Service`。
- 新增 `capability/ai` 聚合包，首期只暴露 `Text() aitext.Service`。
- 将文本能力包迁移到 `capability/ai/aitext`，保持 `framework.ai.text.v1` 能力 ID 和文本生成行为不变。
- 让源码插件、宿主模块、动态插件 guest SDK 和 `WASM` host service handler 都通过 `AI().Text()` 消费文本能力。
- 将调用方可见的 `GenerateRequest` 与 provider 内部请求拆分，使 `SourcePluginID` 由 plugin-scoped service 或动态插件上下文注入。
- 用静态检索和编译门禁验证旧 `capability/aitext` 包路径与 `Services.AIText()` 根入口不再出现在生产代码中。

**Non-Goals:**

- 不新增 `Image()`、`Embedding()` 或任何图片、向量、音频能力实现。
- 不改变 `linapro-ai-core` 的渠道、模型、档位、调用日志数据模型、管理 API 或前端页面。
- 不改变动态插件 `plugin.yaml` 中 `service: ai`、`method: text.generate`、`host:ai:text`、`purpose:<name>` 的授权语义。
- 不新增数据库迁移、Seed、Mock 数据、语言包、菜单、路由或用户可见文案。
- 不引入泛化 `Generate(capabilityType, payload)` 网关，也不把不同 `AI` 子能力合并成一个弱类型运行时分发接口。

## Decisions

### 1. 根能力目录只暴露 `AI()`

目标接口：

```go
type Services interface {
    AI() ai.Service
}
```

`AI()` 是 `AI` 能力族的稳定入口。根服务目录不再随着 `AI` 子能力增加而扩展为 `AIText()`、`AIImage()`、`AIEmbedding()`。这个抽象满足当前已确认变化点：文本能力已经存在，图片和向量属于明确后续能力族，但它不引入 provider manager 或弱类型 registry。

备选方案是保留 `AIText()` 并未来继续追加根方法。该方案实现最少，但会把根 `capability.Services` 变成能力族方法堆叠，不符合 `AI` 能力统一入口。

### 2. `ai.Service` 只做聚合，不做业务网关

目标接口：

```go
type Service interface {
    Text() aitext.Service
}
```

`ai.Service` 的职责是命名空间聚合和插件身份绑定，不负责渠道选择、档位解析、授权判断或弱类型方法分发。文本能力仍由 `aitext.Service` 维护自己的 DTO、状态、错误和 provider factory。

拒绝方案：

```go
type Service interface {
    Generate(ctx context.Context, capabilityType string, payload any) (any, error)
}
```

该方案表面统一，但会把文本、图片、向量的授权、审计、成本、用量和响应结构推到运行时判断里，降低契约清晰度。

### 3. 文本包迁移到 `capability/ai/aitext`

文本能力包移动后，`framework.ai.text.v1`、`CapabilityAITextV1`、`Tier`、`ThinkingEffort`、`Message`、`GenerateResponse`、`ProviderFactory` 和 fallback/provider manager 语义保持不变。迁移不新增兼容包，因为项目没有历史兼容负担，保留旧包会让调用方继续分叉。

实现时需要更新 `linapro-ai-core`、`pluginbridge` 编解码、WASM handler、hostservices、guest SDK、测试桩和启动装配中的导入路径。

### 4. 消费请求与 provider 请求分离

普通消费方看到的 `aitext.GenerateRequest` 不再包含 `SourcePluginID`。来源身份由服务层补齐：

```go
type GenerateRequest struct {
    Purpose string
    Tier Tier
    Messages []Message
    MaxOutputTokens int
    Temperature *float64
    ThinkingEffort *ThinkingEffort
    Metadata map[string]string
}

type ProviderRequest struct {
    GenerateRequest
    SourcePluginID string
}
```

源码插件通过 `ServicesForPlugin(services, pluginID).AI().Text()` 获取 plugin-scoped service；动态插件由 host-call 上下文提供 `pluginID`。这样普通调用方不能漏填、错填或伪造 `SourcePluginID`，审计来源也更稳定。

### 5. 动态插件协议保持稳定

动态插件 `plugin.yaml` 和 host service 协议不变：

```yaml
hostServices:
  - service: ai
    methods:
      - text.generate
```

guest SDK 调用形态调整为 `guest.Default().AI().Text().GenerateText(...)`。该调用最终仍通过 `ai.text.generate` 进入宿主，并继续执行 service、method、`purpose` 资源、输出上限和 `host:ai:text` 授权校验。

### 6. 影响分析

| 领域 | 判断 |
|------|------|
| `i18n` | 无运行时用户可见文案、菜单、路由、按钮、API 文档源文本或语言包变更；任务和审查记录必须明确无影响。 |
| 数据权限 | 不新增数据操作接口，不改变管理 API 或日志查询数据可见性；动态插件授权仍沿用现有 host service resource 校验。 |
| 缓存一致性 | 不新增缓存、不改变档位解析缓存权威源或失效策略；迁移后必须复用原共享 service 实例。 |
| 插件边界 | 会修改 `apps/lina-plugins/linapro-ai-core/` 导入路径，修改前必须检查插件根目录 `AGENTS.md`。 |
| 开发工具 | 不新增脚本或跨平台工具入口。 |
| 测试策略 | 使用 Go 编译门禁、相关单元测试和静态检索验证重构无旧入口残留。 |

## Risks / Trade-offs

- [Risk] `AI()` 聚合接口仍会在新增 `Image()` 或 `Embedding()` 时破坏所有实现 `ai.Service` 的类型。→ Mitigation：让 `ai.Service` 的实现面保持极窄，主要由宿主 directory、scoped directory、guest directory 和测试桩实现；根 `capability.Services` 不再受每个 `AI` 子能力扩展影响。
- [Risk] 包路径迁移影响面较大，容易遗漏测试桩或插件导入。→ Mitigation：任务中加入静态检索，确认生产代码不再 import `lina-core/pkg/plugin/capability/aitext`，不再调用 `AIText()`。
- [Risk] provider 请求拆分可能遗漏动态插件来源注入。→ Mitigation：新增或更新 `WASM` host service 测试，断言动态插件调用最终写入或传递正确 `SourcePluginID`。
- [Risk] 过度抽象成 `AI` 网关。→ Mitigation：规格明确禁止弱类型 `Generate(capabilityType, payload)`，`ai.Service` 只做子能力聚合。

## Migration Plan

1. 新增 `capability/ai` 聚合包和 `Service` 接口，迁移 `aitext` 包到 `capability/ai/aitext`。
2. 将 `capability.Services`、`pluginhost.Services`、hostservices directory、guest directory 和测试桩从 `AIText()` 改为 `AI().Text()`。
3. 拆分 `GenerateRequest` 和 provider 请求，调整 `linapro-ai-core` provider adapter、`WASM` host service handler 和相关测试。
4. 更新 `pluginbridge` 编解码和动态 guest SDK 导入路径，保持协议字段和授权语义不变。
5. 运行 Go 编译门禁和静态检索；若失败，按调用方逐项迁移。
6. 回滚方式为恢复旧包路径和 `AIText()` 入口；由于不修改数据库或 HTTP API，回滚不涉及数据迁移。

## Open Questions

- 是否需要将宿主内部普通调用也纳入 scoped identity，还是只有插件调用注入 `SourcePluginID`？默认建议宿主内部调用保留空来源或使用稳定的宿主来源标识，不伪装成插件。
- 是否需要在本变更中补充治理扫描，自动阻断旧 `capability/aitext` 导入？默认先使用任务级静态检索和审查记录，后续如果旧路径反复回流再升级为治理规则。
