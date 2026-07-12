## MODIFIED Requirements

### Requirement: 插件管理 UI 必须隐藏 builtin 插件的普通治理入口

系统 SHALL 在插件管理列表中展示`distribution=builtin`插件，并将其标识为内建能力。UI MUST 为 builtin 插件展示「内置插件」类标识；若该插件同时命中宿主自动启用管理，UI MUST 继续展示既有「自动启用」类标识。UI MUST 隐藏安装、启用、禁用、卸载、手动升级和租户供应策略更新操作，而不是显示禁用态按钮。UI MUST 允许查看详情；当插件已安装且存在管理页时，MUST 允许通过「管理」按钮进入管理界面。

#### Scenario: 普通插件管理列表展示 builtin 插件

- **WHEN** 管理员打开普通插件管理页面
- **THEN** 表格中可以显示`distribution=builtin`插件
- **AND** 该行展示内置插件标识

#### Scenario: builtin 行隐藏写操作并保留详情与管理

- **WHEN** 页面展示`distribution=builtin`插件行
- **THEN** 页面不显示安装、启用、禁用、卸载、手动升级或租户供应策略更新操作
- **AND** 页面显示详情操作
- **AND** 当插件已安装且存在管理页时显示可用的管理操作

#### Scenario: builtin 详情隐藏写操作

- **WHEN** 页面展示`distribution=builtin`插件详情
- **THEN** 页面不显示安装、启用、禁用、卸载、手动升级或租户供应策略更新操作
- **AND** 页面不得仅通过置灰按钮表达不可操作状态

#### Scenario: 前端不得依赖隐藏操作作为安全边界

- **WHEN** 用户绕过 UI 直接调用`builtin`插件写操作 API
- **THEN** 服务端仍按插件升级治理规范拒绝该操作
- **AND** 前端刷新后继续展示服务端返回的只读状态
