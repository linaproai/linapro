## ADDED Requirements

### Requirement: 发布构建命令必须支持跨平台宿主二进制构建

系统 SHALL 允许运维人员通过 `make build` 构建指定目标平台的宿主二进制。目标平台 SHALL 由 `hack/config.yaml` 中的 `build.platforms` 数组统一配置，每个数组项 SHALL 支持 `<goos>/<goarch>` 形式，也 SHALL 支持 `auto`；`auto` SHALL 按当前执行环境的 `runtime.GOOS/runtime.GOARCH` 自动解析为一个目标平台。命令行 SHALL 支持通过 `platforms=<goos>/<goarch>[,<goos>/<goarch>...]` 覆盖配置文件。多平台构建时，前端生产资源、宿主 `manifest` 嵌入资源与动态插件 `Wasm` 产物 SHALL 只准备一次；宿主后端二进制 SHALL 按目标平台分别交叉编译并输出到可区分的平台目录。

#### Scenario: 构建单一非本机架构宿主二进制
- **当** 运维人员运行 `make build platforms=linux/arm64`
- **则** 构建流程使用 `GOOS=linux` 与 `GOARCH=arm64` 编译宿主后端二进制
- **且** 构建流程产出标准宿主二进制文件

#### Scenario: 使用 platforms 构建多平台宿主二进制
- **当** 运维人员运行 `make build platforms=linux/amd64,linux/arm64`
- **则** 构建流程分别使用 `GOOS=linux GOARCH=amd64` 与 `GOOS=linux GOARCH=arm64` 编译宿主后端二进制
- **且** 每个平台的二进制输出路径互不覆盖

#### Scenario: 使用 auto 构建当前执行环境平台宿主二进制
- **当** 运维人员运行 `make build platforms=auto`
- **则** 构建流程将 `auto` 解析为当前执行环境的 `runtime.GOOS/runtime.GOARCH`
- **且** 构建流程使用解析后的 `GOOS` 与 `GOARCH` 编译宿主后端二进制

### Requirement: 镜像构建命令必须支持多架构 Docker 镜像构建与推送

系统 SHALL 允许运维人员通过 `make image platforms=linux/amd64,linux/arm64 registry=<registry> push=1` 构建并推送多架构 `Docker` 镜像。多架构镜像构建 SHALL 使用每个目标平台对应的宿主二进制，而不是将单个平台二进制复用于所有镜像平台。多平台镜像推送 SHALL 通过 `Docker buildx` 生成并推送远端多架构 manifest。未启用推送时，多平台镜像构建 SHALL 快速失败并提示必须设置 `push=1`，避免只写入本地构建缓存而没有可用镜像产物。

#### Scenario: 构建并推送 amd64 与 arm64 多架构镜像
- **当** 运维人员运行 `make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v1.0.0 push=1`
- **则** 构建流程先分别准备 `linux/amd64` 与 `linux/arm64` 宿主二进制
- **且** `Dockerfile` 在每个平台构建上下文中选择对应平台的宿主二进制
- **且** 镜像构建流程通过 `Docker buildx` 推送 `ghcr.io/linaproai/linapro:v1.0.0` 的多架构 manifest

#### Scenario: 未启用 push 的多平台镜像构建快速失败
- **当** 运维人员运行 `make image platforms=linux/amd64,linux/arm64 push=0`
- **则** 镜像构建流程在调用 `Docker buildx` 前失败
- **且** 错误消息说明多平台镜像构建需要 `push=1`

### Requirement: 发布构建命令必须支持自定义构建配置

系统 SHALL 允许运维人员通过 `make build config=<path>` 与 `make image config=<path>` 指定仓库相对路径的 image-builder 配置文件，用于替代默认 `hack/config.yaml`。该配置文件 SHALL 使用与 `hack/config.yaml` 相同的 `build` 与 `image` 结构，允许通过 `image.dockerfile` 指定自定义 Dockerfile。未指定 `config` 时，系统 SHALL 继续读取 `hack/config.yaml`。

#### Scenario: 使用自定义构建配置构建宿主二进制
- **当** 运维人员运行 `make build config=.github/workflows/nightly-build/config.yaml`
- **则** 构建流程使用该配置文件中的 `build.platforms`、`build.outputDir` 与 `build.binaryName`

#### Scenario: 使用自定义运行时配置构建镜像
- **当** 运维人员运行 `make image config=.github/workflows/nightly-build/config.yaml platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=nightly-20260507 push=1`
- **则** 镜像构建流程使用该配置文件中的 `image.dockerfile`
- **且** 自定义 Dockerfile 可决定容器内 `/app/config.yaml` 的来源

### Requirement: Nightly workflow 必须发布 GHCR 多架构镜像

系统 SHALL 提供 GitHub Actions nightly build workflow，每天凌晨自动运行一次，使用仓库标准 `make image` 入口构建 `linux/amd64` 与 `linux/arm64` 宿主二进制并发布多架构 `Docker` 镜像到 `ghcr.io`。workflow SHALL 使用 `Docker buildx` 发布远端多架构 manifest，镜像标签 SHALL 至少包含按日期生成的不可变 nightly 标签和浮动 `nightly` 标签。workflow SHALL 允许手动触发，以便发布链路可在需要时重新执行。

#### Scenario: nightly 定时发布多架构镜像
- **当** GitHub Actions nightly build workflow 按计划触发
- **则** workflow 安装前端、Go 与 Docker buildx 构建环境
- **且** workflow 运行 `make image config=.github/workflows/nightly-build/config.yaml platforms=linux/amd64,linux/arm64 registry=ghcr.io/<owner> image=linapro tag=nightly-<yyyymmdd> push=1`
- **且** workflow 将同一多架构 manifest 同步标记为 `ghcr.io/<owner>/linapro:nightly`

#### Scenario: 手动触发 nightly 发布链路
- **当** 运维人员通过 GitHub Actions 手动触发 nightly build workflow
- **则** workflow 使用与定时任务相同的多架构构建与发布流程
- **且** 发布完成后可通过 `docker buildx imagetools inspect` 验证远端 manifest 包含 `linux/amd64` 与 `linux/arm64`
