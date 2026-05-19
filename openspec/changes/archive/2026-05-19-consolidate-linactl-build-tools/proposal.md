## Why

`hack/tools/image-builder`、`hack/tools/build-wasm`与`hack/tools/runtime-i18n`已经主要通过`linactl`间接使用，但仍保留独立`Go`模块、独立入口和重复的测试/文档引用，增加了开发工具维护面。
将这些项目专用组件收敛到`hack/tools/linactl/internal/`可以统一长期维护工具边界，降低`go.work`、`CI`夹具和命令入口的复杂度。

## What Changes

- 将镜像构建实现迁移到`hack/tools/linactl/internal/imagebuilder`，`linactl image`与`linactl image.build`直接调用该内部包。
- 将动态插件`Wasm`打包实现迁移到`hack/tools/linactl/internal/wasmbuilder`，`linactl wasm`直接调用该内部包。
- 将运行时`i18n`治理扫描实现迁移到`hack/tools/linactl/internal/runtimei18n`，`linactl i18n.check`直接调用该内部包。
- 移除`hack/tools/image-builder`、`hack/tools/build-wasm`与`hack/tools/runtime-i18n`独立工具模块及其`go.work`条目。
- 更新`linactl`依赖、测试、`CI`夹具、`E2E`辅助代码和中英文工具文档中的旧路径引用。
- 保持公开命令入口不变：`make image`、`make image.build`、`make wasm`、`make i18n.check`、`linactl image`、`linactl image.build`、`linactl wasm`和`linactl i18n.check`继续作为用户使用入口。

## Capabilities

### New Capabilities

- `linactl-build-tool-consolidation`：定义`linactl`统一承载镜像构建、动态插件`Wasm`打包与运行时`i18n`治理扫描能力的开发工具边界。

### Modified Capabilities

- `release-image-build`：镜像构建命令的实现边界从独立`image-builder`工具调整为`linactl`内部组件，但公开命令语义保持不变。

## Impact

- 影响`hack/tools/linactl`、`hack/tools/image-builder`、`hack/tools/build-wasm`、`hack/tools/runtime-i18n`、根`go.work`、`hack/tools/README.md`、`hack/tools/README.zh-CN.md`、`linactl`文档、相关`CI`夹具和直接调用旧工具路径的测试辅助代码。
- `linactl`将接受对`apps/lina-core/pkg/pluginbridge`的编译期依赖，因为该工具只服务当前`LinaPro`仓库。
- 不涉及后端运行时服务、`REST API`、数据库结构、前端页面、用户可见运行时文案、运行时缓存或数据权限逻辑。
