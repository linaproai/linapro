## Context

当前仓库已有 `Release Test and Build` workflow，在 `push.tags` 后运行完整测试和镜像发布。这个链路能阻止错误 tag 发布镜像，但 tag push 事件发生时标签已经存在，GitHub Actions 失败不能回滚或拒绝标签创建。要实现“打标签前拒绝”，需要把实际 tag 创建收口到受控 workflow，并在 GitHub 仓库侧通过 tag ruleset 禁止普通用户直接创建 release tag。

`apps/lina-core/manifest/config/metadata.yaml` 已将 `framework.version` 标注为框架元数据和发布流程的版本基线，当前发布镜像 tag 又直接使用 Git tag，因此 release tag、框架版本和镜像 tag 应保持同一个值。

## Goals / Non-Goals

**Goals:**
- 在 CI 和受控打标入口中统一校验 Git tag 与 `framework.version` 完全一致。
- 使用跨平台 Go 工具读取 YAML 和校验版本，避免在 workflow 中散写不可复用脚本。
- 让错误 tag 在现有 release workflow 中快速失败，不进入测试、镜像构建或发布步骤。
- 提供可以配合 GitHub tag ruleset 的手动 release tag workflow，使 tag 在创建前被校验。
- 明确文档化仓库规则配置，避免维护者误以为 tag push workflow 本身可以拒绝 tag 创建。

**Non-Goals:**
- 不通过代码自动修改 GitHub 仓库 settings 或 ruleset。
- 不改变 `framework.version` 的升级策略，也不自动改写 `metadata.yaml`。
- 不要求 `openapi.version` 与 `framework.version` 同步。
- 不新增前端可见功能、后端运行时 API、数据库表或缓存行为。

## Decisions

1. 新增 `linactl release.tag.check` 作为唯一版本一致性校验入口。
   - 理由：仓库规范要求长期维护脚本优先使用 Go 工具链实现，`linactl` 已是跨平台开发入口。
   - 备选：在 workflow 中用 shell、Python 或 yq 解析 YAML。放弃原因是重复、跨平台价值低，且 GitHub Actions 和本地校验逻辑容易漂移。

2. `release.tag.check` 默认读取 `apps/lina-core/manifest/config/metadata.yaml` 的 `framework.version`，并通过参数接收待校验 tag。
   - 默认 tag 来源按顺序使用 `tag=<value>` 参数、`GITHUB_REF_NAME` 环境变量。
   - 校验要求包括：版本非空、tag 非空、两者完全相等、版本符合 release 格式、可作为 Docker tag。

3. Release tag 格式限定为 Docker tag 兼容的 SemVer 子集。
   - 推荐允许 `vMAJOR.MINOR.PATCH` 和 `vMAJOR.MINOR.PATCH-prerelease`，例如 `v0.2.0`、`v0.2.1-rc.1`。
   - 不允许 SemVer build metadata 的 `+build`，因为 Docker tag 不支持 `+`，会破坏当前 GHCR 镜像 tag 复用逻辑。

4. 保留现有 tag push release workflow，并加入最早执行的 `release-tag-version-check` job。
   - 所有测试和镜像发布 job 都通过 `needs` 依赖该 job。
   - 这样即使有人绕过推荐流程直接 push tag，错误 tag 也不会发布任何 release 制品。

5. 新增 `Create Release Tag` 手动 workflow，用于受控创建 tag。
   - workflow 读取 `framework.version`，运行同一 `linactl release.tag.check` 校验，再创建并 push 同名 tag。
   - workflow 基于仓库变量 `RELEASE_APP_CLIENT_ID` 和仓库密钥 `RELEASE_APP_PRIVATE_KEY` 动态生成 GitHub App installation token。对应 GitHub App 必须安装到当前仓库，并具备 `Contents` 读写权限。
   - 仓库侧应配置 tag ruleset，仅允许受控发布 GitHub App 创建 `v*` tag。GitHub ruleset bypass 绑定的是角色、团队、用户或 GitHub App 等 actor，不是 token 字符串。实际 ruleset 需要维护者在 GitHub 设置中配置，本变更只提供文档化步骤。

## Risks / Trade-offs

- [Risk] 如果未配置 GitHub tag ruleset，人工仍可直接创建错误 tag。→ Mitigation：现有 release workflow 前置校验会阻止错误 tag 发布制品，文档明确 ruleset 是真正拒绝创建的必要条件。
- [Risk] 使用默认 `GITHUB_TOKEN` 创建 tag 可能不会触发另一个 tag push workflow。→ Mitigation：受控打标 workflow 使用 GitHub App installation token 创建 tag，并且 ruleset bypass 配置到对应 GitHub App actor 上；tag push release workflow 仍保留版本一致性兜底检查。
- [Risk] 版本格式过宽会生成非法 Docker tag。→ Mitigation：校验命令同时检查 release 版本格式和 Docker tag 字符集。
- [Risk] workflow 增加一个前置 job 会改变 release job 依赖图。→ Mitigation：只新增校验依赖，不改变各验证 job 内部命令和发布命令。

## Migration Plan

1. 合入代码后，现有 tag push release workflow 会开始执行版本一致性兜底检查。
2. 维护者创建发布 GitHub App，将其安装到当前仓库，授予 `Contents` 读写权限，并配置仓库变量 `RELEASE_APP_CLIENT_ID` 与仓库密钥 `RELEASE_APP_PRIVATE_KEY`。
3. 维护者在 GitHub 仓库配置 tag ruleset，禁止普通用户直接创建、更新、删除 `v*` tag，并将发布 GitHub App 加入 bypass list。
4. 后续 release 使用 `Create Release Tag` workflow 创建 tag；该 workflow 生成 GitHub App installation token 并以 GitHub App actor 创建匹配 `framework.version` 的 tag。
