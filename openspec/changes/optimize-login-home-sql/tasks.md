## 1. 反馈记录

- [x] **FB-1**: 登录后首页 SQL 数量偏高，重点优化插件 release 重复读取和在线会话校验多次 SQL 往返
- [x] **FB-2**: 插件管理页面打开时 `/plugins/dynamic` 运行态列表重复按插件查询 `sys_plugin_release`

## 2. 实现任务

- [x] 2.1 优化 DB-only 在线会话校验，减少有效请求的 `sys_online_session` 查询次数。
- [x] 2.2 为会话校验补充单元测试，覆盖有效、过期、租户不匹配和近期活跃场景。
- [x] 2.3 优化插件 catalog release 读取复用，减少同一请求或列表投影内的 `sys_plugin_release` 重复查询。
- [x] 2.4 为插件 release 读取复用补充单元测试。

## 3. 验证与审查

- [x] 3.1 运行 `openspec validate optimize-login-home-sql --strict`。
- [x] 3.2 运行 `cd apps/lina-core && go test ./internal/service/session -count=1`。
- [x] 3.3 运行覆盖插件 catalog 变更包的 Go 测试。
- [x] 3.4 运行 `cd apps/lina-core && go test ./internal/cmd -count=1` 或等价启动绑定编译烟测。
- [x] 3.5 记录 i18n、缓存一致性、数据权限、开发工具跨平台影响评估，并执行 `lina-review` 审查。

## 4. 完成记录

- i18n：未修改前端页面、公开 API 文档、插件 manifest 文案或运行时语言包；无新增、删除或变更翻译键。
- 缓存一致性：会话校验仍以 `sys_online_session` 数据库行为为权威源；插件 release/dependency 只使用请求级或列表级快照，随上下文释放。平台启用状态快照仅用于运行期读路径，启动一致性校验通过权威启用上下文绕过进程内快照；集群模式仍由既有 runtime cache revision 刷新机制处理跨实例失效。
- 数据权限：未新增或修改业务数据接口；本次只优化鉴权会话与插件元数据内部读取路径，不改变数据权限边界。
- 开发工具跨平台：未新增或修改开发工具、脚本、CI 入口或默认开发命令。
- 验证：`openspec validate optimize-login-home-sql --strict` 通过；`cd apps/lina-core && go test ./internal/service/session ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/cmd -count=1` 通过；`git diff --check` 通过。
- 审查：已按 `lina-review` 范围检查本变更的会话校验、插件元数据读取复用、启动一致性权威读取和测试覆盖，未发现需要阻断的问题。

## 5. 反馈 FB-2 完成记录

- 实现：`/plugins/dynamic` 运行态列表在读取插件 registry 与 release 投影前创建请求级 catalog 快照，并将该快照上下文传递给 runtime upgrade state 投影，避免每个插件重复按 release id 或 plugin/version 查询 `sys_plugin_release`。
- 测试：新增 runtime 单元测试捕获 `ListRuntimeStates` SQL，断言运行态列表不再产生 `sys_plugin_release` 点查。
- i18n：未修改前端页面、API 文档、插件 manifest 文案或语言包；不涉及翻译键新增、修改或删除。
- 缓存一致性：本次只新增请求级只读快照复用，快照随请求释放；权威数据源仍是 `sys_plugin` 与 `sys_plugin_release`，不引入跨请求或跨实例缓存，不改变集群失效模型。
- 数据权限：未新增或修改业务数据操作接口；`/plugins/dynamic` 仍返回插件运行态元数据，不扩大数据权限边界。
- 开发工具跨平台：未新增或修改开发工具、脚本、CI 入口或默认开发命令。
- 验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestListRuntimeStates' -count=1` 通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run '^$' -count=1` 编译烟测通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1` 通过；`cd apps/lina-core && go test ./internal/cmd -count=1` 通过；`openspec validate optimize-login-home-sql --strict` 通过；`git diff --check` 通过。
- 审查：已按 `lina-review` 范围检查本次 runtime 运行态列表快照复用、SQL 捕获测试、Go 编译门禁、i18n、缓存一致性、数据权限和开发工具跨平台影响，未发现阻断问题。
- 剩余风险：`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime ./internal/service/plugin -count=1` 中 runtime 包的 `TestExecuteDynamicWasmBridgeHostCallDemoUsesStructuredHostServices` 与 `TestExecuteDeclaredCronJobUsesWasmBridge` 失败，错误位于 demo dynamic WASM 内部 `dboxFtoa64`/`panicshift` 执行路径；该失败与本次只读 catalog 快照复用无交集，已用聚焦测试和编译烟测覆盖本次变更。
