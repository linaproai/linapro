## Context

动态插件请求进入 guest 前需要完成宿主侧认证、权限和数据范围快照构建。当前实现已经使用共享`session.Store.TouchOrValidate`验证 session hot state，并且 session 模块已有`sessionLastActiveUpdateWindow`和 coordination hot state 投影写回节流；因此 session 写频率不是本变更的主要新增点。

真正需要收敛的是权限 owner 边界：动态路由认证在`runtime_route_auth.go`中直接读取`sys_user_role`、`sys_role`、`sys_role_menu`和`sys_menu`，重复实现了`role`模块已经持有的 token access snapshot 和`permission-access`修订号逻辑。该路径既有性能成本，也有权限、租户和数据范围语义漂移风险。

同时，动态插件 guest 执行目前只有默认超时和单实例内存上限，缺少统一全局并发和按插件并发护栏。30 插件规模下，单插件异常流量可能占满 guest 实例资源，影响其他插件和宿主请求。

## Goals / Non-Goals

**Goals:**

- 动态路由认证只通过`role`模块发布的访问投影契约读取权限、角色名、数据范围和超管状态。
- 动态路由权限 freshness 复用`permission-access`修订号，集群不可确认时 fail-closed。
- 动态路由 session 校验继续使用启动期共享`session.Store`，不得自建 session 缓存或绕过热状态。
- 为 guest 执行增加全局并发、按插件并发、获取超时和内存页上限。
- 资源护栏和 datahost 表契约缓存保持结构化错误、数据权限和缓存一致性可审查。

**Non-Goals:**

- 不改变 JWT claims、登录、登出、强制下线或 token 撤销协议。
- 不新增动态插件可调用的 host service 方法。
- 不把当前用户菜单/按钮 RBAC 作为已授权动态领域管理方法的二次校验；领域方法仍执行数据权限和目标边界。
- 不改变 cachecap 默认 SQL 后端，也不引入 Redis 作为新必需依赖。
- 不实现新的 session last-active 降频机制；现有 session store 已覆盖该语义。

## Decisions

### 1. 访问投影由`role`模块发布，plugin runtime 只消费窄契约

`role`模块新增或收窄一个动态路由可用的访问投影方法，输入必须包含 token ID、user ID 和 tenant ID，输出为 DTO：权限字符串、角色名、数据范围、unsupported 数据范围标记和超管标记。该 DTO 由`role`模块从 token access snapshot 构建，缓存键与 revision 仍由`role`模块持有。

plugin runtime 通过构造函数显式接收该窄接口，删除对角色 DAO、DO、Entity 和私有缓存结构的直接访问。这样动态路由、宿主 API 和 host service 领域方法使用同一个权限 owner。

替代方案是在 plugin runtime 内自建权限缓存。该方案会复制`permission-access`域、租户维度和 fail-closed 策略，容易造成权限收紧后两条路径不一致；因此拒绝。

### 2. Session 只做共享校验，不新增运行时本地缓存

动态路由保留每请求 session 有效性校验。`session.Store.TouchOrValidate`负责 DB-only 或 coordination-backed hot state 的有效性、TTL 刷新和 PostgreSQL 投影节流。plugin runtime 不缓存“session 有效”结论，也不跳过登出、强制下线、用户禁用或 token 过期后的拒绝路径。

### 3. Guest 并发护栏归属 plugin runtime

新增护栏配置：

| 配置项 | 默认值 | 语义 |
|--------|--------|------|
| `plugin.runtime.guest.maxConcurrent` | `max(32, 4×CPU)` | 同时存活 guest 执行全局上限 |
| `plugin.runtime.guest.maxConcurrentPerPlugin` | `0` | 每插件并发上限，`0`表示不限 |
| `plugin.runtime.guest.acquireTimeout` | `3s` | 等待执行许可的最长时间 |
| `plugin.runtime.guest.memoryLimitPages` | `4096` | 单 guest 实例内存页上限，保持 256 MiB 默认 |

配置读取层必须解析`time.Duration`字符串，错误通过启动校验或运行时配置更新校验暴露。超限等待超时返回稳定业务错误，不进入 guest。

### 4. Host call 请求内授权快照复用

同一次`ExecuteBridge`中，host call 授权快照应构建一次并挂在 host call context 中。该优化不改变授权来源：仍以当前 active release 的宿主确认授权快照为准；每次 host call 仍校验 service、method 和资源标识。

### 5. Datahost 表契约缓存后置但可纳入本变更

datahost 当前每次构建授权表契约都会清理并读取表字段。可以按插件、表名和插件迁移版本缓存表契约。权威源是当前数据库 schema 和插件迁移账本；插件 install、upgrade、uninstall SQL 成功提交后必须按插件失效。缓存命中不得放宽字段白名单、数据权限、分页、排序、软删除和审计治理。

## Risks / Trade-offs

- `role`访问投影契约过宽 → 只返回动态路由身份快照和 host service 上下文需要的稳定 DTO，不暴露 role DAO、Entity 或私有缓存条目。
- 权限 freshness 不可确认时影响可用性 → 采用与宿主受保护 API 一致的 fail-closed 策略，安全优先。
- 并发护栏默认值过低影响吞吐 → 默认值按 CPU 和固定下限计算，允许运行时配置调整，并通过指标或日志暴露排队与超时。
- Datahost 表契约缓存遗漏 DDL 变化 → 只在插件生命周期 SQL 成功提交后缓存失效；缓存未命中或失效失败时回源读取 live schema。
- 错误码和配置文案引入`i18n`工作 → 按宿主`i18n`规则维护英文源文本、非英文翻译和 apidoc/README 说明，避免裸中文错误。

## Migration Plan

先在`role`模块发布窄访问投影并迁移动动态路由认证，再增加 guest 并发护栏和稳定错误码，最后优化 host call 授权快照复用与 datahost 表契约缓存。每一步都保留原安全边界：权限不可确认时拒绝，session 不可确认时拒绝，host service 授权失败时拒绝，datahost schema 缓存缺失时回源。

## Open Questions

- datahost 表契约缓存是否必须在本批落地，还是在权限快照和 guest 护栏完成后作为后续优化。当前任务把它列为后置子任务，实施时可基于使用频率裁剪。
- guest 护栏配置是走现有运行时配置服务还是仅静态配置启动生效。建议优先复用现有运行时配置和`runtime-config`修订号，但如果实现成本超出本批，可先静态启动配置并在任务记录说明后续动态化计划。
