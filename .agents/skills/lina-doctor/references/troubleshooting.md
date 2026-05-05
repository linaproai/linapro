# Lina Doctor 排障指南

## 安装通道快照

本节记录 `lina-doctor` 实施前验证过的外部安装通道,用于后续排障与 `tool-matrix.md` 对齐。

| 目标 | 探查命令 | 结果 | 结论 |
| --- | --- | --- | --- |
| `openspec` Homebrew | `brew info openspec` | `homebrew/core` formula 存在,stable `1.3.1`,依赖 `node` | macOS 主通道可用 |
| `gf` Homebrew | `HOMEBREW_NO_AUTO_UPDATE=1 brew info gf` | `homebrew/core` formula 存在,stable `2.10.0`,build dependency 为 `go` | macOS 可作为 `gf` 安装通道,但跨平台仍以 `go install` 为主 |
| `openspec` npm 旧包 | `npm view openspec` | 包存在但版本为 `0.0.0`,仓库指向旧 `openspecio/openspec` | 不得作为 OpenSpec CLI 安装通道 |
| `openspec` npm 官方包 | `npm view @fission-ai/openspec` | 版本 `1.3.1`,bin 为 `openspec`,要求 Node `>=20.19.0` | npm 回落通道使用 `npm i -g @fission-ai/openspec@latest` |
| `skills` npm 包 | `npm view skills` | latest `1.5.3`,bin 包含 `skills` 与 `add-skill`,要求 Node `>=18` | `npx skills add github.com/gogf/skills -g` 的 npm 入口可用 |
| `gogf/skills` GitHub 仓库 | `curl https://api.github.com/repos/gogf/skills` | 仓库公开可访问,默认分支 `main`,未归档,主题包含 `goframe` | `goframe-v2` 上游仓库可用 |

## 上游通道失效

若 `npx skills add github.com/gogf/skills -g` 失败,先查看 `/tmp/lina-doctor-goframe-v2.log` 中的末尾输出并按根因处理。

- `network`: 检查 GitHub 与 npm registry 访问,必要时设置 npm registry 镜像后重试。
- `permission`: 检查 npm 全局 prefix 是否需要 root 权限,优先配置用户级 prefix。
- `package_not_found`: 确认 `skills` npm 包仍提供 `skills` bin,再确认 `github.com/gogf/skills` 仓库仍可访问。
- `unknown`: 人工检查 `skills` CLI 是否调整了命令签名。

人工兜底原则: 最终验证标准是 `~/.claude/skills/goframe-v2/SKILL.md` 存在,安装通道可以替换,但不得把 `goframe-v2` 写回 LinaPro 项目仓库。

## 根因类别

| 根因 | 常见关键词 | 处理动作 |
| --- | --- | --- |
| `network` | `timeout`、`Could not resolve host`、`proxy`、`i/o timeout` | 检查代理、DNS、registry 和镜像变量 |
| `permission` | `EACCES`、`permission denied`、`must be run as root` | 使用用户级 prefix，或在确认后使用必要的提升权限 |
| `package_not_found` | `Unable to locate package`、`formula was not found`、`not in registry` | 检查包名和通道，按`tool-matrix.md`改用备用命令 |
| `shim_conflict` | `already installed`、`conflicts with`、`node command in PATH points to` | 检查 PATH 和版本管理器 shim |
| `unknown` | 未匹配 | 查看日志尾部并手动复现命令 |

## Node Shim 冲突处理

如果检测到`nvm`、`fnm`或`volta`，优先使用对应工具安装 Node。不要同时用 OS 包管理器覆盖 Node，否则`node --version`可能仍解析到旧 shim。

建议排查：

```bash
command -v node
node --version
command -v nvm || true
command -v fnm || true
command -v volta || true
```

## 镜像建议

`lina-doctor`只提示镜像，不自动写入环境变量。

| 变量 | 上游默认 | 中国大陆常用镜像 |
| --- | --- | --- |
| `GOPROXY` | `https://proxy.golang.org,direct` | `https://goproxy.cn,direct` |
| npm registry | `https://registry.npmjs.org/` | `https://registry.npmmirror.com` |
| `PLAYWRIGHT_DOWNLOAD_HOST` | Playwright 默认 CDN | `https://npmmirror.com/mirrors/playwright/` |

## 常见平台问题

- `winget`静默失败：复制 PowerShell wrapper 命令到原生 PowerShell 中重试，查看完整错误。
- `scoop`缓存损坏：运行`scoop cache rm *`后重试。
- `apt`源失效：先运行`sudo apt-get update`确认源可用。
- `gf`安装后找不到：确认`$HOME/go/bin`是否在`PATH`，按`path-and-shell.md`追加。
- `Playwright browsers`下载卡住：设置`PLAYWRIGHT_DOWNLOAD_HOST`后重新执行可选步骤。
