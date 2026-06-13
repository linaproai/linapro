## Why

方案 E 的前置变更已经归档，当前代码中`runtime/route.go`仍接近千行并混合动态路由匹配、鉴权、权限查询、请求封装和响应写回；`wasm`侧也仍有公共 host call helper 归属在单一领域文件中，容易让后续新增领域 host service 继续复制样板。

本变更用于完成方案 E 的可选复杂度治理，把已落地的 registry 分发模式和动态路由分发职责固化为可维护的内部结构。

## What Changes

- 将`capabilityContextForHostCall`等跨领域 host call 公共工具迁出`users`领域文件，放回`wasm`公共层或 registry 公共层。
- 基于既有`hostservicedispatch`registry 模式补齐静态边界，确保新增普通领域 host service 只需要新增领域文件和显式注册条目，不需要修改统一入口分发分支。
- 拆分`apps/lina-core/internal/service/plugin/internal/runtime/route.go`，让入口分发、路由匹配、鉴权与权限查询、请求 envelope、响应写回分别落在职责明确的文件中。
- 增加治理测试或静态测试，约束`route.go`不超过`400`行，并阻断公共 helper 回流到领域文件。
- 保持 HTTP API 路径、动态插件 route contract、WASM guest 协议、host service wire 字符串、数据权限、缓存失效和`i18n`资源语义不变。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-host-service-extension`：补充 WASM host service 公共 helper 归属和 registry 驱动新增领域的可维护性约束。
- `plugin-service-layout`：补充动态插件 route 分发内部职责拆分和入口文件复杂度上限要求。

## Impact

- 影响代码：
  - `apps/lina-core/internal/service/plugin/internal/wasm/**`
  - `apps/lina-core/internal/service/plugin/internal/runtime/route*.go`
  - 相关同包 Go 单元测试或静态治理测试
- 不影响范围：
  - 不修改`api/`DTO、HTTP 方法、公开 URL、权限标签或 OpenAPI 元数据。
  - 不修改 SQL、DAO、DO、Entity、插件 manifest wire、动态插件 artifact 格式或 guest SDK 公开协议。
  - 不修改`apps/lina-plugins/*`业务插件目录。
  - 不新增运行期依赖，不新增缓存域，不新增用户可见文案或`i18n`资源。
