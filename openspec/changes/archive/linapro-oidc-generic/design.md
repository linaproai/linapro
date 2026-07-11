# Design

## 插件身份与分发

- ID：`linapro-oidc-generic`；`type: source`，`distribution: managed`，`scope_nature: platform_only`，`default_install_mode: global`。
- 硬依赖 `linapro-extlogin-core >=0.1.0`；菜单 `parent_key: plugin:linapro-extlogin-core:auth-login`。
- 否决 builtin 强制安装，与「第三方登录可装卸」一致。

## Provider 编码

- v1 固定 `provider = "oidc:default"`；`subject = id_token.sub`（必填）；`SubjectKind = oidc_sub`。
- 插件 init 声明 `ProvideExternalIdentity("oidc:default")`。
- 否决共用 `provider=oidc` 拼 issuer，以及用 issuer URL 当 provider（运维与解绑差、issuer 微调断链）。

## OIDC 协议面

- Authorization Code + PKCE (S256)；Discovery 解析授权/令牌/JWKS；优先 JWKS 验 `id_token`（iss/aud/exp/sub/nonce）。
- scopes 默认 `openid email profile`（可配，必须含 openid）；state 为 HMAC 自包含 payload。
- 凭证未配置：login-start 回 SPA 错误，不 302 到 IdP。
- Discovery/JWKS 进程内短缓存；失败不永久缓存。依赖进入插件 `go.mod`，宿主无感。

## 配置与路由

- 配置经 `hostconfigcap.SysConfig` 持久化插件作用域键（issuer、client_id/secret 脱敏、redirect、scopes、display_name、allow_auto_provision 默认关、落地路径、connection_key 固定 default）。
- 公开：`GET /portal/linapro-oidc-generic/login|callback`；受保护 settings GET/PUT。
- 成功：`CreateLoginHandoffFromHost` 后 SPA 仅 handoff；禁止 JWT 进 URL；错误回跳禁止 `err.Error()` 原文。

## 前端与协议边界

- settings 页与 `auth.login.after` 槽位；显示名优先配置，缺省 i18n。
- 入口使用工作台统一 Vben 组件与主题 token，全宽纵向布局（协议/目录形态，非社交图标行）。
- 插件只负责协议交互与已验证身份提交；不持有链接表、不铸造 JWT、不接受裸 provider+subject 绑定。
- 与 google/discord 产品并列、实现可抄结构但不强制抽公共 OIDC 库。

## 风险与迁移

- Discovery 网络失败：短缓存 + 明确错误码。connection key 文档声明 v1 不可变。
- 新插件无历史迁移；安装顺序 core → generic → 配置 → 启用；卸载不级联删用户。
