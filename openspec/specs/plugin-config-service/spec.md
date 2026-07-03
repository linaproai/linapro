# 插件配置服务规范

## Purpose

定义 LinaPro 向源码插件和动态插件暴露的通用只读配置访问契约。该能力允许插件通过业务无关的 API 读取宿主静态配置，同时将插件特定的配置结构、默认值和验证逻辑保留在各插件内部。
## Requirements
### Requirement:插件配置服务提供通用只读配置访问

系统 SHALL 通过`apps/lina-core/pkg/plugin/capability/config`向源码插件提供业务无关的只读配置访问服务。该服务必须允许源码插件通过任意配置键读取宿主配置文件内容，不得为特定插件或业务模块暴露插件特定的`GetXxx()`配置方法。

#### Scenario:插件读取任意配置键

- **WHEN** 源码插件通过插件配置服务读取现有配置键
- **THEN** 系统返回该键的配置值
- **AND** 配置服务不要求键位于插件特定前缀下

#### Scenario:公共组件不包含插件业务配置方法

- **WHEN** 为源码插件添加或修改私有配置结构
- **THEN** 开发者在插件内部定义配置结构、默认值和验证逻辑
- **AND** 无需在`apps/lina-core/pkg/plugin/capability/config`中添加插件特定的`GetXxx()`方法或插件业务配置类型

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

### Requirement:插件配置服务保持只读边界

系统 SHALL 将插件配置服务限制为只读访问。该服务不得暴露在运行时写入、保存、热重载或变更配置的方法。

#### Scenario:插件只能读取配置

- **当** 源码插件依赖 `apps/lina-core/pkg/pluginservice/config` 时
- **则** 公共服务仅提供读取、扫描和解析配置的方法
- **且** 不提供修改配置文件或系统运行时配置的能力

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

### Requirement:插件运行时配置文件必须遵循插件作用域目录约定

系统 SHALL 以插件 ID 为作用域解析插件运行时配置。插件配置服务 MUST 优先读取宿主主静态配置文件中的`plugin.<plugin-id>`配置段；当该配置段不存在时，生产阶段实际配置文件位于宿主配置根的`plugins/<plugin-id>/config.yaml`，例如`/opt/linapro/config/plugins/<plugin-id>/config.yaml`；开发阶段实际配置文件位于`apps/lina-plugins/<plugin-id>/manifest/config/config.yaml`；动态插件 artifact 可提供`manifest/config/config.yaml`作为发布默认配置。插件独立配置文件名 MUST 保持为`config.yaml`。插件业务配置不得通过宿主通用配置服务新增插件专用`GetXxx()`方法，也不得把插件业务默认值维护为宿主治理配置。

#### Scenario: 主框架静态插件配置优先于生产配置文件

- **WHEN** 宿主主静态配置中存在`plugin.plugin-a.storage.endpoint`
- **AND** 生产配置根下同时存在`plugins/plugin-a/config.yaml`且包含`storage.endpoint`
- **THEN** `Services.Plugins().Config()`读取`storage.endpoint`时返回主静态配置中`plugin.plugin-a.storage.endpoint`的值
- **AND** 系统不从生产配置文件补齐该插件配置段内的单个缺失 key

#### Scenario: 主框架静态插件配置缺失时读取生产配置文件

- **WHEN** 宿主主静态配置中不存在`plugin.plugin-a`
- **AND** 生产部署配置根为`/opt/linapro/config`
- **AND** `/opt/linapro/config/plugins/plugin-a/config.yaml`存在
- **THEN** `Services.Plugins().Config()`读取该文件作为`plugin-a`的实际运行配置
- **AND** 该路径与宿主 GoFrame 配置目录约定保持一致但仍按插件 ID 隔离

#### Scenario: 生产配置文件缺失时读取插件源码目录配置

- **WHEN** 宿主以开发工作区方式运行源码插件`plugin-a`
- **AND** 宿主主静态配置中不存在`plugin.plugin-a`
- **AND** 生产配置根下不存在`plugins/plugin-a/config.yaml`
- **AND** `apps/lina-plugins/plugin-a/manifest/config/config.yaml`存在
- **THEN** `Services.Plugins().Config()`读取该文件作为`plugin-a`的实际运行配置
- **AND** 宿主不得要求开发者把该配置复制到`apps/lina-core/manifest/config`

#### Scenario: 外部配置覆盖发布默认配置

- **WHEN** 动态插件 artifact 携带`manifest/config/config.yaml`
- **AND** 宿主主静态配置中存在`plugin.<plugin-id>`
- **THEN** 系统使用宿主主静态配置中的`plugin.<plugin-id>`作为实际运行配置
- **AND** artifact 中的配置仅作为发布默认配置来源

#### Scenario: 文件配置覆盖发布默认配置

- **WHEN** 动态插件 artifact 携带`manifest/config/config.yaml`
- **AND** 宿主主静态配置中不存在`plugin.<plugin-id>`
- **AND** 生产配置根下同时存在`plugins/<plugin-id>/config.yaml`
- **THEN** 系统使用生产外部配置作为实际运行配置
- **AND** artifact 中的配置仅作为发布默认配置来源

#### Scenario: 模板配置不参与运行期读取

- **WHEN** 插件仅提供`manifest/config/config.example.yaml`但未提供任何实际运行配置
- **AND** 宿主主静态配置中不存在`plugin.<plugin-id>`
- **THEN** 插件配置服务不得自动把模板文件内容作为运行期配置返回
- **AND** 插件业务默认值由插件代码或读取方法默认值承担

### Requirement:宿主公开配置必须通过独立服务读取

系统 SHALL 通过 `HostServices.HostConfig()` 向源码插件暴露宿主配置只读读取能力。源码插件通过该服务读取宿主配置时不得受公开 key 白名单限制；空 key 或 `.` MUST 按宿主配置组件语义返回完整静态配置快照。非 root 配置键的读取顺序 MUST 为当前上下文可见的`sys_config`有效快照、GoFrame 当前静态配置源中的`config.yaml`值、系统已有默认值、`nil`。该服务不得提供写入、保存、热重载或运行时修改宿主配置的方法。动态插件通过 `hostconfig.get` 读取宿主配置时，仍 MUST 先通过 `hostServices` 授权快照校验对应 key，动态插件不得绕过 manifest 授权读取宿主配置。

#### Scenario:源码插件读取任意宿主配置键

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取宿主配置键 `database.default.link`
- **THEN** 系统按宿主当前配置源返回该键的配置值
- **AND** 该读取不要求 key 预先登记到公开白名单

#### Scenario:源码插件优先读取系统配置快照

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取 `custom.feature.limit`
- **AND** 当前上下文可见的`sys_config`中存在`custom.feature.limit`
- **AND** 静态`config.yaml`中也存在`custom.feature.limit`
- **THEN** 系统返回`sys_config`中的有效值

#### Scenario:源码插件读取静态配置 fallback

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取 `workspace.basePath`
- **AND** 当前上下文可见的`sys_config`中不存在`workspace.basePath`
- **AND** 静态`config.yaml`中存在`workspace.basePath`
- **THEN** 系统返回静态配置值

#### Scenario:源码插件读取系统默认值 fallback

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取 `sys.jwt.expire`
- **AND** 当前上下文可见的`sys_config`中不存在`sys.jwt.expire`
- **AND** 静态`config.yaml`中不存在`sys.jwt.expire`
- **AND** 系统默认值元数据存在`sys.jwt.expire`
- **THEN** 系统返回系统默认值
- **AND** 该读取不要求 key 预先登记到公开白名单

#### Scenario:源码插件读取缺失宿主配置键

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取不存在的宿主配置键
- **AND** `sys_config`、静态`config.yaml`和系统默认值元数据都没有该 key
- **THEN** 系统返回未找到语义
- **AND** 不因 key 未登记到白名单而返回权限错误

#### Scenario:源码插件读取完整宿主配置快照

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取空 key 或 `.`
- **THEN** 系统返回宿主当前配置源中的完整配置快照
- **AND** 该读取不要求逐个 key 预先登记到公开白名单

#### Scenario:动态插件宿主配置读取仍受授权快照限制

- **WHEN** 动态插件通过 `hostconfig.get` 读取宿主配置键
- **THEN** 宿主先按当前 release 的 `hostServices` 授权快照校验该 key
- **AND** 未授权 key 的读取必须被拒绝

#### Scenario:动态插件授权后使用统一读取优先级

- **WHEN** 动态插件通过 `hostconfig.get` 读取已授权的宿主配置键
- **AND** 该 key 同时存在当前上下文可见的`sys_config`值和静态`config.yaml`值
- **THEN** 系统返回`sys_config`中的有效值
- **AND** 授权通过后的读取优先级与源码插件 HostConfig 保持一致

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

