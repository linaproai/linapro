## Context

`Storage()`和`Files()`已经同时存在于`capability.Services`和`pluginbridge.Services`中。现有稳定规范也已经要求源码插件和动态插件共享`storagecap.Service`，并要求`files`动态领域 service 名与`Files()`领域目录一致。

当前主要问题不是缺少能力，而是边界表达不够清晰：

- `Storage()`在动态插件侧通过`service: storage`声明，容易被误解为宿主文件中心或上传下载模块。
- `Files()`命名天然接近“文件读写”，但实际只提供宿主文件中心投影读取和可见性校验。
- `linapro-demo-source`源码插件示例仍通过`upload.path`和本地路径管理附件，未体现源码插件也应通过`pluginhost.Services.Storage()`消费统一对象存储领域能力。

## Goals / Non-Goals

**Goals:**

- 将`Storage()`明确为插件私有对象存储领域能力，源码插件和动态插件使用同一`storagecap.Service`语义。
- 将`Files()`明确为宿主文件中心资源投影领域能力，负责`sys_file`可见性、批量投影和存在性不泄露校验。
- 迁移`linapro-demo-source`示例，使源码插件附件与动态插件附件使用相同对象存储领域模型。
- 更新插件公开契约文档和两个 demo README，使开发者能按场景选择`Storage()`或`Files()`。

**Non-Goals:**

- 不重命名`Storage()`、`Files()`、`storagecap`、`filecap`或动态`hostServices`协议名。
- 不让`Storage()`生成`sys_file`记录，也不让插件对象进入宿主文件管理列表。
- 不让`Files()`新增上传、下载、删除、列表插件对象等对象存储方法。
- 不改宿主文件管理模块的数据模型、HTTP API 或现有上传路径兼容规则。
- 不迁移已有历史附件对象或数据库记录；源码 demo 的历史本地附件兼容如需保留，只在实现中做受控读取或清理策略。

## Decisions

### 1. `Storage()`是插件私有对象存储领域，不是文件中心别名

`Storage()`负责插件自有对象内容的`put/get/delete/list/stat`。调用方看到的是插件 logical path，宿主内部映射到`pluginID + tenant`作用域下的 provider object key。该能力不返回本地绝对路径、provider 私有 key、宿主文件管理表主键或`sys_file`实体。

备选方案是把插件附件全部纳入宿主文件管理模块。该方案会让插件私有对象进入宿主文件中心治理范围，扩大`sys_file`语义，也会让插件卸载清理、provider 切换和动态插件路径授权变复杂，因此不采用。

### 2. `Files()`只代表宿主文件中心投影与可见性校验

`Files()`继续通过`filecap.Service`提供`BatchGet`和`EnsureVisible`等能力。它用于插件引用已有宿主文件资源时读取名称、大小、MIME、业务场景等投影，并按租户和数据权限保持不可见资源不泄露。

备选方案是让`Files()`也承担上传下载。该方案会把两个不同生命周期的对象混在一个接口里：宿主文件中心资源由`sys_file`和文件管理数据权限治理，插件私有对象由插件 ID、租户和 storage provider 治理；混合后调用方难以判断数据边界，因此不采用。

### 3. 源码插件示例应改为显式注入`storagecap.Service`

`linapro-demo-source`的业务 service 构造函数应接收`storagecap.Service`作为独立依赖，并只保存所需的最窄接口。`backend/plugin.go`在`registerRoutes`中从`registrar.Services().Storage()`获取插件作用域存储服务，与现有`I18n()`和`TenantFilter()`一样显式校验并注入。

源码插件仍是可信插件形态，不需要在`plugin.yaml hostServices`中声明`storage`资源；但`storagecap.Service`自身仍必须按插件 ID 和租户上下文隔离对象 key。

### 4. 动态插件文档用两个并列示例表达边界

动态插件文档应同时展示：

- `service: storage`用于插件私有对象存储，带`resources.paths`。
- `service: files`用于宿主文件中心资源投影，不带路径资源。

这样避免开发者把`Files()`理解成对象内容读写，也避免把`Storage()`理解成宿主文件中心。

## Risks / Trade-offs

- 源码示例迁移可能影响已有本地附件下载演示 → 通过源码插件单元测试或窄包编译测试覆盖创建、更新、下载、删除和卸载清理路径；如需历史兼容，在任务记录中说明读取和清理策略。
- `Storage()`不进入`sys_file`可能让插件私有附件无法在宿主文件管理页面统一查看 → 这是有意边界；插件私有对象由插件业务页面和插件卸载策略治理。
- 文档边界过于抽象仍可能被误用 → README 使用场景表和动态`hostServices`样例双重说明，减少判断成本。
- 本次变更涉及文档、源码插件和插件能力边界 → 需要运行`openspec validate`、相关 Go 测试、文档同步检查，并在审查中记录`i18n`、数据权限、缓存和跨平台影响。

## Migration Plan

1. 补充 OpenSpec 增量规范，固化`Storage()`与`Files()`职责边界。
2. 迁移`linapro-demo-source`附件服务：通过`storagecap.Service`写入、读取、删除和清理插件私有附件对象。
3. 更新`linapro-demo-source`和`linapro-demo-dynamic` README，保持中英文镜像一致。
4. 更新`apps/lina-core/pkg/plugin`README，补充公开契约边界说明。
5. 运行源码插件后端相关 Go 测试、OpenSpec 严格校验和格式检查。

回滚策略：恢复源码示例的旧本地路径 helper 与 README 文档；由于不改数据库模型、公开 API 方法签名或动态 wire 协议，不需要数据迁移回滚。

## Open Questions

无。若实现时发现历史本地附件需要长期兼容，应在任务记录中说明兼容读取边界；不应因此改变`Storage()`和`Files()`的领域职责。
