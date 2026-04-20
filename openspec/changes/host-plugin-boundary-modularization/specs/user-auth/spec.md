## MODIFIED Requirements

### Requirement: 用户名密码登录
系统 SHALL 支持用户名 + 密码登录，验证成功后返回 JWT Token。登录过程（无论成功或失败）SHALL 发射统一登录生命周期事件。登录成功后 SHALL 在宿主维护的 `sys_online_session` 表中创建会话记录；若 `monitor-loginlog` 已启用，则该插件基于登录事件完成登录日志落库。

#### Scenario: 登录成功
- **WHEN** 用户提交正确的用户名和密码到 `POST /api/v1/auth/login`
- **THEN** 系统返回 JWT Token，响应格式为 `{code: 0, message: "ok", data: {token: "..."}}`
- **AND** 宿主在 `sys_online_session` 表中创建会话记录（包含 token_id、用户信息、IP、浏览器、操作系统等）
- **AND** 宿主发射登录成功事件；若 `monitor-loginlog` 已启用，则由该插件写入成功登录日志

#### Scenario: 登录失败且日志插件缺失
- **WHEN** 用户登录失败，且 `monitor-loginlog` 未安装、未启用或初始化失败
- **THEN** 系统仍返回正确的登录失败结果
- **AND** 宿主不因缺少具体登录日志持久化实现而报错

### Requirement: 用户登出
系统 SHALL 支持用户登出操作。登出操作 SHALL 发射统一登录生命周期事件，并删除宿主维护的在线会话记录。

#### Scenario: 登出成功
- **WHEN** 已登录用户调用 `POST /api/v1/auth/logout`
- **THEN** 系统返回成功响应，从 `sys_online_session` 表中删除该用户的会话记录，前端清除本地存储的 Token
- **AND** 宿主发射登出成功事件；若 `monitor-loginlog` 已启用，则由该插件写入对应日志

### Requirement: 认证中间件
系统 SHALL 提供认证中间件，保护需要登录才能访问的 API。中间件 MUST 同时校验 JWT 签名和宿主维护的会话有效性，不得依赖 `monitor-online` 是否已安装。

#### Scenario: 在线用户插件未安装时访问受保护接口
- **WHEN** 请求头携带有效的 `Authorization: Bearer <token>`，`sys_online_session` 中存在对应会话记录，且 `monitor-online` 未安装或未启用
- **THEN** 中间件仍通过 UPDATE 操作更新会话的 `last_active_time` 为当前时间
- **AND** 请求正常处理，用户信息注入到请求上下文中
