## 1. 子组件骨架与依赖边界

- [x] 1.1 创建 `pkg/pluginbridge/{contract,codec,artifact,hostcall,hostservice,guest}` 子组件目录，并为每个 Go 包补齐符合规范的包注释和文件用途注释
- [x] 1.2 定义子组件依赖方向，先迁移底层 contract/artifact/codec 能力，确保任何子组件都不 import 根包 `pluginbridge`
- [x] 1.3 将 protobuf wire、WASM section 低层读取等纯实现细节下沉到对应子组件的 `internal` 包，避免新增兜底 `util/common/helper` 包

## 2. 合约、产物和编解码迁移

- [x] 2.1 将 `BridgeSpec`、`RouteContract`、`BridgeRequestEnvelopeV1`、`BridgeResponseEnvelopeV1`、`IdentitySnapshotV1`、`CronContract`、`ExecutionSource` 等稳定合约迁移到 `contract` 子组件
- [x] 2.2 将 bridge request/response/route/identity/HTTP snapshot 编解码迁移到 `codec` 子组件，并保留现有 round trip 测试
- [x] 2.3 将 WASM section 常量、`RuntimeArtifactMetadata`、`ReadCustomSection`、`ListCustomSections` 迁移到 `artifact` 子组件，并更新 i18n、apidoc、runtime 调用路径
- [x] 2.4 增加 facade 与子组件一致性测试，覆盖 bridge envelope 和 WASM section 代表性入口

## 3. hostcall 与 hostservice 迁移

- [x] 3.1 将 host call opcode、状态码、`HostCallResponseEnvelope` 和通用 host call codec 迁移到 `hostcall` 子组件
- [x] 3.2 将 `HostServiceSpec`、capability 推导、host service manifest 编解码和 service/method 常量迁移到 `hostservice` 子组件
- [x] 3.3 将 runtime、storage、network、data、cache、lock、config、notify、cron host service payload codec 迁移到 `hostservice` 子组件，并保留字段编号和默认值语义
- [x] 3.4 更新 Wasm host function、runtime 和 plugindb 相关宿主代码，优先 import `hostcall` / `hostservice` / `codec` 等精确子组件

## 4. guest SDK 与根包 facade

- [x] 4.1 将 guest runtime、guest controller dispatcher、context response helper、BindJSON/WriteJSON、ErrorClassifier 迁移到 `guest` 子组件
- [x] 4.2 将 guest host service client helper 迁移到 `guest` 子组件，并保持 `Runtime()`、`Storage()`、`HTTP()`、`Data()`、`Cache()`、`Lock()`、`Config()`、`Notify()`、`Cron()` 等兼容入口
- [x] 4.3 将根包 `pluginbridge` 收敛为薄 facade，使用 type alias、const alias 和 wrapper 函数转发到子组件，根目录生产源码控制在 1 到 3 个文件
- [x] 4.4 更新动态插件样例或补充兼容测试，确保根包旧入口和 `guest` 子组件入口均可编译使用

## 5. 规范、测试和验证

- [x] 5.1 运行并修复 `go test ./pkg/pluginbridge/...`
- [x] 5.2 运行并修复插件 runtime、WASM host function 和 plugindb 相关测试：`go test ./internal/service/plugin/internal/runtime/... ./internal/service/plugin/internal/wasm/... ./pkg/plugindb/...`
- [x] 5.3 对 `apps/lina-plugins/plugin-demo-dynamic` 执行普通 Go 测试和 `GOOS=wasip1 GOARCH=wasm go build ./...`
- [x] 5.4 运行 `openspec validate pluginbridge-subcomponent-refactor --strict`，确保 proposal、design、specs 和 tasks 可归档
- [x] 5.5 记录 i18n 影响判断：本变更不新增、修改或删除运行时语言包、插件 manifest i18n 或 apidoc i18n 资源
- [x] 5.6 记录缓存一致性判断：本变更不新增业务缓存，不改变插件运行时缓存、i18n 资源缓存或 WASM 编译缓存的权威数据源与失效机制
- [x] 5.7 调用 `lina-review` 完成代码与规范审查

## 实施记录

- i18n 影响判断：本次仅重组 `pkg/pluginbridge` Go 包结构和宿主调用 import，不新增、修改或删除运行时语言包、插件 `manifest/i18n` 资源或 `apidoc` i18n JSON 资源；动态插件 i18n / apidoc WASM section 名称保持不变。
- 缓存一致性判断：本次不新增业务缓存，不改变插件运行时缓存、i18n 资源缓存或 WASM 编译缓存的权威数据源、缓存键、失效触发点、跨实例同步机制和故障降级策略；仅将相关常量和函数迁移到子组件后由原调用点继续引用。
- 验证记录：已通过 `cd apps/lina-core && go test ./pkg/pluginbridge/... ./internal/service/plugin/internal/runtime/... ./internal/service/plugin/internal/wasm/... ./pkg/plugindb/...`。
- 验证记录：已通过 `cd apps/lina-plugins/plugin-demo-dynamic && go test ./... && GOOS=wasip1 GOARCH=wasm go build ./...`。
- 验证记录：已通过 `openspec validate pluginbridge-subcomponent-refactor --strict`。
- lina-review 审查结论：未发现阻断问题；本次未新增 API 端点、SQL、前端 UI 或 E2E 用例，不涉及 RESTful/API DTO/apidoc 翻译变更；未新增或扩大数据操作接口，仅调整插件桥接组件包结构与宿主内部 import；根包生产源码已收敛为 `pluginbridge.go` 与 `pluginbridge_guest_wasip1.go` 两个文件，生产代码已改用精确子组件 import，子组件未 import 根包 `pluginbridge`。

## Feedback

- [x] **FB-1**: 在项目规范和 `lina-review` 技能中补充 bugfix 反馈修复必须具备单元测试或 E2E 测试覆盖的要求

## Feedback Implementation Notes

- FB-1 实施记录：已在 `AGENTS.md` 的开发流程关键规则和 E2E 测试规范中补充 bugfix 反馈修复必须新增或更新单元测试 / E2E 测试的硬性要求。
- FB-1 实施记录：已在 `lina-review` 技能中新增 Bugfix 反馈测试覆盖审查步骤，并将该项加入审查报告模板；缺少测试覆盖或缺少测试运行证据会作为严重问题报告。
- FB-1 验证记录：本次为项目规范和技能说明变更，不涉及后端运行时代码、前端 UI、API、SQL、缓存或 i18n 资源变更，不需要新增单元测试或 E2E 测试；已通过文本定位检查和 `openspec validate pluginbridge-subcomponent-refactor --strict`。
- FB-1 lina-review 审查结论：范围来源为 `git status --short`、`git ls-files --others --exclude-standard` 与本次反馈上下文；未发现阻断问题。本次变更未新增业务 bugfix 修复代码，因此新增的 bugfix 测试覆盖要求不适用于本次规范治理任务本身。
