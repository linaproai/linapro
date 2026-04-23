## ADDED Requirements

### Requirement: 登录框位置内置系统参数元数据

系统 MUST 提供受保护的 public frontend 内置参数 `sys.auth.loginPanelLayout`，用于维护登录页默认登录框布局。

#### Scenario: 初始化登录框位置参数
- **WHEN** 管理员执行宿主初始化 SQL
- **THEN** `sys_config` 包含键名为 `sys.auth.loginPanelLayout` 的内置参数记录
- **AND** 该记录默认值为 `panel-right`
- **AND** 该记录包含可读名称和取值说明 `panel-left`、`panel-center`、`panel-right`

### Requirement: 登录框位置参数校验并暴露给公共前端配置接口

系统 MUST 校验 `sys.auth.loginPanelLayout` 的值域，并通过公共前端配置接口把生效值暴露给未登录页面。

#### Scenario: 拒绝非法登录框位置值
- **WHEN** 用户创建、更新或导入 `sys.auth.loginPanelLayout`，其值不是 `panel-left`、`panel-center`、`panel-right` 之一
- **THEN** 系统拒绝该变更并返回参数校验错误

#### Scenario: 公共前端配置返回登录框位置
- **WHEN** 浏览器请求 `GET /config/public/frontend`
- **THEN** 响应中的 `auth.panelLayout` 等于 `sys.auth.loginPanelLayout` 的当前生效值
- **AND** 未登录页面可以在不读取任意其他 `sys_config` 数据的前提下消费该配置

### Requirement: 登录页说明文案参数默认值与长度约束

系统 MUST 为受保护的 public frontend 内置参数 `sys.auth.pageDesc` 提供默认说明文案，并允许保存最长 500 个字符的非空说明文案，以支持登录页展示更完整的能力说明。

#### Scenario: 初始化登录页说明文案参数
- **WHEN** 管理员执行宿主初始化 SQL
- **THEN** `sys_config` 包含键名为 `sys.auth.pageDesc` 的内置参数记录
- **AND** 该记录默认值为 `面向业务演进，提供开箱即用的管理入口与灵活可插拔的扩展机制`

#### Scenario: 保存 500 个字符以内的登录页说明文案
- **WHEN** 管理员在系统参数管理中创建、更新或导入 `sys.auth.pageDesc`，且值长度为 1 到 500 个字符
- **THEN** 系统接受该变更并保存参数值
- **AND** 公共前端配置接口返回的 `auth.pageDesc` 与已保存值一致

#### Scenario: 拒绝超长登录页说明文案
- **WHEN** 管理员在系统参数管理中创建、更新或导入 `sys.auth.pageDesc`，且值长度超过 500 个字符
- **THEN** 系统拒绝该变更并返回参数校验错误
