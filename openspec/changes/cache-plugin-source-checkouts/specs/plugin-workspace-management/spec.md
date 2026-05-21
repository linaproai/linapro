## MODIFIED Requirements

### Requirement: 插件安装必须从配置来源复制到固定目录

系统 SHALL 提供 `make plugins.install`，从 `hack/config.yaml` 声明的来源仓库读取插件目录，并复制到 `apps/lina-plugins/<plugin-id>`。安装命令不得把来源仓库作为 submodule 或嵌套 Git 仓库写入插件目录。插件来源仓库读取 SHALL 优先复用 `temp/plugin-sources/<source>` 下的工具缓存；缓存不存在时首次 clone，缓存存在且有效时通过 `git fetch --prune origin` 获取增量更新，并将缓存工作树重置到配置的 `ref` 后再复制插件目录。

#### Scenario: 安装配置中的缺失插件
- **WHEN** 用户运行 `make plugins.install`
- **AND** `hack/config.yaml` 声明插件 `multi-tenant`
- **AND** 远端 `<root>/multi-tenant/plugin.yaml` 存在
- **AND** 本地 `apps/lina-plugins/multi-tenant` 不存在
- **THEN** 命令将远端插件目录复制到 `apps/lina-plugins/multi-tenant`
- **AND** 目标目录不包含来源仓库的 `.git` 元数据
- **AND** 命令更新插件锁定状态文件

#### Scenario: 安装前自动初始化 submodule 工作区
- **WHEN** 用户运行 `make plugins.install`
- **AND** `apps/lina-plugins` 仍是 submodule
- **THEN** 命令自动执行等价于 `make plugins.init` 的工作区初始化
- **AND** 命令继续安装配置中的插件
- **AND** 插件工作区转换后不再保留 submodule Git metadata
- **AND** 用户不需要手动补跑 `make plugins.init`

#### Scenario: 目标插件目录已存在
- **WHEN** 用户运行 `make plugins.install`
- **AND** `apps/lina-plugins/<plugin-id>` 已存在
- **THEN** 命令默认失败并提示改用 `make plugins.update`
- **AND** 除非用户显式传入 `force=1`，否则不得覆盖已有目录

#### Scenario: 来源缓存存在时安装复用 fetch
- **WHEN** 用户再次运行 `make plugins.install`
- **AND** `temp/plugin-sources/<source>` 是有效 Git 仓库
- **AND** 缓存的 `origin` URL 与 `hack/config.yaml` 中该 source 的 `repo` 一致
- **THEN** 命令通过 `git fetch --prune origin` 更新该缓存
- **AND** 命令不得再创建新的 `temp/plugin-source-<source>-*` 一次性 clone 目录

### Requirement: 插件更新必须保护本地改动

系统 SHALL 提供 `make plugins.update`，按 `hack/config.yaml` 中的来源重新拉取插件目录并更新本地普通目录。更新命令在目标插件存在本地未提交改动或内容与锁定摘要不一致时，默认 SHALL 阻断覆盖。插件来源仓库读取 SHALL 复用 `temp/plugin-sources/<source>` 工具缓存，并通过 `git fetch --prune origin` 更新缓存后解析配置的 `ref`。

#### Scenario: 更新无本地改动的插件
- **WHEN** 用户运行 `make plugins.update`
- **AND** `apps/lina-plugins/<plugin-id>` 存在
- **AND** 该插件目录没有本地未提交改动
- **AND** 远端 source ref 解析到新的 commit 或内容摘要
- **THEN** 命令用远端插件目录更新本地插件目录
- **AND** 命令更新插件锁定状态文件中的 commit、版本和内容摘要

#### Scenario: 更新遇到本地 dirty 插件
- **WHEN** 用户运行 `make plugins.update`
- **AND** `apps/lina-plugins/<plugin-id>` 存在本地未提交改动
- **THEN** 命令默认失败
- **AND** 错误列出 dirty 插件 ID
- **AND** 命令不覆盖本地插件内容

#### Scenario: 强制更新 dirty 插件
- **WHEN** 用户运行 `make plugins.update force=1`
- **AND** 目标插件存在本地未提交改动
- **THEN** 命令允许覆盖该插件目录
- **AND** 命令输出被强制覆盖的插件 ID
- **AND** 命令更新插件锁定状态文件

### Requirement: 插件状态检查必须提供只读诊断

系统 SHALL 提供 `make plugins.status`，只读检查插件工作区状态、配置插件状态、本地改动、插件版本、锁定状态和远端更新状态。除自动初始化缺失或历史 submodule 插件工作区外，该命令不得修改 `apps/lina-plugins`、`.gitmodules`、父仓库 Git index 或锁定状态文件。该命令 MAY 更新 `temp/plugin-sources/<source>` 下的工具缓存以获取远端状态，但 MUST NOT 创建命令结束即删除的 `temp/plugin-source-<source>-*` 一次性 clone 目录。

#### Scenario: 状态检查普通插件工作区
- **WHEN** 用户运行 `make plugins.status`
- **AND** `apps/lina-plugins` 是普通目录
- **THEN** 命令输出工作区类型为普通目录
- **AND** 命令列出配置中每个插件的本地存在性、source、ref、插件版本和本地 dirty 状态

#### Scenario: 状态检查自动初始化 submodule 工作区
- **WHEN** 用户运行 `make plugins.status`
- **AND** `apps/lina-plugins` 仍是 submodule
- **THEN** 命令自动执行等价于 `make plugins.init` 的工作区初始化
- **AND** 命令继续输出普通插件工作区状态
- **AND** 命令不写入插件目录或锁定状态文件

#### Scenario: 远端不可达
- **WHEN** 用户运行 `make plugins.status`
- **AND** 某个 source 远端仓库不可访问
- **THEN** 命令仍输出本地状态
- **AND** 该 source 下插件的远端更新状态标记为 unknown
- **AND** 命令返回可用于诊断的远端访问错误

#### Scenario: 本地存在未配置插件
- **WHEN** 用户运行 `make plugins.status`
- **AND** `apps/lina-plugins` 下存在未在 `hack/config.yaml` 声明的插件目录
- **THEN** 命令列出这些未配置插件
- **AND** 命令不删除或修改这些目录

#### Scenario: 来源缓存损坏或来源变更
- **WHEN** 用户运行 `make plugins.status`
- **AND** `temp/plugin-sources/<source>` 不是有效 Git 仓库，或缓存 `origin` URL 与配置 `repo` 不一致
- **THEN** 命令删除并重建该 source 的工具缓存
- **AND** 删除范围仅限 `temp/plugin-sources/<source>`
- **AND** 命令仍不得修改 `apps/lina-plugins`、`.gitmodules`、父仓库 Git index 或锁定状态文件
