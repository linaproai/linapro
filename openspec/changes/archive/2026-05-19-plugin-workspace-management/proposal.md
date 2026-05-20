## Why

当前 `apps/lina-plugins` 在官方仓库中作为 submodule 挂载，适合 LinaPro 官方插件独立维护，但不适合用户项目直接管理插件源码。用户项目需要用普通目录承载插件代码，并通过仓库配置声明插件来源、安装列表和更新策略，避免 submodule 指针、官方远端和本地提交边界带来的复杂度。

## What Changes

- 新增插件工作区管理命令，支持将 `apps/lina-plugins` 从 submodule 转换为普通目录并保留现有插件代码。
- 在 `hack/config.yaml` 新增 `plugins.sources` 配置，按来源仓库声明 `repo`、`root`、`ref` 和字符串数组 `items`。
- 新增插件安装命令，从官方或自定义来源仓库拉取配置中的插件到 `apps/lina-plugins/<plugin-id>`。
- 新增插件更新命令，按配置来源更新普通目录中的插件代码，并在本地存在未提交改动时默认阻断覆盖。
- 新增插件状态命令，检查工作区是否仍为 submodule、配置插件是否存在、插件版本、本地改动、安装来源和远端更新状态。
- 生成工具维护的插件锁定状态文件，用于记录插件安装来源、解析后的 commit 和内容摘要；`hack/config.yaml` 仍是唯一需要用户手写维护的配置入口。

## Capabilities

### New Capabilities

- `plugin-workspace-management`: 定义用户项目插件工作区去 submodule 化、配置化插件来源、安装、更新、状态检查和锁定状态治理。

### Modified Capabilities

- `project-setup`: 开发工具配置需要支持通过 `hack/config.yaml` 统一声明插件来源和待安装插件列表，并通过跨平台 `linactl` 命令执行。

## Impact

- 影响 `hack/config.yaml` 配置结构和 `hack/tools/linactl` 命令集合。
- 影响根 `Makefile`、`make.cmd` 或拆分 makefile 的插件命令包装入口。
- 影响 `.gitmodules`、父仓库 Git index、`.git/config` 和 `.git/modules/apps/lina-plugins` 的 submodule 元数据治理。
- 影响 `apps/lina-plugins` 目录的插件源码管理方式；用户项目中该目录将作为普通目录由主仓库直接提交。
- 不新增运行时 REST API、数据库表、后端服务缓存或业务数据权限边界。
- 不改变插件运行时目录规范、插件清单结构、插件生命周期安装/启用/卸载语义。
