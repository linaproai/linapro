## Context

当前动态插件 host service 已经形成了统一 envelope、`plugin.yaml hostServices`授权声明和`capability.Services`领域契约，但新增一个领域能力仍要在多处手写同步：

- `pkg/plugin/capability/<x>cap`定义领域契约，并在`capability.go`目录中暴露。
- `pkg/plugin/pluginbridge/pluginbridge_directory.go`把 guest 目录 getter 接到 typed client。
- `pkg/plugin/pluginbridge/internal/domainhostcall/domainhostcall_<x>.go`维护 guest 调用代码。
- `pkg/plugin/pluginbridge/protocol/protocol_hostservice_*_codec.go`维护请求响应 DTO 与编解码。
- `pkg/plugin/pluginbridge/internal/hostservice/hostservice_descriptor.go`维护 service/method/capability/resource 元数据。
- `internal/service/plugin/internal/wasm/wasm_host_service.go`维护 service 级分发 switch，并在`wasm_host_service_<x>.go`维护领域 dispatch。
- `pkg/plugin`双语`README`表格由人工维护，且用户已明确不需要恢复`generated:host-services`标记。

现有 descriptor 与测试能发现一部分漂移，但 descriptor 位于`pluginbridge/internal/hostservice`，宿主 WASM 内部包无法合法导入它作为统一元数据来源。继续把治理源放在 internal 下，会迫使宿主 dispatch、协议 codec 和 guest client 继续各自手写镜像。

本变更属于`apps/lina-core`核心宿主和动态插件桥接能力的内部架构调整，不修改工作台展示，不修改 HTTP API，不修改 SQL，不修改插件实例目录资源。

## Goals / Non-Goals

**Goals:**

- 让新增普通领域 host service 的核心修改点收敛到领域契约、公开 catalog、guest typed client、host dispatch handler 和必要目录注册。
- 让 service/method/capability/resource/payload/guest/dispatch 元数据从公开`protocol/hostservices`catalog派生，`internal/hostservice`不再拥有独立 descriptor 事实源。
- 让宿主 WASM host service 入口由 registry 查找 handler，消除`wasm_host_service.go`的 service 级大 switch。
- 让普通领域能力默认使用统一 JSON envelope，减少每个领域都新增专用`protowire`codec 的需求。
- 保留性能或资源语义明确的特殊 host service 专用 codec，包括`storage`、`cache`、`lock`、`data/recordstore`、`network`以及已有需要二进制稳定 wire 的能力。
- 通过 Go 测试和静态扫描覆盖 catalog、guest client、codec、dispatch registry 和 import 边界，保证遗漏同步点会失败。

**Non-Goals:**

- 不把`capability/<x>cap`领域契约改为生成代码；领域接口仍由能力 owner 手写。
- 不恢复或新增`README`中的`generated:host-services`标记，也不把 README 表格作为本变更首期生成目标。
- 不改变动态插件`plugin.yaml hostServices`声明、service/method wire 字符串、授权快照、错误 envelope、审计语义或数据权限边界。
- 不引入完整代码生成器；首期用结构化 catalog、显式注册和治理测试解决主要扇出。
- 不新增外部依赖，不新增数据库结构，不修改前端页面或运行时多语言资源。

## Decisions

### D1：公开`protocol/hostservices`作为描述源

新增`apps/lina-core/pkg/plugin/pluginbridge/protocol/hostservices`包，承载 host service catalog。catalog 定义每个 service family、method、capability、资源类型、payload 形态、guest client 发布状态和 host dispatcher 发布状态。

目标结构：

```text
apps/lina-core/pkg/plugin/pluginbridge/
  protocol/
    protocol_hostservice.go
    protocol_hostservice_json.go
    protocol_hostservice_binary_codec.go
    hostservices/
      catalog.go
      users.go
      dict.go
      files.go
      sessions.go
      storage.go
      cache.go
      <x>.go
  internal/
    hostservice/
      hostservice_descriptor.go
```

`internal/hostservice`继续负责 manifest validation、capability derivation 和测试辅助，但 descriptor 数据必须从`protocol/hostservices`转换或直接引用，不再维护第二份手写表。

**Rationale：**宿主`internal/service/plugin/internal/wasm`不能导入`pluginbridge/internal/hostservice`，但可以导入公开`protocol/hostservices`。把元数据前移到 protocol 子组件可以同时服务协议、guest 和宿主 dispatch 治理。

**Alternatives considered：**

- 继续以`internal/hostservice`descriptor为源：会卡在 Go internal import 边界，宿主 dispatch 仍无法合法依赖。
- 直接生成所有代码：能减少更多手写，但会引入生成入口、跨平台工具和 generated 文件治理；当前主要痛点可先用 catalog 与 registry 收敛。

### D2：普通领域 host service 使用统一 JSON envelope

普通领域服务不再默认新增`protocol_hostservice_<x>_codec.go`专用`protowire`codec，而是通过`HostServiceJSONRequest`和`HostServiceJSONResponse`承载紧凑 JSON。guest typed client 和 host dispatch handler 负责把领域 DTO marshal/unmarshal 到该 envelope。

特殊服务保留专用 codec：

| 服务类型 | 保留原因 |
| --- | --- |
| `storage` | 包含二进制内容、路径资源授权和流量敏感 payload。 |
| `cache` | TTL、原子递增和资源 namespace 语义需要稳定紧凑 wire。 |
| `lock` | 锁 ticket、续租和释放语义需要稳定专用 payload。 |
| `data`/`recordstore` | 查询计划、事务和数据权限治理需要结构化专用协议。 |
| `network` | URL、header、body 和资源模式校验需要专用协议。 |
| `ai`等多模态大 payload | 仅在已有规范要求或 payload 规模明确时保留专用协议。 |

**Rationale：**多数领域能力只传递小型 DTO 或投影，专用`protowire`codec 增加的手写同步点大于收益。统一 JSON envelope 让新增领域能力不必同步新增 codec 文件，同时仍保留 typed client 和 typed domain contract。

**Alternatives considered：**

- 所有服务统一 JSON：会降低 storage/data/cache/lock/network 这类资源或性能敏感服务的协议清晰度。
- 所有服务继续专用`protowire`：能保持 wire 细粒度，但正是当前扇出的主要来源。

### D3：宿主 dispatch 改为显式 registry

新增`apps/lina-core/internal/service/plugin/internal/wasm/hostservicedispatch`子组件，提供 registry、handler context、response helper 和重复注册检测。`wasm_host_service.go`只保留 envelope 解码、授权校验、上下文构造、registry lookup 和统一错误响应。

首期目标结构：

```text
apps/lina-core/internal/service/plugin/internal/wasm/
  wasm_host_service.go
  wasm_host_service_registry.go
  wasm_host_service_context.go
  hostservicedispatch/
    registry.go
    context.go
    response.go
    register.go
  wasm_host_service_users.go
  wasm_host_service_dict.go
  wasm_host_service_files.go
  wasm_host_service_sessions.go
  wasm_host_service_storage.go
  wasm_host_service_cache.go
  wasm_host_service_<x>.go
```

注册必须通过`wasm_host_service_registry.go`显式调用各领域注册函数完成，不使用`init()`隐式注册。现有领域 handler 首期保留在`wasm`父包作为适配层，因为它们依赖父包私有`hostCallContext`、运行时快照、manifest artifact、job collector 和授权匹配 helper；强行移入子包会把这些内部执行状态扩大为跨包契约。后续若需要物理迁移到`hostservicedispatch/<domain>`子包，必须先抽取窄上下文契约，并保持共享运行期依赖仍由启动期或父包显式提供。

**Rationale：**registry 在这里不是为了预留未知扩展，而是为了消除当前二十多个 service 的手写 switch 和 dispatcher 覆盖漂移。显式注册保留 DI 可见性，能在编译和测试阶段暴露遗漏。

**Alternatives considered：**

- 继续在`wasm_host_service.go`维护 switch：简单但每新增 service 必然改入口文件，且 descriptor 覆盖测试只能追着 switch 形态变化。
- 使用`init()`自注册：减少装配代码，但隐藏依赖来源，不符合显式依赖注入和启动期共享实例规则。
- 立即把所有 handler 搬进`hostservicedispatch/<domain>`子包：目录更整齐，但会迫使`hostCallContext`和运行期快照扩大公开面；首期收益不抵风险，因此先以 registry 子组件和父包显式适配层收敛入口扇出。

### D4：guest typed client 继续保留强类型边界

动态插件 guest 侧继续通过 typed client 调用宿主能力，避免把 raw service/method 字符串泄露给业务插件。为了降低单包文件拥挤，普通领域 client 可以从当前平铺文件迁移到领域子包，再由目录层集中装配。

目标结构：

```text
apps/lina-core/pkg/plugin/pluginbridge/
  pluginbridge_directory.go
  internal/domainhostcall/
    domainhostcall.go
    users/users.go
    dict/dict.go
    files/files.go
    sessions/sessions.go
    <x>/<x>.go
```

`pluginbridge_directory.go`或能力 guest 目录只负责把统一 invoker 注入 typed client，不承载领域调用逻辑。`pluginbridge`根包不重新拥有业务能力语义。

**Rationale：**新增领域能力仍需要 typed client，这是插件作者体验和调用安全的一部分；要消除的是 codec、descriptor 和 dispatch 入口的镜像，而不是取消类型化 SDK。

**Alternatives considered：**

- 暴露 raw host service invoker 给业务插件：扇出更小，但会把 service/method、payload 和授权失败细节推给插件作者，削弱框架能力。
- 为每个领域生成 guest client：可作为后续方向，但首期不增加代码生成治理。

### D5：能力目录保持领域 owner

`pkg/plugin/capability/<x>cap`和`pkg/plugin/capability/capability.go`仍是领域能力契约与源码插件消费面的 owner。动态插件 bridge 只是 transport 适配层，不能把`pluginbridge`变成与`capability`平行的业务能力目录。

目标结构：

```text
apps/lina-core/pkg/plugin/capability/
  capability.go
  usercap/usercap.go
  dictcap/dictcap.go
  filecap/filecap.go
  <x>cap/<x>cap.go
```

**Rationale：**当前问题是动态插件 bridge 扇出，不是领域契约 owner 错位。保留领域契约手写能保持业务语义清晰，也符合源码插件和动态插件共享能力目录的既有规范。

### D6：README 表格不作为生成目标

本变更只要求双语`README`在能力变更后人工同步审查，不恢复生成标记，也不新增 README 表格生成器。catalog 可以为未来文档工具提供数据源，但本次任务不实现。

**Rationale：**用户已明确删除`generated:host-services`标记且不需要恢复。本变更聚焦核心扇出：协议 catalog、guest client 和宿主 dispatch。

## Risks / Trade-offs

- [Risk] 统一 JSON envelope 可能让普通领域 payload 的字段演进缺少`protowire`字段编号约束。→ Mitigation：普通领域 DTO 只承载插件可见投影，采用结构化 JSON schema 测试和 typed client round trip；特殊服务继续使用专用 codec。
- [Risk] registry 引入新的间接层，调试时需要从 service/method 定位 handler。→ Mitigation：registry key 使用明确`service/method`，注册测试输出缺失或重复 key，handler 包名按领域命名。
- [Risk] catalog 与现有 descriptor 迁移期间短暂并存，可能形成双事实源。→ Mitigation：第一步先让 descriptor 从 catalog 派生，并添加测试阻断 descriptor 内新增手写表。
- [Risk] dispatch 子包拆分可能增加 DI 传递量。→ Mitigation：只允许 registry 构造入口显式接收已有 WASM host service 共享依赖；禁止聚合接口依赖结构体，禁止 handler 自行`New()`关键服务。
- [Risk] 普通领域和特殊服务的 codec 分类不清会产生新例外。→ Mitigation：catalog 中必须声明 payload kind；新增专用 codec 必须在设计或任务记录中说明性能、资源或 wire 稳定性依据。

## Migration Plan

1. 新增`protocol/hostservices`catalog 类型与当前 descriptor 等价的数据表。
2. 让`pluginbridge/internal/hostservice`从 catalog 派生 descriptor、capability 和 manifest validation 数据，保留现有公开行为。
3. 新增统一 JSON envelope codec，并把一个代表性普通领域服务（建议`users`）迁移为 catalog + JSON envelope + registry dispatch。
4. 新增`hostservicedispatch`registry，并把`wasm_host_service.go`入口切到 registry lookup；先保留旧 handler 函数作为父包显式适配层，避免扩大 WASM 私有执行上下文公开面。
5. 迁移其余普通领域服务，保留特殊服务专用 codec，但统一注册到 registry。
6. 更新治理测试：catalog 覆盖 guest client、protocol payload、registry handler、特殊 codec 白名单、无 service 级 switch、无 Go internal import 边界违规。
7. 运行`openspec validate consolidate-hostservice-domain-bridge --strict`、相关 Go 测试和静态扫描。

Rollback 不要求保留运行期兼容层；如实现阶段某一步失败，可回退该步骤的未完成文件并保持旧 dispatch 路径继续工作。

## Open Questions

- 首期是否迁移全部普通领域服务，还是先迁移`users`、`dict`、`files`、`sessions`等代表性领域后再继续批量迁移？默认任务按全部普通领域迁移规划。
- `ai`当前已有多模态专用 payload。实现时应逐方法评估是否继续作为特殊服务保留，默认不在本变更中强制改为 JSON envelope。
