## MODIFIED Requirements

### Requirement: 动态普通领域 host service 协议名必须与领域目录一致

系统 SHALL 要求动态插件普通领域`hostServices.service`协议名与已发布的动态领域目录名称保持一致。集合型领域 MUST 使用复数领域名：`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`。对应能力字符串 MUST 分别使用`host:users`、`host:files`、`host:jobs`、`host:notifications`、`host:plugins`和`host:sessions`。`i18n`不属于动态插件可声明的普通领域 host service；动态插件多语言资源由宿主统一管理。项目不保留旧单数 service 别名。

#### Scenario: 动态插件声明集合型领域服务

- **WHEN** 动态插件在`plugin.yaml`中声明用户、文件、任务、通知、插件治理或在线会话领域 host service
- **THEN** service 必须分别使用`users`、`files`、`jobs`、`notifications`、`plugins`或`sessions`
- **AND** 宿主 descriptor、授权快照、guest 调用和`WASM`dispatcher 必须使用同一 service 名

#### Scenario: 动态插件声明插件生命周期治理方法

- **WHEN** 动态插件在`plugin.yaml`中声明`service: plugins`和`method: lifecycle.tenant_delete.ensure`
- **THEN** 宿主校验该方法属于`plugins`领域已发布方法
- **AND** 运行时必须先校验`host:plugins`能力和授权快照中的精确 method
- **AND** 通过校验后才能进入`plugincap.LifecycleService`

#### Scenario: 动态插件声明单一命名空间领域服务

- **WHEN** 动态插件声明`auth`、`authz`、`apidoc`、`bizctx`、`dict`、`infra`、`route`、`ai`、`org`或`tenant`
- **THEN** service 继续使用该领域命名空间名称
- **AND** 不得为了形式统一将其机械复数化

#### Scenario: 动态插件声明 i18n host service

- **WHEN** 动态插件在`plugin.yaml`中声明`service: i18n`
- **THEN** host service 校验必须拒绝该声明
- **AND** public protocol alias、guest 公共目录和`WASM`dispatcher 不得暴露`i18n`服务

### Requirement: 动态插件普通领域 host service 必须覆盖已发布的动态普通领域能力

系统 SHALL 让动态插件通过`hostServices`获得已发布动态普通领域能力的覆盖。动态插件领域 host service MUST 使用语言无关的领域服务名和方法名，MUST 使用`resourceKind: none`表达方法授权，运行时 MUST 从宿主注入的同一个`capability.Services`目录进入对应`*cap.Service`。动态插件协议 MUST NOT 暴露`AdminServices`目录、数据库查询构造器、`DAO/DO/Entity`、HTTP 请求对象、宿主内部 service 或`i18n`运行时翻译服务。

#### Scenario: 动态插件声明普通领域能力

- **WHEN** 动态插件声明`auth`、`authz`、`users`、`dict`、`files`、`sessions`、`jobs`、`infra`、`apidoc`、`bizctx`、`route`、`notifications`、`plugins`、`ai`、`org`或`tenant`等领域 host service
- **THEN** 宿主清单校验 MUST 识别对应领域方法已经发布
- **AND** 声明 MUST 只包含`methods`
- **AND** 声明 MUST NOT 包含`resources`、`paths`、`tables`或`keys`
- **AND** 运行时授权快照 MUST 只按`service + method`校验调用

#### Scenario: 动态插件调用普通领域能力

- **WHEN** 动态插件通过已授权领域方法读取用户、权限、字典、文件、会话、任务、基础设施、API 文档、业务上下文、路由元数据、通知消息、插件治理、AI、组织或租户投影
- **THEN** `WASM`host service handler MUST 构造`CapabilityContext`
- **AND** handler MUST 使用`Capability.ServicesForPlugin(..., pluginID)`取得插件绑定的能力目录
- **AND** 请求 MUST 进入对应`*cap.Service`普通消费面或该普通消费面拥有的子服务
- **AND** 领域实现 MUST 继续执行租户、数据权限、可见性、批量上限、缓存和`i18n`治理

#### Scenario: 动态插件不能通过普通领域面获得管理能力

- **WHEN** 动态插件声明普通领域 host service
- **THEN** 宿主 MUST NOT 因该声明暴露创建、更新、删除、状态变更、授权关系变更、执行任务、撤销会话或发送通知等管理动作
- **AND** 未来如需动态插件管理方法，MUST 通过显式发布的领域管理方法和独立授权语义进入

#### Scenario: 重叠动态能力收敛到领域能力

- **WHEN** 既有动态 host service 与普通领域能力语义重叠
- **THEN** 宿主实现 MUST 优先复用`capability.Services`中对应领域能力或插件绑定子服务
- **AND** 不得继续维护一套与领域能力平行且语义漂移的动态专用实现

### Requirement: 动态插件 i18n 资源必须由宿主管理

系统 SHALL 允许动态插件继续通过`manifest/i18n`交付多语言资源，但运行时资源发现、合并、缓存、失效和前端语言包分发 MUST 由宿主统一管理。动态插件后端 MUST NOT 通过 host service 读取 locale、翻译消息或检索 message key，也 MUST NOT 自行读取`manifest/i18n`资源完成运行时翻译。

#### Scenario: 动态插件交付多语言资源

- **WHEN** 动态插件 artifact 包含`manifest/i18n/<locale>/*.json`资源
- **THEN** 宿主资源扫描和多语言聚合流程负责发现并合并这些资源
- **AND** 动态插件不需要声明`service: i18n`

#### Scenario: 动态插件返回可本地化业务结果

- **WHEN** 动态插件后端需要返回用户可见状态、错误或提示
- **THEN** 响应应返回稳定 message key、message params、英文 fallback 或原始业务数据
- **AND** 最终展示本地化由宿主错误治理、前端运行时语言包或宿主统一展示层完成
