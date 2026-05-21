## ADDED Requirements

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
