## ADDED Requirements

### Requirement: Nightly workflow must archive completed OpenSpec changes
系统 SHALL 提供一个 GitHub Actions nightly 工作流，用于在默认分支上自动扫描 `openspec/changes/` 中已完成且未归档的活跃变更，并通过 Codex 执行 `lina-auto-archive` 技能完成归档。

#### Scenario: Scheduled archive run
- **WHEN** nightly OpenSpec 归档工作流按计划触发
- **THEN** workflow 在仓库默认分支 checkout 代码
- **AND** workflow 使用 `loads/codex:latest` 运行 Codex
- **AND** workflow 调用 `lina-auto-archive` 扫描并归档可自动处理的已完成变更

#### Scenario: No completed active changes
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** `openspec list --json` 未报告任何 `complete`、`completed` 或 `done` 状态的活跃变更
- **THEN** workflow 不调用 Codex 归档任务
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

### Requirement: Nightly workflow must inject Codex credentials and endpoint at runtime
系统 SHALL 通过 GitHub Secret 在运行时生成 Codex 认证文件并注入 provider `base_url`，并 SHALL NOT 将真实 `OPENAI_API_KEY` 或真实 `base_url` 写入版本库中的 `.github/codex` 配置文件。

#### Scenario: Runtime credential setup
- **WHEN** nightly OpenSpec 归档工作流准备运行 Codex
- **THEN** workflow 从仓库内 Codex 配置模板复制 `config.toml`
- **AND** workflow 使用 `secrets.OPENAI_BASE_URL` 替换临时 `CODEX_HOME/config.toml` 中的 `base_url` 占位符
- **AND** workflow 使用 `secrets.OPENAI_API_KEY` 在临时 `CODEX_HOME` 中生成 `auth.json`
- **AND** 生成的认证文件和包含真实 `base_url` 的运行时配置不位于会被提交的仓库工作区路径中

#### Scenario: Missing OpenAI API key
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** `OPENAI_API_KEY` secret 为空或未配置
- **THEN** workflow 在执行 Codex 前失败
- **AND** workflow 不提交任何变更

#### Scenario: Missing OpenAI base URL
- **WHEN** nightly OpenSpec 归档工作流触发
- **AND** `OPENAI_BASE_URL` secret 为空或未配置
- **THEN** workflow 在执行 Codex 前失败
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
