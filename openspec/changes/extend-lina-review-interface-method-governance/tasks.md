## 1. 规范与审查规则

- [x] 1.1 更新 `AGENTS.md` 服务层接口规范，要求后端接口方法定义唯一且语义清晰。
- [x] 1.2 更新 `lina-review` 技能，新增后端接口方法定义审查规则。
- [x] 1.3 增加 `backend-conformance` 增量规范，覆盖重复方法、近义方法、歧义方法和兼容期说明。

## 2. 验证与审查

- [x] 2.1 运行 `openspec validate extend-lina-review-interface-method-governance --strict`。
- [x] 2.2 运行静态检索，确认 `AGENTS.md` 与 `lina-review` 均包含接口方法定义治理规则。
- [x] 2.3 记录 i18n、缓存一致性、数据权限、开发工具跨平台和测试影响评估，并执行 `lina-review` 审查。

## 3. 执行记录

- i18n 影响评估：本变更只调整治理规范和审查技能说明，不新增、修改或删除运行时用户可见文案、菜单、路由、API 文档源文本、插件清单或语言包资源。
- 缓存一致性影响评估：本变更不新增或修改生产缓存、缓存键、失效路径、跨实例协调或运行时状态。
- 数据权限影响评估：本变更不新增或修改数据操作接口、服务数据访问路径、插件 host service 数据访问或权限边界。
- 开发工具跨平台影响评估：本变更不新增或修改开发工具、脚本、CI 命令实现或默认测试入口。
- 测试策略：项目治理类反馈不新增单元测试或 E2E；使用 OpenSpec 严格校验、静态检索和审查结论作为验证证据。
- Review：已按 `lina-review` 口径完成审查。审查范围限定为 `AGENTS.md`、`.agents/skills/lina-review/SKILL.md` 和 `openspec/changes/extend-lina-review-interface-method-governance`；未发现阻塞问题。该变更不涉及 Go 生产代码、REST API、后端数据权限、运行时缓存、前端 UI 或 i18n 资源，Go 编译门禁和 E2E 不适用。

## 4. Feedback

- [x] **FB-1**: 补充暴露给源码插件的 `IsEnabledAuthoritative` 接口方法详细注释，并增加中文用途说明。
- [x] **FB-2**: 移除插件状态契约中与 `PluginStateService` 重复的 `EnablementReader` 接口定义。
- [x] **FB-3**: 将 `pluginstate` adapter 启用状态方法注释调整为英文在上、中文在下的双语格式。
- [x] **FB-4**: 将 `AGENTS.md` 中分散的多语言 `i18n` 治理要求集中整理到 `.agents/rules/i18n.md`。
- [x] **FB-5**: 收敛 `AGENTS.md` 中的 `i18n` 多语言治理长段，只保留关键要求并引用 `.agents/rules/i18n.md`。
- [x] **FB-6**: 将 `AGENTS.md` 中的 `i18n` 治理内容进一步收敛为规则文件入口，并移除接口文档标签规范中的重复治理条目。
- [x] **FB-7**: 收敛 `lina-review` 技能中的 `i18n` 多语言治理长段，改为读取并遵循 `.agents/rules/i18n.md`。
- [x] **FB-8**: 移除 `AGENTS.md` 中除规则入口外散落的 `i18n` 多语言治理措辞，避免形成隐性第二规则源。
- [x] **FB-9**: 补强 `i18n` 外部规则入口、反馈与审查读取要求和治理验证要求，降低规则迁移后的漏触发风险。
- [x] **FB-10**: 将 `AGENTS.md` 改造为保留规范标题与强制规则加载入口，并迁移领域细则到 `.agents/rules/*.md`。
- [x] **FB-11**: 补充动态插件目录结构规范，并将插件规则拆分为通用资源、源码插件和动态插件三类约束。
- [x] **FB-12**: 调整动态插件目录规范，要求动态插件与源码插件保持 `api`、`controller`、`service` 同构开发结构，仅运行时对接方式不同。
- [x] **FB-13**: 收敛 `lina-review` 技能中与项目规范和规则文件重复的领域细则，改为明确引用 `AGENTS.md` 与 `.agents/rules/*.md`。

## 5. Feedback Execution Record

- FB-1 影响分析：修改源码插件可见的插件状态契约、host service reader 边界、pluginstate adapter 注释和核心插件服务接口注释；不修改方法签名、运行时逻辑、REST API、数据库访问或用户可见前端行为。
- FB-1 i18n 影响评估：本次仅新增 Go 文档注释中的接口用途说明，不新增、修改或删除运行时语言包、插件 manifest i18n、apidoc i18n 或前端 UI 文案。
- FB-1 缓存一致性影响评估：本次不修改缓存键、快照刷新、失效路径、跨实例同步或运行时启用状态读取逻辑；注释明确该方法绕过进程内快照读取持久化治理状态。
- FB-1 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、写操作、下载、聚合统计或插件数据访问路径。
- FB-1 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现或默认测试入口。
- FB-1 验证：已运行 `cd apps/lina-core && go test ./pkg/pluginservice/contract ./pkg/pluginservice/pluginstate ./internal/service/pluginhostservices ./internal/service/plugin -count=1`，通过；已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行静态检索确认 `IsEnabledAuthoritative` 公开契约包含详细说明和中文用途说明。
- FB-1 Review：已按 `lina-review` 口径完成审查。审查范围限定为本次修改的插件状态接口注释、adapter 注释、host service reader 注释、核心插件服务接口注释和本变更任务记录；未发现阻塞问题。该反馈不修改运行时行为、REST API、数据权限、缓存实现、前端 UI 或 i18n 资源；Go 编译门禁、OpenSpec 校验和静态检索均已通过。
- FB-2 影响分析：移除 `pkg/pluginservice/contract` 中与 `PluginStateService` 方法集合重复的 `EnablementReader`，并将 `pkg/pluginservice/pluginstate` adapter 与 `internal/service/pluginhostservices.New` 的插件状态入参统一为 `contract.PluginStateService`；不改变运行时启用状态判断逻辑、REST API、数据库访问或前端行为。
- FB-2 i18n 影响评估：本次仅收敛 Go 接口定义和入参类型，不新增、修改或删除运行时语言包、插件 manifest i18n、apidoc i18n 或前端 UI 文案。
- FB-2 缓存一致性影响评估：本次不修改插件启用状态缓存、快照刷新、权威读取、失效路径、跨实例同步或缓存键；仅移除重复接口定义，保留原有 adapter 的空值保护和 pluginID 归一化行为。
- FB-2 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计或插件数据访问路径。
- FB-2 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现或默认测试入口。
- FB-2 验证：已运行 `cd apps/lina-core && go test ./pkg/pluginservice/contract ./pkg/pluginservice/pluginstate ./internal/service/pluginhostservices ./internal/cmd -count=1`，通过；已运行 `cd apps/lina-core && go test ./internal/service/plugin -run '^$' -count=1`，通过编译烟测；已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行静态检索确认 `apps/lina-core` 目标范围内无 `contract.EnablementReader` 或 `PluginStateReader` 残留。完整 `go test ./internal/service/plugin -count=1` 因当前测试数据库缺少 `plugin_linapro_tenant_core_user_membership` 表失败，失败点与本次接口收敛无关，已用该包 `-run '^$'` 编译烟测补充覆盖。
- FB-2 Review：已按 `lina-review` 口径完成审查。审查范围限定为插件状态公共契约、pluginstate adapter、host service directory 构造入参和本变更任务记录；未发现阻塞问题。该反馈不修改运行时行为、REST API、数据权限、缓存实现、前端 UI 或 i18n 资源；Go 变更已通过相关包测试、启动绑定包测试和插件服务包编译烟测。
- FB-3 影响分析：仅调整 `apps/lina-core/pkg/pluginservice/pluginstate/pluginstate_enablement.go` 中 `IsEnabled` 与 `IsEnabledAuthoritative` 的方法注释顺序和英文说明，不修改方法签名、运行时逻辑、REST API、数据库访问或前端行为。
- FB-3 i18n 影响评估：本次只修改 Go 文档注释，不新增、修改或删除运行时语言包、插件 manifest i18n、apidoc i18n 或前端 UI 文案。
- FB-3 缓存一致性影响评估：本次不修改插件启用状态缓存、权威读取、快照刷新、失效路径、跨实例同步或缓存键；注释继续区分普通 snapshot 查询与权威状态查询语义。
- FB-3 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计或插件数据访问路径。
- FB-3 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现或默认测试入口。
- FB-3 验证：已运行 `cd apps/lina-core && go test ./pkg/pluginservice/pluginstate -count=1`，通过；已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- apps/lina-core/pkg/pluginservice/pluginstate/pluginstate_enablement.go openspec/changes/extend-lina-review-interface-method-governance/tasks.md`，通过。
- FB-3 Review：已按 `lina-review` 口径完成审查。审查范围限定为 `pluginstate_enablement.go` 方法注释和本反馈任务记录；未发现阻塞问题。该反馈不修改运行时行为、REST API、数据权限、缓存实现、前端 UI、i18n 资源或开发工具脚本；Go 编译门禁、OpenSpec 校验和 diff 空白检查均已通过。
- FB-4 影响分析：新增 `.agents/rules/i18n.md`，将 `AGENTS.md` 中分散在开发流程、API 响应时间字段契约、后端错误封装、API 文档标签、服务接口注释、缓存一致性和文档语言治理中的 `i18n` 要求按主题集中整理；不修改生产代码、运行时配置、语言包资源、API DTO、插件清单或前端行为。
- FB-4 i18n 影响评估：本次只复制并重排治理规则文档，不新增、修改或删除运行时前端语言包、宿主 `manifest/i18n`、插件 `manifest/i18n`、`apidoc i18n JSON` 或用户可见产品文案。
- FB-4 缓存一致性影响评估：本次不修改生产缓存、缓存键、翻译包失效路径、跨实例同步、分布式协调或运行时状态；仅在规则文档中集中记录既有 `i18n` 缓存一致性要求。
- FB-4 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界。
- FB-4 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤。
- FB-4 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- .agents/rules/i18n.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md`，通过；已运行静态检索确认 `.agents/rules/i18n.md` 包含治理范围、宿主与插件边界、语言注册、API 文档本地化、API 文档资源隔离、API 文档翻译完整性、运行时 UI、错误本地化、API 时间、运行时依赖、缓存一致性、文档语言治理和审查验证等分类。
- FB-4 Review：已按 `lina-review` 口径完成审查。审查范围限定为 `.agents/rules/i18n.md` 和本反馈任务记录；未发现阻塞问题。该反馈不修改生产 Go 代码、REST API、数据权限、运行时缓存、前端 UI、语言包资源或开发工具脚本；Go 编译门禁和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-5 影响分析：收敛 `AGENTS.md` 中开发流程部分的 `i18n` 持续治理长段，以及接口文档标签规范中的 `apidoc` 详细治理条目；保留关键门禁要求和 `.agents/rules/i18n.md` 引用，不修改生产代码、运行时配置、语言包资源、API DTO、插件清单或前端行为。
- FB-5 i18n 影响评估：本次只调整治理规则索引文档，不新增、修改或删除运行时前端语言包、宿主 `manifest/i18n`、插件 `manifest/i18n`、`apidoc i18n JSON` 或用户可见产品文案；完整 `i18n` 规则仍由 `.agents/rules/i18n.md` 承载。
- FB-5 缓存一致性影响评估：本次不修改生产缓存、缓存键、翻译包失效路径、跨实例同步、分布式协调或运行时状态；`AGENTS.md` 仍保留翻译包缓存显式 `scope` 的关键要求，并引用集中规则文档。
- FB-5 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界。
- FB-5 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤。
- FB-5 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules/i18n.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md`，通过；已运行静态检索确认 `AGENTS.md` 中旧的 `i18n持续治理要求`、`API 文档本地化资源隔离`、`API 文档翻译完整性校验` 和 `OpenAPI/API 文档源文本遵循 i18n 配置` 长段标题已移除，同时保留 `i18n 持续治理关键要求`、`OpenAPI/API 文档源文本与本地化资源遵循 i18n 配置` 和 `.agents/rules/i18n.md` 引用。
- FB-5 Review：已按 `lina-review` 口径完成审查。审查范围限定为 `AGENTS.md`、`.agents/rules/i18n.md` 和本反馈任务记录；未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据权限、运行时缓存、前端 UI、语言包资源或开发工具脚本；Go 编译门禁和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-6 影响分析：进一步将 `AGENTS.md` 中的 `i18n` 治理说明收敛为入口条款，并删除接口文档标签规范中的 `apidoc` 翻译治理重复条目；不修改生产代码、运行时配置、语言包资源、API DTO、插件清单或前端行为。
- FB-6 i18n 影响评估：本次只调整治理规则入口文档，不新增、修改或删除运行时前端语言包、宿主 `manifest/i18n`、插件 `manifest/i18n`、`apidoc i18n JSON` 或用户可见产品文案；完整 `i18n` 治理要求统一由 `.agents/rules/i18n.md` 承载。
- FB-6 缓存一致性影响评估：本次不修改生产缓存、缓存键、翻译包失效路径、跨实例同步、分布式协调或运行时状态；`i18n` 缓存治理细则继续由 `.agents/rules/i18n.md` 承载。
- FB-6 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界。
- FB-6 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤。
- FB-6 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules/i18n.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md`，通过；已运行静态检索确认 `AGENTS.md` 中旧的 `i18n持续治理要求`、`API 文档本地化资源隔离`、`API 文档翻译完整性校验` 和 `OpenAPI/API 文档源文本` 重复治理条目已移除，仅保留 `i18n 治理规则入口` 指向 `.agents/rules/i18n.md`。
- FB-6 Review：已按 `lina-review` 口径完成审查。审查范围限定为 `AGENTS.md`、`.agents/rules/i18n.md` 和本反馈任务记录；未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据权限、运行时缓存、前端 UI、语言包资源或开发工具脚本；Go 编译门禁和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-7 影响分析：收敛 `.agents/skills/lina-review/SKILL.md` 中重复维护的 API 文档国际化、字典语言包、i18n 影响面、硬编码双语映射、源文本命名空间、基础层边界、接口依赖、缓存卫生、方向和语言注册等治理长段，改为审查时读取 `.agents/rules/i18n.md`；不修改生产代码、运行时配置、语言包资源、API DTO、插件清单或前端行为。
- FB-7 i18n 影响评估：本次只调整治理技能文档，不新增、修改或删除运行时前端语言包、宿主 `manifest/i18n`、插件 `manifest/i18n`、`apidoc i18n JSON` 或用户可见产品文案；完整 `i18n` 治理细则统一由 `.agents/rules/i18n.md` 承载。
- FB-7 缓存一致性影响评估：本次不修改生产缓存、缓存键、翻译包失效路径、跨实例同步、分布式协调或运行时状态；`lina-review` 仅保留按 `.agents/rules/i18n.md` 审查相关缓存治理的流程入口。
- FB-7 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界。
- FB-7 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤。
- FB-7 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules/i18n.md .agents/skills/lina-review/SKILL.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md`，通过；已运行静态检索确认 `lina-review` 中不再保留 `i18n` 治理长段，只保留读取 `.agents/rules/i18n.md` 的入口、触发条件和报告模板引用。
- FB-7 Review：已按 `lina-review` 口径完成审查。审查范围限定为 `.agents/skills/lina-review/SKILL.md`、`AGENTS.md`、`.agents/rules/i18n.md` 和本反馈任务记录；未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据权限、运行时缓存、前端 UI、语言包资源或开发工具脚本；Go 编译门禁和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-8 影响分析：将 `AGENTS.md` 中除 `i18n 治理规则入口` 外散落的多语言治理相关措辞改为通用表述，包括缓存关键运行时数据列表、API 时间字段背景、关键服务列表、接口错误封装要求和接口注释边界；不修改生产代码、运行时配置、语言包资源、API DTO、插件清单或前端行为。
- FB-8 i18n 影响评估：本次只调整治理索引文档，不新增、修改或删除运行时前端语言包、宿主 `manifest/i18n`、插件 `manifest/i18n`、`apidoc i18n JSON` 或用户可见产品文案；`AGENTS.md` 仅保留单一规则入口，完整多语言治理要求仍由 `.agents/rules/i18n.md` 承载。
- FB-8 缓存一致性影响评估：本次不修改生产缓存、缓存键、失效路径、跨实例同步、分布式协调或运行时状态；缓存治理总则仍保留在 `AGENTS.md`，多语言资源相关缓存细则由 `.agents/rules/i18n.md` 承载。
- FB-8 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界。
- FB-8 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤。
- FB-8 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules/i18n.md .agents/skills/lina-review/SKILL.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md`，通过；已运行 `grep -nE "i18n|国际化|本地化|多语言|语言包|翻译|apidoc" AGENTS.md`，确认 `AGENTS.md` 只剩 `i18n 治理规则入口` 一处命中。
- FB-8 Review：已按 `lina-review` 口径完成审查。审查范围限定为 `AGENTS.md`、`.agents/rules/i18n.md`、`.agents/skills/lina-review/SKILL.md` 和本反馈任务记录；未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据权限、运行时缓存、前端 UI、语言包资源或开发工具脚本；Go 编译门禁和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-9 影响分析：补强 `AGENTS.md` 外部规则文件入口、`.agents/rules/i18n.md` 影响判断和验证要求、`lina-feedback` 反馈影响分析要求、`lina-review` 基础/深度 i18n 审查触发条件，并新增 `spec-governance` 增量规范记录外部规则文件加载治理；按用户要求未修改由 OpenSpec 工具维护的 `.agents/skills/openspec-*` 与 `.agents/prompts/opsx/*` 文件。
- FB-9 i18n 影响评估：本次仅调整项目治理规范和审查/反馈技能说明，不新增、修改或删除运行时前端语言包、宿主 `manifest/i18n`、插件 `manifest/i18n`、`apidoc i18n JSON` 或用户可见产品文案；规则文件新增宿主 API 文档英文源文本要求、运行时 i18n 检查命令和 apidoc 翻译完整性验证要求。
- FB-9 缓存一致性影响评估：本次不修改生产缓存、缓存键、翻译包失效路径、跨实例同步、分布式协调或运行时状态；仅补充规则层面的 i18n 缓存治理入口和验证记录要求。
- FB-9 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界。
- FB-9 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤；仅引用既有跨平台 `go run ./hack/tools/linactl i18n.check` 治理入口。
- FB-9 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules/i18n.md .agents/skills/lina-feedback/SKILL.md .agents/skills/lina-review/SKILL.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md openspec/changes/extend-lina-review-interface-method-governance/specs/spec-governance/spec.md`，通过；已运行静态检索确认 OpenSpec 工具维护的 `.agents/skills/openspec-*` 与 `.agents/prompts/opsx/*` 文件没有残留本次插入的外部规则加载内容；已运行静态检索确认 `AGENTS.md`、`.agents/rules/i18n.md`、`lina-feedback`、`lina-review` 和 `spec-governance` 均包含外部规则入口、i18n 影响判断或验证要求。
- FB-9 Review：已按 `lina-review` 口径完成审查。审查范围限定为 `AGENTS.md`、`.agents/rules/i18n.md`、`.agents/skills/lina-feedback/SKILL.md`、`.agents/skills/lina-review/SKILL.md`、本变更 `tasks.md` 和 `specs/spec-governance/spec.md`；未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据权限、运行时缓存、前端 UI、语言包资源或 OpenSpec 工具维护文件；Go 编译门禁和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-10 影响分析：将 `AGENTS.md` 改造为顶层规范入口，保留文档、开发流程、架构、接口、开发工具、后端、SQL、接口层、服务层、控制器层、软删除、前端、单元测试、E2E、I18N 和 UI 等原有规范标题，并在每个标题下声明唯一事实来源规则文件、强制触发场景和执行门禁；新增 `.agents/rules/documentation.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/frontend-ui.md` 和 `.agents/rules/testing.md` 承载迁移后的领域细则；更新 `lina-feedback` 和 `lina-review`，要求按 `AGENTS.md` 强制规则加载矩阵读取所有命中规则文件；扩展 `spec-governance` 增量规范记录通用规则加载治理。
- FB-10 已读取规则文件：已按 `AGENTS.md` 新入口读取 `.agents/rules/documentation.md`、`.agents/rules/openspec.md`、`.agents/rules/testing.md` 和 `.agents/rules/i18n.md`；本反馈属于项目治理文档变更，确认不涉及生产 Go、HTTP API、SQL/DAO、运行时缓存、数据权限、源码插件实现、前端运行时页面或开发工具执行逻辑，其他领域规则无实现影响。
- FB-10 i18n 影响评估：本次只调整项目治理规范、规则文件和反馈/审查技能说明，不新增、修改或删除运行时用户可见文案、前端 UI 文案、API 文档源文本、宿主或插件 `manifest/i18n`、`apidoc i18n JSON`、插件清单、语言配置或翻译缓存；`AGENTS.md` 继续要求所有功能变更评估 `i18n` 影响，并将细则指向 `.agents/rules/i18n.md`。
- FB-10 缓存一致性影响评估：本次不修改生产缓存、缓存键、失效路径、快照刷新、跨实例同步、分布式协调或运行时状态；仅将缓存一致性治理细则迁移到 `.agents/rules/cache-consistency.md` 并由 `AGENTS.md` 强制触发读取。
- FB-10 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界；仅将数据权限治理细则迁移到 `.agents/rules/data-permission.md` 并由 `AGENTS.md` 强制触发读取。
- FB-10 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤；仅将开发工具跨平台治理细则迁移到 `.agents/rules/dev-tooling.md` 并由 `AGENTS.md` 强制触发读取。
- FB-10 测试策略：本反馈属于项目治理类变更，不新增单元测试或 E2E；使用 OpenSpec 严格校验、规则文件存在性检查、静态检索和 `git diff --check` 作为验证证据。
- FB-10 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules .agents/skills/lina-feedback/SKILL.md .agents/skills/lina-review/SKILL.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md openspec/changes/extend-lina-review-interface-method-governance/specs/spec-governance/spec.md`，通过；已运行文件存在性检查确认 `AGENTS.md` 强制加载矩阵引用的 13 个 `.agents/rules/*.md` 文件均存在；已运行静态检索确认 `AGENTS.md` 保留原有规范标题并声明对应规则文件，`lina-feedback` 与 `lina-review` 均包含通用规则加载矩阵和未读取命中规则文件无效的门禁，`spec-governance` 包含通用规则加载、审查报告和阻断场景要求。
- FB-10 Review：已按 `lina-review` 口径完成审查。审查范围包括 `AGENTS.md`、`.agents/rules/*.md`、`.agents/skills/lina-feedback/SKILL.md`、`.agents/skills/lina-review/SKILL.md`、本变更 `tasks.md` 和 `specs/spec-governance/spec.md`；已读取规则文件为 `.agents/rules/documentation.md`、`.agents/rules/openspec.md`、`.agents/rules/testing.md` 和 `.agents/rules/i18n.md`。未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据库结构、数据权限、运行时缓存、前端 UI、语言包资源或开发工具脚本；Go 编译门禁和 E2E 不适用，OpenSpec 校验、diff 空白检查、规则文件存在性检查和静态检索均已通过。
- FB-11 影响分析：将 `AGENTS.md` 中原“源码插件目录结构规范”扩展为“插件目录结构规范”，并在强制加载矩阵中明确插件通用资源、源码插件目录、动态插件目录和插件生命周期资源均由 `.agents/rules/plugin.md` 承载；重写 `.agents/rules/plugin.md`，区分插件通用资源、源码插件目录结构、源码插件数据库访问、动态插件目录结构、动态插件 manifest 与授权、动态插件产物资源视图、宿主接缝、验证和审查要求；在 `spec-governance` 中新增插件目录规则必须区分通用资源、源码插件和动态插件的场景。
- FB-11 已读取规则文件：已按 `AGENTS.md` 新入口读取 `.agents/rules/plugin.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md` 和 `.agents/rules/i18n.md`；本反馈属于项目治理文档变更，未修改动态插件或源码插件生产实现。
- FB-11 i18n 影响评估：本次只调整插件目录治理规范和 OpenSpec 记录，不新增、修改或删除运行时用户可见文案、前端 UI 文案、API 文档源文本、宿主或插件 `manifest/i18n`、`apidoc i18n JSON`、插件清单、语言配置或翻译缓存；规则中仅声明插件 i18n 资源继续遵守 `.agents/rules/i18n.md`。
- FB-11 缓存一致性影响评估：本次不修改生产缓存、缓存键、动态插件运行时派生缓存、失效路径、快照刷新、跨实例同步、分布式协调或运行时状态；规则中仅声明动态插件刷新、升级、启用、禁用和卸载后应按既有运行时与缓存规则处理派生资源。
- FB-11 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界；规则中仅声明动态插件 `data` host service 资源边界继续遵守 `.agents/rules/data-permission.md`。
- FB-11 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤；规则中仅声明动态插件 guest 或 bridge 变更需要运行 WASM 构建、`linactl wasm` 或等价验证。
- FB-11 测试策略：本反馈属于项目治理类变更，不新增单元测试或 E2E；使用 OpenSpec 严格校验、静态检索、文件内容检查和 `git diff --check` 作为验证证据。
- FB-11 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules/plugin.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md openspec/changes/extend-lina-review-interface-method-governance/specs/spec-governance/spec.md`，通过；已运行静态检索确认 `AGENTS.md` 和 `.agents/rules/plugin.md` 均包含插件通用资源、源码插件专属结构、动态插件专属结构、WASM 构建入口、`pluginbridge`、`hostServices` 和动态插件产物资源视图要求；已运行静态检索确认 `spec-governance` 包含插件目录规则区分通用资源、源码插件和动态插件的场景。
- FB-11 Review：已按 `lina-review` 口径完成审查。审查范围包括 `AGENTS.md`、`.agents/rules/plugin.md`、本变更 `tasks.md` 和 `specs/spec-governance/spec.md`；已读取规则文件为 `.agents/rules/plugin.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md` 和 `.agents/rules/i18n.md`。未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据库结构、数据权限、运行时缓存、前端 UI、语言包资源、动态插件产物或开发工具脚本；Go 编译门禁、WASM 构建和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-12 影响分析：调整 `AGENTS.md` 插件目录结构触发场景、`.agents/rules/plugin.md` 插件规则、`lina-feedback` 和 `lina-review` 的规则加载入口，并更新 `spec-governance` 增量规范；明确源码插件和动态插件必须共享 `backend/api/`、`backend/plugin.go`、`backend/internal/controller/`、`backend/internal/service/` 后端业务开发结构，动态插件的 `main.go`、`go.mod`、WASM 构建入口、`pluginbridge`、`hostServices` 和发布产物视图只表达运行时对接与交付差异。
- FB-12 已读取规则文件：已按 `AGENTS.md` 新入口读取 `.agents/rules/plugin.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md` 和 `.agents/rules/i18n.md`；本反馈属于项目治理文档变更，未修改动态插件或源码插件生产实现。
- FB-12 i18n 影响评估：本次只调整插件目录治理规范、反馈/审查技能入口和 OpenSpec 记录，不新增、修改或删除运行时用户可见文案、前端 UI 文案、API 文档源文本、宿主或插件 `manifest/i18n`、`apidoc i18n JSON`、插件清单、语言配置或翻译缓存。
- FB-12 缓存一致性影响评估：本次不修改生产缓存、缓存键、动态插件运行时派生缓存、失效路径、快照刷新、跨实例同步、分布式协调或运行时状态；仅保留动态插件启停、刷新和卸载后继续按既有运行时与缓存规则处理派生资源的治理要求。
- FB-12 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界；规则中继续要求动态插件涉及宿主数据访问时通过 `hostServices` 资源边界和宿主授权协议接入。
- FB-12 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤；规则中仅说明动态插件 guest、controller、service、bridge 或 WASM 入口变更需要运行 WASM 构建、`linactl wasm`、插件构建 smoke 或等价验证。
- FB-12 测试策略：本反馈属于项目治理类变更，不新增单元测试或 E2E；使用 OpenSpec 严格校验、静态检索和 `git diff --check` 作为验证证据。
- FB-12 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- AGENTS.md .agents/rules/plugin.md .agents/skills/lina-feedback/SKILL.md .agents/skills/lina-review/SKILL.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md openspec/changes/extend-lina-review-interface-method-governance/specs/spec-governance/spec.md`，通过；已运行静态检索确认旧口径 `动态插件不要求 controller/service`、`不应强制创建`、`不得被强制套用`、`错误要求采用源码插件 controller/service` 等冲突措辞无命中；已运行静态检索确认 `AGENTS.md`、`.agents/rules/plugin.md`、`lina-feedback`、`lina-review` 和 `spec-governance` 均包含插件同构开发、共享后端开发结构、动态插件必须维护 `backend/api/`、`backend/internal/controller/`、`backend/internal/service/` 以及桥接层不得替代 controller/service 分层的要求。
- FB-12 Review：已按 `lina-review` 口径完成审查。审查范围包括 `AGENTS.md`、`.agents/rules/plugin.md`、`.agents/skills/lina-feedback/SKILL.md`、`.agents/skills/lina-review/SKILL.md`、本变更 `tasks.md` 和 `specs/spec-governance/spec.md`；已读取规则文件为 `.agents/rules/plugin.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md` 和 `.agents/rules/i18n.md`。未发现阻塞问题。该反馈不修改生产 Go 代码、REST API 行为、数据库结构、数据权限、运行时缓存、前端 UI、语言包资源、动态插件产物或开发工具脚本；Go 编译门禁、WASM 构建和 E2E 不适用，OpenSpec 校验、diff 空白检查和静态检索均已通过。
- FB-13 根因分析：`lina-review` 已声明 `AGENTS.md` 和 `.agents/rules/*.md` 是唯一事实来源，但文件后半部分仍复制维护接口契约、后端 Go、SQL、E2E、缓存、数据权限和开发工具等领域规则细则，导致规则来源重复、后续维护容易漂移。
- FB-13 影响分析：重写 `.agents/skills/lina-review/SKILL.md`，将其收敛为审查编排、范围收集、规则加载、报告结构和阻塞语义；新增 `spec-governance` 场景，要求 `lina-review` 不重复维护领域规则细则；不修改生产代码、运行时配置、API、数据库、插件产物或前端行为。
- FB-13 已读取规则文件：已按 `AGENTS.md` 新入口读取 `.agents/rules/documentation.md`、`.agents/rules/openspec.md`、`.agents/rules/testing.md` 和 `.agents/rules/i18n.md`；本反馈属于项目治理文档变更，确认不涉及生产 Go、HTTP API、SQL/DAO、运行时缓存、数据权限、源码插件实现、前端运行时页面或开发工具执行逻辑，其他领域规则无实现影响。
- FB-13 i18n 影响评估：本次只调整审查技能说明、OpenSpec 任务记录和治理规范，不新增、修改或删除运行时用户可见文案、前端 UI 文案、API 文档源文本、宿主或插件 `manifest/i18n`、`apidoc i18n JSON`、插件清单、语言配置或翻译缓存。
- FB-13 缓存一致性影响评估：本次不修改生产缓存、缓存键、失效路径、快照刷新、跨实例同步、分布式协调或运行时状态；仅要求缓存审查细则回到 `.agents/rules/cache-consistency.md`。
- FB-13 数据权限影响评估：本次不新增或修改数据操作接口、查询条件、详情读取、写操作、下载、聚合统计、插件 host service 数据访问或角色数据权限边界；仅要求数据权限审查细则回到 `.agents/rules/data-permission.md`。
- FB-13 开发工具跨平台影响评估：本次不新增或修改开发工具、脚本、CI 命令实现、默认测试入口或平台专属运维步骤；仅要求开发工具审查细则回到 `.agents/rules/dev-tooling.md`。
- FB-13 测试策略：本反馈属于项目治理类变更，不新增单元测试或 E2E；使用 OpenSpec 严格校验、静态检索和 `git diff --check` 作为验证证据。
- FB-13 验证：已运行 `openspec validate extend-lina-review-interface-method-governance --strict`，通过；已运行 `git diff --check -- .agents/skills/lina-review/SKILL.md openspec/changes/extend-lina-review-interface-method-governance/tasks.md openspec/changes/extend-lina-review-interface-method-governance/specs/spec-governance/spec.md`，通过；已运行静态检索确认 `lina-review` 不再保留 `Unix timestamp`、`time.Time`、`ON DUPLICATE`、`Dependencies`、`bizerr`、`gerror`、`context.Background`、`cluster.enabled`、`TC{NNN}`、`internal/util`、`文件命名规范`、`显式依赖注入`、`日志上下文传播`、`数据权限接入` 等领域细则关键词。
