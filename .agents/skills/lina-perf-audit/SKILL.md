---
name: lina-perf-audit
description: "Manual trigger only. Run a comprehensive LinaPro backend API performance audit only when the user explicitly requests lina-perf-audit or confirms a full audit. This workflow resets the database via make init/mock, restarts services, installs and enables all built-in plugins, seeds load-test data, and launches concurrent sub-agents — it takes significant time (tens of minutes to hours) and consumes substantial tokens. Do not invoke from other skills, CI, scheduled tasks, Git hooks, or vague performance requests."
---

# LinaPro 性能审计

使用此技能对 `apps/lina-core` 和 `apps/lina-plugins` 下的所有内置插件进行全面的 LinaPro 后端 API 审计。审计内容包括 N+1 查询、缺少索引、无限制的列表响应、重复读取、可合并的 SQL 调用、循环中的阻塞操作，以及读/查询端点中执行了 INSERT、 UPDATE、DELETE、REPLACE、TRUNCATE 或 ALTER 等写入 SQL 的问题。此技能仅生成报告和问题卡片，不修复生产代码。

此技能为纯 Markdown 格式，以便 Claude Code、Codex 及其他 AI 编码工具均可读取。

**交互语言**：与用户交互的内容语言以用户上下文使用的语言为准，用户使用英文则使用英文，用户使用中文则使用中文。

## 手动触发门槛

此技能会对本地开发环境产生破坏性影响，且运行成本较高。它会重置数据库、重新加载模拟数据、重启服务、安装并启用所有内置插件、写入临时审计日志、运行多个子代理，并可能消耗大量 Token 预算。

仅在用户明确要求全面审计时才继续执行，例如：

- `run lina-perf-audit`
- `执行 LinaPro 性能审计`
- `对所有后端 API 进行全面性能审计`
- `使用性能审计技能检查所有后端 API 的 N+1 问题`

对于模糊请求，如 `API 好像有点慢`、`性能怎么样` 或 `检查接口性能`，应先请求用户确认。确认消息中必须提及数据库重置、服务重启、耗时、子代理扇出和 Token 成本。在确认前，不得运行 `make stop`、`make init`、`make mock`、`setup-audit-env.sh`、`prepare-builtin-plugins.sh` 或 `stress-fixture.sh`。

## 适用场景

- 用户明确指定 `lina-perf-audit` 或要求运行完整的 LinaPro API 性能审计。
- 用户明确要求对 `lina-core` 和内置插件进行系统性后端 API 审计。
- 用户明确要求在 `temp/lina-perf-audit/<run-id>/` 下生成运行报告，并在 `perf-issues/` 下生成持久化问题卡片。
- 用户在了解成本和本地破坏性影响后确认了模糊的性能请求。

## 不适用场景

- 禁止从其他技能（如 `lina-review`、`lina-feedback` 或 `lina-e2e`）触发。这些技能可以建议用户手动运行 `lina-perf-audit`，但不得自行调用。
- 禁止从 CI、定时任务、Git 钩子、自动化流程或后台工作流中运行。
- 除非用户明确要求全面审计，否则不要为单个端点调查运行此技能。应使用针对性的手动分析。
- 对于模糊的性能关切，在用户确认前不要运行。
- 如果用户无法接受数据库重置、模拟数据重载、服务重启或 Token/时间成本，则不要运行。

## 参考文件

仅在需要时加载以下文件：

- `references/sub-agent-prompt.md`：阶段 1 子代理的提示模板，包括端点或小型模块分片。
- `references/severity-rubric.md`：HIGH / MEDIUM / LOW 分级标准和反模式签名。
- `references/report-template.md`：`audits/<module>.md`、`SUMMARY.md` 和 `meta.json` 模板。
- `references/issue-card-template.md`：持久化 `perf-issues/*.md` 问题卡片模板。
- `references/fingerprint-rule.md`：精确的指纹生成、去重、更新和跨运行规则。

## 内置脚本

确定性的辅助脚本维护在此技能目录下的 `scripts/` 中。从仓库根目录运行，路径格式如 `bash .agents/skills/lina-perf-audit/scripts/scan-endpoints.sh ...`。不要将这些脚本复制到 `hack/` 或在技能目录外维护第二套脚本；技能目录是审计工作流的所有权边界。

## 工作流

严格按三个阶段运行审计。

### 阶段 0：准备

1. 确认用户已明确批准全面审计。
2. 创建 `YYYYMMDD-HHMMSS` 格式的唯一 `run_id`，并设置 `run_dir=temp/lina-perf-audit/<run_id>`。
3. 运行 `make stop`。
4. 通过 `make init confirm=init rebuild=true` 和 `make mock confirm=mock` 重置本地数据。
5. 运行 `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id <run_id>`，修补审计日志配置、启动后端、等待就绪并获取 `admin` Token。
6. 运行 `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir <run_dir>`，发现、同步、安装、启用所有内置插件并加载模拟数据。
7. 在宿主模拟数据和插件模拟数据就绪后，运行 `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir <run_dir>`。
8. 运行 `bash .agents/skills/lina-perf-audit/scripts/scan-endpoints.sh --run-dir <run_dir>`，从宿主和插件 API DTO 生成 `catalog.json`。
9. 运行 `bash .agents/skills/lina-perf-audit/scripts/probe-fixtures.sh --run-dir <run_dir>`，生成 `fixtures.json` 并对不可达的声明路由快速失败。
10. 将没有后端 API 的跳过插件记录到 `meta.json` 中，原因为 `no backend API`。

无论成功或失败，都必须通过 `bash .agents/skills/lina-perf-audit/scripts/restore-audit-env.sh --run-dir <run_dir>` 恢复临时日志设置并停止服务。

### 阶段 1：并发子代理审计

使用子代理执行端点审计任务。默认为 `catalog.json` 中的每个模块分配一个子代理。如果模块过大无法在单个提示中处理，将其拆分为端点分片或小型模块分片；每个子代理提示必须保持在 5KB 以下，且仅包含其分配的端点子集。

每个子代理接收：

- `module`
- `endpoints[]`
- `fixtures`
- `log_path`
- `token`
- `run_dir`

每个子代理串行调用其分配的端点，读取 GoFrame `Trace-ID` 响应头，在 `server.log` 中搜索匹配的 SQL 行，检查相关的控制器/服务源码，对发现进行分类，并在 `temp/lina-perf-audit/<run-id>/audits/` 下编写一个审计 Markdown 文件。对于每个 GET/读取/查询端点，子代理还必须验证请求追踪中不包含写入 SQL 语句（`INSERT`、`UPDATE`、`DELETE`、`REPLACE`、`TRUNCATE`、`ALTER`、`DROP` 或 `CREATE`）。在读取请求中观察到写入 SQL 即为发现，即使端点响应很快也是如此，除非追踪中同时包含读取 SQL 且每个写入语句仅操作 `sys_online_session` 和/或 `plugin_monitor_operlog`。

如果 `Trace-ID` 不可用，子代理必须使用按调用时间窗口加请求 URL 的回退搜索方式，并将证据标记为 `trace ID unavailable, evidence quality reduced`。不得仅因 `Trace-ID` 缺失而跳过端点。

启动阶段 1 前，加载 `references/sub-agent-prompt.md` 作为基础模板。

### 阶段 2：汇总与性能问题卡片

1. 运行 `bash .agents/skills/lina-perf-audit/scripts/aggregate-reports.sh --run-dir <run_dir>`。
2. 脚本读取所有 `audits/*.md`，将结果合并到 `temp/lina-perf-audit/<run-id>/SUMMARY.md`，并根据 `references/severity-rubric.md` 审查报告严重性，将发现分类为 HIGH、MEDIUM 和 LOW。
3. 脚本在 `perf-issues/<severity>-<module>-<slug>.md` 下为每个发现生成或更新一张持久化问题卡片。
4. 脚本根据 `references/fingerprint-rule.md` 按指纹对卡片进行去重。
5. 脚本重新生成 `perf-issues/INDEX.md`，按严重性排列所有 `open` 和 `in-progress` 状态的卡片。
6. 脚本写入最终的 `meta.json`，包含开始/结束时间、Git 提交、压力测试数据状态、子代理数量、子代理状态、跳过的插件、日志设置和恢复结果。
7. 确认 `SUMMARY.md` 使用仓库相对路径链接到每张新建或更新的问题卡片。

`temp/lina-perf-audit/<run-id>/` 是每次运行的快照。`perf-issues/` 是跨运行的持久化积压列表，不得放在 `temp/` 下。

## 破坏性端点处理

子代理必须处理破坏性端点，同时不损坏共享的审计数据。

- 对于 DELETE、重置、清空、卸载等破坏性操作，当同一模块存在匹配的创建端点时，先创建一个专用的审计测试数据。
- 仅将返回的资源 ID 或稳定键用于正在审计的破坏性调用。
- 即使破坏性调用失败，也应尝试清理。
- 在报告条目标记 `autonomous fixture completed, shared data not polluted`。
- 如果同一模块没有匹配的创建端点，将端点标记为 `SKIPPED: no matching create endpoint, manual follow-up required`。
- 禁止以卸载内置插件、清除共享系统日志或销毁全局状态来替代模块自身的测试数据处理。

## 严重性分类

使用三个严重性级别：

- `HIGH`：列表/详情中存在 N+1 问题且有源码证据、大数据量表缺少索引、非批量端点超过 1 秒、循环中存在阻塞的远程/文件/事务操作，或任何 GET/读取/查询端点的请求追踪中执行了非运维性写入 SQL。
- `MEDIUM`：小样本 N+1、缺少分页、重复读取相同数据，或多个 SELECT 调用可通过 JOIN 或 WHERE IN 合并。
- `LOW`：SQL 次数略高但查询有索引且快速、可在应用层过滤的内容可下推到数据库，或仅在静态分析中发现的风险。

每个发现必须包含方法和路径、模块、Trace ID 或回退标记、SQL 次数、适用时的写入 SQL 次数、关键 SQL 片段、相对源文件和行号，以及至少一条具体的改进建议。

## 读取请求副作用检查

GET 端点以及描述为列表、查询、树、选项、计数、健康检查、当前状态或详情的端点应为只读。其追踪的 SQL 不得包含写入操作。检查范围为首个有效 Token 为 `INSERT`、`UPDATE`、`DELETE`、`REPLACE`、`TRUNCATE`、`ALTER`、`DROP` 或 `CREATE` 的 SQL 语句。

- 仅计算由审计端点追踪生成的 SQL。不计算阶段 0 的设置、登录、压力测试数据插入或破坏性端点处理的自主测试数据创建/删除调用。
- 如果读取端点仅写入 `sys_online_session` 和/或 `plugin_monitor_operlog` 且同时读取数据，将该会话心跳或操作日志写入视为预期的运维性副作用。仅在模块审计文件中记录为 PASS 说明，不要为此生成发现、汇总违规或 `perf-issues/` 卡片。
- 如果追踪仅包含上述预期写入表但不包含任何读取 SQL，则仍应报告，因为它不再符合正常的读取加运维性副作用模式。
- 如果读取端点写入任何其他业务、插件状态、运行时状态或存储表，报告为 `HIGH`，反模式签名前缀为 `read-write-side-effect`。
- 改进建议应推荐将副作用移至显式的 POST/PUT/DELETE 操作、将端点拆分为读和写操作，或用非变更性的缓存/读取模型替代持久化写入。
- 如果源码暗示存在写入但运行时 SQL 在采样路径中未显示，报告为 `LOW` 并标注仅静态分析，不要伪造运行时证据。

## 报告结构

每次运行输出：

```text
temp/lina-perf-audit/<run-id>/
  catalog.json
  fixtures.json
  server.log
  audits/<module-or-shard>.md
  SUMMARY.md
  meta.json
```

持久化问题卡片：

```text
perf-issues/
  HIGH-<module>-<slug>.md
  MEDIUM-<module>-<slug>.md
  LOW-<module>-<slug>.md
  INDEX.md
```

运行报告使用 `references/report-template.md`，问题卡片使用 `references/issue-card-template.md`。

## 问题卡片生命周期

每个性能问题对应一张持久化 Markdown 卡片。阶段 2 负责所有卡片的创建和更新。

- 新问题：创建 `perf-issues/<severity>-<module>-<slug>.md`，状态为 `open`。
- 已有指纹：更新 `last_seen_run`、递增 `seen_count`、追加历史记录，不创建重复卡片。
- 已有卡片状态为 `fixed` 或 `obsolete` 且问题再次出现：将状态改回 `open` 并追加回归历史记录。
- 已有卡片状态为 `in-progress`：保持状态并更新观察字段。
- 所有更新后重新生成 `perf-issues/INDEX.md`。
- 每张卡片正文必须包含固定章节：`问题描述`、`复现方式`、`证据`、`改进方案` 和 `历史记录`。
- 卡片描述性内容和 `perf-issues/INDEX.md` 标题必须使用中文编写。API 路径、SQL 片段、Trace ID、指纹、frontmatter 字段名和状态枚举值保持不变。

允许的状态为 `open`、`in-progress`、`fixed` 和 `obsolete`。此技能不得引入超出这些字段更新的自动状态机。

## 交叉引用规则

- `SUMMARY.md` 必须使用仓库相对路径链接到当前运行中创建或更新的所有问题卡片。
- 每张问题卡片的历史记录必须引用 `run_id` 和 `temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`。
- 持久化卡片应使用仓库相对源路径，例如 `apps/lina-core/internal/service/user/user.go:142`。
- 持久化卡片中不要使用本地机器的绝对路径。
- 不要将问题卡片移入 OpenSpec 归档目录。后续的 OpenSpec 变更可能会引用或更新卡片，但此技能的归档不会归档 `perf-issues/`。
