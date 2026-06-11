# 插件存储服务规范

## Purpose
待定 - 由归档变更 dynamic-plugin-host-service-extension 创建。归档后更新目的。
## Requirements
### Requirement: 动态插件通过逻辑存储空间访问文件

系统 SHALL 为动态插件提供受隔离的存储服务，插件只能通过宿主授权的逻辑存储空间访问文件，不能直接指定宿主物理路径。

#### Scenario: 插件写入授权存储空间

- **WHEN** 插件调用存储服务写入文件
- **THEN** 请求必须指向该插件已授权的逻辑存储空间或对象引用
- **AND** 宿主按插件隔离目录或对象前缀保存文件
- **AND** 宿主返回文件标识、大小和元数据摘要

#### Scenario: 插件读取授权存储对象

- **WHEN** 插件调用存储服务读取文件
- **THEN** 宿主仅允许读取当前插件被授权访问的逻辑对象
- **AND** 宿主不得向 guest 暴露宿主物理文件路径

#### Scenario: 插件尝试访问未授权路径

- **WHEN** 插件尝试通过路径拼接、目录穿越或未授权`resourceRef`访问文件
- **THEN** 宿主拒绝该调用
- **AND** 宿主不暴露宿主真实文件系统结构

### Requirement: 宿主存储服务实施大小、类型和公开性治理

系统 SHALL 对动态插件的文件读写操作实施大小限制、类型限制和可见性治理。

#### Scenario: 宿主校验文件写入约束

- **WHEN** 插件向某个逻辑存储空间写入或覆盖文件
- **THEN** 宿主根据该空间策略校验最大大小、允许类型和覆盖规则
- **AND** 不符合策略的请求被拒绝

#### Scenario: 插件请求对外暴露文件

- **WHEN** 插件请求生成文件的对外访问地址
- **THEN** 宿主仅对声明为公开或允许签名访问的逻辑存储空间返回可访问地址
- **AND** 私有存储空间不得返回永久公开链接

### Requirement: 插件存储必须通过 storagecap 领域契约提供

系统 SHALL 定义`storagecap.Service`作为源码插件和动态插件共享的插件对象存储领域契约。该契约 MUST 提供`put`、`get`、`delete`、`list`和`stat`能力，并使用领域 DTO 表达 logical path、content type、overwrite、对象大小、etag、更新时间、可见性和列表上限。公共插件调用面 MUST NOT 暴露`pluginbridge/protocol`存储 DTO、宿主物理路径或文件管理模块内部模型。

#### Scenario: 源码插件获取插件存储服务

- **WHEN** 源码插件通过`pluginhost.Services.Storage()`获取存储服务
- **THEN** 宿主返回绑定当前插件 ID 的`storagecap.Service`
- **AND** 该服务不要求源码插件声明`hostServices`资源

#### Scenario: 动态插件获取插件存储服务

- **WHEN** 动态插件通过`guest.Services.Storage()`获取存储服务
- **THEN** guest API 返回实现`storagecap.Service`的客户端
- **AND** 插件业务代码不得依赖`protocol.HostServiceStorageObject`或等价 transport DTO

### Requirement: 插件存储必须按插件和租户作用域生成对象 key

系统 SHALL 在`storagecap.Service`内部将插件可见 logical path 映射为包含插件 ID 和租户维度的 provider object key。不同插件或不同租户使用相同 logical path 时 MUST 不互相覆盖、读取、列出或删除。无租户上下文时 MUST 使用平台级插件存储作用域。宿主 MUST NOT 向插件返回本地绝对路径或 provider 私有 object key。

#### Scenario: 不同插件同名对象隔离

- **WHEN** 插件`plugin-a`写入 logical path `reports/demo.json`
- **AND** 插件`plugin-b`写入 logical path `reports/demo.json`
- **THEN** provider object key 必须不同
- **AND** 两个插件读取同一 logical path 时只能读取各自对象

#### Scenario: 不同租户同名对象隔离

- **WHEN** 同一插件在租户`1001`写入 logical path `reports/demo.json`
- **AND** 同一插件在租户`1002`读取 logical path `reports/demo.json`
- **THEN** 租户`1002`不得读取租户`1001`写入的对象

#### Scenario: 插件读取对象元数据

- **WHEN** 插件调用`storagecap.Service.Stat`读取对象元数据
- **THEN** 响应可以包含 logical path、size、content type、etag、updated at 和 visibility
- **AND** 响应不得包含本地绝对路径、provider credential、bucket 私有配置或宿主文件管理表主键

### Requirement: 动态插件存储分发必须保留路径授权

系统 SHALL 要求动态插件存储分发器在进入`storagecap.Service`之前校验`hostServices`授权快照中的 storage 方法和 path resource。授权 path MUST 匹配插件可见 logical path，而不是 provider object key 或宿主物理路径。

#### Scenario: 动态插件写入授权路径

- **WHEN** 动态插件声明并获得授权访问 storage path `reports/`
- **AND** 插件写入 logical path `reports/demo.json`
- **THEN** WASM 分发器允许请求进入`storagecap.Service`
- **AND** `storagecap.Service`将 logical path 映射为当前插件和租户作用域下的 provider object key

#### Scenario: 动态插件访问未授权路径

- **WHEN** 动态插件只被授权访问 storage path `reports/`
- **AND** 插件尝试读取 logical path `secrets/demo.json`
- **THEN** WASM 分发器在进入`storagecap.Service`之前拒绝调用
- **AND** provider 不得收到该请求

#### Scenario: 动态插件尝试目录穿越

- **WHEN** 动态插件通过`../`、绝对路径、反斜杠混用或等价路径穿越方式调用 storage
- **THEN** 系统拒绝请求
- **AND** 宿主不得暴露真实文件系统结构或 provider object key

### Requirement: 插件存储必须提供 provider 扩展机制

系统 SHALL 定义`storagecap.Provider`和`storagecap.ProviderFactory`，允许主框架和源码插件提供对象存储后端。Provider MUST 只负责根据 provider object key 执行对象读写、删除、列表和元数据读取，不得接收或解释动态插件`hostServices`授权快照。源码插件可以通过`storagecap.Provide(pluginID, factory)`注册 OSS、MinIO、S3 或其他对象存储 provider。

#### Scenario: 主框架注册默认本地 provider

- **WHEN** 宿主启动且没有配置 active storage provider plugin
- **THEN** 系统使用主框架内置本地磁盘 provider
- **AND** 插件存储能力可在单机模式下正常读写对象

#### Scenario: 源码插件注册存储 provider

- **WHEN** 源码插件调用`storagecap.Provide(pluginID, factory)`注册存储 provider
- **THEN** 系统记录该 provider 工厂
- **AND** provider 只有在被配置为 active provider 且插件处于可服务状态时才承接对象存储调用

#### Scenario: Provider 不接收动态授权信息

- **WHEN** 动态插件调用 storage 并通过授权校验
- **THEN** `storagecap.Service`向 provider 传入 provider object key 和对象操作参数
- **AND** provider 不得接收`hostServices`授权快照、授权 path 列表或动态插件原始 envelope

### Requirement: 插件存储 active provider 必须显式选择

系统 SHALL 对插件存储 provider 使用显式 active provider 策略。未配置 active provider 时 MUST 使用主框架内置本地 provider。配置 active provider plugin ID 时，系统 MUST 仅在该插件启用且 provider 构造成功时使用该 provider；插件未启用、provider 未注册或构造失败时 MUST 返回明确错误，不得静默回退本地 provider。

#### Scenario: 未配置 active provider

- **WHEN** 宿主未配置 active storage provider plugin ID
- **THEN** `storagecap.Service`使用主框架内置本地 provider
- **AND** Provider 状态显示本地 provider 为当前有效 provider

#### Scenario: 配置的 provider 插件未启用

- **WHEN** 宿主配置 active storage provider plugin ID 为`linapro-storage-oss`
- **AND** 该插件未启用或不可服务
- **THEN** `storagecap.Service`返回明确 provider 不可用错误
- **AND** 系统不得自动改用本地 provider 执行写入

#### Scenario: 多个 provider 已注册

- **WHEN** 多个源码插件注册 storage provider
- **AND** 宿主只配置其中一个 provider plugin ID 为 active provider
- **THEN** 系统只使用被配置的 provider
- **AND** 未被选中的 provider 不参与对象读写

### Requirement: 本地磁盘 provider 必须明确集群语义

主框架内置本地磁盘 provider SHALL 被视为单机默认实现。若`cluster.enabled=true`且本地 provider 未配置为共享存储路径或未被明确允许用于集群，系统 MUST 提供明确诊断或阻断策略，不得向调用方承诺跨节点对象一致性。

#### Scenario: 单机模式使用本地 provider

- **WHEN** `cluster.enabled=false`且未配置 active storage provider plugin
- **THEN** 插件存储使用主框架本地磁盘 provider
- **AND** 对象读写在当前节点本地磁盘完成

#### Scenario: 集群模式未配置分布式 provider

- **WHEN** `cluster.enabled=true`
- **AND** active provider 仍为未声明共享语义的本地磁盘 provider
- **THEN** 系统必须返回明确诊断或按配置阻断插件存储写入
- **AND** 系统不得让调用方误以为对象已在所有节点可见

### Requirement: 插件存储列表必须有有界性能契约

系统 SHALL 要求`storagecap.Service.List`具备明确 limit 上限和路径前缀约束。Provider 实现 MUST 不得对插件存储根执行无界全量遍历。超过 limit 的请求 MUST 被截断到上限或返回明确错误。

#### Scenario: 插件按前缀列出对象

- **WHEN** 插件调用`List`并提供 logical path prefix 和 limit
- **THEN** 系统只列出当前插件和租户作用域下匹配该 prefix 的对象
- **AND** 返回数量不得超过系统定义的最大 limit

#### Scenario: 插件请求无界列表

- **WHEN** 插件调用`List`但未提供 limit 或提供超过上限的 limit
- **THEN** 系统使用默认 limit 或最大 limit 约束结果
- **AND** provider 不得无界遍历整个插件存储空间

