## ADDED Requirements

### Requirement: E2E 测试发现必须支持 Host-only 与 Plugin-full 模式

E2E 测试套件 SHALL 明确区分宿主测试和源码插件自有测试。宿主测试发现和执行不得要求 `apps/lina-plugins` 存在；插件测试发现和执行在显式请求插件范围时必须要求官方插件工作区已初始化。

#### Scenario: Host-only E2E 不要求插件工作区
- **WHEN** `apps/lina-plugins` 不存在或为空
- **AND** 开发者运行宿主 E2E 范围
- **THEN** Playwright 只发现宿主归属测试
- **AND** 测试发现不因插件测试目录缺失失败

#### Scenario: 显式插件 E2E 缺少插件工作区
- **WHEN** `apps/lina-plugins` 不存在或为空
- **AND** 开发者运行 `plugins` 或 `plugin:<plugin-id>` 测试范围
- **THEN** 测试入口快速失败
- **AND** 错误提示初始化官方插件 submodule

#### Scenario: Plugin-full E2E 发现插件自有测试
- **WHEN** `apps/lina-plugins` submodule 初始化完成
- **AND** 开发者运行 `plugins` 或 `plugin:<plugin-id>` 测试范围
- **THEN** Playwright 发现对应插件的 `hack/tests/e2e/TC*.ts`
- **AND** TC 编号唯一性和目录归属验证覆盖宿主与插件测试
