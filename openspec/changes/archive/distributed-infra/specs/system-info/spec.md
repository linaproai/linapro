# 系统信息规范

## Purpose
待定 - 由归档变更 v0.5.0 创建。归档后更新目的。
## Requirements
### Requirement: 系统信息页面展示
系统 SHALL 提供一个"系统信息"页面，展示四个信息区块：关于项目、基本信息、后端组件、前端组件。页面样式参考 vben5 框架的 About 页面，使用 Card + Descriptions 布局。

#### Scenario: 正常加载系统信息页面
- **WHEN** 用户点击"系统信息 > 系统信息"菜单
- **THEN** 页面展示四个卡片区块，分别显示项目介绍、系统运行时基本信息、后端技术组件列表、前端技术组件列表

### Requirement: 后端系统信息 API
系统 SHALL 提供 `GET /api/v1/system/info` 接口，返回系统运行时信息。该接口 MUST 在鉴权路由组内，仅登录用户可访问。

#### Scenario: 获取系统运行时信息
- **WHEN** 前端请求 `GET /api/v1/system/info`
- **THEN** 接口返回包含以下字段的 JSON 数据：Go 版本、GoFrame 版本、操作系统及架构、数据库版本、系统启动时间、系统运行时长

### Requirement: 关于项目区块
关于项目区块 SHALL 展示项目名称、以”`面向可持续交付的 AI 原生全栈框架`”为核心的项目描述、当前版本号、开源许可证和项目主页链接。这些信息在前端配置对象中定义。

#### Scenario: 展示项目基本信息
- **WHEN** 系统信息页面加载完成
- **THEN** "关于项目"区块显示项目名称 "LinaPro"、项目描述、版本号、许可证类型，项目主页为可点击的外链

#### Scenario: 展示统一项目定位
- **WHEN** 系统信息页面加载完成
- **THEN** 项目描述明确将 `LinaPro` 表述为”`面向可持续交付的 AI 原生全栈框架`”
- **AND** 若描述管理后台相关能力，则将其表述为默认管理工作台或内建通用模块
- **AND** 不再将整个项目描述为单一后台管理系统

### Requirement: 基本信息区块
基本信息区块 SHALL 展示从后端 API 获取的运行时数据，包括 Go 版本、GoFrame 版本、操作系统、数据库版本、启动时间、运行时长。

#### Scenario: 展示运行时信息
- **WHEN** 后端 API 返回运行时数据
- **THEN** "基本信息"区块以键值对形式展示所有运行时字段

### Requirement: 后端组件区块
后端组件区块 SHALL 展示后端使用的技术组件列表，每个组件包含名称、版本号和官网链接。组件列表在前端配置对象中定义。

#### Scenario: 展示后端组件列表
- **WHEN** 系统信息页面加载完成
- **THEN** "后端组件"区块以网格布局展示 GoFrame、MySQL、JWT 等后端组件，每个组件的名称和版本号可见，官网链接可点击跳转

### Requirement: 前端组件区块
前端组件区块 SHALL 展示前端使用的技术组件列表，每个组件包含名称、版本号和官网链接。组件列表在前端配置对象中定义。

#### Scenario: 展示前端组件列表
- **WHEN** 系统信息页面加载完成
- **THEN** "前端组件"区块以网格布局展示 Vue、Vben5、Ant Design Vue、TypeScript 等前端组件，每个组件的名称和版本号可见，官网链接可点击跳转

### Requirement: 外链地址配置化
所有第三方组件的官网链接 SHALL 在前端配置对象中集中定义，修改链接时无需改动页面组件代码。

#### Scenario: 修改外链地址
- **WHEN** 开发者修改前端配置对象中某个组件的链接地址
- **THEN** 系统信息页面对应组件的链接自动更新为新地址

### Requirement:系统信息页面必须按当前语言显示项目介绍和组件描述
系统 SHALL 根据当前请求语言返回系统信息页面和系统信息 API 上的项目描述、组件描述和其他显示文案。系统信息国际化必须保持项目定位和组件标识稳定，仅本地化面向用户的描述性文本。

#### Scenario:系统信息以英文显示
- **当** 用户以 `en-US` 打开系统信息页面或请求系统信息 API 时
- **则** "关于项目"区块中的项目描述使用英文本地化结果
- **且** 前端和后端组件描述使用英文本地化结果
- **且** 组件名称、版本号和链接保持原始值

#### Scenario:组件描述缺失时回退到默认语言
- **当** 组件缺少当前语言的描述文案时
- **则** 系统回退到默认语言描述
- **且** 组件在对应区块中仍正常显示

### Requirement:系统信息国际化必须覆盖公共项目定位文案
系统 SHALL 在多语言场景下保持项目名称、项目介绍和框架定位描述的语义一致性，确保 `LinaPro` 始终被描述为面向可持续交付的 AI 原生全栈框架，不会在其他语言中偏离为单一后台管理系统或其他产品定位。

#### Scenario:统一项目定位跨语言保留
- **当** 用户切换系统语言并查看系统信息页面时
- **则** 项目定位文案仅在语言表达上变化
- **且** LinaPro 不被描述为单一后台管理系统或偏离框架定位的其他产品定位

### Requirement: 系统信息必须暴露 coordination 状态
系统 SHALL 在系统信息或健康诊断中暴露集群 coordination 状态。响应至少包含 cluster enabled、coordination backend、Redis 健康、当前 node ID、primary 状态和最近错误。

#### Scenario: 集群 Redis 健康状态
- **WHEN** 运维查询系统信息
- **AND** 宿主以 `cluster.coordination=redis` 运行
- **THEN** 响应包含 coordination backend `redis`
- **AND** 响应包含 Redis ping 状态
- **AND** 响应包含当前节点 primary 状态

#### Scenario: Redis 最近错误可见
- **WHEN** Redis coordination 最近发生连接错误
- **THEN** 系统信息响应包含最近错误摘要
- **AND** 不暴露 Redis 密码或敏感连接串

### Requirement: 缓存协调诊断必须包含 Redis revision 状态
系统 SHALL 在缓存协调诊断中展示 Redis-backed revision 状态，包括 domain、scope、tenant scope、本地 observed revision、shared revision、last synced at、recent error 和 stale seconds。

#### Scenario: 查询 revision 诊断
- **WHEN** 运维查询缓存协调状态
- **THEN** 响应展示每个已配置或已触及 domain/scope 的 revision 信息
- **AND** 可识别节点是否落后于 Redis shared revision

### Requirement: 诊断字段必须同步 apidoc i18n
如果系统信息或健康 API 新增 coordination 诊断字段，系统 SHALL 同步维护 apidoc i18n JSON。不得只修改响应结构而遗漏接口文档翻译资源。

#### Scenario: 新增 coordination 字段文档
- **WHEN** API 响应新增 `coordination.backend` 或 `coordination.redisHealthy`
- **THEN** 对应 apidoc i18n JSON 包含字段说明
- **AND** `openspec validate` 和静态检查不发现缺失的 apidoc i18n 资源
