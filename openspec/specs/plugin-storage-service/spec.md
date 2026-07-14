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

### Requirement: 插件私有文件对象必须归属 Storage 领域

系统 SHALL 将插件自有附件、插件业务二进制对象、导入导出临时对象和插件卸载清理对象归属`Storage()`领域能力。源码插件和动态插件 MUST 通过`storagecap.Service`写入、读取、删除、列出和读取对象元数据，不得通过宿主文件中心`Files()`领域或宿主本地物理路径管理这些插件私有对象。

#### Scenario: 源码插件管理私有附件

- **WHEN** 源码插件在业务记录中保存插件自有附件
- **THEN** 插件业务服务必须通过`pluginhost.Services.Storage()`注入的`storagecap.Service`写入和读取附件对象
- **AND** 插件业务记录只能保存插件 logical path 和业务所需展示元数据
- **AND** 插件不得直接读取`upload.path`、拼接宿主本地绝对路径或依赖宿主文件管理模块内部模型管理附件内容

#### Scenario: 动态插件管理私有对象

- **WHEN** 动态插件需要保存插件自有附件或导出对象
- **THEN** 插件必须在`plugin.yaml hostServices`中声明`service: storage`、所需方法和`resources.paths`
- **AND** 插件业务代码必须通过`pluginbridge.Services.Storage()`或等价入口消费`storagecap.Service`
- **AND** 运行时授权 path 必须匹配插件 logical path，而不是宿主物理路径、provider object key 或宿主文件 ID

#### Scenario: 插件私有对象不进入宿主文件中心

- **WHEN** 插件通过`storagecap.Service.Put`写入对象
- **THEN** 系统不得因为该写入自动创建`sys_file`记录
- **AND** 该对象不得默认进入宿主文件管理列表、宿主文件下载 API 或`Files()`批量投影结果
- **AND** 插件如需在自身页面展示对象名称、大小或类型，必须保存业务元数据或读取`storagecap.Service.Stat`结果

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

### Requirement: 动态插件 Storage Put 必须支持有界内存分片上传

系统 SHALL 允许动态插件通过`Storage().Put`写入大文件或未知大小输入，并在 guest SDK 内部按输入大小选择单次`storage.put`或分片上传。分片上传 MUST 使用`put.init`、`put.chunk`、`put.commit`和`put.abort`host service 方法完成传输；宿主 MUST 在 commit 时把已接收临时对象作为`io.Reader`交给`storagecap.Service.Put`写入最终 logical path。系统 MUST NOT 对最终对象大小设置动态 host service 固定上限；单个 chunk payload 可以有固定上限以保护 host call 内存边界。

#### Scenario: 小文件通过单次 host call 写入

- **WHEN** 动态插件调用`Storage().Put`且 guest SDK 能确认输入大小不超过单次上传阈值
- **THEN** guest SDK 可以继续使用`storage.put`一次性提交请求
- **AND** 宿主必须通过`storagecap.Service.Put`写入最终 logical path

#### Scenario: 大文件通过分片写入

- **WHEN** 动态插件调用`Storage().Put`且输入大小超过单次上传阈值
- **THEN** guest SDK 必须先调用`put.init`创建上传会话
- **AND** 按顺序调用`put.chunk`传输分片
- **AND** 调用`put.commit`提交总大小并写入最终 logical path
- **AND** 宿主在分片阶段不得把完整对象常驻内存

#### Scenario: 未知大小 reader 通过分片写入

- **WHEN** 动态插件调用`Storage().Put`且输入 reader 的总大小未知
- **THEN** guest SDK 必须使用分片上传路径
- **AND** 不得为了判断大小而先把完整 reader 读入内存

#### Scenario: 分片上传失败后清理会话

- **WHEN** 动态插件分片上传在任一 chunk 或 commit 阶段失败
- **THEN** guest SDK 必须尽力调用`put.abort`
- **AND** 宿主必须删除对应临时上传文件并释放会话状态

### Requirement: 动态插件 Storage 分片上传必须保持路径授权和会话绑定

系统 SHALL 对`put.init`、`put.chunk`、`put.commit`和`put.abort`执行与`storage.put`等价的 service、method 和`storage.resources.paths`授权校验。授权 path MUST 匹配最终插件 logical path，而不是 upload ID、provider object key、宿主物理路径或文件中心 ID。宿主 MUST 将 upload ID 绑定到当前插件 ID、最终 logical path 和上传会话状态，并在后续 chunk、commit 和 abort 中校验一致性。

#### Scenario: 插件声明分片上传方法

- **WHEN** 动态插件需要通过`Storage().Put`写入大文件或未知大小 reader
- **THEN** 插件必须在`plugin.yaml hostServices`的`service: storage`声明`put.init`、`put.chunk`、`put.commit`和`put.abort`
- **AND** 宿主授权快照必须包含这些方法后，guest SDK 才能完成分片上传流程

#### Scenario: 分片上传访问授权路径

- **WHEN** 动态插件已获授权访问 storage path `reports/`
- **AND** 插件通过分片上传写入 logical path `reports/demo.bin`
- **THEN** `put.init`、`put.chunk`、`put.commit`和`put.abort`都必须按`reports/demo.bin`执行路径授权
- **AND** 授权通过后才能访问上传会话或写入最终对象

#### Scenario: 分片上传访问未授权路径

- **WHEN** 动态插件只获授权访问 storage path `reports/`
- **AND** 插件尝试通过分片上传写入 logical path `secrets/demo.bin`
- **THEN** 宿主必须在进入上传会话或`storagecap.Service.Put`之前拒绝调用
- **AND** provider 不得收到该写入请求

#### Scenario: Upload ID 被挪用到其他路径

- **WHEN** 动态插件使用为`reports/a.bin`创建的 upload ID
- **AND** 后续 chunk、commit 或 abort 请求声明 logical path `reports/b.bin`
- **THEN** 宿主必须拒绝该请求
- **AND** 不得把临时对象提交到任一最终 logical path

#### Scenario: 分片 offset 不连续

- **WHEN** 动态插件提交的 chunk offset 不等于会话当前累计大小
- **THEN** 宿主必须拒绝该 chunk
- **AND** 不得把乱序或重叠内容写入临时对象

### Requirement: Storage 必须支持插件私有对象批量元数据
系统 SHALL 提供`Storage.BatchStat`和动态`storage.stat.batch`，按路径集合批量返回当前插件和租户作用域内的对象元数据。请求 MUST 限制路径数量、单路径长度和总字节数，响应 MUST 使用领域元数据 DTO。

#### Scenario: 批量读取对象元数据
- **WHEN** 插件请求多个 storage 路径的元数据
- **THEN** 系统一次性返回可见路径的元数据和不透明缺失集合
- **AND** 不得逐路径调用单对象 stat 作为常规实现

### Requirement: Storage 必须支持有界游标列表
系统 SHALL 提供`Storage.ListCursor`和动态`storage.list.cursor`，按前缀、cursor 和 limit 返回插件私有对象列表。实现 MUST 不提供无界前缀遍历。

#### Scenario: 游标读取下一页对象
- **WHEN** 插件携带 prefix、cursor 和 limit 请求 storage 列表
- **THEN** 系统返回当前页对象元数据和 next cursor
- **AND** 每页数量不得超过领域上限

### Requirement: Storage 必须支持批量删除明确路径集合
系统 SHALL 提供`Storage.DeleteMany`和动态`storage.delete.batch`，只删除当前插件和租户作用域下明确路径集合。删除缺失对象 SHOULD 作为 no-op 成功处理，除非 provider 返回不可恢复错误。

#### Scenario: 批量删除私有对象
- **WHEN** 插件提交多个明确 storage 路径删除
- **THEN** 系统在资源授权和插件/租户作用域内删除这些路径
- **AND** 不接受宿主物理路径或无边界前缀删除

### Requirement: Storage provider 选择不得依赖主配置项
系统 SHALL 不通过主框架配置文件选择`Storage`provider。宿主 MUST 在当前已注册 storage provider 插件中选择唯一一个平台可服务插件；没有可服务 provider 插件时 MUST 使用内置本地文件 provider。

#### Scenario: 未安装或未启用 storage provider 插件
- **WHEN** 插件调用`Storage`
- **AND** 当前没有可服务的 storage provider 插件
- **THEN** 系统使用内置本地文件 provider 处理请求

#### Scenario: 唯一 provider 插件可服务
- **WHEN** 插件调用`Storage`
- **AND** 当前恰好一个已注册 storage provider 插件处于平台可服务状态
- **THEN** 系统使用该 provider 插件处理请求

#### Scenario: 多个 provider 插件同时可服务
- **WHEN** 插件调用`Storage`
- **AND** 当前多个已注册 storage provider 插件处于平台可服务状态
- **THEN** 系统拒绝本次 storage 请求并返回`CodeStorageProviderConflict`
- **AND** 不得静默选择任意一个 provider 或回退到本地文件 provider

### Requirement: 官方云 storage provider 插件必须可交付并接入管理目录

系统 SHALL 提供官方源码插件实现主流云对象存储 provider，并使其管理配置页挂载到宿主 `setting`（系统设置）稳定目录。官方交付范围至少包括腾讯云 COS、阿里云 OSS、AWS S3 厂商插件，以及 S3 兼容协议插件。插件 MUST 通过 `storagecap.Provide` 注册，MUST NOT 改变插件可见 `storagecap.Service` 契约。

#### Scenario: 安装云插件后出现配置入口

- **WHEN** 管理员安装并同步 `linapro-storage-oss`（或 cos / obs / qiniu / aws / azure / s3）
- **THEN** 「系统设置」目录下 MUST 出现对应配置菜单
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

