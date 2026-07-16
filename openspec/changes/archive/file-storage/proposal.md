## Why

当前系统需要一套统一、可扩展且便于治理的文件存储能力。早期上传功能分散且实现方式不一致：用户头像通过独立接口上传到本地文件系统，通知公告中的图片则以 Base64 编码直接存储在数据库字段中。Base64 方式对数据库存储压力大，也让附件上传、预览、下载、删除、去重和后续对象存储扩展缺乏统一的能力支点。

随着文件管理模块落地，系统还需要让存储路径、访问语义和租户隔离规则更加清晰。普通文件上传曾使用 `t/<tenantId>/...` 作为相对存储路径，其中 `t` 仅是 tenant 的缩写，对调用方和运维排查不够直观。需要在保留租户分区组织方式的前提下，简化新上传文件的物理路径，同时确保历史文件记录无需迁移即可继续访问。

在此基础上，文件中心（`sys_file`）的内容读写曾固定走宿主本地 `storage.Service`（namespace `files`），与插件 `Storage()` / `storagecap` 的「0 云→local、1 云→云、≥2→冲突」后端选择分离。产品期望：**所有字节级文件存储（上传/下载/删除）统一对象后端**；无云插件时默认本地。列表/检索仍走 `sys_file` 元数据。

对象存储虽已通过 `storagecap.Provider` 与统一后端选择实现了服务端中转的通用云接入，但浏览器、插件前端与外部客户端上传/下载大文件时，全部字节仍必须经过 `lina-core`。这会放大宿主带宽与 CPU 成本，拉长上传时延，并限制大文件体验。需要在不破坏现有领域边界与唯一 active provider 语义的前提下，交付与服务端中转对称的**客户端直连对象访问**能力：宿主签发短时访问能力、客户端直连云、完成后由宿主确认并落领域元数据。

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
- **Provider 插件补齐**：官方云插件实现 DirectAccess；local 明确不支持，调用方自动降级 multipart/stream 中转。
- **前端通用上传双模式**：`FileUpload` / `ImageUpload` 在支持直传时走 init→直连→complete，否则静默回退现有 multipart；业务页面无厂商分支。
- **运维与安全**：配置页补充直传所需 CORS/桶策略说明；多云冲突、配置无效 fail-closed 语义与现网一致；会话过期与 complete 校验防脏元数据。

**非破坏性**：现有 multipart 上传、宿主中转下载、插件 `Storage().Put/Get` 流式路径全部保留，作为 local 与能力缺失时的默认路径。

## Capabilities

### New Capabilities

- `file-storage`: 通用文件存储抽象层，支持本地存储（默认）和对象存储扩展，基于租户与 SHA-256 散列值实现文件去重与物理文件复用。
- `file-management`: 文件管理业务模块，提供文件列表、上传、下载、删除、预览、详情、按类型和使用场景筛选等功能。
- `file-upload-storage-path`: 统一普通文件上传的租户分区存储路径规则、历史路径兼容策略和验证要求。
- `unified-object-backend`: 全站对象字节后端选择与文件中心内容接入；与插件 `Storage()` 共享 `ResolveProvider` 规则。
- `client-direct-object-access`: 客户端直连对象访问的通用契约与生命周期（能力探测、访问模式、init/complete/abort、安全 scope、local 降级、与领域元数据确认的边界）。

### Modified Capabilities

- `user-avatar`: 用户头像上传改为调用通用文件上传接口，使用 `avatar` 场景标识并复用统一文件管理能力。
- `notice-editor`: 通知公告富文本编辑器图片上传改为调用通用文件上传接口，附件上传纳入统一文件存储与管理流程。
- `plugin-storage-service`: 文件中心内容与插件 Storage 共享 Resolve 规则；`storagecap.Service` / `Provider` 增加直连访问签发与确认。
- `cloud-storage-provider-plugins`: 各官方云 provider 实现 DirectAccess；配置页补充直传 CORS 运维说明。

## Impact

- **数据库**：新增 `sys_file` 文件管理表（含 `engine` 等字段）；上传路径简化不引入额外 schema 变更，也不迁移历史数据。
- **后端 API**：新增文件上传、列表、下载、删除、后缀列表、场景列表等 API 端点；新增 direct-upload init/complete/abort；公开上传、下载和访问路径保持稳定。
- **后端代码**：新增 `api/file/`、`controller/file/`、`service/file/` 等模块；调整本地存储路径生成逻辑；文件中心内容层依赖 `ResolveProvider` 适配；直连会话存储与 complete 校验；`capabilityhost` local storage provider 支持 `files/` 命名空间路由；`httpstartup` DI 顺序为 file 服务注入 storage runtime + local provider。
- **云存储插件**：七个 `linapro-storage-*` 实现 DirectAccess；README Non-goals 调整；配置页 i18n 增加 CORS 提示。
- **前端**：系统管理菜单下新增“文件管理”子菜单，新增通用 FileUpload / ImageUpload 组件，并改造通知公告编辑器与用户头像上传流程；通用上传 hook 双模式（direct 优先、proxy/multipart 回退）；可选下载直链消费。
- **兼容性**：历史 `t/<tenantId>/...` 路径与新路径可并存访问，访问流程始终以数据库中的 `sys_file.path` 为权威路径；启用唯一云插件后，历史本地对象在云未命中时回退本地读取。
- **多云约束**：多云同时启用时，文件中心上传/下载/删除也会冲突失败（与插件 Storage 一致）。
- **安全与数据权限**：init/complete 继承文件中心权限与租户边界；直链仅短时；不改变列表数据权限模型。
- **破坏性变更**：无对外 API 或数据库层面的破坏性变更；仅新上传文件的物理相对路径格式调整，以及内容字节后端选择统一。
- **非影响项**：列表/检索仍走 `sys_file` 与数据权限；路径规则与统一后端选择不引入额外缓存一致性风险；不引入跨 provider 迁移工具；不把云 SDK 引入宿主 go.mod。
