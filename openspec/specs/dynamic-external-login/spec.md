# dynamic-external-login Specification

## Purpose
TBD - created by archiving change publish-external-login-to-dynamic-plugins. Update Purpose after archive.
## Requirements
### Requirement: 动态插件可调用外部登录换会话

宿主 SHALL 将 `auth` host service 方法 `external_login.login_by_verified_identity` 发布给动态插件。guest 经授权 host call 后，宿主 MUST 使用调用上下文中的 pluginID 盖章，并完成与源码插件等价的外部登录会话铸造路径。

#### Scenario: 授权动态插件成功登录

- **WHEN** 已启用动态插件声明并获授权调用 `external_login.login_by_verified_identity`，且对请求 `provider` 拥有 resource 授权，且身份已链接或允许开户
- **THEN** 宿主 SHALL 返回 token 对或 preToken+租户候选

#### Scenario: 未授权方法被拒绝

- **WHEN** 动态插件未获授权该方法
- **THEN** 宿主 SHALL 拒绝 host call，MUST NOT 铸造会话

### Requirement: 动态插件 provider ownership

动态插件对外部登录 provider 的 ownership SHALL 通过 hostServices 授权资源表达：请求中的 `provider` MUST 匹配该插件在 `auth` 服务下获授权的 `resources[].ref`。源码插件继续通过 `ProvideExternalIdentity` 声明 ownership。

#### Scenario: 动态插件冒用未授权 provider 被拒绝

- **WHEN** 动态插件获授权调用外部登录方法但 `provider` 不在其授权 resources 中
- **THEN** 宿主 SHALL 拒绝，MUST NOT 铸造会话

### Requirement: 动态插件可调用从外部身份建号

宿主 SHALL 将 `users` host service 方法 `users.create_from_external` 发布给动态插件。经授权后行为 SHALL 与源码路径 `usercap.CreateFromExternal` 等价（最小权限、无操作员、邮箱冲突哨兵）。

#### Scenario: 授权后可建号

- **WHEN** 动态插件获授权调用 `users.create_from_external` 并提供合法邮箱或 UsernameAnchor
- **THEN** 宿主 SHALL 创建最小权限用户并返回用户 ID

