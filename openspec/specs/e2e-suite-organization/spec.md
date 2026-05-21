# E2E 测试套件组织规范

## Purpose

定义 Playwright E2E 测试套件的目录归属、辅助文件放置和 TC 治理规则，确保测试树与稳定的 LinaPro 能力边界保持对齐且易于维护。
## Requirements
### Requirement:E2E 测试用例必须按稳定能力边界组织
E2E 测试套件 SHALL 按当前稳定的工作台能力边界和插件归属组织测试目录。不得继续将大多数能力测试堆积到超载的遗留兜底目录中。二级目录可用于更细粒度的能力拆分，但一级目录必须仍反映稳定的能力边界。

#### Scenario:宿主归属的能力测试落入匹配的能力目录
- **当** 团队添加或迁移宿主归属的能力测试文件时
- **则** 该文件必须落入与当前工作台能力边界对齐的目录，如 `iam/`、`settings/`、`scheduler/`、`extension/`、`dashboard/` 或 `about/`

#### Scenario:插件归属的能力测试落入匹配的插件能力目录
- **当** 团队添加或迁移插件能力测试文件时
- **则** 该文件必须落入表达插件能力边界的目录，如 `monitor/operlog/`、`monitor/loginlog/`、`org/dept/` 或 `content/notice/`

#### Scenario:二级目录表达子域而非恢复遗留桶
- **当** 一个能力包含多个清晰的子域时
- **则** 测试套件可使用二级目录表达这些子域
- **且** 不得重新引入新的超载兜底目录来替代稳定的能力边界

### Requirement:非测试文件不得混入 E2E 测试树
`hack/tests/e2e/` 目录树 SHALL 仅包含真实的测试用例文件。共享辅助、等待工具、调试脚本和执行治理脚本必须位于专用支持目录中，不得与 `TC*.ts` 文件混放。

#### Scenario:共享辅助位于支持目录
- **当** 测试需要共享 API 辅助、等待工具或数据构建器时
- **则** 这些文件必须位于 `fixtures/`、`support/`、`scripts/` 或等效的专用支持目录
- **且** 不得位于 `hack/tests/e2e/` 下

#### Scenario:调试脚本不污染测试发现
- **当** 团队添加临时调试或调查脚本时
- **则** 该文件必须位于专用调试目录
- **且** 不得出现在 E2E 发现范围内

### Requirement:TC 编号和目录归属必须自动验证
E2E 测试套件 SHALL 提供自动化的清单和验证来检查 TC 命名、全局唯一性和目录归属，确保重复的 TC ID、无效文件和错放的测试不会在仓库中滞留。

#### Scenario:TC 标识符全局唯一
- **当** 验证器扫描所有 `TC*.ts` 文件时
- **则** 系统必须检测并报告任何重复的 TC 标识符

#### Scenario:无效文件自动报告
- **当** 验证器扫描 `hack/tests/e2e/` 时
- **则** 系统必须报告任何不遵循 `TC{NNNN}-{brief-name}.ts` 约定的文件
- **且** 必须报告任何位于允许的能力目录映射之外的测试文件

### Requirement: Nightly 完整 E2E 必须覆盖宿主与官方插件测试

Nightly 验证链路中的完整 E2E SHALL 同时执行宿主 E2E 和官方插件自有 E2E。测试入口 SHALL 使用现有 E2E 治理范围选择 `e2e` 与 `plugins`，用于覆盖 release workflow 不执行的完整浏览器回归范围。

#### Scenario: Nightly Full E2E 选择宿主和插件范围
- **WHEN** nightly workflow 执行完整 E2E
- **THEN** E2E runner 选择宿主 `e2e` 范围
- **AND** E2E runner 选择官方插件 `plugins` 范围
- **AND** Playwright 发现并执行 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/TC*.ts` 中的插件自有用例

#### Scenario: 插件 E2E 缺失阻止 Nightly 镜像发布
- **WHEN** nightly workflow 执行完整 E2E
- **AND** 官方插件工作区缺失、为空或插件 E2E 范围解析为空
- **THEN** E2E 阶段失败
- **AND** nightly 镜像发布 job 不得执行

#### Scenario: E2E 失败证据被上传
- **WHEN** nightly 完整 E2E 完成或失败
- **THEN** workflow 上传 Playwright report、test-results、后端日志和前端日志
- **AND** artifact 名称能区分 nightly 执行来源

### Requirement: E2E 官方插件验证必须与 Host-only 测试语义区分

E2E 测试套件 SHALL 保留 host-only 与 plugin-full 的语义区分。启用完整 E2E 的 workflow SHALL 同时运行 host-only E2E 和 plugin-full E2E；plugin-full E2E 不得被 host-only E2E 替代，官方插件工作区缺失时不得静默降级为 host-only E2E。

#### Scenario: 启用完整 E2E 时同时运行 Host-only 和 Plugin-full E2E
- **WHEN** workflow 执行完整 E2E 门禁
- **THEN** workflow SHALL 运行 host-only E2E 入口
- **AND** workflow SHALL 运行 plugin-full E2E 入口
- **AND** 下游镜像发布 job SHALL 等待两个 E2E job 均成功

#### Scenario: Plugin-full E2E 不降级为 Host-only E2E
- **WHEN** workflow 执行完整 E2E
- **AND** 官方插件测试无法被发现
- **THEN** workflow 报告 plugin-full 验证失败
- **AND** workflow 不得把只运行宿主 E2E 的结果视为完整验证通过

#### Scenario: Host-only 测试入口不影响 Plugin-full 完整验证
- **WHEN** 仓库存在 host-only E2E 入口
- **AND** workflow 执行 plugin-full E2E
- **THEN** workflow 不得使用 host-only E2E 入口替代 plugin-full E2E
- **AND** workflow 的测试日志应能显示插件测试范围已被选择

### Requirement: Plugin-full E2E 必须聚焦插件框架通用能力与插件自有用例
Plugin-full E2E SHALL validate host-level plugin framework behavior through generic dynamic-plugin fixtures and SHALL validate official source-plugin functionality only through plugin-owned tests under each plugin directory. It MUST NOT rely on indiscriminately re-running the complete host-only suite as its primary coverage strategy.

#### Scenario: Plugin-full 覆盖插件能力
- **WHEN** plugin-full E2E 执行
- **THEN** 它必须覆盖官方插件自有 E2E 测试
- **AND** 根 `hack/tests/e2e/extension/plugin/` 只能覆盖宿主插件框架、动态测试插件运行时和通用插件治理能力
- **AND** 官方源码插件的菜单、权限、路由、i18n、任务、mock data 或运行时资源验证必须闭环在 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`
- **AND** 源码插件自有 E2E 的选择入口必须保持为 `plugins` 和 `plugin:<plugin-id>`，不得按官方插件业务模块新增长期 scope

#### Scenario: Host-only 继续覆盖宿主全量能力
- **WHEN** 完整 E2E 验证链路执行
- **THEN** host-only E2E 必须继续覆盖宿主全量能力范围
- **AND** plugin-full E2E 不得替代 host-only E2E 的宿主基线职责

#### Scenario: 根 E2E 不耦合官方插件 ID
- **WHEN** 开发者新增或修改根 `hack/tests` 下的 E2E 用例、POM、support helper、runner 配置或执行 manifest
- **THEN** 这些文件不得硬编码任何具体源码插件 ID、源码插件路径、源码插件菜单、源码插件 mock data、源码插件测试数据、源码插件配置 baseline 或源码插件 i18n key
- **AND** 插件相关测试数据、插件相关配置和插件专属 baseline 必须放在对应 `apps/lina-plugins/<plugin-id>/hack/tests/` 目录
- **AND** 需要验证具体源码插件行为的用例必须移动到对应 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`
- **AND** 根 E2E 只能保留通用 runner、通用发现机制和宿主插件框架能力所需的非插件专属测试资产
- **AND** 根 E2E 治理校验必须阻止新的具体源码插件信息耦合进入根测试代码与配置

#### Scenario: 主框架宿主基线不默认依赖源码插件
- **WHEN** host-only E2E 或 host-only module E2E 执行
- **THEN** 它不得默认要求 `apps/lina-plugins` 已经初始化
- **AND** 依赖官方插件安装、启用、路由、菜单、i18n 或源码插件 mock 数据的用例必须归入 plugin-full 范围

### Requirement: E2E 测试文件编号必须按模块目录本地递增
E2E test file prefixes SHALL be scoped to the owning module directory instead of using a globally increasing sequence across the whole repository.

#### Scenario: 宿主模块内递增
- **WHEN** 开发者在 `hack/tests/e2e/<module>/` 或其子模块目录中新增测试文件
- **THEN** 文件名必须使用 `TC{NNN}-{brief-name}.ts`
- **AND** `TC` 编号必须从当前目录内的 `TC001` 开始连续递增
- **AND** 不得因为其他宿主模块或源码插件目录中已有更大的 `TC` 编号而跳号

#### Scenario: 源码插件模块内递增
- **WHEN** 开发者在 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/<module>/` 中新增插件自有测试文件
- **THEN** 文件名必须使用 `TC{NNN}-{brief-name}.ts`
- **AND** `TC` 编号必须只在该插件的当前模块目录内连续递增
- **AND** 不得与宿主模块或其他插件模块的测试编号产生耦合

#### Scenario: 治理校验阻止全局编号回归
- **WHEN** `pnpm -C hack/tests test:validate` 执行
- **THEN** 校验必须拒绝 `TC{NNNN}-*.ts` 这类旧全局四位编号文件
- **AND** 校验必须拒绝同一模块目录内重复、缺失或未从 `TC001` 开始连续递增的测试文件编号

### Requirement: 官方插件生命周期回归必须使用代表性完整链路和批量 smoke 组合
Official plugin lifecycle regression SHALL avoid repeating the same full UI lifecycle for every official plugin when a representative full lifecycle plus per-plugin contract smoke provides equivalent behavioral coverage.

#### Scenario: 代表插件执行完整 UI 生命周期
- **WHEN** 官方插件生命周期回归执行
- **THEN** 至少一个代表性官方插件必须覆盖安装、启用、页面可访问、停用、卸载和菜单挂载变化的完整 UI 链路

#### Scenario: 其他官方插件执行 contract smoke
- **WHEN** 非代表官方插件参与生命周期回归
- **THEN** 测试必须验证该插件可同步、可安装、可启用、菜单或路由可挂载、核心页面可访问
- **AND** 测试不得为每个插件重复完整 UI 停用、卸载和缺页验证，除非该插件存在独有生命周期风险

#### Scenario: 独有生命周期风险必须保留专门覆盖
- **WHEN** 某个官方插件具有独有的安装、启用、卸载、数据保留、任务注册、权限或运行时资源风险
- **THEN** 该插件必须保留针对该风险的专门 E2E 或 API-level 回归覆盖

