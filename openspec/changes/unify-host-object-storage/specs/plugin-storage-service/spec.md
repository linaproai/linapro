## MODIFIED Requirements

### Requirement: 插件存储必须提供 provider 扩展机制

系统 SHALL 定义 `storagecap.Provider` 和 `storagecap.ProviderFactory`，允许主框架和源码插件提供对象存储后端。Provider MUST 只负责根据 provider object key 执行对象读写、删除、列表和元数据读取，不得接收或解释动态插件 `hostServices` 授权快照。源码插件可以通过 `storagecap.Provide(pluginID, factory)` 注册 OSS、MinIO、S3 或其他对象存储 provider。主框架内置本地 provider MUST 复用宿主内部 `storage.Service` 执行本地对象读写，不得维护另一套独立的本地文件系统实现。

#### Scenario: 主框架注册默认本地 provider

- **WHEN** 宿主启动且没有配置 active storage provider plugin
- **THEN** 系统使用主框架内置本地磁盘 provider
- **AND** 插件存储能力可在单机模式下正常读写对象
- **AND** 内置本地 provider 通过宿主内部 `storage.Service` 执行本地读写

#### Scenario: 源码插件注册存储 provider

- **WHEN** 源码插件调用 `storagecap.Provide(pluginID, factory)` 注册存储 provider
- **THEN** 系统记录该 provider 工厂
- **AND** provider 只有在被配置为 active provider 且插件处于可服务状态时才承接对象存储调用

#### Scenario: Provider 不接收动态授权信息

- **WHEN** 动态插件调用 storage 并通过授权校验
- **THEN** `storagecap.Service` 向 provider 传入 provider object key 和对象操作参数
- **AND** provider 不得接收 `hostServices` 授权快照、授权 path 列表或动态插件原始 envelope

### Requirement: Storage 和 Files 领域边界必须保持独立

系统 SHALL 保持 `Storage()` 和 `Files()` 两个领域能力的职责独立。`Storage()` 拥有插件对象内容生命周期，`Files()` 拥有宿主文件中心资源投影和可见性校验。任一领域的公开契约 MUST NOT 混入另一个领域的内部标识、存储模型或生命周期职责。底层对象存储实现可以复用宿主内部 `storage.Service`，但该复用不得改变两个领域的公开契约和数据边界。

#### Scenario: Storage 不暴露文件中心标识

- **WHEN** 插件调用 `storagecap.Service.Put`、`Get`、`List` 或 `Stat`
- **THEN** 响应只能包含插件 logical path、对象大小、content type、etag、更新时间和可见性等对象存储元数据
- **AND** 响应不得包含 `sys_file.id`、`sys_file.path`、宿主文件 URL、本地绝对路径或宿主文件管理实体
- **AND** 响应不得包含宿主内部 `storage.Service` namespace 或 provider 私有 object key

#### Scenario: Storage 生命周期由插件业务治理

- **WHEN** 插件业务记录被删除、租户插件被禁用或插件卸载时需要清理插件自有对象
- **THEN** 插件或宿主插件生命周期清理逻辑必须通过 `storagecap.Service.Delete` 或有界 `List` 后删除对象
- **AND** 清理逻辑不得直接删除宿主上传目录、provider 根目录或宿主文件中心记录

#### Scenario: Files 不承担插件对象写入

- **WHEN** 插件需要写入、覆盖、删除或列出插件私有对象内容
- **THEN** 插件必须使用 `Storage()` 领域能力
- **AND** 系统不得为该场景新增 `Files()` 上传、对象内容读取、对象删除或对象列表方法
