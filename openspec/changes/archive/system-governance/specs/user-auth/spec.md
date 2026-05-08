## MODIFIED Requirements

### Requirement: 用户名密码登录
系统 SHALL 支持用户名 + 密码登录，验证成功后返回 JWT Token。登录过程（无论成功或失败）SHALL 自动记录登录日志。登录成功后 SHALL 在 `sys_online_session` 表中创建会话记录。

#### Scenario: 登录成功
- **WHEN** 用户提交正确的用户名和密码到 `POST /api/v1/auth/login`
- **THEN** 系统返回 JWT Token，响应格式为 `{code: 0, message: "ok", data: {token: "..."}}`，写入一条状态为成功的登录日志，并在 `sys_online_session` 表中创建会话记录（包含 token_id、用户信息、IP、浏览器、操作系统等）

#### Scenario: 用户名不存在
- **WHEN** 用户提交不存在的用户名
- **THEN** 系统返回错误信息，提示用户名或密码错误，并写入一条状态为失败的登录日志

#### Scenario: 密码错误
- **WHEN** 用户提交错误的密码
- **THEN** 系统返回错误信息，提示用户名或密码错误（不区分用户名还是密码错误），并写入一条状态为失败的登录日志

#### Scenario: 用户已停用
- **WHEN** 状态为停用的用户尝试登录
- **THEN** 系统返回错误信息，提示用户已停用，并写入一条状态为失败的登录日志

### Requirement: 用户登出
系统 SHALL 支持用户登出操作。登出操作 SHALL 自动记录登录日志，并删除对应的在线会话记录。

#### Scenario: 登出成功
- **WHEN** 已登录用户调用 `POST /api/v1/auth/logout`
- **THEN** 系统返回成功响应，从 `sys_online_session` 表中删除该用户的会话记录，前端清除本地存储的 Token，并写入一条状态为成功的登录日志（msg 为"登出成功"）

### Requirement: 认证中间件
系统 SHALL 提供认证中间件，保护需要登录才能访问的 API。中间件 MUST 同时校验 JWT 签名和会话有效性。

#### Scenario: 携带有效 Token 且会话存在
- **WHEN** 请求头携带有效的 `Authorization: Bearer <token>` 且 `sys_online_session` 表中存在对应会话记录
- **THEN** 中间件通过 UPDATE 操作更新会话的 `last_active_time` 为当前时间，通过受影响行数（>0）判断会话存在，请求正常处理，用户信息注入到请求上下文中

#### Scenario: 携带有效 Token 但会话不存在（已被强制下线或超时清理）
- **WHEN** 请求头携带有效的 JWT Token，但 `sys_online_session` 表中不存在对应会话记录（已被强制下线或因不活跃超时被自动清理）
- **THEN** UPDATE 操作受影响行数为 0，系统返回 401 状态码

#### Scenario: 未携带 Token 访问受保护接口
- **WHEN** 请求未携带 Authorization 头访问受保护 API
- **THEN** 系统返回 401 状态码

#### Scenario: 携带无效 Token 访问受保护接口
- **WHEN** 请求携带无效或过期的 Token
- **THEN** 系统返回 401 状态码
