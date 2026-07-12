## Why

文件中心（`sys_file`）内容读写仍固定走宿主本地 `storage.Service`（namespace `files`），与插件 `Storage()` / `storagecap` 的「0 云→local、1 云→云、≥2→冲突」后端选择分离。产品期望：**所有字节级文件存储（上传/下载/删除）统一对象后端**；无云插件时默认本地。列表/检索仍走 `sys_file` 元数据。

## What Changes

- 引入宿主级统一对象后端选择（复用 `storagecap.ResolveProvider`）。
- 文件中心 Put/Get/Delete **内容**经统一后端；键前缀 `files/` 与插件私有对象隔离。
- 本地 provider 支持将 `files/` 键路由到 `NamespaceFiles`（兼容既有磁盘布局）。
- 读取时对历史本地文件做 **云未命中→本地回退**（过渡期）。
- 新写入 `sys_file.engine` 记录生效 provider id（`local` 或插件 id）。
- 列表/检索/浏览元数据路径不变。

## Capabilities

### New Capabilities

- `unified-object-backend`: 全站对象字节后端选择与文件中心内容接入。

### Modified Capabilities

- （可选交叉）`plugin-storage-service`：明确文件中心内容与插件 Storage 共享 Resolve 规则。

## Impact

- `internal/service/file`：内容层依赖 ResolveProvider 适配。
- `capabilityhost` local storage provider：`files/` 命名空间路由。
- `httpstartup` DI 顺序：file 服务注入 storage runtime + local provider。
- 多云同时启用时，文件中心上传/下载/删除也会冲突失败（与插件 Storage 一致）。
