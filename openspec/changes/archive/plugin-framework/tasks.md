# Tasks

## Summary

- [x] 建立插件框架核心能力：统一 plugin.yaml、源码/动态插件生命周期、动态 WASM 运行时、host service、能力目录、插件 UI、菜单权限、启动引导、依赖、升级、工作区和包边界。
- [x] 拆解 catalog/store/types，消除 catalog 反向回调环（split-plugin-catalog-store-types）。新建 plugintypes 叶子包、store 治理持久化组件，catalog 收窄为清单事实源。副作用调用上提到编排入口。边界测试固化 plugintypes 零兄弟依赖、catalog 不依赖 runtime/integration/dao。
- [x] 构造函数直化与 setter 清零（straighten-plugin-wiring-state）。删除所有内部 service wiring setter 和 ValidateRequiredDependencies；切断 runtime/integration/lifecycle 互持宽 service；消除包级可变运行期状态；WASM host service 改为显式实例；runtime revision controller 迁入 cachecoord/revisionctrl；合并 capabilityhost 微包。plugin.RuntimeDelegate 打破启动环。
- [x] 生命周期编排下沉（sink-plugin-lifecycle-orchestration）。重建 internal/lifecycle 接收 catalog/store/runtime/integration/migration/dependency/i18n/cache publisher/topology 窄接口。SQL migration executor 独立为 internal/migration。Install/Uninstall/UpdateStatus/源码生命周期/自动启用/租户钩子分批迁入。列表投影收敛为 buildPluginProjection 单一入口。publishPluginChange 统一缓存失效。业务控制参数从 context key 改为显式 options。
- [x] 升级编排统一（unify-plugin-upgrade-orchestration）。新建 internal/upgrade 吸收 sourceupgrade 和 runtimeupgrade。source/dynamic 共享依赖校验、失败诊断和缓存发布骨架。失败诊断统一 sys_plugin_migration 约定。治理守卫只在门面执行一次。删除 internal/sourceupgrade 和 internal/runtimeupgrade 目录。
- [x] 运行时认证快照（plugin-runtime-auth-snapshot-guardrails）。role 模块发布动态路由访问投影契约，runtime 通过构造函数注入。session 校验继续使用共享 session.Store。host call 授权快照限制在单次请求内。datahost 表契约缓存按 fingerprint 失效。
- [x] 读模型性能优化（plugin-runtime-read-model-performance）。清单读模型缓存覆盖源码/动态/release manifest。ScanManifests 稳态成本收敛为目录枚举加 stat 守卫。WASM 编译缓存按插件失效。集群 peer 有界差异对账。
- [x] 运行时组合简化（simplify-plugin-runtime-composition）。delegate 未绑定返回明确错误。cache/upgrade adapter nil service 语义收紧。kvcache 后端按拓扑显式创建。
- [x] WASM 路由瘦身（slim-plugin-wasm-route-dispatch）。route.go 拆分为 route_match/route_auth/route_envelope/route_response/route_context。公共 host call helper 迁回 wasm 公共层。静态测试约束 route.go 不超过 400 行。
- [x] 移除动态 i18n host service（remove-dynamic-plugin-i18n-host-service）。从动态插件 host service catalog 和 WASM dispatcher 中移除 i18n。源码插件保留 I18n() 能力。

## Verification

- [x] `cd apps/lina-core && go test ./internal/service/plugin/... -count=1` 通过
- [x] `cd apps/lina-core && go test ./internal/service/i18n/... -count=1` 通过
- [x] `cd apps/lina-core && go test ./internal/service/cachecoord/... -count=1` 通过
- [x] `cd apps/lina-core && go test ./internal/cmd -count=1` 通过
- [x] 各迭代 openspec validate --strict 通过
- [x] 静态边界测试覆盖：plugintypes 零兄弟依赖、catalog 不依赖 runtime/integration/dao、store 不泄漏 DAO/DO/Entity、无 wiring setter、无包级可变状态、无旧 runtimecache import、无 wasm.Configure*、route.go 不超过 400 行、无旧 sourceupgrade/runtimeupgrade import

## Governance

- [x] 审查：lina-review 覆盖每个迭代，规则域包括 OpenSpec、后端 Go、架构、插件、缓存一致性、数据权限、测试、i18n、文档、开发工具跨平台
- [x] i18n：无运行时用户可见文案或语言包新增；动态插件 i18n host service 移除属于边界调整
- [x] 数据权限：平台治理守卫和租户边界未放宽；host service 授权快照语义不变
- [x] 缓存一致性：权威源、触发点、跨实例同步、最大陈旧窗口、故障降级未因重构改变
- [x] DI：所有内部 service 构造函数逐项显式注入，不使用 Deps/Options 聚合结构体
- [x] 不修改 HTTP API、DTO、SQL schema、前端页面、插件 manifest wire 或 apps/lina-plugins/* 源码目录
