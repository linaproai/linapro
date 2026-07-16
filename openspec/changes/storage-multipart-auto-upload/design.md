## Context

文件中心已具备：

- 服务端中转：`POST /file/upload`（整包 `multipart/form-data`）→ spool → `storage.Put` → `sys_file`
- 客户端直传：`direct-upload/init` → 单次 `presigned_url` / `form_post` → `complete` / `abort`
- 前端 `uploadApi`：init 优先直传，`proxy` 回退整包中转
- 配置：`sys.upload.maxSize` 默认 100MB；`sys.upload.directUrlTTL` 默认 1h

插件 Storage 已具备：

- 领域 `storagecap.Service.Put(io.Reader)` + 可选 `DirectAccessProvider`
- 动态插件 guest 传输分片 `put.init/chunk/commit/abort`（宿主临时文件再 **单次** `Service.Put`）
- 云插件 `PutObject` / 单次 PresignPut；**无** CreateMultipartUpload 契约

归档 `file-storage` 已将「大文件单次 presign / multipart」标为二期。本变更交付该二期能力，并覆盖中转与直传两侧的自动分片。

约束：

- 厂商细节不得进入宿主；前端不得绑云 SDK
- 唯一 active storage provider；多云冲突 fail-closed
- 直传执行失败不静默改中转（避免双写）
- object key 仍由宿主分配；complete 前不写成功态 `sys_file`

## Goals / Non-Goals

**Goals:**

1. 领域层可探测的对象 Multipart 能力（Provider 可选接口 + Service/宿主透传）。
2. 文件中心 **proxy_chunked** 与 **direct_multipart** 完整生命周期。
3. **自动分片**：`size >= threshold` 且后端支持时自动选用分片 encoding；否则 single。
4. `sys.upload.maxSize` 默认 **200MB**；`sys.upload.multipartThresholdMB` 默认 **100MB**。
5. 前端 `uploadApi` 四策略自动执行；业务组件 API 不变。
6. 与现有秒传、权限、租户、路径规则兼容。

**Non-Goals:**

- 断点续传 / 跨浏览器恢复（可二期；会话结构预留 parts 列表即可）。
- 客户端携带 STS + 厂商 SDK 作为默认直传路径。
- 修改插件 guest 公共 `Storage().Put` 签名；guest 仍可只暴露单流 Put。
- 下载分片 / 分片并行下载。
- 跨实例会话强一致的首期强制实现（一期进程内；接口隔离以便后续共享 store）。
- 批量迁移历史对象或改变既有 `sys_file.path` 语义。

## Decisions

### D1：能力分层 — 可选 MultipartProvider

与 `DirectAccessProvider` 同模式，不强制所有 Provider 实现：

```text
MultipartUploadProvider (optional on storagecap.Provider)
  SupportsMultipart(ctx) bool
  CreateMultipart(ctx, ProviderMultipartCreateInput) (*MultipartSession, error)
  UploadPart(ctx, ProviderMultipartPartInput) (*MultipartPartResult, error)
  CompleteMultipart(ctx, ProviderMultipartCompleteInput) (*ProviderObject, error)
  AbortMultipart(ctx, ProviderMultipartAbortInput) error
  // 直传
  CreateMultipartPartAccess(ctx, ProviderMultipartPartAccessInput) (*DirectAccess, error)
```

- `storagecap.Service` 提供 logical-path 版镜像方法 + `SupportsMultipart` 探测。
- 宿主 `storage.Service` 对 `NamespaceFiles` 解析 active provider 后透传；其它 namespace 首期可不支持或仅 local 拼盘。
- local：`SupportsMultipart=false`（云 Multipart 语义）；**proxy_chunked 仍可用**（宿主临时拼装 + 单次 Put）。

**替代方案**：强制 Provider 全实现 Multipart — 拒绝，会拖垮 local 与差异云。

### D2：两层「支持分片」探测

| 标志 | 含义 | 谁为 true |
|------|------|-----------|
| `proxyChunked` | 宿主能否接收客户端分片中转 | 文件中心始终 true（本变更交付后） |
| `cloudMultipart` | active provider 是否支持云 Multipart | S3 族等实现后 true |

自动策略（文件中心 plan/init）：

```text
if size > maxSize: reject
if size < threshold:
  prefer direct_single else proxy_single
else:  # size >= threshold
  if SupportsDirectPut && cloudMultipart: direct_multipart
  elif proxyChunked:                      proxy_chunked
  elif SupportsDirectPut:                 direct_single  # 文档化风险；或拒绝超阈值无分片
  else:                                   proxy_single   # 若仍 ≤ 网关可承受
```

**推荐默认（本设计采用）**：`size >= threshold` 且无 `cloudMultipart` 时，**优先 `proxy_chunked`**，避免大文件单次直签 PUT；仅当 proxy_chunked 也不可用时才 fail closed。

### D3：统一规划入口 — 扩展 `direct-upload/init`

扩展现有 init 响应（非新建平行 plan API，降低前端分支）：

```json
{
  "instantReuse": false,
  "uploadSessionId": "...",
  "strategy": {
    "channel": "direct|proxy",
    "encoding": "single|multipart"
  },
  "multipart": {
    "partSize": 8388608,
    "minPartSize": 5242880,
    "maxParts": 10000,
    "maxConcurrency": 3
  },
  "access": { "... single direct only ..." }
}
```

- `channel=proxy` + `encoding=single`：走 `/file/upload`（兼容现状）。
- `channel=proxy` + `encoding=multipart`：走 `/file/upload/chunked/*`，`uploadSessionId` 为中转会话。
- `channel=direct` + `encoding=single`：现有 access 单次传输。
- `channel=direct` + `encoding=multipart`：init 时宿主 `CreateMultipart`；客户端通过 part-url 接口取各 part 签名。

**替代方案**：独立 `/file/upload/plan` — 可后续拆分；一期扩展 init 减少往返与重复校验。

### D4：中转分片 HTTP 契约

```text
POST /file/upload/chunked/init     // 可由 strategy 隐含；也可显式（若 init 已建会话则可省略）
POST /file/upload/chunked/part     // form: uploadSessionId, partNumber, file
POST /file/upload/chunked/complete
POST /file/upload/chunked/abort
```

实现策略：

- 若 `cloudMultipart`：init 时 CreateMultipart；part 时 UploadPart；complete 时 CompleteMultipart + 写库。
- 否则：part 写入临时文件（有序或按 partNumber）；complete 单次 Put + 写库。
- 权限标签与 `/file/upload` 相同：`system:file:upload`；租户/用户绑定会话。

### D5：直传分片 — part 级短时签名

```text
POST /file/direct-upload/part-url  { uploadSessionId, partNumber }
  → DirectUploadAccess (presigned_url for one part)
POST /file/direct-upload/complete  { uploadSessionId, parts: [{partNumber, etag}] }
POST /file/direct-upload/abort
```

- 不在 init 返回全部 part URL（过期与 payload 风险）。
- 不把 temporary_credentials 作为浏览器默认路径。
- complete 校验：会话未过期、parts 完整且连续、size 与声明一致、云 Complete 成功、Stat 可选二次确认。

### D6：配置默认值

| Key | 默认 | 说明 |
|-----|------|------|
| `sys.upload.maxSize` | **200**（MB） | 原 100 → 200；config.yaml 静态默认同步 |
| `sys.upload.multipartThresholdMB` | **100** | `size >= threshold * 1MiB` 触发自动分片评估 |
| `sys.upload.multipartPartSizeMB` | **8** | 中间 part 大小；MUST ≥ 云 minPart（S3 5MB） |
| `sys.upload.multipartMaxConcurrency` | **3** | 前端直传并行度建议值 |

校验：

- 全部为正整数。
- `multipartThresholdMB < maxSize`（否则自动分片不可达，启动/写配置时告警或拒绝）。
- `multipartPartSizeMB >= 5`（与 S3 兼容下限对齐；更严云在 provider 内再抬升）。

代码常量：`defaultUploadMaxSize = 200`；SQL seed 同步更新。

### D7：前端 `uploadApi` 自动执行

```text
init → 读 strategy
  instantReuse → return
  direct+single → 现有 executeDirectTransfer → complete
  direct+multipart → slice(file, partSize) → 并行 part-url+PUT → complete(parts)
  proxy+single → /file/upload
  proxy+multipart → chunked part 循环 → complete
失败 → abort；不跨 channel 静默降级
```

进度：已完成 part 字节 + 进行中 part 的 xhr progress。

### D8：会话存储一期进程内

- 扩展 `directUploadSession`：encoding、cloudUploadId、providerId、parts 状态、partSize。
- 新增 proxy chunked session store（可与 direct 同形不同 map，或统一 `uploadSession` 抽象）。
- 多实例：文档说明 sticky / 单实例；后续共享 cache 不改 API。

### D9：动态插件 WASM upload

- 公共 API 不变：`Put` 仍自动 init/chunk/commit。
- 优化（本变更可选任务）：commit 时若 `SupportsMultipart`，将临时文件按 partSize 切 UploadPart，避免云侧单次 PutObject 大对象限制。
- 不要求动态插件 manifest 新增方法。

### D10：官方云插件优先级

| Provider | 一期 Multipart | 备注 |
|----------|----------------|------|
| s3 / aws | 必须 | SDK Multipart + PresignUploadPart |
| cos / oss / obs | 必须 | 对齐 S3 或原生分片 API |
| azure | 尽量 | Block blob 阶段块 + commit |
| qiniu | 可延后 | 不支持则 `SupportsMultipart=false` |

配置页：CORS 说明补充「分片 PUT 同样需要允许的 Method/Header」。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 阈值=100 且 maxSize 未抬高导致自动分片不可达 | maxSize 默认 200；配置校验 threshold < maxSize |
| S3 最小 part 5MB / 最后一块例外 | partSize 默认 8MB；complete 前校验 part 规则 |
| 集群多实例会话丢失 | 一期进程内 + 文档；API 边界隔离便于二期共享 store |
| 直传 part 签名过期 | part-url 按需签发；TTL 复用 directUrlTTL；客户端可重签同 part |
| 孤儿 Multipart 占用云资源 | abort API + 会话 TTL 到期 best-effort AbortMultipart；二期 GC job |
| 中转分片仍占宿主带宽 | 预期内；≥threshold 且云可用时优先 direct_multipart |
| complete 伪造 parts | 会话绑定 uploadId/key/tenant；云 Complete 失败则拒绝写库 |
| 与 form-data「multipart」术语混淆 | 文档/API 使用 `chunked` / `encoding=multipart`（对象分片）区分 HTTP 表单 |
| Azure/七牛语义差 | 可选接口 + 显式 unsupported；自动策略走 proxy_chunked |
| 默认 maxSize 200 加大攻击面 | 仍鉴权 + body limit 对齐 maxSize；可运行时调低 |

## Migration Plan

1. 部署配置 seed：`maxSize=200`，新增 multipart 三键；已有运行时覆盖值不强制迁移。
2. 先上 Provider Multipart + 探测（行为对旧客户端无感）。
3. 再上 chunked / part-url API 与扩展 init 字段（旧前端忽略 strategy 时：仍可按 access.mode 走旧双模式；**新前端必读 strategy**）。
4. 前端发布后自动分片生效。
5. 回滚：关闭阈值（阈值调到 ≥ maxSize 或极大）或 feature 开关（若实现 `sys.upload.multipartEnabled`，默认 true）；禁用后 encoding 恒为 single。

**兼容**：旧客户端不传/忽略 strategy 时，init 若返回 direct single access 行为与现网一致；proxy 仍整包上传。新字段均为 additive。

## Open Questions

1. 是否增加 `sys.upload.multipartEnabled` 总开关？（建议：要，默认 true，便于紧急回滚。）
2. 直传 multipart 失败时是否允许**用户显式**重试为 proxy_chunked？（默认否，与一期「不静默降级」一致；UI 可提示重新上传。）
3. 插件 `Storage()` 是否暴露公共 Multipart API，还是仅内部优化 Put？（建议一期仅内部 + 能力探测；公共 API 二期按需。）
