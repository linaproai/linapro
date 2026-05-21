## Why

当前 Go 单元测试总体耗时偏高，主要原因不是单一 CI 编排问题，而是单元测试主路径混入了真实动态 Wasm 执行、完整插件生命周期、共享 PostgreSQL 状态和大量无测试包扫描等重型工作。需要在保留 `-race` 并发安全检测价值的前提下，重新划分单元、集成和 smoke 测试边界，降低每次全量单测的总体时间开销。

## What Changes

- 建立 Go 单元测试执行效率规范，明确单元测试、集成型单元测试和真实链路 smoke 的边界。
- 将真实 bundled dynamic Wasm 样例执行限制为少量 smoke 覆盖，普通单测优先使用 synthetic artifact、fake executor 或轻量测试替身。
- 为插件 runtime、catalog、integration 和生命周期测试引入可复用的轻量 fixture，减少重复写 artifact、同步 manifest、安装启用插件和清理治理表的成本。
- 调整 `linactl test.go` 的测试发现与执行策略，避免对大量无 `_test.go` 的包执行完整单测入口，同时保留必要的编译 smoke。
- 保留主测试路径中的 `-race` 覆盖，不以移除 race 作为降耗手段；优化重点放在测试设计和重复重型 fixture 治理上。
- 修正 PostgreSQL CI health check 用户配置，减少 `role "root" does not exist` 等无关日志噪声，提升后续性能分析可读性。

## Capabilities

### New Capabilities

- `go-unit-test-execution-efficiency`: 约束 Go 单元测试的执行效率、测试层级、重型 fixture 复用、真实 Wasm smoke 边界、race 覆盖和可观测耗时报告。

### Modified Capabilities

- 无。

## Impact

- 影响 `hack/tools/linactl` 中 `test.go` 命令的测试包发现、执行计划和报告输出。
- 影响 `apps/lina-core/internal/service/plugin*`、`apps/lina-core/pkg/plugin*` 和 `hack/tools/linactl/internal/wasmbuilder` 等高耗时测试的 fixture 和用例结构。
- 影响 `.github/workflows/reusable-backend-unit-tests.yml` 中 PostgreSQL health check、Go cache 或测试执行参数配置。
- 不改变生产 API、运行时业务行为、数据库结构或前端页面；本变更不新增用户可见文案，预期不影响 i18n 资源。
