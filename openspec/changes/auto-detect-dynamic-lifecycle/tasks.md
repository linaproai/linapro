## 1. Guest 元数据发现能力

- [x] 1.1 在 `pluginbridge/guest` 中提取 controller handler 元数据发现入口，复用现有 dispatcher 支持的签名、request type 和 internal path 推导规则
- [x] 1.2 为元数据发现补充单元测试，覆盖 envelope handler、typed handler、非法签名、重复 request type/path 和方法名到 internal path 的推导
- [x] 1.3 保持 `pluginbridge` 根包 facade 兼容，确保现有动态插件 guest 代码继续编译

## 2. build-wasm 自动生成生命周期契约

- [x] 2.1 修改 `hack/tools/build-wasm`，从动态插件 backend controller 元数据中识别支持的 `Before*`/`After*` lifecycle operation
- [x] 2.2 为已发现生命周期方法生成默认 `LifecycleContract`，并写入既有 backend lifecycle WASM custom section
- [x] 2.3 将 `backend/lifecycle/*.yaml` 改为可选 override，按 operation 合并 requestType、internalPath 和 timeoutMs
- [x] 2.4 对不存在 handler 的 override、重复 operation、非法 operation、非法 timeout 和旧 `Can*`/guard 命名输出明确构建错误
- [x] 2.5 补充 `build-wasm` 单元测试，覆盖无 YAML 自动生成、override 覆盖超时、override 指向不存在 handler、重复 operation 和旧命名拒绝

## 3. 官方动态示例插件迁移

- [x] 3.1 删除 `plugin-demo-dynamic/backend/lifecycle/*.yaml` 的重复声明，保留 controller lifecycle 方法作为契约来源
- [x] 3.2 更新 `plugin-demo-dynamic` 中英文 README，说明 lifecycle contracts 由 controller 方法自动发现生成
- [x] 3.3 补充或更新官方 demo 打包测试，确认无 YAML 时 artifact 仍包含 14 个 lifecycle contracts

## 4. 验证与审查

- [x] 4.1 运行 `cd apps/lina-core && go test ./pkg/pluginbridge/... -count=1`
- [x] 4.2 运行 `cd hack/tools/build-wasm && go test ./... -count=1`
- [x] 4.3 运行覆盖动态插件 runtime artifact 解析和生命周期执行的后端测试，至少包含 `cd apps/lina-core && go test ./internal/service/plugin/internal/runtime ./internal/service/plugin -count=1`
- [x] 4.4 运行官方动态示例插件 WASM 构建 smoke，确认产物可生成并包含自动发现的 lifecycle contracts
- [x] 4.5 运行 `openspec validate auto-detect-dynamic-lifecycle --strict`
- [x] 4.6 运行 `git diff --check`
- [x] 4.7 记录 i18n、缓存一致性、数据权限和开发工具脚本影响判断；完成实现后调用 `lina-review` 审查

## Progress Notes

- 2026-05-16: 完成动态插件 lifecycle 构建期自动发现。`pluginbridge/guest` 新增 controller handler metadata 发现入口和 root facade 兼容导出；`build-wasm` 从 backend controller 方法自动生成 `LifecycleContract`，并将 `backend/lifecycle/*.yaml` 降级为按 operation 合并的可选 override，缺失 handler、重复 operation、非法 operation、非法 timeout 和旧 `Can*`/guard 命名会构建失败。自动发现限定在 `backend/controller*.go` 和 `backend/internal/controller/**`，避免 service 层同名方法被误识别；override 合并后必须保留 requestType 或 dispatcher internalPath 中至少一个真实可达 lookup key，避免生成 guest dispatcher 无法命中的契约。官方 `plugin-demo-dynamic` 删除 14 个重复 lifecycle YAML 声明，README/README.zh-CN 同步改为自动发现说明，打包测试确认无 YAML 时 artifact 仍包含 14 个 source-compatible lifecycle contracts。验证通过：`cd apps/lina-core && go test ./pkg/pluginbridge/... -count=1`、`cd hack/tools/build-wasm && go test ./... -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime ./internal/service/plugin -count=1`、`cd hack/tools/build-wasm && go run . --plugin-dir ../../../apps/lina-plugins/plugin-demo-dynamic --output-dir <tmpdir>`、`openspec validate auto-detect-dynamic-lifecycle --strict`、`git diff --check`。
- 2026-05-16: i18n 影响：仅更新 `plugin-demo-dynamic` 中英文 README，不新增或修改运行时 UI 语言包、manifest/i18n 或 apidoc i18n 资源。缓存一致性影响：本变更只改变 WASM 构建期 lifecycle contract 来源，运行时 artifact 校验和、缓存失效和集群收敛继续复用既有插件运行时机制，不新增缓存路径。数据权限影响：不新增 REST API 或业务数据访问；动态 lifecycle handler 若调用 hostServices，仍受既有授权、租户和数据权限边界约束。开发工具脚本影响：修改 Go 版 `hack/tools/build-wasm`，未新增平台专属脚本或默认开发命令入口。
- 2026-05-16: 完成 `lina-review` 审查。审查发现并修复 3 个实现风险：`RegisterController` 二次注册时缺少既有 requestType/path 冲突检查；自动发现扫描整个 backend 可能误识别 service 层同名方法；YAML override 同时覆盖 requestType 和 internalPath 时可能生成 guest dispatcher 无法命中的 lifecycle contract。已分别补充单元测试并复验通过。最终验证通过：`cd apps/lina-core && go test ./pkg/pluginbridge/... -count=1`、`cd hack/tools/build-wasm && go test ./... -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime ./internal/service/plugin -count=1`、`cd hack/tools/build-wasm && go run . --plugin-dir ../../../apps/lina-plugins/plugin-demo-dynamic --output-dir <tmpdir>`、`openspec validate auto-detect-dynamic-lifecycle --strict`、`git diff --check`。
