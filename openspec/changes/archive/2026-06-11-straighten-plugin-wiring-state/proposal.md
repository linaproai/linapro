## Why

`split-plugin-catalog-store-types`已经拆掉`catalog`环根，但`plugin.New()`、`runtime`、`integration`、`lifecycle`、`wasm`和`runtimecache`仍依赖构造后 setter、包级可变状态和反向持有门面来维持运行。继续推进完整重构前，必须先把装配边界改为构造函数显式注入，并把缓存敏感状态收敛到启动期共享实例。

## What Changes

- **BREAKING** 重构插件服务内部构造顺序：删除`runtime`、`integration`、`lifecycle`等内部 service 的 wiring setter 和`ValidateRequiredDependencies`运行期兜底，改为构造函数逐项显式接收运行期依赖。
- 切断`runtime`与`integration`互相持有完整 service 的循环：把菜单同步、hook 分发、权限菜单过滤、动态 job 执行和依赖校验改为组合根注入的窄契约。
- 将`integration`的`defaultSharedState`、插件生命周期 observer、runtime cache observed revision、runtime reconciler once/mutex 等包级可变状态改为 service 实例或启动期显式共享对象。
- 将`wasm` host service 的`Configure*`包级配置入口和`atomic.Pointer`快照改为由构造函数或显式运行期实例持有，生产 host call 不再读取包级可变配置。
- 将`plugin/runtimecache`迁出插件树，归入缓存协调职责边界；`plugin`和`i18n`通过同一 revision controller 组件实例化各自缓存域。
- 合并`capabilityhost/internal/*cap`单文件子包到`capabilityhost`包内文件，保持接口型依赖逐项显式传入，不引入 options/deps 聚合结构体。
- 更新测试 fixture，删除对生产接线流程的复刻；新增静态边界测试固化 setter、包级可变状态和旧 runtimecache import 清零。
- 不下沉生命周期编排、不统一 source/dynamic upgrade、不瘦身`route.go`；这些继续由后续 C/D/E 变更承接。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-service-layout`：新增插件服务内部构造、状态归属和子包布局治理要求。
- `plugin-host-service-extension`：补充动态插件 WASM host service 运行期依赖不得通过包级配置读取的要求。
- `distributed-cache-coordination`：补充插件 runtime revision controller 作为通用缓存协调组件的归属和实例来源要求。

## Impact

- 影响后端内部代码：`apps/lina-core/internal/service/plugin/**`、`apps/lina-core/internal/service/i18n/**`、`apps/lina-core/internal/cmd/**`，以及必要的缓存协调内部包。
- 影响测试：插件 service 全量单元测试、`internal/cmd`启动绑定测试、WASM host service 测试、i18n revision 测试和静态边界测试。
- 不修改 HTTP API、DTO、OpenAPI 文案、数据库 schema、插件 manifest wire 格式或前端页面。
- 不修改`apps/lina-plugins/<plugin-id>/`插件源码目录。
