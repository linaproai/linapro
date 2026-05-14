## ADDED Requirements

### Requirement: Release workflow 必须在镜像发布前完成完整测试门禁

系统 SHALL 提供 `Release Test and Build` GitHub Actions workflow，用于替代只发布镜像的 release workflow。该 workflow 在 tag push 触发后 SHALL 至少覆盖 `Nightly Test and Build` 中的测试验证面，包括 host-only 与 plugin-full 的 Windows 命令冒烟、Go 单元测试、前端单元测试、host-only build smoke、host-only E2E、plugin-full E2E 和 Redis cluster smoke；只有所有测试 job 成功后，才允许构建并推送 release 多架构镜像。

#### Scenario: Release tag 触发完整测试后发布镜像
- **WHEN** GitHub Actions 收到 tag push 事件
- **THEN** release workflow 先并行或按依赖运行 host-only 与 plugin-full 的 Windows 命令冒烟、Go 单元测试、前端单元测试、host-only build smoke、host-only E2E、plugin-full E2E 和 Redis cluster smoke
- **AND** release 镜像发布 job 通过 `needs` 依赖所有测试 job
- **AND** 所有测试 job 成功后才执行 GHCR 登录、`make image push=1`、`latest` 标签发布和远端 manifest inspect

#### Scenario: Release 覆盖 Nightly Host-only 验证
- **WHEN** nightly workflow 维护 host-only 验证 job
- **THEN** release workflow SHALL 包含等价的 host-only Windows 命令冒烟、Go 单元测试、前端单元测试、host-only build smoke 和 host-only E2E job
- **AND** release 镜像发布 job SHALL 通过 `needs` 等待这些 host-only job 成功

#### Scenario: 任一测试失败阻止 release 镜像推送
- **WHEN** release workflow 中任一测试 job 失败、取消或超时
- **THEN** release 镜像发布 job 不得执行
- **AND** workflow 不得推送 release tag 对应镜像
- **AND** workflow 不得更新 `latest` 浮动镜像标签

#### Scenario: Release workflow 名称表达测试和构建职责
- **WHEN** 仓库维护 release 发布 workflow
- **THEN** workflow 文件名使用 `.github/workflows/release-test-and-build.yml`
- **AND** workflow 展示名称为 `Release Test and Build`
- **AND** 不再保留职责重叠的旧 `release-build.yml`

### Requirement: Release 完整插件镜像必须验证官方插件工作区

系统 SHALL 将 release 发布链路视为 plugin-full 验证路径。发布完整插件镜像前，workflow SHALL 确认官方插件工作区可用，并确保官方插件清单和官方插件测试参与验证。若官方插件工作区缺失、为空或未初始化，workflow SHALL 在镜像构建前快速失败。

#### Scenario: 官方插件工作区存在时继续发布验证
- **WHEN** release workflow checkout 完成
- **AND** `apps/lina-plugins` 包含官方插件目录与 `plugin.yaml`
- **THEN** workflow 继续执行 Go 单元测试、完整 E2E 和镜像构建前置验证
- **AND** 官方插件自有测试属于 release 验证范围

#### Scenario: 官方插件工作区缺失时快速失败
- **WHEN** release workflow 需要发布完整插件镜像
- **AND** `apps/lina-plugins` 不存在、为空或缺少官方插件清单
- **THEN** workflow 在镜像构建前失败
- **AND** 错误消息说明当前缺少官方插件工作区
- **AND** 错误消息包含初始化 submodule 的操作提示

#### Scenario: Release checkout 初始化官方插件 submodule
- **WHEN** 官方插件工作区通过 git submodule 提供
- **THEN** release workflow 的 checkout 步骤使用递归 submodule 初始化
- **AND** 后续测试和镜像构建读取同一个 `apps/lina-plugins` 工作区
