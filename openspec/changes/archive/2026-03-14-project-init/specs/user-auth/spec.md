## ADDED Requirements

### Requirement: 用户名密码登录
系统 SHALL 支持用户名 + 密码登录，验证成功后返回 JWT Token。

#### Scenario: 登录成功
- **WHEN** 用户提交正确的用户名和密码到 `POST /api/v1/auth/login`
- **THEN** 系统返回 JWT Token，响应格式为 `{code: 0, message: "ok", data: {token: "..."}}`

#### Scenario: 用户名不存在
- **WHEN** 用户提交不存在的用户名
- **THEN** 系统返回错误信息，提示用户名或密码错误

#### Scenario: 密码错误
- **WHEN** 用户提交错误的密码
- **THEN** 系统返回错误信息，提示用户名或密码错误（不区分用户名还是密码错误）

#### Scenario: 用户已停用
- **WHEN** 状态为停用的用户尝试登录
- **THEN** 系统返回错误信息，提示用户已停用

### Requirement: 用户登出
系统 SHALL 支持用户登出操作。

#### Scenario: 登出成功
- **WHEN** 已登录用户调用 `POST /api/v1/auth/logout`
- **THEN** 系统返回成功响应，前端清除本地存储的 Token

### Requirement: JWT Token 管理
系统 SHALL 使用 JWT（HS256）进行身份认证，Token 包含用户基本信息。

#### Scenario: Token 载荷内容
- **WHEN** 系统签发 JWT Token
- **THEN** Token 载荷包含 `userId`、`username`、`status` 字段

#### Scenario: Token 有效期
- **WHEN** Token 签发后超过配置的有效期（默认 24 小时）
- **THEN** Token 失效，使用该 Token 的请求返回 401

### Requirement: 认证中间件
系统 SHALL 提供认证中间件，保护需要登录才能访问的 API。

#### Scenario: 携带有效 Token 访问受保护接口
- **WHEN** 请求头携带有效的 `Authorization: Bearer <token>` 访问受保护 API
- **THEN** 请求正常处理，用户信息注入到请求上下文中

#### Scenario: 未携带 Token 访问受保护接口
- **WHEN** 请求未携带 Authorization 头访问受保护 API
- **THEN** 系统返回 401 状态码

#### Scenario: 携带无效 Token 访问受保护接口
- **WHEN** 请求携带无效或过期的 Token
- **THEN** 系统返回 401 状态码

### Requirement: 密码安全存储
系统 SHALL 使用 bcrypt 算法对用户密码进行哈希存储，禁止明文存储。

#### Scenario: 密码哈希存储
- **WHEN** 创建用户或修改密码
- **THEN** 密码经过 bcrypt 哈希后存入数据库，数据库中不存在明文密码

### Requirement: 统一响应格式
所有 API 响应 SHALL 使用统一的 JSON 格式：`{code: number, message: string, data: any}`。

#### Scenario: 成功响应
- **WHEN** API 请求处理成功
- **THEN** 返回 `{code: 0, message: "ok", data: {...}}`

#### Scenario: 错误响应
- **WHEN** API 请求处理失败
- **THEN** 返回 `{code: <非零错误码>, message: "<错误描述>", data: null}`

### Requirement: CORS 跨域支持
系统 SHALL 在开发环境下允许跨域请求。

#### Scenario: 跨域请求处理
- **WHEN** 前端从不同端口发起 API 请求
- **THEN** 后端正确处理 CORS，允许跨域访问

### Requirement: 默认管理员账号
系统 SHALL 在数据库初始化时创建默认管理员账号。

#### Scenario: 默认管理员可登录
- **WHEN** 系统首次初始化后，使用用户名 `admin` 密码 `admin123` 登录
- **THEN** 登录成功，获取有效 Token
