## Why

Nightly 完整 E2E 墙钟被 host-only 单 job（约 40 分钟）和 plugin-full 不均分片（最慢约 22 分钟）主导。现有并行 worker 参数对「整目录 serial」与「插件全量 serial」收益极低；分片策略写死在 workflow YAML 中，新增模块或插件后容易失衡，无法作为长期治理能力演进。

## What Changes

- 在 E2E `execution-manifest` 中引入 **CI 分片治理（ciShards）** 作为单一事实来源：host 按能力边界分片，plugin 按发现到的源码插件做负载均衡装箱。
- 新增 `ci-shard` 运行入口与 `emit-ci-shards` 矩阵输出，供 GitHub Actions 动态生成 host/plugin E2E matrix。
- 验证器强制：host 分片对宿主 TC **完备且不重叠**；plugin 分片在插件工作区就绪时 **完备且不重叠**。
- Nightly/完整 E2E 调用方默认启用更高的并行池 worker，并在 E2E job 中启用 Go module 与 Playwright 浏览器缓存。
- 更新 E2E 与 CI 文档，明确分片治理、并行 worker 与扩展规则。

## Capabilities

### New Capabilities

- （无独立新能力域；本变更以既有 E2E 执行效率能力扩展为主）

### Modified Capabilities

- `e2e-suite-execution-efficiency`: 增加 host-only CI 分片、plugin 负载均衡分片、ciShards 治理与验证、默认并行 worker 与缓存加速要求。

## Impact

- `hack/tests/config/execution-manifest.json`：新增 `ciShards` 配置。
- `hack/tests/scripts/`：分片规划、发射矩阵、校验与 `run-suite` 入口。
- `.github/workflows/reusable-test-verification-suite.yml`、`reusable-e2e-tests.yml`、`nightly-test-and-build.yml`：动态 matrix、workers、缓存。
- `hack/tests/README.md` / `README.zh-CN.md`：文档同步。
- 不改变业务 API、数据库 schema 或前端用户可见行为；不降低 E2E 覆盖范围。
