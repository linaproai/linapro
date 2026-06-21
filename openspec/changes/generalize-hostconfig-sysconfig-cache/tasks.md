## 1. 规范与设计

- [x] 1.1 编写 OpenSpec proposal、design、tasks 和增量规范，明确源码插件与动态插件的 HostConfig 边界。

## 2. 后端实现

- [x] 2.1 将`config`运行时配置快照从硬编码受管 key 泛化为按租户上下文加载`sys_config`有效行。
- [x] 2.2 调整`GetRaw()`读取顺序，使源码插件可通过`HostConfig()`读取自定义`sys_config`有效 key，同时保留静态配置 fallback 和内置默认语义。
- [x] 2.3 调整`sysconfig`写路径，使创建、更新、导入和删除任意配置记录后按需 bump runtime-config revision。

## 3. 验证

- [x] 3.1 补充单元测试覆盖自定义`sys_config`读取、租户 fallback、revision 刷新和静态配置 fallback。
- [x] 3.2 补充或确认动态插件`hostconfig.get`未授权 key 仍被拒绝，授权 key 可读取自定义`sys_config`。
- [x] 3.3 运行 OpenSpec 严格校验和相关 Go 测试门禁，记录缓存一致性、数据权限、i18n、开发工具和 DI 影响。

## Feedback

- [x] **FB-1**: 自定义`sys_config`只修改 key、不修改 value 时也必须推进 runtime-config revision，避免旧 HostConfig 快照继续返回旧 key。
- [x] **FB-2**: 按反馈将`config_runtime_params.go`、`config_runtime_params_cache.go`和`config_runtime_params_revision.go`合并到`config_raw.go`，并将相关单元测试合并到合适的测试文件。
- [x] **FB-3**: 审查并修正`apps/lina-core/internal/service/config`目录下不符合源码关联命名规范的单元测试文件。
- [x] **FB-4**: 将调度模块内置运行时参数`cron.shell.enabled`和`cron.log.retention`统一修正为`sys.cron.shell.enabled`和`sys.cron.log.retention`。
- [x] **FB-5**: 为`hostconfigcap.Service.Get`增加缺失值默认参数，保持`hostconfig.get`协议最小化。

## 验证记录

- OpenSpec：`openspec validate generalize-hostconfig-sysconfig-cache --strict`通过。
- Go 测试：`cd apps/lina-core && go test ./internal/service/config ./internal/service/sysconfig ./internal/service/plugin/internal/wasm ./internal/cmd -count=1`通过。
- 反馈验证：`cd apps/lina-core && go test ./internal/service/sysconfig -count=1`通过，覆盖自定义`sys_config`改 key 后旧 key 失效、新 key 生效的快照刷新场景。
- 文件合并验证：`cd apps/lina-core && go test ./internal/service/config -count=1`通过，覆盖合并后的`config_raw.go`与`config_raw_test.go`。
- 测试文件命名治理：`for f in apps/lina-core/internal/service/config/*_test.go; do base=${f%_test.go}.go; [ -f "$base" ] || echo "$f -> missing ${base##*/}"; done`无输出；`cd apps/lina-core && go test ./internal/service/config -count=1`通过；`openspec validate generalize-hostconfig-sysconfig-cache --strict`通过。
- FB-3 影响分析：仅调整`apps/lina-core/internal/service/config`目录下测试文件组织和 OpenSpec 任务记录，不改变运行时代码、API、SQL、插件协议或缓存实现；治理类反馈不新增运行时测试场景，使用文件存在性检查、包级 Go 测试和 OpenSpec 严格校验闭环。
- FB-3 i18n：无运行时用户可见文案、API 文档源文本、语言包或翻译缓存变更。
- FB-3 缓存一致性：无生产缓存逻辑变更；仅移动既有缓存相关单元测试到源码关联文件。
- FB-3 数据权限：无数据读写边界、租户边界或插件授权边界变更。
- FB-3 开发工具跨平台：无 Makefile、脚本、CI、代码生成入口或跨平台工具变更。
- FB-3 DI 来源检查：无新增运行期依赖、构造函数参数、服务装配或独立服务图。
- FB-4 根因：调度任务能力早期接入`sys_config`时按模块域写入`cron.shell.enabled`和`cron.log.retention`，在`sys_config`被泛化为宿主运行时配置权威源后未同步纳入`sys.`系统命名空间。
- FB-4 实现：将调度模块内置运行时参数修正为`sys.cron.shell.enabled`和`sys.cron.log.retention`，同步更新`sys_config` seed、受管运行时参数常量、API 文档源、宿主 apidoc 翻译资源、E2E 测试 helper、调度 E2E 用例和 OpenSpec 基线规格；项目无历史包袱，不保留旧 key alias 或兼容读取。
- FB-4 影响分析：修改`apps/lina-core/internal/service/config/config_raw.go`、`apps/lina-core/internal/service/config/config_raw_test.go`、`apps/lina-core/manifest/sql/011-scheduled-job-management.sql`、`apps/lina-core/api/job/v1/job_create.go`、`apps/lina-core/manifest/i18n/zh-CN/apidoc/core-api-job.json`、`hack/tests/support/api/job.ts`、`hack/tests/e2e/scheduler/job/TC*.ts`和相关 OpenSpec 规格；影响调度 Shell 全局开关、调度日志默认保留策略、公开前端配置中调度投影和系统参数管理页按 key 查询；新增单元测试固定受管内置运行时参数必须使用`sys.`前缀。
- FB-4 缓存一致性：权威数据源仍为`sys_config`；运行时读取继续复用 runtime-config 共享 revision、本地`gcache`快照和 tenant scope；本次仅改内置 key 名称，不新增缓存结构、失效触发点或跨实例同步路径。
- FB-4 数据权限：无新增数据操作接口、列表、详情、导出、聚合或执行动作；系统参数查询/更新沿用既有配置管理权限与数据边界。
- FB-4 i18n：更新了 API 文档源文本中的 key 名称及宿主`zh-CN` apidoc 翻译资源；未修改运行时用户可见菜单、按钮、表单或语言包文案；未修改翻译缓存逻辑。
- FB-4 开发工具跨平台：无 Makefile、脚本、CI、代码生成入口或跨平台工具变更。
- FB-4 DI 来源检查：无新增运行期依赖、构造函数参数、服务装配或独立服务图；继续复用启动期注入的`config.Service`实例。
- FB-4 验证：`cd apps/lina-core && go test ./internal/service/config ./internal/service/sysconfig ./internal/service/jobmgmt ./internal/packed ./api/job/v1 ./internal/cmd -count=1`通过；`cd apps/lina-vben && pnpm vitest run apps/web-antd/src/runtime/public-frontend.test.ts --dom`通过；`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit -p tsconfig.json`通过；`openspec validate generalize-hostconfig-sysconfig-cache --strict`通过；`git diff --check`通过；静态检索确认非归档源码、SQL、测试和当前规格中不再引用旧 key。
- FB-5 根因：`hostconfigcap.Service.Get`原始返回值虽然是`*gvar.Var`，但缺失 key 时没有统一的调用侧默认值入口；仅保留`String`、`Bool`、`Int`和`Duration`typed helper 会导致其他可由`gvar.Var`转换的类型缺少同等默认值能力。
- FB-5 实现：将`hostconfigcap.Service.Get`签名固定为`Get(ctx, key, defaultValue any)`，缺失或`nil`值且`defaultValue`非`nil`时返回默认值的`gvar.Var`包装；传入`nil`保持缺失返回`nil`。同步更新源码插件调用点、测试替身、`pluginbridge`适配器、WASM host service 调用和`pkg/plugin`中英文 README。`pluginbridge`适配器同时覆盖 host call 返回 JSON`null`时的默认值语义。动态`hostconfig.get`协议仍不接收默认值，授权继续由`hostServices.resources.keys`控制。
- FB-5 影响分析：修改`apps/lina-core/pkg/plugin/capability/hostconfigcap`、`apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall`、`apps/lina-core/internal/service/plugin/internal/wasm`、宿主插件测试替身、`linapro-monitor-loginlog`、`linapro-ai-core`和`linapro-demo-dynamic`的 HostConfig 调用点，以及当前 OpenSpec 规格和插件 README；新增`domainhostcall_config_test.go`覆盖 bridge-backed 默认值语义；不新增 HTTP API、SQL、前端页面、脚本或代码生成入口。
- FB-5 i18n：无运行时用户可见文案、菜单、按钮、表单、表格、API 文档源文本、插件清单、语言包或翻译缓存变更；仅同步技术 README 和 OpenSpec 文档。
- FB-5 缓存一致性：无缓存权威数据源、revision、失效触发点、跨实例同步、TTL 或故障降级变更；默认值只在既有 HostConfig 读取结果缺失后于调用层包装返回。
- FB-5 数据权限：无新增数据操作接口、列表、详情、导出、聚合、下载、写操作或租户可见性变更；动态插件读取仍按`hostconfig.get`授权 key 拦截，源码插件仍通过稳定`HostConfig()`能力读取。
- FB-5 开发工具跨平台：无 Makefile、脚本、CI、代码生成、资源打包、服务启停或跨平台工具变更。
- FB-5 DI 来源检查：无新增运行期依赖、构造函数参数、服务装配或独立服务图；继续复用启动期注入的`hostconfigcap.Service`、`pluginbridge`适配器和 WASM runtime 的`hostConfigService`。
- FB-5 外部规则加载：已按`AGENTS.md`读取`.agents/rules/openspec.md`、`.agents/rules/backend-go.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、`.agents/rules/i18n.md`和`.agents/rules/api-contract.md`；API 契约规则确认无 HTTP API、DTO、OpenAPI 元数据或前端调用契约影响。
- FB-5 验证：`cd apps/lina-core && go test ./pkg/plugin/capability/hostconfigcap ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend -count=1`通过；`cd apps/lina-plugins/linapro-ai-core && GOWORK=off go test ./backend -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/internal/service/dynamic -count=1`通过；插件测试使用`GOWORK=off`是因为根`go.work`未包含这些独立插件模块；`openspec validate generalize-hostconfig-sysconfig-cache --strict`通过；静态检索确认未残留旧的`hostconfigcap.Service.Get(ctx, key)`签名或可变默认值参数。
- 缓存一致性：权威数据源为`sys_config`；快照复用 runtime-config revision、本地`gcache`和`cachecoord`单机/集群分支；缓存 key 增加 tenant scope；创建、更新、导入和删除任意`sys_config`有效值后推进 revision；写入节点立即清理当前 scope 快照，其他节点沿用共享 revision 和 watcher 刷新。
- 数据权限/租户边界：读取按当前 tenant scope 加载平台行或平台+租户行，租户行覆盖平台行；动态插件仍由`hostconfig.get`的`resources.keys`授权拦截；源码插件通过稳定`HostConfig()`读取，不直接访问 DAO/Entity。
- i18n：无运行时用户可见文案、API 文档源文本、语言包或翻译缓存变更。
- 开发工具跨平台：无 Makefile、脚本、CI、代码生成入口或跨平台工具变更。
- DI 来源检查：无新增运行期依赖；继续复用启动期注入的`config.Service`实例、`hostconfigcap.Service`适配器和 WASM runtime 的`hostConfigService`。未新增独立服务图或中间路径`New()`构造关键服务。
