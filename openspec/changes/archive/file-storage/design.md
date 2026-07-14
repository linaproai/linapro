# Design

## 数据模型与场景建模

系统通过 `sys_file` 表统一记录所有上传文件的元信息，包括存储文件名、原始文件名、后缀、大小、散列值、访问 URL、相对存储路径、存储引擎、使用场景和上传者等字段。使用场景直接落在 `scene` 字段中，不再维护独立的关联表，以降低查询和写入复杂度。

系统预定义场景至少包括 `avatar`、`notice_image` 和 `notice_attachment`。同一物理文件内容即使在不同场景复用，也通过独立元数据记录表达业务上下文，确保列表筛选、详情展示和后续业务审计保持清晰。

新写入时 `sys_file.engine = providerID`（`local` 或云插件 id）。列表/下载不依赖 `engine` 做路由，以 `ResolveProvider` 结果与双读策略为准；`engine` 供运维与后续迁移使用。

## 存储抽象与统一对象后端

系统保留 **Files / Storage 领域门面**，统一 **物理后端**选择：

| 路径 | 职责 |
|------|------|
| 文件中心 `file.Service` + `sys_file` | 业务编排、元数据、列表/筛选/权限 |
| 插件 `storagecap.Service` | 插件对象存储门面 |
| `storagecap.ResolveProvider` | 共享后端选择：0 云→local、1 云→该云、≥2→冲突 |

文件中心内容层依赖宽接口宿主 `storage.Service`（不另建文件域窄接口）。启动注入 `storage.NewResolvingService(local, runtime, localProvider)`：对 `NamespaceFiles` 内部走 `ResolveProvider`；其它 namespace 委托本地实现。插件仍经 `storagecap.Service` 调用同一 `ResolveProvider`。

Provider 键约定：

| 来源 | Provider Key |
|------|--------------|
| 文件中心 path `P` | `files/P` |
| 插件 logical path | 既有插件/租户作用域 key（不变） |

本地 provider 路由：

- `files/...` → `NamespaceFiles` + 去掉前缀后的相对 key（**不加** `.capability-storage`）
- 其它 → 既有 `NamespacePlugins` + `.capability-storage/...`

读写语义：

- **Put**：只写活动后端。
- **Get**：活动后端未命中且活动后端非 local 时，再试 local `files/`（仅文件中心适配层，过渡兼容）。
- **Delete**：活动后端删除后，best-effort 再删 local 同 key（避免残留）。

## 本地路径策略

本地存储以 `upload.path` 作为根目录，并按租户与年月组织物理文件。普通文件写入新的物理文件时，相对存储路径采用 `<tenantId>/<yyyy>/<MM>/<generated-file-name>`，保留租户 ID 目录以支持物理隔离、排查和后续按租户治理，同时去掉原先仅作为缩写的 `t` 目录层。`upload.maxSize` 用于限制单次上传大小。

文件访问与下载始终以数据库中的 `sys_file.path` 为权威相对路径（在统一后端之上再拼 `files/` 前缀）。历史记录如果已经保存为 `t/<tenantId>/...`，系统继续按原记录读取存储后端，不对旧文件执行迁移或路径重写。因此旧路径和新路径可以长期并存访问。

## 去重与兼容性约束

上传流程继续基于当前租户和文件 SHA-256 散列值进行去重。命中重复内容时，系统复用已有物理文件，仅新增文件元数据记录，并保留原有物理路径；如果历史重复文件使用的是 `t/<tenantId>/...` 路径，也不会为了生成新格式路径而再次写入文件。这一策略同时保证了存储去重语义、历史路径兼容性和路径规则调整的低风险落地。

路径规则调整只影响真正写入的新物理文件，不改变公开上传、下载和访问 API，也不修改既有文件记录的 `path`。插件 host storage 中非文件中心对象的路径规则保持不变。

## 后端 API 与业务编排

文件管理能力通过统一 RESTful API 暴露：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/file/upload` | 上传文件（`multipart/form-data`），支持可选 `scene` 参数 |
| GET | `/file` | 文件列表（分页+筛选：文件名、后缀、使用场景、时间范围），支持按大小、上传时间排序 |
| GET | `/file/download/{id}` | 下载文件 |
| DELETE | `/file/{ids}` | 删除文件（支持批量） |
| GET | `/file/suffixes` | 获取数据库中已存在的文件后缀列表 |
| GET | `/file/scenes` | 获取系统预定义的使用场景列表 |

上传接口返回文件 ID、存储文件名、原始文件名、后缀、大小和完整访问 URL。列表接口同样返回可直接使用的完整 URL，避免前端自行拼接。场景列表接口返回系统预定义场景集合，而不是仅返回数据库中已存在的数据，以保证前端筛选和上传组件行为稳定。

内容字节的上传、下载打开与物理删除经 `FileObjectStore`（ResolveProvider + `files/` 键 + 本地回退读）完成；列表、检索与元数据浏览继续基于 `sys_file`（及数据权限），不以云桶全量 List 作为业务目录来源。

通知公告富文本编辑器通过 `uploadHandler` 调用统一上传接口并使用 `notice_image` 场景插入完整 URL；通知公告附件通过通用文件上传组件使用 `notice_attachment` 场景；用户头像上传统一切换到 `avatar` 场景，并复用同一套文件存储与管理能力。

## 前端模块与交互设计

前端在 `src/components/upload/` 下提供通用 `FileUpload`、`ImageUpload` 组件和配套上传逻辑，用于处理拖拽上传、图片卡片回显、进度、校验和错误反馈。组件通过 `v-model:value` 绑定文件 ID，并通过 `scene` 参数声明业务场景。

系统管理下新增“文件管理”页面，统一承载文件列表、搜索、上传、批量删除、下载、详情和预览能力。页面默认开启预览模式，图片和 PDF 直接预览，其他不可预览文件展示可复制和可点击的 URL 地址；文件类型筛选使用 Select，下拉选项来自 `/file/suffixes` 接口且不包含点号。详情弹窗展示文件完整信息、使用场景、文件路径等元数据，并通过统一宽度约束保持可读性。

## 风险与迁移

| 风险 | 缓解 |
|------|------|
| 多云冲突影响文件中心 | 配置页提示只启用一个；错误码可读（含 `FILE_STORAGE_CONFLICT`） |
| 切换云后旧本地文件 | Get 本地回退 |
| 云 provider 未配置 | 与插件 Storage 相同 fail 语义 |

迁移预期：未启云插件时行为与本地一致；启用唯一云插件后新文件上云、旧本地文件仍可读；不做批量迁移工具。

## 运行时治理评估

统一后端选择引入可诊断的存储冲突错误源码文案；文件访问继续以 `sys_file` 元数据与 `sys_file.path` 为权威，不引入额外分布式缓存一致性问题。数据权限边界保持不变：上传记录仍归属当前租户，读取和下载继续依赖既有元数据查询、租户过滤和权限校验流程。
