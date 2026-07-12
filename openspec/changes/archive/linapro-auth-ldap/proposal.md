## Why

企业环境仍大量使用 LDAP/Active Directory 作为账号目录。现有协议插件均为 OAuth/OIDC 浏览器跳转，无法覆盖「在 Lina 表单输入目录账号密码并 bind」的路径。外部身份地基已就绪，应新增独立 managed 源码插件，在不改宿主密码登录内核的前提下提供通用 LDAP 登录。

## What Changes

- 新增 managed 源码插件 `linapro-auth-ldap`：单目录 LDAP/AD 配置、LDAPS/StartTLS bind 验真、登录页独立入口弹层、管理设置页。
- v1 单 connection：`provider = ldap:default`；subject 取可配置稳定目录属性（推荐 entryUUID/objectGUID，回退 uid）。
- 平台全局；硬依赖 `linapro-extlogin-core`。
- 主密码登录不动；`auth.login.after` 提供目录登录弹层；`POST` 插件公开 API 完成 bind → `LoginByVerifiedIdentity` → 返回 handoff JSON（非 JWT）。
- 自动开户默认关闭；密码只用于 bind，永不落库、不进日志。
- 不接管宿主密码路径；不做目录全量同步/组映射；不多目录 UI；不交付动态插件形态。

## Capabilities

### New Capabilities

- `linapro-auth-ldap`：配置、bind 验真、provider 归属、登录 API/弹层、依赖与安全边界。

### Modified Capabilities

- （无）

## Impact

- 新增 `apps/lina-plugins/linapro-auth-ldap/`；依赖 `linapro-extlogin-core`；宿主原则上零领域契约变更。
- 前端复用 `auth.login.after` + 弹层与 `completeExternalLoginFromHandoff`。
- 插件 `go.mod` 引入 LDAP 客户端库。
