## Context

`Nightly Test and Build` 已经把发布前最关键的验证拆成独立 job：Windows 命令冒烟、Go 单测、前端单测、完整 E2E、Redis cluster smoke，最后才发布 nightly 镜像。`Release Build` 当前只在 tag push 后执行 Windows 命令冒烟和镜像发布，`release-image` 没有等待 Go/前端/E2E/集群验证。

完整 E2E runner 已经具备官方插件测试发现能力：Playwright 配置同时匹配宿主 `hack/tests/e2e` 和 `apps/lina-plugins/<plugin-id>/hack/tests/e2e`，`pnpm test` 的 `full` 模式选择 `e2e` 与 `plugins` 两个范围。当前活跃变更 `official-plugins-submodule-decoupling` 正在定义 host-only 与 plugin-full 两种插件工作区模式；release 发布链路需要明确选择 plugin-full 验证，否则发布可能只覆盖宿主能力。

## Goals / Non-Goals

**Goals:**

- 让 release 镜像发布与 nightly 一样先通过完整测试门禁。
- 将发布 workflow 重命名为 `release-test-and-build.yml`，名称和职责保持一致。
- 发布镜像前显式验证官方插件工作区可用，并运行官方插件自有 E2E。
- 让 `release-image` job 只能在所有测试 job 成功后执行。
- 保持现有 tag 校验、多架构镜像推送、`latest` 标签发布和 manifest inspect 行为。

**Non-Goals:**

- 不改变发布镜像的 Dockerfile、镜像内容或运行时配置语义。
- 不新增业务 API、数据库 schema、前端页面或运行时文案。
- 不在本变更中完成官方插件 submodule 解耦实现；本变更只消费其 plugin-full 约定。
- 不改变 E2E 用例内容，除非实现时发现官方插件测试入口缺少必要 preflight。

## Decisions

### 决策 1：release workflow 复用 nightly 测试门禁结构

发布前测试应覆盖 nightly 已经证明有价值的验证面，而不是维护一套独立的轻量发布测试。release workflow 将保留 `windows-command-smoke`，并增加 `backend-unit-tests`、`frontend-unit-tests`、`e2e-tests` 和 `redis-cluster-smoke`。

备选方案：

- 只在 release 中运行 smoke E2E：速度更快，但无法证明官方插件和跨模块行为在发布标签上可用。
- 只依赖 nightly 最近一次结果：无法证明当前 tag 对应 commit 已通过验证，且 nightly 与 release 触发时点可能不同。

### 决策 2：release-image 通过 `needs` 等待所有验证

`release-image` 必须声明依赖所有测试 job。这样任何测试失败、取消或超时都会阻止 GHCR release tag 和 `latest` tag 推送，避免先发布后回滚。

备选方案：

- 在单一 job 中串行执行测试和发布：日志集中但总耗时长，失败隔离差，也不能复用 reusable workflows。
- 使用独立 workflow_run 触发发布：可以拆分职责，但 tag 与测试结果的对应关系更难审计。

### 决策 3：官方插件发布验证采用 plugin-full 语义

Release 发布的是完整 LinaPro 框架交付物时，必须验证官方插件目录已初始化并参与测试。checkout 应支持 `submodules: recursive`，并增加 preflight 检查 `apps/lina-plugins` 下至少存在官方插件清单和插件 E2E 目录。若工作区缺失或为空，release 应快速失败并提示初始化 submodule，而不是降级为 host-only。

备选方案：

- 允许 release 在 host-only 模式发布：适合未来专门的 host-only 镜像，但与“发布前验证官方插件稳定性和兼容性”的目标冲突。
- 在 E2E 失败时才暴露插件缺失：错误定位晚，容易被误判为普通测试发现问题。

### 决策 4：减少 nightly/release 漂移

实现时优先复用 reusable workflow 或抽取共同 job，避免 nightly 和 release 的测试步骤长期分叉。短期可以复制 nightly 的 E2E 与 Redis smoke job 到 release workflow；若复制导致明显维护成本，应抽取为 reusable workflow。

备选方案：

- 完全复制 nightly 文件再局部修改：实现最快，但后续工具版本、超时、artifact 策略容易漂移。
- 立即大规模重构所有 CI：一致性最好，但会扩大变更面。

## Risks / Trade-offs

- [Risk] release 耗时显著增加。→ 使用并行 job 编排，保留合理 timeout，并上传 E2E/服务日志便于失败定位。
- [Risk] tag push 后测试失败时 tag 已存在。→ 本变更先阻止镜像发布；后续可增加手动 release dispatch 或环境审批，将“创建 tag”也纳入发布流程治理。
- [Risk] 官方插件 submodule 迁移未完成时 preflight 设计可能需要调整。→ 与 `official-plugins-submodule-decoupling` 的 plugin-full 约定对齐，当前仓库内置插件目录也应满足同一 preflight。
- [Risk] release 和 nightly 对 checkout/submodule 策略不一致。→ 发布完整插件镜像时 release 必须初始化插件；nightly 可后续按同样策略补齐或继续覆盖当前主仓库插件目录。
- [Risk] E2E flake 会阻塞发布。→ 保留 Playwright trace/video/report 上传；若发现高频 flake，应修测试而不是绕过 release 门禁。

## Migration Plan

1. 新增或重命名 release workflow 为 `.github/workflows/release-test-and-build.yml`。
2. 将 workflow name 改为 `Release Test and Build`，保留 tag push 触发和 release tag 合法性校验。
3. 引入完整测试 job，并让 `release-image` 通过 `needs` 依赖全部测试 job。
4. 为 release checkout 增加官方插件工作区初始化/验证策略，确保插件测试在发布前真实执行。
5. 保留发布镜像步骤，确认只有全部测试成功后才执行 GHCR login、`make image push=1`、`latest` 发布和 manifest inspect。
6. 运行 workflow 静态校验、相关脚本校验和 OpenSpec 校验。

Rollback 策略：如 release 流程出现无法及时修复的 CI 环境问题，可临时恢复旧 release workflow 文件，但不得绕过失败测试直接发布 `latest`；应改用明确标记的临时镜像 tag 完成诊断。

## Open Questions

- 官方插件 submodule 远端和默认 checkout 策略是否会在 `official-plugins-submodule-decoupling` 中先落地。
- 是否要把 release 触发从 tag push 进一步升级为 `workflow_dispatch` 加环境审批，以真正做到“创建 release tag 前验证”。
