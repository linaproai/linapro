## Context

仓库已有 `Nightly Build` 工作流负责完整测试和镜像发布。OpenSpec 归档治理最初依赖人工在本地触发 `lina-auto-archive` 和 `lina-archive-consolidate`，后续演进为 monthly GitHub Actions 自动化流程。实践发现基础归档步骤依赖 AI Coding 工具执行时，AI 进程可能返回成功却未完成实际归档，因此将基础归档改为确定性 GitHub Action 直接运行 `openspec archive -y`，AI 工具仅用于可选的归档聚合增强。

`.github/codex/` 和 `.github/cc/` 提供 AI Coding 工具配置模板。真实认证文件不能进入版本库，代理服务 endpoint 也不应固化在仓库中，因此需要将可提交配置模板与运行时密钥、运行时 endpoint 拆开治理。

后端源码方面，宿主 `internal/service` 和 `lina-core/pkg` 公共组件的主文件职责尚未完全统一，部分组件将接口契约、类型定义和大量实现逻辑混在同一个主文件中。接口方法注释已有"紧邻方法声明"的要求，但详细程度不足；文件顶部注释已有存在性要求，但尚未要求说明主要实现逻辑和注意事项。`hack/tools/linactl` 命令实现文件也缺乏按命令名命名和子组件组织的统一规范。

## Goals / Non-Goals

**Goals:**

- 提供独立的 monthly OpenSpec 归档工作流，不与 nightly build 镜像发布链路耦合。
- 将基础归档步骤改为确定性自动化，消除 AI 工具在基础归档阶段的不可靠性；AI 工具仅用于可选的归档聚合增强。
- 通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择归档聚合使用的 AI Coding 工具，支持 `codex`、`cc` 和 `copilot`，并将工具细节封装到 reusable workflow 中。
- 通过预检查、确定性归档断言、变更范围保护和 OpenSpec 校验降低自动化误写风险。
- 只在本次自动归档产生变更时执行归档聚合，避免无意义重写聚合文档。
- 确定性归档阶段的失败不阻塞已成功归档结果的 PR 写回；AI 聚合阶段的失败不阻塞已通过校验的确定性归档 PR。
- 将后端主文件定位为"组件契约入口"，覆盖宿主服务、源码插件后端服务和 `lina-core/pkg` 公共组件三类后端源码。
- 强化接口方法注释，使接口定义文件本身足以说明方法的使用方式、输入输出、错误和重要约束。
- 强化文件顶部注释，使每个文件说明自身职责、主要实现逻辑和注意事项。
- 将规则写入 `AGENTS.md` 和 `lina-review` 审查标准，并通过分批整改使现有代码逐步收敛。

**Non-Goals:**

- 不修改 `lina-auto-archive` 或 `lina-archive-consolidate` 技能的语义。
- 不让每个 PR 合并后立即归档。
- 不改变 REST API、数据库结构、路由绑定、权限语义、缓存语义或业务行为。
- 不为治理目的重命名导出接口、修改服务构造签名或改变调用方契约，除非现有代码本身违反更高优先级规范且需要单独任务说明。
- 不一次性机械移动全仓所有方法；每批只处理一组业务模块或公共组件。
- 不新增前端 UI、运行时语言包、manifest i18n 或 apidoc i18n 资源。
- 不在本迭代实现归档聚合输入 hash 索引；若后续发现聚合文档抖动，再单独治理。

## Decisions

### Decision 1: 确定性基础归档与可选 AI 聚合分离

基础归档步骤改为共享 composite action `.github/actions/monthly-openspec-auto-archive`，使用固定 OpenSpec CLI 版本执行 `openspec list --json`、按名称稳定排序 completed active changes、逐个运行 `openspec archive -y <change>`。每个候选归档后重新执行 `openspec list --json` 确认已离开活跃列表；当 OpenSpec 提供任务计数且 `completedTasks != totalTasks` 时直接记录失败。记录成功和失败的变更到 JSON outputs 和 job summary。

工具专属 reusable workflow 在 AI runtime 准备之前调用此确定性归档 action，然后检测 OpenSpec diff，只有存在 diff 时才准备对应 AI runtime 并执行 archive consolidation。AI 聚合失败或产生无效 OpenSpec 时，workflow 记录诊断日志、恢复确定性归档快照并重新校验，通过后继续归档 PR 写回。

替代方案是继续让 AI 工具执行基础归档，但实践中 AI 进程返回成功却未完成归档的情况无法通过工具退出码可靠检测，因此不采用。

### Decision 2: 独立 monthly 工作流与 UTC cron 映射

新增 `Monthly OpenSpec Archive` workflow，使用 `schedule` 和 `workflow_dispatch` 触发。GitHub Actions cron 使用 UTC，Asia/Shanghai 没有夏令时，北京时间每月 1 日 00:00 对应 UTC 上月最后一天 16:00。workflow 使用 31/30/28/29 日分组触发，并在检测 job 开始处通过 `github.event.schedule` 过滤闰年的 `2/28 16:00 UTC` 重复触发。

`workflow_dispatch` 不受月度 schedule 窗口限制，可从任意暴露该 workflow 的分支 ref 进入。手动触发时 workflow checkout 触发分支执行检测和归档，并以触发分支作为归档 PR 目标分支，PR 来源分支包含触发分支的安全化标识（格式为 `automation/monthly-openspec-archive-<branch-slug>`）。

备选方案是 PR merge 后立即触发，但归档聚合属于语义文档重写，放在每次 merge 后会增加文档噪声和 API 成本，因此不作为第一版实现。

### Decision 3: 主工作流路由与工具专属 reusable workflow

主 workflow `.github/workflows/monthly-openspec-archive.yml` 负责 schedule/manual 触发、月初门禁、OpenSpec 完成候选预检查和 AI Coding 工具路由。它读取 GitHub Variables 中的 `AI_CODING_TOOL`，未配置时默认使用 `codex`。合法取值为 `codex`、`cc`、`copilot`。

主 workflow 在检测阶段通过白名单 `case` 分支解析工具值，使用两个固定的路由 job 并通过 job-level `if` 保证每次只运行匹配工具的 reusable workflow。GitHub Actions 不支持在 `uses` 中使用表达式动态拼接 reusable workflow 路径。

Codex、Claude Code 和 GitHub Copilot CLI 的实现细节拆入独立 reusable workflow，各自完整执行 checkout、确定性归档、运行时准备、可选聚合、校验、变更范围保护、归档 PR 创建或更新和日志上传。工具差异保留在工具专属 workflow 中，公共治理步骤通过本地 composite action 复用。

### Decision 4: 本地 composite action 复用公共治理步骤

与 AI Coding 工具无关的公共治理步骤通过本地 composite action 复用：

- `.github/actions/monthly-openspec-setup`：统一执行 checkout 后的时区设置和 Node 准备。
- `.github/actions/monthly-openspec-auto-archive`：确定性扫描 completed active changes 并逐个运行 `openspec archive -y`。
- `.github/actions/monthly-openspec-detect-changes`：统一检测自动归档后 `openspec/` 是否产生变更。
- `.github/actions/monthly-openspec-assert-archive-complete`：自动归档后统一确认没有完成状态的活跃 OpenSpec 变更残留。
- `.github/actions/monthly-openspec-validate`：在自动归档、归档聚合和写回前按固定 OpenSpec CLI 版本执行全量校验。
- `.github/actions/monthly-openspec-finalize-pr`：统一执行 OpenSpec 校验、生成变更范围保护以及归档 PR 创建或更新。

本地 composite action 需要仓库 checkout 后才能被 runner 解析，因此默认分支 checkout 保留为各工具 workflow 的第一步。公共 action 只承载无工具差异的确定性治理逻辑，避免多份 workflow 中重复维护相同 shell 脚本。

### Decision 5: 公共提示词与运行时密钥隔离

归档聚合提示词统一维护在 `.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`，所有工具专属 reusable workflow 通过 stdin 引用同一份 prompt 文件。基础归档不再使用 AI prompt，改为确定性 action 直接执行。

运行时创建 `$RUNNER_TEMP` 下的工具 home，从仓库配置模板生成临时配置：

- `codex` 分支复制 `.github/codex/config.template.toml`，用 `vars.OPENAI_BASE_URL` 替换占位符，用 `secrets.OPENAI_API_KEY` 写入临时 `auth.json`。
- `cc` 分支读取 `.github/cc/settings.template.json`，用 `secrets.ANTHROPIC_AUTH_TOKEN`、`vars.ANTHROPIC_BASE_URL` 和 `vars.ANTHROPIC_CUSTOM_MODEL` 生成临时 `settings.json`。
- `copilot` 分支用 `secrets.COPILOT_GITHUB_TOKEN` 认证，用 `vars.COPILOT_MODEL` 选择模型（默认 `auto`），用 `vars.COPILOT_REASONING_EFFORT` 选择推理等级（合法值：空、`low`、`medium`、`high`、`xhigh`）；使用隔离的临时 `COPILOT_HOME`。

这样 GitHub Actions 可以复用仓库配置，同时避免密钥和代理 endpoint 进入工作区 diff、artifact 或提交历史。

### Decision 6: 阶段化失败策略

workflow 采用阶段化失败策略：

- 确定性归档后立即执行 `openspec list --json` 断言无完成状态活跃变更残留；存在残留时记录失败变更并继续处理剩余候选。
- 确定性归档后执行 `openspec validate --all`，失败时停止进入 AI 聚合阶段。
- AI 归档聚合后执行 `openspec validate --all`，失败或 AI 命令失败时记录诊断、恢复确定性归档快照并重新校验，通过后继续归档 PR 写回。
- 最终 PR 写回前保留 OpenSpec 校验作为兜底。
- 确定性归档有部分成功时，先创建或更新归档 PR，然后在所有候选处理完成后失败退出。

### Decision 7: PR 写回与仓库策略降级

workflow 采用 GitHub Actions bot 创建或更新维护 PR 的方式写回归档结果，使用固定维护分支 `automation/monthly-openspec-archive`（手动触发时包含源分支标识）。PR 标题为 `chore(openspec): archive completed changes`。

创建或更新 PR 前检查 diff 范围，只允许 PR 包含 `openspec/**` 变更。当仓库策略阻止 `GITHUB_TOKEN` 创建 PR 时，workflow 输出已推送的归档分支和手动 PR URL 并成功结束，不作为硬失败。

### Decision 8: 主文件作为契约入口，具体实现迁出

主文件定义为与组件目录同名的 Go 文件，例如 `user/user.go`、`plugin/plugin.go`、`pkg/dialect/dialect.go`。主文件保留：

- 组件级 package 注释。
- 组件级常量、枚举、输入输出类型、领域类型别名。
- `Service` 或公共契约接口，以及必要的窄接口。
- `serviceImpl` 或组件默认实现结构体。
- 编译期接口断言。
- `New`、`NewXxx` 或等价构造函数。
- 极短且无业务分支的类型方法可以保留，例如枚举 `String()` 或 `Int()`，但不得承载数据库访问、请求处理、缓存刷新、业务编排或外部调用。

具体业务实现迁到同包其他文件，优先按职责命名为 `<component>_<domain>.go`，例如 `user_list.go`、`plugin_runtime.go`。只有组件职责较小且没有更清晰子域时，才使用 `<component>_impl.go`。

替代方案是强制所有实现进入 `<component>_impl.go`。该方案文件数量少，但容易把复杂组件重新聚合成大文件，不利于按职责阅读，因此不采用。

### Decision 9: `lina-core/pkg` 使用同等主文件治理

`apps/lina-core/pkg` 是宿主和插件之间的稳定公共组件层，主文件同样应作为公共契约入口。`pkg/<component>/<component>.go` 可保留导出类型、公共接口、构造函数和轻量方法；复杂解析、编解码、校验、运行时执行和适配逻辑迁移到同包职责文件。

替代方案是只治理 `internal/service`。该方案覆盖面不足，会让公共契约组件继续混合契约与实现，影响插件开发者阅读稳定 API，因此不采用。

### Decision 10: 接口方法注释按"使用契约"编写

接口方法注释不要求机械罗列所有参数名称，但必须让调用者理解：

- 方法做什么，以及是否有副作用。
- 关键输入参数的语义、默认行为和约束。
- 返回值表达什么，空结果或零值如何理解。
- 可能返回哪些业务错误、权限错误、数据权限拒绝、配置错误或底层错误。
- 是否涉及权限、数据权限、租户、缓存、i18n、事务、幂等或并发注意事项。

简单无错误、无副作用的方法可以保持简洁，但不能只有重复方法名的一句话。复杂接口应在注释中说明关键边界，不把使用者迫使到实现文件里寻找约束。

### Decision 11: 文件顶部注释按文件职责分层

主文件使用 package 注释说明组件整体职责、边界和开发者阅读入口；非主文件使用文件注释说明该文件承载的实现切片、主要逻辑和注意事项，并与 `package` 声明保持一个空行。

注释不应复制实现细节或重复代码逐行说明。它应解释文件为什么存在、处理哪类逻辑、有什么重要约束。

### Decision 12: 按模块分批整改

整改任务按业务模块和公共组件分批推进。每批任务只移动同包实现、补充注释和运行对应包测试，不把多个无关模块混在同一提交/任务里。每批任务记录：本批修改范围、行为是否保持不变、i18n/缓存一致性/数据权限影响判断、运行的 `go test` 或替代编译验证、`lina-review` 发现与处理结果。

### Decision 13: `linactl` 命令文件命名与子组件组织

`hack/tools/linactl/` 下承载具体命令实现的源码文件必须按命令名称命名为 `command_<command>.go`，其中 `<command>` 保持命令的点分段语义。当命令名与 Go 工具链文件后缀规则冲突时使用命令专属后缀（如 `test` 使用 `command_testcmd.go`，`wasm` 使用 `command_wasmcmd.go`）。

根目录尽可能只保留 `command_*.go` 指令入口、`command.go` 注册与参数解析、`app.go`/`main.go` 启动装配、基础类型和必要的平台适配文件；复杂共享实现迁移到 `hack/tools/linactl/internal/<组件名称>/` 子组件。

## Risks / Trade-offs

- 自动聚合文档可能重复重写同一批日期归档目录 -> 仅在本次自动归档产生变更时运行聚合，降低无变更日抖动。
- 月度自动归档可能让刚完成的变更最多停留在活跃目录约一个月 -> 保留 `workflow_dispatch` 手动触发，并在反馈流程中仍以"是否归档"判定活跃变更，必要时人工归档。
- UTC cron 无法直接表达北京时间月初 -> 使用分组 cron 加二月闰年去重门禁。
- AI Coding 工具执行失败或 OpenSpec 校验失败 -> workflow 在当前阶段失败并保留日志，不继续执行后续阶段；AI 聚合失败时恢复确定性归档快照继续 PR 写回。
- Actions bot 需要 `pull-requests: write` 权限 -> 缺少权限时 workflow 失败；仓库策略阻止 PR 创建时降级为手动 PR 链接。
- 大规模文件迁移导致审查噪音 -> 按模块分批，只做同包移动和注释补强，不混入业务重构。
- 方法移动引入遗漏或重复定义 -> 每批运行覆盖变更包的 `go test <package> -count=1`，复杂组件增加包级编译验证。
- 注释过度膨胀降低可读性 -> 注释聚焦使用契约和文件职责，不逐行复述实现。
- 历史代码无法一次全部达标 -> 新规则先纳入规范和 review，现有代码通过任务清单逐步整改。
- `pkg` 公共组件迁移影响外部调用方 -> 保持包名、导出符号、函数签名和行为不变，仅调整文件承载位置。

## Migration Plan

1. 提交确定性归档 action、工具专属 reusable workflow、公共 composite action、配置模板和提示词文件。
2. 在 GitHub 仓库配置 `AI_CODING_TOOL` variable（默认 `codex`）；按所选工具配置对应 secrets 和 variables。
3. 通过 `workflow_dispatch` 手动触发一次，验证确定性归档、工具配置、OpenSpec 校验与写回权限。
4. 手动触发验证稳定后，保留 schedule 自动运行。
5. 更新 OpenSpec 增量规范、`AGENTS.md` 和 `lina-review` 审查标准。
6. 建立当前主文件职责问题清单，按宿主服务、`pkg` 公共组件、源码插件分组。
7. 先处理小型组件，验证迁移模式和 review 标准是否可执行。
8. 按业务模块处理宿主核心服务、系统治理、任务调度、插件运行时等复杂区域。
9. 按插件分组处理源码插件后端服务。
10. 执行全量静态扫描、OpenSpec 校验、分包 Go 编译门禁和 `lina-review` 终审。

回滚策略：归档自动化变更可通过恢复 workflow 和 action 文件回滚；源码治理变更为同包文件移动和注释补强，单批任务可通过恢复该批文件移动和注释变更回滚；若某批发现行为风险，暂停该批并保留已经通过的其他批次。
