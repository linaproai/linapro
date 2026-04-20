## MODIFIED Requirements

### Requirement: 操作日志自动记录
系统 SHALL 在宿主审计链路中对所有写操作（POST/PUT/DELETE）以及标记了 `operLog` 标签的查询操作自动发射统一审计事件。`monitor-operlog` 已安装并启用时，该插件订阅事件并将日志持久化到 `sys_oper_log` 表；插件不可用时，宿主核心请求链路仍正常执行。

#### Scenario: 操作日志插件已启用
- **WHEN** 用户发起受审计的请求且 `monitor-operlog` 已安装并启用
- **THEN** 宿主发射统一审计事件
- **AND** `monitor-operlog` 订阅该事件后写入一条对应的操作日志记录

#### Scenario: 操作日志插件缺失或停用
- **WHEN** 用户发起受审计的请求但 `monitor-operlog` 未安装、未启用或初始化失败
- **THEN** 宿主仍然正常完成原始业务请求
- **AND** 宿主不因缺少具体操作日志落库实现而返回错误

## ADDED Requirements

### Requirement: 操作日志治理界面由源码插件交付

系统 SHALL 将操作日志查询、详情、导出、清理与页面能力作为 `monitor-operlog` 源码插件交付。

#### Scenario: 插件启用时暴露治理入口
- **WHEN** `monitor-operlog` 已安装并启用
- **THEN** 宿主暴露操作日志查询、详情、导出、清理接口以及前端页面
- **AND** 插件菜单挂载到宿主 `系统监控` 目录，顶层 `parent_key` 为 `monitor`

#### Scenario: 插件缺失时隐藏治理入口
- **WHEN** `monitor-operlog` 未安装或未启用
- **THEN** 宿主不显示操作日志菜单和页面入口
- **AND** 普通业务请求链路继续正常运行
