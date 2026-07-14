# Tasks

## 维护摘要

- 完成 E2E 目录按稳定能力边界重组、根路径去官方插件耦合、模块本地 `TC{NNN}` 编号、分层执行入口与 isolation/baseline 治理。
- 完成 host-only / plugin-full 职责收敛，以及 `ciShards` 驱动的 host 能力分片 + plugin 负载装箱、动态 CI matrix、`ci-shard` 入口与完备性校验。
- 完成登录态复用、轻量认证 fixture、状态驱动等待、默认 parallel workers、Go/Playwright 缓存，以及分片暴露的宿主/插件稳定性配套修正与 headless shell 初始化对齐。

## 验证

- [x] `pnpm -C hack/tests test:validate`、`tsc --noEmit`、`openspec validate ... --strict`、`git diff --check` 通过；治理覆盖目录归属、编号、隔离类别、ciShards 完备性与根路径插件耦合。
- [x] `pnpm test:ci-shards:unit` 与 `emit-ci-shards` dry-run：host 分片完备、plugin 装箱覆盖全部源码插件 TC；`ci-shard host <name> --list` 可解析。
- [x] smoke / module / host-only / `extension:plugin` / `plugins` / 各 CI 分片可独立回归；host-only 与 plugin-full 墙钟较基线显著下降并可复核最慢分片。
- [x] 关键支撑：前端 route refresh 与前端单测、相关 Go 精确门禁与启动绑定测试通过；完整套件不靠放宽断言或固定等待掩盖问题。
- [x] 归档前 `lina-review` 复核目录、环境职责、fixture、ciShards、baseline、稳定性修正与验证证据。

## 治理影响

- `execution-manifest` 为 smoke、scope、serial、isolation、allowlist 与 **ciShards** 的单一事实来源；YAML 不得作为分片唯一来源。
- 根 `hack/tests` 禁止硬编码官方插件业务别名/专属资产；插件自有 E2E 仅通过 `plugins` / `plugin:<plugin-id>` 选择。
- Main CI / Release 默认不打开完整 E2E；Nightly 等完整验证默认 parallel workers ≥ 2，E2E job 启用 Go module 与 Playwright 浏览器缓存。
- 跨域：`project-setup` 默认安装 Chromium headless shell；默认上传上限 100MB 与运行时 fallback / packed asset 对齐。
- i18n / 数据权限 / 生产 API 契约：无新增运行时用户可见文案与 schema 变更；稳定性修正仅覆盖测试支撑与最小运行时一致性。
