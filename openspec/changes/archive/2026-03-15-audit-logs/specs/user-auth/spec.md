## MODIFIED Requirements

### Requirement: 用户名密码登录
系统 SHALL 支持用户名 + 密码登录，验证成功后返回 JWT Token。登录过程（无论成功或失败）SHALL 自动记录登录日志。

#### Scenario: 登录成功
- **WHEN** 用户提交正确的用户名和密码到 `POST /api/v1/auth/login`
- **THEN** 系统返回 JWT Token，响应格式为 `{code: 0, message: "ok", data: {token: "..."}}`，并写入一条状态为成功的登录日志

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
系统 SHALL 支持用户登出操作。登出操作 SHALL 自动记录登录日志。

#### Scenario: 登出成功
- **WHEN** 已登录用户调用 `POST /api/v1/auth/logout`
- **THEN** 系统返回成功响应，前端清除本地存储的 Token，并写入一条状态为成功的登录日志（msg 为"登出成功"）
