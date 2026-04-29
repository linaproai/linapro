## ADDED Requirements

### Requirement: User list role names must match backend-localized role display
用户管理列表 SHALL 使用后端返回的角色展示名称，并与角色管理页面在同一语言下的内置角色展示保持一致。

#### Scenario: User list shows administrator role in English
- **WHEN** 管理员在 `en-US` 环境下打开用户管理页面
- **THEN** 用户列表中 `admin` 用户关联的 `admin` 角色展示为与角色管理页面一致的英文名称
- **AND** 前端不得基于中文角色名或角色 key 维护额外映射

#### Scenario: Role selector keeps governance semantics
- **WHEN** 管理员打开用户新增或编辑表单
- **THEN** 角色选择器继续使用后端角色选项数据
- **AND** 保存用户角色关系时仍提交稳定角色 ID，而不是本地化展示文本
