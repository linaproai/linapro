## ADDED Requirements

### Requirement: Release tag 必须与框架元数据版本一致

系统 SHALL 将 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version` 作为 release tag 的唯一版本基线。任何 release tag 发布链路在执行测试、构建、镜像推送或更新浮动标签前，MUST 校验 Git tag 名称与 `framework.version` 完全一致。

#### Scenario: Tag 名称与框架版本一致时继续发布
- **WHEN** GitHub Actions 收到 release tag push 事件
- **AND** tag 名称等于 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version`
- **THEN** release workflow 继续执行后续测试和镜像发布门禁

#### Scenario: Tag 名称与框架版本不一致时阻止发布
- **WHEN** GitHub Actions 收到 release tag push 事件
- **AND** tag 名称不等于 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version`
- **THEN** release workflow 在测试和镜像发布前失败
- **AND** workflow 不得执行 GHCR 登录、`make image push=1`、`latest` 标签发布或远端 manifest inspect

#### Scenario: 框架版本格式不合法时阻止发布
- **WHEN** release workflow 校验 `framework.version`
- **AND** `framework.version` 不是 Docker tag 兼容的 release 版本格式
- **THEN** release workflow 失败
- **AND** 错误消息说明 `framework.version`、Git tag 名称和允许的版本格式

### Requirement: Release tag 校验必须通过跨平台工具复用

系统 SHALL 通过仓库跨平台工具入口执行 release tag 与框架元数据版本一致性校验。GitHub Actions、本地发布检查和后续发布自动化 SHALL 复用同一个校验命令，避免在不同 workflow 或脚本中维护重复的 YAML 解析和版本格式规则。

#### Scenario: 本地校验命令读取框架版本
- **WHEN** 维护者运行 release tag 校验命令并传入 tag 名称
- **THEN** 命令读取 `apps/lina-core/manifest/config/metadata.yaml`
- **AND** 命令比较传入 tag 与 `framework.version`
- **AND** 命令输出可读的成功或失败信息

#### Scenario: Workflow 复用同一个校验命令
- **WHEN** release workflow 需要校验 tag 名称
- **THEN** workflow 调用仓库跨平台校验命令
- **AND** workflow 不得使用独立的 ad hoc YAML 解析逻辑替代该命令

### Requirement: 受控发布入口必须在创建 tag 前校验框架版本

系统 SHALL 提供受控的 GitHub Actions 手动发布入口，用于读取 `framework.version` 并创建同名 release tag。该入口 MUST 在创建 tag 前运行 release tag 校验；校验失败时不得创建或推送 tag。

#### Scenario: 受控入口成功创建匹配 tag
- **WHEN** 维护者手动触发受控 release tag workflow
- **AND** `framework.version` 合法且对应远端 tag 尚不存在
- **THEN** workflow 创建名称等于 `framework.version` 的 Git tag
- **AND** workflow 将该 tag 推送到 GitHub 仓库

#### Scenario: 受控入口发现 tag 已存在
- **WHEN** 维护者手动触发受控 release tag workflow
- **AND** 远端已存在名称等于 `framework.version` 的 Git tag
- **THEN** workflow 失败
- **AND** workflow 不得移动、覆盖或删除既有 tag

#### Scenario: 仓库规则阻止人工直接创建 release tag
- **WHEN** 仓库配置了匹配 release tag 的 GitHub tag ruleset
- **THEN** 普通用户不得直接创建、更新或删除受保护 release tag
- **AND** 只有受控发布 actor 可以在校验通过后创建 release tag
- **AND** 受控发布 workflow 必须使用仓库变量 `RELEASE_APP_CLIENT_ID` 和仓库密钥 `RELEASE_APP_PRIVATE_KEY` 生成 GitHub App installation token
- **AND** ruleset bypass 必须配置到该 GitHub App actor，而不是配置到 token 字符串本身
