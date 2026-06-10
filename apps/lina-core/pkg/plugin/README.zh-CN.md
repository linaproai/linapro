# LinaPro 插件公开契约

`apps/lina-core/pkg/plugin`承载`lina-core`稳定的插件侧契约。源码插件、动态插件 guest、动态插件构建器以及宿主集成代码都应通过这里的公开边界接入。宿主侧编排和持久化实现位于`apps/lina-core/internal/service/plugin`。

## 公开组件

| 组件 | 功能职责 |
| --- | --- |
| `capability` | 定义统一的`capability.Services`目录以及用户、文件、i18n、任务、AI、租户、组织、存储、缓存、锁等子包契约。源码插件直接接收这些能力；动态插件通过桥接传输访问等价能力。 |
| `pluginhost` | 定义源码插件的声明期契约、运行期 service 访问、生命周期回调、hook 注册、HTTP 路由贡献、定时任务贡献、菜单过滤、权限过滤和托管静态资源常量。 |
| `pluginbridge` | 提供动态插件 guest SDK。guest 代码在发现或构建阶段使用`pluginbridge.Declarations`，在运行阶段使用`pluginbridge.Services`。 |

## 领域能力

`capability.Services`是宿主持有的领域能力运行期目录。源码插件通过`pluginhost.Services`消费这些入口；动态插件声明匹配的`hostServices`，再通过`pluginbridge.Services`调用。可信源码插件管理命令保留在`capability.AdminServices`中，宿主内部 scope helper 不进入普通插件可见入口。

| 领域能力 | 职责边界 | 运行期与校验路径 |
| --- | --- | --- |
| `APIDoc` | 解析本地化 API 文档中的路由文本和标题 operation key。 | 通过能力目录或`apidoc`host call 提供；校验关注路由和 operation key 载荷。 |
| `Auth`/`Authz` | 处理租户 token 选择或切换、代用户 token、权限投影和权限判断。 | 运行期检查当前用户、租户和权限 key；管理命令保留在`AdminServices.Auth()`。 |
| `AI` | 聚合文本、图片、向量、音频、视觉、文档、安全审核和视频等类型化 AI 子能力。 | 调用按 method 授权，不接受资源声明，并把插件身份注入 provider 请求用于审计和治理。 |
| `Users` | 提供用户批量读取、用户搜索和可见性确认。 | 宿主实现必须让用户存在性和可见性检查保持在调用方边界内。 |
| `BizCtx` | 投影当前业务请求上下文。 | 作为只读运行期上下文桥接当前用户、租户、语言和请求元数据。 |
| `Dict` | 解析字典值标签，并校验类型化值可见性。 | 宿主校验保持在可见字典类型和值范围内。 |
| `Files` | 提供文件批量读取和可见性确认。 | 宿主校验避免插件探测或使用可见边界之外的文件 ID。 |
| `HostConfig` | 读取受治理的宿主运行期配置值。 | 动态声明必须列出`resources.keys`；源码插件获得窄化只读 service。 |
| `I18n` | 读取 locale、翻译消息并查找消息 key。 | 用于运行期本地化流程，不暴露翻译存储内部结构。 |
| `Infra` | 读取基础设施组件状态投影。 | 校验关注可见组件 ID 和只读状态投影。 |
| `Jobs` | 读取定时任务元数据并注册动态任务。 | 声明期任务契约先校验，再进入运行期任务发现和执行。 |
| `Notifications` | 读取通知消息并发送受治理通知。 | 读取调用不需要资源声明；`messages.send`需要`resources[].ref`边界。 |
| `Plugins` | 暴露插件注册表投影、插件作用域配置、启用状态和租户生命周期 hook。 | 运行期检查插件可见性、provider 启用、权威启用状态和租户生命周期前置条件。 |
| `Route` | 暴露当前执行的动态路由元数据。 | 用于运行期路由分发，不暴露宿主 router 内部实现。 |
| `Sessions` | 搜索和批量读取在线会话投影。 | 宿主校验保持会话可见性在调用方边界内。 |
| `Storage` | 提供插件作用域对象存储操作。 | 动态声明使用`resources.paths`；运行期分发校验路径边界和选中的存储 provider。 |
| `Cache` | 提供插件作用域缓存读取、写入、删除、自增和过期操作。 | 动态声明使用`resources[].ref`；运行期分发校验 namespace、key 和 TTL 载荷。 |
| `Lock` | 提供插件可见的分布式锁获取、续租和释放。 | 动态声明使用`resources[].ref`；运行期分发校验锁资源和租约载荷。 |
| `Manifest` | 读取插件作用域 manifest 或 artifact 资源。 | 动态声明使用`resources.paths`；源码和动态路径都通过插件作用域资源视图解析。 |
| `RecordStore` | 提供 guest 侧受治理的插件自有表读取、写入和事务构建器。 | 通过`datahost`分发；校验覆盖插件自有表、查询计划、过滤、排序、写入载荷、事务和审计元数据。 |
| `Org` | 提供可选组织投影，例如部门分配、部门名称和岗位 ID。 | provider 可用性显式暴露；组织 provider 缺失时 fallback service 返回安全中性值。 |
| `Tenant` | 提供可选租户上下文、可见性确认、成员校验、可访问租户列表和租户切换校验。 | provider 可用性显式暴露；宿主过滤器应用租户范围，但不暴露租户存储内部结构。 |

## 宿主领域实现

`apps/lina-core/internal/service/plugin`是宿主侧插件领域组件。根包提供统一门面，覆盖插件发现、管理列表、安装、启用、停用、卸载、运行期升级、源码插件升级、运行期路由分发、前端资源托管、依赖检查和能力装配。

## 声明期能力与运行期能力

### 声明期能力

声明期能力是插件的静态声明和注册输出。宿主在业务执行前使用这些内容构建治理状态。

源码插件通过`pluginhost.Declarations`表达声明期契约，包括`Assets()`、`Lifecycle()`、`Hooks()`、`HTTP()`、`Jobs()`和`Governance()`。

动态插件通过`plugin.yaml`、WASM 自定义 section、`pluginbridge.Declarations.Routes().Group(...)`、`pluginbridge.Declarations.Jobs().Register(...)`以及嵌入的`protocol`契约表达声明期契约，例如路由、任务、生命周期处理器、后端资源、前端资源、SQL、i18n 资源和`hostServices`。

### 运行期能力

运行期能力是在插件业务逻辑执行时可用的 service。

源码插件通过`pluginhost.Services`访问运行期能力；该接口内嵌普通`capability.Services`，并额外提供可信源码插件专用能力，例如`Admin()`和`TenantFilter()`。

动态插件通过`pluginbridge.Services`访问运行期能力。调用会通过`pluginbridge/protocol`编码，经由`WASI host call`传输，由派生的`HostCapabilities`和已确认的`HostServices`授权，再由`apps/lina-core/internal/service/plugin/internal/wasm`分发执行。

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
    methods: [get, put]
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

| Service | 资源声明 | 派生能力 | Methods |
| --- | --- | --- | --- |
| `runtime` | 无 | `host:runtime` | `log.write`<br/>`state.get`<br/>`state.set`<br/>`state.delete`<br/>`info.now`<br/>`info.uuid`<br/>`info.node` |
| `storage` | `resources.paths` | `host:storage` | `put`<br/>`get`<br/>`delete`<br/>`list`<br/>`stat` |
| `network` | `resources[].url` | `host:http:request` | `request` |
| `data` | `resources.tables` | `host:data:read`<br/>`host:data:mutate`| `list`<br/>`get`<br/>`create`<br/>`update`<br/>`delete`<br/>`transaction` |
| `cache` | `resources[].ref` | `host:cache` | `get`<br/>`set`<br/>`delete`<br/>`incr`<br/>`expire` |
| `lock` | `resources[].ref` | `host:lock` | `acquire`<br/>`renew`<br/>`release` |
| `hostconfig` | `resources.keys` | `host:hostconfig` | `get` |
| `manifest` | `resources.paths` | `host:manifest` | `get` |
| `apidoc` | 无 | `host:apidoc` | `route_text.resolve`<br/>`route_texts.resolve`<br/>`route_title_operation_keys.find` |
| `auth` | 无 | `host:auth:token` | `tenant.select`<br/>`tenant.switch`<br/>`impersonation_token.issue`<br/>`impersonation_token.revoke` |
| `authz` | 无 | `host:authz` | `permissions.batch_get`<br/>`permissions.has`<br/>`users.platform_admin.check` |
| `users` | 无 | `host:users` | `users.batch_get`<br/>`users.search`<br/>`users.visible.ensure` |
| `bizctx` | 无 | `host:bizctx` | `current.get` |
| `dict` | 无 | `host:dict` | `labels.resolve` |
| `files` | 无 | `host:files` | `files.batch_get`<br/>`files.visible.ensure` |
| `i18n` | 无 | `host:i18n` | `locale.get`<br/>`messages.translate`<br/>`messages.keys.find` |
| `infra` | 无 | `host:infra` | `status.batch_get` |
| `jobs` | 无 | `host:jobs` | `jobs.batch_get`<br/>`jobs.register` |
| `notifications` | 读取无资源；`messages.send`使用`resources[].ref` | `host:notifications` | `messages.batch_get`<br/>`messages.send` |
| `plugins` | 无 | `host:plugins` | `plugins.batch_get`<br/>`plugins.tenant.list`<br/>`plugins.enabled.check`<br/>`plugins.provider_enabled.check`<br/>`plugins.enabled_authoritative.check`<br/>`config.get`<br/>`lifecycle.tenant_plugin_disable.ensure`<br/>`lifecycle.tenant_plugin_disabled.notify`<br/>`lifecycle.tenant_delete.ensure`<br/>`lifecycle.tenant_deleted.notify` |
| `route` | 无 | `host:route` | `metadata.get` |
| `sessions` | 无 | `host:sessions` | `sessions.search`<br/>`sessions.batch_get` |
| `ai` | 无 | `host:ai:text`<br/>`host:ai:image`<br/>`host:ai:embedding`<br/>`host:ai:audio`<br/>`host:ai:vision`<br/>`host:ai:document`<br/>`host:ai:safety`<br/>`host:ai:video` | `text.generate`<br/>`image.generate`<br/>`image.edit`<br/>`embedding.create`<br/>`audio.transcribe`<br/>`audio.synthesize`<br/>`vision.analyze`<br/>`document.analyze`<br/>`document.cite`<br/>`safety.moderate`<br/>`video.generate`<br/>`video.edit`<br/>`video.extend`<br/>`video.operation.get`<br/>`video.operation.cancel` |
| `org` | 无 | `host:org` | `capability.available`<br/>`capability.status`<br/>`users.dept_assignments.list`<br/>`users.dept_info.get`<br/>`users.dept_name.get`<br/>`users.dept_ids.get`<br/>`users.post_ids.get` |
| `tenant` | 无 | `host:tenant` | `capability.available`<br/>`capability.status`<br/>`tenants.current`<br/>`tenants.platform_bypass`<br/>`tenants.visible.ensure`<br/>`users.tenant_membership.validate`<br/>`users.tenants.list`<br/>`tenants.switch.validate` |
| `secret` | `resources[].ref` | `host:secret` | `resolve` reserved |
| `event` | `resources[].ref` | `host:event:publish` | `publish` reserved |
| `queue` | `resources[].ref` | `host:queue:enqueue` | `enqueue` reserved |

## 维护说明

当插件公开契约或动态 `host service` 描述符发生变化时，需要同步更新本目录下的`README.md`和`README.zh-CN.md`。
