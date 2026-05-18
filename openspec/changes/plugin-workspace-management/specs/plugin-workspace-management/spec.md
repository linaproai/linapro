## ADDED Requirements

### Requirement: 插件工作区必须支持从 submodule 转为普通目录

系统 SHALL 提供跨平台开发工具命令，将固定目录 `apps/lina-plugins` 从 Git submodule 工作区转换为普通目录，并保留该目录下已有插件代码。转换命令不得删除插件源码内容；若 `.gitmodules` 中存在其他 submodule，命令只移除 `apps/lina-plugins` 对应配置。

#### Scenario: 转换已初始化的插件 submodule
- **WHEN** 用户运行 `make plugins.init`
- **AND** `apps/lina-plugins` 是已初始化的 submodule
- **THEN** 命令移除父仓库中 `apps/lina-plugins` 的 submodule/gitlink 配置
- **AND** 命令移除 `.gitmodules` 中 `apps/lina-plugins` 对应 section
- **AND** 命令清理 `.git/config` 和 `.git/modules/apps/lina-plugins` 中对应 submodule 元数据
- **AND** `apps/lina-plugins` 目录下已有插件文件保留为普通工作区文件

#### Scenario: .gitmodules 中存在其他 submodule
- **WHEN** 用户运行 `make plugins.init`
- **AND** `.gitmodules` 同时包含 `apps/lina-plugins` 和其他 submodule 配置
- **THEN** 命令只移除 `apps/lina-plugins` 对应 section
- **AND** `.gitmodules` 文件继续保留其他 submodule 配置

#### Scenario: 工作区已是普通目录
- **WHEN** 用户运行 `make plugins.init`
- **AND** `apps/lina-plugins` 已经是普通目录且不是 submodule
- **THEN** 命令不删除目录内容
- **AND** 命令输出普通目录状态
- **AND** 命令以成功状态结束

### Requirement: 插件来源必须通过 hack/config.yaml 声明

系统 SHALL 通过 `hack/config.yaml` 的 `plugins.sources` 声明插件来源。每个 source SHALL 包含 `repo`、`root`、`ref` 和字符串数组 `items`。`apps/lina-plugins` 是固定插件维护目录，不得在配置中重复声明工作区路径。

#### Scenario: 读取官方和自定义插件来源
- **WHEN** `hack/config.yaml` 包含 `plugins.sources.official` 和 `plugins.sources.custom`
- **AND** 每个 source 均配置 `repo`、`root`、`ref` 和 `items`
- **THEN** 插件工具按 source 解析每个 `items` 中的插件 ID
- **AND** 每个插件的远端路径为 `<repo>@<ref>:<root>/<plugin-id>`
- **AND** 本地目标路径为 `apps/lina-plugins/<plugin-id>`

#### Scenario: items 只接受字符串数组
- **WHEN** `plugins.sources.<name>.items` 包含对象、数字或其他非字符串值
- **THEN** 插件工具在执行任何写入前失败
- **AND** 错误说明 `items` 只支持插件 ID 字符串数组

#### Scenario: items 使用通配符安装来源下全部插件
- **WHEN** `plugins.sources.<name>.items` 配置为字符串数组且包含 `"*"`
- **THEN** 插件工具扫描该 source 的 `<root>` 目录下一层子目录
- **AND** 仅将包含 `plugin.yaml` 的子目录展开为插件 ID
- **AND** 展开后的插件 ID 仍必须符合插件 ID 命名规则
- **AND** 裸 YAML alias 写法 `- *` 不属于支持格式，用户必须写作带引号的 `"*"`

#### Scenario: 通配符不能与显式插件 ID 混用
- **WHEN** 同一个 source 的 `items` 同时包含 `"*"` 和具体插件 ID
- **THEN** 插件工具在执行任何写入前失败
- **AND** 错误说明 wildcard source 不能混用显式插件 ID

#### Scenario: 同名插件跨来源重复声明
- **WHEN** 多个 source 的 `items` 中声明同一个插件 ID
- **THEN** 插件工具在执行任何写入前失败
- **AND** 错误列出冲突插件 ID 和对应 source 名称

#### Scenario: root 使用不安全路径
- **WHEN** source 的 `root` 是绝对路径、包含 `..`、为空或包含平台 drive path
- **THEN** 插件工具拒绝该配置
- **AND** 不访问远端仓库或写入本地插件目录

### Requirement: 插件安装必须从配置来源复制到固定目录

系统 SHALL 提供 `make plugins.install`，从 `hack/config.yaml` 声明的来源仓库读取插件目录，并复制到 `apps/lina-plugins/<plugin-id>`。安装命令不得把来源仓库作为 submodule 或嵌套 Git 仓库写入插件目录。

#### Scenario: 安装配置中的缺失插件
- **WHEN** 用户运行 `make plugins.install`
- **AND** `hack/config.yaml` 声明插件 `multi-tenant`
- **AND** 远端 `<root>/multi-tenant/plugin.yaml` 存在
- **AND** 本地 `apps/lina-plugins/multi-tenant` 不存在
- **THEN** 命令将远端插件目录复制到 `apps/lina-plugins/multi-tenant`
- **AND** 目标目录不包含来源仓库的 `.git` 元数据
- **AND** 命令更新插件锁定状态文件

#### Scenario: 安装前仍是 submodule
- **WHEN** 用户运行 `make plugins.install`
- **AND** `apps/lina-plugins` 仍是 submodule
- **THEN** 命令失败并提示先运行 `make plugins.init`
- **AND** 命令不写入任何插件目录

#### Scenario: 目标插件目录已存在
- **WHEN** 用户运行 `make plugins.install`
- **AND** `apps/lina-plugins/<plugin-id>` 已存在
- **THEN** 命令默认失败并提示改用 `make plugins.update`
- **AND** 除非用户显式传入 `force=1`，否则不得覆盖已有目录

### Requirement: 插件更新必须保护本地改动

系统 SHALL 提供 `make plugins.update`，按 `hack/config.yaml` 中的来源重新拉取插件目录并更新本地普通目录。更新命令在目标插件存在本地未提交改动或内容与锁定摘要不一致时，默认 SHALL 阻断覆盖。

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

系统 SHALL 提供 `make plugins.status`，只读检查插件工作区状态、配置插件状态、本地改动、插件版本、锁定状态和远端更新状态。该命令不得修改 `apps/lina-plugins`、`.gitmodules`、父仓库 Git index 或锁定状态文件。

#### Scenario: 状态检查普通插件工作区
- **WHEN** 用户运行 `make plugins.status`
- **AND** `apps/lina-plugins` 是普通目录
- **THEN** 命令输出工作区类型为普通目录
- **AND** 命令列出配置中每个插件的本地存在性、source、ref、插件版本和本地 dirty 状态

#### Scenario: 状态检查发现仍是 submodule
- **WHEN** 用户运行 `make plugins.status`
- **AND** `apps/lina-plugins` 仍是 submodule
- **THEN** 命令输出 submodule 状态
- **AND** 命令提示先运行 `make plugins.init`
- **AND** 命令不修改 submodule 配置

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

### Requirement: 插件锁定状态必须由工具维护

系统 SHALL 在 `apps/lina-plugins/.linapro-plugins.lock.yaml` 记录插件安装和更新状态。锁定状态文件由工具生成和更新，用于诊断和更新判断；用户维护的插件来源配置仍以 `hack/config.yaml` 为准。

#### Scenario: 安装后写入锁定状态
- **WHEN** `make plugins.install` 成功安装插件
- **THEN** 锁定状态记录插件 ID、source、repo、root、ref、resolved commit、插件清单版本和内容摘要

#### Scenario: 配置与锁定状态不一致
- **WHEN** 用户运行 `make plugins.status`
- **AND** 锁定状态中存在但 `hack/config.yaml` 已不再声明的插件
- **THEN** 命令输出 orphaned lock entry 诊断
- **AND** 命令不自动删除锁定状态或本地插件目录
