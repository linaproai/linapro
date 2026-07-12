---
name: lina-community-fix-ci
description: >-
  排查并修复给定的 LinaPro GitHub Actions 失败问题；修复完成后保留本地改动，
  禁止自动提交、推送或创建 PR。
  必须用户手动触发，禁止自动触发该技能。
compatibility: 需要 GitHub CLI `gh` 已登录且具备读取 Actions run/log 权限；需要 `git` 与本地可执行的 make/测试命令。
---

# Lina Community Fix CI

手动排查并修复 `LinaPro` 仓库中指定的 `GitHub Actions` 失败。该技能负责收集失败证据、定位根因、实施最小必要修复，并在本地尽量复现对应校验；**修复完成后不得自动提交代码**。

## 核心规则

1. **只手动触发**：用户必须明确要求排查或修复 CI 失败；不得因看到 CI 红、PR 检查失败或对话中出现 Actions 链接而自动启动本技能。
2. **默认仓库**：`linaproai/linapro`。用户指定其他仓库时使用用户指定值。
3. **目标明确**：优先处理用户给出的 run URL、run ID、job 名、PR 编号或 workflow 名；未给出时先列出最近失败 run，让用户确认目标，不得盲修所有失败。
4. **先证据后改码**：必须先读失败 job 日志、失败步骤、相关 workflow 定义和本地对应代码，再判断根因；禁止在未说明根因或合理假设前直接改代码。
5. **最小必要修复**：只改与本次 CI 失败直接相关的文件；禁止顺手重构、格式化无关文件、升级无关依赖或扩大修复范围。
6. **禁止自动提交**：修复完成后**不得**执行 `git add`、`git commit`、`git push`、`gh pr create`、`gh pr edit`，也不得调用 `lina-community-commit-push-and-pr`。本地改动保留给用户审查。
7. **禁止静默改写历史**：不得 `git reset --hard`、`git checkout --`、`git clean`、`--force` 推送或丢弃用户已有未提交改动。
8. **尊重工作区**：执行前检查 `git status`；若存在无关未提交改动，只修改本次修复相关路径，不还原、不混入无关变更。
9. **规则门禁**：修改代码前按 `AGENTS.md` 判断命中的规则域，读取对应 `.agents/rules/*.md`；未读取命中规则不得改生产代码、测试或 workflow。
10. **本地验证优先**：能本地复现的失败必须本地复现并验证修复；无法本地复现时明确说明原因、已做检查和下一步建议。
11. **不伪造通过**：本地验证未覆盖到失败步骤时，不得声称“CI 已修复”或“下次 run 一定通过”；只能报告已确认的本地结果和剩余风险。

## 输入识别

自然识别以下用户请求：

- `lina-community-fix-ci`
- `修复 CI 失败`
- `排查 GitHub Actions 失败`
- `fix this failing action`
- `处理 PR #123 的 CI`
- `修复 run 1234567890`
- 粘贴的 Actions run URL、job URL 或失败日志片段

### 目标解析优先级

按顺序尽量解析出以下信息：

| 优先级 | 输入 | 解析方式 |
| --- | --- |
| 1 | Actions run URL / job URL | 提取 `owner/repo`、`run_id`，必要时提取 `job_id` |
| 2 | run ID | 在默认或指定仓库查询该 run |
| 3 | PR 编号 | 查询该 PR 最新失败 check run / workflow run |
| 4 | workflow 名 + 分支 | 查询该 workflow 在指定分支上最近失败 run |
| 5 | 仅说“修 CI” | 列出最近失败 run，请用户确认后再修 |

示例命令：

```bash
# run URL 示例：https://github.com/linaproai/linapro/actions/runs/1234567890
# job URL 示例：.../actions/runs/1234567890/job/9876543210

gh run view "$RUN_ID" -R "$REPO" --json databaseId,workflowName,displayTitle,headBranch,headSha,event,status,conclusion,url,jobs
gh pr checks "$PR_NUMBER" -R "$REPO"
gh run list -R "$REPO" --status failure --limit 10 \
  --json databaseId,workflowName,displayTitle,headBranch,headSha,event,status,conclusion,url,createdAt
```

## 前置检查

在修改任何文件前先做只读检查：

```bash
gh auth status
git rev-parse --show-toplevel
git remote -v
git status --short --branch
git branch --show-current
git rev-parse HEAD
```

确认：

1. 当前工作区是目标仓库可信检出，或用户明确允许在当前检出中修复。
2. `gh` 能读取目标仓库 Actions。
3. 当前分支/提交与失败 run 的关系清楚：
   - 同分支同提交：可直接修。
   - 同分支但本地已领先/落后：先报告差异，再决定基于当前工作区修复还是先对齐失败 SHA。
   - 完全无关分支：停止并说明需要切换到失败分支/检出对应提交，除非用户要求在当前分支前向移植修复。

如果权限不足、不在可信仓库、工作区有高风险脏状态且无法安全隔离修复，停止并报告阻断原因。

## 失败证据收集

### 1. 定位失败 run 与 job

```bash
gh run view "$RUN_ID" -R "$REPO" \
  --json databaseId,workflowName,displayTitle,headBranch,headSha,event,status,conclusion,url,jobs,createdAt,updatedAt

gh run view "$RUN_ID" -R "$REPO" --log-failed
```

日志过大时：

```bash
gh run view "$RUN_ID" -R "$REPO" --job "$JOB_ID" --log
# 或下载完整日志后本地检索
gh api "repos/$REPO/actions/runs/$RUN_ID/logs" > "/tmp/ci-run-$RUN_ID.zip"
```

必须提取：

- workflow 名称与文件路径（如 `.github/workflows/main-ci.yml`）
- 失败 job 名称
- 失败 step 名称
- 关键错误行、断言失败、编译错误、lint 结果、超时信息
- 触发事件（`push` / `pull_request` / `workflow_dispatch` / `schedule`）
- 失败提交 `headSha` 与分支 `headBranch`
- 是否 matrix 中的部分组合失败

### 2. 读取 workflow 定义

根据失败 workflow 读取本地或失败 SHA 对应定义：

```bash
# 优先本地当前工作区
sed -n '1,200p' .github/workflows/<workflow>.yml

# 需要对照失败提交时
git show "$HEAD_SHA:.github/workflows/<workflow>.yml"
```

重点确认：

- 实际执行了哪些 reusable workflow / composite action
- 关键输入参数（如是否开启 E2E、OpenSpec、Windows smoke）
- 使用的 make 目标、脚本路径、工作目录、环境变量
- 缓存、依赖安装、数据库/Redis 等外部依赖假设

### 3. 映射到本地验证命令

将失败步骤映射为可本地执行的等价命令。常见映射：

| CI 信号 | 优先本地验证 |
| --- | --- |
| Go 编译/单测失败 | 对应包 `go test` 或项目约定 make 目标 |
| Go lint 失败 | 与 workflow 相同的 lint 命令 |
| 前端单测失败 | 对应前端测试命令 |
| i18n check 失败 | i18n 校验命令 |
| make command smoke 失败 | 失败日志中的具体 make 命令 |
| OpenSpec 完成性检查失败 | `openspec` 相关校验 |
| 插件相关 job 失败 | 插件目录内对应测试/命令 |
| Redis/集成测试失败 | 确认本地依赖后运行对应测试 |
| E2E 失败 | 按 `lina-e2e` 与项目测试规范运行相关用例 |
| 仅 CI 环境问题（runner、权限、secrets、配额） | 不改业务代码；报告为环境/配置问题 |

不得猜测命令；以 workflow YAML、composite action 和失败日志中的实际命令为准。

## 根因分析

在改代码前输出简短根因分析：

```markdown
### CI 失败根因分析
- 目标：run <id> / job <name> / step <name>
- 仓库与引用：<repo> @ <branch> (<sha>)
- 失败摘要：<一句话>
- 根因判断：<代码缺陷 / 测试脆弱 / 工作流配置 / 环境依赖 / 权限或密钥 / flaky / 信息不足>
- 证据：<日志关键行、源码路径、workflow 路径>
- 拟修复范围：<文件或模块列表>
- 本地复现计划：<命令>
- 是否需要改 workflow：是/否（默认否；只有证据表明 workflow 本身错误才改）
```

分类处理：

1. **代码/测试缺陷**：修复源码或测试，补齐必要断言或稳定化测试。
2. **workflow/脚本配置错误**：最小修改 `.github/workflows/`、`.github/actions/` 或 `hack/` 脚本；同时读取 `dev-tooling` 等命中规则。
3. **依赖/环境假设错误**：优先让本地与 CI 假设对齐；确需改安装/启动逻辑时做最小修复。
4. **Flaky / 竞态 / 超时**：先证明是稳定性问题，再做确定性修复（等待条件、隔离、重试边界、超时调整需有依据）。
5. **外部系统或权限问题**：不改业务代码冒充修复；报告所需权限、secret、runner 或上游服务问题。
6. **信息不足**：继续收集日志/工件；仍不足时停止并向用户要补充信息，不进行猜测式大改。

## 修复实施

### 1. 读取命中规则

根据拟修改路径读取 `AGENTS.md` 与命中的 `.agents/rules/*.md`。常见映射：

- Go 后端：`backend-go.md`
- API/DTO：`api-contract.md`
- SQL/DAO：`database.md`
- 前端：`frontend-ui.md`
- 测试：`testing.md`
- i18n：`i18n.md`
- 插件：`plugin.md`
- 工具脚本/Makefile/CI 脚本：`dev-tooling.md`
- OpenSpec 文档/任务：`openspec.md`
- 文档：`documentation.md`

插件目录变更前，还要检查并遵守 `apps/lina-plugins/<plugin-id>/AGENTS.md`（若存在）。

### 2. 实施最小修复

- 只改消除失败所必需的代码、测试、脚本或 workflow。
- 保持现有风格和项目约定。
- 不把“让 CI 变绿”建立在删除有效测试、削弱断言、盲目 skip、或忽略错误之上，除非有充分证据证明该检查本身错误且修复检查定义更合适。
- 若失败暴露的是真实产品缺陷，按缺陷修复，而不是只改测试来迁就错误行为。
- 若修复触及用户可观察行为，评估是否需要补充或更新测试。

### 3. 本地验证

按映射命令验证：

1. 先跑与失败步骤最接近的最小命令，确认原失败点已消失。
2. 再跑直接相邻的回归检查（同包测试、同 workflow job 内前置步骤对应命令）。
3. 有明确相关的 lint/format/i18n/编译门禁时一并验证。

验证记录必须包含命令、退出码和关键输出摘要。失败则继续分析，不得在红的情况下宣布完成。

### 4. 明确禁止的收尾动作

修复与验证结束后，**停止**。以下动作全部禁止，除非用户在本技能之外另行明确要求：

- `git add` / `git commit` / `git commit --amend`
- `git push` / `git push --force` / `git push --force-with-lease`
- `gh pr create` / `gh pr comment` / `gh pr edit` / `gh workflow run`
- 自动调用 `lina-community-commit-push-and-pr`
- 删除或还原用户已有的无关本地改动

可以向用户说明“修复已完成，本地未提交；如需提交和开 PR，可手动触发 `lina-community-commit-push-and-pr`”，但不得自行执行。

## 特殊场景

### PR 来自外部 fork

- 可以基于失败日志分析，并在当前可写工作区准备修复。
- 不得把密钥写入日志、代码或报告。
- 不运行不可信 PR 分支中来路不明的安装脚本；优先使用仓库既有 make/测试入口。

### 多 job 失败

- 先按日志判断是否同一根因。
- 同一根因：一次修复后分别验证相关命令。
- 不同根因：分项列出，按阻塞优先级逐个修；中途向用户同步进度。
- 不要只修第一个失败就宣称全部完成。

### 仅 schedule / nightly 失败

- 确认是否与主 CI 路径不同（E2E、镜像构建、发布步骤等）。
- 本地无法承担的长时任务要说明，并给出尽可能接近的替代验证。

### 失败已在后续提交修复

- 对比失败 SHA 与当前分支是否已包含修复。
- 若已修复：报告证据，不再重复改码。
- 若未完全修复：只补剩余缺口。

## 最终报告

处理结束后用简洁结构汇报：

```markdown
## CI 修复结果

- 目标仓库：
- 失败 run：
- workflow / job / step：
- 失败提交与分支：
- 根因：
- 修改文件：
- 本地验证：
  - 命令与结果：
- 未覆盖风险：
- Git 状态：仍有未提交改动 / 工作区干净
- 提交状态：**未提交**（本技能禁止自动提交）

### 建议下一步
- 用户审查 diff
- 如需提交/推送/开 PR：手动触发 `lina-community-commit-push-and-pr` 或自行提交
- 如需重跑 Actions：由用户在 GitHub 重跑，或用户明确要求后再执行 `gh run rerun`
```

最终报告不得包含密钥、token、完整敏感日志或无关大段 diff。

## 护栏规则

- **先日志后改码** — 没有失败证据不改文件
- **先根因后方案** — 不靠删测试或 skip 换绿灯
- **最小 diff** — 不修无关问题
- **遵守 AGENTS.md** — 命中规则必须先读
- **本地能验则必验** — 验证失败不得宣称完成
- **永不自动提交** — 修完即停，等待用户决定是否提交
- **永不自动推送/开 PR** — 与提交同样禁止
- **不破坏用户工作区** — 不丢弃、不混入无关改动
)
