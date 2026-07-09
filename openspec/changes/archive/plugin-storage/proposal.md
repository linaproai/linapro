## Why

`Storage()` 和 `Files()` 已经同时存在于 `capability.Services` 和 `pluginbridge.Services` 中，但边界表达不够清晰：`Storage()` 容易被误解为宿主文件中心，`Files()` 命名接近"文件读写"但实际只提供宿主文件中心投影读取。同时动态插件 `Storage().Put` 在 guest 侧把 `io.Reader` 完整读入内存再单次传给宿主，与 `storagecap.Service` 已支持的流式写入不一致。

## What Changes

- 明确 `Storage()` 是插件私有对象存储领域能力，`Files()` 是宿主文件中心资源投影领域能力。
- 迁移 `linapro-demo-source` 示例，使源码插件附件通过 `pluginhost.Services.Storage()` 管理。
- 为动态插件 `storage` host service 增加 `put.init`/`put.chunk`/`put.commit`/`put.abort` 分片方法。
- guest SDK 的 `Storage().Put` 根据输入大小自动选择单次或分片上传。

## Impact

- 影响 `apps/lina-core/pkg/plugin/pluginbridge/protocol`、`apps/lina-core/internal/service/plugin/internal/wasm`、`apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall`、`apps/lina-plugins/linapro-demo-source` 和 `apps/lina-plugins/linapro-demo-dynamic`。
- 不修改 HTTP API、数据库结构、宿主文件中心 `Files()` 领域或源码插件 `storagecap.Service` 公开契约。
- 不新增 SQL、DAO、DO、Entity 或宿主文件中心数据模型。
