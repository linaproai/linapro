## Context

`.github/workflows/nightly-test-and-build.yml`当前同时承载每日定时健康验证和镜像发布。它的`workflow_dispatch`入口也会先运行`reusable-test-verification-suite.yml`，`nightly-image` job 通过`needs`等待验证套件成功后才调用`reusable-image-publish.yml`。这符合自动 nightly 发布的质量门禁，但不适合维护者在测试门禁失败、需要补发当天镜像或需要重新推送`nightly`浮动标签时使用。

现有`.github/workflows/reusable-image-publish.yml`已经封装了`GHCR`登录、`linux/amd64`与`linux/arm64`多架构构建、日期型`nightly-<yyyymmdd>`不可变标签、`nightly`浮动标签和官方插件完整构建模式。因此本变更只需要新增一个轻量调用方 workflow，而不是复制镜像构建步骤。

## Goals / Non-Goals

**Goals:**

- 提供独立的手动`nightly`镜像构建入口。
- 手动入口不依赖测试验证套件或其他前置测试 job。
- 继续复用统一镜像发布 workflow，保持标签、平台、插件完整构建和发布权限一致。
- 保留现有定时`Nightly Test and Build`测试门禁语义。

**Non-Goals:**

- 不移除现有`Nightly Test and Build`的`workflow_dispatch`入口。
- 不改变 release 镜像发布策略、`latest`标签或发布前测试门禁。
- 不修改`make image`、`image-builder`、`Dockerfile`或运行时应用代码。

## Decisions

1. 新增`.github/workflows/manual-nightly-image-build.yml`，而不是在现有 nightly workflow 中按事件类型条件跳过测试。

   原因是独立 workflow 能在`GitHub Actions`列表中清晰暴露“手动补发镜像”能力，并避免在现有 nightly workflow 中引入复杂的条件分支。定时 nightly 继续保持完整测试链路。

2. 手动 workflow 只声明一个`manual-nightly-image` job，并直接`uses: ./.github/workflows/reusable-image-publish.yml`。

   原因是镜像构建、标签生成、浮动标签发布和 manifest inspect 已由 reusable workflow 统一维护；新增调用方只负责触发条件、权限和并发边界。

3. 并发组使用独立的`manual-nightly-image`标识。

   原因是手动补发入口应避免多个手动构建互相覆盖日志和竞争同一`nightly`浮动标签；同时不改变现有定时 nightly workflow 的并发设置。

## Risks / Trade-offs

- 手动入口跳过测试后可能发布来自未验证提交的`nightly`镜像。缓解方式：workflow 名称、注释和 OpenSpec 记录明确其用途是维护者手动补发；自动 nightly 仍保留测试门禁。
- 手动入口和定时 nightly 可能在同一天发布相同不可变日期标签。缓解方式：镜像发布逻辑复用现有标签策略，重复运行会重新推送同名日期标签并更新`nightly`浮动标签，这与补发目的相符。
- 新增 workflow 仍使用`GitHub Actions`内联 shell 步骤间接来自 reusable workflow。缓解方式：本变更不新增默认开发工具或仓库脚本入口，仅新增 CI 编排文件。
