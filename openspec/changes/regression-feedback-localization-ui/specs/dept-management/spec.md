## ADDED Requirements

### Requirement: System-generated unassigned department node must be localized
部门相关页面中由系统自动生成的“未分配部门”虚拟节点 SHALL 使用当前请求语言本地化，并由后端或插件运行时 i18n 资源维护。

#### Scenario: Unassigned department displays in English
- **WHEN** 管理员在 `en-US` 环境下打开包含部门树筛选的页面
- **THEN** 系统生成的未分配部门虚拟节点展示为 `Unassigned Department`
- **AND** 页面不得显示中文 `未分配部门`

#### Scenario: Virtual node identity remains stable
- **WHEN** 管理员选择未分配部门虚拟节点过滤用户或岗位
- **THEN** 前端仍提交既有虚拟节点 ID
- **AND** 后端过滤语义不因展示语言变化而改变
