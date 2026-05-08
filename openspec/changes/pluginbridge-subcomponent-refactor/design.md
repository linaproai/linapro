## Context

`pkg/pluginbridge` 是动态插件 Wasm bridge 的共享基础设施。它同时被宿主运行时、WASM host function、`plugindb`、动态插件样例和 guest 代码使用。当前根包目录下约 41 个生产 Go 文件，包含多个不同层级的职责：

- 稳定 ABI 与 manifest 合约：`BridgeSpec`、`RouteContract`、`CronContract`、`HostServiceSpec` 等。
- bridge envelope 编解码：request、response、route、identity、HTTP request snapshot 和 protobuf wire 工具。
- WASM artifact 辅助：custom section 常量和读取能力。
- host call 协议：opcode、host_call envelope、状态码。
- host service 协议：runtime、storage、network、data、cache、lock、config、notify、cron 等 payload 类型与 codec。
- guest SDK：guest runtime、controller dispatcher、context、BindJSON、WriteJSON、host service client helper。

这些能力都属于插件桥接体系，但使用者不同。把它们放在同一个根包会让“插件作者应使用的 API”和“宿主运行时维护者关注的协议细节”混在一起。项目是新项目，不需要为了历史目录形态保留复杂结构，但现有动态插件样例和宿主内部调用已经依赖 `pluginbridge` 根包 API，因此实现阶段需要以兼容 facade 作为迁移缓冲。

## Goals / Non-Goals

**Goals:**

- 将 `pkg/pluginbridge` 拆分为类似 `pkg/pluginservice` 的公开子组件包，使目录结构直接表达职责边界。
- 让根包生产源码收敛为薄 facade，目标是根目录保留 1 到 3 个生产源码文件。
- 保持现有根包 API 兼容，不要求一次性修改所有外部插件代码。
- 让宿主内部代码逐步使用更精确的子组件 import，降低未来维护时的阅读范围。
- 将低层实现细节下沉到子组件内部，避免继续在根包暴露大量实现文件。
- 通过测试证明重构不改变 ABI 常量、序列化字节、WASM section 读取、host service payload codec 和 guest helper 行为。

**Non-Goals:**

- 不改变动态插件 Wasm bridge 协议、host call 入口、host service 方法名或 payload 字段编号。
- 不删除 `pluginbridge.Data()` 等已存在的兼容 helper；推荐路径变化需另行规划。
- 不引入新的外部依赖或代码生成流程。
- 不修改数据库 schema、SQL、REST API、前端页面或插件清单业务语义。
- 不在本次变更中归档或重写既有插件运行时规范的全部语言。

## Decisions

### 决策一：采用公开子组件包，而不是只使用根包 internal

目标结构：

```text
pkg/pluginbridge/
  pluginbridge.go      # 根包 facade：alias + wrapper
  contract/            # ABI、route、cron、execution source 等稳定合约
  codec/               # bridge request/response envelope 编解码
  artifact/            # Wasm section 常量、custom section 读取、runtime metadata
  hostcall/            # host_call opcode、通用 host call envelope 和状态码
  hostservice/         # host service spec、capability 推导、payload codec
  guest/               # guest runtime、controller dispatcher、BindJSON、host service clients
```

理由：

- `pluginservice` 已采用按能力拆公开包的模式，项目内使用者容易理解。
- `pluginbridge` 的使用者确实分层：插件作者主要用 `guest`，宿主 runtime 主要用 `codec`、`artifact`、`hostcall`、`hostservice`，manifest 校验主要用 `contract` 和 `hostservice`。
- 公开子组件能让文档、测试和 import 都围绕职责组织，收益明显高于仅移动两三个实现文件到 `internal`。

替代方案：只把 `pluginbridge_codec_wire.go` 和 `pluginbridge_wasm_section.go` 搬进根包 internal。该方案风险低，但根目录文件数几乎不变，无法解决用户理解复杂度问题，因此不作为主方案。

### 决策二：保留根包 facade 作为兼容层

根包 `pluginbridge` 继续暴露现有常量、类型和函数。实现方式优先使用：

- `type X = contract.X`
- `const X = contract.X`
- `func EncodeRequestEnvelope(...) { return codec.EncodeRequestEnvelope(...) }`
- `func Runtime() guest.RuntimeHostService { return guest.Runtime() }`

理由：

- 宿主内部和动态插件样例已有大量 `pluginbridge.X` 调用。
- 动态插件 guest 代码可能被用户复制使用，突然要求 import 新子包会造成不必要迁移成本。
- facade 允许实施阶段先拆包，再逐步把宿主内部调用迁到更精确的子包。

替代方案：删除根包并强制所有调用方迁移到子包。该方案更干净，但变更面和风险过大，不适合本次结构治理。

### 决策三：固定依赖方向，避免包循环

依赖方向约束：

```text
contract
  ↑
codec ──→ internal/wire
  ↑
artifact ──→ internal/wasmsection
  ↑
hostservice ──→ contract, codec/internal wire
  ↑
hostcall ──→ hostservice
  ↑
guest ──→ contract, codec, hostcall, hostservice
  ↑
pluginbridge facade ──→ all subcomponents
```

具体实现可以根据代码耦合微调，但必须保证底层包不 import 根包 facade。任何子组件的 `internal` 包都只能服务该子组件或其父路径下的 sibling 包，不承载跨领域的兜底工具。

理由：

- Go 中子包如果反向 import 根包，再由根包 facade import 子包，会形成循环依赖。
- 依赖方向明确后，迁移可以按叶子到上层逐步推进。

### 决策四：优先迁移宿主内部 import，保留插件侧兼容路径

实施后，宿主内部代码应优先使用精确子组件：

- runtime artifact 解析使用 `pluginbridge/artifact`
- Wasm 执行器使用 `pluginbridge/codec`、`pluginbridge/hostcall`、`pluginbridge/hostservice`
- manifest 和 route 合约校验使用 `pluginbridge/contract`、`pluginbridge/hostservice`

动态插件样例可以在一轮内迁移到 `pluginbridge/guest`，但根包 facade 仍必须覆盖旧调用路径。

理由：

- 宿主内部 import 是项目可控代码，应先体现新边界。
- 插件侧兼容路径可降低用户迁移压力。

### 决策五：验证以协议不变为核心，不以文件数量为唯一成功标准

必须覆盖：

- `EncodeRequestEnvelope` / `DecodeRequestEnvelope` 字节级 round trip 不变。
- 各 host service payload `Marshal` / `Unmarshal` round trip 不变。
- WASM custom section 读取错误边界不变。
- `HostCallResponseEnvelope` 和 structured host service envelope 不变。
- guest runtime、typed controller dispatcher、BindJSON/WriteJSON 行为不变。
- 根包 facade 与子组件直接调用结果一致。

理由：

- `pluginbridge` 是 ABI 包，真正风险在协议兼容性，不在代码移动本身。
- 文件数下降只是可维护性结果，不能替代行为验证。

## Risks / Trade-offs

- [Risk] 子组件拆分产生 import cycle。
  Mitigation：先移动 contract 类型和纯实现工具，再迁移上层包；根包 facade 最后接入，禁止子组件 import 根包。

- [Risk] alias/wrapper 遗漏导致现有调用方编译失败。
  Mitigation：实施过程中运行 `go test ./pkg/pluginbridge/... ./internal/service/plugin/... ./pkg/plugindb/...`，并构建动态插件样例。

- [Risk] 序列化字段编号或默认值在拆分中被误改。
  Mitigation：保留并迁移现有 codec round trip 测试，新增 facade 与子组件结果一致性测试。

- [Risk] 子组件包过细导致用户 import 选择困难。
  Mitigation：只建立 5 到 6 个稳定子组件；根包 facade 和 `guest` 包作为插件作者主要入口，文档中说明推荐 import。

- [Risk] 测试文件移动后覆盖范围下降。
  Mitigation：测试随职责迁移到对应子包，同时根包保留 facade 兼容测试。

## Migration Plan

1. 创建子组件包骨架和文件顶部包注释。
2. 迁移 `contract`、`artifact`、`codec` 等低依赖能力，并保持根包 facade 编译通过。
3. 迁移 `hostservice` 和 `hostcall`，保留现有序列化测试并补充 facade 一致性测试。
4. 迁移 `guest` SDK，更新动态插件样例或补充兼容测试。
5. 更新宿主内部 import 到子组件包，保留根包对外兼容。
6. 运行相关 Go 测试、wasip1/wasm 构建和 OpenSpec 校验。

回滚策略：由于不涉及数据和运行时配置，若拆分导致构建或协议测试失败，可回退到根包实现文件；facade 层保留使回滚不需要迁移用户代码。

## Open Questions

- 是否在本轮同步新增 `README.md` / `README.zh_CN.md` 说明各子组件用途；如果新增目录说明文档，必须按项目规范同步双语维护。
- `plugindb` 的推荐入口已经独立存在，是否需要在本轮进一步弱化 `pluginbridge.Data()` 文档描述，还是只保持兼容不改推荐路径。
