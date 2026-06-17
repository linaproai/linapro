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
