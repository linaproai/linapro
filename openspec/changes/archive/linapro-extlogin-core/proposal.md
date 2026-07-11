## Why

外部身份登录的存储与开户策略曾长在宿主核心（`sys_user_external_identity`、开户策略、链接解析），与「宿主保持通用、业务通过插件扩展」相悖；每接入新 IdP 的存储或策略差异都要改 `lina-core`。需要把外部身份领域抽到 `linapro-extlogin-core`，宿主只保留薄 `extlogin` 契约、provider SPI 与 token/session 铸造。

同时，能力面若仅按 Google/Discord 裁剪、core 误标 builtin、裸 Bind、token 进 URL，则无法支撑微信/QQ 等协议与按需装卸。需要完整领域契约、ticket 证明链绑定、安全 handoff、managed 分发与未安装降级。

## What Changes

- 新增并完善源码插件 `linapro-extlogin-core`：私有链接表、Resolve/Provision/Bind 编排、`extidcap` 宽领域契约、ticket、catalog、handoff。
- 宿主新增 `extidspi`（manager-backed lazy provider）；`LoginByExternalIdentity` 改为委托 provider；token/session/租户/hook 留宿主。
- 宿主移除 `sys_user_external_identity` 表与工件；保留最小权限 `ProvisionExternal`（含 UsernameAnchor）供插件反向建号。
- **最终分发**：`distribution: managed`（从早期 builtin 校正为按需安装，宿主不强制装入）。
- 绑定仅 `BindByTicket`；协议回调 SPA 仅 handoff，禁止 JWT 进 URL；SPA 向 core 交换 handoff。
- 扩展 `VerifiedIdentity`（SubjectKind、SecondarySubjects、AppContext 等）。
- 协议插件（google/discord 及后续）硬依赖 core；登录入口按形态分槽（`auth.login.after` 协议全宽按钮、`auth.login.social` 平台图标行）。
- 无邮箱开户、自隔离绑定/解绑/列举、并发唯一索引正确性。

## Capabilities

### New Capabilities

- `external-identity-provider-seam`：宿主 SPI、fail-closed、token 留宿主、建号 seam、ownership。
- `linapro-extlogin-core-identity-store`：链接存储、开户编排、无邮箱 anchor、绑定解绑、并发与禁用/卸载处置、数据权限边界。
- `linapro-extlogin-core-domain`：managed 领域 owner、`extidcap` 完整契约、ticket 绑定、扩展模型、未安装降级。
- `linapro-extlogin-core-handoff`：一次性 handoff 与 SPA 交换，宿主不暴露兑换 HTTP。
- `linapro-extlogin-core-protocol-plugins`：协议插件依赖 core、只做验签与 ownership、入口随启停隐藏。
- `login-external-auth-presentation`：登录页外部入口按形态分槽展示。

### Modified Capabilities

- （无既有 baseline 强制 delta；行为以本归档与 `openspec/specs` 为准。）

## Impact

- 宿主：auth/extidspi 装配、用户建号 anchor、移除外部身份表、登录页双槽位、文档。
- 插件：`linapro-extlogin-core` 全量；google/discord handoff 与依赖。
- 前端：handoff 交换、去掉 query token 依赖。
- 数据权限：链接为用户自隔离；登录不泄露他账号存在性。
