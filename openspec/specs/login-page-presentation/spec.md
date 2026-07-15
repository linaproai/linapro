# 登录页展示规范

## Purpose
待定 - 由归档变更 login-page-simplification-and-positioning 创建。归档后更新目的。
## Requirements
### Requirement:当前阶段仅暴露用户名/密码登录入口

系统在当前阶段 SHALL 将用户名/密码登录作为默认公开登录能力，并允许展示已交付的忘记密码与创建账号认证子流程入口。手机号登录、扫码登录不得继续作为正式公开能力展示。

#### Scenario:标准登录页显示用户名密码与账号辅助入口
- **当** 未认证用户访问 `/auth/login` 时
- **则** 页面显示用户名、密码、记住我和登录控件
- **且** 页面显示忘记密码入口与创建账号入口
- **且** 页面不显示手机号登录或扫码登录入口

#### Scenario:用户访问未交付的认证子路由
- **当** 用户访问 `/auth/code-login` 或 `/auth/qrcode-login` 时
- **则** 系统重定向回标准登录页 `/auth/login`
- **且** 页面仍不显示手机号登录或扫码登录入口

#### Scenario:用户访问已交付的账号辅助子路由
- **当** 用户访问 `/auth/forget-password` 或 `/auth/register` 时
- **则** 系统渲染对应认证子页，而不是重定向到 `/auth/login`

### Requirement:登录面板默认右对齐布局并支持位置配置

系统 SHALL 默认以居中布局渲染登录面板，并允许宿主公共前端配置切换到左、中或右布局。

#### Scenario:无覆盖时登录面板默认居中
- **当** 浏览器加载登录页且宿主未提供登录面板位置覆盖时
- **则** 登录页使用 `panel-center` 布局
- **且** 登录面板显示在主页面区域的中间

#### Scenario:宿主配置覆盖登录面板位置
- **当** 宿主公共前端配置返回 `auth.panelLayout` 为 `panel-left`、`panel-center` 或 `panel-right` 时
- **则** 登录页渲染对应的布局模式
- **且** 登录页工具栏中的布局切换器仍允许在三种布局选项间切换

### Requirement:默认登录页描述支持宿主配置

系统 SHALL 在宿主未提供覆盖时显示默认登录页描述，并在宿主公共前端配置提供时显示配置值。

#### Scenario:无覆盖时显示默认描述
- **当** 浏览器加载登录页且宿主未提供 `auth.pageDesc` 覆盖时
- **则** 登录页显示描述 `Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model`

#### Scenario:宿主配置覆盖登录页描述
- **当** 宿主公共前端配置返回非空的 `auth.pageDesc` 时
- **则** 登录页显示返回的描述

### Requirement:登录页必须支持宿主国际化文案和语言切换刷新

系统 SHALL 根据当前语言渲染登录页标题、描述和副标题，结合前端静态语言资源和宿主返回的本地化公共前端设置。当前语言变化时，登录页必须刷新显示的文案，无需新的登录会话。

#### Scenario:登录页以英文显示宿主文案
- **当** 浏览器语言为 `en-US` 且宿主提供该语言的公共前端配置文案时
- **则** 登录页显示英文标题、描述和登录副标题
- **且** 静态表单字段文案继续从前端静态语言包渲染

#### Scenario:语言切换后登录页文案刷新
- **当** 用户在登录前或登录后切换工作区语言时
- **则** 登录页或认证布局中的宿主文案刷新为新语言结果
- **且** 登录页组件结构无需重新配置

### Requirement:登录页国际化缺失必须回退到默认文案

当宿主未为当前语言提供本地化登录页文本时，系统 SHALL 回退到默认语言文案或内置静态文案。

#### Scenario:当前语言缺少登录页描述翻译
- **当** 当前语言没有 `auth.pageDesc` 的可用本地化结果时
- **则** 登录页回退到默认语言描述或内置默认描述文案
- **且** 登录页布局和认证流程保持可用

### Requirement:登录页 Slogan 插画支持系统参数配置

系统 SHALL 通过内置公共前端参数 `sys.auth.sloganImage` 配置登录页侧栏 slogan 插画地址，并在公共前端配置中暴露为 `auth.sloganImage`。

语义约定：
- 默认值 SHALL 为 Vben 内置插画静态地址 `/slogan.svg`
- 参数为空字符串时 SHALL 不展示侧栏插画
- 参数为非空图片地址时 SHALL 渲染该图片

#### Scenario:默认参数展示内置 slogan 插画
- **当** `sys.auth.loginPanelLayout` 为 `panel-left` 或 `panel-right`
- **且** `sys.auth.sloganImage` 保持默认 `/slogan.svg`（或宿主未配置时回退该默认）时
- **则** 公共前端配置 `auth.sloganImage` 为 `/slogan.svg`
- **且** 登录页侧栏以图片方式展示该内置插画

#### Scenario:空参数时不展示 slogan 插画
- **当** 管理员将 `sys.auth.sloganImage` 保存为空字符串
- **且** 登录面板布局为 `panel-left` 或 `panel-right` 时
- **则** 公共前端配置 `auth.sloganImage` 为空字符串
- **且** 登录页侧栏不渲染 slogan 插画

#### Scenario:配置自定义 slogan 图片地址
- **当** 管理员将 `sys.auth.sloganImage` 更新为有效图片地址
- **且** 登录面板布局为 `panel-left` 或 `panel-right` 时
- **则** 公共前端配置返回该地址
- **且** 登录页侧栏以图片方式展示该 slogan

#### Scenario:居中布局不展示侧栏 slogan
- **当** 登录面板布局为 `panel-center` 时
- **则** 登录页不展示侧栏 slogan 区域（无论 slogan 参数是否配置）

