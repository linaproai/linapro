## ADDED Requirements

### Requirement: Release 完整 E2E 必须覆盖宿主与官方插件测试

Release 发布链路中的完整 E2E SHALL 同时执行宿主 E2E 和官方插件自有 E2E。测试入口 SHALL 使用现有 E2E 治理范围选择 `e2e` 与 `plugins`，不得只运行宿主测试后继续发布完整插件镜像。

#### Scenario: Full E2E 选择宿主和插件范围
- **WHEN** release workflow 执行完整 E2E
- **THEN** E2E runner 选择宿主 `e2e` 范围
- **AND** E2E runner 选择官方插件 `plugins` 范围
- **AND** Playwright 发现并执行 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/TC*.ts` 中的插件自有用例

#### Scenario: 插件 E2E 缺失阻止发布
- **WHEN** release workflow 执行完整 E2E
- **AND** 官方插件工作区缺失、为空或插件 E2E 范围解析为空
- **THEN** E2E 阶段失败
- **AND** release 镜像发布 job 不得执行

#### Scenario: E2E 失败证据被上传
- **WHEN** release 完整 E2E 完成或失败
- **THEN** workflow 上传 Playwright report、test-results、后端日志和前端日志
- **AND** artifact 名称能区分 release 与 nightly 执行来源

### Requirement: Release 官方插件验证必须与 Host-only 测试语义区分

E2E 测试套件 SHALL 保留 host-only 与 plugin-full 的语义区分。Release 完整插件发布 SHALL 使用 plugin-full 语义，官方插件工作区缺失时不得静默降级为 host-only E2E。

#### Scenario: Release 不降级为 Host-only E2E
- **WHEN** release workflow 执行完整 E2E
- **AND** 官方插件测试无法被发现
- **THEN** workflow 报告 plugin-full 验证失败
- **AND** workflow 不得把只运行宿主 E2E 的结果视为 release 完整验证通过

#### Scenario: Host-only 测试入口不影响 release 完整验证
- **WHEN** 仓库存在 host-only E2E 入口
- **AND** release workflow 发布完整插件镜像
- **THEN** release workflow 不得使用 host-only E2E 入口替代 plugin-full E2E
- **AND** release workflow 的测试日志应能显示插件测试范围已被选择
