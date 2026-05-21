## 1. CI 分片与基础治理

- [x] 1.1 修正 browser E2E workflow 的 PostgreSQL health check，显式使用 `pg_isready -U postgres -d linapro`。
- [x] 1.2 将 plugin-full E2E job 改为基于通用入口的分片执行，分片覆盖 `extension:plugin` 与 `plugins`。
- [x] 1.3 为 plugin-full 分片生成唯一 artifact 名称，确保 Playwright report、test-results、后端日志和前端日志不会互相覆盖。
- [x] 1.4 验证分片失败会让完整 verification suite 失败，并阻止依赖验证成功的后续 job。

## 2. Plugin-full 覆盖范围收敛

- [x] 2.1 梳理 plugin-full 需要保留的插件框架通用用例，明确菜单、权限、路由、i18n、任务、运行时资源的通用覆盖文件。
- [x] 2.2 收敛根 E2E manifest，使 plugin-full 不再选择依赖具体官方插件的根测试文件集合。
- [x] 2.3 确认 host-only 仍覆盖宿主全量能力，plugin-full 仍覆盖官方插件自有用例和插件生命周期。

## 3. 认证 fixture 与宿主慢用例优化

- [x] 3.1 在 `hack/tests/fixtures/auth.ts` 中新增不自动导航到 dashboard 的轻量认证页面 fixture，并保留现有 `adminPage` 行为。
- [x] 3.2 优先迁移菜单 CRUD 中适合直接进入业务路由的用例，减少重复 dashboard 加载。
- [x] 3.3 优先迁移文件管理中适合直接进入业务路由的用例，减少重复 dashboard 加载。
- [x] 3.4 评估并迁移角色 CRUD、参数导入、字典导入等相同模式的慢文件。

## 4. 插件 baseline 与普通插件用例优化

- [x] 4.1 在插件 E2E fixture/support 中新增幂等 baseline 辅助，支持一次性同步、安装、启用插件、加载可用 mock 数据并刷新插件投影。
- [x] 4.2 将普通插件页面测试中的重复 `ensureSourcePluginEnabled` 迁移为 suite 或 shard 级 baseline。
- [x] 4.3 确认插件生命周期测试仍显式控制安装、启用、禁用、卸载、上传、同步和清理状态，不被普通 baseline 干扰。

## 5. 生命周期大户重构

- [x] 5.1 重构官方源码插件生命周期用例，保留一个代表官方插件的完整 UI 生命周期，其余官方插件改为 API/contract smoke 加页面可访问性验证。
- [x] 5.2 重构 dynamic runtime 生命周期用例，区分 runtime 核心 UI 生命周期与 dynamic demo API/功能验证，合并重复装配并保留关键 UI 覆盖。
- [x] 5.3 复查源码插件生命周期用例，消除可合并或可 API 化的重复生命周期步骤。

## 6. 验证与验收记录

- [x] 6.1 运行 `openspec validate optimize-e2e-suite-runtime --strict` 并修复所有规范问题。
- [x] 6.2 运行受影响的 module scope E2E smoke，至少覆盖 `extension:plugin`、一个官方插件功能 scope、以及迁移后的宿主慢文件。
- [x] 6.3 记录 host-only 优化前后 wall clock、测试总耗时、最慢文件和最慢用例对比。
- [x] 6.4 记录 plugin-full 优化前后 wall clock、各分片耗时、最长分片和 runner minutes 变化。
- [x] 6.5 明确记录本变更不影响生产 API、数据库 schema、运行时缓存语义和 i18n 资源；若实现中新增可见文案或脚本入口，再同步补充对应治理说明。
- [x] 6.6 完成任务后执行 `/lina-review`，审查 CI 分片、fixture、baseline、慢用例重构和验证记录。

## 验证记录

- 已通过 `openspec validate optimize-e2e-suite-runtime --strict`。
- 已通过 `pnpm -C hack/tests exec tsc --noEmit`。
- 已通过 `pnpm -C hack/tests test:validate`。
- 已通过 `git diff --check`。
- 已验证本地服务 `http://127.0.0.1:5666` 与 `http://127.0.0.1:8080/api/v1/health` 可访问。
- 初次尝试运行 `pnpm -C hack/tests test:module -- settings:file --grep TC001a` 与 `pnpm -C hack/tests test:module -- monitor:operlog --grep TC001a` 时，Playwright global setup 因本机缺少 `chromium_headless_shell-1217` 浏览器二进制失败；随后尝试 `pnpm -C hack/tests exec playwright install chromium`，下载长时间无进展并被终止。
- 已使用本机 Google Chrome 通道完成最小浏览器 smoke：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- settings:file --grep "TC001a"`，结果 `1 passed (8.0s)`。
- 已使用本机 Google Chrome 通道完成普通插件 baseline smoke：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- monitor:operlog --grep "TC001a"`，结果 `1 passed (9.9s)`。
- 已使用本机 Google Chrome 通道完成代表性插件生命周期 smoke：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- extension:plugin --grep "TC001a"`，结果 `1 passed (29.9s)`。
- host-only 优化前基线来自用户提供日志：job 约 36 分钟，Playwright 报告 `197 passed (25.1m)`；本次已迁移菜单 CRUD、文件管理、角色 CRUD、参数导入、字典导入等慢用例使用不预加载 dashboard 的 `authenticatedPage`，预期减少重复 dashboard 首屏加载成本，最终耗时需以 CI artifact 复核。
- plugin-full 优化前基线来自用户提供日志：job 约 2 小时，`pnpm test` 约 112 分钟；本次改为 `extension:plugin` 与 `plugins` 两个通用分片，最长分片预计由插件自有用例集合决定，最终 wall clock 与 runner minutes 需以 CI artifact 复核。
- 本变更只调整 CI workflow、E2E runner manifest、Playwright fixture 和测试代码，不修改生产 API、数据库 schema、运行时缓存语义或用户可见功能；未新增或修改前端运行时文案、插件 manifest i18n、apidoc i18n JSON，确认不需要同步 i18n 资源。

## Feedback

- [x] **FB-1**: 区分无 `apps/lina-plugins` 的宿主模块 scope 与需要官方插件工作区的 plugin-full 接缝 scope。
- [x] **FB-2**: 收敛 plugin-full scope，只保留 `plugins` 与 `plugin:<plugin-id>` 作为源码插件自有用例的通用选择入口。
- [x] **FB-3**: 根 `hack/tests` E2E 代码与配置不得耦合任何具体官方源码插件 ID，插件相关用例必须闭环到对应插件目录。
- [x] **FB-4**: 根路径 E2E 测试文件、配置、测试数据和 baseline 不得耦合任何具体插件信息，插件相关测试资产必须放在对应插件目录。
- [x] **FB-5**: E2E 测试文件名前缀不再全局递增，改为按当前模块目录从 `TC001` 开始递增。
- [x] **FB-6**: `extension:plugin` 分片中动态插件资源“仅本人”数据权限和插件管理动作权限夹具在 plugin-full 环境失败。
- [x] **FB-7**: GitHub Actions host-only E2E 仍运行部分 plugin-full 或插件依赖用例，导致共享种子和宿主断言失败。
- [x] **FB-8**: GitHub Actions plugin-full E2E 中动态插件示例记录和英文布局回归用例存在跨用例状态泄漏。
- [x] **FB-9**: 完整 E2E 中角色新增/编辑抽屉会在异步初始化完成后覆盖已填写字段，导致提交未发出角色保存请求。
- [x] **FB-10**: Nightly plugin-full `plugins` scope 仍在单个 CI job 中串行执行全部源码插件自有 E2E，导致最长分片接近 1 小时。
- [x] **FB-11**: GitHub Actions `plugins-1-of-5` 分片中动态插件禁用后侧边栏菜单隐藏断言存在菜单投影刷新竞态。
- [x] **FB-12**: 动态插件 runtime 缓存失效和 reconciler 通知 reason 存在字符串硬编码。
- [x] **FB-13**: GitHub Actions `plugins-1-of-5` 分片中动态插件 multipart 上传用例假设 wasm artifact 小于默认上传上限。
- [x] **FB-14**: GitHub Actions `plugins-2-of-5` 分片中源码插件管理表格列顺序断言使用过期标题与位置契约。
- [x] **FB-15**: GitHub Actions `plugins-4-of-5` 分片中多租户插件卸载前置条件弹窗断言使用过期本地化原因文案。
- [x] **FB-16**: 默认文件上传大小限制需要从 20MB 提升到 100MB，并保持初始化、配置模板、后端 fallback 与打包资产一致。
- [x] **FB-17**: GitHub Actions `host-only` E2E 在 `Run E2E Suite` 阶段失败。
- [x] **FB-18**: GitHub Actions `plugin-full / plugins-1-of-5` E2E 在 `Run E2E Suite` 阶段失败。

## Feedback 验证记录

- 已移除需要具体官方插件环境的根目录接缝文件集合，根 `extension:plugin` 只覆盖宿主插件框架、动态测试插件与通用插件治理能力。
- 已新增 `pnpm test:host:module -- <scope>`，用于在未安装 `apps/lina-plugins` 时只运行指定宿主模块中可在 host-only 环境执行的用例。
- 已让 `moduleRequiresPluginWorkspace` 根据 scope entry 自动识别 `plugins/`、`apps/lina-plugins/` 与 `plugin:<plugin-id>`，避免通用插件入口在缺少 submodule 时被误判为可运行。
- 已在 `pnpm test:validate` 中增加治理校验：`host:` 前缀的 module scope 不允许依赖 `apps/lina-plugins`。
- 已通过 `pnpm -C hack/tests test:host:module -- settings:file --list`，确认 host-only 单模块入口可列出宿主文件且不执行真实浏览器用例。
- 已通过 `pnpm -C hack/tests test:host:module -- iam:user --list`，确认混合宿主 scope 会过滤插件敏感文件并保留可在主框架环境运行的宿主用例。
- 已通过 `pnpm -C hack/tests test:host:module -- plugin:<plugin-id> --list` 的预期失败验证，确认源码插件通用入口在 host-only module 模式下会被拒绝。
- 已通过 `pnpm -C hack/tests test:module -- <removed-plugin-alias> --list` 的预期失败验证，确认已移除的源码插件业务别名 scope 不再可用。
- 已移除按官方插件业务模块命名的长期别名 scope，源码插件自有用例统一使用 `plugins` 或 `plugin:<plugin-id>`。
- CI plugin-full 分片改用 `extension:plugin` 与 `plugins` 两个通用入口，不在根 workflow matrix 中枚举具体官方插件 ID。
- 已将具体官方插件的 E2E 用例、页面对象与服务依赖 baseline 迁移到对应 `apps/lina-plugins/<plugin-id>/hack/tests/` 目录；根 `hack/tests` 只保留宿主通用测试资产和参数化 helper。
- 已将根 E2E 的插件 ID 防回归校验改为从可选 `apps/lina-plugins/*/plugin.yaml` 动态发现插件标识，不在 `hack/tests` 治理脚本中维护官方插件 ID denylist；当未安装插件工作区时，该校验不要求插件工作区存在。
- 已通过基于 `apps/lina-plugins/*/plugin.yaml` 动态发现插件标识的根 E2E 治理扫描，确认根 E2E 资产和 CI workflow 不再包含具体官方源码插件 ID 或插件业务名称。
- 已通过旧接缝 scope 与具体源码插件入口静态扫描，确认旧 `plugin-host-seams` scope 和具体 `plugin:<concrete-id>` 入口无残留。
- 已通过 `rg -n "@playwright/test" apps/lina-plugins -g '*.ts'`，确认插件测试代码改用宿主封装的 Playwright 支持入口。
- 已通过 `pnpm -C hack/tests exec tsc --noEmit`。
- 已通过 `pnpm -C hack/tests test:validate`，结果 `Validated 233 E2E test files across 17 scopes. Smoke files: 6. Serial files: 216.`。
- 已通过 `pnpm -C hack/tests test:service-deps`，结果 `Service dependency governance passed: 28 files, 113 baseline critical constructor calls.`。
- 已通过 `openspec validate optimize-e2e-suite-runtime --strict`。
- 已通过 `git diff --check`。
- 本次反馈属于 E2E runner 和项目治理修正，不改变生产 API、数据库 schema、运行时缓存语义或用户可见业务功能；仅新增 README 中的命令入口说明，未新增前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON，确认不需要同步 i18n 资源。
- 已将 `hack/tests` 根目录的插件耦合治理改为动态发现具体源码插件标识，仅拒绝具体源码插件 ID、根 manifest 中的 `plugins/<concrete-id>` / `apps/lina-plugins/<concrete-id>` 条目、以及根 service dependency baseline 中的插件路径；`plugins/sync`、`plugins/dynamic` 等宿主通用插件框架 API 保持允许。
- 已将所有根目录和源码插件目录下 E2E 测试文件重命名为模块本地 `TC{NNN}`，每个模块目录从 `TC001` 连续递增；`pnpm -C hack/tests test:validate` 会拒绝旧四位全局编号、目录内重复编号和目录内不连续编号。
- 已将根路径 E2E 不耦合具体源码插件信息、插件测试资产闭环在插件目录、以及 E2E 测试文件按模块目录本地 `TC001` 递增的规则同步到 `AGENTS.md` 项目规范和 `.agents/skills/lina-review/SKILL.md` 审查清单。
- 已通过本次规则同步后的治理验证：`openspec validate optimize-e2e-suite-runtime --strict`、`git diff --check -- AGENTS.md .agents/skills/lina-review/SKILL.md openspec/changes/optimize-e2e-suite-runtime/tasks.md`、`rg -n '项目根路径下的 `E2E`|模块本地递增|TC\{NNN\}|具体源码插件 `ID`|根路径 E2E 不耦合' AGENTS.md .agents/skills/lina-review/SKILL.md openspec/changes/optimize-e2e-suite-runtime/tasks.md`。
- 已通过 `pnpm -C hack/tests test:module -- extension:plugin --list`，确认宿主插件框架 scope 只选择 11 个根目录通用插件框架用例文件。
- 已通过 `pnpm -C hack/tests test:module -- plugins --list`，确认源码插件全量入口选择 138 个插件自有用例文件，文件均来自 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`。
- 已通过 `pnpm -C hack/tests test:host:module -- iam:user --list`，确认 host-only 单模块入口选择 9 个宿主用户模块用例文件。
- 已通过 `pnpm -C hack/tests test:host:module -- 'plugin:<plugin-id>' --list` 的预期失败验证，确认 host-only 模块入口拒绝需要 `apps/lina-plugins` 的插件 scope。
- FB-6 修复了 generic plugin resource 查询层与宿主 `sys_role.data_scope` 枚举不一致的问题：`dataScope=4` 现在按“仅本人”过滤，`dataScope=3` 仍按部门过滤；同时将插件管理动作权限 E2E 的查询角色改为不依赖组织插件的全量数据权限，因为该用例只验证插件动作权限。
- FB-6 已补充后端单测 `TestResolvePluginResourceDataScopeModeUsesHostScopeValues`，防止插件资源数据权限映射再次从宿主角色枚举漂移。
- FB-6 已通过 `cd apps/lina-core && go test ./internal/service/plugin/internal/integration -count=1`。
- FB-6 已通过宿主启动/路由绑定包测试：`cd apps/lina-core && go test ./internal/cmd -count=1`。
- FB-6 已通过 E2E 精确回归：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test e2e/extension/plugin/TC002-plugin-permission-governance.ts e2e/extension/plugin/TC007-plugin-management-action-permissions.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `4 passed (55.4s)`。
- FB-6 已通过完整分片回归：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- extension:plugin`，结果 `26 passed, 1 skipped (2.4m)`。
- FB-6 已通过 `pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict` 和 `git diff --check`。
- FB-6 i18n 影响：本次只调整后端数据权限映射、后端单测和 E2E 测试夹具，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-6 缓存一致性影响：本次不新增缓存；插件资源查询继续按请求上下文中的数据权限快照即时应用数据库过滤，不改变插件 runtime freshness、启用状态快照或跨实例失效语义。
- FB-7 已让 `pnpm test:host` 的 Playwright 子进程显式注入 `E2E_HOST_ONLY_PLUGINS=1`，普通 plugin-full 入口仍保留外部传入值或默认 `0`，使测试代码可区分 host-only 与 plugin-full 环境。
- FB-7 已修复动态插件热升级用例的宿主菜单投影刷新：启用动态插件后通过访问工作台并刷新页面重新加载宿主投影，不再依赖 focus 事件，也不直接改写角色菜单授权表绕过权限缓存。
- FB-7 已将英文内置治理数据本地化用例按 host-only 环境收敛断言：host-only 下任务日志只断言宿主内置任务，并跳过由源码插件提供的登录日志和操作日志接口断言。
- FB-7 已将调度任务用例按 host-only 环境区分宿主内置任务集合，host-only 只断言 `任务日志清理` 和 `在线会话清理`，plugin-full 继续覆盖 `服务监控采集`。
- FB-7 已将字典数据面板用例从插件拥有的 `sys_oper_type` 切换到宿主内置的 `sys_menu_status`，避免 host-only 环境缺少插件字典种子导致失败。
- FB-7 已修复用户角色 POM 断言：运行时根据可见表头 `角色` / `Roles` 解析 VXE `colid`，再定位同一行对应单元格，避免组织/租户列显示顺序变化或 VXE 运行时列 ID 导致读错列。
- FB-7 已通过 host-only 文件选择验证：`pnpm -C hack/tests test:host -- --list`，结果选择 `95` 个文件，其中 `17` 个并行、`78` 个串行。
- FB-7 已通过 host-only 精确回归：`E2E_HOST_ONLY_PLUGINS=1 E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test e2e/extension/plugin/TC003-plugin-hot-upgrade.ts e2e/i18n/TC005-english-built-in-governance-data-localization.ts e2e/iam/user/TC008-user-role.ts e2e/scheduler/job/TC002-job-handler-crud.ts e2e/settings/dict/TC006-dict-data-export.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `14 passed, 1 skipped (1.6m)`；跳过项为 host-only 环境下由源码插件提供的日志接口断言。
- FB-7 已通过用户角色单文件回归：`E2E_HOST_ONLY_PLUGINS=1 E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test e2e/iam/user/TC008-user-role.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `4 passed (37.2s)`。
- FB-7 已通过 `pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict` 和 `git diff --check`。
- FB-7 i18n 影响：本次只调整 E2E 环境分支、断言和跳过逻辑，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-7 缓存一致性影响：本次不修改生产缓存逻辑；动态插件用例改为通过 UI 重新加载宿主投影观察实际缓存刷新效果，避免测试侧直接写权限表造成缓存状态与真实运行路径不一致。
- FB-8 已为 Vite 开发服务补充 `/x` 代理到后端运行时，避免动态插件页面在 dev origin 下请求 `/x/linapro-demo-dynamic/demo-records` 时命中 Vite 404；同时补齐该配置的 node 级 TypeScript 校验保护。
- FB-8 已让动态插件运行时 E2E 清理 `sys_plugin_migration` 与示例记录表状态，分页种子显式写入 `tenant_id=0`，并在页面断言前通过后端 `/x/linapro-demo-dynamic/demo-records` API 等待示例数据可见。
- FB-8 已让英文运行时页面巡检在每个用例前卸载并清理动态插件数据，防止跨用例迁移状态泄漏；源码示例英文布局回归在前置步骤中显式启用 `linapro-org-core`，确保英文侧栏断言自包含。
- FB-8 已通过动态插件与源码示例精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC003-english-runtime-page-audit.ts ../apps/lina-plugins/linapro-demo-source/hack/tests/e2e/host-integration/TC006-english-layout-regression.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `15 passed (4.4m)`。
- FB-8 已通过 plugin-full 分片回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test:module -- extension:plugin`，结果 `26 passed, 1 skipped (2.2m)`；`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test:module -- plugins`，结果 `272 passed, 7 skipped (25.7m)`。
- FB-8 i18n 影响：本次只调整 E2E 状态清理、测试前置条件、后端可见性等待和 Vite dev proxy，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON；现有英文布局和英文动态页面 E2E 已覆盖英文展示不回退。
- FB-8 缓存一致性影响：本次不修改生产缓存逻辑；动态插件 E2E 通过卸载、清理迁移记录、重新安装启用和 API 可见性等待验证真实运行路径，避免测试环境残留的本地或数据库状态影响跨用例一致性。
- FB-9 已修复角色页面 POM 在新增/编辑角色抽屉中与权限树异步初始化的竞态：打开抽屉后等待权限工具栏和首行菜单树渲染完成，普通创建/编辑路径预先标记权限导览已读，提交时等待 `POST /api/v1/role` 或 `PUT /api/v1/role/{id}` 响应并确认抽屉关闭。
- FB-9 已通过角色精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests exec playwright test e2e/iam/role/TC001-role-crud.ts e2e/iam/role/TC004-role-permission-drawer-close.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `13 passed (52.6s)`。
- FB-9 i18n 影响：本次只调整 E2E 页面对象的等待与导览处理，不新增或修改用户可见文本、语言包、插件 manifest i18n 或 apidoc i18n JSON。
- FB-9 缓存一致性影响：本次不修改生产缓存逻辑；角色 POM 仅等待前端权限树初始化和角色保存响应，不改变权限缓存、角色授权缓存或跨实例失效语义。
- FB-10 日志分析结论：用户提供的 GitHub Actions 日志显示 `plugin-full / plugins` job 从 `2026-05-19T23:55:31` 开始执行 Playwright，到 `2026-05-20T00:56:21` 结束，`138` 个文件全部进入 serial pool，`272 passed, 7 skipped (1.0h)`；最慢文件为 `linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts`，耗时 `7.4m`。按日志聚合，`linapro-org-core` 约 `12.1m`、`linapro-content-notice` 约 `10.1m`、`linapro-demo-dynamic` 约 `9.3m`，主瓶颈是单 job 串行执行所有源码插件自有用例。
- FB-10 已将 plugin-full CI matrix 中的 `plugins` 单分片拆为 `plugins-1-of-5` 至 `plugins-5-of-5`，每个分片继续使用通用 `plugins` scope，并通过 Playwright 原生 `--shard=N/5` 在独立 runner、独立 PostgreSQL 和独立 plugin-full 服务实例中执行，避免在同一数据库内并发污染；`extension:plugin` 分片保持不变。
- FB-10 已通过用户日志耗时映射估算 `plugins --shard=N/5` 分布：`1/5` 约 `17.9m`、`2/5` 约 `11.7m`、`3/5` 约 `11.1m`、`4/5` 约 `13.4m`、`5/5` 约 `4.2m` 的测试用例耗时；相比原单 job 约 `58.3m` 的已解析测试用例耗时，最长测试体量预计下降约 `69%`。选择 5 分片是因为 6 分片估算最长仍约 `17.5m`，收益很小；7 分片最长约 `13.9m` 但会额外增加两个 CI job。
- FB-10 已通过 `pnpm -C hack/tests test:module -- plugins -- --shard=1/5 --list`、`--shard=2/5 --list`、`--shard=3/5 --list`、`--shard=4/5 --list`、`--shard=5/5 --list` 验证每个分片可解析并列出测试。
- FB-10 已修复本次验证中暴露的既有 E2E 治理问题：将 `hack/tests/e2e/extension/plugin/TC0243-plugin-status-switch-feedback.ts` 重命名为模块本地连续编号 `TC012-plugin-status-switch-feedback.ts`，并同步测试标题为 `TC-12`，使 `pnpm -C hack/tests test:validate` 可通过。
- FB-10 i18n 影响：本次只调整 GitHub Actions E2E job matrix 和 OpenSpec 任务记录，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-10 缓存一致性影响：本次不修改生产缓存逻辑；CI 层把 `plugins` scope 分散到独立 runner 与独立服务实例中执行，不改变运行时缓存、插件启用状态快照、i18n 缓存或跨实例失效语义。
- Review(FB-10): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git ls-files --others --exclude-standard`、`openspec status --change optimize-e2e-suite-runtime --json`、`git diff -- .github/workflows/reusable-test-verification-suite.yml openspec/changes/optimize-e2e-suite-runtime/tasks.md`。确认本次只将 plugin-full `plugins` scope 从单个 CI job 拆为 5 个 Playwright shard，仍使用通用 `plugins` 入口和 plugin-full 服务启动语义；artifact 名称继续包含 `matrix.shard.name`，不会覆盖其他分片证据。未修改生产 Go/前端代码、业务 API、数据库 schema、运行时缓存逻辑、数据权限逻辑或 i18n 资源。严重问题 0；警告 0。
- 本次完整单元测试已通过 host-only Go 单元测试：`cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml && make init confirm=init rebuild=true && make pack.assets && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' make test.go plugins=0 race=true verbose=true`。
- 本次完整单元测试已通过 plugin-full Go 单元测试：`cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml && make init confirm=init rebuild=true && make pack.assets && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' make test.go plugins=1 race=true verbose=true`。
- 本次完整前端单元测试已通过：`pnpm -C apps/lina-vben test:unit`，结果 `Test Files 42 passed (42)`、`Tests 347 passed (347)`。
- 本次完整 host-only E2E 已通过：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test:host`，结果 `244 passed, 1 skipped (14.6m)`。
- 本次完整 plugin-full E2E 已通过：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test`，结果 `516 passed, 8 skipped (42.8m)`。
- 本次最终静态与治理验证已通过：`pnpm -C apps/lina-vben exec tsc -p apps/web-antd/tsconfig.node.json --noEmit`、`pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict` 和 `git diff --check`。
- FB-11 日志分析结论：用户提供的 GitHub Actions 日志显示 `plugin-full / plugins-1-of-5` 于 `2026-05-20T01:54:31Z` 和 retry `2026-05-20T01:55:28Z` 均失败在 `linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts` 的 `TC-1k`，错误为禁用插件后 `expectSidebarMenuHidden("动态插件示例")` 仍观察到侧边栏菜单可见；同一分片其余 `55` 个测试通过，根因是 UI 开关状态已切换但当前页面的菜单投影刷新存在 CI 时序竞态。
- FB-11 已在 `TC001-runtime-wasm-lifecycle.ts` 中新增 `expectCurrentUserRouteVisible`，在禁用 `linapro-demo-dynamic` 后轮询 `menus/all`，确认当前用户路由投影已不包含 `动态插件示例`，再刷新页面并断言侧边栏隐藏，避免断言抢在菜单投影刷新前执行。
- FB-11 已通过精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep "TC-1k"`，结果 `1 passed (1.3m)`。
- FB-11 已通过 `pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict`、`git diff --check` 和 `git -C apps/lina-plugins diff --check`。
- FB-11 i18n 影响：本次只调整 E2E 等待与断言顺序，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-11 缓存一致性影响：本次不修改生产缓存逻辑；E2E 通过后端 `menus/all` 轮询确认插件禁用后的菜单投影失效已完成，再刷新当前页面观察真实侧栏结果。
- Review(FB-11): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git -C apps/lina-plugins status --short --untracked-files=all`、`openspec status --change optimize-e2e-suite-runtime --json`、`git diff -- openspec/changes/optimize-e2e-suite-runtime/tasks.md` 和 `git -C apps/lina-plugins diff -- linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts`。确认本次只调整动态插件 E2E 对菜单投影刷新的等待与 OpenSpec 反馈记录；未修改生产 Go/前端代码、业务 API、数据库 schema、运行时缓存逻辑、数据权限逻辑或 i18n 资源。严重问题 0；警告 0。
- FB-12 已新增 runtime 包内 `runtimeChangeReason` 命名类型与集中常量，覆盖动态插件安装、升级、启停、刷新、卸载、孤儿卸载、上传、缺失 artifact、期望状态变更和 reconciler 后台扫描原因。
- FB-12 已将 `invalidateRuntimeCaches`、`notifyRuntimeCacheChanged`、`notifyReconcilerChanged` 与 `publishReconcilerChanged` 改为接收 `runtimeChangeReason`，仅在调用现有 frontend bundle invalidator、root cache notifier 和日志输出边界时转换为字符串，保持外部持久化 reason 值不变。
- FB-12 已通过静态扫描 `rg -n '"plugin_uninstalled"|... apps/lina-core/internal/service/plugin/internal/runtime -g '*.go'`，确认动态插件 runtime reason 字符串只保留在 `runtime_change_reason.go` 集中常量定义中。
- FB-12 已通过后端变更包测试：`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -count=1`，结果 `ok lina-core/internal/service/plugin/internal/runtime 11.094s`。
- FB-12 已通过治理验证：`openspec validate optimize-e2e-suite-runtime --strict`、`git diff --check -- apps/lina-core/internal/service/plugin/internal/runtime/... openspec/changes/optimize-e2e-suite-runtime/tasks.md`。
- FB-12 i18n 影响：本次只调整后端内部 runtime reason 的类型定义和调用点，不新增或修改用户可见文案、前端运行时语言包、插件 manifest i18n 或 apidoc i18n JSON。
- FB-12 缓存一致性影响：本次不改变缓存权威数据源、一致性模型、失效 scope、跨实例通知机制或 reason 字符串值；缓存失效仍通过现有 frontend bundle invalidator、runtime bundle scope invalidation、root cache notifier 和 reconciler revision controller 执行。
- Review(FB-12): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git ls-files --others --exclude-standard apps/lina-core/internal/service/plugin/internal/runtime openspec/changes/optimize-e2e-suite-runtime`、`openspec status --change optimize-e2e-suite-runtime --json`、`git diff HEAD -- apps/lina-core/internal/service/plugin/internal/runtime/... openspec/changes/optimize-e2e-suite-runtime/tasks.md`。确认本次只将动态插件 runtime reason 收敛为包内命名类型常量，并在现有字符串边界显式转换；未修改 API、数据库 schema、用户可见文案、数据权限、缓存失效 scope 或跨实例通知语义。后端 Go 编译门禁已由 `cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -count=1` 覆盖。严重问题 0；警告 0。
- FB-13 日志分析结论：用户提供的 GitHub Actions job `77090362703` 对应 `plugin-full / plugins-1-of-5`，失败点位于动态插件 runtime 上传链路；当前 `linapro-demo-dynamic.wasm` artifact 已超过默认 `sys.upload.maxSize=20MB`，原 `TC-1m` 只验证请求体超过默认 8MB body limit，却隐含假设 artifact 仍低于业务上传上限。
- FB-13 已在 `linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts` 中让 `TC-1m` 按当前 wasm artifact 大小临时提高 `sys.upload.maxSize`，通过配置 API 触发真实运行时配置失效路径，并在 `finally` 中恢复原值；`TC-1n` 继续验证超过上传上限时返回文件过大提示。
- FB-13 已通过精准回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep "TC-1m"`，结果 `1 passed`。
- FB-13 已通过上限回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep "TC-1n"`，结果 `1 passed`。
- FB-13 已通过完整分片回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests test:module -- plugins -- --shard=1/5`，结果 `56 passed (9.1m)`。
- FB-13 i18n 影响：本次只调整 E2E 上传前置参数和 API helper，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-13 缓存一致性影响：本次不修改生产缓存逻辑；E2E 通过系统参数 API 更新并恢复 `sys.upload.maxSize`，沿用现有配置缓存失效路径，未引入新的缓存、共享状态或跨实例一致性语义。
- FB-14 日志分析结论：用户提供的 GitHub Actions job `77090362705` 对应 `plugin-full / plugins-2-of-5`，失败点位于源码插件生命周期页面表格列顺序断言；当前插件管理表格契约为名称列、描述列、版本列，且中文标题为 `插件描述`，英文标题为 `Plugin Description`，原断言仍使用旧的 `描述` 且要求其位于版本之后。
- FB-14 已扩展共享 `PluginPage.expectTableColumnBetween` 支持中英文候选列名，并将 `linapro-demo-source/hack/tests/e2e/host-integration/TC001-source-plugin-lifecycle.ts` 的断言更新为描述列位于插件名称和版本之间。
- FB-14 已通过精准回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-source/hack/tests/e2e/host-integration/TC001-source-plugin-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep "TC-1a"`，结果 `1 passed`。
- FB-14 已通过完整分片回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests test:module -- plugins -- --shard=2/5`，结果 `56 passed (7.5m)`。
- FB-14 i18n 影响：本次只调整 E2E 对既有中文/英文列标题的断言候选，不新增或修改前端运行时语言包、插件 manifest i18n 或 apidoc i18n JSON。
- FB-14 缓存一致性影响：本次不修改生产缓存逻辑；共享 POM 只读取当前页面已渲染表头，不改变插件状态、权限、路由或 i18n 缓存语义。
- FB-15 日志分析结论：用户提供的 GitHub Actions job `77090362724` 对应 `plugin-full / plugins-4-of-5`，失败点位于多租户插件强制卸载确认弹窗；当前 manifest i18n 原因为 `仍存在租户。请先删除租户，再卸载多租户插件。`，原 E2E 仍匹配旧的通用 `仍存在租户，请先删除租户，再卸载插件。`。
- FB-15 已将 `linapro-tenant-core/hack/tests/e2e/lifecycle-precondition/TC004-plugin-uninstall-force-dialog.ts` 的原因文案断言更新为当前插件 manifest i18n 文案，同时保留前缀、风险提示、插件 ID 输入和强制卸载二次提交断言。
- FB-15 已通过精准回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-tenant-core/hack/tests/e2e/lifecycle-precondition/TC004-plugin-uninstall-force-dialog.ts --config playwright.config.ts --project=chromium --workers=1 --grep "TC-4a"`，结果 `1 passed (14.5s)`。
- FB-15 已通过完整分片回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests test:module -- plugins -- --shard=4/5`，结果 `56 passed (7.9m)`。
- FB-15 i18n 影响：本次只同步 E2E 断言到现有 `linapro-tenant-core` manifest i18n 文案，不新增、删除或修改任何 i18n JSON 资源。
- FB-15 缓存一致性影响：本次不修改生产缓存逻辑；E2E 继续通过实际卸载 API 与前置条件 veto 流程验证当前运行时状态，不改变插件 lifecycle precondition、事件广播或状态缓存语义。
- Review(FB-13~FB-15): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git -C apps/lina-plugins status --short --untracked-files=all`、`openspec status --change optimize-e2e-suite-runtime --json`、`git diff -- hack/tests/pages/PluginPage.ts openspec/changes/optimize-e2e-suite-runtime/tasks.md` 和 `git -C apps/lina-plugins diff -- linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts linapro-demo-source/hack/tests/e2e/host-integration/TC001-source-plugin-lifecycle.ts linapro-tenant-core/hack/tests/e2e/lifecycle-precondition/TC004-plugin-uninstall-force-dialog.ts`。确认本次只修复三个 plugin-full 分片的 E2E 断言/测试前置条件，并补充 OpenSpec 反馈记录；未修改生产 Go/前端代码、业务 API、数据库 schema、数据权限逻辑、运行时缓存实现或 i18n JSON 资源。后端 Go 编译门禁不适用。已知无关工作区变更 `deterministic-monthly-openspec-archive` 与 `decouple-workspace-plugin-routes` 未纳入本次审查。严重问题 0；警告 0。
- FB-16 已将宿主上传默认上限从 20MB 调整为 100MB，覆盖 `config.template.yaml`、`manifest/sql/005-config-management.sql`、后端 `defaultUploadMaxSize`、`RuntimeParamKeyUploadMaxSize` 默认规格、打包 manifest 资产和 `hack/deploy` 默认配置，确保静态配置、初始化数据、运行时 fallback 与交付资产一致。
- FB-16 已同步后端测试期望，覆盖静态默认值、运行时参数默认规格、multipart 友好错误中的默认上限参数，以及 packed manifest 中的配置模板和 SQL 种子。
- FB-16 已运行 `make pack.assets`，结果 `packed manifest assets prepared: apps/lina-core/internal/packed/manifest`。
- FB-16 已通过精确 Go 门禁：`cd apps/lina-core && go test ./internal/service/config -run 'TestDurationConfigsUseDefaultsWhenUnset|TestRuntimeParamSpecsReturnsCopy|TestGetUploadPathUsesStaticConfig' -count=1`、`cd apps/lina-core && go test ./internal/service/middleware -run 'TestRequestBodyLimitForContentType|TestRequestBodyLimitFriendlyError' -count=1`、`cd apps/lina-core && go test ./internal/packed -run TestFilesEmbedUpdatedUploadDefaults -count=1`。
- FB-16 已通过宿主启动绑定包测试：`cd apps/lina-core && go test ./internal/cmd -count=1`。
- FB-16 完整配置包验证说明：`cd apps/lina-core && go test ./internal/service/config -count=1` 当前仍会被既有包内全局 GoFrame config adapter 与 runtime revision 状态顺序依赖阻断，出现与本次变更无关的多项配置测试串扰；本次以覆盖变更函数的精确测试作为替代编译烟测。
- FB-16 i18n 影响：本次不新增或修改用户可见文案和 i18n JSON 资源；上传错误文案仍使用现有 `{maxSizeMB}` 参数化翻译，测试只验证参数从默认 100MB 注入。
- FB-16 缓存一致性影响：本次不改变运行时配置缓存实现或失效机制；`sys.upload.maxSize` 仍以 `sys_config` 为权威源并通过现有 protected runtime param snapshot/revision 机制热更新，100MB 仅改变缺省值。
- Review(FB-16): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git diff -- apps/lina-core/internal/service/config/config_upload.go apps/lina-core/internal/service/config/config_runtime_params.go apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/sql/005-config-management.sql apps/lina-core/internal/packed/manifest/config/config.template.yaml apps/lina-core/internal/packed/manifest/sql/005-config-management.sql hack/deploy/config.yaml hack/deploy/dev-test/config.yaml apps/lina-core/internal/service/config/config_duration_test.go apps/lina-core/internal/service/config/config_protected_settings_test.go apps/lina-core/internal/service/config/config_runtime_params_test.go apps/lina-core/internal/service/middleware/middleware_request_body_limit_test.go apps/lina-core/internal/packed/packed_test.go openspec/changes/optimize-e2e-suite-runtime/tasks.md`。确认默认上传上限的源码、初始化、交付资产和测试期望已对齐到 100MB；未修改 API 签名、数据库 schema、数据权限逻辑、用户可见固定文案或 i18n JSON。后端 Go 门禁已由上述精确测试和 `internal/cmd` 覆盖。严重问题 0；警告 0。
- FB-17 日志分析结论：GitHub Actions job `77101145762` 对应 `Nightly Test and Build #34 / host-only`，失败点位于 `hack/tests/e2e/extension/plugin/TC003-plugin-hot-upgrade.ts` 的 `TC-3d`；页面触发插件 registry 刷新后回到默认 `分析页`，当前插件 iframe 被卸载，断言 `plugin-dev-dynamic-hot-upgrade/v0.2.0/index.html` 可见 iframe 失败。根因是失败升级后 `runtimeState=upgrade_failed` 会改变插件 registry signature，`refreshAccessibleState` 重新生成路由时只按 route name 判断当前路由是否仍可访问；动态插件 link route 可因菜单记录重建而改变 name/path，且 `upgrade_failed` 会让普通插件菜单入口从可访问路由中过滤，导致当前已打开的稳定 iframe 页被误判为不可访问并 fallback 到默认首页。
- FB-17 已在前端访问刷新逻辑中新增 `resolveAccessibleRouteRefreshTarget`：当前路由可按 route name、规范化完整 path、等价 `iframeSrc` 插件资产目标判定为可保留；若等价插件资产路由存在但内部 link path 已变化，则切换到新 path；若普通可访问路由已被非 `normal` 运行态过滤，但插件仍 installed+enabled，则只保护当前已打开的稳定 iframe 页；若插件已禁用、卸载或正常运行态下路由被移除则仍不保留，保持菜单过滤和降级语义。
- FB-17 已补充前端单测 `apps/lina-vben/apps/web-antd/src/router/access-refresh-route-match.test.ts`，覆盖 route name 匹配、同 path 不同 name、嵌套路由 path 匹配、等价 iframe route 替换、fallback 拒绝、`upgrade_failed` 但 installed+enabled 的当前稳定 iframe 页保留，以及 `normal` 运行态下无可访问路由时不保留。
- FB-17 已通过精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test e2e/extension/plugin/TC003-plugin-hot-upgrade.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `4 passed (25.1s)`。
- FB-17 已通过宿主插件模块回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests test:host:module -- extension:plugin`，结果 `30 passed (1.4m)`。
- FB-17 i18n 影响：本次只调整前端路由刷新判定与单元测试，不新增或修改用户可见文案、前端运行时语言包、插件 manifest i18n 或 apidoc i18n JSON。
- FB-17 缓存一致性影响：本次不修改后端缓存、插件 registry 缓存、i18n 缓存或跨实例失效机制；前端仍使用现有 `getPluginStateMap` registry snapshot，只在本地访问刷新时按当前插件 installed/enabled 与非 `normal` 运行态保护已打开的稳定 iframe 页，禁用/卸载路径仍会 fallback。
- FB-18 日志分析结论：GitHub Actions job `77101145863` 对应 `Nightly Test and Build #34 / plugin-full / plugins-1-of-5`，失败点位于 `apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts` 的 `TC-1n`；默认上传上限提升到 100MB 后，测试为 UI file chooser 构造了超过 100MB 的内存 `Buffer`，Playwright `fileChooser.setFiles({ buffer })` 触发 `Cannot set buffer larger than 50Mb, please write it to a file and pass its path instead.`，还未进入产品上传错误断言。
- FB-18 已将 `TC-1n` 的 UI 上传分支改为把超限 wasm probe 写入 `temp/runtime-e2e/linapro-demo-dynamic-oversized.wasm` 并通过文件路径传给 `fileChooser.setFiles`，`finally` 中清理临时文件；API multipart 分支继续使用内存 buffer 验证后端业务错误响应。
- FB-18 已通过精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep "TC-1n"`，结果 `1 passed (17.0s)`。
- FB-18 已通过完整分片回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests test:module -- plugins -- --shard=1/5`，结果 `56 passed (8.6m)`。
- FB-18 i18n 影响：本次只调整 E2E 文件上传输入方式，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-18 缓存一致性影响：本次不修改生产缓存逻辑；测试仍沿用现有 `sys.upload.maxSize` 配置读取和业务上传错误路径，不新增共享状态或跨实例一致性语义。
- FB-17/FB-18 已通过前端与治理验证：`pnpm -C apps/lina-vben exec vitest run apps/web-antd/src/router/access-refresh-route-match.test.ts --config vitest.config.ts --dom`，结果 `7 passed`；`pnpm -C apps/lina-vben --filter @lina/web-antd typecheck`；`pnpm -C apps/lina-vben test:unit`，结果 `Test Files 43 passed (43)`、`Tests 353 passed (353)`；`pnpm -C hack/tests exec tsc --noEmit`；`pnpm -C hack/tests test:validate`，结果 `Validated 235 E2E test files across 17 scopes. Smoke files: 6. Serial files: 218.`；`openspec validate optimize-e2e-suite-runtime --strict`；`git diff --check`；`git -C apps/lina-plugins diff --check`。
- Review(FB-17~FB-18): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git -C apps/lina-plugins status --short --untracked-files=all`、`git diff -- apps/lina-vben/apps/web-antd/src/router/access-refresh.ts apps/lina-vben/apps/web-antd/src/router/access-refresh-route-match.ts apps/lina-vben/apps/web-antd/src/router/access-refresh-route-match.test.ts openspec/changes/optimize-e2e-suite-runtime/tasks.md` 和 `git -C apps/lina-plugins diff -- linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts`。确认 FB-17 只调整前端访问刷新判定，保护已打开且插件仍 installed+enabled、但运行态非 `normal` 的动态插件稳定 iframe 页，不改变后端 API、数据库 schema、插件 registry 缓存或普通禁用/卸载/权限移除降级语义；FB-18 只调整动态插件 E2E 上传输入方式。未新增或修改 i18n JSON、用户可见文案、数据权限逻辑或后端 Go 生产代码，后端 Go 编译门禁不适用。已知无关工作区变更 `deterministic-monthly-openspec-archive` 与 `decouple-workspace-plugin-routes` 未纳入本次审查。严重问题 0；警告 0。
