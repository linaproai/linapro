## ADDED Requirements

### Requirement: 外部登录结果经一次性 handoff 交付

协议插件完成宿主外部登录后，SHALL 将 access/refresh/pre-token 结果登记为 **linapro-extid-core** 管理的一次性 handoff 码（`extidcap.CreateLoginHandoffFromHost`），并通过 SPA 回跳仅传递 `handoff`（及安全 status），MUST NOT 将 accessToken 或 refreshToken 放入回跳 URL 查询串或 hash 查询。

#### Scenario: 成功回跳仅含 handoff

- **WHEN** Google/Discord 回调登录成功
- **THEN** 回跳 URL SHALL 包含 handoff 码且 MUST NOT 包含 accessToken 或 refreshToken 字面量

### Requirement: SPA 交换 handoff

`linapro-extid-core` SHALL 提供公开 handoff 交换入口（插件 API 前缀下、无需会话）：消费一次性码并返回与密码登录兼容的 token 对或 preToken+租户候选。宿主 MUST NOT 暴露 handoff 兑换 HTTP。码过期、已消费或不存在时 SHALL 返回统一失败，MUST NOT 泄露会话内容。

#### Scenario: 交换成功进入工作台

- **WHEN** SPA 向 `linapro-extid-core` 提交未消费的有效 handoff
- **THEN** 该插件 SHALL 返回 token 对或 preToken，且该 handoff 之后不可再兑换

#### Scenario: 重复交换失败

- **WHEN** 同一 handoff 被第二次提交
- **THEN** 系统 SHALL 拒绝并返回错误

#### Scenario: 宿主无 handoff 兑换接口

- **WHEN** 审查宿主 `api/auth` 公开路由
- **THEN** MUST NOT 存在 `/auth/external-login/exchange` 或等价 handoff 兑换端点

### Requirement: 错误回跳不泄露内部错误

协议插件错误回跳的 message SHALL 使用安全摘要或 i18n 错误码派生文案，MUST NOT 直接使用内部 `err.Error()` 原文面向终端用户。

#### Scenario: 回调失败使用安全文案

- **WHEN** 协议插件回调因配置错误或验签失败而回跳 SPA
- **THEN** 回跳 message SHALL 为安全摘要或本地化错误码文案，MUST NOT 包含内部堆栈或原始 `err.Error()` 细节
