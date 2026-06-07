# 插件包

`apps/lina-core/pkg/plugin`是`Lina Core`插件体系的公开`Go`契约边界。它把插件可见契约集中在一个位置，运行期实现、持久化、生命周期编排和工作台适配仍由宿主侧服务承载。

该组件主要承担三类工作：

- 发布源码插件和动态插件共享的稳定宿主能力接口。
- 发布源码插件用于注册路由、钩子、定时任务、生命周期回调、嵌入资产和治理过滤器的宿主契约。
- 发布动态插件用于`Wasm`路由执行、`guest`运行时辅助、产物元数据和`plugin.yaml`中受治理`hostServices`调用的桥接契约。

## 边界

该包是契约边界，不是业务实现层。它不得向普通插件暴露宿主`DAO`、`DO`、`Entity`、原始`gdb.Model`构造器、具体工作台页面结构或宿主实现包。

源码插件通常使用`pluginhost`以及`pluginhost.Services`返回的类型化服务。动态插件通常使用`pluginbridge/guest`中的`guest`侧宿主服务客户端和低层路由执行辅助能力。

## 目录结构

| 路径 | 职责 |
|------|------|
| `capability/` | 源码插件和动态插件共享的稳定宿主能力目录。负责聚合`Services`、`AdminServices`、插件作用域服务绑定和类型化领域契约。 |
| `pluginbridge/` | 动态插件桥接命名空间。负责公开协议门面、`guest`运行时辅助、桥接契约和`Wasm`宿主调用编解码。 |
| `pluginhost/` | 源码插件宿主命名空间。负责嵌入式源码插件的编译期注册 facade 和运行期回调契约。 |

## `capability`包

`capability`定义稳定的宿主能力目录。宿主运行时服务实现这些接口，插件代码应依赖这些窄契约。

| 路径 | 职责 |
|------|------|
| `capability/` | 聚合`Services`、`AdminServices`和用于插件作用域视图的`ServicesForPlugin`。 |
| `capability/aicap/` | 根`AI`命名空间，聚合类型化`AI`子能力，并通过`ForPlugin`注入来源插件身份。 |
| `capability/aicap/aicommon/` | 共享`AI`值对象、能力类型、方法、档位、资产引用、提供方投影、操作引用、用量和不可用状态辅助能力。 |
| `capability/aicap/aitext/` | 文本生成能力契约、提供方契约、降级行为和来源身份绑定。 |
| `capability/aicap/aiimage/` | 图片生成和图片编辑能力契约。 |
| `capability/aicap/aiembedding/` | 向量嵌入创建能力契约。 |
| `capability/aicap/aiaudio/` | 音频转写和语音合成能力契约。 |
| `capability/aicap/aivision/` | 图片、截图、图示和帧内容的视觉分析能力契约。 |
| `capability/aicap/aidocument/` | 文档分析和带引用文档问答能力契约。 |
| `capability/aicap/aisafety/` | 面向文本和资产输入的安全审核能力契约。 |
| `capability/aicap/aivideo/` | 视频生成、编辑、延展和提供方操作能力契约。 |
| `capability/apidoccap/` | 面向源码插件和动态路由的接口文档文本查询与路由操作键辅助能力。 |
| `capability/authcap/` | 认证与授权命名空间，聚合令牌和授权子能力。 |
| `capability/authcap/authz/` | 授权领域读取和管理契约，不暴露宿主角色、菜单或权限表。 |
| `capability/authcap/token/` | 租户令牌选择、租户切换和模拟登录令牌交接契约，不暴露宿主`JWT`内部实现。 |
| `capability/bizctxcap/` | 只读业务上下文投影，用于用户、租户、模拟登录和平台绕过状态。 |
| `capability/cachecap/` | 插件作用域运行时缓存契约，提供字符串和整数值原语。 |
| `capability/capmodel/` | 共享的存储无关领域原语，例如`CapabilityContext`、操作者和来源元数据、批量结果、分页结果和本地化标签。 |
| `capability/configcap/` | 受治理运行时配置能力契约，用于读取和管理宿主拥有的配置投影。 |
| `capability/dictcap/` | 字典领域标签解析和刷新契约，不暴露字典表。 |
| `capability/filecap/` | 受治理文件投影和文件删除契约，不暴露物理路径或存储表。 |
| `capability/hostconfigcap/` | 面向已授权源码插件和动态`hostconfig`调用的只读宿主配置访问能力。 |
| `capability/i18ncap/` | 面向源码插件的运行时语言和翻译查询契约。 |
| `capability/infracap/` | 基础设施状态投影和刷新契约，不泄漏具体运行时后端。 |
| `capability/jobcap/` | 定时任务投影以及受治理任务执行或状态管理契约。 |
| `capability/manifestcap/` | 面向源码插件和动态插件的插件作用域`manifest/`资源只读访问能力。 |
| `capability/notifycap/` | 通知投影和受治理消息发送、删除契约，不暴露通知表。 |
| `capability/orgcap/` | 可选组织能力、提供方注册、用户部门和岗位投影，以及组织范围接缝。 |
| `capability/plugincap/` | 插件治理投影，以及插件本地配置、状态、生命周期、注册表和租户默认供给管理契约。 |
| `capability/recordstore/` | 面向已授权动态插件`data`服务表的`guest`侧受治理类`ORM`门面。 |
| `capability/routecap/` | 附着在当前请求上的动态路由元数据投影。 |
| `capability/sessioncap/` | 在线会话搜索、批量读取和吊销契约，不暴露会话存储。 |
| `capability/tenantcap/` | 可选租户能力、提供方注册、租户解析、可见性校验、租户切换，以及租户范围接缝。 |
| `capability/usercap/` | 用户领域投影、搜索、可见性校验和受治理状态管理，不暴露`sys_user`。 |

## `pluginbridge`职责

`pluginbridge`是动态插件协议区域，由宿主运行时、动态插件构建器和`Wasm guest`代码共同使用。

| 路径 | 职责 |
|------|------|
| `pluginbridge/` | 命名空间包，说明桥接协议和`guest`辅助能力位于子包中。 |
| `pluginbridge/contract/` | 稳定桥接`ABI`常量、路由契约、请求和响应信封、身份快照、执行来源值、生命周期契约、定时任务契约和校验辅助能力。 |
| `pluginbridge/guest/` | `guest`运行时辅助能力，用于导出的内存分配和执行函数、路由分发、请求绑定、路由匹配、`guest`侧宿主服务客户端和原始宿主调用传输。 |
| `pluginbridge/protocol/` | 公开低层协议门面，负责桥接信封、产物元数据、宿主调用载荷、`hostServices`载荷、编解码、能力常量和公开`hostServices`目录。 |

工具若需要语言无关的`hostServices`目录，应使用`pluginbridge/protocol.HostServiceDescriptors()`。它是开发者和工具用于服务与方法发现的公开来源。

## `pluginhost`职责

`pluginhost`是源码插件契约区域。源码插件随宿主编译，并通过分组 facade 注册自己的后端贡献。

| 区域 | 职责 |
|------|------|
| `SourcePlugin` | 根源码插件分组注册契约。提供`Assets()`、`Lifecycle()`、`Hooks()`、`HTTP()`、`Cron()`和`Governance()`。 |
| `Services` | 传入源码插件注册和回调流程的运行时服务目录。它嵌入普通`capability.Services`，并增加源码插件专属的`Admin()`和`TenantFilter()`接缝。 |
| 资产注册 | `UseEmbeddedFiles`绑定插件拥有的嵌入文件，使宿主可以提供`manifest`和公开资产。 |
| 生命周期注册 | 为安装、升级、禁用、卸载、租户禁用、租户删除和安装模式变更注册前置条件、自定义升级、清理和后置通知回调。 |
| 钩子注册 | 为已发布后端扩展点注册回调式钩子处理器。 |
| `HTTP`注册 | 在插件`API`命名空间下注册源码插件路由，并捕获路由绑定供宿主治理使用。 |
| 定时任务注册 | 注册受保护的定时任务，执行时检查插件启用状态，并提供主节点判断能力。 |
| 治理注册 | 注册宿主治理流水线使用的菜单和权限过滤器。 |

已发布后端扩展点包括`auth.login.succeeded`、`auth.login.failed`、`auth.logout.succeeded`、`plugin.installed`、`plugin.enabled`、`plugin.disabled`、`plugin.uninstalled`、`plugin.upgraded`、`system.started`、`http.route.register`、`cron.register`、`menu.filter`和`permission.filter`。

## `plugin.yaml`中的`hostServices`

动态插件通过`plugin.yaml`中的`hostServices`声明宿主访问。每条声明包含服务名、方法名，以及该服务要求的资源声明形态。

示例：

```yaml
hostServices:
  - service: runtime
    methods:
      - log.write
      - state.get
      - state.set
  - service: storage
    methods:
      - put
      - get
      - list
    paths:
      - exports/
  - service: data
    methods:
      - list
      - get
    tables:
      - plugin_demo_reports
  - service: network
    methods:
      - request
    resources:
      - ref: https://api.example.com/v1/*
  - service: ai
    methods:
      - text.generate
    resources:
      - ref: purpose:report.summary
        attributes:
          defaultTier: standard
```

资源声明形态：

| 资源类型 | 声明字段 | 服务 |
|----------|----------|------|
| `none` | 不声明`paths`、`tables`、`keys`或`resources`。 | `runtime`、`cron`、`config`、`org`、`tenant` |
| `path` | `paths` | `storage`、`manifest` |
| `table` | `tables` | `data` |
| `key` | `keys` | `hostconfig` |
| `resource` | `resources[].ref`以及服务专属治理字段。 | `network`、`cache`、`lock`、`notify`、`ai` |

生产校验会要求`data`服务表属于插件自有命名空间。动态插件不得声明`sys_*`这类宿主核心表。

`network`资源使用已授权的`http`或`https` URL pattern。`ai`资源使用`purpose:<name>`引用，并可包含`defaultTier`、`maxOutputTokens`、`maxPayloadBytes`、`maxInputAssets`、`maxOutputAssets`、`maxAssetBytes`、`allowOperation`、`allowOperationCancel`和`allowedMimeTypes`等受治理属性。

`config`、`hostconfig`和`manifest`在省略方法时默认使用`get`。动态`guest`配置辅助方法，例如`Exists`、`String`、`Bool`、`Int`和`Duration`，都会映射到`config.get`；`plugin.yaml`中仍应声明`config.get`。

## 动态插件`hostServices`目录

本章节列出动态插件可在`plugin.yaml`中声明的`hostServices`服务名、各服务下的方法名，以及每个方法的用途。机器可读的公开来源是`pluginbridge/protocol.HostServiceDescriptors()`。

### 服务总览

| 服务 | 资源类型 | 用途 |
|------|----------|------|
| `runtime` | `none` | 运行时日志、插件作用域状态、宿主时间、宿主生成标识和节点身份。 |
| `cron` | `none` | 宿主侧发现阶段的动态插件定时任务声明。 |
| `storage` | `path` | 已授权逻辑路径下受治理的插件存储对象操作。 |
| `network` | `resource` | 面向已授权 URL pattern 的受治理出站`HTTP`请求。 |
| `data` | `table` | 面向插件自有授权表的受治理读取和变更。 |
| `cache` | `resource` | 受治理的缓存读取、写入、整数递增和过期策略更新。 |
| `lock` | `resource` | 受治理的分布式锁获取、续期和释放。 |
| `notify` | `resource` | 受治理的通知消息发送。 |
| `config` | `none` | 只读访问当前插件运行时配置。 |
| `hostconfig` | `key` | 只读访问显式授权的宿主配置键。 |
| `manifest` | `path` | 只读访问插件作用域`manifest/`资源。 |
| `ai` | `resource` | 通过`purpose:<name>`资源授权的受治理类型化`AI`调用。 |
| `org` | `none` | 组织能力状态和用户组织投影。 |
| `tenant` | `none` | 租户能力状态、当前租户、可见性、成员关系和切换校验。 |

### 方法说明

#### `runtime`

| 方法 | 用途 |
|------|------|
| `log.write` | 写入一条插件结构化运行时日志。 |
| `state.get` | 读取一个插件作用域运行时状态值。 |
| `state.set` | 写入一个插件作用域运行时状态值。 |
| `state.delete` | 删除一个插件作用域运行时状态值。 |
| `info.now` | 返回宿主时间信息。 |
| `info.uuid` | 返回一个宿主生成的唯一标识。 |
| `info.node` | 返回宿主节点身份信息。 |

#### `cron`

| 方法 | 用途 |
|------|------|
| `register` | 向当前宿主侧发现收集器注册一个动态插件定时任务契约。 |

#### `storage`

| 方法 | 用途 |
|------|------|
| `put` | 写入一个受治理存储对象。 |
| `get` | 读取一个受治理存储对象。 |
| `delete` | 删除一个受治理存储对象。 |
| `list` | 列出某个已授权前缀下的受治理存储对象。 |
| `stat` | 读取一个受治理存储对象的元数据。 |

#### `network`

| 方法 | 用途 |
|------|------|
| `request` | 执行一次受治理出站`HTTP`请求。 |

#### `data`

| 方法 | 用途 |
|------|------|
| `list` | 对已授权插件自有表执行一次受治理分页列表查询。 |
| `get` | 从已授权插件自有表按键读取一条受治理记录。 |
| `create` | 在已授权插件自有表中创建一条受治理记录。 |
| `update` | 更新已授权插件自有表中的一条受治理记录。 |
| `delete` | 删除已授权插件自有表中的一条受治理记录。 |
| `transaction` | 对结构化数据变更执行一次受治理事务。 |

#### `cache`

| 方法 | 用途 |
|------|------|
| `get` | 读取一个受治理缓存值。 |
| `set` | 写入一个受治理缓存值。 |
| `delete` | 移除一个受治理缓存值。 |
| `incr` | 对一个受治理整数缓存值执行递增。 |
| `expire` | 更新一个受治理缓存项的过期策略。 |

#### `lock`

| 方法 | 用途 |
|------|------|
| `acquire` | 获取一个受治理分布式锁。 |
| `renew` | 续期一个受治理分布式锁。 |
| `release` | 释放一个受治理分布式锁。 |

#### `notify`

| 方法 | 用途 |
|------|------|
| `send` | 发送一条受治理通知消息。 |

#### `config`

| 方法 | 用途 |
|------|------|
| `get` | 以`JSON`形式读取一个当前插件配置值。 |

`config`在`plugin.yaml`中只发布`get`。`Exists`、`String`、`Bool`、`Int`和`Duration`等`guest`辅助方法只是`config.get`之上的便捷适配。

#### `hostconfig`

| 方法 | 用途 |
|------|------|
| `get` | 读取一个已授权宿主配置值。 |

#### `manifest`

| 方法 | 用途 |
|------|------|
| `get` | 读取一个插件作用域清单资源。 |

#### `ai`

| 方法 | 用途 |
|------|------|
| `text.generate` | 执行一次受治理文本生成请求。 |
| `image.generate` | 执行一次受治理图片生成请求。 |
| `image.edit` | 执行一次受治理图片编辑请求。 |
| `embedding.create` | 执行一次受治理向量嵌入请求。 |
| `audio.transcribe` | 执行一次受治理音频转写请求。 |
| `audio.synthesize` | 执行一次受治理语音合成请求。 |
| `vision.analyze` | 执行一次受治理视觉分析请求。 |
| `document.analyze` | 执行一次受治理文档分析请求。 |
| `document.cite` | 执行一次受治理带引用文档请求。 |
| `safety.moderate` | 执行一次受治理安全审核请求。 |
| `video.generate` | 执行一次受治理视频生成请求。 |
| `video.edit` | 执行一次受治理视频编辑请求。 |
| `video.extend` | 执行一次受治理视频延展请求。 |
| `video.operation.get` | 读取一个受治理提供方操作。 |
| `video.operation.cancel` | 取消一个受治理提供方操作。 |

#### `org`

| 方法 | 用途 |
|------|------|
| `capability.available` | 判断组织能力是否可用。 |
| `capability.status` | 读取组织能力状态。 |
| `users.dept_assignments.list` | 批量列出用户部门归属。 |
| `users.dept_info.get` | 读取单个用户的部门标识和名称。 |
| `users.dept_name.get` | 读取单个用户的部门名称。 |
| `users.dept_ids.get` | 读取单个用户的部门标识集合。 |
| `users.post_ids.get` | 读取单个用户的岗位标识集合。 |

#### `tenant`

| 方法 | 用途 |
|------|------|
| `capability.available` | 判断租户能力是否可用。 |
| `capability.status` | 读取租户能力状态。 |
| `tenants.current` | 读取当前请求租户。 |
| `tenants.platform_bypass` | 判断租户过滤是否可以被绕过。 |
| `tenants.visible.ensure` | 校验当前用户是否可以访问指定租户。 |
| `users.tenant_membership.validate` | 校验单个用户的租户成员关系。 |
| `users.tenants.list` | 列出单个用户可见的租户。 |
| `tenants.switch.validate` | 校验一个租户切换目标。 |

当前还存在`secret.resolve`、`event.publish`和`queue.enqueue`预留治理条目。它们属于描述符目录，用于后续治理对齐，但不是已发布的`guest`可调用方法；在宿主分发器和`guest SDK`支持补齐前，不应把它们用于可执行动态插件调用。

## 开发者指南

- 源码插件或宿主包需要普通读取型插件可见能力时，使用`capability.Services`。
- 只有可信源码插件管理命令才使用`capability.AdminServices`。依赖应保持窄接口，并通过领域服务传递`CapabilityContext`。
- 动态插件`guest`代码中使用`pluginbridge/guest.Default()`或`pluginbridge/guest.New()`获取`runtime`、`storage`、`network`、`recordstore`、`cache`、`lock`、`config`、`notify`、`cron`、`hostconfig`、`manifest`、`org`、`tenant`、`plugin`和`AI`客户端。
- 源码插件 registrar 使用`pluginhost.SourcePlugin`声明嵌入文件、路由、生命周期回调、钩子、定时任务和治理过滤器。
- 动态路由和产物契约使用`pluginbridge/contract`，`guest`路由执行使用`pluginbridge/guest`，低层载荷或需要公开协议目录的工具使用`pluginbridge/protocol`。
- 新增能力契约应保持存储无关，高频读取应优先批量化，并明确数据可见性。不得新增要求插件了解宿主表名或宿主缓存键的能力方法。
