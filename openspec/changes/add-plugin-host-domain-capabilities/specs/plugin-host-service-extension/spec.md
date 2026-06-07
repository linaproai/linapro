## ADDED Requirements

### Requirement: hostServices 必须支持领域服务和领域方法

系统 SHALL 允许动态插件通过`hostServices`声明宿主发布的领域服务和领域方法。领域协议服务名 MUST 使用语言无关的领域名，例如`user`、`authz`、`dict`、`org`、`tenant`、`plugin`和`ai`，不得使用 Go 包名或宿主内部实现名。每个领域方法 MUST 映射到领域能力接口或受控领域适配器。

#### Scenario: 动态插件声明用户领域读取

- **WHEN** 动态插件在`plugin.yaml`中声明`service: user`和`methods: [users.batch_get, users.search]`
- **THEN** 宿主校验该领域服务和方法已发布
- **AND** 安装授权确认后将归一化声明写入运行时授权快照

#### Scenario: 动态插件调用未知领域方法

- **WHEN** 动态插件调用未发布、未声明或未授权的领域方法
- **THEN** 宿主返回能力拒绝或能力不可用错误
- **AND** 不进入任何领域业务逻辑

### Requirement: host service 调用必须传递 CapabilityContext

系统 SHALL 在每一次动态`hostServices`领域调用中构造并传递`CapabilityContext`。该上下文 MUST 包含插件`ID`、执行来源、actor、tenant、授权快照、资源或投影标识、系统调用标识和审计摘要。

#### Scenario: 请求型 host service 调用

- **WHEN** 动态插件在请求路由中调用领域`host service`
- **THEN** WASM host service handler 将当前用户、租户、插件`ID`、路由来源和授权快照传入领域适配器
- **AND** 领域适配器基于上下文执行数据权限和审计治理

#### Scenario: 系统型 host service 调用

- **WHEN** 动态插件在生命周期、hook 或定时任务中调用领域`host service`
- **THEN** WASM host service handler 使用宿主创建的系统 actor 构造上下文
- **AND** 需要用户上下文的领域方法必须拒绝或按领域定义的系统调用边界执行

### Requirement: 动态领域管理方法使用安装授权模型

系统 SHALL 允许动态插件在`hostServices`中声明宿主显式发布的领域管理方法。安装或启用阶段确认授权后，运行时不再额外校验当前用户是否拥有对应工作台菜单或按钮权限；领域管理方法 MUST 继续校验目标资源可见性、租户边界、数据权限、状态机、数量上限和审计来源。

#### Scenario: 动态插件调用授权管理方法

- **WHEN** 动态插件调用已授权的领域管理方法
- **THEN** host service handler 校验`service + method`存在于运行时授权快照
- **AND** 请求进入对应领域`AdminService`或命令适配器
- **AND** 领域命令执行目标边界、状态机和审计校验

#### Scenario: 动态插件越权访问目标资源

- **WHEN** 动态插件已获方法授权但请求操作跨租户、不可见或状态不允许的目标资源
- **THEN** 领域方法拒绝该操作
- **AND** 响应使用结构化业务错误
- **AND** 宿主记录失败审计摘要
