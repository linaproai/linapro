## Why

企业环境仍大量使用 LDAP/Active Directory 作为账号目录。现有协议插件均为 OAuth/OIDC 浏览器跳转，无法覆盖「在 Lina 表单输入目录账号密码并 bind」的路径。外部身份地基（`linapro-extlogin-core` + `extlogin` + handoff）已就绪，应新增独立 managed 源码插件，在不改宿主密码登录内核的前提下提供通用 LDAP 登录。

## What Changes

- 新增 managed 源码插件 `linapro-auth-ldap`：单目录 LDAP/AD 配置、LDAPS/StartTLS bind 验真、登录页独立入口弹层、管理设置页。
- v1 **单 connection**：`provider = ldap:default`；subject 取稳定目录属性（可配置，默认优先 `entryUUID`/`objectGUID`/`uid` 策略见 design）。
- 平台全局（`platform_only` + `global`）；硬依赖 `linapro-extlogin-core`。
- 登录交互：主密码登录不动；`auth.login.after` 提供「目录登录」弹层；`POST` 插件公开 API 完成 bind → `LoginByVerifiedIdentity` → 返回 **handoff**（JSON，非 JWT）。
- 自动开户 **默认关闭、可配置开启**；密码只用于 bind，**永不落库、不进日志**。
- **不**接管宿主密码登录路径；**不**做目录全量同步/组映射；**不**做多目录 UI；**不**做 dynamic 插件形态。

## Capabilities

### New Capabilities

- `ldap-auth-plugin`: `linapro-auth-ldap` 协议插件能力——配置、bind 验真、provider 归属、登录 API/弹层、依赖与安全边界。

### Modified Capabilities

- （无已归档 baseline 修改。）

## Impact

- **新增插件** `apps/lina-plugins/linapro-auth-ldap/`
- **依赖** `linapro-extlogin-core`
- **宿主** 原则上零领域契约变更
- **前端** 复用 `auth.login.after` + 弹层；handoff 兑换复用既有 `completeExternalLoginFromHandoff`
- **第三方依赖** 插件 `go.mod` 引入 LDAP 客户端库（如 `github.com/go-ldap/ldap/v3`）
