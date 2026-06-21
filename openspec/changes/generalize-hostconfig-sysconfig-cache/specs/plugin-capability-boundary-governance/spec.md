## MODIFIED Requirements

### Requirement: 配置公开面只能包含插件自身配置和宿主配置

系统 SHALL 将插件公开配置能力限定为两类：`Services.Plugins().Config()`读取当前插件自身配置，`Services.HostConfig()`读取宿主开放配置。根`Services.Config()` MUST NOT 作为普通插件公开入口存在。`Services.Plugins().Config()`可以读取宿主主静态配置文件中与当前插件 ID 完整匹配的`plugin.<plugin-id>`配置段，并将其视为当前插件自身配置的最高优先级来源；除此之外，`Services.Plugins().Config()`MUST NOT 读取任意宿主配置树、宿主运行时配置中心数据或其他插件的配置段。

`Services.HostConfig()`对于源码插件 SHALL 读取当前上下文可见的宿主配置，包括`sys_config`中的有效 key 和静态宿主配置 key。源码插件使用该能力时不需要声明逐 key manifest 白名单，但 MUST 通过稳定`HostConfig()`能力访问，不得直接访问宿主`dao.SysConfig`、`entity.SysConfig`、`internal/service/sysconfig`或其他内部实现。动态插件通过`hostconfig.get`读取宿主配置时 MUST 继续声明并获得对应 key 授权。

#### Scenario: 源码插件读取自定义系统配置

- **WHEN** 源码插件通过`Services.HostConfig()`读取`custom.feature.limit`
- **AND** 当前上下文可见的`sys_config`中存在`custom.feature.limit`
- **THEN** 系统返回`sys_config`中的有效值
- **AND** 源码插件不需要为该 key 声明 manifest 授权

#### Scenario: 动态插件读取自定义系统配置必须授权

- **WHEN** 动态插件通过`hostconfig.get`读取`custom.feature.limit`
- **AND** 插件 manifest 的`hostServices`未授权该 key
- **THEN** 系统 MUST 拒绝该读取

#### Scenario: 动态插件读取已授权系统配置

- **WHEN** 动态插件通过`hostconfig.get`读取`custom.feature.limit`
- **AND** 插件 manifest 的`hostServices`已授权该 key
- **AND** 当前上下文可见的`sys_config`中存在`custom.feature.limit`
- **THEN** 系统返回`sys_config`中的有效值

#### Scenario: 插件读取自身独立配置文件

- **WHEN** 插件需要读取自身`config.yaml`或 artifact 内配置
- **AND** 宿主主静态配置中不存在当前插件的`plugin.<plugin-id>`配置段
- **THEN** 插件通过`Services.Plugins().Config()`访问
- **AND** `Plugins().Config()`不得读取任意宿主配置树或运行时配置中心数据

#### Scenario: 插件读取主框架静态配置中的自身配置段

- **WHEN** 插件`plugin-a`需要读取自身业务配置
- **AND** 宿主主静态配置中存在`plugin.plugin-a`
- **THEN** 插件通过`Services.Plugins().Config()`访问该配置段内的子 key
- **AND** 插件调用方不需要通过`Services.HostConfig()`读取`plugin.plugin-a.*`

