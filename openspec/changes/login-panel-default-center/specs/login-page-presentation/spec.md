## MODIFIED Requirements

### Requirement:登录面板默认居中布局并支持位置配置

系统 SHALL 默认以居中布局渲染登录面板，并允许宿主公共前端配置切换到左、中或右布局。

#### Scenario:无覆盖时登录面板默认居中
- **当** 浏览器加载登录页且宿主未提供登录面板位置覆盖时
- **则** 登录页使用 `panel-center` 布局
- **且** 登录面板显示在主页面区域的中间

#### Scenario:宿主配置覆盖登录面板位置
- **当** 宿主公共前端配置返回 `auth.panelLayout` 为 `panel-left`、`panel-center` 或 `panel-right` 时
- **则** 登录页渲染对应的布局模式
- **且** 登录页工具栏中的布局切换器仍允许在三种布局选项间切换
