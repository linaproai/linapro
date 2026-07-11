## Context

工作区已具备：

- 宿主 `authcap.ExternalLogin.LoginByVerifiedIdentity`：协议插件提交已验证身份 → 会话/pre-token
- `extidspi` + manager：链接解析/开户委托 `linapro-extlogin-core`，未启用 fail-closed
- 协议参考实现 `linapro-oidc-google` / `linapro-oidc-discord`：portal 登录/回调、settings 经 `sys_config`、slot 登录入口、handoff 回跳
- `extidcap`：catalog 注册、`CreateLoginHandoffFromHost`、SPA `handoff/exchange`

缺口：无法对接可配置 issuer 的企业 OIDC IdP；品牌插件硬编码 Google/Discord 端点与 UX。

已确认产品决议：

1. v1 单 connection，provider 预留 `oidc:{key}`
2. 平台全局，非租户级
3. LDAP 另插件、另变更
4. JIT 默认关、可配置开
5. managed source 插件，OIDC 优先于 LDAP

## Goals / Non-Goals

**Goals:**

- 交付可安装启用的 `linapro-oidc-generic` 源码插件，完成真实 OIDC 授权码登录（非 stub verifier）。
- 单 connection 配置：Issuer（Discovery）、Client ID/Secret、可选 Redirect URL、scopes、显示名/图标、AllowAutoProvision。
- 安全：PKCE、state（HMAC）、nonce、`iss`/`aud`/`exp`/签名校验；secret 设置页脱敏。
- 与现有协议插件同构集成：`ProvideExternalIdentity`、extlogin-core 依赖、handoff、catalog、菜单目录、slot。
- 凭证未配置时 fail-closed（不 302 到伪造 client_id）。

**Non-Goals:**

- 多 connection 管理 UI、租户级 IdP、LDAP、SAML、One Tap。
- 改宿主 `extlogin` 契约或链接表归属。
- 用 generic 替换 google/discord 或抽取跨插件共享库包（可内部复用模式，不强制抽公共 module）。
- 动态 WASM 插件交付、组 claim → 角色映射、RP-Initiated Logout（可后续）。
- SSO 第三方业务 receiver 映射（google 的 backend_redirects）——v1 可不做；仅 SPA handoff + 可选 default 落地路径。

## Decisions

### D1. 插件 ID 与分发

- ID：`linapro-oidc-generic`
- `type: source`，`distribution: managed`，`scope_nature: platform_only`，`default_install_mode: global`
- `dependencies.plugins: [{ id: linapro-extlogin-core, version: ">=0.1.0" }]`
- 菜单 `parent_key: plugin:linapro-extlogin-core:auth-login`

**替代**：builtin 强制安装——否决，与「第三方登录可装卸」一致。

### D2. Provider 编码

- v1 固定 connection key：`default`（常量，设置页可不展示或只读展示）
- `provider = "oidc:default"`
- `subject = id_token.sub`（必填；无 sub 则失败）
- `SubjectKind = oidc_sub`（经 extid 开户路径时）
- `ProvideExternalIdentity("oidc:default")` 在插件 init 声明

**替代**：

- 共用 `provider=oidc` + subject 拼 issuer——否决，运维与解绑差
- 用 issuer URL 当 provider——否决，issuer 微调断链

多连接时将来为每个 key 再 `ProvideExternalIdentity("oidc:"+key)`；v1 只声明一个。

### D3. OIDC 协议面

| 项 | 选择 |
| --- | --- |
| 流 | Authorization Code + **PKCE (S256)** |
| 发现 | `GET {issuer}/.well-known/openid-configuration`（issuer 规范化去尾 `/`） |
| 身份 | 优先验 **id_token**（JWKS）；userinfo 可选补 displayName |
| scopes 默认 | `openid email profile`（可配置，必须含 `openid`） |
| state | HMAC 签名自包含 payload（对齐 google：不依赖跨站 Cookie 存活） |
| nonce | 写入 authorize，校验 id_token.nonce |
| 凭证未配置 | login-start 回 SPA 错误，不跳 IdP |

Discovery 结果可短缓存（进程内 TTL，如 10–15 分钟），密钥轮换靠 JWKS 刷新；缓存失败则每次拉取。

**库选择**：优先标准库 + 成熟 OIDC/JWT 校验（如 `github.com/coreos/go-oidc/v3` 或等价），避免手写 JWT 校验；依赖进入插件 `go.mod`，宿主无感。

### D4. 配置存储

对齐 google：经宿主 `hostconfigcap.SysConfig` 持久化插件作用域键，例如：

| Key | 含义 |
| --- | --- |
| `plugin.linapro-oidc-generic.connection_key` | 固定 `default`（只读语义） |
| `plugin.linapro-oidc-generic.display_name` | 登录按钮/目录显示名 |
| `plugin.linapro-oidc-generic.issuer` | OIDC issuer URL |
| `plugin.linapro-oidc-generic.client_id` | Client ID |
| `plugin.linapro-oidc-generic.client_secret` | Client Secret（投影脱敏） |
| `plugin.linapro-oidc-generic.redirect_url` | 可选覆盖；空则由请求 host 推导 portal callback |
| `plugin.linapro-oidc-generic.scopes` | 空格或 JSON 列表；默认 openid email profile |
| `plugin.linapro-oidc-generic.allow_auto_provision` | `"1"` 开，默认关 |
| `plugin.linapro-oidc-generic.default_backend_redirect` | 可选 SPA 落地路径 |
| `plugin.linapro-oidc-generic.enabled_login` | 是否展示登录入口（默认：凭证齐全即展示） |

若现有 google 用 manifest SQL 种子 `sys_config`：generic 同等处理（平台 `tenant_id=0`）；无 SQL 则首次保存懒创建（以 google 既有模式为准，实施时对齐同一插件惯例）。

### D5. HTTP 路由

```
PUBLIC  /portal/linapro-oidc-generic/login      GET  → BuildAuthorizeURL + 302 IdP
PUBLIC  /portal/linapro-oidc-generic/callback   GET  → code 交换 + LoginByVerifiedIdentity + handoff 302 SPA
PROTECTED /x/.../api/v1/settings                GET/PUT  权限 linapro-oidc-generic:settings:view|update
```

回调成功：`extidcap.CreateLoginHandoffFromHost` → SPA 仅带 `handoff` query；**禁止** JWT 进 URL。

错误回跳：安全错误码 + i18n 键，禁止 `err.Error()` 原文。

### D6. 前端

- `frontend/pages/settings.vue`：Issuer、Client ID/Secret、Redirect、scopes、自动开户、显示名；复制回调 URL
- `frontend/slots/auth.login.after/generic-oidc-login-entry.vue`：图标+Tooltip 或「企业登录」文案；凭证未配置时隐藏或禁用
- 菜单：`plugin:linapro-oidc-generic:settings` + update 按钮权限；i18n `menu.json` / `plugin.json` / `error.json` en-US + zh-CN

### D7. 与 google 插件的关系

- **产品并列**，不互相依赖
- **实现可抄结构**（settings / oauth 分包、state codec、config resolver），但插件内闭环，本变更不强制抽取 `oidclib` 公共包（降低跨插件耦合与 go.mod 复杂度）
- 后续若第三、第四个 OIDC 插件出现再评估共享库

### D8. 测试策略

- 单元：authorize URL（含 PKCE challenge）、state 编解码、id_token 校验（固定密钥 JWKS fixture）、未配置 fail-closed、handoff 路径 mock
- 集成：settings get/save 脱敏；依赖 core 声明静态检查
- E2E：安装 core+generic 后菜单可见；未配置时入口不跳假 IdP；可选 mock IdP 或 wire 级 callback 探测（与 google TC 风格对齐，放插件 `hack/tests/e2e`）

### D9. 宿主改动边界

- **默认零宿主领域改动**
- 仅当源码插件聚合清单需要显式 import 时增加 blank import（与现有插件相同）
- 不修改 `extlogin` / `extidspi` / 密码登录路径

## Risks / Trade-offs

| 风险 | 缓解 |
| --- | --- |
| Discovery/JWKS 网络失败导致登录不可用 | 短缓存 + 明确错误码；不缓存错误永久 |
| connection key 日后多连接改名断链 | v1 固定 `default` 且文档声明 key 不可变 |
| Client secret 进日志 | 禁止打印 secret；错误日志只记 provider/subject 哈希级信息 |
| 与 google 设置项/SSO 规则不一致导致体验分裂 | v1 明确裁剪 backend_redirects；文档写清差异 |
| 用户未装 extlogin-core | 依赖治理阻止启用；README 安装顺序 |
| IdP 不返回 email | 允许 email 空，走无邮箱开户策略（若开 JIT）或仅已绑定用户登录 |

## Migration Plan

- 新插件，无历史数据迁移
- 安装：`linapro-extlogin-core` → `linapro-oidc-generic` → 配置 issuer/凭证 → 启用
- 回滚：禁用/卸载插件；链接表中 `provider=oidc:default` 行随 core 表保留，不级联删用户
- 卸载清数据策略跟随插件 SQL uninstall（若有 settings 种子则清理 sys_config 键）

## Open Questions

- 无阻塞项；以下实施期按惯例选定即可：
  - OIDC Go 库具体选型（go-oidc vs 自研最小校验）——优先 go-oidc
  - settings 是否需要「测试连接」按钮——v1 可不做，靠登录探测
  - 登录按钮文案默认「企业登录」还是 display_name——优先 display_name，空则 i18n 默认「OIDC 登录」
