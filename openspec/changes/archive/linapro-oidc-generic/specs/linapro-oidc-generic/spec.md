## ADDED Requirements

### Requirement: 插件标识、分发与依赖

系统 SHALL 提供 managed 源码插件 `linapro-oidc-generic`。该插件 MUST 在 `plugin.yaml` 中声明对 `linapro-extlogin-core` 的版本依赖（`>=0.1.0`）。未满足依赖时 MUST 无法成功启用。插件 MUST NOT 作为宿主 builtin 强制安装。

#### Scenario: 缺少 extlogin-core 时无法启用

- **WHEN** 管理员尝试启用 `linapro-oidc-generic` 且 `linapro-extlogin-core` 未安装或未启用
- **THEN** 宿主依赖治理 MUST 阻止启用并给出依赖提示

#### Scenario: 按需安装

- **WHEN** 部署未安装任何第三方登录协议插件
- **THEN** 宿主启动 MUST NOT 因缺少 `linapro-oidc-generic` 而失败，且密码登录不受影响

### Requirement: Provider 归属与编码

插件 SHALL 通过 `ProvideExternalIdentity` 声明稳定 provider ID `oidc:default`（v1 单 connection，key 固定为 `default`）。权威身份键 MUST 为 `(provider, subject)`，其中 `subject` 为 IdP `id_token` 的 `sub` claim。插件 MUST NOT 使用 email 作为解析键，MUST NOT 持有外部身份链接表。

#### Scenario: 仅拥有声明的 provider

- **WHEN** 插件完成身份验签并调用 `LoginByVerifiedIdentity`
- **THEN** 请求中的 `provider` MUST 为 `oidc:default`，且宿主 ownership 校验 MUST 通过

#### Scenario: 无 sub 时失败

- **WHEN** IdP 返回的 `id_token` 缺少非空 `sub`
- **THEN** 插件 MUST 拒绝登录并返回明确错误，不得用 email 或 name 冒充 subject

### Requirement: OIDC 授权码登录与安全校验

插件 SHALL 在公开 portal 路径提供登录发起与回调：

- `GET /portal/linapro-oidc-generic/login`
- `GET /portal/linapro-oidc-generic/callback`

登录流 MUST 使用 Authorization Code + PKCE (S256)。插件 MUST 通过 OIDC Discovery 解析授权/令牌/JWKS 端点（基于已配置 issuer）。回调 MUST 校验 state、nonce，并用 JWKS 验证 `id_token` 的签名及 `iss`/`aud`/`exp`（及实现要求的标准声明）。验签成功后 MUST 调用宿主 `ExternalLogin().LoginByVerifiedIdentity`，再经 `linapro-extlogin-core` handoff 将**一次性 handoff 码**回跳 SPA；回跳 URL MUST NOT 包含 access/refresh JWT。

#### Scenario: 未配置凭证 fail-closed

- **WHEN** Client ID 或 Client Secret 为空或为占位值，用户访问 login-start
- **THEN** 系统 MUST NOT 将浏览器 302 到 IdP，MUST 回跳 SPA 并携带安全错误语义

#### Scenario: 成功登录仅 handoff

- **WHEN** 用户完成 IdP 授权且本地身份解析/开户策略允许登录
- **THEN** 回调重定向到 SPA 的 query MUST 包含 handoff 码且 MUST NOT 包含 accessToken 或 refreshToken

#### Scenario: PKCE 与 state

- **WHEN** 插件构造 authorize URL
- **THEN** URL MUST 包含 `code_challenge`/`code_challenge_method=S256` 与不可伪造的 state；回调 state 校验失败时 MUST 拒绝换票

### Requirement: 自动开户默认关闭

插件 SHALL 将 `AllowAutoProvision` 暴露为管理员可配置项，**默认关闭**。仅当配置显式开启时，调用 `LoginByVerifiedIdentity` 才允许 `AllowAutoProvision=true`。未链接身份且未开启自动开户时，行为 MUST 与宿主/extlogin-core 既有未开户拒绝语义一致。邮箱冲突防接管策略 MUST 继续由宿主/core 执行，插件不得静默绑定同邮箱他人账号。

#### Scenario: 默认不自动开户

- **WHEN** 新安装插件且管理员未修改自动开户设置，未链接的外部身份尝试登录
- **THEN** 登录 MUST 失败为未开户/未链接类错误，且 MUST NOT 创建新本地用户

#### Scenario: 开启后允许 JIT

- **WHEN** 管理员开启自动开户且身份未链接、邮箱不与他人冲突（或无邮箱）
- **THEN** 宿主/core 开户路径 MAY 创建最小权限用户并完成登录（由既有引擎策略决定）

### Requirement: 管理设置

插件 SHALL 提供受权限保护的设置 API 与前端设置页，至少支持配置：Issuer、Client ID、Client Secret（读取脱敏）、可选 Redirect URL、scopes、显示名称、自动开户开关、可选 SPA 落地路径。保存后 MUST 在后续登录请求生效而无需重启宿主（请求时解析配置）。Client Secret 投影 MUST NOT 返回明文；空或掩码提交 MUST 保留原 secret。

#### Scenario: Secret 脱敏

- **WHEN** 管理员 GET 设置且已保存 secret
- **THEN** 响应 MUST 使用固定掩码指示已配置，MUST NOT 包含 secret 明文

#### Scenario: 菜单挂载

- **WHEN** `linapro-extlogin-core` 与本插件均已安装启用
- **THEN** 设置菜单 MUST 出现在 `plugin:linapro-extlogin-core:auth-login` 目录下，并具备 view/update 权限点

### Requirement: 登录入口与降级

插件 SHALL 通过工作台 `auth.login.after` 槽位贡献登录入口。插件未启用、依赖不满足、或凭证未配置时，入口 MUST 不展示或不可发起有效 IdP 跳转。显示文案 MUST 优先使用配置的显示名称，缺省时使用插件 i18n 默认文案。

#### Scenario: 禁用后入口消失

- **WHEN** `linapro-oidc-generic` 被禁用
- **THEN** 登录页 MUST NOT 继续展示该插件登录入口

#### Scenario: 凭证齐全时展示

- **WHEN** 插件已启用且 Issuer、Client ID、Client Secret 均已有效配置
- **THEN** 登录页 MUST 展示可点击的企业 OIDC 登录入口

#### Scenario: 登录入口使用工作台统一认证样式

- **WHEN** 插件在登录页渲染企业 OIDC 登录入口
- **THEN** 入口按钮 MUST 复用工作台统一前端组件和主题`token`，并与其他外部登录入口按全宽单行纵向排列，且在桌面与移动视口中 MUST NOT 出现溢出或元素重叠

### Requirement: 协议边界

插件 MUST 只负责 OIDC 协议交互与已验证身份提交。插件 MUST NOT 铸造宿主 JWT、MUST NOT 直接写 `sys_user` 链接表、MUST NOT 在公开 HTTP 上接受客户端自报的裸 `provider`+`subject` 绑定。绑定已有用户 MUST 走 extlogin-core ticket 体系（若本变更提供绑定入口；否则可不暴露绑定 HTTP，仅登录路径）。

#### Scenario: 无链接表

- **WHEN** 审查插件 SQL/DAO
- **THEN** 插件 MUST NOT 新增外部身份链接业务表；链接归属仍为 `linapro-extlogin-core`

### Requirement: 国际化与文档

插件 `i18n.enabled` 为 true 时 MUST 提供 en-US 与 zh-CN 的 plugin/menu/error（及所需 apidoc）资源。MUST 提供双语 README，说明依赖 `linapro-extlogin-core`、安装顺序、provider 编码 `oidc:default`、回调 URL、安全约束与审查清单。

#### Scenario: 中英文菜单

- **WHEN** 工作台语言为 en-US 或 zh-CN 且插件已安装
- **THEN** 设置菜单标题 MUST 显示对应语言本地化文案，而非仅依赖库内中文种子
