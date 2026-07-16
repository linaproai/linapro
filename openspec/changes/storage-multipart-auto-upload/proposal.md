## Why

文件中心与插件 Storage 目前仅支持「整对象」上传路径：服务端中转走单次 `multipart/form-data` + 单次 `Put`，客户端直连走单次 `presigned_url` / `form_post`。大文件会放大宿主带宽与超时风险，单次直签 PUT 也难以稳定覆盖百兆级对象。需要在不破坏既有双模式与厂商中立契约的前提下，为**服务端中转**与**客户端直传**同时补齐对象分片上传，并按可配置阈值**自动选择**分片路径。

## What Changes

- **对象分片领域能力**：在 `storagecap.Provider` / `storagecap.Service` 与宿主 `storage.Service` 上增加可探测的 Multipart 会话能力（Create / UploadPart / Complete / Abort，以及直传用的 part 级访问签发），local 与未实现 Multipart 的 provider 明确降级为不支持云 Multipart。
- **服务端中转分片**：文件中心新增 chunked 上传 API（init / part / complete / abort）；宿主在 commit 时优先走 Provider Multipart，否则拼临时对象后单次 Put。
- **客户端直传分片**：扩展 direct-upload 生命周期，支持云 Multipart（part 级短时签名 + complete 携带 parts）；complete 时由宿主调用云 Complete 再写 `sys_file`。
- **自动分片策略**：统一上传规划（复用/扩展 `direct-upload/init` 或等价 plan）根据文件大小、阈值与能力探测返回 `channel`（direct|proxy）与 `encoding`（single|multipart）；前端 `uploadApi` 自动执行，业务组件无感。
- **上传默认上限调整**：`sys.upload.maxSize` 默认值由 **100MB 调整为 200MB**；新增 `sys.upload.multipartThresholdMB`（默认 **100MB**）等分片相关运行时配置，且阈值必须小于有效 `maxSize` 才有自动分片空间。
- **官方云 provider 补齐**：S3 兼容族（s3/aws）与支持 Multipart 的国内云优先实现；Azure / 七牛等差异后端按可生产路径实现或明确 `SupportsMultipart=false`。
- **非破坏**：现有 `/file/upload` 整包上传、单次直传、插件 `Storage().Put` 单流路径全部保留；动态插件既有 `put.init/chunk/commit` 传输分片语义保留，可在底层升级为 Provider Multipart。

## Capabilities

### New Capabilities

- `object-multipart-upload`：对象存储 Multipart 领域契约、能力探测、中转/直传分片协议、自动分片阈值策略与配置项。

### Modified Capabilities

- `client-direct-object-access`：直连 put 从「仅单对象」扩展为「单对象或 Multipart」；init/complete 载荷与前端 dual-mode 策略扩展为四策略（direct/proxy × single/multipart）。
- `plugin-storage-service`：`storagecap` 增加可选 Multipart 会话与探测；大对象 Put 可在支持时走 Multipart 写后端。
- `cloud-storage-provider-plugins`：官方云插件实现 Multipart（或显式不支持）；配置页可补充分片/CORS 运维说明。
- `host-object-storage`：宿主内部 `storage.Service` 透传 NamespaceFiles 的 Multipart / part 签发能力。
- `file-upload-storage-path`：分片路径下 object key / `sys_file.path` 规则与整包上传保持一致。

## Impact

- **后端 API**：`api/file/v1` 新增/扩展 chunked 与 direct multipart 相关接口；`storagecap`、`internal/service/storage`、`internal/service/file`。
- **配置**：`manifest/sql` 配置 seed、`config` 运行时参数校验、`sys.upload.maxSize` 默认 200、新增 multipart 阈值/分片大小等键。
- **云插件**：`apps/lina-plugins/linapro-storage-*` Provider 实现与测试。
- **前端**：`apps/lina-vben/.../api/core/upload.ts` 与通用上传组件路径；业务页 props 不变。
- **动态插件**：WASM storage upload 会话可复用/对接 Provider Multipart（可选优化，不改变 guest 公共 API）。
- **测试 / i18n / 数据权限**：init/complete/chunked 继承上传权限与租户；错误码与配置说明双语；单测 + 关键路径 E2E。
- **无 BREAKING 预期**：默认阈值 100MB、maxSize 200MB 下，≤100MB 行为与现网一致；仅更大文件与配置变更后行为变化。
