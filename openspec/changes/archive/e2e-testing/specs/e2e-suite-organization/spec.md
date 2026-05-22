## ADDED Requirements

### Requirement: E2E 用例必须按稳定能力边界和插件归属组织
E2E 套件必须按当前稳定的工作台能力边界与插件归属组织目录结构，不能继续把大多数能力堆放在历史遗留的 catch-all 目录中。第一层目录必须表达稳定能力边界，第二层目录只用于表达清晰子域。

#### Scenario: 宿主能力用例进入匹配的能力目录
- **WHEN** 团队新增或迁移宿主能力测试文件
- **THEN** 该文件必须落在与当前工作台能力边界一致的目录中，例如 `iam/`、`settings/`、`scheduler/`、`extension/`、`dashboard/` 或 `about/`

#### Scenario: 插件能力用例进入匹配的插件目录
- **WHEN** 团队新增或迁移具体插件能力测试文件
- **THEN** 该文件必须落在表达该插件能力边界的目录中
- **AND** 具体官方源码插件自有测试必须闭环在 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`

#### Scenario: 二级目录表达子域而不是重新制造大杂烩目录
- **WHEN** 某个能力包含多个清晰子域
- **THEN** 套件可以使用二级目录表达这些子域
- **AND** 不得借此重新引入新的过载 catch-all 目录

### Requirement: 根路径 E2E 必须与具体官方源码插件解耦
根 `hack/tests` 下的 E2E 用例、fixture、support helper、runner 配置和 execution manifest 必须只承载宿主通用测试资产与插件框架通用能力，不得硬编码任何具体官方源码插件信息。

#### Scenario: 根路径测试资产不硬编码具体插件信息
- **WHEN** 开发者新增或修改根 `hack/tests` 下的 E2E 文件、配置、测试数据或 baseline
- **THEN** 这些资产不得硬编码具体官方插件 ID、插件路径、插件菜单、插件 mock data、插件 i18n key 或插件专属 baseline
- **AND** 需要验证具体源码插件行为的用例必须移动到对应插件目录

#### Scenario: host-only 环境不默认依赖插件工作区
- **WHEN** host-only E2E 或 host-only module E2E 执行
- **THEN** 它不得默认要求 `apps/lina-plugins` 已初始化
- **AND** 依赖官方插件安装、启用、路由、菜单、i18n 或 mock 数据的用例必须归入 plugin-full 范围

### Requirement: 非测试文件不得混入 E2E 用例树
`hack/tests/e2e/` 目录树只能包含真实测试用例文件。共享 helper、wait util、debug script 和 execution governance script 必须放在专用支持目录中，不得与 `TC*.ts` 混放。

#### Scenario: 共享 helper 位于支持目录
- **WHEN** 测试需要共享 API helper、wait util 或数据构造器
- **THEN** 这些文件必须位于 `fixtures/`、`support/`、`scripts/` 或等效专用目录
- **AND** 不得位于 `hack/tests/e2e/` 下

#### Scenario: debug script 不进入测试发现范围
- **WHEN** 团队新增临时调试或排障脚本
- **THEN** 该文件必须位于专用 debug 目录
- **AND** 不得进入 E2E discovery scope

### Requirement: TC 编号和目录归属必须自动校验
E2E 套件必须提供自动化盘点与校验，以保证 TC 命名、模块目录编号连续性和目录归属长期保持可治理。

#### Scenario: 宿主模块目录内编号本地递增
- **WHEN** 开发者在 `hack/tests/e2e/<module>/` 或其子目录中新增测试文件
- **THEN** 文件名必须采用 `TC{NNN}-{brief-name}.ts`
- **AND** 编号必须从当前目录的 `TC001` 开始连续递增
- **AND** 不得因为其他宿主模块或插件目录中存在更大编号而跳号

#### Scenario: 插件模块目录内编号本地递增
- **WHEN** 开发者在 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/<module>/` 中新增插件自有测试文件
- **THEN** 文件名必须采用 `TC{NNN}-{brief-name}.ts`
- **AND** 编号只在该插件的当前模块目录内连续递增
- **AND** 不得与宿主模块或其他插件模块形成编号耦合

#### Scenario: 治理校验拒绝旧全局编号和错误归属
- **WHEN** validator 扫描所有 E2E 文件
- **THEN** 它必须拒绝旧 `TC{NNNN}-*.ts` 四位全局编号文件
- **AND** 拒绝目录内重复、缺失或未从 `TC001` 开始连续递增的编号
- **AND** 报告落在错误能力目录中的测试文件

### Requirement: plugin-full 覆盖必须聚焦插件框架通用能力与插件自有用例
plugin-full E2E 必须通过宿主级通用动态插件 fixture 验证插件框架行为，并通过各插件目录下的自有用例验证官方源码插件功能。它不得再以无差别重跑宿主全量套件作为主要覆盖策略。

#### Scenario: 根路径 `extension:plugin` 只覆盖通用插件框架能力
- **WHEN** plugin-full 执行根路径通用插件 scope
- **THEN** 根 `hack/tests/e2e/extension/plugin/` 只能覆盖宿主插件框架、动态测试插件运行时和通用插件治理能力
- **AND** 官方源码插件的菜单、权限、路由、i18n、任务、mock data 或运行时资源验证必须闭环在对应插件目录

#### Scenario: host-only 继续承担宿主基线职责
- **WHEN** 完整 E2E 验证链路执行
- **THEN** host-only E2E 必须继续覆盖宿主全量能力范围
- **AND** plugin-full 不得替代 host-only 的宿主基线职责

#### Scenario: 源码插件自有用例使用通用选择入口
- **WHEN** 开发者或 CI 选择源码插件自有 E2E 范围
- **THEN** 必须使用 `plugins` 或 `plugin:<plugin-id>` 入口
- **AND** 不得重新引入按官方插件业务模块命名的长期 alias scope

### Requirement: 官方插件生命周期回归必须采用代表性完整链路与批量 smoke 组合
官方插件生命周期回归不得为每个官方插件重复执行相同的完整 UI 生命周期；当“代表性完整生命周期 + 每插件 contract smoke”能够提供等价覆盖时，必须优先采用该组合。

#### Scenario: 代表插件执行完整 UI 生命周期
- **WHEN** 官方插件生命周期回归执行
- **THEN** 至少一个代表性官方插件必须覆盖安装、启用、页面可访问、停用、卸载和菜单挂载变化的完整 UI 链路

#### Scenario: 其他官方插件执行 contract smoke
- **WHEN** 非代表官方插件参与生命周期回归
- **THEN** 测试必须验证该插件可同步、可安装、可启用、菜单或路由可挂载以及核心页面可访问
- **AND** 不得为每个插件重复完整的停用、卸载和缺页验证，除非该插件存在独有生命周期风险

#### Scenario: 独有生命周期风险保留专门覆盖
- **WHEN** 某个官方插件存在独有的安装、启用、卸载、数据保留、任务注册、权限或运行时资源风险
- **THEN** 该插件必须保留针对该风险的专门 E2E 或 API-level 回归覆盖
