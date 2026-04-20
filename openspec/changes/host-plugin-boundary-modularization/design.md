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

- `org-management`
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

当前用户管理直接依赖组织树、部门字段与岗位选项。若 `org-management` 未安装，宿主仍必须保证用户管理可用。因此需要先抽象组织能力提供者：

- 用户列表中的部门列改为可选展示。
- 用户表单中的部门树与岗位选择器改为基于组织插件能力探测按需显示。
- 用户详情中的组织信息改为可选扩展数据，不再视为宿主硬字段。
- 宿主业务逻辑只依赖“组织能力接口”，不直接依赖 `dept` / `post` service 实现。
- `org-management` 缺失时，用户管理页面退化为无左树筛选、无部门/岗位字段的核心用户管理视图。

### Online Session vs Online User Plugin

`在线用户` 虽然要独立为源码插件，但宿主认证链路仍依赖在线会话有效性校验。当前认证中间件直接调用会话存储做 `TouchOrValidate`，因此：

- 宿主保留认证会话内核、`sys_online_session` 真相源以及会话存储抽象。
- 宿主继续负责登录创建会话、登出删除会话、请求触达时刷新活跃时间、超时判定与清理任务。
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

### Server Monitor Migration

`monitor-server` 可以作为相对完整的独立插件迁移：

- 插件拥有采集器、清理任务、数据表、查询接口和页面。
- 宿主仅提供任务调度底座和插件生命周期能力。
- 插件启停应联动采集任务注册与撤销。

## Plugin Manifest Rules

- 插件 `id` 必须使用 `kebab-case`，但不要求使用 `plugin-` 前缀。
- 官方插件统一使用领域-能力风格：`org-management`、`content-notice`、`monitor-online` 等。
- 插件菜单 key 继续使用 `plugin:<plugin-id>:<menu-key>` 格式，保证治理一致性。
- 插件 `parent_key` 必须指向宿主稳定目录键或同插件内部菜单键。
- 官方插件的顶层挂载关系固定为：`org-management -> org`、`content-notice -> content`、`monitor-online -> monitor`、`monitor-server -> monitor`、`monitor-operlog -> monitor`、`monitor-loginlog -> monitor`。

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

- 迁 `org-management`。
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
