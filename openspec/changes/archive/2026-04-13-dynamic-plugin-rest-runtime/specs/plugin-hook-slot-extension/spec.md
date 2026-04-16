## ADDED Requirements

### Requirement: 动态插件路由治理元数据集中在`g.Meta`

系统 SHALL 要求动态插件将后端动态路由的治理元数据集中定义在`api`层请求结构体的`g.Meta`中，避免额外引入第二套分散的路由治理配置源。

#### Scenario: 动态插件声明最小治理字段

- **WHEN** 开发者定义一个动态插件后端接口
- **THEN** 该接口可在`g.Meta`中声明`access`、`permission`、`operLog`
- **AND** `access`仅支持`public`和`login`
- **AND** 未声明`access`时按`login`处理

#### Scenario: 公开路由治理边界受限

- **WHEN** 开发者声明一个`public`动态路由
- **THEN** 该路由不得声明`permission`
- **AND** 该路由不得依赖宿主登录态注入
- **AND** 宿主装载阶段会拒绝非法配置

### Requirement: 动态插件权限声明复用宿主现有权限体系

系统 SHALL 将动态路由中的`permission`声明自动接入宿主现有的`sys_menu.perms`权限体系，而不是引入独立的动态权限存储模型。

#### Scenario: 动态路由权限被物化为隐藏菜单项

- **WHEN** 一个动态路由声明了合法的`permission`
- **THEN** 宿主在插件菜单同步阶段自动生成对应的隐藏权限菜单项
- **AND** 这些权限菜单项挂载在该插件专属的动态路由权限目录下
- **AND** 权限值直接复用动态路由声明的`permission`

#### Scenario: 动态路由权限随插件生命周期同步

- **WHEN** 动态插件被启用、禁用、卸载或切换激活版本
- **THEN** 宿主同步新增、更新或移除对应的隐藏权限菜单项
- **AND** 默认管理员角色继续自动拥有这些权限项

### Requirement: 动态插件不直接组合宿主治理中间件

系统 SHALL 保持动态插件为受限业务扩展模型，不向动态插件开放宿主`Auth`、`Ctx`、`OperLog`等治理中间件的自由拼装能力。

#### Scenario: 动态治理由宿主统一执行

- **WHEN** 一个动态插件路由被外部请求命中
- **THEN** 登录校验、权限校验和业务上下文注入由宿主基于路由合同统一执行
- **AND** 动态插件只声明治理需求，不直接调用宿主治理中间件

### Requirement: 动态插件复用公共 bridge 组件降低编写复杂度

系统 SHALL 将动态插件 bridge envelope、二进制 codec、guest 侧处理器适配、错误响应辅助等可复用逻辑抽象到`apps/lina-core/pkg`公共组件中，避免插件作者在每个动态插件中重复编写底层`ABI`与编解码样板。

#### Scenario: 插件运行时复用公共组件

- **WHEN** 开发者编写`backend/runtime/wasm`动态插件运行时
- **THEN** 开发者可以复用`apps/lina-core/pkg/pluginbridge`中的请求／响应信封、二进制编解码和处理器适配逻辑
- **AND** 插件业务代码主要实现路由处理函数，不需要重复实现底层内存读写与 bridge envelope 打包流程

#### Scenario: 公共组件不包含编译阶段流程

- **WHEN** 宿主、构建器或动态插件样例复用`apps/lina-core/pkg/pluginbridge`
- **THEN** 该组件只提供稳定合同、编解码、运行时辅助与无副作用校验逻辑
- **AND** 该组件不包含源码扫描、`go build`调用或产物写入等编译阶段流程
