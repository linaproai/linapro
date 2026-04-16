## Context

`Lina`当前已经具备动态插件`Wasm`产物治理、前端静态资源托管、插件安装与启停、版本切换和后端`hook`装载能力，但后端动态`REST`扩展此前仍停留在“合同治理已建立、真实执行未接入”的阶段。

结合本轮探索与实现，动态插件后端扩展的定位进一步收敛为：

- 动态插件是**受限业务扩展**，不是源码插件的等价替代；
- 动态插件公开路径只允许位于固定前缀`/api/v1/extensions/{pluginId}/...`下；
- 路由治理元数据统一收敛在`api`层请求结构体的`g.Meta`中；
- 宿主负责登录校验、权限校验、业务上下文注入和文档投影；
- 插件运行时只负责受限业务处理，不直接获得宿主`ghttp`上下文，也不自由组合宿主中间件。

此前我们评估过“在`WASI`内直接运行完整`GoFrame ghttp.Server`”的路线，但在当前代码与运行时约束下并不现实。最终落地方案改为：

- 宿主继续掌握路由解析与治理；
- 动态插件提供一个**极简受限的`Wasm bridge`运行时**；
- 宿主将治理后的统一请求快照序列化后写入`Wasm`内存；
- 动态插件在受限桥接层内完成业务处理并返回统一响应快照。

这个方案更符合此前讨论的目标：既要让`AI agent`能低成本生成插件，又要保证动态扩展简单、边界清晰、运行安全。

## Goals / Non-Goals

**Goals**

- 固化动态插件公开路由命名空间：`/api/v1/extensions/{pluginId}/...`。
- 将动态路由治理元数据统一放在`g.Meta`中维护。
- 让运行时产物携带完整路由合同与桥接合同，宿主无需在请求时扫描源码。
- 让宿主在固定前缀下完成快速匹配、鉴权、权限校验和业务上下文注入。
- 让动态插件固定前缀路由复用宿主统一的`RouterGroup + Middleware`注册方式，避免单独维护一条旁路分发入口。
- 让动态路由优先通过真实`Wasm bridge`执行，未声明可执行桥接时再回退到`501`占位。
- 让动态插件权限继续复用宿主现有`sys_menu.perms`体系。
- 让系统`OpenAPI`按运行时状态投影动态插件公开接口。
- 让宿主到动态插件的桥接`DTO`使用高效二进制编解码，禁止使用`json`或纯文本协议承载请求／响应信封。
- 将动态插件可复用的桥接合同、编解码与运行时辅助逻辑沉淀到`lina-core/pkg`公共组件，降低插件业务代码编写复杂度。

**Non-Goals**

- 不在`WASI`中运行完整`GoFrame`私有`HTTP`服务器。
- 不让动态插件占用宿主任意全局路径空间。
- 不向动态插件开放宿主`Auth`、`Ctx`、`OperLog`等中间件自由拼装能力。
- 不为动态插件保留一条脱离宿主统一中间件注册方式的专用旁路入口。
- 不支持`WebSocket`、`SSE`、流式响应和长连接协议。
- 不为每条动态路由在宿主再定义一层额外业务`DTO`映射。
- 不让动态插件直接拿到宿主原始登录令牌或宿主内部服务实例。
- 不在`lina-core`运行时业务组件中保留源码扫描、编译调用、产物生成等编译阶段逻辑。
- 不要求插件作者为每条动态路由重复编写内存分配、桥接信封装配和二进制编解码样板代码。

## Decisions

### 公开路径与匹配模型

**决策**：动态插件只暴露固定前缀公开路径，宿主仅对`/api/v1/extensions/*dynamicPath`进入动态分发链路。

#### 路径结构

```text
/api/v1/extensions/{pluginId}/{plugin-internal-path}
```

其中：

- `{pluginId}`是动态插件唯一标识；
- `{plugin-internal-path}`是插件在`g.Meta.path`中声明的内部业务路径；
- 插件内部路径禁止再次声明宿主公开前缀，避免路径空间污染。

#### 匹配流程

1. 宿主先从公开路径中解析出`pluginId`与内部路径。
2. 宿主只在该插件当前启用版本的`manifest.Routes`中匹配。
3. 匹配维度为`method + internalPath`。
4. 路径支持静态段优先与`/path/{id}`形式的参数段匹配。

**理由**

- 固定前缀把性能影响限制在极小范围内，普通宿主路由不受影响。
- 宿主不需要对全部请求做动态插件解析，只需对固定命名空间请求进入单插件匹配。
- 公开路径与插件内部路径解耦后，插件实现更简单，宿主治理边界也更稳定。

### 路由治理元数据统一收敛到`g.Meta`

**决策**：动态路由治理元数据不再拆出独立配置模块，而是直接复用`api`层请求结构体中的`g.Meta`。

当前支持的最小治理字段如下：

| 字段 | 来源 | 说明 |
| --- | --- | --- |
| `path` | `g.Meta.path` | 插件内部路径 |
| `method` | `g.Meta.method` | 标准`HTTP`方法 |
| `tags` | `g.Meta.tags` | 文档标签 |
| `summary` | `g.Meta.summary` | 路由摘要 |
| `description` | `g.Meta.dc` | 路由详细描述 |
| `access` | `g.Meta.access` | 访问级别，仅支持`public`、`login` |
| `permission` | `g.Meta.permission` | 权限点，格式固定为`{pluginId}:{resource}:{action}` |
| `operLog` | `g.Meta.operLog` | 预留治理字段，当前只做校验与保留 |
| `requestType` | 请求结构体名 | 用于生成稳定`OpenAPI operationId` |

#### 合同校验

宿主与构建器共享同一套校验规则：

- `access`缺省时按`login`归一化；
- `public`路由不得声明`permission`；
- `public`路由不得声明`operLog`；
- `permission`必须匹配`{pluginId}:{resource}:{action}`格式；
- `permission`必须以当前插件`id`为前缀；
- `operLog`当前只允许预定义数值；
- 同一插件内`method + path`不得重复。

**理由**

- 元数据集中在`g.Meta`后，动态路由声明、文档、治理与源码位置一致，理解成本最低。
- 不需要再维护第二套外置路由治理清单，更适合`AI agent`自动生成。

### 构建阶段直接提取并嵌入运行时产物

**决策**：动态插件的编译阶段逻辑统一由`hack/build-wasm`承担。只有独立构建器负责扫描`backend/api/**/*.go`、提取`g.Meta`路由合同、编译可执行`Wasm` guest，并嵌入运行时产物自定义节；`lina-core`宿主只负责装载、校验与执行产物，不再保留任何编译阶段实现或调用逻辑。

#### 运行时产物新增区段

| 区段名 | 作用 |
| --- | --- |
| `lina.plugin.backend.routes` | 保存动态路由合同 |
| `lina.plugin.backend.bridge` | 保存动态路由桥接 ABI 合同 |

宿主装载产物时恢复出：

- `manifest.Routes`
- `manifest.RuntimeArtifact.BridgeSpec`

#### 运行时组件边界

- `apps/lina-core/internal/service/plugin`是运行时业务组件，只负责动态插件产物装载、运行时合同校验、生命周期治理和请求执行。
- 源码扫描、`g.Meta`静态提取、`go build`调用、自定义区段写入和样例产物生成都属于编译阶段逻辑，必须统一收敛到`hack/build-wasm`或`hack/`下的独立脚本／工具。
- 宿主运行时可以消费构建阶段已经嵌入产物的合同与区段，但不得反向调用构建器，也不得为了请求执行去读取插件源码目录。
- 如果需要宿主与构建器复用合同结构、校验规则或产物区段常量，应把这类无副作用的公共模型放入`apps/lina-core/pkg`，而不是把构建流程放回`internal/service/plugin`。

**理由**

- 请求链路不再依赖源码扫描；
- 路由合同天然与`active release`绑定；
- 启用、禁用、卸载、回滚后，宿主读取的都是当前真实生效版本；
- 编译职责集中到单一构建器后，可以避免宿主与工具链双份实现带来的产物偏差与维护成本。

### 宿主治理前置，桥接执行后置

**决策**：宿主先完成治理，再通过统一执行器将请求转入动态插件运行时。

#### 宿主统一中间件注册方式

动态插件固定前缀路由虽然仍由宿主掌握治理权，但它们在宿主侧的接入方式要与普通宿主路由保持一致：统一挂在`RouterGroup`下，通过`group.Middleware(...)`声明宿主中间件链，再由最终处理器进入动态插件执行器。

其中：

- 通用中间件（如`NeverDoneCtx`、`HandlerResponse`、`CORS`、`Ctx`）继续复用宿主现有注册链；
- 动态插件特有的治理阶段（如固定前缀命中后的路由匹配、按路由`access`决定的鉴权、权限校验）也以宿主中间件形式挂载，而不是在独立分发入口中手工串联；
- 动态插件仍**不能自由选择或拼装宿主中间件**，中间件编排权完全由宿主掌握。

**理由**

- 宿主所有`HTTP`入口都遵循同一套路由与中间件注册方式，可降低理解和维护成本；
- 动态路由治理逻辑仍然保留在宿主，但从“手工串联流程”演进为“宿主中间件链编排”，更符合现有后端工程结构；
- 通用中间件的改动可以自然覆盖动态插件固定前缀入口，减少后续遗漏。

#### 宿主治理职责

- 解析固定前缀公开路径；
- 校验插件存在、已安装、已启用且存在`active release`；
- 匹配动态路由；
- 对`login`路由执行令牌校验、会话校验与业务上下文注入；
- 对声明了`permission`的路由复用宿主权限体系做权限校验。

#### `access`语义

| 值 | 宿主行为 |
| --- | --- |
| `login` | 解析`Bearer Token`、校验会话、注入身份快照，必要时校验权限 |
| `public` | 不解析令牌、不注入用户上下文、不允许声明`permission` |

#### 关于`DTO`

宿主与动态插件之间**不再为每个路由额外定义一层业务 DTO**。真正稳定的中间层只有一份统一桥接信封：

- `DynamicRouteBridgeRequestEnvelopeV1`
- `DynamicRouteBridgeResponseEnvelopeV1`

其中宿主只传递：

- 路由匹配快照；
- 请求快照；
- 身份快照。

具体业务参数解析仍由插件在其自身运行时内处理。这样既避免宿主为每条动态路由做额外 DTO 适配，也保留了插件业务灵活性。

信封结构是宿主与插件共享的逻辑模型；跨`Wasm`内存传递时必须使用版本化二进制编解码，不能把信封再编码成`json`或纯文本负载。

**理由**

- 动态插件只是业务扩展，不需要宿主重建一套完整接口层。
- 宿主只治理“如何进入插件”，插件自己决定“如何消费快照并执行业务”。

### 桥接`DTO`采用高效二进制编解码

**决策**：`v1`桥接协议的请求／响应信封采用二进制`protobuf`线格式承载，`BridgeSpec.requestCodec`与`BridgeSpec.responseCodec`只允许声明`protobuf`，宿主拒绝可执行 bridge 使用`json`、`text`、`plain`等文本类编解码协议。

#### 编解码边界

- 二进制编解码只作用于`DynamicRouteBridgeRequestEnvelopeV1`与`DynamicRouteBridgeResponseEnvelopeV1`这类宿主到插件的桥接`DTO`。
- 客户端原始请求体作为`[]byte`放入请求快照透传；如果业务接口本身接收`JSON`请求体，也由插件业务处理原始请求体，宿主不得把整个桥接信封降级为`JSON`。
- `protobuf`消息定义、编解码器、请求／响应信封装配、错误响应辅助和 guest 侧处理器适配器统一抽象到`apps/lina-core/pkg/pluginbridge`公共组件。
- 插件运行时只需复用公共组件注册业务处理函数，不应重复编写底层内存读写、信封编解码和响应打包样板。

**理由**

- 请求／响应快照会出现在每次动态路由调用链路中，使用二进制协议能降低序列化体积和`CPU`开销。
- 固定`protobuf`线格式后，宿主与动态插件可以共享稳定`schema`，避免`map[string]interface{}`或文本协议带来的运行时反射与歧义。
- 公共组件把高风险`ABI`细节封装起来，插件作者主要关注业务处理逻辑，更适合`AI agent`生成动态插件。

### 统一请求／响应快照模型

**决策**：宿主与动态插件之间统一通过稳定`v1`桥接信封交换请求与响应快照。

#### 请求快照内容

- 方法、公开路径、内部路径；
- 原始路径、查询串；
- `Host`、`Scheme`、`RemoteAddr`、`ClientIP`；
- 已脱敏请求头；
- `Cookie`集合；
- 请求体字节；
- 可选身份快照。

#### 安全约束

- 宿主会剥离`Authorization`头；
- 动态插件不能直接使用宿主原始令牌；
- 认证结果通过身份快照显式注入；
- 动态插件无法直接获得宿主`ghttp.Request`或宿主内部服务对象。

#### 响应快照内容

- `StatusCode`
- `ContentType`
- `Headers`
- `Body`

宿主只负责把响应快照写回客户端，不参与插件内部业务序列化。

**理由**

- 统一快照模型能稳定桥接 ABI，避免未来反复改动宿主分发入口。
- 插件作者只需理解一份请求／响应合同，适合受限动态扩展模型。

### 真实`Wasm bridge`执行模型

**决策**：当动态插件包含`backend/runtime/wasm`运行时包并成功构建时，宿主优先通过真实`Wasm bridge`执行动态路由。

#### 插件侧约定

动态插件可以在如下目录提供受限运行时入口：

```text
backend/runtime/wasm
```

构建器使用：

```bash
go build -buildmode=c-shared -o runtime-plugin.wasm ./backend/runtime/wasm
```

并设置：

```bash
GOOS=wasip1
GOARCH=wasm
```

若该目录不存在，则仍输出默认桥接合同，但`routeExecution=false`，宿主保持`501`回退行为。

#### 桥接 ABI

当前固定字段如下：

| 字段 | 值 |
| --- | --- |
| `protocolVersion` | `v1` |
| `requestCodec` | `protobuf` |
| `responseCodec` | `protobuf` |
| `initializeEntrypoint` | `_initialize` |
| `requestAllocExport` | `lina_dynamic_route_alloc` |
| `executeEntrypoint` | `lina_dynamic_route_execute` |
| `routeExecution` | 构建期根据是否存在可执行运行时决定 |

#### 宿主执行流程

1. 通过`wazero`加载当前`active release`的`Wasm`产物；
2. 跳过自动`start function`执行；
3. 如存在`_initialize`导出，则显式调用一次；
4. 调用`lina_dynamic_route_alloc(size)`向 guest 申请请求缓冲区；
5. 宿主将序列化后的请求信封写入 guest 内存；
6. 调用`lina_dynamic_route_execute(ptr, len)`执行动态路由；
7. 从返回的打包`ptr/len`中读取响应负载；
8. 反序列化为统一响应快照并回写客户端。

#### 宿主回退策略

- 若运行时未声明可执行 bridge，则返回`501`占位响应；
- 若桥接导出缺失、内存写入失败、执行失败或响应解码失败，则返回宿主`500`错误。

**理由**

- 不再追求在`WASI`中完整复刻宿主`GoFrame`路由栈，而是用受限 ABI 解决真实执行问题。
- `AI agent`只需生成一个简单的运行时包和导出函数，实现成本更低。
- 宿主仍牢牢掌握鉴权、权限与治理边界，整体安全性更高。

### 动态插件本地中间件位于 guest 运行时内部

**决策**：动态插件如需中间件，只能在其自身`Wasm bridge`运行时内部组织，不与宿主治理中间件混合。

当前样例插件在`backend/runtime/wasm`内实现了：

- 运行时恢复中间件；
- 响应头补充中间件；
- 路由分派与业务调用。

这证明动态插件仍可以有轻量的局部执行链，但这条链只在 guest 内部生效，宿主无需也不会暴露中间件自由编排能力。

**理由**

- 保持宿主治理职责与插件业务职责边界清晰；
- 兼顾“足够灵活”和“不能太强”的设计目标。

### 权限继续复用`sys_menu.perms`

**决策**：动态路由声明的`permission`自动物化为隐藏菜单权限项，继续复用宿主现有角色授权体系。

#### 同步策略

- 插件启用时生成对应隐藏权限节点；
- 插件禁用、卸载或激活版本变更时同步更新；
- 默认管理员角色自动拥有这些权限；
- 动态路由权限目录挂载在插件专属菜单空间下。

**理由**

- 不引入第二套动态权限存储模型；
- 后端权限校验与前端能力控制可继续复用现有基础设施。

### `OpenAPI`按运行时状态投影

**决策**：系统接口文档自动投影当前已启用动态插件的公开接口，并根据运行时是否可执行展示不同响应语义。

#### 文档投影规则

- 公开路径显示为`/api/v1/extensions/{pluginId}/...`；
- 标签、摘要、描述来自动态路由合同；
- `login`路由声明`BearerAuth`；
- 可执行运行时展示`200`与`500`响应语义；
- 未声明可执行 bridge 的运行时仅展示`501`占位说明。

**理由**

- 文档展示的是用户实际可访问的公开接口，而不是插件内部私有路径。
- 接口文档能准确反映当前激活版本是否具备真实执行能力。

## Risks / Trade-offs

### 运行时桥接能力受限

**风险**

动态插件不能像源码插件那样直接复用宿主完整运行时能力，某些复杂场景会受限。

**缓解措施**

- 明确动态插件只承载受限业务扩展；
- 复杂治理、复杂宿主集成继续交给源码插件；
- 动态插件只暴露最小桥接合同，减少运行时不确定性。

### `Wasm bridge`调试成本高于普通宿主代码

**风险**

`Wasm`内存读写、导出函数、编解码协议和构建链都比普通`Go`服务更难调试。

**缓解措施**

- 固定`v1 + protobuf`二进制桥接协议；
- 通过`apps/lina-core/pkg/pluginbridge`封装信封`schema`、编解码器与 guest 侧`ABI`辅助逻辑；
- 将桥接 ABI 明确写入产物元数据；
- 通过单一独立构建器统一产物生成规则，减少环境差异；
- 用样例插件和单元测试覆盖典型桥接路径。

### `operLog`尚未进入真实审计落地

**风险**

当前`operLog`已成为合同字段，但宿主尚未基于动态执行结果落地真实操作日志。

**缓解措施**

- 先固定字段语义与校验规则；
- 后续若需要补齐审计，可在现有快照与桥接边界上继续扩展，而无需重做路径分发与执行框架。

## Validation

本轮实现需要覆盖以下验证点：

- 动态路由合同提取、校验和产物嵌入；
- 宿主装载后`manifest.Routes`恢复；
- 固定前缀路由分发与路径参数匹配；
- 登录校验、权限校验与身份快照注入；
- 动态权限菜单物化与生命周期同步；
- `OpenAPI`运行时感知投影；
- 宿主真实`Wasm bridge`执行；
- 宿主拒绝`json`或纯文本桥接`DTO`编解码合同；
- 公共`pluginbridge`组件能被宿主执行器与样例动态插件复用；
- 样例动态插件通过受限 bridge 返回真实业务响应；
- `E2E`验证插件生命周期与动态路由真实执行闭环。

## Host Functions（宿主回调能力）

### Context

单向`Wasm bridge`协议只解决了"宿主把请求发给 Guest、Guest 返回响应"的问题，但 Guest 无法回调宿主，导致无法执行日志记录、状态持久化、数据库读写等操作，业务场景极度受限。本节为`Wasm` Guest 增加 **Host Functions** 能力，让插件能安全地调用宿主提供的受控服务，从而支撑真实业务扩展场景。

### 架构概览

```text
Guest (WASM)                              Host (Go/wazero)
  |                                         |
  |-- lina_env.host_call(op, ptr, len) --->|  Guest invokes host
  |                                         |  1. Read request from Guest memory
  |                                         |  2. Dispatch by opcode
  |                                         |  3. Execute (log/DB/state)
  |<-- lina_host_call_alloc(size) ---------|  4. Callback Guest to alloc buffer
  |-- return ptr ------------------------->|  5. Write response to Guest memory
  |<-- return packed(ptr, len) ------------|  6. Return response pointer+length
  |                                         |
```

**关键设计决策**：

- **单一入口函数**：所有宿主调用通过`lina_env.host_call(opcode, reqPtr, reqLen) → uint64`统一分发，新增能力不改变`Wasm`导入签名。
- **独立响应缓冲区**：Guest 导出`lina_host_call_alloc`，使用独立的`guestHostCallResponseBuffer`，避免与主请求缓冲区冲突。
- **能力声明+运行时校验**：插件在`plugin.yaml`中声明所需能力，宿主在每次 Host Call 时校验。
- **复用`wazero`再入特性**：Host Function 回调中调用 Guest 导出函数（`lina_host_call_alloc`），`wazero`原生支持。

### 能力清单（Phase 1）

| 能力标识 | Opcode | 说明 |
|----------|--------|------|
| `host:log` | `0x0001` | 通过宿主`logger`组件输出结构化日志 |
| `host:state` | `0x0101`–`0x0103` | 插件隔离的键值状态存储（`sys_plugin_state`表） |
| `host:db:query` | `0x0201` | 只读`SQL`查询（仅`SELECT`，可访问所有表） |
| `host:db:execute` | `0x0202` | 写入`SQL`（`INSERT`/`UPDATE`/`DELETE`，禁止`DDL`） |

### 能力声明模型

插件在`plugin.yaml`中声明所需能力：

```yaml
capabilities:
  - host:log
  - host:state
  - host:db:query
```

构建器在构建阶段校验能力字符串合法性，并嵌入`Wasm`自定义段`lina.plugin.backend.capabilities`。宿主装载产物时恢复出`manifest.HostCapabilities`，在每次 Host Call 分发时按 opcode 映射到能力标识做`O(1)`校验。未声明的能力调用返回`capability_denied`状态。

### Host Call 协议

#### 请求流程

1. Guest 调用`lina_env.host_call(opcode, reqPtr, reqLen)`。
2. Host Function 从 Guest 内存读取`reqLen`字节的请求负载。
3. Host 按 opcode 查找能力映射，校验当前插件是否声明了对应能力。
4. Host 分发到对应处理器执行操作。
5. Host 将响应序列化后调用 Guest 导出`lina_host_call_alloc(respLen)`分配缓冲区。
6. Host 将响应写入 Guest 内存，返回打包的`packed(respPtr, respLen)`。

#### 响应信封

所有 Host Call 响应统一使用`HostCallResponseEnvelope`：

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | `uint32` | 状态码：`0=success`、`1=capability_denied`、`2=invalid_request`、`3=internal_error` |
| `payload` | `bytes` | 序列化的响应负载（按 opcode 不同有不同结构） |

#### 编码方式

请求与响应负载使用手写`protowire`编码（与现有桥接`codec`保持一致），每种 opcode 有独立的请求/响应消息结构。

### 各能力实现细节

#### `host:log`（结构化日志）

Guest 通过`HostLog(level, message, fields)`发送日志请求，宿主使用项目封装的`logger`组件输出，自动附加`[plugin:{pluginID}]`前缀标识日志来源。

#### `host:state`（插件隔离状态存储）

基于`sys_plugin_state`表实现键值存储，所有操作自动按`pluginID`隔离：

- `StateGet(key)`：查询单个状态值
- `StateSet(key, value)`：写入或更新状态值（`INSERT ... ON DUPLICATE KEY UPDATE`）
- `StateDelete(key)`：删除单个状态值

表结构：

```sql
CREATE TABLE IF NOT EXISTS sys_plugin_state (
    id           INT PRIMARY KEY AUTO_INCREMENT,
    plugin_id    VARCHAR(64)   NOT NULL DEFAULT '',
    state_key    VARCHAR(255)  NOT NULL DEFAULT '',
    state_value  LONGTEXT,
    created_at   DATETIME,
    updated_at   DATETIME,
    UNIQUE KEY uk_plugin_state (plugin_id, state_key)
);
```

#### `host:db:query`（只读数据库查询）

- 仅允许`SELECT`语句（前缀校验）
- 黑名单拒绝`DDL`关键词（`DROP`、`ALTER`、`CREATE`、`TRUNCATE`、`GRANT`、`REVOKE`）
- `maxRows`上限 1000 行
- 返回列名、行数据和行数

#### `host:db:execute`（数据库写入）

- 仅允许`INSERT`、`UPDATE`、`DELETE`、`REPLACE`语句
- 黑名单拒绝`DDL`关键词和`SELECT`语句
- 返回受影响行数和最后插入`ID`

**DB 访问模式**：开放访问——插件可查询/操作所有表。安全性由管理员决定是否授予该能力来控制。`SQL`语句前缀校验 +`DDL`关键词黑名单防护。

### 安全边界

- 每次 Host Call 都经过能力校验，未声明的能力调用立即拒绝。
- 状态存储按`pluginID`强隔离，插件无法访问其他插件的状态。
- 数据库访问受`SQL`前缀校验和`DDL`黑名单双重防护。
- Host Function 注册在`wazero Runtime`级别（非模块实例级别），与模块缓存机制兼容。
- 插件运行时上下文通过`context.Context`传递，每个请求独立的`hostCallContext`携带`pluginID`、`capabilities`和`service`引用。

### Guest SDK

`pluginbridge`包提供面向 Guest 的高级`API`封装（`//go:build wasip1`）：

- `HostLog(level, message, fields) error`
- `HostStateGet(key) (string, bool, error)`
- `HostStateSet(key, value) error`
- `HostStateDelete(key) error`
- `HostDBQuery(sql, args, maxRows) (*HostDBQueryResult, error)`
- `HostDBExecute(sql, args) (int64, int64, error)`
- `HostStateGetInt(key) (int, bool, error)` / `HostStateSetInt(key, value) error`

内部通过`//go:wasmimport lina_env host_call`导入宿主函数，`invokeHostCall(opcode, reqBytes)`处理指针传递和响应解码。

### Phase 2 预留

| 能力 | 说明 |
|------|------|
| `host:http` | 出站`HTTP`请求，域名白名单控制 |
| `host:event` | 系统事件发布/订阅 |
| `host:config` | 读取宿主配置项 |
| 管理`UI` | 能力审批/授权界面 |
