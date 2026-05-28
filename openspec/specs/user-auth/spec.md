# 用户认证

## Purpose

用户认证功能负责系统登录、退出、会话验证、登录状态维护以及登录后用户信息和权限数据返回，确保前后端认证流程稳定一致。
## Requirements
### Requirement: 用户名密码登录

系统 SHALL 支持用户名 + 密码登录，登录请求 MUST 携带受控的用户会话客户端类型`clientType`。允许值仅包括`web`、`mobile`、`desktop`、`cli`；系统 MUST 拒绝缺失或未知的客户端类型。验证成功后系统返回 JWT Token。登录过程中（无论成功或失败）SHALL 发出统一的登录生命周期事件，且事件 MUST 包含登录请求中的`clientType`。登录成功后，SHALL 在宿主维护的`sys_online_session`表中创建会话记录；如果`linapro-monitor-loginlog`已启用，插件根据登录事件完成登录日志入库。

#### Scenario: 登录成功
- **当** 用户向`POST /api/v1/auth/login`提交正确的用户名、密码和`clientType=web`时
- **则** 系统返回 JWT Token，响应格式为`{code: 0, message: "ok", data: {accessToken: "...", refreshToken: "..."}}`
- **且** 宿主在`sys_online_session`表中创建会话记录（包含 token_id、用户信息、IP、浏览器、操作系统、client_type 等）
- **且** 宿主签发的 access token 和 refresh token claims 均包含`clientType=web`
- **且** 宿主发出登录成功事件，事件`clientType`为`web`；如果`linapro-monitor-loginlog`已启用，插件写入登录成功日志

#### Scenario: 登录失败且日志插件缺失
- **当** 用户提交错误凭证和合法`clientType=mobile`
- **且** `linapro-monitor-loginlog`未安装、未启用或初始化失败
- **则** 系统仍返回正确的登录失败结果
- **且** 宿主发出登录失败事件，事件`clientType`为`mobile`
- **且** 宿主不因缺少特定登录日志持久化实现而报错

#### Scenario: 拒绝未知客户端类型
- **WHEN** 用户向`POST /api/v1/auth/login`提交`clientType=plugin`
- **THEN** 系统拒绝登录请求
- **AND** 不签发 JWT
- **AND** 不创建在线会话

### Requirement: 用户退出

系统 SHALL 支持用户退出操作。退出操作 SHALL 发出统一的登录生命周期事件并删除宿主维护的在线会话记录。退出事件 MUST 使用当前认证上下文中的会话`clientType`，不得硬编码或重新推断客户端类型。

#### Scenario: 退出成功
- **当** 已登录用户调用`POST /api/v1/auth/logout`时
- **则** 系统返回成功响应，从`sys_online_session`表中删除该用户的会话记录，前端清除本地存储的 Token
- **且** 宿主发出退出成功事件，事件`clientType`等于该 token claims 和会话记录中的`clientType`
- **且** 如果`linapro-monitor-loginlog`已启用，插件写入对应日志

### Requirement:认证中间件
系统 SHALL 提供认证中间件来保护需要登录才能访问的 API。中间件必须同时验证 JWT 签名和宿主维护的会话有效性，不得依赖 `linapro-monitor-online` 是否安装。

#### Scenario:在线用户插件未安装时访问受保护接口
- **当** 请求头携带有效的 `Authorization: Bearer <token>`，`sys_online_session` 中存在对应的会话记录，且 `linapro-monitor-online` 未安装或未启用时
- **则** 中间件仍通过 UPDATE 操作将会话的 `last_active_time` 更新为当前时间
- **且** 请求正常处理，用户信息注入请求上下文

### Requirement:登录后返回用户信息

用户登录成功后，系统 SHALL 返回用户信息，包括角色和菜单树。

#### Scenario:登录成功返回用户信息
- **当** 用户使用正确的用户名密码登录时
- **则** 系统返回 userId、username、realName、email、avatar
- **且** 系统返回 roles 字段，包含用户所有角色的 key 列表
- **且** 系统返回 menus 字段，包含用户可访问的菜单树
- **且** 系统返回 permissions 字段，包含用户拥有的权限标识列表
- **且** 系统返回 homePath 字段，指定用户的首页路径

#### Scenario:超级管理员登录
- **当** admin 角色用户登录时
- **则** 系统返回所有菜单（不检查 sys_role_menu 关联）
- **且** roles 包含 "admin"
- **且** permissions 包含 "*:*:*" 表示拥有所有权限

#### Scenario:普通用户登录
- **当** 非超级管理员用户登录时
- **则** 系统根据用户角色查询 sys_role_menu 获取菜单 ID 列表
- **且** 系统根据菜单 ID 列表构建菜单树
- **且** 系统过滤掉状态为停用（status = 0）的菜单
- **且** 系统过滤掉隐藏的菜单（visible = 0）

#### Scenario:用户登录时无角色
- **当** 用户登录时未分配任何角色
- **则** 系统返回空菜单树
- **且** roles 为空数组
- **且** permissions 为空数组

#### Scenario:用户所有角色被停用
- **当** 用户的所有角色都被停用时
- **则** 系统返回空菜单树
- **且** roles 为空数组
- **且** permissions 为空数组

### Requirement: 认证生命周期事件可供插件订阅

系统 SHALL 将登录成功、登录失败和退出成功等认证生命周期事件作为受控 Hook 发布给已启用的插件。事件 payload 中的`clientType` MUST 来自受控用户会话客户端类型，不得包含`service`、`plugin`或其他非用户客户端主体值。

#### Scenario: 登录成功后发布认证事件
- **当** 用户使用`clientType=desktop`登录成功时
- **则** 宿主向订阅了`auth.login.succeeded`的插件分发事件
- **且** 事件包含宿主暴露的用户身份和客户端上下文
- **且** 事件`clientType`为`desktop`

#### Scenario: 退出成功后发布认证事件
- **当** `clientType=cli`的用户退出成功时
- **则** 宿主向订阅了`auth.logout.succeeded`的插件分发事件
- **且** 事件`clientType`为`cli`
- **且** 事件分发不改变原始退出成功语义

### Requirement:插件认证扩展失败不影响认证结果
系统 SHALL 保证插件在认证事件上的扩展失败不会改变登录或退出的最终结果。

#### Scenario:登录成功后插件 Hook 报错
- **当** 某插件在登录成功事件处理中失败时
- **则** 用户仍收到登录成功结果
- **且** 系统记录插件失败信息以便排查

### Requirement:JWT Token 有效期配置
系统 SHALL 支持通过 `config.yaml` 中的 `jwt.expire` 时长字符串配置 JWT Token 有效期。

#### Scenario:使用新时长配置 Token 有效期
- **当** 管理员在 `config.yaml` 中设置 `jwt.expire=24h` 时
- **则** 系统必须使用该时长值作为 JWT Token 有效期

### Requirement:菜单树结构

系统 SHALL 返回满足前端路由生成要求的菜单树。

#### Scenario:菜单树包含必要字段
- **当** 系统返回菜单树时
- **则** 每个菜单节点包含 id、parentId、name、path、component、icon、type、sort、visible、status 字段
- **且** 目录类型（type="D"）的菜单包含 children 子节点
- **且** 菜单类型（type="M"）的菜单为叶子节点
- **且** 按钮类型（type="B"）不在菜单树中返回

#### Scenario:菜单树按 sort 字段排序
- **当** 系统返回菜单树时
- **则** 同级菜单按 sort 字段升序排列

### Requirement:权限标识列表

系统 SHALL 返回用户的所有权限标识。

#### Scenario:权限标识聚合
- **当** 用户拥有多个角色时
- **则** 系统聚合所有角色的权限标识（去重）
- **且** 权限标识来自菜单表中 type="M" 或 type="B" 的 perms 字段

#### Scenario:超级管理员权限
- **当** 用户是超级管理员（拥有 admin 角色）时
- **则** permissions 返回 ["*:*:*"]
- **且** 前端判定该权限标识拥有所有权限

### Requirement:运行时配置的 JWT 过期时间
系统 SHALL 允许 `sys.jwt.expire` 在运行时控制新签发 JWT Token 的生命周期，无运行时覆盖时回退到静态配置。

#### Scenario:运行时 JWT 过期时间生效
- **当** 管理员维护 `sys.jwt.expire=24h` 时
- **则** 新签发的 JWT Token 使用该时长作为有效过期时间

### Requirement:运行时配置的登录 IP 黑名单
系统 SHALL 允许 `sys.login.blackIPList` 在运行时控制登录 IP 黑名单。

#### Scenario:登录请求被配置的黑名单拒绝
- **当** 登录请求来自 `sys.login.blackIPList` 匹配的 IP 或 CIDR 范围时
- **则** 系统拒绝登录尝试
- **且** 登录日志记录登录 IP 被禁止的失败原因

### Requirement: 集群模式 token 撤销必须使用 Redis
系统 SHALL 在集群模式下使用 Redis coordination KV 存储 JWT revoke 状态。revoked token key 的 TTL MUST 等于 token 剩余有效期。

#### Scenario: 登出写入 Redis revoke
- **WHEN** 用户在集群模式下调用 logout
- **THEN** 系统写入 Redis revoke key
- **AND** key TTL 不超过 JWT 剩余有效期
- **AND** 后续使用该 token 的请求被拒绝

#### Scenario: 切换租户撤销旧 token
- **WHEN** 用户调用 switch-tenant 并获得新 token
- **THEN** 系统将旧 token ID 写入 Redis revoke store
- **AND** 旧 token 在所有节点立即不可用

### Requirement: token 撤销读取失败必须 fail-closed
系统 SHALL 在集群模式认证链中读取 Redis revoke 状态。读取失败时 MUST fail-closed，不得仅凭 JWT 签名放行。

#### Scenario: Redis revoke 读取失败
- **WHEN** 请求携带签名有效的 JWT
- **AND** Redis revoke store 读取失败
- **THEN** 认证链拒绝请求
- **AND** 返回结构化认证失败或服务不可用错误

### Requirement: pre_token 必须使用 Redis 单次 TTL 状态

系统 SHALL 在集群模式下使用 Redis 存储`pre_token`、候选租户、用户会话`clientType`和 single-use 消费状态。`pre_token` MUST 短期有效且只能使用一次。

#### Scenario: pre_token 首次选择租户
- **WHEN** 用户使用`clientType=mobile`完成密码验证并获得有效`pre_token`
- **AND** 用户使用该`pre_token`调用 select-tenant
- **THEN** 系统原子消费 Redis 中的`pre_token`
- **AND** 签发正式 JWT
- **AND** 正式 JWT 和在线会话的`clientType`均为`mobile`
- **AND** 后续同一`pre_token`不可再次使用

#### Scenario: pre_token 重放
- **WHEN** 客户端第二次使用同一`pre_token`
- **THEN** 系统拒绝请求
- **AND** 不签发正式 JWT

### Requirement: 认证短期状态必须集中使用 coordination KV
登录验证码、一次性 token、认证 challenge 或后续同类短 TTL 安全状态 SHALL 在集群模式下使用 coordination KV。模块不得自行创建 Redis key 或直接访问 Redis client。

#### Scenario: 新增一次性认证 challenge
- **WHEN** 后续认证流程需要保存一次性 challenge
- **THEN** 实现通过 coordination KV 写入带 TTL 的状态
- **AND** 不直接导入 Redis client

