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

### Requirement: 动态插件 guest 执行必须具备并发和内存护栏

系统 SHALL 为动态插件 guest 执行提供宿主侧全局并发上限、可选按插件并发上限、执行许可获取超时和单实例内存页上限。动态路由、生命周期、hook、cron discovery 和 cron job 等所有 guest 执行入口 MUST 经过同一护栏。超限等待超时 MUST 返回稳定结构化业务错误，不得进入 guest。

#### Scenario: 全局 guest 并发达到上限

- **WHEN** 当前节点同时运行的 guest 执行数已达到`plugin.runtime.guest.maxConcurrent`
- **AND** 新请求在`plugin.runtime.guest.acquireTimeout`内未获得执行许可
- **THEN** 系统返回稳定资源繁忙错误
- **AND** 不实例化或执行该请求的 guest

#### Scenario: 单插件并发达到上限

- **WHEN** 插件`P`同时运行的 guest 执行数已达到`plugin.runtime.guest.maxConcurrentPerPlugin`
- **AND** 其他插件仍有可用全局额度
- **THEN** 插件`P`的新请求按插件上限排队或超时
- **AND** 其他插件请求不得因插件`P`达到自身上限而被拒绝

#### Scenario: 内存页超过上限

- **WHEN** 动态插件 guest 执行需要超过`plugin.runtime.guest.memoryLimitPages`允许的内存
- **THEN** 宿主拒绝或终止本次 guest 执行
- **AND** 返回结构化资源耗尽错误

#### Scenario: 护栏配置非法

- **WHEN** 管理员配置小于 1 的全局并发、负数按插件并发、非法 duration 或非法内存页上限
- **THEN** 系统拒绝启动或拒绝该运行时配置更新
- **AND** 不使用非法配置覆盖当前有效护栏
