# Tasks

## Summary

- [x] 建立插件框架核心能力：统一 plugin.yaml、源码/动态插件生命周期、动态 WASM 运行时、host service、能力目录、插件 UI、菜单权限、启动引导、依赖、升级、工作区和包边界；插件管理列表操作列「管理」入口（page-registry 判定、多菜单首位跳转、未安装/无管理页置灰）。
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
- [x] 「管理」入口：路由解析 helper 单测；插件管理 E2E TC019（有管理页可跳转、无管理页禁用）；PluginPage POM 管理按钮定位；OpenSpec 严格校验

## Feedback

- [x] FB-1（「管理」入口）：插件未安装时按钮应置灰不可点击；根因：初始仅按是否有管理页启用按钮，未绑定安装状态；处理：未安装强制 disabled 并提示先安装；验证：单测与 E2E 覆盖
- [x] FB-2（「管理」入口）：多菜单插件应进入侧边栏顺序第一个菜单而非最后一个；根因：曾按 routePath 字母序或注册表末项选择；处理：改以 accessMenus 深度优先首位匹配，禁止字母序；验证：多菜单跳转断言

## Governance

- [x] 审查：lina-review 覆盖每个迭代，规则域包括 OpenSpec、后端 Go、架构、插件、缓存一致性、数据权限、测试、i18n、文档、开发工具跨平台、前端 UI
- [x] i18n：动态插件 i18n host service 移除属于边界调整；「管理」入口新增宿主前端管理按钮与提示文案（中英文）
- [x] 数据权限：平台治理守卫和租户边界未放宽；host service 授权快照语义不变；「管理」入口不新增数据访问接口
- [x] 缓存一致性：权威源、触发点、跨实例同步、最大陈旧窗口、故障降级未因重构改变；「管理」入口无缓存影响
- [x] DI：所有内部 service 构造函数逐项显式注入，不使用 Deps/Options 聚合结构体
- [x] 「管理」入口不修改 HTTP API、DTO、SQL schema 或列表读模型字段；前端仅用 page-registry 与路由表
- [x] Builtin 插件分发治理：新增`distribution`字段支持`builtin`声明；启动期自动安装/启用/安全升级 builtin 源码插件；服务端拒绝普通管理写操作；`sys_plugin.distribution`进入基线表结构。普通可管理枚举由`marketplace`重命名为`managed`，缺省归一化为`managed`，旧`marketplace`在有效契约中拒绝，无兼容分支。
- [x] 插件管理默认展示 builtin：普通列表投影默认含`distribution=builtin`；UI 内置标识（可与自动启用并存）；安装/启停/升级/租户策略隐藏；卸载置灰+tooltip；详情与管理可用；`includeBuiltin`兼容；写拒绝边界不变；TC016/列表单测对齐。
- [x] 插件领域能力扩展阶段 0/1：冻结四类矩阵，实现`Users.Current`、`Users.BatchResolve`、`Authz.BatchHasPermissions`、`Dict.EnsureValuesVisible`、`Sessions.Current`，同步动态 host service catalog/guest/dispatcher/README。
- [x] 插件领域能力扩展阶段 1.5-5：完成候选搜索、组织/租户/插件治理投影、插件私有资源批量、通知类型化和 AI 状态，覆盖 40+ 个新方法；FB-1 至 FB-3 收敛 Storage provider 选择为零配置、新增冲突错误码。
- [x] 插件领域服务统一：废除`AdminServices`和各领域`AdminService`；动态 wire method 一次性标准化；治理能力内聚到对应领域`Service`；FB-1 至 FB-59 大量反馈闭环。
- [x] 插件服务契约收敛：收敛`plugin.Service`方法集合；`ListManagedJobs`替换多个 job 查询；`UpdateStatus`替换 Enable/Disable/SetStatus；FB-1 至 FB-6 收敛 facet 为私有定义。
- [x] 窄接口移动到消费者：合并生产者自组合接口；移动跨包窄依赖到消费者；收敛`i18n.Service`删除无业务入口管理诊断 API；FB-1 至 FB-50 大量反馈闭环。
- [x] 验证：`go test`覆盖 capability、pluginhost、pluginbridge、capabilityhost/wasm、各领域 owner adapter 等包；`openspec validate`通过；静态检索确认旧 AdminService/旧 wire method/旧顶层入口无残留。

- [x] 交付 plugin-owned 领域能力：通用 descriptor 注册、owner-aware hostServices/catalog/授权快照/dispatcher、依赖与反向阻断、跨插件 import 边界扫描、`linapro-ai-core` backend/cap 契约与 bridge SDK、core AI 生产入口清理。
- [x] 交付宿主分层简化：JSON envelope 政策、dedicated codec 方法级冻结、wire 常量单一来源、upgrade 归 lifecycle、README 同步。
- [x] 验证：core 插件相关包测试、`linapro-ai-core`/`linapro-demo-dynamic` 包测试、`plugins.check`、`make wasm`、i18n 检查、E2E TC017/TC007、OpenSpec 严格校验与 lina-review。
- [x] FB 摘要：owner invoker 收敛到`aicap.Service`；动态 descriptor 仅发布可运行方法；wire 常量放弃 go generate；CapabilityID 改为`plugin.linapro-ai-core.ai.*`。
- [x] 治理：i18n 覆盖 owner 错误码/授权展示与前端 owner/version 标签；缓存按关键运行时数据治理；数据权限在 owner 查询路径接入；无新增业务表；API 主要扩展 owner-aware 投影；前端授权弹窗可观察行为由 E2E 覆盖；DI 以启动期 descriptor registry 与 domainServices 显式传递。

- [x] 发布动态外部登录与从外部建号 host service（`external_login.login_by_verified_identity`、`users.create_from_external`），对齐安装授权后源码/动态同权同信；更新 README/规则与 fail-closed 单测为未授权拒绝/授权成功。
- [x] 验证：相关 go test、OpenSpec strict；治理：i18n 文档表述同步，无新增用户可见运行时文案要求；数据权限与 ownership 校验保持 fail-closed。
- [x] 依赖两轴：`PluginSnapshot.Enabled`、`RequireEnabled`/`OnlyEnabledDependents`、启用正向与禁用反向矩阵、错误码/文案分流；resolver + lifecycle 单测/集成测通过。
- [x] 静默路由刷新：access-refresh 在 generateAccess 后按可访问性判定是否 force 导航；宿主静态页启停不 remount；单测锁定静默/fallback/replacement 决策。
- [x] 验证：相关 Go 测试包、前端 access-refresh 单测、`openspec validate --strict`、lina-review；治理：禁用反向新增 i18n 错误文案；无 DB schema；静默刷新无后端/权限契约变更。
