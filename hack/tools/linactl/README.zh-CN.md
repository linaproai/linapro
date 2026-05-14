# linactl

`linactl`是`LinaPro`的跨平台开发命令入口。它将仓库长期维护的任务编排放在`Go`工具中，确保`Windows`、`Linux`和`macOS`可以运行同一套命令，而不依赖`GNU Make`或`POSIX Shell`工具。

## 使用方式

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . prepare-packed-assets
go run . wasm p=plugin-demo-dynamic
go run . plugins.status
go run . i18n.check
go run . init confirm=init
go run . tidy
go run . build platforms=linux/amd64,linux/arm64
go run . release.tag.check tag=v0.2.0
go run . release.tag.check print-version=1
```

## Windows 入口

仓库根目录提供`make.cmd`作为`Windows`薄包装入口：

```cmd
make.cmd help
make.cmd status
make.cmd plugins.status
make.cmd i18n.check
make.cmd init confirm=init
make.cmd tidy
make.cmd release.tag.check tag=v0.2.0
```

在`PowerShell`中，需要显式添加当前目录前缀：

```powershell
.\make.cmd help
.\make.cmd status
.\make.cmd i18n.check
.\make.cmd release.tag.check tag=v0.2.0
```

## 参数

`linactl`支持现有`make`风格的`key=value`参数，降低命令迁移成本。

| 参数 | 示例 | 用途 |
|------|------|------|
| `confirm` | `confirm=init` | 确认高风险初始化命令。 |
| `rebuild` | `rebuild=true` | 在`init`时重建配置中的数据库。 |
| `platforms` | `platforms=linux/amd64,linux/arm64` | 指定构建目标平台。 |
| `plugins` | `plugins=0` | 覆盖构建、开发、镜像和 Go 测试命令的自动插件完整模式探测。 |
| `tag` | `tag=v0.2.0` | 指定 `release.tag.check` 校验的 release tag。 |
| `print-version` | `print-version=1` | 输出已校验的 `framework.version`，供发布自动化使用。 |
| `p` | `p=multi-tenant` | 为 Wasm 构建或插件工作区管理命令选择单个插件。 |
| `source` | `source=official` | 为插件工作区管理命令选择单个已配置来源。 |
| `force` | `force=1` | 允许插件安装或更新命令覆盖已存在或存在本地改动的插件目录。 |
| `verbose` | `verbose=1` | 构建任务展示子命令输出。 |

未传入`plugins`时，构建和开发命令会在`apps/lina-plugins`存在插件清单时启用插件完整模式。插件完整模式会基于宿主专用的根目录`go.work`生成或刷新已忽略的`temp/go.work.plugins`，并通过`GOWORK`解析源码插件`Go`模块。

## Release Tag 校验

`release.tag.check` 会读取 `apps/lina-core/manifest/config/metadata.yaml`，并校验 release tag 与 `framework.version` 完全一致。

```bash
make.cmd release.tag.check tag=v0.2.0
make release.tag.check tag=v0.2.0
make release.tag.check metadata=apps/lina-core/manifest/config/metadata.yaml tag=v0.2.0
```

在 GitHub Actions 中，如果未传入 `tag`，该命令也会使用 `GITHUB_REF_NAME` 作为待校验标签。

## 插件工作区命令

插件工作区管理始终使用固定目录 `apps/lina-plugins`。在 `hack/config.yaml` 中配置来源：

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - "multi-tenant"
        - "org-center"
```

`items` 只接受插件 ID 字符串。使用带引号的 `"*"` 可安装 source `root` 下一层的全部插件目录；不要写裸的 `- *`，因为 YAML 会把它当作 alias 语法。如果同一仓库中的插件需要不同 `ref`，应拆成多个 source。

常用命令：

```bash
make plugins.init
make plugins.install
make plugins.install p=multi-tenant
make plugins.update source=official
make plugins.update force=1
make plugins.status
```

`plugins.init` 会将 `apps/lina-plugins` 从 `submodule` 转成普通目录并保留文件。`plugins.install` 和 `plugins.update` 会把配置来源克隆到 `temp/`，再复制插件目录到 `apps/lina-plugins/<plugin-id>`，并更新工具生成的 `apps/lina-plugins/.linapro-plugins.lock.yaml` 锁文件。

## 验证

```bash
cd hack/tools/linactl
go test ./...
go run . help
go run . wasm dry-run=true
go run . plugins.status
go run . release.tag.check tag=v0.2.0
```
