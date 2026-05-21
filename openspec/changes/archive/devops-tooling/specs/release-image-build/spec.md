## MODIFIED Requirements

### Requirement: 镜像构建命令必须支持多架构 Docker 镜像构建与推送

系统 SHALL 允许运维人员通过 `make image platforms=linux/amd64,linux/arm64 registry=<registry> push=1` 构建并推送多架构 Docker 镜像。镜像构建实现 SHALL 由 linactl 内部镜像构建组件执行，而不是依赖独立 `hack/tools/image-builder` 工具模块。

### Requirement: Nightly image publishing must support a manual no-test entrypoint

系统 SHALL 提供一个独立的 GitHub Actions 手动 workflow，用于构建并发布 nightly 镜像。该 workflow MUST 仅通过 `workflow_dispatch` 触发，MUST 直接调用统一镜像发布 workflow，MUST 不依赖测试验证套件。

### Requirement: Nightly demo image must provide a memory-only Compose launcher

系统 SHALL 在 `hack/deploy/docker-compose.yaml` 提供内存态 nightly 演示启动入口，使用 PostgreSQL 服务作为演示数据库，不挂载宿主数据目录或声明持久化卷。

### Requirement: Release tag 必须与框架元数据版本一致

系统 SHALL 将 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version` 作为 release tag 的唯一版本基线。任何 release tag 发布链路在执行测试、构建、镜像推送或更新浮动标签前，MUST 校验 Git tag 名称与 `framework.version` 完全一致。

### Requirement: Release tag 校验必须通过跨平台工具复用

系统 SHALL 通过仓库跨平台工具入口执行 release tag 与框架元数据版本一致性校验。GitHub Actions、本地发布检查和后续发布自动化 SHALL 复用同一个校验命令。

### Requirement: 受控发布入口必须在创建 tag 前校验框架版本

系统 SHALL 提供受控的 GitHub Actions 手动发布入口，用于读取 `framework.version` 并创建同名 release tag。该入口 MUST 在创建 tag 前运行 release tag 校验。

### Requirement: Release workflow 必须复用共享测试模板并运行简要测试门禁

系统 SHALL 提供 `Release Test and Build` GitHub Actions workflow，在 tag push 触发后复用共享测试验证套件，并采用与 Main CI 一致的不含 E2E 的简要测试范围。release 镜像发布 job 通过 `needs` 依赖 tag 校验和共享测试套件。

### Requirement: Release workflow 必须在发布门禁完成后创建 GitHub Release

系统 SHALL 在 release tag 校验、共享测试验证套件和 GHCR 镜像发布全部成功后，为触发 workflow 的 Git tag 创建 GitHub Release。Release 标题 SHALL 使用 `LinaPro Release <tag>` 格式。
