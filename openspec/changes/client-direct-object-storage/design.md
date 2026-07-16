## Context

当前对象访问分层：

| 层 | 职责 |
|---|---|
| 文件中心 `Files()` / `sys_file` | 业务文件元数据、scene、权限、公开 URL、hash 去重 |
| 插件 `Storage()` / `storagecap.Service` | 插件 logical path + 租户/插件隔离 |
| `storagecap.Provider` | 仅 scoped object key 的 Put/Get/Delete/List/Stat |
| 宿主 `storage.Service` | 中立 namespace/key 本地与 resolving 封装 |
| `ResolveProvider` | 0 云→local、1 云→该云、≥2→冲突 |

字节路径今天全部经宿主：浏览器 multipart → file service → Provider.Put；下载经 file access/download → Provider.Get。云插件 README 将 Presigned URLs 列为 Non-goal。

产品要求在**同一套通用后端选择与隔离模型**上增加客户端直连，避免业务/前端绑定厂商。

## Goals / Non-Goals

**Goals:**

- 提供与 `Put`/`Get` 对称的**可探测**客户端直连能力（上传 + 下载）
- 文件中心完整生命周期：`init` → 客户端直传 → `complete`/`abort`；元数据仍由宿主写入
- 插件 `Storage()` 可签发/确认直连访问，path 与租户隔离不变
- 官方云 provider 实现 DirectAccess；local 与不支持能力时**自动降级**服务端中转
- 前端 `FileUpload`/`ImageUpload` 无厂商分支；失败可诊断（CORS 等）
- 密钥与过宽凭证永不进入浏览器；会话绑定租户、用户、key、size/type 约束

**Non-Goals:**

- 跨 provider 对象迁移 / 多云同时可写
- 浏览器内嵌各云官方 SDK 作为默认路径（STS mode 可作为可选高级载荷，页面默认不用）
- 第一期强制云端 multipart 分片直传全厂商打齐（单对象 presign/form 优先；multipart 作为契约可选扩展）
- 改变 `sys_file` 列表数据权限模型
- 将云 SDK 引入 `lina-core` go.mod
- 动态 WASM guest 在沙箱内“直连云”（guest 仍走 host；浏览器插件 UI 可走宿主 HTTP API）

## Decisions

### 1. 契约形态：Provider 能力探测 + 中立 DirectAccess 描述

**决策**：在 `storagecap` 增加可选能力接口（或 Provider 方法集扩展），而非新建平行总线。

概念：

```text
SupportsDirectAccess(op) -> bool
CreateDirectAccess(DirectAccessInput) -> DirectAccess
// DirectAccess 中立字段：
//   Mode: presigned_url | form_post | temporary_credentials
//   Operation: put | get | (optional) multipart_*
//   Method, URL, Headers, FormFields, ExpiresAt
//   凭证字段仅 temporary_credentials 使用，且必须短时+最小权限
```

**理由**：与现有 Provider 扩展点一致；local 可不实现；宿主 adapter 统一探测后降级。

**备选**：每厂商独立 HTTP API → 拒绝，破坏通用性。  
**备选**：仅 STS → 前端过重，默认路径否决。

### 2. 领域生命周期：宿主会话，而不是“只签 URL”

**决策**：文件中心与插件 Storage 均采用：

```text
Init(业务约束) → 返回 uploadSessionId + DirectAccess 或 mode=proxy
客户端执行传输（若 direct）
Complete(sessionId, etag?, clientHash?) → 校验对象 → 写领域状态
Abort(sessionId) → 尽力清理会话/未完成对象（best-effort）
```

**理由**：只发 URL 不 complete 会导致：无 `sys_file`、无权限记录、hash 去重失效、孤儿对象。会话将 key、size 上限、content-type、tenant、actor、scene 绑定在服务端（内存/缓存，TTL 对齐签名过期）。

**会话存储**：优先进程内 + 可选共享缓存键（与集群模式对齐时用既有 cache/coordination）；单机 dev 内存即可。会话不落 `sys_file` 直至 complete 成功。

### 3. 传输 mode 优先级

| 优先级 | Mode | 使用场景 |
|---|---|---|
| 1 | `presigned_url`（PUT/GET） | S3/AWS/COS/OSS/OBS/Azure SAS 等 |
| 2 | `form_post` | 需 POST policy 的厂商/浏览器限制场景 |
| 3 | `temporary_credentials` | 大文件 SDK / 高级客户端；工作台默认不启用 |
| fallback | `proxy` | local、不支持、冲突、配置失败前的明确指示 |

前端通用执行器按 mode 分支（少量），**不按 providerId 分支**。

### 4. 文件中心 API 形状

在现有 `api/file/v1` 旁新增（路径以最终 api-contract 为准，语义固定）：

- `POST /api/v1/files/direct-upload/init`
- `POST /api/v1/files/direct-upload/complete`
- `POST /api/v1/files/direct-upload/abort`（可选但建议）
- 下载：现有 download/access 在可直链时返回 `302` 到预签名 URL，或响应体携带 `directUrl` 由前端决定；**第一期推荐**：API 增加 `preferDirect` 查询/字段，JSON 下载元数据接口返回直链，二进制流接口保持中转兼容旧客户端。

Init 请求：`scene`、`fileName`、`size`、`contentType`、可选 `contentHash`（SHA-256 hex）。  
若 hash 已存在且可复用 → 直接返回已有文件信息（秒传），不签发上传。

Complete：校验会话、`Stat` 对象 size（及可选 etag）、写 `sys_file`、engine=active provider。

### 5. 插件 Storage 直连

`storagecap.Service` 增加例如：

- `CreateDirectPut` / `CreateDirectGet`
- `ConfirmDirectPut`（对 Storage 领域：确认后对象即可 `Get`；无 `sys_file`）

动态插件：host service 方法授权匹配 logical path；会话绑定 pluginId + path。

### 6. Provider 实现分工

| Provider | Put 直传 | Get 直链 | 备注 |
|---|---|---|---|
| local | 否 | 否 | 永远 proxy |
| s3/aws | presigned PUT/GET | 是 | multipart 二期可选 |
| cos/oss/obs | presigned 或 form | 是 | 按 SDK 最稳路径 |
| azure | SAS | 是 | |
| qiniu | upload token / form | 下载域名签名 | 与现有 Get 语义对齐 |

宿主**不** import 云 SDK；签名逻辑只在插件内。

### 7. 安全约束（硬性）

Init/CreateDirectAccess 必须钉死：

- scoped key 由宿主生成（调用方不得指定任意云 key）
- content-length 上限 ≤ `sys.upload.maxSize`（及插件策略）
- content-type 白名单（若 scene 有限制）
- 过期默认 15m（可配置上界，如 ≤1h）
- 操作单一（put 凭证不能 get 列表）
- Complete 必须 Head/Stat 成功且 size 匹配（允许云侧尾差策略写死在实现说明中）
- 私有桶默认；不签发永久公开写 URL

### 8. 前端策略

`useUpload` / `customRequest`：

1. 调 init  
2. 若 `mode=proxy` 或 init 指示不支持 → 现有 multipart  
3. 若 direct → 按 mode 上传 → complete  
4. complete 失败 → 展示可读错误；可选 abort  

CORS 失败：网络层错误映射为运维提示（i18n），不静默当成功。

**降级策略决策**：init 返回 proxy 时自动中转；**直传执行过程中** CORS/网络失败**不**自动改走 multipart（避免双写与用户困惑），提示检查桶 CORS 或改用中转开关（系统配置 `sys.upload.preferDirect` 默认 true，可关）。

### 9. 与 hash 去重

- 客户端可在 init 前算 SHA-256；提供则先查复用  
- 未提供 hash：complete 后可用对象 ETag 不可靠等同内容 hash；**完整内容 hash 可选异步补齐或不强求秒传**  
- 第一期：有 client hash 才走秒传；无 hash 的直传仍创建新记录（或 complete 后宿主流式下载重算——成本高，**默认不做**）

### 10. 配置与运维面

各云插件设置页 Alert 增加：

- 浏览器直传需配置桶 CORS（允许工作台 Origin、PUT/GET/POST、必要 Header）
- 唯一 active provider 与 fail-closed 说明保持

可选宿主配置：

- `sys.upload.preferDirect`（bool，默认 true）
- `sys.upload.directUrlTTL`（duration，有界）

### 11. 测试策略

- 宿主：fake Provider 实现 DirectAccess + 文件 init/complete 单测  
- 各插件：签名生成单测（可用 HTTP mock / 本地 minio 集成可选）  
- 前端：hook 单测 mode 分支  
- E2E：local 路径仍 multipart；若环境无真实云，用 stub 或跳过云直传 E2E 并保留 API 契约测试  

## Risks / Trade-offs

| 风险 | 缓解 |
|---|---|
| 桶 CORS 未配导致直传失败 | 配置页说明 + 明确错误；preferDirect 可关 |
| 孤儿对象（init 后未 complete） | 短 TTL；abort；生命周期 GC 任务可选二期 |
| complete 与实际上传竞态 | complete 重试 Stat；会话未过期可重复 complete 幂等 |
| 伪造 complete | 会话服务端持有 key；Stat 校验；鉴权同上传权限 |
| 七牛/Azure 与 S3 语义差异 | 中立 DirectAccess 字段；插件内适配 |
| 集群多实例会话丢失 | 会话写入共享 cache；或 complete 携带签名服务端可重验的 mac（第二选择） |
| 大文件单次 presign 限制 | 文档说明上限；二期 multipart |

## Migration Plan

1. 先合入契约与宿主 init/complete + local/proxy 路径（行为与现网一致）。  
2. 逐个云插件实现 DirectAccess；未实现前 `SupportsDirectAccess=false`。  
3. 前端双模式上线，`preferDirect` 默认 true。  
4. 文档与插件 README 更新 Non-goals → 支持说明。  
5. 回滚：关闭 `preferDirect` 或禁用直传代码路径即可全站回到中转。

## Open Questions

- 下载直链默认 302 还是 JSON 返回 URL：建议**元数据/下载意图 API 返回 URL，流式 endpoint 保持中转**，避免破坏 `<img src="/uploads/...">` 现有公开访问模型；公开 `/uploads/*` 是否改签直链可作为后续优化。  
- 集群会话：第一期若仅内存，多副本需 sticky 或共享 cache——实现时优先复用已有 cache capability。  
- 云 multipart 是否纳入本变更 tasks：建议 **Phase 内契约预留、实现可选**；tasks 以单对象直传打满七云 + 文件中心 + 前端为主。
