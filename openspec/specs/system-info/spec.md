# system-info Specification

## Purpose
TBD - created by archiving change v0.5.0. Update Purpose after archive.
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

### Requirement: System information page must display project introduction and component descriptions by current language
The system SHALL return project description, component descriptions, and other display copy on the system information page and system information API according to the current request language. System information i18n MUST keep project positioning and component identifiers stable, localizing only user-facing descriptive text.

#### Scenario: System information displays in English
- **WHEN** a user opens the system information page or requests the system information API with `en-US`
- **THEN** the project description in the About section uses an English localized result
- **AND** frontend and backend component descriptions use English localized results
- **AND** component names, version numbers, and links keep their original values

#### Scenario: Missing component descriptions fall back to the default language
- **WHEN** a component lacks description copy in the current language
- **THEN** the system falls back to the default-language description
- **AND** the component still displays normally in the corresponding section

### Requirement: System information i18n must cover public project positioning copy
The system SHALL keep project name, project introduction, and framework positioning descriptions semantically consistent across multilingual scenarios, ensuring that `LinaPro` is always described as an AI-native full-stack framework engineered for sustainable delivery and does not drift into a single admin system or other product boundary in another language.

#### Scenario: Unified project positioning is preserved across languages
- **WHEN** a user switches the system language and views the system information page
- **THEN** the project positioning copy changes only in language expression
- **AND** LinaPro is not described as a single backend management system or any other product positioning that deviates from framework positioning

