## Why

`split-plugin-catalog-store-types`、`straighten-plugin-wiring-state`和`sink-plugin-lifecycle-orchestration`已经完成 catalog 拆解、构造直化和生命周期编排下沉，但插件升级仍分散在`sourceupgrade`、`runtimeupgrade`、`store`升级投影、`runtime`发布切换和根门面 preview/execute 文件中。继续保留两套 source/dynamic 平行升级实现，会让失败记账、治理守卫、缓存失效和 release 提升语义难以统一审查。

## What Changes

- **BREAKING** 新建`apps/lina-core/internal/service/plugin/internal/upgrade`作为插件升级编排 owner，吸收`internal/sourceupgrade`、`internal/runtimeupgrade`、`store`中 runtime upgrade 状态投影相关职责，以及根门面 runtime upgrade preview/execute 编排。
- 统一 source 与 dynamic 插件升级的 preview、execute、失败记账、release 提升、治理资源同步和缓存发布骨架，source/dynamic 仅作为执行策略差异存在。
- 消除`executeRuntimeUpgradeByType`通过公开`UpgradeSourcePlugin`再入 source 升级路径导致的平台治理守卫和缓存发布重复执行风险，内部路径改为直接调用 upgrade 子组件窄方法。
- 将`sys_plugin_migration`升级失败诊断收敛为一套读写约定，复用既有 upgrade phase 常量和结构化错误语义，不新增 SQL schema。
- 保持插件管理 API、DTO、OpenAPI 文案、前端页面、插件 manifest wire 和动态插件 guest 协议不变。
- 不拆分`runtime/route.go`，不继续改造 WASM host service 分发层；这些仍属于后续可选 E 阶段。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-upgrade-governance`：补充 source/dynamic 升级统一编排、失败记账单一约定、治理守卫单次执行和内部调用路径语义要求。
- `plugin-service-layout`：补充插件升级编排必须归属`internal/upgrade`子组件，根门面不得继续承载升级长流程或依赖`sourceupgrade`/`runtimeupgrade`平行包。

## Impact

- 影响后端内部代码：`apps/lina-core/internal/service/plugin/**`，重点是根门面升级文件、`internal/sourceupgrade`、`internal/runtimeupgrade`、`internal/runtime`升级/发布切换、`internal/store`升级状态投影、`internal/testutil`和相关测试。
- 影响测试：插件 service 根包升级测试、source upgrade 测试、dynamic runtime upgrade 测试、runtime release/reconciler 测试、store upgrade projection 测试、静态边界测试和启动绑定测试。
- 不修改 HTTP API、路由、DTO、OpenAPI 文案、数据库 schema、SQL 文件、前端页面、插件 manifest wire、动态插件 guest 协议或`apps/lina-plugins/<plugin-id>/`源码目录。
- `i18n`资源原则上无新增文案；若升级错误 owner 迁移触及错误码或 message key，必须保持既有 key/fallback 或同步记录资源影响。
- 数据权限语义不变，升级入口仍由插件根门面平台治理守卫保护，内部调用不得放宽插件治理可见性或租户边界。
- 缓存一致性语义不变，升级成功或失败后的派生缓存发布继续经统一`publishPluginChange`入口和`plugin-runtime`revision controller。
