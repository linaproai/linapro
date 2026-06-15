## ADDED Requirements

### Requirement: Role 模块必须发布动态路由访问投影契约

系统 SHALL 由`role`模块发布面向动态插件运行时的访问投影契约。该契约 MUST 基于 token access snapshot 和`permission-access`修订号返回当前 token、用户和租户下的权限字符串、角色名、数据范围、unsupported 数据范围标记和超管标记。插件运行时 MUST 不直接读取`sys_user_role`、`sys_role`、`sys_role_menu`或`sys_menu`来构建同类投影。

#### Scenario: 动态路由读取 role 访问投影

- **WHEN** 动态插件登录路由需要构建用户身份快照
- **THEN** 插件运行时调用`role`模块发布的访问投影契约
- **AND** 返回值来自与宿主受保护 API 一致的 token access snapshot
- **AND** 插件运行时不导入角色 DAO、DO 或 Entity

#### Scenario: 权限拓扑变化后动态路由使用新快照

- **WHEN** 管理员修改角色菜单、用户角色、插件权限菜单或数据范围并发布`permission-access`修订号
- **THEN** 动态插件路由后续权限校验使用新修订号下的访问投影
- **AND** 不无限期沿用旧权限、旧角色名或旧数据范围

#### Scenario: 权限 freshness 不可确认时拒绝动态路由

- **WHEN** 集群模式下动态路由需要访问投影
- **AND** Redis`permission-access`修订号不可读取且本地权限缓存超过最大陈旧窗口
- **THEN** `role`访问投影契约返回失败
- **AND** 动态插件请求不得进入 guest 执行

### Requirement: 访问投影必须携带租户维度且不泄漏内部模型

系统 SHALL 在动态路由访问投影的缓存键、输入校验和输出 DTO 中保留租户维度。该契约返回的投影 MUST 是`role`模块自有 DTO 或值对象，不得暴露`DAO`、`DO`、`Entity`、`gdb.Model`、私有缓存结构或可修改共享快照。

#### Scenario: 同一用户不同租户访问投影隔离

- **WHEN** 用户`U`同时持有租户`A`和租户`B`的有效 token
- **THEN** 租户`A`动态路由访问投影不得被租户`B`复用
- **AND** 修改租户`A`权限不得污染租户`B`的无关访问投影

#### Scenario: 调用方不能修改共享快照

- **WHEN** 插件运行时接收`role`访问投影结果
- **THEN** 返回的切片、map 或嵌套对象必须与共享缓存条目隔离
- **AND** 调用方修改本次请求对象不得污染后续请求
