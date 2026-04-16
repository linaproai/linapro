## Context

`Lina`当前已经有一套最小的动态插件宿主回调链路：guest 通过`lina_env.host_call(opcode, reqPtr, reqLen)`进入宿主，宿主按 opcode 校验`capabilities`并分发到`log`、`state`、`db`这几类处理器。这一版已经证明“`Wasm`插件可以通过宿主提供的接口安全扩展能力”是可行的，但它也暴露了当前模型的三个问题：

- 能力粒度过低：当前只有`host:log`、`host:state`和原始`host:db:*`，缺少复杂业务插件真正需要的文件、网络和更安全的数据访问能力。
- 扩展方式不可持续：当前模型本质上是“每个能力一个 opcode ＋ 一套单独编解码”，继续沿用会把`pluginbridge`和运行时分发逻辑推成难维护的 syscall 列表。
- 治理边界不完整：现有`capabilities`只能表达“这个插件大概能做什么”，却不能表达“它具体能访问哪个上游、哪个存储空间、哪个数据资源”，这会让安全治理和审计都停留在粗粒度层面。

本次设计同时参考了两个成熟模式：

- `Envoy`的`Proxy-Wasm`路线证明了“稳定底层 ABI ＋ 宿主扩展函数／异步调用原语”是可持续的扩展模型，典型能力包括`dispatch_http_call`、`grpcCall`、`shared_data`、`shared_queue`和`foreign function`。
- `Higress`在此基础上进一步提供`HttpClient`、`RedisClient`等高层 SDK 封装，说明插件开发者通常并不希望直接操作底层 host call，而更希望拿到一组面向业务的稳定能力接口。

对`Lina`来说，直接照搬服务网格的 API 面并不合适。`Lina`是后台管理系统，不需要暴露原始 socket、原始文件路径或宿主内部服务实例；它真正需要的是一套面向后台业务的宿主服务模型，让动态插件在可审计、可授权、可回滚的前提下使用宿主能力。

另外需要明确一条项目约束：`Lina`当前是全新项目，没有历史债务，也没有必须延续的旧插件生态。因此，当前仓库里已经存在的最小 Host Call 实现只是一轮探索性落地，不构成后续设计的兼容边界。本次设计可以直接选择最优模型，对现有`host:log`、`host:state`、`host:db:*`进行重构、合并或移除，不需要为“旧协议继续可用”额外保留长期分支。

## Goals / Non-Goals

**Goals:**

- 把当前动态插件能力扩展模型从“离散 Host Functions 集合”演进成“稳定 ABI ＋ 宿主服务注册表 ＋ guest SDK”的分层结构。
- 把当前已实现的最小 Host Call 直接重构到统一宿主服务模型，而不是在旧设计外层继续叠加兼容层。
- 将插件能力声明拆成“粗粒度 capability 授权”和“细粒度资源引用授权”两层治理。
- 本迭代按优先级交付十类宿主能力中的前四类核心能力和后六类低优先级能力，其中必须先完成`runtime`、`storage`、`network`、`data`四类核心能力。
- 让宿主服务调用显式复用插件当前的执行上下文，包括`pluginId`、执行来源、当前路由或 Hook、用户身份快照、数据范围和调用超时。
- 让运行时产物、清单快照、资源引用记录和审计链路都能反映插件实际申请和使用的宿主能力范围。

**Non-Goals:**

- 不向动态插件暴露宿主绝对文件路径、原始文件系统句柄或任意目录遍历能力。
- 不向动态插件暴露原始 socket、任意域名直连或宿主内部网络拓扑细节。
- 不让动态插件直接拿到宿主`ghttp`上下文、数据库连接对象或内部`service`实例。
- 不在本次迭代内引入流式 Host Call、`WebSocket`、`SSE`、双向流`gRPC`等长连接能力。
- 不为当前探索性实现额外维护长期兼容协议、双写分支或迁移兜底链路。

## Decisions

### 决策一：保持单一 ABI 入口，但新增统一的宿主服务调用通道

本次不改变当前`Wasm`导入模型，guest 仍然只通过`lina_env.host_call`进入宿主。但在该入口之上，不再延续“每个能力一个 opcode”的公开协议模型，而是统一收敛为一条结构化的“宿主服务调用”通道。当前已实现的`log`、`state`、`db`处理器可以直接并入新的宿主服务分发器，不需要保留为对插件公开的长期协议承诺。

建议的分层如下：

```text
Guest Business Code
        │
        ▼
Guest SDK Helpers
        │
        ▼
lina_env.host_call
        │
        ▼
Generic service invoke
        │
        ▼
Host Service Dispatcher
        │
        ├── runtime service
        ├── storage service
        ├── network service
        ├── data service
        └── other platform services
```

这里的关键不是“保留旧 opcode 再补一层新能力”，而是直接把宿主对 guest 的公开协议收敛为一条通用服务通道。也就是说，当前只保留一个统一的`service invoke`入口，后续扩展能力时不再继续堆叠新的专用 opcode。

选择这条路线的原因：

- 能保留当前`Wasm`宿主导入边界，不影响现有 guest 编译模型。
- 能把当前最小实现直接重构到统一协议下，避免后续同时维护“旧 opcode 语义”和“新服务语义”两套模型。
- 能把未来的扩展点集中到服务注册表和治理层，而不是不断改底层 ABI。

备选方案与取舍：

- 继续按能力增加专用 opcode：实现最直接，但协议和代码会迅速碎片化，放弃。
- 为每个宿主服务增加独立`wasmimport`函数：guest API 表面上更直观，但 ABI 面会失控，兼容成本更高，放弃。
- 让 guest 通过宿主 HTTP API 回环调用能力：会把内部治理绕回公开接口，且容易形成自调用与认证歧义，放弃。

### 决策二：采用“宿主服务声明 ＋ 内部能力推导 ＋ 资源授权”双层治理模型

仅靠`capabilities`字符串列表，无法表达复杂宿主服务的真实治理边界。比如声明了`host:http:request`之后，插件到底能调用哪个上游、响应体能有多大，这些都不是一个字符串能说明的；但如果再要求作者同时维护`capabilities`和`hostServices`两份声明，又会引入重复输入和额外维护成本。

因此本次采用“作者侧单一声明 + 宿主内部双层治理”的模型：

1. 插件作者只声明细粒度`hostServices`策略：
   在`plugin.yaml`中结构化声明 service、method、资源目标（如`resourceRef`、URL 模式或`resources.tables`）和治理参数，用于表达“这个插件申请访问什么资源、希望以什么方式访问”。声明本身只代表申请，不代表自动授权；宿主需要在安装或启用时展示这些申请项，并确认最终授权结果。
2. 宿主内部自动推导粗粒度 capability：
   宿主根据`hostServices.methods`自动推导`host:runtime`、`host:storage`、`host:http:request`、`host:data:read`、`host:data:mutate`等能力分类，用于运行时快速拒绝未授权的大类能力，但这份 capability 集合不再要求插件作者额外维护。

推荐的清单形态如下：

```yaml
hostServices:
  - service: storage
    methods: [put, get, delete, list, stat]
    resources:
      paths:
        - reports/
        - exports/daily.json
  - service: network
    methods: [request]
    resources:
      - url: https://*.example.com/api
  - service: data
    methods: [list, get, create, update, delete]
    resources:
      tables:
        - biz_ticket
        - biz_ticket_comment
```

宿主在安装／启用和运行时要分别校验：

补充落地约束：

- 面向插件作者的示例 `plugin.yaml` 中，凡是本次迭代新增的 `hostServices`、`resources.paths`、`resources.tables`、URL pattern 等字段，样例文件都应提供就地注释说明，确保清单本身即可作为开发模板；
- 动态插件 guest 侧的控制器实现应保持轻量，复杂业务负载逻辑统一下沉到 `backend/internal/service/<component>/` 维护，控制器只负责桥接请求上下文、调用 service 并回写响应。
- `pkg/pluginhost` 中历史遗留的 source-plugin `ResourceSpec` 若未被任何源码插件使用，应优先删除而不是继续为其维护与当前 `catalog.ResourceSpec` 平行的第二套模型；当前 data hostService 的表授权与动态运行时 backend resource 契约统一以宿主内部 `catalog` / runtime artifact 模型为准。

- 声明是否完整且合法；
- 宿主是否确认授权该 service／method／治理目标（例如`resourceRef`、URL 模式或`table`）；
- 宿主是否需要对声明做收窄、覆盖或拒绝；

宿主运行时继续校验：

- capability 是否已由当前 release 的`hostServices`快照推导并授予；
- service／method 是否声明；
- 目标资源标识（`resourceRef`、URL 模式或`table`）是否已被宿主确认授权；
- 当前执行上下文是否满足服务方法要求；
- 调用参数是否满足该资源的治理策略。

选择双层模型的原因：

- 安全边界更清晰，既能在能力层快速拒绝，也能在资源层精确治理。
- 更接近`Higress`用逻辑集群／资源封装网络和缓存能力的模式。
- 可以直接复用`sys_plugin_resource_ref`作为治理投影，而不必再新建一套资源归属体系。

### 决策三：宿主服务注册表要显式区分请求型上下文和系统型上下文

插件调用宿主服务时，并不总是处在一个带用户身份的 HTTP 请求里。动态插件可能在以下两类上下文里运行：

- 请求型上下文：由动态路由、前端触发的后端调用进入，天然带有当前用户身份、角色权限、数据范围和请求链路信息。
- 系统型上下文：由 Hook、定时任务、安装／升级流程等宿主任务触发，不一定存在当前用户。

因此，每个宿主服务方法都需要显式声明自己的上下文要求：

- `request-bound`：必须有当前用户身份和请求上下文，例如需要复用数据范围的数据查询。
- `system-bound`：允许在无用户上下文时调用，例如写插件私有缓存、落地中间文件。
- `both`：两类上下文都能调用，但宿主会根据是否存在用户上下文决定额外治理。

宿主在分发宿主服务调用时，统一传入一个`ExecutionContext`快照，至少包括：

- `pluginId`
- 执行来源（route、hook、cron、lifecycle）
- 当前 route 或 Hook 标识
- 当前用户身份快照（若存在）
- 当前数据范围快照（若存在）
- request ID 和 deadline

这一步是为了让后续的`storage`、`network`、`data`服务都能在相同模型下工作，并且能把错误信息做成结构化输出。

### 决策四：宿主需要提供完整的 Host Call 能力地图，并按层分阶段交付

为了方便审查和后续实现收敛，本次把“宿主到底需要提供哪些 Host Call 能力”拆成完整能力地图，而不是只描述首批要做的三类服务。这里的“完整”有两层含义：

- 一是给出宿主对动态插件的完整能力边界，明确哪些能力已经存在、哪些在本迭代高优先级交付、哪些在本迭代低优先级交付、哪些明确禁止提供。
- 二是把每类能力对应的交互模型写清楚，避免把所有能力都错误地建模成同一种 Host Call。

#### 4.1 宿主 Host Call 交互模型分层

宿主对 guest 的能力暴露分为三层：

| 分层 | 交互方式 | 适用能力 | 说明 |
| --- | --- | --- | --- |
| 通用服务调用 | `service invoke`统一 envelope，同步请求／响应 | `runtime`、`storage`、`network`、`data`、`cache`、`lock`、`notify` | 所有宿主能力统一走这一层 |
| 隐式上下文注入 | 不暴露独立 hostcall，由 bridge 或执行器注入 | 用户身份、数据范围、当前路由、Hook 元数据、deadline | 这些信息是运行时上下文，不应被建模成宿主资源操作 |

其中：

- 包括当前已实现的`host:log`、`host:state`、`host:db:*`在内，所有宿主能力都可以统一重构到`service invoke`通道；
- 后续新增复杂敏感能力不得再暴露新的专用 opcode；
- `ExecutionContext`、当前用户身份、请求元数据等信息由宿主注入，不单独开放为一个“可任意读取宿主内部状态”的 hostcall。

#### 4.2 宿主需要提供的完整能力清单

##### A. 基础宿主能力

| 服务 | capability 建议值 | methods | 上下文要求 | 资源绑定 | 用途 |
| --- | --- | --- | --- | --- | --- |
| `runtime.log` | `host:runtime` | `write` | `both` | 无 | 输出带`pluginId`前缀的结构化日志 |
| `runtime.state` | `host:runtime` | `get`、`set`、`delete` | `both` | 无 | 读写按`pluginId`隔离的键值状态 |
| `runtime.info` | `host:runtime` | `now`、`uuid`、`node` | `both` | 无 | 提供宿主时间、唯一 ID 和节点基础信息 |

##### B. 本迭代高优先级交付能力

| 服务 | capability 建议值 | 首期 methods | 上下文要求 | 资源绑定 | 说明 |
| --- | --- | --- | --- | --- | --- |
| `runtime` | `host:runtime` | `log.write`、`state.get`、`state.set`、`state.delete`、`info.now`、`info.uuid`、`info.node` | `both` | 无 | 统一承载日志、状态和宿主基础信息能力 |
| `storage` | `host:storage` | `put`、`get`、`delete`、`list`、`stat` | `both` | `host-storage` | 用逻辑路径或路径前缀替代宿主物理路径，承接文件和对象存储访问 |
| `network` | `host:http:request` | `request` | `both` | `host-upstream` | 只开放受授权上游的同步 HTTP 请求 |
| `data` | `host:data:read`、`host:data:mutate` | `list`、`get`、`create`、`update`、`delete`、`transaction` | `request-bound`优先，部分`system-bound` | `host-data-table` | 用宿主确认授权的数据表替代直接数据库连接和原始 SQL |

##### C. 本迭代低优先级交付能力

这些能力纳入本迭代规划，但优先级低于`runtime`、`storage`、`network`、`data`四类核心能力。实现顺序上必须先完成前四类核心能力，再继续推进这些能力。

| 服务 | capability 建议值 | 首期 methods | 上下文要求 | 资源绑定 | 规划原因 |
| --- | --- | --- | --- | --- | --- |
| `cache` | `host:cache` | `get`、`set`、`delete`、`incr`、`expire` | `both` | `host-cache` | 很多插件需要短期缓存，但不应直接拿 Redis 客户端 |
| `lock` | `host:lock` | `acquire`、`renew`、`release` | `system-bound`优先 | `host-lock` | 可复用宿主现有分布式锁能力，避免插件自行实现并发协调 |
| `notify` | `host:notify` | `send` | `both` | `host-notify-channel` | 为插件提供站内信、邮件、Webhook 等统一通知出口 |

##### D. 明确不提供的能力

以下能力即使用户有短期诉求，也不应设计成宿主 Host Call：

| 不提供能力 | 原因 |
| --- | --- |
| 宿主绝对路径读写、任意目录遍历 | 会直接破坏宿主文件系统边界 |
| 原始 socket、任意域名直连、内网自由探测 | 风险过高，且绕过上游治理 |
| 宿主`ghttp.Request`、数据库连接、Go `service`实例直出 | 会把动态插件退化成“不受限源码插件” |
| 任意 shell 执行、进程管理、系统命令调用 | 超出后台业务插件可接受安全边界 |
| 无资源绑定的通用 SQL root 能力 | 无法与数据权限、审计和资源治理模型兼容 |

#### 4.3 宿主逻辑资源与绑定模型

结构化宿主服务的治理对象，不是单独一个`service`名称，而是`service + method + governed target`的组合。对`storage`，governed target 是逻辑`path`；对`network`，governed target 是`URL pattern`；对`data`，governed target 是`table`；对`cache`、`lock`和`notify`等低优先级服务，governed target 仍暂按逻辑`resourceRef`规划。所有这些声明统一视为权限申请。真实资源绑定与最终授权由宿主管理员、安装流程或平台预置配置完成，并以当前 release 快照为准。

| 资源类型 | 对应 service | 插件侧逻辑引用示例 | 宿主实际绑定对象 | 核心治理字段 |
| --- | --- | --- | --- | --- |
| `host-storage` | `storage` | `reports/` | 宿主为插件隔离出的逻辑存储路径空间 | 路径边界、目录前缀、默认大小与平台保护 |
| `host-upstream` | `network` | `https://*.example.com/api` | URL 模式命中的 HTTP 地址集合 | URL 模式本身；安装/启用时由宿主确认授权 |
| `host-data-table` | `data` | `sys_plugin_node_state` | 宿主确认授权的数据表 | 可执行操作、表级审计、数据范围、事务边界 |
| `host-cache` | `cache` | `ticket-cache` | 基于 MySQL `MEMORY` 表的分布式 KV 缓存空间 | TTL、key 前缀、值长度上限、淘汰与清理策略 |
| `host-lock` | `lock` | `ticket-lock` | 按插件隔离的分布式锁命名空间 | 租期、续租上限、持有者票据约束、竞争策略 |
| `host-notify-channel` | `notify` | `ops-mail` | 统一通知域中的站内信、邮件、Webhook 通道 | 模板约束、速率限制、接收者范围 |

统一绑定流程如下：

1. 插件在`plugin.yaml`中声明自己依赖哪些受治理目标；对`storage`是逻辑`path`，对`network`是`URL pattern`，对`data`是`table`，对`cache`、`lock`、`notify`等低优先级服务仍是逻辑`resourceRef`，但其语义在本期进一步明确为：`cache.resourceRef = 缓存命名空间`、`lock.resourceRef = 逻辑锁名`、`notify.resourceRef = 通知通道标识`；这些声明统一表示权限申请；
2. 构建器只校验声明是否合法，不在 guest 侧固化真实物理资源地址，也不把声明视为已授权；
3. 宿主在安装或启用插件时向管理员展示申请的 service／method／目标标识（`path`、`URL pattern`、`resourceRef`或`table`）及治理参数，并允许批准、收窄或拒绝；
4. 宿主把最终确认结果绑定到真实受治理资源，并形成当前 release 的授权快照；
5. 运行时仅根据这份授权快照解析目标标识，插件始终看不见底层系统对象，也不能依赖未获确认的声明。

这种模型的关键收益是：插件开发者面向稳定逻辑能力编程，宿主平台面向真实资源治理和审计，两者之间通过 release 快照解耦。

#### 4.4 各能力组的设计边界

##### 基础运行时能力组

基础运行时能力组只负责让插件“能运行、能记录、能保留少量状态、能读取少量宿主基础元数据”，不负责承载复杂资源访问。它包含：

- `runtime.log`
- `runtime.state`
- `runtime.info`
- 通用`service invoke`入口本身

这一组能力的原则是：能力面越小越稳定，不为单个业务需求继续膨胀，也不为历史实现保留平行协议。

##### 资源访问能力组

资源访问能力组是本次设计的核心，负责承接插件真正的敏感能力需求。它包含：

- `storage`
- `network`
- `data`

这一组能力必须同时具备以下特征：

- 有清晰的受治理目标绑定；
- 能被宿主审计；
- 能根据请求型／系统型上下文做差异化校验；
- 失败时不会污染宿主主流程。

##### 平台协同能力组

平台协同能力组在本迭代属于低优先级交付能力，但它们代表宿主平台对复杂动态插件开放的上限。它包含：

- `cache`
- `lock`
- `notify`

这一组能力之所以在本迭代一并纳入规划，是因为很多复杂插件会直接提出这些需求。若只做“未来预留”而不进入本轮任务管理，后续很容易再次出现范围漂移和协议临时扩展。

#### 4.5 完整能力地图对应的 guest API 形态

为了避免插件作者直接拼装底层 envelope，guest 侧需要形成与上表一一对应的 SDK：

| 宿主能力 | guest SDK 建议形态 |
| --- | --- |
| `runtime.log` | `pluginbridge.Runtime().Log(...)` |
| `runtime.state` | `pluginbridge.Runtime().StateGet/Set/Delete(...)` |
| `runtime.info` | `pluginbridge.Runtime().Now/UUID/Node(...)` |
| `storage` | `pluginbridge.Storage().Put/Get/Delete/List/Stat(...)` |
| `network` | `pluginbridge.HTTP().Request(...)` |
| `data` | `plugindb.Open().Table(...).WhereEq/WhereIn/WhereLike(...).Page(...).All/One/Count/Insert/Update/Delete/Transaction(...)` |
| `cache` | `pluginbridge.Cache().Get/Set/Delete/Incr(...)` |
| `lock` | `pluginbridge.Lock().Acquire/Renew/Release(...)` |
| `notify` | `pluginbridge.Notify().Send(...)` |

其中：

- 本迭代必须先实现`runtime`、`storage`、`network`、`data`四组高优先级 SDK；
- `cache`、`lock`、`notify`三组 SDK 作为低优先级能力跟进实现。
- `runtime`、`storage`和`network`首期继续由`pluginbridge`直接暴露高层 helper；`data`则新增`pkg/plugindb`作为推荐入口，通过更接近 GoFrame ORM 的链式 API 降低插件开发者的学习成本。

### 决策五：首批扩展能力分三组落地，且全部走逻辑资源而不是底层系统对象

#### 1. `storage` service

`storage` service 负责文件和对象存储相关能力，但插件只能看见逻辑存储空间，不能看见宿主真实路径。

首批建议方法：

- `put`
- `get`
- `delete`
- `list`
- `stat`

治理原则：

- 插件不再通过`resourceRef`声明存储空间，而是直接通过`resources.paths`声明需要访问的逻辑路径或路径前缀。
- 这些 path 是插件可见的逻辑路径，不是宿主文件系统绝对路径。
- 宿主始终把逻辑 path 映射到插件隔离的内部存储根目录，不向 guest 暴露底层目录、对象存储桶或文件模块实现。

本期落地的第一版宿主绑定语义如下：

- `plugin.yaml`中的`storage.resources`统一声明为`paths`，例如：
  - `reports/`
  - `exports/daily.json`
- `reports/`表示目录前缀授权，允许访问`reports/...`下的对象；
- `exports/daily.json`表示单路径授权，只允许访问这个逻辑对象；
- guest 运行时直接提交目标逻辑路径，例如`reports/demo.json`；
- 宿主先对目标路径做相对路径归一化和越界校验，再根据当前 release 授权快照判断该路径是否命中某条已授权 path；
- 宿主底层继续把这些逻辑路径映射到插件专属隔离目录，例如`.host-services/storage/<pluginId>/...`，但该物理路径不对插件暴露；
- `put/get/delete/list/stat`五个方法全部返回结构化响应，其中`get/stat`对“对象不存在”返回`found=false`，而不是把缺失对象混入协议级错误。

这样做的原因，是文件能力本质上应当是“受治理的逻辑存储访问”，而不是“宿主文件系统 syscall”。

##### `storage` paths 的明确匹配规则

为了方便安装审查、运行时实现和安全复核，`storage.resources.paths`在本期明确为以下规则：

1. **声明合法性**
   - path 必须是逻辑相对路径，不能是宿主绝对路径；
   - path 不允许包含越界语义，例如`..`、盘符前缀或试图跳出插件隔离目录的形式；
   - `storage` 不再要求声明独立`resourceRef`，也不再把`attributes`作为插件公开声明面的必要组成部分。

2. **标准化规则**
   - 宿主在匹配前统一把 path 标准化为 `/` 分隔的相对逻辑路径；
   - 宿主会清理重复分隔符以及`.`、`..`等路径噪声；
   - 任何归一化后仍表现为越界的 path 都直接拒绝。

3. **授权匹配维度**
   - 以`/`结尾的 path 视为“目录前缀授权”；例如`reports/`允许访问`reports/a.json`与`reports/2026/summary.json`；
   - 不以`/`结尾的 path 视为“单路径授权”；例如`exports/daily.json`只允许访问该对象本身；
   - 前缀匹配必须按路径边界生效；`reports/`匹配`reports/a.json`，但不匹配`reports-v2/a.json`；
   - `list`方法只能列举自己已命中的目录前缀范围，不能借前缀外探测其他路径。

4. **默认拒绝原则**
   - 只要目标逻辑路径未命中任何已授权 path，宿主就拒绝本次请求；
   - 插件不能通过路径大小写差异、重复分隔符或`./`、`../`绕过授权边界；
   - 宿主拒绝的是“未命中授权边界的逻辑路径”，而不是文件不存在这类业务结果。

##### `storage` paths 匹配示例

| 已授权 path | 目标 path | 结果 | 说明 |
| --- | --- | --- | --- |
| `reports/` | `reports/a.json` | 允许 | 命中目录前缀 |
| `reports/` | `reports/2026/summary.json` | 允许 | 命中子目录 |
| `reports/` | `reports-v2/a.json` | 拒绝 | 未命中路径边界 |
| `exports/daily.json` | `exports/daily.json` | 允许 | 命中单路径 |
| `exports/daily.json` | `exports/monthly.json` | 拒绝 | 单路径授权不扩散 |
| `reports/` | `../reports/a.json` | 拒绝 | 越界路径在归一化前后都必须拒绝 |

#### 2. `network` service

`network` service 负责出站网络访问，首期只支持同步 HTTP 请求，不开放原始 socket。

首批建议方法：

- `request`

治理原则：

- 插件只声明自己要访问的 URL 模式，例如`https://*.example.com/api`。
- 所有`host-upstream`、`host-storage`、`host-data-table`、`host-cache`、`host-lock`和`host-notify-channel`都遵循同一规则：插件声明的是权限申请，真正授权在安装／启用阶段确认。
- 一旦宿主确认授权某个 URL 模式，插件即可直接对命中的 URL 发起 HTTP 请求，不需要再声明方法白名单、头白名单或独立的上游引用名。
- 宿主仍保留平台级默认保护，例如受保护 hop-by-hop 头过滤、默认 timeout 与默认响应体限制，但这些不再作为插件声明参数。

本期落地的第一版宿主绑定语义如下：

- 首批`network` service 仅实现同步 HTTP 请求，对 guest 暴露统一的`request`方法。
- `plugin.yaml`中的`network.resources`只需要声明 URL 模式，例如：
  - `https://api.example.com/v1`
  - `https://*.example.com/api`
- 宿主匹配时校验 scheme、主机模式和路径前缀；其中主机支持`*`模糊匹配，路径按前缀匹配。
- guest 运行时直接提交目标绝对 URL；宿主根据当前 release 授权快照判断该 URL 是否命中已授权模式。
- 上游返回`4xx/5xx`时协议仍返回成功 envelope，并把真实 HTTP 状态码写入结构化响应；只有宿主治理拒绝、超时、URL 非法或体积超限才进入协议级错误。

##### `network` URL pattern 的明确匹配规则

为了方便安装审查、运行时实现和后续安全复核，`network` 的 URL pattern 匹配规则在本期明确为以下几条：

1. **声明合法性**
   - URL pattern 必须是绝对 URL，且必须包含`http`或`https` scheme 与非空 host；
   - `network` 不再声明`allowMethods`、`headerAllowList`、`timeoutMs`、`maxBodyBytes`或独立`upstreamRef`；
   - URL pattern 本身就是插件申请访问边界的最小治理对象。

2. **标准化规则**
   - 宿主在匹配前会对目标 URL 做标准化：去除 fragment，空 path 视为`/`；
   - path 在比较前按统一规则归一化，消除重复分隔符、`.`、`..`等路径噪声；
   - host 比较不区分大小写。

3. **授权匹配维度**
   - **scheme**：必须精确匹配；`http`与`https`互不通配；
   - **host**：按 hostname 做 glob 匹配，当前实现使用与文件通配一致的`*`语义；例如`*.example.com`可匹配`a.example.com`，也可匹配`a.b.example.com`；
   - **port**：若 pattern 显式声明 port，则目标 URL 必须使用同一 port；若 pattern 未声明 port，则表示不额外限制 port；
   - **path**：按归一化后的前缀匹配；`/api`匹配`/api`与`/api/orders`，但不匹配`/api-v2`；pattern 为`/`时表示该 host 范围下所有 path 都可访问。

4. **不参与授权边界的维度**
   - query string 不参与授权匹配，因此`https://api.example.com/v1?debug=1`与`https://api.example.com/v1?debug=0`对授权来说是同一治理目标；
   - fragment 不参与授权匹配，也不会作为网络权限边界的一部分；
   - 若未来需要对 query、header、method 做更细粒度治理，应作为新的宿主治理能力演进，而不是回退到本期已移除的复杂声明模型。

5. **默认拒绝原则**
   - 只要 scheme、host、port、path 任一维度未命中授权 pattern，宿主就拒绝本次请求；
   - 插件不能通过改写 query、fragment 或大小写差异绕过授权；
   - 宿主拒绝的是“未命中授权边界的目标 URL”，而不是业务层的 HTTP `4xx/5xx` 响应。

##### `network` URL pattern 匹配示例

| 已授权 URL pattern | 目标 URL | 结果 | 说明 |
| --- | --- | --- | --- |
| `https://api.example.com/v1` | `https://api.example.com/v1/users` | 允许 | scheme、host 一致，且 path 命中前缀 |
| `https://api.example.com/v1` | `https://api.example.com/v1?debug=1` | 允许 | query 不参与授权匹配 |
| `https://api.example.com/v1` | `https://api.example.com/v10/users` | 拒绝 | `/v10` 不属于 `/v1` 的 path 前缀 |
| `https://*.example.com/api` | `https://foo.example.com/api/orders` | 允许 | host wildcard 命中，path 前缀命中 |
| `https://*.example.com/api` | `https://foo.example.com/api-v2/orders` | 拒绝 | `/api-v2` 不命中 `/api` 前缀 |
| `https://api.example.com:8443/v1` | `https://api.example.com:8443/v1/ping` | 允许 | 显式 port 完全匹配 |
| `https://api.example.com:8443/v1` | `https://api.example.com/v1/ping` | 拒绝 | pattern 指定了 port，目标 URL 未命中该 port |
| `https://api.example.com/v1` | `http://api.example.com/v1/ping` | 拒绝 | `http` 与 `https` 不互通配 |

#### 3. `data` service

`data` service 负责插件的数据访问能力，但默认不再继续把“原始 SQL”作为未来扩展的主通道。

首批建议方法：

- `list`
- `get`
- `create`
- `update`
- `delete`
- `transaction`

治理原则：

- 插件通过`resources.tables`声明申请访问的数据表，而不是直接获取宿主数据库连接。
- 宿主按`pluginId + table + method`治理授权边界；插件运行时直接以表名发起结构化 CRUD / transaction 请求。
- 请求型调用默认复用当前用户的数据范围和权限上下文。
- 系统型调用只能访问显式声明并被宿主确认授权的数据表。

本期对 `data service` 追加两个实现约束：

- guest 协议层禁止暴露 raw SQL、通用 SQL 执行器或任意条件片段拼接能力；允许表名直传，但表名只能来自清单中声明并经宿主确认授权的`resources.tables`集合；
- 宿主实现层优先通过受控 DAO 对象与 GoFrame `gdb` ORM 组件完成查询和写入，再通过宿主封装的拦截层统一治理最终数据库提交。

##### 数据表直连声明的正式定义

在本设计里，`data`不再引入额外的命名数据资源层，而是采用“插件声明表名、宿主确认授权、运行时按表执行”的直连模型：

- 清单侧只声明`methods`和`resources.tables`；
- 插件运行时请求直接携带`table`；
- 宿主仅允许访问当前 release 授权快照中的表名；
- 宿主内部仍可为每张表装配受控 DAO/ORM 策略、字段映射、数据范围注入和审计规则；
- 插件不能获取数据库连接、不能提交 raw SQL，也不能跳过宿主治理直接拼接任意数据库能力。

这样做的目的是进一步降低插件开发门槛：对插件作者来说，最稳定、最直观的治理对象就是“我要访问哪张表，以及要用哪些结构化方法”；对宿主来说，表级授权足以作为本期的最小可落地边界，后续若需要更强治理，再在宿主内部增加表级元数据策略即可，而不必额外暴露`resourceRef`概念。

##### 宿主内部执行模型

虽然 guest 侧不能触达 raw SQL，但宿主内部仍需要把结构化数据请求落到数据库。这里的建议实现顺序是：

1. guest 请求中的`table`先与当前 release 授权快照做匹配；
2. 表名再映射到宿主维护的受控 DAO / DO / `gdb.Model` 组装计划；
3. 查询、过滤、分页、排序、字段投影、写入字段和事务边界都在宿主代码中显式生成；
4. 最终数据库执行统一经过宿主侧的 `gdb` 拦截层。

这样做的原因是，数据治理的重点不是“隐藏 SQL 文本”本身，而是把可执行的数据操作收敛成宿主可验证、可审计、可限制的对象模型。

##### `gdb` Driver / DB Wrapper 拦截点

GoFrame 的 `gdb` 执行链路最终会进入 `DoCommit(ctx, gdb.DoCommitInput)`。本项目后续实现 `data service` 时，建议采用“自定义 Driver + 自定义 DB wrapper”的方式，把动态插件数据服务的最终数据库执行统一拦截到这一层，而不是在每个 DAO 调用点零散加判断。

具体建议如下：

- 自定义 Driver 负责返回宿主包装过的 `gdb.DB`；
- 包装后的 `gdb.DB` 覆盖或代理 `DoCommit(ctx, gdb.DoCommitInput)`；
- 宿主在上下文中注入 `pluginId`、`table`、`data method`、执行来源、请求用户和事务标识；
- `DoCommit` 拦截层在真正提交前执行二次权限核对、事务资源边界校验、字段级审计与风险策略；
- 对事务型调用，宿主统一跟踪 `BEGIN / COMMIT / ROLLBACK`，确保一次结构化 `transaction` 只作用于当前授权的数据表集合；
- 宿主审计记录以结构化资源动作和字段摘要为主，而不是把原始 SQL 作为 guest 协议的一部分暴露出来。

这意味着 `data service` 的宿主实现可以继续享受 GoFrame DAO/ORM 的抽象能力，但 guest 看到的始终只是结构化宿主服务，而不是数据库驱动接口。

##### 受限 ORM 风格 guest SDK 与 `plugindb` 分层

为了让动态插件的数据访问体验尽量贴近宿主当前 GoFrame ORM 使用方式，同时不破坏既有 ABI 和治理边界，本次为`data service`补充一层受限 ORM 风格 guest SDK，而不是把完整`gdb.DB`、`gdb.Model`或宿主 DAO 直接暴露给插件。

推荐分层如下：

```text
Dynamic Plugin Business Code
        │
        ▼
pkg/plugindb
        │
        ▼
pkg/pluginbridge
        │
        ▼
wasm data host service dispatcher
        │
        ▼
pkg/plugindb/host
        │
        ▼
internal/service/plugin/internal/datahost
        │
        ▼
GoFrame gdb.Model / DAO / DO
```

其中：

- `pkg/plugindb`是插件作者的推荐入口，提供`Open().Table(...).WhereEq(...).All()`这类受限 ORM 风格 API；
- `pkg/pluginbridge`继续作为稳定 ABI 与底层 hostService codec 层，不再承担数据访问的长期推荐高层体验；
- `pkg/plugindb/host`负责沉淀宿主侧可复用的 Driver / DB wrapper、审计上下文和`DoCommit`治理能力；
- `internal/service/plugin/internal/datahost`继续负责数据契约、资源校验、数据范围注入和执行编排。

##### 强类型枚举与 query plan 模型

`plugindb`相关实现必须严格遵守项目“枚举值使用独立类型和常量管理”的规范。所有带枚举语义的值都必须定义为独立 Go 命名类型与常量，禁止在 builder、query plan、执行器和审计逻辑中直接写字符串字面量。首批至少包括：

- `DataPlanAction`
- `DataFilterOperator`
- `DataOrderDirection`
- `DataMutationAction`
- `DataAccessMode`

guest SDK 内部不直接拼装原始 SQL，也不直接暴露完整`gdb.Model`，而是将链式调用收敛为结构化 query plan，例如：

```text
Table("sys_plugin_node_state")
  -> Fields(...)
  -> WhereEq / WhereIn / WhereLike
  -> OrderAsc / OrderDesc
  -> Page(...)
  -> One / All / Count / Insert / Update / Delete / Transaction
        │
        ▼
typed DataQueryPlan
        │
        ▼
host 侧校验与 gdb.Model 映射
```

query plan 的核心收益是：

- guest API 可以更接近 GoFrame ORM，而不是只停留在简单表操作 helper；
- 宿主仍然只接收结构化、可验证、可审计的请求对象；
- 后续若从当前`list/get/create/update/delete/transaction` method 进一步收敛为统一`plan` method，也不会影响插件作者的上层使用方式。

##### 迁移与兼容策略

本次不要求立即删除`pluginbridge.Data()`；它在过渡期仍可保留为兼容层。但后续文档、demo 和推荐用法应切换到`pkg/plugindb`。迁移顺序为：

1. 先补齐`pkg/plugindb/shared`强类型枚举与 query plan 模型；
2. 再实现`pkg/plugindb` guest SDK，并在底层先复用现有结构化`data` hostService；
3. 将宿主 Driver / DB wrapper 治理能力上提到`pkg/plugindb/host`；
4. 更新 demo、文档和测试，将`plugindb.Open()`作为主路径。

#### 4. `cache` service

`cache` service 负责为动态插件提供宿主统一治理的分布式 KV 缓存能力，但本期明确不提供节点内本机缓存封装。若只是单节点进程内缓存，插件完全可以自行实现；宿主需要提供的是“多节点共享、受治理、可审计”的缓存平面。

本期对 `cache` service 的明确约束如下：

- 宿主缓存底座使用 MySQL `MEMORY` 引擎数据表，而不是宿主进程内 map 或插件本地内存；
- `host-cache` 资源标识继续沿用低优先级服务的逻辑`resourceRef`模型，其语义正式定义为“缓存命名空间”；
- 运行时对动态插件的实际缓存隔离键使用 `pluginId + namespace + cacheKey` 组合，而不是只靠插件提交的原始 key；
- 由于 `MEMORY` 表字段长度与单行容量受限，宿主必须在写入前严格校验命名空间、key 和 value 的字节长度，超限直接报错，不得截断写入；
- `MEMORY` 表在数据库重启后内容丢失，这一语义与“缓存”本身一致，因此可接受；它不承担持久化状态职责。

本期建议的首版表结构如下；该表同时遵循项目 SQL 规范：使用自增整型 `id` 作为主键，且所有字段都必须显式带 `COMMENT`。

```sql
CREATE TABLE IF NOT EXISTS sys_kv_cache (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    owner_type  VARCHAR(16) NOT NULL DEFAULT '' COMMENT '所属类型：plugin=动态插件 module=宿主模块',
    owner_key   VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属标识：pluginId 或模块名',
    namespace   VARCHAR(64) NOT NULL DEFAULT '' COMMENT '缓存命名空间，对应 host-cache 资源标识',
    cache_key   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '缓存键',
    value_kind  TINYINT NOT NULL DEFAULT 1 COMMENT '值类型：1=字节串 2=整数',
    value_bytes VARBINARY(4096) NOT NULL COMMENT '缓存字节值，供 get/set 使用',
    value_int   BIGINT NOT NULL DEFAULT 0 COMMENT '缓存整数值，供 incr 使用',
    expire_at   DATETIME NULL DEFAULT NULL COMMENT '过期时间，NULL 表示永不过期',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_owner_namespace_key (owner_type, owner_key, namespace, cache_key),
    KEY idx_expire_at (expire_at)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='宿主分布式KV缓存表';
```

选择该结构的原因：

- `value_kind + value_bytes + value_int` 可以把普通 `set/get` 与 `incr` 的并发语义显式分开，避免把数值缓存退化为字符串读改写；
- `namespace` 继续复用宿主当前授权模型，不需要为 `cache` 额外发明一套不同于低优先级服务的资源声明外形；
- `owner_type + owner_key` 允许未来在动态插件之外复用同一宿主缓存组件，而不把表设计锁死在插件场景。

本期建议的默认长度上限为：

- `namespace <= 64 bytes`
- `cache_key <= 128 bytes`
- `value_bytes <= 4096 bytes`

宿主对超限写入统一返回显式错误，且不写入任何部分数据。`expire` 与周期性过期清理由宿主侧 `internal/service/kvcache` 组件统一处理。

#### 5. `lock` service

`lock` service 不再重新发明一套独立分布式锁实现，而是直接复用宿主现有 `sys_locker` 与 `locker.Service` 能力，再在动态插件运行时外层补一层票据化适配。

本期对 `lock` service 的明确约束如下：

- 不新增独立锁表，直接复用现有宿主分布式锁表与续租逻辑；
- `host-lock` 的 `resourceRef` 正式定义为“逻辑锁名”；
- 宿主落库时自动将逻辑锁名命名空间化为 `plugin:<pluginId>:<resourceRef>`，避免插件之间、插件与宿主内部锁之间互相冲突；
- 动态插件不会直接拿到宿主内存中的锁实例对象，而是拿到宿主签发的锁票据（ticket）；
- `renew` 与 `release` 都必须校验 ticket 与锁名、插件身份的匹配关系，不能只凭锁名操作。

这样做的原因是：当前宿主已经具备可用的分布式锁能力，本期真正新增的是“把它安全地暴露给动态插件”的治理层，而不是再维护第二套锁实现。

#### 6. `notify` service

`notify` service 必须与当前 `sys_notice` 的“通知公告内容管理”语义彻底解耦。`sys_notice` 继续负责公告内容的创建、编辑、草稿和发布，而统一通知域负责真实的消息分发、投递审计和用户收件箱呈现。

本项目是全新项目，没有历史债务，因此本期不保留 `sys_user_message` 兼容层，直接做结构重构：

- `sys_notice` 保留，继续作为公告内容表；
- `sys_user_message` 删除，不再作为长期消息中心表；
- 新增统一通知域表，承载消息主记录、投递记录和通知通道；
- `/user/message` 这套外部 API 与前端页面暂时保持不变，但其底层改为查询新的通知域表；
- 当前消息中心依赖的 `sourceType/sourceId` 语义必须保留，以便继续打开通知公告预览。

本期建议的首版通知域表如下；同样遵循“自增整型 `id` 主键 + 全字段注释”的项目 SQL 规范。

```sql
CREATE TABLE IF NOT EXISTS sys_notify_channel (
    id           BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    channel_key  VARCHAR(64) NOT NULL DEFAULT '' COMMENT '通道标识',
    name         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '通道名称',
    channel_type VARCHAR(32) NOT NULL DEFAULT '' COMMENT '通道类型：inbox=email=webhook',
    status       TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=停用',
    config_json  LONGTEXT NOT NULL COMMENT '通道配置JSON',
    remark       VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at   DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    UNIQUE KEY uk_channel_key (channel_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知通道表';

CREATE TABLE IF NOT EXISTS sys_notify_message (
    id             BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源插件ID，宿主内建流程为空',
    source_type    VARCHAR(32) NOT NULL DEFAULT '' COMMENT '来源类型：notice=公告 plugin=插件 system=系统',
    source_id      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源业务ID',
    category_code  VARCHAR(32) NOT NULL DEFAULT '' COMMENT '消息分类：notice=通知 announcement=公告 other=其他',
    title          VARCHAR(255) NOT NULL DEFAULT '' COMMENT '消息标题',
    content        LONGTEXT NOT NULL COMMENT '消息正文',
    payload_json   LONGTEXT NOT NULL COMMENT '扩展载荷JSON',
    sender_user_id BIGINT NOT NULL DEFAULT 0 COMMENT '发送者用户ID',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知消息主表';

CREATE TABLE IF NOT EXISTS sys_notify_delivery (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    message_id      BIGINT NOT NULL DEFAULT 0 COMMENT '通知消息ID',
    channel_key     VARCHAR(64) NOT NULL DEFAULT '' COMMENT '投递通道标识',
    channel_type    VARCHAR(32) NOT NULL DEFAULT '' COMMENT '投递通道类型',
    recipient_type  VARCHAR(32) NOT NULL DEFAULT '' COMMENT '接收者类型：user=email=webhook',
    recipient_key   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '接收者标识，如用户ID邮箱地址或Webhook标识',
    user_id         BIGINT NOT NULL DEFAULT 0 COMMENT '站内信用户ID，非站内信时为0',
    delivery_status TINYINT NOT NULL DEFAULT 0 COMMENT '投递状态：0=待发送 1=成功 2=失败',
    is_read         TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读：0=未读 1=已读',
    read_at         DATETIME NULL DEFAULT NULL COMMENT '已读时间',
    error_message   VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '失败原因',
    sent_at         DATETIME NULL DEFAULT NULL COMMENT '发送完成时间',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at      DATETIME NULL DEFAULT NULL COMMENT '删除时间',
    KEY idx_message_id (message_id),
    KEY idx_user_inbox (user_id, channel_type, is_read),
    KEY idx_channel_status (channel_key, delivery_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知投递记录表';
```

本期的通知域职责划分明确如下：

- `sys_notice`：公告内容管理；
- `notify.Service`：通知发送编排与通道治理；
- `sys_notify_message`：消息主记录；
- `sys_notify_delivery`：投递记录与用户 inbox；
- `usermsg.Service`：对现有 `/user/message` API 的兼容 facade。

这意味着：

- `notice` 发布时不再直接写 `sys_user_message`，而是调用 `notify.Service.Send(...)`；
- `notify` 首个真实落地的通道是 `inbox`，`email/webhook` 先保留通道与 sender 扩展位；
- 当前消息中心展示、未读计数、已读、删除、清空等能力都迁移到 `sys_notify_delivery`；
- 通知公告预览仍通过 `sourceType=notice` 与 `sourceId=<noticeId>` 串联前端行为，避免重构后破坏既有 UI 交互。

### 决策六：运行时产物和清单都要携带宿主服务治理快照

为了让宿主在装载、切换、审计和回滚时拥有完整真相源，本次不把宿主服务治理信息只留在源码目录或运行时内存里，而是要求：

- `plugin.yaml`仅显式声明`hostServices`；
- 构建器对声明做归一化和静态校验；
- 运行时产物把归一化结果嵌入专用自定义节；
- 宿主装载产物时恢复为 active release 的宿主服务治理快照；
- 宿主将服务相关的路径、URL 模式、表名与低优先级服务 `resourceRef` 同步到`sys_plugin_resource_ref`。

建议新增一个专用产物区段，例如：

- `lina.plugin.backend.host-services`

该区段保存：

- service 名称
- methods
- `storage.resources.paths`
- `network` URL pattern
- `data.resources.tables`
- `cache/lock/notify` 的逻辑 `resourceRef`
- 策略参数
- 协议版本与治理参数

这样可以保证：

- 宿主不需要在请求链路反查插件源码目录；
- 升级、回滚和多节点收敛都以 release 快照为准；
- 管理员可以直接查看插件声明过哪些宿主服务与资源。

### 决策七：沿用`sys_plugin_resource_ref`作为宿主服务资源治理投影

仓库里已经有`sys_plugin_resource_ref`和对应同步逻辑，用于承载插件 release 的治理资源索引。本次不再为宿主服务资源新建一套平行表，而是扩展现有资源类型语义，一次性纳入以下几类逻辑资源：

- `host-storage`
- `host-upstream`
- `host-data-table`
- `host-cache`
- `host-lock`
- `host-notify-channel`

这样做的好处：

- 安装、升级、卸载和回滚时可以沿用现有资源同步与软删除逻辑。
- 管理员查看插件治理信息时，不需要切换到另一套资源视图。
- 高优先级和低优先级宿主服务都能共用一套资源授权、审计和回滚模型，不需要等后续能力落地时再重构资源治理层。
- 该表表达的是“当前 release 被宿主纳入治理的资源索引”，并不依赖作者侧是否显式写过名为`resourceRef`的字段；因此在`data`、`storage`、`network`等已收敛到表名、路径、URL 模式的模型下，继续沿用该表仍然有明确价值。

### 决策八：宿主服务调用统一纳入限额和审计

复杂宿主能力一旦放给动态插件，最容易出问题的不是“能不能调通”，而是“出了问题能不能及时止损并追责”。因此，本次明确规定宿主服务调用都必须带上统一的治理字段：

- timeout
- 最大请求／响应体
- 资源引用
- 调用结果状态
- 调用耗时
- 是否命中权限／资源拒绝

宿主统一记录一条调用审计摘要，但默认不记录敏感 payload 原文，只保留必要的错误摘要和诊断标识。

### 决策九：第一阶段只做同步调用，异步票据模型后置

`Envoy`和`Higress`都大量使用异步回调模型，但`Lina`当前动态路由运行时仍然是同步请求／响应桥接模型。为了避免一次性把复杂度拉满，本次第一阶段明确只支持同步宿主服务调用，并通过严格的 timeout 和大小限制控制风险。

异步票据、回调恢复、流式处理等能力后续再评估是否进入下一轮迭代。这样可以先验证最核心的业务诉求：文件、网络、数据这三类复杂能力能否被稳定、可治理地发布给动态插件。

## Risks / Trade-offs

- [风险] 一次性把七类宿主能力全部纳入迭代，可能导致交付顺序失控。→ Mitigation：明确分成高优先级四类和低优先级三类，任务顺序和验收顺序都以前四类为前置。
- [风险] 数据服务若设计得过于理想化，落地时可能与真实插件诉求脱节。→ Mitigation：首批样例直接走结构化`data` service，必要时补充命名查询和命令模型，但不回退到 raw SQL 协议。
- [风险] 文件和网络能力天然敏感，容易越权。→ Mitigation：必须同时做“由`hostServices`推导出的 capability 校验”、`hostServices`策略校验、`resourceRef`/`path`/`table`/URL pattern 授权校验和上下文校验，任何一层不满足都拒绝执行。
- [风险] 统一通知域重构会影响当前通知公告发布与消息中心链路。→ Mitigation：保留`sys_notice`作为内容表，删除`sys_user_message`后将`/user/message`收敛为 facade，并对通知公告与消息中心相关 E2E 用例做完整回归。
- [风险] `MEMORY` 缓存表若不做长度控制，运行时容易出现写入失败或隐式截断风险。→ Mitigation：宿主在 `cache` service 与 `kvcache` 组件双层执行字节长度校验，超限直接返回显式错误，禁止截断写入。
- [风险] 宿主服务元数据如果只存在清单或只存在数据库，会形成双真相。→ Mitigation：以运行时产物内嵌快照作为 release 真相源，数据库只保存治理投影和审计记录。
- [风险] 现有探索性实现已经编码到若干包内，重构时可能出现局部返工。→ Mitigation：明确本项目是绿地项目，直接以目标模型重构，不为旧协议额外保留分支。

## Migration Plan

1. 在`pluginbridge`中定义统一的宿主服务调用 envelope 和 guest 低层 helper。
2. 扩展`plugin.yaml`与构建器，支持`hostServices`声明的静态校验和产物自定义节写入，并拒绝旧的顶层`capabilities`作者输入。
3. 在运行时解析链路中恢复宿主服务治理快照，并接入宿主服务注册表与统一分发器。
4. 将当前最小 Host Call 实现重构为`runtime`、`storage`、`network`、`data`四类高优先级宿主服务，并补齐上下文校验、资源授权和审计能力。
5. 在`data`能力上新增`pkg/plugindb/shared`强类型枚举与 query plan 模型，以及`pkg/plugindb`受限 ORM 风格 guest SDK。
6. 将`data service`当前自定义 Driver / DB wrapper 与审计上下文能力上提到`pkg/plugindb/host`，形成宿主可复用治理层。
7. 更新动态插件样例、开发文档和自动化测试，将`plugindb.Open()`作为主路径，并验证授权成功、授权拒绝、事务边界和资源限制场景。
8. 在前四类核心能力稳定后，继续实现`cache`、`lock`、`notify`三类低优先级宿主服务，其中`cache`基于 MySQL `MEMORY` 表落地分布式 KV 缓存，`lock`复用现有宿主分布式锁，`notify`重构为统一通知域。
9. 删除`sys_user_message`及其旧实现依赖，将`notice`发布链路与`/user/message` facade 统一迁移到新的通知域表，并回归通知公告与消息中心相关 E2E。

## Merged Scope: `config-duration-unification`

为保持当前仓库仅存在一个活跃 OpenSpec 变更，原 `config-duration-unification` 的设计决策并入本迭代统一维护。该并入范围与动态插件实现解耦，但属于同一轮交付和归档记录。

### 背景

`lina-core` 的时长配置原先同时存在两种表达方式：一类使用整数并将单位写入字段名（如 `expireHour`、`cleanupMinute`、`intervalSeconds`），另一类直接使用 duration 字符串（如 `30s`、`5m`）。这种混用方式迫使业务层重复做 `time.Hour`、`time.Minute`、`time.Second` 换算，也让配置含义不统一。

### 并入后的统一决策

1. 配置文件统一使用 duration 字符串，代码内部统一使用 `time.Duration`；相关键名固定为 `jwt.expire`、`session.timeout`、`session.cleanupInterval`、`monitor.interval`。
2. 不保留任何旧键兼容逻辑；当前项目属于全新项目，错误配置应在启动期尽早暴露，而不是长期保留过渡分支。
3. 由配置层显式解析 duration，而不是继续依赖通用 `Scan` 隐式换算，以便集中给出默认值、非法输入错误和最小粒度约束。
4. 会话清理与监控采集统一基于解析后的 `time.Duration` 生成调度表达，优先采用 `@every <duration>` 风格，而不是再次将 duration 拆回分钟或秒去拼 cron 表达式。

### 风险与迁移

- duration 字符串若漏写单位或单位错误，会在启动期解析失败；通过默认值、明确报错和单元测试覆盖来缓解。
- `@every` 调度与旧 cron 表达式的首轮触发时机存在差异；通过保留启动后的即时采集逻辑来控制行为变化。
- 迁移顺序保持为：先更新默认配置与模板，再收敛配置服务解析，随后调整认证、在线会话和监控模块消费方式，最后补齐测试与规范文档。

## Open Questions

- 第一阶段的`data` service 是否只支持结构化 CRUD 与事务编排，还是要同时支持命名查询模板？
- 第一阶段的`network` service 是否只支持同步 HTTP，还是要把`gRPC`一并纳入首批能力面？
- 插件管理界面是否要在本轮同时展示`hostServices`治理信息，还是先只在清单快照和后端接口层可见？

## Feedback Follow-Ups

### `internal/service` 组件统一接口返回模型

宿主与插件 `internal/service` 目录下的组件统一收敛为“导出接口 + 私有实现结构体”模式：各组件对外只暴露 `Service` 接口和 `New()` 构造函数，具体实现结构体改为包内私有类型。这样做有两个直接收益：

- 调用方不再依赖具体 `Service` 结构体，后续更容易按组件替换实现、注入 mock 或按需拆分子实现；
- 组件公开能力边界可以直接在 `Service` 接口中集中呈现，维护者无需再在多个实现文件中反向查找所有 exported 方法。

该约束适用于 `lina-core` 与插件侧所有宿主 `service` 组件，以及插件动态运行时内部的 `catalog` / `runtime` / `integration` / `frontend` / `lifecycle` / `openapi` 等子组件。包内私有辅助方法和状态仍保留在实现结构体上，不进入公开接口面。

接口定义位置也统一收敛：不额外新增 `*_contract.go` 文件，而是直接内联在各组件主文件中维护，例如 `internal/service/auth/auth.go`、`internal/service/plugin/plugin.go`、`backend/internal/service/dynamic/dynamic.go`。主文件顶部按“公开类型/输入输出模型 -> `Service` 接口 -> `var _ Service = (*serviceImpl)(nil)` -> `serviceImpl` -> `New()`”的顺序组织，这样组件入口文件本身就能完整展示该组件的公开契约与主实现装配关系。

### 宿主与源码插件静态接口统一声明式权限模型

宿主静态 API 与源码插件静态路由需要共用同一套“接口声明 -> 中间件校验”的权限模型。对于进入受保护路由组的静态接口，权限要求统一通过接口元数据中的 `permission` 声明，不再允许控制器方法自行散落式调用权限判断。这样可以保证：

- 宿主控制器、源码插件控制器与动态插件静态桥接接口使用一致的声明方式；
- 权限页面中维护的按钮或菜单权限码可以直接映射到接口声明，避免出现“前端按钮已隐藏，但后端接口仍依赖人工补校验”的维护风险；
- 当前用户自助接口（如用户信息、个人资料、消息中心、登出等）仍可保留为“仅登录态”能力，不强制要求按钮级权限码。

对于缺少按钮级权限码但已有菜单权限码的只读接口，可直接复用对应菜单权限声明；权限上下文在运行时应同时识别有效菜单权限与按钮权限，避免只因为权限类型不同就退回控制器手工判断。

### 访问上下文缓存与失效策略

声明式权限中间件会在每个请求上读取当前用户的角色、菜单与权限集合。如果每次都重新查询角色、角色菜单和菜单权限，会形成稳定的多次数据库读放大。为降低这部分开销，运行时需要引入“按登录 token 缓存访问上下文”的策略：

- 登录成功后即预热当前 token 的访问上下文缓存；
- 后续接口权限校验优先读取 token 对应的缓存快照，避免重复装配角色 / 菜单 / 权限；
- 登出、强制下线、会话失效时立即清理 token 对应缓存；
- 角色、菜单、插件权限拓扑发生变化时，需要提升全局权限拓扑版本并使已有缓存失效，保证后续请求能够回源装配最新权限；
- 集群模式下允许各节点保留各自的本地 token 缓存，但权限拓扑版本必须通过共享存储对齐，避免只有当前节点感知到权限变更。

该策略的目标不是把权限数据永久固化到 token 中，而是在保证权限变更最终可见的前提下，将高频请求路径上的重复数据库读取收敛为“缓存命中 + 轻量版本校验”。
