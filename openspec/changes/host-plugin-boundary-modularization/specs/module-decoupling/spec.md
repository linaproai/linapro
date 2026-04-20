## ADDED Requirements

### Requirement: 非核心管理模块以源码插件形式交付

系统 SHALL 将组织管理、内容管理和系统监控中的非核心模块作为源码插件交付，使开发者能够按需安装和启用。

#### Scenario: 规划组织与内容模块
- **WHEN** 宿主交付默认后台能力
- **THEN** 部门管理与岗位管理由 `org-management` 源码插件提供
- **AND** 通知公告由 `content-notice` 源码插件提供

#### Scenario: 规划系统监控模块
- **WHEN** 宿主交付系统监控相关能力
- **THEN** 在线用户、服务监控、操作日志和登录日志分别由独立源码插件提供
- **AND** 它们的插件 ID 分别为 `monitor-online`、`monitor-server`、`monitor-operlog`、`monitor-loginlog`

### Requirement: 监控插件必须支持独立安装与启停

系统 SHALL 将在线用户、服务监控、操作日志和登录日志视为 4 个相互独立的源码插件，而不是单一监控插件套件。

#### Scenario: 仅安装部分监控插件
- **WHEN** 管理员只安装或启用部分监控插件
- **THEN** 宿主只展示这些已安装且已启用插件对应的监控菜单
- **AND** 未安装的监控插件不会阻塞其他监控插件运行

#### Scenario: 停用单个监控插件
- **WHEN** 管理员停用 `monitor-online`、`monitor-server`、`monitor-operlog` 或 `monitor-loginlog` 中的任意一个
- **THEN** 宿主仅隐藏该插件对应的功能入口
- **AND** 其他监控插件与宿主核心链路继续正常运行

### Requirement: 插件缺失时宿主必须平滑降级

系统 SHALL 在源码插件缺失、未安装或未启用时保证宿主主体功能继续可用。

#### Scenario: 组织插件缺失时访问用户管理
- **WHEN** `org-management` 未安装或未启用
- **THEN** 用户管理页面与接口仍然正常工作
- **AND** 与部门、岗位相关的字段、筛选项、树选择器与表单项被安全隐藏或忽略

#### Scenario: 日志插件缺失时宿主继续处理请求
- **WHEN** `monitor-operlog` 或 `monitor-loginlog` 未安装或未启用
- **THEN** 宿主核心请求链路仍然正常执行
- **AND** 与对应日志持久化相关的能力进入受控降级流程
- **AND** 不因日志插件缺失导致认证、鉴权或普通业务请求失败

### Requirement: 在线用户插件不得承载认证主链路

系统 SHALL 保证 `monitor-online` 只承载在线用户治理能力，而不会接管宿主认证主链路。

#### Scenario: 在线用户插件缺失
- **WHEN** `monitor-online` 未安装或未启用
- **THEN** 宿主仍然正常执行登录、登出、受保护接口鉴权和会话超时清理
- **AND** 宿主继续使用自身会话真相源维护 `last_active_time` 与超时判定

#### Scenario: 在线用户插件执行强制下线
- **WHEN** `monitor-online` 已安装并执行强制下线治理
- **THEN** 插件通过宿主提供的会话治理能力失效指定会话
- **AND** 插件不拥有 JWT 校验、会话触达刷新或超时清理真相源

### Requirement: 日志插件通过宿主事件承接非核心日志落库

系统 SHALL 将登录日志与操作日志的落库责任解耦为“宿主发射事件 + 插件按需订阅持久化”。

#### Scenario: 登录日志插件已启用
- **WHEN** 用户发生登录成功、登录失败或登出成功事件
- **THEN** 宿主先发射统一登录事件
- **AND** `monitor-loginlog` 订阅该事件后完成落库与后续查询治理

#### Scenario: 操作日志插件已启用
- **WHEN** 用户触发写操作或带 `operLog` 标签的受审计查询
- **THEN** 宿主先发射统一审计事件
- **AND** `monitor-operlog` 订阅该事件后完成落库与后续查询治理

#### Scenario: 日志插件未启用
- **WHEN** `monitor-loginlog` 或 `monitor-operlog` 未安装、未启用或初始化失败
- **THEN** 宿主继续处理原始认证或请求流程
- **AND** 宿主不因缺少具体日志持久化实现而返回错误
