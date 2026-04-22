## Overview

本次变更以“开源阶段宿主瘦身”为目标，采用“宿主稳定挂载点 + 官方源码插件按语义落位”的设计，而不是将所有扩展能力统一堆入插件中心。宿主负责提供框架内核、治理底座和稳定导航骨架；组织、内容、监控等非核心模块作为源码插件交付，并通过插件生命周期参与菜单、路由、权限和前端页面装配。

## Goals

- 让宿主边界清晰可读：开发者能一眼知道哪些是框架内核，哪些是可选模块。
- 保持后台主场景易用：一级菜单命名继续沿用管理后台通用认知，而不是抽象平台术语。
- 支持源码插件独立安装、启用、停用、卸载，并在插件缺失时保证宿主平滑降级。
- 为后续继续新增官方插件和业务模块提供稳定菜单挂载点，不再频繁改一级目录。

## Non-Goals

- 本次不引入商业插件市场、签名授权、计费分发等平台化能力。
- 本次不将任务调度能力插件化。
- 本次不要求一次性完成所有代码迁移实现；重点先固定边界、菜单与迁移顺序。

## Menu Architecture

### Stable Top-Level Catalogs

宿主固定提供以下一级目录及稳定父级 `menu_key`，并将它们作为宿主拥有的真实 `sys_menu` 目录记录长期维护：

- `工作台` -> `dashboard`
- `权限管理` -> `iam`
- `组织管理` -> `org`
- `系统设置` -> `setting`
- `内容管理` -> `content`
- `系统监控` -> `monitor`
- `任务调度` -> `scheduler`
- `扩展中心` -> `extension`
- `开发中心` -> `developer`

这些目录记录始终由宿主创建和拥有；空目录是否显示由菜单投影层决定，而不是依赖插件动态创建或删除父级目录。

### Menu Tree

```text
工作台
  - 分析页
  - 工作台
权限管理
  - 用户管理
  - 角色管理
  - 菜单管理
组织管理
  - 部门管理
  - 岗位管理
系统设置
  - 字典管理
  - 参数设置
  - 文件管理
内容管理
  - 通知公告
系统监控
  - 在线用户
  - 服务监控
  - 操作日志
  - 登录日志
任务调度
  - 任务管理
  - 分组管理
  - 执行日志
扩展中心
  - 插件管理
开发中心
  - 接口文档
  - 系统信息
```

### Navigation Rules

- 一级目录由宿主创建，不由插件动态新增。
- 插件功能菜单必须挂到语义对应的宿主目录下。
- `扩展中心` 仅展示插件治理入口，不挂实际业务菜单。
- 若某一级目录无任何可见子菜单，则该父目录在导航投影中自动隐藏，但宿主仍保留对应稳定目录记录。
- 插件状态变化后，宿主触发当前用户菜单与动态路由刷新，最终收敛到最新状态。

## Host Boundary

### Host-Retained Capabilities

以下能力必须保留在宿主：

- 认证、JWT、登录态解析、用户上下文。
- 用户管理、角色管理、菜单管理、权限校验。
- 插件注册表、安装/卸载、启停、菜单同步、治理资源同步。
- 字典能力、参数配置能力、文件基础能力。
- 任务调度平台能力。
- 宿主统一事件 / Hook 机制。
- 宿主拥有的一级目录记录与菜单投影治理逻辑。
- `扩展中心` 与 `开发中心` 提供的治理和开发辅助能力。

### Source-Plugin Capabilities

本次固定以下官方源码插件：

- `org-center`
  - 菜单挂载：`org`
  - 承载：部门管理、岗位管理。
- `content-notice`
  - 菜单挂载：`content`
  - 承载：通知公告。
- `monitor-online`
  - 菜单挂载：`monitor`
  - 承载：在线用户查询与强制下线治理。
- `monitor-server`
  - 菜单挂载：`monitor`
  - 承载：服务监控采集、清理与展示。
- `monitor-operlog`
  - 菜单挂载：`monitor`
  - 承载：操作日志查询、导出、清理与详情。
- `monitor-loginlog`
  - 菜单挂载：`monitor`
  - 承载：登录日志查询、导出、清理与详情。

## Critical Decoupling Strategy

### Organization Decoupling

当前用户管理直接依赖组织树、部门字段与岗位选项。若 `org-center` 未安装，宿主仍必须保证用户管理可用。因此需要先抽象组织能力提供者：

- 用户列表中的部门列改为可选展示。
- 用户表单中的部门树与岗位选择器改为基于组织插件能力探测按需显示。
- 用户详情中的组织信息改为可选扩展数据，不再视为宿主硬字段。
- 宿主业务逻辑只依赖“组织能力接口”，不直接依赖 `dept` / `post` service 实现。
- `orgcap` 作为宿主能力接缝，只保留宿主拥有的接口、DTO、空实现和调用入口；`org-center` 安装后通过稳定 Provider 注册其真实实现。
- 即使插件业务表与宿主共用同一个数据库，宿主也不得直接持有或查询 `org-center` 的物理表、ORM 工件和关联写入逻辑。
- `org-center` 缺失时，用户管理页面退化为无左树筛选、无部门/岗位字段的核心用户管理视图。

### Online Session vs Online User Plugin

`在线用户` 虽然要独立为源码插件，但宿主认证链路仍依赖在线会话有效性校验。当前认证中间件直接调用会话存储做 `TouchOrValidate`，因此：

- 宿主保留认证会话内核、`sys_online_session` 真相源以及会话存储抽象。
- 宿主继续负责登录创建会话、登出删除会话、请求触达时刷新活跃时间、超时判定与清理任务。
- 宿主对 `monitor-online` 发布独立的会话 DTO / filter / result 契约，避免直接把内部 `session` 类型别名暴露给插件。
- `monitor-online` 只负责读取会话投影、展示在线用户列表和执行强制下线治理。
- `monitor-online` 不拥有 JWT 校验、会话超时语义、`last_active_time` 维护或清理任务真相源。
- 插件未安装时，宿主仍然能正常登录、登出、校验会话超时。

### Login Log / Oper Log Eventization

当前登录日志与操作日志都由宿主核心链路直接依赖具体日志服务。为实现插件化，需要改为：

- 宿主定义统一登录事件契约，覆盖登录成功、登录失败和登出成功等场景。
- 宿主定义统一审计事件契约，覆盖写操作与带 `operLog` 标签的受审计查询操作。
- 宿主核心认证链路和请求链路只负责发射事件，不再直接依赖具体 `loginlog` / `operlog` 落库实现。
- `monitor-loginlog` 与 `monitor-operlog` 订阅事件后完成落库、查询、导出和清理。
- 插件未安装时，宿主事件发射链路仍可执行，但不强制落库，也不允许因此阻塞认证、鉴权或普通业务请求。

为避免宿主继续保留插件专用的操作日志中间件，同时又让源码插件拿到完整请求结果，需要进一步收敛 HTTP 接缝：

- 宿主在源码插件统一 HTTP 注册入口上同时发布路由注册器与全局中间件注册器，而不是为 `monitor-operlog` 单独保留专用宿主挂点。
- GoFrame HOOK 适合前后置观察，但不适合作为操作日志唯一实现，因为提前结束请求会跳过后置 HOOK，无法稳定获取完整完成态结果。
- `monitor-operlog` 通过宿主封装的全局中间件注册器注册自己的审计中间件，并复用宿主统一审计事件分发。
- 宿主对插件自注册的全局中间件统一包裹启停开关；插件停用时直接旁路其逻辑，不要求重建 HTTP 路由树。

### Capability Seams Instead of Placeholder Branches

本次拆分不接受“宿主保留完整业务骨架 + 在各处散落 `if pluginEnabled` 判断”的实现方式。宿主与插件之间需要采用更稳定、更整洁的接缝：

- 宿主保留统一的能力接口，例如组织能力 `orgcap`，由宿主核心模块只依赖接口、DTO 和空实现；插件安装后通过注册的 Provider/Adapter 提供真实能力实现。
- 宿主通过统一 Hook 事件对登录日志、操作日志等链路发射事件，而不是在认证和中间件链路中保留具体插件落库分支。
- 源码插件通过路由注册器和 Cron 注册器接管自己的 HTTP API 与定时任务，不要求宿主为这些插件预留静态绑定代码。
- 宿主仅保留真正属于框架内核的真相源与治理底座，例如 `sys_online_session` 会话真相源、插件生命周期、菜单治理和任务调度底座。
- 当插件缺失时，宿主通过能力接口的“零值 / 空集合 / 无额外字段”语义平滑降级，而不是到处判断后拼接不同代码路径。

### Plugin-Local ORM Generation

源码插件后端的数据库访问需要在插件目录内闭环，而不是再次回流到宿主 `dao/model` 工件或长期依赖散落的 `g.DB().Model(...)` 语句：

- 每个官方源码插件的 `backend/` 目录都维护独立的 `hack/config.yaml`，保证开发者进入插件后端目录后即可执行 `gf gen dao`。
- 插件本地生成并维护 `internal/dao`、`internal/model/do`、`internal/model/entity`，由插件 service 直接消费。
- 插件读取宿主共享表（如 `sys_user`、`sys_dict_data`）时，也通过插件本地 codegen 工件完成查询，避免重新依赖宿主 `dao/model` 包。
- 一旦某个业务表完成插件化迁移，宿主必须同步删除该表对应的 `dao/do/entity` 与直接查表实现，避免形成“双份 ORM 工件 + 双份存储入口”。
- 当前版本不直接查库的源码插件仍保留本地 codegen 配置，便于后续扩展时继续沿用统一结构。

### Plugin-Owned Storage Lifecycle and Naming

源码插件的数据存储必须在宿主边界之外清晰可辨，不能继续伪装成宿主内建业务表：

- 宿主 `make init` 仅初始化宿主内核表和必需 Seed 数据，不创建任何源码插件业务表。
- 宿主 `make mock` 不再写入插件业务表；插件演示数据应由插件安装 SQL、插件自有 mock 资源或插件专属命令装载。
- 官方源码插件的业务物理表统一使用插件作用域命名 `plugin_<plugin_id_snake_case>`；单表插件优先直接使用该完整表名，多表插件再按需追加业务后缀，例如 `plugin_org_center_dept`、`plugin_content_notice`、`plugin_monitor_loginlog`；`sys_` 前缀仅保留给宿主核心表。
- 插件若需读取宿主共享治理数据，只允许显式依赖宿主共享表（如 `sys_user`、`sys_dict_data`、`sys_online_session`），不得反向要求宿主持有插件业务表。
- 插件安装/卸载 SQL、插件本地 `gf gen dao` 配置、插件 service 与测试数据必须围绕上述插件作用域表名保持一致。

### Server Monitor Migration

`monitor-server` 可以作为相对完整的独立插件迁移：

- 插件拥有采集器、清理任务、数据表、查询接口和页面。
- 宿主仅提供任务调度底座和插件生命周期能力。
- 插件启停应联动采集任务注册与撤销。

## Plugin Manifest Rules

- 插件 `id` 必须使用 `kebab-case`，但不要求使用 `plugin-` 前缀。
- 官方插件统一使用领域-能力风格：`org-center`、`content-notice`、`monitor-online` 等。
- 插件菜单 key 继续使用 `plugin:<plugin-id>:<menu-key>` 格式，保证治理一致性。
- 插件 `parent_key` 必须指向宿主稳定目录键或同插件内部菜单键。
- 官方插件的顶层挂载关系固定为：`org-center -> org`、`content-notice -> content`、`monitor-online -> monitor`、`monitor-server -> monitor`、`monitor-operlog -> monitor`、`monitor-loginlog -> monitor`。

## Migration Order

### Phase 1: Governance Foundation

- 固定一级目录与宿主父级 `menu_key`。
- 创建并维护宿主拥有的 9 个一级目录记录。
- 调整菜单 SQL、菜单投影和导航隐藏规则。
- 固定插件 ID 与菜单挂载约束。

### Phase 2: Event and Boundary Extraction

- 抽登录事件契约与发布器。
- 抽审计事件契约与发布器。
- 抽组织能力接口。
- 划清认证会话内核与在线用户治理边界。

### Phase 3: Independent Monitor Plugins

- 迁 `monitor-operlog`。
- 迁 `monitor-loginlog`。
- 迁 `monitor-server`。
- 最后迁 `monitor-online`。

### Phase 4: Organization and Content Plugins

- 迁 `org-center`。
- 迁 `content-notice`。

## Risks and Mitigations

### Risk: User Management Loses Required Fields

- 风险：组织插件未安装时，用户管理页面和接口仍假定部门/岗位必然存在。
- 缓解：先做能力探测与字段降级，再迁插件。

### Risk: Authentication Depends on Online User Plugin

- 风险：将在线用户整体搬出宿主会破坏认证主链路。
- 缓解：保留宿主会话内核，只迁展示与治理能力。

### Risk: Menu Tree Becomes Empty or Fragmented

- 风险：插件未安装时出现空父目录，或一级目录只存在于前端投影导致 `parent_key` 无法稳定解析。
- 缓解：一级目录宿主化、稳定目录记录常驻、空父目录仅在导航投影层隐藏、语义挂载规则固定。

### Risk: Too Many Small Plugins Increase Maintenance Cost

- 风险：监控拆成 4 个插件后文档、测试与样例维护增加。
- 缓解：接受该拆分作为产品要求，同时统一插件模板、菜单挂载规则和测试规范，降低边际成本。
