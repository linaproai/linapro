## Context

LinaPro 插件平台经过多轮演进，从初始的清单契约和生命周期模型逐步扩展到动态 WASM 运行时、宿主服务授权、能力目录、领域能力模型和管理页面性能优化。Phase A-D 针对插件服务内部复杂度进行系统治理，解决 catalog 反向回调环、setter 注入链、包级可变状态、生命周期长流程滞留根门面、升级逻辑分散、运行时认证绕过权限边界、清单解析无缓存和路由文件过长等问题。

本项目定位为全新项目，不考虑旧接口兼容。所有设计决策均以一次性迁移和强治理为目标。

## 1. catalog/store/types 拆解

`catalog` 同时承担清单扫描校验、治理表读写和反向回调，成为 `catalog`、`runtime`、`integration` 之间 setter 回注环的根源。

**决策**：新建 `plugintypes` 叶子包承载纯类型和值对象；新建 `store` 作为插件治理表读写的唯一 owner；`catalog` 收窄为清单扫描、解析、校验和访问。副作用调用（menuSyncer、hookDispatcher、resourceRefSyncer）上提到编排入口。

**边界验证**：`plugintypes` 不得导入 `catalog`/`store`/`runtime`/`integration`/`dao`/`do`/`entity`；`catalog` 不得导入 `runtime`/`integration`/`dao`/`do`/`entity`；`store` 导出 API 不得泄漏 DAO/DO/Entity。

## 2. 构造函数直化与 setter 清零

`plugin.New()` 仍先构造多个半初始化 service，再通过 `Set*` 回注依赖。`runtime` 与 `integration` 互相持有完整 service，`wasm` 使用包级 `atomic.Pointer` 配置快照。

**决策**：构造函数逐项显式接收运行期依赖，删除所有 wiring setter 和 `ValidateRequiredDependencies`。使用窄契约切断 service 互持。共享状态由组合根显式创建。WASM host service 使用显式 runtime 实例。runtime revision controller 迁入 `cachecoord/revisionctrl`。合并 `capabilityhost/internal/*cap` 微包。

**关键实现**：`plugin.RuntimeDelegate` 作为组合根专用窄代理打破启动环；middleware 删除未使用的完整 `pluginsvc.Service` 字段。

## 3.0 依赖生命周期两轴模型

插件硬依赖曾用统一 `dependency.Resolver`：安装/启用共用 `CheckInstall`（只认 Installed），卸载/禁用共用 `CheckReverse`（只认已安装下游），导致「下游已全部禁用仍不能禁用 core」等反直觉行为。

**最终模型**：

| 操作 | 正向 | 反向 |
|------|------|------|
| 安装 | 依赖 installed + version | — |
| 卸载 | — | 下游 installed |
| 启用 | 依赖 installed + enabled + version | — |
| 禁用 | — | 下游 enabled |
| 升级 | 同安装轴候选版本 | 下游 installed 版本契约 |

实现要点：`PluginSnapshot.Enabled` 从 registry 全局 `Status == enabled` 写入；`InstallCheckInput.RequireEnabled` / `ReverseCheckInput.OnlyEnabledDependents` 区分轴；新增 `not_enabled` / `dependency_not_enabled`；卸载反向保留 `PLUGIN_REVERSE_DEPENDENCY_BLOCKED`，禁用反向使用 `PLUGIN_REVERSE_ENABLED_DEPENDENCY_BLOCKED`；任何轴只阻断不级联。租户粒度与现网 `UpdateStatus` 对齐，只认全局 registry 启用态。行为 BREAKING：下游仅禁用后允许禁用依赖插件；卸载语义不变。

## 3. 生命周期编排下沉

`plugin_lifecycle.go`、`plugin_lifecycle_source.go` 和 `plugin_auto_enable.go` 仍在根门面承载长状态机。

**决策**：重建 `internal/lifecycle` 为编排 owner，接收 catalog、store、runtime、integration、migration、dependency、i18n、cache publisher、topology 等窄接口。SQL migration executor 独立为 `internal/migration`。根门面只保留平台治理守卫和委托。

**关键设计**：
- Install/Uninstall/UpdateStatus 分批迁入 lifecycle
- source/dynamic veto 汇总统一为同一套 helper
- 列表投影收敛为 `buildPluginProjection` 单一入口（list/summary/detail/dependency_snapshot 四种模式）
- `publishPluginChange` 统一封装 revision 发布、管理读模型失效和派生缓存失效
- 业务控制参数从 context key 改为显式 options

## 4. 升级编排统一

source/dynamic 两套升级骨架分散在 `sourceupgrade`、`runtimeupgrade`、`store` 升级投影和根门面。

**决策**：新建 `internal/upgrade` 作为统一升级编排 owner，吸收 `sourceupgrade` 和 `runtimeupgrade`。source/dynamic 差异作为策略函数存在，共享依赖校验、反向依赖保护、失败诊断和缓存发布骨架。

**关键设计**：
- 失败诊断统一使用 `sys_plugin_migration` 约定（`phase=upgrade`、`migration_key=upgrade-phase-<phase>`）
- 治理守卫只在门面入口执行一次，内部策略不得再入公开方法
- cache publisher 失败不把失败目标 release 写成权威缓存来源
- 删除 `internal/sourceupgrade` 和 `internal/runtimeupgrade` 目录

## 5. 运行时认证快照

动态路由认证在 `runtime_route_auth.go` 中直接读取 `sys_user_role`、`sys_role`、`sys_role_menu` 和 `sys_menu`，绕过 role 模块的 token access snapshot 和 permission-access 修订号。

**决策**：由 role 模块发布面向动态插件运行时的窄访问投影契约，动态路由认证通过构造函数显式注入该契约。session 校验继续使用启动期共享 `session.Store`。

**关键设计**：
- 访问投影复用 token access snapshot、permission-access 修订号和 fail-closed 策略
- host call 授权快照限制在单次请求内复用
- datahost 表契约缓存按插件迁移账本和授权方法 fingerprint 失效

## 6. 读模型性能优化

动态路由热路径重复解析同一 artifact，详情/依赖检查/OpenAPI 投影反复全量 ScanManifests，单插件变更触发全部 WASM 编译缓存失效。

**决策**：新增清单读模型缓存（源码 manifest、动态 desired manifest、release manifest、YAML 快照），ScanManifests 稳态成本收敛为目录枚举加 stat 守卫，WASM 编译缓存按插件/artifact 路径失效，集群 peer 执行有界差异对账。

**关键设计**：
- `pluginID` 到 artifact 路径的索引化查询
- 动态路由请求上下文复用已解析 manifest
- 编译过程移出全局缓存写锁

## 7. 运行时组合简化

`RuntimeDelegate` 未绑定时返回 nil 或原始入参，cache/upgrade adapter nil service 静默成功，kvcache 依赖进程级默认 provider。

**决策**：delegate 保留为组合接缝，但绑定状态必须可诊断，未绑定运行期调用返回明确错误。kvcache 后端选择改为按 `cluster.enabled` 显式创建。

## 8. WASM 路由瘦身

`route.go` 接近千行，混合路由匹配、鉴权、权限查询、请求封装和响应写回。公共 host call helper 位于领域文件。

**决策**：拆分为 `route_match.go`（路由匹配）、`route_auth.go`（鉴权与权限）、`route_envelope.go`（请求封装）、`route_response.go`（响应写回）和 `route_context.go`（上下文状态）。`capabilityContextForHostCall` 迁回 wasm 公共层。静态测试约束 `route.go` 不超过 400 行。

## 9. 动态 i18n host service 收敛

动态插件同时拥有 `service: i18n` 会让 guest 侧误以为可以主动读取 locale 或翻译消息。

**决策**：从动态插件 host service catalog、协议别名、guest SDK 目录和 WASM dispatcher 中移除 i18n 服务。`plugin.yaml hostServices` 校验拒绝 `service: i18n`。源码插件保留 `I18n()` 能力。

## 缓存一致性

- 权威数据源：插件 registry/release、manifest artifact/source tree、runtime revision controller
- 一致性模型：单节点同步失效 + `plugin-runtime` revision 跨实例广播
- 失效触发点：治理写入成功后通过 `publishPluginChange` 统一发布
- 跨实例同步：复用 `cachecoord.Service` 和 `revisionctrl.Controller`
- 最大陈旧窗口：由 `ensureRuntimeCacheFresh` 和 revision 观察窗口约束
- 故障降级：保守隐藏

## DI 拓扑

- `plugin.New()` 按 catalog/store/migration/lifecycle/runtime/integration/sourceupgrade/upgrade 顺序分阶段构造
- 所有内部 service 通过构造函数逐项显式接收依赖，不使用 Deps/Options 聚合结构体
- `plugin.RuntimeDelegate` 由 HTTP 组合根创建，先传入 auth/role/menu/apidoc 等消费者，真实 pluginSvc 构造完成后绑定
- WASM runtime 由 `plugin.New` 内部调用 `wasm.NewRuntime` 创建并持有
- revision controller 由 `cachecoord` 服务与拓扑 owner 提供共享 revision 后端

## 10. Builtin 插件分发治理

**决策**：在插件 manifest 中新增`distribution`字段，缺省归一化为`managed`，支持`builtin`声明项目内建源码插件。合法值仅`managed|builtin`；旧值`marketplace`在有效契约中拒绝。`builtin`必须同时满足源码插件和编译期注册，动态插件不能声明`builtin`。

**关键设计**：
- `sys_plugin`基线表结构使用`distribution varchar(32) not null default 'managed'`
- 普通插件管理列表默认展示`builtin`与`managed`；列表/详情投影继续暴露`distribution`；`includeBuiltin`保留为兼容查询字段（忽略或始终视为包含），不再用于隐藏 builtin
- 写操作（安装、启用/禁用、卸载、手动升级、租户供应策略变更）由服务端 guard 统一拒绝，拒绝语义不因列表可见而放宽
- 启动期独立执行`BootstrapBuiltinPlugins(ctx)`，在插件路由、cron、前端包预热前自动安装、启用和安全升级 builtin 源码插件
- 生命周期变化继续复用现有依赖解析、SQL 迁移、资源同步、缓存失效、enabled snapshot 和集群主节点边界

**管理 UI 只读治理**：
- `distribution === 'builtin'` 时展示「内置插件」badge（中/英 i18n）；若同时命中宿主`plugin.autoEnable`，继续展示既有「自动启用」类标识，二者可并存
- 安装/启停/升级/租户策略入口对 builtin 行继续隐藏；具备卸载权限时仍展示「卸载」按钮，但置为`disabled`并附 tooltip 说明不可卸载，以与可卸载行保持操作列按钮数量与列宽一致
- 详情入口始终可用；「管理」仍按「已安装 + 存在管理页」判定，不因 distribution 禁用
- 前端隐藏或置灰不得作为安全边界；绕过 UI 调用写 API 时服务端仍按升级治理规范拒绝

## 10.0 插件注册表变更后的静默路由刷新

插件启停后宿主会重建菜单与动态路由；早期在当前路由仍可访问时默认 `router.replace({ force: true })`，导致插件管理页等宿主静态页 remount、筛选/滚动丢失。

**决策**：在 `access-refresh` 决策层、于 `generateAccess` 之后按当次结果判定——当前路由仍可访问且无需路径纠正、无强制默认路由、无当前页 pending plugin generation remount 时静默跳过 force 导航；不可访问则 fallback；存在 `replacementPath` 则纠正导航。不按插件管理页白名单特例。禁用/卸载的 Tab 清理仍在业务动作处执行。自动静默不得在入队时把 `skipRouteNavigation` 无脑 OR 进队列以免吞掉必要导航。

## 10.1 插件管理列表「管理」入口

运维人员进入某个插件的业务管理页（如 LDAP 设置、登录日志、通知管理）时，不应只能从左侧菜单自行定位。列表操作列需要提供直达入口，并在无管理页或未安装时明确置灰。

**决策**：

1. **判定来源：前端 page-registry**。以`getPluginPages()`中归属该`pluginId`的可导航页面为准；排除`frontend/pages/components/**`以及文件名含`modal`/`drawer`的辅助组件。构建期已有稳定注册表，无需后端扩展列表字段，符合列表首屏性能约束。
2. **多管理页目标选择**。以当前会话`accessMenus`深度优先遍历顺序选择该插件第一个匹配菜单路径；若 access 菜单尚无匹配则回退`router.getRoutes()`注册顺序中的第一个匹配路径。**禁止**按`routePath`字母序排序，否则会出现如`/ai/invocations`排在`/ai/providers`前、误进非首位菜单的问题。侧边栏菜单顺序即用户感知的“第一个菜单”，与`plugin.yaml`的`sort`一致。
3. **跳转路径解析**。路径匹配支持完整相等或后缀匹配，以兼容相对菜单路径挂到父目录后的完整 URL。当前会话找不到任何匹配路由时，保持在列表页并给出用户可见提示。按钮启用只表达“插件声明了管理页且已安装”；真正可访问性仍受启用状态与权限约束。
4. **按钮状态**。已安装且存在可导航管理页 → 可点击；未安装 → 禁用并提示先安装；已安装但无管理页 → 禁用并提示无管理页面。

**非目标**：不为每个插件强制新增管理页；不改变菜单同步、权限过滤或动态插件资产托管语义；不在列表接口返回`managementPath`等新字段。

**取舍**：仅 iframe/资产页、未进入 page-registry 的动态插件可能被判定为无管理页（当前托管工作台主要源码插件管理页走 page-registry）；多管理页只进一个，优先菜单顺序首位，后续若需要可改为下拉。

## 11. 插件领域能力扩展

**阶段 0/1 冻结**：冻结插件领域能力扩展的四类矩阵（方法发布、错误语义、规模上限、动态授权资源），实现第一批高频只读能力：`Users.Current`、`Users.BatchResolve`、`Authz.BatchHasPermissions`、`Dict.EnsureValuesVisible`、`Sessions.Current`。

**阶段 1.5-5 完成**：继续完成剩余能力，覆盖候选搜索（`Users.Search`扩展、`Dict.ListValues`、`Files.Search`、`Jobs.Search`等）、组织/租户/插件治理投影、插件私有资源批量（`Storage.BatchStat`、`Cache.GetMany`等）、通知类型化和 AI 状态，共 40+ 个新方法。

**关键决策**：
- 动态普通领域方法继续使用 JSON envelope，不新增 per-domain 专用 codec
- 数据权限在 owner 或 provider 批量路径完成，不在 adapter 中内存过滤
- 缓存和运行时状态只复用现有 owner，不新增权威缓存
- README 和治理测试作为每批完成门禁

## 12. 插件领域服务统一

**决策**：废除`capability.AdminServices`和各领域`AdminService`，每个领域只保留一个插件可见`Service`入口。动态插件可声明性只由`host service registry`注册事实表达，不新增额外方法可声明性字段。

**关键设计**：
- 动态 wire method 一次性标准化，不保留旧方法兼容别名
- Auth 领域动态命名保持全局一致，使用`service: auth`声明认证和授权方法
- 治理能力内聚到对应领域`Service`，删除`pluginhost.Services.PluginLifecycle()`和`PluginState()`顶层入口
- 领域实现归属真实 owner，`capabilityhost`和 WASM host service 只保留薄适配职责

## 13. 插件服务契约收敛

**决策**：收敛`plugin.Service`方法集合，删除仅作为语义包装或无生产入口的方法。新增按真实消费场景划分的插件 service 私有 facet 接口，对外只保留统一`Service`入口。

**关键设计**：
- 插件 job 查询合并为`ListManagedJobs(ctx, ManagedJobQuery)`，用查询参数表达可执行、已安装、插件 ID 和 handler 是否需要返回
- 插件状态变更合并为`UpdateStatus(ctx, pluginID, UpdateStatusOptions)`，保留目标状态和动态插件授权确认输入
- Auth hook 的三个包装方法删除，调用方使用已有`DispatchHookEvent`表达具体事件

## 14. 窄接口移动到消费者

**决策**：将生产者包中仅用于自组合的分类接口合并回默认`Service`接口，对确有复杂度收益的跨包窄依赖接口，将定义移动到消费者包。

**关键设计**：
- 接口按 owner 分为三类：生产者完整契约、稳定产品/运行期契约、消费方窄依赖
- 消费方优先复用目标组件已有`Service`或稳定契约；仅当完整契约不能清晰表达消费边界时才在消费方包内定义窄依赖接口
- 收敛`i18n.Service`，删除无业务入口的 i18n 管理诊断 API 和源码插件消息搜索方法

## Plugin-Owned Domain Capabilities

Non-core domain capabilities are owned by domain owner plugins. Public contracts live under `apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap`. Core keeps only the plugin kernel, dependency governance, generic capability descriptors, dynamic routing, authorization, audit, and lifecycle controls.

Owner-aware dynamic `hostServices` use structured `owner` and `version` fields rather than encoding ownership into the service string. Core merges static core-owned catalog entries with owner descriptors, stores authorization snapshots keyed by `owner/service/version`, and dispatches through a generic invoker path instead of domain-specific switches. Consuming plugins must declare a hard `dependencies.plugins` entry for the owner; reverse disable/uninstall/upgrade checks protect downstream consumers and must not N+1 on first-screen lists.

AI is the first owner pilot. `linapro-ai-core` owns text and multimodal contracts, provider SPI, and the dynamic guest bridge SDK. Core no longer ships production `aicap` contracts, `ProvideAIText`, or AI-specific codecs/dispatchers. Capability IDs use `plugin.linapro-ai-core.ai.<family>.v1`. Dynamic descriptors only publish runnable methods; remaining multimodal methods stay in the owner contract until invokers exist.

Import boundary scanning via `linactl plugins.check` allows cross-plugin production imports only into owner `backend/cap/...`. Runtime caches for registries, authorization snapshots, and owner availability follow critical runtime data rules: authoritative sources, post-commit invalidation, cluster coordination, and rebuild-or-deny on failure.

## Host Layer Simplification

New core-owned host service methods must use JSON envelopes. Existing dedicated codecs are frozen as a method-level allowlist. Wire constants for services and methods live only under `protocol/hostservices` and are referenced by the catalog; no `go generate`. Historical `HostServiceCapabilityJSON*` aliases are removed in favor of `HostServiceJSON*`. Upgrade preview/execute is owned by the lifecycle facade; the root plugin package no longer constructs or holds a parallel `upgrade.Service`, while public type aliases remain stable for management API callers.

## 同权同信与动态外部登录

**决策**：经宿主安装或升级治理并处于启用状态的动态插件，与源码插件适用同一信任级与能力准入模型；不得仅因 `type=dynamic` 永久拒绝发布某一 core-owned 领域能力。

**关键设计**：
- 动态插件可经 hostServices 授权调用 `external_login.login_by_verified_identity` 与 `users.create_from_external`，guest 走真实 host call 而非永久 stub。
- 源码 provider ownership 继续 `ProvideExternalIdentity(providerID)`；动态 ownership 由 `auth` 服务下 `resources[].ref` 声明 provider ID，WASM dispatcher 校验后盖章 pluginID 铸会话。
- 调用链：dynamic guest → domainhostcall → wasm dispatcher（授权 + ownership）→ capability 或等价 auth/users 实现。
- 安全仍依赖安装治理、方法级授权、provider ownership 与启用检查；被攻破的已授权动态插件与源码插件同信模型，后续可叠加宿主验签加固。
