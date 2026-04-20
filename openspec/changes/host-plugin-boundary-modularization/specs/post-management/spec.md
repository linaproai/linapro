## ADDED Requirements

### Requirement: 岗位管理由组织源码插件交付

系统 SHALL 将岗位管理能力作为 `org-management` 源码插件交付，而不是继续作为宿主默认内建模块。

#### Scenario: 组织插件启用时提供岗位管理
- **WHEN** `org-management` 已安装并启用
- **THEN** 宿主暴露岗位管理 API、页面与菜单
- **AND** 岗位管理菜单挂载到宿主 `组织管理` 目录，顶层 `parent_key` 为 `org`

#### Scenario: 组织插件缺失时隐藏岗位管理入口
- **WHEN** `org-management` 未安装或未启用
- **THEN** 宿主不显示岗位管理菜单和页面入口
- **AND** 用户管理等宿主能力按照组织降级规则继续可用
