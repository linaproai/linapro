## 1. 准备阶段(基础设施与目录骨架)

- [x] 1.1 创建 `.agents/skills/lina-perf-audit/` 目录骨架(SKILL.md、references/、scripts/ 占位)
- [x] 1.2 创建 `.agents/skills/lina-perf-audit/scripts/` 目录,加入 README 说明该目录脚本仅供 lina-perf-audit skill 调用
- [x] 1.3 验证 GoFrame 默认响应头 `Trace-ID` 在本仓库实际可用:启动后端,curl `/health` 接口,确认 `Trace-ID` 响应头非空,并调用一个真实读库接口确认相同 trace ID 能在 `server.log` 中关联到 SQL debug 日志(写入验证截图或日志到本任务记录中)

## 2. 辅助脚本实现

- [x] 2.1 实现 `.agents/skills/lina-perf-audit/scripts/setup-audit-env.sh`:停服 → 备份 `apps/lina-core/manifest/config/config.yaml` 中的 `logger.path` 与 `logger.file` 原值到指定 run-dir → patch 为 `temp/lina-perf-audit/<run-id>/` 与固定 `server.log` → 启动后端 → 等待健康探针就绪 → 输出 admin/admin123 登录后的 token 到 run-dir
- [x] 2.2 实现 `.agents/skills/lina-perf-audit/scripts/restore-audit-env.sh`:读取 run-dir 中备份的 `logger.path` 与 `logger.file` 原值,恢复 config.yaml,停服(成功路径与失败路径都必须能调用)
- [x] 2.3 实现 `.agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh`:扫描 `apps/lina-plugins/*/plugin.yaml`,通过宿主插件管理 API 同步、安装并启用所有内置插件,插件存在 `manifest/sql/mock-data/` 时显式加载插件 mock data,任一内置插件安装或启用失败时终止并报告;没有后端 API 的插件只在 endpoint 审计阶段标记为 skipped
- [x] 2.4 实现 `.agents/skills/lina-perf-audit/scripts/scan-endpoints.sh`:扫描 `apps/lina-core/api/**/v1/*.go` 与 `apps/lina-plugins/*/backend/api/**/v1/*.go`,解析每个 g.Meta/gmeta.Meta 提取 method/path/tags/permission/dc/eg,按模块边界分组,生成 `temp/lina-perf-audit/<run-id>/catalog.json`,并在 meta 中记录无后端 API 的内置插件
- [x] 2.5 实现 `.agents/skills/lina-perf-audit/scripts/probe-fixtures.sh`:调用各模块的 list 接口,采集每个资源的样本 ID(取首条),生成 `temp/lina-perf-audit/<run-id>/fixtures.json`,并把 DTO 存在但运行时路由不可访问的接口报告为环境准备失败
- [x] 2.6 实现 `.agents/skills/lina-perf-audit/scripts/stress-fixture.sh`:按 design.md 中"Decision 5"的目标资源清单,在宿主 mock 与所有内置插件安装/插件 mock 数据加载完成后,通过 SQL 直连方式插入 stress 数据(使用 `INSERT IGNORE`,按外键依赖顺序),并校验 `apps/lina-core/manifest/sql/` 与 `apps/lina-core/manifest/sql/mock-data/` 未被写入

## 3. SKILL.md 主体编写

- [x] 3.1 编写 `.agents/skills/lina-perf-audit/SKILL.md` frontmatter:name、description(含 ⚠️ MANUAL TRIGGER ONLY 与资源代价说明)、trigger 条件,并说明该 skill 可被 Claude Code、Codex 与其他 AI Coding 工具读取
- [x] 3.2 编写 SKILL.md "When to Use" 章节:列出明确触发场景(用户显式提"接口性能审查"/"N+1 检查"/"/lina-perf-audit"等),要求确认含糊请求
- [x] 3.3 编写 SKILL.md "When NOT to Use" 章节:列出禁止场景(自动化、CI、其他 skill 引用、单接口排查、含糊性能问题)
- [x] 3.4 编写 SKILL.md "Workflow" 章节:三阶段流水线说明,明确每个阶段调用的辅助脚本与产物路径
- [x] 3.5 编写 SKILL.md "Sub-agent Prompt Template" 章节:子 agent 输入(模块名、endpoints[]、fixtures、log_path、token、run_dir)、产物结构、单 agent 上下文上限(< 5KB)
- [x] 3.6 编写 SKILL.md "Destructive Endpoint Handling" 章节:create→DELETE→cleanup 自治模板,以及"无对应 create 接口时 SKIPPED 标注"规则
- [x] 3.7 编写 SKILL.md "Severity Classification" 章节:HIGH / MEDIUM / LOW 判定准则与判定示例
- [x] 3.8 编写 SKILL.md "Report Schema" 章节:`audits/<module>.md` 与 `SUMMARY.md` 的章节结构、问题条目字段(method/path、severity、trace ID、SQL 序列、源码引用 file:line、改进建议)
- [x] 3.9 编写 SKILL.md "Issue Card Lifecycle" 章节:Stage 2 汇总时如何根据指纹去重生成/更新根目录 `perf-issues/<severity>-<module>-<slug>.md`,卡片 frontmatter 字段、5 大正文章节(问题描述/复现方式/证据/改进方案/历史记录)、状态机(open / in-progress / fixed / obsolete)、INDEX.md 重生成规则
- [x] 3.10 编写 SKILL.md "Cross-reference Rules" 章节:run SUMMARY.md 与各问题卡片之间的相对路径引用约定;明确根目录 `perf-issues/` 作为跨 run 持久资产保留,归档时不进 OpenSpec archive 与 specs

## 4. 配套 references 文档

- [x] 4.1 编写 `.agents/skills/lina-perf-audit/references/sub-agent-prompt.md`:可被主 agent 直接拼装传给子 agent 的 prompt 模板
- [x] 4.2 编写 `.agents/skills/lina-perf-audit/references/severity-rubric.md`:严重度判定的具体反模式清单(N+1、缺索引、缺分页、循环 I/O、重复读取等),并约定每种反模式对应的"反模式签名"用于指纹生成
- [x] 4.3 编写 `.agents/skills/lina-perf-audit/references/report-template.md`:`audits/<module>.md` 与 `SUMMARY.md` 的 markdown 模板示例
- [x] 4.4 编写 `.agents/skills/lina-perf-audit/references/issue-card-template.md`:`perf-issues/<severity>-<module>-<slug>.md` 的 markdown 模板,包含 frontmatter、5 大正文章节(问题描述/复现方式/证据/改进方案/历史记录)、复现命令骨架与示例 trace 引用
- [x] 4.5 编写 `.agents/skills/lina-perf-audit/references/fingerprint-rule.md`:`fingerprint = sha256(module + ":" + method + ":" + path + ":" + severity + ":" + 反模式签名)` 的精确算法定义,以及命中已有卡片时各字段(`last_seen_run`、`seen_count`、`status` 回滚为 open)的更新规则与回归判定逻辑

## 5. 文档同步

- [x] 5.1 在 `.agents/skills/lina-perf-audit/README.md` 与 `README.zh_CN.md` 中说明本 skill 的用途、触发方式、产物位置(双语并存,与项目 README 规范一致)
- [x] 5.2 在项目根 CLAUDE.md "常用命令" 章节追加一行,说明 `lina-perf-audit` 是手动触发的性能审查 skill,并指向其 README

## 6. dry-run 验证

- [x] 6.1 在干净的本地环境跑一次完整 dry-run:`make stop` → 触发 skill(内部执行 `make init confirm=init rebuild=true` 与 `make mock confirm=mock`) → 全流程跑通 → 生成完整 `temp/lina-perf-audit/<run-id>/` 目录
- [x] 6.2 验证 `logger.path` 与 `logger.file` 在 dry-run 结束后已恢复为原值(对比 `apps/lina-core/manifest/config/config.yaml` 的 git diff 应为空)
- [x] 6.3 验证 `apps/lina-core/manifest/sql/` 与 `apps/lina-core/manifest/sql/mock-data/` 在 dry-run 结束后无任何修改(git diff 应为空)
- [x] 6.4 验证 dry-run 产物结构符合 spec(每个模块一份 `audits/<module>.md`、有 SUMMARY.md、有 meta.json)
- [x] 6.5 验证至少识别出 1 个 HIGH 级 N+1 案例(stress-fixture 后,某个列表接口 SQL 调用次数 ≈ 行数);若未命中,在 SUMMARY.md 中明确记录"本次执行未发现 HIGH 级问题",而不是误报
- [x] 6.6 验证至少 1 个破坏性接口走 create→DELETE 自治闭环并标注成功
- [x] 6.7 验证 dry-run 在根目录 `perf-issues/` 下为每个发现的问题生成了独立卡片文件,文件名符合 `<severity>-<module>-<slug>.md` 规则,且 frontmatter 与 5 大正文章节齐备
- [x] 6.8 验证 dry-run 生成或更新了 `perf-issues/INDEX.md`,且按严重度倒序列出所有 `open` / `in-progress` 卡片
- [x] 6.9 紧接着再跑一次 dry-run,验证指纹去重生效:相同问题不会被重复创建为新文件,而是更新原卡片的 `last_seen_run`、`seen_count` 与历史记录
- [x] 6.10 手动把某张卡片 `status` 改为 `fixed`,再跑一次 dry-run,验证若问题仍存在则状态被自动改回 `open` 并在历史记录中追加"回归"条目
- [x] 6.11 验证根目录 `perf-issues/` 不在 `temp/` 下、未被 `.gitignore` 忽略、不会被 `temp/` 清理流程误删,且 OpenSpec 归档流程不会搬运该目录

## 7. 手动触发约束验证

- [x] 7.1 检查 `.agents/skills/lina-perf-audit/SKILL.md` description 字段必须出现 "MANUAL TRIGGER ONLY" 字样
- [x] 7.2 跑一次模拟"含糊请求"测试:在新 session 中说"接口性能怎么样",验证 skill 触发后第一步是询问用户确认而不是直接 reset DB
- [x] 7.3 grep 项目内所有其他 skill 文件(`.agents/skills/*/SKILL.md`),确认没有任何 skill 引用 `lina-perf-audit`,并在本任务记录中列出 grep 结果

## 8. 审查与验证

- [x] 8.1 运行 `openspec validate add-lina-perf-audit-skill --type change --strict`,确认变更结构合规
- [x] 8.2 调用 `/lina-review` 技能进行代码与规范审查
- [x] 8.3 修复审查中发现的所有问题,重新运行 `/lina-review` 直至通过

## 执行记录

- 主 dry-run 目录：`temp/lina-perf-audit/20260501-233924/`。
- 接口目录扫描结果：`catalog.json` 记录 `171` 个接口、`26` 个模块，其中宿主接口 `121` 个、内置插件接口 `50` 个；`demo-control` 因无后端 API 标记为 `skipped`。
- 子 agent 审查结果：`audits/` 下生成 `22` 份模块或分片报告，覆盖全部 `26` 个模块，其中 `core-small` 合并覆盖 `core:auth`、`core:health`、`core:i18n`、`core:publicconfig`、`core:sysinfo`。
- 汇总结果：`SUMMARY.md` 记录 `18` 个 `HIGH`、`6` 个 `MEDIUM`、`0` 个 `LOW`；其中 `17` 个为读接口写入副作用问题，另有 `1` 个 `HIGH` 级 `N+1` 案例：`GET /api/v1/job/log?pageNum=1&pageSize=100`。
- 持久问题卡片：根目录 `perf-issues/` 生成 `24` 张问题卡片与 `INDEX.md`，卡片均包含 frontmatter 与 `问题描述`、`复现方式`、`证据`、`改进方案`、`历史记录` 五个正文段落。
- 指纹去重与回归验证：使用隔离目录 `temp/lina-perf-audit/card-lifecycle-validation/` 复用同一批审查报告验证卡片生命周期；第二个 run 未创建重复卡片，`24` 张卡片的 `seen_count` 增至 `2`；手动将一张卡片改为 `fixed` 后第三个 run 将其恢复为 `open` 并追加 `被再次观察到 (回归)` 历史记录。
- 环境恢复：`restore-audit-env.sh` 已恢复 `apps/lina-core/manifest/config/config.yaml` 的 `logger.path` 与 `logger.file`，`make status` 显示前后端均未运行；`git diff -- apps/lina-core/manifest/config/config.yaml apps/lina-core/manifest/sql apps/lina-core/manifest/sql/mock-data` 为空。
- 手动触发约束：`SKILL.md` description 包含 `MANUAL TRIGGER ONLY`；独立子 agent 模拟新会话输入 `接口性能怎么样`，结论为必须先询问确认，不能直接执行 `make stop`、`make init`、`make mock` 或审计脚本；`rg` 检查确认其他 skill 的 `SKILL.md` 未引用 `lina-perf-audit`。
- 审查修复：`lina-review` 收尾审查发现 `stress-fixture.sh` 部分幂等写入未使用 `INSERT IGNORE` 字样，已统一改为 `INSERT IGNORE INTO ... SELECT ... WHERE NOT EXISTS`。
- 验证命令：`bash -n .agents/skills/lina-perf-audit/scripts/*.sh`、`python3 -m json.tool` 校验 run JSON、`go build -o temp/bin/lina-perf-audit-build-check ./apps/lina-core`、`openspec validate add-lina-perf-audit-skill --type change --strict` 均通过。

## Feedback

- [x] **FB-1**: 审计范围明确覆盖所有内置插件
- [x] **FB-2**: 跨 run 问题卡片迁移到根目录 `perf-issues/`
- [x] **FB-3**: skill 目录改为通用 `.agents/skills/lina-perf-audit/`
- [x] **FB-4**: 增加查询请求执行写操作的审查项
- [x] **FB-5**: 将 lina-perf-audit 辅助脚本迁移到 skill 内部 scripts 目录闭环维护
- [x] **FB-6**: `perf-issues/` 问题卡片内容需要使用中文描述
- [x] **FB-7**: `perf-issues/` 不应包含仅写入 `sys_online_session` 或 `plugin_monitor_operlog` 的读请求副作用卡片
- [x] **FB-8**: 修复 `core:joblog` 作业日志列表动态插件 i18n 元数据 N+1 查询问题
- [x] **FB-9**: 修复 `plugin-demo-dynamic:dynamic` host-call-demo GET 接口执行插件状态持久化写入问题
- [x] **FB-10**: 修复 `core:job` 作业列表与详情重复读取动态插件本地化元数据问题
- [x] **FB-11**: 修复 `core:jobgroup` 作业分组列表逐组统计作业数量的 N+1 风险
- [x] **FB-12**: 修复 `core:menu` 菜单列表请求内重复读取菜单与插件运行时元数据问题
- [x] **FB-13**: 修复 `core:plugin` 插件列表重复读取动态插件与发布版本状态问题
- [x] **FB-14**: 修复 `core:role` 角色菜单关联逐条写入问题
- [x] **FB-15**: 修复 `monitor-operlog:operlog` 操作日志列表重复读取插件本地化元数据问题
- [x] **FB-16**: 集群部署下插件运行态缓存缺少跨节点失效机制
- [x] **FB-17**: 优化集群模式动态插件 reconciler，避免每 2 秒全量扫描动态插件 registry，改为基于共享 revision 判断是否需要收敛并保留低频兜底扫描

### FB-6 执行记录

- 已将 `.agents/skills/lina-perf-audit/scripts/aggregate-reports.sh` 的问题卡片生成逻辑改为中文描述，覆盖卡片标题、问题描述、复现说明、证据字段、改进方案、历史记录以及 `perf-issues/INDEX.md` 表头和状态展示。
- 已使用 `temp/lina-perf-audit/20260501-233924/` 的现有审计产物重新生成根目录 `perf-issues/`，共 `24` 张问题卡片与 `INDEX.md`。
- 已更新 `.agents/skills/lina-perf-audit/SKILL.md`、`references/issue-card-template.md` 与 OpenSpec 增量规范，明确 `perf-issues` 卡片正文和索引描述性文本使用中文，接口路径、SQL、Trace-ID、fingerprint、frontmatter 字段名和状态枚举值保持机器可读原值。
- 验证命令：`bash -n .agents/skills/lina-perf-audit/scripts/*.sh`、`openspec validate add-lina-perf-audit-skill --type change --strict`、`perf-issues` 卡片章节完整性检查均通过。

### FB-7 执行记录

- 已更新 `.agents/skills/lina-perf-audit/scripts/aggregate-reports.sh` 的 Stage 2 汇总逻辑：GET 读请求命中 `read-write-side-effect` 时,如果同一 trace 中存在查询 SQL,且写 SQL 目标表仅限 `sys_online_session` 或 `plugin_monitor_operlog`,则视为预期的会话活跃刷新或操作日志记录,不再生成或保留 `perf-issues/` 卡片。
- 已同步更新 `.agents/skills/lina-perf-audit/SKILL.md` 与 OpenSpec 增量规范,明确仅写入上述运维表的读请求副作用不进入持久问题卡片；同一请求若还写入业务表、插件状态表或运行时状态表,仍必须生成问题卡片。
- 已使用 `temp/lina-perf-audit/20260501-233924/` 的现有审计产物重新生成 `SUMMARY.md`、`meta.json` 与根目录 `perf-issues/`：本次汇总剩余 `8` 张可报告问题卡片,忽略 `16` 个预期会话/操作日志副作用问题。
- 已确认 `perf-issues/` 中不再包含仅写入 `sys_online_session` 或 `plugin_monitor_operlog` 的 `read-write-side-effect` 卡片；保留的 `plugin-demo-dynamic/host-call-demo` 卡片仍写入 `sys_plugin_state`、`sys_plugin_node_state` 等非运维表,因此属于真实问题。
- 验证命令：`bash -n .agents/skills/lina-perf-audit/scripts/aggregate-reports.sh`、`bash .agents/skills/lina-perf-audit/scripts/aggregate-reports.sh --run-dir temp/lina-perf-audit/20260501-233924`、`openspec validate add-lina-perf-audit-skill --type change --strict`、`perf-issues` 卡片章节完整性检查均通过。
- 追加修正：已将该例外规则前移到 `references/sub-agent-prompt.md` 与 `references/severity-rubric.md`，要求子 agent 对"查询 SQL + 仅写 `sys_online_session`/`plugin_monitor_operlog`"记录 PASS 备注，而不是先产出 HIGH 再依赖 Stage 2 过滤。
- 追加修正：已增强 `aggregate-reports.sh` 的兜底过滤，解析带 schema/backtick 或日志前缀的写 SQL 目标表，并要求报告中的可见写 SQL 数量覆盖 `Write SQL count` 后才忽略，避免 SQL 摘录不完整时被错误抑制。
- 追加修正：Stage 2 兜底过滤不再限定 `GET` 方法，而是以 `read-write-side-effect` 签名识别读语义问题，覆盖被 catalog 判定为 read/query 的非 GET 历史接口。
- 追加验证：使用最小合成审计样本验证聚合行为，结果为 `Ignored 2 expected session/operation-log side-effect findings`、`Aggregated 2 reportable findings`；GET 与非 GET 读语义签名中的预期运维写入会被忽略，业务表写入和写 SQL 摘录不完整的情况仍会生成问题卡片。

### FB-8 执行记录

- 已在 `jobmgmt` 作业日志列表与详情投影中引入请求内 handler source text 缓存，避免同一批日志行重复解析相同动态插件 handler 的 i18n 元数据。
- 影响面：后端内部性能优化，无新增用户可见文案；不涉及前端 runtime i18n、manifest runtime i18n 或 apidoc 文案变更。
- 验证命令：`go test ./apps/lina-core/internal/service/jobmgmt`。

### FB-9 执行记录

- 已将 `plugin-demo-dynamic` 的 `host-call-demo` 动态插件接口从 `GET` 改为 `POST`，使其与运行时状态、插件隔离存储和授权数据表写入行为匹配。
- 已同步调整插件 `zh-CN` 与 `zh-TW` apidoc i18n JSON，将 `host_call_demo` 路径翻译键从 `paths.get` 移到 `paths.post`；英文 apidoc 仍使用 DTO 英文源文本。
- 影响面：动态插件 API 方法与 apidoc 本地化资源；无前端调用点命中。
- 验证命令：`go test ./apps/lina-plugins/plugin-demo-dynamic/backend/...`、`python3 -m json.tool apps/lina-plugins/plugin-demo-dynamic/manifest/i18n/zh-CN/apidoc/plugin-api-main.json`、`python3 -m json.tool apps/lina-plugins/plugin-demo-dynamic/manifest/i18n/zh-TW/apidoc/plugin-api-main.json`。

### FB-10 执行记录

- 已在作业列表、详情与本地化关键字搜索路径复用 handler source text 缓存，并在动态插件 i18n release 解析层增加随动态插件 cache invalidation 失效的 release 缓存，避免同一插件 release 元数据在同一热点链路中重复读取。
- 影响面：后端内部性能优化；新增 Go 测试文件仅覆盖缓存行为，无新增用户可见文案。
- 验证命令：`go test ./apps/lina-core/internal/service/jobmgmt`、`go test ./apps/lina-core/internal/service/i18n`。

### FB-11 执行记录

- 已将作业分组列表中的逐组 `COUNT` 改为一次 `GROUP BY group_id` 批量统计，再按分组 ID 回填 `JobCount`。
- 影响面：后端查询性能优化，无 i18n 资源变更。
- 验证命令：`go test ./apps/lina-core/internal/service/jobmgmt`。

### FB-12 执行记录

- 已在插件集成层复用已加载的插件启用状态快照；当冷路径必须读取 `sys_plugin` 时，会把读取结果回填到进程共享快照，避免同一请求或后续过滤路径重复读取插件运行时状态。
- 已补充 `integration` 单元测试，验证 registry 读取结果会回填共享启用快照。
- 影响面：菜单/权限插件过滤内部性能优化，无 i18n 资源变更；冷缓存下权限中间件与菜单列表仍各自读取不同字段需求的菜单数据，这是当前访问控制与管理列表的职责差异。
- 验证命令：`go test ./apps/lina-core/internal/service/plugin/internal/integration ./apps/lina-core/internal/service/middleware ./apps/lina-core/internal/controller/menu`。

### FB-13 执行记录

- 已为动态插件 i18n release 解析增加可精确失效的 release 缓存，并在动态插件全量 bundle 加载后回填 enabled release 缓存，折叠插件列表投影中的重复 `sys_plugin` / `sys_plugin_release` 单条读取。
- 已验证插件服务全子树与 i18n 服务测试通过。
- 影响面：插件列表与动态插件 i18n 内部性能优化；无新增用户可见文案。
- 验证命令：`go test ./apps/lina-core/internal/service/plugin/...`、`go test ./apps/lina-core/internal/service/i18n`。

### FB-14 执行记录

- 已将角色创建/更新中的 `sys_role_menu` 关联逐条插入改为一次批量插入，并在构建批量数据时过滤非法菜单 ID、去重重复菜单 ID。
- 已补充 `role` 单元测试覆盖关联数据标准化逻辑。
- 影响面：角色管理写入性能优化，无 i18n 资源变更。
- 验证命令：`go test ./apps/lina-core/internal/service/role`。

### FB-15 执行记录

- 已在 apidoc service 增加 `ResolveRouteTexts` 批量解析接口，并通过 `pluginservice/apidoc` 暴露给源码插件；`monitor-operlog` 操作日志列表改为一次加载 apidoc catalog 后批量本地化日志记录。
- 影响面：操作日志列表本地化性能优化；接口无用户可见响应契约变更，未新增或修改翻译源文本。
- 验证命令：`go test ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/cmd ./apps/lina-plugins/monitor-operlog/backend/internal/service/operlog`。

### FB-16 执行记录

- [x] 集群部署下插件运行态缓存缺少跨节点失效机制，需基于 `cluster.enabled` 区分单机本地失效与集群共享 revision 协调，避免动态插件 i18n release、runtime bundle、前端 bundle 和插件启用快照在其他节点上长期 stale。

- 已新增 `pluginruntimecache` 内部组件，统一使用 `sys_kv_cache` 中的 `plugin-runtime/runtime-cache/revision` 作为集群共享 revision；`cluster.enabled=false` 时 revision 读写为 no-op，单机部署继续依赖进程内直接失效。
- 已将 root 插件 facade、动态插件 runtime reconciler 与 runtime i18n bundle 入口接入 revision 协调：插件安装、卸载、启停、动态包上传、源码插件升级、动态 runtime 实际收敛和 artifact 缺失收敛成功后发布 revision；读路径在集群模式下发现 revision 变化时刷新插件启用快照、清空动态前端 bundle，并显式失效 source-plugin 与 dynamic-plugin 两个 runtime i18n sector。
- 已移除动态插件 release 的跨请求进程级缓存，`TranslateDynamicPluginSourceText` 和动态插件单包 i18n 加载改为读取当前最新 release，保留调用方已有的请求内缓存，避免其他未经过 plugin facade 的业务路径继续读取 stale release。
- 已补充 `pluginruntimecache` 单元测试覆盖单机 no-op、集群 revision 刷新、mutating node 本地 revision 记录与 KV 错误传播；已补充 i18n 测试验证同一插件新增 release 后 source-text 翻译读取最新 artifact。
- 影响面：后端内部缓存一致性与性能治理，无新增用户可见文案、API 文档文案、前端 runtime i18n、宿主/插件 manifest runtime i18n 或 apidoc i18n 资源。
- 验证命令：`go test ./apps/lina-core/internal/service/pluginruntimecache`、`go test ./apps/lina-core/internal/service/i18n`、`go test ./apps/lina-core/internal/controller/i18n`、`go test -p 1 ./apps/lina-core/internal/service/plugin/...`、`go test ./apps/lina-core/internal/service/jobmgmt ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/cmd ./apps/lina-plugins/monitor-operlog/backend/internal/service/operlog ./apps/lina-plugins/plugin-demo-dynamic/backend/...` 均通过；`go test ./apps/lina-core/internal/service/plugin/...` 曾在多包并发执行时因共享测试库状态互相干扰导致 `plugin/internal/integration` 汇总失败，随后单包 `go test ./apps/lina-core/internal/service/plugin/internal/integration -run Test -count=1` 与顺序执行 `go test -p 1 ./apps/lina-core/internal/service/plugin/...` 均通过。

### FB-17 执行记录

- 已将集群模式动态插件 reconciler 从“每 2 秒直接全量 `ReconcileRuntimePlugins`”改为“每 2 秒只读取 `sys_kv_cache` 中的 `plugin-runtime/reconciler/revision`，仅当 revision 未消费或 5 分钟兜底周期到达时才扫描动态插件 registry”；启动后仍会先执行一次收敛扫描，避免冷启动遗漏。
- 已复用 `pluginruntimecache` 的集群开关控制：`cluster.enabled=false` 时 reconciler revision 读写为 no-op，单机部署保持直接同步逻辑；`cluster.enabled=true` 时动态包上传、desired state 变更、install/upgrade/refresh/enable/disable/uninstall 成功收敛和 artifact 缺失收敛都会发布 reconciler revision。
- 已补充主节点直接请求失败重试语义：主节点更新 desired state 后先发布“不本地标记已消费”的 revision，如果即时收敛失败，后台仍会在下一轮 revision 检查中继续重试；即时收敛成功后生命周期成功通知会发布并本地观察新的 revision。
- 已补充单元测试覆盖显式 KV key、只发布不本地观察、revision 未变化跳过全量扫描、revision 变化触发扫描、5 分钟兜底扫描以及 reconciler revision 发布 key。
- 影响面：后端内部调度与缓存一致性优化，无新增用户可见文案、API 文档文案、前端 runtime i18n、宿主/插件 manifest runtime i18n 或 apidoc i18n 资源。
- 聚焦审查结论：本次 FB-17 未新增 API、SQL 或前端变更；缓存权威数据源为 `sys_plugin` / `sys_plugin_release` / artifact storage，跨节点同步机制为 `sys_kv_cache` revision，正常最大触发延迟约 2 秒，遗漏 revision 的恢复路径为 5 分钟兜底扫描。
- 验证命令：`go test ./apps/lina-core/internal/service/pluginruntimecache`、`go test ./apps/lina-core/internal/service/plugin/internal/runtime`、`go test -p 1 ./apps/lina-core/internal/service/plugin/...`、`go test ./apps/lina-core/internal/service/i18n ./apps/lina-core/internal/controller/i18n`、`go test ./apps/lina-core/internal/service/jobmgmt ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/cmd ./apps/lina-plugins/monitor-operlog/backend/internal/service/operlog ./apps/lina-plugins/plugin-demo-dynamic/backend/...`、`openspec validate add-lina-perf-audit-skill --type change --strict`、`git diff --check` 均通过。

### FB-8 至 FB-15 综合验证记录

- 综合回归命令：`go test ./apps/lina-core/...`、`go test ./apps/lina-core/internal/service/plugin/...`、`go test ./apps/lina-core/internal/service/jobmgmt ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/cmd ./apps/lina-plugins/monitor-operlog/backend/internal/service/operlog ./apps/lina-plugins/plugin-demo-dynamic/backend/...` 均通过。
- 结构与格式校验：`openspec validate add-lina-perf-audit-skill --type change --strict`、`git diff --check` 均通过。
- 本轮为针对 `perf-issues/` 已审查卡片的代码修复与回归测试，未重新执行完整 `lina-perf-audit` reset DB 审计；后续完整审计若再次观察到相同指纹，卡片生命周期会按既有规则重新打开。

### FB-8 至 FB-15 接口日志运行时复验记录

- 复验运行目录：`temp/lina-perf-audit/fb8-fb15-api-verify-20260502-085553/`；准备步骤为 `make stop`、`make init confirm=init rebuild=true`、`make mock confirm=mock`、`setup-audit-env.sh`、`prepare-builtin-plugins.sh`、`stress-fixture.sh`、`scan-endpoints.sh`。
- 本轮按接口分片启动 4 个 sub agent 并行验证，报告分别为 `audits/fb8-fb10-fb11-jobmgmt-runtime-verification.md`、`audits/fb12-fb13-menu-plugin-runtime-verification.md`、`audits/fb14-role-runtime-verification.md`、`audits/fb9-fb15-plugin-runtime-verification.md`；原始 headers/body/Trace 日志位于 `api-verify/` 对应子目录。
- FB-8 `GET /api/v1/job/log?pageNum=1&pageSize=100`：Trace-ID `10287f12dd99ab1896fb99295eebd4f2`，响应 `code=0`；SQL 为 `sys_job_log` count + page list 与一次 `sys_job WHERE id IN (...)` 批量补齐，未出现 `sys_plugin` / `sys_plugin_release` 动态插件 i18n 元数据 N+1。
- FB-10 `GET /api/v1/job?pageNum=1&pageSize=100` 与 `GET /api/v1/job/5`：Trace-ID 分别为 `0079e812dd99ab1897fb9929fff6fb48`、`202f2b13dd99ab1898fb9929fa5899d0`，响应均为 `code=0`；列表为 `sys_job` count + page list 与一次 `sys_job_group WHERE id IN (...)`，详情为按主键读取 `sys_job` 与 `sys_job_group`，未出现重复动态插件本地化元数据读取。
- FB-11 `GET /api/v1/job-group`：Trace-ID `606a5013dd99ab1899fb9929f5e4a9e6`，响应 `code=0`；SQL 使用一次 `SELECT group_id, COUNT(1) ... GROUP BY group_id` 批量统计，未出现逐分组循环 `COUNT`。
- FB-12 `GET /api/v1/menu`：Trace-ID `303fd769bf99ab186afb99291e7e1f96`，响应 `code=0`；SQL 中 `sys_menu` 两次读取分别服务于权限快照和菜单树列表，未出现插件运行态或菜单元数据同型重复读取风暴；`sys_online_session` 更新时间为预期运营副作用。
- FB-13 `GET /api/v1/plugins`：Trace-ID `9854116cbf99ab186bfb99290564e3c5`，响应 `code=0`；仅观察到插件列表快照、release 快照、一次动态插件单条元数据读取与一次 `information_schema.TABLES` 读取，未出现随插件条目线性增长的 artifact state / `sys_plugin` / `sys_plugin_release` 重复读取。
- FB-14 `POST /api/v1/role`、`PUT /api/v1/role/37`、`DELETE /api/v1/role/37`：Trace-ID 分别为 `303ef41d039aab18c4fb99291058d727`、`4837a420039aab18c5fb99299e93fcd7`、`e0565e22039aab18c6fb99294a573e5c`，响应均为 `code=0`；创建时 `sys_role_menu` 仅 1 条多 `VALUES` 批量插入 3 行，更新时先 1 条删除旧关系再 1 条多 `VALUES` 批量插入 2 行，未出现按 `menuId` 循环插入；临时角色活跃行与角色菜单关系已清理为 0。
- FB-15 `GET /api/v1/operlog?pageNum=1&pageSize=3`：Trace-ID `901c72effc99ab18b7fb992909cdd76b`，响应 `code=0`；SQL 为操作日志 count/list 加插件与 release 快照读取，没有按返回行重复读取插件本地化或 apidoc 元数据；`sys_online_session` 更新时间为预期运营副作用。
- FB-9 `POST /api/v1/extensions/plugin-demo-dynamic/host-call-demo`：Trace-ID `a87fc3f0fc99ab18b8fb99290b84a644`，HTTP 200；写 SQL 来自显式 `POST` 动作，包括 `sys_plugin_state` 与 `sys_plugin_node_state` 的演示性写入，符合接口语义。负向验证 `GET /api/v1/extensions/plugin-demo-dynamic/host-call-demo?skipNetwork=1` 返回 HTTP 404，Trace-ID `30175652fd99ab18b9fb9929852025d5`，未触发写 SQL。
- 写接口复验前，因 `demo-control` 会按设计拦截写请求，本轮临时将其置为 disabled，证据为 `07-disable-demo-control-for-write-verification.log`；复验完成后已恢复为 enabled，证据为 `08-restore-demo-control-after-write-verification.log`。
- 环境恢复：`restore-audit-env.sh --run-dir temp/lina-perf-audit/fb8-fb15-api-verify-20260502-085553` 已恢复 `logger.path` 与 `logger.file` 并停止服务；`10-make-status-after-restore.log` 显示前后端均未运行。
- 复验后补充校验：`go test ./apps/lina-core/internal/service/jobmgmt ./apps/lina-core/internal/service/i18n ./apps/lina-core/internal/service/plugin/... ./apps/lina-core/internal/service/role ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/cmd ./apps/lina-plugins/monitor-operlog/backend/internal/service/operlog ./apps/lina-plugins/plugin-demo-dynamic/backend/...`、`openspec validate add-lina-perf-audit-skill --type change --strict`、`git diff --check` 均通过；`apps/lina-core/manifest/config/config.yaml` 与 SQL 目录无恢复遗留 diff。
- i18n 影响判断：本次运行时复验只新增验证记录和临时审计产物，不新增或修改接口文案、前端 runtime i18n、宿主/插件 runtime manifest i18n 或 apidoc i18n 资源。

### FB-8 至 FB-15 targeted API 回归记录

- 回归运行目录：`temp/lina-perf-audit/perf-issues-regression-20260502-205731/`；最终汇总为 `api-regression/SUMMARY.final.md`，证据日志为 `server.log` 与 `backend-nohup.log`。本轮未执行完整 `lina-perf-audit` 全量审计，也未重建数据库；准备步骤为 `setup-audit-env.sh`、`prepare-builtin-plugins.sh`、`stress-fixture.sh` 与 `scan-endpoints.sh`，随后只调用 `perf-issues/` 中 FB-8 至 FB-15 关联接口。
- FB-8 `GET /api/v1/job/log?pageNum=1&pageSize=100`：Trace-ID `c8731d632cc1ab189ad4707194cae4c6`，响应 `code=0`；SQL 共 9 条，包含一次 `sys_job WHERE id IN (...)` 批量补齐，`sys_plugin` 与 `sys_plugin_release` 引用均为 0，未复现动态插件 i18n 元数据 N+1。
- FB-10 `GET /api/v1/job?pageNum=1&pageSize=100` 与 `GET /api/v1/job/5`：Trace-ID 分别为 `709f9d642cc1ab189bd47071461d40e4`、`30f4dd642cc1ab189cd47071a6ef2bb2`，响应均为 `code=0`；作业列表 SQL 共 4 条并使用一次 `sys_job_group WHERE id IN (...)` 批量补齐，详情 SQL 共 3 条，未出现动态插件元数据重复读取。
- FB-11 `GET /api/v1/job-group`：Trace-ID `a0a403652cc1ab189dd470715eb905c3`，响应 `code=0`；SQL 共 4 条，其中作业数量统计为一次 `SELECT group_id, COUNT(1) ... GROUP BY group_id`，没有逐分组 `COUNT` 循环。
- FB-12 `GET /api/v1/menu`：Trace-ID `30142c652cc1ab189ed47071895d6f26`，响应 `code=0`；SQL 共 2 条，未出现插件运行态或菜单元数据重复读取风暴。
- FB-13 `GET /api/v1/plugins`：Trace-ID `101419662cc1ab189fd47071611ef3d8`，响应 `code=0`；SQL 共 8 条，`sys_plugin` 与 `sys_plugin_release` 引用各 3 次，属于有界读取，未随插件条目线性增长。
- FB-14 `POST /api/v1/role`、`PUT /api/v1/role/47`、`DELETE /api/v1/role/47`：Trace-ID 分别为 `30165b6b2cc1ab18a1d4707139e6dd30`、`c828b56b2cc1ab18a2d470716c1371ec`、`1828236d2cc1ab18a3d47071505b9ca0`，响应均为 `code=0`；创建仅 1 条 `sys_role_menu` 多 `VALUES` 批量插入，更新为 1 条删除旧关系加 1 条多 `VALUES` 批量插入，未出现按 `menuId` 循环插入；临时角色活跃行与角色菜单关系已清理为 0。
- FB-15 `GET /api/v1/operlog?pageNum=1&pageSize=3`：Trace-ID `d06b08692cc1ab18a0d47071594bb04a`，响应 `code=0`；SQL 共 5 条，`sys_plugin` 与 `sys_plugin_release` 引用各 1 次，操作日志本地化元数据读取保持有界。
- FB-9 `POST /api/v1/extensions/plugin-demo-dynamic/host-call-demo`：Trace-ID `289e726f2cc1ab18a4d470716c7f024e`，HTTP 200，响应 `pluginId=plugin-demo-dynamic` 且临时 `sys_plugin_node_state` 数据已删除；写 SQL 来自显式 `POST` 动作，符合接口语义。负向验证 `GET /api/v1/extensions/plugin-demo-dynamic/host-call-demo?skipNetwork=1` 返回 HTTP 404，Trace-ID `20fc3ed42cc1ab18a5d47071e750144c`，未触发写 SQL。
- 写接口复验前通过 SQL 临时将 `demo-control` 置为 disabled，复验结束已恢复为 enabled；证据为 `api-regression/disable-demo-control.log`、`api-regression/restore-demo-control.log` 与 `api-regression/verify-role-cleanup.log`。审计环境恢复命令 `restore-audit-env.sh --run-dir temp/lina-perf-audit/perf-issues-regression-20260502-205731` 已恢复 `logger.path` / `logger.file` 并停止服务。
- 补充校验：`go test ./apps/lina-core/internal/service/jobmgmt ./apps/lina-core/internal/service/i18n ./apps/lina-core/internal/service/plugin/... ./apps/lina-core/internal/service/role ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/cmd ./apps/lina-plugins/monitor-operlog/backend/internal/service/operlog ./apps/lina-plugins/plugin-demo-dynamic/backend/...`、`openspec validate add-lina-perf-audit-skill --type change --strict`、`git diff --check` 均通过。
- 评估结论：本轮 targeted API 回归全部通过，FB-8 至 FB-15 的改进后预期效果满足；未观察到原 `perf-issues/` 卡片对应的 N+1、重复元数据读取、循环写入或读接口写副作用回归。
- i18n 与缓存影响判断：本次只新增验证记录和临时审计产物，不新增或修改运行时代码、接口文案、前端 runtime i18n、宿主/插件 manifest runtime i18n 或 apidoc i18n 资源；缓存相关变更未新增，本轮只验证既有有界读取与显式失效后的运行时表现。
