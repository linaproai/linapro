# 安装脚本说明

该目录承载 `LinaPro` 官方源码快速安装入口，用于从 `GitHub/Codeload` 下载仓库归档，并把项目安全解压到本地工作目录。

## 目录结构

```text
hack/scripts/install/
  install.sh       macOS 与 Linux 安装脚本
  install.ps1      Windows PowerShell 安装脚本
  test_install.py  安装脚本行为与参数契约 smoke test
  README.md        英文说明文档
  README.zh_CN.md  中文说明文档
```

## 设计目标

安装脚本只聚焦于一个有限的 `bootstrap` 流程：

- 从 `GitHub/Codeload` 下载源码归档
- 先解压到临时目录，再复制到最终目标目录
- 当调用方未传入 `--ref` 或 `-Ref` 时，自动解析最新稳定标签版本
- 默认保护非空目录，只有显式指定覆盖模式时才允许 `overlay` 安装
- 输出 `Go`、`Node.js`、`pnpm`、`MySQL`、`make` 的环境体检结果
- 输出 `make init confirm=init`、`make dev` 等后续推荐命令

脚本不会自动安装系统依赖，也不会自动执行数据库初始化或服务启动命令。

## 官方入口映射

| 平台 | 远程入口 | 仓库脚本 |
| --- | --- | --- |
| `macOS` / `Linux` | `curl -fsSL https://linapro.ai/install.sh \| bash` | `hack/scripts/install/install.sh` |
| `Windows PowerShell` | `irm https://linapro.ai/install.ps1 \| iex` | `hack/scripts/install/install.ps1` |

站点托管入口应当始终作为仓库脚本的薄包装或重定向，避免安装行为产生漂移。

默认情况下，用户不带参数执行安装时，脚本会优先解析目标仓库最新稳定语义化标签版本；只有在无法识别稳定标签时，才回退到 `main`。

## 使用方式

### `macOS` 与 `Linux`

```bash
bash ./hack/scripts/install/install.sh
bash ./hack/scripts/install/install.sh --ref v0.1.0 --dir ~/Workspace/linapro
bash ./hack/scripts/install/install.sh --current-dir --force
```

### `Windows PowerShell`

```powershell
.\hack\scripts\install\install.ps1
.\hack\scripts\install\install.ps1 -Ref v0.1.0 -Dir C:\Workspace\linapro
.\hack\scripts\install\install.ps1 -CurrentDir -Force
```

## 参数说明

| Shell | PowerShell | 说明 |
| --- | --- | --- |
| `--repo` | `-Repo` | 覆盖默认的 `GitHub` 仓库地址，例如 `owner/name`。 |
| `--ref` | `-Ref` | 指定要下载的分支、标签或提交引用。 |
| `--dir` | `-Dir` | 安装到显式指定的目标目录。 |
| `--name` | `-Name` | 在当前工作目录下新建一个子目录并安装进去。 |
| `--current-dir` | `-CurrentDir` | 直接把项目解压到当前目录。 |
| `--force` | `-Force` | 允许覆盖安装到非空目录。 |
| `--help` | `-Help` | 输出内置帮助信息。 |

## 本地归档覆盖

两个脚本都支持环境变量 `LINAPRO_INSTALL_ARCHIVE_PATH`。
它允许你在不访问网络的情况下，直接使用本地归档文件验证安装流程。
其中 `Shell` 脚本读取本地 `.tar.gz` 文件，`PowerShell` 脚本读取本地 `.zip` 文件。

两个脚本还支持 `LINAPRO_INSTALL_STABLE_REF`，可以显式覆盖自动解析得到的稳定标签版本，主要用于测试或受控包装脚本场景。

## 验证方式

更新安装脚本逻辑，或需要在本地与 `CI` 中复用同一条验证命令时，请执行：

```bash
make test-install
```

该目标会运行 `python3 hack/scripts/install/test_install.py`，当前覆盖以下场景：

- 使用本地归档安装到命名目录
- 使用当前目录模式完成安装
- 未指定 `--force` 时拒绝安装到非空目录
- 校验 `install.sh` 与 `install.ps1` 的参数契约一致性
