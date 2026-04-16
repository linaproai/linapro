# plugin-hook-slot-extension Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
### Requirement: 宿主发布稳定的后端扩展点契约
系统 SHALL 发布一组具名、版本化、可审计的后端扩展点，供插件在宿主关键业务事件或宿主治理链路上注册回调并扩展行为。

#### Scenario: 宿主维护正式的后端扩展点目录
- **WHEN** 宿主对外发布插件后端扩展能力
- **THEN** 宿主必须维护正式的后端扩展点目录
- **AND** 一期至少公开 `auth.login.succeeded`、`auth.logout.succeeded`、`system.started`、`plugin.installed`、`plugin.enabled`、`plugin.disabled`、`plugin.uninstalled`
- **AND** 每个扩展点都必须说明触发时机、上下文、执行顺序、执行模式与失败隔离策略

#### Scenario: 登录成功事件触发 Hook
- **WHEN** 用户登录成功且宿主发布 `auth.login.succeeded` Hook
- **THEN** 宿主按约定上下文向已启用插件分发该事件
- **AND** 上下文至少包含用户标识、登录时间、客户端信息、请求上下文与当前插件运行代际信息

#### Scenario: 登出成功事件触发 Hook
- **WHEN** 用户登出成功且宿主发布 `auth.logout.succeeded` Hook
- **THEN** 宿主向订阅该 Hook 的已启用插件分发事件
- **AND** 插件只能读取宿主公开的上下文字段

### Requirement: Hook 执行失败必须与主流程隔离
系统 SHALL 对插件 Hook 的超时、异常和返回错误实施隔离，不得让插件扩展破坏宿主主链路。

#### Scenario: 插件 Hook 执行失败
- **WHEN** 某插件在登录成功 Hook 中超时、崩溃或返回错误
- **THEN** 用户登录主流程仍然返回成功
- **AND** 宿主记录该插件的执行失败信息
- **AND** 其他插件的 Hook 仍按顺序继续执行或按策略被安全跳过

#### Scenario: 当前生效动态 release 的 Hook 契约持续可用
- **WHEN** 一个动态插件已经切换到某个 active release，随后又出现 staged 升级、失败回滚或 active release 重载
- **THEN** 宿主仍然按当前 active release 内嵌声明的 Hook 契约分发事件
- **AND** 宿主不会因为只恢复了顶层 manifest 元数据而丢失该 release 的 Hook 声明

### Requirement: 宿主以回调注册方式发布通用后端扩展点
系统 SHALL 为源码插件提供最小化、回调注册式的后端扩展接口，避免插件作者为常见扩展场景维护复杂声明式属性。

#### Scenario: 事件 Hook 与注册式接口属于同一类后端扩展点
- **WHEN** 宿主同时发布事件型 Hook 与注册型回调扩展点
- **THEN** 这两类能力都必须被视为“宿主已发布后端扩展点上的回调注册”
- **AND** 插件开发者统一通过 Go 类型常量选择扩展点并注册回调函数
- **AND** 宿主只允许通过执行模式决定回调是否阻塞主流程或异步执行

#### Scenario: 宿主维护正式的后端回调扩展点目录
- **WHEN** 宿主对外发布源码插件后端扩展能力
- **THEN** 宿主必须提供统一的 Go 注册入口与回调注册方法
- **AND** 一期至少提供 `http.route.register`、`http.request.after-auth`、`cron.register`、`menu.filter`、`permission.filter`
- **AND** 这些扩展点必须与已发布 Hook 一样纳入技术文档维护

#### Scenario: 宿主以统一的 Go 类型目录管理所有后端扩展点
- **WHEN** 宿主同时发布事件型 Hook 与注册型回调扩展点
- **THEN** 宿主必须使用统一的 Go `type` 与常量目录维护这些后端扩展点
- **AND** 插件代码、宿主调度代码与技术文档都必须引用同一组类型常量
- **AND** 不允许在宿主实现或插件示例中散落硬编码的后端扩展点字符串

#### Scenario: 宿主为后端回调扩展点声明执行模式
- **WHEN** 插件向宿主注册某个后端扩展点回调
- **THEN** 注册接口必须显式声明该回调的执行模式
- **AND** 执行模式至少区分 `blocking` 与 `async`
- **AND** 宿主必须校验当前扩展点是否支持所声明的执行模式
- **AND** 不支持的执行模式必须在注册阶段被拒绝

#### Scenario: 宿主以接口类型暴露回调输入对象
- **WHEN** 宿主向插件暴露 Hook、After-Auth、Route、Cron、菜单过滤或权限过滤这类回调输入对象
- **THEN** 宿主必须优先暴露抽象接口而不是具体结构体指针
- **AND** 插件回调只依赖宿主公开的方法契约
- **AND** 宿主后续扩展字段或能力时不应要求插件直接耦合内部结构体实现

#### Scenario: 插件通过回调注册受宿主治理的 HTTP 路由
- **WHEN** 一个源码插件注册自己的 HTTP 路由
- **THEN** 宿主在启动时自动装配该路由
- **AND** 插件被禁用后，这些路由请求会被宿主拒绝或降级
- **AND** 插件作者无需手工修改宿主控制器或路由骨架代码
- **AND** 插件作者仅需要在 `apps/lina-plugins/lina-plugins.go` 中维护插件后端包的显式导入关系

#### Scenario: 宿主为插件提供独立的无前缀路由分组
- **WHEN** 宿主向源码插件开放 HTTP 路由注册能力
- **THEN** 宿主必须提供独立于主服务 `/api/v1` 分组的插件路由分组
- **AND** 该插件路由分组本身不得内置任何固定路由前缀
- **AND** 插件可以自行选择是否挂载到 `/api/v1`、其他业务前缀或无前缀路径

#### Scenario: 宿主向插件公开可选的主服务中间件目录
- **WHEN** 宿主向源码插件开放 HTTP 路由注册能力
- **THEN** 宿主必须向插件公开已发布的主服务中间件目录
- **AND** 一期至少包括 `NeverDoneCtx`、`HandlerResponse`、`CORS`、`Ctx`、`Auth`、`OperLog`
- **AND** 插件可以按自身路由分组需要选择任意子集并决定组合顺序
- **AND** 插件也可以将宿主中间件与插件自定义中间件组合使用

#### Scenario: 插件可拆分多个治理策略不同的路由分组
- **WHEN** 同一个源码插件既需要公开免鉴权接口，又需要公开受保护接口
- **THEN** 插件必须可以在一次路由注册回调中声明多个独立路由分组
- **AND** 路由分组注册方式必须与宿主主服务保持一致，支持 `group.Group(prefix, func(group *ghttp.RouterGroup) { ... })` 风格
- **AND** 每个路由分组都可以自行选择宿主已发布中间件的任意子集与组合顺序
- **AND** 每个路由分组都可以继续追加自己的子路径前缀

#### Scenario: 插件在鉴权完成后参与请求扩展
- **WHEN** 一个受保护请求通过 JWT 鉴权并完成用户上下文注入
- **THEN** 宿主向已启用插件分发 `http.request.after-auth` 回调
- **AND** 插件可以读取宿主公开的用户身份与请求对象
- **AND** 插件失败不会中断当前请求主流程

#### Scenario: 插件注册受宿主启停控制的定时任务
- **WHEN** 一个源码插件注册自己的定时任务
- **THEN** 宿主通过统一的 `cron` 组件完成注册
- **AND** 插件被禁用后，宿主不会执行该插件的定时任务回调
- **AND** 插件作者无需在宿主 `cmd` 层手工追加定时任务代码

#### Scenario: 定时任务注册器暴露主节点识别能力
- **WHEN** 宿主向插件暴露定时任务注册输入对象
- **THEN** 该对象必须提供“当前节点是否为主节点”的识别方法
- **AND** 插件可以基于该方法决定是否执行仅主节点生效的定时逻辑

#### Scenario: 插件按菜单与权限过滤器参与宿主治理链路
- **WHEN** 宿主生成菜单列表或权限列表
- **THEN** 宿主向已启用插件发布 `menu.filter` 与 `permission.filter` 回调
- **AND** 插件只能对宿主公开的菜单/权限描述进行过滤判断
- **AND** 过滤失败不会破坏宿主原有菜单与权限计算流程

### Requirement: 宿主发布前端 Slot 扩展点
系统 SHALL 为前端页面和布局发布受控的 Slot 扩展点，允许插件在宿主公开位置插入 UI 内容。

#### Scenario: 宿主维护正式的前端 Slot 插槽目录
- **WHEN** 宿主对外发布前端 Slot 能力
- **THEN** 宿主必须维护正式的前端 Slot 插槽目录
- **AND** 一期至少公开 `layout.user-dropdown.after` 与 `dashboard.workspace.after`
- **AND** 每个插槽都必须说明宿主位置、渲染容器、推荐用途、排序规则与失败降级策略

### Requirement: 宿主优先在公共界面层发布通用前端 Slot
系统 SHALL 优先在布局界面、登录界面、工作台界面与 CRUD 通用界面发布通用前端 Slot，避免把扩展点绑定到单一业务模块。

#### Scenario: 宿主发布首批通用公共 Slot
- **WHEN** 宿主扩充前端 Slot 能力
- **THEN** 一期至少新增 `layout.header.actions.before`、`layout.header.actions.after`、`auth.login.after`、`dashboard.workspace.before`、`crud.toolbar.after`、`crud.table.after`
- **AND** 这些 Slot 必须依附已有公共界面层而不是某个业务模块私有页面
- **AND** 开发文档必须同步说明每个 Slot 的宿主位置与推荐内容

#### Scenario: 插件在登录页公开区插入内容
- **WHEN** 一个已启用插件声明向 `auth.login.after` 插入前端内容
- **THEN** 宿主在登录页表单之后渲染该内容
- **AND** 插件被禁用后该内容立即隐藏

#### Scenario: 插件在 CRUD 通用界面插入内容
- **WHEN** 一个已启用插件声明向 `crud.toolbar.after` 或 `crud.table.after` 插入前端内容
- **THEN** 宿主在通用 Grid 工具栏区或表格区下方渲染该内容
- **AND** 所有复用该通用界面的页面都能自动获得扩展位

#### Scenario: 插件向宿主布局插入内容
- **WHEN** 一个已启用插件声明向 `layout.user-dropdown.after` 插入前端内容
- **THEN** 宿主在该 Slot 对应位置尝试加载插件声明的前端入口
- **AND** 源码插件的 Slot 内容必须来自真实前端源码文件，而不是仅依赖声明式 JSON 配置
- **AND** 这些源码文件默认放在 `frontend/slots/` 目录下，并由宿主在构建时发现和挂载
- **AND** 插件内容只在宿主公开的容器范围内渲染
- **AND** 插件不能越权访问未公开的宿主内部实现

#### Scenario: 插件向右上角菜单栏插入页面入口
- **WHEN** 一个已启用插件声明向 `layout.user-dropdown.after` 插入插件菜单入口
- **THEN** 宿主在右上角菜单栏展示该入口文案
- **AND** 点击该入口后宿主以内页导航方式打开插件 Tab 页面
- **AND** 该过程不会触发整页刷新

#### Scenario: 登录态在线启用插件后立即激活右上角入口路由
- **WHEN** 管理员在当前已登录会话中启用一个会向 `layout.user-dropdown.after` 注入入口的插件
- **THEN** 宿主无需重新登录即可同步刷新该入口对应的动态路由
- **AND** 用户点击该入口后不会进入 404 页面
- **AND** 宿主直接以内页 Tab 方式打开插件页面

#### Scenario: 当前会话在插件状态变化后重新获得焦点
- **WHEN** 当前已登录会话之外的其他操作改变了会注入 `layout.user-dropdown.after` 的插件状态，且当前标签页重新获得焦点
- **THEN** 宿主自动同步该 Slot 的可见性与对应动态路由
- **AND** 已启用插件的右上角入口重新显示且可正常打开
- **AND** 已禁用插件的右上角入口及时隐藏

#### Scenario: 插件 Slot 契约不匹配
- **WHEN** 插件声明的前端入口与 Slot 所要求的契约不兼容
- **THEN** 宿主跳过该插件内容渲染
- **AND** 宿主记录契约不匹配错误
- **AND** 当前页面其他宿主内容正常渲染

### Requirement: Hook 与 Slot 执行顺序可预测
系统 SHALL 为同一 Hook 或 Slot 上的多个插件定义稳定的执行顺序。

#### Scenario: 多个插件订阅同一 Hook
- **WHEN** 多个插件同时订阅同一个后端 Hook 或前端 Slot
- **THEN** 宿主按照 manifest 显式优先级或统一的默认排序规则执行
- **AND** 相同输入下的执行顺序在各节点保持一致

### Requirement: Hook 与 Slot 标识必须使用专门类型定义
系统 SHALL 使用专门的类型定义合法的插件安装位置，避免在宿主实现和插件示例中散落硬编码字符串。

#### Scenario: 后端扩展点在 Go 中声明
- **WHEN** 宿主实现后端 Hook 插槽或注册式回调扩展点
- **THEN** 宿主必须使用 Go `type` 与常量声明合法后端扩展点标识
- **AND** 插件后端示例通过该类型常量引用扩展点，而不是直接写事件名字符串
- **AND** 宿主内部服务层不得再额外维护一套语义重复的别名常量

#### Scenario: 前端 Slot 插槽在 TypeScript 中声明
- **WHEN** 宿主实现前端 Slot 插槽
- **THEN** 宿主必须使用 TypeScript 常量与类型声明合法 Slot 标识
- **AND** 宿主页面、Slot 装载器与插件前端示例通过统一类型引用插槽，而不是直接写 Slot 名字符串

#### Scenario: 插件声明未知插槽
- **WHEN** 插件声明一个未被宿主发布的 Hook 或 Slot 标识
- **THEN** 宿主拒绝该声明或跳过装载
- **AND** 宿主记录“插槽未发布或契约不支持”的错误信息

### Requirement: 宿主提供面向插件开发者的插槽技术文档
系统 SHALL 将前后端插槽目录、类型定义与示例用法沉淀到插件开发者可直接查阅的技术文档中。

#### Scenario: 发布插件开发指南
- **WHEN** 宿主新增、调整或正式发布 Hook/Slot 插槽
- **THEN** 宿主同步更新 `apps/lina-plugins/README.md`
- **AND** 文档中明确区分“已发布插槽”和“后续规划插槽”
- **AND** 文档中给出 Go 与前端源码插件的推荐引用方式

### Requirement: 动态插件路由治理元数据集中在`g.Meta`

系统 SHALL 要求动态插件将后端动态路由的治理元数据集中定义在`api`层请求结构体的`g.Meta`中，避免额外引入第二套分散的路由治理配置源。

#### Scenario: 动态插件声明最小治理字段

- **WHEN** 开发者定义一个动态插件后端接口
- **THEN** 该接口可在`g.Meta`中声明`access`、`permission`、`operLog`
- **AND** `access`仅支持`public`和`login`
- **AND** 未声明`access`时按`login`处理

#### Scenario: 公开路由治理边界受限

- **WHEN** 开发者声明一个`public`动态路由
- **THEN** 该路由不得声明`permission`
- **AND** 该路由不得依赖宿主登录态注入
- **AND** 宿主装载阶段会拒绝非法配置

### Requirement: 动态插件权限声明复用宿主现有权限体系

系统 SHALL 将动态路由中的`permission`声明自动接入宿主现有的`sys_menu.perms`权限体系，而不是引入独立的动态权限存储模型。

#### Scenario: 动态路由权限被物化为隐藏菜单项

- **WHEN** 一个动态路由声明了合法的`permission`
- **THEN** 宿主在插件菜单同步阶段自动生成对应的隐藏权限菜单项
- **AND** 这些权限菜单项挂载在该插件专属的动态路由权限目录下
- **AND** 权限值直接复用动态路由声明的`permission`

#### Scenario: 动态路由权限随插件生命周期同步

- **WHEN** 动态插件被启用、禁用、卸载或切换激活版本
- **THEN** 宿主同步新增、更新或移除对应的隐藏权限菜单项
- **AND** 默认管理员角色继续自动拥有这些权限项

### Requirement: 动态插件不直接组合宿主治理中间件

系统 SHALL 保持动态插件为受限业务扩展模型，不向动态插件开放宿主`Auth`、`Ctx`、`OperLog`等治理中间件的自由拼装能力。

#### Scenario: 动态治理由宿主统一执行

- **WHEN** 一个动态插件路由被外部请求命中
- **THEN** 登录校验、权限校验和业务上下文注入由宿主基于路由合同统一执行
- **AND** 动态插件只声明治理需求，不直接调用宿主治理中间件

### Requirement: 动态插件复用公共 bridge 组件降低编写复杂度

系统 SHALL 将动态插件 bridge envelope、二进制 codec、guest 侧处理器适配、错误响应辅助等可复用逻辑抽象到`apps/lina-core/pkg`公共组件中，避免插件作者在每个动态插件中重复编写底层`ABI`与编解码样板。

#### Scenario: 插件运行时复用公共组件

- **WHEN** 开发者编写`backend/runtime/wasm`动态插件运行时
- **THEN** 开发者可以复用`apps/lina-core/pkg/pluginbridge`中的请求／响应信封、二进制编解码和处理器适配逻辑
- **AND** 插件业务代码主要实现路由处理函数，不需要重复实现底层内存读写与 bridge envelope 打包流程

#### Scenario: 公共组件不包含编译阶段流程

- **WHEN** 宿主、构建器或动态插件样例复用`apps/lina-core/pkg/pluginbridge`
- **THEN** 该组件只提供稳定合同、编解码、运行时辅助与无副作用校验逻辑
- **AND** 该组件不包含源码扫描、`go build`调用或产物写入等编译阶段流程

