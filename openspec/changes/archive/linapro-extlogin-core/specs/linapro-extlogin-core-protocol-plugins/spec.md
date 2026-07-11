## ADDED Requirements

### Requirement: 协议插件硬依赖 linapro-extlogin-core

每个第三方登录协议插件（含 `linapro-oidc-google`、`linapro-oidc-discord`、`linapro-oidc-generic`、`linapro-auth-ldap` 及未来协议插件）SHALL 在 `plugin.yaml` 的 `dependencies.plugins` 中声明对 `linapro-extlogin-core` 的版本约束。未满足依赖时 SHALL 无法成功启用该协议插件。

#### Scenario: 缺少 core 时无法启用协议插件

- **WHEN** 用户尝试启用协议插件且 `linapro-extlogin-core` 未安装或未启用
- **THEN** 宿主依赖治理 SHALL 阻止启用并给出依赖提示

### Requirement: 协议插件只负责验签与 provider 归属

协议插件 SHALL 声明 `ProvideExternalIdentity(providerID)`，完成 IdP 协议验签，并调用宿主 `ExternalLogin` 与/或 `extidcap` ticket 能力。协议插件 MUST NOT 持有链接表或实现开户策略。

#### Scenario: 登录路径委托宿主与 core

- **WHEN** 协议插件完成身份验签
- **THEN** 它 SHALL 通过宿主外部登录 seam 换取会话结果，链接解析与开户由 core 引擎完成

### Requirement: 登录入口随插件启停隐藏

协议插件的登录入口 SHALL 通过工作台登录槽位或等价机制贡献；插件未启用时入口 MUST NOT 展示。

#### Scenario: 禁用协议插件后按钮消失

- **WHEN** 某协议插件被禁用
- **THEN** 登录页 MUST NOT 继续展示该插件登录入口
