## MODIFIED Requirements

### Requirement: Release workflow 必须复用共享测试模板并运行简要测试门禁

系统 SHALL 提供 `Release Test and Build` GitHub Actions workflow。该 workflow 在 tag push 触发后 SHALL 复用共享测试验证套件，并采用与 Main CI 一致的不含 E2E 的简要测试范围：host-only 与 plugin-full 的 Windows 命令冒烟、Go 单元测试、前端单元测试、插件命令冒烟、常用 make 命令冒烟、Redis integration、host-only build smoke 和 Redis cluster smoke。Release workflow 不 SHALL 运行 host-only E2E 或 plugin-full E2E。

#### Scenario: Release tag 触发简要测试后发布镜像

- **WHEN** GitHub Actions 收到 tag push 事件
- **THEN** release workflow 先完成 release tag 与 framework version 校验
- **AND** release workflow 调用共享测试验证套件
- **AND** 所有测试 job 成功后才执行 GHCR 登录、镜像构建、latest 标签发布和远端 manifest inspect
