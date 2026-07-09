## 1. Storage 与 Files 边界澄清

`Storage()` 定位为插件私有对象存储领域能力，源码插件和动态插件使用同一 `storagecap.Service` 语义。`Files()` 定位为宿主文件中心资源投影领域能力，负责 `sys_file` 可见性、批量投影和存在性不泄露校验。

**决策**：不重命名 `Storage()`/`Files()`；不让 `Storage()` 生成 `sys_file` 记录；不让 `Files()` 新增对象存储方法。迁移 `linapro-demo-source` 示例，使源码插件附件通过 `Storage()` 管理。

## 2. 动态插件分片上传

动态插件 `Storage().Put` 当前在 guest 侧 `io.ReadAll` 后单次传给宿主，受 guest 内存约束。

**决策**：保留 `storage.put` 单次上传路径，新增 `put.init`/`put.chunk`/`put.commit`/`put.abort` 分片方法。guest SDK 根据输入大小和可知性自动选择传输策略。上传会话由 WASM storage host service 管理，提交时通过 `storagecap.Service.Put` 流式写入最终 logical path。分片授权继续使用 `storage.resources.paths` 匹配。
