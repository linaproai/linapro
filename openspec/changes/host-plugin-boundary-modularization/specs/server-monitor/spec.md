## ADDED Requirements

### Requirement: 服务监控由独立源码插件交付

系统 SHALL 将服务监控能力作为 `monitor-server` 源码插件交付，而不是继续作为宿主默认内建模块。

#### Scenario: 服务监控插件启用时提供能力
- **WHEN** `monitor-server` 已安装并启用
- **THEN** 宿主暴露服务监控采集、清理、查询和页面能力
- **AND** 插件菜单挂载到宿主 `系统监控` 目录，顶层 `parent_key` 为 `monitor`

#### Scenario: 服务监控插件缺失时平滑降级
- **WHEN** `monitor-server` 未安装或未启用
- **THEN** 宿主不显示服务监控菜单和页面入口
- **AND** 其他监控插件与宿主核心能力继续正常运行
