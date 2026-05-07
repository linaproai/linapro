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
