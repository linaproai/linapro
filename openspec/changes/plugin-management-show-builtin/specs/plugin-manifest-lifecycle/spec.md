## MODIFIED Requirements

### Requirement: 插件管理投影必须暴露并过滤分发治理类型

系统 SHALL 在插件列表和详情 API 的服务端投影中返回`distribution`字段。普通插件管理列表 MUST 同时返回`distribution=managed`与`distribution=builtin`插件，并按现有分页与筛选条件投影。列表查询 MUST 仍为只读操作，并不得因包含`builtin`而触发治理表写入或生命周期副作用。兼容查询字段`includeBuiltin`若仍存在，MUST NOT 再用于隐藏`builtin`插件。

#### Scenario: 普通管理列表包含 builtin 插件

- **WHEN** 管理员调用默认插件列表查询
- **THEN** 响应可同时包含`distribution=managed`与`distribution=builtin`的插件
- **AND** 响应中的每个插件项都包含`distribution`字段

#### Scenario: 包含 builtin 的列表查询保持只读

- **WHEN** 管理员调用包含内建插件的插件列表查询
- **THEN** 查询不得安装、启用、升级、同步或修复任何插件治理数据

#### Scenario: 插件详情返回 distribution

- **WHEN** 管理员查询任一插件详情
- **THEN** 响应包含该插件当前注册表或发布投影中的`distribution`
