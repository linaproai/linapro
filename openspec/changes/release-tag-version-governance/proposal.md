## Why

当前 release tag 可以由人工直接创建，现有 `Release Test and Build` 只能在 tag push 之后运行检测，无法保证 GitHub 仓库标签名称与 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version` 始终一致。标签与实际框架版本不一致会让源码、镜像、升级依赖和问题排查产生歧义。

## What Changes

- 为 release tag 增加版本一致性门禁：release tag 名称必须等于 `metadata.yaml` 的 `framework.version`。
- 提供跨平台 `linactl` 校验命令，统一解析 `metadata.yaml` 并校验 tag 名、版本格式和 Docker tag 兼容性。
- 在现有 tag push release workflow 中加入前置校验，校验失败时阻止测试、镜像构建、GHCR 推送和 `latest` 更新。
- 增加受控 release workflow 入口，通过手动触发读取 `framework.version` 并在检测通过后创建 tag，作为配合 GitHub tag ruleset 的推荐打标路径。
- 在仓库文档中说明 tag ruleset 配置建议：禁止普通用户直接创建、更新、删除 release tag，仅允许受控发布身份创建。

## Capabilities

### New Capabilities

### Modified Capabilities
- `release-image-build`: 发布链路必须治理 release tag 与框架元数据版本一致性，并提供受控打标入口。

## Impact

- 影响 `.github/workflows/release-test-and-build.yml`、新增受控 release tag workflow，以及可能复用的 reusable workflow 调用关系。
- 影响 `hack/tools/linactl`，新增跨平台 release tag 校验命令及单元测试。
- 影响 README/AGENTS 发布说明，新增 GitHub tag ruleset 配置建议。
- 不新增后端运行时 API、数据库 schema、前端页面、运行时缓存或业务数据权限逻辑。
