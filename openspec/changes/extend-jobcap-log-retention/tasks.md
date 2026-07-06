# Tasks

## Feedback

- [x] **FB-1**: 为 `jobcap` 创建和更新定时任务补齐任务级日志清理策略输入，并覆盖源码插件能力适配和动态 host service 转发测试
- [x] **FB-2**: 为 `jobcap` 的 `Get`、`BatchGet` 和 `List` 查询投影返回任务级日志清理策略，并覆盖源码插件能力适配和动态 host service 查询响应测试
- [x] **FB-3**: 在 `linapro-demo-source` 和 `linapro-demo-dynamic` 示例插件中补齐 `jobcap` 领域能力完整方法使用测试，并校验动态插件 `jobs.*` host service 授权声明

## 执行记录

### FB-1 影响分析

- 修改文件：`apps/lina-core/pkg/plugin/capability/jobcap/jobcap.go`、`apps/lina-core/internal/service/jobmgmt/capabilityadapter/jobmgmt_capability.go`、`apps/lina-core/internal/service/jobmgmt/capabilityadapter/jobmgmt_capability_test.go`、`apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_test.go`。
- 受影响模块：插件 `Jobs()` 领域能力、源码插件能力适配、动态插件 `jobs.create` 和 `jobs.update` host service 转发。
- `i18n` 影响：无运行时用户可见文案、前端 UI、API 文档源文本、插件清单或语言包资源影响。
- 缓存一致性影响：无缓存读写、派生状态或失效机制影响。
- 数据权限影响：复用现有任务创建、更新和可见性校验边界；新增字段不改变租户、角色数据权限或目标可见性策略。
- 开发工具跨平台影响：无开发工具、脚本、构建入口或跨平台执行路径影响。
- 数据库影响：不修改 `sys_job` 表结构，复用既有 `log_retention_override` 字段。
- `DI` 来源检查：未新增运行期依赖、构造函数参数、启动装配或共享实例。
- `apps/lina-core/pkg/plugin` README 同步审查：高层 `Jobs` 能力说明未列字段级输入，现有描述仍准确，不需要修改。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`。

### FB-1 验证

- `go test ./pkg/plugin/capability/jobcap ./internal/service/jobmgmt/capabilityadapter ./internal/service/plugin/internal/wasm -count=1`：通过。
- `go test ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`：通过。
- `openspec validate extend-jobcap-log-retention --strict`：通过。
- `git diff --check`：通过。
- `rg -n "[ \t]+$" openspec/changes/extend-jobcap-log-retention`：通过，未发现未跟踪 OpenSpec 文件尾随空白。
- `find openspec/changes/extend-jobcap-log-retention -type f ... tail -c 1`：通过，未发现未跟踪 OpenSpec 文件缺少末尾换行。
- `make lint`：通过。

### FB-3 根因

前两轮实现已经补齐 `jobcap` 创建、更新和查询投影的日志清理策略，但示例插件只覆盖了内置任务注册和核心 host service 转发测试。`linapro-demo-source` 没有在自身测试中演示源码插件通过 `pluginhost` 的 `Services().Jobs()` 使用完整 `jobcap.Service` 宽接口；`linapro-demo-dynamic` 也只在 `plugin.yaml` 中声明了 `jobs.register`，缺少对运行期 `pluginbridge.Default().Jobs()` 客户端完整方法和对应 `jobs.*` 授权声明的样例级回归测试。

### FB-3 影响分析

- 修改文件：`apps/lina-plugins/linapro-demo-source/backend/plugin_jobcap_usage_test.go`、`apps/lina-plugins/linapro-demo-dynamic/backend/internal/service/dynamic/dynamic.go`、`apps/lina-plugins/linapro-demo-dynamic/backend/internal/service/dynamic/dynamic_host_services.go`、`apps/lina-plugins/linapro-demo-dynamic/backend/internal/service/dynamic/dynamic_jobs_capability_test.go`、`apps/lina-plugins/linapro-demo-dynamic/backend/api/api_contract_test.go`、`apps/lina-plugins/linapro-demo-dynamic/plugin.yaml`、`apps/lina-plugins/linapro-demo-dynamic/README.md`、`apps/lina-plugins/linapro-demo-dynamic/README.zh-CN.md`、`openspec/changes/extend-jobcap-log-retention/proposal.md`、`openspec/changes/extend-jobcap-log-retention/design.md`、`openspec/changes/extend-jobcap-log-retention/tasks.md`。
- 受影响模块：源码示例插件 Jobs registrar 能力使用测试、动态示例插件 `Jobs()` guest client 初始化、动态示例插件 `jobs.*` host service 授权声明和双语说明文档。
- `i18n` 影响：动态示例插件启用了 `i18n`，本次 `plugin.yaml` 只补齐机器可读 host service method 值，README 仅同步技术说明；无运行时 UI 文案、菜单、路由、按钮、API 文档源文本、`manifest/i18n` 或 `apidoc i18n JSON` 资源变更。
- 缓存一致性影响：无缓存读写、缓存失效、派生状态或跨实例一致性影响。
- 数据权限影响：新增测试使用 fake `jobcap.Service` 验证调用契约，不直接访问宿主数据；动态插件仅声明既有受治理 `jobs.*` host service 方法，真实读写、执行和可见性边界仍由宿主 `jobcap` 实现和 WASM dispatcher 授权检查负责。
- 开发工具跨平台影响：不修改脚本、构建入口、CI 或开发工具；插件模块测试因仓库根 `go.work` 未纳入插件模块，使用 `GOWORK=off` 以插件独立模块方式验证。
- 数据库影响：不修改 SQL、DAO、DO、Entity、表结构、索引或查询路径；测试不执行数据库访问。
- HTTP API 契约影响：不新增或修改 HTTP 路由、方法、DTO、`g.Meta` 或响应结构；动态插件新增 API 包契约测试仅读取 `plugin.yaml` 并校验 `jobs.*` 授权声明。
- 接口性能影响：不新增运行期列表、批量、聚合或查询装配路径；测试覆盖 `BatchGet` 和 `List` 的有界输入形态，不引入 `N+1`。
- `DI` 来源检查：动态插件 `serviceImpl` 新增 `jobsSvc jobcap.Service` 字段，由 `New()` 调用 `newJobsHostService()` 初始化；`newJobsHostService()` 复用包级 `guestServices = pluginbridge.Default()` 并调用 `guestServices.Jobs()`。依赖 owner 为 `apps/lina-core/pkg/plugin/pluginbridge` 的 guest-side capability directory，传递路径为 `pluginbridge.Default()` -> `newJobsHostService()` -> `serviceImpl.jobsSvc`，不新增宿主启动期依赖、共享后端、缓存实例或服务图。
- `apps/lina-core/pkg/plugin` README 同步审查：核心 README 已列出 `jobs.batch_get`、`jobs.list`、`jobs.visible.ensure`、`jobs.create`、`jobs.update`、`jobs.delete`、`jobs.run`、`jobs.status.set` 和 `jobs.register`，无需修改。
- E2E 影响：不修改前端页面、路由、表单、表格、按钮或用户操作流程；授权声明通过动态插件 API 契约测试读取真实 `plugin.yaml` 校验，未新增 E2E。
- 文档治理影响：同步修改动态插件 `README.md` 与 `README.zh-CN.md`，保持中英文镜像事实一致；OpenSpec 文档继续使用中文。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/database.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`、`.agents/rules/cache-consistency.md`。

### FB-3 验证

- `GOWORK=off go test ./backend -count=1`（`apps/lina-plugins/linapro-demo-source`）：通过。
- `GOWORK=off go test ./backend/internal/service/dynamic ./backend/api -count=1`（`apps/lina-plugins/linapro-demo-dynamic`）：通过。
- `GOWORK=off go test ./... -count=1`（`apps/lina-plugins/linapro-demo-source`）：通过。
- `GOWORK=off go test ./... -count=1`（`apps/lina-plugins/linapro-demo-dynamic`）：通过。
- `go test ./pkg/plugin/capability/jobcap ./internal/service/jobmgmt/capabilityadapter ./internal/service/plugin/internal/wasm ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`（`apps/lina-core`）：通过。
- `openspec validate extend-jobcap-log-retention --strict`：通过。
- `git diff --check && git -C apps/lina-plugins diff --check`：通过。
- `rg -n "[ \t]+$" openspec/changes/extend-jobcap-log-retention apps/lina-plugins/linapro-demo-source/backend/plugin_jobcap_usage_test.go apps/lina-plugins/linapro-demo-dynamic/backend/internal/service/dynamic/dynamic_jobs_capability_test.go`：通过，未发现尾随空白。
- `find/tail` 末尾换行检查：通过，未发现 OpenSpec 新文件或新增测试文件缺少末尾换行。
- `make lint`：通过。

### FB-2 根因

创建和更新路径已经把 `jobcap.SaveInput.LogRetentionOverride` 转发给宿主任务 owner，但 `jobcap.JobInfo` 查询投影仍只包含 `ID`、`Name`、`Group` 和 `Status`，`BatchGet` 与 `List` 的数据库查询也只投影了 `id`、`name`、`group_id` 和 `status`。因此源码插件能力调用方和动态插件 `jobs.batch_get`、`jobs.list` 响应都无法读取已持久化的 `sys_job.log_retention_override`。

### FB-2 影响分析

- 修改文件：`apps/lina-core/pkg/plugin/capability/jobcap/jobcap.go`、`apps/lina-core/internal/service/jobmgmt/capabilityadapter/jobmgmt_capability.go`、`apps/lina-core/internal/service/jobmgmt/capabilityadapter/jobmgmt_capability_test.go`、`apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_test.go`、`openspec/changes/extend-jobcap-log-retention/design.md`、`openspec/changes/extend-jobcap-log-retention/specs/plugin-host-domain-capabilities/spec.md`、`openspec/changes/extend-jobcap-log-retention/tasks.md`。
- 受影响模块：插件 `Jobs()` 领域能力查询投影、源码插件能力适配、动态插件 `jobs.batch_get` 和 `jobs.list` host service JSON 响应。
- `i18n` 影响：无运行时用户可见文案、前端 UI、API 文档源文本、插件清单或语言包资源影响；仅新增 Go 契约字段和 OpenSpec 说明。
- 缓存一致性影响：无缓存读写、派生状态或失效机制影响。
- 数据权限影响：查询仍复用现有 tenant filter 与 user data-scope 过滤，新增字段只在已可见任务投影中返回，不改变租户、角色数据权限或任务存在性暴露边界。
- 开发工具跨平台影响：无开发工具、脚本、构建入口或跨平台执行路径影响。
- 数据库影响：不修改 `sys_job` 表结构、迁移 SQL、DAO、DO 或 Entity，复用既有 `log_retention_override` 字段；`BatchGet` 与 `List` 仅在同一次投影查询中增加该列，不新增查询次数。
- 接口性能影响：`BatchGet` 继续受 ID 数量上限约束，`List` 继续受分页上限约束；没有逐行数据库查询或前端瀑布式补查，规避 `N+1`。
- `DI` 来源检查：未新增运行期依赖、构造函数参数、启动装配或共享实例。
- `apps/lina-core/pkg/plugin` README 同步审查：高层 `Jobs` 能力说明未列字段级查询投影，现有描述仍准确，不需要修改。
- E2E 影响：无前端页面、路由、表单、表格或用户可观察工作流变化，使用 Go 单元测试覆盖源码插件能力适配和动态 host service 响应。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/database.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`。

### FB-2 验证

- `go test ./pkg/plugin/capability/jobcap ./internal/service/jobmgmt/capabilityadapter ./internal/service/plugin/internal/wasm ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`：通过。
- `openspec validate extend-jobcap-log-retention --strict`：通过。
- `git diff --check`：通过。
- `rg -n "[ \t]+$" openspec/changes/extend-jobcap-log-retention`：通过，未发现未跟踪 OpenSpec 文件尾随空白。
- `find openspec/changes/extend-jobcap-log-retention -type f ... tail -c 1`：通过，未发现未跟踪 OpenSpec 文件缺少末尾换行。
- `make lint`：通过。
