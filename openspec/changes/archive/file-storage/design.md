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
- **Multipart**：`NamespaceFiles` 透传 active provider 的 Create/UploadPart/Complete/Abort 与 part 级签发；多云冲突与 Put 相同 fail-closed；非文件命名空间首期可不支持 Multipart。

文件中心 MUST 通过宿主 `storage.Service` 完成对象字节的 Multipart 写入，不得在 file 服务内直接 import 云 SDK。

## 本地路径策略

本地存储以 `upload.path` 作为根目录，并按租户与年月组织物理文件。普通文件写入新的物理文件时，相对存储路径采用 `<tenantId>/<yyyy>/<MM>/<generated-file-name>`，保留租户 ID 目录以支持物理隔离、排查和后续按租户治理，同时去掉原先仅作为缩写的 `t` 目录层。

文件访问与下载始终以数据库中的 `sys_file.path` 为权威相对路径（在统一后端之上再拼 `files/` 前缀）。历史记录如果已经保存为 `t/<tenantId>/...`，系统继续按原记录读取存储后端，不对旧文件执行迁移或路径重写。因此旧路径和新路径可以长期并存访问。

## 去重与兼容性约束

上传流程继续基于当前租户和文件 SHA-256 散列值进行去重。命中重复内容时，系统复用已有物理文件，仅新增文件元数据记录，并保留原有物理路径；如果历史重复文件使用的是 `t/<tenantId>/...` 路径，也不会为了生成新格式路径而再次写入文件。

路径规则调整只影响真正写入的新物理文件，不改变公开上传、下载和访问 API，也不修改既有文件记录的 `path`。插件 host storage 中非文件中心对象的路径规则保持不变。直传 init（含 single 与 multipart）与中转分片生成的对象 key 使用同一套租户分区路径规则，complete 写入的 `sys_file.path` 与整包上传一致。

## 后端 API 与业务编排

文件管理能力通过统一 RESTful API 暴露：

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/file/upload` | 上传文件（`multipart/form-data` 整包），支持可选 `scene` 参数 |
| GET | `/file` | 文件列表（分页+筛选：文件名、后缀、使用场景、时间范围），支持按大小、上传时间排序 |
| GET | `/file/download/{id}` | 下载文件 |
| DELETE | `/file/{ids}` | 删除文件（支持批量） |
| GET | `/file/suffixes` | 获取数据库中已存在的文件后缀列表 |
| GET | `/file/scenes` | 获取系统预定义的使用场景列表 |
| POST | `/file/direct-upload/init` | 上传规划与直传初始化（鉴权、channel×encoding 策略、会话、签发或 proxy） |
| POST | `/file/direct-upload/part-url` | 直传 multipart：按 partNumber 签发短时写访问 |
| POST | `/file/direct-upload/complete` | 直传完成（single：Stat 校验；multipart：CompleteMultipart + parts）并写 `sys_file` |
| POST | `/file/direct-upload/abort` | 直传中止（best-effort 清理会话与云 Multipart） |
| POST | `/file/upload/chunked/init` | 中转分片初始化（亦可由 init 策略隐含） |
| POST | `/file/upload/chunked/part` | 中转分片上传（sessionId + partNumber + file） |
| POST | `/file/upload/chunked/complete` | 中转分片完成（云 Multipart 优先，否则拼装 + Put） |
| POST | `/file/upload/chunked/abort` | 中转分片中止 |

上传接口返回文件 ID、存储文件名、原始文件名、后缀、大小和完整访问 URL。列表接口同样返回可直接使用的完整 URL。场景列表接口返回系统预定义场景集合。chunked 与 direct multipart 权限标签与 `/file/upload` 相同：`system:file:upload`；租户/用户绑定会话。

内容字节的上传、下载打开与物理删除经统一对象后端完成；列表、检索与元数据浏览继续基于 `sys_file`（及数据权限）。在 active 云后端支持直连时，上传/下载字节路径可绕过宿主；≥ 分片阈值且支持云 Multipart 时优先 direct_multipart。

通知公告富文本编辑器通过 `uploadHandler` 调用统一上传接口并使用 `notice_image` 场景；通知公告附件使用 `notice_attachment`；用户头像使用 `avatar`。

## 客户端直连对象访问

### 契约形态

在 `storagecap` 增加可选能力探测与中立 DirectAccess 描述：

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
Init(业务约束) → 返回 uploadSessionId + strategy + DirectAccess 或 proxy/multipart 指引
客户端执行传输（single 或 multipart）
Complete(sessionId, parts?/etag?, clientHash?) → 校验对象 → 写领域状态
Abort(sessionId) → 尽力清理会话/未完成对象与云 Multipart（best-effort）
```

只发 URL 不 complete 会导致无 `sys_file`、无权限记录、hash 去重失效与孤儿对象。会话将 key、size 上限、content-type、tenant、actor、scene、encoding、云 uploadId 等绑定在服务端；第一期会话为进程内 store（由 `file.New` 创建），集群共享 cache 可作为后续增强。会话不落 `sys_file` 直至 complete 成功。

### 传输 mode 与四策略

| 优先级 | Mode | 使用场景 |
|---|---|---|
| 1 | `presigned_url`（PUT/GET） | S3/AWS/COS/OSS/OBS/Azure SAS 等 |
| 2 | `form_post` | 需 POST policy 的厂商/浏览器限制场景 |
| 3 | `temporary_credentials` | 大文件 SDK / 高级客户端；工作台默认不启用 |
| fallback | `proxy` | local、不支持、冲突、配置失败前的明确指示 |

init 响应扩展中立策略字段：

```text
strategy.channel:  direct | proxy
strategy.encoding: single | multipart
multipart?: { partSize, minPartSize, maxParts, maxConcurrency }
```

| channel | encoding | 客户端路径 |
|---------|----------|------------|
| direct | single | 既有 access 单次传输 → complete |
| direct | multipart | part-url 循环 + 有限并发 PUT → complete(parts) |
| proxy | single | `/file/upload` 整包 |
| proxy | multipart | `/file/upload/chunked/*` |

前端通用执行器按 strategy 与 mode 分支，不按 providerId 分支。直传执行过程中 CORS/网络失败**不**自动改走中转（避免双写）。

### 安全约束（硬性）

- scoped key 由宿主生成（调用方不得指定任意云 key）
- content-length 上限 ≤ `sys.upload.maxSize`（及插件策略）
- content-type 白名单（若 scene 有限制）
- 过期受 `directUrlTTL` 约束；part URL 按需签发
- 操作单一（put 凭证不能 get 列表）；part 访问绑定会话与 partNumber
- Complete 必须 Head/Stat 或云 Complete 成功且 size/parts 匹配
- 私有桶默认；不签发永久公开写 URL；密钥永不进入浏览器

### 与 hash 去重

- 客户端可在 init 前算 SHA-256；提供则先查复用（秒传，仍在 init 阶段）
- 有 client hash 才走秒传；无 hash 的直传仍创建新记录
- 默认不做 complete 后宿主流式下载重算 hash

### 插件 Storage 与 filecap

`storagecap.Service` 提供 CreateDirectPut / ConfirmDirectPut / CreateDirectGet，以及 logical-path 版 Multipart 方法与探测。动态插件 path 授权覆盖直连/分片会话。guest 公共 `Storage().Put` 签名不变；底层 commit 在 SupportsMultipart 时可改为 UploadPart 流式提交（一期可延期，现有 temp-file + Put 仍可用）。第一期不扩展 `filecap`；浏览器走宿主 HTTP `direct-upload/*` 与 `chunked/*`。

### Provider 实现分工

| Provider | Put 直传 | Get 直链 | Multipart | 备注 |
|---|---|---|---|---|
| local | 否 | 否 | 云 Multipart=false | 永远 proxy；proxy_chunked 可用（拼装 + Put） |
| s3/aws | presigned PUT/GET | 是 | 必须 | SDK Multipart + PresignUploadPart |
| cos/oss/obs | presigned 或 form | 是 | 必须 | 对齐 S3 或原生分片 API |
| azure | SAS | 是 | 尽量 | Block blob；不支持则 false |
| qiniu | upload token / form | 下载域名签名 | 可延后 | 不支持则 `SupportsMultipart=false` |

宿主**不** import 云 SDK；签名与 Multipart 逻辑只在插件内。配置页补充浏览器直传/分片 CORS 运维说明。

## 对象 Multipart 与自动分片

### 能力分层 — 可选 MultipartProvider

与 `DirectAccessProvider` 同模式，不强制所有 Provider 实现：

```text
MultipartUploadProvider (optional on storagecap.Provider)
  SupportsMultipart(ctx) bool
  CreateMultipart / UploadPart / CompleteMultipart / AbortMultipart
  CreateMultipartPartAccess  // 直传 part 级签发
```

- `storagecap.Service` 提供 logical-path 版镜像方法 + 探测。
- 宿主 `storage.Service` 对 `NamespaceFiles` 解析 active provider 后透传。
- local：`SupportsMultipart=false`（云 Multipart 语义）；**proxy_chunked 仍可用**。

### 两层「支持分片」探测

| 标志 | 含义 | 谁为 true |
|------|------|-----------|
| `proxyChunked` | 宿主能否接收客户端分片中转 | 文件中心始终 true（交付后） |
| `cloudMultipart` | active provider 是否支持云 Multipart | S3 族等实现后 true |

自动策略（init）：

```text
if size > maxSize: reject
if size < threshold:
  prefer direct_single else proxy_single
else:  # size >= threshold
  if SupportsDirectPut && cloudMultipart: direct_multipart
  elif proxyChunked:                      proxy_chunked
  elif SupportsDirectPut:                 direct_single  # 文档化风险
  else:                                   proxy_single
```

`size >= threshold` 且无 `cloudMultipart` 时，**优先 `proxy_chunked`**，避免大文件单次直签 PUT。

### 中转分片实现

- 若 `cloudMultipart`：init 时 CreateMultipart；part 时 UploadPart；complete 时 CompleteMultipart + 写库。
- 否则：part 写入临时文件；complete 单次 Put + 写库。
- complete 前不将对象视为已登记业务文件；abort 丢弃会话并 best-effort AbortMultipart。

### 直传分片实现

- 不在 init 返回全部 part URL（过期与 payload 风险）。
- 不把 temporary_credentials 作为浏览器默认路径。
- complete 校验：会话未过期、parts 完整且连续、size 与声明一致、云 Complete 成功。

### 配置默认值

| Key | 默认 | 说明 |
|-----|------|------|
| `sys.upload.maxSize` | **200**（MB） | 单文件上限 |
| `sys.upload.multipartThresholdMB` | **100** | `size >= threshold` 触发自动分片评估 |
| `sys.upload.multipartPartSizeMB` | **8** | 中间 part 大小；MUST ≥ 云 minPart（S3 5MB） |
| `sys.upload.multipartMaxConcurrency` | **3** | 前端直传并行度建议值 |
| `sys.upload.multipartEnabled` | **true**（若实现） | 紧急回滚总开关 |

校验：全部为正整数；`multipartThresholdMB < maxSize`；`multipartPartSizeMB >= 5`。body limit 对齐运行时 `maxSize`。

## 前端模块与交互设计

前端在 `src/components/upload/` 下提供通用 `FileUpload`、`ImageUpload` 组件和配套上传逻辑，用于处理拖拽上传、图片卡片回显、进度、校验和错误反馈。组件通过 `v-model:value` 绑定文件 ID，并通过 `scene` 参数声明业务场景。默认 `uploadApi` 内置四策略执行，组件无需改 props。

```text
init → 读 strategy
  instantReuse → return
  direct+single → executeDirectTransfer → complete
  direct+multipart → slice → 并行 part-url+PUT → complete(parts)
  proxy+single → /file/upload
  proxy+multipart → chunked part 循环 → complete
失败 → abort；不跨 channel 静默降级
```

进度：已完成 part 字节 + 进行中 part 的 xhr progress。

系统管理下新增“文件管理”页面，统一承载文件列表、搜索、上传、批量删除、下载、详情和预览能力。页面默认开启预览模式；文件类型筛选使用 Select，下拉选项来自 `/file/suffixes`。详情弹窗展示完整元数据。

## 风险与迁移

| 风险 | 缓解 |
|------|------|
| 多云冲突影响文件中心 | 配置页提示只启用一个；错误码可读（含 `FILE_STORAGE_CONFLICT`） |
| 切换云后旧本地文件 | Get 本地回退 |
| 云 provider 未配置 | 与插件 Storage 相同 fail 语义 |
| 桶 CORS 未配导致直传/分片失败 | 配置页说明 + 明确错误 |
| 孤儿对象（init 后未 complete） | 短 TTL；abort；会话到期 best-effort AbortMultipart；GC 可选二期 |
| complete 与实际上传竞态 | complete 重试 Stat；会话未过期可重复 complete 幂等 |
| 伪造 complete / parts | 会话服务端持有 key/uploadId；Stat 或云 Complete 校验；鉴权同上传权限 |
| 七牛/Azure 与 S3 语义差异 | 中立 DirectAccess / Multipart 字段；插件内适配；不支持则 false + proxy_chunked |
| 集群多实例会话丢失 | 一期进程内 + 文档 sticky；API 边界隔离便于二期共享 store |
| 阈值与 maxSize 配置不当使自动分片不可达 | maxSize 默认 200；校验 threshold < maxSize |
| S3 最小 part 5MB | partSize 默认 8MB；complete 前校验 part 规则 |
| 中转分片仍占宿主带宽 | 预期内；≥threshold 且云可用时优先 direct_multipart |
| 默认 maxSize 200 加大攻击面 | 仍鉴权 + body limit 对齐 maxSize；可运行时调低 |
| 与 HTTP 表单 multipart 术语混淆 | API 使用 `chunked` / `encoding=multipart` 区分对象分片 |

迁移预期：未启云插件时行为与本地一致（含 proxy_chunked）；启用唯一云插件后新文件上云、旧本地文件仍可读；支持直连的云后端下前端优先直传；≥ 阈值自动分片。不做批量迁移工具。回滚：禁用云插件、调高阈值/关闭 multipartEnabled、或关闭直传路径回到中转。

兼容：旧客户端忽略 strategy 时，init 若返回 direct single access 行为与既有双模式一致；proxy 仍整包上传。新字段均为 additive。

**非目标（刻意不做）**：断点续传 / 跨浏览器恢复；浏览器默认 STS + 厂商 SDK；修改插件 guest 公共 Put 签名；下载分片；跨实例会话强一致首期强制实现；批量迁移历史对象。

## 运行时治理评估

统一后端选择引入可诊断的存储冲突错误源码文案；文件访问继续以 `sys_file` 元数据与 `sys_file.path` 为权威，不引入额外分布式缓存一致性问题。数据权限边界保持不变：上传记录仍归属当前租户，读取和下载继续依赖既有元数据查询、租户过滤和权限校验；直传/分片 init/complete/chunked 继承上传权限与租户，会话绑定租户。直连与分片会话一期为进程内存储，无业务缓存一致性变更。DI：`file.New` 自建会话 store；`storage.NewResolvingService` 复用启动期 runtime；无新运行期依赖 owner。开发工具跨平台：无 Makefile/脚本变更要求。i18n：API 文档、错误码与配置/CORS 运维说明走双语治理。
