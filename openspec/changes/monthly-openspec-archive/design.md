## Context

仓库已有 `Nightly Build` 工作流，负责完整测试和镜像发布。OpenSpec 归档治理目前仍依赖人工在本地触发 `lina-auto-archive` 和 `lina-archive-consolidate`，而两个技能都需要 AI Coding 工具执行上下文、OpenSpec CLI 和仓库文件系统写权限。

`.github/codex/` 和 `.github/cc/` 提供 AI Coding 工具配置模板。真实认证文件不能进入版本库，代理服务 endpoint 也不应固化在仓库中，因此需要将可提交配置模板与运行时密钥、运行时 endpoint 拆开治理。

## Goals / Non-Goals

**Goals:**
- 提供独立的 monthly OpenSpec 归档工作流，不与现有 nightly build 镜像发布链路耦合。
- 通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择运行两个既有技能的 AI Coding 工具，支持 `codex`、Claude Code 的 `cc` 和 GitHub Copilot CLI 的 `copilot` 取值，并将工具细节封装到 reusable workflow 中。
- 通过预检查、变更范围保护和 OpenSpec 校验降低自动化误写风险。
- 只在本次自动归档产生变更时执行归档聚合，避免无意义重写聚合文档。
- 运行时从 GitHub Secret 注入对应工具所需的 API key/token 和 provider `base_url`，不提交真实认证文件或真实服务 endpoint。

**Non-Goals:**
- 不修改 `lina-auto-archive` 或 `lina-archive-consolidate` 技能的语义。
- 不让每个 PR 合并后立即归档。
- 不新增后端、前端、数据库或运行时配置能力。
- 不在本迭代实现归档聚合输入 hash 索引；若后续发现聚合文档抖动，再单独治理。

## Decisions

### 独立 monthly 工作流

新增 `Monthly OpenSpec Archive` workflow，使用 `schedule` 和 `workflow_dispatch` 触发。定时任务选择每月 1 日 00:00 Asia/Shanghai 运行，降低每日触发带来的 AI 工具、OpenSpec 校验、PR 维护和 runner 资源消耗；维护者仍可通过 `workflow_dispatch` 在需要时手动执行归档。

GitHub Actions cron 使用 UTC，Asia/Shanghai 没有夏令时。北京时间每月 1 日 00:00 对应 UTC 上月最后一天 16:00。由于 GitHub cron 不能表达“每月最后一天”，workflow 使用 31/30/28/29 日分组触发，并在检测 job 开始处通过 `github.event.schedule` 过滤闰年的 `2/28 16:00 UTC` 重复触发；`2/29 16:00 UTC` 覆盖闰年 3 月 1 日北京时间 00:00。`workflow_dispatch` 不受该门禁限制。

备选方案是 PR merge 后立即触发，但归档聚合属于语义文档重写，放在每次 merge 后会增加文档噪声和 API 成本，因此不作为第一版实现。

### 主工作流路由

主 workflow `.github/workflows/monthly-openspec-archive.yml` 负责 schedule/manual 触发、默认分支限制、月初门禁、OpenSpec 完成候选预检查和 AI Coding 工具路由。它读取 GitHub Variables 中的 `AI_CODING_TOOL`，未配置时默认使用 `codex`。合法取值为：

- `codex`：调用 `.github/workflows/monthly-openspec-archive-codex.yml`。
- `cc`：调用 `.github/workflows/monthly-openspec-archive-cc.yml`。
- `copilot`：调用 `.github/workflows/monthly-openspec-archive-copilot.yml`。

主 workflow 在检测阶段通过白名单 `case` 分支解析工具值，避免把任意 reusable workflow、任意镜像名或任意命令直接暴露给仓库变量。GitHub Actions 不支持在 `jobs.<job_id>.uses` 中使用表达式动态拼接 reusable workflow 路径，因此主 workflow 使用两个固定的路由 job，并通过 job-level `if` 保证每次只运行匹配工具的 reusable workflow。切换工具时只需要修改 GitHub Actions Variables 中的 `AI_CODING_TOOL`，不需要修改 workflow YAML。

### 工具专属 reusable workflow

Codex 和 Claude Code 的实现细节拆入独立 reusable workflow：

- `.github/workflows/monthly-openspec-archive-codex.yml`：使用 `loads/codex:latest`、`.github/codex/config.template.toml`、`codex exec`、OpenAI secrets 和 Codex 专属日志 artifact。
- `.github/workflows/monthly-openspec-archive-cc.yml`：使用 `loads/cc:latest`、`.github/cc/settings.template.json`、`claude -p`、Anthropic secrets/model 配置和 Claude Code 专属日志 artifact。
- `.github/workflows/monthly-openspec-archive-copilot.yml`：使用 `@github/copilot`、`copilot -p`、`COPILOT_GITHUB_TOKEN` 认证、`COPILOT_MODEL` 模型配置、`COPILOT_REASONING_EFFORT` 推理等级配置和 Copilot CLI 专属日志 artifact。

reusable workflow 是 job 级调用，不与主 workflow 共享 checkout 后的工作区或 runner 临时目录。因此工具专属 workflow 必须各自完整执行 checkout、运行时准备、`lina-auto-archive`、差异检测、条件性 `lina-archive-consolidate`、OpenSpec 校验、变更范围保护、归档 PR 创建或更新和日志上传。工具差异保留在工具专属 workflow 中，公共治理步骤通过本地 composite action 复用，避免通过 artifact 在 workflow 之间传递 patch，也让后续新增 AI Coding 工具只需要新增一个工具 workflow 和复用现有公共 action。

工具专属 workflow 中与 AI Coding 工具无关的公共治理步骤通过本地 composite action 复用：

- `.github/actions/monthly-openspec-setup`：统一执行 checkout 后的时区设置和 Node 准备。
- `.github/actions/monthly-openspec-detect-changes`：统一检测自动归档后 `openspec/` 是否产生变更。
- `.github/actions/monthly-openspec-assert-archive-complete`：自动归档后统一确认没有完成状态的活跃 OpenSpec 变更残留。
- `.github/actions/monthly-openspec-validate`：在自动归档、归档聚合和写回前按固定 OpenSpec CLI 版本执行全量校验。
- `.github/actions/monthly-openspec-finalize-pr`：统一执行 OpenSpec 校验、生成变更范围保护以及归档 PR 创建或更新。

本地 composite action 需要仓库 checkout 后才能被 runner 解析，因此默认分支 checkout 保留为各工具 workflow 的第一步；`monthly-openspec-setup` 复用 checkout 之后的时区和 Node 准备。工具 workflow 仍保留各自的认证配置、AI 容器调用和日志 artifact 上传，因为这些步骤与 Codex/Claude Code 的运行时约束强相关。公共 action 只承载无工具差异的确定性治理逻辑，避免两份 workflow 中重复维护相同 shell 脚本。

AI 工具执行日志必须同时满足实时排障和事后审查两个目标。Codex 分支保留 `--output-last-message` 作为最终消息 artifact，同时将 `docker run` 的 stdout/stderr 通过 `tee` 写入 task console log，因此 `codex exec` 的过程日志会出现在当前 Actions step 中，并由外层 `set -o pipefail` 保留失败退出码。Claude Code 分支不能再把 `claude -p` stdout 直接重定向到文件；应在容器内用 `tee` 将 stdout/stderr 写入 task markdown artifact，同时显式捕获并返回 `claude -p` 的原始退出码，避免日志透传掩盖 AI 工具进程失败。Copilot CLI 分支将 `copilot -p` 的 stdout/stderr 通过 `tee` 写入 step 日志和 console artifact，并使用 `--no-ask-user`、有限 `--allow-tool` 规则和最终 OpenSpec 范围守卫保证无人值守执行不会静默等待人工确认或写回非 OpenSpec 文件。

AI 工具的退出码不能作为 monthly 阶段成功的唯一判断依据。工具专属 workflow 在 `Run Lina Auto Archive` 后立即通过 `.github/actions/monthly-openspec-assert-archive-complete` 重新执行 `openspec list --json`，若仍存在 `complete`、`completed` 或 `done` 状态的活跃变更，则判定自动归档阶段未完成并失败，不再执行差异检测、归档聚合或 PR 写回。自动归档阶段还会立即执行 `openspec validate --all`，避免已经无效的 OpenSpec 状态继续进入聚合阶段。`Run Lina Archive Consolidate` 后也立即执行同一 OpenSpec 校验，聚合输出无效时直接失败并停止在 PR 收尾之前。最终 PR action 保留 OpenSpec 校验作为写回前兜底，但前置阶段检查负责让失败尽早暴露在对应阶段。

### 公共提示词文件

自动归档与归档聚合提示词统一维护在 `.github/prompts/`：

- `.github/prompts/monthly-openspec-auto-archive.zh-CN.md`
- `.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`

Codex 和 Claude Code reusable workflow 都通过 stdin 读取同一份 prompt 文件。这样避免在工具 workflow 中重复维护 prompt 正文，同时保留 prompt 作为普通仓库文件，避免通过 reusable workflow outputs 传递多行文本带来的转义、大小限制和额外 job 复杂度。

### 运行时 AI home

工具专属 workflow 不直接使用仓库内真实认证文件。运行时创建 `$RUNNER_TEMP` 下的工具 home：

- `codex` 分支复制 `.github/codex/config.template.toml`，将其中的 `base_url` 占位符替换为 `vars.OPENAI_BASE_URL`，再用 `secrets.OPENAI_API_KEY` 写入临时 `auth.json`。
- `cc` 分支读取 `.github/cc/settings.template.json`，用 `secrets.ANTHROPIC_AUTH_TOKEN`、`vars.ANTHROPIC_BASE_URL` 和 `vars.ANTHROPIC_CUSTOM_MODEL` 生成临时 `settings.json`。
- `copilot` 分支用 `secrets.COPILOT_GITHUB_TOKEN` 认证 GitHub Copilot CLI，用 `vars.COPILOT_MODEL` 选择模型，并用 `vars.COPILOT_REASONING_EFFORT` 选择推理等级；未配置模型时使用 `auto` 并交由 GitHub Copilot CLI 默认模型路由处理，未配置推理等级时不传递 `--reasoning-effort`，保持 CLI 默认推理配置。

这样 GitHub Actions 可以复用仓库配置，同时避免密钥和代理 endpoint 进入工作区 diff、artifact 或提交历史。

### 预检查与条件执行

主 workflow 先运行 `openspec list --json`，仅当存在 `complete`、`completed` 或 `done` 状态的活跃变更时才调用工具专属 reusable workflow。归档技能仍会执行更保守的任务勾选和状态检查，预检查只用于节省无效运行成本。

工具专属 workflow 自动归档后通过 `git diff --quiet -- openspec` 判断本次是否产生 OpenSpec 变更。只有产生变更时才调用归档聚合技能。

### 写回策略

workflow 采用 GitHub Actions bot 创建或更新维护 PR 的方式写回归档结果。workflow 限制只在默认分支运行，并设置 `contents: write` 与 `pull-requests: write`。PR 使用固定维护分支 `automation/monthly-openspec-archive`，每次有新的归档结果时用当前默认分支重新生成该分支的单个归档提交，并创建或更新标题为 `chore(openspec): archive completed changes` 的 PR。

创建或更新 PR 前检查 diff 范围，monthly 自动运行时只允许 PR 包含 `openspec/**` 变更，避免 AI Coding 工具意外修改业务代码、workflow 配置或本地密钥治理规则。OpenSpec 校验失败或存在允许范围外的文件变更时，workflow 失败且不创建或更新归档 PR。

## Risks / Trade-offs

- 自动聚合文档可能重复重写同一批日期归档目录 → 仅在本次自动归档产生变更时运行聚合，降低无变更日抖动。
- 月度自动归档可能让刚完成的变更最多停留在活跃目录约一个月 → 保留 `workflow_dispatch` 手动触发，并在反馈流程中仍以“是否归档”判定活跃变更，必要时人工归档。
- UTC cron 无法直接表达北京时间月初 → 使用分组 cron 加二月闰年去重门禁，二月 28/29 日在闰年和平年都能覆盖且不会重复执行。
- AI Coding 工具执行失败、自动归档后仍存在完成状态活跃变更，或 OpenSpec 校验失败 → workflow 在当前阶段失败并保留日志，不继续执行后续阶段，也不写回不完整结果。
- Actions bot 需要 `pull-requests: write` 权限创建或更新维护 PR → 缺少权限时 workflow 会失败并保留日志，不会绕过分支保护直接写入默认分支。
- `loads/codex:latest` 或 `loads/cc:latest` 若不可拉取会导致 workflow 无法启动 → workflow 使用用户提供镜像，失败时由 GitHub Actions 明确暴露镜像拉取问题。
- `@github/copilot` 当前要求 Node.js 24 或更高版本 → monthly 公共 setup 固定使用 Node.js 24，避免 Copilot CLI 安装后启动失败。
- `AI_CODING_TOOL` 配置错误会导致运行失败 → workflow 仅允许 `codex`、`cc` 和 `copilot`，并在准备运行时时输出明确错误。
- 本地 composite action 增加了一层跳转 → 用明确命名和窄输入承载公共治理逻辑，工具 workflow 只保留工具差异，降低重复脚本漂移风险。
- reusable workflow 的工作区不共享 → 工具 workflow 自己完成写回闭环，避免主 workflow 和子 workflow 之间传递 patch artifact。
- `lina-archive-consolidate` 在归档数量较多时包含完整 OpenSpec 流程提示 → CI prompt 明确要求按无人值守 monthly 场景执行，并在无法安全继续时失败而不是交互等待。

## Migration Plan

1. 提交 Codex/Claude Code 配置模板、主路由 workflow 和工具专属 reusable workflow。
2. 在 GitHub 仓库配置 `AI_CODING_TOOL` variable；未配置时默认使用 `codex`。
3. 使用 `codex` 时配置 `OPENAI_API_KEY` secret 和 `OPENAI_BASE_URL` variable；使用 `cc` 时配置 `ANTHROPIC_AUTH_TOKEN` secret，并配置 `ANTHROPIC_BASE_URL`、`ANTHROPIC_CUSTOM_MODEL` variables；使用 `copilot` 时配置 `COPILOT_GITHUB_TOKEN` secret，并按需配置 `COPILOT_MODEL`、`COPILOT_REASONING_EFFORT` variables。
4. 通过 `workflow_dispatch` 手动触发一次，验证镜像、工具配置、OpenSpec CLI 与写回权限。
5. 手动触发验证稳定后，保留 schedule 自动运行。
