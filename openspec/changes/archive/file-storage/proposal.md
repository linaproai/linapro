## Why

当前系统需要一套统一、可扩展且便于治理的文件存储能力。早期上传功能分散且实现方式不一致：用户头像通过独立接口上传到本地文件系统，通知公告中的图片则以 Base64 编码直接存储在数据库字段中。Base64 方式对数据库存储压力大，也让附件上传、预览、下载、删除、去重和后续对象存储扩展缺乏统一的能力支点。

随着文件管理模块落地，系统还需要让存储路径、访问语义和租户隔离规则更加清晰。普通文件上传曾使用 `t/<tenantId>/...` 作为相对存储路径，其中 `t` 仅是 tenant 的缩写，对调用方和运维排查不够直观。需要在保留租户分区组织方式的前提下，简化新上传文件的物理路径，同时确保历史文件记录无需迁移即可继续访问。

在此基础上，文件中心（`sys_file`）的内容读写曾固定走宿主本地 `storage.Service`（namespace `files`），与插件 `Storage()` / `storagecap` 的「0 云→local、1 云→云、≥2→冲突」后端选择分离。产品期望：**所有字节级文件存储（上传/下载/删除）统一对象后端**；无云插件时默认本地。列表/检索仍走 `sys_file` 元数据。

对象存储虽已通过 `storagecap.Provider` 与统一后端选择实现了服务端中转的通用云接入，但浏览器、插件前端与外部客户端上传/下载大文件时，全部字节仍必须经过 `lina-core`。这会放大宿主带宽与 CPU 成本，拉长上传时延，并限制大文件体验。需要在不破坏现有领域边界与唯一 active provider 语义的前提下，交付与服务端中转对称的**客户端直连对象访问**能力：宿主签发短时访问能力、客户端直连云、完成后由宿主确认并落领域元数据。

即便具备整对象中转与单次直签路径，大文件仍会放大宿主带宽与超时风险，单次直签 PUT 也难以稳定覆盖百兆级对象。因此在双模式与厂商中立契约之上，进一步为**服务端中转**与**客户端直传**同时补齐对象分片上传，并按可配置阈值**自动选择**分片路径；业务组件保持无感。

## What Changes

- **通用文件存储与上传能力**：设计统一的文件存储接口（Storage Interface），默认实现本地存储，后续可扩展 OSS（阿里云、腾讯云、MinIO 等），并由统一上传 API 处理各类文件上传请求。
- **文件管理数据模型与业务模块**：新增 `sys_file` 表记录上传文件元信息，前端新增文件管理页面，统一提供文件列表、上传、下载、删除、预览、筛选和详情展示能力。
- **现有上传场景统一接入**：将通知公告富文本图片、通知公告附件和用户头像上传统一接入通用文件上传接口，移除分散的上传实现和 Base64 图片存储方式。
- **本地存储路径简化**：普通文件新写入物理文件时，使用 `<tenantId>/<yyyy>/<MM>/<generated-file-name>` 作为相对存储路径，不再增加 `t` 目录层。
- **历史路径兼容与物理文件复用**：保留 `sys_file.path` 中已记录的历史路径，不迁移旧文件；上传和访问继续以数据库记录路径为准，并保持按租户与文件 hash 的去重语义，相同内容重复上传时继续复用已有物理文件。
- **统一对象后端选择**：引入宿主级统一对象后端选择（复用 `storagecap.ResolveProvider`）。文件中心 Put/Get/Delete **内容**经统一后端；provider 键前缀 `files/` 与插件私有对象隔离。
- **本地 `files/` 命名空间路由**：本地 provider 支持将 `files/` 键路由到 `NamespaceFiles`（兼容既有磁盘布局）。
- **云未命中本地回退**：读取时对历史本地文件做云未命中→本地回退（过渡期）；新写入 `sys_file.engine` 记录生效 provider id（`local` 或插件 id）。
- **元数据路径不变**：列表/检索/浏览元数据路径不变，仍以 `sys_file` 为权威。
- **通用直连契约**：扩展 `storagecap.Provider`（及文件中心解析路径）支持可探测的客户端直连访问签发：`presigned_url` / `form_post` / `temporary_credentials` 等中立 mode，禁止向前端暴露长期 AK/SK 或厂商专有调用面。
- **文件中心直传生命周期**：新增文件中心 `init` → 客户端上传 → `complete`（及可选 `abort`）API；init 完成鉴权、配额、scene、后缀、maxSize、key 规划与会话绑定；complete 做对象存在性/大小校验并写入 `sys_file`（含 hash 秒传路径）。
- **下载直链（对称能力）**：对已落库且内容在云上的对象，支持签发短时 GET 访问（预签名或等价）；local 与不可签发时保持现有中转。
- **插件 Storage 直连**：`storagecap.Service` 增加直连 put/get 访问签发与确认语义；隔离与 path 授权规则与现有 `Put`/`Get` 一致。
- **Provider 插件补齐**：官方云插件实现 DirectAccess；local 明确不支持，调用方自动降级中转。
- **对象分片领域能力**：在 `storagecap.Provider` / `storagecap.Service` 与宿主 `storage.Service` 上增加可探测的 Multipart 会话能力（Create / UploadPart / Complete / Abort，以及直传用的 part 级访问签发）；local 与未实现 Multipart 的 provider 明确降级为不支持云 Multipart。
- **服务端中转分片**：文件中心新增 chunked 上传 API（init / part / complete / abort）；宿主在 commit 时优先走 Provider Multipart，否则拼临时对象后单次 Put。
- **客户端直传分片**：扩展 direct-upload 生命周期，支持云 Multipart（part 级短时签名 + complete 携带 parts）；complete 时由宿主调用云 Complete 再写 `sys_file`。
- **自动分片策略**：统一上传规划（扩展 `direct-upload/init`）根据文件大小、阈值与能力探测返回 `channel`（direct|proxy）与 `encoding`（single|multipart）；前端 `uploadApi` 自动执行，业务组件无感。
- **上传默认上限与分片配置**：`sys.upload.maxSize` 默认值 **200MB**；新增 `sys.upload.multipartThresholdMB`（默认 **100MB**）、`multipartPartSizeMB`、`multipartMaxConcurrency` 等运行时配置；阈值必须小于有效 `maxSize` 才有自动分片空间。
- **官方云 provider Multipart**：S3 兼容族（s3/aws）与 cos/oss/obs 实现 Multipart；Azure / 七牛等差异后端按可生产路径实现或明确 `SupportsMultipart=false`。
- **前端通用上传四策略**：`FileUpload` / `ImageUpload` 经 `uploadApi` 自动执行 direct/proxy × single/multipart；业务页面无厂商分支、无强制分片 props。
- **运维与安全**：配置页补充直传/分片所需 CORS/桶策略说明；多云冲突、配置无效 fail-closed；会话过期与 complete 校验防脏元数据；直传失败不静默改中转。

**非破坏性**：现有整包 `/file/upload`、单次直传、插件 `Storage().Put` 单流路径与动态插件 `put.init/chunk/commit` 传输分片语义全部保留；动态插件底层可升级为 Provider Multipart 而不改 guest 公共 API。

## Capabilities

### New Capabilities

- `file-storage`: 通用文件存储抽象层，支持本地存储（默认）和对象存储扩展，基于租户与 SHA-256 散列值实现文件去重与物理文件复用。
- `file-management`: 文件管理业务模块，提供文件列表、上传、下载、删除、预览、详情、按类型和使用场景筛选等功能。
- `file-upload-storage-path`: 统一普通文件上传的租户分区存储路径规则、历史路径兼容策略和验证要求（含分片路径与整包一致）。
- `unified-object-backend`: 全站对象字节后端选择与文件中心内容接入；与插件 `Storage()` 共享 `ResolveProvider` 规则。
- `client-direct-object-access`: 客户端直连对象访问的通用契约与生命周期（能力探测、访问模式、init/complete/abort、channel×encoding 策略、安全 scope、local 降级、与领域元数据确认的边界）。
- `object-multipart-upload`: 对象存储 Multipart 领域契约、能力探测、中转/直传分片协议、自动分片阈值策略与配置项。
- `host-object-storage`: 宿主内部 `storage.Service` 透传 `NamespaceFiles` 的 Multipart / part 签发能力。

### Modified Capabilities

- `user-avatar`: 用户头像上传改为调用通用文件上传接口，使用 `avatar` 场景标识并复用统一文件管理能力。
- `notice-editor`: 通知公告富文本编辑器图片上传改为调用通用文件上传接口，附件上传纳入统一文件存储与管理流程。
- `plugin-storage-service`: 文件中心内容与插件 Storage 共享 Resolve 规则；`storagecap.Service` / `Provider` 增加直连访问签发与可选 Multipart 会话；大对象 Put 可在支持时走 Multipart 写后端。
- `cloud-storage-provider-plugins`: 各官方云 provider 实现 DirectAccess 与 Multipart（或显式不支持）；配置页补充直传/分片 CORS 运维说明。

## Impact

- **数据库**：新增 `sys_file` 文件管理表（含 `engine` 等字段）；上传路径简化不引入额外 schema 变更，也不迁移历史数据；配置 seed 同步 `maxSize=200` 与 multipart 相关键。
- **后端 API**：新增文件上传、列表、下载、删除、后缀列表、场景列表等 API 端点；新增 direct-upload init/complete/abort 与 part-url；新增 chunked init/part/complete/abort；公开上传、下载和访问路径保持稳定。
- **后端代码**：`api/file/`、`controller/file/`、`service/file/`；`storagecap` Multipart 可选接口；宿主 `storage.Service` 透传；本地路径生成与统一后端适配；直连/分片会话存储与 complete 校验；`capabilityhost` local storage provider 支持 `files/` 命名空间路由；`httpstartup` DI 为 file 服务注入 storage runtime + local provider，file 自建会话 store。
- **配置**：`sys.upload.maxSize` 默认 200；`multipartThresholdMB` 默认 100；`multipartPartSizeMB` / `multipartMaxConcurrency` 等运行时校验。
- **云存储插件**：`linapro-storage-*` 实现 DirectAccess 与 Multipart（或 `SupportsMultipart=false`）；配置页 i18n 增加 CORS/分片提示。
- **前端**：系统管理“文件管理”子菜单与通用 FileUpload / ImageUpload；`uploadApi` 四策略自动执行；通知公告编辑器与用户头像接入统一上传。
- **动态插件**：WASM storage upload 会话公共 API 不变；底层 commit 走 Provider Multipart 可作为后续优化。
- **兼容性**：历史 `t/<tenantId>/...` 路径与新路径可并存；启用唯一云插件后历史本地对象云未命中回退本地；≤ 默认阈值时行为与整包/单次直传一致。
- **多云约束**：多云同时启用时，文件中心上传/下载/删除与分片路径也会冲突失败。
- **安全与数据权限**：init/complete/chunked 继承文件中心上传权限与租户边界；直链与 part URL 仅短时；不改变列表数据权限模型。
- **破坏性变更**：无对外 API 或数据库层面的破坏性变更；仅新上传物理相对路径格式、内容字节后端选择统一，以及更大文件在配置开启后的自动分片行为变化。
- **非影响项**：列表/检索仍走 `sys_file` 与数据权限；会话一期进程内，无新业务缓存语义；不引入跨 provider 迁移工具；不把云 SDK 引入宿主 go.mod。
