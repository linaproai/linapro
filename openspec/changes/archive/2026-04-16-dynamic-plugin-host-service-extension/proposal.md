## Why

当前动态插件的宿主回调能力仍停留在最小可用阶段，主要只有`host:log`、`host:state`和原始`host:db:*`几类能力。复杂业务插件一旦需要文件处理、出站网络访问、受治理的数据访问等能力，就会立刻撞上两类问题：一类是能力不够，另一类是如果继续沿着“每加一项能力就加一个专用 opcode”的方式演进，运行时协议、安全治理和开发体验都会快速失控。

`Envoy`与`Higress`已经证明了一条更稳的路线：先保持稳定的底层`Wasm`宿主调用边界，再由宿主发布可治理的扩展能力，并在 guest 侧提供更高层的 SDK 封装。`Lina`需要把当前零散的 Host Functions 演进成一套面向后台业务的宿主服务模型，才能在不暴露宿主内部实现的前提下，支持更复杂的动态插件场景。

## What Changes

- 在现有`lina_env.host_call`基础上，引入“稳定 ABI ＋ 结构化宿主服务”的扩展模型，后续敏感能力统一走宿主服务注册表，而不是继续线性增加离散 opcode。
- 为动态插件新增结构化的`hostServices`声明，用于描述插件要访问的宿主服务、方法、资源引用或数据表范围以及治理参数；宿主内部粗粒度 capability 分类由这些声明自动推导，不再要求作者重复维护顶层`capabilities`。
- 对所有资源型 hostServices 采用“声明即申请，安装/启用时由宿主确认授权”的治理模型；其中`storage`服务直接声明逻辑路径或路径前缀`resources.paths`，`network`服务直接声明 URL 模式，`data`服务在`resources`节点下以`tables`声明数据表申请，其余低优先级服务（`cache`、`lock`、`notify`）继续沿用逻辑`resourceRef`规划；宿主在安装或启用阶段展示、确认并固化最终授权快照。
- 本迭代按优先级分两层推进宿主服务：高优先级先完成`runtime`、存储／文件、出站网络、数据访问四类能力；低优先级继续纳入缓存、锁和通知能力。
- 当前已实现的`host:log`、`host:state`、`host:db:*`只作为现状参考，不构成兼容约束；宿主可以直接重构为统一宿主服务模型。
- 将宿主服务调用的上下文透传、资源授权、限流／限额、审计记录和错误模型纳入动态插件运行时治理，并补齐 guest SDK、样例插件和自动化测试。

## Capabilities

### New Capabilities

- `plugin-host-service-extension`：定义动态插件调用宿主服务的统一协议、注册表、鉴权与审计模型。
- `plugin-storage-service`：定义动态插件的文件／对象存储访问能力、逻辑空间隔离和公开性治理。
- `plugin-network-service`：定义动态插件的出站 HTTP 调用能力、上游引用绑定和请求限制策略。
- `plugin-data-service`：定义动态插件的数据访问能力、表级授权、数据范围注入与事务边界。
- `plugin-cache-service`：定义动态插件的宿主缓存访问能力、命名缓存空间和 TTL 治理。
- `plugin-lock-service`：定义动态插件的宿主锁能力、锁资源绑定和续租／释放约束。
- `plugin-notify-service`：定义动态插件的宿主通知能力、通知通道绑定和模板治理。

### Modified Capabilities

- `plugin-runtime-loading`：动态插件运行时产物需要携带宿主服务声明、资源授权和协议版本信息，宿主装载时需恢复这些治理快照。
- `plugin-manifest-lifecycle`：`plugin.yaml`需要支持结构化宿主服务声明和资源申请快照（含`resourceRef`与`data.resources.tables`），并以`hostServices`作为唯一作者侧宿主能力声明入口，以便安装、升级、卸载和审计时统一治理。
- `user-auth`：JWT Token 有效期配置并入当前迭代，统一改为使用 `jwt.expire` duration 字符串配置。
- `online-user`：在线会话超时阈值与清理周期并入当前迭代，统一改为使用 `session.timeout`、`session.cleanupInterval` duration 字符串配置。
- `server-monitor`：服务监控采集周期并入当前迭代，统一改为使用 `monitor.interval` duration 字符串配置。

## Impact

- 后端运行时：`apps/lina-core/pkg/pluginbridge`、`apps/lina-core/internal/service/plugin/internal/wasm`、`apps/lina-core/internal/service/plugin/internal/runtime`
- 清单与治理：`apps/lina-core/internal/service/plugin/internal/catalog`、`apps/lina-core/internal/service/plugin/internal/integration`
- 构建链路：`hack/build-wasm`及相关动态插件打包工具
- 插件样例与文档：动态插件样例、插件开发文档、宿主服务使用说明
- 测试：插件运行时单测、集成测试，以及`hack/tests/e2e/plugin/`下的插件治理回归用例
- 配置与通用规范：`apps/lina-core/manifest/config/config.yaml`、`apps/lina-core/manifest/config/config.template.yaml`、`apps/lina-core/internal/service/config/`、`apps/lina-core/internal/service/auth/`、`apps/lina-core/internal/service/role/`、`apps/lina-core/internal/service/cron/`以及相关项目规范文档

## Merged Scope: `config-duration-unification`

为满足项目“单一活跃迭代统一管理”的约束，原 `config-duration-unification` 变更并入当前 `dynamic-plugin-host-service-extension` 迭代，不再保留独立活跃目录。并入范围如下：

- 将 `jwt`、`session`、`monitor` 下的时长配置统一为带单位的 duration 字符串。
- 将整数加单位后缀的配置键调整为统一语义命名：`jwt.expire`、`session.timeout`、`session.cleanupInterval`、`monitor.interval`。
- 配置服务统一将这些配置解析为 `time.Duration`，业务层直接消费解析结果，不再自行做小时、分钟、秒换算。
- 旧整数配置键不保留兼容逻辑，配置文件、实现代码和规范文档仅保留新的 duration 写法。
- 同步保留该子范围内已完成的测试、规范和项目约束更新，包括“后端时长使用 `time.Duration`”与“禁止忽略 `error` 返回值”两项规范。
