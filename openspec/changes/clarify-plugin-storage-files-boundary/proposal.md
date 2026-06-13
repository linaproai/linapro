## Why

当前插件体系已经同时提供`Storage()`和`Files()`两类能力，但源码插件示例仍存在自行拼接本地上传路径的实现，动态插件示例也容易让开发者把`storage`理解为宿主文件中心能力。需要明确把`Storage()`定位为独立的插件对象存储领域能力，并与`Files()`宿主文件领域能力划清边界，避免源码插件、动态插件和后续插件开发出现语义混用。

## What Changes

- 明确`Storage()`是源码插件和动态插件共享的插件私有对象存储领域能力，用于插件自有附件、导入导出临时对象、业务二进制对象和卸载清理对象。
- 明确`Files()`是宿主文件中心领域能力，只用于读取和校验`sys_file`文件资源投影，不承担插件私有对象的写入、读取、删除和列表职责。
- 规范动态插件`hostServices`声明：插件对象存储使用`service: storage`和`resources.paths`；宿主文件资源投影使用`service: files`和文件领域方法。
- 规划迁移`linapro-demo-source`源码插件示例，使其通过`pluginhost.Services.Storage()`注入`storagecap.Service`管理插件私有附件，不再直接读取`upload.path`、拼接宿主本地路径或依赖`os/gfile`管理插件附件。
- 更新`apps/lina-core/pkg/plugin`与源码/动态插件说明文档，确保中英文文档对`Storage()`和`Files()`边界表达一致。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-storage-service`：补充`Storage()`作为独立插件对象存储领域能力的边界，明确其不产生`sys_file`元数据、不进入宿主文件管理列表、不暴露宿主物理路径。
- `plugin-host-domain-capabilities`：补充`Files()`宿主文件领域能力边界，明确其只处理宿主文件中心资源投影、批量读取和可见性校验，不承担插件私有对象存储职责。

## Impact

- 影响规范：
  - `openspec/specs/plugin-storage-service/spec.md`
  - `openspec/specs/plugin-host-domain-capabilities/spec.md`
- 预计影响代码与文档：
  - `apps/lina-plugins/linapro-demo-source/backend/**`
  - `apps/lina-plugins/linapro-demo-source/README.md`
  - `apps/lina-plugins/linapro-demo-source/README.zh-CN.md`
  - `apps/lina-plugins/linapro-demo-dynamic/README.md`
  - `apps/lina-plugins/linapro-demo-dynamic/README.zh-CN.md`
  - `apps/lina-core/pkg/plugin/README.md`
  - `apps/lina-core/pkg/plugin/README.zh-CN.md`
- 不影响范围：
  - 不修改`storagecap.Service`、`filecap.Service`公开方法签名。
  - 不修改动态插件`hostServices` wire 字符串、WASM ABI、现有动态插件 artifact 格式。
  - 不新增 SQL、DAO、DO、Entity 或宿主文件中心数据模型。
  - 不改变宿主文件管理 HTTP API 或动态插件公开路由契约。
