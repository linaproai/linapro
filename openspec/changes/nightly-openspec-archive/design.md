## Context

仓库已有 `Nightly Build` 工作流，负责完整测试和镜像发布。OpenSpec 归档治理目前仍依赖人工在本地触发 `lina-auto-archive` 和 `lina-archive-consolidate`，而两个技能都需要 Codex 执行上下文、OpenSpec CLI 和仓库文件系统写权限。

`.github/codex/` 已存在本地模板，但目录当前被 `.gitignore` 整体忽略，GitHub Actions checkout 后无法读取这些模板。真实 `auth.json` 不能进入版本库，代理服务的 `base_url` 也不应固化在仓库中，因此需要将可提交配置模板与运行时密钥、运行时 endpoint 拆开治理。

## Goals / Non-Goals

**Goals:**
- 提供独立的 nightly OpenSpec 归档工作流，不与现有 nightly build 镜像发布链路耦合。
- 使用 `loads/codex:latest` 和仓库内 Codex 配置模板运行两个既有技能。
- 通过预检查、变更范围保护和 OpenSpec 校验降低自动化误写风险。
- 只在本次自动归档产生变更时执行归档聚合，避免每天无意义重写聚合文档。
- 运行时从 GitHub Secret 注入 API key 和 Codex provider `base_url`，不提交真实认证文件或真实服务 endpoint。

**Non-Goals:**
- 不修改 `lina-auto-archive` 或 `lina-archive-consolidate` 技能的语义。
- 不让每个 PR 合并后立即归档。
- 不新增后端、前端、数据库或运行时配置能力。
- 不在本迭代实现归档聚合输入 hash 索引；若后续发现聚合文档抖动，再单独治理。

## Decisions

### 独立 nightly 工作流

新增 `Nightly OpenSpec Archive` workflow，使用 `schedule` 和 `workflow_dispatch` 触发。定时任务选择 Asia/Shanghai 凌晨后段运行，对应 UTC 前一日晚上，避免与现有 `Nightly Build` 的 00:00 Asia/Shanghai 执行窗口重叠。

备选方案是 PR merge 后立即触发，但归档聚合属于语义文档重写，放在每次 merge 后会增加文档噪声和 API 成本，因此不作为第一版实现。

### 运行时 Codex home

workflow 不直接使用仓库内 `.github/codex/auth.json`。运行时创建 `$RUNNER_TEMP/codex-home` 作为 `CODEX_HOME`，复制 `.github/codex/config.toml`，将其中的 `base_url` 占位符替换为 `secrets.OPENAI_BASE_URL`，再用 `secrets.OPENAI_API_KEY` 写入临时 `auth.json`。这样 GitHub Actions 可以复用仓库配置，同时避免密钥和代理 endpoint 进入工作区 diff、artifact 或提交历史。

### 预检查与条件执行

workflow 先运行 `openspec list --json`，仅当存在 `complete`、`completed` 或 `done` 状态的活跃变更时才调用 Codex。归档技能仍会执行更保守的任务勾选和状态检查，预检查只用于节省无效运行成本。

自动归档后通过 `git diff --quiet -- openspec` 判断本次是否产生 OpenSpec 变更。只有产生变更时才调用归档聚合技能。

### 写回策略

第一版采用 GitHub Actions bot 直接提交回触发分支。workflow 限制只在默认分支运行，并设置 `contents: write`。提交前检查 diff 范围，nightly 自动运行时只允许提交 `openspec/**` 变更，避免 Codex 意外修改业务代码、workflow 配置或本地密钥治理规则。

如果仓库分支保护禁止 Actions bot 直推，后续可以把提交步骤替换为创建维护 PR；本迭代保留直接提交作为最小闭环。

## Risks / Trade-offs

- 自动聚合文档可能重复重写同一批日期归档目录 → 仅在本次自动归档产生变更时运行聚合，降低无变更日抖动。
- Codex 执行失败或 OpenSpec 校验失败 → workflow 失败并保留日志，不提交不完整结果。
- Actions bot 直推可能受分支保护限制 → 该失败是可见的治理失败，后续可改为创建 PR。
- `loads/codex:latest` 若不可拉取会导致 workflow 无法启动 → workflow 使用用户提供镜像，失败时由 GitHub Actions 明确暴露镜像拉取问题。
- `lina-archive-consolidate` 在归档数量较多时包含完整 OpenSpec 流程提示 → CI prompt 明确要求按无人值守 nightly 场景执行，并在无法安全继续时失败而不是交互等待。

## Migration Plan

1. 提交 Codex 配置模板和 nightly workflow。
2. 在 GitHub 仓库配置 `OPENAI_API_KEY` 和 `OPENAI_BASE_URL` secrets。
3. 通过 `workflow_dispatch` 手动触发一次，验证镜像、Codex 配置、OpenSpec CLI 与写回权限。
4. 手动触发验证稳定后，保留 schedule 自动运行。
