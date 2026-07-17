# cloud-storage-provider-plugins Specification

## Purpose

定义官方云 storage provider 插件在客户端直连与对象 Multipart 方面的实现与运维说明要求。

## Requirements

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
