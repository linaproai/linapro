## Context

`apps/lina-core/internal/service/config` 是宿主级配置服务组件，集中承载以下几类能力：

- `config.yaml` 静态配置读取与默认值回退；
- `sys_config` 受保护运行时参数读取、校验、缓存与快照同步；
- 公共前端设置白名单组装与受保护配置校验；
- 插件动态存储路径、元数据与 OpenAPI 等宿主配置读取；
- 单节点 / 集群模式下的运行时参数修订号同步策略。

当前该包已经有一批测试，但实测 `cd apps/lina-core && go test ./internal/service/config -cover` 覆盖率为 `71.9%`，距离 `80%` 目标仍有缺口。覆盖明细显示缺口主要集中在以下几类代码：

- `config_plugin.go`：插件动态存储路径默认值、兼容回退和 override 路径几乎未覆盖；
- `config_public_frontend.go`：`PublicFrontendSettingSpecs`、`IsProtectedConfigParam`、`ValidateProtectedConfigValue`、时区解析等辅助分支覆盖不足；
- `config_runtime_params_cache.go`：缓存命中、缓存失效、异常回退、错误缓存值清理等冷门路径覆盖不足；
- `config_runtime_params_revision.go`：集群修订控制器的读取、同步、递增和共享 KV 错误分支未形成完整测试；
- 少量 getter / helper（如 `GetJwtSecret`、`GetSessionTimeout`、`GetUploadPath`、`mustScanMetadataConfig`）仍缺少默认值与空对象分支验证。

本次变更不增加新功能，而是将这些宿主关键路径补齐测试保护，并把 `80%+` 覆盖率作为配置管理组件的明确交付门槛。

## Goals / Non-Goals

**Goals:**
- 让 `apps/lina-core/internal/service/config` 包级单元测试覆盖率达到并稳定保持在 `80%` 及以上。
- 优先补齐当前低覆盖热点子模块的默认值、回退、异常、缓存与集群同步分支测试。
- 让新增测试可重复执行、不依赖共享全局状态残留，避免缓存、配置适配器和全局 override 相互污染。
- 在不改变现有对外语义的前提下，按需做小范围测试友好性整理，降低后续继续补测的门槛。

**Non-Goals:**
- 不新增配置管理页面、API、数据库结构或运行时参数能力。
- 不把本次工作扩展为整个 `apps/lina-core` 的全局覆盖率治理，仅针对 `config` 组件包级目标。
- 不为了凑覆盖率而引入无业务意义的测试桩代码或破坏现有实现抽象。
- 不把单元测试替换为 E2E；本次重点仍是服务层与纯函数路径的快速回归验证。

## Decisions

### 1. 以包级覆盖率作为验收基线，而不是单文件逐个设阈值

本次验收统一采用：

- `cd apps/lina-core && go test ./internal/service/config -cover`

并要求最终结果 `>= 80%`。

这样做的原因是：

- Go 原生覆盖率命令简单直接，易于本地执行和 CI 接入；
- 当前缺口分布在多个文件，包级指标更符合本轮“整体补齐”的目标；
- 单文件阈值虽然更细，但会显著增加维护成本，本轮先不引入。

备选方案是为每个文件单独设最低覆盖率，但目前收益不足，且容易让实现被测试策略绑死，因此不采用。

### 2. 优先补低覆盖且高风险分支，而不是平均补测所有文件

本轮优先顺序按“覆盖缺口 x 风险权重”确定：

1. `config_plugin.go`：涉及插件动态存储路径，当前几乎无覆盖，且含兼容回退与 override 逻辑；
2. `config_public_frontend.go`：涉及受保护键判断、统一校验入口、白名单元数据与时区解析；
3. `config_runtime_params_revision.go` / `config_runtime_params_cache.go`：涉及集群同步、缓存重建和退化回退，是最容易在重构时引入隐性回归的路径；
4. 其余 getter / helper：补默认值、空对象、防御性分支。

这样可以在最少测试增量下，尽快把覆盖率和风险面同时拉高。备选方案是先给每个文件平均补 1~2 个测试，但容易继续遗漏真正关键的异常路径，因此不采用。

### 3. 复用现有 fake service 与状态重置模式，必要时只做最小测试友好性改造

现有 `config_runtime_params_test.go` 已经提供了 `fakeRuntimeParamKVCacheService`、运行时参数写入夹具，以及对局部配置状态的测试模式。本轮继续沿用该思路：

- 通过假实现或可替换依赖控制 `kvcache` 行为；
- 在每个测试中显式重置静态缓存、运行时快照、revision 状态和 override；
- 对必须观察的进程级状态（如 plugin storage override、runtime snapshot cache）使用测试专用 helper 做成对恢复。

如果现有代码对测试隔离不够友好，只允许做最小化整理，例如提炼 reset helper、隔离全局变量访问或暴露必要的包内辅助函数；不做影响生产语义的结构性重写。

### 4. 测试场景按“主路径 + 回退路径 + 异常路径”成组设计

为了避免覆盖率上去但回归保护仍然薄弱，本次每类子模块至少覆盖以下三类路径中的两类，关键模块覆盖三类：

- **主路径**：正常读取配置、命中缓存、解析成功；
- **回退路径**：配置缺失、兼容字段回退、默认值生效、缓存已存在直接复用；
- **异常路径**：共享 KV 失败、缓存值损坏、非法输入、空对象或无法解析的状态。

例如：

- 插件配置测试要覆盖默认目录、`runtime.storagePath` 回退、override 清理；
- 公共前端配置测试要覆盖受保护键识别、统一校验入口分发、非法枚举/布尔值报错；
- 运行时修订控制器测试要覆盖 clustered `GetInt`/`Incr` 正常路径和错误传播路径；
- 快照缓存测试要覆盖本地缓存命中、revision 变化后重建、无效缓存条目移除与 fallback。

### 5. 本轮只把覆盖率结果记录为实现验证，不额外引入新的测试框架

仓库当前已经使用 Go 原生 `testing`、`go test -cover` 与必要的包内假实现即可满足需求，因此本轮不新增断言框架或 monkey patch 依赖。这样可以保持测试风格一致，并避免为了补测试再次扩大依赖面。

## Risks / Trade-offs

- **[全局状态相互污染]** → 配置适配器、静态缓存、运行时 snapshot 和 override 都是进程级状态；通过统一 reset helper 和 `t.Cleanup` 成对恢复来缓解。
- **[为测试改代码过度设计]** → 仅允许最小化的测试友好性整理，不引入脱离业务语义的抽象层。
- **[覆盖率达到但价值不足]** → 任务拆分明确要求优先覆盖低覆盖热点和异常分支，而不是只补简单 happy path。
- **[集群/缓存路径测试脆弱]** → 优先通过 fake `kvcache.Service` 和包内状态控制构造确定性测试，避免依赖真实外部环境。

## Migration Plan

1. 先补充 `config` 包低覆盖模块的测试夹具与状态重置辅助方法。
2. 分批为 plugin / public frontend / runtime revision / runtime snapshot cache / getter helpers 增加单元测试。
3. 每完成一组测试后执行 `go test ./internal/service/config -cover`，持续观察包级覆盖率变化。
4. 在任务完成时记录最终覆盖率结果，确认达到 `80%` 及以上后再进入评审。

## Open Questions

- 是否需要在后续迭代把 `config` 包覆盖率校验接入统一 CI 门禁？本轮先在变更实现阶段完成命令级验证，不强制扩展到 CI。
- 若实现中发现个别分支只能通过较重的重构才能稳定测试，是否接受局部提炼依赖边界？本轮默认接受“最小必要重构”，但不扩展为架构重写。
