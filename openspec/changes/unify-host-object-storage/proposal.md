## Why

当前宿主文件中心和插件对象存储分别维护本地磁盘读写实现，底层对象存储职责重复，但二者上层领域语义不同。需要将对象内容读写收敛到宿主中立存储组件，同时保持文件中心和插件 `Storage()` 能力各自的权限、元数据和协议边界。

## What Changes

- 新增宿主内部对象存储能力，由中立 `storage.Service` 负责对象 `Put`、`Get`、`Delete`、`Stat`、有界 `List` 和本地 provider 实现。
- 将文件中心物理文件读写从 `file.Storage` 收敛到新的宿主对象存储服务；文件中心继续拥有 `sys_file` 元数据、hash 复用、业务场景、数据权限和公开下载 URL 语义。
- 将插件 `storagecap.Service` 保持为插件公开领域契约，但其底层 provider-backed 本地读写复用新的宿主对象存储实现。
- 保持插件 `Storage()` 与 `Files()` 领域边界独立：插件私有对象不自动进入宿主文件中心，文件中心也不依赖插件公开能力接口。
- 不改变 HTTP API、数据库结构、插件公开 DTO、动态插件 host service 方法名或用户可观察页面行为。

## Capabilities

### New Capabilities

- `host-object-storage`: 宿主内部中立对象存储能力，覆盖 provider owner、命名空间/key 语义、路径安全、元数据、列表性能和复用边界。

### Modified Capabilities

- `file-upload-storage-path`: 文件中心上传路径继续保持现有语义，但物理对象写入由宿主对象存储服务承载。
- `plugin-storage-service`: 插件 `storagecap.Service` 继续拥有插件 logical path、插件/租户隔离和授权语义，但 provider 本地读写复用宿主对象存储服务。

## Impact

- 影响范围：
  - `apps/lina-core/internal/service/file`
  - `apps/lina-core/internal/service/plugin/internal/capabilityhost`
  - `apps/lina-core/internal/cmd/internal/httpstartup`
  - `apps/lina-core/pkg/plugin/capability/storagecap`
- 不影响：
  - HTTP API、OpenAPI 元数据和前端调用契约。
  - 数据库迁移、`sys_file` 表结构和 DAO 生成工件。
  - 插件公开 `storagecap.Service`、`filecap.Service` DTO 与动态 host service 方法名。
- 验证方式：
  - `openspec validate unify-host-object-storage --strict`
  - 覆盖文件服务、插件 storage provider/adapter、WASM storage host service 和启动装配的 Go 测试。
  - 静态检索确认 `file.Storage` 被移除，文件中心不依赖插件 `storagecap.Service`。
