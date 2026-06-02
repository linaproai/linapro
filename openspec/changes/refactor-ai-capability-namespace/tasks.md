## 1. 规则、边界和现状确认

- [x] 1.1 实施前重新读取 `AGENTS.md` 和命中的规则文件，至少覆盖 `openspec`、`documentation`、`architecture`、`plugin`、`backend-go`、`testing`、`i18n`、`cache-consistency`、`data-permission`，并记录 `i18n`、缓存一致性、数据权限、开发工具跨平台和测试策略影响分析
  - 实现记录：已重新读取 `AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`i18n.md`、`cache-consistency.md`、`data-permission.md`、`dev-tooling.md`，并使用 `openspec-apply-change`、`goframe-v2`、`karpathy-guidelines` 技能。影响分析：本变更为 Go 后端能力契约和插件 host service 接缝重构，不新增 HTTP API、数据库、前端页面、菜单、路由或运行时文案；`i18n` 无资源影响；缓存一致性不改变档位解析缓存权威源、失效或共享修订号策略；数据权限不新增数据操作接口，动态插件授权仍沿用 `hostServices` service/method/resource 校验；开发工具跨平台无影响，不修改脚本或工具入口；测试策略为 Go 单测、编译门禁、静态导入检索、OpenSpec 严格校验和 `git diff --check`。
- [x] 1.2 修改 `apps/lina-plugins/linapro-ai-core/` 前检查该插件根目录是否存在 `AGENTS.md`，存在时先读取并遵守插件本地规范
  - 实现记录：已执行 `find apps/lina-plugins/linapro-ai-core -maxdepth 1 -name AGENTS.md ...`，未发现插件根目录 `AGENTS.md` 普通文件或符号链接；本次修改该插件时继续遵守项目顶层规范和命中的规则文件。
- [x] 1.3 静态梳理当前 `AIText()`、`capability/aitext`、`SourcePluginID`、`HostServiceMethodAITextGenerate` 和 guest `AIText()` 的调用点，确认迁移范围和不涉及数据库、HTTP API、前端页面或语言包的无影响判断
  - 实现记录：已使用 `rg` 扫描 `apps/lina-core` 与 `apps/lina-plugins`。迁移范围集中在 `pkg/plugin/capability`、`pkg/plugin/capability/guest`、`pkg/plugin/pluginbridge`、`internal/service/plugin/internal/hostservices`、`internal/service/plugin/internal/wasm`、`internal/service/plugin` 启动装配和测试桩，以及 `apps/lina-plugins/linapro-ai-core/backend` 的 `aitext` 导入与 provider 请求来源投影。`SourcePluginID` 在 `i18n`、runtime cache、operlog 等既有上下文中的命中不属于本次迁移。确认本次不涉及数据库、HTTP API、前端页面、菜单或语言包。
- [x] 1.4 评估本变更是否适合 subagent 并行推进；若不采用，记录原因和串行执行边界
  - 实现记录：本变更横跨同一组 Go 包路径迁移、接口签名和测试桩编译收敛，强依赖统一上下文和连续编译反馈；并行拆分容易产生不一致的包路径、接口实现和来源注入语义。本轮不采用 subagent，按本地串行方式完成迁移、测试和审查。

## 2. 宿主 AI 命名空间重构

- [x] 2.1 新增 `apps/lina-core/pkg/plugin/capability/ai/` 聚合包，定义 `ai.Service`、默认实现和 `Text() aitext.Service` 入口
  - 实现记录：已新增 `apps/lina-core/pkg/plugin/capability/ai/ai.go`，`ai.Service` 仅暴露 `Text() aitext.Service`，默认实现只做类型化子能力聚合；未引入弱类型 `Generate`/`Invoke` 网关。已通过 `go test ./pkg/plugin/capability/... -count=1`。
- [x] 2.2 将文本能力包迁移到 `apps/lina-core/pkg/plugin/capability/ai/aitext/`，保持 `framework.ai.text.v1`、DTO、错误码、provider factory、fallback 和状态语义不变
  - 实现记录：已将原 `capability/aitext` 迁移到 `capability/ai/aitext`，保留 `framework.ai.text.v1`、`ProviderPluginID`、档位、状态、错误码、fallback、provider factory 和 provider manager 语义；不保留旧兼容包路径。已通过 `go test ./pkg/plugin/capability/... -count=1`。
- [x] 2.3 将根 `capability.Services`、`pluginhost.Services`、源码插件 hostservices directory 和 scoped directory 从 `AIText()` 改为 `AI().Text()`
  - 实现记录：已将根 `capability.Services` 改为 `AI() ai.Service`，`pluginhost.Services` 通过嵌入自动同步；源码插件 hostservices directory 保存 `ai.Service`，scoped directory 在 `AI()` 上绑定 `pluginID` 并通过 `AI().Text()` 访问文本能力。已通过 `go test ./internal/service/plugin/internal/hostservices ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/testutil ./internal/service/plugin ./internal/cmd -count=1`。
- [x] 2.4 确保 `AI()` 和 `Text()` 在未配置 provider 时返回可降级 fallback service，而不是要求调用方处理 nil
  - 实现记录：`ai.New(nil)`、`ai.ForPlugin(nil, pluginID)` 和 `ai.Service.Text()` 均返回 `aitext.New(nil)` fallback；hostservices base/scoped directory 的 `AI()` 在缺少服务时返回 fallback namespace。新增 `capability/ai` 单测覆盖 fallback 可用性，并通过 `go test ./pkg/plugin/capability/... -count=1`。
- [x] 2.5 拆分文本生成消费请求与 provider 内部请求，使普通 `GenerateRequest` 不再包含 `SourcePluginID`，由 scoped service 注入 provider 请求来源
  - 实现记录：已从调用方可见的 `aitext.GenerateRequest` 移除 `SourcePluginID`，新增 provider 内部 `aitext.ProviderRequest`；`aitext.ForPlugin` 将 scoped `pluginID` 注入 provider 请求，宿主内部普通调用保持空来源。新增 `TestForPluginInjectsSourcePluginID`，并通过 `go test ./pkg/plugin/capability/... -count=1`。

## 3. 动态插件和官方插件同步

- [x] 3.1 将动态插件 guest 目录从 `AIText()` 调整为 `AI().Text()`，并保持底层 `service: ai`、`method: text.generate`、`purpose:<name>` 调用不变
  - 实现记录：已将 guest `Services` 根入口改为 `AI() AIService`，`AIService.Text()` 返回原文本生成 guest client；host service 调用仍构造 `service: ai`、`method: text.generate` 与 `purpose:<name>` resourceRef。已通过 `go test ./pkg/plugin/capability/... -count=1`。
- [x] 3.2 更新 `WASM` host service handler，使动态插件来源身份继续来自 host-call 上下文，并通过 `services.AI().Text()` 获取文本能力
  - 实现记录：`hostfn_service_ai.go` 不再向公共 `GenerateRequest` 写入 `SourcePluginID`，而是通过 `capability.ServicesForPlugin(..., hcc.pluginID).AI().Text()` 获取 scoped 文本能力；动态插件来源由 scoped service 注入。已更新 WASM AI host service 测试并通过 `go test ./internal/service/plugin/internal/wasm -count=1`。
- [x] 3.3 更新 `pluginbridge` 协议编解码和测试中对 `aitext` 的导入路径，确认 host service 授权和 DTO 字段无协议变化
  - 实现记录：已将 `pluginbridge` AI codec 的 `aitext` 导入更新为 `capability/ai/aitext`，未改变 `HostServiceAITextGenerateRequest` 字段、`service: ai`、`method: text.generate`、`host:ai:text` 或 `purpose` 资源校验。已通过 `go test ./pkg/plugin/pluginbridge/internal/hostservice -count=1`。
- [x] 3.4 更新 `apps/lina-plugins/linapro-ai-core/` 中 provider 注册、service、adapter、controller 测试和 helper 的 `aitext` 导入路径与 provider 请求结构
  - 实现记录：已更新 `linapro-ai-core` 导入路径和 `GenerateText(ctx, aitext.ProviderRequest)` provider 签名；调用日志继续从 provider 请求读取 `SourcePluginID`。已通过 `GOWORK=off go test ./backend/internal/service/ai -count=1`。
- [x] 3.5 更新宿主内部测试桩、集成测试和启动装配测试中实现 `capability.Services` 的类型，确保它们实现 `AI() ai.Service`
  - 实现记录：已更新 root plugin tests、testutil、integration fake、WASM fake 和启动依赖测试中的能力目录实现，均通过 `AI() ai.Service` 暴露 fallback 或 fake 文本能力。已通过 `go test ./internal/service/plugin/internal/hostservices ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/testutil ./internal/service/plugin ./internal/cmd -count=1`。

## 4. 验证和审查

- [x] 4.1 补充或更新 Go 单元测试，覆盖 `AI().Text()` 能力获取、fallback 可用性、源码插件 scoped 来源注入、动态插件来源注入和旧 `AIText()` 入口移除
  - 实现记录：新增 `capability/ai` fallback 单测，新增 `aitext.ForPlugin` provider 来源注入单测，更新 WASM AI host service 测试断言动态插件来源由 scoped 目录注入；旧 `AIText()` 入口移除后相关测试桩已改为 `AI() ai.Service`。已通过 `go test ./pkg/plugin/capability/... ./internal/service/plugin/internal/wasm -count=1`。
- [x] 4.2 运行宿主相关 Go 编译门禁，至少覆盖 `go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/internal/hostservice ./internal/service/plugin/internal/wasm ./internal/service/plugin ./internal/cmd -count=1`
  - 验证记录：已在 `apps/lina-core` 运行 `go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/internal/hostservice ./internal/service/plugin/internal/wasm ./internal/service/plugin ./internal/cmd -count=1`，结果通过。DI 来源检查：未新增新的运行期依赖 owner；`ai.Service` 聚合由 source-plugin hostservices directory 使用启动期传入的同一个 `aitext.Service` 构造，`aitext.Service` 仍由 HTTP startup 通过 `aitext.New(pluginSvc)` 创建并共享，plugin-scoped 来源通过 `ServicesForPlugin(...).AI()` 包装注入，不创建独立关键服务图。
- [x] 4.3 在 `apps/lina-plugins/linapro-ai-core/` 运行插件后端 Go 测试，覆盖 provider adapter、档位调用和调用日志来源投影不回归
  - 验证记录：已在 `apps/lina-plugins/linapro-ai-core` 运行 `GOWORK=off go test ./backend/... -count=1`，结果通过；`backend/internal/service/ai` 覆盖 provider adapter、档位调用和调用日志来源投影。
- [x] 4.4 运行静态检索，确认生产代码不再 import `lina-core/pkg/plugin/capability/aitext`，不再调用 `Services.AIText()` 或 guest `AIText()` 根入口
  - 验证记录：已运行 `rg -n "lina-core/pkg/plugin/capability/aitext|AIText\\(\\)|\\.AIText\\(" apps/lina-core apps/lina-plugins -g '*.go'`，无命中；已检查旧 `apps/lina-core/pkg/plugin/capability/aitext` 目录不存在。`SourcePluginID` 仅保留在 `ProviderRequest`、provider 日志投影、其他既有上下文和测试验证中。
- [x] 4.5 记录 E2E 影响判断：本变更无用户可观察 UI 行为变化，预期不新增 E2E；如实现中触及页面或用户文案，则按 `testing` 与 `i18n` 规则补充验证
  - 验证记录：本次只修改 Go 后端能力契约、host service 接缝、guest SDK 和官方插件 provider 签名；未修改前端页面、路由、菜单、表单、按钮、运行时文案或 API 文档源文本，无用户可观察 UI 行为变化，不触发 E2E 新增要求。`i18n` 无资源影响，测试策略以 Go 单测、编译门禁、静态扫描、OpenSpec 校验和 diff 检查覆盖。
- [x] 4.6 运行 `openspec validate refactor-ai-capability-namespace --strict` 和 `git diff --check`
  - 验证记录：已运行 `openspec validate refactor-ai-capability-namespace --strict`，结果有效；插件子仓库 `git -C apps/lina-plugins diff --check` 通过；本次变更范围 `git diff --check -- apps/lina-core/pkg/plugin/capability apps/lina-core/pkg/plugin/pluginbridge apps/lina-core/internal/service/plugin apps/lina-core/internal/cmd apps/lina-core/internal/service/user openspec/changes/refactor-ai-capability-namespace` 通过。最终全局 `git diff --check` 被非本次变更文件 `.agents/skills/lina-community-issue-review/SKILL.md` 中既有 conflict marker 阻断，命中行包括 22、30、37、293、295、297、510、512、513；该文件在本轮开始时已处于 dirty 状态，本次未修改，不作为本变更完成阻断。
- [x] 4.7 完成实现后调用 `lina-review`，审查宿主边界、插件目录归属、AI 聚合抽象复杂度、动态插件授权语义、DI 来源、测试覆盖和无 `i18n`、数据权限、缓存一致性影响判断
  - 审查记录：已使用 `lina-review`，从 `git status --short`、未跟踪文件展开、`apps/lina-plugins` 子仓库状态和 `openspec status --change refactor-ai-capability-namespace --json` 收集范围；已重新读取 `AGENTS.md`、`openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`i18n.md`、`cache-consistency.md`、`data-permission.md`、`dev-tooling.md`。`apps/lina-plugins/linapro-ai-core/` 根目录未发现插件本地 `AGENTS.md`。审查结论：未发现本变更范围内的阻塞问题；`ai.Service` 只做类型化子能力聚合，未引入弱类型网关；动态插件授权协议未改变；DI 来源复用启动期共享 `aitext.Service` 与 plugin service runtime；无新增 HTTP API、数据库、前端、运行时文案、缓存或数据权限边界变更。剩余风险：全局 `git diff --check` 被非本次变更文件 `.agents/skills/lina-community-issue-review/SKILL.md` 的既有 conflict marker 阻断；本次变更范围 diff check、插件子仓库 diff check 和 `openspec validate refactor-ai-capability-namespace --strict` 均通过。
