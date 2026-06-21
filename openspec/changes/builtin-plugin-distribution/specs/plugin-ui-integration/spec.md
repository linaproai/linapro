## ADDED Requirements

### Requirement: 插件管理 UI 必须隐藏 builtin 插件的普通治理入口

系统 SHALL 将`distribution=builtin`插件视为项目内建能力，而不是普通插件管理对象。插件管理列表默认不得展示`builtin`插件。若管理员通过受控诊断入口或详情 API 看到`builtin`插件，UI MUST 隐藏安装、启用、禁用、卸载、手动升级和租户供应策略更新操作，而不是显示禁用态按钮。

#### Scenario: 普通插件管理列表不显示 builtin 插件

- **WHEN** 管理员打开普通插件管理页面
- **THEN** 页面列表请求不包含`builtin`诊断参数
- **AND** 表格中不显示`distribution=builtin`插件

#### Scenario: builtin 详情隐藏写操作

- **WHEN** 页面展示`distribution=builtin`插件详情
- **THEN** 页面不显示安装、启用、禁用、卸载、手动升级或租户供应策略更新操作
- **AND** 页面不得仅通过置灰按钮表达不可操作状态

#### Scenario: 前端不得依赖隐藏操作作为安全边界

- **WHEN** 用户绕过 UI 直接调用`builtin`插件写操作 API
- **THEN** 服务端仍按插件升级治理规范拒绝该操作
- **AND** 前端刷新后继续展示服务端返回的只读状态
