## ADDED Requirements

### Requirement: Nightly workflow must archive completed OpenSpec changes
系统 SHALL 提供一个 GitHub Actions nightly 工作流，用于在默认分支上自动扫描 `openspec/changes/` 中已完成且未归档的活跃变更，并通过可配置的 AI Coding 工具执行 `lina-auto-archive` 技能完成归档。

#### Scenario: Scheduled archive run
- **WHEN** nightly OpenSpec 归档工作流按计划触发
- **THEN** workflow 在仓库默认分支 checkout 代码
- **AND** workflow 使用 `AI_CODING_TOOL` 指定的 AI Coding 工具运行归档任务
- **AND** workflow 调用 `lina-auto-archive` 扫描并归档可自动处理的已完成变更

#### Scenario: No completed active changes
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** `openspec list --json` 未报告任何 `complete`、`completed` 或 `done` 状态的活跃变更
- **THEN** workflow 不调用 AI Coding 工具归档任务
- **AND** workflow 成功结束且不创建提交

### Requirement: Nightly workflow must consolidate only after new archive changes
系统 SHALL 仅在本次 nightly 自动归档产生 OpenSpec 文件变更后执行 `lina-archive-consolidate` 技能，避免无新增归档时重复重写聚合归档文档。

#### Scenario: Archive produced changes
- **WHEN** `lina-auto-archive` 执行后 `openspec/` 下存在新的文件变更
- **THEN** workflow 调用 `lina-archive-consolidate` 聚合已归档变更
- **AND** workflow 在聚合后继续执行 OpenSpec 校验

#### Scenario: Archive produced no changes
- **WHEN** `lina-auto-archive` 执行完成
- **AND** `openspec/` 下没有新的文件变更
- **THEN** workflow 跳过 `lina-archive-consolidate`
- **AND** workflow 不创建提交

### Requirement: Nightly workflow must select the AI Coding tool from GitHub Variables
系统 SHALL 通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择 nightly OpenSpec 归档使用的 AI Coding 工具，并 SHALL 在未配置该变量时默认使用 `codex`。

#### Scenario: Default Codex tool
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** GitHub Variables 未配置 `AI_CODING_TOOL`
- **THEN** 主 workflow 调用 Codex reusable workflow
- **AND** Codex reusable workflow 使用 `loads/codex:latest` 和 `codex exec` 运行 AI 任务

#### Scenario: Explicit Codex tool
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `codex`
- **THEN** 主 workflow 调用 Codex reusable workflow
- **AND** Codex reusable workflow 使用 `loads/codex:latest` 和 `codex exec` 运行 AI 任务

#### Scenario: Explicit Claude Code tool
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 为 `cc`
- **THEN** 主 workflow 调用 Claude Code reusable workflow
- **AND** Claude Code reusable workflow 使用 `loads/cc:latest` 和 `claude -p` 运行 AI 任务

#### Scenario: Unsupported tool value
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** GitHub Variables 中 `AI_CODING_TOOL` 不是 `codex` 或 `cc`
- **THEN** 主 workflow 在执行任何工具 reusable workflow 前失败
- **AND** workflow 不提交任何变更

### Requirement: Nightly workflow must isolate tool implementations in reusable workflows
系统 SHALL 将不同 AI Coding 工具的运行时准备、镜像调用、认证配置和日志上传细节封装在工具专属 reusable workflow 中，并 SHALL 让主 workflow 只负责触发、候选检测和路由。

#### Scenario: Codex implementation is isolated
- **WHEN** 所选工具为 Codex
- **THEN** 主 workflow 调用 `.github/workflows/nightly-openspec-archive-codex.yml`
- **AND** Codex reusable workflow 独立完成 checkout、归档、聚合、校验、变更范围保护、提交和日志上传

#### Scenario: Claude Code implementation is isolated
- **WHEN** 所选工具为 Claude Code
- **THEN** 主 workflow 调用 `.github/workflows/nightly-openspec-archive-cc.yml`
- **AND** Claude Code reusable workflow 独立完成 checkout、归档、聚合、校验、变更范围保护、提交和日志上传

#### Scenario: Only one tool workflow runs
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** `AI_CODING_TOOL` 为任一合法值
- **THEN** workflow 仅运行匹配该工具的 reusable workflow
- **AND** workflow 不运行其他工具的 reusable workflow

### Requirement: Nightly workflow must share prompt files across AI tools
系统 SHALL 将 nightly OpenSpec 自动归档和归档聚合提示词维护为 `.github/prompts/` 下的公共文件，并 SHALL 让所有工具专属 reusable workflow 引用同一份提示词内容。

#### Scenario: Shared auto archive prompt
- **WHEN** 任一工具专属 reusable workflow 执行 `lina-auto-archive`
- **THEN** workflow 从 `.github/prompts/nightly-openspec-auto-archive.zh-CN.md` 读取提示词
- **AND** workflow 不在工具专属 workflow 中内联维护重复的自动归档提示词正文

#### Scenario: Shared archive consolidate prompt
- **WHEN** 任一工具专属 reusable workflow 执行 `lina-archive-consolidate`
- **THEN** workflow 从 `.github/prompts/nightly-openspec-archive-consolidate.zh-CN.md` 读取提示词
- **AND** workflow 不在工具专属 workflow 中内联维护重复的归档聚合提示词正文

### Requirement: Nightly workflow must inject AI tool credentials and endpoint at runtime
系统 SHALL 通过 GitHub Secret 在运行时生成所选 AI Coding 工具的认证配置并注入 provider `base_url`，并 SHALL NOT 将真实 API key/token 或真实 `base_url` 写入版本库中的 `.github/codex` 或 `.github/cc` 配置文件。

#### Scenario: Runtime credential setup
- **WHEN** nightly OpenSpec 归档工作流准备运行 Codex
- **THEN** workflow 从仓库内 Codex 配置模板 `.github/codex/config.template.toml` 复制临时 `config.toml`
- **AND** workflow 使用 `secrets.OPENAI_BASE_URL` 替换临时 AI home 中 `config.toml` 的 `base_url` 占位符
- **AND** workflow 使用 `secrets.OPENAI_API_KEY` 在临时 AI home 中生成 `auth.json`
- **AND** workflow 在 Codex 容器内将该临时 AI home 映射为 `CODEX_HOME`
- **AND** 生成的认证文件和包含真实 `base_url` 的运行时配置不位于会被提交的仓库工作区路径中

#### Scenario: Runtime Claude Code setup
- **WHEN** nightly OpenSpec 归档工作流准备运行 Claude Code
- **THEN** workflow 从仓库内 Claude Code 配置模板读取 `settings.template.json`
- **AND** workflow 使用 `secrets.ANTHROPIC_BASE_URL` 替换临时 `settings.json` 中的 `base_url` 占位符
- **AND** workflow 使用 `secrets.ANTHROPIC_AUTH_TOKEN` 替换临时 `settings.json` 中的认证 token 占位符
- **AND** workflow 使用 `vars.ANTHROPIC_CUSTOM_MODEL` 或同名 secret 替换临时 `settings.json` 中的模型占位符
- **AND** 生成的认证文件和包含真实 `base_url` 的运行时配置不位于会被提交的仓库工作区路径中

#### Scenario: Missing OpenAI API key
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** 所选工具为 Codex
- **AND** `OPENAI_API_KEY` secret 为空或未配置
- **THEN** workflow 在执行 Codex 前失败
- **AND** workflow 不提交任何变更

#### Scenario: Missing OpenAI base URL
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** 所选工具为 Codex
- **AND** `OPENAI_BASE_URL` secret 为空或未配置
- **THEN** workflow 在执行 Codex 前失败
- **AND** workflow 不提交任何变更

#### Scenario: Missing Anthropic auth token
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** 所选工具为 Claude Code
- **AND** `ANTHROPIC_AUTH_TOKEN` secret 为空或未配置
- **THEN** workflow 在执行 Claude Code 前失败
- **AND** workflow 不提交任何变更

#### Scenario: Missing Anthropic base URL
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** 所选工具为 Claude Code
- **AND** `ANTHROPIC_BASE_URL` secret 为空或未配置
- **THEN** workflow 在执行 Claude Code 前失败
- **AND** workflow 不提交任何变更

#### Scenario: Missing Anthropic model
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** 所选工具为 Claude Code
- **AND** `ANTHROPIC_CUSTOM_MODEL` variable 和同名 secret 均为空或未配置
- **THEN** workflow 在执行 Claude Code 前失败
- **AND** workflow 不提交任何变更

### Requirement: Nightly workflow must guard generated changes
系统 SHALL 在自动提交前验证本次变更范围，并仅允许 OpenSpec 归档治理文件被 nightly 自动任务修改。

#### Scenario: Allowed OpenSpec changes
- **WHEN** nightly OpenSpec 归档工作流完成归档和聚合
- **AND** 工作区变更仅包含 `openspec/**`
- **THEN** workflow 可以提交变更到默认分支

#### Scenario: Unexpected file changes
- **WHEN** nightly OpenSpec 归档工作流完成归档和聚合
- **AND** 工作区存在允许范围外的文件变更
- **THEN** workflow 失败
- **AND** workflow 不提交任何变更

### Requirement: Nightly workflow must validate OpenSpec artifacts before commit
系统 SHALL 在提交自动归档结果前执行 OpenSpec 校验，校验失败时必须停止提交。

#### Scenario: OpenSpec validation passes
- **WHEN** nightly OpenSpec 归档工作流产生待提交变更
- **AND** `openspec validate --all` 执行成功
- **THEN** workflow 提交自动归档和聚合结果

#### Scenario: OpenSpec validation fails
- **WHEN** nightly OpenSpec 归档工作流产生待提交变更
- **AND** `openspec validate --all` 执行失败
- **THEN** workflow 失败
- **AND** workflow 不提交自动归档结果
