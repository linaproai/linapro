# cloud-storage-provider-plugins Specification

## Purpose
TBD - created by archiving change cloud-storage-providers. Update Purpose after archive.
## Requirements
### Requirement: 系统必须提供云对象存储 provider 源码插件（厂商 + 协议）

系统 SHALL 交付以下官方 source 插件，并分别通过 `storagecap.Provide(pluginID, factory)` 注册 `storagecap.Provider` 实现：

- `linapro-storage-cos`：腾讯云 COS（厂商 SDK）
- `linapro-storage-oss`：阿里云 OSS（厂商 SDK）
- `linapro-storage-obs`：华为云 OBS（厂商 SDK）
- `linapro-storage-qiniu`：七牛云 Kodo（厂商 SDK；region 可选，留空时 MUST 根据 AccessKey 与 bucket 自动探测；endpoint 可选下载域名）
- `linapro-storage-aws`：AWS S3（厂商插件；配置面向 AWS，region 必填，默认由 SDK 解析官方 endpoint，MUST NOT 将自定义 endpoint / path-style 作为主配置项）
- `linapro-storage-azure`：Azure Blob Storage（厂商插件；配置面向 Azure account/container/shared key，MUST NOT 将 Azure 并入 S3 协议插件）
- `linapro-storage-s3`：S3 兼容协议插件（MinIO、R2、Ceph RGW 等；MUST 要求自定义 endpoint，MUST 支持 path-style；region 可选，缺省时使用签名占位默认值）

上述插件 MUST 为 `type: source`、`scope_nature: platform_only`、平台级全局安装模式，MUST NOT 将云 SDK 依赖引入 `lina-core` 宿主模块路径。

#### Scenario: 通用 S3 协议插件要求 endpoint

- **WHEN** 管理员配置 `linapro-storage-s3`
- **THEN** 设置 MUST 包含 endpoint 与 bucket 等协议字段
- **AND** region 可为可选；留空时运行时 MUST 使用默认签名 region（如 `us-east-1`）
- **AND** MUST NOT 将菜单与展示名绑定为「AWS S3」

#### Scenario: AWS 厂商插件使用 AWS 心智配置

- **WHEN** 管理员配置 `linapro-storage-aws`
- **THEN** 设置 MUST 包含 access key、secret、region、bucket
- **AND** 管理面 MUST NOT 以主表单暴露自定义 endpoint / force path-style（第一期）

#### Scenario: 华为云 OBS 厂商插件配置

- **WHEN** 管理员配置 `linapro-storage-obs`
- **THEN** 设置 MUST 包含 access key、secret、region、bucket
- **AND** endpoint 可选；留空时运行时 MUST 使用华为云默认 OBS endpoint 规则

#### Scenario: 七牛云 Kodo 厂商插件配置

- **WHEN** 管理员配置 `linapro-storage-qiniu`
- **THEN** 设置 MUST 包含 access key、secret、bucket
- **AND** region 可为可选；留空时运行时 MUST 根据 AccessKey 与 bucket 自动探测机房
- **AND** endpoint 可为可选自定义下载域名

#### Scenario: Azure Blob 厂商插件配置

- **WHEN** 管理员配置 `linapro-storage-azure`
- **THEN** 设置 MUST 包含 account name、account key、container
- **AND** endpoint 可选；留空时运行时 MUST 使用 `https://{account}.blob.core.windows.net/`
- **AND** MUST NOT 要求 S3 语义字段（region 作为主必填项、force path-style）

#### Scenario: 仅启用 COS 插件时使用 COS 后端

- **WHEN** `linapro-storage-cos` 已安装且为唯一可服务 storage provider 插件
- **AND** 业务插件调用 `storagecap.Service.Put` 写入对象
- **THEN** 对象内容 MUST 写入 COS 配置中的目标桶
- **AND** 调用方 MUST 仍只使用 logical path，不得感知 bucket 或云侧 object key 拼接细节

#### Scenario: 未启用任何云插件时回退本地

- **WHEN** 全部云 storage provider 插件均未处于可服务状态
- **THEN** `storagecap.Service` MUST 使用宿主内置 local provider
- **AND** 不得因未安装云插件而拒绝 Storage 调用

#### Scenario: 同时启用多个云插件时冲突

- **WHEN** 两个或以上云 storage provider 插件同时可服务
- **THEN** `storagecap.Service` 对象操作 MUST 失败并返回 `CodeStorageProviderConflict`（或等价稳定业务码）
- **AND** MUST NOT 静默选择其中一个 provider
- **AND** MUST NOT 静默回退 local

### Requirement: Provider 必须实现完整对象操作契约

每个云 storage provider 插件 MUST 实现 `storagecap.Provider` 的完整方法集，包括 `Put`、`Get`、`Delete`、`DeleteMany`、`List`、`ListCursor`、`Stat` 和 `BatchStat`。Provider MUST 只接收并解释 scoped object key，MUST NOT 接收或解析动态插件 `hostServices` 授权快照、插件 logical path 或租户/插件隔离规则（隔离由宿主 adapter 负责）。

#### Scenario: 批量元数据与游标列表

- **WHEN** 宿主 adapter 调用 provider 的 `BatchStat` 或 `ListCursor`
- **THEN** 实现 MUST 返回契约约定的 found/missing 与 next cursor 语义
- **AND** MUST NOT 对桶执行无界全量遍历作为生产路径

#### Scenario: 覆盖与存在冲突

- **WHEN** 调用方以 `Overwrite=false` 写入已存在对象 key
- **THEN** provider MUST 映射为 `storagecap` 对象已存在错误（或等价稳定码）
- **AND** MUST NOT 静默覆盖

### Requirement: 云存储插件必须提供管理配置页与 settings API

每个云 storage provider 插件 SHALL 提供管理后台配置页面与受权限保护的 settings API（查询与保存）。配置页 MUST 挂载到宿主 `setting` 目录（`parent_key: setting`）。页面布局 MUST 与授权登录类设置页保持一致：外层卡片、顶部说明 Alert、水平表单、密钥掩码与「已配置可留空保持」语义。

#### Scenario: 保存并掩码回显密钥

- **WHEN** 管理员在配置页填写访问密钥并保存
- **THEN** 系统 MUST 将密钥持久化到平台级 `sys_config`（或宿主 SysConfig 缝）
- **AND** 再次查询 settings 时密钥 MUST 仅以掩码形式返回
- **AND** 响应 MUST 标明密钥已配置

#### Scenario: 空密钥保持原值

- **WHEN** 密钥已配置
- **AND** 管理员保存时提交空密钥或掩码占位
- **THEN** 系统 MUST 保持原密钥不变

#### Scenario: 权限控制

- **WHEN** 用户不具备对应插件的 settings 查看或更新权限
- **THEN** 系统 MUST 拒绝 settings 查询或保存
- **AND** 对应用户 MUST NOT 在导航中看到无权限的配置菜单（或进入后被拒绝）

### Requirement: 配置页必须说明唯一启用与 fail-closed 语义

每个云存储配置页 MUST 向管理员展示明确说明：全局 Storage 后端同时只允许一个 storage provider 插件可服务；多开将导致 Storage 调用冲突失败；插件已启用但配置不完整时 MUST NOT 静默回退本地存储。

#### Scenario: 页面展示运维提示

- **WHEN** 管理员打开任一云存储配置页
- **THEN** 页面 MUST 显示关于唯一启用与配置不完整 fail-closed 的说明信息

### Requirement: 云存储插件必须支持连通性探测

每个云 storage provider 插件 SHALL 提供连通性探测能力（独立 API 或 settings 子动作），在不持久化脏数据的前提下验证当前表单或已存配置能否访问目标桶。探测失败 MUST 返回可读错误，MUST NOT 修改已持久化密钥以外的无关状态。

#### Scenario: 配置正确时探测成功

- **WHEN** 管理员提交该插件合法的必填配置（厂商插件含 region；协议插件含 endpoint）与密钥并触发测试连接
- **THEN** 系统 MUST 对目标桶执行只读探测并返回成功

#### Scenario: 测试连接失败以弹窗展示原因

- **WHEN** 管理员在任一云存储插件配置页点击测试连接且探测失败
- **THEN** 页面 MUST 以弹窗（或等价模态提示）展示完整可读失败原因，便于排查
- **AND** MUST NOT 仅依赖页面顶部常驻 Alert 区域展示该错误详情
- **AND** 成功路径的短 toast 提示 MAY 保留

#### Scenario: 配置错误时探测失败

- **WHEN** 管理员使用错误密钥或错误 bucket 触发测试连接
- **THEN** 系统 MUST 返回失败结果
- **AND** MUST NOT 将探测失败解释为静默成功

### Requirement: 配置不完整时对象操作必须失败且不回退本地

当某一云 storage provider 插件为唯一可服务 provider，但其运行时配置缺失或无效时，系统 SHALL 使 `storagecap.Service` 对象读写返回明确错误，MUST NOT 自动改用内置 local provider。

#### Scenario: 唯一云插件启用但未配置 bucket

- **WHEN** 仅该云插件可服务
- **AND** bucket 或密钥未配置
- **AND** 业务调用 `Storage().Put` 或 `Get`
- **THEN** 调用 MUST 失败并返回可诊断错误
- **AND** 对象 MUST NOT 被写入本地磁盘 provider

### Requirement: 官方云 storage provider 必须实现客户端直连访问

下列官方云 storage provider 插件 MUST 实现客户端直连访问能力（至少覆盖 put 与 get 之一的可生产路径，推荐两者均支持）：`linapro-storage-cos`、`linapro-storage-oss`、`linapro-storage-obs`、`linapro-storage-qiniu`、`linapro-storage-aws`、`linapro-storage-azure`、`linapro-storage-s3`。实现 MUST 输出 `storagecap` 中立 DirectAccess 描述，MUST NOT 要求宿主或业务前端依赖该插件私有 HTTP API 才能完成直传。

#### Scenario: S3 协议插件签发 presigned put

- **WHEN** `linapro-storage-s3` 为唯一可服务 provider 且配置完整
- **AND** 宿主请求对某 scoped key 创建 put 直连访问
- **THEN** 插件 MUST 返回可用的 `presigned_url`（或文档化的等价 mode）描述
- **AND** 使用该描述在过期前 MUST 能将对象写入目标 key

#### Scenario: 配置不完整时直连签发失败且不回退 local

- **WHEN** 某云插件为唯一可服务 provider 但密钥或 bucket 无效
- **AND** 宿主请求创建直连访问
- **THEN** 签发 MUST 失败并返回可诊断错误
- **AND** MUST NOT 静默改为 local 直连或 local 成功语义

### Requirement: 云存储配置页必须说明浏览器直传 CORS 要求

每个官方云 storage provider 配置页 MUST 向管理员展示浏览器直传所需的 CORS（或厂商等价跨域）配置说明，包括至少：允许的方法（如 GET/PUT/POST/HEAD）、需要暴露或允许的 Header 类型、以及需填入工作台 Origin 的提示。说明 MUST 与唯一启用/fail-closed 提示一并可见或可发现。

#### Scenario: 管理员打开配置页可见 CORS 说明

- **WHEN** 管理员打开任一官方云存储配置页
- **THEN** 页面 MUST 展示或可展开直传 CORS 相关运维说明

### Requirement: 支持 Multipart 的官方云 provider 必须实现中立 Multipart 接口

下列官方云 storage provider 插件在可生产配置下 MUST 实现 `storagecap` 可选 Multipart 接口（含 SupportsMultipart、CreateMultipart、UploadPart、CompleteMultipart、AbortMultipart，以及供直传使用的 part 级访问签发能力，若该插件已实现 DirectAccess）：

- `linapro-storage-s3`
- `linapro-storage-aws`
- `linapro-storage-cos`
- `linapro-storage-oss`
- `linapro-storage-obs`

实现 MUST 只接收 scoped object key 与 Multipart 会话参数，MUST NOT 解析插件 hostServices 授权快照。无法在合理成本内对齐的插件（如部分 Azure/七牛路径）MAY 在未实现时使 SupportsMultipart 返回 false，但 MUST NOT 抛出未说明的 panic 或静默写错对象。

#### Scenario: S3 兼容 provider 创建并完成 Multipart

- **WHEN** `linapro-storage-s3` 或 `linapro-storage-aws` 为唯一可服务 provider
- **AND** 调用方 CreateMultipart 后上传至少一个 part 并 Complete
- **THEN** provider MUST 在桶中生成完整对象
- **AND** Stat/Get 可观察到与单次 Put 等价的对象内容

#### Scenario: 未实现 Multipart 的 provider 明确不支持

- **WHEN** 某官方云插件尚未实现 Multipart
- **THEN** SupportsMultipart MUST 为 false
- **AND** 文件中心自动策略 MUST 可回退到 proxy_chunked 或 single，而不是调用未实现方法导致 500

### Requirement: 云存储配置说明必须覆盖分片直传 CORS

每个已实现 part 级直传签发的云 storage 配置页 MUST 向管理员说明：浏览器分片直传需要对桶 CORS 允许相应 HTTP 方法与头部（与单次 PUT 直传一致或更完整），且全局仍只允许一个可服务 storage provider。

#### Scenario: 配置页含分片/CORS 提示

- **WHEN** 管理员打开已支持 Multipart 直传的云存储配置页
- **THEN** 页面 MUST 展示 CORS 或分片直传相关运维说明（双语受 i18n 治理）

