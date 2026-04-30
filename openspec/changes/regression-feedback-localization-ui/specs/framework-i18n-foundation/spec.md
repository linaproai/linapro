## ADDED Requirements

### Requirement: English regression sweep must cover framework-delivered pages and seed display content
默认管理工作台 SHALL 为人工回归反馈中覆盖的框架交付页面提供英文环境回归保障，确保系统生成内容、默认种子内容和静态 UI 文案不残留中文。

#### Scenario: English regression pages contain no Chinese system copy
- **WHEN** 管理员切换到 `en-US` 并打开工作台、用户管理、角色管理、部门管理、岗位管理、字典管理、系统参数、服务监控和定时任务页面
- **THEN** 页面中的框架交付标题、按钮、表单标签、表格列、系统生成节点、默认内置记录展示和确认弹窗文案均使用英文
- **AND** 用户可编辑业务字段仅在已明确纳入框架交付投影时才被本地化

#### Scenario: English layout regressions are screenshot checked
- **WHEN** Playwright 在 `en-US` 环境下截取岗位表单、字典表单和服务监控磁盘表格
- **THEN** 关键标签、选项、表头和列值不发生不可读换行或互相遮挡
- **AND** 截图检查结果作为本变更验收依据之一

#### Scenario: Version information menu title is localized consistently
- **WHEN** 管理员查看开发中心下的版本信息页面入口
- **THEN** `zh-CN` 展示为 `版本信息`
- **AND** `en-US` 展示为 `Version Info`
- **AND** `zh-TW` 展示为 `版本資訊`

### Requirement: Runtime locale JSON values must avoid markdown-only code markers
运行时翻译 JSON SHALL 避免在用户可见字符串中使用 markdown 式反引号标记，因为普通 UI 渲染不会执行代码高亮，原始反引号会影响阅读体验。

#### Scenario: Locale JSON strings are displayed as plain UI text
- **WHEN** 前端、宿主或插件交付的 locale JSON 字符串包含文件路径、参数示例、通配符或扩展名
- **THEN** 字符串 SHALL 直接展示对应内容本身
- **AND** 字符串 SHALL NOT 使用反引号包裹这些内容
- **AND** 自动化检查 SHALL 阻止 locale JSON 字符串重新引入反引号
