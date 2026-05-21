# Design

## 测试目录与资产归属

E2E 资产被划分为两个明确边界：根 `hack/tests/` 负责宿主能力、宿主插件框架和动态测试插件运行时的通用测试能力；`apps/lina-plugins/<plugin-id>/hack/tests/` 负责具体官方源码插件的自有用例、页面对象、测试数据和专属 baseline。目录组织以稳定能力边界为第一层，例如 `iam/`、`settings/`、`scheduler/`、`extension/`、`dashboard/`、`about/`，再按清晰子域使用第二层目录。这样导航从用户能力出发，而不是从历史 URL bucket 或物理源码位置出发。

根路径 E2E 不再允许耦合具体官方插件 ID、插件路径、插件菜单、插件 i18n key、插件 mock data 或插件专属 baseline。需要验证具体官方插件行为时，测试必须回收到对应插件目录，通过 `plugins` 或 `plugin:<plugin-id>` 入口选择。这个边界让 host-only 主框架环境不必默认依赖 `apps/lina-plugins`，也避免根测试资产随着插件扩张持续膨胀。

`hack/tests/e2e/` 只保留真实用例文件。共享 API helper、wait util、fixture 辅助、governance script 和 debug script 统一迁入 `fixtures/`、`support/`、`scripts/` 与 `debug/`。目录 ownership、非测试文件、根路径插件耦合和 manifest 引用都通过治理校验持续守卫。

## 编号与治理模型

早期 suite 使用全局 `TC{NNNN}` 编号，便于单一目录阶段的唯一性检查，但随着宿主模块和多个插件模块并行演进，全局编号会把无关目录耦合在一起，带来重命名冲突和长期维护成本。最终治理模型改为模块目录本地 `TC{NNN}` 连续递增：每个拥有目录从 `TC001` 开始，独立维护本目录内的编号连续性，宿主目录与各插件目录互不影响。

治理校验不再只检查“全局是否重复”，而是同时检查：目录内编号是否从 `TC001` 开始连续递增、是否仍残留旧四位全局编号、测试文件是否落在允许的能力目录、根路径是否重新引入具体插件耦合、以及 execution manifest 是否引用了失效文件或错误 scope。这使目录重构、插件迁移和后续增量迭代都具备可持续的自动守卫。

## 执行入口与环境职责

执行入口围绕“默认全量、按需快反馈、按环境分责”组织：

- `pnpm test` / `pnpm test:full`：完整回归入口。
- `pnpm test:smoke`：关键路径快速反馈。
- `pnpm test:module -- <scope>`：按 manifest scope 解析模块范围。
- `pnpm test:host`：不依赖官方插件工作区的宿主全量回归。
- `pnpm test:host:module -- <scope>`：仅执行可在 host-only 环境运行的宿主 scope。

execution manifest 是 smoke 文件、module scope、串行边界、隔离类别和 allowlist 的唯一事实来源。full 与 module 模式都采用同一套“并行池 + 串行池”拆分逻辑，module 运行不会绕过隔离规则。host-only module 在解析到 `plugins`、`plugin:<plugin-id>` 或任何需要 `apps/lina-plugins` 的 scope 时必须直接失败，避免在错误环境中产生伪失败。

plugin-full 的职责被显式收敛：根路径 scope `extension:plugin` 只负责宿主插件框架、动态测试插件运行时和通用插件治理；源码插件自有测试统一通过 `plugins` 或 `plugin:<plugin-id>` 入口选择。这样 plugin-full 不再用“重新跑一遍宿主全量套件”的方式覆盖插件能力，而是由 host-only 保持宿主基线、由插件目录维护插件闭环回归。

## CI 分片与执行证据

在本地和 CI 中都继续复用现有 runner 与 manifest 体系，不引入新的测试框架。plugin-full 通过通用 scope 进行分片：宿主插件框架 scope 独立运行，`plugins` scope 继续使用通用入口并借助 Playwright shard 拆成多个独立 job。每个分片使用独立服务实例、独立 PostgreSQL、独立 artifact 名称，失败会阻断完整 verification suite 的下游 job。

为了让 wall clock 优化可复核，runner 在 full、module 和 shard 模式下都输出并行文件数、串行文件数、worker 数量和串行隔离类别摘要；CI 继续上传 Playwright report、test-results 与前后端日志。PostgreSQL service health check 改为显式 `pg_isready -U postgres -d linapro`，避免依赖 runner OS 用户产生无效健康检查噪声。

## 认证态、页面装配与等待策略

高频已登录用例默认复用预生成 `storageState`。`adminPage` 保持兼容语义，继续代表“进入默认工作台后操作页面”；同时新增不自动导航 dashboard 的 `authenticatedPage`，让慢文件可以直接进入目标业务路由，消除重复的首页加载和二次跳转成本。登录、登出、未认证跳转、登录失败等认证场景仍显式走真实登录流程，避免被共享登录态掩盖。

等待策略统一收敛到 page object 和 shared support 层，优先使用表格装载完成、抽屉/弹窗可见、toast 完成、route 变化、dropdown/confirm overlay 状态和 API 完成作为就绪信号。固定时长等待只允许保留在有明确业务理由的位置，并需要注释说明。业务状态断言优先使用稳定 API 字段、ID、code、permission key、labelKey 或计数值；展示文案和多语言 copy 作为单独的表现层断言。

## 共享状态隔离与前置条件装配

E2E 运行时把“跨文件共享全局状态”当成一等治理对象。manifest 为插件生命周期、运行时 i18n 缓存、系统参数、公共前端配置、字典数据、权限矩阵、共享数据库种子、文件系统插件产物等风险面声明机器可读的 isolation category。validator 再通过静态启发式扫描高风险 API、helper 和关键字，发现高风险用例未被归入串行边界或未声明分类时立即失败。确有并行安全例外时，只能通过带理由的 allowlist 显式声明。

前置条件由 fixture 持有而不是由文件间隐式共享。普通插件功能用例通过 suite/shard 级幂等 baseline 统一完成插件同步、安装、启用、可用 mock 数据加载和插件投影刷新；生命周期用例仍自行控制安装、启用、禁用、卸载、上传、升级和清理状态，避免 baseline 污染被测对象。测试创建的数据必须带唯一前缀并自行清理，确保单文件独立运行。

## 生命周期覆盖策略

完整 UI 生命周期只保留给代表性官方插件与必须验证宿主壳、iframe、runtime 切换的动态插件场景，其余官方插件采用 API/contract smoke、菜单/路由挂载检查和核心页面可访问性验证，以在不丢失语义的前提下降低重复生命周期成本。对运行时缓存、ETag 与条件请求相关场景，断言关注协议语义：请求是否携带条件头，响应是否是合法的 `304`，或合法的 `200 + 新 ETag + 有效 body`。只有在测试独占资源版本时才要求固定结果码。

## 反馈驱动的稳定性修正

优化与分片之后，回归会更快暴露真实边界问题，因此设计上允许把“保证 suite 稳定运行所需的最小生产/运行时修正”纳入同一能力体系。这类修正包括：插件资源数据权限映射与宿主枚举对齐、host-only 环境对插件依赖用例的过滤、动态插件菜单投影刷新等待、Vite `/x` 代理和动态示例数据清理、角色权限抽屉异步初始化等待、动态插件上传探测改为文件路径输入、插件页面列顺序/卸载前置条件断言更新，以及动态插件热升级后当前稳定 iframe 页的可访问路由刷新保护。

同一思路也覆盖测试支撑配置的一致性修正：上传默认上限需要在配置模板、初始化 SQL、运行时 fallback 与 packed asset 中一致对齐到 100MB，运行时 cache/reconciler reason 统一收敛为命名类型常量，避免测试与生产路径对同一约束出现分叉。这样 E2E 优化不只是“跑得更快”，而是形成可持续揭示和修复真实回归问题的长期机制。
