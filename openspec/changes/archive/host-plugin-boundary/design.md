# Design

## Host Boundary

`apps/lina-core`只保留框架核心和管理基座：认证、JWT、登录态、用户、角色、菜单、插件注册与生命周期治理、任务调度平台、配置、字典、文件基础能力、统一事件、稳定顶级菜单目录和插件扩展 seam。新增管理后台能力时，必须先判断是否承担框架级复用和统一治理责任；不属于宿主基座的能力优先按源码插件设计。

宿主不能长期持有源码插件业务表、插件本地 ORM 产物、插件 mock 数据或插件业务表直接查询逻辑。官方源码插件在自己的`backend/`目录维护`hack/config.yaml`、`internal/dao`、`internal/model/do`和`internal/model/entity`；插件业务表使用`plugin_<plugin_id_snake_case>`命名空间，安装、mock 和卸载资源由插件生命周期负责。

## Stable Menu Mounts

宿主维护 9 个真实一层目录记录：`dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer`。这些目录是插件`parent_key`的稳定锚点，不能由插件动态创建新的同级目录系统。

组织插件挂载到`org`，内容插件挂载到`content`，监控插件挂载到`monitor`，插件治理入口仍留在`extension`。当目录下没有任何可见子菜单时，导航投影层隐藏空目录，但数据库中的宿主稳定目录记录继续存在。

## Source Plugin Decoupling

组织、内容和监控能力按官方源码插件交付：`org-center`承载部门和岗位，`content-notice`承载公告通知，`monitor-online`承载在线用户查询和强制下线，`monitor-server`承载服务监控，`monitor-operlog`承载操作日志，`monitor-loginlog`承载登录日志。

宿主保留认证会话真源、`sys_online_session`、登录/登出、受保护接口认证、会话超时判断和清理任务。`monitor-online`只能消费宿主提供的会话投影和管理能力，不能接管 JWT 校验、`last_active_time`维护或清理真源。

登录日志和操作日志通过“宿主事件 + 插件订阅持久化”解耦。宿主认证链路发布登录成功、登录失败和登出事件；请求链路发布统一审计事件；日志插件缺失时宿主核心链路继续执行，不因缺少具体日志持久化实现而失败。

## Capability Seams

宿主与插件通过稳定 seam 协作，而不是在宿主业务代码中散落插件状态判断。组织能力通过宿主拥有的接口、DTO、空实现和 Provider 注册接入；登录和审计通过事件 Hook 接入；源码插件通过 HTTP 路由注册、全局中间件注册、Cron 注册、菜单过滤和权限过滤接入。

HTTP seam 必须给源码插件提供统一注册入口，包含 route registrar 和 global middleware registrar。插件不直接持有裸`*ghttp.Server`，也不需要修改宿主 controller 或路由骨架。插件可以声明多个路由分组，自行组合宿主发布的主服务中间件和插件自定义中间件；禁用后宿主通过运行时开关旁路插件注册的全局中间件。

Cron seam 允许插件注册由宿主启停控制的定时任务，并暴露主节点识别能力。菜单和权限过滤 seam 只允许插件基于宿主暴露的描述做过滤和判定，过滤失败不能破坏宿主原有计算过程。

## Demo Control Guard

`demo-control`是官方源码插件，不是宿主硬编码开关。默认交付配置不启用它；当`plugin.autoEnable`包含`demo-control`时，宿主启动阶段安装并启用该插件，演示保护在`/*`请求链上生效。

启用后，`GET`、`HEAD`、`OPTIONS`继续放行；`POST`、`PUT`、`DELETE`默认返回只读演示提示并终止后续业务处理。插件保留最小会话白名单：`POST /api/v1/auth/login`和`POST /api/v1/auth/logout`；同时保留对其他插件的安装、卸载、启用、禁用白名单，但禁止改变`demo-control`自身治理状态。这个策略依赖项目统一的 RESTful Method 语义。

## Pluginbridge Subcomponents

`pkg/pluginbridge`子组件化是宿主边界治理中的插件公共包样板，但完整桥接协议、ABI、transport、guest SDK和host service wire契约由插件框架分组长期承载。这里保留的是边界设计原因：桥接层只能表达动态插件传输和协议职责，不能重新拥有宿主业务能力语义。

- `contract`：ABI、路由、Cron、执行来源和稳定 DTO。
- `codec`：桥接请求、响应、路由、身份和 HTTP 快照 envelope 编解码。
- `artifact`：WASM section 常量、custom section 读取和运行时 artifact 元数据。
- `hostcall`：host call opcode、状态码和通用 envelope。
- `hostservice`：host service spec、capability 派生、manifest codec 和各类 host service payload codec。
- `guest`：guest runtime、typed controller dispatcher、`BindJSON`、`WriteJSON`、错误分类和 host service client。

根包`pluginbridge`只保留 facade、文档和必要兼容入口，通过类型别名、常量别名和 wrapper 委托到子组件，不复制协议实现。宿主内部优先使用精确子组件导入；动态插件 guest 代码可以继续使用根包兼容入口。

拆分不得改变桥接协议行为。ABI 常量、WASM custom section 名、protobuf 字段编号、host call 状态码、host service 方法字符串、payload codec 结果和 guest helper 行为必须保持不变，并通过`go test ./pkg/pluginbridge/...`、host runtime/wasm/plugindb 测试和动态插件样例 build 验证。

## Cross-Domain Impacts

- `menu-management`承载稳定一层目录、插件语义挂载和空目录隐藏的最终菜单契约，当前契约由`openspec/specs/menu-management/spec.md`承载，历史 owner 为`archive/user-management`。
- `user-management`承载组织插件缺失时的部门/岗位字段、筛选、树选择器和详情降级，当前契约由`openspec/specs/user-management/spec.md`承载，历史 owner 为`archive/user-management`。
- `dept-management`和`post-management`承载`org-center`中的部门与岗位能力，当前契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 为`archive/org-structure`。
- `notice-management`承载`content-notice`公告通知能力，当前契约由`openspec/specs/notice-management/spec.md`承载，历史 owner 为`archive/notification`。
- `online-user`、`server-monitor`、`oper-log`和`login-log`承载监控插件能力，当前契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 为`archive/system-governance`。
- `user-auth`承载登录、登出、JWT 和认证中间件契约，当前契约由`openspec/specs/user-auth/spec.md`承载，历史 owner 为`archive/user-auth`。
- `config-management`承载 TraceID 静态配置和运行时配置治理，当前契约由`openspec/specs/config-management/spec.md`承载，历史 owner 为`archive/system-config`。
- `plugin-manifest-lifecycle`、`plugin-runtime-loading`、`plugin-hook-slot-extension`承载插件 ID、挂载规则、生命周期、WASM section 和 hook 能力，当前契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 为`archive/plugin-framework`。
- `pluginbridge-subcomponent-architecture`承载动态插件桥接子组件、协议出口、guest SDK 和 ABI 行为，当前契约由`openspec/specs/pluginbridge-subcomponent-architecture/spec.md`承载，历史 owner 为`archive/plugin-framework`。

## Risks And Boundaries

- 组织、内容和监控拆成多个源码插件会增加文档、测试和样例维护成本；通过统一插件模板、菜单挂载规则和 E2E 规范降低长期成本。
- `monitor-online`不能接管认证主链路；宿主会话内核必须独立可运行，否则插件缺失会破坏登录和鉴权。
- Method-based demo policy 依赖 RESTful 约束；违反 Method 语义的接口会被错误放行或拦截，因此 API 合同治理必须持续执行。
- `pluginbridge`子组件拆分可能引入 import cycle 或 facade 遗漏；通过分层迁移、禁止低层导入根包和 facade 一致性测试控制风险。
