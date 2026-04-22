## MODIFIED Requirements

### Requirement: 操作日志自动记录
系统 SHALL 通过 `monitor-operlog` 源码插件在宿主统一 HTTP 注册入口上声明的全局审计中间件，对所有写操作（POST/PUT/DELETE）以及标记了 `operLog` 标签的查询操作自动发射统一审计事件。宿主只提供受治理的全局中间件注册接缝与统一事件分发，不保留固定的操作日志业务中间件。`monitor-operlog` 已安装并启用时，插件中间件参与请求链并将日志持久化到 `plugin_monitor_operlog` 表；插件不可用时，宿主核心请求链路必须旁路该采集逻辑并继续正常执行。

#### Scenario: 操作日志插件已启用
- **WHEN** 用户发起受审计的请求且 `monitor-operlog` 已安装并启用
- **THEN** `monitor-operlog` 通过宿主封装的全局 HTTP 中间件注册器包裹匹配请求
- **AND** 宿主发射统一审计事件
- **AND** `monitor-operlog` 写入一条对应的操作日志记录

#### Scenario: 操作日志插件缺失或停用
- **WHEN** 用户发起受审计的请求但 `monitor-operlog` 未安装、未启用或初始化失败
- **THEN** 宿主旁路插件自注册的审计中间件逻辑
- **AND** 宿主仍然正常完成原始业务请求
- **AND** 宿主不因缺少具体操作日志落库实现而返回错误

#### Scenario: 下游中间件提前结束请求
- **WHEN** `monitor-operlog` 的全局审计中间件已经包裹一个请求，且后续中间件或处理器写出响应后提前结束当前请求
- **THEN** 审计中间件在 `Next` 返回后仍可读取当前响应快照并发射匹配的审计事件
- **AND** 提前结束请求不会导致本次操作日志漏记

### Requirement: 操作日志类型使用语义字符串常量

系统 SHALL 使用具备业务语义的字符串常量表达操作日志类型，而不是在宿主、插件、接口和存储层传播 `1~6` 这类位置敏感整数编码。

#### Scenario: 审计事件落库时写入语义类型
- **WHEN** 宿主发射一次操作日志审计事件
- **THEN** `monitor-operlog` 使用强类型常量写入 `oper_type`
- **AND** `oper_type` 的持久化值为 `create`、`update`、`delete`、`export`、`import`、`other` 之一

#### Scenario: 操作日志接口返回语义类型
- **WHEN** 管理员查询或导出操作日志
- **THEN** 接口中的 `operType` 字段返回语义字符串值
- **AND** 前端继续通过 `sys_oper_type` 字典渲染对应中文标签

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
