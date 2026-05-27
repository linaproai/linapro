## 1. Host Service 契约与作用域配置

- [x] 1.1 调整源码插件 host services 契约，提供`Config()`、`HostConfig()`和`Manifest()`三个读取入口，并更新源码插件 registrar 的注入结构。
- [x] 1.2 实现插件作用域配置 resolver，支持开发阶段`apps/lina-plugins/<plugin-id>/manifest/config/config.yaml`、生产阶段 GoFrame 配置根`plugins/<plugin-id>/config.yaml`和动态 artifact 默认配置来源。
- [x] 1.3 确保`manifest/config/config.example.yaml`仅作为模板文件，不参与运行期自动默认值读取。
- [x] 1.4 为`HostConfig()`实现宿主公开配置白名单读取，拒绝空键、`.`、完整配置扫描和未公开宿主配置键。
- [x] 1.5 为`Config()`和`HostConfig()`补齐结构化错误、日志和调用来源诊断，避免向插件暴露宿主私有配置路径或敏感值。

## 2. Manifest 资源读取

- [x] 2.1 实现`HostServices.Manifest()`源码插件适配器，支持读取和扫描当前插件实际存在的`manifest/metadata.yaml`及声明型 YAML 资源。
- [x] 2.2 实现 manifest 路径治理，拒绝空根读取、绝对路径、路径穿越、Windows drive path、URL、跨插件路径和未授权路径。
- [x] 2.3 保持`manifest/config/`、`manifest/sql/`和`manifest/i18n/`专用治理边界，避免通用`Manifest()`绕过配置、SQL 或 i18n 管线。
- [x] 2.4 增加 manifest 资源读取单元测试，覆盖合法读取、结构体扫描、路径穿越拒绝、绝对路径拒绝、URL 拒绝和跨插件读取拒绝。

## 3. 动态插件授权、Guest SDK 与运行时资源视图

- [x] 3.1 调整动态插件`plugin.yaml`的`hostServices`校验和授权快照，使`config`、`hostConfig`和`manifest`仅接受`methods: [get]`。
- [x] 3.2 将 guest SDK 的`Exists`、`String`、`Bool`、`Int`、`Duration`和`Scan`等便捷方法映射到`get`，并在 SDK 或共享适配层完成类型转换。
- [x] 3.3 实现动态插件`hostConfig.resources.keys`和`manifest.resources.paths`资源边界校验，并把调用审计记录到统一 host service 摘要。
- [x] 3.4 扩展动态插件运行时资源视图，使 active release artifact 暴露实际携带的`manifest/config/`、`manifest/metadata.yaml`和其他声明型 manifest 资源。
- [x] 3.5 将动态插件 manifest 资源视图和 artifact 默认配置视图绑定 active release checksum 或 generation，并接入`plugin-runtime`修订失效。

## 4. 动态插件打包与示例插件资源

- [x] 4.1 调整动态插件构建器的`go:embed`资源快照生成逻辑，保留`manifest/config/config.example.yaml`、`manifest/config/config.yaml`和声明型 manifest 资源路径。
- [x] 4.2 调整目录扫描兼容回退逻辑，使未声明`go:embed`的动态插件也能打包实际存在的`manifest/config/`和声明型 YAML 资源。
- [x] 4.3 更新官方源码插件和动态示例插件的目录结构，按需新增或迁移`manifest/config/config.example.yaml`模板，不提交`manifest/config/config.yaml`或无业务用途的`manifest/metadata.yaml`。
- [x] 4.4 更新示例插件`plugin.yaml`的`hostServices`声明，使用`config.get`、`hostConfig.get`表达读取权限，仅在存在声明型 manifest 资源时声明`manifest.get`。

## 5. 测试、验证与治理记录

- [x] 5.1 增加或更新 Go 单元测试，覆盖插件配置 resolver 的开发路径、生产路径、artifact 默认配置、模板不参与运行期读取和宿主配置隔离。
- [x] 5.2 增加或更新动态插件 host service 测试，覆盖`methods: [get]`接受、typed 方法拒绝、未授权 key/path 拒绝和空键完整快照拒绝。
- [x] 5.3 增加或更新动态插件构建 smoke 或资源快照测试，覆盖`manifest/config/`和实际存在的声明型 YAML 资源被打包并可由运行时资源视图读取。
- [x] 5.4 运行覆盖变更包的 Go 编译门禁、动态插件构建 smoke、OpenSpec 严格校验和空白检查。
- [x] 5.5 在执行记录中补充影响评估：`i18n`资源无直接影响、数据权限无新增业务数据访问、缓存一致性绑定`plugin-runtime`修订、开发工具跨平台影响和测试策略。
- [x] 5.6 完成实现后执行`lina-review`审查，并按审查结论修复阻断问题。

## Feedback

- [x] **FB-1**: 移除不必要的插件`metadata.yaml`文件，明确`metadata.yaml`是可选声明资源而非所有插件必备资源。
- [x] **FB-2**: 将插件`manifest/config/config.yaml`作为本地实际配置加入忽略规则，仅提交`config.example.yaml`模板。
- [x] **FB-3**: 删除宿主`manifest/config`中已经迁移为插件自有配置的插件业务配置内容，保留宿主插件治理配置。
- [x] **FB-4**: 将动态插件非 WASM host service stub 抽象到主框架`guest`组件，降低插件开发样板代码。
- [x] **FB-5**: 将宿主公开配置读取能力从`HostRuntime`统一重命名为`HostConfig`，避免与动态插件`Runtime`服务混淆。
- [x] **FB-6**: 将`pluginbridge`低层协议实现包收敛到`pluginbridge/internal`，保留根包作为稳定公共入口。
- [x] **FB-7**: 源码示例插件生命周期日志应复用宿主`LifecycleHook`常量，避免插件注册处硬编码生命周期名称。
- [x] **FB-8**: 将`plugindb`非公开实现收敛到`plugindb/internal`，避免低层 host/shared 子包成为公共 API。
- [x] **FB-9**: 修正`pluginhostservices`组件源码文件命名，确保 service 目录文件统一使用组件名前缀。
- [x] **FB-10**: 动态插件 WASM 构建不应引入 host-side PostgreSQL 驱动依赖。

## Execution Record

- 规则加载：已按`AGENTS.md`读取`openspec`、`plugin`、`backend-go`、`testing`、`architecture`、`data-permission`、`cache-consistency`、`dev-tooling`、`i18n`、`api-contract`、`database`、`documentation`、`frontend-ui`规则，并使用`openspec-apply-change`、`lina-feedback`、`lina-review`、`lina-e2e`和`goframe-v2`技能。
- 反馈根因：插件配置归属已迁移到插件自有`manifest/config`后，宿主`apps/lina-core/manifest/config`继续维护`linapro-monitor-server`等插件业务配置会造成宿主侵入和双重事实来源；批量添加`metadata.yaml`会把可选声明资源误变成所有插件必备资源；提交实际`config.yaml`存在敏感配置误提交风险。
- 实现结果：宿主配置仅保留`plugin.autoEnable`、`plugin.dynamic.storagePath`、`plugin.allowForceUninstall`等插件治理配置，删除宿主监控插件业务配置副本；插件实际`manifest/config/config.yaml`已通过`apps/lina-plugins/.gitignore`忽略，仅提交`config.example.yaml`模板；无业务用途的插件`metadata.yaml`已删除。
- 行为修复：`linapro-tenant-core`不再通过插件作用域`Config()`读取宿主 JWT 配置，改为通过显式宿主`Auth()`能力签发和撤销 impersonation token，避免扩大`HostConfig()`或宿主配置读取面。
- `i18n`影响：本轮配置反馈不新增运行时 UI 文案、API 文档源文本、语言包或插件清单文案；插件专属日志与页面本地化仍由插件自有 E2E 覆盖，宿主根级 i18n E2E 不再硬编码插件专属治理数据。
- 缓存一致性影响：配置和`Manifest()`读取保持只读；动态 artifact 默认配置和 manifest 资源视图绑定 active release checksum/generation，并由既有`plugin-runtime`修订触发失效。
- 数据权限影响：本轮反馈不新增业务数据读取、写入、导出、聚合或候选项接口；impersonation 仍由宿主 auth 能力和平台管理员边界控制。
- 开发工具跨平台影响：为`linactl`隐藏 GoFrame 命令补齐命令文件治理，使用 Go 工具链验证；未新增平台专属脚本或默认开发路径依赖。
- 验证结果：`cd apps/lina-core && go test ./... -count=1`通过；`cd hack/tools/linactl && go test ./... -count=1`通过；`go run ./hack/tools/linactl test.go`通过，覆盖 13 个模块、126 个测试包和 168 个编译 smoke 包；`pnpm -C hack/tests test:validate`通过；`go run ./hack/tools/linactl test`通过，E2E 结果为 522 passed、8 skipped；`openspec validate introduce-plugin-config-manifest-resources --strict`通过；`git diff --check && git -C apps/lina-plugins diff --check`通过。
- FB-4 反馈根因：动态插件示例在`!wasip1`普通 Go 编译路径中本地维护`unsupported*HostService`，只是为了让包测试避开真实 WASM host call；该逻辑与具体插件业务无关，属于`pluginbridge/guest`SDK 应统一维护的跨插件兜底能力。
- FB-4 实现结果：`guest`组件新增共享 host service 契约和非 WASM unsupported stub，覆盖`runtime`、`storage`、`network`、`data`、`cache`、`lock`、`config`、`notify`、`cron`、`hostConfig`和`manifest`；root`pluginbridge`在 WASM 和非 WASM 下统一转发`HostConfig()`与`Manifest()`；示例动态插件删除本地`dynamic_host_services_stub.go`和`dynamic_host_services_wasip1.go`，改为单个无 build tag 的`dynamic_host_services.go`绑定。
- FB-4 可复用逻辑审查：示例插件剩余窄接口用于业务服务单元测试替身，不上收到主框架；`plugindb`已有主框架非 WASM stub，不存在插件重复实现；记录结构体与`map[string]any`互转、时间解析和生命周期日志属于示例业务投影或插件路由逻辑，暂不抽象为公共 API。
- FB-4 影响分析：本轮只调整动态插件 guest SDK 和示例插件 host service 绑定，不新增 HTTP API、SQL、前端 UI、用户可见文案、缓存失效路径或业务数据访问面；`i18n`资源无影响，数据权限无新增数据操作影响，缓存一致性无新增缓存状态影响，开发工具跨平台无新增脚本或平台专属入口影响。
- FB-4 验证结果：`cd apps/lina-core && go test ./pkg/pluginbridge/... -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && go test ./backend/internal/service/dynamic -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && go test ./backend/api/... ./backend/internal/controller/dynamic -count=1`通过；`make wasm p=linapro-demo-dynamic out=temp/output`通过；`openspec validate introduce-plugin-config-manifest-resources --strict`通过；相关变更范围`git diff --check`通过。
- FB-6 反馈根因：`pluginbridge`已通过根包 facade 暴露动态插件 SDK、host service 契约、artifact section 常量和协议编解码能力，但宿主运行时仍直接 import `pluginbridge/artifact`、`pluginbridge/codec`、`pluginbridge/hostcall`和`pluginbridge/hostservice`等低层实现子包，导致这些协议实现包成为公共 API，扩大了插件开发者和构建工具可依赖的边界。
- FB-6 实现结果：保留`pluginbridge/contract`和`pluginbridge/guest`作为公共契约与 guest SDK；将`artifact`、`codec`、`hostcall`和`hostservice`移动到`apps/lina-core/pkg/pluginbridge/internal/`；宿主运行时、`plugindb`、`pluginhost`、`apidoc`、`i18n`和 WASM host service 调用全部改走根`pluginbridge`facade；同步更新 panic allowlist 路径和 facade 测试说明。
- FB-6 影响分析：本轮属于公共包边界治理，不改变 WASM ABI、host service payload、动态插件授权模型、HTTP API、SQL、前端 UI 或运行时用户可见文案；`i18n`资源无影响；数据权限无新增业务数据读写、导出、聚合或候选项接口影响；缓存一致性无新增缓存、失效或跨实例状态影响；开发工具跨平台影响仅为`linactl`继续通过根`pluginbridge`facade 引用公开契约，未新增脚本或平台专属入口。
- FB-6 验证结果：`cd apps/lina-core && go test ./pkg/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./pkg/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run '^$' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/apidoc ./internal/service/i18n -run '^$' -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -run TestProductionPanicsMatchAllowlist -count=1`通过；`cd hack/tools/linactl && go test ./... -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && go test ./... -count=1`通过；`openspec validate introduce-plugin-config-manifest-resources --strict`通过；相关变更范围`git diff --check`通过；`cd apps/lina-core && go list ./pkg/pluginbridge/artifact ./pkg/pluginbridge/codec ./pkg/pluginbridge/hostcall ./pkg/pluginbridge/hostservice`确认旧公开子包路径不可再列出，`go list ./pkg/pluginbridge/internal/artifact ./pkg/pluginbridge/internal/codec ./pkg/pluginbridge/internal/hostcall ./pkg/pluginbridge/internal/hostservice`确认新内部实现包存在。
- FB-7 反馈根因：`linapro-demo-source`源码示例插件生命周期注册日志直接传入`"BeforeInstall"`、`"Upgrade"`、`"AfterTenantDelete"`等字符串，而宿主`pluginhost`已经发布`LifecycleHook`命名类型和生命周期常量；示例插件重复硬编码会绕过后端枚举治理，并在生命周期名称调整时形成双重事实来源。
- FB-7 实现结果：`apps/lina-plugins/linapro-demo-source/backend/plugin.go`中的生命周期日志调用全部改为复用`pluginhost.LifecycleHook*`常量，日志 helper 参数从`string`收紧为`pluginhost.LifecycleHook`，输出时通过`operation.String()`保持原有日志格式。
- FB-7 影响分析：本轮只调整源码示例插件后端日志参数类型和调用点，不改变生命周期注册顺序、回调行为、HTTP API、SQL、前端 UI、插件清单、WASM ABI、host service payload 或用户可见运行时文案；`i18n`资源无影响；数据权限无新增业务数据读写、导出、聚合或候选项接口影响；缓存一致性无新增缓存、失效或跨实例状态影响；开发工具跨平台无新增脚本、命令或默认开发路径依赖。
- FB-7 验证结果：`cd apps/lina-plugins && go test lina-plugin-linapro-demo-source/backend -count=1`通过；静态扫描确认`apps/lina-plugins/linapro-demo-source/backend/plugin.go`中不再保留生命周期日志裸字符串。
- FB-5 反馈根因：`HostRuntime`名称容易与动态插件`Runtime`服务混淆，而`HostEnvironment`又会误导为泛化环境变量读取能力；实际需求是读取宿主显式公开的配置键，应命名为`HostConfig`并与插件作用域`Config()`、运行期`Runtime`能力清晰分离。
- FB-5 实现结果：源码插件契约、宿主适配器、WASM host service、`pluginbridge`guest SDK、动态插件`plugin.yaml`声明、示例 API 文档和 OpenSpec 文档统一使用`HostConfig`/`hostConfig`命名；内部协议服务标识规范化为小写`hostconfig`，能力标识为`host:hostconfig`。行为保持不变，仍只允许通过`resources.keys`和宿主白名单读取公开配置键，不提供`Scan`、`Data`、根节点读取或任意宿主配置树访问。
- FB-5 影响分析：本轮涉及插件宿主边界、动态插件 host service 授权声明、Go 后端公共契约、WASM host service、API 文档源文本和`apidoc`本地化资源；不新增 HTTP 路由、SQL、前端 UI 或用户可见运行时文案。数据权限无新增业务数据读写、导出、聚合或候选项接口；缓存一致性无新增缓存状态、失效或跨实例同步路径；开发工具跨平台无新增脚本、命令或默认开发路径依赖。
- FB-5 验证结果：`cd apps/lina-core && go test ./pkg/pluginbridge/... ./pkg/pluginservice/... ./internal/service/config ./internal/service/plugin/internal/wasm ./internal/service/pluginhostservices ./internal/service/plugin ./internal/cmd -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && go test ./backend/api/... ./backend/internal/controller/dynamic ./backend/internal/service/dynamic -count=1`通过；`go run ./hack/tools/linactl i18n.check`通过；`openspec validate introduce-plugin-config-manifest-resources --strict`通过；`git diff --check && git -C apps/lina-plugins diff --check`通过；静态扫描确认除`tasks.md`反馈记录保留历史来源名和被排除的备选命名外，代码、规范正文和文件名无`HostRuntime`、`HostEnvironment`或`host-runtime`残留。
- FB-8 反馈根因：`plugindb`根包已经提供动态插件作者使用的链式数据访问入口，但 typed plan、JSON 编解码、host-side DB 包装和审计上下文仍放在`pkg/plugindb/shared`与`pkg/plugindb/host`公共子包下，导致宿主执行细节成为可被外部包直接依赖的公共 API。
- FB-8 实现结果：将`shared`计划模型与 host-side DB 治理实现分别移动到`pkg/plugindb/internal/plan`和`pkg/plugindb/internal/host`；`plugindb`根包保留插件作者使用的`Open()`链式 API，并提供宿主执行层需要的窄 facade（审计上下文、host DB、typed plan 解码和校验）；`datahost`与测试改为通过根包 facade 访问，不再直接依赖低层子包；`pluginbridge` data codec 移除 typed plan 解码 helper，避免桥接协议包反向依赖`plugindb`内部计划实现。
- FB-8 影响分析：本轮属于公共包边界治理，不改变 HTTP API、WASM host service payload、动态插件授权模型、SQL、前端 UI、运行时用户可见文案或数据权限规则；`i18n`资源无影响；数据权限无新增业务数据读取、写入、导出、聚合或候选项接口影响；缓存一致性无新增缓存、失效或跨实例状态影响，既有 host DB 句柄缓存只移动目录且语义不变；开发工具跨平台无新增脚本、命令或平台专属入口。
- FB-8 验证结果：`cd apps/lina-core && go test ./pkg/plugindb/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/datahost -count=1`通过；`cd apps/lina-core && go test ./pkg/pluginbridge/internal/hostservice -count=1`通过；`cd apps/lina-core && go test ./pkg/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/datahost ./internal/cmd -count=1`通过；`openspec validate introduce-plugin-config-manifest-resources --strict`通过；相关变更范围`git diff --check`通过；`cd apps/lina-core && go list ./pkg/plugindb/host ./pkg/plugindb/shared`确认旧公开子包路径不可再列出；静态扫描确认 Go 代码无`pkg/plugindb/host`或`pkg/plugindb/shared`导入。
- FB-9 反馈根因：`apps/lina-core/internal/service/pluginhostservices`属于`service/`目录下的`pluginhostservices`组件，但`adapters.go`、`adapters_test.go`和`session_data_scope_test.go`未使用组件名前缀，违反后端 Go 规则中 service 组件源文件必须以组件名作为前缀的文件组织要求。
- FB-9 实现结果：将上述文件分别重命名为`pluginhostservices_adapters.go`、`pluginhostservices_adapters_test.go`和`pluginhostservices_session_data_scope_test.go`；未修改运行时逻辑、接口契约、测试逻辑或源码注释语义。
- FB-9 影响分析：本轮属于源码文件命名治理，不改变 HTTP API、WASM host service payload、动态插件授权模型、SQL、前端 UI、运行时用户可见文案或数据权限规则；`i18n`资源无影响；数据权限无新增业务数据读取、写入、导出、聚合或候选项接口影响；缓存一致性无新增缓存、失效或跨实例状态影响；开发工具跨平台无新增脚本、命令或平台专属入口。
- FB-9 验证结果：静态文件名检查确认`apps/lina-core/internal/service/pluginhostservices`下所有 Go 源文件均为`pluginhostservices.go`或`pluginhostservices_*.go`；`cd apps/lina-core && go test ./internal/service/pluginhostservices -count=1`通过；`openspec validate introduce-plugin-config-manifest-resources --strict`通过。
- FB-10 反馈根因：GitHub Actions job `77570966000`在`make test.go plugins="1" race=true verbose=true`中运行`internal/service/plugin/internal/runtime`测试时，多处测试需要通过`linactl wasm`构建`apps/lina-plugins/linapro-demo-dynamic`；`apps/lina-core/pkg/plugindb/plugindb_host.go`缺少`!wasip1`构建约束，导致动态插件 guest 的`GOOS=wasip1 GOARCH=wasm`构建把 host-side `plugindb/internal/host`、GoFrame PostgreSQL 驱动和`github.com/lib/pq`纳入依赖图，最终在`github.com/lib/pq@v1.10.9/connector.go:107:13`触发`undefined: userCurrent`。
- FB-10 实现结果：为`plugindb_host.go`增加`//go:build !wasip1`，确保宿主数据库 facade 只参与非 WASI 构建；新增`TestWasip1GuestBuildDoesNotImportHostDatabaseDependencies`，通过`GOOS=wasip1 GOARCH=wasm go list -deps -tags wasip1 -json ./pkg/plugindb`断言 guest 依赖图不包含`plugindb/internal/host`、`github.com/lib/pq`和 GoFrame PostgreSQL 驱动；补齐插件运行时测试夹具中的 WASM host service 配置，使依赖`config`、`hostConfig`、`manifest`、`cache`、`lock`、`notify`和`storage`的 runtime 测试可单独运行、顺序无关。
- FB-10 影响分析：本轮只调整 Go 构建边界和测试 fixture，不新增或修改 HTTP API、WASM payload、动态插件授权声明、SQL、前端 UI、插件清单或运行时用户可见文案；`i18n`资源无影响；数据权限无新增业务数据读取、写入、导出、聚合或候选项接口影响，既有 data host service 授权边界不变；缓存一致性无新增生产缓存行为，测试 fixture 仅使用单机`kvcache`实例；开发工具跨平台影响为恢复既有`linactl wasm`的 Go 工具链 WASI 构建路径，未新增脚本或平台专属入口；接口契约无影响。
- FB-10 验证结果：`cd apps/lina-core && go test ./pkg/plugindb -run 'TestWasip1GuestBuildDoesNotImportHostDatabaseDependencies|TestQueryBuilderBuildsTypedPlan|TestTransactionRejectsMultipleTables' -count=1`通过；`cd apps/lina-core && GOOS=wasip1 GOARCH=wasm go list -deps -tags wasip1 -f '{{.ImportPath}}' ./pkg/plugindb`确认依赖图不再包含 host-side PostgreSQL 依赖；`make wasm p=linapro-demo-dynamic out=temp/output`通过；`cd apps/lina-core && go test ./pkg/plugindb ./pkg/plugindb/internal/host ./internal/service/plugin/internal/datahost -count=1`通过；`cd apps/lina-core && go test -race ./pkg/plugindb ./pkg/plugindb/internal/host ./internal/service/plugin/internal/datahost -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog|TestExecuteDynamicWasmBridgeReturnsGuestResponse|TestExecuteDynamicWasmBridgeHostCallDemoUsesStructuredHostServices|TestExecuteDynamicWasmBridgeCreatesDemoRecord|TestExecuteDeclaredCronJobUsesWasmBridge|TestBundledDynamicSampleDeclaresBeforeAndAfterLifecycleCallbacks' -count=1`通过；同一 runtime 子集加`-race`通过；`cd apps/lina-core && go test -p=1 ./internal/service/plugin/... ./pkg/plugindb/... -count=1`通过；`make test.go plugins="1" race=true verbose=true`通过，覆盖 13 个模块、126 个测试包和 172 个编译 smoke 包；`openspec validate introduce-plugin-config-manifest-resources --strict`通过；`git diff --check && git -C apps/lina-plugins diff --check`通过。
