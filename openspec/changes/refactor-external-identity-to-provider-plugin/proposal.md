## Why

当前 `feat/plugin-auth` 分支把外部身份登录的**存储与开户策略**长在了宿主核心里：`sys_user_external_identity` 表、其 DAO/DO/Entity 生成工件、SQL 迁移 `013-auth-external-identity.sql`，以及 `tryAutoProvision` 开户策略全部位于 `apps/lina-core`。第三方身份链接（Google/Discord，以及未来的微信/QQ/SAML）本质上是可插拔的外部集成能力，不属于框架必须内建的核心领域；把链接表和开户策略绑死在宿主，导致每接入一个新 IdP 的存储或策略差异都要改动 `lina-core`，与"宿主保持通用、业务通过插件扩展"的定位相悖。

同时，宿主目前对外部身份的开户策略强制要求邮箱（`user_provision_external.go` 对空邮箱返回 `CodeUserProvisionEmailInvalid`），且缺少"已登录用户绑定外部身份到现有账号"的能力入口——这两点使微信/QQ 这类无邮箱或需绑定流程的 IdP 无法仅通过加插件闭环。

## What Changes

- 新增 `linapro-extlogin-core` 源码插件（`distribution: builtin`），承载外部身份链接存储与开户/绑定编排：拥有插件私有表 `plugin_linapro_extlogin_core_user_external_identity`（`plugin_linapro_extlogin_core_*` 前缀）、链接记录 CRUD、`Resolve(provider, subject) → userID`、开户策略与绑定/解绑逻辑。
- 宿主新增 `ExternalIdentityProvider` SPI（`pkg/plugin/capability/authcap/extlogin/extidspi`），定义"外部身份 → 本地用户"解析与开户的稳定能力接缝；仿照既有 `orgspi`/`tenantspi` 的 provider 管理与注入模式。宿主 provider 缺失时 `extlogin` 走 fail-closed，与 tenant/org 能力缺失返回中性值的处理一致。
- 宿主 `LoginByExternalIdentity` 重构：把"查链接表 + `tryAutoProvision`"替换为调用注入的 `ExternalIdentityProvider`；**token 铸造、会话持久化、租户解析、pre-token、登录 hook、IP 黑名单与禁用账号检查等核心 auth 编排全部留在宿主不变**（这是硬边界——任何登录路径都由宿主铸 token）。
- 宿主移除 `sys_user_external_identity` 相关：DAO/DO/Entity 生成工件、`013-auth-external-identity.sql`、`user_provision_external.go` 中的外部身份链接写入，迁移到插件；宿主仅保留 `ProvisionExternalUser` 这一"最小权限建号"的用户域能力（供 provider 反向调用建号）。
- **BREAKING**（仅对 `feat/plugin-auth` 分支内未发布代码）：`google`/`discord` 插件对 `extlogin` seam 的依赖不变，但外部身份存储改由 `linapro-extlogin-core` 提供，二者新增对 `linapro-extlogin-core` 的运行时依赖（provider 未启用时 external login fail-closed）。
- 收敛两个已知 follow-up 到插件内：放宽无邮箱开户（派生用户名/替代锚点）与新增 `BindVerifiedIdentity` 绑定/解绑能力，均落在 `linapro-extlogin-core` 内部，宿主不再参与具体策略。

## Capabilities

### New Capabilities

- `external-identity-provider-seam`: 宿主侧 `ExternalIdentityProvider` SPI 契约——定义外部身份解析、开户、绑定/解绑的稳定能力接缝，provider 由源码插件实现、宿主持有 manager 并注入；覆盖 fail-closed 语义、provider ownership 治理、与 `extlogin` 登录路径的协作边界，以及 token/session 铸造留在宿主的硬约束。
- `oidc-core-identity-store`: `linapro-extlogin-core` 插件能力——插件私有表 `plugin_linapro_extlogin_core_user_external_identity` 的存储与生命周期、`(provider, subject)` 解析、host-owned 最小权限开户编排、无邮箱开户策略、已登录用户绑定/解绑外部身份，以及数据权限与卸载清理边界。

### Modified Capabilities

<!-- external-identity 登录路径此前未沉淀为基线 spec，其行为随本变更首次以 SPI 边界形式规范化，归入上述新能力；不修改现有基线 spec 的需求。 -->

## Impact

- **宿主 `apps/lina-core`**：`internal/service/auth/auth_external_identity.go`（重构为调用 provider）、`auth_provisioner_bind.go`、`internal/service/user/user_provision_external.go`（收敛为纯建号）；移除 `sys_user_external_identity` DAO/DO/Entity 与 `013-auth-external-identity.sql`；新增 `extidspi` 包与宿主装配注入；`hack/config.yaml` DAO 生成清单移除该表。
- **新插件 `apps/lina-plugins/linapro-extlogin-core`**：新建完整源码插件（`plugin.yaml`、`backend/`、`manifest/sql/`、`manifest/i18n/`、`plugin_embed.go`、`hack/config.yaml`），实现 provider SPI。
- **既有插件**：`linapro-oidc-google`、`linapro-oidc-discord` 新增对 `linapro-extlogin-core` 的依赖声明；登录调用路径不变（仍走 `extlogin` seam）。
- **数据库**：`sys_user_external_identity` → 插件私有表 `plugin_linapro_extlogin_core_user_external_identity`；需数据迁移策略（新项目无历史负担，采用插件安装 SQL 建表 + 宿主删表，无存量数据迁移）。
- **数据权限**：链接表读写与绑定/解绑动作须按 `.agents/rules/data-permission.md` 评估——外部身份链接属用户自隔离资源，绑定/解绑仅作用于当前会话用户，登录解析走 `(provider, subject)` 唯一键不泄露他账号存在性。
- **文档**：`pkg/plugin/README.md` Auth 域条目、`apps/lina-plugins/README.md` 插件清单、新插件 README 双语。
