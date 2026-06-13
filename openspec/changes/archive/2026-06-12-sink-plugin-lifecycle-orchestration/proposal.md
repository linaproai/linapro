## Why

`split-plugin-catalog-store-types`和`straighten-plugin-wiring-state`已经拆掉 catalog 环根并清理 setter 装配，但插件生命周期、源码生命周期、自动启用、列表投影和缓存失效编排仍滞留在`plugin`根门面。继续把这些长流程留在 facade 层会阻碍后续统一升级体系，并让高频插件列表路径的扫描、缓存和批量装配边界难以集中治理。

## What Changes

- **BREAKING** 将`plugin_lifecycle.go`、`plugin_lifecycle_source.go`、`plugin_auto_enable.go`中的生命周期编排下沉到重建后的`internal/lifecycle`编排组件，`plugin`根门面保留平台治理守卫和一行委托。
- 将旧`internal/lifecycle`中 SQL migration executor 职责迁出为独立`internal/migration`组件，避免“lifecycle”组件名实不符。
- 统一 lifecycle veto 汇总、动态 lifecycle decision 汇总、卸载收尾、动态插件启用资格判断和 decision/err 处理 helper，减少同构复制逻辑。
- 将`Install`、`Uninstall`、`UpdateStatus`、源码插件安装/卸载/回滚、启动自动启用和租户生命周期钩子拆成命名明确的窄函数，改造本次迁移触碰的编排函数，使其长度和调用链可审查。
- 将插件列表、管理摘要、管理详情和依赖快照装配收敛到单一投影构建入口，并保留批量读取、快照和缓存命中路径，避免列表路径因分散实现产生`N+1`风险。
- 将插件 runtime、管理读模型、frontend/i18n/WASM 派生缓存失效收敛为单一插件变化发布入口，继续复用`plugin-runtime`revision 协调，不创建新的缓存域。
- 删除根门面生产代码中的直接`dao`访问和本变更迁移后不再需要的业务控制流 context key；确需保留的启动快照 context 仅限 startup orchestration 输入，不再作为普通生命周期控制参数。
- 不统一 source/dynamic upgrade 体系，不删除`sourceupgrade`或`runtimeupgrade`包；这些属于后续 D。也不拆分`runtime/route.go`或继续改造 WASM 分发层；这些属于后续 E。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-service-layout`：新增 lifecycle 编排归属、根 facade 约束、列表投影收敛和静态治理要求。
- `plugin-manifest-lifecycle`：补充生命周期编排下沉后必须保持的安装、卸载、启停、自动启用、租户生命周期钩子和 veto 语义。
- `distributed-cache-coordination`：补充插件生命周期变化必须通过单一`plugin-runtime`变化发布入口失效派生缓存和管理读模型。

## Impact

- 影响后端内部代码：`apps/lina-core/internal/service/plugin/**`，重点是根门面、`internal/lifecycle`、新增`internal/migration`、`internal/management`或列表投影所在组件、`internal/testutil`以及相关测试。
- 影响测试：插件 service 全量单元测试、生命周期/源码生命周期/自动启用/租户钩子测试、列表投影测试、缓存失效测试、启动绑定测试和静态边界测试。
- 不修改 HTTP API、DTO、OpenAPI 文案、数据库 schema、SQL 迁移、前端页面、插件 manifest wire 格式或`apps/lina-plugins/<plugin-id>/`源码目录。
- `i18n`资源无新增文案，仍需验证 lifecycle 错误和 veto 汇总复用既有 message key 与 fallback。
- 数据权限语义不变，插件列表和详情仍按既有平台治理和宿主能力可见性规则返回；本变更只改变内部装配和批量投影路径。
