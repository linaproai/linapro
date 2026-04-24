## ADDED Requirements

### Requirement: 登录页必须支持宿主国际化文案与语言切换联动
The system SHALL render login-page title, description, and subtitle according to the active language, combining frontend static language resources with localized public frontend settings returned by the host. When the active language changes, the login page MUST refresh the displayed copy without requiring a new login session.

#### Scenario: 登录页按英文展示宿主文案
- **WHEN** 浏览器当前语言为 `en-US` 且宿主已提供对应语言的公共前端配置文案
- **THEN** 登录页显示英文标题、说明和登录副标题
- **AND** 静态表单字段文案继续通过前端静态语言包渲染

#### Scenario: 切换语言后刷新登录页文案
- **WHEN** 用户在登录前或登录后切换工作台语言
- **THEN** 登录页或认证布局中的宿主文案同步刷新为新语言结果
- **AND** 不需要重新配置登录页组件结构

### Requirement: 登录页国际化缺失时必须回退默认文案
The system SHALL fall back to the default language copy or built-in static copy when the host does not provide localized login-page text for the current language.

#### Scenario: 当前语言缺少登录页说明翻译
- **WHEN** 当前语言下没有可用的 `auth.pageDesc` 本地化结果
- **THEN** 登录页回退显示默认语言说明或内建默认说明文案
- **AND** 登录页布局与认证流程保持可用
