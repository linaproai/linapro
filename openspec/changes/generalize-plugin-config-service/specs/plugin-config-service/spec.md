## ADDED Requirements

### Requirement: 插件配置服务提供通用只读配置访问能力

系统 SHALL 通过 `apps/lina-core/pkg/pluginservice/config` 向源码插件提供业务无关的只读配置访问服务。该服务 MUST 允许源码插件按任意配置 key 读取宿主配置文件内容，并 MUST NOT 为具体插件或业务模块暴露 `GetXxx()` 专用配置方法。

#### Scenario: 插件读取任意配置 key

- **WHEN** 源码插件通过插件配置服务读取一个存在的配置 key
- **THEN** 系统返回该 key 对应的配置值
- **AND** 配置服务不要求该 key 位于插件专属前缀下

#### Scenario: 公共组件不包含插件业务配置方法

- **WHEN** 新增或修改一个源码插件的私有配置结构
- **THEN** 开发者在插件内部定义配置结构、默认值和校验逻辑
- **AND** 不需要在 `apps/lina-core/pkg/pluginservice/config` 中新增插件专用的 `GetXxx()` 方法或插件业务配置类型

### Requirement: 插件配置服务支持结构体扫描和基础类型读取

系统 SHALL 支持源码插件通过通用配置服务将配置段扫描到调用方提供的结构体，并支持读取字符串、布尔值、整数和 `time.Duration` 等常用类型。缺失或空配置 key MUST 使用调用方提供的默认值。

#### Scenario: 插件扫描配置段到私有结构体

- **WHEN** 源码插件调用配置服务扫描一个存在的配置段
- **THEN** 系统将该配置段绑定到插件提供的结构体实例
- **AND** 插件可以在扫描后执行自己的默认值补齐和业务校验

#### Scenario: 插件读取基础类型配置并使用默认值

- **WHEN** 源码插件读取一个缺失或空的字符串、布尔值、整数或 duration 配置 key
- **THEN** 系统返回调用方传入的默认值
- **AND** 不因为缺失 key 返回失败

#### Scenario: Duration 配置解析失败

- **WHEN** 源码插件读取 duration 配置 key 且配置值不是合法的 duration 字符串
- **THEN** 系统返回显式错误
- **AND** 插件调用方可以自行决定 fail-fast、降级或向上包装错误

### Requirement: 插件配置服务保持只读边界

系统 SHALL 将插件配置服务限定为只读访问能力。该服务 MUST NOT 暴露写入、保存、热更新或运行时变更配置的方法。

#### Scenario: 插件只能读取配置

- **WHEN** 源码插件依赖 `apps/lina-core/pkg/pluginservice/config`
- **THEN** 该公共服务只提供读取、扫描和解析配置的方法
- **AND** 不提供修改配置文件或系统运行时配置的能力

### Requirement: 插件业务配置由插件内部维护

系统 SHALL 要求插件自己的配置结构、默认值、校验和业务语义维护在对应插件内部，而不是维护在宿主通用配置服务中。

#### Scenario: Monitor server 插件加载监控配置

- **WHEN** `monitor-server` 源码插件需要读取监控采集间隔和保留倍数
- **THEN** 插件通过通用配置服务读取现有监控配置 key
- **AND** 插件在自身内部维护监控配置结构、默认值和校验逻辑
- **AND** 宿主通用配置服务不暴露 `MonitorConfig` 或 `GetMonitor()`

### Requirement: 插件配置读取不新增分布式缓存一致性负担

系统 SHALL 将本次插件配置服务能力限定在静态配置读取范围内，不新增运行时可变配置缓存。若后续扩展为运行时可变配置读取，系统 MUST 另行设计跨实例修订号、广播失效、共享缓存或等价的集群一致性机制。

#### Scenario: 读取静态配置文件

- **WHEN** 源码插件通过插件配置服务读取静态配置文件内容
- **THEN** 系统可以复用宿主现有配置读取机制
- **AND** 不需要新增分布式缓存失效或跨实例刷新机制

#### Scenario: 后续扩展运行时可变配置

- **WHEN** 插件配置服务未来需要读取可在运行时修改的配置数据
- **THEN** 设计必须明确权威数据源、一致性模型、失效触发点、跨实例同步机制和故障降级策略
- **AND** 不得仅依赖单节点进程内缓存保证集群一致性

### Requirement: 动态插件通过 config host service 读取完整静态配置

系统 SHALL 通过动态插件 `hostServices` 授权模型提供 `config` host service，使动态插件可以通过 `lina_env.host_call` 读取宿主 GoFrame 静态配置内容。该 host service MUST 允许读取完整配置，不要求配置 key 位于插件专属前缀或资源白名单下，并 MUST 保持只读。

#### Scenario: 动态插件声明配置读取服务

- **WHEN** 动态插件在 `plugin.yaml` 中声明 `service: config` 且 `methods` 为 `get`、`exists`、`string`、`bool`、`int`、`duration` 的任意非空子集
- **THEN** 系统接受该 host service 声明并从声明中推导配置读取能力
- **AND** 插件运行时授权快照包含 `config` host service

#### Scenario: 动态插件省略配置读取方法

- **WHEN** 动态插件在 `plugin.yaml` 中声明 `service: config` 且未填写 `methods`
- **THEN** 系统接受该 host service 声明并默认授权当前完整配置只读方法集合
- **AND** 授权快照中的 config host service 方法被规范化为 `get`、`exists`、`string`、`bool`、`int`、`duration`

#### Scenario: 动态插件声明不支持的配置方法

- **WHEN** 动态插件在 `plugin.yaml` 的 `config` host service 中声明 `get`、`exists`、`string`、`bool`、`int`、`duration` 之外的方法
- **THEN** 系统拒绝该 host service 声明
- **AND** 不为写入、保存、热更新或运行时配置变更方法派生配置能力

#### Scenario: 动态插件使用 guest SDK 便捷读取方法

- **WHEN** 动态插件代码通过 guest SDK 调用 `Exists`、`String`、`Bool`、`Int` 或 `Duration` 等配置便捷方法
- **THEN** guest SDK 通过对应的 `config.exists`、`config.string`、`config.bool`、`config.int` 或 `config.duration` 发起 host_call
- **AND** 宿主在只读配置服务内完成存在性判断或类型解析

#### Scenario: 动态插件读取任意配置 key

- **WHEN** 已授权动态插件通过 `config.get` 读取任意存在的配置 key
- **THEN** 系统返回该 key 对应配置值的 JSON 表示
- **AND** 不根据插件 ID、前缀或 key pattern 拒绝读取

#### Scenario: 动态插件读取完整配置快照

- **WHEN** 已授权动态插件通过 `config.get` 传入空 key 或 `.`
- **THEN** 系统返回 GoFrame 配置系统暴露的完整静态配置快照 JSON

#### Scenario: 动态插件读取缺失配置 key

- **WHEN** 已授权动态插件通过任一 config 只读方法读取不存在的配置 key
- **THEN** 系统返回 found=false 的结果
- **AND** 不将缺失 key 视为 host_call 失败

#### Scenario: 动态插件配置服务保持只读

- **WHEN** 动态插件声明或调用 `config` host service
- **THEN** host service 权限声明只支持 `get`、`exists`、`string`、`bool`、`int`、`duration` 这些只读方法
- **AND** 不支持写入、保存、热更新或运行时配置变更方法
