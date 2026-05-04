## Why

当前 `apps/lina-core/pkg/pluginservice/config` 已开始通过 `GetMonitor()` 暴露插件业务相关的强类型配置，这会导致每新增一个插件或插件配置形态，就需要继续修改宿主公共组件。插件需要读取配置文件中的不同内容，但配置服务本身应保持业务无关，只提供稳定、通用、只读的配置访问能力。

## What Changes

- 新增面向源码插件的通用只读配置服务能力，允许插件按任意配置 key 读取、扫描和解析配置文件内容。
- 移除 `pluginservice/config` 对具体插件业务配置结构的直接暴露，例如 `MonitorConfig` 与 `GetMonitor()` 这类插件专用方法。
- 插件自己的配置结构、默认值、校验和业务含义由插件内部维护，宿主公共组件仅负责通用读取与基础类型解析。
- 配置读取保持只读，不在插件配置服务中提供写入、保存或运行时变更能力。
- 配置服务需要清晰说明可信边界：源码插件可读取完整宿主配置；动态或第三方插件若后续复用该能力，必须再通过 host service 授权和审计机制治理。
- **BREAKING**：源码插件不再通过 `configsvc.New().GetMonitor(ctx)` 获取监控配置，需改为通过通用配置访问器读取对应 key 并在插件内部完成结构化解析。

## Capabilities

### New Capabilities

- `plugin-config-service`: 定义插件可使用的通用只读配置访问能力，包括按任意 key 获取配置、扫描配置段、解析基础类型和解析 `time.Duration`。

### Modified Capabilities

- 无。

## Impact

- 影响 `apps/lina-core/pkg/pluginservice/config` 的公开接口设计与实现。
- 影响当前依赖 `pluginservice/config.GetMonitor()` 的源码插件，主要是 `apps/lina-plugins/monitor-server`。
- 需要补充插件配置服务的单元测试，覆盖任意 key 读取、结构体扫描、默认值、缺失 key、duration 解析和错误处理。
- 不涉及数据库变更，不需要新增 SQL。
- 不涉及前端 UI 文案、菜单、路由或页面交互变更；本次不需要新增或修改运行时 i18n、manifest i18n、apidoc i18n 资源。
- 不新增运行时可变配置缓存；静态配置读取可以继续复用宿主现有进程内静态配置读取方式，不引入分布式缓存一致性问题。
