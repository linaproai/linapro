## Context

LinaPro 当前后端 API 规模(截至本提案):

```
lina-core/api                         → 宿主 v1 接口
lina-plugins/*/backend/api            → 所有内置插件后端接口
审计范围                              → 宿主 + 所有内置插件
```

trace 与日志基础设施:

- `apps/lina-core/manifest/config/config.yaml` 中 `logger.extensions.traceIDEnabled=true`,日志保留 trace ID。
- `database.default.debug=true`,`gdb` 会把每条 SQL 输出到日志,前缀带 trace ID。
- `apps/lina-core/pkg/logger/logger_config.go` 已经实现 trace ID 文本/结构化双格式 handler,无需改动。
- GoFrame v2 在 `net/ghttp/ghttp_response.go:136` 的 `Response.Flush()` 中默认把当前 trace ID 写入响应头 `Trace-ID`(常量定义在 `net/ghttp/ghttp.go:160`)。

业务约束:

- 项目处于 SDD 驱动的快速迭代期,接口数量持续增长,需要可重复运行的审查能力,而不是一次性脚本。
- 当前 mock 数据量较小(`apps/lina-core/manifest/sql/mock-data/` 下的 seed 一般每张表几条到十几条),不足以让 N+1 在 SQL 调用次数上呈现可观察的差异。
- 项目 CLAUDE.md 已经规定:同一活跃迭代内的 OpenSpec 文档默认保持同一语言,本次会话使用中文,因此本变更活跃期 proposal/design/specs/tasks 均使用中文,归档时再统一转写为英文。

利益相关者:

- 平台维护者:需要持续看到接口性能基线和退化趋势。
- 插件作者:需要在新插件接入前自查是否引入 N+1。
- 业务开发者:需要明确"什么样的写法会被审查命中"。

## Goals / Non-Goals

**Goals:**

- 提供一个**手动触发**的 agent skill,把"环境准备 → 安装并启用所有内置插件 → 接口枚举 → 串行调用 → trace 反查 → 静态对照 → 报告"的工作流固化下来,并同时检查 GET/查询接口是否在请求执行中产生写副作用。
- 子 agent 之间通过接口任务边界解耦,主 agent 只负责调度和汇总;每个 API 接口审查任务必须由子 agent 执行,每个子 agent 上下文窗口仅承载本 shard 上下文,避免主上下文溢出。
- 报告产物结构稳定:每个模块或接口 shard 一个 markdown,加一份汇总,问题按 HIGH / MEDIUM / LOW 分级,带证据(SQL 序列、源码引用、复现接口路径)和改进建议。
- 不依赖任何新的中间件、不修改 GoFrame 行为、不污染交付配置(只在审计期间临时 patch logger.path 与 logger.file)。
- skill 描述与 spec 中明确"manual-trigger-only"约束,任何自动化路径不得引用本 skill。

**Non-Goals:**

- 不修复任何已知或新发现的接口性能问题,所有修复由独立 OpenSpec 变更后续承接。
- 不构建实时性能监控(本 skill 是离线审查工具,不是 APM)。
- 不替代 Go pprof / 数据库 EXPLAIN 等专业工具,只做"代码 + SQL 调用模式"层面的常见反模式审查。
- 不审查前端性能(本 skill 仅覆盖后端 API)。
- 不归档实际审计报告:报告每次执行都会重新生成,只归档 skill 能力定义和辅助脚本。

## Decisions

### Decision 1: 沉淀为 agent skill,而不是 hack 下的工具脚本

**选择**:把整个工作流写成 `.agents/skills/lina-perf-audit/SKILL.md`,辅助脚本作为 skill bundled resources 放在 `.agents/skills/lina-perf-audit/scripts/`。

**理由**:

1. 工作流核心价值在于"主 agent 编排 + 子 agent 拆分 + 报告汇总"的智能调度,纯 shell 脚本无法表达 sub-agent 之间的任务分发和结果合并。
2. skill 与现有 `lina-review` / `lina-feedback` 治理 skill 同源,保持研发流程一致性。
3. `.agents/skills/` 是项目内通用 skill 根目录,Claude Code 可通过 `.claude/skills` symlink 读取,Codex 与其他 AI Coding 工具也可直接读取。
4. 辅助脚本(stress fixture、环境准备、内置插件准备、endpoint 扫描)是确定性逻辑,放在 skill 自身 `scripts/` 下能让技能的说明、引用资料与可执行入口闭环维护,避免仓库级 `hack/` 目录承担 skill 私有实现细节。

**备选**:仓库根单脚本 `perf-audit.sh` 方案 —— 否决,因为无法表达 sub-agent 调度,也会把 skill 私有实现散落到 skill 边界外。

### Decision 2: 子 agent 颗粒度以接口任务为最小执行单元,主 agent 只做调度与汇总

**选择**:主 agent 不串行审查具体接口,只负责 Stage 0 准备、endpoint catalog 分片、子 agent 状态跟踪与 Stage 2 汇总。每个 API 接口审查任务必须交给子 agent 执行;默认以模块/资源生成任务队列,但当模块接口数量较多或 prompt 超过 5KB 时,必须继续拆成单接口或小接口组 shard,确保每个接口都由某个子 agent 负责。

**理由**:

1. 用户明确要求避免主任务因 API 数量过多而上下文溢出,因此具体接口调用、trace 反查、SQL 计数与源码对照必须下放到子 agent。
2. 同一模块的接口仍共享源码与 fixture 背景,所以主 agent 可以先按模块/资源生成任务队列,再按接口数量和 prompt 大小拆分为子 agent shard,兼顾上下文控制和 fixture 闭环。
3. 破坏性接口需要与同模块 create 接口保持在同一 shard 或由同一子 agent 自治,避免跨 agent 共享临时资源 ID。
4. 每个子 agent 输出 ~3-5KB markdown,主 agent 只读取审计摘要与问题条目,汇总上下文可控。

**备选**:主 agent 串行审查全部接口 —— 否决,接口数量持续增长时会引发上下文窗口溢出。

**备选**:完全固定每接口一个 agent —— 部分采用。对接口很多的模块或超大 prompt,必须拆到单接口;对小模块,允许一个子 agent 负责少量接口,但报告必须逐接口列明责任归属。

**备选**:按风险维度聚合(列表查询/写操作/文件操作各一组) —— 否决,跨模块调用源码会爆炸;模块/资源边界更适合保持 fixture 与源码上下文局部性。

### Decision 2.1: 审计范围覆盖所有内置插件

**选择**:Stage 0 必须扫描 `apps/lina-plugins/*/plugin.yaml`,同步、安装并启用所有内置插件;插件带有 `manifest/sql/mock-data/` 时,安装阶段显式加载插件 mock 数据。endpoint catalog 仍以 `apps/lina-core/api` 与 `apps/lina-plugins/*/backend/api` 的 DTO 元数据为基准,并在运行前通过接口探测确认路由可访问。

**理由**:

1. 用户明确要求本 skill 覆盖所有内置插件,不能只审计当前已经安装或启用的插件。
2. 插件 API 的权限菜单、插件自有表和 mock 数据通常由插件安装流程写入,仅执行宿主 `make init` / `make mock` 不足以让插件接口进入可审计状态。
3. 通过宿主插件管理 API 安装/启用插件,比直接执行插件 SQL 更接近真实运行路径,也能覆盖权限、菜单、生命周期和动态插件治理逻辑。

**约束**:

- 内置插件安装/启用失败时,skill 必须在 Stage 0 失败并输出失败插件清单,不得静默跳过。
- 没有后端 API 的内置插件必须写入 `meta.json` 的 skipped plugin 列表,说明原因是 "no backend API",不能误报为审计失败。

### Decision 3: trace ID 取得方式直接读响应头 `Trace-ID`,不引入审计专用中间件

**选择**:子 agent 用 `curl -i`(或等价方式)拿响应头 `Trace-ID`,然后用该 ID grep 临时日志文件。

**理由**:

1. GoFrame v2 默认在 `Response.Flush()` 中把 trace ID 写到响应头,验证已经过 `/Users/john/go/pkg/mod/github.com/gogf/gf/v2@.../net/ghttp/ghttp_response.go:136`。
2. 不需要新增任何 middleware,不修改生产配置,不引入审计专用的 build tag 或环境变量分支。
3. 只在异常路径(非 Flush 终结的请求,例如 panic 拦截前断开的链路)可能拿不到响应头,这类情况本来就不在性能审查的关注点上。

**备选**:加 audit-only middleware 写自定义响应头 —— 否决,既然 GoFrame 默认已经具备能力,新增中间件就是冗余,而且会引入"审计模式 vs 普通模式"的二元配置维护负担。

**备选**:让请求自带 X-Probe-Id 反查日志 —— 否决,需要请求 → 日志 → trace ID → 再查 SQL 三跳,而响应头方式只需一跳。

### Decision 4: 破坏性接口由子 agent 自治 create→操作→delete 自身

**选择**:对 DELETE / 卸载 / 清空类接口,子 agent 在 prompt 模板中执行"先 POST 创建一条目标资源,记录返回 ID,再 DELETE 该 ID"的闭环。

**理由**:

1. 比"每个接口跑完 reset DB"成本低很多(reset DB 通常需要数秒到十几秒,按接口逐个重置会把大量时间浪费在等待上,自治 fixture 几乎无额外耗时)。
2. 比"破坏性接口跳过人工补"完整,不留尾巴。
3. 同模块内 create 与 delete 接口本来就成对存在,子 agent 拿到的源码已经包含两侧契约,模板易于复用。

**约束**:

- skill prompt 模板必须显式列出"破坏性接口治理流程",并要求子 agent 报告中标注"通过自治 fixture 完成"。
- 对**没有对应 create 接口的破坏性操作**(例如清空全部登录日志、卸载某个具体内置插件),子 agent 必须标注 SKIPPED 并在汇总中归类为"需人工补审"。

### Decision 5: 审计前补一批 stress-fixture,让 N+1 可观察

**选择**:`.agents/skills/lina-perf-audit/scripts/stress-fixture.sh` 在宿主 `make mock confirm=mock` 与所有内置插件安装/插件 mock 数据加载完成之后,向若干列表型核心资源额外插入 50~100 条数据。

**理由**:

1. 当前 mock 数据通常每张表 5~10 条,N+1 与正常调用在 SQL 调用次数上几乎一致,无法用调用次数(NumberOfQueries ≈ NumberOfRows)这类直观特征识别。
2. 50~100 条不会显著拖慢接口,不影响审计耗时,但足够让"循环内每行调一次 dao"的 SQL 调用次数从 5~10 跳到 50~100,任何 grep `count` 的报告都能立刻看到差异。
3. stress-fixture 只在审计期间使用,不进 `manifest/sql/mock-data/`,不污染日常 demo 数据。

**目标资源清单**(初稿,可调):

```
sys_user                 → 100 条
sys_dict_type            → 50 条
sys_dict_data            → 每个 type 下 20 条 ≈ 1000 条
sys_role                 → 30 条
sys_menu                 → 50 条
sys_user_msg             → 100 条
sys_job / sys_job_log    → 50 / 200 条
plugin: sys_dept         → 100 条
plugin: sys_post         → 50 条
plugin: sys_notice       → 100 条
plugin: sys_login_log    → 200 条
```

### Decision 6: 双层产物 —— run 级原始报告 + 跨 run 的问题卡片

skill 必须同时维护**两份独立**的产物体系:

**A. 单次 run 原始报告**(每次执行生成新一份,可被对比/丢弃):

```
temp/lina-perf-audit/<run-id>/
  catalog.json              全量接口清单
  fixtures.json             资源 ID 映射
  server.log                临时落盘日志(setup 阶段把 logger.file 临时固定为 server.log)
  audits/<module-or-shard>.md 每个模块或接口 shard 一份(子 agent 直接产出)
  SUMMARY.md                汇总(主 agent 合并)
  meta.json                 运行环境、起止时间、git commit、开关位
```

`<run-id> = YYYYMMDD-HHMMSS`。每次审计独立目录,可保留多份历史结果对比。

**B. 跨 run 累积的"问题卡片"**(供后续修复跟进):

```
perf-issues/
  HIGH-<module>-<slug>.md       一个文件 = 一个性能问题
  MEDIUM-<module>-<slug>.md
  LOW-<module>-<slug>.md
  INDEX.md                      问题索引(自动生成)
```

每个问题文件结构(模板):

```markdown
---
id: HIGH-user-list-n-plus-1
severity: HIGH
module: user
endpoint: GET /user/list
status: open                    # open | in-progress | fixed | obsolete
first_seen_run: 20260501-153012
last_seen_run: 20260501-153012
seen_count: 1
fingerprint: <sha-of-module+endpoint+pattern>
---

## 问题描述
<- 一段话说明问题表现与影响 ->

## 复现方式
1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id <run-id>`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/<run-id>`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/<run-id>`
5. 登录后 `curl -i 'http://127.0.0.1:8080/user?pageSize=100'`
6. 在临时日志中观察 SQL 数量 ≈ 100,而非常数

## 证据
- trace ID: `xxxxxxxxxxxxxxxx`
- SQL 序列(节选):...
- 源码位置: `apps/lina-core/internal/service/user/user.go:142`(循环内 dao.Dept.Get)

## 改进方案
1. 把循环内的 `dao.Dept.Get(deptId)` 改为预加载: 收集所有 deptId,一次 `dao.Dept.Where(dao.Dept.Columns().Id, gset.NewFrom(ids).Slice()).All()`,再用 map 索引回填。
2. 同步检查 role/post 字段填充逻辑是否存在同类问题。
3. 增加 e2e 用例 TCxxxx 验证修复后 SQL 数量为常数。

## 历史记录
- 20260501-153012: 首次发现(SQL 数量 102,样本 100 条)
```

**指纹与去重规则**:

- `fingerprint = sha256(module + ":" + method + ":" + path + ":" + 严重度 + ":" + 反模式签名)`
- 反模式签名由 skill 根据 SQL 调用次数特征、源码 file:line、问题类型(N+1 / 缺分页 / 缺索引等)归一化生成。
- 写卡片前必须用 fingerprint 查 `perf-issues/` 下既有文件:命中则**更新**(`last_seen_run`、`seen_count` 自增、追加历史记录条目);未命中则新建文件。

**与 run 报告的关系**:

- run 报告(audits/<module>.md、SUMMARY.md)是**当次执行的快照**,可被 .gitignore 后丢弃或保留。
- 问题卡片是**跨 run 累积的修复跟进资产**,即便清理了所有 run 报告,卡片本身仍保留在根目录 `perf-issues/`;由后续单独的 OpenSpec 变更逐个消费(标 fixed 或删除)。
- 主 agent 在 Stage 2 汇总时,**先**生成/更新 `perf-issues/*.md`,**再**把卡片清单的链接写入当次 run 的 SUMMARY.md。

**理由**:

1. 单次 run 原始报告仍放在 `temp/` 下,作为可清理的执行快照。
2. 跨 run 问题卡片放在根目录 `perf-issues/`,避免被清理 `temp/` 时误删,也方便后续修复变更直接引用。
3. 单次 run 报告与跨 run 问题卡片职责清晰分离:前者是"这次跑了什么",后者是"还有哪些没修"。
4. 卡片文件名编码严重度 + 模块 + slug,文件管理器列表里就能按严重度排序。
5. 指纹去重避免重复跑审计后产生大量近似卡片,让"已知未修"的问题数量稳定可统计。
6. OpenSpec 归档时不搬运问题卡片;skill 能力定义本身才是可归档的资产。

### Decision 7: 报告问题严重度三级分类

**HIGH**(必须立即修):

- 列表/详情接口出现 N+1(SQL 调用次数 ≈ 返回行数 + 常数),且能在源码中定位到循环内 dao 调用。
- 缺索引导致大表全表扫描(SQL 涉及 WHERE 字段在 schema 中无索引)。
- 接口响应时间 > 1s 且非导出/批量类。
- 在循环内执行远程调用、文件 I/O、跨数据库事务等阻塞操作。
- GET/查询/详情/选项类读接口在自身 trace 中执行 INSERT、UPDATE、DELETE、REPLACE、TRUNCATE、ALTER、DROP、CREATE 等写 SQL。

**MEDIUM**(尽快修):

- N+1 但行数小(< 20),当前未爆雷但样本扩大后会爆。
- 列表接口缺分页(返回全表)。
- 同一请求内重复读取相同数据(缺请求级缓存)。
- 多次 SELECT 可合并为 JOIN/WHERE IN。

**LOW**(可观察/择期优化):

- SQL 数量略多但都命中索引、单条耗时 < 5ms。
- 可下推到数据库的过滤被搬到了应用层。
- 静态分析提示但实际调用未观察到性能问题。
- 静态分析提示 GET/查询接口可能在条件分支中写库,但本次采样 trace 未触发该分支。

### Decision 7.1: 读请求写副作用作为审计发现,不归入性能豁免

**选择**:所有 GET、list、query、tree、options、count、detail、current 等读语义接口都必须检查 SQL trace 中是否出现写操作。只要当前被审计接口自己的 trace 中出现写 SQL,即使响应很快也记录为 HIGH。

**理由**:

1. 查询请求产生写操作违反 RESTful 语义和项目 API 规范,会破坏重试、缓存、预取、爬虫访问和只读事务假设。
2. 这类问题通常能通过同一套 trace ID + SQL 日志链路发现,与性能审查共享调用、日志和源码对照成本。
3. 即使写操作单次耗时很低,也可能带来锁竞争、审计误差、用户数据变更和幂等性问题,不能按 LOW 性能优化处理。

**判定约束**:

- 只统计被审计接口 trace 内的 SQL,不得把 Stage 0 安装插件、登录、stress fixture 或破坏性接口自治 create/delete 的写 SQL 算入读接口。
- 如果源码存在写库分支但本次 runtime trace 未命中,最多标为 LOW 静态风险,不得伪造 runtime 证据。
- 改进建议必须指向拆分接口语义:把写副作用移动到 POST/PUT/DELETE 动作、拆分 read 与 mark-as-read 等操作,或改为非持久化读取模型。

### Decision 8: 强制约束 manual-trigger-only,在 SKILL.md 与 spec 中双重声明

**选择**:

- SKILL.md frontmatter 的 description 字段第一段就声明 "Manually-invoked... ⚠️ MANUAL TRIGGER ONLY"。
- SKILL.md 内部专门一节 "When NOT to Use",列出禁止场景。
- spec 中以 Requirement + Scenario 形式编码本约束,审查 skill 也将检查任何其他 skill 是否引用本 skill。

**理由**:

- 完整跑一次审计涉及:reset DB、安装并启用所有内置插件、补 stress fixture、重启后端服务、多个子 agent 串行执行(总耗时几十分钟到数小时,token 成本显著)。如果被 lina-review 或类似 skill 自动触发,会在用户毫无预期的情况下消耗大量资源、且破坏本地数据库状态。
- 含糊请求("接口好像有点慢"、"性能怎么样")必须先与用户确认,不能自动展开完整审计。

## Risks / Trade-offs

| 风险 | 应对 |
|---|---|
| GoFrame Server 默认 tracer 在某些极端路径(panic 在 Flush 之前发生)可能不写响应头,导致拿不到 trace ID | 子 agent 在拿不到响应头时回退到"按时间窗 + URL 模糊匹配"找日志,并在该接口报告中标注 "trace ID unavailable, evidence quality reduced" |
| stress-fixture 数据补完后接口耗时变长,可能让其他 agent 等待超时 | stress-fixture 控制在 100 条量级以内;子 agent 调用接口设单次 30s 超时,超时则记录但继续 |
| 子 agent 之间高并发可能导致后端日志交错,grep trace ID 时出现跨 agent 的串扰 | trace ID 是请求级唯一,本身就是用来抗交错的,grep 不会混淆 |
| 主 agent 上下文窗口仍有压力(多个 agent × 报告片段) | 单 agent 报告硬性约束 < 5KB,主 agent 只读 SUMMARY 部分汇总;详细证据保留在文件中按需读取 |
| stress-fixture 与 mock 数据有外键冲突(例如插入用户引用了不存在的部门) | fixture 脚本必须在宿主 mock 与内置插件安装/插件 mock 数据加载完成后运行,优先按依赖顺序补,且使用 `INSERT IGNORE`;失败时报错退出而不是静默跳过 |
| 用户重复触发本 skill,产生大量 temp 目录 | skill 在每次执行结束后提示用户"上次产物在 temp/lina-perf-audit/...",并提供清理命令(由用户决定是否清);本 skill 自身不主动删除历史产物 |
| 子 agent 创建 fixture 后未能 delete(网络异常、断电等),导致数据残留 | fixture create-delete 用 `defer` 模式,子 agent 退出前必须执行清理;残留数据通过下次 reset DB 自动清除,不影响审计正确性 |

## Migration Plan

本变更**纯增量**,不修改任何现有运行时行为或交付配置:

1. 新增 `.agents/skills/lina-perf-audit/SKILL.md` 与配套 references。
2. 新增 `.agents/skills/lina-perf-audit/scripts/` 目录与脚本,作为 skill 内部 bundled resources 维护。
3. 新增 `openspec/specs/lina-perf-audit-skill/spec.md`(归档时落入)。

无回滚需求 —— skill 不被任何代码引用,移除即停用。

## Open Questions

无 —— 所有关键决策已在探索阶段对齐。
