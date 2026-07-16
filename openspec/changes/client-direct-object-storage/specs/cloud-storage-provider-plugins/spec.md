## ADDED Requirements

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
