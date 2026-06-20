## Context

LinaPro 已经将插件业务配置收口到`Services.Plugins().Config()`和动态插件`plugins.config.get`。现有配置解析顺序为生产配置文件`plugins/<plugin-id>/config.yaml`、开发期`apps/lina-plugins/<plugin-id>/manifest/config/config.yaml`、动态 artifact 内`manifest/config/config.yaml`。这保证了插件配置不回退到宿主全局配置树，但也导致只需要少量参数的生产插件必须额外维护独立配置文件。

主框架静态配置服务已经通过`GetRaw(ctx, key)`提供可信内部原始读取能力，`HostConfig()`也基于该能力暴露宿主配置读取。新的需求不是让插件绕过`Plugins().Config()`直接读取宿主配置，而是把`plugin.<plugin-id>`定义为插件作用域配置的一个受限静态来源。

本变更属于宿主通用插件能力，不属于具体工作台页面适配，也不修改业务插件目录内文件。它涉及源码插件能力目录、动态插件 WASM host service 和启动期依赖注入，因此必须保持两类插件语义一致。

## Goals / Non-Goals

**Goals:**

- 让`Services.Plugins().Config()`和动态`plugins.config.get`优先读取主框架静态配置中的`plugin.<plugin-id>`配置段。
- 在`plugin.<plugin-id>`不存在时保持现有文件和 artifact 默认配置回退能力。
- 通过启动期显式注入共享配置工厂，保证源码插件和动态插件读取同一套配置源优先级。
- 保持`Plugins().Config()`只读、插件作用域、非热更新的能力边界。
- 明确`HostConfig()`仍是读取宿主任意配置的独立能力，动态插件`hostconfig.get`仍受 manifest 授权约束。

**Non-Goals:**

- 不提供插件配置写入、保存、热重载或运行时变更能力。
- 不把`plugin.<plugin-id>`纳入数据库配置中心或`sys_config`管理。
- 不为每个插件在`lina-core`添加专用配置结构、`GetXxx()`方法或业务默认值。
- 不改变动态插件 host service 协议、方法名或授权声明。
- 不修改具体业务插件的配置文件、`plugin.yaml`或业务读取逻辑。

## Decisions

### 决策 1：将`plugin.<plugin-id>`建模为插件作用域配置源

`plugin.<plugin-id>`只通过`Plugins().Config()`生效，插件读取时仍传入插件内部 key，例如`storage.endpoint`。配置服务内部将当前插件 ID 映射到宿主静态配置 key`plugin.<plugin-id>`并在该配置段存在时从该段解析子 key。

备选方案是让插件使用`HostConfig()`读取`plugin.<plugin-id>.storage.endpoint`。该方案会把业务配置读取暴露为宿主配置授权问题，动态插件还需要额外声明`hostconfig`资源 key，破坏`Plugins().Config()`作为插件自身配置入口的语义，因此不采用。

### 决策 2：采用配置段级优先，而不是逐 key 混合回退

只要`plugin.<plugin-id>`配置段存在，系统就使用该配置段作为该插件的有效配置源；该段内缺失的 key 返回缺失或调用方默认值，不再继续读取`plugins/<plugin-id>/config.yaml`补齐单个 key。

备选方案是逐 key 优先读取主配置，再按 key 回退文件配置。该方案更灵活，但会让同一个插件的实际配置来源混合，不利于生产排障和变更审计，也容易掩盖配置遗漏，因此不采用。

### 决策 3：通过窄接口注入宿主静态配置读取能力

`plugincap`位于`apps/lina-core/pkg/plugin/capability`，不能直接依赖`internal/service/config`。实现应在`plugincap`中定义或复用窄接口，仅需要读取`GetRaw(ctx, key)`这一能力。HTTP 启动装配从`configSvc`取得该接口，创建带宿主静态配置 reader 的`ConfigServiceFactory`。

备选方案是在`plugincap`内部直接调用`g.Cfg()`。这会绕过启动期共享配置服务，也无法复用受保护运行期参数的现有读取边界，更不利于测试和依赖治理，因此不采用。

### 决策 4：源码插件和动态插件复用同一个配置工厂

当前启动流程中动态插件 runtime 使用启动期创建的`pluginConfigFactory`，源码插件能力目录内部又创建了一个默认工厂。实现应扩展`NewHostServices()`和`capabilityhost.New()`参数，让源码插件能力目录接收启动期同一个`pluginConfigFactory`。

这会增加一个显式构造函数依赖，但能满足缓存敏感服务和运行时配置快照必须复用启动期共享实例的要求，也能让编译期暴露所有未同步调用点。

### 决策 5：不新增缓存或热更新机制

该能力继续读取静态配置源和插件文件源，不新增写路径或热重载。生产配置变化默认通过部署和重启生效。动态插件 artifact 默认配置仍绑定当前 active release 的执行上下文。

缓存一致性判断：本变更不新增跨实例缓存、失效事件或运行时可变数据。权威数据源是进程启动时 GoFrame 可见的主配置文件、外部插件配置文件、开发期插件配置文件或动态 artifact。集群模式下各实例通过相同部署配置和重启流程获得一致配置；未来如支持在线 reload，必须单独设计共享修订号、广播失效或等价机制。

## Risks / Trade-offs

- `plugin.<plugin-id>`与现有宿主`plugin`治理配置共用顶层命名空间，可能造成阅读混淆。缓解方式是在文档中明确`plugin.allowForceUninstall`、`plugin.dynamic`、`plugin.autoEnable`属于宿主治理配置，`plugin.<plugin-id>`属于对应插件业务配置。
- 配置段级优先会导致主配置中一旦声明`plugin.<plugin-id>`，文件配置不再补齐缺失 key。缓解方式是通过单元测试和文档明确该行为，插件继续使用默认值处理缺失 key。
- 源码插件 host services 构造函数增加显式依赖，会影响多个测试 fixture。缓解方式是测试统一传入`plugincap.NewConfigFactory`，并补充共享工厂行为测试。
- 若实现错误地允许`Plugins().Config()`读取任意宿主 key，会破坏插件配置边界。缓解方式是只读取完整`plugin.<plugin-id>`段，并保留空 key、`.`、前后点号等 key 校验。

## Migration Plan

1. 先扩展配置工厂和测试，验证静态`plugin.<plugin-id>`优先级及文件回退语义。
2. 调整启动装配，让源码插件和动态插件复用同一个配置工厂。
3. 更新`pkg/plugin`中英文 README，记录配置来源优先级。
4. 运行`openspec validate prioritize-host-plugin-config --strict`、相关 Go 单元测试和启动装配编译烟测。

回滚方式：恢复配置工厂为仅文件和 artifact 来源，并将`NewHostServices()`参数恢复为内部创建默认工厂。由于不涉及数据库和外部 API，回滚不需要数据迁移。

## Open Questions

- 是否需要在`config.template.yaml`中添加注释示例展示`plugin.<plugin-id>`？当前倾向暂不添加具体插件示例，避免主配置模板膨胀；README 足以说明用法。
