## Why

插件私有对象存储与宿主文件中心长期容易被混用：`Storage()` 被误解为文件管理，`Files()` 被当作对象读写入口；动态插件 `Storage().Put` 在 guest 侧整读进内存，无法可靠写入大对象。与此同时，宿主已具备 `storagecap.Service` / `storagecap.Provider` 扩展点与内置 local provider，但生产与集群场景需要 COS、OSS、OBS、七牛、AWS、Azure 及 S3 兼容等云后端；运维需要在管理后台集中配置凭据与桶信息，配置入口应落在既有「系统设置」下，而不是再造伪领域 core 插件或独立「存储管理」一级目录。

## What Changes

- 明确 `Storage()` 为插件私有对象存储领域能力，`Files()` 为宿主文件中心资源投影领域能力；二者契约与生命周期互不混入。
- 迁移 `linapro-demo-source` 示例，使源码插件附件通过 `pluginhost.Services.Storage()` / `storagecap.Service` 管理，禁止直读 `upload.path` 拼接宿主本地路径。
- 为动态插件 `storage` host service 增加 `put.init` / `put.chunk` / `put.commit` / `put.abort` 分片方法；guest SDK 的 `Storage().Put` 按输入大小自动选择单次或分片上传。
- 落实 `storagecap.Provider` 扩展：源码插件通过 `storagecap.Provide(pluginID, factory)` 注册云后端；运行时以「唯一可服务已注册 provider 插件」自动选中（0 个回退 local，多个冲突拒绝），不依赖主配置 active provider ID。
- 交付官方 source 云存储插件：腾讯云 COS、阿里云 OSS、华为云 OBS、七牛云 Kodo、AWS S3、Azure Blob 厂商插件，以及 S3 兼容协议插件；各自实现完整 `Provider`、settings API/页面、密钥掩码与连通性探测。
- 云存储配置菜单统一挂载到宿主 `setting`（系统设置）；不新增 `storage` 一级稳定目录，不引入 `linapro-storage-core` 壳插件。
- **Out of scope**：宿主文件中心（`Files()` / `sys_file`）对象内容上云、预签名直传 URL、跨 provider 数据迁移、多 active provider 并存、按租户选择不同云后端、动态插件实现 Provider。

## Capabilities

### New Capabilities

- `cloud-storage-provider-plugins`：云对象存储 source 插件（COS/OSS/OBS/七牛/AWS/Azure 厂商 + S3 协议）的 Provider 实现、设置 API/页面、sys_config 密钥治理、唯一启用与 fail-closed 语义。
- `host-storage-menu-catalog`：云存储配置菜单挂载到宿主 `setting`（系统设置）的契约；不维护独立 `storage` 一级目录。

### Modified Capabilities

- `plugin-storage-service`：澄清 Storage 领域归属与 Files 边界；动态插件分片上传；将「源码插件可注册 OSS/S3 等 provider」落实为可交付要求；以唯一可服务自动选中替代主配置 active provider；官方云插件配置面挂到系统设置。
- `plugin-host-domain-capabilities`：Files 仅表示宿主文件中心投影；插件私有附件必须使用 Storage。
- `menu-management`：一级稳定目录不含 `storage`；云存储插件按 `parent_key: setting` 挂载配置页。
- `core-host-boundary-governance`：稳定目录清单不含 `storage`；云对象存储实现由插件扩展，配置入口复用系统设置。

## Impact

- **宿主 `lina-core`**：`storagecap` 公开方法契约保持稳定；强化 Provider 选择与冲突语义；动态 storage host service 与 guest SDK 支持分片上传；菜单种子不维护 `storage` 一级目录。
- **插件 `lina-plugins`**：新增 `linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure`、`linapro-storage-s3`（source、platform_only、managed）；`linapro-demo-source` / demo 动态插件示例与文档更新。
- **运行时**：业务插件继续只调用 `Storage()`；启用恰好一个云插件后对象写入云桶；误多开则 `CodeStorageProviderConflict`；唯一云插件配置无效时 fail-closed，不静默回退 local。
- **依赖**：云 SDK 仅引入各云插件 `go.mod`，宿主不引入云 SDK。
- **i18n**：云插件菜单/设置页/错误文案双语；宿主不再维护 `storage` 菜单标题。
- **测试**：Storage/Files 边界与分片上传单测；云 Provider mock 单测；系统设置下云存储菜单挂载与 settings 掩码 E2E。
- **数据权限**：settings 为平台配置控制面；Storage 运行时仍按插件/租户 key 作用域与动态 path 授权隔离。
- **缓存**：无新增跨节点业务缓存权威数据；配置读路径遵循 `sys_config` 既有语义。
