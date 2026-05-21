## 1. 后端路由契约调整

- [x] 1.1 在动态插件 runtime 中定义 public prefix `/x`，并让路由匹配只支持该入口。
- [x] 1.2 调整 HTTP 路由绑定，使 `/x/{pluginId}/...` 挂载宿主统一中间件和动态插件鉴权/权限中间件，且不依赖 `/api/v1` 路由组。
- [x] 1.3 移除 `/api/v1/extensions/{pluginId}/...` 动态插件分发入口，确保旧路径不再执行动态插件 bridge 路由。
- [x] 1.4 更新动态插件 public path 元数据构建，确保 bridge 请求、审计上下文和中间件拿到实际命中的 public path。

## 2. OpenAPI、文档与 i18n 同步

- [x] 2.1 将动态插件 OpenAPI 投影和插件资源列表 public path 生成切换到 `/x/{pluginId}/...`。
- [x] 2.2 更新 apidoc 动态插件路径识别逻辑，使 `/x/{pluginId}/...` 生成和读取稳定插件翻译 key。
- [x] 2.3 同步更新核心 API DTO 文档、动态示例插件 API 注释、`manifest/i18n` 和 packed apidoc i18n JSON 中关于动态插件公开路径的文案。
- [x] 2.4 静态扫描非归档代码和文档中的 `/api/v1/extensions` 引用，确认仅保留迁移说明或明确断言旧路径不再分发的测试引用。

## 3. 动态示例插件与测试更新

- [x] 3.1 将 `linapro-demo-dynamic` 前端请求基地址切换为 `/x/${pluginId}`，并确保插件自有 API 版本可在内部路径中表达。
- [x] 3.2 更新动态插件 runtime、OpenAPI、controller、pluginbridge 编解码相关单元测试，覆盖 `/x` 路径和旧路径不再分发。
- [x] 3.3 更新官方动态插件 E2E 用例，验证 `/x/{pluginId}/...` 可完成已登录动态插件请求。
- [x] 3.4 更新性能审计脚本或端点扫描脚本中的动态插件路径识别和 public path 生成逻辑。

## 4. 验证与审查

- [x] 4.1 运行 `openspec validate rename-dynamic-plugin-route-prefix --strict`。
- [x] 4.2 运行后端 Go 编译/测试门禁，至少覆盖 `apps/lina-core/internal/cmd`、动态插件 runtime、OpenAPI 投影、apidoc 和 pluginbridge 相关变更包。
- [x] 4.3 运行动态示例插件相关测试或 E2E smoke，覆盖新 `/x` 路径的用户可观察行为。
- [x] 4.4 记录 i18n 影响判断：本次需要同步 apidoc i18n JSON 和动态示例插件文案，不新增运行时菜单翻译键时也需明确说明。
- [x] 4.5 记录缓存一致性判断：本次不新增缓存，必须确认 `/x` 路由仍经过插件 runtime freshness、启用状态和派生缓存失效逻辑。
- [x] 4.6 执行 `/lina-review` 审查，修复发现的问题后再标记任务完成。

## 5. 验证记录

- [x] OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] 后端 Go：`cd apps/lina-core && go test ./internal/cmd ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/openapi ./internal/service/apidoc ./internal/controller/plugin ./pkg/pluginbridge/... ./pkg/pluginservice/contract -count=1`
- [x] 路由回归：`cd apps/lina-core && go test ./internal/cmd -run TestDynamicPluginRootRoutesPrecedeSPAFallback -count=1`
- [x] 动态示例插件 E2E smoke：重启本地开发服务到当前工作区代码后，`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- plugin:linapro-demo-dynamic -- --grep "TC-1j"`；验证启用 `linapro-demo-dynamic` 后 `/x/linapro-demo-dynamic/backend-summary` 返回真实 Wasm bridge 响应。
- [x] E2E 治理：`pnpm -C hack/tests test:validate`
- [x] 路径静态扫描：`rg -n '/api/v1/extensions|api/v1/extensions|extensions/\\$\\{|`extensions/|/x//|LegacyRoutePublicPrefix|bindLegacyDynamicPluginAPIRoutes|parts\\[0\\] == "extensions"|/extensions' apps/lina-core apps/lina-plugins/linapro-demo-dynamic hack/tests .agents/skills/lina-perf-audit/scripts -g '!node_modules' -g '!dist' -g '!build' -g '!public/stoplight/**'`；仅剩服务依赖基线中的非路由文件名 `extensions.go` 和旧路径不再分发的负向单元测试引用。
- [x] i18n 影响：已同步核心 API DTO 文档、核心 `manifest/i18n`、核心 packed apidoc i18n、动态示例插件 API 注释和动态示例插件 `manifest/i18n` 中的公开路径文案；本次未新增运行时菜单翻译键。
- [x] 缓存一致性：本次不新增缓存；`/x` 路由继续通过 `PrepareDynamicRouteMiddleware`、`AuthenticateDynamicRouteMiddleware`、`prepareDynamicRouteRuntime` 和 `ensureRuntimeCacheFresh` 相关路径，仍受插件启用状态、runtime freshness、运行时修订号和派生缓存失效机制约束。
- [x] `/lina-review`：审查范围限定为本变更相关路由、runtime、OpenAPI/apidoc、插件示例、E2E 引用、性能审计脚本和 OpenSpec 文档；未发现阻塞问题。已补充 `TestDynamicPluginRootRoutesPrecedeSPAFallback` 防止 `/x/...` 被前端 SPA fallback 接管，并修正旧兼容措辞注释。

## Feedback

- [x] **FB-1**: 将动态插件 API DTO 中的 `gmeta.Meta` 统一改为 `g.Meta`
- [x] **FB-2**: 优化插件管理列表的列对齐、运行时状态列位置、列宽和运行时状态说明
- [x] **FB-3**: 调整插件管理列表列标题居中、名称/版本列宽和基础信息列顺序

## Feedback Verification

- [x] FB-1 后端 Go：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/api/... -count=1`
- [x] FB-1 WASM API 编译烟测：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off GOOS=wasip1 GOARCH=wasm go test ./backend/api/dynamic/v1 -run '^$' -count=1`
- [x] FB-1 动态插件产物构建：`cd apps/lina-plugins && go run ../../hack/tools/linactl wasm p=linapro-demo-dynamic out=temp/output`
- [x] FB-1 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-1 i18n 影响：本次仅统一 Go DTO 元数据嵌入类型，未新增、修改或删除用户可见文案、菜单、路由、API 文档源文本或 apidoc i18n JSON。
- [x] FB-1 缓存一致性：本次不新增或修改缓存逻辑，不影响动态插件 runtime freshness、启用状态或派生缓存失效机制。
- [x] FB-1 `/lina-review`：审查范围限定为动态插件 API DTO、独立模块依赖元数据和 OpenSpec 反馈记录；未发现阻塞问题。`g.Meta` 是 `gmeta.Meta` 的类型别名，路由元数据、权限标签和 apidoc 源文本未发生语义变化；`frame/g` 引入的依赖变更已通过普通 Go 测试和实际 `wasip1/wasm` 产物构建验证。
- [x] FB-2 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-2 前端 i18n/E2E 治理：`pnpm -C hack/tests test:validate`
- [x] FB-2 前端类型检查：`pnpm -C apps/lina-vben -F @lina/web-antd typecheck`
- [x] FB-2 E2E：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/hack/tests/e2e/extension/plugin/TC013-plugin-management-table-layout.ts --grep "TC-13a" --workers=1`
- [x] FB-2 i18n 影响：本次新增插件管理页运行时状态列表头说明文案，已同步 `zh-CN` 和 `en-US` 前端运行时语言包；不涉及菜单、路由、manifest i18n 或 apidoc i18n。
- [x] FB-2 缓存一致性：本次仅调整前端表格列展示、列宽和 tooltip 文案，不新增或修改缓存逻辑，不影响插件 runtime freshness、启用状态或派生缓存失效机制。
- [x] FB-2 数据权限影响：本次不新增、修改或扩大任何数据操作接口；插件列表查询仍复用既有接口与权限边界。
- [x] FB-2 `/lina-review`：审查范围包含插件管理页列配置、中英文运行时语言包、宿主插件页 POM、新增 E2E 用例和 OpenSpec 反馈记录；未发现阻塞问题。变更不涉及 Go 生产代码、REST API、后端数据权限或缓存逻辑，Go 编译门禁不适用。
- [x] FB-3 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-3 前端 i18n/E2E 治理：`pnpm -C hack/tests test:validate`
- [x] FB-3 前端类型检查：`pnpm -C apps/lina-vben -F @lina/web-antd typecheck`
- [x] FB-3 E2E：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/hack/tests/e2e/extension/plugin/TC013-plugin-management-table-layout.ts --grep "TC-13a" --workers=1`
- [x] FB-3 i18n 影响：本次将插件管理列表描述列表头统一为插件专属“插件描述/Plugin Description”，已同步 `zh-CN` 和 `en-US` 前端运行时语言包；不涉及菜单、路由、manifest i18n 或 apidoc i18n。
- [x] FB-3 缓存一致性：本次仅调整前端表格列顺序、表头对齐和列宽，不新增或修改缓存逻辑，不影响插件 runtime freshness、启用状态或派生缓存失效机制。
- [x] FB-3 数据权限影响：本次不新增、修改或扩大任何数据操作接口；插件列表查询仍复用既有接口与权限边界。
- [x] FB-3 `/lina-review`：审查范围包含插件管理页列配置、中英文运行时语言包、宿主插件页 POM、插件管理列表布局 E2E 和 OpenSpec 反馈记录；未发现阻塞问题。变更不涉及 Go 生产代码、REST API、后端数据权限或缓存逻辑，Go 编译门禁不适用。
