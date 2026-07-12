## 1. 分片治理模型

- [x] 1.1 在 `hack/tests/config/execution-manifest.json` 增加 `ciShards`（host 分片、plugin packing、pluginFullExtra、defaults.parallelWorkers）
- [x] 1.2 在 `execution-governance.mjs` / `ci-shards.mjs` 实现 host 分片解析、插件文件数 bin-pack、分片完备性检查 API
- [x] 1.3 新增 `emit-ci-shards.mjs`（JSON / GitHub outputs）与 `ci-shards-test.mjs` 单测脚本

## 2. Runner 与验证

- [x] 2.1 `run-suite.mjs` 增加 `ci-shard` 模式（host / plugin / plugin-full-extra）
- [x] 2.2 `validate-e2e.mjs` 接入 ciShards 完备性与命名校验
- [x] 2.3 `package.json` 增加便捷脚本（`test:ci-shards` / `test:ci-shard` / `test:ci-shards:unit`）
- [x] 2.4 更新 `hack/tests/README.md` 与 `README.zh-CN.md`

## 3. CI Workflow

- [x] 3.1 `reusable-test-verification-suite.yml`：plan-e2e-shards + host/plugin 动态 matrix
- [x] 3.2 `reusable-e2e-tests.yml`：Go cache、Playwright browser cache
- [x] 3.3 `nightly-test-and-build.yml`：`e2e-parallel-workers: 2`（与 manifest 默认对齐）
- [x] 3.4 确认 main/release 仍默认关闭完整 E2E

## 4. 验证

- [x] 4.1 运行 `pnpm test:ci-shards:unit` 与分片规划 dry-run（`emit-ci-shards`）
- [x] 4.2 确认 host 6 分片完备、plugin 8 路装箱（tenant-core 独占重片，其余约 12–21 文件）
- [x] 4.3 `ci-shard host core --list` 可解析并列出用例
- [ ] 4.4 `pnpm test:validate`：当前分支另有预存失败（根目录引用官方插件 ID / 高风险并行），与本变更无关；ciShards 校验未报错
