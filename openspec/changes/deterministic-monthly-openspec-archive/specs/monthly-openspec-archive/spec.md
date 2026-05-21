## MODIFIED Requirements

### Requirement: Monthly workflow must archive completed OpenSpec changes
系统 SHALL provide a GitHub Actions monthly workflow that scans `openspec/changes/` on the target branch for completed active changes and runs repository auto-archive logic through the configured AI coding tool runtime before archive consolidation.

#### Scenario: Scheduled archive run
- **WHEN** monthly OpenSpec 归档工作流按计划触发
- **AND** 当前 Asia/Shanghai 日期为每月 1 日
- **THEN** workflow 在仓库默认分支 checkout 代码
- **AND** workflow invokes the configured AI Coding tool to run `.github/prompts/monthly-openspec-auto-archive.zh-CN.md`
- **AND** auto archive uses the repository `lina-auto-archive` rules to run `openspec archive -y <change>` for safe completed active changes
- **AND** workflow invokes archive consolidation only after tool-driven auto archive produces OpenSpec file changes that need consolidation

#### Scenario: Monthly schedule window
- **WHEN** GitHub Actions schedule 事件在 UTC 触发
- **THEN** workflow 使用 UTC 月末 cron 分组覆盖 Asia/Shanghai 每月 1 日 00:00
- **AND** workflow 在闰年跳过 `2/28 16:00 UTC` 的重复 schedule 事件
- **AND** workflow 在平年使用 `2/28 16:00 UTC` 覆盖 Asia/Shanghai 3 月 1 日 00:00
- **AND** workflow 在 `2/29 16:00 UTC` 存在时覆盖闰年 Asia/Shanghai 3 月 1 日 00:00

#### Scenario: Manual archive run
- **WHEN** 维护者通过 `workflow_dispatch` 手动触发 monthly OpenSpec 归档工作流
- **THEN** workflow 不受月度 schedule 窗口限制
- **AND** workflow may be dispatched from any branch ref that exposes the workflow
- **AND** workflow checkout 该触发分支并在该分支内容上执行 OpenSpec 完成候选预检查和工具运行时自动归档
- **AND** workflow 以该触发分支作为归档 PR 的目标分支
- **AND** workflow 使用包含该触发分支安全化标识的归档 PR 来源分支

#### Scenario: No completed active changes
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** `openspec list --json` 未报告任何 `complete`、`completed` 或 `done` 状态的活跃变更
- **THEN** workflow skips tool-specific auto archive execution
- **AND** workflow 不调用 AI Coding 工具归档任务
- **AND** workflow 成功结束且不创建或更新归档 PR

#### Scenario: Archive command fails for one candidate
- **WHEN** tool-specific auto archive execution processes multiple completed active changes
- **AND** `openspec archive -y <change>` fails for one candidate
- **THEN** workflow records the failed change name and archive error summary
- **AND** workflow immediately fails the auto archive step
- **AND** workflow does not process additional archive candidates
- **AND** workflow does not run archive consolidation
- **AND** workflow does not create or update an archive PR

### Requirement: Monthly workflow must consolidate only after new archive changes
系统 SHALL 仅在本次 monthly 工具运行时自动归档产生 OpenSpec 文件变更后执行 `lina-archive-consolidate` 技能，避免无新增归档时重复重写聚合归档文档。

#### Scenario: Archive produced changes
- **WHEN** tool-specific auto archive 执行后 `openspec/` 下存在新的文件变更
- **THEN** workflow 调用 `lina-archive-consolidate` 聚合已归档变更
- **AND** workflow 在聚合后执行临时变更清理检查和 OpenSpec 校验
- **AND** workflow stops before PR finalization if archive consolidation, temporary change cleanup, or OpenSpec validation fails

#### Scenario: Archive produced no changes
- **WHEN** tool-specific auto archive 执行完成
- **AND** `openspec/` 下没有新的文件变更
- **THEN** workflow 跳过 `lina-archive-consolidate`
- **AND** workflow 不创建或更新归档 PR

### Requirement: Monthly workflow must select the AI Coding tool from GitHub Variables
系统 SHALL 通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择 monthly OpenSpec 自动归档和归档聚合使用的 AI Coding 工具，并 SHALL 在未配置该变量时默认使用 `codex`。

#### Scenario: Default Codex tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 未配置 `AI_CODING_TOOL`
- **AND** tool-specific auto archive produced OpenSpec file changes
- **THEN** 主 workflow 调用 Codex reusable workflow
- **AND** Codex reusable workflow 使用 `loads/codex:latest` 和 `codex exec` 运行 archive consolidation

#### Scenario: Explicit Codex tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `codex`
- **AND** tool-specific auto archive produced OpenSpec file changes
- **THEN** 主 workflow 调用 Codex reusable workflow
- **AND** Codex reusable workflow 使用 `loads/codex:latest` 和 `codex exec` 运行 archive consolidation

#### Scenario: Explicit Claude Code tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `cc`
- **AND** tool-specific auto archive produced OpenSpec file changes
- **THEN** 主 workflow 调用 Claude Code reusable workflow
- **AND** Claude Code reusable workflow 使用 `loads/cc:latest` 和 `claude -p` 运行 archive consolidation

#### Scenario: Explicit GitHub Copilot CLI tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `copilot`
- **AND** tool-specific auto archive produced OpenSpec file changes
- **THEN** 主 workflow 调用 GitHub Copilot CLI reusable workflow
- **AND** GitHub Copilot CLI reusable workflow 使用 `@github/copilot` 和 `copilot -p` 运行 archive consolidation
- **AND** GitHub Copilot CLI reusable workflow 使用 `COPILOT_MODEL` variable 配置模型，未配置时默认使用 `auto`
- **AND** GitHub Copilot CLI reusable workflow 使用 `COPILOT_REASONING_EFFORT` variable 配置推理等级，未配置时不传递显式推理等级
- **AND** workflow 仅接受空值、`low`、`medium`、`high` 或 `xhigh` 作为 Copilot 推理等级

#### Scenario: Unsupported tool value
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 不是 `codex`、`cc` 或 `copilot`
- **THEN** 主 workflow 在执行任何工具 reusable workflow 前失败
- **AND** workflow 不创建或更新归档 PR

### Requirement: Monthly workflow must isolate tool implementations in reusable workflows
系统 SHALL 将不同 AI Coding 工具的运行时准备、镜像调用、认证配置和日志上传细节封装在工具专属 reusable workflow 中，并 SHALL 让主 workflow 只负责触发、候选检测和路由。Base auto archive and archive consolidation MUST both run through the selected tool-specific runtime.

#### Scenario: Codex implementation is isolated
- **WHEN** 所选工具为 Codex
- **THEN** 主 workflow 调用 `.github/workflows/monthly-openspec-archive-codex.yml`
- **AND** Codex reusable workflow uses `loads/codex:latest` and `codex exec` for both auto archive and archive consolidation
- **AND** Codex reusable workflow 独立完成 checkout、按需聚合、校验、变更范围保护、归档 PR 创建或更新和日志上传

#### Scenario: GitHub Copilot CLI implementation is isolated
- **WHEN** 所选工具为 GitHub Copilot CLI
- **THEN** 主 workflow 调用 `.github/workflows/monthly-openspec-archive-copilot.yml`
- **AND** GitHub Copilot CLI reusable workflow uses `@github/copilot` and `copilot -p` for both auto archive and archive consolidation
- **AND** GitHub Copilot CLI reusable workflow 独立完成 checkout、按需聚合、校验、变更范围保护、归档 PR 创建或更新和日志上传

#### Scenario: Claude Code implementation is isolated
- **WHEN** 所选工具为 Claude Code
- **THEN** 主 workflow 调用 `.github/workflows/monthly-openspec-archive-cc.yml`
- **AND** Claude Code reusable workflow uses `loads/cc:latest` and `claude -p` for both auto archive and archive consolidation
- **AND** Claude Code reusable workflow 独立完成 checkout、按需聚合、校验、变更范围保护、归档 PR 创建或更新和日志上传

#### Scenario: Only one tool workflow runs
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** `AI_CODING_TOOL` 为任一合法值
- **THEN** workflow 仅运行匹配该工具的 reusable workflow
- **AND** workflow 不运行其他工具的 reusable workflow

### Requirement: Monthly workflow must share prompt files across AI tools
系统 SHALL 将 monthly OpenSpec 自动归档和归档聚合提示词维护为 `.github/prompts/` 下的公共文件，并 SHALL 让所有工具专属 reusable workflow 引用同一份自动归档提示词和同一份聚合提示词内容。

#### Scenario: Shared archive consolidate prompt
- **WHEN** 任一工具专属 reusable workflow 执行 `lina-archive-consolidate`
- **THEN** workflow 从 `.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md` 读取提示词
- **AND** workflow 不在工具专属 workflow 中内联维护重复的归档聚合提示词正文

#### Scenario: Shared auto archive prompt
- **WHEN** 任一工具专属 reusable workflow 执行 base auto archive
- **THEN** workflow 从 `.github/prompts/monthly-openspec-auto-archive.zh-CN.md` 读取提示词
- **AND** workflow 通过当前已选择的 Codex、Claude Code 或 GitHub Copilot CLI 运行时执行该提示词

### Requirement: Monthly workflow must stream AI tool execution logs
系统 SHALL 在 monthly OpenSpec 归档聚合执行期间，将 AI Coding 工具进程的标准输出和标准错误实时写入 GitHub Actions step 日志，并在前序步骤均成功时保留 artifact 日志用于事后排查。

#### Scenario: Archive consolidate logs are visible
- **WHEN** 任一工具专属 reusable workflow 执行 `Run Lina Archive Consolidate`
- **THEN** `codex exec`、`claude -p` 或 `copilot -p` 进程的标准输出和标准错误在当前 GitHub Actions step 日志中可见
- **AND** workflow uploads the corresponding AI tool log artifact only when all prior workflow steps succeed
- **AND** 日志透传不得掩盖 AI 工具进程的失败退出码

#### Scenario: Auto archive logs are visible
- **WHEN** 任一工具专属 reusable workflow 执行 base auto archive
- **THEN** `codex exec`、`claude -p` 或 `copilot -p` 进程的标准输出和标准错误在当前 GitHub Actions step 日志中可见
- **AND** 日志透传不得掩盖 AI 工具进程的失败退出码

### Requirement: Monthly workflow must stop on any step failure
系统 SHALL 在 monthly OpenSpec 自动归档、归档聚合、校验、变更范围保护、提交、推送和 PR 创建或更新的任一步骤失败时立即停止后续归档步骤。失败步骤 MUST expose the failing command output in the GitHub Actions log, and workflow MUST NOT restore an earlier archive snapshot or continue to PR finalization after a failed archive consolidation step.

#### Scenario: Auto archive leaves completed changes active
- **WHEN** tool-specific auto archive finishes
- **AND** `openspec list --json` 仍报告 `complete`、`completed` 或 `done` 状态的活跃变更
- **THEN** workflow outputs the remaining change names, status values, and task counts
- **AND** workflow fails immediately
- **AND** workflow does not run archive consolidation
- **AND** workflow does not create or update an archive PR

#### Scenario: Auto archive produces invalid OpenSpec state
- **WHEN** tool-specific auto archive produces OpenSpec file changes
- **AND** `openspec validate --all` 执行失败
- **THEN** workflow fails before archive consolidation
- **AND** workflow 不创建或更新归档 PR

#### Scenario: Archive consolidation produces invalid OpenSpec state
- **WHEN** 任一工具专属 reusable workflow 执行 `Run Lina Archive Consolidate`
- **AND** AI 工具进程返回成功
- **AND** `openspec validate --all` 执行失败
- **THEN** workflow fails immediately at the validation step
- **AND** workflow does not restore an earlier archive snapshot
- **AND** workflow does not create or update an archive PR

#### Scenario: Archive consolidation command fails
- **WHEN** 任一工具专属 reusable workflow 执行 `Run Lina Archive Consolidate`
- **AND** AI 工具进程返回失败
- **THEN** workflow fails immediately at the consolidation step
- **AND** workflow does not run temporary change cleanup
- **AND** workflow does not validate OpenSpec after archive consolidation
- **AND** workflow does not create or update an archive PR

#### Scenario: Repository policy blocks pull request creation
- **WHEN** monthly OpenSpec 归档 workflow 已成功提交并推送归档来源分支
- **AND** GitHub repository policy prevents `GITHUB_TOKEN` from creating or updating pull requests
- **THEN** workflow fails at the pull request command step
- **AND** workflow does not treat repository pull request policy failures as successful archive completion
