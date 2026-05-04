# Image Builder 工具

`image-builder`会根据仓库根目录配置文件`hack/config.yaml`构建生产`Docker`镜像。它是`make image`背后的跨平台执行层。

## 使用方式

推荐使用仓库封装入口：

```bash
make image
make image tag=v0.6.0
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1
```

也可以直接调用工具：

```bash
go run ./hack/tools/image-builder --tag=v0.6.0
go run ./hack/tools/image-builder --tag=v0.6.0 --registry=ghcr.io/linaproai --push=1
```

## 配置

默认值读取自`hack/config.yaml`中的`image`配置段。

| 字段 | 说明 |
|------|------|
| `name` | 不带远端仓库前缀的`Docker`镜像名称。 |
| `tag` | 可选默认标签。为空时通过`git describe`自动推导。 |
| `registry` | 可选远端仓库前缀，例如`ghcr.io/linaproai`。 |
| `push` | 默认推送行为。 |
| `baseImage` | 传递给`Dockerfile`的运行时基础镜像。 |
| `os`/`arch`/`platform` | 目标二进制与`Docker`镜像平台。`auto`表示跟随本机`Go`架构。 |
| `dockerfile` | 相对于仓库根目录的`Dockerfile`路径，默认是`hack/docker/Dockerfile`。 |
| `outputDir`/`binaryName` | 相对于仓库根目录的镜像构建产物位置。 |

命令行参数会覆盖本次调用的配置文件默认值。未通过配置或`registry=...`设置远端仓库时，也可以使用`LINAPRO_IMAGE_REGISTRY`提供仓库前缀。

`apps/lina-core`、`apps/lina-vben`、`apps/lina-plugins`、前端嵌入资源、宿主`manifest`嵌入资源以及`build-wasm`工具路径属于项目结构约定，因此不会暴露到`hack/config.yaml`中。

## 输出

- 前端生产资源会复制到宿主嵌入资源目录。
- 宿主`manifest`资源会被准备到嵌入目录，且不会嵌入本地`config.yaml`。
- 动态插件`Wasm`产物会写入配置的输出目录。
- 宿主二进制会以`CGO_ENABLED=0`编译到配置的目标平台。
- `Docker`会构建`<registry-prefix>/<name>:<tag>`，只有`push=true`时才推送。
