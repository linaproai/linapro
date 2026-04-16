## Why

当前动态插件虽然已经能把 `frontend`、`manifest` 等资源随 `.wasm` 一起交付，但资源来源仍由构建器直接扫描源码目录决定，和源码插件通过 `go:embed` 声明资源的方式不一致。作者在源码插件与动态插件之间切换时，需要理解两套不同的资源打包心智模型，增加了开发和维护成本。

本次变更希望统一“插件作者如何声明资源”的方式：动态插件也通过 `go:embed` 声明自己的 `plugin.yaml`、`frontend`、`manifest` 等资源；但宿主仍继续消费构建阶段产出的快照自定义节，不把上传校验、启用校验、菜单绑定和静态资源托管改成运行时 guest 资源读取链路。

## What Changes

- 为动态插件新增作者侧资源声明约定：插件通过根级资源声明文件使用 `go:embed` 暴露 `plugin.yaml`、`frontend`、`manifest` 等资源集合。
- 调整动态插件构建器：优先从动态插件声明的嵌入文件系统中读取资源，再生成宿主已有的 `manifest`、前端资源、SQL、路由合同等 Wasm 自定义节快照。
- 保持宿主治理模型不变：上传、安装、启用、菜单校验、`/plugin-assets/...` 托管与动态路由治理仍以 Wasm 自定义节快照为真相源，而不是通过 guest 方法动态读取资源。
- 为动态插件样例补齐统一的资源声明方式，并更新开发文档，明确源码插件与动态插件在“作者侧声明”一致、在“宿主侧消费”不同。
- 为构建器和宿主补充兼容规则，确保未声明 `go:embed` 的旧动态插件在迁移期仍可继续通过目录扫描构建。

## Capabilities

### New Capabilities

- `plugin-embed-snapshot-packaging`: 规范动态插件如何通过 `go:embed` 声明资源，并由构建阶段生成宿主可治理的快照自定义节。

### Modified Capabilities

- `plugin-runtime-loading`: 动态插件运行时产物的资源来源从“仅目录扫描”扩展为“优先读取插件声明的嵌入文件系统，并回写为宿主可消费的自定义节快照”。
- `plugin-manifest-lifecycle`: 动态插件的 `plugin.yaml` 与安装/卸载 SQL 在作者侧可由 `go:embed` 声明，但宿主安装、上传与生命周期治理仍以产物内嵌快照为准。
- `plugin-ui-integration`: 动态插件前端托管资源改为允许来源于插件声明的嵌入文件系统，并继续通过宿主托管地址稳定对外服务。

## Impact

- 影响 `apps/lina-plugins/plugin-demo-dynamic` 的资源声明方式与示例代码结构。
- 影响 `hack/build-wasm` 构建器的资源收集逻辑、自定义节生成逻辑与测试。
- 影响宿主动态插件解析与文档说明，但不改变现有上传接口、固定前缀路由治理和 `/plugin-assets/...` 对外地址。
- 需要补充 OpenSpec 规范、开发文档与构建/兼容性测试。
