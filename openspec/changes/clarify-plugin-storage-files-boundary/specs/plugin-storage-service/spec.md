## ADDED Requirements

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

系统 SHALL 保持`Storage()`和`Files()`两个领域能力的职责独立。`Storage()`拥有插件对象内容生命周期，`Files()`拥有宿主文件中心资源投影和可见性校验。任一领域的公开契约 MUST NOT 混入另一个领域的内部标识、存储模型或生命周期职责。

#### Scenario: Storage 不暴露文件中心标识

- **WHEN** 插件调用`storagecap.Service.Put`、`Get`、`List`或`Stat`
- **THEN** 响应只能包含插件 logical path、对象大小、content type、etag、更新时间和可见性等对象存储元数据
- **AND** 响应不得包含`sys_file.id`、`sys_file.path`、宿主文件 URL、本地绝对路径或宿主文件管理实体

#### Scenario: Storage 生命周期由插件业务治理

- **WHEN** 插件业务记录被删除、租户插件被禁用或插件卸载时需要清理插件自有对象
- **THEN** 插件或宿主插件生命周期清理逻辑必须通过`storagecap.Service.Delete`或有界`List`后删除对象
- **AND** 清理逻辑不得直接删除宿主上传目录、provider 根目录或宿主文件中心记录

#### Scenario: Files 不承担插件对象写入

- **WHEN** 插件需要写入、覆盖、删除或列出插件私有对象内容
- **THEN** 插件必须使用`Storage()`领域能力
- **AND** 系统不得为该场景新增`Files()`上传、对象内容读取、对象删除或对象列表方法
