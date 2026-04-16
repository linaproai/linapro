## ADDED Requirements

### Requirement: 会话存储抽象层

系统 SHALL 定义 `SessionStore` 抽象接口用于会话管理，当前使用 MySQL MEMORY 引擎实现。接口 MUST 支持以下操作：创建会话、查询会话、删除会话、列表查询（支持按用户名和 IP 过滤）。

#### Scenario: 登录时创建会话
- **WHEN** 用户通过 `POST /api/v1/auth/login` 成功登录
- **THEN** 系统在 `sys_online_session` 表中创建一条会话记录，包含 token_id（UUID）、user_id、username、dept_name、ip、browser、os、login_time 等字段

#### Scenario: 登出时删除会话
- **WHEN** 已登录用户调用 `POST /api/v1/auth/logout`
- **THEN** 系统从 `sys_online_session` 表中删除该用户对应的会话记录

#### Scenario: 请求校验会话有效性
- **WHEN** 用户携带有效 JWT Token 访问受保护 API
- **THEN** 中间件除校验 JWT 签名外，还 MUST 检查 `sys_online_session` 表中是否存在对应的会话记录；若不存在（已被强制下线），返回 401

### Requirement: 在线用户列表查询

系统 SHALL 提供管理员查询当前所有在线用户的 API，支持按用户名和 IP 地址过滤。

#### Scenario: 查询在线用户列表
- **WHEN** 管理员调用 `GET /api/v1/monitor/online/list`
- **THEN** 系统返回所有在线会话记录列表，每条记录包含：token_id、username、dept_name（部门名称）、ip（登录 IP）、login_location（登录地点）、browser（浏览器）、os（操作系统）、login_time（登录时间）

#### Scenario: 按用户名过滤
- **WHEN** 管理员调用 `GET /api/v1/monitor/online/list?username=admin`
- **THEN** 系统仅返回用户名包含 "admin" 的在线会话记录

#### Scenario: 按 IP 地址过滤
- **WHEN** 管理员调用 `GET /api/v1/monitor/online/list?ip=192.168`
- **THEN** 系统仅返回 IP 地址包含 "192.168" 的在线会话记录

### Requirement: 强制下线

系统 SHALL 支持管理员强制下线指定在线用户，被下线用户的后续请求 MUST 返回 401。

#### Scenario: 强制下线成功
- **WHEN** 管理员调用 `DELETE /api/v1/monitor/online/{tokenId}`
- **THEN** 系统删除该 tokenId 对应的会话记录，返回成功响应

#### Scenario: 被下线用户再次请求
- **WHEN** 已被强制下线的用户携带原 Token 访问任意受保护 API
- **THEN** 中间件检测到会话不存在，返回 401 状态码

#### Scenario: 下线不存在的 tokenId
- **WHEN** 管理员调用 `DELETE /api/v1/monitor/online/{tokenId}` 但该 tokenId 不存在
- **THEN** 系统返回成功响应（幂等操作）

### Requirement: 在线用户前端页面

系统 SHALL 提供在线用户管理页面，展示当前在线用户列表并支持强制下线操作。

#### Scenario: 页面展示在线用户列表
- **WHEN** 管理员访问在线用户页面
- **THEN** 页面展示 VXE-Grid 表格，包含以下列：登录账号、部门名称、IP 地址、登录地点、浏览器（带图标）、操作系统（带图标）、登录时间、操作（强制下线按钮）；工具栏显示在线人数统计

#### Scenario: 搜索过滤
- **WHEN** 管理员在搜索栏输入用户名或 IP 地址并搜索
- **THEN** 表格数据根据筛选条件刷新

#### Scenario: 强制下线交互
- **WHEN** 管理员点击某用户行的"强制下线"按钮
- **THEN** 弹出确认对话框，确认后调用强制下线 API，成功后刷新表格数据

### Requirement: 会话活跃时间跟踪

系统 SHALL 跟踪每个在线用户的最后活跃时间，用于判断会话是否超时。

#### Scenario: 登录时初始化活跃时间
- **WHEN** 用户成功登录，系统创建 `sys_online_session` 会话记录
- **THEN** `last_active_time` 字段 MUST 设置为当前时间

#### Scenario: 每次请求时更新活跃时间
- **WHEN** 已登录用户携带有效 Token 访问受保护 API
- **THEN** 认证中间件 MUST 通过 UPDATE 操作更新该会话的 `last_active_time` 为当前时间，并通过受影响行数判断会话是否存在（>0 存在，=0 不存在或已被清除）

### Requirement: 不活跃会话自动清理

系统 SHALL 提供定时任务自动清理长时间未操作的在线会话，防止会话表无限增长。超时阈值和清理频率 MUST 支持通过配置文件调整。

#### Scenario: 定时清理超时会话
- **WHEN** 定时清理任务执行时（默认每 5 分钟一次）
- **THEN** 系统 MUST 查询 `sys_online_session` 表中 `last_active_time` 距当前时间超过超时阈值（默认 24 小时）的记录，并将其删除

#### Scenario: 超时阈值可配置
- **WHEN** 管理员在 `config.yaml` 中设置 `session.timeoutHour` 配置项
- **THEN** 系统 MUST 使用该配置值作为会话超时阈值，不设置时默认为 24 小时

#### Scenario: 清理频率可配置
- **WHEN** 管理员在 `config.yaml` 中设置 `session.cleanupMinute` 配置项
- **THEN** 系统 MUST 使用该配置值作为清理任务执行间隔，不设置时默认为 5 分钟

### Requirement: 系统监控菜单

系统 SHALL 在导航菜单中新增"系统监控"一级菜单，包含"在线用户"和"服务监控"两个子菜单项。

#### Scenario: 菜单展示
- **WHEN** 管理员登录系统后查看左侧导航
- **THEN** 可见"系统监控"一级菜单，展开后包含"在线用户"和"服务监控"子菜单

#### Scenario: 菜单导航
- **WHEN** 管理员点击"在线用户"或"服务监控"菜单项
- **THEN** 页面跳转到对应的功能页面
