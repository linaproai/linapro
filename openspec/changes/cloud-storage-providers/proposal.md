## Why

宿主已提供 `storagecap.Service` 与 `storagecap.Provider` 扩展点，但对象存储后端目前仅有内置本地磁盘 provider。集群与生产环境需要腾讯云 COS、阿里云 OSS、AWS S3 等云对象存储对接；运维侧也需要与「系统监控」同构的稳定菜单挂载点，在安装对应 provider 插件后集中配置各云凭据与桶信息，而不是再造一层伪领域 core 插件。

## What Changes

- 在宿主稳定一级菜单骨架中新增 `storage`（存储管理）目录；无可见云存储配置子菜单时由导航投影隐藏，有插件挂载后展示。
- 新增官方源码插件：腾讯云 COS、阿里云 OSS、华为云 OBS、七牛云 Kodo、AWS S3、Azure Blob 厂商插件，以及 S3 兼容协议插件；均实现 `storagecap.Provider` 并通过 `storagecap.Provide` 注册；运行时仍由宿主 `ResolveProvider` 选择唯一可服务 provider（0 个回退 local，多个冲突拒绝）。
- 每个云插件提供管理后台配置页（GET/PUT settings），样式对齐「授权登录」子页；配置经 `sys_config` 持久化，密钥掩码投影；菜单 `parent_key: storage`。
- 明确不引入 `linapro-storage-core`：Storage 为 core-owned 能力，父目录归宿主，具体云实现归插件。
- **Out of scope**：宿主文件中心（`Files()` / `sys_file`）上云、预签名直传 URL、跨 provider 数据迁移、多 active provider 并存。

## Capabilities

### New Capabilities

- `cloud-storage-provider-plugins`：云对象存储 source 插件（COS/OSS/OBS/七牛/AWS/Azure 厂商 + S3 协议）的 Provider 实现、设置 API/页面、sys_config 密钥治理与唯一启用语义。
- `host-storage-menu-catalog`：宿主稳定目录 `storage`（存储管理）的种子、i18n、空目录隐藏与插件语义挂载契约。

### Modified Capabilities

- `plugin-storage-service`：将「源码插件可注册 OSS/S3 等 provider」从预留能力落实为可交付要求，并与宿主 `storage` 目录下的配置管理面衔接（不改变 `storagecap.Service` 插件可见契约语义）。
- `menu-management` / `core-host-boundary-governance`：稳定一层目录列表增加 `storage`；插件按 `parent_key: storage` 挂载云存储配置页。

## Impact

- **宿主 `lina-core`**：菜单 SQL 种子与 i18n（`storage` 目录、`developer` 等 sort 顺延如需要）；原则上不改 `storagecap` 公开方法契约；需允许框架文件变更（仓库存在 `.contributing`）。
- **插件 `lina-plugins`**：新增 `linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure`、`linapro-storage-s3`（source、platform_only、managed；厂商插件与 S3 协议插件配置面分离）。
- **运行时**：业务插件继续只调用 `Storage()`；启用恰好一个云插件后对象写入云桶，误多开则 `CodeStorageProviderConflict`。
- **依赖**：各插件引入对应云厂商官方 Go SDK（或成熟社区 SDK），宿主不引入云 SDK。
- **i18n**：宿主菜单文案 + 三插件菜单/设置页/错误文案双语。
- **测试**：宿主目录显隐 E2E/投影行为；至少一家 settings 掩码与保存；Provider 单测（mock）；可选连通性探测。
- **数据权限**：settings 为平台配置控制面；Storage 运行时仍按既有插件/租户 key 作用域隔离。
- **缓存**：无新增跨节点业务缓存权威数据；配置读路径遵循 sys_config 既有语义。
