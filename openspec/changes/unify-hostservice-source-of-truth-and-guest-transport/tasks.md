# 任务清单：单一事实源与统一 guest 传输

## 1. Descriptor 驱动文档生成

- [x] 1.1 新增 host service README 渲染器，基于`pluginbridge/internal/hostservice`descriptor 生成英文与中文 host service 表格；不保留独立 Go 生成入口
- [x] 1.2 在`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`植入`generated:host-services`标记区块，并人工确认双语事实一致
- [x] 1.3 新增 README 漂移测试，确保 descriptor 渲染结果与当前双语 README 逐字节一致；临时构造漂移验证测试能失败并在验证后恢复

## 2. Descriptor 双向覆盖治理

- [x] 2.1 扩展`hostservice_descriptor_test.go`，反向校验宿主 dispatcher selector 中的 service/method 均存在于 descriptor 且声明`Dispatcher=true`
- [x] 2.2 扩展宿主 service 级 switch 与`dispatchXxxHostService`文件集合校验，确保与 descriptor 的 dispatcher service 集合精确一致
- [x] 2.3 扩展 guest client selector 反向校验，确保 guest selector 中的 method 均存在于 descriptor 且声明`GuestClient=true`
- [x] 2.4 正反向验证覆盖治理测试：当前代码通过；临时构造缺失或孤儿 selector 确认测试能失败，验证后删除临时改动

## 3. Guest host call 传输单轨化

- [x] 3.1 复核根目录残留的`pluginbridge_hostcall_*_wasip1.go`、adapter 和 mirror stub 文件清单，确认`recordstore`执行文件不纳入迁移范围
- [x] 3.2 将 runtime、storage、cache、lock、host config、manifest 和 plugins config 等残留基础能力 client 迁入`internal/domainhostcall`注入式构造，复用现有 invoker 与 protocol codec
- [x] 3.3 更新`pluginbridge_directory.go`和相关测试替身，通过统一 invoker 装配基础能力与领域能力 guest client，保持`pluginbridge.Services`getter 签名不变
- [x] 3.4 删除根目录逐域 WASI 单例、adapter 和 mirror stub 残留，更新 descriptor 非 WASI stub 测试期望为“根目录无逐域 stub 残留”

## 4. 验证与收尾

- [x] 4.1 运行`go test ./pkg/plugin/pluginbridge/... -count=1`和`go test ./pkg/plugin/... -count=1`，确认 descriptor、README 漂移、protocol codec 和 guest client 测试通过
- [x] 4.2 执行动态插件样例普通构建与项目正式`wasip1`构建路径，确认 guest 编译闭包和 host service 调用装配不回归
- [x] 4.3 静态检索确认根目录无逐域`pluginbridge_hostcall_*_wasip1.go`、adapter 和 mirror stub 残留；确认 service/method 字符串与 protocol codec 未被重命名
- [x] 4.4 记录影响分析：无 HTTP API/DTO/路由变更、无 SQL/数据库变更、无前端变更、无运行时`i18n`资源变更、无缓存一致性语义变更、无数据权限路径变更；不新增开发工具、脚本或默认开发命令
- [x] 4.5 运行`openspec validate unify-hostservice-source-of-truth-and-guest-transport --strict`通过，调用`lina-review`完成变更审查

## 执行记录

- C1 文档生成：新增`pkg/plugin/pluginbridge/internal/hostservice`内 README 渲染器，基于 descriptor 维护双语 README host service 生成区块；按反馈删除独立`go run`生成入口，漂移治理由 Go 测试内的 descriptor 渲染器完成，不新增脚本或默认开发命令。
- C1/C2 治理测试：新增 README drift 测试，并扩展 descriptor 覆盖测试为 guest selector、dispatcher selector、宿主 service switch 和 dispatcher function 集合双向校验；`go test ./pkg/plugin/pluginbridge/internal/hostservice -count=1`已通过。
- C1/C2 负向验证：临时将`domainhostcall_runtime.go`中`runtime.info.now`的 guest selector 替换为`runtime.info.uuid`，`go test ./pkg/plugin/pluginbridge/internal/hostservice -run TestHostServiceDescriptorsCoverProtocolGuestAndDispatcher -count=1`按预期失败并报告缺失`HostServiceMethodRuntimeInfoNow`，随后恢复并通过；临时修改英文 README 生成区块中`runtime`方法文本，`go test ./pkg/plugin/pluginbridge/internal/hostservice -run TestHostServiceReadmeBlocksDoNotDrift -count=1`按预期失败并提示从 descriptor 渲染器更新生成区块，随后恢复并通过`go test ./pkg/plugin/pluginbridge/internal/hostservice -count=1`。
- C2 guest transport 迁移：复核残留清单为根目录`pluginbridge_hostcall_{runtime,storage,cache,lock,hostconfig,manifest}_wasip1.go`、`pluginbridge_config_adapter.go`、`pluginbridge_hostconfig_adapter.go`、`pluginbridge_manifest_adapter.go`、`pluginbridge_jobs_adapter.go`以及非 WASI mirror stub；`recordstore`执行文件不纳入迁移范围。已新增`pluginbridge_hostcall_clients.go`统一 public facade，将 runtime、storage、network、cache、lock、host config、manifest 和 plugins config 等基础能力收敛到`internal/domainhostcall`注入式构造，保留`pluginbridge.Services`getter 签名和 protocol codec/wire 格式；删除根目录逐域 WASI 单例、adapter 和 mirror stub，非 WASI 不可用行为只保留在 raw`InvokeHostService`stub。
- 验证收尾：`go test ./pkg/plugin/pluginbridge/... -count=1`和`go test ./pkg/plugin/... -count=1`均通过；动态样例普通模块验证使用`GOWORK=off go test ./... -count=1`在`apps/lina-plugins/linapro-demo-dynamic`通过（根`go.work`不包含该独立模块，直接`go test ./...`会被工作区限制拒绝）；项目正式`wasip1`构建入口`make -C apps/lina-plugins wasm p=linapro-demo-dynamic`通过并输出`temp/output/linapro-demo-dynamic.wasm`。静态检索`find pkg/plugin/pluginbridge -maxdepth 1 -type f \( -name 'pluginbridge_hostcall_*_wasip1.go' -o -name '*_adapter.go' \)`无输出；残留关键词检索仅命中`internal/domainhostcall`私有 helper、hostservice 默认方法治理和测试断言文本；`git diff -- apps/lina-core/pkg/plugin/pluginbridge/protocol`仅包含前序`recordstore`注释路径更新，无 service/method 字符串或 codec 语义变更。
- 影响分析：本变更没有新增 HTTP API、路由、DTO、OpenAPI 元数据或前端调用契约；没有 SQL、DAO、数据库时间字段、软删除或索引变更；没有前端页面、路由、组件、E2E 资产或用户可观察 UI 行为变更；没有运行时`i18n`语言包、API 文档本地化资源或错误消息变更，只有双语技术文档生成区块；没有数据权限过滤路径、授权拒绝策略或数据可见性边界变更；没有缓存权威源、失效、刷新、跨实例同步或陈旧窗口语义变更，cache host service 仅迁移 guest client 构造；没有新增运行期服务依赖，guest client 只显式注入既有 raw/json invoker；按反馈删除独立 Go 生成入口，不新增开发工具、脚本或默认开发命令，跨平台治理验证为 Go drift 测试。
- OpenSpec 与审查：`openspec validate unify-hostservice-source-of-truth-and-guest-transport --strict`通过；`lina-review`范围化审查读取`AGENTS.md`、OpenSpec、文档、架构、插件、后端 Go、测试、开发工具、缓存一致性、数据权限、`i18n`和接口契约规则，审查 C 变更文件与验证证据，未发现阻塞问题。
