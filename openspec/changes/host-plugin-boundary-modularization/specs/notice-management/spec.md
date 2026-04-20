## MODIFIED Requirements

### Requirement: 通知公告菜单与权限
系统 SHALL 将通知公告菜单作为 `content-notice` 源码插件菜单挂载到宿主 `内容管理` 目录，而不是挂到 `系统管理`。

#### Scenario: 菜单显示
- **WHEN** `content-notice` 已安装、已启用且当前用户拥有菜单访问权限
- **THEN** `内容管理` 分组下显示 `通知公告` 菜单项
- **AND** 插件治理仍由 `扩展中心 / 插件管理` 负责

#### Scenario: 插件缺失或停用
- **WHEN** `content-notice` 未安装、未启用或当前用户无权访问其菜单
- **THEN** 宿主不显示 `通知公告` 菜单入口
- **AND** 若 `内容管理` 无其他可见子菜单，则父目录一并隐藏

## ADDED Requirements

### Requirement: 通知公告由内容源码插件交付

系统 SHALL 将通知公告能力作为 `content-notice` 源码插件交付，而不是继续作为宿主默认内建模块。

#### Scenario: 内容插件启用时提供通知公告能力
- **WHEN** `content-notice` 已安装并启用
- **THEN** 宿主暴露通知公告相关 API、页面与菜单
- **AND** 该插件继续承载公告内容管理与发布流程

#### Scenario: 内容插件缺失时隐藏通知公告入口
- **WHEN** `content-notice` 未安装或未启用
- **THEN** 宿主不显示通知公告菜单和页面入口
- **AND** 宿主其余核心能力继续正常运行
