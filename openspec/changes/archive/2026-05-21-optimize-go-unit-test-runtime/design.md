## Context

本次变更来源于两份 Go 单元测试日志分析：

- `host-only` 总耗时约 8 分 49 秒，其中 `make test.go` 约 7 分 13 秒。
- `plugin-full` 总耗时约 14 分 32 秒，其中 `make test.go` 约 12 分 47 秒。
- `plugin-full` 中 `lina-core/internal/service/plugin`、`plugin/internal/integration`、`plugin/internal/runtime` 三个包合计报告耗时超过 210 秒。
- 两个真实 bundled dynamic Wasm 相关测试合计接近 98 秒，明显超出普通单元测试粒度。
- 当前 `linactl test.go` 统一执行 `go test -p=1 -race -v ./...`，会扫描大量 `[no test files]` 包，并把 race、串行包执行、真实数据库和真实插件产物成本叠加到所有测试路径。

用户明确要求降低整个单元测试的时间开销，而不是简单为了 main CI 更快牺牲 `-race`。因此本设计把优化目标放在测试设计合理性、重型 fixture 复用和真实链路边界治理上。

## Goals / Non-Goals

**Goals:**

- 保留 Go 单元测试主路径中的 `-race` 并发安全检测能力。
- 降低全量 Go 单元测试总体耗时，优先治理真实 dynamic Wasm 执行、插件生命周期重复初始化和无测试包扫描。
- 让 `linactl test.go` 输出可审计的执行计划与耗时摘要，便于后续持续优化。
- 将真实 bundled Wasm、真实源码插件工作区和完整 PostgreSQL 生命周期测试明确归入集成型单元测试或 smoke 边界。
- 修正测试基础设施日志噪声，使 PostgreSQL health check 不再产生无关 `root` 用户错误。

**Non-Goals:**

- 不移除主测试路径的 `-race` 覆盖。
- 不降低插件 runtime、catalog、integration、缓存、权限和租户相关行为的覆盖要求。
- 不改变生产 API、运行时业务行为、数据库 schema 或前端用户体验。
- 不引入新的第三方测试框架或外部服务依赖。

## Decisions

### Decision 1: 保留 race，优化测试工作量

`-race` 能发现普通断言无法覆盖的并发数据竞争，尤其适用于插件运行时缓存、协调、session、锁、动态插件桥接等组件。本变更不把移除 race 作为降耗手段，而是减少被 race 放大的重复重型工作。

可选方案是将 main CI 默认改成 `race=false`，但这会把并发安全问题后移到 nightly 或 release，不符合当前质量要求，因此不采用。

### Decision 2: 将真实 bundled dynamic Wasm 执行收敛为 smoke

普通单测应验证插件 manifest 解析、runtime 分发、cron 声明、授权和错误路径的业务逻辑，而不应每次都执行真实 bundled Wasm 样例。真实 Wasm 样例仍需要保留最小 smoke，证明打包产物和宿主桥接能够闭环。

具体实现上，普通用例优先使用 `testutil.WriteRuntimeWasmArtifact`、synthetic artifact、fake executor 或轻量 host service 替身；只有 smoke 用例允许调用真实 bundled runtime sample。

### Decision 3: 插件测试 fixture 复用且按作用域隔离

插件测试目前大量重复写入 artifact、同步 manifest、安装/启用插件、刷新 runtime cache 和清理治理表。应提取共享 helper 或包级 fixture，让测试只声明差异化场景，同时使用唯一 plugin id、tenant id、release id 和显式清理保持顺序无关。

对需要真实 PostgreSQL 的测试，继续使用独立数据作用域与幂等清理；对纯逻辑验证，改用内存替身或 mock service，避免每个测试都走完整生命周期。

### Decision 4: `linactl test.go` 先发现测试包，再执行测试计划

`go test ./...` 会遍历无测试包并产生编译与输出成本，插件 workspace 中尤为明显。`linactl test.go` 应先用 Go 工具链或跨平台文件扫描发现包含 `_test.go` 的包，针对这些包执行单测；对没有测试文件但仍需编译保障的模块，使用轻量编译 smoke 或显式报告跳过单测。

该发现逻辑必须使用 Go 实现，保持跨平台，不引入 shell 管道作为默认开发入口。

### Decision 5: 输出耗时摘要作为治理反馈

优化后需要能看到收益和新瓶颈。`linactl test.go` 应记录模块、包数量、跳过的无测试包数量、race/verbose 参数和每个模块耗时摘要。日志不需要替代 `go test` 输出，但必须足够判断后续最慢模块。

## Risks / Trade-offs

- 真实 Wasm 用例减少后可能遗漏打包产物与宿主桥接组合问题 → 保留最小 smoke，并在 fixture 测试中覆盖解析、分发、授权、cron 和错误路径。
- 测试包发现逻辑可能漏跑通过非标准命名触发的测试 → 使用 Go package 元数据或严格 `_test.go` 文件发现，并用 linactl 自身单测覆盖发现规则。
- 共享 fixture 可能引入测试间状态污染 → 每个测试仍必须使用唯一业务标识和 `t.Cleanup`，包级 fixture 只承载不可变或可重建基础资源。
- 分层后测试分类不清可能导致重型测试重新进入单元主路径 → 通过规范、任务记录和 review 检查要求说明测试层级。
- Go cache 或 health check 调整可能改变 CI 环境表现 → 这些改动只影响测试基础设施，不改变生产构建；如出现问题可单独回滚 workflow 配置。
