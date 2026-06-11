## 1. Catalog 与描述源收敛

- [x] 1.1 新增`pkg/plugin/pluginbridge/protocol/hostservices`catalog 类型、payload kind、资源类型和 service/method 元数据结构。
- [x] 1.2 将现有`pluginbridge/internal/hostservice`descriptor 改为从公开 catalog 派生，删除独立手写 service/method 表。
- [x] 1.3 更新 manifest validation、capability derivation 和 host service descriptor 测试，确保行为与现有授权快照语义一致。
- [x] 1.4 增加静态检索或单元测试，阻断宿主 WASM 包导入`pluginbridge/internal/hostservice`。

## 2. 普通领域 JSON Envelope

- [x] 2.1 在`pluginbridge/protocol`新增通用 host service JSON request/response envelope 和 round trip 测试。
- [x] 2.2 在 catalog 中标记普通领域使用 JSON envelope，特殊服务使用专用 codec，并建立特殊服务白名单测试。
- [x] 2.3 迁移`users`作为代表性普通领域，验证 guest typed client、host dispatch handler 和领域 DTO 映射可用。
- [x] 2.4 批量迁移其余普通领域服务，避免新增普通领域 per-domain 专用`protowire`codec。

## 3. WASM Host Service Registry

- [x] 3.1 新增`internal/service/plugin/internal/wasm/hostservicedispatch`registry、handler context、response helper 和重复注册检测。
- [x] 3.2 将`wasm_host_service.go`改为 envelope 解码、授权校验、registry lookup 和统一错误响应，不再维护 service 级大 switch。
- [x] 3.3 将领域 dispatch handler 通过父包显式注册适配层接入`hostservicedispatch`registry，避免扩大 WASM 私有执行上下文公开面。
- [x] 3.4 保留`storage`、`cache`、`lock`、`data/recordstore`、`network`和必要`ai`方法的专用 codec 语义，但统一接入 registry。
- [x] 3.5 增加 registry 单元测试，覆盖已注册方法成功分发、未知 service/method 拒绝、重复注册失败和缺失依赖失败。

## 4. Guest Client 与能力目录

- [x] 4.1 将普通领域 guest client 从平铺`domainhostcall_<x>.go`逐步收敛到领域子包或等价清晰结构，并保持 typed client 对插件作者可用。
- [x] 4.2 更新`pluginbridge_directory.go`或能力 guest 目录装配，只负责注入统一 invoker，不承载领域调用逻辑。
- [x] 4.3 校验`pkg/plugin/capability/<x>cap`仍是领域契约 owner，`pluginbridge`不重新发布业务能力语义。

## 5. 治理验证与回归

- [x] 5.1 增加 catalog 双向覆盖测试，校验 guest client、protocol payload、专用 codec、registry handler 和 catalog 声明一致。
- [x] 5.2 增加静态扫描，确认`wasm_host_service.go`不存在 service 级 host service switch，普通领域没有未标记的 per-domain 专用 codec。
- [x] 5.3 运行`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`。
- [x] 5.4 运行`cd apps/lina-core && go test ./pkg/plugin/... -count=1`。
- [x] 5.5 运行`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`。
- [x] 5.6 运行动态插件样例普通 Go 测试和`GOOS=wasip1 GOARCH=wasm`构建，确认 guest 编译闭包不回归。
  - 验证记录：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./... -count=1`通过。
  - 验证记录：`cd hack/tools/linactl && go run . wasm p=linapro-demo-dynamic out=temp/output`通过，生成`temp/output/linapro-demo-dynamic.wasm`。直接裸`GOOS=wasip1 GOARCH=wasm go build`不是项目维护入口，因为动态插件样例需要`linactl wasm`先生成 guest dispatcher。
- [x] 5.7 运行`openspec validate consolidate-hostservice-domain-bridge --strict`。
  - 验证记录：`openspec validate consolidate-hostservice-domain-bridge --strict`通过。

## 6. 影响分析与审查记录

- [x] 6.1 记录本变更无 HTTP API、SQL、前端页面、运行时用户可见文案、插件实例目录资源影响。
  - 影响分析：本变更只调整`apps/lina-core`内动态插件 host service 协议描述、guest bridge 编码和 WASM dispatch 内部结构；未修改 HTTP API 路由、请求响应 DTO、SQL 迁移、前端页面、运行时用户可见文案、插件 manifest 资源或插件实例目录资源。
  - `i18n`判断：无运行时用户可见文案、API 文档源文本、语言包、插件清单文案或翻译缓存变更；仅新增中文 OpenSpec 文档记录。
- [x] 6.2 记录数据权限影响：动态插件经 host service 访问数据的 handler 必须保持与宿主 API 等价的数据权限、租户边界和目标可见性校验。
  - 影响分析：registry 重构只改变`service/method`到 handler 的查找方式，不改变 handler 内部调用的`capability/<x>cap`契约、授权快照、租户上下文或既有目标可见性校验。`users`路径仍保留`ensureHostCallUsersVisible`等可见性检查，组织和租户路径仍通过原有领域服务边界处理。
- [x] 6.3 记录缓存一致性影响：registry handler 必须复用启动期共享实例或共享后端，且不改变缓存权威源、失效触发点、跨实例同步或陈旧窗口。
  - 影响分析：本变更未新增缓存、快照、订阅、失效或刷新逻辑；`hostservicedispatch`registry 保存的是函数注册表，不持有业务缓存状态。handler 继续复用`hostCallContext`中由宿主启动期传入的共享服务、运行时快照和后端实例，不改变缓存权威源、失效触发点、跨实例同步机制或可接受陈旧窗口。
  - DI 来源检查：未新增关键运行期服务依赖。新增 registry 由`wasm_host_service_registry.go`在父包显式构造，handler 注册函数复用现有`hostServiceRuntime`和`hostCallContext`传递路径；未引入`init()`自注册、包级默认实例或调用关键服务`New()`创建独立服务图。
- [x] 6.4 记录开发工具跨平台影响：本变更首期不新增代码生成入口或脚本；若后续引入生成器，必须另行满足跨平台工具规则。
  - 影响分析：本变更没有新增或修改`Makefile`、`make.cmd`、`hack/tools/`、脚本或 CI；`linactl wasm`仅作为已有跨平台 Go 工具入口执行验证。后续如把 catalog 派生为代码生成器，需要单独走开发工具规则并提供 Windows、Linux、macOS 等价入口。
- [x] 6.5 完成实现和验证后调用`lina-review`进行代码、规范和规则矩阵审查。
  - 审查记录：已按`AGENTS.md`规则矩阵读取并遵守`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`和`.agents/instructions/markdown-format.instructions.md`，并使用`goframe-v2`技能与`lina-review`技能。
  - 审查范围：本变更新增 OpenSpec 目录、`protocol/hostservices`catalog、通用 JSON envelope、`hostservicedispatch`registry、WASM registry 显式注册、普通领域`users`/`org`/`tenant`guest 与 host JSON envelope 迁移及治理测试；`.agents/rules/plugin.md`为本任务外已有暂存改动，未纳入本次实现修改。
  - 审查结论：未发现阻塞问题。WASM 入口文件不再维护 service 级 switch；registry 使用显式注册且无`init()`自注册；`internal/hostservice`descriptor 从公开 catalog 派生；普通领域旧专用 codec 名称只保留在 deny-list 治理测试中；`pluginbridge_directory.go`只注入 invoker 并选择 typed client；`capability/<x>cap`仍是领域契约 owner。
  - 规则域结论：无 HTTP API、SQL、前端、运行时用户可见文案或插件实例目录资源变更；未触发 E2E 质量审查，原因是本次为动态插件 host service 内部 Go 协议与 dispatch 重构，无用户可观察页面或端到端工作流变化。
  - 验证证据：`go test ./pkg/plugin/... -count=1`、`go test ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/wasm/hostservicedispatch -count=1`、`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./... -count=1`、`cd hack/tools/linactl && go run . wasm p=linapro-demo-dynamic out=temp/output`和`openspec validate consolidate-hostservice-domain-bridge --strict`均通过。
