## 1. 插件服务契约

- [x] 1.1 在`plugin`根服务中新增真实消费者 facet、`ManagedJobQuery`和`UpdateStatusOptions`。
- [x] 1.2 用`ListManagedJobs`替换多个插件 job 查询方法，并删除旧 job 查询根方法。
- [x] 1.3 用`UpdateStatus`替换`Enable`、`Disable`和`SetStatus`根方法，并删除 auth hook 包装方法与无生产入口的`SyncSourcePlugins`。

## 2. 消费者迁移

- [x] 2.1 迁移插件管理控制器、HTTP 启动编排、`apidoc`、`cron`和`jobhandler`到统一`Service`入口、包内私有 facet 或消费者本地窄接口。
- [x] 2.2 迁移`RuntimeDelegate`绑定目标和调用点，保持启动期同一插件服务实例共享。

## 3. 测试和静态治理

- [x] 3.1 更新受影响的测试替身和断言，确保测试只实现实际依赖的方法。
- [x] 3.2 使用静态检索确认旧根方法和宽接口依赖不再被生产调用点引用。

## 4. 验证和审查

- [x] 4.1 运行 OpenSpec 严格校验和相关 Go 测试。
- [x] 4.2 记录 DI、缓存一致性、数据权限、`i18n`、开发工具和 E2E 影响分析，并执行`lina-review`。

## 验证记录

- `rg "HandleAuthLogin|HandleAuthLogout" apps/lina-core -g'*.go' -n`：无结果。
- `rg "SyncSourcePlugins\\(" apps/lina-core -g'*.go' -n`：无结果。
- `rg "plugin(Management|Startup|RuntimeHTTP|Integration|Job)Svc" apps/lina-core/internal/cmd/internal/httpstartup -n`：无结果。
- `rg "pluginSvc\\s+pluginsvc\\.Service" apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go -n`：确认`httpRuntime`只保留统一`pluginSvc pluginsvc.Service`字段。
- `rg "\\.ListExecutableJobs|\\.ListExecutableJobsByPlugin|\\.ListJobDeclarationsByPlugin|\\.ListInstalledJobDeclarations" apps/lina-core/internal apps/lina-core/pkg -g'*.go' -n`：仅剩`plugin`根服务内部委托和`internal/integration`自测。
- `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/auth ./internal/service/cron ./internal/service/jobhandler ./internal/controller/plugin ./internal/cmd -count=1`：通过。
- `openspec validate shrink-plugin-service-contract --strict`：通过。

## 影响分析与审查记录

- DI：无新增运行期依赖；`pluginsvc.New`仍在`httpstartup`启动装配中创建同一个插件服务实例，`httpRuntime`只保存统一`pluginSvc pluginsvc.Service`字段；`RuntimeDelegate`绑定目标使用统一`pluginsvc.Service`入口，仍绑定同一实例。
- 缓存一致性：无新增缓存、失效策略或跨实例协调语义；`ManagedJobQuery`和 facet 迁移只改变 Go 类型边界，继续复用原插件 runtime cache 和启动期共享服务实例。
- 数据权限：未新增或修改 HTTP API、数据库查询、资源可见性规则或插件数据访问授权；插件管理控制器仍走原管理方法和原权限链。
- `i18n`：未修改运行时用户可见文案、API 文档源文本、错误消息或语言包。
- 开发工具跨平台：未修改脚本、构建、代码生成、CI 或跨平台工具入口。
- E2E：本次为内部 Go 契约治理，无前端页面、用户可观察行为或端到端工作流变化；使用 Go 测试、静态检索和 OpenSpec 校验覆盖。
- `apps/lina-core/pkg/plugin`文档：本次未修改公共插件能力契约、`pluginbridge`或动态`hostServices`wire method，不需要同步`README.md`/`README.zh-CN.md`。
- `lina-review`：已按`AGENTS.md`读取命中规则并审查本次变更范围，未发现阻塞问题。

## Feedback

- [x] **FB-1**: 将`plugin.Service`拆分出来的 facet 接口改为私有定义，包外只保留导出的`Service`入口。
- [x] **FB-2**: 将`httpRuntime`中拆散的插件 facet 字段收敛为单个`pluginSvc pluginsvc.Service`统一入口。
- [x] **FB-3**: 删除`httpstartup`中剩余的插件 runtime HTTP 和 integration 分类接口，helper 直接使用统一`pluginsvc.Service`入口。
- [x] **FB-4**: 删除插件管理控制器本地重复的`pluginManagementService`接口，`NewV1`和控制器字段直接使用统一`pluginsvc.Service`入口。
- [x] **FB-5**: 将插件根服务状态变更方法命名统一为宿主内部风格的`UpdateStatus`，保留公共 capability 接口`SetStatus`命名。
- [x] **FB-6**: 删除`RuntimeDelegate`单用途窄接口，直接绑定统一`pluginsvc.Service`入口。

## Feedback 验证记录

- `rg "pluginsvc\\.(ManagementService|StartupService|RuntimeHTTPService|IntegrationService|JobService|StateService|CapabilityEnvService|TenantLifecycleService)" apps/lina-core -g'*.go' -n`：无结果。
- `rg "type (ManagementService|StartupService|RuntimeHTTPService|IntegrationService|JobService|StateService|CapabilityEnvService|TenantLifecycleService)" apps/lina-core/internal/service/plugin -g'*.go' -n`：仅剩`apps/lina-core/internal/service/plugin/internal/runtime/runtime.go`中的内部 runtime 子包`IntegrationService`，不是`plugin`根服务 facet。
- `rg "type (managementService|startupService|runtimeHTTPService|integrationService|jobService|stateService|capabilityEnvService|tenantLifecycleService)" apps/lina-core/internal/service/plugin/plugin.go -n`：确认`plugin`根服务 facet 均为私有定义。
- `rg "plugin(Management|Startup|RuntimeHTTP|Integration|Job)Svc" apps/lina-core/internal/cmd/internal/httpstartup -n`：无结果。
- `rg "pluginRuntimeHTTPService|pluginIntegrationService" apps/lina-core/internal/cmd/internal/httpstartup -n`：无结果。
- `rg "pluginSvc\\s+pluginsvc\\.Service" apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go -n`：确认`httpRuntime`只保留统一`pluginSvc pluginsvc.Service`字段。
- `rg "pluginManagementService" apps/lina-core/internal/controller/plugin -n`：无结果。
- `rg "pluginSvc\\s+pluginsvc\\.Service" apps/lina-core/internal/controller/plugin/plugin_new.go -n`：确认插件管理控制器只使用统一`pluginsvc.Service`字段。
- 旧状态选项类型静态检索：无结果。
- `rg "\\.SetStatus\\(" apps/lina-core/internal/service/plugin apps/lina-core/internal/controller/plugin -g'*.go' -n`：无结果，插件根服务和插件管理控制器不再调用`SetStatus`。
- `rg "UpdateStatus\\(" apps/lina-core/internal/service/plugin apps/lina-core/internal/controller/plugin -g'*.go' -n`：确认插件根服务接口、实现、控制器和相关测试统一使用`UpdateStatus`。
- `rg "SetStatus" apps/lina-core/pkg/plugin/capability apps/lina-core/internal/service/user/capabilityadapter apps/lina-core/internal/service/jobmgmt/capabilityadapter -g'*.go' -n`：确认公共 user/job capability 接口和适配器仍保留`SetStatus`命名。
- `rg "type runtimeDelegateService|service runtimeDelegateService|BindService\\(service runtimeDelegateService\\)|serviceSnapshot\\(\\) runtimeDelegateService" apps/lina-core -g'*.go'`：无结果。
- `rg "BindService\\(service Service\\)|service Service|serviceSnapshot\\(\\) Service" apps/lina-core/internal/service/plugin/plugin_runtime_delegates.go`：确认`RuntimeDelegate`直接绑定统一`Service`入口。
- `cd apps/lina-core && go test ./internal/cmd/internal/httpstartup ./internal/cmd -count=1`：通过。
- `cd apps/lina-core && go test ./internal/controller/plugin -count=1`：通过。
- `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/auth ./internal/service/cron ./internal/service/jobhandler ./internal/controller/plugin ./internal/cmd/internal/httpstartup ./internal/cmd -count=1`：通过。
- `cd apps/lina-core && go test ./internal/service/plugin ./internal/cmd/internal/httpstartup -count=1`：通过。
- `openspec validate shrink-plugin-service-contract --strict`：通过。

## Feedback 影响分析

- DI：无新增运行期依赖；启动装配仍只创建同一个`pluginsvc.New`实例，`httpRuntime`和插件管理控制器均保存统一`pluginSvc pluginsvc.Service`字段并传递该实例；测试替身通过嵌入`pluginsvc.Service`保持只覆盖实际调用方法。
- 缓存一致性：无新增缓存、失效策略、快照或跨实例协调语义；仅收敛 Go 接口导出面。
- 数据权限：未新增或修改 HTTP API、数据库查询、数据读写路径、资源可见性或插件授权边界。
- `i18n`：未修改运行时用户可见文案、API 文档源文本、错误消息、插件清单或语言包资源。
- 开发工具跨平台：未修改脚本、构建、代码生成、CI 或跨平台工具入口。
- E2E：无前端页面、用户可观察行为或端到端工作流变化；使用 Go 测试、静态检索和 OpenSpec 校验覆盖本次治理反馈。
- FB-5：仅调整插件根服务内部 Go 契约命名和 OpenSpec 治理文本；无新增运行期依赖、缓存、数据权限、`i18n`、开发工具或 E2E 影响，公共 capability 契约命名保持不变。
- FB-6：删除`RuntimeDelegate`单用途窄接口后仍绑定启动期同一个`pluginsvc.New`实例，不新增运行期依赖、不创建独立服务图、不改变插件 hook、菜单过滤、provider env、租户生命周期或 OpenAPI 投影行为；无缓存、数据权限、`i18n`、开发工具或 E2E 影响。
