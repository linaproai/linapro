## Why

当前`Nightly Test and Build`虽然支持手动触发，但手动运行仍会先执行完整测试验证套件。维护者在需要补发、重推或诊断`nightly`镜像时，需要一个不受前置测试条件阻塞的人工发布入口。

## What Changes

- 新增一个仅通过`workflow_dispatch`触发的`GitHub Actions` workflow，用于手动构建并发布`nightly`镜像。
- 该手动 workflow 直接复用`.github/workflows/reusable-image-publish.yml`，不依赖`reusable-test-verification-suite.yml`或其他测试 job。
- 保持现有定时`Nightly Test and Build`不变，继续在测试验证通过后发布自动`nightly`镜像。
- 手动发布仍使用现有`nightly`标签策略、多架构平台、官方插件完整构建模式和`GHCR`发布权限。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `release-image-build`: `nightly`镜像发布能力必须提供一个独立的手动触发入口，允许维护者跳过前置测试门禁直接构建并发布`nightly`镜像。

## Impact

- 新增`.github/workflows/manual-nightly-image-build.yml`。
- 复用现有`.github/workflows/reusable-image-publish.yml`，不新增镜像构建脚本、后端代码、数据库结构、前端页面或运行时配置。
- 不影响现有定时`nightly`测试与发布链路，也不改变 release 镜像发布链路。
