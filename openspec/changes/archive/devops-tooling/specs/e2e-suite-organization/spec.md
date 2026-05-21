## ADDED Requirements

### Requirement: Nightly 完整 E2E 必须覆盖宿主与官方插件测试

Nightly 验证链路中的完整 E2E SHALL 同时执行宿主 E2E 和官方插件自有 E2E。测试入口 SHALL 使用现有 E2E 治理范围选择 `e2e` 与 `plugins`，用于覆盖 release workflow 不执行的完整浏览器回归范围。

### Requirement: E2E 官方插件验证必须与 Host-only 测试语义区分

E2E 测试套件 SHALL 保留 host-only 与 plugin-full 的语义区分。启用完整 E2E 的 workflow SHALL 同时运行 host-only E2E 和 plugin-full E2E；plugin-full E2E 不得被 host-only E2E 替代，官方插件工作区缺失时不得静默降级为 host-only E2E。

### Requirement: Release 不运行完整 E2E

Release workflow 采用与 Main CI 一致的不含 E2E 的简要测试开关：`include-host-only-e2e-tests` SHALL 为 `false`，`include-plugin-full-e2e-tests` SHALL 为 `false`。完整 E2E 验证由 nightly workflow 覆盖。
