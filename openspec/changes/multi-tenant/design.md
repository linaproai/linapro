## Context

LinaPro 当前架构以"单租户 + 单一管理员"为前提:`sys_user`、`sys_role`、`sys_*` 业务表无租户维度;`bizctx` 仅承载用户身份与 locale;数据权限通过 `datascope`(角色 data_scope)与 `orgcap`(部门树)叠加;插件治理为单层(平台管理员 install + 启用,影响所有人)。

要支持"公司内部多 BU(初期)+ 未来可演化为 SaaS 硬隔离"的多租户形态,核心约束有四:

1. **不分期、一次完整落地**(用户明确要求):项目无历史负担,Pool 模型一次到位代价更小。
2. **尽可能插件化**:租户主体、成员、生命周期等高变更频率部分由 `multi-tenant` 插件持有;宿主仅保留最薄的能力接缝。
3. **保持单租户开箱可用**:不安装 `multi-tenant` 插件时,系统行为与今天等价(所有数据视为 `tenant_id=0` 的 PLATFORM 租户)。
4. **复用现有"稳定能力接缝"模式**:`pkg/orgcap` + `internal/service/orgcap` + plugin Provider 的三段式范式已被验证,本设计直接复用其形状。

stakeholders:
- 框架核心开发者(本仓库维护者)
- 框架使用方运维(平台管理员)与租户管理员(下游)
- 插件作者(需理解 `scope_nature` 与 `LifecycleGuard` 契约)
- AI 协作研发流程(需通过 OpenSpec 治理保证 spec 一致性)

## Goals / Non-Goals

**Goals:**

- 在不破坏单租户开箱体验的前提下,提供完整的多租户能力(schema、解析、JWT/会话、缓存、文件、审计、插件治理、UI)。
- 把"是否启用多租户、如何解析租户、租户主体管理、用户-租户绑定、租户配置覆盖、租户生命周期事件"等所有可变政策封装在 `multi-tenant` 插件内。
- 把"插件治理"升级为"平台管理员 install + 租户管理员 enable"的两层模型,并通过 `scope_nature` + `install_mode` 把插件作用域显式化。
- 把"卸载/禁用前置校验"从宿主硬编码迁移为插件自检(`LifecycleGuard` 否决型钩子),允许平台管理员紧急 `--force` 但强审计。
- 在 1:N membership 模型下提供平台管理员 vs 租户管理员的角色分层、跨租户 impersonation 与可观察性。
- 给字典、配置等"平台默认 + 租户覆盖"语义的资源提供统一的 fallback 读路径辅助方法。
- 全面接入 i18n(zh-CN / en-US)、数据权限(datascope)、缓存一致性(cluster-aware 失效)、审计(loginlog/operlog 双轨)。

**Non-Goals:**

- 不实现 schema-per-tenant 或 DB-per-tenant 隔离模式(预留 `tenant.isolation.mode` 配置占位,值仅 `pool`)。
- 不实现"按租户启用插件 UI 的批量平台面板"高级形态;租户管理员单租户内启用即可,平台管理员的批量操作留作未来扩展。
- 不实现"租户级配额/计费"逻辑;只预留 `plugin_multi_tenant_quota` 表与扩展点,默认无限制。
- 不实现"租户级品牌定制"(logo、配色、域名 SSL 等);只在 schema 层为 `tenant_branding` 留扩展位。
- 不引入分布式事务、跨租户数据迁移、租户合并/分裂等复杂运维操作。
- 不破坏现有"单租户开箱体验":未安装 multi-tenant 插件时,所有 sys_*.tenant_id 默认 0,所有过滤逻辑 no-op,所有现有 e2e 用例不修改也能通过。

## Decisions

### 1. 隔离模型:Pool(单库 + tenant_id 列)

**选择**:所有 `sys_*` 与 `plugin_*` 业务表加 `tenant_id INT NOT NULL DEFAULT 0`,索引升级为 `(tenant_id, ...)` 联合索引。`sys_user` 在 1:N 模式下以 membership 作为用户可见性权威边界,`sys_user.tenant_id` 仅表示主租户/默认登录租户。

**理由**:
- 与"插件主导"目标自洽(其他模式需要宿主级连接路由,无法插件化)。
- Pool 是 Q1 软隔离的标准答案,且 PG 对 (tenant_id, x) 联合索引性能稳定。
- 未来要升级为 schema-per-tenant 时,通过中间件层加 `SET search_path` 即可,业务代码不变。

**替代方案**:
- schema-per-tenant:连接路由侵入宿主,放弃插件主导。
- DB-per-tenant:连接池 + DSN 路由,工程量与运维复杂度过高,首版不做。

**配置位**:`tenant.isolation.mode: pool`(枚举,首版仅 `pool`)。

### 2. 用户-租户基数:1:N membership(默认),保留 1:1 配置开关

**选择**:全局身份(`sys_user.username` 全局唯一)+ `plugin_multi_tenant_user_membership` 多对多绑定。配置项 `tenant.cardinality: multi|single`,默认 `multi`。

**理由**:
- 内部 BU 场景下"同一员工服务多个 BU"是常态。
- 1:N 是 1:1 的真超集;`single` 模式下绑定第二条 membership 时被拒即可。
- Slack/Notion/GitHub Orgs/Atlassian 都采用此模型;Salesforce 历史 1:1 已是行业反例。

**Schema**:
```
sys_user
  + tenant_id INT NOT NULL DEFAULT 0
    -- 在 multi 模式下为"主租户/默认登录租户"
    -- 在 single 模式下为权威唯一租户
    -- 0 = PLATFORM 用户(平台管理员或单租户模式下的所有用户)

plugin_multi_tenant_user_membership
  id, user_id, tenant_id,
  is_tenant_admin BOOL,
  status SMALLINT,
  joined_at, last_active_at
  UNIQUE (user_id, tenant_id)
```

**替代方案**:
- 严格 1:1:简单但限制大;首版直接做但允许配置切回。
- N:M + 租户内 profile(Slack 形态):本次不做,留扩展位 `plugin_multi_tenant_user_profile`。

### 3. 平台管理员 vs 租户管理员的角色分层

**选择**:`sys_role` 增加 `is_platform_role BOOL` 与 `tenant_id INT`;`sys_user_role` 增加 `tenant_id INT`,UNIQUE `(user_id, role_id, tenant_id)`。

**语义**:
- `is_platform_role=true` 的角色:`tenant_id=0`,仅可分配给 `tenant_id=0` 的用户(平台管理员候选);权限点以 `platform:*` 开头。
- `is_platform_role=false` 的角色:`tenant_id>0`,仅在所属租户内可见与可分配。
- 平台管理员在请求中:`bizctx.TenantId=0` 表示"管理平台",`bizctx.TenantId=T` 表示"代某租户操作"(impersonation)。
- impersonation 状态下,操作日志双轨记录:`acting_user_id=平台管理员` + `on_behalf_of_tenant_id=T`。

**bypass 规则**:
- 仅管理平台模式(`bizctx.TenantId=0` 且持有平台角色)全量 bypass `tenantcap.Apply` 与 `datascope`。
- 平台管理员 impersonation 某租户时(`bizctx.TenantId=T`,`bizctx.ActingAsTenant=true`)不全量 bypass,而是按租户 T 的普通租户视图注入过滤;区别仅体现在审计字段 `acting_user_id` / `on_behalf_of_tenant_id`。
- 平台管理员跨租户读仅允许通过显式 `/platform/*` API;平台管理员跨租户写必须通过 impersonation 或专用平台 API,且必须显式指定目标 `tenant_id` 并记审计。

**默认 admin 用户的迁移**:
- `admin` 用户 → `tenant_id=0`,绑定新增的"平台超级管理员"角色(`is_platform_role=true`,permissions=`['*']`)。
- 历史"超级管理员"角色保留为平台角色;新建租户后,自动 provision 一个"租户管理员"模板角色(`tenant_id=新租户`,`is_tenant_admin=true`)。

### 4. bizctx 与租户解析中间件

**选择**:`bizctx.Context` 增加 `TenantId int` 字段;新增 `tenancy` 中间件位于 `auth` 之后、业务处理之前。

**解析责任链**(顺序由配置决定,默认 `[override, jwt, session, header, subdomain, default]`):

```
override     X-Tenant-Override (仅平台管理员)
jwt          Claims.TenantId
session      Session 中持久化的当前租户
header       X-Tenant-Code(仅登录前/无正式 JWT 的请求作为租户 hint)
subdomain    {code}.{root_domain}(仅登录前/无正式 JWT 的请求作为租户 hint)
default      用户的主租户(tenant_id=0 或 membership 第一条)
```

每个解析器实现独立接口 `tenancy.Resolver`,在 plugin 启动时注册到 `tenantcap.Service`。配置在 `config.yaml` 的 `tenant.resolution.*` 段,运行时优先读 `plugin_multi_tenant_resolver_config` 表(管理后台可改)。

**TenantId 优先级固定策略**:
- 平台管理员 `X-Tenant-Override` 最高优先级,但只允许平台管理员且必须进入 impersonation 审计语义。
- 已签发正式 JWT 的业务请求以 `Claims.TenantId` 为权威租户身份;普通用户不得通过 `X-Tenant-Code`、子域名或 session 覆盖 JWT 中的租户。
- session 仅作为服务端持续会话校验与无 JWT 内部链路的兜底,不得覆盖正式 JWT。
- header/subdomain 只作为登录前或 `pre_token` 阶段的租户 hint,用于缩小候选租户或自动选择;如果与用户 membership 不匹配则拒绝。
- default 只在没有 override、JWT、session、header/subdomain hint 时生效;在 1:N 且无主租户可判定时按 `on_ambiguous` 返回 `TENANT_REQUIRED`。
- 切换租户必须调用 `/auth/switch-tenant` 重签 token 并撤销旧 token,禁止通过 header/subdomain 让旧 token 临时切到其他租户。

**未识别请求行为**(`tenant.resolution.on_ambiguous`):
- `prompt`(默认):返回 `TENANT_REQUIRED` 错误码,前端弹挑选器。
- `reject`:401 拒绝。
- `first_owned`:自动选 membership 第一条(便利但不推荐)。

**保留子域名**:`tenant.resolution.subdomain.reserved: [www, api, admin, static, docs]`;命中保留子域时不解析为租户,等同 PLATFORM。

**multi-tenant 插件未安装时**:`tenancy` 中间件存在但走"短路":直接写 `bizctx.TenantId=0` 并放行,所有 `tenantcap.Apply` 退化为 no-op。

**理由**:
- 责任链 + Provider 模式让宿主与插件解耦;插件提供具体 Resolver 实现。
- 配置文件 + 后台管理双源,后台优先 + 配置兜底是行业惯例(Auth0/Okta)。

### 5. tenantcap 接缝与 DAO 注入纪律

**新增**:
- `pkg/tenantcap/`:Provider 接口、注册函数、TenantID 类型、PLATFORM 常量。
- `internal/service/tenantcap/`:`Service` 接口、`Apply(model, col)`、`Current(ctx)`、`Enabled(ctx)`、`EnsureTenantVisible(ctx, tenantID)`、`PlatformBypass(ctx)`(平台管理员判断)。

**Provider 接口**:
```go
type Provider interface {
    ResolveTenant(ctx context.Context, r *ghttp.Request) (TenantID, error)
    ValidateUserInTenant(ctx context.Context, userID int, tenantID TenantID) error
    ListUserTenants(ctx context.Context, userID int) ([]TenantInfo, error)
    SwitchTenant(ctx context.Context, userID int, target TenantID) error
    OnTenantLifecycleEvent(ctx context.Context, evt LifecycleEvent) // optional
}
```

**DAO 注入纪律**:
- 凡读 `sys_*` 或 `plugin_*` 业务表的 host service,必须经 `tenantcap.Apply(ctx, model, "tenant_id")` 包装。
- 与 `datascope.ApplyUserScope` 叠加时,先 tenantcap 后 datascope(租户隔离优先级最高)。
- `Apply` 在 multi-tenant 未启用时是 no-op,启用后注入 `WHERE tenant_id = ?`。
- 平台管理员管理平台上下文(`TenantId=0` 且 `PlatformBypass(ctx) = true`)时跳过注入;平台管理员 impersonation 某租户时仍注入目标租户过滤。
- 写操作:由 service 层在构造 DO 时填 `tenant_id = bizctx.TenantId`;平台管理员显式写时使用专用 service 方法(避免误写到 PLATFORM)。

**理由**:
- 与现有 `datascope.ApplyUserScope` 调用形态一致,开发者无新认知负担。
- 把"是否启用 + 是否 bypass"集中在 Service 层,DAO 调用点不需要写条件判断。

### 6. JWT/Session 的租户绑定

**选择**:
- JWT Claims 增加 `TenantId int`(0 = 平台).
- 会话存储主键变 `(tenant_id, token_id)`,index `(user_id, tenant_id)`。
- 切换租户:调用 `/auth/switch-tenant`,服务端校验 membership → 删除旧 session → 重签 token → 返回新 token。旧 token 被加入 revoke 列表(短期 + cluster 广播)。
- 登录时若用户仅 1 个 membership 且 tenant.cardinality=multi 也是 1 个,自动选中;否则按解析策略走。

**理由**:
- token 中带 tenant_id 是行业惯例(Auth0 Organizations、Cognito)。
- 重签 + 旧 token 立即失效防止"老 token 持平台权限访问租户接口"逃逸。

### 7. 字典/配置的 platform fallback

**读路径**(封装在 `tenantcap.ReadWithPlatformFallback`):
```sql
SELECT * FROM sys_dict_data WHERE dict_type = ? AND tenant_id = current_tenant
UNION ALL
SELECT * FROM sys_dict_data WHERE dict_type = ? AND tenant_id = 0
  AND NOT EXISTS (SELECT 1 FROM sys_dict_data WHERE dict_type = ? AND tenant_id = current_tenant);
```

或应用层两次查询 + 合并,以避免数据库特有语法:
1. 先查 `tenant_id = current_tenant`,如有结果直接返回。
2. 否则 fallback 查 `tenant_id = 0`。

**写路径**:
- 默认写 `bizctx.TenantId`(租户管理员只能写自己租户)。
- 平台管理员显式调用 `WritePlatformDefault(...)` 才能写 `tenant_id=0`。
- 字典类型的"是否允许租户覆盖"通过 `sys_dict_type.allow_tenant_override BOOL` 控制;false 时仅平台管理员可改,所有租户共享。

**适用范围**:
- `sys_dict_type` / `sys_dict_data` / `sys_config`:platform fallback。
- `sys_role`:平台角色全局可见 + 租户角色私有(不 fallback,而是 union)。
- `sys_menu`:平台全局,不接入 tenancy。

**理由**:
- 真实 SaaS 普遍需要"平台默认 + 租户覆盖",硬选一种会失去灵活性。
- 把 fallback 模式封装到 helper,DAO 调用点保持简单。

### 8. 插件治理两层化与 scope_nature

**plugin.yaml 新增字段**:
```yaml
id: <plugin-id>
scope_nature: platform_only | tenant_aware  # 必填
default_install_mode: global | tenant_scoped # 可选,scope_nature=tenant_aware 时生效,默认 tenant_scoped
default_for_new_tenants: true | false        # 可选,install_mode=global 自动启用;tenant_scoped 时控制是否自动开启
```

**sys_plugin 新增列**:
- `scope_nature VARCHAR(32)`:安装时从 plugin.yaml 拷贝,运行时不可改。
- `install_mode VARCHAR(32)`:platform_only 强制 `global`;tenant_aware 由平台管理员选,可在严格条件下切换。

**sys_plugin_state 改造**:
- 增加 `tenant_id INT NOT NULL DEFAULT 0`。
- 主键 `(plugin_id, tenant_id)`。
- `tenant_id=0`:平台级开关(对 platform_only 与 install_mode=global 适用)。
- `tenant_id>0`:租户级开关(对 install_mode=tenant_scoped 适用)。

**IsEnabled 解析**(伪码):
```go
func IsEnabled(ctx, pluginID) bool {
    p := lookupPlugin(pluginID)
    if !p.installed { return false }
    if p.scope_nature == "platform_only" || p.install_mode == "global" {
        return state(pluginID, 0).enabled
    }
    // tenant_scoped
    tid := bizctx.TenantId(ctx)
    if tid == 0 { return false }
    return state(pluginID, tid).enabled
}
```

**install_mode 切换规则**:
| 切换方向 | 是否允许 | 处理 |
|----------|----------|------|
| `global → tenant_scoped` | ✅ | 自动为所有 active 租户初始化 enabled=current global state |
| `tenant_scoped → global` | ⚠️ | 仅平台管理员可执行;二次确认 + 强制审计;立即对所有租户强制启用 |
| `scope_nature` 变更 | ❌ | 仅插件升级版本时随 manifest 变更,需迁移脚本 |

**租户管理员可见的插件清单**:仅 `scope_nature=tenant_aware` 且 `install_mode=tenant_scoped` 的已安装插件。

**理由**:
- 与 Atlassian Forge / Slack Apps 的两层治理对齐。
- `scope_nature` 写入 plugin.yaml 让插件作者承担"作用域天性"判断的责任,而不是平台管理员。

### 9. LifecycleGuard 否决型钩子

**契约**(`pkg/pluginhost/lifecycle_guard.go`):
```go
type CanUninstaller interface {
    CanUninstall(ctx context.Context) (ok bool, reason string, err error)
}
type CanDisabler interface {
    CanDisable(ctx context.Context) (ok bool, reason string, err error)
}
type CanTenantDisabler interface {
    CanTenantDisable(ctx context.Context, tenantID int) (ok bool, reason string, err error)
}
type CanTenantDeleter interface {
    CanTenantDelete(ctx context.Context, tenantID int) (ok bool, reason string, err error)
}
```

**调用与聚合**:
- 宿主在执行卸载/禁用/租户删除前,枚举所有已安装插件 → 类型断言每个 `CanXxx` 接口 → 并发调用(超时 5s/钩子)。
- 任意一个返回 `ok=false` → 整体拒绝;**仍执行所有钩子**收集理由集合(不短路)。
- `ok=true, err!=nil`:理由 = `"plugin <id> guard check failed: <err>"`,默认拒绝。
- `panic`:recover + 默认拒绝 + 记日志。
- `reason` 必须是 i18n key(参考 `framework-i18n-source-text-registry`),前端按 locale 渲染。

**`--force` 通道**:
- 平台管理员可在 UI 上选"强制卸载",必须文本输入插件 ID 二次确认。
- 强制操作单独写审计日志,记录被绕过的所有 reason 与触发用户。
- 配置 `plugin.allow_force_uninstall: true|false`(默认 true,严格合规场景可关)。

**multi-tenant 插件的具体实现**:
- `CanUninstall`:存在 `plugin_multi_tenant_tenant.status='active'` 行 → 拒绝,reason `"plugin.multi-tenant.uninstall_blocked.tenants_exist"`,params `{count}`。
- `CanDisable`(install_mode=global 适用):存在 active 租户 → 拒绝,reason `"plugin.multi-tenant.disable_blocked.tenants_exist"`。
- `CanTenantDelete`(订阅,跨插件协作):若该租户存在未结算资源 → 拒绝(本次未实现具体规则,预留接口)。

**理由**:
- 把"前置校验"从宿主硬编码迁移为插件自检,符合解耦原则。
- 聚合多否决一次性展示是 IDE 平台(Eclipse/IntelliJ)惯例,避免运维人员"修一个发现下一个"。

### 10. 缓存一致性(集群感知)

**约束**:
- 所有运行时缓存(`kvcache` / `pluginruntimecache` / 翻译缓存 / 字典缓存 / 角色缓存)key 默认携带 `tenant_id` 维度。
- `cluster.enabled=true` 时,`cachecoord` 失效消息额外携带 `tenant_id` 字段;按 `(tenant_id, scope, key)` 精细化失效。
- 平台级缓存(`tenant_id=0`)失效广播作用于所有节点的所有租户视图(因为可能影响 fallback 结果)。
- 多租户能力本身的缓存(`plugin_multi_tenant_tenant`、membership 列表)采用版本号 + 短期内存缓存 + 集群广播失效。

**集群一致性兜底**:
- 单机模式(`cluster.enabled=false`):同步刷新 + 进程内缓存,无需广播。
- 集群模式:依赖现有 `cachecoord` 与 leader-election 机制;租户解析器配置变更通过 leader 写 + 广播。

**租户解析器配置缓存**:
- 进程内 LRU(TTL 30s)+ 集群失效广播。
- 每次配置变更(`/platform/tenant/resolver-config`)立即广播失效 + bump 版本号。

**理由**:
- 命中"缓存一致性治理要求"硬性条款。
- 多租户场景下"用户切租户后角色错版本"是高发事故,版本号 + 主动失效是必备防线。

### 11. org-center 插件的协同改造

**Schema 变更**:
- `plugin_org_center_dept` / `_post` / `_user_dept` / `_user_post` 全部加 `tenant_id INT NOT NULL DEFAULT 0`。
- 索引升级为 `(tenant_id, ...)` 联合索引。
- 现有 mock 数据写入 `tenant_id=0`(PLATFORM 默认部门),`make mock` 兼容。

**事件订阅**:
- 实现 `tenantcap.LifecycleSubscriber` 接口(由 multi-tenant 插件定义),订阅 `tenant.created`:
  - 自动创建租户根部门(`name=租户.name`,`tenant_id=新租户`)。
- 订阅 `tenant.deleted`:
  - 级联删除该租户所有 dept/post/user_dept/user_post 记录。
- 订阅 `tenant.suspended`:
  - 不删数据,但 `orgcap.Provider` 在 IsSuspended 租户上下文返回空树。

**bizctx 接入**:
- 所有 dept 树构建、用户-部门解析、`ApplyUserDeptScope` 内部读 `bizctx.TenantId`,过滤本租户数据。
- 平台管理员上下文 + impersonation 的某租户 → 显示该租户的部门树。

**plugin.yaml**:
```yaml
id: org-center
scope_nature: tenant_aware
default_install_mode: global
default_for_new_tenants: true
```

理由:部门管理是企业基础能力,默认对所有租户开启更符合"开箱即用"。

### 12. 文件存储租户隔离

- 本地存储:文件路径前缀 `/storage/t/{tenant_id}/yyyy/mm/dd/...`,`tenant_id=0` 为平台共享。
- 对象存储(若启用):object key 前缀 `t/{tenant_id}/...`。
- 跨租户引用:不允许;若需共享(如平台 logo)显式写 `tenant_id=0`。
- 文件读取接口:鉴权时校验文件 `tenant_id` 与 `bizctx.TenantId` 匹配;平台管理员管理平台模式只能通过显式 `/platform/*` 只读接口跨租户访问,impersonation 时按目标租户校验。

### 13. 任务调度的租户感知

- `sys_job` / `sys_job_log` 加 `tenant_id`;内置系统任务(session 清理、监控数据归档)`tenant_id=0`,业务任务由租户管理员创建,绑定本租户。
- 任务执行时,`gcron` 触发的回调会构造一个 `bizctx`,其中 `TenantId=任务.tenant_id`,`UserId=任务.create_by`(或 0 表示系统)。
- 平台级 handler(如 `cron.session.cleanup`)注册时声明 `scope: platform`;租户级 handler 注册时声明 `scope: tenant`。
- 租户管理员只能创建/查看 `scope=tenant` 的 handler 任务;平台管理员可创建任意。

### 14. 审计与可观察性

- `monitor-loginlog` 表:加 `tenant_id`(用户登录到的租户)、`acting_user_id`(true 操作人,通常 = user_id;impersonation 时 = 平台管理员)、`on_behalf_of_tenant_id`(impersonation 时 = 目标租户)。
- `monitor-operlog` 同上,字段语义对齐。
- 查询接口:租户管理员只能看本租户日志;平台管理员可看全部 + 按 `acting_user_id` 过滤。
- 强制操作(force-uninstall、impersonation 启用、scope 切换)单独审计:`monitor-operlog.action_kind='platform_force_action'`。

### 15. i18n 与运行时翻译缓存

- 否决钩子 reason、租户管理后台、平台管理员标签、租户切换提示、登录页租户挑选 UI 全部走 i18n key。
- 运行时翻译缓存 key 加 `tenant_id` 维度(因为不同租户可能有不同字典翻译)。
- 失效作用域 `scope` 从今天的 `(locale, sector)` 升级为 `(tenant_id, locale, sector)`;invalidation 必须指明 tenant 范围,禁止全租户清空。
- `manifest/i18n/<locale>/apidoc/**.json` 不变(apidoc 走插件/宿主合并,与租户无关)。

### 16. 启动期装配与一致性校验

- `framework-bootstrap-installer` 启动时:
  1. 应用 SQL 迁移(包括 016 新文件)。
  2. 检测 `multi-tenant` 插件是否安装且 enabled → 是 → 装配 `tenantcap.Provider` → tenancy 中间件激活 → 否 → 短路模式。
  3. 校验 `sys_plugin.install_mode` 与 `scope_nature` 一致性(platform_only 必须 install_mode=global,否则启动期 panic 并提示修复)。
  4. 校验 `sys_role.is_platform_role` 与 `tenant_id` 一致性(is_platform_role=true 必须 tenant_id=0)。
- 启动期任何一致性检查失败,服务拒绝启动并打印明确报错。

### 17. e2e 测试组织

新增模块 `hack/tests/e2e/multi-tenant/`,涵盖:

- `tenant-lifecycle/`:创建、暂停、恢复、删除租户,验证 schema 与缓存一致性。
- `tenant-isolation/`:租户 A 用户不可见租户 B 数据(用户、角色、菜单、字典、配置、文件、日志)。
- `tenant-resolution/`:每种解析策略一组用例(header、subdomain mock、JWT、session、default、prompt)。
- `tenant-switching/`:1:N 用户切换租户,旧 token 失效,菜单/权限刷新。
- `platform-admin/`:平台管理员 impersonation、bypass 校验、双轨日志。
- `plugin-governance/`:install_mode 选择、tenant_scoped 启用/禁用、scope_nature 强制。
- `lifecycle-guard/`:卸载否决聚合、`--force` 流程、reason i18n 渲染。
- `tenant-config-override/`:字典/配置 platform fallback 与租户覆盖。
- `org-center-tenancy/`:org-center 插件租户感知、租户事件触发部门初始化。

每个 TC 文件遵循 `TC{NNNN}-` 命名规范。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| **改动面巨大,一次性落地易出回归** | 严格的 e2e 隔离用例 + datascope 已有的查询过滤先例 + tenancy bypass 的统一 helper(避免散点判断);所有 sys_*.tenant_id 默认 0 + Apply no-op 保证未启用 multi-tenant 时行为完全等价 |
| **DAO 注入纪律遗漏导致跨租户数据泄露** | 通过 `backend-conformance` spec 增加硬条款 + `lina-review` 审查清单中加入"DAO 必须经 tenantcap.Apply"检查;e2e 隔离用例覆盖每个查询接口的反例 |
| **缓存 key 遗漏 tenant 维度导致脏读** | 集中在 helper 函数中构造 cache key,各模块通过 helper 而非自行拼接;用例验证"租户 A 切租户 B 后看到 A 的旧数据"反例 |
| **JWT 切换租户导致 race condition** | 旧 token 立即写 revoke 列表 + cluster 广播;客户端切换流程要求"先调 switch-tenant 拿新 token,再用新 token 请求",避免双 token 并发 |
| **`scope_nature` 误标注导致插件错位** | 安装期校验:platform_only 不允许 install_mode=tenant_scoped;租户管理员看不到 platform_only 插件即使 enabled;插件作者文档明确两值语义 |
| **LifecycleGuard 钩子超时阻塞 UI** | 5s 钩子超时 + fail-safe(超时 = ok=false);并发执行所有钩子;UI 显示"插件 X 检查中..." |
| **平台管理员 `--force` 滥用** | 强制审计 + 二次确认输入插件 ID + 配置开关可关;`monitor-operlog` 单独 action_kind 便于运维核对 |
| **租户解析配置错误导致全站 401** | 配置变更前置校验(至少保留 default 解析器在链尾);后台改完后立即可视化测试一次解析(传入示例 request 看结果);业务请求中的正式 JWT 始终优先于普通 hint,避免 header/subdomain 误覆盖已认证租户 |
| **org-center 改造影响现有部门树用例** | mock 数据写入 PLATFORM(0),单租户场景下 dept 树行为与今天等价;e2e 既有用例保持通过 |
| **租户暂停/删除时插件数据残留** | 通过 `CanTenantDelete` 钩子要求所有持有租户数据的插件参与确认;tenant.deleted 事件触发各插件级联清理;cleanup 失败需重试(幂等) |
| **多租户性能下降(联合索引未命中)** | 所有原索引升级为 `(tenant_id, ...)`;用 PG `EXPLAIN` 验证关键查询;性能审计纳入 `lina-perf-audit` 后续轮次 |
| **i18n 资源缺失导致否决理由无法本地化** | i18n 翻译完整性校验测试覆盖租户管理 + LifecycleGuard reason key;CI 阻断缺失翻译 |
| **登录解析"prompt"模式增加额外往返** | 默认行为提供清晰交互(挑选器),且仅在 1:N 用户首次登录时出现;后续走 session 短路 |
| **平台/租户角色权限模型复杂度上升** | 引入 `is_platform_role` 单一开关 + 权限点前缀约定(`platform:*`)简化心智模型;UI 严格隔离平台/租户视图 |
| **未来从 Pool 升 schema-per-tenant 的迁移成本** | 配置项 `tenant.isolation.mode` 已预留;tenancy 中间件已是统一注入点,届时改 mode 实现即可,业务代码不动 |

## Migration Plan

项目无历史负担,采用**重置数据库 + 按新 schema 初始化**:

1. **新增 SQL 文件** `apps/lina-core/manifest/sql/016-multi-tenant-and-plugin-governance.sql`:
   - 所有 `sys_*` 表 `ALTER TABLE ADD COLUMN IF NOT EXISTS tenant_id` + 默认值 0 + 重建索引。
   - `sys_role` 加 `is_platform_role`;`sys_user_role` 加 `tenant_id`。
   - `sys_plugin` 加 `scope_nature` / `install_mode`;`sys_plugin_state` 加 `tenant_id` + 调整主键。
   - seed:平台超级管理员角色、admin 用户绑定平台超管角色。
2. **新增插件 SQL** `apps/lina-plugins/multi-tenant/manifest/sql/001-multi-tenant-schema.sql`:
   - `plugin_multi_tenant_tenant`、`plugin_multi_tenant_user_membership`、`plugin_multi_tenant_config_override`、`plugin_multi_tenant_quota`、`plugin_multi_tenant_resolver_config`。
3. **org-center 插件 SQL 变更** `apps/lina-plugins/org-center/manifest/sql/001-org-center-schema.sql` 同迭代修改:
   - 所有 plugin_org_center_* 表加 `tenant_id`(项目无历史负担,直接改 001 文件)。
4. **执行 `make init`**:重建数据库;执行 `make mock` 加载演示数据。
5. **代码部署**:`tenantcap` 与中间件等核心改动随版本一起上线。
6. **回滚策略**:回退到上一个 git 标签 + 重新 `make init`(项目阶段允许重置数据库)。

## Clarified Decisions

- **租户编码(tenant code)**:只允许 ASCII 小写字母、数字与连字符,格式固定为 `[a-z0-9-]{2,32}`;不允许中文或其他 Unicode 字符。中文、英文展示名称只写入 `name`。
- **TenantId 优先级**:正式 JWT 是普通业务请求的权威租户身份;header/subdomain 仅是登录前 hint;切换租户必须通过 `/auth/switch-tenant` 重签。
- **平台 bypass 与 impersonation**:只有 `TenantId=0` 管理平台模式全量 bypass;impersonation `TenantId=T` 按 T 过滤,仅审计标识平台管理员代操作。
- **暂停租户边界**:租户成员不可读写;平台管理员可通过 `/platform/*` 执行只读排障与恢复操作。业务写入必须通过 impersonation 或专用平台 API。
- **install_mode 切回 global**:平台管理员确认后强制所有租户启用该插件,并强审计。
- **用户列表隔离边界**:1:N 模式下 membership 是租户可见性的权威边界;`sys_user.tenant_id` 仅表示主租户/默认登录租户,不能作为用户列表唯一过滤条件。
- **平台管理员的 impersonation token audience**:本次不拆分单独 audience,仍使用正式 token 结构,通过 `is_impersonation`、`acting_user_id` 与审计字段区分。
- **租户级 i18n 自定义文案**:本次不落地;留 schema 扩展位 `plugin_multi_tenant_i18n_override` 不创建。
- **平台管理员全平台聚合视图**:本次提供"租户总览 + 用户聚合 + 操作日志全量"基础页面,深度看板留扩展。
- **配额(quota)强制时机**:本次仅建表,执行点为 0(默认无限制);具体 hook 由后续迭代定义。
