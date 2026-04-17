## ADDED Requirements

### Requirement: 会话超时时间运行时配置
系统 SHALL 支持通过参数设置中的 `sys.session.timeout` duration 字符串参数控制在线会话超时阈值；当该参数未设置时，继续回退到 `config.yaml` 中的 `session.timeout` 或默认值。

#### Scenario: 使用运行时参数控制会话超时阈值
- **WHEN** 管理员在参数设置中维护 `sys.session.timeout=24h`
- **THEN** 系统 MUST 使用该 duration 值作为在线会话超时阈值

### Requirement: 鉴权链路实时执行会话超时校验
系统 SHALL 在每次鉴权时根据当前会话的 `last_active_time` 与 `sys.session.timeout` 实时判断会话是否已超时，而不只是等待定时清理任务。

#### Scenario: 鉴权时拒绝超时会话
- **WHEN** 已登录用户携带 Token 访问受保护 API 且对应在线会话的 `last_active_time` 已超过 `sys.session.timeout`
- **THEN** 认证中间件 MUST 拒绝本次请求并返回 401
- **AND** 系统清理对应的在线会话记录
