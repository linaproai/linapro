## Context

外部身份登录当前在 `feat/plugin-auth` 分支的形态是：宿主 `extlogin.Service` 契约已经很薄（一个方法 `LoginByVerifiedIdentity` + 三个 DTO），但支撑它的**存储与开户策略**长在宿主核心：

- 链接表 `sys_user_external_identity`（宿主 SQL `013-auth-external-identity.sql` + DAO/DO/Entity 生成工件）。
- `auth_external_identity.go` 的 `LoginByExternalIdentity`（231 行）与 `tryAutoProvision`。
- `user_provision_external.go` 的建号 + 邮箱校验。

逐行分析 `LoginByExternalIdentity` 后可将其拆成两类：

- **A. 天生属于核心 auth（占 ~90%，不可下放）**：IP 黑名单、禁用账号检查、登录 hook 派发、租户解析（`loginTenants`）、多租户 pre-token（`preTokens.Create`）、**token 铸造（`generateTokenPair`）**、会话持久化（`createSession`）、登录时间更新。这些是 auth 服务私有方法，与密码登录共用同一 seam。
- **B. 可下放（仅两块）**：`(provider, subject)` 链接表存储与解析、开户策略。

框架已有 provider SPI 先例：`orgspi`、`tenantspi` 都是"插件实现 provider、宿主持有 manager 并注入、provider 缺失返回中性值/降级"。本设计照抄该模式，把 B 部分迁到 `linapro-extid-core` 插件，A 部分原地不动。项目为全新项目，无历史数据负担。

## Goals / Non-Goals

**Goals:**

- 宿主核心只保留"薄领域接口 + SPI 定义"：`extlogin.Service`（不变）+ 新增 `ExternalIdentityProvider` SPI；链接表与开户策略随 `linapro-extid-core` 装卸。
- provider 缺失时 `extlogin` fail-closed，语义与 tenant/org 能力缺失一致。
- provider ownership 治理（`ProvideExternalIdentity` 盖章、拒绝未拥有 provider 与禁用插件）保持不变。
- 顺带在插件内收敛两个 follow-up：无邮箱开户、绑定/解绑外部身份。
- Google/Discord 插件登录调用路径不变（仍走 `extlogin` seam），仅新增对 `linapro-extid-core` 的运行时依赖。

**Non-Goals:**

- 不改动 token 铸造、会话、租户解析、pre-token、登录 hook 等核心 auth 编排——这是硬边界。
- 不把 external login 能力开放给动态插件（继续 fail-closed stub，安全论证见 PR #54）。
- 不引入新的前端页面结构变化（登录按钮仍走 `auth.login.after` 槽位）。
- 不做存量数据迁移（新项目，插件安装建表 + 宿主删表即可）。

## Decisions

### D1. provider SPI 边界：解析 + 开户 + 绑定，不含 token

`ExternalIdentityProvider` 契约方法围绕"外部身份 ↔ 本地 userID"，示意：

```
Resolve(ctx, ResolveInput{Provider, Subject}) (userID int, found bool, err error)
Provision(ctx, ProvisionInput{Provider, Subject, Email, DisplayName, PluginID, AllowAutoProvision}) (userID int, err error)
Bind(ctx, BindInput{UserID(当前会话), Provider, Subject, Email}) error
Unbind(ctx, UnbindInput{UserID(当前会话), Provider, Subject}) error
List(ctx, UserID(当前会话)) ([]BoundIdentity, err error)
```

宿主 `LoginByExternalIdentity` 改为：provider ownership 盖章（不变）→ `Resolve`，未命中则按 `AllowAutoProvision` 调 `Provision` → 拿到 userID 后进入**原有**租户解析/token 铸造/会话/hook 流程（不变）。`ExternalProvisionInput`/`ProvisionExternalUser` 保留在宿主用户域作为"最小权限建号"原语，供 provider 反向调用；但"查同邮箱、决定是否开户"的策略移入插件。

理由：满足架构规则「新增抽象须隔离已确认变化点」——微信/QQ/SAML 的存储与开户差异都收敛在 provider 后面；宿主契约按「以当前稳定职责为中心保持简洁」只暴露解析/开户/绑定，不泄露 token 内部。

### D2. 链接表迁为插件私有表 `user_external_identity`

去 `sys_` 前缀（插件私有表约定），字段沿用：`user_id`、`provider`、`subject`、`plugin_id`、`email_snapshot`、时间/软删字段按 `database.md`。`(provider, subject)` 唯一索引。建表走插件安装 SQL，卸载走 uninstall SQL。插件维护自身 `hack/config.yaml` DAO 生成，宿主移除该表全部工件。

理由：`plugin.md`「禁止插件重新依赖宿主的 dao/do/entity 生成工件」+「插件 SQL 遵守 database.md」。

### D3. 数据权限：用户自隔离例外

外部身份链接是用户自隔离资源。登录解析走 `(provider, subject)` 唯一键，返回统一 not-provisioned 不泄露他账号存在性；绑定/解绑/列举仅作用于当前会话用户。按 `data-permission.md` 例外要求记录：权威边界=当前会话用户 + `(provider,subject)` 唯一键；拒绝策略=越权目标整体拒绝、重复占用返回冲突；合理例外=当前用户自隔离链接。

### D4. fail-closed 通过 provider 缺失实现

宿主 auth 注入的 provider 为 nil（`linapro-extid-core` 未安装/未启用）时，`LoginByVerifiedIdentity` 返回 not-provisioned，不建号不铸 token。与 `orgspi`/`tenantspi` 的"provider 不在返回中性值"同构，但外部登录的中性值就是 fail-closed 拒绝。

### D5. 插件依赖关系

`linapro-oidc-google`/`linapro-oidc-discord` 声明对 `linapro-extid-core` 的依赖。core 为 `distribution: builtin`（宿主启动自动安装启用，保证依赖可满足）；google/discord 维持现状 `distribution: managed`——它们是带占位凭证的 OAuth 参考实现，不应被宿主启动强制安装，由用户按需启用（对原「三者均 builtin」表述的实施校正）。core 提供存储与 provider 实现，google/discord 仍只负责 OAuth 协议验签 + 调 `extlogin` seam。

### D5b. provider 注册与注入：manager-backed lazy service，非 raw push（校正 task 1.2 记述漂移）

`extidspi.Provider` 的注册与注入**必须照抄 `tenantspi`/`orgspi` 的 manager 模式**（本文件 Context 行 14、D4 行 21 已如此声明），而非 `BindExternalProvisioner` 式的 raw provider 直推。理由是硬约束：`framework-capability-registry` 规范要求源码插件 provider 实例由消费方 service 经 `IsProviderEnabled` **惰性解析**、且插件启用状态是 provider 可用性的唯一权威；D7 要求 disable→fail-closed、re-enable→恢复。raw push（静态字段）既无法在启动期用插件作用域能力（`usercap.Service`）构造 provider，也无法在插件被 disable 后自动失活，违反 D4/D7。

**落地形态**（最小 surface，复刻 tenant）：

- `extidspi` 包新增 `ProviderEnv{PluginID, Users usercap.Service, BizCtx bizctxcap.Service}`、`ProviderFactory`、`Manager`/`NewManager`/`RegisterFactory`、`New(manager, enablement, envFactory) Provider`——`New(...)` 返回的 `Provider` 惰性解析已启用 factory，无启用 provider 时 fail-closed（`Resolve`→`found=false`；`Provision`/`Bind`/`Unbind`/`List`→not-provisioned 错误）。新增能力键常量 `CapabilityExternalIdentityV1`。
- `pluginhost.ProviderDeclarations` 新增 `ProvideExternalIdentityProvider(factory)`——与既有**字符串** `ProvideExternalIdentity(providerID)`（google/discord 的 ownership 盖章）**并存且正交**：字符串 API 门禁"调用方插件是否拥有该 provider ID"，factory API 声明"谁提供 resolve/provision 引擎"。`linapro-extid-core` 只注册 factory、不声明任何 provider ID 字符串（它从不调用 `LoginByVerifiedIdentity`）。google/discord 的字符串 ownership 与测试**零改动**。
- `httpstartup_runtime.go`：新增 `externalIdentityProviderManager = extidspi.NewManager()`；`externalIdentitySvc = extidspi.New(externalIdentityProviderManager, pluginRuntime, pluginRuntime.ExternalIdentityProviderEnv)`；在 `BindExternalProvisioner` 邻近调用 `authSvc.BindExternalIdentityProvider(externalIdentitySvc)`（注入的是 **manager-backed service**，非插件 raw provider）；`RegisterSourcePluginProviderFactories` 增加第 4 个 manager 参数。
- **task 1.2 记述校正**：task 1.2 已实现的 `BindExternalIdentityProvider` seam 与 `identityProvider` 字段**保留不变**（复用），但其"沿用 `BindExternalProvisioner` 直注入范式"的措辞仅指"post-construction 绑定这一动作"，注入值必须是 manager-backed lazy service；HANDOFF 旧述"插件从 `backend/plugin.go` 调 `authSvc.BindExternalIdentityProvider`"作废——插件只在声明期注册 factory，绑定由宿主启动装配完成，插件不得直达 `authSvc`。此为记述校正，非设计分叉：与 Context 行 14/D4 一致。

### D6. 插件可调的 provisioning seam + 用户名 anchor + 事务契约（回应 Momus #1/#2/#3）

**问题**：现状 `ProvisionExternalUser`（`user_provision_external.go:43-83`）① 硬拒空/无 `@` 邮箱（`:45-46`），② 从邮箱本地部分派生用户名（`resolveProvisionUsername:87-88`），③ 只经 host 内部 `auth.ExternalProvisioner` seam（`auth.go` `BindExternalProvisioner` 进程内绑定）暴露，插件无法反向调用；④ 现有 provision+链接写入的单事务原子性依赖两表同属宿主同一 `dao.SysUserExternalIdentity.Transaction`。链接表迁到插件后，这四点全部断裂。`orgspi`/`tenantspi` 不是先例——它们的 adapter 只**消费**宿主读能力，没有反向调用宿主写原语的模式。

**决策**：

- **扩展建号原语接受用户名 anchor**：`ProvisionExternalInput`/`usercap` provision DTO 增加可选 `UsernameAnchor` 字段。邮箱为空时用 anchor 派生用户名（如 `wechat-<subject 短哈希>`），并把 `:45-46` 的 `@` 校验放宽为"邮箱为空且 anchor 提供 → 允许"。邮箱非空仍走原派生。"用户如何被 shape"仍留宿主用户域（`user_provision_external.go:1-8` 职责声明不变），只是 shape 输入锚点从"仅邮箱"扩展为"邮箱或 anchor"。**这是必须的 host 契约变更**，不是 task 2.3 之前误写的"只保留纯建号"。
- **新增插件可调的 provisioning seam**：在 `usercap`（宽接口）新增最小权限外部建号方法（区别于 `usercap.Create` 的操作员建号——后者带租户/角色/创建边界校验，非 provisioning 的 operator-less 语义）。插件 provider 通过注入的 `usercap.Service` 调用它建号，宿主内部仍委托 `ProvisionExternalUser`。
- **事务契约与去重锚点**：provision（宿主 `sys_user` 写）+ 链接写入（插件 `user_external_identity` 写）跨模块，无法共享单 GoFrame TX。权威去重锚点是**链接表 `(provider,subject)` 唯一索引**——不是 `sys_user` 的 email/username（核对 `001-user-auth-bootstrap.sql`：仅 `uk_sys_user_username` 唯一，`email` 无唯一索引且默认 `''`，且 `resolveProvisionUsername` 遇冲突加数字后缀，天然不会触发"冲突复用")。因此正确性保证收敛为"**同一 `(provider,subject)` 最终只有一条有效链接、指向一个账号**"，而非"只建一个 `sys_user`"。实现顺序两选一：① 先以 `(provider,subject)` 抢占链接（先插占位或用唯一索引冲突）再补建号，避免游离账号；② 先建号→再写链接，链接唯一索引冲突时复用已链接账号，竞态下可能留下未链接建号孤儿（按 D7 容忍）。无邮箱 anchor 派生用户名**必须确定性可复现**，使同一 anchor 命中同一 username → 复用同一账号。

### D6a. provisioning seam 的插件作用域：源码插件专用 + 动态 fail-closed stub

新增的 `usercap` 外部建号方法（`usercap.Service.ProvisionExternal`）为**源码插件专用**。唯一消费者 `linapro-extid-core` 为 `type: source`，通过注入的 `usercap.Service` 直接调用。动态（WASM）插件通过 `usercap.Service` guest 契约得到 **fail-closed stub**（`domainhostcall_users.go` 的 `usersService.ProvisionExternal` 直接返回错误，不发 host call），与 `authcap` 的 `ExternalLogin()`（`domainhostcall_auth.go:43-63` 的 `externalLoginService`）同构；**不**登记 protocol 常量、host-service descriptor 或 WASM dispatcher case，即不发布为动态 `users` host service。

理由：

- `design.md` Non-Goal 明确不向动态插件开放 external login 能力（`继续 fail-closed stub`）；外部建号是该 fail-closed 边界的 provisioning 半环，不能拆分后单独对 WASM 开放。
- operator-less、最小权限的建号原语暴露给沙箱 WASM 会扩大攻击面，与本次 fail-closed 重构的安全立场相悖。
- `usercap.Create`（已发布为动态）是 operator 治理建号（带租户/角色/创建边界校验），治理面不同，不构成对本方法的发布先例。
- 架构规则 `architecture.md:22` 的"源码/动态双可用"由 guest fail-closed stub 满足其"受治理降级"语义（参见 `pluginbridge-subcomponent-architecture` 规范：领域降级行为归属 `pkg/plugin/capability`，`pluginbridge` 只按需增加 transport/descriptor 同步点），不要求该方法在 WASM 侧可调用。

验证：`domainhostcall_users_test.go` 的 `TestUsersProvisionExternalFailsClosed` 断言动态路径返回错误且不触达 host transport。

### D7. disable 与 uninstall 的数据处置（回应 Momus #2 场景缺口）

- **disable（保留数据）**：`linapro-extid-core` 被禁用时，`ExternalIdentityProvider` 注入视为缺失 → external login fail-closed（D4），但 `user_external_identity` 表**保留**，重新启用后链接恢复可用。
- **uninstall（清理表 + 孤儿处置）**：卸载 SQL 删表后，宿主 `sys_user` 中经外部登录建的账号成为**无外部链接的孤儿账号**——它们有独立 userID、可被管理员重置密码后用密码登录，因此**不级联删除 sys_user**（避免误删真实用户）。卸载语义 = "断开所有外部身份链接，账号保留"，须在插件 README 与 uninstall SQL 注释写明。

### D8. 并发正确性（回应 Momus #2 TOCTOU）

现状 email/username 均为 `Count`-then-`Insert` 的 TOCTOU（`auth_external_identity.go:52-60`、`user_provision_external.go:108-120`）。迁移后 `user_external_identity` 的 `(provider,subject)` 唯一索引是最终防线：并发下两请求同时 resolve 未命中 → 都尝试 provision+link，其一因唯一索引冲突失败。设计规定：**唯一索引冲突必须被捕获并转为 not-provisioned 复用路径或冲突错误，MUST NOT 冒泡为 500**，须有单测覆盖。

## Risks / Trade-offs

- **多一层间接**：登录路径多一次 provider 调用。可接受——换来 IdP 扩展点收敛，符合架构规则对新增抽象的门槛。
- **插件依赖链**：google/discord → core。若 core 未启用，两者 external login fail-closed；需在插件依赖声明与 README 中说明，避免"装了 google 却登录被拒"的困惑。
- **重构而非纯搬运**：涉及新建 SPI 包、表迁移、宿主装配改注入、既有插件依赖调整；工作量按一个完整 openspec 变更计。缓解：A 部分（token/session）完全不动，回归面集中在 B 部分与装配注入。
- **`ProvisionExternalUser` 归属**：建号 shape 留宿主用户域，开户**策略**（是否开户、同邮箱/anchor 复用、无邮箱走 anchor）在插件。边界不能只靠注释——必须通过 D6 的契约变更（DTO 增 anchor + 新增插件可调 provision 方法）落实，否则 Momus #1/#2 的"插件调不到、无邮箱建不了号"会在实现中卡死。
- **跨边界事务放弃单事务**：D6 用"先建号→再链接 + 建号幂等"替代单 TX；风险是链接写入失败留下未链接账号，靠幂等复用兜底，须单测覆盖失败补偿路径。
- **Bind/Unbind/List 是用户可达能力**：spec 要求"已登录用户绑定/解绑"，须有插件 `backend/api/` DTO + controller + 路由 + apidoc i18n，不能只停在 service 层（Momus #3）。
- **数据权限例外须显式验证**：绑定/解绑越权、重复占用、登录不泄露存在性、并发唯一索引冲突不冒泡 500 四条须有单测覆盖，否则审查不通过。
- **SPI 包路径待对齐**：新 SPI 置于 `authcap/extlogin/extidspi`，而先例是 `<domain>cap/<domain>spi`（`orgcap/orgspi`）。非阻塞，但须在实现时对齐或在设计记录理由。
