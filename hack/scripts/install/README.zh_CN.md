# LinaPro 安装器

本目录维护`LinaPro`单一源码下载入口的仓库内实现：

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

托管的`/install.sh`内容必须与`hack/scripts/install/install.sh`保持一致。该脚本完全自包含，只负责下载指定版本的`LinaPro`仓库源码。运行环境检查与工具安装由`lina-doctor`技能负责。

## 支持平台

| 平台 | 运行环境 |
| --- | --- |
| `macOS` | Darwin 上的`bash` |
| `Linux` | Linux 发行版与 WSL 上的`bash` |
| `Windows` | Git Bash 或 WSL |

Windows 用户必须在 Git Bash 或 WSL 中执行安装命令。原生 PowerShell 与`cmd.exe`不是受支持的入口。

## 目录结构

```text
hack/scripts/install/
  install.sh          托管的 curl|bash 入口
  README.md             英文说明
  README.zh_CN.md       简体中文镜像
```

## 环境变量

| 变量 | 默认值 | 含义 | 示例 |
| --- | --- | --- | --- |
| `LINAPRO_VERSION` | `origin`上的最高稳定 Git 标签 | 要克隆的目标版本标签。若无法自动解析标签，安装器会失败。 | `LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_DIR` | `./linapro` | 克隆项目的目标目录。 | `LINAPRO_DIR=~/Workspace/my-linapro curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_SHALLOW` | 未设置 | 使用浅克隆。默认完整克隆更推荐，因为后续可以直接通过 Git 拉取新发布标签升级。 | `LINAPRO_SHALLOW=1 ...` |
| `LINAPRO_FORCE` | 未设置 | 在内置安全检查通过后，允许替换非空目标目录。 | `LINAPRO_FORCE=1 ...` |

`LINAPRO_NON_INTERACTIVE`和`LINAPRO_SKIP_MOCK`不再由安装器使用。脚本不会提示环境安装，也不会加载 mock 数据。

## 本地等价命令

在已有仓库检出中，可以直接运行同一份安装器源码：

```bash
bash hack/scripts/install/install.sh
```

该命令仍会把请求版本克隆到`LINAPRO_DIR`或`./linapro`；除非显式设置`LINAPRO_DIR`，否则不会覆盖当前检出目录。

## 安装器执行内容

1. 检测当前命令是否运行在受支持的`bash`平台。
2. 解析`LINAPRO_VERSION`，或从远程 Git 仓库选择最高的稳定`vX.Y.Z`标签。
3. 目标目录非空时拒绝覆盖，除非`LINAPRO_FORCE=1`通过安全检查。
4. 执行保留`origin`远程地址并拉取发布标签的 Git 克隆，便于后续按标签升级。
5. 检出选中的标签，并输出项目目录、默认`admin`/`admin123`账号与下一步指引。

## 克隆后的下一步

```bash
cd <project-dir>
# 让你的 AI 工具运行 lina-doctor 并修复缺失的开发工具。
make init && make dev
```

如果 Go、Node、pnpm、OpenSpec、GoFrame CLI、Playwright browsers 或`goframe-v2`技能可能缺失，请在项目初始化前先通过 AI 工具调用`lina-doctor`。

## 基于标签升级

默认安装会保留一个带`origin`远程地址的普通 Git 仓库。后续要把已安装目录切换到新的发布标签：

```bash
git fetch --tags --force origin
git checkout --detach <new-version-tag>
```

除非克隆体积比升级便利性更重要，否则不建议设置`LINAPRO_SHALLOW=1`。如果浅克隆检出无法切换到新标签，先执行一次`git fetch --unshallow --tags --force origin`，再检出新标签。

## 诊断与重试

- 如果最新标签解析失败，使用`LINAPRO_VERSION=v0.x.y`显式重试。
- 如果克隆失败，检查网络连通性，并确认所选标签存在于 GitHub Releases。
- 如果目标目录非空，请选择其他`LINAPRO_DIR`，或确认目标路径后使用`LINAPRO_FORCE=1`重试。
- 如果克隆后缺少开发工具，请通过 AI 工具调用`lina-doctor`技能。

## 部署到 linapro.ai

远程入口发布属于本仓库变更之外的运维任务。

1. `CI/CD`将`hack/scripts/install/install.sh`复制到`linapro.ai`CDN 路径`/install.sh`。
2. 每次`install.sh`变更后都必须刷新 CDN 缓存。
3. 发布后在干净环境中验证：

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```
