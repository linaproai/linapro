## Context

当前动态插件生命周期处理器通过 `backend/lifecycle/*.yaml` 显式声明，并由 `build-wasm` 打包进 WASM artifact 的 `lina.plugin.backend.lifecycle` 自定义段。宿主运行时只读取 artifact 中的 `LifecycleHandlers`，在安装、升级、禁用、卸载、租户禁用、租户删除和安装模式切换等场景中按声明调用动态插件处理器。

这种机制运行时确定性较好，但开发体验有重复配置问题：动态插件 controller 已经暴露 `BeforeInstall`、`AfterInstall` 等方法，仍然需要再写一份 YAML 声明 `operation`、`requestType` 和 `internalPath`。遗漏 YAML 时，插件方法存在但宿主不会调用；YAML 与方法命名不一致时，错误会延迟到生命周期执行阶段暴露。

## Goals / Non-Goals

**Goals:**

- 在 `build-wasm` 阶段自动发现动态插件 guest controller 中符合生命周期命名和 bridge 签名的处理器。
- 自动生成与现有 YAML 等价的 `LifecycleContract`，并继续嵌入既有 WASM lifecycle custom section。
- 保留 `backend/lifecycle/*.yaml` 作为可选 override，用于覆盖默认 `requestType`、`internalPath` 或 `timeoutMs`。
- 宿主运行时继续以 artifact 内的显式 lifecycle contracts 为权威输入，不做盲目试探调用。
- 让官方动态示例插件不再需要维护重复 lifecycle YAML 文件。

**Non-Goals:**

- 不改变生命周期操作名称、执行顺序、前置阻断和后置通知语义。
- 不改变动态插件 bridge ABI、WASM custom section 名称、hostServices 授权模型或运行时错误码。
- 不新增 REST API、数据库表、前端交互或用户可见文案。
- 不让宿主在运行时通过探测调用来发现插件方法。

## Decisions

### 决策 1：采用构建期自动发现，而不是宿主运行时盲探

构建工具比宿主运行时更适合发现生命周期方法。动态插件编译成 WASM 后，宿主只看到 bridge 执行入口和 artifact 元数据，无法直接读取 guest 内部反射注册表。运行时盲探会把“未实现处理器”“路径声明错误”“执行失败”混在一起，尤其对 `Before*` 前置阻断场景存在误放行风险。

因此本变更让 `build-wasm` 在打包阶段生成 lifecycle contracts。宿主运行时仍然只读取 artifact 中的显式 contracts：没有 contract 就跳过，有 contract 就按既有失败关闭策略执行。

### 决策 2：由 pluginbridge guest 侧提供元数据发现能力，并让构建期规则保持一致

当前 guest dispatcher 已能通过反射注册 controller 方法，并根据方法名推导 `requestType` 和 `internalPath`。本变更应把这部分推导能力整理为可复用的只读元数据入口，例如返回 method name、request type、internal path 和签名类别。

`build-wasm` 在不执行插件代码的前提下通过 AST 识别生命周期候选，并使用与 guest dispatcher 一致的签名、request type 与 internal path 推导规则。根包 facade 可以继续导出兼容入口；构建工具必须通过共享的公开 helper 和测试覆盖保持 AST 推导与 guest dispatcher 规则一致，避免运行时分发和构建期契约生成漂移。

### 决策 3：生命周期自动发现只接受稳定操作名和合法 bridge 签名

自动发现只匹配既有 `LifecycleOperation` 集合中的方法名，包括 `BeforeInstall`、`AfterInstall`、`BeforeUpgrade`、`AfterUpgrade`、`BeforeDisable`、`AfterDisable`、`BeforeUninstall`、`AfterUninstall`、`BeforeTenantDisable`、`AfterTenantDisable`、`BeforeTenantDelete`、`AfterTenantDelete`、`BeforeInstallModeChange` 和 `AfterInstallModeChange`。

候选方法还必须满足 guest dispatcher 已支持的 bridge handler 签名。非 bridge handler、旧 `Can*` 命名、guard 命名或任意相似方法都不得自动生成 lifecycle contract。

### 决策 4：YAML 作为 override，而不是另一套来源

默认 lifecycle contract 由方法元数据生成：

```text
Operation    = 方法名
RequestType  = dispatcher 推导出的 request type
InternalPath = /__lifecycle/<kebab-case-operation>
TimeoutMs    = 0，表示使用宿主默认生命周期超时
```

如果存在 `backend/lifecycle/*.yaml`，构建工具先自动生成 contracts，再按 `operation` 合并 YAML override。YAML 可以覆盖 `requestType`、`internalPath` 和 `timeoutMs`，但不得声明插件中不存在的生命周期方法；否则构建失败。这样可以保留特殊路径或超时需求，同时避免 YAML 成为必须维护的重复清单。

### 决策 5：官方 demo 删除重复声明，测试确认 artifact 仍包含完整契约

`plugin-demo-dynamic` 应删除 14 个 `backend/lifecycle/*.yaml` 文件，保留 controller 中的生命周期方法。构建测试需要确认官方 demo 打包后仍包含 14 个 lifecycle contracts，并且 operation、requestType、internalPath 与当前行为兼容。

## Risks / Trade-offs

- [风险] 构建工具无法直接加载某些用户插件的 controller 实例。缓解：先覆盖当前官方动态插件结构，要求动态插件 backend 提供可被构建工具导入或执行元数据导出的入口；无法自动发现时构建失败并提示保留 YAML override 或补齐入口。
- [风险] 自动发现规则与 guest dispatcher 注册规则漂移。缓解：把元数据推导放在 `pluginbridge/guest`，构建工具只调用公共入口，并用单元测试同时覆盖 dispatcher 和 metadata。
- [风险] 删除 YAML 后特殊超时配置丢失。缓解：默认使用宿主生命周期超时；需要自定义超时时继续保留对应 operation 的 YAML override。
- [风险] 运行时不可见的构建期错误影响开发者理解。缓解：构建失败信息必须包含插件 ID、operation、方法名和冲突字段。

## Migration Plan

1. 在 `pluginbridge/guest` 中提取 controller handler metadata 发现能力，并保持现有 dispatcher 行为不变。
2. 修改 `build-wasm`，在读取 plugin manifest 和 backend contracts 时自动收集生命周期 metadata，生成默认 contracts。
3. 将 `backend/lifecycle/*.yaml` 处理改为 override 合并，并为不存在的方法、重复 operation、非法 operation、非法 timeout 返回构建错误。
4. 删除官方 `plugin-demo-dynamic/backend/lifecycle/*.yaml`，更新中英文 README，说明生命周期由 controller 自动发现。
5. 补充构建工具、pluginbridge 和 runtime artifact 测试，运行 OpenSpec 校验和 Go 编译烟测。

Rollback 策略：如果自动发现实现风险过高，可以保留当前 YAML 必填模式；已新增的 metadata 发现能力不参与构建路径即可，不影响宿主运行时。

## Cross-Cutting Assessments

- i18n：本变更不新增用户可见 UI、API 文案、apidoc 或运行时错误码。README 文档若更新，需保持中英文镜像一致。
- 缓存一致性：本变更只改变 WASM 构建期 lifecycle contract 来源，不新增缓存、失效、订阅或跨实例协调路径。运行时 artifact 校验和变化仍由既有插件运行时加载和升级机制处理。
- 数据权限：本变更不新增 REST 接口或业务数据操作。动态 lifecycle 处理器若调用 hostServices，仍受既有授权、租户和数据权限边界约束。
- 开发工具脚本：本变更修改 Go 版 `hack/tools/build-wasm`，不得新增平台专属脚本；验证应运行 `cd hack/tools/build-wasm && go test ./... -count=1`。
