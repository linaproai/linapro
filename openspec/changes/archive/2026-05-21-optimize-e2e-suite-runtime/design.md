## Context

当前 E2E 执行链路已经具备较好的治理基础：`run-suite.mjs` 支持 full、host、smoke 和 module 模式；`execution-manifest.json` 已声明模块范围、串行隔离文件和隔离类别；GitHub Actions 已将浏览器 E2E 封装为 `reusable-e2e-tests.yml`。因此本次优化应优先复用现有 runner 与 manifest，而不是引入新的测试框架或重写 Playwright 组织方式。

从 GitHub Actions 日志看，host-only 模式约 36 分钟，其中 Playwright 测试约 25 分钟，主要由大量 5-10 秒的 UI 用例累积形成。plugin-full 模式约 2 小时，其中 Playwright 测试约 112 分钟，主要由插件生命周期测试、插件自有功能测试和插件敏感宿主回归在单个 job 内串行执行导致。

## Goals / Non-Goals

**Goals:**

- 将 plugin-full E2E wall clock 优先降低到 45 分钟以内，同时保持官方插件覆盖。
- 将 host-only E2E wall clock 优先降低到 30 分钟以内，并保留现有宿主功能覆盖。
- 降低普通插件页面测试的重复前置条件成本。
- 保留插件生命周期、权限矩阵、运行时 i18n、共享数据库种子和文件系统产物等高风险测试的隔离语义。
- 让优化结果可以通过 CI 日志、Playwright report 和测试耗时记录复核。

**Non-Goals:**

- 不改变生产 API、数据库 schema、前端用户功能或插件运行时语义。
- 不为了提速而移除关键 E2E 覆盖；确需缩窄 plugin-full 范围时，必须把具体插件行为回收到所属插件目录内维护。
- 不引入新的外部测试服务或新 CI 平台。
- 不在本变更中调整产品文案或 i18n 资源。

## Decisions

### 1. Plugin-full 先按通用 module scope 做 CI 分片

使用现有 runner 能力拆分 plugin-full job，优先分片：

- `extension:plugin`
- `plugins`

每个分片继续使用 `make dev plugins=1` 和官方插件 submodule checkout，artifact 名称包含分片名。`extension:plugin` 只选择根目录中不依赖具体官方插件的宿主插件框架、动态测试插件与通用插件治理用例；`plugins` 作为全部源码插件自有测试的通用入口。`plugin:<plugin-id>` 仍作为本地或临时 CI 调试时选择单个源码插件自有测试的通用入口，但常规 CI 不再枚举具体官方插件 ID。这样可以最小化 runner 改动，让源码插件自有用例在统一入口下执行，同时避免 manifest 按官方插件业务模块维护长期别名。

备选方案是提高单 job 的 Playwright worker 数量，但当前 manifest 中大量 plugin-full 文件因共享状态被串行化，直接增加 worker 对 wall clock 收益有限，而且更容易引入状态污染。

### 2. Plugin-full 与 host-only 保持职责区分

host-only 继续覆盖宿主全量能力；plugin-full 聚焦根目录通用插件框架测试、动态测试插件运行时，以及每个官方源码插件目录内闭环维护的自有 E2E。根 `hack/tests` 不能依赖任何具体官方插件 ID、路径、菜单、mock data、i18n key 或页面定位器；凡是需要验证具体官方插件行为的用例，必须移动到 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/` 并通过 `plugin:<plugin-id>` 运行。

当只想在没有 `apps/lina-plugins` 的主框架环境中运行某个宿主模块时，使用 `pnpm test:host:module -- <scope>`。该入口复用 host-only 排除清单，只执行指定 scope 中不依赖官方插件工作区的宿主用例。

备选方案是让 plugin-full 继续执行 `pnpm test` 全量，但拆分全部模块。这会降低 wall clock，却会显著增加 runner minutes，且继续混淆 host-only 与 plugin-full 的验证职责。

### 3. 新增不自动导航的认证页面 fixture

保留现有 `adminPage` 兼容旧用例，同时新增轻量 fixture，例如 `authenticatedPage`。该 fixture 只创建带 admin storage state 的 page，不默认进入 `/dashboard/analytics`。迁移高耗时文件时，用例直接进入目标业务路由，避免 dashboard 首屏加载和后续业务页面跳转的重复成本。

备选方案是直接修改 `adminPage` 行为，但这会影响大量既有测试，风险更高。

### 4. 插件 baseline 按 suite 或 shard 级别幂等准备

新增共享辅助能力，让普通插件页面测试可以声明需要的插件集合，并在 suite 或 shard 级别执行一次：

- `syncPlugins`
- install/enable required plugins
- load plugin mock data when present
- refresh plugin projection

生命周期测试仍然自行控制安装、启用、禁用和卸载，避免 baseline 干扰被测状态。

备选方案是在每个测试文件继续 `beforeEach ensureSourcePluginEnabled`，实现简单但成本高，且重复刷新插件投影会持续放大 plugin-full 耗时。

### 5. 生命周期大户使用代表性完整链路加批量 contract smoke

官方插件生命周期测试不应对每个插件都完整跑 UI 安装、启用、停用、卸载、路由缺失和路由恢复。保留一个代表插件执行完整 UI 生命周期，其他官方插件使用 API lifecycle、菜单挂载和页面可访问性 smoke 验证。

Dynamic runtime 测试拆分为核心生命周期与 demo 功能验证。对不需要真实浏览器交互的上传大小、API bridge、数据保留等断言优先使用 API/request 层验证；需要验证宿主壳、iframe、新标签页和 embedded runtime 的场景保留 UI 覆盖。

## Risks / Trade-offs

- [Risk] CI 分片会增加总 runner minutes 和 artifact 数量。→ 使用有限分片起步，优先拆插件相关范围；artifact 名称中加入 shard name，便于定位失败。
- [Risk] 分片后每个 job 独立初始化数据库和服务，可能暴露隐含的跨文件依赖。→ 保持测试文件独立性要求，失败时优先修复 fixture 前置条件，而不是恢复跨文件依赖。
- [Risk] 新 fixture 与旧 `adminPage` 并存可能形成选择混乱。→ 仅先迁移高耗时文件，并在 fixture 注释和任务记录中明确适用场景。
- [Risk] 插件 baseline 可能掩盖生命周期测试期望的初始状态。→ baseline 只用于普通插件功能测试；生命周期目录继续显式清理和装配。
- [Risk] 缩窄 plugin-full 范围可能遗漏宿主与插件框架回归。→ 根目录只保留通用插件框架与动态测试插件覆盖；具体官方插件行为必须闭环到对应插件目录，并通过 `plugins` 或 `plugin:<plugin-id>` 选择。

## Migration Plan

1. 先修改 CI workflow，使 plugin-full 支持 module shard，并修正 PostgreSQL health check。
2. 运行或通过 GitHub Actions 验证各分片可独立启动、独立上传 artifact、失败能阻止下游发布。
3. 新增 `authenticatedPage`，优先迁移 2 个高耗时宿主文件验证收益。
4. 新增插件 baseline 辅助，并迁移普通插件功能测试中的重复 `ensureSourcePluginEnabled`。
5. 重构官方源码插件生命周期与 dynamic runtime 生命周期用例，用耗时日志确认收益。
6. 根据结果决定是否继续迁移更多宿主页面测试或继续拆分插件自有用例目录。
