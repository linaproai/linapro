# 任务清单：修正 pkg/plugin 依赖方向

## 1. ManifestSnapshot 迁入 capmodel（D2）

- [x] 1.1 在`pkg/plugin/capability/capmodel`新增`ManifestSnapshot`类型，字段与 JSON tag 与现`pluginbridge/contract.ManifestSnapshotV1`完全一致，文件归属与注释遵守`.agents/rules/backend-go.md`
- [x] 1.2 将`pluginbridge/contract/contract_lifecycle.go`中的`ManifestSnapshotV1`改为`type ManifestSnapshotV1 = capmodel.ManifestSnapshot`别名，删除原结构体定义，确认`LifecycleRequest`序列化字段不变
- [x] 1.3 `pluginhost/pluginhost_source_plugin_manifest.go`与`pluginhost/internal/manifestview`改为依赖`capmodel.ManifestSnapshot`，`ManifestSnapshot.Values()`返回类型同步切换，删除两处对`pluginbridge/contract`的 import
- [x] 1.4 检查`pluginhost_source_plugin_test.go`等测试对 contract manifest 类型的引用，统一改为 capmodel；宿主侧`internal/service/plugin/internal/{sourceupgrade,catalog,runtime,wasm}`中 manifest snapshot 引用按需更新 import（行为不变）
- [x] 1.5 运行`go build ./...`与`pkg/plugin`、`internal/service/plugin`相关`go test`，确认 D2 步骤独立可编译可通过

## 2. recordstore 迁移到 pluginbridge（D1）

- [x] 2.1 将`pkg/plugin/capability/recordstore/`（含`internal/plan`、全部测试）整体移动到`pkg/plugin/pluginbridge/recordstore/`，包名保持`recordstore`，文件内容除 import 路径外不变
- [x] 2.2 更新`pluginbridge`根包（`pluginbridge.go`、`pluginbridge_directory.go`、stub 与 wasip1 文件中涉及处）对 recordstore 的 import 路径
- [x] 2.3 更新宿主侧消费方 import：`internal/service/plugin/internal/datahost/datahost.go`、`datahost_plan.go`及`pluginbridge/internal/hostservice`描述符测试
- [x] 2.4 更新动态插件样例`apps/lina-plugins/linapro-demo-dynamic/backend/internal/service/dynamic/dynamic_host_services.go`的 import（修改前确认插件根目录无`AGENTS.md`本地规范；如有，先读取并遵守）
- [x] 2.5 更新`capability/capability_test.go`中对 recordstore 的测试引用；确认`capability/**`非测试代码已无任何`pluginbridge`引用
- [x] 2.6 运行宿主`go build ./...`、`go vet ./...`与`go test ./pkg/plugin/... ./internal/service/plugin/...`；对动态插件样例执行普通 Go 构建与`GOOS=wasip1 GOARCH=wasm`构建，验证 guest 执行路径可用

## 3. import 边界治理测试（D3）

- [x] 3.1 在`pkg/plugin`新增边界治理测试：解析`capability/**`非测试源文件 import，断言无`lina-core/pkg/plugin/pluginbridge`与`lina-core/pkg/plugin/pluginhost`前缀；解析`pluginhost/**`非测试源文件 import，断言无`lina-core/pkg/plugin/pluginbridge`前缀；失败信息包含违规文件与 import 路径
- [x] 3.2 验证治理测试的正反向行为：当前代码全部通过；临时构造违规 import 确认测试能捕获（验证后移除）
- [x] 3.3 运行`go test ./pkg/plugin/...`确认治理测试纳入常规测试门禁

执行记录：`go test ./pkg/plugin -run TestPluginPackageDependencyDirection -count=1`通过；临时加入`capability/tmp_boundary_violation.go`和`pluginhost/tmp_boundary_violation.go`后测试按预期失败并输出违规文件与 import 路径，验证后已删除临时文件；`go test ./pkg/plugin/... -count=1`通过。

## 4. README 与文档同步（D4）

- [x] 4.1 按`.agents/rules/documentation.md`更新`pkg/plugin/README.md`与`README.zh-CN.md`：组件职责表中 recordstore 归属调整；新增"动态插件专属能力"小节，记录`Runtime`/`Network`/`RecordStore`仅在`pluginbridge.Services`的原因与新增能力归属判定参照
- [x] 4.2 检查`pkg/plugin`目录下其他 README 或子包文档是否引用旧`capability/recordstore`路径并同步修正

执行记录：`pkg/plugin/README.md`与`README.zh-CN.md`已同步新增动态插件专属能力说明，`RecordStore`明确归属`pluginbridge.Services`和`pkg/plugin/pluginbridge/recordstore`；`find apps/lina-core/pkg/plugin -name 'README*.md'`确认本目录仅有双语 README；静态检索确认双语 README 与`pkg/plugin`Go 源码无旧`capability/recordstore`路径残留。

## 5. 验证与收尾

- [x] 5.1 全量验证：宿主`go build ./...`、`go vet ./...`、`go test ./...`（或项目约定的等价 make 目标）通过；动态插件样例 wasip1 构建通过
- [x] 5.2 静态检索确认仓库内（含文档与脚本）无`capability/recordstore`残留引用；确认`pluginhost`非测试代码无`pluginbridge`残留引用
- [x] 5.3 记录影响分析结论：无 i18n 影响（无运行时文案变更）、无缓存一致性影响、无数据权限路径变更（datahost 授权执行逻辑仅改 import）、无开发工具跨平台影响（治理测试为纯 Go 测试）、无新增运行期依赖（DI 装配不变，记录无影响判断）
- [x] 5.4 运行`openspec validate fix-plugin-pkg-dependency-direction --strict`通过，调用`lina-review`完成变更审查

执行记录：宿主`go build ./...`、`go vet ./...`、`go test ./... -count=1`均通过；动态插件样例`GOWORK=off go build ./...`通过，正式`go -C hack/tools/linactl run . wasm p=linapro-demo-dynamic out=temp/output`通过并产出`temp/output/linapro-demo-dynamic.wasm`，该文件经`file`确认为 WebAssembly 二进制。直接`GOOS=wasip1 GOARCH=wasm go build ./...`会缺少生成期`backend/plugin_wasip1.go`，因此采用项目正式`linactl wasm`路径作为 wasip1 构建验证。静态检索确认生产代码、README、E2E 测试模板和脚本中无旧`capability/recordstore`引用；归档历史与基线`openspec/specs`中的旧路径由本变更归档阶段更新，不作为当前活跃变更实现残留。`pluginhost`与`capability`非测试 Go 源码均无违规`pluginbridge` import。影响分析：本变更无运行时用户可见文案或语言包变更，`i18n`无影响；无缓存状态、失效或刷新路径变更，缓存一致性无影响；`datahost`授权执行逻辑仅更新 recordstore import，数据权限路径和拒绝策略无变化；治理测试为 Go 单元测试，未新增脚本或跨平台工具入口，开发工具跨平台无影响；未新增运行期依赖、构造函数参数或 DI 装配，DI 来源检查结论为无新增运行期依赖。

审查记录：已运行`openspec validate fix-plugin-pkg-dependency-direction --strict`并通过；已按`lina-review`完成本变更审查，结论见本次执行输出。
