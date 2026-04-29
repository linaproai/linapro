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
