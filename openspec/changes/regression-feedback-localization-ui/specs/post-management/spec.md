## ADDED Requirements

### Requirement: Position form status selector must remain readable in English
岗位新增和编辑表单 SHALL 在英文环境下保持状态字段标签和选项布局清晰，避免状态选择项因空间不足换行导致视觉不整齐。

#### Scenario: English position status options stay on one line
- **WHEN** 管理员在 `en-US` 环境下打开新增或编辑岗位表单
- **THEN** 状态字段标签和 `Normal` / `Disabled` 等选项保持同一行可读
- **AND** 表单整体布局不遮挡后续字段或操作按钮

#### Scenario: Position form remains responsive
- **WHEN** 视口宽度不足以展示双列表单
- **THEN** 表单可降级为单列或增加合理宽度
- **AND** 状态选项仍可读且可操作
