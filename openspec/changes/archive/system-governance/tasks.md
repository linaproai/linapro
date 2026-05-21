## 1. OpenSpec 归档自动化基础架构

- [x] 1.1 调整 `.gitignore`，允许提交 `.github/codex/config.template.toml` 和无密钥模板，同时继续忽略真实认证文件
- [x] 1.2 新增或整理 `.github/codex` 配置模板，确保 workflow 可复制配置但不会提交真实 `OPENAI_API_KEY`
- [x] 1.3 新增 `Monthly OpenSpec Archive` GitHub Actions workflow，支持 schedule 和 `workflow_dispatch`
- [x] 1.4 在 workflow 中通过临时 `CODEX_HOME` 注入 Codex 配置和 `OPENAI_API_KEY` secret
- [x] 1.5 实现 OpenSpec 完成状态预检查，无可归档变更时跳过 AI 工具执行
- [x] 1.6 实现 OpenSpec 校验、变更范围保护和自动写回逻辑

## 2. AI 工具路由与 reusable workflow 隔离

- [x] 2.1 主 workflow 通过 GitHub Variables 中的 `AI_CODING_TOOL` 在 `codex` 和 `cc` 之间切换 AI Coding 工具
- [x] 2.2 主 workflow 通过工具专属 reusable workflow 封装 Codex 和 Claude Code 实现细节，主 workflow 只负责检测和路由
- [x] 2.3 新增 `.github/workflows/monthly-openspec-archive-codex.yml` 和 `.github/workflows/monthly-openspec-archive-cc.yml`
- [x] 2.4 调整 `ANTHROPIC_BASE_URL`、`ANTHROPIC_CUSTOM_MODEL` 和 `OPENAI_BASE_URL` 为 GitHub Variables 读取

## 3. 公共组件抽取与提示词复用

- [x] 3.1 新增 `.github/actions/monthly-openspec-setup`、`.github/actions/monthly-openspec-detect-changes`、`.github/actions/monthly-openspec-finalize-pr` 本地 composite action
- [x] 3.2 提取自动归档和归档聚合提示词到 `.github/prompts/` 公共文件，由 Codex 和 Claude Code workflow 共同引用
- [x] 3.3 Codex 配置模板使用 `.github/codex/config.template.toml` 文件名，并更新 workflow、忽略规则和 OpenSpec 引用
- [x] 3.4 产生文件修改时创建或更新维护 PR，而不是直接推送到默认分支

## 4. 月度调度与 Copilot 支持

- [x] 4.1 工作流改为每月 1 日 00:00 Asia/Shanghai 触发一次，保留手动触发，并同步调整文件名与工作流名称
- [x] 4.2 支持通过 GitHub Copilot CLI 执行自动归档和归档聚合，并通过 GitHub Variables 配置 Copilot 模型
- [x] 4.3 支持通过 GitHub Variables 配置推理等级并传递给 `--reasoning-effort`
- [x] 4.4 确保 AI 工具进程输出实时显示到 GitHub Actions step 日志

## 5. 阶段化失败与校验强化

- [x] 5.1 任一阶段失败时立即失败退出，不得继续执行后续阶段
- [x] 5.2 新增 `.github/actions/monthly-openspec-assert-archive-complete` 和 `.github/actions/monthly-openspec-validate` composite action
- [x] 5.3 `openspec validate --all` 必须通过，避免归档工作流因既有主规范格式失败
- [x] 5.4 将 `openspec/specs/**/spec.md` 主规范统一为 OpenSpec 1.3.1 可识别结构

## 6. 确定性基础归档

- [x] 6.1 新增共享 monthly OpenSpec 确定性归档 composite action `.github/actions/monthly-openspec-auto-archive`
- [x] 6.2 将 Codex、Claude Code 和 GitHub Copilot reusable workflow 接入确定性归档，AI 工具仅用于聚合
- [x] 6.3 确定性归档有部分成功时先写入归档 PR，然后在所有候选处理完成后失败退出
- [x] 6.4 升级 artifact upload workflow actions 到最新运行时版本

## 7. 归档阻塞修复与分支策略

- [x] 7.1 修复 `remove-sqlite-support` 使其可被 `openspec archive -y` 正常归档
- [x] 7.2 `workflow_dispatch` 允许从任意分支触发，手动触发以触发分支作为检测和 PR 目标分支
- [x] 7.3 归档 PR 来源分支包含触发分支的安全化标识
- [x] 7.4 AI 归档聚合失败或产生无效 OpenSpec 时恢复确定性归档快照，不阻塞已通过校验的归档 PR
- [x] 7.5 仓库策略阻止 PR 创建时输出手动 PR 链接并成功结束

## 8. 归档自动化验证与审查

- [x] 8.1 运行 `openspec validate` 对所有相关变更执行 strict 校验
- [x] 8.2 对新增 workflow、action 和配置模板执行静态检查，确认 YAML、JSON 和 TOML 格式有效
- [x] 8.3 运行 actionlint 验证 workflow 和 composite action
- [x] 8.4 运行临时副本确定性归档 smoke 测试
- [x] 8.5 记录 i18n、缓存一致性、数据权限、REST API、E2E 和 Go 生产代码影响判断
- [x] 8.6 完成实现后调用 `lina-review` 审查

## 9. 规范与审查标准落地

- [x] 9.1 更新 `AGENTS.md` 后端代码规范，写入主文件契约入口、接口方法详细注释、文件顶部详细说明和 `lina-core/pkg` 公共组件同等治理要求
- [x] 9.2 更新 `.agents/skills/lina-review/SKILL.md`，增加主文件职责、接口方法注释完整度、文件顶部注释质量和分批验证记录审查项
- [x] 9.3 校验 OpenSpec 增量规范与设计文档，确认 `backend-conformance` 变更覆盖宿主、源码插件和 `lina-core/pkg`
- [x] 9.4 记录本轮规范变更的 i18n、缓存一致性、数据权限和开发工具脚本影响判断

## 10. 基线扫描与任务切片确认

- [x] 10.1 扫描宿主 `apps/lina-core/internal/service/**` 主文件，列出仍包含复杂 receiver 方法实现的组件
- [x] 10.2 扫描 `apps/lina-core/pkg/**` 主文件，列出仍包含复杂实现逻辑的公共组件
- [x] 10.3 扫描源码插件 `apps/lina-plugins/*/backend/internal/service/**` 主文件，列出仍包含复杂 receiver 方法实现的插件组件
- [x] 10.4 将扫描结果按模块批次记录到任务备注，作为后续整改范围和终审对账依据

## 11. 宿主基础与安全服务整改

- [x] 11.1 整改 `auth` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 11.2 整改 `session` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 11.3 整改 `middleware` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 11.4 整改 `bizctx` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 11.5 运行基础与安全服务变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 11.6 调用 `lina-review` 审查本批基础与安全服务整改

## 12. 宿主用户、角色与数据权限服务整改

- [x] 12.1 整改 `user` 主文件职责、接口方法注释和文件顶部说明，按列表、详情、资料、导入导出、批量操作等职责迁移实现
- [x] 12.2 整改 `role` 主文件职责、接口方法注释和文件顶部说明，按角色、菜单、数据权限、访问缓存等职责迁移实现
- [x] 12.3 整改 `datascope` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 12.4 整改 `tenantcap` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 12.5 整改 `orgcap` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 12.6 运行用户、角色、数据权限相关变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 12.7 调用 `lina-review` 审查本批用户、角色与数据权限服务整改

## 13. 宿主系统治理服务整改

- [x] 13.1 整改 `config` 主文件职责、接口方法注释和文件顶部说明，保持各配置分组实现文件边界清晰
- [x] 13.2 整改 `sysconfig` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 13.3 整改 `sysinfo` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 13.4 整改 `dict` 主文件职责、接口方法注释和文件顶部说明，按类型、数据、导入导出和 i18n 职责迁移实现
- [x] 13.5 整改 `menu` 主文件职责、接口方法注释和文件顶部说明，按权限树、过滤、校验和 i18n 职责迁移实现
- [x] 13.6 整改 `file` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 13.7 运行系统治理服务变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 13.8 调用 `lina-review` 审查本批系统治理服务整改

## 14. 宿主任务、调度与运行状态服务整改

- [x] 14.1 整改 `cron` 主文件职责、接口方法注释和文件顶部说明，确保调度注册与具体任务逻辑分文件
- [x] 14.2 整改 `jobhandler` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 14.3 整改 `jobmeta` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 14.4 整改 `jobmgmt` 及其 internal 子组件主文件职责、接口方法注释和文件顶部说明
- [x] 14.5 整改 `startupstats` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 14.6 运行任务与调度相关变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 14.7 调用 `lina-review` 审查本批任务、调度与运行状态服务整改

## 15. 宿主缓存、协调与分布式基础服务整改

- [x] 15.1 整改 `cluster` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 15.2 整改 `coordination` 及其 internal 子组件主文件职责、接口方法注释和文件顶部说明
- [x] 15.3 整改 `cachecoord` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 15.4 整改 `kvcache` 及其 internal 子组件主文件职责、接口方法注释和文件顶部说明
- [x] 15.5 整改 `locker` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 15.6 整改 `hostlock` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 15.7 运行缓存、协调与锁相关变更包 Go 编译门禁，并记录缓存一致性影响判断
- [x] 15.8 调用 `lina-review` 审查本批缓存、协调与分布式基础服务整改

## 16. 宿主 i18n、通知与 API 文档服务整改

- [x] 16.1 整改 `i18n` 主文件职责、接口方法注释和文件顶部说明，按 locale、cache、resource、source text、dynamic plugin 等职责保持实现分文件
- [x] 16.2 整改 `notify` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 16.3 整改 `usermsg` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 16.4 整改 `apidoc` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 16.5 运行 i18n、通知和 API 文档相关变更包 Go 编译门禁，并记录 i18n 与缓存一致性影响判断
- [x] 16.6 调用 `lina-review` 审查本批 i18n、通知与 API 文档服务整改

## 17. 宿主插件外层与内部子组件整改

- [x] 17.1 整改 `plugin` 主文件职责、接口方法注释和文件顶部说明，按列表、生命周期、运行时、升级、前端、OpenAPI、host service 等职责迁移实现
- [x] 17.2 整改 `pluginruntimecache` 和 `pluginhostservices` 主文件职责、接口方法注释和文件顶部说明
- [x] 17.3 整改 `plugin/internal/catalog`、`runtime`、`integration`、`frontend`、`openapi`、`lifecycle`、`wasm` 主文件职责、接口方法注释和文件顶部说明
- [x] 17.4 运行插件服务变更包 Go 编译门禁，并记录 i18n、缓存一致性、数据权限和插件桥接影响判断
- [x] 17.5 调用 `lina-review` 审查本批宿主插件服务整改

## 18. `lina-core/pkg` 公共组件整改

- [x] 18.1 整改小型公共组件 `authtoken`、`bizerr`、`closeutil`、`dbdriver`、`excelutil`、`gdbutil`、`logger`、`menutype`、`orgcap`、`pluginfs`、`tenantcap`、`testsupport` 主文件职责和文件顶部说明
- [x] 18.2 整改插件与桥接公共组件 `pluginhost`、`pluginbridge`、`pluginservice`、`plugindb`、`sourceupgrade` 主文件职责、接口方法注释和文件顶部说明
- [x] 18.3 整改数据库、方言与资源公共组件 `dialect`、`i18nresource` 主文件职责、接口方法注释和文件顶部说明
- [x] 18.4 运行 `lina-core/pkg` 公共组件变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 18.5 调用 `lina-review` 审查本批 `lina-core/pkg` 公共组件整改

## 19. 源码插件后端服务整改

- [x] 19.1 整改 `org-center` 插件 `dept`、`post` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 19.2 整改 `multi-tenant` 插件 `tenant`、`membership`、`tenantplugin`、`resolver`、`resolverconfig`、`impersonate`、`provider`、`lifecycleprecondition` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 19.3 整改监控插件 `monitor-loginlog`、`monitor-operlog`、`monitor-online`、`monitor-server` 和内容插件 `content-notice` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 19.4 整改示例与演示插件 `demo-control`、`plugin-demo-source`、`plugin-demo-dynamic` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 19.5 运行源码插件后端变更包 Go 编译门禁，并记录 i18n、缓存一致性、数据权限和插件桥接影响判断
- [x] 19.6 调用 `lina-review` 审查本批源码插件后端服务整改

## 20. `linactl` 命令文件组织治理

- [x] 20.1 在 `AGENTS.md` 开发工具与脚本规范中补充 `linactl` 命令文件命名规则和子组件组织规则
- [x] 20.2 在 `lina-review` 开发工具与脚本跨平台审查中补充命令文件命名和子组件组织审查项
- [x] 20.3 将现有 `hack/tools/linactl` 命令入口按规范拆分到具体命令文件，删除旧兜底文件
- [x] 20.4 将共享实现迁移到 `internal/<组件名称>/` 子组件
- [x] 20.5 运行 `hack/tools/linactl` 包测试和工具 smoke 验证
- [x] 20.6 调用 `lina-review` 审查 `linactl` 命令组织治理

## 21. 全量复核与治理验证

- [x] 21.1 全量扫描宿主、源码插件和 `lina-core/pkg` 主文件，确认复杂实现逻辑已按任务范围迁出或记录明确例外
- [x] 21.2 全量扫描新增或修改的接口定义，确认接口方法注释覆盖功能、输入、输出、错误和关键约束
- [x] 21.3 全量扫描新增或修改的 Go 文件顶部注释，确认主文件和非主文件注释层级正确
- [x] 21.4 运行 `openspec validate --strict` 对所有相关变更执行校验
- [x] 21.5 运行 `git diff --check` 覆盖本变更所有文档和 Go 文件
- [x] 21.6 汇总所有分批 Go 编译门禁结果
- [x] 21.7 记录本变更最终 i18n、缓存一致性、数据权限、开发工具脚本和 E2E 影响判断
- [x] 21.8 调用 `lina-review` 完成终审并修复审查发现
