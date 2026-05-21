# Tasks

## 1. 基线盘点与治理骨架

- [x] 盘点 `hack/tests/e2e/` 现状，整理历史目录到稳定能力目录的迁移映射，记录高频固定等待、共享状态热点、错误归属文件和目录治理缺口。
- [x] 建立 `hack/tests/config/`、`support/`、`fixtures/`、`scripts/`、`debug/` 等治理骨架，并让 execution manifest 成为 smoke、module、serial 边界、隔离类别和 allowlist 的唯一事实来源。
- [x] 实现持续治理校验，覆盖目录归属、非测试文件、manifest 引用、根路径插件耦合、隔离类别完整性和编号规则。

## 2. 目录重组与编号收敛

- [x] 按稳定能力边界拆分历史 `system/`、`monitor/`、`plugin/` 等过载目录，建立 `iam`、`settings`、`scheduler`、`extension`、`org`、`content`、`monitor`、`dashboard`、`about` 等能力目录和必要的二级子域目录。
- [x] 将 helper、debug script、共享等待逻辑和治理脚本迁出 `hack/tests/e2e/`，确保发现范围只包含真实 `TC*.ts` 用例。
- [x] 将具体官方源码插件行为、页面对象、测试数据和 baseline 收敛到 `apps/lina-plugins/<plugin-id>/hack/tests/`，让根路径 E2E 只保留宿主插件框架、动态测试插件和通用插件治理资产。
- [x] 将 E2E 文件命名从全局 `TC{NNNN}` 迁移为模块目录本地 `TC{NNN}` 连续递增，并同步修复测试标题、导入、文档和 manifest 引用。

## 3. 执行入口、环境职责与 CI 分片

- [x] 保持 `pnpm test` 为全量回归入口，并补齐 `test:full`、`test:smoke`、`test:module`、`test:host` 和 `test:host:module` 等分层命令。
- [x] 实现 manifest 驱动的并行池 / 串行池调度与隔离类别摘要输出，确保 full、module 和 host-only 模式共享同一套边界规则。
- [x] 将 plugin-full 入口统一为 `extension:plugin`、`plugins` 与 `plugin:<plugin-id>`，移除官方插件业务别名 scope，并在 host-only module 模式下显式拒绝插件 scope。
- [x] 将 plugin-full CI 收敛为基于通用 scope 的分片执行，为各分片生成唯一 artifact 名称，并保证任一分片失败会阻断完整 verification suite 的下游流程。
- [x] 修正 browser E2E workflow 的 PostgreSQL health check，显式使用 `pg_isready -U postgres -d linapro`。

## 4. 认证态、等待策略与宿主慢用例优化

- [x] 增加预生成管理员登录态与 `storageState` 管理，并让高频已登录 fixture 默认复用登录态，同时保留认证场景的真实登录路径。
- [x] 新增不自动导航 dashboard 的 `authenticatedPage`，优先迁移菜单 CRUD、文件管理、角色 CRUD、参数导入、字典导入等高耗时宿主页面用例，减少重复首页加载和二次跳转。
- [x] 为表格、抽屉、弹窗、toast、route ready、dropdown/confirm overlay 提供共享的状态驱动等待工具，并在高频 page object 中系统替换固定等待。
- [x] 完成第二轮固定等待清理，使业务测试和调试脚本中的 `waitForTimeout` 收敛到零残留或仅保留有明确业务理由的例外。

## 5. 共享状态隔离、基线装配与断言治理

- [x] 盘点插件生命周期、运行时 i18n 缓存、系统参数、公共配置、字典、菜单/角色权限、共享数据库种子和文件系统产物等高风险共享状态，并为对应文件声明机器可读的 isolation category。
- [x] 为 validator 增加高风险启发式检测、allowlist 原因说明和未分类串行风险拦截，确保新增高风险测试不能绕过治理。
- [x] 为普通插件功能用例提供 suite/shard 级幂等 baseline，统一完成插件同步、安装、启用、mock 数据加载和投影刷新；生命周期用例继续显式控制自身状态与清理逻辑。
- [x] 将缓存/ETag 断言调整为协议语义校验，接受合法的 `304` 或 `200 + 新 ETag + body` 分支；将业务状态断言调整为稳定字段、ID、code、permission key、labelKey 或计数器。
- [x] 统一测试文件的自包含前置条件、唯一数据命名和自清理策略，消除跨文件依赖。

## 6. 插件生命周期覆盖与 plugin-full 提速

- [x] 收敛 plugin-full 的职责边界：host-only 继续覆盖宿主全量能力，根 `extension:plugin` 只覆盖宿主插件框架与动态测试插件接缝，具体官方源码插件行为全部由插件目录内自有 E2E 闭环维护。
- [x] 为普通插件页面测试迁移重复的 `ensureSourcePluginEnabled` 逻辑，改为复用幂等 baseline，减少每个用例重复同步、安装、启用和投影刷新成本。
- [x] 将官方源码插件生命周期回归改为“代表性完整 UI 生命周期 + 其余插件 contract smoke / 页面可访问性”组合，降低重复生命周期成本。
- [x] 拆分 dynamic runtime 生命周期回归，区分宿主壳/iframe/runtime 切换所需的关键 UI 覆盖与可由 API/request 层验证的 contract 场景。
- [x] 将 `plugins` scope 进一步拆成 5 个 Playwright shard，显著降低 plugin-full 最长分片耗时并保持通用入口语义不变。

## 7. 反馈驱动的宿主与插件稳定性修复

- [x] 修复 generic plugin resource 数据权限映射与宿主 `data_scope` 枚举不一致的问题，并校正插件管理动作权限用例的前置角色选择。
- [x] 区分 host-only 与 plugin-full 环境语义，为 host-only 模块选择器、宿主断言、调度任务断言和插件依赖过滤补齐显式分支。
- [x] 为动态插件运行时补齐 `/x` 代理、示例数据清理、菜单投影刷新等待和英文页面巡检前置条件，消除跨用例状态泄漏。
- [x] 修复角色新增/编辑抽屉的异步初始化竞态、动态插件禁用后菜单隐藏竞态、源码插件生命周期表格列顺序断言、强制卸载前置条件文案断言等 shard 暴露的问题。
- [x] 将动态插件超限上传 UI 探测改为文件路径输入，并保留 API multipart 错误分支验证；同时在运行时内部收敛 cache/reconciler change reason 常量，避免字符串硬编码扩散。
- [x] 加固插件热升级后的访问刷新逻辑，使当前稳定 iframe 页能在 `upgrade_failed` 等非正常运行态下按等价路由/资源目标正确保留或切换。
- [x] 将默认上传上限统一提升到 100MB，并同步对齐配置模板、初始化 SQL、运行时 fallback、packed asset 与相关测试期望。

## 8. 文档与规则同步

- [x] 补齐 `hack/tests/README.md` 与 `hack/tests/README.zh-CN.md`，说明目录边界、执行入口、manifest 机制、隔离类别、fixture 前置条件、host-only / plugin-full 职责和缓存语义断言规则。
- [x] 补充 E2E 冲突治理记录、宿主根路径插件去耦规则、模块本地编号规则，并同步更新相关项目规范与审查清单。
- [x] 将根路径插件耦合扫描改为从 `apps/lina-plugins/*/plugin.yaml` 动态发现插件标识，避免在治理脚本中维护静态插件 denylist。

## 9. 验证与归档前复核

- [x] 通过 `pnpm -C hack/tests test:validate`、`pnpm -C hack/tests exec tsc --noEmit`、`openspec validate ... --strict` 和 `git diff --check`，最终治理覆盖 235 个 E2E 文件、17 个 scope、6 个 smoke 文件和 218 个 serial 文件。
- [x] 通过 smoke、module、host-only module、`extension:plugin`、`plugins` 及各 shard 的针对性回归，确认新入口、host-only 过滤、plugin baseline、代表性生命周期和反馈修复都可独立验证。
- [x] 通过关键支撑测试，包括前端 route refresh 单测、前端单测全集、相关 Go 精确门禁与 `internal/cmd` 启动绑定测试，确保与回归稳定性直接相关的宿主/插件修正可编译且可执行。
- [x] 记录并复核优化收益：host-only 基线从约 36 分钟收敛到 `244 passed, 1 skipped (14.6m)`；plugin-full 基线从约 2 小时 / 112 分钟 Playwright 收敛到 `516 passed, 8 skipped (42.8m)`，且 `plugins` 分片最长耗时显著下降。
- [x] 逐轮完成 `lina-review`，复核目录结构、环境职责、fixture、分片、baseline、反馈修复和验证证据，确认本组变更达到归档条件。
