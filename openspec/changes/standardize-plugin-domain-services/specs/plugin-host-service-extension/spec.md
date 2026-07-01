## ADDED Requirements

### Requirement: 动态 host service registry 必须作为动态可声明性唯一来源

系统 SHALL 将动态`host service registry`作为动态插件可声明`service + method`的唯一事实来源。方法是否注册决定动态插件能否在`plugin.yaml hostServices`中声明该方法；系统 MUST NOT 通过`dynamic-auth`、`source-only`、`reserved`或等价方法可声明性字段表达方法可见性。

#### Scenario: 已注册方法进入授权流程

- **WHEN** 动态插件声明的`service + method`存在于动态`host service registry`
- **THEN** 构建器或宿主清单校验允许该声明进入安装、启用或升级授权确认流程
- **AND** 运行时仍必须校验授权快照和资源范围

#### Scenario: 未注册方法被拒绝

- **WHEN** 动态插件声明或调用不存在于动态`host service registry`的`service + method`
- **THEN** 构建、安装、启用或运行时必须拒绝该方法
- **AND** 请求不得进入 dispatcher 业务处理器、`capability.Services`或领域 owner

#### Scenario: Go Service 方法未注册为动态方法

- **WHEN** 某个统一 Go `Service`方法未注册到动态`host service registry`
- **THEN** 该方法天然不能被动态插件声明或调用
- **AND** 系统不得要求维护额外字段标记该方法为源码插件专用

### Requirement: 动态插件领域能力暴露必须按 registry、示例和风险分层治理

系统 SHALL 在新增或补齐动态插件领域能力暴露时区分三类事实：动态`host service registry`已注册的方法、示例动态插件已声明并验证的方法，以及统一 Go `Service`存在但未动态发布的方法。已注册但示例未声明的方法 MAY 通过示例和 smoke 测试补齐；统一 Go `Service`存在但未注册的方法 MUST 先按只读投影、源码插件专用、高风险写入/执行或缺少稳定 owner 分类，不得因为 Go 方法存在就直接开放为动态 wire method。

#### Scenario: 补齐已注册但示例未声明的方法

- **WHEN** 某个领域方法已经存在于动态`host service registry`和 catalog，但`linapro-demo-dynamic`未在`plugin.yaml hostServices`中声明
- **THEN** 可以只补充示例插件声明、示例调用和 smoke 验证
- **AND** 不得新增平行 wire method、绕过授权快照或改变领域 owner 实现

#### Scenario: 评估未发布的统一 Go Service 方法

- **WHEN** 某个统一 Go `Service`方法不存在于动态`host service registry`
- **THEN** OpenSpec 设计、任务记录或 README 必须说明该方法暂不动态开放的分类原因
- **AND** 写入、删除、撤销、执行、内容流、系统配置修改和跨领域治理类方法必须先补充事务、审计、缓存失效、租户隔离、幂等、资源成本和错误暴露策略，再考虑注册为动态方法

#### Scenario: 租户生命周期和成员替换不进入动态注册

- **WHEN** 租户 owner 内部存在租户创建、更新、状态变更、删除或用户租户成员关系替换能力
- **THEN** 这些 owner 内部能力不得因为存在 Go provider 方法就进入动态`host service registry`
- **AND** 动态插件只能声明已经注册的租户目录读取、成员校验、成员列表、可见性校验、租户插件治理和过滤上下文方法
- **AND** 若未来需要开放租户生命周期或成员替换动态方法，必须先形成独立的事务、审计、数据权限、租户边界、缓存失效和错误暴露设计

#### Scenario: 对齐 guest helper 与已注册根方法

- **WHEN** guest SDK 的子能力 helper 本地返回不可用，但同等语义已经通过已注册根方法提供
- **THEN** 可以将 helper 委托到已注册根方法
- **AND** 委托后仍必须使用同一动态授权快照、资源范围和结构化错误映射

### Requirement: 插件可见目录的可见性校验必须保持单一批量签名

系统 SHALL 为动态 host service 和源码插件目录的可见性校验保持单一`EnsureVisible(ctx, ids []...) error`签名。方法可以在实现层兼容单个 ID 便捷调用，但公共契约 MUST NOT 同时暴露`EnsureVisible`与`EnsureVisibleMany`两个并列入口。

#### Scenario: 动态插件调用目录可见性校验

- **WHEN** 动态插件通过 host service 调用组织或租户目录可见性校验
- **THEN** host service MUST 只暴露一个批量签名的`EnsureVisible`方法
- **AND** 传入单个目标时也必须以长度为 1 的 ID 切片编码
- **AND** 不得再保留`EnsureVisibleMany`

### Requirement: 动态 wire method 必须一次性标准化且不得保留兼容别名

系统 SHALL 对动态 wire method 进行一次性破坏性标准化。标准化后的 method 名称 MUST 按领域和子资源表达稳定语义；旧 wire method 名称 MUST 删除，不得以兼容别名、双注册或隐式 fallback 继续保留。

#### Scenario: 标准化用户批量读取方法

- **WHEN** 动态插件声明用户批量读取能力
- **THEN** 宿主接受标准化 method，例如`users.batch_get`
- **AND** 宿主不得继续接受旧别名或历史 method 名

#### Scenario: 静态检索发现旧 wire method

- **WHEN** 治理扫描发现生产代码、guest client、dispatcher、catalog、README 或示例插件继续声明旧 wire method 别名
- **THEN** 验证或审查必须失败
- **AND** 调用方必须迁移到标准化 method 名称

### Requirement: 插件治理动作必须通过领域组件注册为动态 host service

系统 SHALL 将插件治理动作作为领域能力组件的一部分注册为动态 host service。插件生命周期治理归属`plugins.lifecycle`，插件启用状态读取归属`plugins.state`，插件配置归属`plugins`领域配置方法；租户插件启停、默认供给和插件表租户过滤归属`tenant`领域。动态插件能否声明和调用这些治理方法只由动态`host service registry`注册事实、`plugin.yaml hostServices`声明、安装或启用授权和运行时授权快照决定；系统 MUST NOT 通过`pluginhost.Services`顶层方法、`pluginbridge`平行业务接口或未归属领域 owner 的 dispatcher 分支暴露治理动作。

#### Scenario: 动态插件读取插件注册表

- **WHEN** 动态插件声明并获授权调用`plugins.registry.list`或等价注册表读取方法
- **THEN** 宿主允许其读取当前上下文可见的插件注册表投影
- **AND** 响应不得隐式授予未声明、未授权或未注册的插件治理动作

#### Scenario: 动态插件声明已注册治理方法

- **WHEN** 动态插件声明`plugins.lifecycle.tenant_delete.ensure`、`tenant.plugins.enabled.set`或等价已注册治理方法
- **THEN** 宿主允许该声明进入安装或启用授权确认流程
- **AND** 运行时 dispatcher 必须在进入领域 owner 前校验授权快照和资源范围

#### Scenario: 动态插件声明未注册治理方法

- **WHEN** 动态插件声明`plugins.lifecycle.set_enabled`、`plugins.install`、`plugins.uninstall`或等价治理生命周期方法
- **THEN** 若该方法未注册到动态`host service registry`，构建、安装或启用校验必须拒绝该声明
- **AND** 运行时请求不得进入 dispatcher 业务处理器、`capability.Services`或领域 owner

## MODIFIED Requirements

### Requirement: hostServices 必须支持领域服务和领域方法

系统 SHALL 允许动态插件通过`hostServices`声明宿主注册到动态`host service registry`的领域服务和领域方法。领域协议服务名 MUST 使用语言无关的领域名，并且普通领域 service 名 MUST 与`pkg/plugin/capability.Services`领域目录名称保持一致；集合型领域使用`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`，命名空间型领域继续使用`authz`、`dict`、`org`、`tenant`、`ai`等领域名。领域协议名不得使用 Go 包名或宿主内部实现名。每个已注册领域方法 MUST 映射到统一领域`Service`或受控领域适配器。

#### Scenario: 动态插件声明用户领域读取

- **WHEN** 动态插件在`plugin.yaml`中声明`service: users`和`methods: [users.batch_get]`
- **THEN** 宿主校验该领域服务和方法已注册到动态`host service registry`
- **AND** 安装授权确认后将归一化声明写入运行时授权快照

#### Scenario: 动态插件调用未知领域方法

- **WHEN** 动态插件调用未注册、未声明或未授权的领域方法
- **THEN** 宿主返回能力拒绝或能力不可用错误
- **AND** 不进入任何领域业务逻辑

### Requirement: 动态领域管理方法使用安装授权模型

系统 SHALL 允许动态插件在`hostServices`中声明宿主已注册的领域管理方法。安装或启用阶段确认授权后，运行时不再额外校验当前用户是否拥有对应工作台菜单或按钮权限；领域管理方法 MUST 继续校验目标资源可见性、租户边界、数据权限、状态机和数量上限。动态领域管理方法 MUST 进入对应统一`Service`，不得依赖`AdminService`或`Services.Admin()`目录。

#### Scenario: 动态插件调用授权管理方法

- **WHEN** 动态插件调用已注册且已授权的领域管理方法
- **THEN** host service handler 校验`service + method`存在于运行时授权快照
- **AND** 请求进入对应领域统一`Service`
- **AND** 领域方法执行目标边界、状态机和数量上限校验

#### Scenario: 动态插件越权访问目标资源

- **WHEN** 动态插件已获方法授权但请求操作跨租户、不可见或状态不允许的目标资源
- **THEN** 领域方法拒绝该操作
- **AND** 响应使用结构化业务错误
- **AND** 响应不暴露不可见资源细节

### Requirement: 宿主服务适配器必须适配到`*cap`能力组件

系统 SHALL 要求源码插件 hostservices directory、动态插件`WASM`host service handler 和 guest SDK 最终适配到`pkg/plugin/capability/<domain>cap`能力组件。适配层 MUST 不再依赖`capability/contract`具体能力服务接口作为运行时服务目录契约，也不得依赖`AdminService`作为管理动作入口。

#### Scenario: 源码插件构造宿主服务目录

- **WHEN** 宿主启动期构造源码插件可消费的`capability.Services`
- **THEN** directory 字段和方法返回目标`*cap.Service`
- **AND** 认证授权能力以`authcap.Service`能力族聚合入口发布，token 子服务归属`authcap/token`，授权子服务归属`authcap/authz`
- **AND** 插件作用域配置 factory 通过`Plugins().Config()`对应子服务发布
- **AND** 其他插件作用域能力 factory 使用对应`manifestcap`或其他目标组件
- **AND** 源码插件同进程租户表过滤能力只能通过`tenantspi.ApplyPluginTableFilter(ctx, filter, model, qualifier)`helper 使用，普通`tenantcap.Service.Filter()`仅提供可序列化上下文，动态插件等价过滤通过可序列化 host service 或`RecordStore`治理完成
- **AND** 宿主服务目录不得继续提供`AdminService`或`Services.Admin()`入口

#### Scenario: 动态 host service 调用领域能力

- **WHEN** 动态插件通过`hostServices`调用`service + method`
- **THEN** `WASM`host service handler 校验 registry、授权和资源后进入对应`*cap.Service`
- **AND** handler 不得通过旧`contract.*Service`、`*cap.AdminService`或其他平行管理目录绕过目标能力组件

### Requirement: Go 包重命名不得改变动态插件协议

系统 SHALL 将`*cap`包重命名和统一`Service`收敛视为 Go 公共包边界重构。除本次明确标准化的 wire method 外，动态插件`plugin.yaml hostServices`声明、运行时授权快照、`service`字符串、资源授权、protobuf envelope 和错误 envelope MUST 保持当前目标模型不变。

#### Scenario: 动态插件声明 AI 文本能力

- **WHEN** 动态插件声明`service: ai`和`method: text.generate`
- **THEN** 宿主仍按`service: ai`和`method: text.generate`校验授权
- **AND** Go 侧能力组件重命名为`aicap`不得要求插件清单改为`service: aicap`

#### Scenario: 动态插件读取插件配置

- **WHEN** 动态插件声明`service: plugins`和标准化插件配置读取 method
- **THEN** 宿主按标准化`plugins.config.*`方法和授权快照执行校验
- **AND** Go 侧插件配置能力收口到`Plugins().Config()`不得改变插件配置属于`plugins`领域的事实

#### Scenario: 动态插件读取宿主配置

- **WHEN** 动态插件声明`service: hostConfig`和授权配置 key
- **THEN** 宿主仍按宿主配置授权快照执行校验
- **AND** Go 侧宿主配置能力使用`HostConfig()`，不得与`Plugins().Config()`混用
