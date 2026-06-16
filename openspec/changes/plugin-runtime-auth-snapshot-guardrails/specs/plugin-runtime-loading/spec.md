## ADDED Requirements

### Requirement: 动态路由认证必须复用宿主权限和会话 owner

系统 SHALL 在动态插件登录路由进入 guest 前复用宿主`role`和`session`owner 完成认证、权限和数据范围快照构建。插件运行时 MUST 通过构造函数显式接收`role`访问投影契约和共享`session.Store`，不得在请求路径中直接查询角色治理表、创建独立 session store 或缓存 session 有效结论。

#### Scenario: 动态路由通过 role 投影执行权限校验

- **WHEN** 动态插件路由声明需要权限`P`
- **THEN** 插件运行时通过`role`访问投影获得当前 token 的权限集合
- **AND** 权限集合不包含`P`且用户不是超管时，请求被拒绝且不进入 guest

#### Scenario: 动态路由使用共享 session hot state

- **WHEN** 动态插件登录路由收到带 JWT 的请求
- **THEN** 插件运行时通过启动期注入的共享`session.Store`校验 tenant、token 和 session timeout
- **AND** session 不存在、已撤销、租户不匹配、token 过期或 hot state 后端 fail-closed 时，请求被拒绝

#### Scenario: 动态路由身份快照携带数据范围

- **WHEN** 动态插件请求通过认证并进入 guest 执行
- **THEN** 传递给 guest 和 host service context 的身份快照包含当前 token 的数据范围、unsupported 标记和超管标记
- **AND** 后续 host service 或 data service 调用继续按该上下文执行数据权限治理
