## ADDED Requirements

### Requirement: JWT Token 有效期配置
系统 SHALL 支持通过 `config.yaml` 中的 `jwt.expire` duration 字符串配置 JWT Token 有效期。

#### Scenario: 使用新的 duration 配置 Token 有效期
- **WHEN** 管理员在 `config.yaml` 中设置 `jwt.expire=24h`
- **THEN** 系统 MUST 使用该 duration 值作为 JWT Token 有效期
