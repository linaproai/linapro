# LinaPro 插件公开契约

`apps/lina-core/pkg/plugin`承载`lina-core`稳定的插件侧契约。源码插件、动态插件 guest、动态插件构建器以及宿主集成代码都应通过这里的公开边界接入。宿主侧编排和持久化实现位于`apps/lina-core/internal/service/plugin`。

## 公开组件

| 组件 | 功能职责 |
| --- | --- |
| `capability` | 定义统一的`capability.Services`目录以及用户、文件、i18n、任务、AI、租户、组织、存储、缓存、锁等子包契约。源码插件直接接收完整目录；动态插件只通过桥接传输访问已发布的能力子集。动态插件 i18n 资源由宿主管理，不发布`i18n`host service。 |
| `pluginhost` | 定义源码插件的声明期契约、运行期 service 访问、生命周期回调、hook 注册、HTTP 路由贡献、定时任务贡献、菜单过滤、权限过滤和托管静态资源常量。 |
| `pluginbridge` | 提供动态插件 guest SDK。guest 代码在发现或构建阶段使用`pluginbridge.Declarations`，在运行阶段使用`pluginbridge.Services`。 |

## 领域能力

`capability.Services`是宿主持有的领域能力运行期目录。源码插件通过`pluginhost.Services`消费这些入口；动态插件声明已经显式发布为动态`hostServices`的匹配入口，再通过`pluginbridge.Services`调用。`I18n()`保留为源码插件运行期能力；动态插件 i18n 资源由宿主发现、合并、缓存和分发。可信源码插件管理命令保留在`capability.AdminServices`中，宿主内部 scope helper 不进入普通插件可见入口。

### `AdminServices`与动态插件

`capability.AdminServices`只通过`pluginhost.Services.Admin()`暴露给可信源码插件。动态插件的`pluginbridge.Services`没有`Admin()`入口，因此不能直接消费`sessioncap.AdminService`或`notifycap.AdminService`这类领域`AdminService`接口。

动态插件只能使用已经显式发布为动态`hostServices`、由插件清单声明、经宿主授权并注册到`WASM host-service`分发器中的具体方法。例如当前`sessions`动态服务发布读取、搜索、批量、用户在线状态和可见性方法，不发布`sessioncap.AdminService.Revoke`。如果确实需要让动态插件使用某个管理命令，应为该动作新增窄化、版本化的`host-service`方法，而不是开放完整`AdminServices`目录。

| 领域能力 | 职责边界 | 运行期与校验路径 |
| --- | --- | --- |
| `APIDoc` | 解析本地化 API 文档中的路由文本和标题 operation key。 | 通过能力目录或`apidoc`host call 提供；校验关注路由和 operation key 载荷。 |
| `Auth`/`Authz` | 处理租户 token 选择或切换、代用户 token、权限投影、单权限判断和批量权限判断。 | 运行期检查当前用户、租户和权限 key；管理命令保留在`AdminServices.Auth()`。 |
| `AI` | 聚合文本、图片、向量、音频、视觉、文档、安全审核和视频等类型化 AI 子能力，并提供方法级状态投影。 | 调用按 method 授权，不接受资源声明，并把插件身份注入 provider 请求用于审计和治理。状态读取只暴露可用性、原因和公开 provider 身份。 |
| `Users` | 提供当前用户投影、用户批量读取、用户批量解析、用户搜索和可见性确认。 | 宿主实现必须让用户存在性和可见性检查保持在调用方边界内。 |
| `BizCtx` | 投影当前业务请求上下文。 | 作为只读运行期上下文桥接当前用户、租户、语言和请求元数据。 |
| `Dict` | 解析字典值标签、列出有界字典候选，并校验类型化值可见性。 | 宿主校验保持在可见字典类型和值范围内。 |
| `Files` | 提供已有`sys_file`资源的宿主文件中心投影、有界搜索和可见性确认。 | 宿主校验避免插件探测或使用可见边界之外的文件 ID；该能力不负责写入、读取、删除或列出插件私有对象。 |
| `HostConfig` | 读取受治理的宿主运行期配置值。 | 动态声明必须列出`resources.keys`；源码插件获得窄化只读 service。 |
| `I18n` | 为源码插件读取 locale、翻译消息并查找消息 key。 | 源码插件通过`pluginhost.Services`接收该能力；动态插件不接收`i18n`host service，因为其 i18n 资源由宿主管理。 |
| `Infra` | 读取基础设施组件状态投影。 | 校验关注可见组件 ID 和只读状态投影。 |
| `Jobs` | 读取定时任务元数据、搜索有界任务候选、校验任务可见性并注册动态任务。 | 声明期任务契约先校验，再进入运行期任务发现和执行；运行期搜索和可见性校验保持只读。 |
| `Notifications` | 读取类型化通知消息投影、按业务来源批量读取消息、校验消息可见性并发送受治理通知。 | 读取调用不需要资源声明且保持 actor 可见；`messages.send`需要`resources[].ref`边界。 |
| `Plugins` | 暴露当前插件投影、插件注册表投影、租户插件分页、能力状态、插件作用域配置、启用状态和租户生命周期 hook。 | 运行期检查插件可见性、provider 启用、权威启用状态和租户生命周期前置条件。 |
| `Route` | 暴露当前执行的动态路由元数据。 | 用于运行期路由分发，不暴露宿主 router 内部实现。 |
| `Sessions` | 读取当前在线会话投影、搜索在线会话、批量读取在线会话投影、校验会话可见性，并批量读取用户在线状态。 | 宿主校验保持会话和用户可见性在调用方边界内。 |
| `Storage` | 提供插件私有对象存储操作，用于插件自有附件、业务二进制对象、导入导出临时文件和卸载清理，并支持显式批量元数据、游标列表和批量删除。 | 源码插件通过`pluginhost.Services`获得插件作用域`Storage()`；动态声明使用`service: storage`和`resources.paths`；写入不会创建`sys_file`记录，也不会暴露 provider key 或本地路径。 |
| `Cache` | 提供插件作用域缓存读取、写入、删除、多键读取/写入/删除、自增和过期操作。 | 动态声明使用`resources[].ref`；运行期分发校验 namespace、key、value 大小和 TTL 载荷。缓存仍是非权威插件运行期数据。 |
| `Lock` | 提供插件可见的分布式锁获取、续租和释放。 | 动态声明使用`resources[].ref`；运行期分发校验锁资源和租约载荷。 |
| `Manifest` | 读取插件作用域 manifest 或 artifact 资源，包括有界多资源读取和路径列表。 | 动态声明使用`resources.paths`；源码和动态路径都通过插件作用域资源视图解析。 |
| `Org` | 提供可选组织投影，例如用户组织档案、有界部门树、部门搜索、岗位选项、可见性校验、部门分配、部门名称和岗位 ID。 | provider 可用性显式暴露；组织 provider 缺失时 fallback service 返回安全中性值。 |
| `Tenant` | 提供可选租户上下文、租户信息、租户批量/搜索投影、可见性确认、成员校验、可访问租户列表、批量用户租户列表和租户切换校验。 | provider 可用性显式暴露；宿主过滤器应用租户范围，但不暴露租户存储内部结构。 |

## 动态插件专属能力

`Runtime`、`Network`和`RecordStore`是`pluginbridge.Services`上的动态插件专属入口。它们不属于`capability.Services`，因为源码插件已经运行在宿主进程内，可以使用宿主原生等价能力，或者使用源码插件数据访问接缝，而不是 guest host-service 包装。

| 能力 | 公开入口 | 边界原因 |
| --- | --- | --- |
| `Runtime` | `pluginbridge.Services.Runtime()` | 动态插件需要通过`WASI host-service`客户端写日志、读写状态、读取时间、生成 UUID 和读取节点身份；源码插件直接使用宿主原生日志和运行期上下文。 |
| `Network` | `pluginbridge.Services.Network()` | 动态插件需要经由 host-service 授权访问受治理的出站 HTTP；源码插件使用宿主原生 HTTP client 或注入的领域 service。 |
| `RecordStore` | `pluginbridge.Services.RecordStore()`与`pkg/plugin/pluginbridge/recordstore` | 动态插件需要 guest 侧 facade 封装 data host-service 协议和类型化查询计划；源码插件使用自有 DAO 或 provider 接缝。 |

新增能力只有在源码插件和动态插件共享同一个稳定宿主持有领域契约时，才应进入`capability.Services`。仅服务动态插件的 host-service client 和 guest SDK 应留在`pluginbridge`下。

### `Storage()`与`Files()`边界

| 场景 | 使用能力 | 边界 |
| --- | --- | --- |
| 插件保存自有附件、生成的导出文件、业务二进制对象或导入临时文件。 | `Storage()` / 动态`service: storage` | 插件传入 logical object path。宿主在委托给当前 storage provider 前按插件 ID 和租户加作用域。对象不会进入宿主文件中心列表，也不会创建`sys_file`元数据。 |
| 插件在记录删除或卸载清理时删除、列出、读取元数据或清理自有对象。 | `Storage()` / 动态`service: storage` | 清理通过 logical prefix 执行`Delete`或有界`List`后删除。插件不得直接删除宿主上传根目录、provider 根目录或文件中心记录。 |
| 插件引用用户已经上传到宿主文件中心的文件。 | `Files()` / 动态`service: files` | 插件只获得`filecap.FileProjection`投影和宿主文件 ID 的可见性校验。响应不得暴露`DAO`、`DO`、`Entity`、provider object key 或本地绝对路径。 |
| 插件命令从请求中接收宿主文件 ID。 | `Files().EnsureVisible` / `files.visible.ensure` | 命令执行写入前必须先校验全部 ID。不存在和不可见使用相同拒绝语义，避免泄露资源存在性。 |

`Storage()`provider 选择不依赖主配置项。宿主在恰好一个 storage provider 插件可服务时使用该插件，没有可服务 provider 时回退到内置本地 provider，多个 provider 插件同时可服务时拒绝 storage 调用。

## 普通消费契约、Provider SPI 与 Guest SDK

插件侧包使用三类独立边界，确保每类调用方只导入自己可以安全依赖的契约：

| 边界 | 包形态 | 目标调用方 | 不得包含 |
| --- | --- | --- | --- |
| 普通消费契约 | `pkg/plugin/capability/<domain>cap` | 通过`pluginhost.Services`调用的源码插件、通过生成或 bridge-backed client 调用的动态插件，以及宿主 adapter | GoFrame 数据库 builder、GoFrame HTTP request 对象、provider factory 注册入口或宿主私有实现状态 |
| 源码插件 Provider SPI | `pkg/plugin/capability/<domain>cap/<domain>spi` | 实现宿主领域 provider 的源码插件，以及宿主能力装配代码 | 动态插件 guest SDK import 或 WASM host-service wire 契约 |
| 动态插件 Guest SDK | `pkg/plugin/pluginbridge`及其动态插件专属子包 | WASM guest 代码和动态插件构建器 | Provider SPI import 或源码插件注册 API |

Provider factory 声明归属`pluginhost.Declarations.Providers()`。源码 provider 插件通过`ProvideTenant`、`ProvideOrg`和`ProvideAIText`等领域方法声明 factory。宿主启动装配负责持有 provider manager 实例，并把共享 manager 注入宿主能力 service；普通`capability`包不保留包级 provider 注册表。

## 宿主领域实现

`apps/lina-core/internal/service/plugin`是宿主侧插件领域组件。根包提供统一门面，覆盖插件发现、管理列表、安装、启用、停用、卸载、运行期升级、源码插件升级、运行期路由分发、前端资源托管、依赖检查和能力装配。

## 声明期能力与运行期能力

### 声明期能力

声明期能力是插件的静态声明和注册输出。宿主在业务执行前使用这些内容构建治理状态。

源码插件通过`pluginhost.Declarations`表达声明期契约，包括`Assets()`、`Lifecycle()`、`Hooks()`、`HTTP()`、`Jobs()`和`Access()`。

动态插件通过`plugin.yaml`、WASM 自定义 section、`pluginbridge.Declarations.Routes().Group(...)`、`pluginbridge.Declarations.Jobs().Register(...)`以及嵌入的`protocol`契约表达声明期契约，例如路由、任务、生命周期处理器、后端资源、前端资源、SQL、i18n 资源和`hostServices`。

### 运行期能力

运行期能力是在插件业务逻辑执行时可用的 service。

源码插件通过`pluginhost.Services`访问运行期能力；该接口内嵌普通`capability.Services`，并额外提供可信源码插件专用能力，例如`Admin()`和`TenantFilter()`。

动态插件通过`pluginbridge.Services`访问已发布的运行期能力。调用会通过`pluginbridge/protocol`编码，经由`WASI host call`传输，由派生的`HostCapabilities`和已确认的`HostServices`授权，再由`apps/lina-core/internal/service/plugin/internal/wasm`分发执行。该运行期目录不暴露`Admin()`、`TenantFilter()`或`I18n()`这类源码插件专属入口。

每次 guest 执行时，宿主会构建请求级`HostServices`授权快照，每次 host call 仍按该快照校验`service`、`method`和资源身份。

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
| `secret` | `resources[].ref` | `host:secret` | `resolve` reserved |
| `event` | `resources[].ref` | `host:event:publish` | `publish` reserved |
| `queue` | `resources[].ref` | `host:queue:enqueue` | `enqueue` reserved |
| `hostconfig` | `resources.keys` | `host:hostconfig` | `get` |
| `manifest` | `resources.paths` | `host:manifest` | `get`<br/>`get_many`<br/>`list` |
| `apidoc` | 无 | `host:apidoc` | `route_text.resolve`<br/>`route_texts.resolve`<br/>`route_title_operation_keys.find` |
| `auth` | 无 | `host:auth:token` | `tenant.select`<br/>`tenant.switch`<br/>`impersonation_token.issue`<br/>`impersonation_token.revoke` |
| `authz` | 无 | `host:authz` | `permissions.batch_get`<br/>`permissions.batch_has`<br/>`permissions.has`<br/>`users.platform_admin.check` |
| `ai` | 无 | `host:ai:text`<br/>`host:ai`<br/>`host:ai:image`<br/>`host:ai:embedding`<br/>`host:ai:audio`<br/>`host:ai:vision`<br/>`host:ai:document`<br/>`host:ai:safety`<br/>`host:ai:video` | `text.generate`<br/>`text.method_status.get`<br/>`ai.methods.status.batch_get`<br/>`image.generate`<br/>`image.edit`<br/>`embedding.create`<br/>`audio.transcribe`<br/>`audio.synthesize`<br/>`vision.analyze`<br/>`document.analyze`<br/>`document.cite`<br/>`safety.moderate`<br/>`video.generate`<br/>`video.edit`<br/>`video.extend`<br/>`video.operation.get`<br/>`video.operation.cancel` |
| `users` | 无 | `host:users` | `users.current.get`<br/>`users.batch_get`<br/>`users.resolve.batch`<br/>`users.search`<br/>`users.visible.ensure` |
| `bizctx` | 无 | `host:bizctx` | `current.get` |
| `dict` | 无 | `host:dict` | `labels.resolve`<br/>`dict.values.list`<br/>`values.visible.ensure` |
| `files` | 无 | `host:files` | `files.batch_get`<br/>`files.search`<br/>`files.visible.ensure` |
| `infra` | 无 | `host:infra` | `status.batch_get` |
| `jobs` | 无 | `host:jobs` | `jobs.batch_get`<br/>`jobs.search`<br/>`jobs.visible.ensure`<br/>`jobs.register` |
| `notifications` | 读取无资源；`messages.send`使用`resources[].ref` | `host:notifications` | `messages.batch_get`<br/>`messages.by_source.batch_get`<br/>`messages.visible.ensure`<br/>`messages.send` |
| `plugins` | 无 | `host:plugins` | `plugins.current.get`<br/>`plugins.batch_get`<br/>`plugins.search`<br/>`plugins.tenant.list`<br/>`plugins.capabilities.status.batch_get`<br/>`plugins.enabled.check`<br/>`plugins.provider_enabled.check`<br/>`plugins.enabled_authoritative.check`<br/>`config.get`<br/>`lifecycle.tenant_plugin_disable.ensure`<br/>`lifecycle.tenant_plugin_disabled.notify`<br/>`lifecycle.tenant_delete.ensure`<br/>`lifecycle.tenant_deleted.notify` |
| `route` | 无 | `host:route` | `metadata.get` |
| `sessions` | 无 | `host:sessions` | `sessions.current.get`<br/>`sessions.search`<br/>`sessions.batch_get`<br/>`sessions.users.online.batch_get`<br/>`sessions.visible.ensure` |
| `org` | 无 | `host:org` | `capability.available`<br/>`capability.status`<br/>`users.dept_assignments.list`<br/>`users.org_profiles.batch_get`<br/>`users.dept_info.get`<br/>`users.dept_name.get`<br/>`users.dept_ids.get`<br/>`users.post_ids.get`<br/>`depts.tree.list`<br/>`depts.search`<br/>`posts.options.list`<br/>`depts.visible.ensure`<br/>`posts.visible.ensure` |
| `tenant` | 无 | `host:tenant` | `capability.available`<br/>`capability.status`<br/>`tenants.current`<br/>`tenants.current_info.get`<br/>`tenants.platform_bypass`<br/>`tenants.visible.ensure`<br/>`tenants.batch_get`<br/>`tenants.search`<br/>`users.tenant_membership.validate`<br/>`users.tenants.list`<br/>`users.tenants.batch_list`<br/>`tenants.visible.batch_ensure`<br/>`tenants.switch.validate` |
<!-- END generated:host-services -->

## 维护说明

当插件公开契约或动态 `host service` 描述符发生变化时，需要同步更新本目录下的`README.md`和`README.zh-CN.md`。
