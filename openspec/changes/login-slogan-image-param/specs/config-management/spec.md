## ADDED Requirements

### Requirement:内置参数 sys.auth.sloganImage

系统 SHALL 提供内置参数 `sys.auth.sloganImage`，用于配置登录页 slogan 插画图片地址。默认值为 `/slogan.svg`（Vben 内置插画）。空值表示不使用插画。

#### Scenario:参数设置页可见 slogan 参数
- **当** 管理员打开参数设置并搜索 `sys.auth.sloganImage` 时
- **则** 列表显示该内置参数
- **且** 参数名称标识为登录展示相关的 slogan 插画配置

#### Scenario:允许清空 slogan 地址以隐藏插画
- **当** 管理员将 `sys.auth.sloganImage` 保存为空值时
- **则** 系统接受该值
- **且** 登录页不展示 slogan 插画
