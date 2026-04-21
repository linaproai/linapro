## MODIFIED Requirements

### Requirement: 登录日志自动记录
系统 SHALL 在登录成功、登录失败和登出成功等认证生命周期节点自动发射统一登录事件。`monitor-loginlog` 已安装并启用时，该插件订阅事件并将日志持久化到 `plugin_monitor_loginlog` 表；插件不可用时，宿主认证链路仍正常执行。

#### Scenario: 登录日志插件已启用
- **WHEN** 用户发生登录成功、登录失败或登出成功事件，且 `monitor-loginlog` 已安装并启用
- **THEN** 宿主发射统一登录事件
- **AND** `monitor-loginlog` 订阅该事件后写入对应的登录日志记录

#### Scenario: 登录日志插件缺失或停用
- **WHEN** 用户发生登录成功、登录失败或登出成功事件，但 `monitor-loginlog` 未安装、未启用或初始化失败
- **THEN** 宿主仍然正常返回认证结果
- **AND** 宿主不因缺少具体登录日志落库实现而返回错误

## ADDED Requirements

### Requirement: 登录日志治理界面由源码插件交付

系统 SHALL 将登录日志查询、详情、导出、清理与页面能力作为 `monitor-loginlog` 源码插件交付。

#### Scenario: 插件启用时暴露治理入口
- **WHEN** `monitor-loginlog` 已安装并启用
- **THEN** 宿主暴露登录日志查询、详情、导出、清理接口以及前端页面
- **AND** 插件菜单挂载到宿主 `系统监控` 目录，顶层 `parent_key` 为 `monitor`

#### Scenario: 插件缺失时隐藏治理入口
- **WHEN** `monitor-loginlog` 未安装或未启用
- **THEN** 宿主不显示登录日志菜单和页面入口
- **AND** 登录与登出流程继续正常运行
