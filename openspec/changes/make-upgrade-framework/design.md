## Context

LinaPro 开发工具统一由 `linactl` 承载，根 `Makefile` / `make.cmd` 只做薄包装。仓库已有 `version`、`release.tag.check`、`plugins.update` 等与版本/Git 相关的命令，但没有“把上游框架合并到当前业务分支”的入口。二次开发者需要手动处理 tag 选择与 merge。

约束：

- 必须跨平台（Windows / Linux / macOS），禁止用 Shell 拼装业务逻辑。
- `command_<name>.go` 命名；复杂逻辑可下沉 `internal/`。
- 已有 `db.upgrade` 使用 `runUpgrade`，新命令必须使用不冲突的 handler 名（如 `runFrameworkUpgrade`）。
- 用户明确要求：默认最新稳定版、`v` 指定版本、`v=main` 拉 main。

## Goals / Non-Goals

**Goals:**

- 提供 `make upgrade` / `linactl upgrade [v=...] [remote=...] [force=1]`。
- 默认解析 remote 上最新稳定 semver tag 并 merge 到当前分支。
- `v=vX.Y.Z`（或 `X.Y.Z`，规范化为带 `v` 的 tag）合并指定稳定版本。
- `v=main` 或其它非版本字符串时，按 remote 分支 ref 合并。
- 安全默认：脏工作区、detached HEAD 拒绝执行（`force=1` 仅跳过脏检查）。
- 单元测试覆盖版本解析、最新稳定 tag 选择与关键安全门禁；Git 集成路径可在临时仓库验证。

**Non-Goals:**

- 不自动处理 merge 冲突（冲突时失败并提示用户手动解决）。
- 不自动执行 `db.upgrade`、插件更新、依赖 tidy 或服务重启。
- 不改写 commit 历史（不 rebase、不 force-push）。
- 不自动选择 fork 的 `upstream` remote（默认 `origin`，显式 `remote=` 覆盖）。
- 不支持预发布 tag 作为“最新稳定版”（`-rc` / `-alpha` 等排除在默认选择之外；若用户显式 `v=` 指定仍可尝试合并）。

## Decisions

### D1: 实现落点为 `linactl upgrade` + Make 薄包装

- **选择**：业务逻辑在 `hack/tools/linactl/command_upgrade.go`，必要时纯函数可同文件或小 internal 包；`hack/makefiles/release.mk`（或 `dev.mk`）增加 `upgrade` 目标转发。
- **备选**：独立 shell 脚本 — 违反 dev-tooling 跨平台与 Go 优先规则。
- **理由**：与现有 `version` / `plugins.update` 一致，Windows `make.cmd` 无需额外入口。

### D2: 稳定版本定义与解析

- **选择**：稳定 tag 匹配 `^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`，规范化为 `vMAJOR.MINOR.PATCH`。默认最新 = 对 remote tags 做 semver 比较取最大。
- **备选**：读取 `metadata.yaml` 的 `framework.version` — 那是本地当前版本，不能表示“上游最新”。
- **理由**：与仓库现有 release tag 约定（`v0.5.0` 等）一致。

### D3: `v` 参数语义

| 输入 | 行为 |
| --- | --- |
| 省略 / 空 | 最新稳定 tag |
| 匹配稳定版本模式 | 合并对应 tag（缺 `v` 前缀则补齐） |
| 其它非空字符串（如 `main`） | 合并 `<remote>/<name>` 分支 |

- **备选**：仅允许 tag 与硬编码 `main` — 过于死板；允许任意 branch 成本低且对长期分支有用。

### D4: Git 操作序列

1. 确认 `git` 可用、当前目录为仓库根。
2. 解析当前分支；detached HEAD 则失败。
3. 非 `force` 时要求工作区干净（`git status --porcelain` 为空）。
4. 确保官方 remote 并 `git fetch linapro --tags --prune`。
5. 解析目标 ref（tag 名或 `linapro/<branch>`）。
6. 记录 pre-merge `HEAD`，执行 `git merge --no-edit --no-commit <ref>`。
7. 将 `apps/lina-plugins` 恢复为 pre-merge 状态（官方插件变更不进入结果）；若升级前不存在该路径，则丢弃官方引入的该路径。
8. 若仅剩插件变更被剔除后索引无差异，则 `merge --abort` 并报告无宿主变更。
9. 否则 `git commit --no-edit` 完成合并，并提示插件需用 `plugins.update` 单独更新。

- **备选**：`git pull` — 语义不如显式 fetch + merge 清晰，且默认 tag 升级不是 pull 跟踪分支。
- **备选**：合并后依赖用户手工处理 submodule — 不可靠，默认会改写 gitlink。

### D5: 升级源硬编码为官方仓库

- **选择**：固定 URL `https://github.com/linaproai/linapro.git`；工具托管 remote 名 `linapro`（每次运行 add 或 set-url 纠正）；不提供 `remote=` 覆盖。
- **备选**：默认 `origin` / 可选 `remote=` — fork 场景会误拉业务方仓库，不符合“升级框架”语义。
- **测试**：通过可覆盖的 `officialFrameworkRepoURL` 变量指向临时 bare 仓，避免单测访问网络。

### D6: 与 `db.upgrade` 命名隔离

- 公开命令名 `upgrade`（框架源码升级）与 `db.upgrade`（数据库 SQL 重放）并存。
- handler 使用 `runFrameworkUpgrade`，避免与 `runUpgrade` 冲突。

## Risks / Trade-offs

- **[Risk] merge 冲突导致半完成状态** → Mitigation：不自动 abort 之外的额外清理；错误信息提示用户解决冲突或 `git merge --abort`。
- **[Risk] 默认 merge 最新稳定版可能跨多个大版本** → Mitigation：文档说明可 `v=` 指定中间版本；命令打印将要合并的目标 ref。
- **[Risk] 无法访问 GitHub 官方仓** → Mitigation：fetch 失败直接报错并打印官方 URL，不回退到 `origin`。
- **[Risk] 脏工作区强制合并损坏未提交改动** → Mitigation：默认拒绝；`force=1` 仅跳过检查且文档标明风险。
- **[Trade-off] 不自动跑 db/plugin 升级** → 保持命令单一职责，避免一次操作耦合过多副作用。
- **[Trade-off] 不支持自定义 remote** → 保证二次开发始终对齐官方源；需要实验性源码时用户可直接使用 git。

## Migration Plan

- 纯增量命令，无数据迁移。
- 回滚：删除 `upgrade` 命令与 Make 目标即可，不影响既有命令。

## Open Questions

- 无阻塞问题。
