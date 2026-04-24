## ADDED Requirements

### Requirement: 宿主嵌入式插件页面必须参与宿主语言上下文与翻译包刷新链路
The system SHALL let host-managed plugin pages participate in the host locale context and runtime message refresh flow. When the active language changes, host-embedded plugin pages and their route metadata MUST refresh to the new language without requiring plugin reinstallation.

#### Scenario: 切换语言后刷新插件路由标题
- **WHEN** 用户在已登录会话中切换工作台语言
- **THEN** 宿主重新构建当前用户可访问的插件路由标题和菜单标题
- **AND** 已启用插件页面在导航和页签中的显示文案切换为新语言

#### Scenario: 宿主嵌入式插件页面加载运行时翻译包
- **WHEN** 宿主以嵌入方式加载插件前端页面
- **THEN** 插件页面能够获取宿主当前语言上下文和对应的运行时翻译资源
- **AND** 插件无需自行重复实现与宿主脱节的语言解析规则

### Requirement: 插件语言资源变更后必须能被宿主及时感知
The system SHALL refresh plugin-related runtime messages after plugin enablement, disablement, or upgrade so that the host UI does not continue showing stale plugin translations.

#### Scenario: 启用插件后立即显示插件翻译
- **WHEN** 管理员在当前会话中启用了一个带有国际化资源的插件
- **THEN** 宿主刷新运行时翻译包与动态菜单
- **AND** 该插件的导航标题和页面文案可立即使用当前语言显示

#### Scenario: 升级插件后切换到新版本翻译资源
- **WHEN** 插件升级到包含新翻译资源的 release 并完成生效切换
- **THEN** 宿主后续提供的运行时翻译包使用新 release 的资源
- **AND** 旧 release 的翻译消息不再继续被前端消费
