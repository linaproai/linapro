## Context

`Nightly Test and Build` 和 `Main CI` 已经通过 `.github/workflows/reusable-test-verification-suite.yml` 统一编排验证 job。Nightly 启用完整 E2E，Main CI 使用不含 E2E 的简要测试范围。`Release Build` 当前只在 tag push 后执行 Windows 命令冒烟和镜像发布，`release-image` 没有复用共享测试模板，也没有按 Main CI 的简要范围等待 Go、前端、命令冒烟、Redis、SQLite 和构建冒烟验证。

完整 E2E runner 已经具备官方插件测试发现能力：Playwright 配置同时匹配宿主 `hack/tests/e2e` 和 `apps/lina-plugins/<plugin-id>/hack/tests/e2e`，`pnpm test` 的 `full` 模式选择 `e2e` 与 `plugins` 两个范围。完整浏览器回归由 nightly 覆盖；release 发布链路只运行与 Main CI 一致的简要测试，并保留 plugin-full 的官方插件单测、前端单测、Windows 命令冒烟和镜像构建语义。

## Goals / Non-Goals

**Goals:**

- 让 release 镜像发布与 nightly 一样复用共享测试验证套件。
- 将发布 workflow 重命名为 `release-test-and-build.yml`，名称和职责保持一致。
- 发布镜像前采用与 Main CI 一致的不含 E2E 的简要测试范围。
- 让 `release-image` job 只能在 tag 校验和共享测试套件成功后执行。
- 保持现有 tag 校验、多架构镜像推送、`latest` 标签发布和 manifest inspect 行为。

**Non-Goals:**

- 不改变发布镜像的 Dockerfile、镜像内容或运行时配置语义。
- 不新增业务 API、数据库 schema、前端页面或运行时文案。
- 不在本变更中完成官方插件 submodule 解耦实现；本变更只消费其 plugin-full 约定。
- 不改变 E2E 用例内容；release 不新增或运行完整 E2E。

## Decisions

### 决策 1：release workflow 复用共享测试验证套件

发布前测试不应维护一套独立展开的 job。release workflow 通过 `reusable-test-verification-suite.yml` 编排验证阶段，和 nightly/main CI 共享同一套测试模板、artifact 命名输入和逐项开关。

备选方案：

- 在 release workflow 中继续手工展开每个测试 job：短期可行，但 main/nightly/release 的 job 名称、artifact、参数和开关会继续漂移。
- 只依赖 nightly 最近一次结果：无法证明当前 tag 对应 commit 已通过任何发布前验证，且 nightly 与 release 触发时点可能不同。

### 决策 2：release 使用 Main CI 的简要测试范围

Release workflow 采用与 Main CI 一致的不含 E2E 的简要测试开关：启用 host-only/plugin-full 的 Windows 命令冒烟、Go 单元测试、前端单元测试，启用插件命令冒烟、常用 make 命令冒烟、Redis integration、SQLite smoke、host-only build smoke 和 Redis cluster smoke，关闭 host-only E2E 与 plugin-full E2E。完整 E2E 继续由 nightly 覆盖。

备选方案：

- release 也运行完整 E2E：覆盖最强，但 tag 发布链路耗时和 flake 风险高，且会重复 nightly 已承担的完整浏览器回归。
- release 只做镜像构建：速度最快，但无法证明当前 tag 对应 commit 通过基础测试门禁。

### 决策 3：release-image 通过 `needs` 等待 tag 校验与共享验证

`release-image` 必须声明依赖 release tag 校验和 `verification-suite`。这样任何简要测试失败、取消或超时都会阻止 GHCR release tag 和 `latest` tag 推送，避免先发布后回滚。

备选方案：

- 在单一 job 中串行执行测试和发布：日志集中但总耗时长，失败隔离差，也不能复用 reusable workflows。
- 使用独立 workflow_run 触发发布：可以拆分职责，但 tag 与测试结果的对应关系更难审计。

### 决策 4：官方插件发布验证采用 plugin-full 简要语义

Release 发布的是完整 LinaPro 框架交付物时，仍必须保留 plugin-full 语义。共享测试套件启用 plugin-full Windows 命令冒烟、Go 单元测试和前端单元测试；镜像发布 job 继续通过 `plugins: "1"` 构建包含官方插件的镜像，并由 reusable image publish workflow 验证官方插件工作区。

备选方案：

- 允许 release 在 host-only 模式发布：适合未来专门的 host-only 镜像，但与当前完整插件镜像发布目标冲突。
- 在 release 中额外运行插件 E2E：覆盖更强，但与本次“使用 Main CI 简要测试、不做完整测试”的目标冲突。

### 决策 5：减少 main/nightly/release 漂移

实现时优先复用 reusable workflow，避免 main、nightly 和 release 的测试步骤长期分叉。Release 调用共享测试套件时必须显式传入全部必填开关，避免通过默认值隐式启停测试。

备选方案：

- 完全复制 main CI 文件再局部修改：实现最快，但后续工具版本、超时、artifact 策略容易漂移。
- 立即大规模重构所有 CI：一致性最好，但会扩大变更面。

## Risks / Trade-offs

- [Risk] release 仍比只构建镜像耗时更长。→ 复用共享测试套件并采用 Main CI 简要范围，不重复 nightly 完整 E2E。
- [Risk] tag push 后测试失败时 tag 已存在。→ 本变更先阻止镜像发布；后续可增加手动 release dispatch 或环境审批，将“创建 tag”也纳入发布流程治理。
- [Risk] release 不运行完整 E2E，可能晚于 nightly 才发现浏览器回归。→ nightly 继续启用 host-only 和 plugin-full E2E；release 保留单测、命令冒烟、Redis、SQLite、构建冒烟和 plugin-full 镜像构建校验。
- [Risk] 官方插件 submodule 迁移未完成时 preflight 设计可能需要调整。→ image publish workflow 继续使用递归 submodule checkout 和官方插件工作区 preflight，测试套件保留 plugin-full 简要验证。

## Migration Plan

1. 新增或重命名 release workflow 为 `.github/workflows/release-test-and-build.yml`。
2. 将 workflow name 改为 `Release Test and Build`，保留 tag push 触发和 release tag 合法性校验。
3. 引入 `verification-suite` job，调用 `reusable-test-verification-suite.yml` 并显式传入与 Main CI 一致的不含 E2E 的简要测试开关。
4. 让 `release-image` 通过 `needs` 依赖 release tag 校验和 `verification-suite`。
5. 保留发布镜像步骤，确认只有简要测试成功后才执行 GHCR login、`make image push=1`、`latest` 发布和 manifest inspect。
6. 运行 workflow 静态校验、共享套件输入覆盖检查、actionlint 和 OpenSpec 校验。

Rollback 策略：如 release 流程出现无法及时修复的 CI 环境问题，可临时恢复旧 release workflow 文件，但不得绕过失败测试直接发布 `latest`；应改用明确标记的临时镜像 tag 完成诊断。

## Open Questions

- 官方插件 submodule 远端和默认 checkout 策略是否会在 `official-plugins-submodule-decoupling` 中先落地。
- 是否要把 release 触发从 tag push 进一步升级为 `workflow_dispatch` 加环境审批，以真正做到“创建 release tag 前验证”。
