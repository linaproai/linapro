## MODIFIED Requirements

### Requirement:默认后台使用稳定的一级目录结构

系统 SHALL 为默认管理后台提供面向项目管理后台主场景的稳定一级目录结构。

#### Scenario:查询默认后台菜单骨架
- **当** 宿主为当前用户投射默认后台菜单时
- **则** 一级目录按以下结构组织：`工作台`、`权限管理`、`组织管理`、`系统设置`、`内容管理`、`系统监控`、`任务调度`、`存储管理`、`扩展中心`、`开发中心`
- **且** 这些一级目录对应的宿主稳定父级 `menu_key` 分别为 `dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`storage`、`extension`、`developer`

#### Scenario:一级目录作为宿主稳定目录记录存在
- **当** 宿主初始化或同步默认后台菜单骨架时
- **则** 这些一级目录由宿主创建和拥有
- **且** 插件只能将子菜单挂载到这些目录下，而非创建新的同级一级目录

#### Scenario:默认后台扩展业务模块
- **当** 开发者持续向项目添加业务模块或官方源码插件时
- **则** 新菜单会优先放置在现有稳定目录中
- **且** 无需频繁重构一级导航命名和结构

### Requirement:插件菜单语义化挂载到宿主目录

系统 SHALL 要求插件菜单落入语义对应的宿主目录，而非统一归入插件管理目录。

#### Scenario:组织插件挂载菜单
- **当** `linapro-org-core` 将其菜单同步到宿主时
- **则** 其菜单挂载到 `组织管理`
- **且** 不挂载到 `扩展中心`

#### Scenario:内容插件挂载菜单
- **当** `linapro-content-notice` 将其菜单同步到宿主时
- **则** 其菜单挂载到 `内容管理`
- **且** 不挂载到 `扩展中心`

#### Scenario:监控插件挂载菜单
- **当** `linapro-monitor-online`、`linapro-monitor-server`、`linapro-monitor-operlog` 或 `linapro-monitor-loginlog` 将菜单同步到宿主时
- **则** 这些菜单挂载到 `系统监控`
- **且** `扩展中心/插件管理` 仍负责安装和启停管理

#### Scenario:云存储 provider 插件挂载菜单
- **当** `linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure` 或 `linapro-storage-s3` 将菜单同步到宿主时
- **则** 其配置管理菜单 MUST 挂载到 `存储管理`（`parent_key: storage`）
- **且** 不挂载到 `扩展中心` 作为配置入口父目录
- **且** `扩展中心/插件管理` 仍负责安装和启停管理

### Requirement:空父目录自动隐藏

系统 SHALL 在某级目录下没有可见子菜单时自动隐藏该目录，避免默认后台出现空壳导航。

#### Scenario:内容管理无可见菜单
- **当** `linapro-content-notice` 未安装、未启用或当前用户无权访问其菜单时
- **则** `内容管理` 不出现在左侧导航中

#### Scenario:部分系统监控插件缺失
- **当** 仅安装了部分监控插件且 `系统监控` 下有可见菜单时
- **则** 左侧导航仅显示可见的监控子菜单
- **且** 父目录 `系统监控` 继续保留
- **且** 如果所有监控子菜单都不可见，父目录也将被隐藏

#### Scenario:存储管理无可见菜单
- **当** 所有云存储 provider 插件未安装、未启用或当前用户无权访问其配置菜单时
- **则** `存储管理` 不出现在左侧导航中
- **且** 宿主不删除 `menu_key = storage` 的稳定目录记录
