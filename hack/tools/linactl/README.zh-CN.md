# linactl

`linactl`是`LinaPro`的跨平台开发命令入口。它将仓库长期维护的任务编排放在`Go`工具中，确保`Windows`、`Linux`和`macOS`可以运行同一套命令，而不依赖`GNU Make`或`POSIX Shell`工具。

## 使用方式

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . prepare-packed-assets
go run . wasm p=plugin-demo-dynamic
go run . init confirm=init
go run . tidy
go run . build platforms=linux/amd64,linux/arm64
```

## Windows 入口

仓库根目录提供`make.cmd`作为`Windows`薄包装入口：

```cmd
make.cmd help
make.cmd status
make.cmd init confirm=init
make.cmd tidy
```

在`PowerShell`中，需要显式添加当前目录前缀：

```powershell
.\make.cmd help
.\make.cmd status
```

## 参数

`linactl`支持现有`make`风格的`key=value`参数，降低命令迁移成本。

| 参数 | 示例 | 用途 |
|------|------|------|
| `confirm` | `confirm=init` | 确认高风险初始化命令。 |
| `rebuild` | `rebuild=true` | 在`init`时重建配置中的数据库。 |
| `platforms` | `platforms=linux/amd64,linux/arm64` | 指定构建目标平台。 |
| `plugins` | `plugins=0` | 覆盖构建、开发、镜像和 Go 测试命令的自动插件完整模式探测。 |
| `p` | `p=plugin-demo-dynamic` | 构建指定动态插件。 |
| `verbose` | `verbose=1` | 构建任务展示子命令输出。 |

未传入`plugins`时，构建和开发命令会在`apps/lina-plugins`存在插件清单时启用插件完整模式。插件完整模式会基于宿主专用的根目录`go.work`生成或刷新已忽略的`temp/go.work.plugins`，并通过`GOWORK`解析源码插件`Go`模块。

## 验证

```bash
cd hack/tools/linactl
go test ./...
go run . help
go run . wasm dry-run=true
```
