# User Auth

## Purpose

用户认证功能负责系统登录、登出、会话校验、登录态维护以及登录后用户信息与权限数据返回，确保前后端鉴权流程稳定一致。
## Requirements
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

### Requirement: 登录后返回用户信息

系统 SHALL 在用户登录成功后返回用户信息，包括角色和菜单树。

#### Scenario: 登录成功返回用户信息
- **WHEN** 用户使用正确的用户名和密码登录
- **THEN** 系统返回 userId、username、realName、email、avatar
- **AND** 系统返回 roles 字段，包含用户所有角色的 key 列表
- **AND** 系统返回 menus 字段，包含用户可访问的菜单树
- **AND** 系统返回 permissions 字段，包含用户所有的权限标识列表
- **AND** 系统返回 homePath 字段，指定用户的首页路径

#### Scenario: 超级管理员登录
- **WHEN** admin 角色的用户登录
- **THEN** 系统返回所有菜单（不检查 sys_role_menu 关联）
- **AND** roles 包含 "admin"
- **AND** permissions 包含 "*:*:*" 表示所有权限

#### Scenario: 普通用户登录
- **WHEN** 非超级管理员用户登录
- **THEN** 系统根据用户的角色查询 sys_role_menu 获取菜单ID列表
- **AND** 系统根据菜单ID列表构建菜单树
- **AND** 系统过滤掉停用状态（status=0）的菜单
- **AND** 系统过滤掉隐藏状态（visible=0）的菜单

#### Scenario: 用户无角色登录
- **WHEN** 没有分配任何角色的用户登录
- **THEN** 系统返回空的菜单树
- **AND** roles 为空数组
- **AND** permissions 为空数组

#### Scenario: 用户角色全部停用
- **WHEN** 用户的所有角色都被停用
- **THEN** 系统返回空的菜单树
- **AND** roles 为空数组
- **AND** permissions 为空数组

### Requirement: 认证生命周期事件可供插件订阅
系统 SHALL 将登录成功与登出成功等认证生命周期事件作为受控 Hook 发布给已启用插件。

#### Scenario: 登录成功后发布认证事件
- **WHEN** 用户登录成功
- **THEN** 宿主向订阅 `auth.login.succeeded` 的插件分发事件
- **AND** 事件包含宿主公开的用户身份与客户端上下文

#### Scenario: 登出成功后发布认证事件
- **WHEN** 用户登出成功
- **THEN** 宿主向订阅 `auth.logout.succeeded` 的插件分发事件
- **AND** 事件分发不改变原有登出成功语义

### Requirement: 插件认证扩展失败不影响认证结果
系统 SHALL 保证插件在认证事件上的扩展失败不会改变登录或登出的最终结果。

#### Scenario: 登录成功 Hook 中插件报错
- **WHEN** 某插件在登录成功事件处理中失败
- **THEN** 用户仍然收到登录成功结果
- **AND** 系统记录该插件失败信息用于排查

### Requirement: JWT Token 有效期配置
系统 SHALL 支持通过 `config.yaml` 中的 `jwt.expire` duration 字符串配置 JWT Token 有效期。

#### Scenario: 使用新的 duration 配置 Token 有效期
- **WHEN** 管理员在 `config.yaml` 中设置 `jwt.expire=24h`
- **THEN** 系统 MUST 使用该 duration 值作为 JWT Token 有效期

## ADDED Requirements

### Requirement: 菜单树结构

系统返回的菜单树必须符合前端路由生成要求。

#### Scenario: 菜单树包含必要字段
- **WHEN** 系统返回菜单树
- **THEN** 每个菜单节点包含 id、parentId、name、path、component、icon、type、sort、visible、status 字段
- **AND** 目录类型（type="D"）的菜单包含 children 子节点
- **AND** 菜单类型（type="M"）的菜单为叶子节点
- **AND** 按钮类型（type="B"）不在菜单树中返回

#### Scenario: 菜单树按排序字段排序
- **WHEN** 系统返回菜单树
- **THEN** 同级菜单按 sort 字段升序排列

### Requirement: 权限标识列表

系统 SHALL 返回用户的所有权限标识。

#### Scenario: 权限标识聚合
- **WHEN** 用户有多个角色
- **THEN** 系统聚合所有角色的权限标识（去重）
- **AND** 权限标识来自菜单表中 type="M" 或 type="B" 的 perms 字段

#### Scenario: 超级管理员权限
- **WHEN** 用户是超级管理员（有 admin 角色）
- **THEN** permissions 返回 ["*:*:*"]
- **AND** 前端判断此权限标识为拥有所有权限
