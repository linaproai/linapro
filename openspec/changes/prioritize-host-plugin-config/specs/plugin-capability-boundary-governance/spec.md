## MODIFIED Requirements

### Requirement: 配置公开面只能包含插件自身配置和宿主配置

系统 SHALL 将插件公开配置能力限定为两类：`Services.Plugins().Config()`读取当前插件自身配置，`Services.HostConfig()`读取宿主开放配置。根`Services.Config()` MUST NOT 作为普通插件公开入口存在。`Services.Plugins().Config()`可以读取宿主主静态配置文件中与当前插件 ID 完整匹配的`plugin.<plugin-id>`配置段，并将其视为当前插件自身配置的最高优先级来源；除此之外，`Services.Plugins().Config()`MUST NOT 读取任意宿主配置树、宿主运行时配置中心数据或其他插件的配置段。

#### Scenario: 插件读取宿主配置

- **WHEN** 插件需要读取宿主开放配置项
- **THEN** 插件通过`Services.HostConfig()`访问
- **AND** `HostConfig()`不得读取当前插件私有`config.yaml`

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

#### Scenario: 插件不能通过自身配置服务读取其他插件配置

- **WHEN** 插件`plugin-a`通过`Services.Plugins().Config()`读取配置 key
- **THEN** 系统只在`plugin.plugin-a`、`plugins/plugin-a/config.yaml`、`apps/lina-plugins/plugin-a/manifest/config/config.yaml`或`plugin-a`当前 artifact 默认配置中解析
- **AND** 系统不得返回`plugin.plugin-b`、`plugins/plugin-b/config.yaml`或其他插件配置来源中的值

#### Scenario: 动态插件宿主配置授权不被插件自身配置替代

- **WHEN** 动态插件需要读取宿主任意配置 key
- **THEN** 动态插件必须声明并获得`hostconfig.get`对应 key 授权
- **AND** `plugins.config.get`只能读取当前插件作用域配置，不得绕过`hostconfig`授权读取宿主任意配置
