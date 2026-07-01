## Context

当前有两套本地磁盘对象读写实现：文件中心的 `file.Storage` 和插件存储的本地 `storagecap.Provider`。它们都负责对象内容的写入、读取、删除和路径安全，但上层职责不同：文件中心拥有 `sys_file` 元数据、hash 复用、数据权限和下载 URL；插件 `storagecap.Service` 拥有插件 logical path、插件/租户作用域、动态授权和 provider 状态。

本设计将重复的对象内容读写收敛为宿主内部中立 `storage` 组件。`file` 和 `storagecap` 都作为领域层依赖该组件，而不是彼此依赖。这样保持高内聚：对象存储归 `storage`，文件领域归 `file`，插件对象领域归 `storagecap`。

## Goals / Non-Goals

**Goals:**

- 新增 `apps/lina-core/internal/service/storage` 作为宿主对象存储 owner，提供中立 `Service` 和本地 provider 实现。
- 移除 `file` 包内的导出 `Storage` 接口和 `LocalStorage` 实现，让文件中心通过 `storage.Service` 写入、读取和删除物理对象。
- 让插件本地 provider 复用同一个 `storage.Service` 或同一中立本地 provider 实现，避免维护第二套本地文件系统逻辑。
- 保持 `file.Service`、`filecap.Service`、`storagecap.Service`、动态插件 host service、HTTP API 和数据库结构不变。
- 使用显式依赖注入从启动装配传递同一个宿主对象存储实例，避免运行期临时构造。

**Non-Goals:**

- 不引入 S3、MinIO、OSS 等新外部 provider。
- 不改变 `storagecap.Service` 的公开 DTO、方法名、授权语义或插件 logical path 规则。
- 不改变文件中心 URL 路径、`sys_file.path` 存储格式、hash 复用语义或数据权限策略。
- 不新增通用 DI 容器、全局 service locator 或聚合依赖结构体。

## Decisions

### 1. 新增中立 `storage.Service`，不复用 `storagecap.Service` 作为底座

`storagecap.Service` 是插件公开领域契约，包含插件 ID、租户作用域、logical path、动态授权前置和 provider 状态语义。文件中心直接依赖它会让 `file` 反向耦合插件协议。新增内部 `storage.Service` 只表达对象 key 读写、删除、元数据和有界列表，调用方负责把领域标识映射为命名空间和 key。

备选方案是让 `file` 直接依赖 `storagecap.Service`。该方案被拒绝，因为它会把文件中心路径、公开 URL 和 hash 复用建立在插件 logical path 协议上，扩大耦合并模糊 `Storage()` 与 `Files()` 的领域边界。

### 2. `storage.Service` 使用 namespace/key 表达隔离边界

中立存储服务接收稳定 `Namespace` 与相对 `Key`，内部负责路径清理、目录穿越拒绝、绝对路径拒绝和本地根目录锚定。文件中心使用如 `files` 命名空间并传入现有 `<tenantId>/<yyyy>/<MM>/<generated>` key；插件 provider 使用如 `plugins` 命名空间并传入现有 provider object key。

这样不会改变已有持久化路径语义：`sys_file.path` 继续保存文件中心相对路径，插件返回值继续只暴露 logical path。

### 3. URL 仍归文件中心，不进入中立存储

文件下载 URL 是 HTTP 文件中心领域语义，不是底层对象存储语义。`storage.Service` 不提供 `Url` 方法；文件中心继续在 `file` 包内生成 `/api/v1/uploads/<path>`，避免插件对象存储或底层 provider 获得宿主文件中心公开 URL 责任。

### 4. 插件 `storagecap.Provider` 保持公开扩展契约

`storagecap.Provider` 和 provider factory 仍保留在 `pkg/plugin/capability/storagecap`，因为它们是插件 provider 扩展点和公开能力契约。内置本地 provider 改为委托宿主 `storage.Service`，但不会把内部 `storage.Service` 暴露给源码插件或动态插件。

### 5. 测试以行为不变和重复实现消除为核心

文件中心测试继续覆盖上传路径、读取、删除和 hash 复用。插件存储测试继续覆盖 logical path 隔离、租户隔离、列表上限、分片上传和 host service 分发。新增或调整测试需要证明两个领域共享底层对象存储实现后，仍不泄露 provider key、本地绝对路径或文件中心 ID。

## Risks / Trade-offs

- 路径语义混淆 → `storage.Service` 只接收 namespace/key，`file` 和 `storagecap` 分别保留领域路径转换，测试覆盖目录穿越和历史路径读取。
- 插件 provider 扩展点被内部实现污染 → `storagecap.Provider` 公开契约保持不变，内置本地 provider 只是一个适配器。
- 文件中心 URL 误下沉到底层存储 → 明确 `storage.Service` 不提供 URL，URL 仍由 `file` 生成。
- 本地列表性能回归 → 中立存储列表必须有 limit 上限，插件 `List` 继续在领域层限制最大数量。
- DI 来源不清 → 启动装配创建一个宿主对象存储实例，并显式注入 `file.New` 与插件 host services 构造路径。

## Migration Plan

1. 新增 `internal/service/storage` 组件和本地实现，迁移文件系统读写、路径解析、`Stat`、有界 `List` 测试。
2. 修改 `file` 服务构造函数依赖 `storage.Service`，删除 `file.Storage` 和 `file.LocalStorage`。
3. 修改插件内置本地 provider，通过 `storage.Service` 执行 provider object key 的读写、删除、元数据和列表。
4. 修改启动装配，统一创建宿主对象存储实例，并分别注入文件服务和插件 host services。
5. 运行 OpenSpec 校验、静态检索和 Go 测试，确认公开契约与行为不变。

## Open Questions

无。当前变更不选择新的外部存储 provider，也不改变配置模型。
