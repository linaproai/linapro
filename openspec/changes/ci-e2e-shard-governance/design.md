## Context

当前 Nightly 在 host-only 上以 `pnpm test:host` 单 job 执行约 100+ 宿主用例，其中绝大多数被 `serial` 声明串行；plugin-full 虽有 5 路 Playwright `--shard`，但 `plugins` 树整体 serial，且 count 均分导致时长严重不均。墙钟≈最慢 E2E job。

项目已有执行治理基础：`execution-manifest.json`、`run-suite.mjs` 的 parallel/serial 池、module/host-module/plugin 入口、`validate-e2e.mjs`。应在该治理面扩展，而不是在 YAML 中堆临时 matrix。

## Goals / Non-Goals

**Goals:**

- Host-only 与 plugin-full 的 CI 分片由 manifest + 算法共同定义，可校验、可本地复现。
- 各分片使用独立 runner（独立 PostgreSQL 与服务进程），使「模块内 serial」在「模块间 job 并行」下仍安全。
- 插件分片不在根目录硬编码官方插件业务别名；按工作区发现结果装箱。
- 验证器阻止「漏测」与「重复覆盖」。
- 缓存与默认 workers 作为长期默认，而非一次性参数 hack。

**Non-Goals:**

- 不在本次强制把大量 serial 用例改写为 parallel（高 flaky 风险；分片后墙钟已可显著下降）。
- 不把 Main CI / Release 默认打开完整 E2E。
- 不改变 Playwright 业务断言语义与 TC 编号规则。
- 不引入按时长历史自适应分片（可后续用 weightOverrides 演进）。

## Decisions

### Decision 1: `ciShards` 进入 execution-manifest

在 `config/execution-manifest.json` 增加：

```json
"ciShards": {
  "defaults": { "parallelWorkers": 2 },
  "host": [
    { "name": "core", "entries": ["e2e/auth", "e2e/dashboard", "e2e/about"] },
    { "name": "iam", "entries": ["e2e/iam"] },
    ...
  ],
  "plugin": {
    "targetShardCount": 8,
    "packing": "file-count-binpack",
    "weightOverrides": {}
  },
  "pluginFullExtra": [
    { "name": "extension-plugin", "scope": "extension:plugin" }
  ]
}
```

- **host**：显式 entries，与能力边界对齐；校验全集 = host-only 发现集。
- **plugin**：`targetShardCount` + bin-pack；单元为「每个源码插件的 TC 集合」，权重默认 = 文件数，可用 `weightOverrides[pluginId]` 微调。
- **pluginFullExtra**：plugin-full 环境特有的宿主框架分片（如 `extension:plugin`），与 host-only 语义区分保留。

### Decision 2: 运行入口 `ci-shard`

```text
node scripts/run-suite.mjs ci-shard host <name>
node scripts/run-suite.mjs ci-shard plugin <name>
node scripts/run-suite.mjs ci-shard plugin-full-extra <name>
```

复用现有 parallel/serial 拆分与 worker 环境变量；不绕过隔离治理。

### Decision 3: 动态 GitHub matrix

`reusable-test-verification-suite` 增加轻量 `plan-e2e-shards` job：

1. checkout（plugin 需要 recursive submodule）
2. 运行 `emit-ci-shards.mjs` 写出 host/plugin matrix JSON
3. host-only / plugin-full E2E 使用 `strategy.matrix.shard: fromJson(...)`

避免 YAML 与插件清单双源漂移。

### Decision 4: 默认 parallelWorkers=2

- Nightly `e2e-parallel-workers` 默认改为 2（与 manifest defaults 对齐）。
- 仅加速 parallel 池；serial 池仍 workers=1。
- 分片后的墙钟收益主要来自 job 并行，workers 为叠加收益。

### Decision 5: E2E job 缓存

- `setup-go` 开启 module cache。
- Playwright browsers 使用 `actions/cache` 缓存 `~/.cache/ms-playwright`。

## Risks / Trade-offs

| 风险 | 缓解 |
| --- | --- |
| 分片增多导致 runner 分钟数上升 | 墙钟下降优先；targetShardCount 可调；Main CI 仍关 E2E |
| bin-pack 仍可能偶发不均 | weightOverrides；后续可接入历史耗时 |
| 动态 matrix 在插件 submodule 失败时无分片 | require-plugin-e2e 与校验失败阻断 |
| 独立 job 重复 setup ~3–4m | 缓存降低安装成本；不与单 job 串行 40m 同量级 |

## Migration Plan

1. 落地 manifest + 脚本 + 校验。
2. 切换 verification suite 为动态 matrix。
3. Nightly 调 workers=2 并开缓存。
4. 本地 `pnpm test:validate` 与 emit dry-run 验证。
5. 观察下一次 Nightly 墙钟与最慢分片，必要时调 `targetShardCount` / weights。

## Open Questions

- 无阻塞问题。后续若需要自适应耗时分片，仅扩展 packing 策略与 weight 来源，不改 CI 契约形状。
