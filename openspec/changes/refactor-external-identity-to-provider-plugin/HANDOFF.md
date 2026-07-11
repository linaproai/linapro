# Handoff / 工作记录: refactor-external-identity-to-provider-plugin

> 目的：本文件是完整的接手记录，供**另一个会话**独立完成本重构剩余部分，无需依赖原对话上下文。
> OpenSpec 正式产物是 proposal.md / design.md / specs/ / tasks.md（Momus 已两轮审查 APPROVE）；本文件是实现进度 + 接手指南。
> 当前分支：父仓库 `feat/external-login-seam`，插件子仓库 `apps/lina-plugins` 在 `feat/oidc-reference-plugins`。

## 一句话状态

**重构主体已完成**：组 1、1b、2、3、4、5、6、8 全部落地，组 7 除 7.7 端到端探测（deferred，见 tasks 7.7）外完成。tasks.md 当前 37/38 勾选。宿主 login 已切到 manager-backed provider seam（fail-closed），旧表 `sys_user_external_identity` 及全部工件已删除，插件 `linapro-extlogin-core` 已创建并进入插件完整聚合（14 manifests）。验证基线：宿主 `go build`/`go vet ./...` 全绿（含修复组 1 遗留的 2 处 `authcap.New` 三参）、auth 外部登录 9 测试 + SPI manager 3 测试 + 插件 identity 7 测试全过、聚合 `go build` 通过、`make i18n.check` exit 0。**关键设计决议见 design D5b**（manager-backed lazy 注入，非 raw push；task 1.2 记述已校正）。唯一遗留：7.7 需运行完整宿主栈做真实登录链路探测（OAuth stub verifier 可离线走通），完成后勾选即可归档。注意：`TestLoginRejectsBlacklistedIP` 与 lint gofmt 告警均为预存在环境问题（工作目录 i18n 资源路径 / Windows CRLF checkout），已在 tasks 7.6 记录。

## 目标（为什么做这个重构）

把外部身份的**存储与开户策略**从宿主核心抽离到新源码插件 `linapro-extlogin-core`，宿主只保留：
1. 薄消费契约 `extlogin.Service`（已存在，不变）
2. 新增 provider SPI `extidspi.Provider`（组 1 已建）
3. **token/session/tenant 铸造留在宿主**（硬边界，任何登录路径都由宿主铸 token）

顺带收敛两个 follow-up 到插件内：无邮箱开户（微信/QQ）、已登录用户绑定/解绑外部身份。

## 已完成（组 1 + 组 1b.1）— 全部可编译、已验证

### 组 1：宿主 SPI 契约 + 绑定 seam

- **新增** `apps/lina-core/pkg/plugin/capability/authcap/extlogin/extidspi/extidspi.go`
  - `Provider` 接口：`Resolve` / `Provision` / `Bind` / `Unbind` / `List`
  - DTO：`ResolveInput` / `ProvisionInput`（含 `UsernameAnchor`、`PluginID`、`AllowAutoProvision`）/ `BindInput` / `UnbindInput` / `BoundIdentity`
  - 契约无 token/session/租户内部类型
- **改** `apps/lina-core/internal/service/auth/auth.go`
  - 接口加 `BindExternalIdentityProvider(provider extidspi.Provider)`
  - `serviceImpl` 加字段 `identityProvider extidspi.Provider`（nil = fail-closed）
  - `ExternalProvisionInput` 加 `UsernameAnchor` 字段
  - import 加 `extidspi`
- **改** `apps/lina-core/internal/service/auth/auth_provisioner_bind.go`
  - 新增 `BindExternalIdentityProvider` 实现（沿用既有 `BindExternalProvisioner` 直注入范式，非 capabilityregistry）
  - import 加 `extidspi`

### 组 1b.1：无邮箱 anchor 建号

- **改** `apps/lina-core/internal/service/user/user.go`：`ProvisionExternalInput` 加 `UsernameAnchor`
- **改** `apps/lina-core/internal/service/user/user_provision_external.go`
  - `ProvisionExternalUser` 放宽 `@` 校验：邮箱空 + anchor 提供 → 允许（email 置空不写无效值）
  - 新增 `resolveProvisionUsernameFromAnchor`：**确定性、无数字后缀**派生（幂等复用基础）
- **改** `apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_provision_adapter.go`：透传 `UsernameAnchor`

### 验证结果（本会话实测）

- `cd apps/lina-core; rtk go build ./...` → Success
- `rtk go vet ./internal/service/auth/... ./internal/service/user/... ./pkg/plugin/capability/authcap/...` → clean
- `rtk go test ./internal/service/auth/... ./internal/service/user/...` → 80 passed, 1 failed
  - 唯一 failed = `TestLoginRejectsBlacklistedIP`，已用 `git stash` 对照证明是**预存在的 i18n 资源扫描失败**（`i18n_locale.go:70 scan host i18n locale resources failed`），与本重构无关，勿追。

## 剩余工作（按安全顺序：~~1b.3~~ → 4 → 5 → 2 → 3 → 6 → 7 → 8，**从组 4 起**）

> 顺序理由：先建插件并注册 provider（4/5），再把宿主 login 切到 provider（2），最后删旧表工件（3）。组 2 必须先于组 3。每组做完必须 `rtk go build ./...` 通过再进下一组。

### 1b.3 — usercap 插件可调最小权限建号方法 —— **DONE**

- `usercap.Service` 新增 `ProvisionExternal(ctx, ProvisionExternalInput) (UserID, error)` + `ProvisionExternalInput` DTO（`pkg/plugin/capability/usercap/usercap.go`）。
- 宿主实现落在 `internal/service/user/capabilityadapter/user_capability.go` 的 `userCapabilityAdapter`（**注意：实际 adapter 在此，不是先前 HANDOFF 猜测的 `capabilityhost/`**），委托 `owner.ProvisionExternalUser`。
- **作用域=源码插件专用**：动态侧 `domainhostcall_users.go` 提供 fail-closed guest stub（仿 `Auth.ExternalLogin()`/`domainhostcall_auth.go`），**不**登记 protocol 常量 / descriptor / WASM dispatcher（决策见 design.md **D6a**，Oracle 背书）。
- 接口新增方法后补齐全部 9 处实现桩：host 5（`capabilityadapter` 真实实现 + `domainhostcall_users` stub + `testutil_services`/`plugin_test`/`integration_test`/`capabilityhost_session_test`/`wasm_host_service_test` fake no-op）、插件 4（tenant-core ×3 + org-core ×1 测试 fake）。
- 新增 `domainhostcall_users_test.go::TestUsersProvisionExternalFailsClosed`（断言 fail-closed 且不触达 transport）。tasks 1b.3 已记 DI/影响/验证。
- **⚠️ 预存在阻塞（非本任务，需组 2 附近顺手修）**：`go vet ./...` 有 2 个 `authcap.New` arity 错误——`wasm_host_service_test.go:664`、`integration_test.go:125`。组 1 auth SPI 给 `authcap.New` 加了第三参 `extlogin.Service`，但这两个测试文件的 `authcap.New(...)` 调用点未同步（`git stash` 对照证明与 1b.3 无关）。

### 4 — 新建 linapro-extlogin-core 插件（在 apps/lina-plugins 子仓库）

- 参照 `apps/lina-plugins/linapro-demo-source/` 结构：`plugin.yaml`（type: source, distribution: builtin, 无菜单、仅能力）、`plugin_embed.go`、`go.mod`、`Makefile`、`hack/config.yaml`（DAO 生成）、`backend/plugin.go`。
- 私有表 `plugin_linapro_extlogin_core_user_external_identity`（**`plugin_linapro_extlogin_core_*` 前缀**）：字段 `user_id`、`provider`、`subject`、`plugin_id`、`email_snapshot`、`created_at`、`updated_at`、`deleted_at`，**`(provider, subject)` 唯一索引**。建表 SQL 放 `manifest/sql/`，卸载 SQL 放 `manifest/sql/uninstall/`。遵守 `.agents/rules/database.md` 软删+时间字段。
- 在插件目录跑 `make dao` 生成 DAO/DO/Entity。
- **验收硬证据**：`apps/lina-plugins/linapro-extlogin-core/plugin.yaml` 必须存在于磁盘。

### 5 — 插件实现 provider SPI + Bind/Unbind/List 交付面 + i18n

- 在 `backend/internal/service/` 实现 `extidspi.Provider`：
  - `Resolve`：按 `(provider,subject)` 唯一键查表。
  - `Provision`：**去重锚点 = `(provider,subject)` 唯一索引（不是 sys_user email——email 无唯一索引）**；无邮箱 → 确定性、碰撞抗性 `UsernameAnchor`（从 `(provider,subject)` 取足够熵，username `VARCHAR(64)`/provision 上限 30 字符有余量）；调组 1b.3 的宿主最小权限建号方法；唯一索引冲突必须捕获、**绝不冒泡 500**；同邮箱已存在启用账号 → 冲突拒绝（防静默接管）。
  - 事务：跨模块无单一 TX，先建号→再写链接（或先抢占链接再补建号），建号幂等靠 `(provider,subject)` 复用。
  - `Bind`/`Unbind`/`List`：**仅作用于当前会话用户**；越权目标整体拒绝；重复占用 `(provider,subject)` 返回冲突。
- Bind/Unbind/List 用户可达交付面：`backend/api/` DTO + 权限标签、controller、路由注册（registrar）、`manifest/i18n/<locale>/apidoc/` API 文档翻译（遵守 `.agents/rules/api-contract.md` 的 `g.Meta`/`dc`/`eg`）。
- `manifest/i18n/{zh-CN,en-US}/error.json` 错误码本地化（开户冲突、越权、重复绑定等）。
- 在 `backend/plugin.go` 通过 registrar/启动装配把 provider 注册到宿主：找到 `BindExternalProvisioner` 的调用点 `apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go:221`，仿照调用 `authSvc.BindExternalIdentityProvider(<插件 provider>)`。
- **数据权限**（`.agents/rules/data-permission.md`）：外部身份链接是用户自隔离资源——登录解析走 `(provider,subject)` 唯一键不泄露他账号存在性；bind/unbind/list 仅当前会话用户。在设计/审查记录例外（权威边界=当前用户+唯一键；拒绝策略=越权整体拒绝/重复占用冲突）。

### 2 — 重构宿主 LoginByExternalIdentity（破坏性）

- 改 `apps/lina-core/internal/service/auth/auth_external_identity.go`：
  - `ProvideExternalIdentity` 盖章保留不变。
  - 把 `sys_user_external_identity` 查询 + `tryAutoProvision` 替换为调注入的 `identityProvider.Resolve`，未命中且 `AllowAutoProvision` → `identityProvider.Provision`。
  - `identityProvider` 为 nil → fail-closed `CodeAuthExternalUserNotProvisioned`。
  - **移除** `auth_external_identity.go:52-60` 的邮箱冲突判定（`CodeAuthExternalEmailConflict`，迁到插件）。
  - **保留不动**：IP 黑名单、禁用账号检查、租户解析、pre-token、`generateTokenPair`、`createSession`、登录时间更新、登录成功/失败 hook。

### 3 — 移除宿主 sys_user_external_identity 表工件（破坏性，必须在组 2 之后）

- 删 `apps/lina-core/manifest/sql/013-auth-external-identity.sql`。
- 删 DAO/DO/Entity：`internal/dao/sys_user_external_identity.go`、`internal/dao/internal/sys_user_external_identity.go`、`internal/model/do/sys_user_external_identity.go`、`internal/model/entity/sys_user_external_identity.go`。
- 从 `apps/lina-core/hack/config.yaml` DAO 生成清单移除 `sys_user_external_identity`。
- 改造 `auth_external_identity_test.go`、`auth_external_provision_test.go` 为 provider-mock 驱动。

### 6 — google/discord 依赖

- `linapro-oidc-google`、`linapro-oidc-discord` 的 `plugin.yaml` 声明依赖 `linapro-extlogin-core`。
- 确认两者登录路径不变（仍走 `extlogin` seam），core 未启用时 external login fail-closed。

### 7 — 测试

- 宿主单测：provider 缺失 fail-closed、provider ownership 拒绝、resolve 命中后走 token 铸造（mock provider）。
- 插件单测：resolve/provision/无邮箱 anchor 开户/邮箱冲突/绑定越权/重复绑定/列举自隔离。
- 并发单测：并发未链接登录只产生一条有效链接、链接写入失败后确定性 anchor 幂等复用、`(provider,subject)` 唯一索引冲突不冒泡 500。
- 生命周期单测：跨 provider 同 subject 不交叉；禁用保留链接、重启恢复；卸载删表但不级联删除 `sys_user` 孤儿账号。
- 三插件 `go build ./...` / `go vet ./...`；i18n JSON 校验。
- 端到端：装 core+google，走 One Tap/授权码登录成功；卸载 core 后 external login fail-closed。

### 8 — 文档

- `apps/lina-core/pkg/plugin/README.md` + `README.zh-CN.md`：Auth 域 external-login + 新 provider SPI 边界 + fail-closed。**注意**：现有 README 文本还写着 "host-owned sys_user_external_identity linkages"，需改为经插件 provider 解析。
- `apps/lina-plugins/README.md` 插件清单加 `linapro-extlogin-core`。
- 新插件 `README.md` + `README.zh-CN.md` 双语。

## Momus 审查确认的关键契约（实现时必须遵守）

1. **去重锚点 = `(provider,subject)` 唯一索引**，绝不假设 `sys_user` email 唯一（核对 `manifest/sql/001-user-auth-bootstrap.sql`：仅 `uk_sys_user_username` 唯一，`email` 无唯一索引且默认 `''`）。
2. **无邮箱 anchor 派生必须确定性 + 碰撞抗性**（`resolveProvisionUsername` 的数字后缀去重会破坏复用，anchor 路径已用独立函数 `resolveProvisionUsernameFromAnchor` 绕开）。
3. **正确性保证 = "同一 `(provider,subject)` 最终只有一条有效链接、指向一个账号"**，而非"只建一个 sys_user"（并发竞态下允许未链接孤儿，按卸载不级联删除规则容忍）。
4. **token/session/tenant 铸造留宿主**，SPI 契约不含任何铸造方法。
5. **disable 保留表 / uninstall 删表但不级联删 sys_user**。

## 恢复 / 验收命令

```
# 复验当前可编译断点
cd D:\work\linapro\apps\lina-core; rtk go build ./...

# 查看已改动文件
rtk git -C D:\work\linapro status --short -- apps/lina-core

# openspec 状态
openspec status --change refactor-external-identity-to-provider-plugin
openspec validate refactor-external-identity-to-provider-plugin

# 最终验收（全部完成后）
cd D:\work\linapro\apps\lina-core; rtk go build ./...; rtk go vet ./...
# 各插件目录：go build ./... / go vet ./... / i18n JSON 校验
# 跑 lina-review 技能审查整个变更集
```

## 约束提醒（所有会话通用）

- 每组做完必须 build 通过再进下一组；组 2 必须先于组 3；绝不留破损态。
- 禁止类型错误抑制、禁止删失败测试凑绿。
- 修改插件目录前先查该插件根是否有 `AGENTS.md` 并遵守。
- 命中的规则文件（plugin/backend-go/database/data-permission/api-contract/i18n/architecture）实现前必须读，不能凭记忆。
- 用 `rtk` 前缀跑 git/go；插件内代码生成用 `make dao`/`make ctrl`；Windows/pwsh 环境。
- 逐项勾选 tasks.md 的 `[x]`；全部完成后重写本 HANDOFF 为最终态；未经明确要求不 commit。
