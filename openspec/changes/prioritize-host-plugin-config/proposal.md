## Why

当前插件业务配置在生产环境主要依赖`plugins/<plugin-id>/config.yaml`，对只需要一两个业务参数的插件来说管理成本偏高。将主框架静态配置文件中的`plugin.<plugin-id>`作为插件业务配置的最高优先级来源，可以让简单插件用一份主配置完成部署，同时保留独立配置文件承载复杂配置的能力。

## What Changes

- 调整插件作用域配置读取优先级，优先读取主框架静态配置文件中的`plugin.<plugin-id>`配置段。
- 当`plugin.<plugin-id>`配置段不存在时，继续按现有顺序读取`plugins/<plugin-id>/config.yaml`、开发期`manifest/config/config.yaml`和动态 artifact 默认配置。
- 明确`plugin.<plugin-id>`属于`Plugins().Config()`的插件作用域配置源，不等同于允许插件通过自身配置服务读取任意宿主配置。
- 源码插件和动态插件共享同一配置工厂实例，确保`Services.Plugins().Config()`和动态`plugins.config.get`语义一致。
- 更新插件公开契约文档，说明插件业务配置的来源优先级和`HostConfig()`边界。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-config-service`：插件作用域配置读取新增主框架静态`plugin.<plugin-id>`最高优先级来源，并保留现有文件和 artifact 回退来源。
- `plugin-capability-boundary-governance`：配置公开面边界调整为允许`Plugins().Config()`读取`plugin.<plugin-id>`这一插件作用域配置段，但仍禁止读取任意宿主配置树或运行时配置中心。

## Impact

- 影响`apps/lina-core/pkg/plugin/capability/plugincap`的配置源解析和测试。
- 影响`apps/lina-core/internal/cmd/internal/httpstartup`、`apps/lina-core/internal/service/plugin/plugin_host_services.go`和`apps/lina-core/internal/service/plugin/internal/capabilityhost`的启动期共享配置工厂装配。
- 影响动态插件`plugins.config.get`的有效配置来源，但不改变 host service 方法名、协议载荷或授权模型。
- 需要同步`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`中插件配置能力说明。
- 不涉及数据库结构、HTTP API、前端 UI、数据权限或插件目录内业务文件变更。
