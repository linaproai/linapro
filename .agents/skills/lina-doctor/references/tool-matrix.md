# Lina Doctor 工具矩阵

本文档是`lina-doctor`生成安装命令的权威来源。命令名、包名和环境变量保持英文原文。

## 版本下限

| 工具 | 下限 | 说明 |
| --- | --- | --- |
| `go` | `1.22.0` | 后端构建与`gf`安装需要 |
| `node` | `20.19.0` | `@fission-ai/openspec`要求 Node `>=20.19.0` |
| `pnpm` | `8.0.0` | 前端与 E2E 测试需要 |
| `git` | 存在即可 | 源码管理与升级需要 |
| `make` | 存在即可 | 项目根命令入口 |
| `openspec` | 存在即可 | SDD 工作流 CLI |
| `gf` | 存在即可 | GoFrame 代码生成 CLI |
| `Playwright browsers` | 存在即可 | E2E 浏览器运行时 |
| `goframe-v2` | 存在即可 | 全局 AI 技能，检查`~/.claude/skills/goframe-v2/SKILL.md` |

## macOS

| 工具 | 命令 | 备注 |
| --- | --- | --- |
| `go` | `brew install go` | 使用 Homebrew |
| `node` | `brew install node` | 若检测到`nvm`、`fnm`或`volta`，改用版本管理器 |
| `git` | `brew install git` | Xcode Command Line Tools 也可能提供 |
| `make` | `brew install make` | 系统也可能已提供 |
| `openspec` | `brew install openspec` | 主通道，Homebrew formula 为`1.3.1` |
| `openspec` fallback | `npm i -g @fission-ai/openspec@latest` | `openspec@0.0.0`不得使用 |
| `gf` | `go install github.com/gogf/gf/v2/cmd/gf@latest` | Homebrew `gf` formula 存在，但跨平台默认用 Go 安装 |
| `pnpm` | `npm i -g pnpm` | 依赖 Node |
| `Playwright browsers` | `cd hack/tests && pnpm exec playwright install` | 可选目标 |
| `goframe-v2` | `npx skills add github.com/gogf/skills -g` | 可选目标，安装到用户全局技能目录 |

## Linux / WSL

| 包管理器 | `go` | `node` | `git` | `make` |
| --- | --- | --- | --- | --- |
| `apt-get` | `sudo apt-get update && sudo apt-get install -y golang-go` | `sudo apt-get update && sudo apt-get install -y nodejs npm` | `sudo apt-get update && sudo apt-get install -y git` | `sudo apt-get update && sudo apt-get install -y make` |
| `dnf` | `sudo dnf install -y golang` | `sudo dnf install -y nodejs npm` | `sudo dnf install -y git` | `sudo dnf install -y make` |
| `yum` | `sudo yum install -y golang` | `sudo yum install -y nodejs npm` | `sudo yum install -y git` | `sudo yum install -y make` |
| `pacman` | `sudo pacman -Sy --needed go` | `sudo pacman -Sy --needed nodejs npm` | `sudo pacman -Sy --needed git` | `sudo pacman -Sy --needed make` |

Linux 和 WSL 默认安装到当前 Linux 环境。只有用户明确要求 Windows 主机工具链时，才进入 Windows PowerShell 包装路径。

| 工具 | 命令 | 备注 |
| --- | --- | --- |
| `openspec` | `npm i -g @fission-ai/openspec@latest` | 依赖 Node `>=20.19.0` |
| `gf` | `go install github.com/gogf/gf/v2/cmd/gf@latest` | 依赖 Go |
| `pnpm` | `npm i -g pnpm` | 依赖 Node |
| `Playwright browsers` | `cd hack/tests && pnpm exec playwright install` | 可选目标 |
| `goframe-v2` | `npx skills add github.com/gogf/skills -g` | 可选目标 |

## Windows Git Bash

Windows 包管理器命令统一通过`powershell.exe -NoProfile -Command "..."`包装。

| 包管理器 | `go` | `node` | `git` | `make` |
| --- | --- | --- | --- | --- |
| `winget` | `powershell.exe -NoProfile -Command "winget install GoLang.Go"` | `powershell.exe -NoProfile -Command "winget install OpenJS.NodeJS"` | `powershell.exe -NoProfile -Command "winget install Git.Git"` | `powershell.exe -NoProfile -Command "winget install ezwinports.make"` |
| `scoop` | `powershell.exe -NoProfile -Command "scoop install go"` | `powershell.exe -NoProfile -Command "scoop install nodejs-lts"` | `powershell.exe -NoProfile -Command "scoop install git"` | `powershell.exe -NoProfile -Command "scoop install make"` |
| `choco` | `powershell.exe -NoProfile -Command "choco install golang -y"` | `powershell.exe -NoProfile -Command "choco install nodejs-lts -y"` | `powershell.exe -NoProfile -Command "choco install git -y"` | `powershell.exe -NoProfile -Command "choco install make -y"` |

| 工具 | 命令 |
| --- | --- |
| `openspec` | `npm i -g @fission-ai/openspec@latest` |
| `gf` | `go install github.com/gogf/gf/v2/cmd/gf@latest` |
| `pnpm` | `npm i -g pnpm` |
| `Playwright browsers` | `cd hack/tests && pnpm exec playwright install` |
| `goframe-v2` | `npx skills add github.com/gogf/skills -g` |
