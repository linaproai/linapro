## Why

当前对象存储虽已通过 `storagecap.Provider` 与统一后端选择实现了**服务端中转**的通用云接入，但浏览器、插件前端与外部客户端上传/下载大文件时，全部字节仍必须经过 `lina-core`。这会放大宿主带宽与 CPU 成本，拉长上传时延，并限制大文件体验。与此同时，预签名直传、临时凭证与 CDN 下载在各云厂商 README 中长期列为 Non-goal，调用方也无法在不感知厂商的前提下获得与后端 `Put`/`Get` 同等通用的直连能力。

需要在**不破坏现有领域边界与唯一 active provider 语义**的前提下，交付与服务端中转对称的**客户端直连对象访问**能力：宿主签发短时访问能力、客户端直连云、完成后由宿主确认并落领域元数据。

## What Changes

- **通用直连契约**：扩展 `storagecap.Provider`（及文件中心解析路径）支持可探测的客户端直连访问签发：`presigned_url` / `form_post` / `temporary_credentials` 等中立 mode，禁止向前端暴露长期 AK/SK 或厂商专有调用面。
- **文件中心直传生命周期**：新增文件中心 `init` → 客户端上传 → `complete`（及可选 `abort`）API；init 完成鉴权、配额、scene、后缀、maxSize、key 规划与会话绑定；complete 做对象存在性/大小校验并写入 `sys_file`（含 hash 秒传路径）。
- **下载直链（对称能力）**：对已落库且内容在云上的对象，支持签发短时 GET 访问（预签名或等价），可配置或按场景选择「宿主中转 / 直链跳转」；local 与不可签发时保持现有中转。
- **插件 Storage 直连**：`storagecap.Service` 增加直连 put/get 访问签发与确认语义，供源码插件后端代客户端编排或插件前端经宿主 API 使用；隔离与 path 授权规则与现有 `Put`/`Get` 一致。
- **Provider 插件补齐**：官方云插件（COS/OSS/OBS/七牛/AWS/Azure/S3）实现 DirectAccess；local 明确 `SupportsDirectAccess=false`，调用方自动降级 multipart/stream 中转。
- **前端通用上传**：`FileUpload` / `ImageUpload` 在支持直传时走 init→直连→complete，否则静默回退现有 multipart；业务页面无厂商分支。
- **运维与安全**：配置页补充直传所需 CORS/桶策略说明；多云冲突、配置无效 fail-closed 语义与现网一致；会话过期与 complete 校验防脏元数据。

**非破坏性**：现有 multipart 上传、宿主中转下载、插件 `Storage().Put/Get` 流式路径全部保留，作为 local 与能力缺失时的默认路径。

## Capabilities

### New Capabilities

- `client-direct-object-access`：客户端直连对象访问的通用契约与生命周期（能力探测、访问模式、init/complete/abort、安全 scope、local 降级、与领域元数据确认的边界）。

### Modified Capabilities

- `plugin-storage-service`：`storagecap.Service` / `Provider` 增加直连访问签发与确认；动态插件 path 授权覆盖直连会话；文档边界从 Non-goal 转为正式能力。
- `cloud-storage-provider-plugins`：各官方云 provider 必须实现完整服务端对象契约之外的 DirectAccess（或明确不可用并走宿主降级策略的统一探测结果）；配置页补充直传 CORS 运维说明。
- `unified-object-backend`：文件中心内容写入/读取在 active 云后端支持直连时，允许字节路径绕过宿主，但后端选择与冲突规则不变；元数据仍以 `sys_file` 为准。
- `file-upload-storage-path`：直传 init 生成的对象 key 必须遵守既有租户分区与路径规则；complete 写入的 `sys_file.path` 与中转上传一致。

## Impact

- **宿主 `lina-core`**：`storagecap` 契约扩展；文件服务新增 direct-upload / direct-download 会话与 API；下载路径可选签发直链；单元测试与 DI 说明。
- **云存储插件**：七个 `linapro-storage-*` 的 `objstore` 实现 DirectAccess；README Non-goals 调整；配置页 i18n 增加 CORS 提示。
- **前端 `lina-vben`**：通用上传 hook 双模式；可选下载直链消费；i18n 错误文案。
- **插件 SDK / 动态 host service**：`Storage()` 直连方法与授权；大对象仍可选用既有分片中转。
- **E2E / 测试**：直传成功路径（可用 mock/fake provider）、local 回退、complete 校验失败、多 provider 冲突。
- **安全与数据权限**：init/complete 继承文件中心权限与租户边界；直链仅短时；不改变列表数据权限模型。
- **无影响项（默认）**：不引入跨 provider 迁移工具；不改变「唯一可服务云插件」选择规则；不把云 SDK 引入宿主 go.mod。
