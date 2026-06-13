## 1. OpenSpec 与基线

- [x] 1.1 记录方案 E 归档后基线、影响判断和不变契约范围，并运行`openspec validate slim-plugin-wasm-route-dispatch --strict`。
  记录：当前无其他活跃 OpenSpec 变更，`runtime/route.go`基线为`995`行，`wasm`已有`hostservicedispatch`registry 与`wasm_host_service_registry.go`，但`capabilityContextForHostCall`仍定义在`wasm_host_service_users.go`。本变更只处理方案 E 的内部复杂度治理，不修改 HTTP API、DTO、SQL、前端页面、业务插件目录、插件 manifest wire、动态插件 artifact 格式或 guest SDK 公开协议。数据权限语义不变，动态 route 继续沿用现有鉴权、权限菜单查询与目标分发边界；缓存语义不变，不新增缓存域或失效路径；`i18n`无运行时用户可见文案和资源影响；开发工具跨平台无影响，未修改脚本、Makefile、CI 或`linactl`。验证：`openspec validate slim-plugin-wasm-route-dispatch --strict`通过。

## 2. WASM host service 公共 helper

- [x] 2.1 将`capabilityContextForHostCall`从用户领域文件迁入`wasm`公共 host service 层，保持函数签名、调用语义和依赖来源不变。
  记录：已将`capabilityContextForHostCall`、`capabilitySourceFromExecution`和`capabilityAuthorizationFromHostServices`迁入`wasm_host_service.go`公共 host service 层，保持小写未导出函数、原签名和基于既有`hostCallContext`的依赖来源不变；未新增运行期依赖、构造函数参数、缓存域或包级默认实例。数据权限、租户、授权快照和错误 envelope 语义不变。
- [x] 2.2 增加或更新静态治理测试，验证公共 helper 不再定义于`wasm_host_service_users.go`，且统一入口仍通过 registry 分发而非 service 级大 switch。
  记录：已在`wasm_host_service_registry_test.go`增加静态测试，验证`capabilityContextForHostCall`不在`wasm_host_service_users.go`定义，并验证`wasm_host_service.go`入口仍调用`dispatchRegisteredHostService(ctx, hcc, request)`且未恢复`switch request.Service`。验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/wasm/hostservicedispatch -count=1`通过。

## 3. 动态 route 瘦身

- [x] 3.1 将`runtime/route.go`中的路由匹配、公开路径解析和 path 参数提取拆入`route_match.go`。
  记录：已新增`route_match.go`承载`dynamicRouteMatch`、`MatchDynamicRoutePath`、`matchDynamicRoute`、`matchDynamicRoutePath`和 path 归一化逻辑。公开路径仍按`/x/{plugin-id}/api/v1/...`匹配，内部 route contract、method 与 path 参数提取语义不变。
- [x] 3.2 将 JWT/session 解析、访问级别判断、角色菜单权限查询和 access context 构造拆入`route_auth.go`。
  记录：已新增`route_auth.go`承载`dynamicRouteClaims`、`dynamicRouteAccessContext`、JWT access token 解析、session touch、角色/菜单权限批量查询和数据范围快照构造。未新增 DAO 查询路径，仍按用户角色、角色菜单、菜单权限三段集合化查询装配，数据权限和租户边界语义不变。
- [x] 3.3 将请求 envelope、header、cookie、query、body 适配拆入`route_envelope.go`，将响应写回拆入`route_response.go`。
  记录：已新增`route_envelope.go`承载 bridge request envelope、header 脱敏、cookie/query/body snapshot 与 request ID 构造；新增`route_response.go`承载 raw bridge response 写回。另将请求上下文状态和动态 route metadata 拆入`route_context.go`，保持同包未导出函数和公开辅助函数语义不变。
- [x] 3.4 增加或更新静态治理测试，验证`route.go`不超过`400`行，并记录拆分未改变 HTTP API、数据权限、缓存和`i18n`语义。
  记录：已在`route_test.go`新增`TestDynamicRouteEntrypointStaysSlim`，验证`route.go`行数不超过`400`，且匹配、鉴权、envelope 和响应写回职责分别归属到拆分文件。当前`route.go`为`250`行。验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -count=1`通过。影响判断：未修改 HTTP API/DTO/路由契约、SQL、前端、动态 guest 协议或用户可见文案；缓存 freshness 检查仍在原 runtime 入口链路中执行；数据权限与权限菜单查询语义保持原实现。

## 4. 验证与审查

- [x] 4.1 运行 WASM 和 runtime 变更包 Go 测试、`openspec validate slim-plugin-wasm-route-dispatch --strict`和`git diff --check`。
  记录：验证通过：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/wasm/hostservicedispatch -count=1`；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -count=1`；`openspec validate slim-plugin-wasm-route-dispatch --strict`；`git diff --check`。按`plugin.md`审查`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`，本次只调整宿主内部`internal/wasm`公共 helper 归属与`internal/runtime/route*.go`文件组织，不修改`pkg/plugin`公开契约、guest SDK、host service 声明格式或中英文文档描述的能力边界，因此无需同步 README。
- [x] 4.2 按`lina-review`执行任务完成审查，记录读取规则、验证证据、DI 来源、缓存一致性、数据权限、开发工具跨平台、测试策略和`i18n`影响判断。
  审查记录：已按`lina-review`任务完成范围审查本变更新增/修改文件：`openspec/changes/slim-plugin-wasm-route-dispatch/**`、`runtime/route.go`、`runtime/route_*.go`、`runtime/route_test.go`、`wasm_host_service.go`、`wasm_host_service_users.go`和`wasm_host_service_registry_test.go`。已读取规则：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/backend-go.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`。结论：未发现阻塞问题。DI 来源：本变更未新增构造函数参数、运行期 service 依赖或启动装配路径；WASM helper 继续使用既有`hostCallContext`与`hostServiceRuntime`，dynamic route 拆分继续使用既有`serviceImpl`字段。缓存一致性：未新增缓存、失效、revision 或跨实例状态；动态 route 仍沿用既有 runtime freshness 和启用状态检查。数据权限：未新增数据操作接口，动态 route 权限查询仍按租户、角色、菜单集合化查询并通过既有`menuFilter`过滤。API 契约：未修改`api/`、DTO、HTTP 方法、公开路径、权限标签或 OpenAPI 元数据。开发工具跨平台：未修改脚本、Makefile、CI 或`linactl`。测试策略：纯内部重构，无 UI 或用户可观察页面变化，未触发 E2E；使用 Go 单元测试、静态治理测试、OpenSpec 严格校验和格式检查覆盖。`i18n`：未新增用户可见文案、翻译资源或 API 文档源文本，无资源影响。额外验证：`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`rg -n "[ \t]+$|\*\*\* Add File|<!--" openspec/changes/slim-plugin-wasm-route-dispatch apps/lina-core/internal/service/plugin/internal/runtime/route_auth.go apps/lina-core/internal/service/plugin/internal/runtime/route_context.go apps/lina-core/internal/service/plugin/internal/runtime/route_envelope.go apps/lina-core/internal/service/plugin/internal/runtime/route_match.go apps/lina-core/internal/service/plugin/internal/runtime/route_response.go`无匹配。剩余风险：工作区包含大量前序方案未提交变更和归档目录，本次审查未回退或重新审查这些既有变更；本变更验证基于当前工作区状态。
