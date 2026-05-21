## ADDED Requirements

### Requirement: linactl 必须统一承载仓库工具实现

系统 SHALL 将镜像构建、动态插件 Wasm 打包与运行时 i18n 治理扫描实现作为 `hack/tools/linactl/internal/` 下的内部子组件维护。`hack/tools/image-builder`、`hack/tools/build-wasm` 与 `hack/tools/runtime-i18n` 不得继续作为仓库默认开发路径中的独立 Go 工具模块存在。

#### Scenario: 镜像构建实现通过 linactl 内部组件执行

- **WHEN** 开发者运行 `linactl image` 或 `linactl image.build`
- **THEN** 命令直接调用 `hack/tools/linactl/internal/imagebuilder` 中的实现
- **AND** 命令不得再通过 `go run ./hack/tools/image-builder` 调用独立工具

#### Scenario: 动态插件 Wasm 打包实现通过 linactl 内部组件执行

- **WHEN** 开发者运行 `linactl wasm`
- **THEN** 命令直接调用 `hack/tools/linactl/internal/wasmbuilder` 中的实现
- **AND** 命令不得再通过进入 `hack/tools/build-wasm` 目录执行 `go run .` 调用独立工具

#### Scenario: 运行时 i18n 治理扫描通过 linactl 内部组件执行

- **WHEN** 开发者运行 `linactl i18n.check`
- **THEN** 命令直接调用 `hack/tools/linactl/internal/runtimei18n` 中的实现
- **AND** 命令不得再通过进入 `hack/tools/runtime-i18n` 目录执行 `go run . scan` 或 `go run . messages` 调用独立工具

### Requirement: 公开开发命令必须保持稳定

系统 SHALL 保持 `make image`、`make image.build`、`make wasm`、`make i18n.check`、`linactl image`、`linactl image.build`、`linactl wasm` 和 `linactl i18n.check` 的公开入口稳定。工具实现迁移不得要求开发者改用新的命令名称。

### Requirement: 工具整合必须更新仓库引用

系统 SHALL 移除默认开发路径中对 `hack/tools/image-builder`、`hack/tools/build-wasm` 和 `hack/tools/runtime-i18n` 的直接引用，包括 `go.work`、CI 夹具、测试辅助和工具文档。
