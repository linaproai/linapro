## Requirements

### Requirement: 默认后台使用稳定的一级目录结构

系统 SHALL 为默认管理后台提供面向项目管理后台主场景的稳定一级目录结构。

#### Scenario: 查询默认后台菜单骨架

- **WHEN** 宿主为当前用户投射默认后台菜单时
- **THEN** 一级目录按以下结构组织：`工作台`、`权限管理`、`组织管理`、`系统设置`、`内容管理`、`系统监控`、`任务调度`、`扩展中心`、`开发中心`
- **AND** 这些一级目录对应的宿主稳定父级 `menu_key` 分别为 `dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer`
- **AND** MUST NOT 将 `storage`（存储管理）作为一级宿主稳定目录

#### Scenario: 一级目录作为宿主稳定目录记录存在

- **WHEN** 宿主初始化或同步默认后台菜单骨架时
- **THEN** 这些一级目录由宿主创建和拥有
- **AND** 插件只能将子菜单挂载到这些目录下，而非创建新的同级一级目录

#### Scenario: 默认后台扩展业务模块

- **WHEN** 开发者持续向项目添加业务模块或官方源码插件时
- **THEN** 新菜单会优先放置在现有稳定目录中
- **AND** 无需频繁重构一级导航命名和结构

### Requirement: 插件菜单语义化挂载到宿主目录

系统 SHALL 要求插件菜单落入语义对应的宿主目录，而非统一归入插件管理目录。

#### Scenario: 组织插件挂载菜单

- **WHEN** `linapro-org-core` 将其菜单同步到宿主时
- **THEN** 其菜单挂载到 `组织管理`
- **AND** 不挂载到 `扩展中心`

#### Scenario: 内容插件挂载菜单

- **WHEN** `linapro-content-notice` 将其菜单同步到宿主时
- **THEN** 其菜单挂载到 `内容管理`
- **AND** 不挂载到 `扩展中心`

#### Scenario: 监控插件挂载菜单

- **WHEN** `linapro-monitor-online`、`linapro-monitor-server`、`linapro-monitor-operlog` 或 `linapro-monitor-loginlog` 将菜单同步到宿主时
- **THEN** 这些菜单挂载到 `系统监控`
- **AND** `扩展中心/插件管理` 仍负责安装和启停管理

#### Scenario: 云存储 provider 插件挂载菜单

- **WHEN** `linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure` 或 `linapro-storage-s3` 将菜单同步到宿主时
- **THEN** 其配置管理菜单 MUST 挂载到 `系统设置`（`parent_key: setting`）
- **AND** 不挂载到独立的 `存储管理` 一级目录
- **AND** 不挂载到 `扩展中心` 作为配置入口父目录
- **AND** `扩展中心/插件管理` 仍负责安装和启停管理

### Requirement: 空父目录自动隐藏

系统 SHALL 在某级目录下没有可见子菜单时自动隐藏该目录，避免默认后台出现空壳导航。

#### Scenario: 内容管理无可见菜单

- **WHEN** `linapro-content-notice` 未安装、未启用或当前用户无权访问其菜单时
- **THEN** `内容管理` 不出现在左侧导航中

#### Scenario: 部分系统监控插件缺失

- **WHEN** 仅安装了部分监控插件且 `系统监控` 下有可见菜单时
- **THEN** 左侧导航仅显示可见的监控子菜单
- **AND** 父目录 `系统监控` 继续保留
- **AND** 如果所有监控子菜单都不可见，父目录也将被隐藏

#### Scenario: 系统设置下无云存储配置菜单

- **WHEN** 所有云存储 provider 插件未安装、未启用或当前用户无权访问其配置菜单时
- **THEN** 「系统设置」下不出现这些云存储配置子菜单
- **AND** 「系统设置」因仍有宿主内建设置菜单（如字典、参数、文件管理）而可继续可见
