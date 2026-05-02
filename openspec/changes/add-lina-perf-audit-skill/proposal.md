## Why

LinaPro 当前已有 `lina-core` 与多个内置插件后端 API，在 SDD 驱动的快速迭代中持续累积，但缺少系统性的性能与读请求副作用审查机制：N+1 查询、循环内调用 dao、缺索引的全表扫描、缺分页的列表查询、阻塞调用、重复读配置、未命中缓存，以及 GET/查询接口执行 UPDATE/INSERT/DELETE 等写操作的问题只能依赖个人经验偶发发现。当前 GoFrame 默认已经把每个请求的 trace ID 写到响应头 `Trace-ID`，配合 `database.debug=true` 输出的 SQL 日志，已经具备 trace ID 串联 SQL 调用链的能力，需要把这套审查流程沉淀为可复用的 AI 治理能力，避免每次手动重做。

## What Changes

- 新增 agent skill `lina-perf-audit`：编排"环境准备 → 安装并启用所有内置插件 → 按接口任务拆分子 agent → trace ID 反查 SQL → 静态源码对照 → 报告汇总"的完整审查流水线,同时检查 GET/查询接口是否执行写 SQL。
- 子 agent 颗粒度：主 agent 只负责接口任务跟踪和汇总；每个 API 接口审查任务都交给子 agent 执行，默认按模块/资源生成任务队列，接口较多时继续拆成单接口或小接口组 shard 并发执行。
- 破坏性接口（DELETE / 卸载 / 清空等）由子 agent 自治：跑前自动 create 一条目标资源，调用结束删自身，不污染其他模块数据。
- trace ID 取得方式：直接读 GoFrame 默认响应头 `Trace-ID`，**不引入新中间件**，不改动生产代码。
- 配套交付物：
  - `.agents/skills/lina-perf-audit/scripts/setup-audit-env.sh` —— 停服 → 备份 `logger.path` / `logger.file` 原值 → patch 为审计专用目录与固定 `server.log` → 启动后端 → 等待健康探针就绪 → 输出 admin/admin123 登录 token。
  - `.agents/skills/lina-perf-audit/scripts/restore-audit-env.sh` —— 读取备份的 `logger.path` / `logger.file` 原值,恢复 `apps/lina-core/manifest/config/config.yaml`,停服(成功路径与失败路径都必须能调用)。
  - `.agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh` —— 扫描 `apps/lina-plugins/*/plugin.yaml`,通过宿主插件管理 API 同步、安装并启用所有内置插件,插件存在 `manifest/sql/mock-data/` 时显式加载 mock data。
  - `.agents/skills/lina-perf-audit/scripts/scan-endpoints.sh` —— 扫描 `apps/lina-core/api` 与 `apps/lina-plugins/*/backend/api`,解析 g.Meta 生成按模块分组的 endpoint catalog。
  - `.agents/skills/lina-perf-audit/scripts/probe-fixtures.sh` —— 调用各模块的 list 接口,采集每种资源的样本 ID,生成 fixtures.json。
  - `.agents/skills/lina-perf-audit/scripts/stress-fixture.sh` —— 在宿主 mock 与所有内置插件安装/插件 mock 完成后叠加压力 fixture(每个列表资源补 50~100 条),让 N+1 在 SQL 调用次数上有可观察的差异。
- 报告结构(双层):
  - **单次 run 原始报告**:每次审计写到 `temp/lina-perf-audit/<run-id>/`,包含 `audits/<module-or-shard>.md` 与 `SUMMARY.md`,按 HIGH / MEDIUM / LOW 三级分类问题。可被丢弃或保留多份对比。
- **跨 run 问题卡片**:每个发现的性能问题或读请求副作用问题独立成一个 markdown 写到仓库根目录 `perf-issues/<severity>-<module>-<slug>.md`,内含问题描述、复现步骤、证据、改进方案与状态字段(`open` / `in-progress` / `fixed` / `obsolete`),供后续 OpenSpec 变更逐个消费;基于指纹去重,重复审计同一问题时**更新**已有卡片而不是新建。
- **强制约束:仅手动触发**:skill 描述、文档与 spec 中明确声明 manual-trigger-only,禁止被其他 skill / CI / 自动化流水线引用,也禁止在用户描述含糊时(如"接口好像有点慢")自动触发,必须先与用户确认。
- 不修复任何接口性能问题:本变更只交付审查能力本身;实际接口优化由后续单独的 OpenSpec 变更逐项处理。

## Capabilities

### New Capabilities

- `lina-perf-audit-skill`: 定义 LinaPro API 性能与读请求副作用审查 agent skill 的对外契约,涵盖触发条件(手动)、子 agent 编排约束、破坏性接口治理规则、trace ID 串联机制、报告产物结构与严重度分级、以及"禁止自动触发"的强制约束。

### Modified Capabilities

(无 —— 本次变更只新增能力,不修改任何已有 capability。)

## Impact

- **新增 skill 文件**:`.agents/skills/lina-perf-audit/SKILL.md` 及其引用的 references,作为 Claude Code、Codex 与其他 AI Coding 工具都可读取的通用 skill。
- **新增 skill 内置审查辅助脚本**:`.agents/skills/lina-perf-audit/scripts/` 目录(setup-audit-env / restore-audit-env / prepare-builtin-plugins / scan-endpoints / probe-fixtures / stress-fixture),脚本作为 skill bundled resources 闭环维护,不再放到仓库级 `hack/scripts/` 目录。
- **新增 spec**:`openspec/specs/lina-perf-audit-skill/spec.md`(归档时落入)。
- **运行期影响**:仅在被手动触发时才会重建本地数据库、加载宿主与所有内置插件 mock 数据、补充 stress fixture、重启后端服务、写入临时审计日志到 `temp/lina-perf-audit/<run-id>/` 与累积的问题卡片到根目录 `perf-issues/`。**不影响生产代码、不修改任何接口运行时行为、不引入新中间件、不改动配置默认值**。
- **i18n 影响**:无。本变更不涉及任何前端文案、菜单/按钮、API DTO 描述、apidoc 资源或语言包变更,审计报告与 skill 文档仅供研发人员阅读使用,不进入运行时多语言体系。
- **依赖**:复用现有 `make init confirm=init rebuild=true` / `make mock confirm=mock` / `make dev` / `make stop` 命令链;依赖 GoFrame 默认 `Trace-ID` 响应头与 `database.debug=true` 行为。
- **不归档时机**:skill 落地后立即可用;按本变更归档时,实际审计**报告本身**不归档(报告每次执行都会重新生成),只归档 skill 能力定义和辅助脚本。
