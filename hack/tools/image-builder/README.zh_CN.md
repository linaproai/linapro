# Image Builder 工具

`image-builder`会基于标准`make build`产物和仓库根目录配置文件`hack/config.yaml`构建生产`Docker`镜像。它是`make image`背后的`Docker`执行层。

## 使用方式

推荐使用仓库封装入口：

```bash
make image
make image tag=v0.6.0
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1
make image platforms=linux/amd64
make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v0.6.0 push=1
make image config=.github/workflows/nightly-build/config.yaml
```

也可以直接调用工具：

```bash
make build
go run ./hack/tools/image-builder --tag=v0.6.0
go run ./hack/tools/image-builder --tag=v0.6.0 --registry=ghcr.io/linaproai --push=1
```

跨平台宿主二进制可通过`make build platforms=linux/arm64`构建。多平台镜像发布使用`Docker buildx`，例如：

```bash
make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v0.6.0 push=1
```

## 配置

构建默认值读取自`hack/config.yaml`中的`build`配置段。

| 字段 | 说明 |
|------|------|
| `platforms` | 宿主二进制与`Docker`镜像平台列表。每个 YAML 数组项使用`goos/goarch`格式，或使用`auto`表示当前系统的`runtime.GOOS/runtime.GOARCH`。命令行覆盖使用英文逗号分隔的`platforms=...`值。 |
| `cgoEnabled` | `make build`构建宿主二进制时是否启用`CGO`。 |
| `outputDir`/`binaryName` | 相对于仓库根目录的标准`make build`产物位置。 |

镜像元数据默认值读取自`hack/config.yaml`中的`image`配置段。

| 字段 | 说明 |
|------|------|
| `name` | 不带远端仓库前缀的`Docker`镜像名称。 |
| `tag` | 可选默认标签。为空时通过`git describe`自动推导。 |
| `registry` | 可选远端仓库前缀，例如`ghcr.io/linaproai`。 |
| `push` | 默认推送行为。 |
| `baseImage` | 传递给`Dockerfile`的运行时基础镜像。 |
| `dockerfile` | 相对于仓库根目录的`Dockerfile`路径，默认是`hack/docker/Dockerfile`。 |

命令行参数会覆盖本次调用的配置文件默认值。`make build`或`make image`可使用`config=<path>`选择仓库级 image-builder 配置文件而不是`hack/config.yaml`。未通过配置或`registry=...`设置远端仓库时，也可以使用`LINAPRO_IMAGE_REGISTRY`提供仓库前缀。

`apps/lina-core`、`apps/lina-vben`、`apps/lina-plugins`、前端嵌入资源、宿主`manifest`嵌入资源以及`build-wasm`工具路径属于项目结构约定，因此不会暴露到`hack/config.yaml`中。

## 输出

- `make build`会把前端生产资源复制到宿主嵌入资源目录。
- `make build`会把宿主`manifest`资源准备到嵌入目录，且不会嵌入本地`config.yaml`。
- `make build`会把动态插件`Wasm`产物写入配置的构建输出目录。
- `make build`会按配置的目标平台编译宿主二进制。
- `make build platforms=linux/amd64,linux/arm64`会分别写入`temp/output/linux_amd64/lina`和`temp/output/linux_arm64/lina`。
- `make image`会把标准宿主二进制 staged 到`Docker`构建上下文，而不是重新构建。
- `make image config=<path>`会使用指定的 image-builder 配置文件而不是`hack/config.yaml`。
- 单平台镜像使用`docker build`，多平台镜像使用`docker buildx build --push`。
- `Docker`会构建`<registry-prefix>/<name>:<tag>`，只有`push=true`时才推送。多平台构建必须启用`push=true`，以便发布远端 manifest。
