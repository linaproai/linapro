## Why

外部身份登录已从宿主抽离到 `linapro-extid-core`，但能力面仍按 Google/Discord 当前用法裁剪，且 core 被标为 `builtin`，与「第三方登录非主框架关键领域、应按需安装并降级」的定位冲突。需要把 `linapro-extid-core` 建成可扩展的第三方登录地基：完整领域契约、证明链绑定、安全 handoff、managed 分发与未安装降级，以便后续微信/QQ/抖音等协议插件只做验签并依赖 core。

## What Changes

- **BREAKING（未发布分支）**：`linapro-extid-core` 的 `distribution` 由 `builtin` 改为 `managed`；宿主启动不再强制安装/启用该插件。
- **BREAKING（未发布分支）**：裸 `BindIdentity(provider, subject)` 废弃；绑定仅接受已验证 ticket（`BindByTicket`）。
- **BREAKING（未发布分支）**：协议插件回调不再把 access/refresh token 放入 SPA 回跳 URL；改为一次性 `handoff` 码，由宿主交换会话结果。
- 在 `linapro-extid-core` 发布 plugin-owned 领域能力 `extidcap`（完整接口面：ticket、编排、链接、目录、策略、档案同步等），未实现方法返回明确 not-supported，接口一次到位。
- 扩展 `VerifiedIdentity` 模型：`SubjectKind`、`SecondarySubjects`、`AppContext`、手机号等，支撑微信/QQ 等非扁平 subject 场景。
- 宿主保留薄 `extlogin` + SPI；新增 handoff 交换与（可选）ticket 登录入口；provider 缺失 fail-closed 与 UI 隐藏规格写清。
- 协议插件（google/discord）显式依赖 `linapro-extid-core`，登录/回跳走 handoff；依赖未满足时不可启用或入口隐藏。
- Vben 登录页改为消费 handoff 交换结果；无第三方能力时隐藏相关区域。

## Capabilities

### New Capabilities

- `external-identity-domain`: `linapro-extid-core` 作为 managed 外部身份领域 owner 的完整能力契约、存储、ticket、目录、策略与降级语义。
- `external-login-handoff`: 宿主一次性登录结果 handoff 与 SPA 交换契约，禁止 JWT 进 URL。
- `external-identity-protocol-plugins`: 协议插件依赖 core、provider 归属、槽位降级与扩展约定。

### Modified Capabilities

- （无已归档 baseline 修改；本变更建立在活跃 `refactor-external-identity-to-provider-plugin` 工作区实现之上。）

## Impact

- **宿主** `apps/lina-core`：auth handoff、extlogin 扩展、SPI/装配文档、i18n 错误码。
- **插件** `linapro-extid-core`：managed、cap 契约、表结构扩展、绑定 API、catalog。
- **插件** `linapro-oidc-google` / `linapro-oidc-discord`：handoff 回跳、依赖与 README。
- **前端** `apps/lina-vben`：login handoff 交换、文案。
- **依赖治理**：协议插件强依赖 core；未装 core 时第三方登录整体不可用并隐藏。
