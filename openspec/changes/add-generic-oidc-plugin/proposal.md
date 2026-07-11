## Why

现有第三方登录仅有品牌插件 `linapro-oidc-google` / `linapro-oidc-discord`，无法对接 Keycloak、Okta、Azure AD 等可配置 issuer 的企业 IdP。外部身份地基（`linapro-extid-core` + 宿主 `extlogin` + handoff）已就绪，应新增官方通用 OIDC 协议插件，在不改动宿主核心领域的前提下按需扩展企业登录能力。

## What Changes

- 新增 managed 源码插件 `linapro-oidc-generic`：通用 OIDC Authorization Code + PKCE 登录、OIDC Discovery、JWKS 验 `id_token`、管理设置页、登录页入口槽位。
- v1 **单 connection**：稳定 `connectionKey`（默认 `default`），`provider` 编码为 `oidc:{connectionKey}`；配置与 ownership 预留多连接扩展，本变更不交付多连接管理 UI。
- 平台全局安装（`platform_only` + `global`），与 google/discord 一致；不实现 per-tenant IdP。
- 硬依赖 `linapro-extid-core`；验签后走宿主 `LoginByVerifiedIdentity`，SPA 回跳仅 handoff；自动开户 **默认关闭、可配置开启**。
- 注册 `extidcap` Provider 目录与 `auth.login.after` 登录按钮；未配置凭证 / 未装 core / 插件禁用时入口 fail-closed 或隐藏。
- 插件清单、双语 README、i18n、设置菜单挂在 `plugin:linapro-extid-core:auth-login` 下。
- **不**实现 LDAP（另变更）；**不**用 generic 替换 google/discord；**不**交付动态插件形态。

## Capabilities

### New Capabilities

- `generic-oidc-plugin`: `linapro-oidc-generic` 协议插件的能力边界——provider 归属与编码、OIDC 登录/回调、验签、settings、登录入口、依赖与降级。

### Modified Capabilities

- （无已归档 baseline 修改；本变更建立在活跃外部身份相关工作区能力之上，不改宿主链接表或 `extlogin` 契约语义。）

## Impact

- **新增插件** `apps/lina-plugins/linapro-oidc-generic/`：backend（portal 登录/回调、settings API、OIDC 服务）、frontend（settings 页、login slot）、manifest（i18n/菜单/可选 config seed）、plugin.yaml、README 双语。
- **插件聚合**：源码插件注册入口纳入新插件（与现有 oidc 插件同一路径）。
- **依赖** `linapro-extid-core`：`dependencies.plugins` + 运行时 catalog/handoff；core 未启用时不可用。
- **宿主** `apps/lina-core`：原则上无领域契约变更；仅可能涉及插件列表/聚合构建若需登记（按现有源码插件接入惯例）。
- **前端宿主槽位**：复用既有 `auth.login.after`，无强制改 `login.vue` 结构。
- **后续**：`linapro-auth-ldap` 单独立项；多 connection / 租户级 IdP 另开变更。
