## 1. 权限与会话边界确认

- [x] 1.1 静态确认动态路由认证中所有角色 DAO、DO、Entity 直连点，并记录迁移目标 owner。
- [x] 1.2 对齐现有`session.Store.TouchOrValidate`语义，确认本变更不新增 session 有效性缓存或写回节流机制。
- [x] 1.3 记录影响分析：数据权限、缓存一致性、`i18n`、开发工具跨平台、测试策略、DI 来源和`apps/lina-core/pkg/plugin`README 是否需要同步。

执行记录：

- 动态路由认证原有角色直连点集中在`apps/lina-core/internal/service/plugin/internal/runtime/runtime_route_auth.go`，包含`sys_user_role`、`sys_role`、`sys_role_menu`和`sys_menu`读取；迁移目标 owner 为`apps/lina-core/internal/service/role`发布的动态路由访问投影契约。静态扫描`rg -n "internal/dao|internal/model/do|internal/model/entity|SysUserRole|SysRoleMenu|SysRole|SysMenu" apps/lina-core/internal/service/plugin/internal/runtime/runtime_route_auth.go`无命中。
- `session.Store.TouchOrValidate`继续作为动态路由唯一 session hot state 校验入口，未新增动态路由本地 session 有效性缓存或写回节流机制；过期、租户不匹配、删除 hot state 和 coordination 读失败 fail-closed 由`internal/service/session`既有测试覆盖。
- 影响分析：数据权限有影响，身份快照继续携带 role owner 的数据范围并传递给 host service context；缓存一致性有影响，role 访问投影复用`permission-access`修订号和租户维度，host call 授权快照限制在单次请求内，datahost 表契约缓存按插件迁移账本和授权方法 fingerprint 失效；`i18n`无 guest 资源治理新增影响，当前保留的运行时行为未新增用户可见文案、错误码、API 文档源文本或语言包资源；开发工具跨平台无新增脚本或构建入口变更；测试策略为后端单元测试、静态扫描、JSON 校验、`make i18n.check`和`openspec validate`；DI 来源为启动期创建的`roleSvc`，经`plugin.New`显式传入`runtime.New`并复用启动期共享实例，`wasm.NewRuntime`不新增配置 reader 依赖；`apps/lina-core/pkg/plugin`中英文 README 已同步说明动态 guest host call 授权快照。

## 2. Role 访问投影契约

- [x] 2.1 在`role`模块发布动态路由访问投影窄契约，返回权限、角色名、数据范围、unsupported 标记和超管标记。
- [x] 2.2 让访问投影复用 token access snapshot、`permission-access`修订号、租户维度和 fail-closed 策略。
- [x] 2.3 将插件 runtime 构造函数改为显式注入访问投影契约，删除动态路由认证对角色治理表的直接访问。
- [x] 2.4 补充权限命中、权限拒绝、租户隔离、权限拓扑变化、freshness 不可确认 fail-closed 和返回对象隔离测试。

执行记录：

- `role.RoleAccessSnapshotService`新增`BuildDynamicRouteAccessProjection`，返回`DynamicRouteAccessProjection`，不暴露`DAO`、`DO`、`Entity`或私有缓存结构。
- 访问投影内部调用`getTokenAccessContext`，通过`dynamicRouteAccessContext`写入 token ID 和租户 scope，复用 token access snapshot、`permission-access`修订号、租户维度、缓存 TTL 和 freshness fail-closed 语义。
- `runtime.RoleAccessProjector`通过`plugin.New`和`runtime.New`显式注入；`runtime_route_auth.go`不再导入角色治理表相关`dao`、`do`或`entity`。
- 覆盖测试包括`role_dynamic_route_access_test.go`、`role_access_cache_test.go`、`runtime_role_access_test.go`和`runtime_route_test.go`，覆盖返回对象隔离、租户桶隔离、freshness 失败、权限命中和权限拒绝。

## 3. 动态路由身份快照迁移

- [x] 3.1 将动态路由身份快照构建改为消费`role`访问投影和共享`session.Store`。
- [x] 3.2 验证登出、强制下线、token 撤销、session 过期、租户不匹配和 Redis hot state 失败时动态路由均拒绝执行。
- [x] 3.3 验证身份快照中的数据范围和 host service context 与宿主受保护 API 同源一致。

执行记录：

- `buildDynamicRouteIdentitySnapshot`继续先调用共享`sessionStore.TouchOrValidate`，再通过`roleAccess.BuildDynamicRouteAccessProjection`构建权限和数据范围快照。
- 登出、强制下线和 token 撤销最终表现为 session hot state 删除或 token 不存在；动态路由沿用`TouchOrValidate`的`false`返回拒绝进入 guest。`session`包既有测试覆盖 session 过期、租户不匹配、coordination hot state 删除和 Redis/coordination 读失败 fail-closed。
- 身份快照的`Permissions`、`RoleNames`、`DataScope`、`DataScopeUnsupported`、`UnsupportedDataScope`和`IsSuperAdmin`均来自 role 投影；`SetUserAccess`写入同源数据范围供后续 host service/data service 使用。

## 4. Guest 资源治理范围收敛

- [x] 4.1 经反馈确认本轮不新增 guest 全局并发、按插件并发、许可等待超时或可配置内存页上限。
- [x] 4.2 移除 guest 资源治理相关运行时配置、错误码、SQL seed、`i18n`资源、README 描述和测试。
- [x] 4.3 保留 WASM runtime 既有默认执行超时和固定单实例内存上限，不新增配置依赖。
- [x] 4.4 记录资源治理后置风险：异常插件流量隔离不在本轮解决，后续如需治理应独立建模。

执行记录：

- 根因：guest 资源治理与权限 owner 收敛属于不同关注点，把并发许可、运行时配置、错误本地化和 SQL seed 纳入本轮会显著抬高动态插件基础链路复杂度。
- 实现取舍：删除`guestExecutionGuard`、`config.GetPluginGuestRuntime`、`plugin.runtime.guest.*`参数、guest 资源错误码和相关测试；`ExecuteBridge`继续直接编译、实例化和执行 guest，并保留请求内 host service 授权快照。
- WASM runtime 仍使用`defaultBridgeExecutionTimeout`和`defaultWasmMemoryLimitPages`作为固定基础约束；这些既有约束不作为本轮可配置治理能力暴露。

## 5. Host call 与 datahost 微优化

- [x] 5.1 在`ExecuteBridge`请求内构建一次 host service 授权快照并传递给 host call context，保持每次调用的 service/method/resource 校验。
- [x] 5.2 为 datahost 表契约缓存设计按插件、表名和迁移状态的键，并在插件 SQL 生命周期成功提交后按插件失效。
- [x] 5.3 补充授权收缩、系统型调用、DDL 后 schema 刷新、缓存不可用回源和数据权限边界测试。
- [x] 5.4 确认 datahost 表契约缓存已在本批落地，不采用后移方案，并在任务记录中说明缓存一致性边界。

执行记录：

- `ExecuteBridge`为每次 guest 执行构建`hostServiceAccessSnapshot`并传递给`hostCallContext`；每次 host call 仍校验`service`、`method`、`resourceRef`或`table`，storage/network/manifest 继续使用既有 matcher，未改变授权来源、数据权限、租户边界或错误 envelope。
- datahost 表契约缓存已在本批落地：`BuildCachedAuthorizedTableContract`接收`pluginID`、`table`和授权方法，缓存键为`pluginID + table + sys_plugin_migration lifecycle ledger fingerprint + authorization methods fingerprint`。权威源为 live schema、host service 授权快照中的 table/method 形状和插件迁移账本；迁移账本读取失败或插件 ID 为空时回源读取 live schema，不缓存失败状态。
- `wasm_host_service_data.go`在 data service dispatch 中传入当前`hostCallContext.pluginID`和请求内授权快照的 data service methods；`handleHostServiceInvoke`仍在 dispatch 前执行 service/method/table 授权，`ExecuteList/Get/Create/Update/Delete/Transaction`仍执行 operation、字段白名单、数据范围、分页上限、软删除和审计治理，因此缓存命中只复用表契约装配，不绕过 data service 安全边界。
- `migration.ExecuteManifestSQLFiles`在 plugin install/upgrade/rollback/uninstall SQL 事务成功提交后按插件调用`InvalidateTableContractCache`，并记录失效 plugin、条目数和原因；缓存键同时包含迁移账本 fingerprint，跨节点或漏失效时也会因账本变化回源重建。
- 新增 datahost 测试覆盖迁移 fingerprint 未变化时复用缓存、DDL 后显式失效刷新 live schema、迁移账本变化后刷新 schema、授权方法收缩后不沿用旧操作集合；host call 授权快照测试继续覆盖请求内复用、授权收缩后新请求使用新快照和系统型调用不伪造用户上下文。当前 host service data 授权形状仅包含 table/method，暂无独立字段级授权清单；字段白名单仍来自 live schema 表契约并由 datahost 执行层校验。

## 6. 验证与审查

- [x] 6.1 运行覆盖`role`、`plugin/internal/runtime`、`wasm`、`datahost`和启动装配变更包的`go test <changed-package> -count=1`。
- [x] 6.2 涉及构造函数、路由绑定或启动装配变更时，运行`cd apps/lina-core && go test ./internal/cmd -count=1`或等价启动绑定测试。
- [x] 6.3 运行`openspec validate plugin-runtime-auth-snapshot-guardrails --strict`。
- [x] 6.4 执行`lina-review`，审查数据权限等价性、缓存一致性七要素、DI 来源检查、错误本地化和测试覆盖。

验证记录：

- `cd apps/lina-core && go test ./internal/service/config ./internal/service/sysconfig -count=1`：通过，覆盖运行时配置解析和 sysconfig 管理校验；guest 资源治理配置已从本变更移除。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`：通过，覆盖 WASM 编译缓存、请求内 host service 授权快照和固定内存上限。
- `cd apps/lina-core && go test ./internal/service/config ./internal/service/sysconfig ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/runtime ./internal/service/role ./internal/service/plugin/internal/datahost ./internal/service/plugin/internal/migration ./internal/service/plugin ./internal/service/session ./internal/service/user ./internal/cmd ./internal/cmd/internal/httpstartup -count=1 -p=1`：通过。
- 仓库根目录执行`make i18n.check`：通过；无 runtime i18n 违规，message coverage 和 frontend key coverage 通过，输出的 module-level`$t()`为仓库既有警告且不属于本次变更文件。
- `jq empty apps/lina-core/manifest/i18n/en-US/config.json apps/lina-core/manifest/i18n/en-US/error.json apps/lina-core/manifest/i18n/zh-CN/config.json apps/lina-core/manifest/i18n/zh-CN/error.json`：通过。
- `rg -n "ON DUPLICATE KEY|INSERT INTO sys_config \([^\n]*\"id\"|INSERT INTO sys_config \([^\n]* id" apps/lina-core/manifest/sql/005-config-management.sql`：无命中；SQL 为 Seed DML，依赖`uk_sys_config_tenant_key`和`ON CONFLICT DO NOTHING`保持幂等，不显式写入自增`id`。
- `git diff --check`：通过。
- `openspec validate plugin-runtime-auth-snapshot-guardrails --strict`：通过。
- `lina-review`：本任务执行内完成审查，结论见最终回复。

## Feedback

- [x] **FB-1**: 合并 guest 护栏 SQL 到既有配置初始化脚本，并移除本次变更中的兼容逻辑。
- [x] **FB-2**: 移除本次迭代新增的 WASM guest 资源治理限制，保留权限、会话、host service 和 datahost 边界收敛。

执行记录：

- 根因：前次实现按`.agents/rules/database.md`的默认迭代 SQL 文件策略新增了`apps/lina-core/manifest/sql/013-plugin-runtime-auth-snapshot-guardrails.sql`，但本项目顶层规范声明全新项目无历史负担，且本轮反馈明确要求不考虑兼容性，应将 guest 护栏 seed 合并到既有配置初始化 SQL。
- SQL 修复：已将四个`plugin.runtime.guest.*`内建配置 seed 合并到`apps/lina-core/manifest/sql/005-config-management.sql`的`sys_config`初始化数据中，并删除独立`013-plugin-runtime-auth-snapshot-guardrails.sql`。SQL 仍为 Seed DML，依赖`uk_sys_config_tenant_key`和`ON CONFLICT DO NOTHING`保持幂等，不显式写入自增`id`，不新增 DAO 生成输入。
- 兼容逻辑清理：移除`wasm_host_call_context.go`中面向旧空`methods`声明的`defaultHostServiceMethods`运行时默认授权兜底；同步移除`pkg/plugin/pluginbridge`host service catalog/validation/capability 层的`DefaultMethods`派生，插件声明和 active release 授权快照现在都只使用显式声明的`methods`。新增`TestHostCallContextRequiresExplicitReadServiceMethods`和`TestValidateHostServiceSpecsRejectsMissingMethods`，验证`hostconfig`和`manifest`仅声明 service/key/path 但缺少 method 时拒绝授权且不派生 capability。`wasm_bridge.go`保留既有`_initialize`执行路径作为当前 WASM ABI 初始化步骤，但去掉 reactor/non-reactor 兼容表述。
- 影响分析：`i18n`无新增影响，本轮未新增或修改运行时文案、错误码、API 文档源文本或语言包资源；缓存一致性无新增影响，host call 请求内授权快照仍为单次执行内派生状态，datahost 缓存策略未改变；数据权限无放宽，host service/data service 仍按显式 method/resource/table 与身份快照校验；开发工具跨平台无影响，未修改脚本、构建或代码生成入口；测试策略为后端单元测试、SQL 静态扫描、OpenSpec 校验、JSON 校验、`i18n`治理检查和完整变更范围 Go 回归。
- 规则加载：本轮已按`AGENTS.md`读取`.agents/rules/openspec.md`、`backend-go.md`、`database.md`、`architecture.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`、`documentation.md`、`testing.md`和`.agents/rules/plugin.md`；未修改`apps/lina-plugins/<plugin-id>/`，未触发插件本地`AGENTS.md`。
- 验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm ./internal/service/config ./internal/service/sysconfig -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/config ./internal/service/sysconfig ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/runtime ./internal/service/role ./internal/service/plugin/internal/datahost ./internal/service/plugin/internal/migration ./internal/service/plugin ./internal/service/session ./internal/service/user ./internal/cmd ./internal/cmd/internal/httpstartup -count=1 -p=1`通过；`openspec validate plugin-runtime-auth-snapshot-guardrails --strict`通过；`git diff --check`通过；`rg -n "ON DUPLICATE KEY|INSERT INTO sys_config \([^\n]*\"id\"|INSERT INTO sys_config \([^\n]* id" apps/lina-core/manifest/sql/005-config-management.sql`无命中；`rg -n "013-plugin-runtime-auth-snapshot-guardrails" apps/lina-core/manifest/sql`无命中。

FB-2 执行记录：

- 根因：本轮用户反馈指出 WASM guest 资源治理并非当前关键能力，而全局并发、单插件并发、许可等待超时、可配置内存页、结构化资源错误、配置 seed 和`i18n`资源会显著增加组件复杂度；该判断成立，资源治理与本变更核心的权限 owner 收敛属于不同关注点。
- 实现调整：删除`guestExecutionGuard`、`config.GetPluginGuestRuntime`、`plugin.runtime.guest.*`运行时参数、guest 资源错误码和相关测试；恢复`wasm.NewRuntime`和`newWasmHostServiceRuntime`为无 guest 配置 reader 的构造路径；`wasm.ExecuteBridge`继续使用固定`defaultBridgeExecutionTimeout`和`defaultWasmMemoryLimitPages`，并保留请求内 host service 授权快照。
- 资源清理：从`apps/lina-core/manifest/sql/005-config-management.sql`移除四条`plugin.runtime.guest.*`seed；从宿主中英文`config.json`和`error.json`移除 guest 配置/错误文案；同步更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，只保留 host service 授权快照说明。
- OpenSpec 调整：`proposal.md`、`design.md`和`specs/plugin-runtime-loading/spec.md`已移除 guest 并发和内存护栏要求，并记录资源治理后置为独立变更。
- 影响分析：`i18n`影响已清理，无新增运行时文案、错误码、API 文档源文本或语言包资源；缓存一致性保留 role 访问投影、host call 请求内授权快照和 datahost 表契约缓存策略，无新增 guest 配置缓存；数据权限无放宽，身份快照、host service 和 data service 仍按 role owner、显式 method/resource/table 与字段白名单校验；开发工具跨平台无影响，未修改脚本、构建或代码生成入口；SQL 无新增 seed，已做幂等和自增主键静态检查；测试策略为后端包测试、pluginbridge 回归、`i18n`检查、OpenSpec 严格校验和静态残留扫描。
- 规则加载：本轮已按`AGENTS.md`读取`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`i18n.md`、`data-permission.md`、`cache-consistency.md`、`database.md`和`documentation.md`；未修改`apps/lina-plugins/<plugin-id>/`，未触发插件本地`AGENTS.md`。
- 验证：`cd apps/lina-core && go test ./internal/service/config ./internal/service/sysconfig ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/runtime ./internal/service/role ./internal/service/plugin/internal/datahost ./internal/service/plugin/internal/migration ./internal/service/plugin ./internal/service/session ./internal/service/user ./internal/cmd ./internal/cmd/internal/httpstartup -count=1 -p=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`通过；`openspec validate plugin-runtime-auth-snapshot-guardrails --strict`通过；`make i18n.check`通过，仍有既有 module-level`$t()`警告且不属于本次变更文件；`jq empty apps/lina-core/manifest/i18n/en-US/config.json apps/lina-core/manifest/i18n/en-US/error.json apps/lina-core/manifest/i18n/zh-CN/config.json apps/lina-core/manifest/i18n/zh-CN/error.json`通过；`git diff --check`通过；`rg -n "plugin\.runtime\.guest|PluginGuestRuntime|guestExecutionGuard|CodePluginGuest|PLUGIN_GUEST|wasm_guest_guard|GetPluginGuestRuntime" apps/lina-core --glob '!**/node_modules/**'`无命中；`rg -n "ON DUPLICATE KEY|INSERT INTO sys_config \([^\n]*\"id\"|INSERT INTO sys_config \([^\n]* id|plugin\.runtime\.guest" apps/lina-core/manifest/sql/005-config-management.sql`无命中。
