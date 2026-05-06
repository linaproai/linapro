## ADDED Requirements

### Requirement: 仓库构建命令必须支持配置化目标平台
系统 SHALL 允许维护者在仓库根配置文件 `hack/config.yaml` 的 `build` 配置段中声明 `make build` 的目标操作系统、目标架构、CGO 开关、输出目录和二进制文件名。`make build` 必须按该配置构建宿主二进制，并继续把前端静态资源、宿主 manifest 资源和动态插件 Wasm 产物收敛到同一个构建输出目录。命令行一次性覆盖参数可以覆盖配置文件中的目标操作系统、目标架构和输出路径。

`image` 配置段 SHALL 只保留 Docker 镜像仓库、标签、registry、push、运行时基础镜像和 Dockerfile 等镜像元数据，不得再保存宿主二进制的目标操作系统、目标架构、输出目录或二进制文件名。镜像构建流程必须直接消费 `make build` 生成的宿主二进制，而不是在 image 工具内部重新执行前端、Wasm 或宿主后端构建。

#### Scenario: make build 使用配置的目标平台
- **WHEN** `hack/config.yaml` 的 `build.os` 为 `linux` 且 `build.arch` 为 `amd64`
- **AND** 用户在仓库根执行 `make build`
- **THEN** 构建命令使用 `GOOS=linux` 与 `GOARCH=amd64` 编译宿主二进制
- **AND** 宿主二进制写入 `build.outputDir/build.binaryName` 指定的位置

#### Scenario: 镜像构建复用标准构建产物
- **WHEN** 用户执行 `make image`
- **THEN** 镜像流程先确保 `make build` 产物存在
- **AND** image 工具把该二进制作为 Docker 构建上下文输入
- **AND** image 工具不再自行执行 `pnpm run build`、`make wasm` 或 `go build`

#### Scenario: image 配置不再包含二进制构建参数
- **WHEN** 维护者查看 `hack/config.yaml`
- **THEN** `image` 配置段不包含 `os`、`arch`、`platform`、`outputDir` 或 `binaryName`
- **AND** 这些构建参数统一在 `build` 配置段维护
