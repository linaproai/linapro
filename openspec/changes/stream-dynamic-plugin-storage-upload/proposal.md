## Why

当前动态插件`Storage().Put`会在 guest 侧把`io.Reader`完整读入内存，再通过单次`storage.put`host call 传给宿主；这与`storagecap.Service`已经支持`io.Reader`流式写入的领域契约不一致，也会让未知大小或大文件上传受 guest 内存约束。

本变更为动态插件补齐分片上传传输能力，使动态插件既保留小文件单次上传路径，又可以在大文件或未知大小输入下以有界内存写入同一个`Storage()`领域能力。

## What Changes

- 为动态插件`storage`host service 增加`put.init`、`put.chunk`、`put.commit`和`put.abort`方法。
- guest SDK 的`Storage().Put`根据输入大小和可知性自动选择单次`storage.put`或分片上传。
- 宿主侧以临时上传会话接收分片，提交时通过`storagecap.Service.Put`把临时对象流式写入最终 logical path。
- 分片上传的授权边界继续使用`storage.resources.paths`匹配最终插件 logical path，不能授权 provider object key、宿主物理路径或文件中心 ID。
- 单次 host call 只限制分片 payload 大小以保护 WASM 调用内存，不对最终对象大小引入固定上限；最终对象大小仍由 storage provider 或逻辑空间策略治理。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-storage-service`：补充动态插件`Storage().Put`的大文件和未知大小输入分片上传要求。

## Impact

- 影响`apps/lina-core/pkg/plugin/pluginbridge/protocol`中的 storage host service 编解码 DTO 与方法目录。
- 影响`apps/lina-core/internal/service/plugin/internal/wasm`中的 storage host service 注册、分发和测试。
- 影响`apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall`中的动态插件 guest storage adapter。
- 影响`apps/lina-plugins/linapro-demo-dynamic/plugin.yaml`中的 storage host service 方法声明，用于示例插件覆盖分片上传路径。
- 不修改 HTTP API、数据库结构、宿主文件中心`Files()`领域或源码插件`storagecap.Service`公开契约。
