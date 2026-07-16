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

路径规则调整只影响真正写入的新物理文件，不改变公开上传、下载和访问 API，也不修改既有文件记录的 `path`。插件 host storage 中非文件中心对象的路径规则保持不变。直传 init 生成的对象 key 与 multipart 上传使用同一套租户分区路径规则，complete 写入的 `sys_file.path` 与中转上传一致。

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
| POST | `/api/v1/files/direct-upload/init` | 直传初始化（鉴权、策略、会话、签发或 proxy） |
| POST | `/api/v1/files/direct-upload/complete` | 直传完成（Stat 校验、写 `sys_file`） |
| POST | `/api/v1/files/direct-upload/abort` | 直传中止（best-effort 清理会话） |

上传接口返回文件 ID、存储文件名、原始文件名、后缀、大小和完整访问 URL。列表接口同样返回可直接使用的完整 URL，避免前端自行拼接。场景列表接口返回系统预定义场景集合，而不是仅返回数据库中已存在的数据，以保证前端筛选和上传组件行为稳定。

内容字节的上传、下载打开与物理删除经统一对象后端完成；列表、检索与元数据浏览继续基于 `sys_file`（及数据权限），不以云桶全量 List 作为业务目录来源。在 active 云后端支持直连时，上传/下载字节路径可绕过宿主，但后端选择与冲突规则不变，元数据仍以 `sys_file` 为准。

通知公告富文本编辑器通过 `uploadHandler` 调用统一上传接口并使用 `notice_image` 场景插入完整 URL；通知公告附件通过通用文件上传组件使用 `notice_attachment` 场景；用户头像上传统一切换到 `avatar` 场景，并复用同一套文件存储与管理能力。

## 客户端直连对象访问

### 契约形态

在 `storagecap` 增加可选能力探测与中立 DirectAccess 描述，而非新建平行总线：

```text
SupportsDirectAccess(op) -> bool
CreateDirectAccess(DirectAccessInput) -> DirectAccess
// Mode: presigned_url | form_post | temporary_credentials | proxy
// Operation: put | get
// Method, URL, Headers, FormFields, ExpiresAt
```

与现有 Provider 扩展点一致；local 可不实现；宿主 adapter 统一探测后降级。拒绝每厂商独立 HTTP API 与默认 STS 路径。

### 领域生命周期

文件中心与插件 Storage 均采用：

```text
Init(业务约束) → 返回 uploadSessionId + DirectAccess 或 mode=proxy
客户端执行传输（若 direct）
Complete(sessionId, etag?, clientHash?) → 校验对象 → 写领域状态
Abort(sessionId) → 尽力清理会话/未完成对象（best-effort）
```

只发 URL 不 complete 会导致无 `sys_file`、无权限记录、hash 去重失效与孤儿对象。会话将 key、size 上限、content-type、tenant、actor、scene 绑定在服务端；第一期会话为进程内 store（由 `file.New` 创建 `directSessions`），集群共享 cache 可作为后续增强。会话不落 `sys_file` 直至 complete 成功。

### 传输 mode 优先级

| 优先级 | Mode | 使用场景 |
|---|---|---|
| 1 | `presigned_url`（PUT/GET） | S3/AWS/COS/OSS/OBS/Azure SAS 等 |
| 2 | `form_post` | 需 POST policy 的厂商/浏览器限制场景 |
| 3 | `temporary_credentials` | 大文件 SDK / 高级客户端；工作台默认不启用 |
| fallback | `proxy` | local、不支持、冲突、配置失败前的明确指示 |

前端通用执行器按 mode 分支，不按 providerId 分支。

### 安全约束（硬性）

- scoped key 由宿主生成（调用方不得指定任意云 key）
- content-length 上限 ≤ `sys.upload.maxSize`（及插件策略）
- content-type 白名单（若 scene 有限制）
- 过期默认 15m（代码常量 `defaultDirectUploadTTL`，上界 `maxDirectUploadTTL=1h`）
- 操作单一（put 凭证不能 get 列表）
- Complete 必须 Head/Stat 成功且 size 匹配
- 私有桶默认；不签发永久公开写 URL
- 密钥与过宽凭证永不进入浏览器

### 前端策略

`useUpload` / 默认 `uploadApi` 双模式：

1. 调 init
2. 若 `mode=proxy` 或 init 指示不支持 → 现有 multipart
3. 若 direct → 按 mode 上传 → complete
4. complete 失败 → 展示可读错误；可选 abort

init 返回 proxy 时自动中转；**直传执行过程中** CORS/网络失败**不**自动改走 multipart（避免双写与用户困惑）。第一期不新增 `sys.upload.preferDirect` 等 sys_config 键，行为固定为「探测支持则直连、否则 proxy」；运维通过启停云插件与桶 CORS 控制直连可用性。

### 与 hash 去重

- 客户端可在 init 前算 SHA-256；提供则先查复用（秒传）
- 有 client hash 才走秒传；无 hash 的直传仍创建新记录
- 默认不做 complete 后宿主流式下载重算 hash

### 插件 Storage 与 filecap

`storagecap.Service` 提供 CreateDirectPut / ConfirmDirectPut / CreateDirectGet；动态插件 path 授权覆盖直连会话。第一期 guest 客户端对 CreateDirect* 返回 `proxy`（保持大对象中转分片）；源码插件可直接使用 host 侧直连方法。第一期不扩展 `filecap`；浏览器走宿主 HTTP `direct-upload/*`；插件附件继续 `Storage()` 或 `Files().Upload` 中转。

### Provider 实现分工

| Provider | Put 直传 | Get 直链 | 备注 |
|---|---|---|---|
| local | 否 | 否 | 永远 proxy |
| s3/aws | presigned PUT/GET | 是 | multipart 二期可选 |
| cos/oss/obs | presigned 或 form | 是 | 按 SDK 最稳路径 |
| azure | SAS | 是 | |
| qiniu | upload token / form | 下载域名签名 | 与现有 Get 语义对齐 |

宿主**不** import 云 SDK；签名逻辑只在插件内。配置页补充浏览器直传 CORS 运维说明。

## 前端模块与交互设计

前端在 `src/components/upload/` 下提供通用 `FileUpload`、`ImageUpload` 组件和配套上传逻辑，用于处理拖拽上传、图片卡片回显、进度、校验和错误反馈。组件通过 `v-model:value` 绑定文件 ID，并通过 `scene` 参数声明业务场景。默认 `uploadApi` 已内置 dual-mode，组件无需改 props。

系统管理下新增“文件管理”页面，统一承载文件列表、搜索、上传、批量删除、下载、详情和预览能力。页面默认开启预览模式，图片和 PDF 直接预览，其他不可预览文件展示可复制和可点击的 URL 地址；文件类型筛选使用 Select，下拉选项来自 `/file/suffixes` 接口且不包含点号。详情弹窗展示文件完整信息、使用场景、文件路径等元数据，并通过统一宽度约束保持可读性。

## 风险与迁移

| 风险 | 缓解 |
|------|------|
| 多云冲突影响文件中心 | 配置页提示只启用一个；错误码可读（含 `FILE_STORAGE_CONFLICT`） |
| 切换云后旧本地文件 | Get 本地回退 |
| 云 provider 未配置 | 与插件 Storage 相同 fail 语义 |
| 桶 CORS 未配导致直传失败 | 配置页说明 + 明确错误 |
| 孤儿对象（init 后未 complete） | 短 TTL；abort；生命周期 GC 可选二期 |
| complete 与实际上传竞态 | complete 重试 Stat；会话未过期可重复 complete 幂等 |
| 伪造 complete | 会话服务端持有 key；Stat 校验；鉴权同上传权限 |
| 七牛/Azure 与 S3 语义差异 | 中立 DirectAccess 字段；插件内适配 |
| 集群多实例会话丢失 | 会话 API 边界已隔离；后续可写共享 cache |
| 大文件单次 presign 限制 | 文档说明上限；二期 multipart |

迁移预期：未启云插件时行为与本地一致；启用唯一云插件后新文件上云、旧本地文件仍可读；支持直连的云后端下前端优先直传；不做批量迁移工具。回滚：禁用云插件或关闭直传代码路径即可回到中转。

## 运行时治理评估

统一后端选择引入可诊断的存储冲突错误源码文案；文件访问继续以 `sys_file` 元数据与 `sys_file.path` 为权威，不引入额外分布式缓存一致性问题。数据权限边界保持不变：上传记录仍归属当前租户，读取和下载继续依赖既有元数据查询、租户过滤和权限校验流程；直传 init/complete 继承上传权限与租户，complete 校验会话租户。直连会话第一期为进程内存储，无业务缓存一致性变更。
