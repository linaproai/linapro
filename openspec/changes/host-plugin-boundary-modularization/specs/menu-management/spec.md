## ADDED Requirements

### Requirement: 默认后台使用稳定一级目录结构

系统 SHALL 为默认管理后台提供面向项目管理后台主场景的稳定一级目录结构。

#### Scenario: 查询默认后台菜单骨架
- **WHEN** 宿主为当前用户投影默认后台菜单
- **THEN** 一级目录按以下结构组织：`工作台`、`权限管理`、`组织管理`、`系统设置`、`内容管理`、`系统监控`、`任务调度`、`扩展中心`、`开发中心`
- **AND** 这些一级目录对应的稳定宿主父级 `menu_key` 精确为 `dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer`

#### Scenario: 一级目录作为宿主稳定目录记录存在
- **WHEN** 宿主初始化或同步默认后台菜单骨架
- **THEN** 这些一级目录由宿主创建并拥有
- **AND** 插件只能向这些目录下挂载子菜单，而不是创建新的同级一级目录

#### Scenario: 默认后台扩展业务模块
- **WHEN** 开发者在项目中继续叠加业务模块或官方源码插件
- **THEN** 新增菜单优先挂到已有稳定目录下
- **AND** 不需要频繁重构一级导航命名与结构

### Requirement: 插件菜单按语义挂载到宿主目录

系统 SHALL 要求插件菜单落在语义对应的宿主目录中，而不是统一归集到插件管理目录。

#### Scenario: 组织插件挂载菜单
- **WHEN** `org-center` 同步其菜单到宿主
- **THEN** 其菜单挂载到 `组织管理`
- **AND** 不挂载到 `扩展中心`

#### Scenario: 内容插件挂载菜单
- **WHEN** `content-notice` 同步其菜单到宿主
- **THEN** 其菜单挂载到 `内容管理`
- **AND** 不挂载到 `扩展中心`

#### Scenario: 监控插件挂载菜单
- **WHEN** `monitor-online`、`monitor-server`、`monitor-operlog` 或 `monitor-loginlog` 同步菜单到宿主
- **THEN** 这些菜单挂载到 `系统监控`
- **AND** 仍由 `扩展中心 / 插件管理` 负责安装与启停治理

### Requirement: 空父目录自动隐藏

系统 SHALL 在某一级目录无可见子菜单时自动隐藏该目录，避免默认后台出现空壳导航。

#### Scenario: 内容管理无可见菜单
- **WHEN** `content-notice` 未安装、未启用或当前用户无权访问其菜单
- **THEN** `内容管理` 不显示在左侧导航中

#### Scenario: 系统监控部分插件缺失
- **WHEN** `系统监控` 下只有部分监控插件已安装且可见
- **THEN** 左侧导航只显示可见的监控子菜单
- **AND** 父目录 `系统监控` 继续保留
- **AND** 若所有监控子菜单都不可见，则父目录一并隐藏
