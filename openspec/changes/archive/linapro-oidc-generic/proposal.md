## Why

第三方登录仅有品牌插件 `linapro-oidc-google` / `linapro-oidc-discord`，无法对接 Keycloak、Okta、Azure AD 等可配置 issuer 的企业 IdP。外部身份地基（`linapro-extlogin-core` + 宿主 `extlogin` + handoff）已就绪，需要官方通用 OIDC 协议插件，在不改动宿主核心领域的前提下按需扩展企业登录能力。

## What Changes

- 新增 managed 源码插件 `linapro-oidc-generic`：通用 OIDC Authorization Code + PKCE、OIDC Discovery、JWKS 验 `id_token`、管理设置页、登录页入口。
- v1 单 connection：`connectionKey` 固定 `default`，`provider = oidc:default`；配置与 ownership 预留多连接扩展，不交付多连接管理 UI。
- 平台全局（`platform_only` + `global`）；硬依赖 `linapro-extlogin-core`；验签后走 `LoginByVerifiedIdentity`，SPA 回跳仅 handoff；自动开户默认关闭、可配置开启。
- 注册 `extidcap` Provider 目录与 `auth.login.after` 登录入口；未配置凭证 / 未装 core / 插件禁用时 fail-closed 或隐藏。
- 设置菜单挂在 `plugin:linapro-extlogin-core:auth-login` 下；双语 README/i18n。
- 不实现 LDAP、不替换 google/discord、不交付动态插件形态。

## Capabilities

### New Capabilities

- `linapro-oidc-generic`：协议插件能力边界——provider 归属与编码、OIDC 登录/回调、验签、settings、登录入口、依赖与降级。

### Modified Capabilities

- （无）

## Impact

- 新增 `apps/lina-plugins/linapro-oidc-generic/`（backend portal/settings/OIDC、frontend settings 与 login slot、manifest、plugin.yaml、双语 README）。
- 依赖 `linapro-extlogin-core`；宿主原则上无领域契约变更。
- 前端复用 `auth.login.after` 全宽协议按钮形态。
