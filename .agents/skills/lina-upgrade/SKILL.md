---
name: lina-upgrade
description: Upgrade the LinaPro framework or source plugins through an AI-guided, script-driven workflow covering framework-wide and per-plugin upgrade tasks.
---

**交互语言**：与用户交互的内容语言以用户上下文使用的语言为准，用户使用英文则使用英文，用户使用中文则使用中文。

# 用途

`lina-upgrade` 引导 AI 工具完成 `LinaPro` 的开发期升级。
它替代了已移除的 `make upgrade` 入口，保持升级流程通过 `bash`、`git` 和 `make` 脚本驱动。

此技能有两个范围：

- `framework`：升级 `apps/lina-core/`、`apps/lina-vben/`、共享清单和生成的宿主工件。
- `source-plugin`：升级 `apps/lina-plugins/` 下的单个或所有源码插件。

不升级动态插件，也不引入新的二进制 `CLI` 或项目级配置文件。

# 调用时机

当用户表达以下意图时调用此技能：

- "升级 LinaPro 到 `v0.6.0`"
- "升级框架"
- "运行框架升级"
- "升级 LinaPro 框架到 `v0.6.0`"
- "升级源码插件 `plugin-demo-source`"
- "upgrade source plugin `plugin-demo-source`"
- "upgrade all source plugins"

如果请求未明确说明 `framework` 或 `source-plugin`，在运行脚本前先提出一个简短的澄清问题。

# AI 必须从用户收集的输入

前提条件：确保 `gf` 和 `openspec` 已安装；如有缺失，先调用 `lina-doctor`。

| 输入 | 何时必需 | 说明 |
| --- | --- | --- |
| 目标版本 | `framework` 升级 | 必须大于 `apps/lina-core/manifest/config/metadata.yaml.framework.version` 中声明的版本。 |
| 范围 | 始终 | 必须为 `framework` 或 `source-plugin`。 |
| 插件 ID | `source-plugin` 升级 | 使用具体的插件 ID 或 `all` 进行批量源码插件升级。 |
| 数据库备份确认 | 检测到 SQL 迁移 | 即使在非交互模式下也要打印备份提醒。 |

# 工作流

按顺序执行以下步骤。当步骤需要人工干预时停止。

1. **预检查**：验证 `git status --short` 是否干净、`metadata.yaml` 是否存在、目标版本是否大于已声明的基线版本。失败场景：工作区有未提交变更或目标版本不明确；请用户提交、暂存或澄清。
2. **基线校验**：运行 `scripts/upgrade-baseline-check.sh`。失败场景：`ERR_TAG_NOT_FOUND` 或 `ERR_HEAD_NOT_DESCENDANT`；显示脚本输出并请用户确认实际基线。
3. **计划生成**：运行 `scripts/upgrade-plan.sh <target-version>`。失败场景：缺少目标标签或变更日志不可读；修复输入后继续。将计划展示给用户确认。
4. **合并执行**：运行 `git merge --no-commit upstream/<target>` 或计划中选择的等价远程引用。失败场景：合并冲突；进入冲突处理。
5. **冲突处理**：使用 `scripts/upgrade-classify.sh <path>` 对每个冲突路径分类。第 1 级冲突始终上报。第 2 级冲突需要 AI 判断置信度；低置信度时上报。第 3 级冲突使用 `git checkout --theirs <path>`。
6. **重新生成**：运行 `scripts/upgrade-regenerate.sh`。失败场景：`make dao` 或 `make ctrl` 失败；报告日志路径并停止。
7. **数据库迁移**：提醒用户备份数据库，然后运行 `make init confirm=init` 执行增量 SQL 集。失败场景：迁移风险或破坏性 SQL；上报。
8. **验证**：运行 `scripts/upgrade-verify.sh`。失败场景：构建、类型检查、代码检查或冒烟测试失败；停止并报告失败的命令。
9. **提交**：验证通过后运行 `git commit -m "chore: upgrade to <target-version>"`。失败场景：用户要求不提交；按要求保留为已暂存或未暂存状态。
10. **报告**：汇总自动解决的冲突、人工决策、迁移状态、测试结果和剩余的用户检查项。

# 源码插件子流程

对于 `source-plugin` 升级，跳过框架合并步骤。检查 `apps/lina-plugins/<plugin-id>/plugin.yaml` 或所有插件清单，将发现的版本与宿主治理数据报告的有效已安装版本进行比较，然后通过宿主命令或专项脚本执行现有的显式源码插件升级服务路径。忽略动态插件。

# AI 必须生成的输出

- 合并前的升级计划：目标版本、基线版本、按级别分组的变更路径、SQL 迁移摘要和破坏性变更说明。
- 执行后的最终报告：已运行的命令、已解决的冲突、已执行的迁移、验证结果和待办操作列表。
- 受阻时的上报报告：具体路径、级别、失败代码以及技能无法安全继续的原因。

# 参考文件

- `references/tier-classification.md`
- `references/conflict-resolution.md`
- `references/escalation-rules.md`
- `references/changelog-conventions.md`

# 允许使用的工具

- `Bash`
- `Read`
- `Edit`
- `Grep`
- `Glob`
