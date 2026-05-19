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

