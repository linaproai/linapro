# framework-upgrade Specification

## Purpose
TBD - created by archiving change make-upgrade-framework. Update Purpose after archive.
## Requirements
### Requirement: 框架升级命令必须提供统一跨平台入口

系统 SHALL 提供 `linactl upgrade` 作为框架源码升级的唯一实现入口，并通过根 `Makefile` 的 `make upgrade` 与 Windows `make.cmd upgrade` 薄包装暴露。业务逻辑 MUST 实现在 `linactl`（Go）中，不得依赖平台专属 shell 脚本承载合并逻辑。

#### Scenario: Make 入口转发到 linactl

- **WHEN** 开发者在仓库根目录运行 `make upgrade`
- **THEN** 根 Makefile 转发到 `linactl upgrade`
- **AND** 可选参数 `v`、`force` 被透传到 `linactl`

#### Scenario: 直接调用 linactl

- **WHEN** 开发者运行 `linactl upgrade`
- **THEN** 命令执行与 `make upgrade` 相同的升级语义

### Requirement: 框架升级不得自动更新插件

系统在执行 `upgrade` 合并官方框架时 MUST 保留本地 `apps/lina-plugins` 的升级前状态（包括 submodule gitlink 或普通目录树内容）。系统 MUST NOT 调用 `plugins.update`、`plugins.install` 或其它插件安装/更新流程。官方仓库中对 `apps/lina-plugins` 的变更 MUST NOT 进入升级合并结果。

#### Scenario: 合并后插件工作区保持本地版本

- **WHEN** 本地 `apps/lina-plugins` 与官方目标 ref 中的插件内容或 submodule 指针不同
- **AND** 开发者运行 `make upgrade` 或 `linactl upgrade`
- **THEN** 宿主框架代码按目标 ref 合并到当前分支
- **AND** `apps/lina-plugins` 保持升级前的本地状态
- **AND** 命令输出说明插件不会自动更新

#### Scenario: 不触发插件更新命令

- **WHEN** 开发者运行 `upgrade`
- **THEN** 系统不执行 `plugins.update` 或 `plugins.install`

### Requirement: 升级源必须固定为官方仓库

系统 SHALL 始终从硬编码的官方仓库 `https://github.com/linaproai/linapro.git` 拉取框架代码。系统 MUST 使用工具托管 remote 名 `linapro`（不存在时自动创建，URL 不正确时自动纠正）。系统 MUST NOT 使用本地 `origin`、fork remote 或用户自定义 `remote=` 参数作为升级源。

#### Scenario: 默认从官方仓库拉取

- **WHEN** 开发者运行 `make upgrade` 或 `linactl upgrade`
- **THEN** 系统确保 remote `linapro` 指向 `https://github.com/linaproai/linapro.git`
- **AND** 从该 remote fetch tags/branches
- **AND** 不依赖本地 `origin` 的 URL

#### Scenario: 拒绝 remote 参数

- **WHEN** 开发者运行 `make upgrade remote=upstream` 或 `linactl upgrade remote=origin`
- **THEN** 命令失败
- **AND** 错误信息说明升级源固定为官方仓库

#### Scenario: origin 为 fork 时仍使用官方源

- **WHEN** 本地 `origin` 指向 fork 或其它非官方仓库
- **AND** 开发者运行默认 `upgrade`
- **THEN** 系统仍从 `linapro`（官方 URL）解析并合并目标

### Requirement: 默认合并最新稳定版本到当前分支

当未指定 `v` 参数时，系统 SHALL 从官方 remote 获取标签，选择最新稳定版本标签，并将其合并到当前本地分支。稳定版本 MUST 匹配 `vMAJOR.MINOR.PATCH`（可接受无前缀 `v` 的输入，但内部规范化为带 `v` 前缀的 tag）；预发布标签（含 `-` 后缀）MUST NOT 作为默认“最新稳定版”候选。

#### Scenario: 默认升级到最新稳定 tag

- **WHEN** 开发者在干净工作区的已命名分支上运行 `linactl upgrade` 或 `make upgrade`
- **AND** 官方仓库上存在多个稳定 tag（例如 `v0.4.0`、`v0.5.0`）以及预发布 tag（例如 `v0.6.0-rc.1`）
- **THEN** 系统 fetch 官方 remote 的 tags
- **AND** 选择最新稳定 tag `v0.5.0`
- **AND** 将该 tag merge 到当前分支
- **AND** 不选择 `v0.6.0-rc.1`

#### Scenario: 官方仓库无稳定 tag 时失败

- **WHEN** 开发者运行默认 `upgrade` 且官方仓库上不存在任何稳定版本 tag
- **THEN** 命令失败
- **AND** 错误信息说明未找到稳定版本

### Requirement: 支持通过 v 参数指定版本或分支

系统 SHALL 接受可选参数 `v`：

- 当 `v` 为稳定版本号（有或无 `v` 前缀）时，合并对应版本 tag 到当前分支
- 当 `v=main` 或其它非稳定版本字符串时，合并 `linapro/<v>` 分支到当前分支

#### Scenario: 指定稳定版本

- **WHEN** 开发者运行 `make upgrade v=v0.5.0` 或 `make upgrade v=0.5.0`
- **THEN** 系统将官方仓库的 tag `v0.5.0` merge 到当前本地分支

#### Scenario: 指定 main 分支

- **WHEN** 开发者运行 `make upgrade v=main`
- **THEN** 系统 fetch 官方 remote
- **AND** 将 `linapro/main` merge 到当前本地分支

### Requirement: 升级前必须执行安全门禁

系统 SHALL 在执行 merge 前校验：

1. 当前不处于 detached HEAD
2. 工作区干净（`git status --porcelain` 无输出），或用户在交互确认中明确同意在脏工作区继续，或显式传入 `force=1` / `force=true`

当工作区不干净且未传 `force` 时，系统 MUST：

1. 向标准输出提示工作区不干净；
2. 请求用户确认是否继续（提示中说明输入 `y` 继续）；
3. 仅当用户输入 `y` / `yes`（大小写不敏感，忽略首尾空白）时继续后续 fetch/merge；
4. 其它输入、空输入或读取失败时 MUST 以非零退出码结束，且 MUST NOT 执行 merge。

系统 MUST NOT 在门禁失败时执行 merge。`force` MUST NOT 跳过 detached HEAD 检查。`force=1` 时 MUST 跳过脏工作区检查与确认提示。

#### Scenario: 脏工作区用户确认 y 后继续

- **WHEN** 工作区存在未提交变更且未传 `force`
- **AND** 用户在确认提示中输入 `y` 或 `yes`
- **THEN** 系统继续后续 fetch/merge 流程

#### Scenario: 脏工作区用户拒绝或未确认

- **WHEN** 工作区存在未提交变更且未传 `force`
- **AND** 用户输入非 `y`/`yes` 的内容、空行，或 stdin 无可用输入（如非交互管道）
- **THEN** 命令失败且不执行 merge

#### Scenario: force 允许脏工作区继续

- **WHEN** 工作区存在未提交变更且传入 `force=1`
- **AND** 当前位于已命名分支
- **THEN** 系统跳过脏工作区检查与确认提示并继续 fetch/merge

#### Scenario: detached HEAD 拒绝

- **WHEN** 当前处于 detached HEAD
- **THEN** 命令失败且不执行 merge
- **AND** 即使传入 `force=1` 也不得继续

### Requirement: 合并失败时保持可恢复

当 `git merge` 因冲突或其它 Git 错误失败时，系统 SHALL 以非零退出码结束，并向前端输出/错误流暴露失败原因。系统 MUST NOT 在失败后自动执行 `git reset --hard` 或丢弃用户本地变更。

#### Scenario: merge 冲突

- **WHEN** 目标 ref 与当前分支产生冲突导致 merge 失败
- **THEN** 命令返回错误
- **AND** 不自动 hard reset 工作区

### Requirement: 升级时仅下载所选目标 ref

系统在执行 `upgrade` 的对象下载阶段 MUST 仅 fetch 本次已解析的目标 ref 可达对象，MUST NOT 使用会下载官方 remote 全部 tags 的命令作为默认路径（例如 `git fetch <remote> --tags`）。

解析目标时：

1. 未指定 `v`：MUST 通过远端 tag 列表发现（例如 `git ls-remote --tags --refs`，不下载对象）并选出最新稳定 tag，再仅 fetch 该 tag；
2. `v` 为稳定版本：MUST 仅 fetch 对应 tag；
3. `v` 为分支名：MUST 仅 fetch 该分支到 `linapro/<branch>` 对应的 remote-tracking ref。

#### Scenario: 默认升级不拉取全部 tags

- **WHEN** 开发者运行默认 `make upgrade` 或 `linactl upgrade` 且官方存在多个稳定 tag
- **THEN** 系统选择最新稳定 tag 并只下载该 tag 所需对象
- **AND** 系统不执行会同步官方全部 tags 的 fetch

#### Scenario: 指定版本只拉取该 tag

- **WHEN** 开发者运行 `make upgrade v=v0.5.0`
- **THEN** 系统只 fetch tag `v0.5.0`（及其合并所需对象）
- **AND** 不 fetch 其它无关 release tags

#### Scenario: 指定分支只拉取该分支

- **WHEN** 开发者运行 `make upgrade v=main`
- **THEN** 系统只更新 `linapro/main` 所需对象
- **AND** 不执行全量 tags fetch

