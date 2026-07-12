## MODIFIED Requirements

### Requirement: 插件存储必须提供 provider 扩展机制

系统 SHALL 定义 `storagecap.Provider` 和 `storagecap.ProviderFactory`，允许主框架和源码插件提供对象存储后端。Provider MUST 只负责根据 provider object key 执行对象读写、删除、列表和元数据读取，不得接收或解释动态插件 `hostServices` 授权快照。源码插件可以通过 `storagecap.Provide(pluginID, factory)` 注册 COS、OSS、S3、MinIO 兼容端或其他对象存储 provider。主框架内置本地 provider MUST 复用宿主内部 `storage.Service` 执行本地对象读写，不得维护另一套独立的本地文件系统实现。

#### Scenario: 主框架注册默认本地 provider

- **WHEN** 当前没有可服务的 storage provider 插件
- **THEN** 系统使用主框架内置本地磁盘 provider
- **AND** 插件存储能力可在单机模式下正常读写对象
- **AND** 内置本地 provider 通过宿主内部 `storage.Service` 执行本地读写

#### Scenario: 源码插件注册存储 provider

- **WHEN** 源码插件调用 `storagecap.Provide(pluginID, factory)` 注册存储 provider
- **THEN** 系统记录该 provider 工厂
- **AND** 仅当该插件处于平台可服务状态，且为当前唯一可服务 storage provider 插件时，才承接对象存储调用

#### Scenario: Provider 不接收动态授权信息

- **WHEN** 动态插件调用 storage 并通过授权校验
- **THEN** `storagecap.Service` 向 provider 传入 provider object key 和对象操作参数
- **AND** provider 不得接收 `hostServices` 授权快照、授权 path 列表或动态插件原始 envelope

## REMOVED Requirements

### Requirement: 插件存储 active provider 必须显式选择

**Reason**: 与现行实现及「Storage provider 选择不得依赖主配置项」冲突。运行时以「唯一可服务已注册 provider 插件」自动选中，0 个回退 local，多个冲突拒绝；不再通过宿主主配置 active provider plugin ID 选择后端。

**Migration**: 运维通过插件管理安装/启停恰好一个云 storage provider 插件（如 `linapro-storage-cos`）选择后端；凭据与 bucket 在「存储管理」目录下各插件配置页维护。多插件同时启用将得到 `CodeStorageProviderConflict`。

## ADDED Requirements

### Requirement: 官方云 storage provider 插件必须可交付并接入管理目录

系统 SHALL 提供官方源码插件实现主流云对象存储 provider，并使其管理配置页挂载到宿主 `storage`（存储管理）稳定目录。官方交付范围至少包括腾讯云 COS、阿里云 OSS、AWS S3 厂商插件，以及 S3 兼容协议插件。插件 MUST 通过 `storagecap.Provide` 注册，MUST NOT 改变插件可见 `storagecap.Service` 契约。

#### Scenario: 安装云插件后出现配置入口

- **WHEN** 管理员安装并同步 `linapro-storage-oss`（或 cos / obs / qiniu / aws / azure / s3）
- **THEN** 「存储管理」目录下 MUST 出现对应配置菜单
- **AND** 业务插件调用 `Storage()` 的代码路径 MUST 无需修改

#### Scenario: 唯一云插件启用后承接写入

- **WHEN** 仅一个官方云 storage provider 插件可服务且配置有效
- **AND** 业务插件调用 `Storage().Put`
- **THEN** 对象 MUST 写入该云后端
- **AND** 响应 MUST 仍只暴露 logical path 元数据

### Requirement: 唯一可服务云 provider 配置无效时不得回退本地

当恰好一个 storage provider 插件可服务但其配置缺失或无效时，系统 SHALL 使 Storage 对象操作失败并返回可诊断错误，MUST NOT 静默回退内置 local provider。

#### Scenario: 云插件启用但密钥缺失

- **WHEN** 唯一可服务 provider 为某一云插件
- **AND** 其密钥或 bucket 未配置
- **AND** 调用方执行 `Storage().Put` 或 `Get`
- **THEN** 调用 MUST 失败
- **AND** MUST NOT 将对象写入本地磁盘 provider
