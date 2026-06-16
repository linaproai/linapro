## Why

动态插件路由当前在插件`runtime`内部直接查询角色、角色菜单和菜单表来构建权限上下文，绕过了`role`模块已经建立的 token access snapshot、`permission-access`修订号和数据范围 owner 边界。该路径既有性能成本，也存在权限、租户和数据范围语义漂移风险。

## What Changes

- 由`role`模块发布面向动态插件运行时的窄访问投影契约，动态路由认证通过构造函数显式注入该契约，不再直接访问角色相关 DAO。
- 动态路由身份快照复用`role`模块的 token access snapshot、`permission-access`修订号、租户维度和 fail-closed 语义。
- 动态路由继续通过启动期共享`session.Store`校验 session hot state；不新增独立 session 缓存，不绕过登出、强制下线、token 撤销或 Redis fail-closed 语义。
- 优化请求内 host call 授权快照复用和 datahost 表契约缓存，但不得放宽 host service 授权、数据权限、租户边界或结构化 data service 约束。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-runtime-loading`：补充动态路由认证必须通过`role`访问投影契约和 session hot state 共享实例的要求。
- `role-management`：补充`role`模块必须发布 token access snapshot 投影给动态插件运行时使用。
- `plugin-host-service-extension`：补充 host call 授权快照复用时不得改变授权、数据权限、租户边界和错误 envelope。
- `plugin-data-service`：补充 datahost 表契约缓存的权威源、失效触发点和数据权限边界。

## Impact

- 影响`apps/lina-core/internal/service/role`、`apps/lina-core/internal/service/plugin/internal/runtime`、`wasm`、`datahost`、host service dispatch 和启动装配路径。
- 涉及认证、权限、数据权限、缓存一致性、显式依赖注入和后端测试。
- 不新增 HTTP API，不改变动态插件 host service 协议，不改变`plugin.yaml`。
- 已有`session.Store.TouchOrValidate`具备 last-active 写回节流和集群 hot state fail-closed 语义，本变更不重复定义新的 session 降频机制。
