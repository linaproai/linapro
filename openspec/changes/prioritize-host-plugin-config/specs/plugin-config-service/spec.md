## MODIFIED Requirements

### Requirement: 插件运行时配置文件必须遵循插件作用域目录约定

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
