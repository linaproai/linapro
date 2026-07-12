# Design

## 边界拆分：宿主 vs 插件

`LoginByExternalIdentity` 中 token 铸造、会话、租户解析、pre-token、登录 hook、IP 黑名单与禁用账号检查属核心 auth（~90%），不可下放。可下放的是 `(provider, subject)` 链接存储与开户/绑定策略。

宿主保留：`extlogin.Service`、`extidspi` manager-backed lazy provider、`usercap.ProvisionExternal`（含 `UsernameAnchor`）、ownership 盖章。插件 `linapro-extlogin-core` 拥有链接表与领域编排。

## Provider SPI 与注入

契约：`Resolve` / `Provision` / `Bind` / `Unbind` / `List`，仅返回 userID 与链接语义，不含 token/session。

注入必须照抄 `tenantspi`/`orgspi`：**manager + factory + IsProviderEnabled 惰性解析**，禁止 raw provider 直推。插件声明期 `ProvideExternalIdentityProvider(factory)`；`ProvideExternalIdentity(providerID)` 字符串 ownership 与引擎 factory 正交。无启用 provider 时 fail-closed。

## 链接表与开户

- 表名 `plugin_linapro_extlogin_core_user_external_identity`（`plugin_<plugin-id-underscored>_*` 前缀）；`(provider, subject)` 部分唯一索引（软删释放键）。
- 建号在宿主、策略在插件；先建号再写链接，权威去重为链接唯一索引；冲突捕获不冒泡 500；无邮箱 anchor 确定性派生。
- 邮箱冲突防接管：宿主铸号原语未过滤 email 计数 + 哨兵错误，策略语义归插件。
- 禁用保留表；卸载 DROP 表不级联删 `sys_user` 孤儿账号。

## managed 领域与 extidcap

最终 `distribution: managed`：宿主启动不强制安装 core；未装/未启用 → 外部登录 fail-closed、UI 隐藏，密码登录不受影响。

三层契约：

1. Host `ExternalLogin`：已验证身份 → 会话。
2. Host `extidspi`：登录路径 Resolve/Provision。
3. Plugin-owned `extidcap`：宽入口 `Service` 聚合 Ticket/Login/Linkage/Providers；Catalog/Handoff 独立门面；绑定仅 ticket。

VerifiedIdentity 扩展 SubjectKind、SecondarySubjects、AppContext、phone 等；权威键仍 `(provider, subject)`。

## Handoff

协议插件拿到宿主 `LoginOutput` 后 `CreateLoginHandoffFromHost`；core 持有一次性码；SPA `POST .../handoff/exchange` 兑换。宿主不暴露 handoff HTTP。插件 API 路径仅业务相对路径，不嵌套冗余 `/plugins/{pluginId}/`。错误回跳用安全文案。多实例 handoff 须走共享 cache（与会话同类），否则文档注明单实例限制。

## 协议插件约定

硬依赖 core；只做 IdP 验签与 `ProvideExternalIdentity`；不持有链接表；登录入口随启停隐藏。

## 登录页展示

- `auth.login.after`：通用 OIDC/LDAP 等 → 全宽纵向按钮。
- `auth.login.social`：Google/Discord 等平台账号 → 「其他登录方式」+ 横向圆形图标。
两区域无注入时各自整块隐藏。

## 演进说明

早期将 core 标为 builtin 以保证 google/discord 依赖可满足；领域完整化后改为 managed，与「第三方登录非主框架关键领域」一致。裸 Bind 与 JWT 进 URL 废弃为 ticket + handoff。动态插件早期对外部登录 fail-closed stub；同权同信原则下，外部登录 host service 发布见插件框架归档中的 dynamic-external-login / plugin-trust-parity。

## 数据权限

链接为用户自隔离：登录走唯一键统一 not-provisioned；Bind/Unbind/List 仅当前会话用户。

## 风险

- 跨模块无单 TX：靠唯一索引 + 幂等 anchor。
- managed 后需先装 core 再装协议插件：依赖治理与 README 安装顺序。
