## Context

工作区已具备：宿主 `extlogin` + `extidspi`、插件 `linapro-extid-core`（链接表与 Resolve/Provision/Bind）、google/discord 协议插件、Vben `completeExternalLogin`。缺口：core 误标 builtin、Bind 无证明、token 进 URL、领域接口按现状裁剪、无 Provider 目录/统一编排模型。

## Goals / Non-Goals

**Goals:**

- 插件 ID 固定 `linapro-extid-core`；`distribution: managed`；未装/未启用 fail-closed + UI 隐藏。
- 协议插件独立，`dependencies.plugins` 硬依赖 `linapro-extid-core`。
- 完整领域 cap `extidcap`：ticket、LoginPrepare、链接生命周期、catalog、policy、profile sync、admin 投影；接口全、实现分优先级。
- 绑定仅 ticket；登录回跳仅 handoff code。
- 宿主仍独占 token/session 铸造。

**Non-Goals:**

- 不实现微信/QQ/抖音协议插件本身。
- 不向动态 WASM 开放 ExternalLogin / CreateFromExternal。
- 不把链接表迁回宿主。
- 不强制改插件 ID 命名（保留 oidc-core 作为领域 owner ID）。

## Decisions

### D1. 三层契约

1. **Host `authcap.ExternalLogin`**：已验证身份 → 会话；新增 handoff 创建/交换（或并列 auth API）。
2. **Host `extidspi`**：宿主登录路径用 Resolve/Provision（可委托 LoginPrepare）；manager + enablement 不变。
3. **Plugin-owned `extidcap`**（`linapro-extid-core/backend/cap/extidcap`）：完整领域面；协议插件与 HTTP API 消费。

### D2. managed + 降级

- core 不由宿主 bootstrap 安装。
- SPI 无启用 provider → Resolve found=false / Provision unavailable → 登录 fail-closed。
- 协议插件依赖未满足 → 安装/启用拦截。
- 前端：无协议插件或依赖失败 → 不渲染第三方登录入口；绑定区隐藏。

### D3. Verified ticket

- 协议插件验签成功后 `IssueVerifiedTicket`；TTL 短、单次消费。
- `BindByTicket` / `LoginPrepare` 只吃 ticket 或已盖章 VerifiedIdentity（服务端路径）。
- 公开 HTTP 禁止裸 provider+subject 绑定。

### D4. Login handoff（闭环在 linapro-extid-core）

- 宿主仅铸造会话（`LoginByVerifiedIdentity`）；**不**暴露 handoff HTTP。
- 协议插件拿到 `LoginOutput` 后调用 `extidcap.CreateLoginHandoffFromHost`；`linapro-extid-core` 持有一次性码。
- SPA 调用 core 公开 API `POST /x/linapro-extid-core/api/v1/handoff/exchange` 兑换（插件 `g.Meta path` 仅声明业务相对路径 `/handoff/exchange`，不得再嵌套 `/plugins/{pluginId}/`）。
- 错误回跳仅安全错误码/本地化消息，禁止 `err.Error()` 原文。

### D5. VerifiedIdentity 扩展字段

权威键仍为 `(provider, subject)`；`SubjectKind`、`SecondarySubjects`、`AppContext`、phone 等进入 ticket 与链接快照，支撑微信 unionid/openid。链接表增加可选列或 JSON meta（有界）。

### D6. Provider catalog

- 协议插件声明期 `RegisterProvider(ProviderDescriptor)`（进程内 registry，随插件 enablement 过滤）。
- `ListProviders` 供登录页/个人中心；无 core 时前端不调用、入口隐藏。

### D7. 宽入口 + 子面隔离，实现分波

- `extidcap.Service` 为宽入口，通过 `Ticket()` / `Login()` / `Linkage()` / `Providers()` 暴露角色子面（对齐宿主 `orgcap`/`authcap` 聚合风格）；**不**使用插件可见 `AdminService` 目录表达风险边界。
- Catalog 注册与 SPA handoff 保持独立 `CatalogService` / `HandoffService` 进程绑定门面，不并入 fat `Service` 方法表。
- 删除可推导 wrapper 与未接线占位：`PreviewLogin`（用 `LoginPrepare`+`DryRun`）、`AdminListByUser`（复用 `ListByUser`）、`RebindByTicket` / `AdminForceUnbind`（需要时再加）。
- Wave A（本变更必达）：ticket、LoginPrepare、BindByTicket、Unbind、List、GetLinkage、catalog 注册/列表、handoff、managed、协议插件与 Vben 切换。
- Wave B（需要时再扩展子面）：Rebind、强制解绑（authz 在调用侧）、ResolveBatch 增强等。

## Risks / Trade-offs

- managed 后用户需手动装 core 再装协议插件 → 依赖文案与 enable 拦截缓解。
- handoff 需共享缓存/内存；多实例部署须用 cache 后端（与会话同类）——本变更优先走宿主 cache/kv 能力若可用，否则进程内 + 文档注明单实例限制并接 cache。
- 表结构扩展对未发布库可直接 migration。

## Migration Plan

- 无生产数据负担；改 distribution、扩展字段并入 `001` 建表 SQL（无单独 002 ALTER）、切换绑定 API 与回跳协议。
- 旧裸 Bind API 删除或返回 gone。

## Open Questions

- 无阻塞项；个人中心绑定 UI 可本变更提供最小 API + 后续页面。
