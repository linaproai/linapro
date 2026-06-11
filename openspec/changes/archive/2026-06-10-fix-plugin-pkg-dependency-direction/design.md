# 设计：修正 pkg/plugin 依赖方向

## Context

`apps/lina-core/pkg/plugin`包含三个公开顶层组件：`capability`（插件消费宿主能力的稳定目录与`*cap`契约）、`pluginhost`（源码插件贡献入口）、`pluginbridge`（动态插件 ABI 与 guest SDK）。基线规范`plugin-package-boundary-governance`已要求三者职责分离，但未约束依赖方向，当前存在两处实际反向依赖：

1. `capability/recordstore/recordstore_exec_wasip1.go`直接 import `pluginbridge/protocol`并内嵌 host-service 调用编码。`recordstore`虽位于`capability/`下，但只被`pluginbridge.Services.RecordStore()`暴露、不在`capability.Services`目录中，本质是动态插件专属 guest SDK。
2. `pluginhost/pluginhost_source_plugin_manifest.go`与`pluginhost/internal/manifestview`以`pluginbridge/contract.ManifestSnapshotV1`作为源码插件升级回调的快照契约，并通过`ManifestSnapshot.Values()`将该 bridge ABI 类型直接暴露给源码插件作者。

`ManifestSnapshotV1`本身是纯字段结构体（16 个 string/int/bool 字段，仅 JSON tag），无任何 bridge ABI 依赖，具备迁入中立原语包的条件。`capability/capmodel`的既有定位即"shared plugin-domain capability primitives"，且`pluginhost`与`pluginbridge`均已依赖`capability`，迁入不引入新依赖方向。

消费方现状（非测试代码）：

- `capability/recordstore`的外部消费方：宿主侧`internal/service/plugin/internal/datahost`（2 个文件，消费`QueryPlan`等计划契约）、`pluginbridge`根包（directory 与公开契约）、`pluginbridge/internal/hostservice`（descriptor 测试）、动态插件`apps/lina-plugins/linapro-demo-dynamic`（1 个文件）。
- `pluginhost`对`pluginbridge/contract`的依赖点：仅`pluginhost_source_plugin_manifest.go`与`internal/manifestview`两处（另有 1 个测试文件）。

项目无历史负担，不需要保留旧 import 路径的兼容入口。

## Goals / Non-Goals

**Goals:**

- `capability/**`非测试代码不再 import `pluginbridge`或`pluginhost`的任何子包，使 capability 成为真正的最底层契约层。
- `pluginhost/**`非测试代码不再 import `pluginbridge`的任何子包，源码插件公开 API 不暴露动态插件 ABI 类型。
- 依赖方向通过可重复执行的治理测试固化，防止回归。
- `Runtime`/`Network`/`RecordStore`作为动态插件专属能力的设计决定在双语 README 中显式记录。
- 所有迁移保持运行时行为、wire 格式、协议语义零变更。

**Non-Goals:**

- 不拆分`tenantcap`/`orgcap`等包中的宿主专用接缝（gdb/ghttp 泄漏问题，后续独立变更）。
- 不改造`aitext`/`tenantcap`/`orgcap`的包级 provider 注册单例（后续独立变更）。
- 不统一 guest 侧双轨传输模式、不引入 hostservice descriptor 代码生成（后续独立变更）。
- 不为源码插件补齐 RecordStore/Runtime/Network 等价能力。

## Decisions

### D1：recordstore 整体迁移到 pluginbridge 子包，而非接缝倒置

**选择**：将`capability/recordstore`（含`internal/plan`）整体迁移为`pluginbridge/recordstore`，迁移后其对`pluginbridge/protocol`的 import 成为同层合法依赖。

**备选与否决**：曾考虑保留包位置、在 recordstore 中定义`Executor`接口由 pluginbridge 注入实现（依赖倒置）。否决理由：recordstore 已有`HostServiceInvoker`函数类型注入接缝（`OpenWithHostServiceInvoker`），但 wasip1 执行文件中对`protocol`的 DTO 与编码依赖是结构性的——倒置需要再造一层与`protocol`等价的中立 DTO，纯属为分层而分层的重复定义，违背架构规则中"禁止为预留未知变化引入抽象层"的要求。该组件事实上只服务动态插件，迁移到 pluginbridge 是让包位置承认事实，复杂度更低。

**影响**：基线规范`plugin-package-boundary-governance`中"record store SDK 位于`pkg/plugin/capability/recordstore`"的既有 Requirement 需要通过 delta spec 修改为新位置；`capability`下`*cap`命名豁免列表同步移除`recordstore`。

### D2：ManifestSnapshotV1 迁入 capmodel，contract 保留类型别名

**选择**：类型定义移到`capability/capmodel`（命名为`ManifestSnapshot`，capmodel 内不带 V1 后缀的 wire 版本语义由别名层承载）；`pluginbridge/contract`保留`type ManifestSnapshotV1 = capmodel.ManifestSnapshot`别名；`pluginhost`与`manifestview`改为直接依赖 capmodel。

**备选与否决**：

- 全量替换不留别名：可行但破坏`protocol`/`contract`既有的 facade 别名转发惯例（`protocol_contract.go`整文件就是别名层），且动态插件生命周期协议侧的命名（`ManifestSnapshotV1`）有 wire 版本语义，保留别名让协议命名留在协议包。
- 迁入独立新包（如`capability/manifestmodel`）：否决，单类型不足以支撑新包，capmodel 定位完全匹配。

**影响**：JSON tag 随类型定义迁移，`LifecycleRequest.fromManifest/toManifest`序列化结果不变；基线规范`plugin-runtime-loading`中"源码插件和动态插件复用同一 manifest snapshot 契约"的 Requirement 语义不变，不需要 delta。

### D3：import 边界治理以 Go 单元测试形态实现

**选择**：在`pkg/plugin`下新增治理测试（如`plugin_boundary_test.go`），用`go/parser`或`golang.org/x/tools/go/packages`扫描`capability/**`、`pluginhost/**`非测试源文件的 import 声明，断言：`capability`不含`lina-core/pkg/plugin/pluginbridge`与`lina-core/pkg/plugin/pluginhost`前缀；`pluginhost`不含`lina-core/pkg/plugin/pluginbridge`前缀。测试文件（`_test.go`）豁免——`capability_test.go`合法地用 bridge 做集成验证。

**备选与否决**：shell 脚本 + grep 形态（需按`.agents/rules/dev-tooling.md`处理跨平台，且不进入`go test`门禁，易被遗忘）；引入 import-linter 类外部工具（新增工具链依赖，违背最小复杂度）。Go 测试形态随`make test`自动执行、跨平台、无新依赖。

### D4：README 记录动态插件专属能力的判定标准

**选择**：在双语 README 的能力目录章节后补一小节，说明`Runtime`、`Network`、`RecordStore`仅存在于`pluginbridge.Services`的原因：源码插件对 Runtime/Network 有宿主原生等价物（日志组件、HTTP 客户端均可直接使用），RecordStore 是对 host-service data 协议的 guest 侧封装、源码插件直接使用自有 DAO。该小节同时作为后续新增能力时"进 capability 还是进 pluginbridge"的判定参照。

## Risks / Trade-offs

- [迁移遗漏 import 导致编译失败] → 消费方已全量枚举（datahost 2 文件、demo 插件 1 文件、pluginbridge 根包、若干测试）；以`go build ./...`与全量`go vet`作为门禁，遗漏即编译期暴露。
- [demo 动态插件是独立 module，宿主编译门禁不覆盖] → 任务中显式包含`apps/lina-plugins/linapro-demo-dynamic`的构建验证（含 wasip1 目标构建，因 recordstore 的 wasip1 文件是本次迁移核心）。
- [治理测试对包路径前缀硬编码，未来包重命名需同步] → 可接受；测试失败信息明确指向边界规则，重命名场景下属于预期提醒而非误报。
- [capmodel 吸收 manifest snapshot 后定位轻微扩张] → capmodel 既有定位即跨域共享原语，manifest 快照与`CapabilityActor`等同级，不构成能力服务聚合点，符合基线规范对公共原语包的约束。
- [`pluginhost`测试文件仍 import pluginbridge] → 治理测试豁免`_test.go`；`pluginhost_source_plugin_test.go`若仅因 manifest 类型 import contract，迁移后顺手改为 capmodel，消除豁免依赖面。

## Migration Plan

单仓单次变更完成，无部署迁移。实施顺序：先 D2（capmodel 类型 + 别名 + pluginhost 切换，独立可编译），再 D1（包迁移 + 全量 import 更新），然后 D3（治理测试，依赖前两步完成才能通过），最后 D4（README）。回滚策略：整个变更为纯重构，git revert 即可。

## Open Questions

无。三个决策点（推进节奏、迁移方式、治理形态）已在探索阶段与用户确认。
