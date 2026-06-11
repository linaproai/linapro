## Requirements

### Requirement: Dynamic Plugins Access Files via Logical Storage Spaces

The system SHALL provide isolated storage service where plugins access files only through authorized logical storage spaces.

#### Scenario: Plugin writes to authorized space
- **WHEN** a plugin calls storage service to write a file
- **THEN** the request must target an authorized logical storage space
- **AND** the host saves to an isolated plugin directory
- **AND** returns file ID, size, and metadata

#### Scenario: Plugin reads authorized storage object
- **WHEN** a plugin reads a file
- **THEN** the host only allows access to authorized logical objects
- **AND** does not expose physical file paths

#### Scenario: Unauthorized path access is rejected
- **WHEN** a plugin attempts path traversal or unauthorized access
- **THEN** the host rejects the call

### Requirement: 插件存储必须通过 storagecap 领域契约提供

系统 SHALL 定义`storagecap.Service`作为源码插件和动态插件共享的插件对象存储领域契约。该契约 MUST 提供`put`、`get`、`delete`、`list`和`stat`能力，并使用领域 DTO 表达 logical path、content type、overwrite、对象大小、etag、更新时间、可见性和列表上限。公共插件调用面 MUST NOT 暴露`pluginbridge/protocol`存储 DTO、宿主物理路径或文件管理模块内部模型。

### Requirement: 插件存储必须按插件和租户作用域生成对象 key

系统 SHALL 在`storagecap.Service`内部将插件可见 logical path 映射为包含插件 ID 和租户维度的 provider object key。不同插件或不同租户使用相同 logical path 时 MUST 不互相覆盖、读取、列出或删除。无租户上下文时 MUST 使用平台级插件存储作用域。宿主 MUST NOT 向插件返回本地绝对路径或 provider 私有 object key。

### Requirement: 动态插件存储分发必须保留路径授权

系统 SHALL 要求动态插件存储分发器在进入`storagecap.Service`之前校验`hostServices`授权快照中的 storage 方法和 path resource。授权 path MUST 匹配插件可见 logical path，而不是 provider object key 或宿主物理路径。

### Requirement: 插件存储必须提供 provider 扩展机制

系统 SHALL 定义`storagecap.Provider`和`storagecap.ProviderFactory`，允许主框架和源码插件提供对象存储后端。Provider MUST 只负责根据 provider object key 执行对象读写、删除、列表和元数据读取，不得接收或解释动态插件`hostServices`授权快照。源码插件可以通过`storagecap.Provide(pluginID, factory)`注册 OSS、MinIO、S3 或其他对象存储 provider。

### Requirement: 插件存储 active provider 必须显式选择

系统 SHALL 对插件存储 provider 使用显式 active provider 策略。未配置 active provider 时 MUST 使用主框架内置本地 provider。配置 active provider plugin ID 时，系统 MUST 仅在该插件启用且 provider 构造成功时使用该 provider；插件未启用、provider 未注册或构造失败时 MUST 返回明确错误，不得静默回退本地 provider。

### Requirement: 本地磁盘 provider 必须明确集群语义

主框架内置本地磁盘 provider SHALL 被视为单机默认实现。若`cluster.enabled=true`且本地 provider 未配置为共享存储路径或未被明确允许用于集群，系统 MUST 提供明确诊断或阻断策略，不得向调用方承诺跨节点对象一致性。

### Requirement: 插件存储列表必须有有界性能契约

系统 SHALL 要求`storagecap.Service.List`具备明确 limit 上限和路径前缀约束。Provider 实现 MUST 不得对插件存储根执行无界全量遍历。超过 limit 的请求 MUST 被截断到上限或返回明确错误。
