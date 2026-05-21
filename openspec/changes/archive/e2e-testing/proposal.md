## Why

当前 E2E 套件已经同时承担宿主能力、插件框架和源码插件回归，但目录结构、执行入口和隔离治理长期沿用早期单体工作台阶段的约定，已无法准确表达当前稳定的能力边界和插件归属。`hack/tests/e2e/` 仍保留过载的历史分组，根目录测试资产混入非测试文件，并且一度需要通过全局递增编号、手工 scope 和人工约定维持秩序；随着插件数量和测试规模持续增长，这种方式不断放大定位成本、迁移成本和变更冲突。

回归耗时也已经成为主要交付瓶颈。host-only 模式需要重复承担登录态初始化、dashboard 首屏加载和业务页面跳转成本，plugin-full 模式则长期把插件框架通用回归、源码插件自有用例和高风险共享状态文件集中在少量串行 job 中执行。与此同时，插件生命周期、运行时 i18n 缓存、系统参数、字典、菜单权限、共享种子数据和文件系统产物等跨用例共享状态会在并行运行时产生伪失败。如果没有机器可校验的隔离分类、fixture 前置条件和执行边界，E2E 套件既无法安全扩展并行度，也无法持续吸收新的高风险回归场景。

因此，本变更把 E2E 套件统一收敛为按稳定能力边界组织、按环境职责选择、按共享状态隔离运行、按证据度量优化收益的一套治理体系，同时补齐优化过程中暴露出的宿主与插件回归稳定性缺口。

## What Changes

- 按当前稳定的工作台能力边界和插件归属重组 `hack/tests/e2e/`，并将具体官方源码插件行为、页面对象、测试数据和 baseline 迁移到对应 `apps/lina-plugins/<plugin-id>/hack/tests/` 目录；根路径 E2E 只保留宿主插件框架、动态测试插件运行时和通用插件治理资产。
- 将 `hack/tests/e2e/` 限定为仅存放真实 `TC*.ts` 用例，把 helper、wait util、debug script、governance script 移入 `fixtures/`、`support/`、`scripts/` 与 `debug/`，并由治理校验持续阻止目录漂移、错误归属和根目录插件耦合回归。
- 将测试文件编号从跨仓库全局四位递增收敛为模块目录本地 `TC{NNN}` 连续递增，并由校验脚本拒绝旧 `TC{NNNN}` 命名、目录内重复编号和断号。
- 保持 `pnpm test` 为全量回归入口，新增 `test:full`、`test:smoke`、`test:module`、`test:host` 和 `test:host:module` 等分层入口；plugin-full 统一使用 `extension:plugin`、`plugins` 与 `plugin:<plugin-id>` 通用 scope，不再维护官方插件业务别名。
- 引入预生成登录态、`authenticatedPage` 轻量 fixture、状态驱动等待工具、串行/并行两阶段调度、隔离类别清单和启发式风险检测，以降低重复登录、固定等待和串行拥塞导致的 wall clock。
- 为普通插件功能用例提供幂等 suite/shard 级 baseline，统一处理插件同步、安装、启用、mock 数据加载和投影刷新；插件生命周期用例继续显式控制自身安装、启用、禁用、卸载、上传、升级和清理状态。
- 将 plugin-full CI 收敛为基于通用 scope 的分片执行与 Playwright shard 调度，独立上传 report、test-results 与服务日志，并通过显式 PostgreSQL 用户/数据库健康检查减少无效等待和日志噪声。
- 将缓存/ETag 校验、业务状态断言和跨文件前置条件改为协议语义、稳定字段和自包含 fixture，避免因全局版本刷新、语言切换或共享数据残留导致误报。
- 针对优化与分片过程中暴露出的宿主/插件回归不稳定问题，补齐页面装配、菜单投影刷新、插件资源权限、运行时上传探测、表格列契约、卸载前置条件文案和默认上传上限等配套修正。

## Capabilities

### New Capabilities

- `e2e-suite-organization`：定义 E2E 目录边界、宿主/插件资产归属、根路径去插件耦合、非测试文件隔离和模块本地 TC 编号治理。
- `e2e-suite-execution-efficiency`：定义分层执行入口、host-only 与 plugin-full 职责边界、登录态复用、轻量认证页面、状态驱动等待、共享状态隔离、插件 baseline、CI 分片和耗时验收机制。

### Modified Capabilities

- 无。

## Breaking Changes

- 根 `hack/tests` 与根 execution manifest 不再允许硬编码具体官方源码插件 ID、插件路径、插件菜单或插件专属 baseline；相关测试资产必须迁移到对应插件目录。
- 源码插件自有用例的长期选择入口只保留 `plugins` 与 `plugin:<plugin-id>`，原按官方插件业务模块命名的 alias scope 被移除。
- E2E 文件命名从全局 `TC{NNNN}` 迁移为模块目录本地 `TC{NNN}` 连续递增，治理校验会拒绝旧编号方案。
- `test:host:module` 只允许执行不依赖 `apps/lina-plugins` 的宿主 scope，插件 scope 会被显式拒绝。

## Impact

- 影响 `hack/tests/e2e/`、`apps/lina-plugins/*/hack/tests/`、`hack/tests/config/execution-manifest.json`、`hack/tests/scripts/*`、`hack/tests/fixtures/*`、`hack/tests/support/*`、`hack/tests/pages/*`、`hack/tests/README*` 以及相关 CI workflow。
- 影响 plugin-full 与 host-only 的执行职责、artifact 结构、Playwright fixture 组合方式、验证脚本、scope 选择规则和回归证据输出。
- 为保障优化后的回归稳定性，补充了少量宿主/插件运行时支持修正，包括插件页面可访问路由刷新判定、动态插件上传探测方式、运行时 reason 常量治理以及默认上传上限与打包资产对齐。
- 不引入新的测试框架，不改变生产 API 契约或数据库 schema；除默认上传上限提升到 100MB 外，本组变更主要聚焦 E2E 基础设施、测试治理和为回归稳定性所需的配套修正。
- 本组变更未新增前端运行时语言包、插件 manifest i18n 或 apidoc i18n 资源。
