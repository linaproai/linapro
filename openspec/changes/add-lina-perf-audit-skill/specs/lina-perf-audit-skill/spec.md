## ADDED Requirements

### Requirement: 仅手动触发约束

`lina-perf-audit` skill SHALL 明确声明为仅手动触发(manual-trigger-only),且 MUST NOT 被任何自动化路径(其他 skill、CI/CD 流水线、定时任务、git hook 等)引用或调用,且 MUST NOT 在用户描述含糊时自动展开完整审计。

#### Scenario: SKILL.md frontmatter 显式声明手动触发

- **WHEN** `.agents/skills/lina-perf-audit/SKILL.md` 的 frontmatter 被加载
- **THEN** description 字段必须包含醒目的 "MANUAL TRIGGER ONLY" 声明
- **AND** 必须列出本 skill 的资源代价(数据库会被重置、子 agent 数量、预估耗时、token 成本量级)
- **AND** 必须列出禁止自动调用的明确清单(其他 skill、CI、自动化流水线)

#### Scenario: 含糊请求必须先与用户确认

- **WHEN** 用户表达模糊性能关注(例如 "接口好像有点慢"、"性能怎么样"、"看下接口性能" 等)
- **THEN** skill 在执行任何 Stage 0 准备步骤前,必须显式向用户确认是否真的要启动完整审计
- **AND** 必须在确认请求中说明本 skill 的代价(reset DB、耗时、token 成本)
- **AND** 不得在确认前调用 `make stop` / `make init` 等任何破坏性操作

#### Scenario: 其他 skill 不得自动调用本 skill

- **WHEN** 任何其他 skill(例如 `lina-review`、`lina-feedback`、`lina-e2e`)在执行过程中检测到性能问题或性能相关需求
- **THEN** 该 skill 不得自动 invoke `lina-perf-audit`
- **AND** 应当向用户输出"建议手动触发 lina-perf-audit 进行系统性审查"的提示
- **AND** 由用户决定是否手动启动

### Requirement: 三阶段流水线编排

`lina-perf-audit` skill MUST 按 Stage 0 准备 → Stage 1 子 agent 并发审查 → Stage 2 主 agent 汇总 的三阶段顺序执行,且单次 run 产物 MUST 落到 `temp/lina-perf-audit/<run-id>/` 下的稳定路径。

#### Scenario: 辅助脚本在 skill 内部闭环维护

- **WHEN** 本 skill 需要确定性辅助脚本完成环境准备、插件准备、endpoint 扫描、fixture 探测或 stress fixture 写入
- **THEN** 这些脚本 MUST 位于 `.agents/skills/lina-perf-audit/scripts/`
- **AND** SKILL.md、references、OpenSpec 文档与问题卡片模板中的脚本命令 MUST 使用该 skill 内部路径
- **AND** 不得在 `hack/scripts/perf-audit/` 维护本 skill 私有脚本的第二份副本

#### Scenario: Stage 0 完成完整环境准备

- **WHEN** 用户确认启动审计
- **THEN** skill 必须依次执行:停服 → reset DB(`make init confirm=init rebuild=true && make mock confirm=mock`)→ 临时 patch logger 落盘配置并启动后端(`.agents/skills/lina-perf-audit/scripts/setup-audit-env.sh`)→ 等待就绪 → 用 admin/admin123 登录拿 token → 运行 `.agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh` 同步、安装并启用所有内置插件且按需加载插件 mock data → 运行 `.agents/skills/lina-perf-audit/scripts/stress-fixture.sh` 补充压力数据 → 运行 `.agents/skills/lina-perf-audit/scripts/scan-endpoints.sh` 生成 catalog → 运行 `.agents/skills/lina-perf-audit/scripts/probe-fixtures.sh` 生成 fixtures
- **AND** 必须为本次执行生成唯一 run-id(格式 `YYYYMMDD-HHMMSS`)
- **AND** 必须把上述产物全部写入 `temp/lina-perf-audit/<run-id>/`

#### Scenario: Stage 1 按接口任务拆分子 agent

- **WHEN** Stage 0 完成
- **THEN** skill 必须按 catalog 中的模块/资源边界生成接口任务队列,覆盖宿主模块与所有内置插件资源模块
- **AND** 每个 API 接口审查任务必须由子 agent 执行,主 agent 不得串行执行具体接口审查
- **AND** 每个子 agent 只负责分配给自己的接口 shard;当模块接口数量较多或 prompt 超过 5KB 时,必须继续拆成单接口或小接口组 shard
- **AND** 不同子 agent 之间允许并发执行
- **AND** 必须把每个子 agent 的产物写入 `temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`

#### Scenario: Stage 2 汇总产物结构稳定

- **WHEN** 所有子 agent 完成
- **THEN** skill 必须由主 agent 把所有 `audits/<module-or-shard>.md` 合并为 `temp/lina-perf-audit/<run-id>/SUMMARY.md`
- **AND** 必须按 HIGH / MEDIUM / LOW 三级对问题分类列出
- **AND** 必须输出本次执行的 `meta.json`,包含起止时间、git commit、stress-fixture 是否启用、子 agent 数量与状态

### Requirement: 所有内置插件审计覆盖

`lina-perf-audit` skill MUST 审计仓库内所有内置插件的后端 API,不得仅审计当前运行环境已经安装或启用的插件。

#### Scenario: Stage 0 安装并启用所有内置插件

- **WHEN** Stage 0 准备审计环境
- **THEN** skill 必须从 `apps/lina-plugins/*/plugin.yaml` 发现所有内置插件
- **AND** 必须通过宿主插件管理 API 同步、安装并启用每个内置插件
- **AND** 当插件存在 `manifest/sql/mock-data/` 时,必须在插件安装阶段显式请求加载插件 mock data
- **AND** 任一内置插件安装或启用失败时,Stage 0 必须失败并输出失败插件清单,不得静默跳过
- **AND** 成功安装但没有后端 API 的内置插件只在 endpoint 审计阶段标记为 skipped

#### Scenario: 内置插件 endpoint catalog 完整

- **WHEN** `scan-endpoints.sh` 生成 endpoint catalog
- **THEN** catalog 必须包含 `apps/lina-core/api/**/v1/*.go` 与 `apps/lina-plugins/*/backend/api/**/v1/*.go` 中所有声明了路由元数据的接口
- **AND** 对没有后端 API 的内置插件,必须在 `meta.json` 中记录 skipped plugin 条目,原因标记为 `no backend API`
- **AND** 对存在 DTO 但运行时路由不可访问的插件接口,必须在 Stage 0 或 probe 阶段报告为环境准备失败,不得进入正常审计结果

### Requirement: trace ID 直接读响应头

子 agent 调用接口时 MUST 直接读取 GoFrame 默认响应头 `Trace-ID`,用作反查日志的索引。本 skill MUST NOT 引入任何审计专用 middleware,也 MUST NOT 修改 GoFrame 默认行为或生产配置。

#### Scenario: 子 agent 通过响应头取得 trace ID

- **WHEN** 子 agent 调用某个接口
- **THEN** 必须读取响应头 `Trace-ID` 字段
- **AND** 用该 trace ID 在 `temp/lina-perf-audit/<run-id>/server.log` 中 grep 关联日志行
- **AND** 不得依赖任何新增 middleware 或自定义响应头

#### Scenario: trace ID 不可用时的降级路径

- **WHEN** 某个接口的响应头中 `Trace-ID` 为空(例如非正常 Flush 路径)
- **THEN** 子 agent 必须按"调用时间窗口 ± 2s + 请求 URL 模糊匹配"的方式定位日志行
- **AND** 在该接口的报告条目中标注 "trace ID unavailable, evidence quality reduced"
- **AND** 不得以拿不到 trace ID 为理由跳过该接口的审查

### Requirement: 破坏性接口自治治理

子 agent 在调用破坏性接口(DELETE / 卸载 / 清空 / 重置等)时,MUST 先 POST 创建一条目标资源、记录返回 ID、再 DELETE 该 ID,SHALL 确保审计不污染其他模块数据。

#### Scenario: 破坏性接口自治闭环

- **WHEN** 子 agent 准备调用某个 DELETE 类接口
- **THEN** 必须先调用同模块内的 create 接口,创建一条专门用于本次调用的资源
- **AND** 用返回的资源 ID 作为路径参数发起 DELETE
- **AND** 在报告条目中标注 "通过自治 fixture 完成,未污染共享数据"
- **AND** 即便 DELETE 调用失败,后续也必须尝试再次清理该 fixture

#### Scenario: 没有对应 create 接口的破坏性操作

- **WHEN** 破坏性接口在同模块内不存在对应 create 接口(例如清空全部登录日志、卸载具体内置插件)
- **THEN** 子 agent 必须在报告条目中标注 "SKIPPED: 无对应 create 接口,需人工补审"
- **AND** 不得绕过模块边界使用其他模块的资源做替代
- **AND** 必须在 SUMMARY.md 的"需人工补审清单"小节列出该接口

### Requirement: 报告问题严重度分级

skill MUST 按 HIGH / MEDIUM / LOW 三级对发现的问题分类,且每条问题 MUST 附带证据(SQL 序列、源码引用、复现接口路径)与改进建议。审计范围 MUST 同时覆盖性能问题与 GET/查询接口执行写 SQL 的读请求副作用问题。

#### Scenario: HIGH 级问题判定

- **WHEN** 接口出现以下情况之一
  - 列表/详情接口的 SQL 调用次数 ≈ 返回行数 + 常数,且能在源码中定位到循环内 dao 调用
  - SQL 涉及的 WHERE 字段在 schema 中无索引,且数据规模可能导致全表扫描
  - 接口响应时间 > 1s 且非导出/批量类
  - 在循环内执行远程调用、文件 I/O、跨数据库事务等阻塞操作
  - GET/查询/详情/选项/统计类读接口在自身请求 trace 中执行非预期的 INSERT、UPDATE、DELETE、REPLACE、TRUNCATE、ALTER、DROP、CREATE 等写 SQL
- **THEN** 必须在该接口报告条目中标注 `severity: HIGH`
- **AND** 必须在 SUMMARY.md 的 HIGH 区块列出该问题
- **AND** 改进建议中必须给出具体的合并查询、批量加载、分页或异步化方案

#### Scenario: MEDIUM 级问题判定

- **WHEN** 接口出现以下情况之一
  - N+1 模式但当前样本行数 < 20 未爆雷
  - 列表接口缺分页(返回全表)
  - 同一请求内重复读取相同数据
  - 多次 SELECT 可合并为 JOIN 或 WHERE IN
- **THEN** 必须标注 `severity: MEDIUM`

#### Scenario: LOW 级问题判定

- **WHEN** 接口出现以下情况之一
  - SQL 数量略多但都命中索引、单条耗时 < 5ms
  - 应用层做了可下推到数据库的过滤
  - 静态分析提示但实际调用未观察到性能问题
  - 静态分析提示 GET/查询接口可能在条件分支写库,但本次采样 trace 未触发该分支
- **THEN** 必须标注 `severity: LOW`
- **AND** 在改进建议中说明"可观察/择期优化"

#### Scenario: 查询请求写副作用必须被识别

- **WHEN** 子 agent 调用 GET/list/query/tree/options/count/detail/current 等读语义接口
- **THEN** 必须基于该请求的 `Trace-ID` 日志检查 SQL 序列中是否存在写操作
- **AND** 写操作关键字至少包括 `INSERT`、`UPDATE`、`DELETE`、`REPLACE`、`TRUNCATE`、`ALTER`、`DROP`、`CREATE`
- **AND** 如果发现非预期写 SQL,必须标注 `severity: HIGH`,反模式签名前缀为 `read-write-side-effect`
- **AND** 报告证据必须包含 write SQL count、写入目标表名、关键写 SQL 片段、源码 file:line 与"读接口存在写副作用"的分析
- **AND** 不得把 Stage 0 环境准备、登录、stress fixture、破坏性接口自治 create/delete 的写 SQL 计入该接口
- **AND** 改进建议必须说明将写副作用拆分到 POST/PUT/DELETE 动作或改为非持久化读取模型
- **AND** 如果该读请求 trace 中除读取 SQL 外,写 SQL 仅写入 `sys_online_session` 或 `plugin_monitor_operlog`,则该写入属于预期的会话活跃刷新或操作日志记录,应仅作为 PASS 备注记录,不得生成审计问题、SUMMARY 违规项或 `perf-issues/` 问题卡片
- **AND** 如果请求 trace 只写入上述运维表但没有读取 SQL,则不符合"查询 + 运维副作用"模式,仍必须作为可报告问题保留
- **AND** 如果同一请求还写入其他业务表、插件状态表或运行时状态表,仍必须按非预期读请求写副作用生成问题卡片

#### Scenario: 报告条目必须可追溯

- **WHEN** 报告中出现任何问题条目
- **THEN** 必须包含:接口 method + path、模块名、trace ID(或降级标注)、SQL 调用计数、write SQL 调用计数(适用时)、关键 SQL 片段、controller/service 源码 file:line 引用
- **AND** 必须给出至少一条具体的改进建议

### Requirement: 不修改生产代码与交付配置

skill 在审计期间 MAY 临时 patch 本地 `apps/lina-core/manifest/config/config.yaml` 的 `logger.path` 与 `logger.file`,但 MUST 在审计结束(无论成功失败)后恢复为原值;且 MUST NOT 修改任何源码、API DTO、SQL 文件或前端代码。

#### Scenario: logger.path/logger.file patch 与恢复

- **WHEN** Stage 0 准备阶段需要让日志落盘以便 grep
- **THEN** skill 必须备份 `logger.path` 与 `logger.file` 原值(包括空值)到 `temp/lina-perf-audit/<run-id>/meta.json`
- **AND** 临时把 `logger.path` 设置为 `temp/lina-perf-audit/<run-id>/`
- **AND** 临时把 `logger.file` 设置为 `server.log`,确保 `temp/lina-perf-audit/<run-id>/server.log` 是稳定日志入口
- **AND** 在 Stage 2 完成后(或任何阶段失败退出前)必须恢复 `logger.path` 与 `logger.file` 为备份值
- **AND** 必须在 SUMMARY.md 中标注本次执行使用的 logger.path、logger.file 与恢复结果

#### Scenario: 禁止修改任何交付资产

- **WHEN** skill 执行过程中发现潜在性能问题
- **THEN** 不得修改任何 `apps/lina-core/api/` / `apps/lina-core/internal/` / `apps/lina-plugins/` / `manifest/sql/` 下的源码或资源
- **AND** 改进建议必须以"建议在后续 OpenSpec 变更中执行..."的形式给出
- **AND** SUMMARY.md 必须明确标注 "本次执行未修改任何交付代码"

### Requirement: 跨 run 问题卡片产物

skill MUST 在每次审计结束时,把所有发现的性能问题以"一问题一文件"的形式写入仓库根目录 `perf-issues/<severity>-<module>-<slug>.md`,且 MUST 按指纹去重 —— 同一问题再次被审计命中时,SHALL 更新已有卡片而非新建文件,以便后续单独的修复变更逐个消费。问题卡片与单次 run 原始报告(`temp/lina-perf-audit/<run-id>/`)是两份独立产物,卡片 MUST 跨 run 累积保留,且 MUST NOT 写入 `temp/`。

#### Scenario: 每个性能问题写到独立卡片文件

- **WHEN** 子 agent 在某次审计中发现一个新的性能问题
- **THEN** 主 agent 在 Stage 2 汇总时 MUST 在仓库根目录 `perf-issues/` 目录下创建独立 markdown 文件
- **AND** 文件名 MUST 形如 `<severity>-<module>-<slug>.md`(severity 取 `HIGH`/`MEDIUM`/`LOW`,slug 为反模式简短描述,如 `n-plus-1-list`、`missing-index-status`)
- **AND** 文件 frontmatter MUST 包含 `id`、`severity`、`module`、`endpoint`、`status`(初值 `open`)、`first_seen_run`、`last_seen_run`、`seen_count`(初值 `1`)、`fingerprint` 字段
- **AND** 文件正文 MUST 至少包含「问题描述」、「复现方式」、「证据」、「改进方案」、「历史记录」5 个章节,每个章节 MUST 有实质内容(改进方案章节 MUST 给出至少一条具体可执行的修复步骤)
- **AND** 问题卡片正文与 `perf-issues/INDEX.md` 的描述性文本 MUST 使用中文编写；接口路径、SQL 片段、Trace-ID、fingerprint、frontmatter 字段名、状态枚举值等机器可读标识保持原值

#### Scenario: 复现方式可独立执行

- **WHEN** 任意工程师只读到某个问题卡片文件,不读其他上下文
- **THEN** 卡片中的「复现方式」章节 MUST 列出从干净环境到观察问题表象的完整命令序列(包括 `make init confirm=init rebuild=true && make mock confirm=mock`、`bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh`、`bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh`、`bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh`、对应 `curl` 请求与预期 SQL 数量等)
- **AND** MUST NOT 依赖任何当次审计 run 残留的中间状态或临时变量

#### Scenario: 指纹去重避免重复卡片

- **WHEN** 后续审计 run 再次命中已经存在卡片的同一问题
- **THEN** 主 agent MUST 用 `fingerprint = sha256(module + ":" + method + ":" + path + ":" + severity + ":" + 反模式签名)` 检索 `perf-issues/` 下既有卡片
- **AND** 命中时 MUST 更新该卡片的 `last_seen_run`、`seen_count` 与「历史记录」章节,MUST NOT 新建重复文件
- **AND** 未命中时 MUST 新建文件
- **AND** 如果命中但卡片 `status: fixed`,MUST 把状态改回 `open` 并在历史记录中追加"被再次观察到 (回归)"的条目

#### Scenario: 问题卡片与 run 报告交叉引用

- **WHEN** 主 agent 完成 Stage 2 汇总
- **THEN** 当次 run 的 `temp/lina-perf-audit/<run-id>/SUMMARY.md` MUST 引用本次新建或更新的所有问题卡片相对路径
- **AND** 每个 `perf-issues/*.md` 的「历史记录」章节 MUST 引用最近一次发现该问题的 run-id 与对应的 `audits/<module-or-shard>.md` 相对路径
- **AND** `perf-issues/INDEX.md` MUST 自动重新生成,列出当前所有 `status: open` 与 `status: in-progress` 的卡片,按严重度倒序

#### Scenario: 问题卡片保留在根目录但不进入 OpenSpec 归档

- **WHEN** 本变更归档,或日常 git 操作发生
- **THEN** `perf-issues/` MUST 位于仓库根目录,不得放入 `temp/` 或被 `temp/` 清理流程覆盖
- **AND** `perf-issues/` MUST NOT 被 `.gitignore` 忽略,以免跨 run 问题卡片被当作临时产物遗忘
- **AND** OpenSpec 归档流程 MUST NOT 把任何卡片文件搬入 `openspec/changes/archive/` 或 `openspec/specs/`
- **AND** 卡片本身随后续修复变更被人工标记 `status: fixed` 或 `status: obsolete`,但卡片管理 MUST NOT 引入额外的状态机自动化

### Requirement: stress fixture 仅用于审计

stress fixture 数据 MUST 仅在审计期间存在,MUST NOT 写入 `apps/lina-core/manifest/sql/mock-data/` 等任何交付目录,审计结束后 SHALL 随下次 reset DB 自动清除。

#### Scenario: stress fixture 通过独立脚本生成

- **WHEN** Stage 0 准备阶段需要补充压力数据
- **THEN** skill 必须调用 `.agents/skills/lina-perf-audit/scripts/stress-fixture.sh`
- **AND** 该脚本必须在宿主 `make mock confirm=mock` 与所有内置插件安装/插件 mock 数据加载完成后才执行
- **AND** 必须按依赖顺序补充资源,使用幂等的 INSERT 方式(`INSERT IGNORE` 或前置存在性判断)
- **AND** 必须不向 `apps/lina-core/manifest/sql/` 或 `apps/lina-core/manifest/sql/mock-data/` 写入任何文件
