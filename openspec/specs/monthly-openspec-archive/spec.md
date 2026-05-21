# monthly-openspec-archive Specification

## Purpose
TBD - created by archiving change monthly-openspec-archive. Update Purpose after archive.
## Requirements
### Requirement: Monthly workflow must archive completed OpenSpec changes
系统 SHALL provide a GitHub Actions monthly workflow that deterministically scans `openspec/changes/` on the default branch for completed active changes and runs the pinned OpenSpec CLI archive command for each archive candidate before any optional AI-assisted consolidation.

#### Scenario: Scheduled archive run
- **WHEN** monthly OpenSpec 归档工作流按计划触发
- **AND** 当前 Asia/Shanghai 日期为每月 1 日
- **THEN** workflow 在仓库默认分支 checkout 代码
- **AND** workflow uses the pinned OpenSpec CLI to run `openspec archive -y <change>` for each completed active change
- **AND** workflow only invokes the configured AI Coding tool after deterministic archiving produces OpenSpec file changes that need consolidation

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
- **AND** workflow checkout 该触发分支并在该分支内容上执行 OpenSpec 完成候选预检查和确定性归档
- **AND** workflow 以该触发分支作为归档 PR 的目标分支
- **AND** workflow 使用包含该触发分支安全化标识的归档 PR 来源分支

#### Scenario: No completed active changes
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** `openspec list --json` 未报告任何 `complete`、`completed` 或 `done` 状态的活跃变更
- **THEN** workflow skips deterministic archive execution
- **AND** workflow 不调用 AI Coding 工具归档任务
- **AND** workflow 成功结束且不创建或更新归档 PR

#### Scenario: Archive command fails for one candidate
- **WHEN** deterministic archive execution processes multiple completed active changes
- **AND** `openspec archive -y <change>` fails for one candidate
- **THEN** workflow continues processing remaining candidates
- **AND** workflow records the failed change name and archive error summary
- **AND** workflow may still create or update an archive PR for successful archive results
- **AND** workflow fails after PR finalization if any completed active change remains unarchived

### Requirement: Monthly workflow must consolidate only after new archive changes
系统 SHALL 仅在本次 monthly deterministic 自动归档产生 OpenSpec 文件变更后执行 `lina-archive-consolidate` 技能，避免无新增归档时重复重写聚合归档文档。

#### Scenario: Archive produced changes
- **WHEN** deterministic auto archive 执行后 `openspec/` 下存在新的文件变更
- **THEN** workflow 调用 `lina-archive-consolidate` 聚合已归档变更
- **AND** workflow 在聚合后继续执行 OpenSpec 校验

#### Scenario: Archive produced no changes
- **WHEN** deterministic auto archive 执行完成
- **AND** `openspec/` 下没有新的文件变更
- **THEN** workflow 跳过 `lina-archive-consolidate`
- **AND** workflow 不创建或更新归档 PR

### Requirement: Monthly workflow must select the AI Coding tool from GitHub Variables
系统 SHALL 通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择 monthly OpenSpec 归档聚合使用的 AI Coding 工具，并 SHALL 在未配置该变量时默认使用 `codex`。该变量 MUST NOT control deterministic base archiving.

#### Scenario: Default Codex tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 未配置 `AI_CODING_TOOL`
- **AND** deterministic archive produced OpenSpec file changes
- **THEN** 主 workflow 调用 Codex reusable workflow
- **AND** Codex reusable workflow 使用 `loads/codex:latest` 和 `codex exec` 运行 archive consolidation

#### Scenario: Explicit Codex tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `codex`
- **AND** deterministic archive produced OpenSpec file changes
- **THEN** 主 workflow 调用 Codex reusable workflow
- **AND** Codex reusable workflow 使用 `loads/codex:latest` 和 `codex exec` 运行 archive consolidation

#### Scenario: Explicit Claude Code tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `cc`
- **AND** deterministic archive produced OpenSpec file changes
- **THEN** 主 workflow 调用 Claude Code reusable workflow
- **AND** Claude Code reusable workflow 使用 `loads/cc:latest` 和 `claude -p` 运行 archive consolidation

#### Scenario: Explicit GitHub Copilot CLI tool
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `copilot`
- **AND** deterministic archive produced OpenSpec file changes
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
系统 SHALL 将不同 AI Coding 工具的运行时准备、镜像调用、认证配置和日志上传细节封装在工具专属 reusable workflow 中，并 SHALL 让主 workflow 只负责触发、候选检测和路由。确定性基础归档 MUST be shared by the tool-specific workflows and MUST NOT be implemented separately by each AI tool.

#### Scenario: Codex implementation is isolated
- **WHEN** 所选工具为 Codex
- **THEN** 主 workflow 调用 `.github/workflows/monthly-openspec-archive-codex.yml`
- **AND** Codex reusable workflow uses the shared deterministic archive action before any Codex runtime setup
- **AND** Codex reusable workflow 独立完成 checkout、可选聚合、校验、变更范围保护、归档 PR 创建或更新和日志上传

#### Scenario: GitHub Copilot CLI implementation is isolated
- **WHEN** 所选工具为 GitHub Copilot CLI
- **THEN** 主 workflow 调用 `.github/workflows/monthly-openspec-archive-copilot.yml`
- **AND** GitHub Copilot CLI reusable workflow uses the shared deterministic archive action before any Copilot runtime setup
- **AND** GitHub Copilot CLI reusable workflow 独立完成 checkout、可选聚合、校验、变更范围保护、归档 PR 创建或更新和日志上传

#### Scenario: Claude Code implementation is isolated
- **WHEN** 所选工具为 Claude Code
- **THEN** 主 workflow 调用 `.github/workflows/monthly-openspec-archive-cc.yml`
- **AND** Claude Code reusable workflow uses the shared deterministic archive action before any Claude Code runtime setup
- **AND** Claude Code reusable workflow 独立完成 checkout、可选聚合、校验、变更范围保护、归档 PR 创建或更新和日志上传

#### Scenario: Only one tool workflow runs
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** `AI_CODING_TOOL` 为任一合法值
- **THEN** workflow 仅运行匹配该工具的 reusable workflow
- **AND** workflow 不运行其他工具的 reusable workflow

### Requirement: Monthly workflow must share prompt files across AI tools
系统 SHALL 将 monthly OpenSpec 归档聚合提示词维护为 `.github/prompts/` 下的公共文件，并 SHALL 让所有工具专属 reusable workflow 引用同一份聚合提示词内容。Base auto archive MUST use deterministic workflow logic instead of a shared AI prompt.

#### Scenario: Shared archive consolidate prompt
- **WHEN** 任一工具专属 reusable workflow 执行 `lina-archive-consolidate`
- **THEN** workflow 从 `.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md` 读取提示词
- **AND** workflow 不在工具专属 workflow 中内联维护重复的归档聚合提示词正文

#### Scenario: Deterministic auto archive does not use AI prompt
- **WHEN** 任一工具专属 reusable workflow 执行 base auto archive
- **THEN** workflow invokes the shared deterministic archive action
- **AND** workflow does not invoke Codex, Claude Code, or GitHub Copilot CLI for base auto archive

### Requirement: Monthly workflow must stream AI tool execution logs
系统 SHALL 在 monthly OpenSpec 归档聚合执行期间，将 AI Coding 工具进程的标准输出和标准错误实时写入 GitHub Actions step 日志，并继续保留 artifact 日志用于事后排查。

#### Scenario: Archive consolidate logs are visible
- **WHEN** 任一工具专属 reusable workflow 执行 `Run Lina Archive Consolidate`
- **THEN** `codex exec`、`claude -p` 或 `copilot -p` 进程的标准输出和标准错误在当前 GitHub Actions step 日志中可见
- **AND** workflow 仍保留对应 AI 工具日志 artifact
- **AND** 日志透传不得掩盖 AI 工具进程的失败退出码

#### Scenario: Deterministic archive logs are visible
- **WHEN** 任一工具专属 reusable workflow 执行 deterministic base auto archive
- **THEN** `openspec archive -y <change>` 的输出在当前 GitHub Actions step 日志中可见
- **AND** workflow records archived and failed candidates in the GitHub step summary

### Requirement: Monthly workflow must fail fast after each archive phase
系统 SHALL 在 monthly OpenSpec 自动归档和归档聚合阶段后执行确定性阶段检查；deterministic archive 留下无法归档的 completed active change 时，workflow 必须输出明确失败原因。AI consolidation 是可选增强阶段，失败或产生无效 OpenSpec 状态时 MUST NOT 阻塞已经通过校验的 deterministic archive PR。

#### Scenario: Deterministic archive leaves completed changes active
- **WHEN** deterministic auto archive finishes
- **AND** `openspec list --json` 仍报告 `complete`、`completed` 或 `done` 状态的活跃变更
- **THEN** workflow outputs the remaining change names, status values, and task counts
- **AND** workflow creates or updates an archive PR first when successful archive results produced OpenSpec changes
- **AND** workflow fails after PR finalization

#### Scenario: Auto archive produces invalid OpenSpec state
- **WHEN** deterministic auto archive produces OpenSpec file changes
- **AND** `openspec validate --all` 执行失败
- **THEN** workflow fails before archive consolidation
- **AND** workflow 不创建或更新归档 PR

#### Scenario: Archive consolidation produces invalid OpenSpec state
- **WHEN** 任一工具专属 reusable workflow 执行 `Run Lina Archive Consolidate`
- **AND** AI 工具进程返回成功
- **AND** `openspec validate --all` 执行失败
- **THEN** workflow captures the invalid consolidation diff and validation log as AI tool diagnostics
- **AND** workflow restores the previously validated deterministic archive result
- **AND** workflow validates the restored deterministic archive state
- **AND** workflow may create or update the archive PR using only the deterministic archive result

#### Scenario: Archive consolidation command fails
- **WHEN** 任一工具专属 reusable workflow 执行 `Run Lina Archive Consolidate`
- **AND** AI 工具进程返回失败
- **THEN** workflow captures the failed consolidation diff and validation log as AI tool diagnostics
- **AND** workflow restores the previously validated deterministic archive result
- **AND** workflow may create or update the archive PR using only the deterministic archive result

#### Scenario: Repository policy blocks pull request creation
- **WHEN** monthly OpenSpec 归档 workflow 已成功提交并推送归档来源分支
- **AND** GitHub repository policy prevents `GITHUB_TOKEN` from creating or updating pull requests
- **THEN** workflow outputs the pushed archive source branch and manual pull request URL
- **AND** workflow succeeds because the archive branch already contains the validated archive result
- **AND** workflow only fails for pull request command errors unrelated to repository pull request policy

### Requirement: Monthly workflow must inject AI tool credentials and endpoint at runtime
系统 SHALL 通过 GitHub Secret 和 GitHub Variables 在运行时生成所选 AI Coding 工具的认证配置并注入 provider `base_url`，并 SHALL NOT 将真实 API key/token 或真实 `base_url` 写入版本库中的 `.github/codex` 或 `.github/cc` 配置文件。

#### Scenario: Runtime credential setup
- **WHEN** monthly OpenSpec 归档工作流准备运行 Codex
- **THEN** workflow 从仓库内 Codex 配置模板 `.github/codex/config.template.toml` 复制临时 `config.toml`
- **AND** workflow 使用 `vars.OPENAI_BASE_URL` 替换临时 AI home 中 `config.toml` 的 `base_url` 占位符
- **AND** workflow 使用 `secrets.OPENAI_API_KEY` 在临时 AI home 中生成 `auth.json`
- **AND** workflow 在 Codex 容器内将该临时 AI home 映射为 `CODEX_HOME`
- **AND** 生成的认证文件和包含真实 `base_url` 的运行时配置不位于会被提交的仓库工作区路径中

#### Scenario: Runtime Claude Code setup
- **WHEN** monthly OpenSpec 归档工作流准备运行 Claude Code
- **THEN** workflow 从仓库内 Claude Code 配置模板读取 `settings.template.json`
- **AND** workflow 使用 `vars.ANTHROPIC_BASE_URL` 替换临时 `settings.json` 中的 `base_url` 占位符
- **AND** workflow 使用 `secrets.ANTHROPIC_AUTH_TOKEN` 替换临时 `settings.json` 中的认证 token 占位符
- **AND** workflow 使用 `vars.ANTHROPIC_CUSTOM_MODEL` 替换临时 `settings.json` 中的模型占位符
- **AND** 生成的认证文件和包含真实 `base_url` 的运行时配置不位于会被提交的仓库工作区路径中

#### Scenario: Runtime GitHub Copilot CLI setup
- **WHEN** monthly OpenSpec 归档工作流准备运行 GitHub Copilot CLI
- **THEN** workflow 使用 `secrets.COPILOT_GITHUB_TOKEN` 认证 GitHub Copilot CLI
- **AND** workflow 使用 `vars.COPILOT_MODEL` 配置 GitHub Copilot CLI 模型
- **AND** 当 `COPILOT_MODEL` 未配置或为 `auto` 时，workflow 不传递显式 `--model` 参数，由 GitHub Copilot CLI 默认模型路由处理
- **AND** workflow 使用 `vars.COPILOT_REASONING_EFFORT` 配置 GitHub Copilot CLI 推理等级
- **AND** 当 `COPILOT_REASONING_EFFORT` 未配置时，workflow 不传递显式 `--reasoning-effort` 参数，由 GitHub Copilot CLI 默认推理配置处理
- **AND** 当 `COPILOT_REASONING_EFFORT` 配置为合法非空值时，workflow 将其传递为 `--reasoning-effort=<value>`
- **AND** 当 `COPILOT_REASONING_EFFORT` 配置为不支持的值时，workflow 在执行 GitHub Copilot CLI 前失败
- **AND** workflow 使用隔离的临时 `COPILOT_HOME` 保存 GitHub Copilot CLI 运行状态和配置

#### Scenario: Missing OpenAI API key
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** 所选工具为 Codex
- **AND** `OPENAI_API_KEY` secret 为空或未配置
- **THEN** workflow 在执行 Codex 前失败
- **AND** workflow 不创建或更新归档 PR

#### Scenario: Missing OpenAI base URL
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** 所选工具为 Codex
- **AND** `OPENAI_BASE_URL` variable 为空或未配置
- **THEN** workflow 在执行 Codex 前失败
- **AND** workflow 不创建或更新归档 PR

#### Scenario: Missing Anthropic auth token
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** 所选工具为 Claude Code
- **AND** `ANTHROPIC_AUTH_TOKEN` secret 为空或未配置
- **THEN** workflow 在执行 Claude Code 前失败
- **AND** workflow 不创建或更新归档 PR

#### Scenario: Missing Anthropic base URL
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** 所选工具为 Claude Code
- **AND** `ANTHROPIC_BASE_URL` variable 为空或未配置
- **THEN** workflow 在执行 Claude Code 前失败
- **AND** workflow 不创建或更新归档 PR

#### Scenario: Missing Anthropic model
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** 所选工具为 Claude Code
- **AND** `ANTHROPIC_CUSTOM_MODEL` variable 为空或未配置
- **THEN** workflow 在执行 Claude Code 前失败
- **AND** workflow 不创建或更新归档 PR

#### Scenario: Missing GitHub Copilot token
- **WHEN** monthly OpenSpec 归档工作流触发
- **AND** 所选工具为 GitHub Copilot CLI
- **AND** `COPILOT_GITHUB_TOKEN` secret 为空或未配置
- **THEN** workflow 在执行 GitHub Copilot CLI 前失败
- **AND** workflow 不创建或更新归档 PR

### Requirement: Monthly workflow must guard generated changes
系统 SHALL 在创建或更新归档 PR 前验证本次变更范围，并仅允许 OpenSpec 归档治理文件被 monthly 自动任务修改。

#### Scenario: Allowed OpenSpec changes
- **WHEN** monthly OpenSpec 归档工作流完成归档和聚合
- **AND** 工作区变更仅包含 `openspec/**`
- **THEN** workflow 可以创建或更新归档 PR，且 PR 目标分支为仓库默认分支

#### Scenario: Unexpected file changes
- **WHEN** monthly OpenSpec 归档工作流完成归档和聚合
- **AND** 工作区存在允许范围外的文件变更
- **THEN** workflow 失败
- **AND** workflow 不创建或更新归档 PR

### Requirement: Monthly workflow must validate OpenSpec artifacts before creating a pull request
系统 SHALL 在创建或更新归档 PR 前执行 OpenSpec 校验，校验失败时必须停止 PR 写回。

#### Scenario: OpenSpec validation passes
- **WHEN** monthly OpenSpec 归档工作流产生待写回变更
- **AND** `openspec validate --all` 执行成功
- **THEN** workflow 使用固定维护分支创建或更新归档 PR

#### Scenario: OpenSpec validation fails
- **WHEN** monthly OpenSpec 归档工作流产生待写回变更
- **AND** `openspec validate --all` 执行失败
- **THEN** workflow 失败
- **AND** workflow 不创建或更新归档 PR

