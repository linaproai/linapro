## ADDED Requirements

### Requirement: Nightly image publishing must support a manual no-test entrypoint

系统 SHALL 提供一个独立的`GitHub Actions`手动 workflow，用于构建并发布`nightly`镜像。该 workflow MUST 仅通过`workflow_dispatch`触发，MUST 直接调用统一镜像发布 workflow，MUST 不依赖测试验证套件、单元测试、`E2E`测试、smoke 测试或其他前置测试 job。现有定时 nightly workflow MUST 继续保留测试门禁。

#### Scenario: 手动触发直接发布 nightly 镜像

- **WHEN** 维护者通过`GitHub Actions`手动触发 no-test nightly 镜像发布 workflow
- **THEN** workflow 直接调用统一镜像发布 workflow 构建并推送`linux/amd64`与`linux/arm64`多架构镜像
- **AND** workflow 发布日期型`nightly-<yyyymmdd>`不可变标签和`nightly`浮动标签
- **AND** workflow 不等待任何测试验证 job 完成

#### Scenario: 定时 nightly 继续受测试门禁保护

- **WHEN** 现有 nightly workflow 通过 schedule 触发
- **THEN** workflow 继续先运行共享测试验证套件
- **AND** 只有测试验证套件成功后才发布`nightly`镜像
