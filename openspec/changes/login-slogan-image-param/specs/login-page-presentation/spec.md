## ADDED Requirements

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
