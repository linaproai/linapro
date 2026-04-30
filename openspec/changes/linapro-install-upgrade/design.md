## Context

### 当前状态

`LinaPro` 仓库当前提供以下"开发态安装与升级"实现:

**安装侧**

- `openspec/specs/framework-bootstrap-installer/spec.md` 中已经定义了 `hack/scripts/install/install.sh` + `install.ps1` 双脚本规范、基于 source archive 下载、默认 stable tag 解析、目录覆盖保护与健康检查等需求
- 但实际仓库 `hack/scripts/install/` 目录尚未落地实现,用户当前唯一可行的安装路径仍是手工 `git clone` + `make init`
- 默认管理工作台仍提示 `admin/admin123`,数据库初始化依赖 `make init`

**升级侧**

- `Makefile`(根)定义 `upgrade` target,通过 `confirm=upgrade` 二次确认
- `apps/lina-core/Makefile` 定义 `upgrade` 代理 target
- `hack/tools/upgrade-source/` 是独立 Go module,实现 `cli` + `frameworkupgrade` + `sourceupgrade` 三层
- 升级基线版本来自 `apps/lina-core/hack/config.yaml.frameworkUpgrade.version`
- 框架升级流程"全量重放"`manifest/sql/` 中所有 SQL 文件
- 源码插件升级走 `scope=source-plugin plugin=<id>` 子命令

### 约束

- `LinaPro` 项目处于全新阶段,**无历史兼容包袱**,允许直接断更老的 `make upgrade` 实现
- 用户分层契约: Tier 1 (插件)零冲突;Tier 2 (修改 core)冲突自理;Tier 3 (生成代码)永远覆盖
- 安装入口必须能够支撑 `curl -fsSL https://linapro.ai/install.sh | bash` 单条命令
- 不引入新的项目级配置文件,`.linapro.yaml` 之类的概念不接受
- 不引入新的二进制 CLI,所有动作通过 `git` / `make` / `bash` 完成
- 升级技能必须能被 `Claude Code` / `Codex` 等 AI 工具加载执行,只有冲突无法解决时才需要人工介入

### 利益相关方

- **框架终端用户**: 在自己的项目仓库中安装、升级 LinaPro,期望安装一条命令、升级零干预
- **框架维护者**: 负责保持 `Tier 1` 公共契约稳定、维护 changelog、发布 stable tag
- **AI 工具(`Claude Code` / `Codex`)**: 加载升级技能,代替人工执行升级流程,失败时把决策点抛给用户

## Goals / Non-Goals

### Goals

- **G1** 安装路径收敛为单条 `curl -fsSL https://linapro.ai/install.sh | bash`,在 `macOS` / `Linux` / `Windows-via-Git-Bash-or-WSL` 三平台行为一致
- **G2** 升级路径完全由 `lina-upgrade` 技能驱动,无冲突时全自动,有冲突且 AI 无法处置时才转人工
- **G3** 不引入新的项目级配置文件或新二进制 CLI 工具
- **G4** 现有 `make upgrade` + `hack/tools/upgrade-source/` 整套实现彻底移除,不留向后兼容别名
- **G5** 升级技能必须能正确处理框架升级与源码插件升级两条子工作流(等价覆盖原 `make upgrade` 的 `scope=framework` + `scope=source-plugin`)
- **G6** 安装与升级流程在用户改了任意 core/vben 文件的前提下仍能尝试自动合并,失败时给出可操作的人工动作清单
- **G7** 文档详细程度必须支持后续 `/opsx:apply` 时各任务直接落地,不需要再补设计决策

### Non-Goals

- **NG1** 不在本变更中实现 `linapro.ai/install.sh` CDN 部署管道,只交付仓库内 `bootstrap.sh` 源文件并文档化部署流程
- **NG2** 不为 Windows 原生 PowerShell 用户维护单独的 `.ps1` 入口,要求他们切换到 Git Bash 或 WSL
- **NG3** 不实现"运行时业务系统升级"(`runtime-upgrade-governance` 中保留为未来方向),本变更范围仅限源码升级
- **NG4** 不维护"用户改了哪些核心文件"的登记表(`.linapro.yaml.modified`),改动信息全部通过 git diff 实时计算
- **NG5** 不为 `Tier 1` 公共契约稳定性做形式化保证(语义化版本承诺需另行立项)
- **NG6** 不在升级技能中尝试复杂的"基于 AI 的代码理解 / 自动重构合并",只做"按 Tier 选择策略 + 三路合并 + 测试验证"
- **NG7** 不实现安装脚本的 i18n 化,所有输出统一英文(README 维持中英镜像)

## Decisions

### D1 安装入口的形式: `curl | bash` + 仓库内 `bootstrap.sh` 双载体

**决策**: `https://linapro.ai/install.sh` 的内容 = 仓库内 `hack/scripts/install/bootstrap.sh` 的发布拷贝,本仓库只负责维护 `bootstrap.sh` 源文件,部署到 CDN 是单独的运维步骤。

**为什么**:

- 一份逻辑两个载体,源码仍在仓库内 review / version 管理
- `bootstrap.sh` 在 git clone 后也能以 `bash hack/scripts/install/bootstrap.sh` 的方式从本地运行,等价于 curl 入口
- CDN 静态托管比动态服务简单,出错时只需要重发 `bootstrap.sh`

**替代方案与拒绝理由**:

- *动态生成服务*: 维护成本高,引入运行时单点;拒绝
- *直接下载 GitHub raw 链接*: 可以,但 URL 较长,失去品牌入口;拒绝
- *仅维护 CDN 版,不在仓库内放*: 失去版本一致性,本地无法 review;拒绝

### D2 `bootstrap.sh` 必须完全自包含,不再 curl 二级脚本

**决策**: `bootstrap.sh` 只做"探测 OS → 解析版本 → git clone → exec dispatch",不再通过 `curl` 拉取任何二级脚本。后续平台脚本随 `git clone` 一同到达本地,版本一致性由 git tag 保证。

**为什么**:

- 减少 curl|bash 链路上的网络故障点
- 防止"install.sh 来自 CDN v0.7,helpers.sh 也是 CDN 上的最新"的版本错配
- 简化签名校验路径(只需要校验单个 bootstrap.sh)
- 单文件易于审计 — 用户可以在执行前直接 `cat | less` 查看

**替代方案与拒绝理由**:

- *bootstrap.sh + curl 二次拉取 helpers.sh*: 网络故障增加、版本错配增加;拒绝

### D3 默认安装版本 = GitHub `releases/latest` 重定向解析

**决策**: 在用户未设置 `LINAPRO_VERSION` 环境变量时,通过 `curl -sI https://github.com/linaproai/linapro/releases/latest` 获取重定向 URL,提取 tag 名作为目标版本。无法解析时硬失败,不回落 main 分支。

**为什么**:

- GitHub 自己定义"latest stable"语义,自动排除 `-rc` / `-alpha` 预发布
- 不消耗 GitHub API rate limit(匿名 60 次/小时)
- 用户体验直接对应"GitHub Releases 页面上看到的最新发布"
- 硬失败优于"静默落到 main 分支" — 让用户对版本有明确感知

**替代方案与拒绝理由**:

- *`git ls-remote --tags --sort=-v:refname`*: 会带上 `-rc.1` 这类预发布,需要正则过滤;拒绝
- *GitHub Releases API*: 走 API 受 rate limit 影响;拒绝
- *静默回落 main 分支*: 与 G3 "可控版本"原则冲突;拒绝

### D4 三平台脚本统一用 `bash` 实现,不再维护 PowerShell

**决策**: 提供 `install-macos.sh` / `install-linux.sh` / `install-windows.sh` 三个 bash 脚本,均假设运行在 bash 环境。Windows 用户必须在 Git Bash 或 WSL 终端中执行,文档明示。

**为什么**:

- Git Bash 随 Git for Windows 一起安装,Windows 上做 `LinaPro` 开发的用户 99% 已具备
- WSL 是更现代的方案,行为完全等价于 Linux
- 维护一份 bash 实现 + 三个平台分支,显著低于 (bash × 2) + (ps1 × 1) 的成本
- 升级技能 `lina-upgrade` 的 scripts 也只需要 bash 实现,Windows 用户在 Git Bash 里调 Claude Code 一样跑得通

**替代方案与拒绝理由**:

- *维护 install.ps1*: 双倍维护成本、PowerShell 与 bash 的探测/字符串处理/网络调用差异大;拒绝
- *仅 WSL,放弃 Git Bash*: 安装门槛提高,Git Bash 用户被边缘化;拒绝
- *提供 .bat 包装*: 无意义,bash 用户可以直接用 bash;拒绝

### D5 默认 full clone,`LINAPRO_SHALLOW=1` opt-in 浅克隆

**决策**: 默认 `git clone --branch <tag>` 不带 `--depth`,获取完整历史。提供 `LINAPRO_SHALLOW=1` 环境变量供 CI 或容器环境降级为浅克隆。

**为什么**:

- LinaPro 是开发者要"住进去"几年的项目,git 历史是文档资源
- 后续 `lina-upgrade` 技能强依赖完整历史(`git diff baseline...HEAD` / `git merge-base --is-ancestor` 等),浅克隆需要先 `--unshallow` 才能升级,反而是潜在故障点
- 用户体验一致:`git log` 任何时候都能用

**替代方案与拒绝理由**:

- *默认浅克隆,升级时再 unshallow*: 多了一步可能失败的网络操作;拒绝
- *从不允许浅克隆*: 牺牲 CI/容器场景的速度;拒绝

### D6 版本基线 = `metadata.yaml.framework.version`

**决策**: 升级技能从 `apps/lina-core/manifest/config/metadata.yaml.framework.version` 读取当前 baseline 版本,不引入新的 `.linapro.yaml`。约定该字段不允许用户手工编辑,违反时由升级技能在校验阶段提示用户修正。

**为什么**:

- `metadata.yaml.framework.version` 已经存在,作用是 API 文档与系统信息页展示
- 单一真值源避免"yaml 与实际状态不一致需要手工修复"的二阶问题
- 用户改动通过 `git diff upstream/<baseline>...HEAD` 动态计算,而不是依赖手工登记
- 升级历史从 git log / git tag 推导,不需要专门维护

**替代方案与拒绝理由**:

- *新增 `.linapro.yaml`*: 多一个文件、多一处维护、自报数据可能过期;拒绝
- *从 git tag 推导基线*: 用户可能在 tag 之上提交了 commit,无法直接判定;拒绝

### D7 baseline 校验四层

**决策**: 升级技能首先调用 `upgrade-baseline-check.sh`,执行四层校验:

| 层 | 检查内容 | 失败处理 |
|---|---|---|
| 1 存在性 | `declared_version` 必须是 upstream 真实 tag | 返回 `ERR_TAG_NOT_FOUND`,AI 询问用户 |
| 2 可达性 | HEAD 必须是该 tag commit 的后代 | 返回 `ERR_HEAD_NOT_DESCENDANT`,AI 询问用户 |
| 3 身份对照(软) | SQL 编号 / 关键路径文件统计,异常时打印警告 | 不阻断,AI 综合判断 |
| 4 汇总 | `OK_BASELINE_CONFIRMED` + commits_ahead / core_changed / sql 计数 | 进入下一步 |

**为什么**:

- 第 1+2 层是硬约束:基线版本必须真实存在且可达,否则后续 `git diff baseline...HEAD` 无意义
- 第 3 层是软警告:用户合法行为也会让 SQL 数量变化,把"看起来怪不怪"的判断交给 AI
- 第 4 层提供量化指标,让 AI 在用户对话中能给出具体数字
- 校验脚本必须**无副作用**,只读 git 状态 + metadata.yaml 一个值,不写任何文件

**替代方案与拒绝理由**:

- *只做存在性校验*: 用户 reset/rebase 后会得到错误的 baseline;拒绝
- *把校验做进 SKILL.md 让 AI 直接调 git*: 失去脚本的可测试性;拒绝

### D8 文件 Tier 分类与冲突解决策略

**决策**: 把 `apps/lina-core/` / `apps/lina-vben/` / `apps/lina-plugins/` 下所有路径明确归入 Tier 1/2/3,升级时按 Tier 自动选择策略:

| Tier | 路径 | 升级策略 |
|---|---|---|
| 1 | `pkg/bizerr/**`、`pkg/logger/**`、`pkg/contract/**`、插件运行时公共接口、`apps/lina-plugins/<your-plugin>/**`(用户自有插件目录) | 不应冲突,出现冲突即转人工 |
| 2 | `apps/lina-core/internal/**`(除自动生成路径)、`apps/lina-vben/apps/web-antd/src/**`(除自动生成路径)、`apps/lina-core/manifest/config/*.yaml`(除版本字段) | git 三路合并,失败由 AI 评估信心,信心不足转人工 |
| 3 | `apps/lina-core/internal/dao/**`、`apps/lina-core/internal/model/{do,entity}/**`、`apps/lina-core/internal/controller/**`(骨架部分)、插件后端的同等路径 | `git checkout --theirs`,然后 `make dao` / `make ctrl` 重生成 |

**为什么**:

- Tier 1 是稳定契约层,设计上不应该被用户改动,所以"出现冲突即转人工"是合理的强约束
- Tier 3 是自动生成代码,任何手工修改都会被下次 `make dao/ctrl` 覆盖,所以直接接受上游版本是安全的
- Tier 2 是用户合法可改的区域,需要 AI 介入做语义合并

**替代方案与拒绝理由**:

- *不分 Tier,所有冲突都转人工*: 失去自动化收益,违反 G2;拒绝
- *用 ML 模型动态分类*: 过度设计,文件路径分类已经足够;拒绝

### D9 升级技能整体工作流(10 步)

**决策**: `lina-upgrade` 技能的标准工作流必须严格按 10 步执行:

```
1. 前置守卫     git status 干净 / metadata.yaml 存在 / 目标版本 ≥ 当前版本
2. baseline 校验  调用 upgrade-baseline-check.sh,失败时 AI 与用户对话
3. 升级计划     拉取 changelog + OpenSpec 归档 + git diff 改动文件 → 用户确认
4. 执行合并     git merge --no-commit upstream/<target>
5. 冲突处理     按 Tier 自动 / 必要时转人工
6. 重生成       make dao + make ctrl
7. 数据迁移     make init 按编号补跑新 SQL
8. 验证         go build + pnpm typecheck + pnpm lint + e2e smoke
9. 提交         git commit -m "chore: upgrade to vX.Y"(metadata.yaml 已被上游覆盖)
10. 报告        总结自动解决了什么、人工要看什么
```

**为什么**:

- 顺序固定可以让 AI 跨会话保持一致行为
- 每一步都有明确的失败处理(转人工或终止)
- 验证阶段在合并提交之前,失败可以 `git merge --abort` 安全退出

**替代方案与拒绝理由**:

- *AI 自由编排步骤*: 失去可预测性,不同 AI 工具产生不同行为;拒绝

### D10 升级转人工的硬规则

**决策**: 写入 `references/escalation-rules.md`,包含 5 条硬规则:

1. Tier 1 区域出现冲突
2. Tier 2 三路合并后 AI 自评信心不足或语义有歧义
3. 数据库迁移会破坏用户数据(如 DROP COLUMN 命中已有数据)
4. e2e smoke 失败且自动回滚未恢复
5. 用户主动声明改过的核心文件被上游也改了

**为什么**:

- 转人工是契约,必须显式列出而不是依赖 AI 主观判断
- 每条规则都对应具体可观测的状态(冲突标记、SQL 语义、测试结果),AI 容易检测

### D11 框架升级 + 源码插件升级两条子工作流

**决策**: `lina-upgrade` 技能内部分发到两条子工作流:

- **框架升级子流程**: 当用户说"升级框架到 vX.Y"时触发,执行 D9 的 10 步
- **源码插件升级子流程**: 当用户说"升级插件 plugin-demo 到 vX.Y"时触发,行为等价于原 `make upgrade scope=source-plugin plugin=<id>`,即:
  - 对比 `apps/lina-plugins/<id>/plugin.yaml.version` 与 `sys_plugin.version`
  - 应用插件的 release 流程(执行 phase=upgrade SQL、同步治理资源)
  - 此子流程不涉及 `git merge`,因为源码插件版本切换是"机械替换"

**为什么**:

- 用户原 `make upgrade scope=source-plugin` 流程不能简单丢弃,必须有等价替代
- 两条子流程的失败模式不同:框架升级看 git 冲突,插件升级看 SQL 与治理同步,需要分别处理
- 把两条子流程都收敛在同一个技能里,用户无需了解 scope 概念

### D12 安装脚本的 prereq 探测策略

**决策**: `checks/prereq.sh` 检测以下工具与版本:

| 工具 | 最低版本 | 缺失时行为 |
|---|---|---|
| `go` | 1.22+ | 提示对应平台安装方式(brew / apt / winget),不自动装(除非 `--auto-install-tools`) |
| `node` | 20+ | 同上 |
| `pnpm` | 8+ | 提示 `npm i -g pnpm` |
| `git` | 2.x | 大概率已存在(curl|bash 流程已用过 git) |
| `make` | 任意 | macOS/Linux 默认有,Windows Git Bash 没有,提示安装(或允许跳过,后续 install-windows.sh 直接调命令) |
| `mysql` 客户端可达 | - | 提示用户配置 `manifest/config/config.yaml` 或 docker 启动 |

**为什么**:

- 提示而不是自动装是更安全的默认,避免在用户系统装不需要的依赖
- `--auto-install-tools` 提供"我懒得管,你帮我装"的路径,但不强制
- Git Bash 没有 `make` 是已知缺陷,通过让 `install-windows.sh` 直接调底层 `go` / `pnpm` 命令绕过

### D13 端口冲突与配置文件保护

**决策**: 安装脚本在执行 `make init` 前探测 5666 / 8080 端口占用,占用时打印警告但不阻塞(因为可能用户已经手工启动过)。已存在的 `apps/lina-core/manifest/config/config.yaml` 不覆盖,只在不存在时从 `config.template.yaml` 复制。

**为什么**:

- 重复执行 `install-*.sh` 必须幂等,不能破坏用户已有配置
- 端口探测是友好提示,不是硬约束
- 配置模板复制是首次安装的便利,后续升级不应触发

### D14 删除既有 `make upgrade` 的范围

**决策**: 完整删除以下内容,不保留向后兼容别名:

```
□ 整个 hack/tools/upgrade-source/ 目录(独立 Go module)
□ 根 Makefile 的 upgrade target 与相关 .PHONY 声明
□ apps/lina-core/Makefile 的 upgrade 代理 target
□ apps/lina-core/hack/config.yaml 中的 frameworkUpgrade 区块(由升级技能不再使用)
□ apps/lina-core/internal/cmd/ 中若引用 frameworkUpgrade.version 的代码改读 metadata.yaml.framework.version
□ 全仓 grep "make upgrade" 清理过时文档引用
□ README.md / README.zh_CN.md / CLAUDE.md 中相关章节
□ .agents/instructions/ / .agents/prompts/ 中如有引用同步迁移
```

**为什么**:

- 项目处于全新阶段,无历史用户依赖 `make upgrade` 命令
- 保留别名只会让升级路径变模糊
- 完整删除可以让代码库更清晰,降低后续维护成本

### D15 文档语言策略

**决策**: 本变更的 `proposal.md` / `design.md` / `tasks.md` / 增量规范文档**全部使用中文**,因为用户用中文交互;归档时统一翻译为英文,这与项目 `OpenSpec document language policy` 一致。

`hack/scripts/install/README.md` 使用英文,`README.zh_CN.md` 使用中文(项目规范要求"目录级主说明文档统一英文 + 中文镜像")。

`SKILL.md` / `references/*.md` / 安装脚本的输出文本统一使用英文,因为这些主要给 AI 工具读取或在终端显示,英文具备更好的工具兼容性。

## Risks / Trade-offs

### R1 `linapro.ai` CDN 部署不在本仓库 PR 范围

**风险**: 合并 PR 后,`https://linapro.ai/install.sh` 不会自动指向新版本的 `bootstrap.sh`,用户依然会拿到旧版本(或 404,如果之前没部署过)。

**缓解**:
- 在 `tasks.md` 中文档化 CDN 部署流程作为附属人工任务
- 在 `README.md` 中暂时同时提供"curl|bash 远程入口"和"git clone 后本地执行"两种方式,远程入口稳定前后者保底
- 在 PR 描述中明确标注需要运维侧动作

### R2 Windows Git Bash 行为差异

**风险**: Git Bash 下 `chmod` 无效、路径格式 `C:\` vs `/c/` 切换、CRLF/LF 转换等,可能让脚本在某些 Windows 配置下行为异常。

**缓解**:
- `.gitattributes` 强制 `*.sh` 使用 LF
- 脚本内统一使用 `/c/foo` 风格路径,需要传给 Windows 原生工具时 `cygpath -w` 转换
- 不依赖文件 `+x` 权限,所有脚本通过 `bash script.sh` 调用
- 在 `install-windows.sh` 顶部输出"You are running in Git Bash / WSL"的探测结果,帮助用户确认环境

### R3 升级技能"无冲突全自动"承诺受 Tier 1 稳定性影响

**风险**: 如果框架未来出现 Tier 1 公共契约的破坏性变更(`pkg/bizerr` / `pkg/logger` / Plugin API 等),升级会从"自动"退化为"频繁人工"。

**缓解**:
- 在 `references/escalation-rules.md` 明确"Tier 1 冲突即转人工"
- 通过 `references/changelog-conventions.md` 要求每个引入 Tier 1 变更的 OpenSpec 归档变更必须明确标注 `BREAKING (Tier 1)`
- 升级计划生成阶段(D9 步骤 3)会汇总所有 Tier 1 标注的变更,用户提前看到风险

### R4 GitHub `releases/latest` 重定向解析依赖网络

**风险**: 用户网络不通 GitHub 时无法解析默认版本,安装直接失败。

**缓解**:
- 文档明示 `LINAPRO_VERSION=v0.1.0` 显式覆盖
- bootstrap.sh 失败时打印的错误消息中包含覆盖示例
- 不引入"静默回落 main"的兜底,因为这与 D3 决策冲突

### R5 baseline 校验失败时 AI 决策的不确定性

**风险**: AI 在面对 `ERR_TAG_NOT_FOUND` 时,可能根据用户回答错误地"猜"基线版本,导致后续 `git diff` 计算的"用户改动清单"包含上游 commit,合并行为不可预期。

**缓解**:
- 校验脚本输出包含 `hint:` 字段列出最近 3 个 stable tag 作为候选
- AI 必须把候选列表展示给用户,不能直接选最近的
- 失败时不允许进入第 4 步合并,只能停在对话状态

### R6 `frameworkUpgrade.version` 字段从 `hack/config.yaml` 移除可能影响下游引用

**风险**: 如果 `apps/lina-core/internal/` 中有代码读取 `hack/config.yaml.frameworkUpgrade.version`(例如启动时校验或在系统信息页展示),移除字段会导致编译失败或运行时错误。

**缓解**:
- `tasks.md` 中明确包含"全仓 grep `frameworkUpgrade` 引用"的清理任务
- 任何对该字段的运行时读取改为读 `metadata.yaml.framework.version`
- 测试覆盖确保启动流程不再依赖该字段

## Migration Plan

### 部署顺序

1. **PR 合并前**: 在仓库内完成 `bootstrap.sh` / `install-*.sh` / `lib/_common.sh` / `checks/prereq.sh` / `lina-upgrade` 技能等所有源文件
2. **PR 合并前**: 删除 `hack/tools/upgrade-source/` 与 `Makefile.upgrade` target,确保 CI 通过
3. **PR 合并前**: 同步更新 `README.md` / `README.zh_CN.md` / `CLAUDE.md`
4. **PR 合并后(运维侧)**: 把仓库内最新的 `hack/scripts/install/bootstrap.sh` 部署到 `https://linapro.ai/install.sh` CDN 路径(并同步部署 `bootstrap.ps1` 入口若未来需要 — 当前不需要)
5. **PR 合并后(运维侧)**: 验证 `curl -fsSL https://linapro.ai/install.sh | bash` 在干净环境中能成功安装

### 回滚策略

- **仓库内回滚**: PR 合并后若发现严重问题,通过 `git revert` 恢复整个变更
- **CDN 回滚**: 保留最近一份可用的 `bootstrap.sh`,部署失败时 CDN 切回上一版本
- **用户侧**: 已经成功安装的用户不受影响,因为安装结果是本地 git 仓库,与 CDN 无关

### 升级到本变更后的体验切换

- 老用户(若存在已安装 `LinaPro` 的开发者): 拉取最新代码后 `make upgrade` 命令将不存在,按新文档使用 `lina-upgrade` 技能即可
- 新用户: 直接使用 `curl -fsSL https://linapro.ai/install.sh | bash` 安装,使用 `lina-upgrade` 技能升级

## Open Questions

### Q1 `apps/lina-core/hack/config.yaml.frameworkUpgrade` 完整保留还是部分删除?

**当前倾向**: 保留 `repositoryUrl` 字段供安装脚本作为 `git clone` 的默认 origin 推断;删除 `version` 字段(已被 `metadata.yaml.framework.version` 替代)。

**待落地时确认**: 实施时需要先 grep `frameworkUpgrade` 在代码中的全部引用,根据实际依赖决定是否完整删除整个 `frameworkUpgrade` 区块。

### Q2 升级技能是否需要 `dry-run` 模式?

**当前倾向**: 提供 `dry-run` 模式作为可选,通过用户在调用技能时声明"先预览"触发。技能输出升级计划后停止,不执行 `git merge`。

**待落地时确认**: 是通过 SKILL.md 中的对话约定实现,还是通过 scripts 脚本的 `--dry-run` 参数实现 — 倾向前者(更轻量)。

### Q3 安装时是否要求用户配置 git remote?

**当前倾向**: `git clone` 默认会建立 `origin` 指向 `linaproai/linapro`,用户保留这个 origin 即可。后续升级技能优先尝试 `upstream` remote,不存在时回退到 `origin`。文档建议用户日常推自己代码时改名 `origin → upstream`,但不强制。

**待落地时确认**: 是否在安装末尾打印"建议你设置 upstream remote"的提示。
