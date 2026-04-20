## MODIFIED Requirements

### Requirement: 会话存储抽象层
系统 SHALL 将在线会话存储、会话有效性校验和会话活跃时间维护视为宿主认证会话内核；`monitor-online` 只消费宿主提供的会话投影和治理能力。

#### Scenario: 在线用户插件未安装
- **WHEN** `monitor-online` 未安装或未启用
- **THEN** 宿主仍在 `sys_online_session` 中创建、删除、校验和清理会话记录
- **AND** 登录、登出、受保护接口鉴权与超时判定继续正常工作

#### Scenario: 在线用户插件已启用
- **WHEN** `monitor-online` 已安装并启用
- **THEN** 插件通过宿主提供的会话投影查询在线用户并执行强制下线治理
- **AND** 插件不拥有 JWT 校验、`last_active_time` 维护或清理任务真相源

### Requirement: 在线用户列表查询
系统 SHALL 在 `monitor-online` 已安装并启用时提供管理员查询当前在线用户的治理能力，支持按用户名和 IP 地址过滤。

#### Scenario: 查询在线用户列表
- **WHEN** `monitor-online` 已安装并启用且管理员请求在线用户列表
- **THEN** 插件返回宿主会话投影中的在线会话记录列表
- **AND** 每条记录仍包含 token_id、username、dept_name、ip、login_location、browser、os、login_time 等治理字段

### Requirement: 强制下线
系统 SHALL 在 `monitor-online` 已安装并启用时支持管理员强制下线指定在线用户，被下线用户的后续请求 MUST 返回 401。

#### Scenario: 插件执行强制下线
- **WHEN** 管理员通过 `monitor-online` 对指定 `tokenId` 执行强制下线
- **THEN** 宿主会话内核失效该会话记录
- **AND** 后续携带该 Token 的请求被宿主认证中间件拒绝

### Requirement: 系统监控菜单
系统 SHALL 在 `monitor-online` 已安装并启用时，将 `在线用户` 菜单作为插件菜单挂载到宿主 `系统监控` 目录，而不是要求与 `服务监控` 共同作为固定内建子菜单出现。

#### Scenario: 菜单展示
- **WHEN** `monitor-online` 已安装、已启用且当前用户有权访问其菜单
- **THEN** `系统监控` 下显示 `在线用户` 子菜单
- **AND** 该规则不要求 `服务监控` 同时存在

#### Scenario: 插件缺失或停用
- **WHEN** `monitor-online` 未安装、未启用或当前用户无权访问其菜单
- **THEN** 宿主隐藏 `在线用户` 菜单入口
- **AND** 宿主认证会话内核继续独立运行
