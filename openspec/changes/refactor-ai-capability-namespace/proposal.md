## Why

当前主框架已通过 `aitext` 提供文本 `AI` 能力，但能力入口直接挂在根 `capability.Services` 上，形态为 `AIText() aitext.Service`。随着后续图片、向量、音频等 `AI` 能力增加，继续在根能力目录追加 `AIImage()`、`AIEmbedding()` 会稀释宿主能力边界，也不利于形成统一的 `AI` 调用入口。

LinaPro 定位是面向可持续交付的 `AI` 原生全栈框架，`AI` 能力应在宿主能力目录中形成清晰命名空间：根目录只暴露 `AI()`，具体文本、图片、向量等能力在 `ai.Service` 下按子能力扩展。

## What Changes

- **BREAKING**：将根能力目录中的 `AIText() aitext.Service` 重构为 `AI() ai.Service`。
- 新增 `apps/lina-core/pkg/plugin/capability/ai/` 能力命名空间，首期只承载文本子能力入口 `Text() aitext.Service`。
- 将现有文本能力包从 `apps/lina-core/pkg/plugin/capability/aitext` 迁移到 `apps/lina-core/pkg/plugin/capability/ai/aitext`，保持 `framework.ai.text.v1`、`Available`、`Status`、`GenerateText` 和 provider factory 语义不变。
- 调整源码插件宿主服务目录、动态插件 guest 目录和 `WASM` host service handler，使源码插件与动态插件都通过 `AI().Text()` 使用文本能力。
- 将文本生成消费请求与 provider 内部请求分离，普通调用方不再填写 `SourcePluginID`；插件来源身份由 plugin-scoped service 或动态插件 host-call 上下文注入。
- 保持动态插件 `hostServices` 协议的 `service: ai` 与 `method: text.generate` 不变，继续使用 `host:ai:text` 细粒度授权。
- 本次不实现 `Image()`、`Embedding()`、图片生成、向量生成、音频处理、流式文本或渠道治理新功能。

## Capabilities

### New Capabilities

- `framework-ai-capability-namespace`：定义宿主 `AI` 能力聚合入口、文本子能力迁移、插件来源身份注入和未来子能力扩展边界。

### Modified Capabilities

- `plugin-host-service-extension`：调整动态插件 guest 侧 `AI` 调用入口，要求 `AI().Text()` 最终仍进入 `ai.text.generate` host service，并保持现有授权、资源和审计语义。

## Impact

- 影响 `apps/lina-core/pkg/plugin/capability/capability.go`、新增 `capability/ai` 聚合包，并迁移 `aitext` 包路径。
- 影响 `apps/lina-core/pkg/plugin/pluginhost`、`apps/lina-core/pkg/plugin/capability/guest`、`apps/lina-core/internal/service/plugin/internal/hostservices`、`apps/lina-core/internal/service/plugin/internal/wasm` 和相关测试桩。
- 影响 `apps/lina-core/pkg/plugin/pluginbridge` 中引用 `aitext` DTO 的编解码文件，但不改变 `service: ai`、`method: text.generate`、`host:ai:text` 或 `purpose:<name>` 授权资源语义。
- 影响 `apps/lina-plugins/linapro-ai-core/` 对 `aitext` 的导入路径和 provider 请求结构；修改前必须检查并遵守该插件根目录 `AGENTS.md`。
- 不新增数据库表、HTTP API、前端页面、菜单、运行时文案或插件资源；`i18n`、数据权限和缓存一致性预期无行为影响，但实现和审查必须记录无影响判断。
- 需要运行宿主能力、动态 host service、源码插件智能中心后端、启动装配和静态导入边界验证，确认生产代码不再依赖旧 `capability/aitext` 路径或 `Services.AIText()` 根入口。
