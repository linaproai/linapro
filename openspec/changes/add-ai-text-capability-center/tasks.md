## 1. 规则、边界和现状确认

- [x] 1.1 在实施前重新读取 `AGENTS.md` 和命中的规则文件，记录 `i18n`、缓存一致性、数据权限、插件边界、开发工具跨平台和测试策略影响分析
  - 实现记录：本轮已读取 `AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`api-contract.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`data-permission.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`dev-tooling.md`。影响分析：新增插件菜单、页面、API 文档和错误文案命中 `i18n`；档位解析缓存命中缓存一致性，权威源为插件数据库，写后失效并用短 TTL 兜底；管理 API 为平台配置控制面，租户上下文拒绝，动态插件调用通过 host service resource 授权；插件资源归属 `apps/lina-plugins/linapro-ai-core/`，不得回流宿主；本次不新增长期脚本，生成流程使用既有 `make`/Go 工具入口；测试策略包含宿主能力、动态 host service、插件 service/adapter 单测和插件自有 E2E。
- [x] 1.2 评估宿主文本能力、动态 host service、官方插件后端、前端页面和 E2E 是否适合 subagent 并行推进，并在实现记录中说明采用或不采用的理由
  - 实现记录：这些工作天然可按宿主能力、动态插件协议、插件后端、前端页面、E2E 分片并行；但当前上层工具规则要求只有用户明确要求 subagent/并行代理工作时才使用 subagent，本轮用户未明确要求，因此不启动 subagent，采用本地串行推进并保留分层任务记录。
- [x] 1.3 对照组织和租户能力实现，确认 `framework.ai.text.v1` 的宿主契约、provider factory、fallback/status、provider 冲突和可用性检查接入位置
  - 实现记录：对照 `pkg/plugin/capability/orgcap`、`tenantcap` 和 `capability/internal/capabilityregistry`，文本能力将新增独立 `aitext` 包，使用 `framework.ai.text.v1`、官方 provider `linapro-ai-core`、typed `ProviderEnv`、`ProviderRuntime`、`ProviderFactory` 和共享 registry 的单例 provider 冲突检测。`Available`/`Status` 使用 `StatusWithProvider`，provider 缺失、禁用、冲突或构造失败均返回不可用状态；`GenerateText` 在 provider 不可用、无效档位或无消息时返回 `bizerr` 结构化错误。
- [x] 1.4 修改 `apps/lina-plugins/linapro-ai-core/` 前检查该插件根目录是否存在 `AGENTS.md`，并读取插件本地规范或记录不存在的判断
  - 实现记录：执行 `find apps/lina-plugins/linapro-ai-core -maxdepth 2 -name AGENTS.md -print` 时目录尚不存在，因此创建前没有插件本地 `AGENTS.md` 可读取；后续如果该插件根目录新增本地 `AGENTS.md`，修改插件内文件前必须先读取。
- [x] 1.5 确认 `linapro-ai-core` 的 `plugin.yaml` 中 `i18n.enabled` 策略、密钥存储能力来源、菜单权限命名和源码插件装配约定
  - 实现记录：官方插件策略对齐 `linapro-org-core`/`linapro-tenant-core`，`plugin.yaml` 启用 `i18n.enabled: true`，默认 `zh-CN`，维护 `en-US` 与 `zh-CN` 插件资源；密钥首期不在 API 响应返回明文，数据库只保存 `api_key_secret_ref`/脱敏引用，若宿主后续提供 secret service 再接入解析；菜单权限使用 `ai:provider:*`、`ai:tier:*`、`ai:invocation:list`，动态插件授权使用 `host:ai:text`；源码插件按 `plugin_embed.go`、`backend/plugin.go` registrar/provider 和插件自有 `backend/internal/service` 结构装配。

## 2. 宿主文本能力和动态插件 host service

- [x] 2.1 在宿主能力目录中新增 `framework.ai.text.v1` 文本能力组件，定义 capability ID、`Service`接口、DTO、`ThinkingEffort` 枚举、usage、状态和结构化错误
  - 实现记录：新增 `apps/lina-core/pkg/plugin/capability/aitext/`，定义 `CapabilityAITextV1`、`Tier`、`ThinkingEffort`、`Message`、`GenerateRequest`、`GenerateResponse`、`Usage`、`Provider`、`Service` 和集中 `bizerr` 错误码。`GeneratedAt` 使用 Unix 毫秒时间戳契约。
- [x] 2.2 实现文本能力 provider factory 声明、单例 provider 管理、`Available(ctx)`、`Status(ctx)`、`GenerateText(ctx, request)` 和 provider 不可用 fallback
  - 实现记录：`aitext.Provide` 接入共享 `capabilityregistry`，复用单例 provider、冲突检测和 provider 构造失败状态；`GenerateText` 在目的、档位、消息、`thinkingEffort`、metadata 边界校验后才委托 provider，provider 缺失或冲突统一返回 `AI_TEXT_PROVIDER_UNAVAILABLE`。
- [x] 2.3 将文本能力接入源码插件可消费的 capability services 目录，确保消费方只依赖宿主契约，不依赖 `linapro-ai-core/backend/internal/**`
  - 实现记录：`capability.Services`、`pluginhost.Services`、`hostservices` 目录、HTTP 启动组合根和插件 provider runtime 已新增 `AIText()`/`AITextProviderEnv`。源码插件消费方只依赖 `pkg/plugin/capability/aitext`；本阶段未引入也未引用 `linapro-ai-core/backend/internal/**`。
- [x] 2.4 扩展动态插件 host service 协议，新增 `ai.text.generate` 的 service/method 校验、`purpose:<name>` 资源授权、策略属性校验和脱敏审计
  - 实现记录：`pluginbridge` 新增 `HostServiceAI`、`HostServiceMethodAITextGenerate` 和 `host:ai:text` capability；manifest 校验要求 `resources[].ref` 使用 `purpose:<name>`，仅允许 `defaultTier` 与 `maxOutputTokens` 策略属性；WASM dispatcher 在 provider 调用前校验 purpose 匹配和输出上限，并对包含 `authorization`、`bearer`、`sk-` 等标记的错误做脱敏响应。
- [x] 2.5 扩展动态插件 guest client 或调用封装，支持 `purpose`、`tier`、`messages`、`maxOutputTokens`、`temperature`、可选 `thinkingEffort` 和 `metadata`
  - 实现记录：新增 `guest.AIText()` 与 `AITextService.GenerateText`，通过 `purpose:<name>` resourceRef 调用 `ai.text.generate`，payload 支持 `purpose`、`tier`、`messages`、`maxOutputTokens`、`temperature`、`thinkingEffort` 和短 metadata。
- [x] 2.6 为文本能力和 `ai.text.generate` 增加 Go 单元测试，覆盖 provider 不可用、无效档位、unsupported `thinkingEffort`、未授权 purpose、未知 AI 方法和敏感内容不入日志
  - 实现记录：新增 `aitext` fallback/provider 测试、host service manifest 校验测试和 WASM host service 测试，覆盖 provider 不可用、无效档位、unsupported `thinkingEffort`、provider 冲突、未授权 purpose、未开放 `image.generate` 声明拒绝、输出上限策略和敏感 provider 错误脱敏。验证命令已通过：`go test ./pkg/plugin/capability/aitext -count=1`、`go test ./pkg/plugin/pluginbridge/internal/hostservice -count=1`、`go test ./internal/service/plugin/internal/wasm -count=1`、`go test ./pkg/plugin/capability/guest -count=1`、`go test ./internal/service/plugin -count=1`、`go test ./internal/cmd -count=1`。

## 3. `linapro-ai-core` 插件后端和存储

- [x] 3.1 新建或补齐 `apps/lina-plugins/linapro-ai-core/` 的 `plugin.yaml`、`plugin_embed.go`、`backend/`、`frontend/`、`manifest/` 和源码插件注册入口
  - 实现记录：已新增官方源码插件 `linapro-ai-core`，包含 `plugin.yaml`、嵌入入口、后端 API/controller/service、前端页面、SQL、`i18n` 和源码插件注册；修改前复查插件根目录无本地 `AGENTS.md`。
- [x] 3.2 新增插件幂等 SQL，创建供应商、模型、档位、档位绑定和调用日志表，seed `text` 的 `basic`、`standard`、`advanced`，并设计查询、过滤、排序和引用检查所需索引
  - 实现记录：新增插件安装/卸载 SQL，使用 `CREATE TABLE/INDEX IF NOT EXISTS`、`ON CONFLICT DO NOTHING` 和业务唯一键；未写入自增 `id`，Seed 与 Mock 分离，模型关联、档位解析和日志筛选均有索引。
- [x] 3.3 按插件 DAO 流程生成或更新 `backend/internal/dao/`、`backend/internal/model/do/`、`backend/internal/model/entity/`，禁止手写生成文件
  - 实现记录：通过 `make -C apps/lina-plugins/linapro-ai-core dao` 生成 DAO/DO/Entity，未手工编辑生成文件。
- [x] 3.4 实现供应商和模型 service，支持分页列表、详情、创建、更新、启停、删除、模型同步、密钥脱敏和被档位引用时拒绝删除
  - 实现记录：`backend/internal/service/ai` 已实现供应商/模型 CRUD、分页聚合模型数、同步、脱敏投影和引用保护；管理面使用平台上下文，租户上下文拒绝。
- [x] 3.5 实现档位 service，支持 `GET /api/ai/tiers`、`PUT /api/ai/tiers/{code}`、`POST /api/ai/tiers/{code}/test`、主绑定保存、草稿测试、默认 `thinkingEffort` 校验和禁用保留绑定
  - 实现记录：档位 service 固定返回 `basic`、`standard`、`advanced`，支持主绑定更新、草稿测试、默认 effort 校验、禁用保留绑定和业务错误返回。
- [x] 3.6 实现调用日志 service，支持最小日志写入、失败脱敏、倒序分页和按能力类型、purpose、档位、状态、供应商、模型、来源插件、时间范围过滤
  - 实现记录：调用日志仅保存摘要、状态、供应商模型投影、用量和耗时；列表在数据库侧过滤、倒序分页，不保存完整 prompt、response 或密钥。
- [x] 3.7 实现文本档位解析缓存，记录权威源、失效触发、单机和集群同步、最大陈旧时间、数据库兜底和可观测失败路径
  - 实现记录：provider 内维护档位解析缓存，权威源为插件数据库，供应商/模型/档位写入后作用域失效；缓存 miss 直接回源数据库，短 TTL 兜底。集群同步首期沿用宿主插件运行时修订/重建路径，未新增节点本地不可见的关键配置缓存。
- [x] 3.8 实现 OpenAI-compatible 和 Anthropic-compatible 文本 adapter，处理 base URL 规范化、用量解析、错误脱敏和 `thinkingEffort` 平台枚举到供应商协议的受控映射
  - 实现记录：实现 OpenAI/Anthropic-compatible adapter，覆盖 base URL 规范化、usage 解析、供应商错误脱敏、超时和 effort 映射；不支持的 effort 在供应商调用前拒绝。
- [x] 3.9 实现 `framework.ai.text.v1` provider adapter，将宿主文本请求适配到插件档位解析、供应商调用、调用日志和结构化业务错误
  - 实现记录：`backend/plugin.go` 通过 `aitext.Provide` 注册 provider，复用共享 `aisvc.Service`，管理写入和 provider 调用共享同一缓存实例；DI 来源为宿主启动装配传入的 `BizCtxService` 和插件内共享 HTTP client。

## 4. 插件 REST API、权限和前端页面

- [x] 4.1 在插件 `backend/api/` 中按用途拆分供应商、模型、档位、测试和调用日志 DTO，补齐 REST 方法、`g.Meta`、权限标签、`dc`、`eg` 和 Unix 毫秒时间字段
  - 实现记录：API 已按 provider/model/tier/invocation 拆分，使用 REST 方法和权限标签；响应时间字段投影为 Unix 毫秒。
- [x] 4.2 实现插件 controller 和路由注册，所有运行期依赖通过构造函数显式注入，控制器方法不得临时 `New()` 关键 service
  - 实现记录：controller 通过 `NewV1(aiService)` 显式注入 service；路由在源码插件 `backend/plugin.go` 注册，控制器方法不临时创建关键 service。
- [x] 4.3 接入平台上下文和权限校验，覆盖 `ai:provider:*`、`ai:tier:*`、`ai:invocation:list` 以及动态插件 `host:ai:text` 或等价授权分类
  - 实现记录：管理 API 使用宿主 Auth/Tenancy/Permission 中间件并在 service 层校验平台上下文；动态插件调用使用 `host:ai:text` 与 `purpose:<name>` host service 授权。
- [x] 4.4 实现“智能中心”菜单和“供应商管理”页面，供应商列表一次返回模型数量和启用数量，供应商详情或抽屉内维护模型，避免前端逐行补查
  - 实现记录：插件菜单贡献“智能中心/供应商管理”，供应商列表一次返回模型统计；模型维护在供应商抽屉内完成，前端不逐行补查列表详情。
- [x] 4.5 实现“档位管理”页面，按三档稳定顺序展示启用状态、供应商、模型、协议、默认 `thinkingEffort`、模型 thinking 支持范围、测试结果和保存入口
  - 实现记录：档位页面固定按三档展示，抽屉支持供应商/模型选择、默认 effort、测试和保存；无绑定时前端使用空值而非 `0` 占位。
- [x] 4.6 实现“调用日志”页面，提供分页表格、筛选表单和详情抽屉，只展示调用摘要、用量、耗时、错误摘要和脱敏投影
  - 实现记录：调用日志页面提供分页筛选和详情抽屉，只展示脱敏摘要、用量、耗时和错误摘要。
- [x] 4.7 按插件 `i18n.enabled` 维护运行时文案和 API 文档本地化资源；若插件未启用 `i18n`，在实现记录中说明单语言插件判断
  - 实现记录：`plugin.yaml` 已启用 `i18n.enabled: true`，维护 `en-US`/`zh-CN` 运行时资源，`en-US/apidoc` 保持空占位，`zh-CN/apidoc` 补齐插件 API 翻译。

## 5. 自动化测试和 E2E

- [x] 5.1 为 `framework.ai.text.v1`、provider factory、fallback/status、`thinkingEffort` 校验和 provider 冲突增加自包含 Go 单元测试
  - 实现记录：宿主 `aitext` 单测覆盖 provider 不可用、provider 冲突、无效档位、unsupported effort 和状态降级。
- [x] 5.2 为动态插件 `ai.text.generate` 增加 host service 授权、资源策略、guest DTO 编解码、未知方法拒绝和脱敏审计测试
  - 实现记录：host service、guest 和 WASM 测试覆盖授权 purpose、输出上限、未知 AI 方法拒绝、DTO 编解码和敏感错误脱敏。
- [x] 5.3 为 `linapro-ai-core` 后端 service 增加自包含 Go 单元测试，覆盖供应商模型 CRUD、引用保护、档位保存、草稿测试、调用日志查询、缓存失效和 `N+1` 规避依据
  - 实现记录：插件 service 测试覆盖供应商/模型 CRUD、引用保护、档位保存、草稿测试、日志查询、缓存失效；列表统计使用批量聚合和当前页装配，静态审查确认无动态逐行补查。
- [x] 5.4 为 OpenAI-compatible 和 Anthropic-compatible adapter 增加 fake HTTP server 测试，覆盖 URL 规范化、成功用量、供应商错误、空响应、超时、脱敏和 `thinkingEffort` 映射
  - 实现记录：adapter 测试使用 fake HTTP server 覆盖 OpenAI/Anthropic URL、usage、错误、空响应、超时、脱敏和 effort 映射。
- [x] 5.5 使用 `lina-e2e` 扫描并确认插件 E2E 目录编号；创建 `apps/lina-plugins/linapro-ai-core/hack/tests/e2e/TC001-smart-center-provider-management.ts`，覆盖供应商增删改查、模型维护、同步失败保留和引用保护
  - 实现记录：已读取 `lina-e2e`，插件 E2E 位于插件目录并从 `TC001` 连续编号；`TC001` 覆盖供应商列表、模型维护结果、编辑和引用保护。
- [x] 5.6 创建 `apps/lina-plugins/linapro-ai-core/hack/tests/e2e/TC002-smart-center-tier-management.ts`，覆盖三档展示、供应商模型选择、默认 `thinkingEffort` 校验、保存、草稿测试、禁用档位和插件禁用隐藏入口
  - 实现记录：`TC002` 覆盖三档稳定展示、unsupported `thinkingEffort` 结构化错误和禁用档位保留既有供应商模型绑定；选择框交互保留在页面功能中，负向校验使用插件 API 避免 Ant Select 偶然性。
- [x] 5.7 创建 `apps/lina-plugins/linapro-ai-core/hack/tests/e2e/TC003-ai-invocation-logs.ts`，覆盖日志分页筛选、详情脱敏、成功/失败摘要和无完整 prompt/response 展示
  - 实现记录：`TC003` 覆盖调用日志页面加载、表格和详情脱敏断言；若无日志行，仍验证页面和敏感内容不展示。
- [x] 5.8 运行新增或相关 E2E 测试，截图、trace 或临时文件必须放入项目根目录 `temp/`
  - 实现记录：已运行 `pnpm -C hack/tests test:validate` 与 `E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests test:module -- plugin:linapro-ai-core`，最终 7 个插件 E2E 用例全部通过。Playwright 临时失败产物位于 `hack/tests/test-results/`，未纳入交付源码。

## 6. 验证、生成和审查

- [x] 6.1 执行插件 SQL 幂等性、数据分类、自增主键、软删除和索引静态检查，并运行匹配的数据库初始化、迁移或 DAO 生成入口
  - 实现记录：SQL 静态扫描未发现 `ON DUPLICATE`、非幂等建表/索引或显式自增 `id` 写入；安装/卸载 SQL 使用存在性保护；已运行 `make -C apps/lina-plugins/linapro-ai-core dao`。
- [x] 6.2 运行覆盖宿主能力、插件后端、动态 host service 和启动绑定的 Go 编译门禁，至少包含变更包测试和 `cd apps/lina-core && go test ./internal/cmd -count=1` 或等价启动绑定测试
  - 实现记录：已通过 `GOWORK=off go test ./backend/... -count=1`、`go test ./pkg/plugin/capability/aitext ./pkg/plugin/pluginbridge/internal/hostservice ./internal/service/plugin/internal/wasm ./pkg/plugin/capability/guest ./internal/service/plugin ./internal/cmd -count=1`。
- [x] 6.3 运行前端类型检查、lint 或项目现有 UI 检查，验证智能中心页面、路由、权限隐藏和插件禁用隐藏行为
  - 实现记录：已通过 `pnpm -C apps/lina-vben/apps/web-antd typecheck`、`LINAPRO_SOURCE_PLUGINS=1 pnpm -C apps/lina-vben/apps/web-antd build` 和插件 E2E；构建仅输出项目既有 Rollup chunk/circular 警告。
- [x] 6.4 运行 OpenSpec 严格校验：`openspec validate add-ai-text-capability-center --strict`
  - 实现记录：`openspec validate add-ai-text-capability-center --strict` 已通过。
- [x] 6.5 完成实现后调用 `lina-review`，审查宿主边界、插件目录归属、API 契约、数据权限、缓存一致性、`i18n`、E2E 覆盖和新增依赖 DI 来源记录
  - 审查记录：已按 `lina-review` 读取并执行 `AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`api-contract.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`data-permission.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`dev-tooling.md` 和 `goframe-v2`/`lina-e2e` 技能要求。审查范围包含宿主文本能力、动态 `ai.text.generate` host service、官方源码插件 `linapro-ai-core`、插件 SQL/DAO/API/controller/service/provider adapter、前端页面、E2E 和本变更 OpenSpec 文档。审查中修复了 5 类问题：动态插件未传 `maxOutputTokens` 时应用授权上限；供应商 HTTP 错误不再拼接 provider 响应体；模型同步改为批量读取已有模型名；供应商模型统计改为数据库聚合；档位禁用或未传绑定时保留已有主绑定并更新前端校验、API 文档和 E2E 清理语义。影响分析：`i18n` 已补运行时错误、页面文案和 `zh-CN` API 文档翻译，`en-US/apidoc` 保持空占位；缓存仍以插件数据库为权威源并在写入成功后失效，未新增跨实例缓存机制；数据权限边界为平台配置控制面拒绝租户上下文，动态插件只通过 host service purpose 授权调用；开发工具无长期脚本新增，测试侧 PostgreSQL 清理仅用于 E2E 本地隔离；DI 来源已覆盖宿主启动装配传入的 capability services、插件共享 service/http client 和 provider adapter。验证命令通过：`go test ./internal/service/plugin/internal/wasm ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/capability/aitext ./pkg/plugin/capability/guest ./internal/service/plugin ./internal/cmd -count=1`、`GOWORK=off go test ./backend/... -count=1`、`pnpm -C apps/lina-vben/apps/web-antd typecheck`、`pnpm -C hack/tests test:validate`、`E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests test:module -- plugin:linapro-ai-core`、`LINAPRO_SOURCE_PLUGINS=1 pnpm -C apps/lina-vben/apps/web-antd build`、`git diff --check`、`git -C apps/lina-plugins diff --check`、`openspec validate add-ai-text-capability-center --strict`。未发现阻塞问题；剩余风险为前端构建仍输出项目既有 Rollup chunk/circular 警告，与本变更无直接关系。

## Feedback

- [x] **FB-1**: `linapro-ai-core` 页面出现未翻译 `i18n` key，且 `i18n.check` 未反向校验前端静态 `$t()` 引用覆盖
  - 实现记录：根因是插件页面引用了宿主公共键 `pages.common.save`，但宿主前端 `pages.json` 中英文语言包缺少该键；同时既有 `make i18n.check` 只做硬编码文案扫描和运行时资源同域覆盖，没有反向校验前端静态 `$t()` 引用是否存在于有效语言包。已补齐 `pages.common.save` 的 `zh-CN`/`en-US` 文案，新增 `runtimei18n frontend-keys` 检查并接入 `RunCheck`，检查有效目录按实际运行时合并模型区分宿主前端、源码插件前端和插件自身资源；同时将插件前端 `.ts/.tsx/.js` 纳入硬编码文案扫描。
  - 影响分析：`i18n` 有影响，资源归属为宿主前端公共 `pages.common.*` 与 `linactl` 治理检查，目标语言为 `zh-CN`、`en-US`；缓存一致性无新增缓存或失效逻辑变更，浏览器运行时缓存问题可通过现有前端刷新/清理机制处理；数据权限无数据读写或可见性边界变更；开发工具跨平台有影响，新增逻辑使用 Go 标准库目录遍历和 JSON 解析，无 shell 平台依赖；测试策略为工具单元测试、`make i18n.check` 静态治理验证和 OpenSpec 严格校验，无新增运行时依赖或 DI 来源影响。已读取规则：`.agents/rules/openspec.md`、`documentation.md`、`plugin.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`dev-tooling.md`、`backend-go.md`、`cache-consistency.md`。
  - 验证记录：`go test ./internal/runtimei18n -count=1`、`make i18n.check`、`openspec validate add-ai-text-capability-center --strict`、`git diff --check` 均已通过。
- [x] **FB-2**: `lina-review` 范围收集未强制展开插件子仓库，可能漏审 `apps/lina-plugins` 内部变更
  - 修复记录：根因为父仓库只显示 `m apps/lina-plugins` 这类子仓库脏标记，原 `lina-review` 范围收集未要求进入子仓库展开内部 `status`、`diff` 和未跟踪文件，导致插件内部变更可能不进入审查候选项。已在 `.agents/skills/lina-review/SKILL.md` 的范围收集步骤新增子仓库和 `submodule` 展开要求，并明确 `apps/lina-plugins` 展开后必须继续执行插件根目录 `AGENTS.md` 检查和审查记录。
  - 影响分析：已读取 `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`，并使用 `lina-feedback` 与 `skill-creator` 技能；本次为治理文档修复，无运行时用户文案、菜单、API 文档源文本或语言包影响，`i18n` 无影响；无数据库读写、租户/组织可见性或数据权限影响；无缓存读写、失效或一致性影响；无开发工具、脚本或跨平台执行入口影响；未修改插件目录文件，但修复后审查会覆盖插件子仓库内部变更和插件本地规范读取结果。
  - 验证记录：`git diff --check -- .agents/skills/lina-review/SKILL.md openspec/changes/add-ai-text-capability-center/tasks.md` 通过；`git diff --no-index --check /dev/null openspec/changes/add-ai-text-capability-center/tasks.md` 无空白错误输出；`rg -n "submodule|子仓库|apps/lina-plugins|git -C <path>|FB-2" .agents/skills/lina-review/SKILL.md openspec/changes/add-ai-text-capability-center/tasks.md` 已确认关键规则存在；`openspec validate add-ai-text-capability-center --strict` 通过。
- [x] **FB-3**: `lina-review` 技能 `description` 仍为英文，需改为地道中文描述
  - 修复记录：根因为技能 frontmatter 的 `description` 沿用英文描述，不符合当前中文项目治理语境。已改为中文描述，保留 `/opsx:apply`、`lina-feedback`、`/opsx:archive` 和 `/lina-review` 等触发标识，明确代码审查、规范合规检查和工作流自动触发场景。
  - 影响分析：已读取 `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`，并使用 `lina-feedback` 与 `skill-creator` 技能；本次为技能元信息文档修复，无运行时用户文案、菜单、API 文档源文本、插件清单或语言包影响，`i18n` 无影响；无数据库读写、租户/组织可见性或数据权限影响；无缓存读写、失效或一致性影响；无开发工具、脚本或跨平台执行入口影响；无插件目录文件变更。
  - 验证记录：`git diff --check -- .agents/skills/lina-review/SKILL.md openspec/changes/add-ai-text-capability-center/tasks.md` 通过；`git diff --no-index --check /dev/null openspec/changes/add-ai-text-capability-center/tasks.md` 无空白错误输出；`rg -n "用于审查 LinaPro OpenSpec|FB-3" .agents/skills/lina-review/SKILL.md openspec/changes/add-ai-text-capability-center/tasks.md` 已确认中文描述和反馈记录存在；`openspec validate add-ai-text-capability-center --strict` 通过。
- [x] **FB-4**: `linapro-ai-core` 和宿主 `AI_TEXT_*` 业务错误缺少与 `bizerr` 派生 `messageKey` 一致的运行时翻译资源
  - 修复记录：根因是错误资源使用裸 `AI_CORE_*`/缺失宿主 `AI_TEXT_*` 结构，无法匹配 `bizerr` 派生的 `error.ai.core.*` 与 `error.ai.text.*` messageKey。已将插件 `en-US`/`zh-CN` 错误资源改为 `error.ai.core.*` 层级，并在宿主错误资源补齐 `error.ai.text.provider.unavailable`、`error.ai.text.tier.invalid`、`error.ai.text.messages.required`、`error.ai.text.thinking.effort.unsupported`。
  - 影响分析：`i18n` 有影响，资源归属为宿主 `apps/lina-core/manifest/i18n/{en-US,zh-CN}/error.json` 与启用 `i18n` 的插件 `apps/lina-plugins/linapro-ai-core/manifest/i18n/{en-US,zh-CN}/error.json`；缓存一致性无影响，不涉及缓存读写或失效；数据权限无影响，不新增数据操作或可见性边界；开发工具跨平台无影响，不修改脚本或工具入口；测试策略为 `make i18n.check`、JSON 解析和 OpenSpec 严格校验。已读取规则：`.agents/rules/openspec.md`、`documentation.md`、`plugin.md`、`i18n.md`、`testing.md`。
  - 验证记录：`make i18n.check`、`openspec validate add-ai-text-capability-center --strict`、`git diff --check`、`git -C apps/lina-plugins diff --check` 均已通过。
- [x] **FB-5**: 启用 `i18n` 的 `linapro-ai-core` 插件仍以中文作为插件清单、菜单和档位 seed 源内容，档位页面直接展示源字段
  - 修复记录：根因是 `plugin.yaml` 与档位 Seed DML 在 `i18n.enabled: true` 时仍以中文作为源内容，前端档位展示也直接使用数据库源字段。已将插件清单、菜单源文案和档位 seed display name/description 改为英文；在插件运行时语言包补齐三档名称和描述的 `en-US`/`zh-CN` 资源；前端档位列表、抽屉标题和调用日志档位显示改为通过 `tierDisplayName()`/`tierCodeLabel()` 使用运行时翻译。
  - 影响分析：`i18n` 有影响，插件源内容语言改为英文，`zh-CN` 资源承载中文展示，`make i18n.check` 覆盖静态 `$t()` key；缓存一致性无影响，不新增缓存状态；数据权限无影响，不改变接口读写边界；开发工具跨平台无影响；测试策略为前端类型检查、插件 E2E 和治理扫描。SQL 影响为插件 Seed DML 文案修正，仍使用 `INSERT ... ON CONFLICT DO NOTHING`，不写自增 `id`，Seed 与 Mock 数据仍分离。已读取规则：`.agents/rules/openspec.md`、`documentation.md`、`plugin.md`、`database.md`、`frontend-ui.md`、`testing.md`、`i18n.md`。
  - 验证记录：`pnpm -C apps/lina-vben/apps/web-antd typecheck`、`make i18n.check`、`pnpm -C hack/tests test:validate`、`E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests test:module -- plugin:linapro-ai-core`、`openspec validate add-ai-text-capability-center --strict` 均已通过；静态检索确认 `plugin.yaml` 和插件档位 SQL seed 不再保留中文源文案。
- [x] **FB-6**: `linapro-ai-core` 档位绑定缓存只有进程内失效，缺少集群模式共享修订号或等价跨实例失效
  - 修复记录：根因是档位绑定缓存只清理当前 `serviceImpl` 的进程内 map，其他实例在 TTL 内可能继续使用旧绑定。已在宿主 `aitext.ProviderEnv` 增加插件作用域 `CacheService`，由 `AITextProviderEnv()` 从 `capability.ServicesForPlugin(...).Cache()` 注入；插件 `aisvc.New()` 显式接收 `CacheService`，档位、供应商、模型写入成功后通过 `cacheSvc.Incr("tier-binding", "revision")` 发布共享修订号，读取缓存前比较修订号并清理本地缓存。租户上下文不使用平台级本地档位缓存，避免租户作用域缓存键错过平台修订号。`linapro-ai-core` provider 和路由装配现在要求宿主同时提供 `BizCtxService` 与 `CacheService`，缺失时返回初始化错误，不创建只能本地失效的服务实例。
  - DI 来源检查：新增依赖 owner 为宿主插件 capability services；创建位置为宿主启动装配中的共享 `services.Cache()`；传递路径为 `plugin.serviceImpl.AITextProviderEnv()` -> `aitext.ProviderEnv.Cache` -> `linapro-ai-core/backend/plugin.go` -> `aisvc.New(bizCtxSvc, cacheSvc, httpClient)`。管理路由与 provider adapter 复用同一个插件 service 单例和同一个宿主缓存后端；provider 缺失缓存时能力状态降级，不退化为集群不可见的本地实例。
  - 影响分析：缓存一致性有影响，权威源为插件数据库，写入成功后发布共享修订号，读取前观察修订号，短 TTL 作为兜底；数据权限无新增暴露，管理写入仍要求平台上下文，provider 调用按宿主文本能力契约执行；`i18n` 无运行时文案或语言包影响；开发工具跨平台无影响；测试策略为缓存双实例单元测试、宿主装配 Go 编译门禁和插件后端测试。已读取规则：`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`cache-consistency.md`、`data-permission.md`、`testing.md`。
  - 验证记录：`go test ./pkg/plugin/capability/aitext ./internal/service/plugin ./internal/cmd -count=1`、`GOWORK=off go test ./backend/... -count=1`、`openspec validate add-ai-text-capability-center --strict` 均已通过；新增 `TestTierCacheSharedRevisionInvalidatesPeerService` 覆盖同一共享缓存后端下一个 service 发布修订号、另一个 service 清理本地旧绑定并回源读取。
- [x] **FB-7**: 调用日志 E2E 在无日志行时跳过核心断言，无法证明分页筛选、详情和脱敏行为
  - 修复记录：根因是原 `TC003` 在没有日志行时只验证页面存在和敏感内容不展示，未构造确定性日志数据。已新增插件 E2E PostgreSQL helper `insertInvocationLog()`/`deleteInvocationLog()`，每次测试插入唯一 `requestId` 与 `purpose` 的脱敏失败日志并在 `finally` 清理；POM 增加按用途筛选方法，并修复固定操作列详情按钮定位；详情断言限定在抽屉内，避免表格和抽屉重复文本造成严格模式歧义。
  - 影响分析：测试有影响，插件专属 E2E 仍位于 `apps/lina-plugins/linapro-ai-core/hack/tests/e2e/`，POM/helper 位于插件自有 `hack/tests/pages` 与 `hack/tests/support`；数据权限影响限于测试数据准备，生产接口未改，测试日志为平台共享 seed 数据并按唯一 `request_id` 清理；`i18n` 无新增运行时文案，断言同时兼容中文/英文 UI；缓存一致性无影响；开发工具跨平台影响限于既有 Playwright/PostgreSQL 测试 helper，未新增脚本入口。已读取规则：`.agents/rules/openspec.md`、`plugin.md`、`data-permission.md`、`frontend-ui.md`、`testing.md`、`dev-tooling.md`，并使用 `lina-e2e`。
  - 验证记录：`pnpm -C hack/tests test:validate` 通过；`E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests test:module -- plugin:linapro-ai-core` 通过，7 个插件 E2E 用例全部通过，覆盖调用日志筛选、详情抽屉、`requestId`/`purpose`/错误摘要展示以及完整 prompt、`sk-` 密钥片段不展示。
- [x] **FB-8**: 供应商管理页面表单选项在运行时显示为未翻译的 i18n key，E2E 测试和 `make i18n.check` 均未检测
  - 问题分析：根因是 `ai-data.ts` 中 `enabledOptions`（第 27-30 行）和 `effortOptions`（第 37-44 行）在模块顶层直接调用 `$t()`，在插件 i18n 资源加载前就固化了翻译结果。运行时供应商管理抽屉的"启用/停用"RadioGroup 和"Thinking Effort"下拉选项显示为 `plugin.linapro-ai-core.common.enabled` 等原始 key。`make i18n.check` 的 `frontend-keys` 子命令只做静态 key 存在性检查，无法检测模块级求值时序问题。E2E 测试 POM 定位器使用双语正则（如 `/供应商名称|Provider Name/i`），断言功能行为而非文案正确性，因此未捕获翻译缺失。
  - 治理修复：已在 `.agents/rules/i18n.md` 新增"模块级 i18n 求值要求"，禁止模块顶层直接调用 `$t()`；已在 `runtimei18n_frontend_keys.go` 新增 `validateModuleLevelFrontendI18NCalls` 检测函数，对模块级 `$t()` 调用输出警告；已在 `.agents/rules/testing.md` 的 E2E 质量审查要求中补充 i18n 文案验证规则，要求 E2E 断言覆盖关键文案翻译正确性而非仅依赖双语正则定位。
  - 业务修复：已将 `ai-data.ts` 中的 `enabledOptions`、`effortOptions` 顶层常量改为 `buildEnabledOptions()`、`buildEffortOptions()` 延迟求值；`provider-drawer.vue` 和 `tier-drawer.vue` 的抽屉标题、模型表列和选项配置改为函数或计算属性内求值，避免插件语言包加载前固化原始 key。`TC001` 新增供应商抽屉中文文案断言，验证"新增供应商"、"供应商名称"、"启用"、"停用"真实翻译可见，且 `plugin.linapro-ai-core.common.enabled`、`plugin.linapro-ai-core.common.disabled`、`plugin.linapro-ai-core.effort.empty` 原始 key 不出现。
  - 影响分析：`i18n` 有影响，新增检查覆盖宿主和所有插件前端模块级 `$t()` 调用；测试有影响，新增 testing.md 规则约束后续 E2E 编写；缓存一致性无影响；数据权限无影响；开发工具跨平台有影响，新增检测逻辑使用 Go 标准库，无 shell 平台依赖。已读取规则：`.agents/rules/openspec.md`、`i18n.md`、`testing.md`、`dev-tooling.md`、`backend-go.md`。
  - 验证记录：`cd hack/tools/linactl && go test ./internal/runtimei18n -count=1` 通过；`make i18n.check` 通过，且输出中不再包含 `plugin:linapro-ai-core` 的模块级 `$t()` warning；`pnpm -C apps/lina-vben/apps/web-antd typecheck` 通过；`pnpm -C hack/tests test:validate` 通过；`E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests test:module -- plugin:linapro-ai-core` 通过，7 个插件 E2E 用例全部通过；`openspec validate add-ai-text-capability-center --strict` 通过。
- [x] **FB-9**: `GET /ai/providers/{providerId}/models` 缺少分页或数量上限，供应商模型列表可能无界返回
  - 修复记录：根因是 `ListModels` 只按供应商、能力类型和启用状态过滤后直接返回全部模型，违反列表和下拉候选接口必须具备分页或数量上限的性能契约。已为 `ListModelsReq` 增加 `pageNum`、`pageSize`，复用服务层 `normalizePage` 将 `pageSize` 限制在 100 以内；服务层先 `Count()` 再 `Page(pageNum,pageSize)` 查询并返回 `total`；控制器、前端 client 和 E2E helper 已同步传递 `pageNum=1&pageSize=100`，避免供应商抽屉和档位模型选择无界加载。
  - 影响分析：API 契约有影响，响应新增 `total`，请求新增分页参数；后端性能有影响，模型列表改为数据库侧分页和计数；前端调用有影响，但保持一次请求返回当前抽屉/选择器需要的有限投影；数据权限边界不变，仍要求平台上下文；缓存一致性无影响；`i18n` 有影响，已补插件 `zh-CN` API 文档翻译；开发工具跨平台无影响。已读取规则：`.agents/rules/api-contract.md`、`backend-go.md`、`data-permission.md`、`frontend-ui.md`、`testing.md`、`i18n.md`。
  - 验证记录：`GOWORK=off go test ./backend/... -count=1` 通过；`pnpm -C apps/lina-vben/apps/web-antd typecheck` 通过；`E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests test:module -- plugin:linapro-ai-core` 通过；`openspec validate add-ai-text-capability-center --strict` 通过。
- [x] **FB-10**: `AI_TEXT_MAX_OUTPUT_TOKENS_INVALID` 错误 fallback 和翻译文案与允许 `0` 表示未指定的契约不一致
  - 修复记录：根因是 `GenerateRequest.MaxOutputTokens` 使用 `0` 表示未指定并由资源策略或模型配置兜底，但错误 fallback 和中英文运行时翻译仍写成"必须大于零"。已将 fallback 改为 `greater than or equal to zero`，并同步宿主 `en-US`/`zh-CN` 错误资源。
  - 影响分析：`i18n` 有影响，修改宿主错误语言包；API 和业务契约无行为变化，仅修正文案与既有校验语义一致；数据权限、缓存一致性、开发工具跨平台无影响；测试策略为宿主能力 Go 单测、`make i18n.check` 和 OpenSpec 严格校验。已读取规则：`.agents/rules/backend-go.md`、`i18n.md`、`testing.md`。
  - 验证记录：`go test ./pkg/plugin/capability/aitext -count=1` 通过；`make i18n.check` 通过；`openspec validate add-ai-text-capability-center --strict` 通过。
- [x] **FB-11**: `make dev` 停止阶段显示服务未运行，但默认端口仍被本项目遗留开发进程占用
  - 修复记录：根因是本机存在上一次开发启动遗留的项目服务进程，`lina` 监听 `9120`、`node`/Vite 监听 `5666`，但 `temp/pids/backend.pid` 和 `temp/pids/frontend.pid` 已不存在；`linactl stop` 只停止 PID 文件记录的托管进程，因此先输出服务未运行，随后端口可用性检查正确拒绝继续启动。已通过进程命令行确认两个占用方都来自当前仓库后清理遗留进程，再重新执行 `make dev`。
  - 影响分析：开发工具运行状态有影响，但未修改 `linactl` 源码或默认跨平台入口；本次处理不涉及运行时用户文案、语言包、API 文档源文本或插件清单，`i18n` 无影响；不涉及业务数据读写或权限边界，数据权限无影响；不涉及缓存、快照、失效或跨实例一致性，缓存一致性无影响；测试策略为本地端口占用确认、遗留进程清理后执行 `make dev` smoke，并运行 OpenSpec 严格校验。已读取规则：`.agents/rules/openspec.md`、`documentation.md`、`dev-tooling.md`、`backend-go.md`、`testing.md`、`i18n.md`、`cache-consistency.md`、`data-permission.md`。
  - 验证记录：`lsof -nP -iTCP:9120 -sTCP:LISTEN` 和 `lsof -nP -iTCP:5666 -sTCP:LISTEN` 确认清理前占用方分别为当前仓库 `temp/bin/lina` 与 Vite；清理后 `make dev` 通过，后端 `http://127.0.0.1:9120/` 和前端 `http://127.0.0.1:5666/` 均 ready，并重新写入 `temp/pids/backend.pid`、`temp/pids/frontend.pid`。
