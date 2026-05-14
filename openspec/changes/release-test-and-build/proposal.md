## Why

当前发布镜像 workflow 在 tag push 后会直接构建并推送多架构镜像，缺少与 nightly 同等级别的 Go 单测、前端单测、完整 E2E 和集群冒烟门禁。发布后才发现宿主或官方插件兼容性问题，会把普通修复升级成紧急修复，影响 LinaPro 作为可持续交付全栈框架的交付可信度。

## What Changes

- 将现有 release 镜像发布 workflow 调整为 `release-test-and-build.yml`，语义从“只发布镜像”提升为“先测试、再发布”。
- 在 release 镜像发布前复用 nightly 的关键验证阶段：Windows 命令冒烟、Go 单元测试、前端单元测试、完整 E2E 和 Redis cluster smoke。
- 明确 release 发布链路必须验证官方插件工作区和官方插件自有测试，避免只验证 host-only 能力后发布包含或声明包含官方插件能力的镜像。
- 为官方插件 submodule/可选工作区演进预留显式 preflight：发布完整插件镜像时必须初始化并验证 `apps/lina-plugins`。
- 保留镜像 tag 校验、多架构构建、`latest` 浮动标签和远端 manifest inspect，但必须在测试成功后执行。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `release-image-build`: 发布镜像 workflow 必须在镜像推送前完成完整测试门禁，并显式验证官方插件工作区。
- `e2e-suite-organization`: release 完整 E2E 必须覆盖宿主和官方插件自有 E2E，且在官方插件工作区缺失时快速失败。

## Impact

- 影响 `.github/workflows/release-build.yml` 的文件命名、触发后的 job 编排、checkout 策略、测试依赖和镜像发布依赖关系。
- 可能复用或抽取 `.github/workflows/nightly-test-and-build.yml` 中的完整 E2E 与 Redis cluster smoke 逻辑，减少 nightly/release 漂移。
- 影响 GitHub Actions 总耗时和资源消耗，release 发布会等待完整验证通过。
- 不新增后端 REST API、数据库结构、前端用户可见文案或运行时缓存逻辑。
