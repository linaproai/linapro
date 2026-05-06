# lina-perf-audit 技能规范

## 目的

定义手动触发的 `lina-perf-audit` 代理技能的治理契约，该技能审计 LinaPro 后端 API 的性能风险和读请求写副作用，覆盖宿主和所有内置插件。

## 需求

### 需求：仅限手动触发

`lina-perf-audit` 技能 SHALL 声明为仅限手动触发，不得由自动化、CI/CD 管线、定时任务、Git 钩子或其他技能调用。模糊的性能请求必须在运行任何破坏性审计准备前要求用户确认。

#### 场景：SKILL.md 前置元数据声明触发约束

- **当** `.agents/skills/lina-perf-audit/SKILL.md` 被加载时
- **则** 其描述包含 `MANUAL TRIGGER ONLY` 短语
- **且** 列出预期的资源成本，包括数据库重置、服务重启、内置插件启用、压力固件、子代理扇出、耗时和 Token 成本
- **且** 列出禁止的调用路径，包括其他技能、CI、定时任务、Git 钩子和模糊的性能请求

#### 场景：模糊请求需要确认

- **当** 用户说了一些模糊的话，如 `API 好像很慢`、`性能怎么样` 或 `检查接口性能` 时
- **则** 技能在运行任何 Stage 0 设置命令前询问是否启动完整审计
- **且** 确认文本提及数据库重置、服务重启、耗时、子代理扇出和 Token 成本
- **且** 技能在确认前不运行 `make stop`、`make init`、`make mock`、`setup-audit-env.sh`、`prepare-builtin-plugins.sh` 或 `stress-fixture.sh`

#### 场景：其他技能不自动调用此技能

- **当** 另一个技能如 `lina-review`、`lina-feedback` 或 `lina-e2e` 观察到性能问题时
- **则** 该技能不得调用 `lina-perf-audit`
- **且** 只能建议用户手动触发 `lina-perf-audit`

### 需求：三阶段审计工作流

`lina-perf-audit` 技能必须在三个阶段执行完整审计：Stage 0 准备、Stage 1 并发子代理审计和 Stage 2 汇总加问题卡片聚合。单次运行产物必须写入 `temp/lina-perf-audit/<run-id>/` 下。

#### 场景：辅助脚本保持在技能边界内

- **当** 设置、插件准备、端点扫描、固件探测、压力固件或报告聚合需要确定性辅助逻辑时
- **则** 脚本位于 `.agents/skills/lina-perf-audit/scripts/` 下
- **且** `SKILL.md`、参考文档、OpenSpec 文档和问题卡片模板使用技能自有脚本路径
- **且** 不在 `hack/` 下维护技能私有脚本的副本

#### 场景：Stage 0 准备完整审计环境

- **当** 用户确认完整审计后
- **则** 技能创建 `YYYYMMDD-HHMMSS` 格式的唯一 `run-id`
- **且** 停止服务，通过 `make init confirm=init rebuild=true` 和 `make mock confirm=mock` 重置本地数据库，通过 `setup-audit-env.sh` 补丁审计日志，通过 `prepare-builtin-plugins.sh` 安装并启用所有内置插件，通过 `stress-fixture.sh` 添加审计专用压力数据，通过 `scan-endpoints.sh` 扫描端点，通过 `probe-fixtures.sh` 探测固件
- **且** 所有生成产物写入 `temp/lina-perf-audit/<run-id>/` 下
- **且** 通过 `restore-audit-env.sh` 在成功或失败时恢复临时日志设置

#### 场景：Stage 1 使用子代理处理端点任务

- **当** Stage 0 完成后
- **则** 技能从 `catalog.json` 构建宿主和内置插件模块的端点任务
- **且** 每个 API 端点审计任务由子代理执行，而非主代理串行执行
- **且** 大模块拆分为端点或小模块分片，确保每个子代理提示保持在配置的提示预算以下
- **且** 每个子代理将其分配的审计输出精确写入 `temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`

#### 场景：Stage 2 产出稳定输出

- **当** 所有子代理完成后
- **则** 主代理运行报告聚合
- **且** `SUMMARY.md` 按 HIGH、MEDIUM 和 LOW 严重程度列出发现
- **且** `meta.json` 记录运行时序、Git 提交、压力固件状态、子代理状态、跳过的插件、日志设置和恢复结果
- **且** 每个新增或更新的持久问题卡片都从运行摘要链接

### 需求：覆盖所有内置插件

技能必须审计仓库中每个内置插件的后端 API，不得将覆盖范围限制在审计开始前已安装或已启用的插件。

#### 场景：Stage 0 安装并启用内置插件

- **当** Stage 0 准备审计环境时
- **则** 从 `apps/lina-plugins/*/plugin.yaml` 发现插件
- **且** 通过宿主插件管理 API 同步、安装和启用每个内置插件
- **且** 当插件提供 `manifest/sql/mock-data/` 时加载插件模拟数据
- **且** 任何安装或启用失败使 Stage 0 失败，并列出失败插件
- **且** 成功发现但无后端 API 的插件记录为跳过，原因为 `no backend API`

#### 场景：端点目录包含宿主和插件 DTO

- **当** `scan-endpoints.sh` 生成 `catalog.json` 时
- **则** 目录包含来自 `apps/lina-core/api/**/v1/*.go` 和 `apps/lina-plugins/*/backend/api/**/v1/*.go` 的路由元数据
- **且** 声明的插件路由在运行时不可达会使设置或固件探测失败，而非静默进入正常审计输出

### 需求：基于 Trace-ID 的 SQL 证据

子代理必须使用 GoFrame 默认的 `Trace-ID` 响应头将端点调用与 SQL 日志行关联。技能不得添加审计专用中间件、自定义响应头或生产行为变更。

#### 场景：子代理从响应获取 trace ID

- **当** 子代理调用端点时
- **则** 读取 `Trace-ID` 响应头
- **且** 使用该 trace ID 在 `temp/lina-perf-audit/<run-id>/server.log` 中查找匹配的 SQL 行
- **且** 不依赖新中间件或自定义头

#### 场景：Trace ID 不可用

- **当** 端点响应不包含 `Trace-ID` 时
- **则** 子代理按请求时间窗口和请求 URL 搜索
- **且** 审计条目注明 `trace ID unavailable, evidence quality reduced`
- **且** 端点不会仅因 trace ID 缺失而被跳过

### 需求：破坏性端点使用自主固件处理

子代理在调用 DELETE、卸载、清除、重置或等效操作等破坏性端点时，必须避免损坏共享审计数据。

#### 场景：破坏性端点有匹配的创建端点

- **当** 子代理审计破坏性端点且同一模块有匹配的创建端点时
- **则** 先创建专用审计固件
- **且** 仅对该固件调用破坏性端点
- **且** 即使破坏性调用失败也尝试清理
- **且** 报告说明自主固件完成且未污染共享数据

#### 场景：破坏性端点无匹配的创建端点

- **当** 破坏性操作无同模块创建端点时
- **则** 子代理将端点标记为 `SKIPPED: no matching create endpoint, manual follow-up required`
- **且** 端点出现在 `SUMMARY.md` 的手动跟进部分
- **且** 子代理不使用其他模块的资源作为替代

### 需求：严重程度分类和读请求副作用检测

技能必须将发现分类为 HIGH、MEDIUM 或 LOW，并包含证据和整改建议。审计必须检查性能风险和执行意外写 SQL 的读/查询端点。

#### 场景：分配 HIGH 严重程度

- **当** 端点存在带源证据的列表/详情 N+1、潜在大数据量的缺失索引、超过 1 秒的非批量响应、阻塞循环工作或读/查询端点 trace 中的意外写 SQL 时
- **则** 发现标记为 `severity: HIGH`
- **且** `SUMMARY.md` 在 HIGH 部分列出
- **且** 整改包括具体的批处理、分页、索引、异步或端点语义修复

#### 场景：分配 MEDIUM 严重程度

- **当** 端点显示小样本 N+1、缺失分页、重复相同数据读取或应通过 `JOIN` 或 `WHERE IN` 合并的多个 SELECT 调用时
- **则** 发现标记为 `severity: MEDIUM`

#### 场景：分配 LOW 严重程度

- **当** 端点在快速索引查询下 SQL 计数略高、可下推的应用层过滤或仅在运行时未观察到的静态证据时
- **则** 发现标记为 `severity: LOW`
- **且** 整改说明问题为观察性或较低优先级

#### 场景：检测读请求写副作用

- **当** 子代理调用 GET、列表、查询、树、选项、计数、详情、当前或等效读端点时
- **则** 检查端点 trace 中首个重要标记为 `INSERT`、`UPDATE`、`DELETE`、`REPLACE`、`TRUNCATE`、`ALTER`、`DROP` 或 `CREATE` 的写 SQL
- **且** 意外写操作报告为 HIGH，反模式签名为 `read-write-side-effect`
- **且** Stage 0 设置、登录、压力固件写入和自主固件创建/删除调用不计入被审计读端点的写操作
- **且** 如果 trace 包含读 SQL 且每条写语句仅触及 `sys_online_session` 或 `plugin_monitor_operlog`，写操作记录为预期的运行 PASS 注释，不创建发现、摘要违规或问题卡片
- **且** 如果 trace 写入其他业务、插件状态、运行时状态或存储表，发现保持可报告

#### 场景：发现可追溯

- **当** 报告包含发现时
- **则** 包含方法和路径、模块名、trace ID 或回退标记、SQL 计数、适用时的写 SQL 计数、关键 SQL 摘录、源文件和行号以及至少一个具体整改建议

### 需求：审计不修改生产交付资产

技能可以在审计运行期间临时补丁本地日志输出，但必须恢复原始日志设置，且不得修改源代码、API DTO、SQL 交付资产、前端代码或默认运行时配置。

#### 场景：日志设置被补丁并恢复

- **当** Stage 0 需要稳定日志输出用于 SQL 证据时
- **则** `setup-audit-env.sh` 备份原始 `logger.path` 和 `logger.file`
- **且** 将 `logger.path` 设置为运行目录，`logger.file` 设置为 `server.log`
- **且** `restore-audit-env.sh` 在成功或失败时恢复精确的原始值
- **且** `SUMMARY.md` 记录日志路径、日志文件和恢复结果

#### 场景：交付资产不被审计更改

- **当** 技能发现性能问题时
- **则** 仅写入报告和问题卡片
- **且** 不修改 `apps/lina-core/api/`、`apps/lina-core/internal/`、`apps/lina-plugins/`、`manifest/sql/` 或前端源码
- **且** 整改文本说明修复应通过后续 OpenSpec 变更实施

### 需求：持久跨运行问题卡片

技能必须将每个发现的性能或读副作用问题写入仓库根目录 `perf-issues/` 下的持久 markdown 卡片，并在运行间按指纹去重。

#### 场景：新问题创建一张卡片

- **当** Stage 2 观察到新发现时
- **则** 创建 `perf-issues/<severity>-<module>-<slug>.md`
- **且** 卡片前置元数据包含 `id`、`severity`、`module`、`endpoint`、`status`、`first_seen_run`、`last_seen_run`、`seen_count` 和 `fingerprint`
- **且** 卡片正文包含 `问题描述`、`复现方式`、`证据`、`改进方案` 和 `历史记录` 部分
- **且** 描述性卡片文本和 `perf-issues/INDEX.md` 标题使用中文，而 API 路径、SQL 摘录、Trace ID、指纹、前置元数据字段名和状态值保持机器可读的原始值

#### 场景：复现步骤自包含

- **当** 工程师仅阅读单张问题卡片时
- **则** `复现方式` 部分给出从干净本地环境到端点请求和预期 SQL 观察所需的命令
- **且** 不依赖原始运行中的未声明临时变量

#### 场景：现有指纹更新现有卡片

- **当** 后续运行观察到相同指纹的发现时
- **则** Stage 2 更新 `last_seen_run`、递增 `seen_count` 并追加历史
- **且** 不创建重复文件
- **且** 如果现有卡片有 `status: fixed` 或 `status: obsolete`，将状态改回 `open` 并记录回归历史条目

#### 场景：问题卡片和运行报告相互引用

- **当** Stage 2 完成时
- **则** 运行 `SUMMARY.md` 链接到该运行中创建或更新的所有卡片
- **且** 每个卡片历史条目引用运行 ID 和相关的 `audits/<module-or-shard>.md` 路径
- **且** `perf-issues/INDEX.md` 按严重程度列出所有 `open` 和 `in-progress` 卡片

#### 场景：问题卡片不是 OpenSpec 归档产物

- **当** OpenSpec 变更归档时
- **则** `perf-issues/` 保留在仓库根目录
- **且** OpenSpec 归档不将问题卡片移入 `openspec/changes/archive/` 或 `openspec/specs/`

### 需求：压力固件仅限审计使用

压力固件数据必须仅在审计运行期间存在，不得写入宿主或插件交付 SQL 目录。

#### 场景：压力固件脚本生成临时规模数据

- **当** Stage 0 需要额外行以使 N+1 行为可观察时
- **则** 技能运行 `.agents/skills/lina-perf-audit/scripts/stress-fixture.sh`
- **且** 脚本在宿主模拟数据和插件模拟数据就绪后运行
- **且** 使用 `INSERT IGNORE` 或前置存在性检查等幂等方式按依赖顺序插入数据
- **且** 不向 `apps/lina-core/manifest/sql/`、`apps/lina-core/manifest/sql/mock-data/` 或插件交付 SQL 目录写入文件
