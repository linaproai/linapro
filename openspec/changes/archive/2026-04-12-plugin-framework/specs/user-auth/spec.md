## ADDED Requirements

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
