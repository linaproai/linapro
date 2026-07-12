## Context

现有两套路径：

1. 文件中心 → `storagesvc` namespace `files`（仅本地）
2. 插件 `Storage()` → `storagecap.ResolveProvider`（local 或唯一云插件）

目标：统一 **物理后端**，保留 **Files / Storage 领域门面**。

## Decisions

### 1. 统一 Backend，不合并业务 API

- 文件中心仍 `file.Service` + `sys_file`，**依赖宽接口**宿主 `storage.Service`（不另建文件域窄接口）。
- 启动注入 `storage.NewResolvingService(local, runtime, localProvider)`：对 `NamespaceFiles` 内部走 `ResolveProvider`；其它 namespace 委托本地实现。
- 插件仍 `storagecap.Service`，同样 `ResolveProvider`。
- 两者共享后端选择规则，但业务门面分离。

### 2. Provider key 前缀

| 来源 | Provider Key |
|---|---|
| 文件中心 path `P` | `files/P` |
| 插件 logical path | 既有插件/租户作用域 key（不变） |

本地 provider：

- `files/...` → `NamespaceFiles` + 去掉前缀后的相对 key（**不加** `.capability-storage`）
- 其它 → 既有 `NamespacePlugins` + `.capability-storage/...`

### 3. 历史兼容

- Get：活动后端未命中且活动后端非 local 时，再试 local `files/`（仅文件中心适配层）。
- Delete：活动后端删除后，best-effort 再删 local 同 key（避免残留）。
- Put：只写活动后端。

### 4. Engine 字段

新写入 `sys_file.engine = providerID`（`local` 或云插件 id）。列表/下载不依赖 engine 做路由，以 Resolve 结果 + 双读为准；engine 供运维与后续迁移使用。

## Risks

| 风险 | 缓解 |
|---|---|
| 多云冲突影响文件中心 | 配置页已提示只启用一个；错误码可读 |
| 切换云后旧本地文件 | Get 本地回退 |
| 云 provider 未配置 | 与插件 Storage 相同 fail 语义 |

## Migration

1. 部署本变更。
2. 未启云插件：行为与本地一致。
3. 启用唯一云插件：新文件上云；旧本地文件仍可读。
4. 不做批量迁移工具（后续变更）。
