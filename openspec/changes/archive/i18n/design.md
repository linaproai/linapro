# Design

## Context and Goals

LinaPro需要把国际化能力沉淀为宿主级基础设施，而不是让默认工作台、业务模块和插件各自维护语言解析、资源装载和文案治理。设计目标是同时解决四类问题：一是让宿主、插件和前端在统一约定下发现、装载和分发翻译资源；二是让动态元数据、错误消息和导入导出内容按当前请求语言稳定投影；三是降低运行时翻译热路径、缓存和前端语言切换的性能成本；四是把默认交付范围收敛为真正需要长期维护的双语基线，同时保留项目自行扩展更多语言的机制。

本设计不再保留历史双轨兼容层。最终形态以“默认交付仅维护`zh-CN`与`en-US`，其他语言通过资源目录和默认配置扩展”为准；繁体中文不再作为框架默认资源与验收目标，但浏览器中文语言标签仍统一回退到默认中文。

## Resource Model and Locale Governance

### Three-Layer I18n Model

国际化文本按来源与生命周期拆分为三层：

- **静态界面文案**：由默认管理工作台和共享前端包的本地`JSON`资源维护。
- **动态元数据**：菜单、字典、系统参数、系统信息、角色、任务、插件元数据等由后端按请求语言投影。
- **业务内容**：确实需要多语言正文的业务模块，在模块自身边界内维护多语言内容模型。

这种拆分避免前端用稳定业务键或数据库值反向翻译宿主治理数据，也避免宿主基础服务越界理解业务实体。

### File-Only Source of Truth

翻译资源统一以文件为权威来源。宿主、源码插件和动态插件均通过`manifest/i18n/<locale>/`维护运行时语言资源，通过`manifest/i18n/<locale>/apidoc/**/*.json`维护`API`文档翻译资源。运行时不再依赖`sys_i18n_locale`、`sys_i18n_message`、`sys_i18n_content`等翻译持久化表，避免数据库覆盖与文件资源并存造成审计、缺失检查、回写和缓存失效复杂化。

资源组织按语义域拆分，例如：

```text
manifest/i18n/
  en-US/
    framework.json
    menu.json
    dict.json
    config.json
    error.json
    artifact.json
    public-frontend.json
    apidoc/
      common.json
      core-api-user.json
```

运行时资源既允许层级`JSON`，也允许点分键格式；宿主加载后统一展平成扁平键治理，对前端接口再输出嵌套对象结构。

### Built-in Locale Discovery and Default Scope

宿主从`manifest/i18n/<locale>/*.json`自动发现可用语言，再通过默认配置中的`i18n`段维护默认语言、多语言开关、排序和原生名称。默认交付收敛为双语：

```yaml
i18n:
  default: zh-CN
  enabled: true
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文
```

设计约束如下：

- 新增内置语言只能通过补充宿主、插件、`apidoc`与前端资源，以及可选的默认配置元数据完成。
- 禁止为了注册语言而新增后端`Go`枚举、`SQL seed`或前端硬编码语言清单。
- 默认交付不再保留`zh-TW`运行时资源、静态语言包或专项翻译基线。
- 当`i18n.enabled=false`时，宿主只接受默认语言，工作台隐藏语言切换器，仅加载默认语言资源。
- 浏览器语言只作为首次访问默认值参考，所有中文语言标签（包括`zh-TW`、`zh-Hans-CN`）统一映射到`zh-CN`。

### Translation Key Conventions

动态元数据翻译键由稳定业务锚点推导，避免建立额外映射表。主要规则包括：

- 菜单：`menu.<menu_key>.title`
- 字典类型：`dict.<dict_type>.name`
- 字典项：`dict.<dict_type>.<value>.label`
- 系统参数：`config.<config_key>.name`、`config.<config_key>.remark`
- 参数字段表头：`config.field.<name>`
- 插件：`plugin.<plugin_id>.name`、`plugin.<plugin_id>.description`
- 系统信息：`systemInfo.component.<section>.<name>.description`
- 公共前端配置：`publicFrontend.<group>.<field>`
- 错误消息：`error.<domain>.<case>`
- 内置角色：`role.builtin.<key>.name`

## Runtime Distribution and Performance

### Locale Resolution

宿主通过`LocaleResolver`在请求入口统一解析语言，优先级为：

1. 查询参数`lang`
2. 请求头`Accept-Language`
3. 默认配置中的`i18n.default`

解析结果写入统一业务上下文，供控制器、服务、插件桥接、导入导出和运行时翻译包聚合共用。语言不可用时总是回退到默认语言，不隐式混入其他语言内容。

### Runtime Bundle API and ETag Negotiation

宿主提供运行时翻译包与语言列表接口：

- 翻译包接口按语言返回宿主、源码插件和启用中的动态插件聚合后的消息对象。
- 语言列表接口返回多语言开关、默认语言标记、显示名、原生名和固定`ltr`方向。
- 翻译包响应头输出`ETag: "<locale>-<bundleVersion>"`与`Cache-Control: private, must-revalidate`。
- 当前端携带匹配的`If-None-Match`时，接口返回`304 Not Modified`且不携带消息体。

任何与翻译资源相关的缓存失效都会推动`bundleVersion`递增，确保同一语言的不同内容拥有不同的版本标识。

### Layered Cache and Scoped Invalidation

`runtimeBundleCache`按“语言 × 扇区”重构为分层缓存：

- `host`：宿主资源，启动后加载。
- `plugins`：源码插件资源，在源码插件注册变更时刷新。
- `dynamic`：动态插件资源，在安装、启用、禁用、升级、卸载时刷新。
- `merged`：按优先级聚合后的视图，对前端输出使用。

失效必须显式传入作用域，例如按语言、扇区、插件`ID`清理，而不是无差别清空所有语言和所有扇区。这样既满足单机模式下的细粒度缓存控制，也为分布式环境下的跨实例失效广播保留明确作用域边界。

### Zero-Copy Translation Hot Path

`Translate`、`TranslateSourceText`、`TranslateOrKey`和`TranslateWithDefaultLocale`在缓存命中时仅持有读锁并直接返回单值结果，不再克隆整包消息。只有`BuildRuntimeMessages`和`ExportMessages`这类需要向调用方返回消息集合的方法才保留克隆语义。这样可以消除高频路径上无意义的内存复制，稳定翻译调用延迟。

### Frontend Persistent Cache and Load Flow

前端`runtime-i18n.ts`改为使用`requestClient`请求运行时翻译包，并将`{ locale, etag, messages, savedAt }`持久化到`localStorage`，默认`TTL`为`7`天。页面加载或语言切换时优先使用持久化缓存立即渲染，再后台发送带`If-None-Match`的请求校验是否需要更新。

`loadMessages`按失败语义拆分：

1. 运行时翻译包失败时，优先使用持久缓存或降级提示。
2. 公共前端配置失败时，不阻塞主流程。
3. 第三方组件语言包必须等待成功加载后再继续初始化。

## Service Boundaries and Shared Components

### Interface Split

原有`i18n.Service`职责过重，拆分为四类小接口：

- `LocaleResolver`：解析请求语言和上下文语言。
- `Translator`：负责翻译查询与错误本地化。
- `BundleProvider`：负责运行时翻译包、语言列表和版本输出。
- `Maintainer`：负责导出、缺失检查、资源诊断与缓存失效。

`Service`保留为四类接口的组合类型，业务模块按最小依赖声明字段类型，降低测试替身和模块耦合成本。

### ResourceLoader Reuse

在`pkg/i18nresource/`中提供共享`ResourceLoader`，通过`Subdir`、`LocaleSubdir`、`PluginScope`、`LayoutMode`和`ValueMode`等参数装配宿主运行时资源与`apidoc`资源加载流程。运行时翻译包和`API`文档不再各自维护重复的插件遍历和资源合并实现。

### WASM Section Parsing Ownership

动态插件`WASM`自定义`section`读取统一由`pkg/pluginbridge`提供`ReadCustomSection`与`ListCustomSections`能力，`i18n`、`apidoc`和插件运行时都通过这一能力读取动态插件产物中的资源片段，避免在`i18n`包内重复维护`WASM`解析逻辑。

### Business Projection Ownership

业务模块自行维护本地化投影规则，`internal/service/i18n`只提供语言解析、翻译查找、资源加载、缓存和缺失检查等基础能力。菜单、字典、系统参数、任务、角色、插件等模块各自在自身`*_i18n.go`或等价文件中决定：

- 如何推导翻译键。
- 哪些字段属于宿主治理元数据，应该按语言投影。
- 哪些字段属于用户编辑数据，应保留数据库原值。
- 哪些记录属于内置受保护对象，允许读时本地化但禁止删除。

### Source Namespace Registration

源码拥有的翻译命名空间通过`RegisterSourceTextNamespace(prefix, reason string)`显式注册。业务模块在各自`init()`中声明自己的前缀，缺失翻译检查和资源诊断再基于注册表识别“由代码拥有的翻译键”，而不是在`i18n`基础服务中硬编码业务模块名单。

## Message Governance and Localization Rules

### Message Classification

运行时消息被划分为六类：

- `UserMessage`：用户可见错误、校验失败、前端提示。
- `UserArtifact`：导入导出文件、模板、表头、失败原因、枚举文本。
- `UserProjection`：宿主治理元数据和只读展示投影。
- `DeveloperDiagnostic`：插件协议、`WASM`宿主调用、清单校验、`CLI`诊断等开发者可读文本。
- `OpsLog`：运维日志与指标，使用稳定英文与结构化字段。
- `UserData`：用户输入或外部系统数据，默认保持原值。

### Structured Error Model

宿主通过`apps/lina-core/pkg/bizerr`统一构造业务错误。业务模块在各自`*_code.go`中集中定义稳定错误码和英文兜底文本，请求响应中统一输出：

- `code`：GoFrame类型错误码。
- `message`：按当前请求语言解析后的显示文案。
- `errorCode`：稳定业务语义标识。
- `messageKey`与`messageParams`：供前端和插件复用的本地化锚点。

这样前端优先渲染`messageKey`，无法渲染时再使用后端已本地化的`message`，避免再次依赖中文源文本做判断。

### Import/Export and Plugin Contract

导入导出流程在请求级别解析语言并复用翻译结果，由业务服务传入本地化后的工作表名、列标题、枚举文本和失败原因。插件桥接、宿主服务和动态插件返回的错误契约保留结构化字段，默认使用稳定英文开发者诊断文本，并在进入工作台时根据`messageKey`按当前语言展示。

### Hardcoded Content Cleanup and Scanning

通过静态扫描和代码审查治理后端、插件、前端中可返回给调用方的硬编码中文。扫描重点覆盖`gerror`构造、`panic`边界、导入导出表头数组、前端表单与表格标题、弹窗提示和插件错误构造，允许名单仅用于注释、测试数据、用户示例和运维类英文文本。

## Default Workspace Experience

### Startup Language and Static Bundles

默认管理工作台在无已保存偏好且未显式传入初始化语言时，按浏览器首选语言决定首次启动语言。所有中文类标签映射到`zh-CN`，其他语言默认使用`en-US`。默认工作台只交付`zh-CN`与`en-US`静态语言包，不再保留`zh-TW`默认静态资源。

### Fixed LTR Direction

工作台文档方向固定为`ltr`。语言切换时同时更新`<html dir>`与`Ant Design Vue ConfigProvider`的`direction="ltr"`。设计上不维护静态`RTL`语言分支，避免新增语言时再补充方向逻辑。

### Runtime Refresh on Language Switch

语言切换不仅刷新静态语言包，还要同步刷新以下内容：

- 运行时翻译包。
- 公共前端配置中的品牌与登录页文案。
- 动态菜单、路由标题和相关工作台上下文。
- 宿主嵌入式插件页面的语言上下文。

### English Sweep and Layout Adaptation

针对默认交付页面执行英文回归，确保标题、按钮、标签、表头、空状态、系统生成节点、内置记录显示和确认弹窗不残留中文。对英文较长的表单标签、表格列头、抽屉标题和按钮区域做布局适配，保证在常见桌面分辨率下仍可读、可操作。

### Built-in Protection and Workbench Governance

内置字典、系统参数和类似治理对象允许编辑但禁止删除；后端返回结构化业务错误，前端同步隐藏或阻止删除入口。动态插件按钮挂载到所属插件菜单下，工作台默认内容从示例模板收敛为真实导航与业务语义。启用`demo-control`插件时，插件治理写操作被阻断，但读取能力保留。

## Plugin and Repository Governance

### Plugin Locale Lifecycle

插件通过标准`manifest/i18n/<locale>/`目录交付语言资源。宿主在源码插件同步、动态插件安装、升级、启用、禁用与卸载过程中同步维护这些资源对运行时翻译聚合的影响。新增语言时，只要插件补齐对应目录资源，无需修改宿主代码或插件生命周期代码。

### Project Positioning and Host Boundary

LinaPro统一定位为“面向可持续交付的`AI`原生全栈框架”。`apps/lina-core`是核心宿主服务，负责提供通用模块接口、系统治理和插件扩展能力；默认管理工作台只是框架默认入口与适配层，而不是宿主能力边界本身。仅影响工作台展示结构的需求应优先通过前端适配或工作台适配接口解决。

### README and Bootstrap Governance

所有目录级主说明文档统一维护英文`README.md`与中文镜像`README.zh-CN.md`。数据库`init`与`mock`命令要求显式确认参数，并在执行`SQL`时采用首错即停策略，防止误操作和部分成功带来的不一致状态。

## Risks and Trade-offs

- 默认交付移除`zh-TW`会降低即装即用的语言覆盖面，但能显著降低内建能力和插件示例的维护成本；需要繁体中文的项目仍可按资源约定自行扩展。
- 分层缓存与作用域失效增加了实现复杂度，但可以避免任意改动触发全量清缓存，并为分布式环境的跨实例同步保留清晰边界。
- 共享`ResourceLoader`和接口拆分引入额外抽象层，但能显著减少重复实现和跨模块耦合。
- 固定`ltr`牺牲了未来自动切换文档方向的灵活性，但当前宿主优先目标是降低默认交付复杂度。
- 去除数据库翻译覆盖能力意味着不能通过运行时接口快速热修语言资源，但“文件资源单一事实来源”更利于审计、缺失检查和长期交付治理。
