## ADDED Requirements

### Requirement: E2E 用例必须按稳定能力边界组织目录
E2E 套件 SHALL 按当前工作台稳定能力边界与插件归属组织测试目录,不得继续把多数能力用例堆放在历史遗留的大杂烩目录中。目录允许使用二级子模块表达更细的能力切分,但首层目录必须反映稳定能力边界。

#### Scenario: 宿主能力用例落在对应能力目录
- **WHEN** 新增或迁移一个宿主能力测试文件
- **THEN** 该文件 MUST 落在与当前工作台能力边界一致的目录中,例如 `iam/`、`settings/`、`scheduler/`、`extension/`、`dashboard/`、`about/`

#### Scenario: 插件能力用例落在对应插件能力目录
- **WHEN** 新增或迁移一个插件能力测试文件
- **THEN** 该文件 MUST 落在表达插件能力边界的目录中,例如 `monitor/operlog/`、`monitor/loginlog/`、`org/dept/`、`content/notice/`

#### Scenario: 二级目录表达子域而不是重新回退到旧分组
- **WHEN** 某个能力下存在多个明显子域
- **THEN** 套件 MAY 使用二级目录表达子域,但不得为了省事重新引入新的过载总目录来替代稳定能力边界

### Requirement: 非测试用例文件不得混入 E2E 用例树
`hack/tests/e2e/` 目录树 SHALL 只承载真正的测试用例文件。共享 helper、等待工具、调试脚本和执行治理脚本必须放置在专用支持目录中,不得与 `TC*.ts` 文件混放。

#### Scenario: 共享 helper 必须位于支持目录
- **WHEN** 测试需要共享 API helper、等待工具或数据构造器
- **THEN** 这些文件 MUST 位于 `fixtures/`、`support/`、`scripts/` 或等价的专用支持目录,而不是位于 `hack/tests/e2e/` 下

#### Scenario: 调试脚本不得污染用例扫描结果
- **WHEN** 团队新增一次性调试脚本或排查脚本
- **THEN** 该文件 MUST 位于专用调试目录,并且不得出现在 E2E 用例扫描范围内

### Requirement: TC 编号与目录归属必须可自动校验
E2E 套件 SHALL 提供自动化盘点能力,用于校验 TC 文件命名、全局唯一性和目录归属是否合法,避免重复 TC ID、非法文件和错误目录长期滞留。

#### Scenario: TC 编号必须全局唯一
- **WHEN** 盘点脚本扫描全部 `TC*.ts` 文件
- **THEN** 系统 MUST 能检测并报告任何重复的 TC ID

#### Scenario: 非法文件必须被自动报告
- **WHEN** 盘点脚本扫描 `hack/tests/e2e/`
- **THEN** 系统 MUST 能报告任何不符合 `TC{NNNN}-{brief-name}.ts` 约定的文件或不在允许目录映射内的测试文件
