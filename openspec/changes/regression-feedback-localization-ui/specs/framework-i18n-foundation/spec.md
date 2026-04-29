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
