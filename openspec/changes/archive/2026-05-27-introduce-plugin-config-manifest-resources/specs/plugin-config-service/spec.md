## MODIFIED Requirements

### Requirement:插件配置服务提供通用只读配置访问

系统 SHALL 通过`apps/lina-core/pkg/pluginservice/config`向源码插件提供业务无关的只读插件配置访问服务。该服务 MUST 默认作用域到当前插件 ID，只允许源码插件读取当前插件自己的运行期配置，不得暴露任意宿主静态配置树，也不得要求插件配置键位于宿主全局配置前缀下。

#### Scenario:源码插件读取自身配置键

- **WHEN** 源码插件`plugin-a`通过插件配置服务读取`storage.endpoint`
- **AND** `plugin-a`的运行期配置中存在该键
- **THEN** 系统返回`plugin-a`作用域下的配置值
- **AND** 插件调用方不需要在读取键中携带`plugin-a`前缀

#### Scenario:源码插件不能通过配置服务读取宿主私有配置

- **WHEN** 源码插件通过插件配置服务读取宿主私有配置键，例如`database.default.link`或`server.jwt.secret`
- **AND** 该键不属于当前插件运行期配置来源
- **THEN** 系统不得返回宿主全局配置值
- **AND** 调用结果 MUST 表达为未找到或拒绝访问，而不是回退到宿主`g.Cfg()`完整配置树

#### Scenario:公共组件不包含插件业务配置方法

- **WHEN** 为源码插件添加或修改私有配置结构
- **THEN** 开发者在插件内部定义配置结构、默认值和验证逻辑
- **AND** 无需在`apps/lina-core/pkg/pluginservice/config`中添加插件特定的`GetXxx()`方法或插件业务配置类型

### Requirement:插件配置服务支持结构体扫描和基本类型读取

系统 SHALL 支持源码插件通过通用配置服务将当前插件配置段扫描到调用方提供的结构体中，并支持读取字符串、布尔值、整数和`time.Duration`等常见类型。缺失或空白的配置键必须使用调用方提供的默认值，且所有读取均限定在当前插件配置作用域内。

#### Scenario:插件将自身配置段扫描到私有结构体

- **WHEN** 源码插件调用配置服务扫描`storage`配置段
- **AND** 当前插件运行期配置中存在`storage`段
- **THEN** 系统将该段绑定到插件提供的结构体实例
- **AND** 插件可在扫描后应用自己的默认值和业务验证

#### Scenario:插件读取带默认值的基本类型配置

- **WHEN** 源码插件读取缺失或空白的字符串、布尔值、整数或时长配置键
- **THEN** 系统返回调用方提供的默认值
- **AND** 缺失的键不会导致失败

#### Scenario:时长配置解析失败

- **WHEN** 源码插件读取值不是有效时长字符串的时长配置键
- **THEN** 系统返回明确错误
- **AND** 插件调用方可选择快速失败行为、回退行为或上游错误包装

### Requirement:插件业务配置在插件内部维护

系统 SHALL 要求每个插件的配置结构、默认值、验证和业务语义在该插件内部维护，而非在宿主通用配置服务中。插件源码目录 SHOULD 使用`manifest/config/config.example.yaml`作为可提交模板文件，使用`manifest/config/config.yaml`作为本地或部署环境的实际运行配置文件；`config.yaml`默认不得提交到源码仓库，`config.example.yaml`不得被宿主自动当作运行期默认值加载。

#### Scenario:监控服务插件加载监控配置

- **WHEN** `linapro-monitor-server`源码插件需要读取监控采集间隔和保留倍数
- **THEN** 插件通过通用配置服务读取自身`manifest/config/config.yaml`或生产部署配置中的监控配置键
- **AND** 插件在内部维护监控配置结构、默认值和验证逻辑
- **AND** 宿主通用配置服务不暴露`MonitorConfig`或`GetMonitor()`

#### Scenario:宿主配置不再维护插件业务配置副本

- **WHEN** 插件已经通过自己的`manifest/config/config.example.yaml`声明业务配置结构
- **THEN** 宿主`apps/lina-core/manifest/config`不得继续维护该插件业务配置副本
- **AND** 宿主仍可保留`plugin.autoEnable`、`plugin.dynamic.storagePath`、`plugin.allowForceUninstall`等插件治理配置

#### Scenario:模板配置不参与运行期读取

- **WHEN** 插件仅提供`manifest/config/config.example.yaml`但未提供任何实际运行配置
- **THEN** 插件配置服务不得自动把模板文件内容作为运行期配置返回
- **AND** 插件业务默认值由插件代码或读取方法默认值承担

#### Scenario:实际配置文件默认不提交

- **WHEN** 插件仓库维护`manifest/config/config.example.yaml`模板
- **THEN** 仓库忽略`manifest/config/config.yaml`
- **AND** 开发者可以在本地创建`manifest/config/config.yaml`进行调试
- **AND** 不会把本地实际配置或敏感值作为插件源码资源提交

### Requirement:插件配置读取不增加分布式缓存一致性负担

系统 SHALL 将此插件配置服务能力限制为只读配置读取，不得添加运行时写入、保存或热重载配置能力。开发阶段可读取插件源码目录下的实际配置文件；生产阶段配置变更默认通过部署、重启或未来显式 reload 流程生效。如果未来扩展到运行时可变配置读取或热更新，系统必须单独设计跨实例修订号、广播失效、共享缓存或等效的集群一致性机制。

#### Scenario:读取静态插件配置文件

- **WHEN** 源码插件或动态插件通过插件配置服务读取自身运行期配置
- **THEN** 系统可复用宿主构造的插件作用域配置 resolver
- **AND** 不需要因为单次只读读取新增分布式缓存失效或跨实例刷新机制

#### Scenario:动态发布默认配置绑定发布版本

- **WHEN** 动态插件从 artifact 中的`manifest/config/config.yaml`读取默认运行配置
- **THEN** 该配置视图 MUST 绑定当前 active release 的 checksum 或 generation
- **AND** 动态插件升级、禁用、卸载或同版本刷新后不得继续使用旧发布的默认配置视图

#### Scenario:未来运行时可变配置扩展

- **WHEN** 插件配置服务未来需要读取可在运行时修改且不经重启生效的配置数据
- **THEN** 设计必须定义权威数据源、一致性模型、失效触发点、跨实例同步机制和故障回退策略
- **AND** 不得仅依赖单节点进程内缓存来保证集群一致性

## ADDED Requirements

### Requirement:插件运行时配置文件必须遵循插件作用域目录约定

系统 SHALL 以插件 ID 为作用域解析插件运行时配置。开发阶段实际配置文件位于`apps/lina-plugins/<plugin-id>/manifest/config/config.yaml`；生产阶段推荐实际配置文件位于宿主配置根的`plugins/<plugin-id>/config.yaml`，例如`/opt/linapro/config/plugins/<plugin-id>/config.yaml`。插件配置文件名 MUST 保持为`config.yaml`，并不得回流到`apps/lina-core/manifest/config`。

#### Scenario:开发阶段读取插件源码目录配置

- **WHEN** 宿主以开发工作区方式运行源码插件`plugin-a`
- **AND** `apps/lina-plugins/plugin-a/manifest/config/config.yaml`存在
- **THEN** `HostServices.Config()`读取该文件作为`plugin-a`的实际运行配置
- **AND** 宿主不得要求开发者把该配置复制到`apps/lina-core/manifest/config`

#### Scenario:生产阶段读取宿主配置根下的插件配置

- **WHEN** 生产部署配置根为`/opt/linapro/config`
- **AND** `/opt/linapro/config/plugins/plugin-a/config.yaml`存在
- **THEN** `HostServices.Config()`优先读取该文件作为`plugin-a`的实际运行配置
- **AND** 该路径与宿主 GoFrame 配置目录约定保持一致但仍按插件 ID 隔离

#### Scenario:生产配置覆盖发布默认配置

- **WHEN** 动态插件 artifact 携带`manifest/config/config.yaml`
- **AND** 生产配置根下同时存在`plugins/<plugin-id>/config.yaml`
- **THEN** 系统使用生产外部配置作为实际运行配置
- **AND** artifact 中的配置仅作为发布默认配置或开发示例来源

### Requirement:宿主公开配置必须通过独立服务读取

系统 SHALL 通过`HostServices.HostConfig()`向插件暴露宿主公开配置。该服务 MUST 使用宿主显式白名单控制可读取键，不得与`HostServices.Config()`使用相同的完整 API，不得支持完整宿主配置扫描、根节点读取或任意宿主配置键读取。

#### Scenario:插件读取白名单宿主公开配置

- **WHEN** 插件通过`HostServices.HostConfig()`读取白名单键`workspace.basePath`
- **THEN** 系统返回该公开宿主配置值
- **AND** 该读取不要求插件访问宿主完整配置树

#### Scenario:插件读取未公开宿主配置被拒绝

- **WHEN** 插件通过`HostServices.HostConfig()`读取未在白名单中的宿主键
- **THEN** 系统拒绝该读取或返回未找到
- **AND** 系统不得回退到宿主`g.Cfg()`任意 key 读取

#### Scenario:HostConfig 不提供完整扫描

- **WHEN** 插件尝试通过`HostServices.HostConfig()`读取空键、`.`或请求完整数据快照
- **THEN** 系统拒绝该请求
- **AND** `HostConfig()`不得暴露`Scan`、`Data`或等价完整树读取方法

### Requirement:动态插件通过配置宿主服务读取插件作用域配置

系统 SHALL 通过动态插件`hostServices`授权模型提供`config`宿主服务，使动态插件可通过统一 host service 协议读取当前插件自己的运行期配置。动态插件配置服务 MUST 只支持`get`读取动作，guest SDK 的`Exists`、`String`、`Bool`、`Int`、`Duration`或`Scan`等 helper 必须映射到`config.get`并在 guest SDK 或共享适配层完成类型转换。

#### Scenario:动态插件声明插件配置读取服务

- **WHEN** 动态插件在`plugin.yaml`中声明`service: config`且`methods: [get]`
- **THEN** 系统接受宿主服务声明并从中派生插件配置读取能力
- **AND** 插件运行时授权快照包含`config.get`

#### Scenario:动态插件声明 typed 配置方法被拒绝

- **WHEN** 动态插件在`plugin.yaml`的`config`宿主服务中声明`exists`、`string`、`bool`、`int`、`duration`或其他非`get`方法
- **THEN** 系统拒绝宿主服务声明
- **AND** 系统提示这些 typed 方法应由 guest SDK helper 映射到`get`

#### Scenario:动态插件读取自身配置键

- **WHEN** 已授权的动态插件通过`config.get`读取`storage.endpoint`
- **THEN** 系统返回当前插件作用域下该键配置值的 JSON 表示
- **AND** 不按宿主全局配置树解析该键

#### Scenario:动态插件不能读取完整宿主配置快照

- **WHEN** 已授权的动态插件向`config.get`传递空键或`.`
- **THEN** 系统拒绝该读取
- **AND** 系统不得返回 GoFrame 配置系统暴露的完整静态配置快照

#### Scenario:动态插件读取缺失配置键

- **WHEN** 已授权的动态插件通过`config.get`读取当前插件配置中缺失的键
- **THEN** 系统返回`found=false`结果
- **AND** 不将缺失的键视为 host service 调用失败

## REMOVED Requirements

### Requirement:动态插件通过配置宿主服务读取完整静态配置

**Reason**: 完整宿主静态配置读取会暴露宿主私有配置和其他插件配置，不符合插件作用域隔离与动态插件授权最小化要求。

**Migration**: 动态插件通过`config.get`读取当前插件自己的运行期配置；确实需要宿主公开配置时，通过`hostConfig.get`读取宿主白名单键。
