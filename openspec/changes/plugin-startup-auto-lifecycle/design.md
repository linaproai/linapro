## Context

当前宿主启动链路会在 `apps/lina-core/internal/cmd/cmd_http.go` 中先执行插件发现同步，但这一步只负责把 manifest 写入注册表，不会把插件推进到“已安装/已启用”状态。现状上有两个关键限制：

- 源码插件在 `apps/lina-core/internal/service/plugin/plugin_lifecycle_source.go` 中仍采用显式安装/卸载语义；仅被发现并不会自动安装。
- 动态插件虽然已经具备 `desired_state/current_state` 与主节点 reconciler，但只有管理 API 调用安装/启用动作时，才会把状态推进到 `installed` / `enabled`。

这意味着：

1. 某些要求“宿主启动即就绪”的插件，仍然依赖管理员进入插件页后再点击安装/启用。
2. 源码插件、动态插件的启动期行为缺少统一入口，部署和演示环境需要额外人工补操作。
3. 源码插件路由注册、插件 cron 接线、动态前端 bundle 预热等动作，都依赖最终启用态；如果 bootstrap 发生得太晚，系统即使已监听端口，也可能仍未达到“关键插件已就绪”的状态。

本次设计的约束如下：

- 启动期策略必须在 UI 可用前生效，因此配置源必须是宿主静态配置，而不是依赖插件管理页面或数据库中的后置录入。
- 集群模式下共享生命周期动作必须由主节点执行，避免重复跑插件 SQL、重复写治理资源或重复切换动态 release。
- 当前项目是全新项目，可以直接调整插件生命周期语义与启动顺序，不需要为旧行为做兼容包袱。
- 优先复用现有 `plugin` 服务、动态 reconciler、授权快照模型和 enabled snapshot，不新建额外治理表。

## Goals / Non-Goals

**Goals:**

- 提供宿主级 `plugin.startup` 配置，让运维或开发者按 `pluginId` 声明插件在启动期至少应达到的目标状态。
- 同时覆盖源码插件和动态插件，避免形成两套独立的启动自动化机制。
- 让启动期 bootstrap 发生在插件路由、插件 cron、动态前端 bundle 预热之前，保证系统对外服务前已经尽可能接近目标插件状态。
- 复用动态插件现有授权快照模型，使需要 host service 授权的动态插件也能参与启动自动化。
- 在单节点和集群模式下都具备可预测的失败处理：支持可选插件降级启动，也支持关键插件 fail-fast。

**Non-Goals:**

- 不在本次变更中实现“自动下载/自动构建”动态插件产物；启动期只消费已经可被宿主发现的源码插件目录或动态产物文件。
- 不在本次变更中把启动策略写入 `plugin.yaml`；该策略是环境级部署决策，而不是插件源码契约。
- 不在本次变更中提供插件管理页面对启动策略的可视化编辑；本轮仅定义宿主静态配置和后端启动行为。
- 不把启动策略做成“强制精确状态收敛”的破坏性控制器；移除策略或降低目标状态时，不自动替用户执行卸载/禁用。

## Decisions

### 决策一：以宿主静态配置 `plugin.startup` 作为唯一启动策略入口

宿主在 `apps/lina-core/manifest/config/config.template.yaml` 下新增 `plugin.startup` 配置段，并在 `apps/lina-core/internal/service/config/config_plugin.go` 中统一解析。建议结构如下：

```yaml
plugin:
  dynamic:
    storagePath: "temp/output"
  startup:
    blockUntilReady: true
    readyTimeout: "30s"
    policies:
      - pluginId: "demo-control"
        desiredState: "enabled"
        required: true
      - pluginId: "report-runtime"
        desiredState: "installed"
        required: false
        authorization:
          services:
            - service: "data"
              methods: ["list", "get"]
              tables: ["plugin_report_runtime_record"]
```

配置语义：

- `desiredState` 只允许 `manual` / `installed` / `enabled`。
- `required=true` 表示该插件未达到目标状态时，宿主启动必须失败；否则只记录告警并继续启动。
- `blockUntilReady` 与 `readyTimeout` 控制宿主是否在启动期间等待插件收敛结果；时间长度使用带单位字符串并解析为 `time.Duration`。
- `authorization` 复用现有 `HostServiceAuthorizationInput` 结构，避免为动态插件再定义第二套授权描述模型。

这样设计的原因：

- 静态配置能在宿主开放 HTTP 服务前直接读取，满足“启动即生效”。
- 启动策略是环境部署决策，不应该固化进插件 `plugin.yaml`；同一个插件在开发、演示、生产环境可能需要不同启动策略。
- 复用已有授权模型可以降低认知成本，也能直接对接当前动态插件授权快照持久化逻辑。

备选方案与取舍：

- 把 `autoInstall/autoEnable` 写进 `plugin.yaml`：会把环境策略固化到插件源码，且无法区分不同部署环境，放弃。
- 把启动策略存入 `sys_config`：首次启动和插件未安装时无法稳定依赖，且会让“启动前决策”变成“启动后读取”，放弃。

### 决策二：增加独立的插件启动期 bootstrap 阶段，并前移到插件接线之前

宿主启动顺序调整为：

1. 启动 cluster 选主。
2. 扫描并同步插件 manifest 到注册表。
3. 执行 `plugin startup bootstrap`：解析 `plugin.startup.policies`，对命中的插件推进安装/启用。
4. 刷新 enabled snapshot。
5. 再进行插件 cron 接线、源码插件 HTTP 路由注册、动态 bundle 预热和 runtime reconciler 启动。

对应到实现层，新增类似 `pluginSvc.BootstrapStartupPolicies(ctx)` 的总入口，由它负责：

- 拉取配置并构建 `pluginId -> policy` 映射。
- 对源码插件执行同步的安装/启用推进。
- 对动态插件写入授权快照与目标状态，并在主节点上同步触发一次 targeted reconcile。
- 在 `blockUntilReady=true` 时等待目标插件达到可接受状态或超时。
- 最终统一刷新 enabled snapshot，确保后续路由和 cron 注册看到的是 bootstrap 之后的状态。

这样设计的原因：

- 源码插件路由和 cron 注册本身是“先注册、运行时看 enabled snapshot 决定是否放行”的模式，所以 bootstrap 必须先于 snapshot 刷新完成。
- 动态 bundle 预热、开放 API 前的 readiness 语义，都依赖最终启用态；如果 bootstrap 放在更后面，会出现“服务已启动但关键插件未就绪”的窗口。

备选方案与取舍：

- 继续复用列表页/API 被动触发安装：不能满足“启动即就绪”，放弃。
- 让 runtime reconciler 后台慢慢收敛，不阻塞启动：对演示控制、启动期 Hook、首个请求即依赖插件的场景不可接受，放弃作为默认方案。

### 决策三：启动策略采用“最低目标状态”语义，而不是破坏性精确收敛

`desiredState` 被定义为启动期的“最低目标状态”而非“绝对最终状态”：

- `manual`：宿主不做任何启动期动作。
- `installed`：若插件当前未安装，则推进到已安装；若当前已经启用，不会把它降回已安装。
- `enabled`：若插件当前未安装则先安装，再推进到启用；若当前已启用则保持现状。

同时，移除策略或把 `desiredState` 从 `enabled` 改为 `installed/manual` 时，宿主不会自动执行禁用或卸载。真正的降级动作仍由管理员显式操作触发。

这样设计的原因：

- 启动策略的核心目标是“减少人工补操作”，不是替代完整生命周期治理。
- 避免配置变更在下一次重启时产生破坏性副作用，例如误卸载插件、误关闭正在使用的扩展能力。
- 与现有动态插件 `desired_state/current_state` 机制兼容：只有当当前状态低于策略目标时，才需要写入新的 reconcile 目标。

备选方案与取舍：

- 使用“精确目标状态”语义：虽然更简单，但会让重启变成潜在的破坏性操作，风险过高，放弃。

### 决策四：源码插件与动态插件采用不同推进路径，但共享同一策略模型

同一份 `plugin.startup.policies` 对两类插件都生效，但具体执行分流如下：

- 源码插件：
  - `installed`：调用现有源码插件安装编排。
  - `enabled`：若未安装则先安装，再更新源码插件启用状态。
  - 集群模式下，共享副作用（SQL、菜单、资源索引）只由主节点执行；从节点只读取主节点写入后的稳定状态并刷新本地 snapshot。

- 动态插件：
  - `installed`：必要时持久化授权快照，然后把运行时目标推进到 `installed`。
  - `enabled`：若低于启用态，则写入授权快照并把运行时目标推进到 `enabled`；主节点同步触发 targeted reconcile，从节点只等待共享状态收敛。
  - 继续复用现有 `desired_state/current_state/generation/release_id` 机制，不新增第二套启动专用状态表。

这样设计的原因：

- 源码插件与动态插件的生命周期引擎已经不同：源码插件是同步本地编排，动态插件是主节点 reconcile。统一策略模型可以降低运维使用成本，但实现上仍应尊重现有引擎边界。
- 集群模式下，源码插件虽然是“编译进宿主”的，但安装 SQL、菜单、资源引用仍是共享治理动作，必须避免多节点重复执行。

备选方案与取舍：

- 强行把源码插件也改造成完全 reconcile 化：会引入额外复杂度，且当前并无必要，放弃。
- 为源码插件和动态插件定义两套配置：会增加理解成本和配置重复，放弃。

### 决策五：动态插件授权缺失时，允许按 required 语义选择 fail-fast 或降级到“仅安装”

动态插件如果声明了受治理 host services，且启动策略目标是 `enabled`，则必须满足以下规则：

- 若配置中提供了 `authorization`，宿主在启动期先持久化 release 授权快照，再推进到启用。
- 若未提供 `authorization`：
  - `required=true` 时，认为无法达到目标状态，宿主启动失败。
  - `required=false` 时，宿主记录告警，最多把插件推进到 `installed`，并保留待人工确认授权的状态。

这样设计的原因：

- 启动即启用的前提，是宿主已经明确允许该动态插件访问哪些受治理资源。
- 对非关键插件允许“安装成功但待授权”的降级路径，可以避免把所有动态插件都绑定为 fail-fast。
- 这与当前管理页中的授权确认语义一致，只是把“人工点击确认”换成“静态配置提供确认结果”。

备选方案与取舍：

- 缺少授权也允许直接启用：会让插件以不完整授权快照运行，运行时表现不确定，放弃。
- 缺少授权一律启动失败：对可选插件过于严格，放弃。

## Risks / Trade-offs

- [Risk] 集群模式下选主尚未稳定，启动期 bootstrap 可能等不到主节点完成共享动作。→ Mitigation：引入 `readyTimeout`，主节点负责同步推进，共享动作超时则按 `required` 语义 fail-fast 或降级告警。
- [Risk] 启动策略与管理员手工操作可能产生认知偏差。→ Mitigation：采用“最低目标状态”语义，不做破坏性反向收敛，并在日志/插件详情中明确记录“本次启动由 startup policy 推进”的结果。
- [Risk] 动态插件授权配置较长，手工维护容易出错。→ Mitigation：复用现有授权数据结构和校验逻辑，配置解析阶段即校验 service/method/path/table 的合法性。
- [Risk] 把 bootstrap 前移会延长宿主冷启动耗时。→ Mitigation：只对命中策略的插件执行动作；非必需插件允许降级；`blockUntilReady` 可控。
- [Risk] 源码插件安装流程在集群场景中改为主节点独占后，从节点第一次启动时可能短暂看不到启用态。→ Mitigation：从节点在 bootstrap 后增加一次等待/刷新，再进行路由与 cron 注册。

## Migration Plan

1. 扩展 `plugin.startup` 配置模型与模板，但默认不配置任何 policy，使现有项目升级后行为保持不变。
2. 在 `plugin` 服务中新增 startup bootstrap 入口和 source/dynamic 分流逻辑，并补充配置校验。
3. 调整 `cmd_http.go` 启动顺序，把 bootstrap 插入到插件路由、cron、bundle 预热之前。
4. 补充测试：
   - 源码插件 `manual/installed/enabled` 三种策略；
   - 动态插件安装/启用、授权缺失、授权命中；
   - `required=true/false`；
   - cluster 主从节点下的共享动作与等待逻辑。
5. 发布与回滚：
   - 发布时默认无策略，不会自动改变线上插件状态；
   - 若需要回滚该能力，只需移除 `plugin.startup` 配置并回退代码版本；已自动安装/启用的插件状态不会被回滚逻辑自动破坏，仍可通过管理 API 显式治理。

## Open Questions

- 是否需要在插件管理列表或详情接口中增加“启动策略命中结果/最近一次 bootstrap 来源”的只读展示字段？本轮实现并不依赖该 UI，但运维可观测性会因此更好。
- 非必需动态插件在缺少授权时，是否需要明确把列表态投影为“已安装待授权（来源：startup policy）”而不是普通 installed？如果需要，可能要补充更细的状态文案，但不一定需要新增底层状态机。
