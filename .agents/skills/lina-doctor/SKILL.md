---
name: lina-doctor
description: "LinaPro 环境检查与依赖安装技能, covers environment, install, fix, lina-doctor, dependencies, toolchain"
---

# Lina Doctor 环境检查与依赖安装

**交互语言**：与用户交互的内容语言以用户上下文使用的语言为准，用户使用英文则使用英文，用户使用中文则使用中文。

## 用途

`lina-doctor`用于在已经克隆的`LinaPro`仓库中诊断和修复开发态工具链。它负责检查并按用户确认安装 Go、Node、pnpm、Git、Make、OpenSpec、GoFrame CLI、Playwright browsers 和`goframe-v2`技能。

`lina-doctor`不负责下载`LinaPro`源码，不替代`hack/scripts/install/bootstrap.sh`，也不执行`make init`、`make mock`或数据库初始化。

## 触发场景

当用户表达以下意图时调用本技能：

- 检查`LinaPro`环境、环境检查、安装依赖、修复工具链。
- 缺少`go`、`node`、`pnpm`、`openspec`、`gf`、`Playwright browsers`或`goframe-v2`。
- `environment check`、`install dependencies`、`fix environment`、`setup linapro environment`、`lina-doctor`。

如果用户要求安装 MySQL、Docker、VS Code、Claude Code 或 Codex，不要尝试安装；说明这些工具不属于本技能范围。

## 输入与环境变量

开始前读取并向用户说明以下输入：

- `--check-only`：只运行诊断，不生成安装计划，不执行安装。
- `LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1`：跳过 Playwright browsers 安装计划。
- `LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1`：跳过`goframe-v2`技能安装计划。
- `LINAPRO_DOCTOR_NON_INTERACTIVE=1`：跳过逐项确认，但仍必须打印每条命令。
- `LINAPRO_DOCTOR_TIMEOUT`：单步安装超时时间，默认`300s`。
- `LINAPRO_DOCTOR_DEBUG=1`：启用 shell 调试输出。

## 工作流

1. 确认当前目录是`LinaPro`仓库根目录；若不是，提示用户先进入克隆后的项目目录。
2. 读取`LINAPRO_DOCTOR_*`环境变量，识别`--check-only`、跳过项、非交互模式和单步超时。
3. 执行`scripts/doctor-detect.sh`，采集 OS、包管理器、shell、Node 版本管理器、镜像变量和仓库根状态。
4. 执行`scripts/doctor-check.sh`，输出严格 JSON 诊断结果。
5. 若`--check-only`被启用，直接返回诊断 JSON 和人工建议，不生成安装计划，不执行安装。
6. 基于诊断 JSON 执行`scripts/doctor-plan.sh`，按拓扑序生成结构化安装计划，并展示镜像建议。
7. 向用户展示每个计划项的命令、包管理器、`sudo`/PowerShell 需求、关键/可选属性和跳过原因。
8. 执行`scripts/doctor-install.sh`，按计划逐项确认、安装、复检；关键工具失败即停止，可选目标失败记录非阻塞 escalation。
9. 执行`scripts/doctor-verify.sh`，汇总关键工具 smoke 结果、可选目标状态、PATH 修复提示和镜像建议。
10. 输出最终报告：已满足工具、安装成功项、跳过项、非阻塞失败项、下一步`make init`、`make dev`、`openspec list`和`pnpm test`。

## 输出

技能执行时必须产出以下信息：

- 诊断 JSON：来自`doctor-check.sh`，用于 AI 工具消费。
- 安装计划：来自`doctor-plan.sh`，列出待执行命令和依赖关系。
- 最终报告：来自`doctor-verify.sh`，列出已满足、已安装、已跳过和非阻塞失败项。
- Escalation 报告：安装或复检失败时输出失败工具、命令、包管理器、日志尾部、根因和人工动作。

## 参考文档

- `references/tool-matrix.md`：工具与平台安装命令矩阵。
- `references/install-strategy.md`：拓扑顺序、包管理器选择和执行策略。
- `references/path-and-shell.md`：PATH 修复与 shell 配置建议。
- `references/troubleshooting.md`：常见失败根因与排障动作。

## 可用工具

使用`Bash`、`Read`、`Edit`、`Grep`和`Glob`完成诊断、脚本调用、文件读取和小范围修复。
