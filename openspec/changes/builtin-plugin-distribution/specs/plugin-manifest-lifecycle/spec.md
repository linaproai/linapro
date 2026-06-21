## ADDED Requirements

### Requirement: 插件清单必须声明分发治理类型

系统 SHALL 支持插件 manifest 的`distribution`字段。缺省值 MUST 归一化为`marketplace`；合法值仅包含`marketplace`和`builtin`。非法值 MUST 在 manifest 校验阶段失败。`builtin`仅允许`type=source`，且必须存在同 ID 的编译期源码插件注册绑定。插件注册表和发布 manifest snapshot MUST 保存归一化后的`distribution`。

#### Scenario: 普通插件缺省为 marketplace

- **WHEN** 插件 manifest 未声明`distribution`
- **THEN** 系统将该插件归一化为`distribution=marketplace`
- **AND** 同步到`sys_plugin`和发布 manifest snapshot 的值均为`marketplace`

#### Scenario: 内建源码插件声明 builtin

- **WHEN** 源码插件 manifest 声明`distribution: builtin`
- **AND** 该插件通过编译期源码插件注册表绑定同一插件 ID
- **THEN** manifest 校验通过
- **AND** 系统将该插件注册表和发布 manifest snapshot 的`distribution`保存为`builtin`

#### Scenario: 动态插件不能声明 builtin

- **WHEN** 动态插件 manifest 声明`distribution: builtin`
- **THEN** manifest 校验失败
- **AND** 系统不得将该插件降级为`marketplace`继续同步

#### Scenario: 未注册源码插件不能声明 builtin

- **WHEN** 源码插件 manifest 声明`distribution: builtin`
- **AND** 源码插件注册表中不存在同 ID 绑定
- **THEN** manifest 校验失败
- **AND** 错误包含插件 ID 和缺失源码注册绑定的事实

#### Scenario: 非法 distribution 被拒绝

- **WHEN** 插件 manifest 声明`distribution: managed`
- **THEN** manifest 校验失败
- **AND** 错误列出合法值`marketplace`和`builtin`

### Requirement: 插件管理投影必须暴露并过滤分发治理类型

系统 SHALL 在插件列表和详情 API 的服务端投影中返回`distribution`字段。普通插件管理列表默认仅返回`distribution=marketplace`插件。显式诊断查询包含`builtin`插件时，查询 MUST 仍为只读操作，并不得触发治理表写入或生命周期副作用。

#### Scenario: 普通管理列表隐藏 builtin 插件

- **WHEN** 管理员调用默认插件列表查询
- **THEN** 响应仅包含`distribution=marketplace`的插件
- **AND** 响应中的每个插件项都包含`distribution`字段

#### Scenario: 诊断查询包含 builtin 插件

- **WHEN** 具备诊断权限的管理员调用包含内建插件的只读查询
- **THEN** 响应可以包含`distribution=builtin`的插件
- **AND** 查询不得安装、启用、升级、同步或修复任何插件治理数据

#### Scenario: 插件详情返回 distribution

- **WHEN** 管理员查询任一插件详情
- **THEN** 响应包含该插件当前注册表或发布投影中的`distribution`
