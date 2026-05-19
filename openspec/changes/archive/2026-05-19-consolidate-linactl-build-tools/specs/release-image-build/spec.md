## MODIFIED Requirements

### Requirement:镜像构建命令必须支持多架构 Docker 镜像构建与推送

系统 SHALL 允许运维人员通过`make image platforms=linux/amd64,linux/arm64 registry=<registry> push=1`构建并推送多架构`Docker`镜像。多架构镜像构建 SHALL 使用每个目标平台对应的宿主二进制，而不是将单个平台二进制复用于所有镜像平台。多平台镜像推送 SHALL 通过`Docker buildx`生成并推送远端多架构 manifest。未启用推送时，多平台镜像构建 SHALL 快速失败并提示必须设置`push=1`，避免只写入本地构建缓存而没有可用镜像产物。镜像构建实现 SHALL 由`linactl`内部镜像构建组件执行，而不是依赖独立`hack/tools/image-builder`工具模块。

#### Scenario:构建并推送 amd64 与 arm64 多架构镜像

- **当** 运维人员运行`make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v1.0.0 push=1`
- **则** 构建流程先分别准备`linux/amd64`与`linux/arm64`宿主二进制
- **且**`Dockerfile`在每个平台构建上下文中选择对应平台的宿主二进制
- **且** 镜像构建流程通过`Docker buildx`推送`ghcr.io/linaproai/linapro:v1.0.0`的多架构 manifest

#### Scenario:未启用 push 的多平台镜像构建快速失败

- **当** 运维人员运行`make image platforms=linux/amd64,linux/arm64 push=0`
- **则** 镜像构建流程在调用`Docker buildx`前失败
- **且** 错误消息说明多平台镜像构建需要`push=1`

#### Scenario:镜像命令不再调用独立 image-builder 工具

- **当** 运维人员运行`linactl image`或`linactl image.build`
- **则** 命令通过`linactl`内部镜像构建组件完成镜像构建或 staging
- **且** 命令不得再执行`go run ./hack/tools/image-builder`
