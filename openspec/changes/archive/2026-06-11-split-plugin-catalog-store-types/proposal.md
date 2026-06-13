## Why

`apps/lina-core/internal/service/plugin/internal/catalog`同时承担插件清单扫描校验、治理表读写、发布/授权投影、依赖文法和跨包副作用回调，已经成为`catalog`、`runtime`、`integration`之间 setter 回注环的根源。先拆解该环根，可以在不改变插件外部契约的前提下为后续插件服务装配直化、生命周期编排下沉和升级治理统一创造稳定前置。

## What Changes

- **BREAKING**：不保留`catalog`旧内部路径、setter、别名或过渡 wrapper；项目无兼容性负担，直接迁移内部调用方。
- 新增`plugin/internal/plugintypes`叶子包，承载插件 ID、状态、类型、scope、generation、版本等纯类型和值对象。
- 新增`plugin/internal/store`子组件，作为`sys_plugin*`插件治理表读写、发布/授权/治理投影、node/release state 持久化的唯一 owner。
- 收窄`catalog`为插件清单事实源，只保留清单扫描、解析、校验和访问能力。
- 删除`catalog`中的`Set*` wiring 方法、`runtime`/`integration`反向回调字段和由`catalog`触发的菜单同步、hook 分发、资源引用同步等副作用。
- 将`menuSyncer`、`hookDispatcher`、`resourceRefSyncer`等副作用调用时机上提到当前编排入口；本变更不要求同步完成 lifecycle 子包下沉。
- 将`backendLoader`、`dynamicManifestLoader`、`artifactParser`等扫描输入改为可下沉的清单/资源读取能力或调用入口显式参数，不再作为`catalog`实例字段。
- 建立 import 边界治理测试或静态扫描，固化`plugintypes`、`catalog`、`store`、`runtime`、`integration`之间的依赖方向。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-service-layout`：插件服务内部子组件职责边界新增约束，要求`catalog`、`store`、`plugintypes`职责分离并移除`catalog`反向回调环。
- `plugin-manifest-lifecycle`：插件清单生命周期新增实现约束，要求清单扫描/校验与治理表写入、副作用同步分离，清单查询不得隐藏触发治理副作用。

## Impact

- 影响代码：`apps/lina-core/internal/service/plugin`及其`internal/catalog`、`internal/runtime`、`internal/integration`、`internal/dependency`、测试支撑和相关调用方。
- 不修改 HTTP API、路由、DTO、OpenAPI 元数据、前端页面、插件`plugin.yaml`语义、host service wire 字符串或数据库 schema。
- 不新增 SQL 迁移，不修改`dao`、`do`、`entity`生成代码。
- 本变更属于宿主插件服务内部架构治理；运行时行为目标保持不变，但实现必须显式记录缓存、数据权限、`i18n`和测试影响判断。
