## Context

当前根`Makefile`已经通过`linactl`承载长期维护的跨平台开发命令。`linactl image`和`linactl image.build`仍通过`go run ./hack/tools/image-builder`调用独立工具；`linactl wasm`仍通过进入`hack/tools/build-wasm`目录执行`go run .`来打包动态插件；`linactl i18n.check`仍通过进入`hack/tools/runtime-i18n`目录分别执行`go run . scan`和`go run . messages`来完成运行时`i18n`治理检查。

这些独立工具本质上已经是`LinaPro`仓库专用能力。继续保留独立模块会让`go.work`、`CI`夹具、测试辅助和文档路径保持多套入口。用户已确认可以接受`build-wasm`成为`linactl`的编译期依赖，因此本次不再为了`linactl`的通用轻量性保留独立模块边界。`runtime-i18n`依赖更轻，且公开使用入口已经是`make i18n.check`或`linactl i18n.check`，适合一并收敛为内部组件。

## Goals / Non-Goals

**Goals:**

- 将镜像构建与动态插件`Wasm`打包实现迁移到`hack/tools/linactl/internal/`下的职责明确子包。
- 将运行时`i18n`治理扫描实现迁移到`hack/tools/linactl/internal/runtimei18n`。
- 保持公开命令语义稳定，避免用户从`make`或`linactl`入口感知迁移。
- 移除旧独立工具模块，减少`go.work`、`README`、`CI`和测试辅助中的路径分叉。
- 保留现有单元测试覆盖，并补齐`linactl`侧对合并后命令行为的验证。

**Non-Goals:**

- 不改变镜像构建参数、`Docker`构建语义、动态插件 artifact 格式或插件运行时桥接协议。
- 不新增运行时后端服务、`REST API`、数据库迁移或前端页面。
- 不为旧`go run ./hack/tools/image-builder`或`go run ./hack/tools/build-wasm`入口提供长期兼容层。

## Decisions

### 1. 使用`linactl/internal`独立包承载实现

镜像构建实现放入`hack/tools/linactl/internal/imagebuilder`，动态插件打包实现放入`hack/tools/linactl/internal/wasmbuilder`，运行时`i18n`治理扫描实现放入`hack/tools/linactl/internal/runtimei18n`。命令文件仍只负责编排参数、插件工作区准备和输出，复杂实现收敛到内部包。

备选方案是把两个工具源码直接放在`linactl`根包中。该方案会让命令入口文件和复杂构建逻辑混在一起，不符合`linactl`子组件组织规范。

### 2. 允许`linactl`编译期依赖`lina-core/pkg/pluginbridge`

`wasmbuilder`需要使用`pluginbridge`中的 artifact section、生命周期、路由和 host service contract 定义。将这部分依赖保留为编译期依赖可以消除子进程工具边界，减少路径和工作区准备复杂度。

备选方案是继续通过子进程调用独立`build-wasm`，但这会保留本次希望消除的独立模块。

### 3. 保持公开命令不变并删除旧独立入口

用户使用入口仍为`make image`、`make image.build`、`make wasm`和对应`linactl`命令。旧工具目录删除后，仓库内测试、`CI`和文档必须全部切换到公开入口或内部包测试。

备选方案是保留旧目录作为兼容 wrapper。由于这些工具是仓库内部开发工具，不是对外稳定`CLI`，保留 wrapper 会延长双入口维护期。

### 4. `runtime-i18n`作为`linactl`治理组件执行

`linactl i18n.check`直接调用`internal/runtimei18n`的扫描与消息覆盖检查函数，并保留两个检查都尝试执行的行为：即使硬编码文案扫描失败，也继续执行消息覆盖检查，最终合并返回错误。`allowlist.json`随内部组件迁移到`hack/tools/linactl/internal/runtimei18n/allowlist.json`，默认路径由`linactl`仓库根目录解析。

## Risks / Trade-offs

- `linactl`依赖变重 → 通过接受项目专用工具定位，并在`linactl/go.mod`中显式维护依赖来控制影响。
- 旧路径引用遗漏 → 通过`rg`扫描`image-builder`、`build-wasm`和旧`go run`入口，并更新测试、文档与`CI`夹具。
- `wasmbuilder`迁移后 guest runtime 构建工作区错误 → 保留既有`temp/go.work.plugins`选择逻辑，并运行`linactl wasm`或对应单元测试验证。
- `runtime-i18n`迁移后 allowlist 路径错误 → 通过内部包单元测试和`linactl i18n.check`smoke 验证默认路径。
- 工作区已有大量无关改动 → 本次只触碰构建工具整合相关文件，不回退或重写其他活跃变更。

## Migration Plan

1. 迁移`image-builder`源码到`linactl/internal/imagebuilder`并改为可调用包函数。
2. 迁移`build-wasm/internal/builder`源码到`linactl/internal/wasmbuilder`并让`linactl wasm`直接调用。
3. 迁移`runtime-i18n`源码到`linactl/internal/runtimei18n`并让`linactl i18n.check`直接调用。
4. 更新`go.work`、`linactl/go.mod`、测试辅助、`E2E`、`CI`夹具和文档引用。
5. 删除旧独立工具目录。
6. 运行`linactl`单元测试、命令 smoke、OpenSpec 校验和静态引用扫描。

## Open Questions

无。
