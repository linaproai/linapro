## Context

`apps/lina-core/pkg/pluginservice/config` 是宿主发布给源码插件使用的公共配置服务入口。当前该组件通过 `MonitorConfig` 类型别名和 `GetMonitor()` 方法暴露了 `monitor-server` 插件需要的配置形态，导致宿主公共组件开始持有插件业务配置 schema。

这与插件体系的边界目标冲突：源码插件可以复用宿主能力，但插件业务配置结构、默认值和校验逻辑应归属于插件自身。宿主公共组件应只提供稳定的读取能力，不应随插件数量增长继续增加 `GetXxx()` 专用方法。

本次变更不改变配置文件的整体来源，仍以 GoFrame 配置系统为权威数据源；也不新增运行时配置写入能力。

## Goals / Non-Goals

**Goals:**

- 将 `pluginservice/config` 设计为业务无关的通用只读配置访问器。
- 允许源码插件读取任意配置 key，而不是限制在特定插件前缀下。
- 允许动态插件通过声明 `config` host service 读取完整静态配置内容。
- 支持结构体扫描、原始值读取、存在性判断、基础类型读取和 `time.Duration` 解析。
- 将插件专用配置结构、默认值、校验和业务语义移动到插件内部。
- 移除当前公共组件中的 `MonitorConfig` 别名和 `GetMonitor()` 专用方法。
- 为配置读取错误、缺失 key 和 duration 解析补齐单元测试。

**Non-Goals:**

- 不提供配置写入、保存、热更新或运行时配置管理能力。
- 不调整系统配置管理模块的数据库配置能力。
- 不引入新的配置文件格式、配置中心或外部依赖。
- 不迁移现有配置 key 的业务含义；例如 `monitor.interval` 可继续由 `monitor-server` 插件通过通用读取器读取。

## Decisions

### Decision 1: 公共接口采用通用 key 访问，而不是业务方法

`pluginservice/config.Service` 暴露通用方法：

- `Get(ctx, key)`：读取原始 GoFrame 配置值。
- `Exists(ctx, key)`：判断 key 是否存在。
- `Scan(ctx, key, target)`：将配置段扫描到调用方提供的结构体。
- `String/Bool/Int/Duration(ctx, key, defaultValue)`：读取基础类型并支持缺省值。

插件内部维护自己的 `Config` 结构和 `Load(ctx)` 方法。以 `monitor-server` 为例，它可以在插件内部扫描 `monitor` 配置段，然后再读取 `monitor.interval` 解析为 `time.Duration` 并执行秒级对齐校验。

替代方案：继续为每个插件在公共组件中增加 `GetXxx()`。该方案会让宿主公共组件依赖插件业务 schema，随着插件增长持续膨胀，因此不采用。

### Decision 2: 允许读取任意 key，但保持只读和可信源码插件边界

源码插件与宿主在同一进程、同一代码仓库中构建，属于可信扩展。配置服务不对 key 增加前缀限制，插件可读取完整配置文件内容。

替代方案：强制限制到 `plugins.<plugin-id>` 前缀。该方案可以形成更强隔离，但不符合当前需求，也会阻碍源码插件复用宿主已有通用配置。安全治理通过只读边界、代码审查和未来动态插件 host service 授权来解决。

### Decision 3: Duration 解析由公共服务提供，业务校验由插件负责

公共服务负责把配置字符串解析为 `time.Duration`，并保证默认值语义稳定。具体业务约束，例如“必须大于 0”“必须至少 1 秒”“必须按整秒对齐”，由插件或调用方在自己的配置加载方法中校验。

替代方案：公共服务内置所有 duration 校验策略。该方案会把业务规则混入通用组件，无法覆盖不同配置项的差异，因此不采用。

### Decision 4: 错误返回优先，插件调用点自行决定失败策略

通用读取方法返回 `error`，不在公共服务中直接 `panic`。插件启动期或定时任务注册期可以选择 fail-fast；普通业务路径可将错误包装为调用端可见的业务错误或内部错误。

替代方案：沿用宿主内部配置服务的 `mustScanConfig` 风格直接 panic。该方案适合宿主启动期静态配置加载，但作为插件通用公共接口过于强硬，不利于插件自行控制降级策略，因此不采用。

### Decision 5: 动态插件通过 config host service 读取完整静态配置

动态插件无法直接导入 `pkg/pluginservice/config`，因此通过 `lina_env.host_call` 增加 `config` host service。动态插件在 `plugin.yaml` 中声明：

```yaml
hostServices:
  - service: config
```

该声明授予动态插件读取完整静态配置的能力，不再要求配置 key 资源白名单。`methods` 可以省略；省略时等价于授予当前完整的配置只读方法集合，即 `get`、`exists`、`string`、`bool`、`int`、`duration`。如果插件希望收窄权限，也可以显式声明其中一部分方法：

```yaml
hostServices:
  - service: config
    methods: [get, exists, string, bool, int, duration]
```

请求 payload 携带 key；`get` 返回配置值的 JSON 表示，并支持 key 为空或 `.` 时返回 GoFrame 配置系统暴露的完整配置快照。`exists` 返回 found 标记；`string`、`bool`、`int`、`duration` 返回对应类型的字符串表示。wasip1 guest SDK 的 `Exists`、`String`、`Bool`、`Int`、`Duration` 直接调用对应 host service 方法，手工构造 host_call 的动态插件也可以调用这些只读方法。服务不提供写入、保存、热更新或运行时配置管理。

替代方案：要求动态插件逐项声明可读 key 或 key pattern。该方案隔离性更强，但不符合当前“动态插件也可以读全量配置”的目标，因此不采用。

## Risks / Trade-offs

- 全量配置可读可能暴露敏感配置给源码插件和动态插件 → 该能力必须通过 `hostServices` 显式声明，纳入动态插件安装/启用阶段的授权快照和 host_call 审计链路；调用仍保持只读。
- 移除 `GetMonitor()` 是破坏性接口变更 → 当前项目无历史兼容约束；同步迁移 `monitor-server` 插件调用点和测试即可。
- 业务校验下沉到插件后可能出现校验风格不一致 → 在 `monitor-server` 迁移时建立插件内部配置加载模式，后续插件复用该模式。
- 通用配置服务不做运行时缓存失效 → 本次只读取静态配置文件，不新增可变缓存；若未来支持运行时动态配置，需要单独接入集群模式下的跨实例修订号或广播失效机制。

## Migration Plan

1. 在 `pkg/pluginservice/config` 中替换公开接口，提供通用只读配置访问方法。
2. 移除 `MonitorConfig` 类型别名和 `GetMonitor()` 方法。
3. 在 `monitor-server` 插件内部新增私有配置加载逻辑，读取现有 `monitor` 配置段并完成默认值、duration 解析和业务校验。
4. 更新 `monitor-server` 的定时任务注册和清理逻辑，改用插件内部配置加载方法。
5. 增加或更新单元测试，覆盖通用配置服务和 `monitor-server` 配置加载。
6. 为动态插件新增 `config` host service 常量、能力推导、编解码、guest helper 和 host dispatcher。
7. 运行受影响 Go 测试，必要时运行完整后端测试。

## Open Questions

- 无。
