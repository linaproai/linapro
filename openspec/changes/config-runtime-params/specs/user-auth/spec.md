## ADDED Requirements

### Requirement: JWT Token 有效期运行时配置
系统 SHALL 支持通过参数设置中的 `sys.jwt.expire` duration 字符串参数控制新签发 JWT Token 的有效期；当该参数未设置时，继续回退到 `config.yaml` 中的 `jwt.expire` 或默认值。

#### Scenario: 使用运行时参数控制 Token 有效期
- **WHEN** 管理员在参数设置中维护 `sys.jwt.expire=24h`
- **THEN** 系统 MUST 使用该 duration 值作为新签发 JWT Token 的有效期

### Requirement: 登录 IP 黑名单运行时配置
系统 SHALL 支持通过参数设置中的 `sys.login.blackIPList` 控制登录 IP 黑名单，阻止命中的来源地址完成登录。

#### Scenario: 拒绝命中黑名单的登录请求
- **WHEN** 当前登录请求来源 IP 命中 `sys.login.blackIPList` 中配置的 IP 或 CIDR 规则
- **THEN** 系统拒绝登录并返回失败结果
- **AND** 登录日志记录失败原因“登录IP已被禁止”
