## Requirements

### Requirement: 插件私有文件对象必须归属 Storage 领域

系统 SHALL 将插件自有附件、插件业务二进制对象、导入导出临时对象和插件卸载清理对象归属 `Storage()` 领域能力。源码插件和动态插件 MUST 通过 `storagecap.Service` 管理这些对象，不得通过宿主文件中心 `Files()` 领域或宿主本地物理路径管理。

### Requirement: Storage 和 Files 领域边界必须保持独立

系统 SHALL 保持 `Storage()` 和 `Files()` 两个领域能力的职责独立。`Storage()` 拥有插件对象内容生命周期，`Files()` 拥有宿主文件中心资源投影和可见性校验。任一领域的公开契约 MUST NOT 混入另一个领域的内部标识、存储模型或生命周期职责。

### Requirement: 动态插件 Storage Put 必须支持有界内存分片上传

系统 SHALL 允许动态插件通过 `Storage().Put` 写入大文件或未知大小输入，并在 guest SDK 内部按输入大小选择单次 `storage.put` 或分片上传。分片上传 MUST 使用 `put.init`、`put.chunk`、`put.commit` 和 `put.abort` host service 方法完成传输。系统 MUST NOT 对最终对象大小设置动态 host service 固定上限。

### Requirement: 动态插件 Storage 分片上传必须保持路径授权和会话绑定

系统 SHALL 对 `put.init`、`put.chunk`、`put.commit` 和 `put.abort` 执行与 `storage.put` 等价的 service、method 和 `storage.resources.paths` 授权校验。授权 path MUST 匹配最终插件 logical path。宿主 MUST 将 upload ID 绑定到当前插件 ID、最终 logical path 和上传会话状态。
