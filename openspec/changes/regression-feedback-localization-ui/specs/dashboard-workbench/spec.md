## ADDED Requirements

### Requirement: Workbench page copy must use runtime i18n resources
工作台页面 SHALL 将默认交付内容接入运行时 i18n 资源，英文环境不得显示中文项目动态、待办、快捷入口或说明文案。

#### Scenario: Workbench displays English default copy
- **WHEN** 管理员在 `en-US` 环境下打开工作台页面
- **THEN** 工作台首页的标题、指标、快捷入口、项目卡片、动态和待办内容使用英文运行时文案
- **AND** 页面不得出现中文默认系统文案

#### Scenario: Workbench copy changes with language switch
- **WHEN** 管理员从 `zh-CN` 切换到 `en-US`
- **THEN** 工作台页面无需重新登录即可刷新为英文文案
- **AND** 路由标题、标签页标题和页面内容语言保持一致

#### Scenario: Workbench quick navigation targets core management pages
- **WHEN** 管理员打开工作台页面
- **THEN** 快捷导航展示 `用户管理`、`菜单管理`、`系统参数`、`扩展中心`、`接口文档` 和 `定时任务`
- **AND** 这些入口分别跳转到 `/system/user`、`/system/menu`、`/system/config`、`/system/plugin`、`/about/api-docs` 和 `/system/job`

#### Scenario: Workbench demo activity and todos are LinaPro-specific
- **WHEN** 管理员查看工作台最新动态和待办事项
- **THEN** 示例内容围绕 LinaPro 核心宿主、插件扩展、接口文档、任务调度、OpenSpec 和自动化验证展示
- **AND** 最新动态和待办事项保持既有列表样式、间距和交互外观不变

#### Scenario: Workbench project cards reflect the delivered admin stack
- **WHEN** 管理员打开工作台页面
- **THEN** 项目卡片展示 `LinaPro`、`GoFrame`、`Vue`、`Vben`、`Ant Design` 和 `TypeScript`
- **AND** 每张卡片跳转到对应项目官网或文档地址
- **AND** 卡片日期统一展示为 `2026-05-01`
- **AND** 卡片说明保持简短激励文案，不作为项目介绍长文本展示

#### Scenario: Workbench project cards use local logo assets
- **WHEN** 管理员查看工作台项目卡片
- **THEN** `LinaPro` 使用从 `/logo.png` 无损转换得到的 `/logo.webp`
- **AND** `GoFrame` 使用从 `/Users/john/Temp/goframe-logo.png` 无损转换得到的 `/goframe-logo.webp`
- **AND** `Vben` 使用下载到本地 public 目录的 `/vben-logo.webp`
- **AND** 英文描述文案保持单行展示，超出卡片宽度时使用省略号

### Requirement: Management workbench logo must provide subtle dark-mode edge depth
管理工作台 SHALL 在深色模式下为左上角品牌 Logo 图标边缘提供极轻微青色光亮效果，以增强深色界面的层次感，同时不得影响亮色模式、Logo 可点击区域或导航布局稳定性。

#### Scenario: Dark-mode logo edge glow remains subtle and layout-safe
- **WHEN** 管理员在深色模式下打开管理工作台
- **THEN** 左上角品牌 Logo 图标边缘显示极轻微青色光亮效果
- **AND** 光亮效果 SHALL follow the icon pixels rather than rendering as a glow around the full image box
- **AND** 光亮效果不得导致 Logo 文本、侧边栏菜单或顶部导航发生布局偏移

#### Scenario: Light-mode logo keeps the original visual treatment
- **WHEN** 管理员切换到亮色模式
- **THEN** 左上角品牌 Logo 图标不显示深色模式专用青色边缘光亮效果

### Requirement: User theme preference must override the public frontend default
管理工作台 SHALL 在启动同步公开前端配置时优先保留用户显式设置过的主题偏好；只有用户尚未设置主题偏好时，才采用系统参数 `sys.ui.theme.mode` 提供的默认主题模式。

#### Scenario: Existing user theme preference survives page refresh
- **GIVEN** 用户已在工作台中显式设置主题模式为 `dark`
- **AND** 系统公开前端配置 `sys.ui.theme.mode` 为 `light`
- **WHEN** 用户刷新管理工作台页面
- **THEN** 工作台继续使用用户设置的 `dark` 主题模式
- **AND** 启动同步公开前端配置不得把本地用户主题偏好改回 `light`

#### Scenario: System default applies when no user theme preference exists
- **GIVEN** 用户本地没有显式主题偏好
- **AND** 系统公开前端配置 `sys.ui.theme.mode` 为 `light`、`dark` 或 `auto`
- **WHEN** 用户打开或刷新管理工作台页面
- **THEN** 工作台使用系统公开前端配置提供的主题模式作为默认值
