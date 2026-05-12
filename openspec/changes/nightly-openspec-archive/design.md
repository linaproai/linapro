## Context

仓库已有 `Nightly Build` 工作流，负责完整测试和镜像发布。OpenSpec 归档治理目前仍依赖人工在本地触发 `lina-auto-archive` 和 `lina-archive-consolidate`，而两个技能都需要 AI Coding 工具执行上下文、OpenSpec CLI 和仓库文件系统写权限。

`.github/codex/` 和 `.github/cc/` 提供 AI Coding 工具配置模板。真实认证文件不能进入版本库，代理服务 endpoint 也不应固化在仓库中，因此需要将可提交配置模板与运行时密钥、运行时 endpoint 拆开治理。

## Goals / Non-Goals

**Goals:**
- 提供独立的 nightly OpenSpec 归档工作流，不与现有 nightly build 镜像发布链路耦合。
- 通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择运行两个既有技能的 AI Coding 工具，支持 `codex` 和 Claude Code 的 `cc` 取值，并将工具细节封装到 reusable workflow 中。
- 通过预检查、变更范围保护和 OpenSpec 校验降低自动化误写风险。
- 只在本次自动归档产生变更时执行归档聚合，避免每天无意义重写聚合文档。
- 运行时从 GitHub Secret 注入对应工具所需的 API key/token 和 provider `base_url`，不提交真实认证文件或真实服务 endpoint。

**Non-Goals:**
- 不修改 `lina-auto-archive` 或 `lina-archive-consolidate` 技能的语义。
- 不让每个 PR 合并后立即归档。
- 不新增后端、前端、数据库或运行时配置能力。
- 不在本迭代实现归档聚合输入 hash 索引；若后续发现聚合文档抖动，再单独治理。

## Decisions

### 独立 nightly 工作流

新增 `Nightly OpenSpec Archive` workflow，使用 `schedule` 和 `workflow_dispatch` 触发。定时任务选择 Asia/Shanghai 凌晨后段运行，对应 UTC 前一日晚上，避免与现有 `Nightly Build` 的 00:00 Asia/Shanghai 执行窗口重叠。

备选方案是 PR merge 后立即触发，但归档聚合属于语义文档重写，放在每次 merge 后会增加文档噪声和 API 成本，因此不作为第一版实现。

### 主工作流路由

主 workflow `.github/workflows/nightly-openspec-archive.yml` 负责 schedule/manual 触发、默认分支限制、OpenSpec 完成候选预检查和 AI Coding 工具路由。它读取 GitHub Variables 中的 `AI_CODING_TOOL`，未配置时默认使用 `codex`。合法取值为：

- `codex`：调用 `.github/workflows/nightly-openspec-archive-codex.yml`。
- `cc`：调用 `.github/workflows/nightly-openspec-archive-cc.yml`。

主 workflow 在检测阶段通过白名单 `case` 分支解析工具值，避免把任意 reusable workflow、任意镜像名或任意命令直接暴露给仓库变量。GitHub Actions 不支持在 `jobs.<job_id>.uses` 中使用表达式动态拼接 reusable workflow 路径，因此主 workflow 使用两个固定的路由 job，并通过 job-level `if` 保证每次只运行匹配工具的 reusable workflow。切换工具时只需要修改 GitHub Actions Variables 中的 `AI_CODING_TOOL`，不需要修改 workflow YAML。

### 工具专属 reusable workflow

Codex 和 Claude Code 的实现细节拆入独立 reusable workflow：

- `.github/workflows/nightly-openspec-archive-codex.yml`：使用 `loads/codex:latest`、`.github/codex/config.template.toml`、`codex exec`、OpenAI secrets 和 Codex 专属日志 artifact。
- `.github/workflows/nightly-openspec-archive-cc.yml`：使用 `loads/cc:latest`、`.github/cc/settings.template.json`、`claude -p`、Anthropic secrets/model 配置和 Claude Code 专属日志 artifact。

reusable workflow 是 job 级调用，不与主 workflow 共享 checkout 后的工作区或 runner 临时目录。因此工具专属 workflow 必须各自完整执行 checkout、运行时准备、`lina-auto-archive`、差异检测、条件性 `lina-archive-consolidate`、OpenSpec 校验、变更范围保护、自动提交和日志上传。这样虽然保留少量公共治理步骤重复，但避免了通过 artifact 在 workflow 之间传递 patch 的复杂度，也让后续新增 AI Coding 工具只需要新增一个工具 workflow 和一个主路由 job。

### 公共提示词文件

自动归档与归档聚合提示词统一维护在 `.github/prompts/`：

- `.github/prompts/nightly-openspec-auto-archive.zh-CN.md`
- `.github/prompts/nightly-openspec-archive-consolidate.zh-CN.md`

Codex 和 Claude Code reusable workflow 都通过 stdin 读取同一份 prompt 文件。这样避免在工具 workflow 中重复维护 prompt 正文，同时保留 prompt 作为普通仓库文件，避免通过 reusable workflow outputs 传递多行文本带来的转义、大小限制和额外 job 复杂度。

### 运行时 AI home

工具专属 workflow 不直接使用仓库内真实认证文件。运行时创建 `$RUNNER_TEMP` 下的工具 home：

- `codex` 分支复制 `.github/codex/config.template.toml`，将其中的 `base_url` 占位符替换为 `secrets.OPENAI_BASE_URL`，再用 `secrets.OPENAI_API_KEY` 写入临时 `auth.json`。
- `cc` 分支读取 `.github/cc/settings.template.json`，用 `secrets.ANTHROPIC_AUTH_TOKEN`、`secrets.ANTHROPIC_BASE_URL` 和 `vars.ANTHROPIC_CUSTOM_MODEL`（或同名 secret）生成临时 `settings.json`。

这样 GitHub Actions 可以复用仓库配置，同时避免密钥和代理 endpoint 进入工作区 diff、artifact 或提交历史。

### 预检查与条件执行

主 workflow 先运行 `openspec list --json`，仅当存在 `complete`、`completed` 或 `done` 状态的活跃变更时才调用工具专属 reusable workflow。归档技能仍会执行更保守的任务勾选和状态检查，预检查只用于节省无效运行成本。

工具专属 workflow 自动归档后通过 `git diff --quiet -- openspec` 判断本次是否产生 OpenSpec 变更。只有产生变更时才调用归档聚合技能。

### 写回策略

第一版采用 GitHub Actions bot 直接提交回触发分支。workflow 限制只在默认分支运行，并设置 `contents: write`。提交前检查 diff 范围，nightly 自动运行时只允许提交 `openspec/**` 变更，避免 AI Coding 工具意外修改业务代码、workflow 配置或本地密钥治理规则。

如果仓库分支保护禁止 Actions bot 直推，后续可以把提交步骤替换为创建维护 PR；本迭代保留直接提交作为最小闭环。

## Risks / Trade-offs

- 自动聚合文档可能重复重写同一批日期归档目录 → 仅在本次自动归档产生变更时运行聚合，降低无变更日抖动。
- AI Coding 工具执行失败或 OpenSpec 校验失败 → workflow 失败并保留日志，不提交不完整结果。
- Actions bot 直推可能受分支保护限制 → 该失败是可见的治理失败，后续可改为创建 PR。
- `loads/codex:latest` 或 `loads/cc:latest` 若不可拉取会导致 workflow 无法启动 → workflow 使用用户提供镜像，失败时由 GitHub Actions 明确暴露镜像拉取问题。
- `AI_CODING_TOOL` 配置错误会导致运行失败 → workflow 仅允许 `codex` 和 `cc`，并在准备运行时时输出明确错误。
- 工具专属 reusable workflow 会重复 checkout、guard、commit 等公共治理步骤 → 当前重复是有意设计，用少量重复换取工具实现隔离；若后续工具数量继续增长，可再把公共治理逻辑抽成共享 shell 脚本或 composite action。
- reusable workflow 的工作区不共享 → 工具 workflow 自己完成写回闭环，避免主 workflow 和子 workflow 之间传递 patch artifact。
- `lina-archive-consolidate` 在归档数量较多时包含完整 OpenSpec 流程提示 → CI prompt 明确要求按无人值守 nightly 场景执行，并在无法安全继续时失败而不是交互等待。

## Migration Plan

1. 提交 Codex/Claude Code 配置模板、主路由 workflow 和工具专属 reusable workflow。
2. 在 GitHub 仓库配置 `AI_CODING_TOOL` variable；未配置时默认使用 `codex`。
3. 使用 `codex` 时配置 `OPENAI_API_KEY` 和 `OPENAI_BASE_URL` secrets；使用 `cc` 时配置 `ANTHROPIC_AUTH_TOKEN`、`ANTHROPIC_BASE_URL` secrets 与 `ANTHROPIC_CUSTOM_MODEL` variable 或 secret。
4. 通过 `workflow_dispatch` 手动触发一次，验证镜像、工具配置、OpenSpec CLI 与写回权限。
5. 手动触发验证稳定后，保留 schedule 自动运行。
