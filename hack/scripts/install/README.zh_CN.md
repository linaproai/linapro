# LinaPro 安装器

本目录维护 `LinaPro` 单一安装入口的仓库内实现：

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

托管的 `/install.sh` 内容必须与 `hack/scripts/install/bootstrap.sh` 保持一致。`bootstrap.sh` 是完全自包含的单文件入口：负责解析目标版本、克隆指定标签，并分发到克隆仓库内的平台脚本。

## 支持平台

| 平台 | 运行环境 |
| --- | --- |
| `macOS` | Darwin 上的 `bash` |
| `Linux` | Linux 发行版与 WSL 上的 `bash` |
| `Windows` | 仅支持 Git Bash 或 WSL |

Windows 用户必须在 Git Bash 或 WSL 中执行安装命令。原生 PowerShell 与 `cmd.exe` 不再作为安装入口维护。

## 目录结构

```text
hack/scripts/install/
  bootstrap.sh          托管的 curl|bash 入口
  install-macos.sh      macOS 克隆后安装脚本
  install-linux.sh      Linux 与 WSL 克隆后安装脚本
  install-windows.sh    Windows Git Bash 克隆后安装脚本
  checks/prereq.sh      共享前置检查
  lib/_common.sh        共享安装辅助函数
  README.md             英文说明
  README.zh_CN.md       简体中文镜像
```

## 环境变量

| 变量 | 默认值 | 含义 | 示例 |
| --- | --- | --- | --- |
| `LINAPRO_VERSION` | GitHub 最新稳定发布版本 | 要克隆的目标版本标签。若无法自动解析标签，安装器会失败。 | `LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_DIR` | `./linapro` | 克隆项目的目标目录。 | `LINAPRO_DIR=~/Workspace/my-linapro curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_NON_INTERACTIVE` | 未设置 | 平台脚本需要确认时跳过交互并使用默认行为。 | `LINAPRO_NON_INTERACTIVE=1 ...` |
| `LINAPRO_SKIP_MOCK` | 未设置 | 执行 `make init`，但跳过 `make mock`。 | `LINAPRO_SKIP_MOCK=1 ...` |
| `LINAPRO_SHALLOW` | 未设置 | 使用 `git clone --depth 1`。后续第一次升级需要先执行 `git fetch --unshallow`。 | `LINAPRO_SHALLOW=1 ...` |

`LINAPRO_FORCE=1` 是隐藏恢复开关，仅用于明确要替换非空目标目录的场景。

## 本地等价命令

在已有仓库检出中，可以直接运行同一份本地入口：

```bash
bash hack/scripts/install/bootstrap.sh
```

该命令仍会把请求版本克隆到 `LINAPRO_DIR` 或 `./linapro`；除非显式设置 `LINAPRO_DIR`，否则不会覆盖当前检出目录。

## 安装器执行内容

1. 解析 `LINAPRO_VERSION`，或跟随 GitHub `releases/latest` 重定向获取最新稳定版本。
2. 目标目录非空时拒绝覆盖，除非设置 `LINAPRO_FORCE=1`。
3. 执行 `git clone --branch <tag> https://github.com/linaproai/linapro.git "$LINAPRO_DIR"`。
4. 分发到 `install-macos.sh`、`install-linux.sh` 或 `install-windows.sh`。
5. 检查 `go >= 1.22`、`node >= 20`、`pnpm >= 8`、`git`、`make`、MySQL 客户端，以及 `5666` / `8080` 端口。
6. 执行后端 `go mod download`、前端 `pnpm install`、`make init confirm=init`，并在未设置 `LINAPRO_SKIP_MOCK=1` 时执行 `make mock confirm=mock`。
7. 输出项目目录、默认 `admin` / `admin123` 账号，以及下一步 `make dev` 命令。

## 诊断与重试

- 如果最新发布版本解析失败，使用 `LINAPRO_VERSION=v0.x.y` 显式重试。
- 如果克隆失败，检查网络连通性，并确认所选标签存在于 GitHub Releases。
- 如果前置检查失败，按 `checks/prereq.sh` 打印的平台提示安装缺失工具，然后在已克隆仓库内重新执行平台脚本。
- 如果端口 `5666` 或 `8080` 被占用，请在运行 `make dev` 前停止冲突进程。
- 如果数据库初始化失败，检查 `apps/lina-core/manifest/config/config.yaml` 与 MySQL 连通性，然后重新执行 `make init confirm=init`。

## 部署到 linapro.ai

远程入口发布属于本仓库变更之外的运维任务。

1. `CI/CD` 将 `hack/scripts/install/bootstrap.sh` 复制到 `linapro.ai` CDN 路径 `/install.sh`。
2. `/install.ps1` 仅作为未来 PowerShell 入口的占位，目前不会通过该流程发布 PowerShell 安装器。
3. 每次发布新的 `LinaPro` 稳定标签后，都必须同步刷新 CDN 缓存。
4. 发布后在干净环境中验证：

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```
