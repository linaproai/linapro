# LinaPro 插件公开契约

`apps/lina-core/pkg/plugin`承载`lina-core`稳定的插件侧契约。源码插件、动态插件 guest、动态插件构建器以及宿主集成代码都应通过这里的公开边界接入。宿主侧编排和持久化实现位于`apps/lina-core/internal/service/plugin`。

本包拥有插件内核、源码插件贡献 API、动态插件传输、core-owned 能力目录、通用 capability descriptor 接收、owner-aware 动态路由、授权、审计、缓存失效和生命周期治理。由业务插件拥有的非核心领域契约应位于该 owner 插件的`backend/cap/<domain>cap`目录。

## 公开组件

| 组件 | 功能职责 |
| --- | --- |
| `capability` | 定义统一的`capability.Services`目录以及用户、文件、i18n、任务、租户、组织、存储、缓存、锁、manifest、route、plugins 和业务上下文等 core-owned 能力子包契约。源码插件从`pluginhost`注册器和回调输入中接收该目录；动态插件只通过桥接传输访问已发布的能力子集。动态插件 i18n 资源由宿主管理，不发布`i18n`host service。 |
| `pluginhost` | 定义源码插件的声明期契约、运行期 service 访问、生命周期回调、hook 注册、HTTP 路由贡献、定时任务贡献、菜单过滤、权限过滤、托管静态资源常量和通用 capability descriptor 接收入口。它不得成为 plugin-owned 能力的领域专属 provider facade owner。 |
| `pluginbridge` | 提供动态插件 guest SDK、动态插件声明、core-owned host-service client 和 owner-aware host-call envelope。guest 代码在发现或构建阶段使用`pluginbridge.Declarations`，在运行阶段使用`pluginbridge.Services`。plugin-owned 领域 guest SDK 位于 owner 插件的`backend/cap/<domain>cap/bridge`或等价公开包。 |

## 领域能力

`capability.Services`是 core-owned 领域能力运行期目录。源码插件通过`pluginhost`注册器和回调输入返回的`capability.Services`消费这些入口；动态插件声明已经显式发布为动态`hostServices`的匹配入口，再通过`pluginbridge.Services`调用。每个 core-owned 领域只暴露一个插件可见`Service`；方法级契约承载风险、授权、数据权限、上下文、性能和缓存治理。领域方法依赖标准业务`ctx`获取当前用户、租户、权限和数据权限信息；动态`hostServices`授权保留在 dispatcher 内部校验。`I18n()`保留为源码插件运行期能力；动态插件 i18n 资源由宿主发现、合并、缓存和分发。插件生命周期归属`Plugins().Lifecycle()`，插件启用状态查询归属`Plugins().State()`；租户插件治理和租户过滤上下文归属`Tenant()`。宿主内部 scope helper 不通过普通插件可见入口暴露。

plugin-owned 非核心能力不归属`capability.Services`。owner 插件在`apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap`下发布普通消费契约、provider `SPI`、动态 guest SDK、descriptor helper 和版本策略。`lina-core`接收 owner descriptor，建立`owner + service + version + method`索引，校验依赖和授权快照，分发 owner-aware 动态调用，记录审计 envelope，并在生命周期变化后失效运行期缓存。`AI`是第一组 plugin-owned 能力族，归属`apps/lina-plugins/linapro-ai-core/backend/cap/aicap`。

### 统一`Service`与动态插件

插件可见领域不再提供独立管理目录。源码插件通过`pluginhost`输入返回的`capability.Services`目录上的`Services.<Domain>()`或窄化注入的`*cap.Service`执行读取、写入、执行和治理动作。插件生命周期和启用状态分别通过`Services.Plugins().Lifecycle()`和`Services.Plugins().State()`访问；租户插件治理和租户过滤上下文分别通过`Services.Tenant().Plugins()`和`Services.Tenant().Filter()`访问。`capability.Services`不再提供独立租户服务视图，也不暴露顶层`PluginLifecycle()`、`PluginState()`、`TenantPluginGovernance()`、`TenantFilter()`或`TenantTableFilter()`方法。同进程源码插件和宿主 adapter 需要约束 GoFrame 查询时，使用`tenantspi.ApplyPluginTableFilter(ctx, Services.Tenant().Filter(), model, qualifier)`或从注入的`tenantcap.Service.Filter()`派生过滤上下文；普通`tenantcap.Service`和动态插件只暴露可序列化租户过滤上下文。动态插件只能在治理方法已经注册、声明、授权并由 dispatcher 分发时，按普通领域`hostServices`调用这些插件和租户治理方法。

动态插件只能使用已经注册到动态`host service`目录、由插件清单声明、经宿主授权并由`WASM host-service`分发器处理的具体方法。如果确实需要让动态插件使用某个受治理动作，应在统一领域中发布窄化、版本化的动态方法，而不是引入平行管理服务。

| 领域能力 | 职责边界 | 运行期与校验路径 |
| --- | --- | --- |
| `APIDoc` | 解析本地化 API 文档中的路由文本和标题 operation key。 | 通过能力目录或`apidoc`host call 提供；校验关注路由和 operation key 载荷。 |
| `Auth` | 通过`Token()`和`Authz()`子能力处理租户 token 选择或切换、代用户 token、权限投影、单权限判断和批量权限判断。 | 运行期检查当前用户、租户、权限 key 和方法级授权。 |
| `AI` | 由`linapro-ai-core`以 plugin-owned 方式拥有；聚合文本、图片、向量、音频、视觉、文档、安全审核和视频等类型化 AI 子能力，并提供方法级状态投影。 | 源码插件 import `linapro-ai-core/backend/cap/aicap`；动态插件声明`owner: linapro-ai-core`、`service: ai`和`version: v1`。`lina-core`只治理 descriptor 发现、依赖检查、授权快照、owner-aware 路由、审计和缓存失效。 |
| `Users` | 提供用户基础读取、有界用户列表、可见性确认、受治理用户写入、状态与凭证动作，并通过`Users().Assignment()`聚合用户角色关联关系。 | 宿主实现必须让用户存在性和可见性检查保持在调用方边界内。用户角色关联方法默认仅属于源码/Go 契约，除非单独注册为动态`hostServices`；当前动态`users`服务只发布投影、批量、解析、列表和可见性方法。 |
| `BizCtx` | 投影当前业务请求上下文。 | 作为只读运行期上下文桥接当前用户、租户、语言和请求元数据。 |
| `Dict` | 解析字典值标签、列出有界字典候选，并校验类型化值可见性。 | 宿主校验保持在可见字典类型和值范围内。 |
| `Files` | 提供宿主文件中心投影、有界搜索、可见性确认、内容读取、受治理上传，以及从插件存储创建`sys_file`记录。 | 宿主校验避免插件探测或使用可见边界之外的文件 ID。上传会创建宿主文件中心元数据，插件私有对象仍由`Storage()`持有。 |
| `HostConfig` | 按宿主优先级链读取宿主配置值，并通过`SysConfig()`提供受治理的`sys_config`投影和写入方法。 | 动态`get`和单 key `sys_config`方法声明必须列出`resources.keys`。该能力独立于插件作用域业务配置。 |
| `I18n` | 为源码插件读取`locale`并翻译消息。 | 源码插件通过`pluginhost`输入中的`capability.Services`接收该能力；动态插件不接收`i18n`host service，因为其 i18n 资源由宿主管理。 |
| `Jobs` | 读取定时任务元数据、搜索有界任务候选、校验任务可见性，并执行受治理的运行期任务动作。 | 声明期任务契约通过`pluginbridge.Declarations.Jobs().Register(...)`提交；运行期服务不暴露`Register`。 |
| `Notifications` | 列表和读取类型化通知消息投影、按业务来源批量读取消息、校验可见性、删除消息、更新已读状态并发送受治理通知。 | actor 作用域的读取、删除和已读状态调用不需要资源声明；`messages.send`需要`resources[].ref`边界。 |
| `Plugins` | 暴露当前插件投影、插件注册表投影、租户插件分页、插件启用状态、provider 启用状态、插件作用域配置和受治理插件生命周期编排。 | 运行期检查插件可见性、插件启用/provider 状态、插件作用域配置来源隔离、生命周期前置条件，以及已发布治理方法的动态`hostServices`授权。 |
| `Route` | 暴露当前执行的动态路由元数据。 | 用于运行期路由分发，不暴露宿主 router 内部实现。 |
| `Sessions` | 读取当前在线会话投影、搜索在线会话、批量读取在线会话投影、校验会话可见性，并批量读取用户在线状态。 | 宿主校验保持会话和用户可见性在调用方边界内。 |
| `Storage` | 提供插件私有对象存储操作，用于插件自有附件、业务二进制对象、导入导出临时文件和卸载清理，并支持显式批量元数据、游标列表和批量删除。 | 源码插件通过`pluginhost`输入中的`capability.Services`获得插件作用域`Storage()`；动态声明使用`service: storage`和`resources.paths`；写入不会创建`sys_file`记录，也不会暴露 provider key 或本地路径。 |
| `Cache` | 提供插件作用域缓存读取、写入、删除、多键读取/写入/删除、自增和过期操作。 | 动态声明使用`resources[].ref`；运行期分发校验 namespace、key、value 大小和正 TTL 载荷。缓存仍是非权威插件运行期数据。 |
| `Lock` | 提供插件可见的分布式锁获取、续租和释放。 | 动态声明使用`resources[].ref`；运行期分发校验锁资源和租约载荷。 |
| `Manifest` | 读取插件作用域 manifest 或 artifact 资源，包括有界多资源读取和路径列表。 | 动态声明使用`resources.paths`；源码和动态路径都通过插件作用域资源视图解析。 |
| `Org` | 提供可选组织投影，例如用户组织档案、有界部门树、部门搜索、岗位选项、可见性校验、部门分配、部门名称和岗位 ID。 | provider 可用性显式暴露；组织 provider 缺失时 fallback service 返回安全中性值。 |
| `Tenant` | 提供可选租户上下文、租户信息、租户批量/搜索投影、可见性确认、成员校验、可访问租户列表、租户插件治理和插件自有表租户过滤上下文。 | provider 可用性显式暴露；租户生命周期写入和成员关系替换留在租户 owner 或宿主内部 SPI。同进程 Go 调用方持有 GoFrame model 时，可以把普通过滤上下文传给`tenantspi.ApplyPluginTableFilter(...)`。动态插件只接收可序列化过滤上下文，并必须通过`RecordStore`或 owner 侧过滤完成租户隔离。 |

### 插件分发治理

插件 manifest 和生命周期回调快照包含`distribution`，宿主会将其归一化为`managed`或`builtin`。省略时视为`managed`。`builtin`是仅适用于源码插件的治理模式，用于随宿主编译交付的项目组成能力：宿主会在启动时自动安装、启用和安全升级，普通插件管理写操作会被拒绝。动态插件必须保持`managed`语义，不能自声明内建治理。

## 动态插件专属能力

`Runtime`、`Network`和`RecordStore`是`pluginbridge.Default()`或`pluginbridge.New()`返回的`pluginbridge.Services`目录上的动态插件专属入口。它们不属于`capability.Services`，因为源码插件已经运行在宿主进程内，可以使用宿主原生等价能力，或者使用源码插件数据访问接缝，而不是 guest host-service 包装。

| 能力 | 公开入口 | 边界原因 |
| --- | --- | --- |
| `Runtime` | `pluginbridge.Default().Runtime()`或`pluginbridge.New().Runtime()` | 动态插件需要通过`WASI host-service`客户端写日志、读写状态、读取时间、生成 UUID 和读取节点身份；源码插件直接使用宿主原生日志和运行期上下文。 |
| `Network` | `pluginbridge.Default().Network()`或`pluginbridge.New().Network()` | 动态插件需要经由 host-service 授权访问受治理的出站 HTTP；源码插件使用宿主原生 HTTP client 或注入的领域 service。 |
| `RecordStore` | `pluginbridge.Default().RecordStore()`或`pluginbridge.New().RecordStore()`，以及`pkg/plugin/pluginbridge/recordstore` | 动态插件需要 guest 侧 facade 封装 data host-service 协议和类型化查询计划；源码插件使用自有 DAO 或 provider 接缝。 |

新增能力只有在源码插件和动态插件共享同一个稳定 core-owned 领域契约时，才应进入`capability.Services`。仅服务动态插件的 host-service client 和 guest SDK 应留在`pluginbridge`下。plugin-owned 非核心能力必须从 owner 插件的`backend/cap/<domain>cap`边界发布公开契约和 guest SDK，不得在`lina-core`中新增领域包、codec 或 dispatcher 分支。

### `Storage()`与`Files()`边界

| 场景 | 使用能力 | 边界 |
| --- | --- | --- |
| 插件保存自有附件、生成的导出文件、业务二进制对象或导入临时文件。 | `Storage()` / 动态`service: storage` | 插件传入 logical object path。宿主在委托给当前 storage provider 前按插件 ID 和租户加作用域。对象不会进入宿主文件中心列表，也不会创建`sys_file`元数据。 |
| 插件在记录删除或卸载清理时删除、列出、读取元数据或清理自有对象。 | `Storage()` / 动态`service: storage` | 清理通过 logical prefix 执行`Delete`或有界`List`后删除。插件不得直接删除宿主上传根目录、provider 根目录或文件中心记录。 |
| 插件引用用户已经上传到宿主文件中心的文件。 | `Files()` / 动态`service: files` | 插件只获得`filecap.FileInfo`信息和宿主文件 ID 的可见性校验。响应不得暴露`DAO`、`DO`、`Entity`、provider object key 或本地绝对路径。 |
| 插件命令从请求中接收宿主文件 ID。 | `Files().EnsureVisible` / `files.visible.ensure` | 命令执行写入前必须先校验全部 ID。不存在和不可见使用相同拒绝语义，避免泄露资源存在性。 |
| 插件需要上传内容并登记到宿主文件中心。 | `Files().Upload` / `files.upload` | 宿主通过文件 owner 写入，让`sys_file`记录租户、上传人、场景、hash 和存储元数据。动态直传有大小上限；更大的动态 payload 应先使用`Storage().Put`。 |
| 插件已经把对象写入私有存储，并需要宿主文件中心记录。 | `Files().CreateFromStorage` / `files.create_from_storage` | 宿主从插件作用域`Storage()`对象复制到文件中心存储。动态插件还必须为源路径声明`storage.get`。该操作不会移动或删除源对象，也不会暴露 provider key 或本地路径。 |

`Storage()`provider 选择不依赖主配置项。宿主在恰好一个 storage provider 插件可服务时使用该插件，没有可服务 provider 时回退到内置本地 provider，多个 provider 插件同时可服务时拒绝 storage 调用。

## 插件配置来源

源码插件通过`Services.Plugins().Config()`读取当前插件业务配置，动态插件通过`plugins.config.get`读取当前插件业务配置。这些入口都限定在插件作用域内，不暴露任意宿主配置 key，也不会读取兄弟插件配置。

配置来源优先级按配置段生效：

| 优先级 | 来源 | 运行期行为 |
| --- | --- | --- |
| 1 | 宿主主静态配置段`plugin.<plugin-id>` | 整个配置段作为该插件的有效配置来源。段内缺失的子 key 返回不存在或调用方默认值，不再回退到文件来源。 |
| 2 | 宿主配置根下的生产文件`plugins/<plugin-id>/config.yaml` | 仅在`plugin.<plugin-id>`不存在时使用。 |
| 3 | 开发期文件`apps/lina-plugins/<plugin-id>/manifest/config/config.yaml` | 仅在宿主静态配置和生产文件来源都不存在时使用。 |
| 4 | 动态 artifact 默认配置`manifest/config/config.yaml` | 作为当前动态插件执行上下文的最终回退来源。 |

`manifest/config/config.example.yaml`只作为模板，不会作为运行期默认值加载。`HostConfig()`仍然是独立的宿主配置能力；非 root 宿主 key 按当前`sys_config`快照、宿主静态配置、宿主默认值、缺失`nil`的顺序读取。动态`hostconfig.get`调用仍必须在`hostServices`中声明`resources.keys`授权。源码插件调用`HostConfig().Get(ctx, key, defaultValue)`时必须显式传入默认值；传入`nil`表示在宿主优先级链读取后保持缺失 key 返回`nil`的语义。

## 普通消费契约、Provider SPI 与 Guest SDK

插件侧包使用三类独立边界，确保每类调用方只导入自己可以安全依赖的契约：

| 边界 | 包形态 | 目标调用方 | 不得包含 |
| --- | --- | --- | --- |
| Core-owned 普通消费契约 | `pkg/plugin/capability/<domain>cap` | 通过`pluginhost`输入中的`capability.Services`调用的源码插件、通过生成或 bridge-backed client 调用的动态插件，以及宿主 adapter | GoFrame HTTP request 对象、provider factory 注册入口、宿主私有实现状态，或 GoFrame 数据库 builder |
| Core-owned 源码插件 Provider SPI | `pkg/plugin/capability/<domain>cap/<domain>spi` | 实现宿主领域 provider 的源码插件，以及宿主能力装配代码 | 动态插件 guest SDK import 或 WASM host-service wire 契约 |
| Core-owned 动态插件 Guest SDK | `pkg/plugin/pluginbridge`及其动态插件专属子包 | WASM guest 代码和动态插件构建器 | Provider SPI import 或源码插件注册 API |
| Plugin-owned 普通消费契约 | `apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap` | 依赖 owner 插件的源码插件、owner adapter，以及通过显式注入使用 owner 契约的宿主模块 | owner 插件`backend/internal`、DAO、DO、Entity、controller、provider 密钥、私有缓存或宿主 dispatcher |
| Plugin-owned Provider SPI 与 Guest SDK | `apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap/spi`和`apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap/bridge` | owner 插件 provider 注册 helper、源码 provider adapter，以及调用 owner-aware host service 的动态 guest | core `pluginhost`领域专属 facade、core `pluginbridge`领域专属 codec、owner 内部 service import 或绕过授权 |

Core-owned provider factory 声明归属`pluginhost.Declarations.Providers()`。源码 provider 插件通过`ProvideTenant`和`ProvideOrg`等领域方法声明 factory。宿主启动装配负责持有 provider manager 实例，并把共享 manager 注入宿主能力 service；普通`capability`包不保留包级 provider 注册表。plugin-owned provider 使用 owner helper 将 typed factory 包装为通用 capability descriptor。`pluginhost`只接收这些 descriptor，不得为非核心领域新增`ProvideAIText`等`Provide<Domain>`facade。

## 宿主领域实现

`apps/lina-core/internal/service/plugin`是宿主侧插件领域组件。根包提供统一门面，覆盖插件发现、管理列表、安装、启用、停用、卸载、运行期升级、源码插件升级、运行期路由分发、前端资源托管、依赖检查和能力装配。

## 声明期能力与运行期能力

### 声明期能力

声明期能力是插件的静态声明和注册输出。宿主在业务执行前使用这些内容构建治理状态。

源码插件通过`pluginhost.Declarations`表达声明期契约，包括`Assets()`、`Lifecycle()`、`Hooks()`、`HTTP()`、`Jobs()`和`Access()`。

动态插件通过`plugin.yaml`、WASM 自定义 section、`pluginbridge.Declarations.Routes().Group(...)`、`pluginbridge.Declarations.Jobs().Register(...)`以及嵌入的`protocol`契约表达声明期契约，例如路由、任务、生命周期处理器、后端资源、前端资源、SQL、i18n 资源和`hostServices`。

### 运行期能力

运行期能力是在插件业务逻辑执行时可用的 service。

源码插件通过`pluginhost`输入返回的`capability.Services`目录访问运行期能力。插件生命周期和状态通过`Services.Plugins().Lifecycle()`和`Services.Plugins().State()`暴露。租户插件治理和租户过滤上下文通过`Services.Tenant().Plugins()`和`Services.Tenant().Filter()`暴露。同进程表过滤通过显式`tenantspi`helper 调用完成，不作为单独的源码插件服务目录方法或租户服务镜像暴露。

动态插件通过`pluginbridge.Services`访问已发布的运行期能力。调用会通过`pluginbridge/protocol`编码，经由`WASI host call`传输，由派生的`HostCapabilities`和已确认的`HostServices`授权，再由`apps/lina-core/internal/service/plugin/internal/wasm`分发执行。动态插件不接收顶层兼容快捷入口或`I18n()`，但可以在`hostServices`授权后调用`Plugins()`和`Tenant()`下已经注册的治理方法。

每次 guest 执行时，宿主会构建请求级`HostServices`授权快照，每次 host call 仍按该快照校验`service`、`method`和资源身份。owner-aware plugin-owned 调用还会先校验`owner`、`version`、调用插件的`dependencies.plugins`声明、owner 插件启用状态、descriptor 注册状态和 method 授权，再分发到 owner handler。

## 动态插件 Host Service 声明

最小结构：

```yaml
hostServices:
  - service: runtime
    methods:
      - log.write
```

资源作用域结构：

```yaml
hostServices:
  - service: storage
    methods: [get, put, put.init, put.chunk, put.commit, put.abort]
    resources:
      paths:
        - reports/
  - service: data
    methods: [list, get]
    resources:
      tables:
        - plugin_acme_demo_report
  - service: hostconfig
    methods: [get]
    resources:
      keys:
        - i18n.default
  - service: network
    methods: [request]
    resources:
      - url: https://*.example.com/api
  - service: notifications
    methods: [messages.send]
    resources:
      - ref: inbox
        attributes:
          channel: inbox
```

owner-aware plugin-owned 结构：

```yaml
dependencies:
  plugins:
    - id: linapro-ai-core
      version: ">=0.1.0"

hostServices:
  - service: ai
    owner: linapro-ai-core
    version: v1
    methods:
      - text.generate
      - text.method_status.get
```

未声明`owner`的行表示 core-owned service。plugin-owned service 必须声明`owner`和`version`；宿主会合并 core-owned 静态 catalog 和 owner descriptor 投影，用于校验、授权、升级预览和运行期分发。下面的生成 catalog 来自宿主 catalog 实现，并且只列出 core-owned services。

## 可声明 Host Services

<!-- BEGIN generated:host-services -->
| Service | 资源声明 | 派生能力 | Methods |
| --- | --- | --- | --- |
| `runtime` | 无 | `host:runtime` | `log.write`<br/>`state.get`<br/>`state.get_many`<br/>`state.set`<br/>`state.set_many`<br/>`state.delete`<br/>`state.delete_many`<br/>`info.now`<br/>`info.uuid`<br/>`info.node` |
| `storage` | `resources.paths` | `host:storage` | `put`<br/>`put.init`<br/>`put.chunk`<br/>`put.commit`<br/>`put.abort`<br/>`get`<br/>`delete`<br/>`delete.batch`<br/>`list`<br/>`list.cursor`<br/>`stat`<br/>`stat.batch` |
| `network` | `resources[].url` | `host:http:request` | `request` |
| `data` | `resources.tables` | `host:data:read`<br/>`host:data:mutate` | `list`<br/>`get`<br/>`batch_get`<br/>`create`<br/>`update`<br/>`delete`<br/>`transaction` |
| `cache` | `resources[].ref` | `host:cache` | `get`<br/>`get_many`<br/>`set`<br/>`set_many`<br/>`delete`<br/>`delete_many`<br/>`incr`<br/>`expire` |
| `lock` | `resources[].ref` | `host:lock` | `acquire`<br/>`renew`<br/>`release` |
| `hostconfig` | `resources.keys` | `host:hostconfig` | `get`<br/>`sys_config.get`<br/>`sys_config.value.set`<br/>`sys_config.reset` |
| `manifest` | `resources.paths` | `host:manifest` | `get`<br/>`get_many`<br/>`list` |
| `apidoc` | 无 | `host:apidoc` | `route_text.resolve`<br/>`route_texts.resolve`<br/>`route_title_operation_keys.find` |
| `auth` | 无 | `host:auth:token`<br/>`host:auth:authz` | `token.tenant.select`<br/>`token.tenant.switch`<br/>`token.impersonation_token.issue`<br/>`token.impersonation_token.revoke`<br/>`authz.permissions.batch_get`<br/>`authz.permissions.batch_has`<br/>`authz.permissions.has`<br/>`authz.users.platform_admin.check`<br/>`authz.role_permissions.replace` |
| `users` | 无 | `host:users` | `users.current.get`<br/>`users.batch_get`<br/>`users.resolve.batch`<br/>`users.list`<br/>`users.visible.ensure`<br/>`users.create`<br/>`users.update`<br/>`users.delete`<br/>`users.status.set`<br/>`users.password.reset`<br/>`users.assignment.roles.replace` |
| `bizctx` | 无 | `host:bizctx` | `current.get` |
| `dict` | 无 | `host:dict` | `dict.refresh`<br/>`dict.type.get`<br/>`dict.type.batch_get`<br/>`dict.type.list`<br/>`dict.type.visible.ensure`<br/>`dict.type.keys.visible.ensure`<br/>`dict.type.create`<br/>`dict.type.update`<br/>`dict.type.delete`<br/>`dict.value.get`<br/>`dict.value.batch_get`<br/>`dict.value.labels.resolve`<br/>`dict.value.list`<br/>`dict.value.visible.ensure`<br/>`dict.value.values.visible.ensure`<br/>`dict.value.create`<br/>`dict.value.update`<br/>`dict.value.delete`<br/>`dict.value.by_type.delete` |
| `files` | 无 | `host:files` | `files.batch_get`<br/>`files.list`<br/>`files.visible.ensure`<br/>`files.upload`<br/>`files.create_from_storage`<br/>`files.metadata.update`<br/>`files.delete`<br/>`files.delete_many` |
| `jobs` | 无 | `host:jobs` | `jobs.batch_get`<br/>`jobs.list`<br/>`jobs.visible.ensure`<br/>`jobs.create`<br/>`jobs.update`<br/>`jobs.delete`<br/>`jobs.run`<br/>`jobs.status.set`<br/>`jobs.register` |
| `notifications` | 除`messages.send`使用`resources[].ref`外无需资源声明 | `host:notifications` | `messages.batch_get`<br/>`messages.list`<br/>`messages.by_source.batch_get`<br/>`messages.visible.ensure`<br/>`messages.send`<br/>`messages.delete`<br/>`messages.by_source.delete`<br/>`messages.mark_read`<br/>`messages.mark_unread` |
| `plugins` | 无 | `host:plugins` | `plugins.current.get`<br/>`plugins.batch_get`<br/>`plugins.registry.list`<br/>`plugins.tenant.list`<br/>`config.get`<br/>`plugins.state.enabled.check`<br/>`plugins.state.provider_enabled.check`<br/>`plugins.state.enabled_authoritative.check`<br/>`plugins.lifecycle.tenant_plugin_disable.ensure`<br/>`plugins.lifecycle.tenant_plugin_disabled.notify`<br/>`plugins.lifecycle.tenant_delete.ensure`<br/>`plugins.lifecycle.tenant_deleted.notify` |
| `route` | 无 | `host:route` | `metadata.get` |
| `sessions` | 无 | `host:sessions` | `sessions.current.get`<br/>`sessions.list`<br/>`sessions.batch_get`<br/>`sessions.users.online.batch_get`<br/>`sessions.visible.ensure`<br/>`sessions.revoke`<br/>`sessions.revoke_many` |
| `org` | 无 | `host:org` | `capability.available`<br/>`capability.status`<br/>`org.assignment.user_profiles.batch_get`<br/>`org.department.tree.list`<br/>`org.department.batch_get`<br/>`org.department.list`<br/>`org.post.batch_get`<br/>`org.post.options.list`<br/>`org.department.visible.ensure_many`<br/>`org.post.visible.ensure_many`<br/>`org.department.create`<br/>`org.department.update`<br/>`org.department.delete`<br/>`org.post.create`<br/>`org.post.update`<br/>`org.post.delete`<br/>`org.assignment.by_user.replace`<br/>`org.assignment.by_user.cleanup` |
| `tenant` | 无 | `host:tenant` | `capability.available`<br/>`capability.status`<br/>`tenant.context.current`<br/>`tenant.context.info`<br/>`tenant.context.platform_bypass`<br/>`tenant.directory.batch_get`<br/>`tenant.directory.list`<br/>`tenant.membership.validate`<br/>`tenant.membership.list_by_user`<br/>`tenant.directory.visible.ensure_many`<br/>`tenant.plugins.enabled.set`<br/>`tenant.plugins.defaults.provision`<br/>`tenant.filter.context` |
<!-- END generated:host-services -->

## 维护说明

当插件公开契约或动态 `host service` 描述符发生变化时，需要同步更新本目录下的`README.md`和`README.zh-CN.md`。

### Host Service 载荷与常量

- 新增 core-owned host service 方法必须使用统一 JSON envelope（`HostServiceJSONRequest` / `HostServiceJSONResponse` 或空载荷）。不得为新增方法引入 dedicated binary codec。
- 存量 dedicated codec 方法集合已冻结；catalog 治理测试会拒绝名单外的 dedicated 扩张。
- `service` / `method` 线值常量只维护在 `pluginbridge/protocol/hostservices/wire_constants.go`。catalog 必须引用这些常量（禁止重复写 wire 字符串字面量）；catalog 测试会校验。
- WASM 分发侧新增 JSON 方法应优先复用 `decodeCapabilityJSONRequest` 与 `capabilityJSONResponse` helper。