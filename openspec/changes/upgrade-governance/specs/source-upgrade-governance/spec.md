## ADDED Requirements

### Requirement: 框架元数据必须集中维护并可被系统信息页直接展示
框架 SHALL 在 `apps/lina-core/manifest/config/metadata.yaml` 中集中维护框架名称、版本号、介绍、项目官网、仓库地址和开源协议，并由系统信息接口直接返回给系统信息页顶部项目卡片展示，避免前后端各自硬编码不同的项目信息。

#### Scenario: 系统信息接口返回框架元数据
- **WHEN** 管理工作台请求系统信息接口
- **THEN** 返回结果中必须包含框架名称、版本号、介绍、项目官网、仓库地址和开源协议
- **AND** 这些字段的值必须来自宿主 `metadata.yaml`

### Requirement: 框架必须提供开发态的正式源码升级入口
框架 SHALL 提供独立于 `init` 和 `mock` 的正式源码升级入口，用于二开项目从旧框架版本升级到新框架版本；该入口 MUST 作为开发态工具收敛在仓库根目录 `hack/upgrade-framework/` 下，并由 `make upgrade` 调用，而不是作为 `lina-core` 运行时命令对外暴露。

#### Scenario: 通过升级命令执行框架升级
- **WHEN** 操作者执行正式升级命令
- **THEN** 系统必须按源码升级流程执行版本检查、代码覆盖和宿主 SQL 全量执行
- **AND** 不得要求操作者手工判断 SQL 执行位置
- **AND** 不得依赖 `lina-core` 运行时命令实现开发态升级逻辑

### Requirement: 源码升级前必须完成安全检查
框架 SHALL 在正式升级开始前提醒操作者做好备份，并检查 Git 工作区是否干净；若当前工作区存在未提交或未暂存的修改，命令 MUST 拒绝继续执行。

#### Scenario: 升级前工作区存在本地修改
- **WHEN** 操作者执行升级命令时，当前 Git 工作区存在已修改、未暂存或未提交文件
- **THEN** 命令必须拒绝继续升级
- **AND** 提示先提交或 stash 当前修改

### Requirement: 源码升级必须只读取 hack 配置中的升级元数据
框架 SHALL 以当前项目 `apps/lina-core/hack/config.yaml` 中记录的 `frameworkUpgrade.version` 作为当前项目的升级基线，并与目标标签代码中 `apps/lina-core/hack/config.yaml` 里的目标升级版本进行比较；上游仓库地址也必须优先来自同一文件中的 `frameworkUpgrade.repositoryUrl`。若目标版本不高于当前版本，则命令 MUST 安全退出且不得执行覆盖或 SQL；源码升级实现 MUST 不读取宿主运行时配置文件。

#### Scenario: 目标版本不高于当前版本
- **WHEN** 升级命令解析出的目标框架版本小于或等于当前项目 `hack/config.yaml` 中记录的升级版本
- **THEN** 命令必须提示当前项目已使用相同或更高版本的框架
- **AND** 不得覆盖代码，也不得执行 SQL

#### Scenario: 升级命令读取上游仓库地址
- **WHEN** 操作者未显式传入 `--repo`
- **THEN** 命令必须从当前项目 `apps/lina-core/hack/config.yaml` 的 `frameworkUpgrade.repositoryUrl` 读取默认上游仓库地址
- **AND** 不得回退读取宿主运行时配置文件

### Requirement: 源码升级必须从第一条宿主 SQL 开始全量执行
框架 SHALL 在完成目标标签代码覆盖后，从宿主 `manifest/sql/` 的第一条 SQL 文件开始按顺序执行全部宿主 SQL；该流程 MUST 不依赖数据库中的 SQL 游标或额外升级元数据表。

#### Scenario: 升级命令全量执行宿主 SQL
- **WHEN** 升级命令开始执行宿主 SQL
- **THEN** 必须从排序后的第一条宿主 SQL 文件开始执行
- **AND** 必须按文件顺序执行到最后一条宿主 SQL 文件

#### Scenario: 升级过程中某个 SQL 文件失败
- **WHEN** 升级命令执行某个宿主 SQL 文件失败
- **THEN** 命令必须立即停止后续 SQL 执行
- **AND** 必须返回失败的 SQL 文件和错误信息
