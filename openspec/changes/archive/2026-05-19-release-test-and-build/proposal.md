## Why

当前发布镜像 workflow 在 tag push 后需要先通过测试验证再构建并推送多架构镜像。Nightly 和 Main CI 已经通过共享测试验证套件统一编排验证 job；release workflow 也应复用同一模板，避免维护独立展开的测试 job，同时按 Main CI 的简要测试范围执行发布前门禁，不在 tag 发布链路中重复完整 E2E。

## What Changes

- 将现有 release 镜像发布 workflow 调整为 `release-test-and-build.yml`，语义从“只发布镜像”提升为“先测试、再发布”。
- 在 release 镜像发布前像 nightly 一样复用 `.github/workflows/reusable-test-verification-suite.yml`，并采用 Main CI 的简要测试开关。
- 明确 release 发布链路不运行 host-only E2E 或 plugin-full E2E；完整浏览器回归由 nightly 继续覆盖。
- 保留 plugin-full 的官方插件简要验证和镜像构建语义，避免发布只覆盖 host-only 能力的完整插件镜像。
- 保留镜像 tag 校验、多架构构建、`latest` 浮动标签和远端 manifest inspect，但必须在测试成功后执行。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `release-image-build`: 发布镜像 workflow 必须在镜像推送前复用共享测试验证套件，并使用与 Main CI 一致的不含 E2E 的简要测试范围。
- `e2e-suite-organization`: 完整 E2E 由 nightly 覆盖宿主和官方插件自有 E2E；release 不运行完整 E2E。

## Impact

- 影响 `.github/workflows/release-build.yml` 的文件命名、触发后的 job 编排、共享测试模板参数和镜像发布依赖关系。
- 复用 `.github/workflows/reusable-test-verification-suite.yml`，减少 main/nightly/release 编排漂移。
- 影响 GitHub Actions 总耗时和资源消耗，release 发布会等待简要验证通过，但不重复 nightly 的完整 E2E。
- 不新增后端 REST API、数据库结构、前端用户可见文案或运行时缓存逻辑。
