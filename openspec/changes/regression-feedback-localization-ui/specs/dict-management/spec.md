## ADDED Requirements

### Requirement: Dictionary forms must keep long English labels readable
字典管理新增和编辑表单 SHALL 在英文环境下为较长标签提供足够空间，避免 `Dictionary Type` 和 `Tag Style` 等标签换行影响页面美观。

#### Scenario: Dictionary type label stays readable
- **WHEN** 管理员在 `en-US` 环境下打开新增或编辑字典类型表单
- **THEN** `Dictionary Type` 标签保持可读且不被强制拆行
- **AND** 表单输入区域与标签对齐清晰

#### Scenario: Tag style label stays readable
- **WHEN** 管理员在 `en-US` 环境下打开新增或编辑字典数据表单
- **THEN** `Tag Style` 标签保持可读且不被强制拆行
- **AND** 标签样式选择器不会遮挡其他字段

### Requirement: Tag style picker options must use valid translations
字典数据表单的 Tag Style 下拉框 SHALL 展示当前语言下的人类可读选项文本，不得直接显示运行时 i18n key。

#### Scenario: Tag style options render English labels
- **WHEN** 管理员在 `en-US` 环境下展开 Tag Style 下拉框
- **THEN** 下拉选项展示 `Default`、`Primary`、`Success` 等英文标签
- **AND** 不显示 `pages.*`、`component.*` 或其他 i18n key 原文
