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

### Requirement: Built-in dictionary records must be editable but not deletable
字典类型和字典数据 SHALL 标识系统内置记录，并禁止删除系统内置记录，同时继续允许管理员修改内置记录的可编辑字段。

#### Scenario: Built-in dictionary type delete action is disabled
- **WHEN** 管理员查看字典类型列表中的系统内置字典类型
- **THEN** 该行删除按钮置灰且不会触发删除确认
- **AND** 鼠标悬停删除按钮时展示系统内置数据不支持删除的提示
- **AND** 编辑按钮仍可打开编辑表单

#### Scenario: Built-in dictionary data delete action is disabled
- **WHEN** 管理员查看字典数据列表中的系统内置字典数据
- **THEN** 该行删除按钮置灰且不会触发删除确认
- **AND** 鼠标悬停删除按钮时展示系统内置数据不支持删除的提示
- **AND** 编辑按钮仍可打开编辑表单

#### Scenario: Backend rejects built-in dictionary deletion
- **WHEN** 调用端绕过前端直接请求删除系统内置字典类型或字典数据
- **THEN** 后端 SHALL 返回结构化业务错误并保留记录
- **AND** 非内置字典类型和字典数据仍可按既有权限与校验规则删除
